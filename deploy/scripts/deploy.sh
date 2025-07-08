#!/bin/bash

# Kubernetes deployment script
set -e

# Configuration
NAMESPACE="blog"
ENVIRONMENT=${1:-production}
KUBECONFIG=${2:-~/.kube/config}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    log_step "Checking prerequisites..."

    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi

    # Check if kubectl can connect to cluster
    if ! kubectl --kubeconfig=$KUBECONFIG cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
        exit 1
    fi

    log_info "Prerequisites check passed"
}

# Create namespace if it doesn't exist
create_namespace() {
    log_step "Creating namespace if it doesn't exist..."

    if ! kubectl --kubeconfig=$KUBECONFIG get namespace $NAMESPACE &> /dev/null; then
        kubectl --kubeconfig=$KUBECONFIG apply -f deploy/kubernetes/namespace.yaml
        log_info "Namespace $NAMESPACE created"
    else
        log_info "Namespace $NAMESPACE already exists"
    fi
}

# Apply Kubernetes manifests
apply_manifests() {
    log_step "Applying Kubernetes manifests..."

    # Apply in order
    kubectl --kubeconfig=$KUBECONFIG apply -f deploy/kubernetes/configmap.yaml
    kubectl --kubeconfig=$KUBECONFIG apply -f deploy/kubernetes/secret.yaml
    kubectl --kubeconfig=$KUBECONFIG apply -f deploy/kubernetes/deployment.yaml
    kubectl --kubeconfig=$KUBECONFIG apply -f deploy/kubernetes/service.yaml

    # Apply ingress only in production
    if [ "$ENVIRONMENT" = "production" ]; then
        kubectl --kubeconfig=$KUBECONFIG apply -f deploy/kubernetes/ingress.yaml
    fi

    log_info "Kubernetes manifests applied successfully"
}

# Wait for deployment to be ready
wait_for_deployment() {
    log_step "Waiting for deployment to be ready..."

    kubectl --kubeconfig=$KUBECONFIG rollout status deployment/blog-deployment -n $NAMESPACE --timeout=300s

    if [ $? -eq 0 ]; then
        log_info "Deployment is ready"
    else
        log_error "Deployment failed to become ready"
        exit 1
    fi
}

# Check deployment health
check_health() {
    log_step "Checking deployment health..."

    # Get pod status
    kubectl --kubeconfig=$KUBECONFIG get pods -n $NAMESPACE -l app=blog

    # Check service endpoints
    kubectl --kubeconfig=$KUBECONFIG get endpoints -n $NAMESPACE blog-service

    log_info "Health check completed"
}

# Show deployment info
show_info() {
    log_step "Deployment Information:"

    echo "Namespace: $NAMESPACE"
    echo "Environment: $ENVIRONMENT"
    echo "Kubeconfig: $KUBECONFIG"

    # Get service URL
    if [ "$ENVIRONMENT" = "production" ]; then
        echo "Ingress: Check your ingress controller for external access"
    else
        echo "Service: blog-service.$NAMESPACE.svc.cluster.local"
    fi

    # Get pod names
    PODS=$(kubectl --kubeconfig=$KUBECONFIG get pods -n $NAMESPACE -l app=blog -o jsonpath='{.items[*].metadata.name}')
    echo "Pods: $PODS"
}

# Main deployment process
main() {
    log_info "Starting deployment to $ENVIRONMENT environment..."

    check_prerequisites
    create_namespace
    apply_manifests
    wait_for_deployment
    check_health
    show_info

    log_info "Deployment completed successfully!"
}

# Run main function
main