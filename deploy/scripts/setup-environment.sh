
#!/bin/bash
# scripts/setup-environment.sh - Initial environment setup

set -e

# Create .env template
echo "📝 Creating .env template..."
cat > deploy/configs/.env.example << 'EOF'
# Docker Configuration
DOCKER_REGISTRY=
DOCKER_NAMESPACE=yourusername

# Application Configuration
APP_DOMAIN=your-domain.com
JWT_SECRET=your-jwt-secret-key

# Database Configuration
DB_PASSWORD=secure-postgres-password

# Email Configuration (optional)
SMTP_HOST=
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=
EMAIL_SENDER=

# SSL/TLS Configuration
CERT_MANAGER_EMAIL=your-email@domain.com
EOF

echo "✅ Environment setup completed!"
echo "📝 Next steps:"
echo "1. Copy configs/.env.example to .env and fill in your values"
echo "2. Update manifests with your specific configuration"
echo "3. Run make build to build your application"
