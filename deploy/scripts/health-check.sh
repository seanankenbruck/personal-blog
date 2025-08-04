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