#!/bin/bash
# scripts/generate-configmap.sh - ConfigMap generation for static content blog

set -e

echo "🔧 Generating Kubernetes configmap from environment variables..."

# Check if .env exists
if [ ! -f .env ]; then
    echo "❌ .env file not found!"
    echo "Please copy .env.example to .env and fill in your values:"
    echo "  cp deploy/configs/.env.example .env"
    echo "  # Edit .env with your configuration"
    exit 1
fi

# Source environment variables
source .env

# Set defaults for static content blog
SERVER_PORT="${SERVER_PORT:-8080}"
GIN_MODE="${GIN_MODE:-release}"
CONTENT_DIR="${CONTENT_DIR:-/content/posts}"

# Generate configmap YAML with values
cat > deploy/manifests/configmaps/generated-configmap.yaml << EOF
# Generated from .env file on $(date)
apiVersion: v1
kind: ConfigMap
metadata:
  name: blog-config
  namespace: blog-app
data:
  SERVER_PORT: "${SERVER_PORT}"
  GIN_MODE: "${GIN_MODE}"
  CONTENT_DIR: "${CONTENT_DIR}"
EOF

echo "✅ Generated deploy/manifests/configmaps/generated-configmap.yaml"
echo "ℹ️  Template available at deploy/manifests/configmaps/app-config-template.yaml"
echo ""
echo "ℹ️  Configuration generated for static content blog"