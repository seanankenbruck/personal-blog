#!/bin/bash

# AWS Infrastructure Cleanup Script
# This script removes all resources created by the deployment script
# Use with caution - this will delete everything!

set -e  # Exit on any error

# Configuration
APP_NAME="personal-blog"
ENVIRONMENT="production"
AWS_REGION="us-east-1"
AWS_PROFILE="infra-manager"
DOMAIN_NAME="ankenbruckdevops.com"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check if AWS CLI is installed
    if ! command -v aws &> /dev/null; then
        error "AWS CLI not found. Please install AWS CLI first."
        exit 1
    fi
    
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        error "jq not found. Please install jq for JSON parsing."
        exit 1
    fi
    
    # Test AWS profile
    if ! aws sts get-caller-identity --profile "${AWS_PROFILE}" --region "${AWS_REGION}" &> /dev/null; then
        error "AWS profile '${AWS_PROFILE}' not configured or invalid. Please check your credentials."
        exit 1
    fi
    
    log "Prerequisites check passed!"
}

# Confirmation prompt
confirm_deletion() {
    echo
    warn "⚠️  DANGER ZONE ⚠️"
    warn "This script will permanently delete the following AWS resources:"
    echo
    echo "  🗂️  CloudFormation Stack: ${APP_NAME}-infrastructure"
    echo "  🐳 ECS Service and Tasks"
    echo "  🗄️  RDS Database (with all data)"
    echo "  📦 ECR Repository (with all images)"
    echo "  🌐 Load Balancer and Target Groups"
    echo "  🔒 SSL Certificates (if not in use elsewhere)"
    echo "  📡 Route 53 DNS Records"
    echo "  🔑 IAM Roles created by the deployment"
    echo "  📝 Systems Manager Parameters"
    echo "  📊 CloudWatch Log Groups"
    echo
    warn "This action CANNOT be undone!"
    echo
    read -p "Are you absolutely sure you want to continue? Type 'DELETE' to confirm: " confirmation
    
    if [ "$confirmation" != "DELETE" ]; then
        log "Cleanup cancelled. No resources were deleted."
        exit 0
    fi
    
    echo
    warn "Starting cleanup in 10 seconds... Press Ctrl+C to cancel"
    sleep 10
}

# Get AWS account ID
get_account_id() {
    AWS_ACCOUNT_ID=$(aws sts get-caller-identity \
        --profile "${AWS_PROFILE}" \
        --region "${AWS_REGION}" \
        --query Account \
        --output text)
    
    if [ -z "$AWS_ACCOUNT_ID" ]; then
        error "Could not retrieve AWS account ID"
        exit 1
    fi
    
    info "AWS Account ID: ${AWS_ACCOUNT_ID}"
}

# Stop ECS service first to ensure graceful shutdown
stop_ecs_service() {
    log "Stopping ECS service..."
    
    # Get cluster name from CloudFormation if it exists
    CLUSTER_NAME=$(aws cloudformation describe-stacks \
        --profile "${AWS_PROFILE}" \
        --region "${AWS_REGION}" \
        --stack-name "${APP_NAME}-infrastructure" \
        --query 'Stacks[0].Outputs[?OutputKey==`ECSClusterName`].OutputValue' \
        --output json \
        --no-cli-pager 2>/dev/null || echo "[]")
    
    if [ ! -z "$CLUSTER_NAME" ] && [ "$CLUSTER_NAME" != "None" ]; then
        log "Found ECS cluster: ${CLUSTER_NAME}"
        
        # Scale service to 0
        aws ecs update-service \
            --profile "${AWS_PROFILE}" \
            --region "${AWS_REGION}" \
            --cluster "${CLUSTER_NAME}" \
            --service "${APP_NAME}-service" \
            --desired-count 0 \
            2>/dev/null || warn "ECS service may not exist or already stopped"
        
        log "Waiting for ECS tasks to stop..."
        sleep 30
        
        # Delete the service
        aws ecs delete-service \
            --profile "${AWS_PROFILE}" \
            --region "${AWS_REGION}" \
            --cluster "${CLUSTER_NAME}" \
            --service "${APP_NAME}-service" \
            --force \
            2>/dev/null || warn "ECS service may already be deleted"
        
        log "ECS service deletion initiated"
    else
        warn "ECS cluster not found or already deleted"
    fi
}

