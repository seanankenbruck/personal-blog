# Makefile for Personal Blog Kubernetes Deployment

# Configuration
NAMESPACE := blog-app
IMAGE_NAME := personal-blog
TAG := latest
REGISTRY :=
DOCKER_NAMESPACE := smankenb

# Derived variables
ifeq ($(REGISTRY),)
	FULL_IMAGE := $(DOCKER_NAMESPACE)/$(IMAGE_NAME):$(TAG)
else
	FULL_IMAGE := $(REGISTRY)/$(DOCKER_NAMESPACE)/$(IMAGE_NAME):$(TAG)
endif

.PHONY: help setup build deploy undeploy update health-check clean logs

## Show this help message
help:
	@echo 'Personal Blog Kubernetes Deployment'
	@echo ''
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  %-20s %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

## Setup initial environment and directory structure
setup:
	@echo "🔧 Setting up deployment environment..."
	@chmod +x deploy/scripts/*.sh
	@./deploy/scripts/setup-environment.sh
	@echo "✅ Setup completed!"

## Build local binary
build-local:
	@echo "🔨 Building local binary..."
	@go build -o blog cmd/main.go
	@echo "✅ Binary built: ./blog"

## Build and push Docker image
build:
	@echo "🏗️ Building Docker image: $(FULL_IMAGE)"
	@chmod +x deploy/scripts/build-image.sh
	@./deploy/scripts/build-image.sh $(TAG)

## Deploy application to Kubernetes
deploy:
	@echo "🚀 Deploying to Kubernetes..."
	@chmod +x deploy/scripts/deploy.sh
	@./deploy/scripts/deploy.sh

## Undeploy application from Kubernetes
undeploy:
	@echo "🗑️ Removing deployment..."
	@chmod +x deploy/scripts/undeploy.sh
	@./deploy/scripts/undeploy.sh

## Update application with new image
update:
	@echo "🔄 Updating application..."
	@chmod +x deploy/scripts/update-app.sh
	@./deploy/scripts/update-app.sh $(TAG)

## Perform health check on deployed application
health-check:
	@echo "🏥 Performing health check..."
	@chmod +x deploy/scripts/health-check.sh
	@./deploy/scripts/health-check.sh

## Show application logs
logs:
	@echo "📋 Fetching application logs..."
	@kubectl logs -f deployment/blog-app -n $(NAMESPACE)

## Show deployment status
status:
	@echo "📊 Deployment Status:"
	@echo "Namespace: $(NAMESPACE)"
	@kubectl get all -n $(NAMESPACE)
	@echo ""
	@echo "🌐 Ingress:"
	@kubectl get ingress -n $(NAMESPACE)

## Port forward for local access (for testing)
port-forward:
	@echo "🔗 Setting up port forwarding..."
	@echo "Access your blog at: http://localhost:8080"
	@kubectl port-forward service/blog-service 8080:80 -n $(NAMESPACE)

## Open a shell in the application pod
shell:
	@echo "🐚 Opening shell in application pod..."
	@kubectl exec -it deployment/blog-app -n $(NAMESPACE) -- /bin/sh

## View content directory in pod
content-shell:
	@echo "📁 Opening content directory..."
	@kubectl exec -it deployment/blog-app -n $(NAMESPACE) -- ls -la /content/posts

## Backup content files
backup-content:
	@echo "💾 Creating content backup..."
	@mkdir -p backups
	@kubectl exec deployment/blog-app -n $(NAMESPACE) -- tar -czf /tmp/content-backup.tar.gz -C /content .
	@kubectl cp $(NAMESPACE)/deployment/blog-app:/tmp/content-backup.tar.gz backups/content-backup-$(shell date +%Y%m%d-%H%M%S).tar.gz
	@echo "✅ Content backup created in backups/ directory"

## Scale application
scale:
	@if [ -z "$(REPLICAS)" ]; then \
		echo "❌ Please specify REPLICAS=number"; \
		exit 1; \
	fi
	@echo "⚖️ Scaling application to $(REPLICAS) replicas..."
	@kubectl scale deployment/blog-app --replicas=$(REPLICAS) -n $(NAMESPACE)

## Clean up all resources
clean: undeploy
	@echo "🧹 Cleaning up all resources..."
	@echo "✅ Cleanup completed"

## Generate secrets from .env file
generate-secrets:
	@echo "🔐 Generating secrets from .env file..."
	@if [ ! -f .env ]; then \
		echo "❌ .env file not found. Copy .env.example and fill in values."; \
		exit 1; \
	fi
	@./deploy/scripts/generate-secrets.sh

## Apply only configuration changes
config-update:
	@echo "⚙️ Updating configuration..."
	@kubectl apply -f manifests/configmaps/ -n $(NAMESPACE)
	@kubectl apply -f manifests/secrets/ -n $(NAMESPACE)
	@kubectl rollout restart deployment/blog-app -n $(NAMESPACE)

## Monitor resource usage
monitor:
	@echo "📈 Monitoring resource usage..."
	@watch kubectl top pods -n $(NAMESPACE)

## Run full deployment pipeline
deploy-full: build deploy health-check
	@echo "🎉 Full deployment completed!"

## Quick development deployment (skips build)
deploy-dev:
	@echo "🚀 Quick development deployment..."
	@kubectl apply -f deploy/manifests/ -n $(NAMESPACE)

# Development targets
dev-build:
	@echo "🔨 Building development image..."
	@docker buildx build --platform linux/arm64 -t $(FULL_IMAGE)-dev --load .

dev-push:
	@echo "📤 Pushing development image..."
	@docker push $(FULL_IMAGE)-dev