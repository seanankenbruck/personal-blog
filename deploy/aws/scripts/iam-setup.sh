#!/bin/bash

# IAM Setup Script for infra-manager user
# This script creates an IAM group with all required permissions for AWS infrastructure deployment

set -e  # Exit on any error

# Configuration
GROUP_NAME="infra-managers"
USER_NAME="infra-manager"
AWS_REGION="us-east-1"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

# Check if AWS CLI is configured
check_aws_cli() {
    if ! command -v aws &> /dev/null; then
        error "AWS CLI not found. Please install AWS CLI first."
    fi
    
    if ! aws sts get-caller-identity &> /dev/null; then
        error "AWS CLI not configured. Please run 'aws configure' first."
    fi
    
    log "AWS CLI check passed"
}

# Create IAM group
create_iam_group() {
    log "Creating IAM group: ${GROUP_NAME}"
    
    if aws iam get-group --group-name "${GROUP_NAME}" &> /dev/null; then
        warn "Group ${GROUP_NAME} already exists, skipping creation"
    else
        aws iam create-group --group-name "${GROUP_NAME}"
        log "Group ${GROUP_NAME} created successfully"
    fi
}

# Create and attach IAM policies
create_policies() {
    log "Creating and attaching IAM policies..."
    
    # Policy 1: EC2 and VPC
    cat > ec2-vpc-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "EC2FullAccess",
      "Effect": "Allow",
      "Action": [
        "ec2:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
    
    # Policy 2: ECS
    cat > ecs-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ECSFullAccess",
      "Effect": "Allow",
      "Action": [
        "ecs:*"
      ],
      "Resource": "*"
    },
    {
      "Sid": "ECSPassRole",
      "Effect": "Allow",
      "Action": [
        "iam:PassRole"
      ],
      "Resource": [
        "arn:aws:iam::*:role/*-execution-role",
        "arn:aws:iam::*:role/*-task-role"
      ]
    }
  ]
}
EOF
    
    # Policy 3: RDS
    cat > rds-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "RDSFullAccess",
      "Effect": "Allow",
      "Action": [
        "rds:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
    
    # Policy 4: ECR
    cat > ecr-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ECRFullAccess",
      "Effect": "Allow",
      "Action": [
        "ecr:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
    
    # Policy 5: CloudFormation
    cat > cloudformation-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "CloudFormationFullAccess",
      "Effect": "Allow",
      "Action": [
        "cloudformation:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
    
    # Policy 6: IAM (Limited)
    cat > iam-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "IAMRoleManagement",
      "Effect": "Allow",
      "Action": [
        "iam:CreateRole",
        "iam:DeleteRole",
        "iam:GetRole",
        "iam:ListRoles",
        "iam:UpdateRole",
        "iam:TagRole",
        "iam:UntagRole",
        "iam:PutRolePolicy",
        "iam:DeleteRolePolicy",
        "iam:GetRolePolicy",
        "iam:ListRolePolicies",
        "iam:AttachRolePolicy",
        "iam:DetachRolePolicy",
        "iam:ListAttachedRolePolicies",
        "iam:PassRole",
        "iam:GetAccountSummary"
      ],
      "Resource": "*"
    }
  ]
}
EOF
    
    # Policy 7: Route 53
    cat > route53-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Route53FullAccess",
      "Effect": "Allow",
      "Action": [
        "route53:*"
      ],
      "Resource": "*"
    },
    {
      "Sid": "Route53DomainsReadOnly",
      "Effect": "Allow",
      "Action": [
        "route53domains:GetDomainDetail",
        "route53domains:ListDomains"
      ],
      "Resource": "*"
    }
  ]
}
EOF
    
    # Policy 8: Certificate Manager
    cat > acm-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ACMFullAccess",
      "Effect": "Allow",
      "Action": [
        "acm:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
    
    # Policy 9: Systems Manager
    cat > ssm-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "SSMParameterStoreAccess",
      "Effect": "Allow",
      "Action": [
        "ssm:GetParameter",
        "ssm:GetParameters",
        "ssm:GetParametersByPath",
        "ssm:PutParameter",
        "ssm:DeleteParameter",
        "ssm:DescribeParameters",
        "ssm:GetParameterHistory",
        "ssm:AddTagsToResource",
        "ssm:RemoveTagsFromResource",
        "ssm:ListTagsForResource"
      ],
      "Resource": "*"
    }
  ]
}
EOF
    
    # Policy 10: Elastic Load Balancing
    cat > elb-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ELBFullAccess",
      "Effect": "Allow",
      "Action": [
        "elasticloadbalancing:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
    
    # Policy 11: CloudWatch Logs
    cat > logs-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "CloudWatchLogsAccess",
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "logs:DescribeLogGroups",
        "logs:DescribeLogStreams",
        "logs:FilterLogEvents",
        "logs:GetLogEvents",
        "logs:DeleteLogGroup",
        "logs:DeleteLogStream",
        "logs:PutRetentionPolicy"
      ],
      "Resource": "*"
    }
  ]
}
EOF
    
    # Policy 12: STS
    cat > sts-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "STSGetCallerIdentity",
      "Effect": "Allow",
      "Action": [
        "sts:GetCallerIdentity"
      ],
      "Resource": "*"
    }
  ]
}
EOF
    
    # Array of policies to create and attach
    declare -a policies=(
        "InfraManagerEC2VPC:ec2-vpc-policy.json"
        "InfraManagerECS:ecs-policy.json"
        "InfraManagerRDS:rds-policy.json"
        "InfraManagerECR:ecr-policy.json"
        "InfraManagerCloudFormation:cloudformation-policy.json"
        "InfraManagerIAM:iam-policy.json"
        "InfraManagerRoute53:route53-policy.json"
        "InfraManagerACM:acm-policy.json"
        "InfraManagerSSM:ssm-policy.json"
        "InfraManagerELB:elb-policy.json"
        "InfraManagerLogs:logs-policy.json"
        "InfraManagerSTS:sts-policy.json"
    )
    
    # Create and attach each policy
    for policy_info in "${policies[@]}"; do
        IFS=':' read -r policy_name policy_file <<< "$policy_info"
        
        log "Creating policy: ${policy_name}"
        
        # Create the policy
        aws iam create-policy \
            --policy-name "${policy_name}" \
            --policy-document "file://${policy_file}" \
            --description "Policy for infra-manager deployment automation" || warn "Policy ${policy_name} may already exist"
        
        # Get AWS account ID
        AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
        
        # Attach policy to group
        aws iam attach-group-policy \
            --group-name "${GROUP_NAME}" \
            --policy-arn "arn:aws:iam::${AWS_ACCOUNT_ID}:policy/${policy_name}" || warn "Policy ${policy_name} may already be attached"
        
        log "Policy ${policy_name} attached to group ${GROUP_NAME}"
    done
}