# Clean up ECR repository
cleanup_ecr_repository() {
    log "Cleaning up ECR repository..."
    
    # List all images in the repository
    IMAGES=$(aws ecr list-images \
        --profile "${AWS_PROFILE}" \
        --region "${AWS_REGION}" \
        --repository-name "${APP_NAME}" \
        --query 'imageIds' \
        --output json \
        --no-cli-pager 2>/dev/null || echo "[]")
    
    if [ "$IMAGES" != "[]" ] && [ "$IMAGES" != "" ]; then
        log "Deleting all images from ECR repository..."
        aws ecr batch-delete-image \
            --profile "${AWS_PROFILE}" \
            --region "${AWS_REGION}" \
            --repository-name "${APP_NAME}" \
            --image-ids "$IMAGES" \
            --no-cli-pager 2>/dev/null || warn "Could not delete some ECR images"
    fi
    
    # The ECR repository itself will be deleted by CloudFormation
    log "ECR repository cleanup completed"
}

# Remove Route 53 DNS records
cleanup_route53_records() {
    log "Cleaning up Route 53 DNS records..."
    
    # Get hosted zone ID for the domain
    HOSTED_ZONE_ID=$(aws route53 list-hosted-zones \
        --profile "${AWS_PROFILE}" \
        --query "HostedZones[?Name=='${DOMAIN_NAME}.'].Id" \
        --output text 2>/dev/null | sed 's|/hostedzone/||' || echo "")
    
    if [ ! -z "$HOSTED_ZONE_ID" ] && [ "$HOSTED_ZONE_ID" != "None" ]; then
        log "Found hosted zone ID: ${HOSTED_ZONE_ID}"
        
        # Get ALB DNS name if it exists
        ALB_DNS_NAME=$(aws cloudformation describe-stacks \
            --profile "${AWS_PROFILE}" \
            --region "${AWS_REGION}" \
            --stack-name "${APP_NAME}-infrastructure" \
            --query 'Stacks[0].Outputs[?OutputKey==`ALBDNSName`].OutputValue' \
            --output text 2>/dev/null || echo "")
        
        if [ ! -z "$ALB_DNS_NAME" ] && [ "$ALB_DNS_NAME" != "None" ]; then
            # Delete the A record pointing to ALB
            cat > delete-dns-record.json << EOF
{
  "Changes": [
    {
      "Action": "DELETE",
      "ResourceRecordSet": {
        "Name": "${DOMAIN_NAME}",
        "Type": "A",
        "AliasTarget": {
          "DNSName": "${ALB_DNS_NAME}",
          "EvaluateTargetHealth": false,
          "HostedZoneId": "$(aws elbv2 describe-load-balancers \
            --profile "${AWS_PROFILE}" \
            --region "${AWS_REGION}" \
            --names "${APP_NAME}-alb" \
            --query 'LoadBalancers[0].CanonicalHostedZoneId' \
            --output text 2>/dev/null || echo "Z35SXDOTRQ7X7K")"
        }
      }
    }
  ]
}
EOF
            
            aws route53 change-resource-record-sets \
                --profile "${AWS_PROFILE}" \
                --hosted-zone-id "${HOSTED_ZONE_ID}" \
                --change-batch file://delete-dns-record.json \
                2>/dev/null || warn "Could not delete DNS record (may not exist)"
            
            rm -f delete-dns-record.json
            log "Route 53 DNS record deletion initiated"
        else
            warn "ALB DNS name not found, skipping DNS record deletion"
        fi
    else
        warn "Hosted zone not found for ${DOMAIN_NAME}"
    fi
}

