#!/bin/bash
# scripts/deploy.sh - Main deployment script

set -e

NAMESPACE="blog-app"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MANIFEST_DIR="$(dirname "$SCRIPT_DIR")/manifests"

echo "🚀 Deploying Personal Blog to Kubernetes..."

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
echo "📋 Checking prerequisites..."
if ! command_exists kubectl; then
    echo "❌ kubectl is required but not installed."
    exit 1
fi

if ! command_exists docker; then
    echo "❌ Docker is required but not installed."
    exit 1
fi

# Check if cluster is accessible
if ! kubectl cluster-info >/dev/null 2>&1; then
    echo "❌ Cannot connect to Kubernetes cluster."
    exit 1
fi

echo "✅ Prerequisites check passed"

# Create namespace
echo "📦 Creating namespace..."
kubectl apply -f "$MANIFEST_DIR/namespace.yaml"

# Deploy storage
echo "💾 Setting up persistent storage..."
kubectl apply -f "$MANIFEST_DIR/storage/"

# Wait for PVC to be bound
echo "⏳ Waiting for PVC to be bound..."
kubectl wait --for=condition=Bound pvc/postgres-pvc -n "$NAMESPACE" --timeout=60s

# Deploy secrets and config
echo "🔐 Applying secrets and configuration..."
kubectl apply -f "$MANIFEST_DIR/secrets/"
kubectl apply -f "$MANIFEST_DIR/configmaps/"

# Deploy database
echo "🗄️ Deploying PostgreSQL..."
kubectl apply -f "$MANIFEST_DIR/database/"

# Wait for database to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
kubectl wait --for=condition=available deployment/postgres -n "$NAMESPACE" --timeout=300s

# Deploy cache
echo "🚀 Deploying Redis..."
kubectl apply -f "$MANIFEST_DIR/cache/"

# Wait for Redis to be ready
echo "⏳ Waiting for Redis to be ready..."
kubectl wait --for=condition=available deployment/redis -n "$NAMESPACE" --timeout=120s

# Deploy application
echo "🌐 Deploying blog application..."
kubectl apply -f "$MANIFEST_DIR/application/"

# Wait for application to be ready
echo "⏳ Waiting for blog application to be ready..."
kubectl wait --for=condition=available deployment/blog-app -n "$NAMESPACE" --timeout=300s

# Deploy ingress
echo "🌍 Setting up ingress..."
kubectl apply -f "$MANIFEST_DIR/ingress/"

echo "✅ Deployment completed successfully!"
echo ""
echo "📊 Deployment status:"
kubectl get pods -n "$NAMESPACE"
echo ""
echo "🔗 Services:"
kubectl get services -n "$NAMESPACE"
echo ""
echo "🌐 Ingress:"
kubectl get ingress -n "$NAMESPACE"

echo ""
echo "🎉 Your blog is now deployed!"
echo "📝 Next steps:"
echo "1. Update your DNS to point to your cluster"
echo "2. Check the application logs: kubectl logs -f deployment/blog-app -n $NAMESPACE"
echo "3. Access your blog at: https://your-domain.com"