# Add user to group
add_user_to_group() {
    log "Adding user ${USER_NAME} to group ${GROUP_NAME}"
    
    # Check if user exists
    if ! aws iam get-user --user-name "${USER_NAME}" &> /dev/null; then
        warn "User ${USER_NAME} does not exist. Creating user..."
        aws iam create-user --user-name "${USER_NAME}"
        log "User ${USER_NAME} created"
    fi
    
    # Add user to group
    aws iam add-user-to-group \
        --group-name "${GROUP_NAME}" \
        --user-name "${USER_NAME}"
    
    log "User ${USER_NAME} added to group ${GROUP_NAME}"
}

# Create access keys for the user
create_access_keys() {
    log "Creating access keys for user ${USER_NAME}"
    
    # Check if access keys already exist
    if aws iam list-access-keys --user-name "${USER_NAME}" --query 'AccessKeyMetadata[0].AccessKeyId' --output text | grep -q "AK"; then
        warn "Access keys already exist for user ${USER_NAME}"
        log "To create new access keys, delete existing ones first:"
        log "aws iam delete-access-key --user-name ${USER_NAME} --access-key-id <existing-key-id>"
        return
    fi
    
    # Create new access keys
    ACCESS_KEY_OUTPUT=$(aws iam create-access-key --user-name "${USER_NAME}")
    
    echo
    log "=== ACCESS KEYS CREATED ==="
    echo "Access Key ID: $(echo $ACCESS_KEY_OUTPUT | jq -r '.AccessKey.AccessKeyId')"
    echo "Secret Access Key: $(echo $ACCESS_KEY_OUTPUT | jq -r '.AccessKey.SecretAccessKey')"
    echo
    warn "IMPORTANT: Save these credentials securely. The secret key will not be shown again!"
    echo
    log "Configure AWS CLI with these credentials:"
    log "aws configure --profile infra-manager"
    echo
}

# List attached policies for verification
verify_setup() {
    log "Verifying group setup..."
    
    echo
    log "Group: ${GROUP_NAME}"
    log "Attached policies:"
    aws iam list-attached-group-policies --group-name "${GROUP_NAME}" --query 'AttachedPolicies[].PolicyName' --output table
    
    echo
    log "Group members:"
    aws iam get-group --group-name "${GROUP_NAME}" --query 'Users[].UserName' --output table
}

# Cleanup function
cleanup() {
    log "Cleaning up temporary files..."
    rm -f *.json
}

# Main function
main() {
    log "Starting IAM setup for infra-manager..."
    
    check_aws_cli
    create_iam_group
    create_policies
    add_user_to_group
    create_access_keys
    verify_setup
    cleanup
    
    log "IAM setup completed successfully!"
    echo
    log "Next steps:"
    log "1. Configure AWS CLI with the new credentials"
    log "2. Test the deployment script: ./deploy-aws.sh deploy"
}

# Trap to clean up files on exit
trap cleanup EXIT

# Run main function
main "$@"