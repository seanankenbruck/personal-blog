#!/bin/bash

# Build and push Docker image script
set -e

# Configuration
IMAGE_NAME="seanankenbruck/blog"
VERSION=${1:-latest}
REGISTRY=${2:-docker.io}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    log_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Build the image
log_info "Building Docker image: ${IMAGE_NAME}:${VERSION}"
docker build -t ${IMAGE_NAME}:${VERSION} -f deploy/docker/Dockerfile .

if [ $? -eq 0 ]; then
    log_info "Docker image built successfully"
else
    log_error "Failed to build Docker image"
    exit 1
fi

# Tag for registry
FULL_IMAGE_NAME="${REGISTRY}/${IMAGE_NAME}:${VERSION}"
docker tag ${IMAGE_NAME}:${VERSION} ${FULL_IMAGE_NAME}

# Push to registry (if not local)
if [ "$REGISTRY" != "local" ]; then
    log_info "Pushing image to registry: ${FULL_IMAGE_NAME}"

    # Check if logged in to registry
    if ! docker info | grep -q "Username"; then
        log_warn "Not logged in to Docker registry. Please run 'docker login' first."
        log_info "Image built successfully but not pushed. Run 'docker push ${FULL_IMAGE_NAME}' to push manually."
        exit 0
    fi

    docker push ${FULL_IMAGE_NAME}

    if [ $? -eq 0 ]; then
        log_info "Image pushed successfully to registry"
    else
        log_error "Failed to push image to registry"
        exit 1
    fi
else
    log_info "Skipping push (local registry)"
fi

# Clean up old images (optional)
if [ "$3" = "--cleanup" ]; then
    log_info "Cleaning up old images..."
    docker image prune -f
fi

log_info "Build completed successfully!"
log_info "Image: ${FULL_IMAGE_NAME}"