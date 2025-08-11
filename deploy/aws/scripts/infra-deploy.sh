#!/bin/bash

# AWS Blog Deployment Script
# This script deploys the Go blog application to AWS using ECS, RDS, and ALB

set -e  # Exit on any error

# Configuration
APP_NAME="personal-blog"
ENVIRONMENT="production"
AWS_REGION="us-east-1"
DOMAIN_NAME="ankenbruckdevops.com"
DB_NAME="blog_db"
DB_USERNAME="blogadmin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging function
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

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check if AWS CLI is installed and configured
    if ! command -v aws &> /dev/null; then
        error "AWS CLI not found. Please install and configure AWS CLI first."
    fi
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        error "Docker not found. Please install Docker first."
    fi
    
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        error "jq not found. Please install jq for JSON parsing."
    fi
    
    # Test AWS credentials
    if ! aws sts get-caller-identity &> /dev/null; then
        error "AWS credentials not configured or invalid. Please run 'aws configure' first."
    fi
    
    # Validate region has enough AZs
    AZ_COUNT=$(aws ec2 describe-availability-zones \
        --region "${AWS_REGION}" \
        --query 'AvailabilityZones[?State==`available`]' \
        --output json 2>/dev/null | jq length 2>/dev/null || echo "0")
    
    if [ "$AZ_COUNT" -lt 2 ]; then
        error "Region ${AWS_REGION} only has ${AZ_COUNT} availability zone(s). This deployment requires at least 2 AZs."
    fi
    
    log "Prerequisites check passed!"
}

# Generate random password for database
generate_db_password() {
    openssl rand -base64 32 | tr -d "=+/" | cut -c1-25
}

# Check if region has enough availability zones
check_availability_zones() {
    log "Checking availability zones in region ${AWS_REGION}..."
    
    # Get list of available AZs
    AZ_COUNT=$(aws ec2 describe-availability-zones \
        --region "${AWS_REGION}" \
        --query 'AvailabilityZones[?State==`available`]' \
        --output json | jq length)
    
    if [ "$AZ_COUNT" -lt 2 ]; then
        error "Region ${AWS_REGION} only has ${AZ_COUNT} availability zone(s). This deployment requires at least 2 AZs."
        log "Please choose a different region with at least 2 availability zones."
        log "Regions with multiple AZs include: us-east-1, us-west-2, eu-west-1, ap-southeast-1"
        exit 1
    fi
    
    log "Region ${AWS_REGION} has ${AZ_COUNT} availability zones ✓"
}

