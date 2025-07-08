# Terraform configuration for blog application infrastructure
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }

  backend "s3" {
    bucket = "blog-terraform-state"
    key    = "blog/terraform.tfstate"
    region = "us-west-2"
  }
}

# Configure AWS Provider
provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "blog"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}

# VPC and Networking
module "vpc" {
  source = "./modules/vpc"

  environment = var.environment
  vpc_cidr    = var.vpc_cidr
  azs         = var.availability_zones
}

# EKS Cluster
module "eks" {
  source = "./modules/eks"

  environment        = var.environment
  cluster_name       = var.cluster_name
  vpc_id            = module.vpc.vpc_id
  subnet_ids        = module.vpc.private_subnet_ids
  node_groups       = var.node_groups
  depends_on        = [module.vpc]
}

# RDS PostgreSQL
module "rds" {
  source = "./modules/rds"

  environment     = var.environment
  vpc_id          = module.vpc.vpc_id
  subnet_ids      = module.vpc.database_subnet_ids
  security_groups = [module.vpc.database_security_group_id]

  db_name     = var.database_name
  db_username = var.database_username
  db_password = var.database_password
  db_instance_class = var.database_instance_class
}

# Application Load Balancer
module "alb" {
  source = "./modules/alb"

  environment = var.environment
  vpc_id      = module.vpc.vpc_id
  subnet_ids  = module.vpc.public_subnet_ids

  domain_name = var.domain_name
  certificate_arn = module.acm.certificate_arn
}

# ACM Certificate
module "acm" {
  source = "./modules/acm"

  domain_name = var.domain_name
  environment = var.environment
}

# Route53 DNS
module "route53" {
  source = "./modules/route53"

  domain_name = var.domain_name
  alb_dns_name = module.alb.alb_dns_name
  alb_zone_id  = module.alb.alb_zone_id
}

# Kubernetes Provider
provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_ca_certificate)
  token                  = data.aws_eks_cluster_auth.cluster.token
}

provider "helm" {
  kubernetes {
    host                   = module.eks.cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks.cluster_ca_certificate)
    token                  = data.aws_eks_cluster_auth.cluster.token
  }
}

# Data sources
data "aws_eks_cluster_auth" "cluster" {
  name = module.eks.cluster_name
}

# Kubernetes namespaces
resource "kubernetes_namespace" "blog" {
  metadata {
    name = "blog"
    labels = {
      name = "blog"
      app  = "blog"
      environment = var.environment
    }
  }
}

# Kubernetes ConfigMap
resource "kubernetes_config_map" "blog_config" {
  metadata {
    name      = "blog-config"
    namespace = kubernetes_namespace.blog.metadata[0].name
  }

  data = {
    DATABASE_HOST = module.rds.db_endpoint
    DATABASE_PORT = "5432"
    DATABASE_NAME = var.database_name
    SMTP_HOST     = var.smtp_host
    SMTP_PORT     = var.smtp_port
  }

  depends_on = [kubernetes_namespace.blog]
}

# Kubernetes Secret
resource "kubernetes_secret" "blog_secrets" {
  metadata {
    name      = "blog-secrets"
    namespace = kubernetes_namespace.blog.metadata[0].name
  }

  data = {
    DATABASE_USER     = base64encode(var.database_username)
    DATABASE_PASSWORD = base64encode(var.database_password)
    JWT_SECRET        = base64encode(var.jwt_secret)
    SMTP_USERNAME     = base64encode(var.smtp_username)
    SMTP_PASSWORD     = base64encode(var.smtp_password)
  }

  depends_on = [kubernetes_namespace.blog]
}