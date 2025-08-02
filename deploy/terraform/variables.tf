# Terraform variables for blog application

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"

  validation {
    condition     = contains(["development", "staging", "production"], var.environment)
    error_message = "Environment must be one of: development, staging, production."
  }
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "Availability zones"
  type        = list(string)
  default     = ["us-west-2a", "us-west-2b", "us-west-2c"]
}

variable "cluster_name" {
  description = "EKS cluster name"
  type        = string
  default     = "blog-cluster"
}

variable "node_groups" {
  description = "EKS node groups configuration"
  type = map(object({
    instance_types = list(string)
    capacity_type  = string
    min_size       = number
    max_size       = number
    desired_size   = number
  }))
  default = {
    general = {
      instance_types = ["t3.medium"]
      capacity_type  = "ON_DEMAND"
      min_size       = 1
      max_size       = 3
      desired_size   = 2
    }
  }
}

variable "database_name" {
  description = "Database name"
  type        = string
  default     = "blog_db"
}

variable "database_username" {
  description = "Database username"
  type        = string
  default     = "blog_user"
}

variable "database_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

variable "database_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
}

variable "domain_name" {
  description = "Domain name for the application"
  type        = string
}

variable "jwt_secret" {
  description = "JWT secret key"
  type        = string
  sensitive   = true
}

variable "smtp_host" {
  description = "SMTP host"
  type        = string
  default     = "smtp.gmail.com"
}

variable "smtp_port" {
  description = "SMTP port"
  type        = string
  default     = "587"
}

variable "smtp_username" {
  description = "SMTP username"
  type        = string
  sensitive   = true
}

variable "smtp_password" {
  description = "SMTP password"
  type        = string
  sensitive   = true
}

# Tags
variable "common_tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default = {
    Project     = "blog"
    ManagedBy   = "terraform"
    Environment = "production"
  }
}