# Create infrastructure CloudFormation stack
create_infrastructure() {
    log "Creating infrastructure with CloudFormation..."
    
    # Check AZ availability first
    check_availability_zones
    
    DB_PASSWORD=$(generate_db_password)
    
    cat > infrastructure.yaml << 'EOF'
AWSTemplateFormatVersion: '2010-09-09'
Description: 'Infrastructure for Sean Blog Application'

Parameters:
  AppName:
    Type: String
    Default: sean-blog
  Environment:
    Type: String
    Default: production
  DomainName:
    Type: String
    Default: ankenbruckdevops.com
  DBPassword:
    Type: String
    NoEcho: true
    Description: Database password
  AZ1:
    Type: String
    Default: us-east-1a
    AllowedValues:
      - us-east-1a
      - us-east-1b
  AZ2:
    Type: String
    Default: us-east-1c
    AllowedValues:
      - us-east-1c
      - us-east-1d

Resources:
  # VPC and Networking
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      EnableDnsHostnames: true
      EnableDnsSupport: true
      Tags:
        - Key: Name
          Value: !Sub "${AppName}-vpc"

  PublicSubnet1:
    Type: AWS::EC2::Subnet
    Properties:
      VpcId: !Ref VPC
      CidrBlock: 10.0.1.0/24
      AvailabilityZone: !Ref AZ1
      MapPublicIpOnLaunch: true
      Tags:
        - Key: Name
          Value: !Sub "${AppName}-public-subnet-1"

  PublicSubnet2:
    Type: AWS::EC2::Subnet
    Properties:
      VpcId: !Ref VPC
      CidrBlock: 10.0.2.0/24
      AvailabilityZone: !Ref AZ2
      MapPublicIpOnLaunch: true
      Tags:
        - Key: Name
          Value: !Sub "${AppName}-public-subnet-2"

  PrivateSubnet1:
    Type: AWS::EC2::Subnet
    Properties:
      VpcId: !Ref VPC
      CidrBlock: 10.0.3.0/24
      AvailabilityZone: !Ref AZ1
      Tags:
        - Key: Name
          Value: !Sub "${AppName}-private-subnet-1"

  PrivateSubnet2:
    Type: AWS::EC2::Subnet
    Properties:
      VpcId: !Ref VPC
      CidrBlock: 10.0.4.0/24
      AvailabilityZone: !Ref AZ2
      Tags:
        - Key: Name
          Value: !Sub "${AppName}-private-subnet-2"

  InternetGateway:
    Type: AWS::EC2::InternetGateway
    Properties:
      Tags:
        - Key: Name
          Value: !Sub "${AppName}-igw"

  AttachGateway:
    Type: AWS::EC2::VPCGatewayAttachment
    Properties:
      VpcId: !Ref VPC
      InternetGatewayId: !Ref InternetGateway

  PublicRouteTable:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
        - Key: Name
          Value: !Sub "${AppName}-public-rt"

  PublicRoute:
    Type: AWS::EC2::Route
    DependsOn: AttachGateway
    Properties:
      RouteTableId: !Ref PublicRouteTable
      DestinationCidrBlock: 0.0.0.0/0
      GatewayId: !Ref InternetGateway

  PublicSubnetRouteTableAssociation1:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      SubnetId: !Ref PublicSubnet1
      RouteTableId: !Ref PublicRouteTable

  PublicSubnetRouteTableAssociation2:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      SubnetId: !Ref PublicSubnet2
      RouteTableId: !Ref PublicRouteTable

  # Security Groups
  ALBSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Security group for Application Load Balancer
      VpcId: !Ref VPC
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 80
          ToPort: 80
          CidrIp: 0.0.0.0/0
        - IpProtocol: tcp
          FromPort: 443
          ToPort: 443
          CidrIp: 0.0.0.0/0
      Tags:
        - Key: Name
          Value: !Sub "${AppName}-alb-sg"

  ECSSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Security group for ECS tasks
      VpcId: !Ref VPC
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 8080
          ToPort: 8080
          SourceSecurityGroupId: !Ref ALBSecurityGroup
      Tags:
        - Key: Name
          Value: !Sub "${AppName}-ecs-sg"

  RDSSecurityGroup:
    Type: AWS::EC2::SecurityGroup
    Properties:
      GroupDescription: Security group for RDS database
      VpcId: !Ref VPC
      SecurityGroupIngress:
        - IpProtocol: tcp
          FromPort: 5432
          ToPort: 5432
          SourceSecurityGroupId: !Ref ECSSecurityGroup
      Tags:
        - Key: Name
          Value: !Sub "${AppName}-rds-sg"

  # RDS Database
  DBSubnetGroup:
    Type: AWS::RDS::DBSubnetGroup
    Properties:
      DBSubnetGroupDescription: Subnet group for RDS database
      SubnetIds:
        - !Ref PrivateSubnet1
        - !Ref PrivateSubnet2
      Tags:
        - Key: Name
          Value: !Sub "${AppName}-db-subnet-group"

  Database:
    Type: AWS::RDS::DBInstance
    Properties:
      DBInstanceIdentifier: !Sub "${AppName}-db"
      DBInstanceClass: db.t3.micro
      Engine: postgres
      EngineVersion: '16.9'
      MasterUsername: blogadmin
      MasterUserPassword: !Ref DBPassword
      DBName: blog
      AllocatedStorage: 20
      StorageType: gp2
      VPCSecurityGroups:
        - !Ref RDSSecurityGroup
      DBSubnetGroupName: !Ref DBSubnetGroup
      BackupRetentionPeriod: 7
      MultiAZ: false
      PubliclyAccessible: false
      StorageEncrypted: true
      Tags:
        - Key: Name
          Value: !Sub "${AppName}-database"

  # ECS Cluster
  ECSCluster:
    Type: AWS::ECS::Cluster
    Properties:
      ClusterName: !Sub "${AppName}-cluster"
      CapacityProviders:
        - FARGATE
        - FARGATE_SPOT
      DefaultCapacityProviderStrategy:
        - CapacityProvider: FARGATE
          Weight: 1

  # Application Load Balancer
  ApplicationLoadBalancer:
    Type: AWS::ElasticLoadBalancingV2::LoadBalancer
    Properties:
      Name: !Sub "${AppName}-alb"
      Scheme: internet-facing
      Type: application
      Subnets:
        - !Ref PublicSubnet1
        - !Ref PublicSubnet2
      SecurityGroups:
        - !Ref ALBSecurityGroup
      Tags:
        - Key: Name
          Value: !Sub "${AppName}-alb"

  TargetGroup:
    Type: AWS::ElasticLoadBalancingV2::TargetGroup
    Properties:
      Name: !Sub "${AppName}-targets"
      Port: 8080
      Protocol: HTTP
      VpcId: !Ref VPC
      TargetType: ip
      HealthCheckPath: /
      HealthCheckProtocol: HTTP
      HealthCheckIntervalSeconds: 30
      HealthCheckTimeoutSeconds: 5
      HealthyThresholdCount: 2
      UnhealthyThresholdCount: 3
      Tags:
        - Key: Name
          Value: !Sub "${AppName}-targets"

  HTTPListener:
    Type: AWS::ElasticLoadBalancingV2::Listener
    Properties:
      DefaultActions:
        - Type: forward
          TargetGroupArn: !Ref TargetGroup
      LoadBalancerArn: !Ref ApplicationLoadBalancer
      Port: 80
      Protocol: HTTP

  # ECR Repository
  ECRRepository:
    Type: AWS::ECR::Repository
    Properties:
      RepositoryName: !Sub "${AppName}"
      ImageScanningConfiguration:
        ScanOnPush: true
      LifecyclePolicy:
        LifecyclePolicyText: |
          {
            "rules": [
              {
                "rulePriority": 1,
                "description": "Keep last 10 images",
                "selection": {
                  "tagStatus": "any",
                  "countType": "imageCountMoreThan",
                  "countNumber": 10
                },
                "action": {
                  "type": "expire"
                }
              }
            ]
          }

Outputs:
  VPCId:
    Description: VPC ID
    Value: !Ref VPC
    Export:
      Name: !Sub "${AWS::StackName}-VPC-ID"

  PublicSubnet1Id:
    Description: Public Subnet 1 ID
    Value: !Ref PublicSubnet1
    Export:
      Name: !Sub "${AWS::StackName}-PublicSubnet1-ID"

  PublicSubnet2Id:
    Description: Public Subnet 2 ID
    Value: !Ref PublicSubnet2
    Export:
      Name: !Sub "${AWS::StackName}-PublicSubnet2-ID"

  PrivateSubnet1Id:
    Description: Private Subnet 1 ID
    Value: !Ref PrivateSubnet1
    Export:
      Name: !Sub "${AWS::StackName}-PrivateSubnet1-ID"

  PrivateSubnet2Id:
    Description: Private Subnet 2 ID
    Value: !Ref PrivateSubnet2
    Export:
      Name: !Sub "${AWS::StackName}-PrivateSubnet2-ID"

  ECSClusterName:
    Description: ECS Cluster Name
    Value: !Ref ECSCluster
    Export:
      Name: !Sub "${AWS::StackName}-ECSCluster-Name"

  ALBDNSName:
    Description: Application Load Balancer DNS Name
    Value: !GetAtt ApplicationLoadBalancer.DNSName
    Export:
      Name: !Sub "${AWS::StackName}-ALB-DNSName"

  TargetGroupArn:
    Description: Target Group ARN
    Value: !Ref TargetGroup
    Export:
      Name: !Sub "${AWS::StackName}-TargetGroup-ARN"

  ECRRepositoryURI:
    Description: ECR Repository URI
    Value: !Sub "${AWS::AccountId}.dkr.ecr.${AWS::Region}.amazonaws.com/${ECRRepository}"
    Export:
      Name: !Sub "${AWS::StackName}-ECR-URI"

  DatabaseEndpoint:
    Description: RDS Database Endpoint
    Value: !GetAtt Database.Endpoint.Address
    Export:
      Name: !Sub "${AWS::StackName}-DB-Endpoint"

  ECSSecurityGroupId:
    Description: ECS Security Group ID
    Value: !Ref ECSSecurityGroup
    Export:
      Name: !Sub "${AWS::StackName}-ECS-SG-ID"
EOF

    # Deploy the CloudFormation stack
    log "Deploying CloudFormation stack..."
    aws cloudformation deploy \
        --template-file infrastructure.yaml \
        --stack-name "${APP_NAME}-infrastructure" \
        --parameter-overrides \
            AppName="${APP_NAME}" \
            Environment="${ENVIRONMENT}" \
            DomainName="${DOMAIN_NAME}" \
            DBPassword="${DB_PASSWORD}" \
        --capabilities CAPABILITY_IAM \
        --region "${AWS_REGION}"
    
    log "Infrastructure stack deployed successfully!"
    
    # Store database password in AWS Systems Manager Parameter Store
    aws ssm put-parameter \
        --name "/${APP_NAME}/${ENVIRONMENT}/db-password" \
        --value "${DB_PASSWORD}" \
        --type "SecureString" \
        --overwrite \
        --region "${AWS_REGION}"
    
    log "Database password stored in Parameter Store"
}

