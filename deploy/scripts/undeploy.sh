#!/bin/bash
# scripts/undeploy.sh - Cleanup script

set -e

NAMESPACE="blog-app"

echo "🗑️ Undeploying Personal Blog from Kubernetes..."

# Delete ingress first
echo "🌐 Removing ingress..."
kubectl delete -f deploy/manifests/ingress/ --ignore-not-found=true

# Delete application
echo "📱 Removing application..."
kubectl delete -f deploy/manifests/application/ --ignore-not-found=true

# Delete cache
echo "🚀 Removing Redis..."
kubectl delete -f deploy/manifests/cache/ --ignore-not-found=true

# Delete database
echo "🗄️ Removing PostgreSQL..."
kubectl delete -f deploy/manifests/database/ --ignore-not-found=true

# Delete configs and secrets
echo "🔐 Removing configuration..."
kubectl delete -f deploy/manifests/configmaps/ --ignore-not-found=true
kubectl delete -f deploy/manifests/secrets/ --ignore-not-found=true

# Ask about persistent storage
echo "❓ Do you want to delete persistent storage? (PVC and PV)"
echo "⚠️  WARNING: This will permanently delete your database data!"
read -p "Delete storage? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "💾 Removing persistent storage..."
    kubectl delete -f deploy/manifests/storage/ --ignore-not-found=true
else
    echo "💾 Keeping persistent storage..."
fi

# Delete namespace
echo "📦 Removing namespace..."
kubectl delete namespace "$NAMESPACE" --ignore-not-found=true

echo "✅ Cleanup completed!"
