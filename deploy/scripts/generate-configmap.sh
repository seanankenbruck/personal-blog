#!/bin/bash
# scripts/generate-configmap.sh - Secure configmap generation

set -e

echo "🔐 Generating Kubernetes configmap from environment variables..."

# Check if .env exists
if [ ! -f .env ]; then
    echo "❌ .env file not found!"
    echo "Please copy .env.example to .env and fill in your values:"
    echo "  cp configs/.env.example .env"
    echo "  # Edit .env with your real values"
    exit 1
fi

# Source environment variables
source .env

# Validate required variables
required_vars=("DB_HOST" "DB_PORT" "DB_USER" "DB_NAME" "DB_SSLMODE" "REDIS_HOST" "REDIS_PORT" "SERVER_PORT" "GIN_MODE" "APP_HOST" "SMTP_HOST" "SMTP_PORT" "SMTP_USER" "EMAIL_SENDER")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ Required variable $var is not set in .env"
        exit 1
    fi
done

# Generate configmap YAML with real values (all values quoted as strings)
cat > deploy/manifests/configmaps/generated-configmap.yaml << EOF
# Generated from .env file on $(date)
apiVersion: v1
kind: ConfigMap
metadata:
  name: blog-config
  namespace: blog-app
data:
  DB_HOST: "${DB_HOST}"
  DB_PORT: "${DB_PORT}"
  DB_USER: "${DB_USER}"
  DB_NAME: "${DB_NAME}"
  DB_SSLMODE: "${DB_SSLMODE}"
  REDIS_HOST: "${REDIS_HOST}"
  REDIS_PORT: "${REDIS_PORT}"
  SERVER_PORT: "${SERVER_PORT}"
  GIN_MODE: "${GIN_MODE}"
  APP_HOST: "${APP_HOST}"
  SMTP_HOST: "${SMTP_HOST}"
  SMTP_PORT: "${SMTP_PORT}"
  SMTP_USER: "${SMTP_USER}"
  EMAIL_SENDER: "${EMAIL_SENDER}"
EOF

echo "✅ Generated deploy/manifests/configmaps/generated-configmap.yaml"
echo "ℹ️  Template available at deploy/manifests/configmaps/app-config-template.yaml"
echo ""
echo "🔒 Security reminders:"
echo "  - generated-configmap.yaml contains real configuration values and is git-ignored"
echo "  - Only the template file should be committed to git"
echo "  - Never commit .env file with real values"