# Build and push Docker image
build_and_push_image() {
    log "Building and pushing Docker image..."

    # Store (date +%Y%m%d-%H%M%S) for image tag
    IMAGE_TAG=$(date +%Y%m%d-%H%M%S)
    
    # Get ECR repository URI from CloudFormation stack
    ECR_URI=$(aws cloudformation describe-stacks \
        --stack-name "${APP_NAME}-infrastructure" \
        --query 'Stacks[0].Outputs[?OutputKey==`ECRRepositoryURI`].OutputValue' \
        --output text \
        --region "${AWS_REGION}")
    
    if [ -z "$ECR_URI" ]; then
        error "Could not retrieve ECR repository URI from CloudFormation stack"
    fi
    
    # Get AWS account ID
    AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
    
    # Login to ECR
    aws ecr get-login-password --region "${AWS_REGION}" | \
        docker login --username AWS --password-stdin "${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
    
    # Build the image
    log "Building Docker image..."
    docker build -t "${APP_NAME}:latest" .
    
    # Tag the image
    docker tag "${APP_NAME}:latest" "${ECR_URI}:latest"
    docker tag "${APP_NAME}:latest" "${ECR_URI}:${IMAGE_TAG}"
    
    # Push the image
    log "Pushing Docker image to ECR..."
    docker push "${ECR_URI}:latest"
    docker push "${ECR_URI}:${IMAGE_TAG}"
    
    log "Docker image pushed successfully!"
}

