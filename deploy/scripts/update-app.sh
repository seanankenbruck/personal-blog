#!/bin/bash
# scripts/update-app.sh - Update application deployment

set -e

NAMESPACE="blog-app"
TAG=${1:-"latest"}

echo "🔄 Updating blog application..."

# Build and push new image
echo "🏗️ Building new image..."
./deploy/scripts/build-image.sh "$TAG"

# Restart deployment to pull new image
echo "🚀 Restarting application deployment..."
kubectl rollout restart deployment/blog-app -n "$NAMESPACE"

# Wait for rollout to complete
echo "⏳ Waiting for rollout to complete..."
kubectl rollout status deployment/blog-app -n "$NAMESPACE" --timeout=300s

echo "✅ Application update completed!"
echo "📊 Current status:"
kubectl get pods -n "$NAMESPACE" -l app=blog-app