# Remove SSL certificates (only if not in use)
cleanup_ssl_certificates() {
    log "Checking SSL certificates..."
    
    # List certificates for the domain
    CERT_ARNS=$(aws acm list-certificates \
        --profile "${AWS_PROFILE}" \
        --region "${AWS_REGION}" \
        --query "CertificateSummaryList[?DomainName=='${DOMAIN_NAME}'].CertificateArn" \
        --output text 2>/dev/null || echo "")
    
    if [ ! -z "$CERT_ARNS" ]; then
        for CERT_ARN in $CERT_ARNS; do
            # Check if certificate is in use
            IN_USE=$(aws acm describe-certificate \
                --profile "${AWS_PROFILE}" \
                --region "${AWS_REGION}" \
                --certificate-arn "$CERT_ARN" \
                --query 'Certificate.InUseBy' \
                --output text 2>/dev/null || echo "")
            
            if [ "$IN_USE" == "None" ] || [ -z "$IN_USE" ]; then
                log "Deleting unused SSL certificate: ${CERT_ARN}"
                aws acm delete-certificate \
                    --profile "${AWS_PROFILE}" \
                    --region "${AWS_REGION}" \
                    --certificate-arn "$CERT_ARN" \
                    2>/dev/null || warn "Could not delete certificate (may be in use)"
            else
                warn "SSL certificate ${CERT_ARN} is in use, skipping deletion"
            fi
        done
    else
        info "No SSL certificates found for ${DOMAIN_NAME}"
    fi
}

# Delete Systems Manager parameters
cleanup_ssm_parameters() {
    log "Cleaning up Systems Manager parameters..."
    
    # Delete database password parameter
    aws ssm delete-parameter \
        --profile "${AWS_PROFILE}" \
        --region "${AWS_REGION}" \
        --name "/${APP_NAME}/${ENVIRONMENT}/db-password" \
        2>/dev/null || warn "Database password parameter may not exist"
    
    # Delete any other parameters under the app path
    PARAMETERS=$(aws ssm get-parameters-by-path \
        --profile "${AWS_PROFILE}" \
        --region "${AWS_REGION}" \
        --path "/${APP_NAME}/${ENVIRONMENT}" \
        --query 'Parameters[].Name' \
        --output text 2>/dev/null || echo "")
    
    if [ ! -z "$PARAMETERS" ]; then
        for PARAM in $PARAMETERS; do
            aws ssm delete-parameter \
                --profile "${AWS_PROFILE}" \
                --region "${AWS_REGION}" \
                --name "$PARAM" \
                2>/dev/null || warn "Could not delete parameter: $PARAM"
            log "Deleted parameter: $PARAM"
        done
    fi
    
    log "SSM parameters cleanup completed"
}

# Delete CloudWatch log groups
cleanup_cloudwatch_logs() {
    log "Cleaning up CloudWatch log groups..."
    
    # Delete ECS log group
    aws logs delete-log-group \
        --profile "${AWS_PROFILE}" \
        --region "${AWS_REGION}" \
        --log-group-name "/ecs/${APP_NAME}" \
        2>/dev/null || warn "ECS log group may not exist"
    
    # Find and delete other related log groups
    LOG_GROUPS=$(aws logs describe-log-groups \
        --profile "${AWS_PROFILE}" \
        --region "${AWS_REGION}" \
        --log-group-name-prefix "${APP_NAME}" \
        --query 'logGroups[].logGroupName' \
        --output text 2>/dev/null || echo "")
    
    if [ ! -z "$LOG_GROUPS" ]; then
        for LOG_GROUP in $LOG_GROUPS; do
            aws logs delete-log-group \
                --profile "${AWS_PROFILE}" \
                --region "${AWS_REGION}" \
                --log-group-name "$LOG_GROUP" \
                2>/dev/null || warn "Could not delete log group: $LOG_GROUP"
            log "Deleted log group: $LOG_GROUP"
        done
    fi
    
    log "CloudWatch logs cleanup completed"
}

