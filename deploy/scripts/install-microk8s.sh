#!/bin/bash
# install-microk8s.sh - Complete installation of MicroK8s

set -e

echo "🔄 Install MicroK8s on Raspberri Pi Cluster"
echo "================================="
echo ""

# Configuration
CONTROL_IP="10.42.42.100"
WORKER1_IP="10.42.42.101"  
WORKER2_IP="10.42.42.102"
SSH_USER="sankenbruck"

echo "Control Node: $CONTROL_IP"
echo "Worker 1:     $WORKER1_IP"
echo "Worker 2:     $WORKER2_IP"
echo ""

read -p "Proceed with migration? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Migration cancelled"
    exit 1
fi


# Step 2: Install MicroK8s
echo ""
echo "📦 Step 2: Installing MicroK8s on all nodes..."
echo "============================================="

# Install MicroK8s on all nodes
for node_ip in "$CONTROL_IP" "$WORKER1_IP" "$WORKER2_IP"; do
    echo "Installing MicroK8s on $node_ip..."
    ssh "$SSH_USER@$node_ip" "
        # Install snapd if not present
        if ! command -v snap >/dev/null 2>&1; then
            sudo apt update -qq
            sudo apt install -y snapd
            sudo systemctl enable --now snapd.socket
            # Wait for snapd to be ready
            sudo snap wait system seed.loaded
        fi
        
        # Install MicroK8s
        sudo snap install microk8s --classic
        
        # Add user to microk8s group
        sudo usermod -a -G microk8s sankenbruck
        
        # Wait for MicroK8s to be ready
        sudo microk8s status --wait-ready --timeout=300
        
        echo 'MicroK8s installed on $node_ip'
    " &
done
wait

echo "✅ MicroK8s installed on all nodes"

# Step 3: Configure MicroK8s cluster
echo ""
echo "🔗 Step 3: Configuring MicroK8s cluster..."
echo "========================================="

# Get join command from control node
echo "Getting cluster join command..."
JOIN_COMMAND=$(ssh "$SSH_USER@$CONTROL_IP" "sudo microk8s add-node" | grep "microk8s join" | head -1)
echo "Join command: $JOIN_COMMAND"

# Join worker nodes
echo "Joining worker nodes to cluster..."
ssh "$SSH_USER@$WORKER1_IP" "sudo $JOIN_COMMAND" &
ssh "$SSH_USER@$WORKER2_IP" "sudo $JOIN_COMMAND" &
wait

# Wait for all nodes to be ready
echo "Waiting for cluster to be ready..."
ssh "$SSH_USER@$CONTROL_IP" "sudo microk8s kubectl wait --for=condition=ready node --all --timeout=300s"

echo "✅ MicroK8s cluster configured"

Step 4: Enable essential addons
echo ""
echo "🔌 Step 4: Enabling MicroK8s addons..."
echo "===================================="

ssh "$SSH_USER@$CONTROL_IP" "
    echo 'Enabling DNS addon...'
    sudo microk8s enable dns
    
    echo 'Enabling ingress addon...'
    sudo microk8s enable ingress
    
    echo 'Enabling storage addon...'
    sudo microk8s enable storage
    
    echo 'Enabling metallb load balancer...'
    sudo microk8s enable metallb:$CONTROL_IP-$CONTROL_IP
    
    echo 'Waiting for addons to be ready...'
    sudo microk8s kubectl wait --for=condition=ready pod --all -n kube-system --timeout=300s
"

echo "✅ MicroK8s addons enabled"

# Step 5: Setup kubectl access
echo ""
echo "⚙️  Step 5: Setting up kubectl access..."
echo "======================================"

# Copy MicroK8s config
scp "$SSH_USER@$CONTROL_IP:~/.kube/config" ~/.kube/microk8s-config 2>/dev/null || {
    # If .kube/config doesn't exist, create it
    ssh "$SSH_USER@$CONTROL_IP" "
        mkdir -p ~/.kube
        sudo microk8s config > ~/.kube/config
        chmod 600 ~/.kube/config
    "
    scp "$SSH_USER@$CONTROL_IP:~/.kube/config" ~/.kube/microk8s-config
}

# Update server IP in config
sed -i.bak "s/127.0.0.1/$CONTROL_IP/g" ~/.kube/microk8s-config
sed -i.bak "s/localhost/$CONTROL_IP/g" ~/.kube/microk8s-config