# Create ECS task definition and service
deploy_ecs_service() {
    log "Deploying ECS service..."
    
    # Get values from CloudFormation stack
    ECR_URI=$(aws cloudformation describe-stacks \
        --stack-name "${APP_NAME}-infrastructure" \
        --query 'Stacks[0].Outputs[?OutputKey==`ECRRepositoryURI`].OutputValue' \
        --output text \
        --region "${AWS_REGION}")
    
    DB_ENDPOINT=$(aws cloudformation describe-stacks \
        --stack-name "${APP_NAME}-infrastructure" \
        --query 'Stacks[0].Outputs[?OutputKey==`DatabaseEndpoint`].OutputValue' \
        --output text \
        --region "${AWS_REGION}")
    
    CLUSTER_NAME=$(aws cloudformation describe-stacks \
        --stack-name "${APP_NAME}-infrastructure" \
        --query 'Stacks[0].Outputs[?OutputKey==`ECSClusterName`].OutputValue' \
        --output text \
        --region "${AWS_REGION}")
    
    PRIVATE_SUBNET_1=$(aws cloudformation describe-stacks \
        --stack-name "${APP_NAME}-infrastructure" \
        --query 'Stacks[0].Outputs[?OutputKey==`PrivateSubnet1Id`].OutputValue' \
        --output text \
        --region "${AWS_REGION}")
    
    PRIVATE_SUBNET_2=$(aws cloudformation describe-stacks \
        --stack-name "${APP_NAME}-infrastructure" \
        --query 'Stacks[0].Outputs[?OutputKey==`PrivateSubnet2Id`].OutputValue' \
        --output text \
        --region "${AWS_REGION}")
    
    ECS_SECURITY_GROUP=$(aws cloudformation describe-stacks \
        --stack-name "${APP_NAME}-infrastructure" \
        --query 'Stacks[0].Outputs[?OutputKey==`ECSSecurityGroupId`].OutputValue' \
        --output text \
        --region "${AWS_REGION}")
    
    TARGET_GROUP_ARN=$(aws cloudformation describe-stacks \
        --stack-name "${APP_NAME}-infrastructure" \
        --query 'Stacks[0].Outputs[?OutputKey==`TargetGroupArn`].OutputValue' \
        --output text \
        --region "${AWS_REGION}")
    
    # Create IAM role for ECS task execution
    cat > trust-policy.json << 'EOF'
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

    # Create execution role if it doesn't exist
    if ! aws iam get-role --role-name "${APP_NAME}-execution-role" &> /dev/null; then
        aws iam create-role \
            --role-name "${APP_NAME}-execution-role" \
            --assume-role-policy-document file://trust-policy.json \
            --region "${AWS_REGION}"
        
        aws iam attach-role-policy \
            --role-name "${APP_NAME}-execution-role" \
            --policy-arn arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy \
            --region "${AWS_REGION}"
        
        # Add SSM permissions for accessing database password
        cat > ssm-policy.json << EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ssm:GetParameter",
        "ssm:GetParameters"
      ],
      "Resource": "arn:aws:ssm:${AWS_REGION}:*:parameter/${APP_NAME}/${ENVIRONMENT}/*"
    }
  ]
}
EOF
        
        aws iam put-role-policy \
            --role-name "${APP_NAME}-execution-role" \
            --policy-name "SSMParameterAccess" \
            --policy-document file://ssm-policy.json
    fi
    
    # Get AWS account ID for role ARN
    AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
    
    # Create task definition
    cat > task-definition.json << EOF
{
  "family": "${APP_NAME}-task",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::${AWS_ACCOUNT_ID}:role/${APP_NAME}-execution-role",
  "containerDefinitions": [
    {
      "name": "${APP_NAME}",
      "image": "${ECR_URI}:latest",
      "essential": true,
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "DB_HOST",
          "value": "${DB_ENDPOINT}"
        },
        {
          "name": "DB_PORT",
          "value": "5432"
        },
        {
          "name": "DB_USER",
          "value": "${DB_USERNAME}"
        },
        {
          "name": "DB_NAME",
          "value": "${DB_NAME}"
        },
        {
          "name": "SERVER_PORT",
          "value": "8080"
        }
      ],
      "secrets": [
        {
          "name": "DB_PASSWORD",
          "valueFrom": "arn:aws:ssm:${AWS_REGION}:${AWS_ACCOUNT_ID}:parameter/${APP_NAME}/${ENVIRONMENT}/db-password"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/${APP_NAME}",
          "awslogs-region": "${AWS_REGION}",
          "awslogs-stream-prefix": "ecs",
          "awslogs-create-group": "true"
        }
      }
    }
  ]
}
EOF
    
    # Register task definition
    aws ecs register-task-definition \
        --cli-input-json file://task-definition.json \
        --region "${AWS_REGION}"
    
    # Create ECS service
    aws ecs create-service \
        --cluster "${CLUSTER_NAME}" \
        --service-name "${APP_NAME}-service" \
        --task-definition "${APP_NAME}-task" \
        --desired-count 2 \
        --launch-type FARGATE \
        --network-configuration "awsvpcConfiguration={subnets=[${PRIVATE_SUBNET_1},${PRIVATE_SUBNET_2}],securityGroups=[${ECS_SECURITY_GROUP}],assignPublicIp=DISABLED}" \
        --load-balancers "targetGroupArn=${TARGET_GROUP_ARN},containerName=${APP_NAME},containerPort=8080" \
        --region "${AWS_REGION}" || warn "Service may already exist, continuing..."
    
    log "ECS service deployed successfully!"
    
    # Clean up temporary files
    rm -f trust-policy.json ssm-policy.json task-definition.json
}

