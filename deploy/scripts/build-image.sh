#!/bin/bash

# Build and push script for blog application
# Usage: ./scripts/build-image.sh [tag]

set -e

# Configuration
REGISTRY=${DOCKER_REGISTRY:-""}  # Set to registry URL, leave empty for Docker Hub
IMAGE_NAME="personal-blog"
NAMESPACE=${DOCKER_NAMESPACE:-$USER}  # Docker Hub username or registry namespace
TAG=${1:-"latest"}

# Full image name
if [ -n "$REGISTRY" ]; then
    FULL_IMAGE_NAME="$REGISTRY/$NAMESPACE/$IMAGE_NAME:$TAG"
else
    FULL_IMAGE_NAME="$NAMESPACE/$IMAGE_NAME:$TAG"
fi

echo "Building image: $FULL_IMAGE_NAME"

# Build for ARM64 (Raspberry Pi)
docker buildx build \
    --platform linux/arm64 \
    --tag "$FULL_IMAGE_NAME" \
    --push \
    .

echo "Image built and pushed successfully: $FULL_IMAGE_NAME"

# Update the deployment file with new image
if [ -f "manifests/application/blog-deployment.yaml" ]; then
    echo "Updating deployment file with new image..."
    sed -i.bak "s|image: .*/$IMAGE_NAME:.*|image: $FULL_IMAGE_NAME|g" manifests/application/blog-deployment.yaml
    echo "Updated blog-deployment.yaml"
fi

echo "Build completed successfully!"
echo "To deploy: kubectl apply -f manifests/ -n blog-app"