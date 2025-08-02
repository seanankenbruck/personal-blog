# Makefile for Blog Application
.PHONY: help build run test clean docker-build docker-run deploy-staging deploy-production terraform-init terraform-plan terraform-apply

# Variables
APP_NAME := blog
DOCKER_IMAGE := seanankenbruck/blog
VERSION := $(shell git describe --tags --always --dirty)
REGISTRY := docker.io

# Colors
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(GREEN)Blog Application Makefile$(NC)"
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'

# Development
build: ## Build the Go application
	@echo "$(GREEN)Building application...$(NC)"
	go build -o bin/$(APP_NAME) ./cmd/main.go

run: ## Run the application locally
	@echo "$(GREEN)Running application...$(NC)"
	go run ./cmd/main.go

test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	@echo "$(GREEN)Running linter...$(NC)"
	golangci-lint run

clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning build artifacts...$(NC)"
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker
docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t $(DOCKER_IMAGE):$(VERSION) -f deploy/docker/Dockerfile .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

docker-push: ## Push Docker image to registry
	@echo "$(GREEN)Pushing Docker image...$(NC)"
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_IMAGE):latest

docker-run: ## Run application with Docker Compose
	@echo "$(GREEN)Running with Docker Compose...$(NC)"
	cd deploy/docker && docker-compose up -d

docker-stop: ## Stop Docker Compose services
	@echo "$(GREEN)Stopping Docker Compose services...$(NC)"
	cd deploy/docker && docker-compose down

docker-logs: ## Show Docker Compose logs
	@echo "$(GREEN)Showing Docker Compose logs...$(NC)"
	cd deploy/docker && docker-compose logs -f

# Kubernetes
k8s-deploy: ## Deploy to Kubernetes
	@echo "$(GREEN)Deploying to Kubernetes...$(NC)"
	./deploy/scripts/deploy.sh

k8s-deploy-staging: ## Deploy to staging environment
	@echo "$(GREEN)Deploying to staging...$(NC)"
	./deploy/scripts/deploy.sh staging

k8s-deploy-production: ## Deploy to production environment
	@echo "$(GREEN)Deploying to production...$(NC)"
	./deploy/scripts/deploy.sh production

k8s-status: ## Check Kubernetes deployment status
	@echo "$(GREEN)Checking deployment status...$(NC)"
	kubectl get pods -n blog
	kubectl get services -n blog
	kubectl get ingress -n blog

k8s-logs: ## Show application logs
	@echo "$(GREEN)Showing application logs...$(NC)"
	kubectl logs -f deployment/blog-deployment -n blog

k8s-delete: ## Delete Kubernetes deployment
	@echo "$(GREEN)Deleting Kubernetes deployment...$(NC)"
	kubectl delete namespace blog

# Terraform
terraform-init: ## Initialize Terraform
	@echo "$(GREEN)Initializing Terraform...$(NC)"
	cd deploy/terraform && terraform init

terraform-plan: ## Plan Terraform changes
	@echo "$(GREEN)Planning Terraform changes...$(NC)"
	cd deploy/terraform && terraform plan

terraform-apply: ## Apply Terraform changes
	@echo "$(GREEN)Applying Terraform changes...$(NC)"
	cd deploy/terraform && terraform apply -auto-approve

terraform-destroy: ## Destroy Terraform infrastructure
	@echo "$(GREEN)Destroying Terraform infrastructure...$(NC)"
	cd deploy/terraform && terraform destroy -auto-approve

# Database
db-migrate: ## Run database migrations
	@echo "$(GREEN)Running database migrations...$(NC)"
	./deploy/scripts/migrate.sh

db-backup: ## Backup database
	@echo "$(GREEN)Backing up database...$(NC)"
	./deploy/scripts/backup.sh

# Monitoring
monitoring-start: ## Start monitoring stack
	@echo "$(GREEN)Starting monitoring stack...$(NC)"
	cd deploy/monitoring && docker-compose up -d

monitoring-stop: ## Stop monitoring stack
	@echo "$(GREEN)Stopping monitoring stack...$(NC)"
	cd deploy/monitoring && docker-compose down

# Security
security-scan: ## Run security scan
	@echo "$(GREEN)Running security scan...$(NC)"
	trivy image $(DOCKER_IMAGE):latest

# CI/CD
ci-build: ## Build for CI/CD
	@echo "$(GREEN)Building for CI/CD...$(NC)"
	./deploy/scripts/build.sh $(VERSION)

ci-deploy: ## Deploy for CI/CD
	@echo "$(GREEN)Deploying for CI/CD...$(NC)"
	./deploy/scripts/deploy.sh production

# Utilities
generate-secrets: ## Generate secure secrets
	@echo "$(GREEN)Generating secure secrets...$(NC)"
	@echo "JWT_SECRET=$$(openssl rand -base64 32)"
	@echo "DATABASE_PASSWORD=$$(openssl rand -base64 32)"
	@echo "SESSION_SECRET=$$(openssl rand -base64 32)"

check-deps: ## Check dependencies
	@echo "$(GREEN)Checking dependencies...$(NC)"
	go mod tidy
	go mod verify

update-deps: ## Update dependencies
	@echo "$(GREEN)Updating dependencies...$(NC)"
	go get -u ./...
	go mod tidy

# Development workflow
dev-setup: ## Setup development environment
	@echo "$(GREEN)Setting up development environment...$(NC)"
	go mod download
	cp deploy/configs/development.env .env
	@echo "$(YELLOW)Please update .env with your configuration$(NC)"

dev: ## Start development environment
	@echo "$(GREEN)Starting development environment...$(NC)"
	docker-compose -f deploy/docker/docker-compose.yml up -d postgres redis
	@echo "$(YELLOW)Waiting for services to be ready...$(NC)"
	sleep 10
	go run ./cmd/main.go

# Production workflow
prod-deploy: docker-build docker-push k8s-deploy-production ## Full production deployment
	@echo "$(GREEN)Production deployment completed!$(NC)"

# Staging workflow
staging-deploy: docker-build docker-push k8s-deploy-staging ## Full staging deployment
	@echo "$(GREEN)Staging deployment completed!$(NC)"