# Configure SSL certificate
setup_ssl_certificate() {
    log "Setting up SSL certificate..."
    
    # Check if certificate already exists
    CERT_ARN=$(aws acm list-certificates \
        --region "${AWS_REGION}" \
        --query "CertificateSummaryList[?DomainName=='${DOMAIN_NAME}'].CertificateArn" \
        --output text)
    
    if [ -z "$CERT_ARN" ]; then
        log "Requesting SSL certificate for ${DOMAIN_NAME}..."
        CERT_ARN=$(aws acm request-certificate \
            --domain-name "${DOMAIN_NAME}" \
            --subject-alternative-names "*.${DOMAIN_NAME}" \
            --validation-method DNS \
            --region "${AWS_REGION}" \
            --query CertificateArn \
            --output text)
        
        warn "SSL certificate requested. You need to validate it via DNS before HTTPS will work."
        warn "Certificate ARN: ${CERT_ARN}"
        warn "Please check AWS Certificate Manager console to complete DNS validation."
    else
        log "SSL certificate already exists: ${CERT_ARN}"
    fi
    
    # Get ALB ARN
    ALB_ARN=$(aws elbv2 describe-load-balancers \
        --names "${APP_NAME}-alb" \
        --region "${AWS_REGION}" \
        --query 'LoadBalancers[0].LoadBalancerArn' \
        --output text)
    
    # Check if HTTPS listener already exists
    HTTPS_LISTENER=$(aws elbv2 describe-listeners \
        --load-balancer-arn "${ALB_ARN}" \
        --region "${AWS_REGION}" \
        --query 'Listeners[?Port==`443`].ListenerArn' \
        --output text)
    
    if [ -z "$HTTPS_LISTENER" ]; then
        # Get target group ARN
        TARGET_GROUP_ARN=$(aws cloudformation describe-stacks \
            --stack-name "${APP_NAME}-infrastructure" \
            --query 'Stacks[0].Outputs[?OutputKey==`TargetGroupArn`].OutputValue' \
            --output text \
            --region "${AWS_REGION}")
        
        # Create HTTPS listener
        aws elbv2 create-listener \
            --load-balancer-arn "${ALB_ARN}" \
            --protocol HTTPS \
            --port 443 \
            --certificates CertificateArn="${CERT_ARN}" \
            --default-actions Type=forward,TargetGroupArn="${TARGET_GROUP_ARN}" \
            --region "${AWS_REGION}"
        
        log "HTTPS listener created successfully!"
    else
        log "HTTPS listener already exists"
    fi
}

