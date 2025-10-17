#!/bin/bash
# scripts/generate-secrets.sh - Secret generation for static content blog

set -e

echo "🔐 Generating Kubernetes secrets..."

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

# Generate empty secrets YAML (no secrets needed for static content)
cat > deploy/manifests/secrets/generated-secrets.yaml << EOF
# Generated for static content blog on $(date)
# No secrets required for static content deployment
apiVersion: v1
kind: Secret
metadata:
  name: blog-secrets
  namespace: blog-app
type: Opaque
data:
  # No secrets needed for static content blog
  # This file is kept for future extensibility
EOF

# Create template file for reference (safe to commit)
cat > deploy/manifests/secrets/app-secrets-template.yaml << 'EOF'
# Template file - DO NOT put real secrets here
# Use scripts/generate-secrets.sh to create the actual secrets file
apiVersion: v1
kind: Secret
metadata:
  name: blog-secrets
  namespace: blog-app
type: Opaque
data:
  # No secrets needed for static content blog
  # This file is kept for future extensibility
EOF

echo "✅ Generated deploy/manifests/secrets/generated-secrets.yaml"
echo "ℹ️  Template available at deploy/manifests/secrets/app-secrets-template.yaml"
echo ""
echo "ℹ️  No secrets required for static content blog deployment"