# Delete IAM roles created by the deployment
cleanup_iam_roles() {
    log "Cleaning up IAM roles..."
    
    # Delete execution role
    EXECUTION_ROLE="${APP_NAME}-execution-role"
    
    # Detach policies from execution role
    aws iam detach-role-policy \
        --profile "${AWS_PROFILE}" \
        --role-name "$EXECUTION_ROLE" \
        --policy-arn "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy" \
        2>/dev/null || warn "Could not detach policy from execution role"
    
    # Delete inline policies
    aws iam delete-role-policy \
        --profile "${AWS_PROFILE}" \
        --role-name "$EXECUTION_ROLE" \
        --policy-name "SSMParameterAccess" \
        2>/dev/null || warn "Could not delete inline policy from execution role"
    
    # Delete the role
    aws iam delete-role \
        --profile "${AWS_PROFILE}" \
        --role-name "$EXECUTION_ROLE" \
        2>/dev/null || warn "Execution role may not exist"
    
    log "IAM roles cleanup completed"
}

# Delete CloudFormation stack (this will delete most resources)
delete_cloudformation_stack() {
    log "Deleting CloudFormation stack..."
    
    # Check if stack exists
    if aws cloudformation describe-stacks \
        --profile "${AWS_PROFILE}" \
        --region "${AWS_REGION}" \
        --stack-name "${APP_NAME}-infrastructure" &> /dev/null; then
        
        log "Stack ${APP_NAME}-infrastructure found, initiating deletion..."
        
        aws cloudformation delete-stack \
            --profile "${AWS_PROFILE}" \
            --region "${AWS_REGION}" \
            --stack-name "${APP_NAME}-infrastructure"
        
        log "Waiting for CloudFormation stack deletion to complete..."
        log "This may take 10-15 minutes..."
        
        # Wait for stack deletion with timeout
        aws cloudformation wait stack-delete-complete \
            --profile "${AWS_PROFILE}" \
            --region "${AWS_REGION}" \
            --stack-name "${APP_NAME}-infrastructure" || warn "Stack deletion may have timed out or failed"
        
        log "CloudFormation stack deletion completed"
    else
        warn "CloudFormation stack ${APP_NAME}-infrastructure not found"
    fi
}

# Verify cleanup completion
verify_cleanup() {
    log "Verifying cleanup completion..."
    
    # Check if CloudFormation stack still exists
    if aws cloudformation describe-stacks \
        --profile "${AWS_PROFILE}" \
        --region "${AWS_REGION}" \
        --stack-name "${APP_NAME}-infrastructure" &> /dev/null; then
        warn "CloudFormation stack still exists. Deletion may still be in progress."
    else
        log "✅ CloudFormation stack deleted"
    fi
    
    # Check ECS cluster
    if aws ecs describe-clusters \
        --profile "${AWS_PROFILE}" \
        --region "${AWS_REGION}" \
        --clusters "${APP_NAME}-cluster" \
        --query 'clusters[?status==`ACTIVE`]' \
        --output text 2>/dev/null | grep -q "ACTIVE"; then
        warn "ECS cluster still active"
    else
        log "✅ ECS cluster deleted"
    fi
    
    # Check ECR repository
    if aws ecr describe-repositories \
        --profile "${AWS_PROFILE}" \
        --region "${AWS_REGION}" \
        --repository-names "${APP_NAME}" &> /dev/null; then
        warn "ECR repository still exists"
    else
        log "✅ ECR repository deleted"
    fi
    
    log "Cleanup verification completed"
}

# Display cleanup summary
display_summary() {
    echo
    log "🧹 Cleanup Summary"
    echo "=================="
    echo
    log "The following resources have been cleaned up:"
    echo "  ✅ ECS Service and Tasks"
    echo "  ✅ ECR Repository and Images" 
    echo "  ✅ CloudFormation Stack (VPC, Subnets, Security Groups, etc.)"
    echo "  ✅ RDS Database"
    echo "  ✅ Application Load Balancer"
    echo "  ✅ Route 53 DNS Records"
    echo "  ✅ SSL Certificates (if unused)"
    echo "  ✅ Systems Manager Parameters"
    echo "  ✅ CloudWatch Log Groups"
    echo "  ✅ IAM Execution Roles"
    echo
    info "Note: Some resources may take additional time to fully delete."
    info "You can check the AWS Console to verify all resources are removed."
    echo
    log "Cleanup completed successfully! 🎉"
}