# Setup Route 53 DNS
setup_dns() {
    log "Setting up Route 53 DNS..."
    
    # Get hosted zone ID for the domain
    HOSTED_ZONE_ID=$(aws route53 list-hosted-zones \
        --query "HostedZones[?Name=='${DOMAIN_NAME}.'].Id" \
        --output text | sed 's|/hostedzone/||')
    
    if [ -z "$HOSTED_ZONE_ID" ]; then
        warn "No hosted zone found for ${DOMAIN_NAME}. Please create one manually in Route 53."
        return
    fi
    
    # Get ALB DNS name
    ALB_DNS_NAME=$(aws cloudformation describe-stacks \
        --stack-name "${APP_NAME}-infrastructure" \
        --query 'Stacks[0].Outputs[?OutputKey==`ALBDNSName`].OutputValue' \
        --output text \
        --region "${AWS_REGION}")
    
    echo "URLs:"
    echo "  Load Balancer: http://${ALB_DNS_NAME}"
    echo "  Domain: https://${DOMAIN_NAME}"
    echo
    echo "AWS Resources:"
    echo "  ECS Cluster: ${APP_NAME}-cluster"
    echo "  RDS Database: ${APP_NAME}-db"
    echo "  ECR Repository: ${APP_NAME}"
    echo
    echo "Next Steps:"
    echo "1. Validate SSL certificate in AWS Certificate Manager console"
    echo "2. Ensure Route 53 hosted zone is properly configured"
    echo "3. Test the application functionality"
    echo "4. Set up monitoring and alerting"
    echo
    echo "Useful Commands:"
    echo "  View ECS service: aws ecs describe-services --cluster ${APP_NAME}-cluster --services ${APP_NAME}-service"
    echo "  View logs: aws logs describe-log-groups --log-group-name-prefix /ecs/${APP_NAME}"
    echo "  Update service: aws ecs update-service --cluster ${APP_NAME}-cluster --service ${APP_NAME}-service --force-new-deployment"
    echo
}

