# Personal Blog Application

A modern, test-driven personal blog built with Go, featuring a beautiful design, full OpenTelemetry feature set and comprehensive deployment infrastructure.

## 🌟 Features

- **Modern Web Application**: RESTful API using Gin web framework
- **Beautiful UI**: An inspired design with deep greens and blues
- **Database**: PostgreSQL with GORM for data persistence
- **Authentication**: JWT-based user authentication
- **Email Subscriptions**: Newsletter subscription system with email confirmation
- **Content Management**: Markdown support for blog posts
- **Responsive Design**: Mobile-friendly interface
- **OpenTelemetry**: Instrumentation for metrics, logs, and traces
- **Comprehensive Testing**: Full test suite with high coverage

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   Application   │    │   PostgreSQL    │
│   (Nginx/ALB)   │───▶│   (Go Binary)   │───▶│   Database      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   SSL/TLS       │    │   Static Files  │    │   Redis Cache   │
│   (Let's Encrypt)│    │   (CSS/JS/Images)│    │   (Optional)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📁 Project Structure

```
blog/
├── cmd/                    # Application entry point
│   └── main.go           # Main application
├── internal/              # Application code
│   ├── config/           # Configuration management
│   ├── domain/           # Domain models
│   ├── handler/          # HTTP handlers
│   ├── middleware/       # Middleware components
│   ├── repository/       # Data access layer
│   └── service/          # Business logic
├── static/                # Static assets
│   ├── styles.css        # Main stylesheet
│   └── css/              # Additional CSS
├── templates/             # HTML templates
├── deploy/                # 🚀 Deployment configurations
│   ├── docker/           # Docker configurations
│   │   ├── Dockerfile    # Multi-stage build
│   │   ├── docker-compose.yml
│   │   └── nginx/        # Nginx config
│   ├── kubernetes/       # K8s manifests
│   ├── terraform/        # Infrastructure as Code
│   ├── scripts/          # Deployment scripts
│   ├── configs/          # Environment configs
│   └── monitoring/       # Monitoring stack
├── .github/workflows/    # CI/CD pipelines
├── Makefile             # Easy deployment commands
└── README.md            # This file
```

## 🚀 Quick Start

### Prerequisites

- Go 1.24.5 or later
- PostgreSQL 12 or later
- Docker & Docker Compose (optional)
- Make (optional, for convenience)

### Local Development

```bash
# Clone the repository
git clone https://github.com/seanankenbruck/blog.git
cd blog

# Setup development environment
make dev-setup

# Start development environment
make dev
```

### Docker Development

```bash
# Build and run with Docker Compose
make docker-run

# View logs
make docker-logs

# Stop services
make docker-stop
```

## 🔧 Configuration

### Environment Variables

Create environment-specific configuration files:

```bash
# Development
cp deploy/configs/development.env .env

# Production
cp deploy/configs/production.env .env
```

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://user:pass@host:5432/db` |
| `JWT_SECRET` | JWT signing secret | `your-secret-key` |
| `SMTP_HOST` | SMTP server host | `smtp.gmail.com` |
| `SMTP_PORT` | SMTP server port | `587` |
| `SMTP_USERNAME` | SMTP username | `your-email@gmail.com` |
| `SMTP_PASSWORD` | SMTP password | `your-app-password` |

## 🐳 Docker Deployment

### Building the Image

```bash
# Build locally
make docker-build

# Build with specific version
make docker-build VERSION=v1.0.0
```

### Running with Docker Compose

```bash
# Start all services
cd deploy/docker
docker-compose up -d

# View logs
docker-compose logs -f blog-app

# Stop services
docker-compose down
```

### Services Included

- **blog-app**: Main application
- **postgres**: PostgreSQL database
- **redis**: Redis cache (optional)
- **nginx**: Reverse proxy (optional)
- **adminer**: Database management UI

## ☸️ Kubernetes Deployment

### Prerequisites

1. **Kubernetes Cluster**: EKS, GKE, or local (k3s)
2. **kubectl**: Configured to access your cluster
3. **Docker Image**: Pushed to a registry

### Deployment Steps

```bash
# Deploy to staging
./deploy/scripts/deploy.sh staging

# Deploy to production
./deploy/scripts/deploy.sh production

# Check deployment status
make k8s-status
```

### Manual Deployment

```bash
# 1. Create namespace and apply configurations
kubectl apply -f deploy/kubernetes/namespace.yaml
kubectl apply -f deploy/kubernetes/configmap.yaml
kubectl apply -f deploy/kubernetes/secret.yaml

# 2. Deploy application
kubectl apply -f deploy/kubernetes/deployment.yaml
kubectl apply -f deploy/kubernetes/service.yaml

# 3. Deploy ingress (production only)
kubectl apply -f deploy/kubernetes/ingress.yaml

# 4. Check deployment status
kubectl get pods -n blog
kubectl get services -n blog
```

## 🏗️ Infrastructure as Code (Terraform)

### Prerequisites

1. **Terraform**: Version 1.0 or later
2. **AWS Account**: With appropriate permissions
3. **S3 Bucket**: For Terraform state storage

### Deployment Steps

```bash
# Initialize Terraform
make terraform-init

# Plan changes
make terraform-plan

# Apply changes
make terraform-apply
```

### Infrastructure Components

- **VPC**: Network isolation
- **EKS Cluster**: Kubernetes cluster
- **RDS**: PostgreSQL database
- **ALB**: Application load balancer
- **ACM**: SSL certificates
- **Route53**: DNS management

## 🔄 CI/CD Pipeline

The repository includes GitHub Actions workflows for:

- **Testing**: Run tests on pull requests
- **Building**: Build and push Docker images
- **Deploying**: Deploy to staging/production
- **Security**: Vulnerability scanning

### Manual Deployment

```bash
# Trigger manual deployment
gh workflow run deploy.yml -f environment=production
```

## 📊 Monitoring & Observability

### Metrics Collection

- **Prometheus**: Metrics collection
- **Grafana**: Visualization and dashboards
- **AlertManager**: Alerting

### Logging

- **Loki**: Log aggregation
- **Fluentd**: Log forwarding
- **Grafana**: Log visualization

### Health Checks

```bash
# Application health
curl http://your-domain.com/health

# Kubernetes health
kubectl get pods -n blog
kubectl describe pod -n blog <pod-name>
```

## 🛡️ Security

### Secrets Management

```bash
# Generate secure secrets
make generate-secrets

# Update Kubernetes secrets
kubectl create secret generic blog-secrets \
  --from-literal=JWT_SECRET=your-secret \
  --from-literal=DATABASE_PASSWORD=your-password \
  -n blog
```

### Security Features

- **Network Policies**: Pod-to-pod communication
- **Security Groups**: AWS security groups
- **SSL/TLS**: End-to-end encryption
- **RBAC**: Kubernetes role-based access
- **Non-root Containers**: Security-hardened Docker images

## 🗄️ Database Management

### Migrations

```bash
# Run migrations
make db-migrate

# Create migration
go run cmd/migrate/main.go create migration_name
```

### Backups

```bash
# Create backup
make db-backup

# Restore backup
pg_restore -h host -U user -d database backup.sql
```

## 🚨 Troubleshooting

### Common Issues

1. **Database Connection**
   ```bash
   kubectl logs -n blog deployment/blog-deployment
   kubectl describe pod -n blog <pod-name>
   ```

2. **Image Pull Issues**
   ```bash
   kubectl describe pod -n blog <pod-name>
   docker pull seanankenbruck/blog:latest
   ```

3. **Ingress Issues**
   ```bash
   kubectl get ingress -n blog
   kubectl describe ingress blog-ingress -n blog
   ```

### Debug Commands

```bash
# Check pod status
kubectl get pods -n blog -o wide

# View logs
kubectl logs -f -n blog deployment/blog-deployment

# Execute commands in pod
kubectl exec -it -n blog <pod-name> -- /bin/sh

# Port forward for debugging
kubectl port-forward -n blog svc/blog-service 8080:80
```

## 📈 Scaling

### Horizontal Pod Autoscaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: blog-hpa
  namespace: blog
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: blog-deployment
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## 💰 Cost Optimization

### Resource Optimization

- **Spot Instances**: For non-critical workloads
- **Resource Limits**: Proper CPU/memory limits
- **Auto-scaling**: Based on demand

### Monitoring Costs

- **CloudWatch**: AWS cost monitoring
- **Kubecost**: Kubernetes cost analysis
- **Resource Quotas**: Prevent overspending

## 🔄 Backup & Recovery

### Backup Strategy

1. **Database**: Daily automated backups
2. **Application**: Docker image versioning
3. **Configuration**: Git version control
4. **Infrastructure**: Terraform state backups

### Disaster Recovery

1. **RTO**: 15 minutes (Recovery Time Objective)
2. **RPO**: 1 hour (Recovery Point Objective)
3. **Multi-region**: Cross-region backups
4. **Testing**: Monthly recovery drills

## 🧪 Testing

Run the test suite:
```bash
go test ./...
```

Run tests with coverage:
```bash
make test-coverage
```

## 📚 Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Terraform Documentation](https://www.terraform.io/docs/)
- [Docker Documentation](https://docs.docker.com/)
- [AWS EKS Documentation](https://docs.aws.amazon.com/eks/)

## 🤝 Support

For deployment issues or questions:

1. Check the troubleshooting section
2. Review application logs
3. Check Kubernetes events
4. Open an issue on GitHub

## 📄 License

MIT

---

**Note**: This deployment guide assumes you have the necessary permissions and access to the required services. Always follow your organization's security policies and procedures.