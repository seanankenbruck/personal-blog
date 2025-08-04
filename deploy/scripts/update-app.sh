#!/bin/bash
# scripts/update-app.sh - Update application deployment

set -e

NAMESPACE="blog-app"
TAG=${1:-"latest"}

echo "🔄 Updating blog application..."

# Build and push new image
echo "🏗️ Building new image..."
./scripts/build-image.sh "$TAG"

# Restart deployment to pull new image
echo "🚀 Restarting application deployment..."
kubectl rollout restart deployment/blog-app -n "$NAMESPACE"

# Wait for rollout to complete
echo "⏳ Waiting for rollout to complete..."
kubectl rollout status deployment/blog-app -n "$NAMESPACE" --timeout=300s

echo "✅ Application update completed!"
echo "📊 Current status:"
kubectl get pods -n "$NAMESPACE" -l app=blog-app

#!/bin/bash
# scripts/setup-environment.sh - Initial environment setup

set -e

echo "🔧 Setting up deployment environment..."

# Create directory structure
echo "📁 Creating directory structure..."
mkdir -p manifests/{storage,secrets,configmaps,database,cache,application,ingress}
mkdir -p scripts
mkdir -p configs

# Create .env template
echo "📝 Creating .env template..."
cat > configs/.env.example << 'EOF'
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
echo "3. Run ./scripts/build-image.sh to build your application"
echo "4. Run ./scripts/deploy.sh to deploy to Kubernetes"

---
#!/bin/bash
# scripts/health-check.sh - Health check script

set -e

NAMESPACE="blog-app"

echo "🏥 Performing health check..."

# Check namespace
echo "📦 Checking namespace..."
if kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
    echo "✅ Namespace exists"
else
    echo "❌ Namespace not found"
    exit 1
fi

# Check deployments
echo "🚀 Checking deployments..."
deployments=("postgres" "redis" "blog-app")
for deployment in "${deployments[@]}"; do
    if kubectl get deployment "$deployment" -n "$NAMESPACE" >/dev/null 2>&1; then
        ready=$(kubectl get deployment "$deployment" -n "$NAMESPACE" -o jsonpath='{.status.readyReplicas}')
        desired=$(kubectl get deployment "$deployment" -n "$NAMESPACE" -o jsonpath='{.spec.replicas}')
        if [ "$ready" = "$desired" ]; then
            echo "✅ $deployment: $ready/$desired replicas ready"
        else
            echo "⚠️ $deployment: $ready/$desired replicas ready"
        fi
    else
        echo "❌ $deployment: not found"
    fi
done

# Check services
echo "🔗 Checking services..."
services=("postgres-service" "redis-service" "blog-service")
for service in "${services[@]}"; do
    if kubectl get service "$service" -n "$NAMESPACE" >/dev/null 2>&1; then
        echo "✅ $service: exists"
    else
        echo "❌ $service: not found"
    fi
done

# Check ingress
echo "🌐 Checking ingress..."
if kubectl get ingress blog-ingress -n "$NAMESPACE" >/dev/null 2>&1; then
    echo "✅ Ingress exists"
    kubectl get ingress blog-ingress -n "$NAMESPACE"
else
    echo "❌ Ingress not found"
fi

# Check persistent volume claims
echo "💾 Checking storage..."
if kubectl get pvc postgres-pvc -n "$NAMESPACE" >/dev/null 2>&1; then
    status=$(kubectl get pvc postgres-pvc -n "$NAMESPACE" -o jsonpath='{.status.phase}')
    echo "✅ PostgreSQL PVC: $status"
else
    echo "❌ PostgreSQL PVC not found"
fi

echo "🏥 Health check completed!"