# Update existing deployment
update_deployment() {
    log "Updating existing deployment..."
    
    # Build and push new image
    build_and_push_image
    
    # Force new deployment
    CLUSTER_NAME=$(aws cloudformation describe-stacks \
        --stack-name "${APP_NAME}-infrastructure" \
        --query 'Stacks[0].Outputs[?OutputKey==`ECSClusterName`].OutputValue' \
        --output text \
        --region "${AWS_REGION}")
    
    aws ecs update-service \
        --cluster "${CLUSTER_NAME}" \
        --service "${APP_NAME}-service" \
        --force-new-deployment \
        --region "${AWS_REGION}"
    
    log "Deployment update initiated!"
    
    # Wait for deployment to complete
    log "Waiting for deployment to complete..."
    aws ecs wait services-stable \
        --cluster "${CLUSTER_NAME}" \
        --services "${APP_NAME}-service" \
        --region "${AWS_REGION}"
    
    log "Deployment update completed successfully!"
}

# Cleanup function
cleanup_deployment() {
    log "Cleaning up deployment..."
    
    # Delete ECS service
    CLUSTER_NAME=$(aws cloudformation describe-stacks \
        --stack-name "${APP_NAME}-infrastructure" \
        --query 'Stacks[0].Outputs[?OutputKey==`ECSClusterName`].OutputValue' \
        --output text \
        --region "${AWS_REGION}" 2>/dev/null || true)
    
    if [ ! -z "$CLUSTER_NAME" ]; then
        log "Stopping ECS service..."
        aws ecs update-service \
            --cluster "${CLUSTER_NAME}" \
            --service "${APP_NAME}-service" \
            --desired-count 0 \
            --region "${AWS_REGION}" 2>/dev/null || true
        
        aws ecs delete-service \
            --cluster "${CLUSTER_NAME}" \
            --service "${APP_NAME}-service" \
            --force \
            --region "${AWS_REGION}" 2>/dev/null || true
    fi
    
    # Delete CloudFormation stack
    log "Deleting CloudFormation stack..."
    aws cloudformation delete-stack \
        --stack-name "${APP_NAME}-infrastructure" \
        --region "${AWS_REGION}"
    
    # Delete IAM role
    aws iam detach-role-policy \
        --role-name "${APP_NAME}-execution-role" \
        --policy-arn arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy \
        --region "${AWS_REGION}" 2>/dev/null || true
    
    aws iam delete-role-policy \
        --role-name "${APP_NAME}-execution-role" \
        --policy-name "SSMParameterAccess" \
        --region "${AWS_REGION}" 2>/dev/null || true
    
    aws iam delete-role \
        --role-name "${APP_NAME}-execution-role" \
        --region "${AWS_REGION}" 2>/dev/null || true
    
    # Delete SSM parameter
    aws ssm delete-parameter \
        --name "/${APP_NAME}/${ENVIRONMENT}/db-password" \
        --region "${AWS_REGION}" 2>/dev/null || true
    
    log "Cleanup initiated. CloudFormation stack deletion in progress..."
}

# Show logs function
show_logs() {
    log "Fetching recent application logs..."
    
    aws logs describe-log-groups \
        --log-group-name-prefix "/ecs/${APP_NAME}" \
        --region "${AWS_REGION}" \
        --query 'logGroups[].logGroupName' \
        --output text | while read LOG_GROUP; do
        
        if [ ! -z "$LOG_GROUP" ]; then
            log "Logs from ${LOG_GROUP}:"
            aws logs filter-log-events \
                --log-group-name "$LOG_GROUP" \
                --start-time $(date -d '1 hour ago' +%s)000 \
                --region "${AWS_REGION}" \
                --query 'events[].message' \
                --output text | tail -20
            echo
        fi
    done
}