# Backup existing kubectl config and set new one
if [ -f ~/.kube/config ]; then
    cp ~/.kube/config "~/.kube/config.backup.$(date +%Y%m%d-%H%M%S)"
fi

cp ~/.kube/microk8s-config ~/.kube/config
export KUBECONFIG=~/.kube/config

echo "✅ kubectl configured for MicroK8s"

# Step 6: Test cluster functionality  
echo ""
echo "🧪 Step 6: Testing cluster functionality..."
echo "========================================"

echo "📊 Cluster status:"
kubectl get nodes -o wide

echo ""
echo "📦 System pods:"
kubectl get pods -n kube-system

echo ""
echo "🧪 DNS test:"
if kubectl run test-dns --image=busybox:1.28 --rm -it --restart=Never --timeout=30s -- nslookup kubernetes.default; then
    echo "✅ DNS working perfectly!"
else
    echo "❌ DNS test failed"
    exit 1
fi

echo ""
echo "🧪 Service connectivity test:"
KUBE_DNS_IP=$(kubectl get service -n kube-system kube-dns -o jsonpath='{.spec.clusterIP}')
if kubectl run test-service --image=busybox:1.28 --rm -it --restart=Never --timeout=30s -- nc -zv "$KUBE_DNS_IP" 53; then
    echo "✅ Service routing working perfectly!"
else
    echo "❌ Service connectivity test failed"
    exit 1
fi

# Step 7: Deploy k8s manifests
echo ""
echo "🚀 Step 7: Deploying k8s manifests..."
echo "============================================"

if [ -d "manifests" ]; then
    echo "Creating blog-app namespace..."
    kubectl create namespace blog-app --dry-run=client -o yaml | kubectl apply -f -
    
    echo "Deploying application manifests..."
    
    # Deploy in order: storage, secrets, configs, database, cache, application, ingress
    kubectl apply -f manifests/storage/ || echo "⚠️  Storage manifests not found or failed"
    kubectl apply -f manifests/secrets/ || echo "⚠️  Secrets manifests not found or failed"
    kubectl apply -f manifests/configmaps/ || echo "⚠️  ConfigMaps manifests not found or failed"
    
    # Wait for storage to be ready before deploying database
    sleep 10
    
    kubectl apply -f manifests/database/ || echo "⚠️  Database manifests not found or failed"
    kubectl apply -f manifests/cache/ || echo "⚠️  Cache manifests not found or failed"
    
    # Wait for database to be ready before deploying application
    echo "Waiting for database to be ready..."
    kubectl wait --for=condition=available deployment/postgres -n blog-app --timeout=300s || echo "⚠️  Database not ready"
    
    kubectl apply -f manifests/application/ || echo "⚠️  Application manifests not found or failed"
    kubectl apply -f manifests/ingress/ || echo "⚠️  Ingress manifests not found or failed"
    
    # Wait for application to be ready
    echo "Waiting for application to be ready..."
    kubectl wait --for=condition=available deployment/blog-app -n blog-app --timeout=300s || echo "⚠️  Application not ready"
    
    echo "📊 Application status:"
    kubectl get all -n blog-app
    
    echo ""
    echo "📋 Application logs:"
    kubectl logs deployment/blog-app -n blog-app --tail=10 || echo "⚠️  Could not get application logs"
    
else
    echo "ℹ️  No manifests directory found. Deploy application manually with:"
    echo "  kubectl apply -f deploy/manifests/"
fi


# Final summary
echo ""
echo "🎉 MicroK8s install and blog deployment completed!"
echo "=============================================="
echo ""
echo "📊 Final cluster status:"
kubectl get nodes
echo ""
kubectl get pods -n blog-app 2>/dev/null || echo "No blog-app namespace found"
echo ""
echo "🔗 Access Information:"
echo "  kubectl context: $(kubectl config current-context)"
echo "  Cluster endpoint: https://$CONTROL_IP:16443"
echo "  Dashboard: microk8s dashboard-proxy (run on control node)"
echo ""
echo "🚀 Next steps:"
echo "  1. Verify application is working: kubectl get pods -n blog-app"
echo "  2. Check application logs: kubectl logs deployment/blog-app -n blog-app"
echo "  3. Access application via ingress or port-forward"
echo "  4. Set up external DNS/load balancer if needed"
echo ""
echo "💡 MicroK8s specific commands:"
echo "  ssh $SSH_USER@$CONTROL_IP"
echo "  sudo microk8s kubectl get pods    # Alternative kubectl"
echo "  sudo microk8s dashboard-proxy     # Web dashboard"
echo ""
