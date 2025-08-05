#!/bin/bash
# scripts/generate-secrets.sh - Secure secret generation

set -e

echo "🔐 Generating Kubernetes secrets from environment variables..."

# Check if .env exists
if [ ! -f .env ]; then
    echo "❌ .env file not found!"
    echo "Please copy .env.example to .env and fill in your values:"
    echo "  cp configs/.env.example .env"
    echo "  # Edit .env with your real secrets"
    exit 1
fi

# Source environment variables
source .env

# Validate required variables
required_vars=("DB_PASSWORD" "JWT_SECRET" "SMTP_PASSWORD")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ Required variable $var is not set in .env"
        exit 1
    fi
done

# Generate secrets YAML with real values
cat > deploy/manifests/secrets/generated-secrets.yaml << EOF
# WARNING: This file contains real secrets and should NOT be committed to git
# Generated from .env file on $(date)
apiVersion: v1
kind: Secret
metadata:
  name: blog-secrets
  namespace: blog-app
type: Opaque
data:
  DB_PASSWORD: $(echo -n "${DB_PASSWORD}" | base64 -w 0)
  JWT_SECRET: $(echo -n "${JWT_SECRET}" | base64 -w 0)
EOF

# Add optional SMTP password if provided
if [ -n "$SMTP_PASSWORD" ]; then
    echo "  SMTP_PASSWORD: $(echo -n "$SMTP_PASSWORD" | base64 -w 0)" >> deploy/manifests/secrets/generated-secrets.yaml
fi

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
  # These will be populated by generate-secrets.sh from .env file
  DB_PASSWORD: "<base64-encoded-password>"
  JWT_SECRET: "<base64-encoded-jwt-secret>"
  SMTP_PASSWORD: "<base64-encoded-smtp-password>"  # optional
EOF

echo "✅ Generated deploy/manifests/secrets/generated-secrets.yaml"
echo "ℹ️  Template available at deploy/manifests/secrets/app-secrets-template.yaml"
echo ""
echo "🔒 Security reminders:"
echo "  - generated-secrets.yaml contains real secrets and is git-ignored"
echo "  - Only the template file should be committed to git"
echo "  - Never commit .env file with real values"