# Main execution
main() {
    case "${1:-deploy}" in
        "deploy")
            log "Starting full AWS deployment..."
            check_prerequisites
            #create_infrastructure
            build_and_push_image
            #deploy_ecs_service
            #setup_ssl_certificate
            #setup_dns
            #health_check
            #display_info
            ;;
        "update")
            log "Starting deployment update..."
            check_prerequisites
            update_deployment
            health_check
            ;;
        "cleanup")
            log "Starting cleanup..."
            cleanup_deployment
            ;;
        "logs")
            log "Showing application logs..."
            show_logs
            ;;
        "status")
            log "Checking deployment status..."
            CLUSTER_NAME=$(aws cloudformation describe-stacks \
                --stack-name "${APP_NAME}-infrastructure" \
                --query 'Stacks[0].Outputs[?OutputKey==`ECSClusterName`].OutputValue' \
                --output text \
                --region "${AWS_REGION}" 2>/dev/null || echo "Not found")
            
            if [ "$CLUSTER_NAME" != "Not found" ]; then
                aws ecs describe-services \
                    --cluster "${CLUSTER_NAME}" \
                    --services "${APP_NAME}-service" \
                    --region "${AWS_REGION}" \
                    --query 'services[0].{ServiceName:serviceName,Status:status,RunningCount:runningCount,DesiredCount:desiredCount,TaskDefinition:taskDefinition}'
            else
                log "Deployment not found"
            fi
            ;;
        *)
            echo "Usage: $0 [deploy|update|cleanup|logs|status]"
            echo
            echo "Commands:"
            echo "  deploy  - Full deployment (default)"
            echo "  update  - Update existing deployment with new code"
            echo "  cleanup - Remove all AWS resources"
            echo "  logs    - Show recent application logs"
            echo "  status  - Show deployment status"
            exit 1
            ;;
    esac
}

# Trap to clean up temporary files on exit
trap 'rm -f infrastructure.yaml task-definition.json trust-policy.json ssm-policy.json dns-record.json' EXIT

# Run main function with all arguments
main "$@"
        --output text \
        --region "${AWS_REGION}")
    
    ALB_HOSTED_ZONE_ID=$(aws elbv2 describe-load-balancers \
        --names "${APP_NAME}-alb" \
        --region "${AWS_REGION}" \
        --query 'LoadBalancers[0].CanonicalHostedZoneId' \
        --output text)
    
    # Create Route 53 record
    cat > dns-record.json << EOF
{
  "Changes": [
    {
      "Action": "UPSERT",
      "ResourceRecordSet": {
        "Name": "${DOMAIN_NAME}",
        "Type": "A",
        "AliasTarget": {
          "DNSName": "${ALB_DNS_NAME}",
          "EvaluateTargetHealth": false,
          "HostedZoneId": "${ALB_HOSTED_ZONE_ID}"
        }
      }
    }
  ]
}
EOF
    
    aws route53 change-resource-record-sets \
        --hosted-zone-id "${HOSTED_ZONE_ID}" \
        --change-batch file://dns-record.json
    
    log "DNS record created successfully!"
    rm -f dns-record.json
}

# Health check function
health_check() {
    log "Performing health check..."
    
    ALB_DNS_NAME=$(aws cloudformation describe-stacks \
        --stack-name "${APP_NAME}-infrastructure" \
        --query 'Stacks[0].Outputs[?OutputKey==`ALBDNSName`].OutputValue' \
        --output text \
        --region "${AWS_REGION}")
    
    log "Waiting for application to be healthy..."
    for i in {1..30}; do
        if curl -f "http://${ALB_DNS_NAME}" &> /dev/null; then
            log "Application is healthy!"
            break
        fi
        log "Health check attempt $i/30 failed, waiting 10 seconds..."
        sleep 10
    done
    
    log "Application URL: http://${ALB_DNS_NAME}"
    log "Domain URL: https://${DOMAIN_NAME} (after SSL validation)"
}

# Display deployment info
display_info() {
    log "Deployment completed successfully!"
    echo
    echo "=== Deployment Information ==="
    echo "Application Name: ${APP_NAME}"
    echo "Environment: ${ENVIRONMENT}"
    echo "AWS Region: ${AWS_REGION}"
    echo
    
    ALB_DNS_NAME=$(aws cloudformation describe-stacks \
        --stack-name "${APP_NAME}-infrastructure" \
        --query 'Stacks[0].Outputs[?OutputKey==`ALBDNSName`].OutputValue' \