# Show what would be deleted (dry run)
dry_run() {
    log "DRY RUN MODE - Showing what would be deleted:"
    echo
    
    # Check CloudFormation stack
    if aws cloudformation describe-stacks \
        --profile "${AWS_PROFILE}" \
        --region "${AWS_REGION}" \
        --stack-name "${APP_NAME}-infrastructure" &> /dev/null; then
        echo "📦 CloudFormation Stack: ${APP_NAME}-infrastructure"
        
        # Show stack resources
        aws cloudformation list-stack-resources \
            --profile "${AWS_PROFILE}" \
            --region "${AWS_REGION}" \
            --stack-name "${APP_NAME}-infrastructure" \
            --query 'StackResourceSummaries[].{Type:ResourceType,Status:ResourceStatus}' \
            --output table 2>/dev/null || echo "  Could not list stack resources"
    else
        echo "📦 CloudFormation Stack: Not found"
    fi
    
    echo
    
    # Check ECR repository
    if aws ecr describe-repositories \
        --profile "${AWS_PROFILE}" \
        --region "${AWS_REGION}" \
        --repository-names "${APP_NAME}" &> /dev/null; then
        echo "🐳 ECR Repository: ${APP_NAME}"
        
        IMAGE_COUNT=$(aws ecr list-images \
            --profile "${AWS_PROFILE}" \
            --region "${AWS_REGION}" \
            --repository-name "${APP_NAME}" \
            --query 'length(imageIds)' \
            --output text 2>/dev/null || echo "0")
        echo "   └── Images: ${IMAGE_COUNT}"
    else
        echo "🐳 ECR Repository: Not found"
    fi
    
    echo
    
    # Check IAM roles
    if aws iam get-role \
        --profile "${AWS_PROFILE}" \
        --role-name "${APP_NAME}-execution-role" &> /dev/null; then
        echo "🔑 IAM Role: ${APP_NAME}-execution-role"
    else
        echo "🔑 IAM Role: Not found"
    fi
    
    echo
    
    # Check SSM parameters
    PARAM_COUNT=$(aws ssm get-parameters-by-path \
        --profile "${AWS_PROFILE}" \
        --region "${AWS_REGION}" \
        --path "/${APP_NAME}/${ENVIRONMENT}" \
        --query 'length(Parameters)' \
        --output text 2>/dev/null || echo "0")
    echo "📝 SSM Parameters: ${PARAM_COUNT} parameters under /${APP_NAME}/${ENVIRONMENT}/"
    
    echo
    log "End of dry run. Use 'clean-infra.sh delete' to actually delete these resources."
}

# Main execution
main() {
    case "${1:-help}" in
        "delete")
            log "Starting AWS infrastructure cleanup..."
            check_prerequisites
            get_account_id
            confirm_deletion
            stop_ecs_service
            cleanup_ecr_repository
            cleanup_route53_records
            cleanup_ssl_certificates
            cleanup_ssm_parameters
            cleanup_cloudwatch_logs
            cleanup_iam_roles
            delete_cloudformation_stack
            verify_cleanup
            display_summary
            ;;
        "dry-run"|"preview")
            log "Running cleanup dry run..."
            check_prerequisites
            get_account_id
            dry_run
            ;;
        "help"|*)
            echo "AWS Infrastructure Cleanup Script"
            echo
            echo "Usage: $0 [command]"
            echo
            echo "Commands:"
            echo "  delete     - Delete all AWS infrastructure (DESTRUCTIVE!)"
            echo "  dry-run    - Show what would be deleted without deleting"
            echo "  preview    - Same as dry-run"
            echo "  help       - Show this help message"
            echo
            echo "Configuration:"
            echo "  App Name:    ${APP_NAME}"
            echo "  Environment: ${ENVIRONMENT}"
            echo "  AWS Region:  ${AWS_REGION}"
            echo "  AWS Profile: ${AWS_PROFILE}"
            echo "  Domain:      ${DOMAIN_NAME}"
            echo
            warn "⚠️  The 'delete' command will permanently destroy all resources!"
            warn "⚠️  Always run 'dry-run' first to see what will be deleted."
            ;;
    esac
}

# Run main function with all arguments
main "$@"