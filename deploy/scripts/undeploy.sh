#!/bin/bash
# scripts/undeploy.sh - Cleanup script for static content blog

set -e

NAMESPACE="blog-app"

echo "🗑️ Undeploying Personal Blog from Kubernetes..."

# Delete ingress first
echo "🌐 Removing ingress..."
kubectl delete -f deploy/manifests/ingress/ --ignore-not-found=true

# Delete application
echo "📱 Removing application..."
kubectl delete -f deploy/manifests/application/ --ignore-not-found=true

# Delete configs and secrets
echo "🔐 Removing configuration..."
kubectl delete -f deploy/manifests/configmaps/ --ignore-not-found=true
kubectl delete -f deploy/manifests/secrets/ --ignore-not-found=true

# Delete namespace
echo "📦 Removing namespace..."
kubectl delete namespace "$NAMESPACE" --ignore-not-found=true

echo "✅ Cleanup completed!"