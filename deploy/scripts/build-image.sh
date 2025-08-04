#!/bin/bash

# Build and push script for blog application
# Usage: ./scripts/build-image.sh [tag]

set -e

# Configuration
REGISTRY=${DOCKER_REGISTRY:-""}  # Set to registry URL, leave empty for Docker Hub
IMAGE_NAME="personal-blog"
NAMESPACE="smankenb"  # Docker Hub username or registry namespace
TAG=${1:-"latest"}

# Full image name
if [ -n "$REGISTRY" ]; then
    FULL_IMAGE_NAME="$REGISTRY/$NAMESPACE/$IMAGE_NAME:$TAG"
else
    FULL_IMAGE_NAME="$NAMESPACE/$IMAGE_NAME:$TAG"
fi

echo "Building image: $FULL_IMAGE_NAME"

# Create a new builder instance if it doesn't exist
if ! docker buildx inspect multiarch-builder >/dev/null 2>&1; then
    echo "Creating multi-architecture builder..."
    docker buildx create --name multiarch-builder --use
fi

# Use the multi-architecture builder
docker buildx use multiarch-builder

# Build for ARM64 with optimizations
docker buildx build \
    --platform linux/arm64 \
    --tag "$FULL_IMAGE_NAME" \
    --cache-from type=registry,ref="$NAMESPACE/$IMAGE_NAME:cache" \
    --cache-to type=registry,ref="$NAMESPACE/$IMAGE_NAME:cache,mode=max" \
    --push \
    --progress=plain \
    .

echo "Image built and pushed successfully: $FULL_IMAGE_NAME"

# Update the deployment file with new image
if [ -f "deploy/manifests/application/blog-deployment.yaml" ]; then
    echo "Updating deployment file with new image..."
    sed -i.bak "s|image: .*/$IMAGE_NAME:.*|image: $FULL_IMAGE_NAME|g" deploy/manifests/application/blog-deployment.yaml
    echo "Updated blog-deployment.yaml"
fi

echo "Build completed successfully!"
echo "To deploy: kubectl apply -f deploy/manifests/ -n blog-app"