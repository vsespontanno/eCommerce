# 🛒 eCommerce Microservices Platform

Production-ready microservices-based eCommerce platform built with Go, gRPC, Kubernetes, and GitOps principles.

## 📋 Table of Contents

- [Architecture](#-architecture)
- [Microservices](#-microservices)
- [Tech Stack](#-tech-stack)
- [Prerequisites](#-prerequisites)
- [Quick Start](#-quick-start)
- [Deployment](#-deployment)
- [GitOps Workflow](#-gitops-workflow)
- [Monitoring](#-monitoring)
- [Development](#-development)

---

## 🏗 Architecture

```
┌────────────────────────────────────────────────────────────────┐
│                         Nginx Ingress                          │
│                    (TLS Termination & Routing)                 │
└───────────────────────────┬────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
   ┌────▼────┐         ┌────▼────┐         ┌────▼────┐
   │   SSO   │         │Products │         │  Cart   │
   │ Service │◄────────┤ Service │◄────────┤ Service │
   └────┬────┘         └────┬────┘         └────┬────┘
        │                   │                   │
        │              ┌────▼────┐          ┌───▼────┐
        │              │ Wallet  │          │ Order  │
        │              │ Service │◄─────────┤Service │
        │              └────┬────┘          └───┬────┘
        │                   │                   │
        │                   └───────┬───────────┘
        │                           │
        │                      ┌────▼────────┐
        │                      │    Saga     │
        │                      │Orchestrator │
        │                      └────┬────────┘
        │                           │
        └───────────────┬───────────┴─────────┐
                        │                     │
                   ┌────▼─────┐          ┌────▼────┐
                   │PostgreSQL│          │  Kafka  │
                   │ Cluster  │          │ Cluster │
                   └──────────┘          └─────────┘
```

### Communication Patterns

- **Synchronous**: gRPC for inter-service communication
- **Asynchronous**: Kafka for event-driven workflows
- **Client Access**: HTTP/REST via gRPC-Gateway
- **Service Discovery**: Kubernetes DNS

---

## 🎯 Microservices

### 1. **SSO Service** (Authentication & Authorization)
- **Port**: gRPC 50051, HTTP 8080
- **Features**:
  - User registration & login
  - JWT token generation & validation
  - Password hashing with bcrypt
  - Token expiry management
- **Database**: PostgreSQL (`users` table)
- **API**: `/auth/register`, `/auth/login`, 

### 2. **Products Service** (Product Catalog)
- **Port**: gRPC 50051/50052, HTTP 8081
- **Features**:
  - Product CRUD operations
  - Inventory management
  - Price management
  - Product search & filtering
  - Saga products reserver
- **Database**: PostgreSQL (`products` table)
- **Cache**: Redis for product catalog

### 3. **Wallet Service** (User Balance)
- **Port**: gRPC 50051/50054, HTTP 8080
- **Features**:
  - Balance management
  - Transaction history
  - Balance reservations
  - Refunds
- **Database**: PostgreSQL (`wallets`, `transactions` tables)

### 4. **Cart Service** (Shopping Cart)
- **Port**: HTTP 8080
- **Features**:
  - Add/remove items
  - Update quantities
  - Cart persistence
  - Price calculation
- **Database**: PostgreSQL (`cart` table)
- **Cache**: Redis for active carts

### 5. **Order Service** (Order Management)
- **Port**: gRPC 50051
- **Features**:
  - Order creation
  - Order status tracking
  - Order history
  - Order cancellation
- **Database**: PostgreSQL (`orders`, `order_items` tables)

### 6. **Saga Orchestrator** (Distributed Transactions)
- **Port**: gRPC 50051
- **Features**:
  - Saga pattern implementation
  - Compensating transactions
  - Order workflow orchestration
  - Event sourcing
- **Database**: PostgreSQL (`sagas` table)
- **Message Broker**: Kafka

---

## 🛠 Tech Stack

### Backend
- **Language**: Go 1.23
- **Framework**: gRPC, gRPC-Gateway
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Message Broker**: Kafka 3.x
- **Authentication**: JWT

### Infrastructure
- **Container**: Docker
- **Orchestration**: Kubernetes (Yandex Cloud)
- **CI/CD**: GitHub Actions
- **GitOps**: ArgoCD
- **Configuration**: Kustomize
- **Ingress**: Nginx Ingress Controller
- **Monitoring**: Prometheus + Grafana
- **Registry**: Yandex Container Registry

### Development
- **Linting**: golangci-lint
- **Security**: Trivy, Gosec
- **Migrations**: goose
- **Logging**: zap (structured logging)

---

## 📦 Prerequisites

### Required Tools
- **Go**: 1.23+
- **Docker**: 20.10+
- **kubectl**: 1.28+
- **Yandex Cloud CLI**: latest
- **Git**: 2.x

### Optional Tools
- **ArgoCD CLI**: for GitOps management
- **k9s**: Kubernetes TUI
- **Postman/Insomnia**: API testing

### Yandex Cloud Resources
- Managed Kubernetes cluster
- Container Registry
- PostgreSQL cluster (or use in-cluster StatefulSet)
- Load Balancer

---

## 🚀 Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/vsespontanno/eCommerce.git
cd eCommerce
```

### 2. Set Up Environment Variables

Create `.env` file in each service directory:

```bash
# Example for sso-service
JWT_SECRET=your-secret-key-here
GRPC_PORT=50051
HTTP_PORT=8080
PG_USER=postgres
PG_PASSWORD=your-password
PG_NAME=ecommerce
PG_HOST=postgres.datastore.svc.cluster.local (or localhost)
PG_PORT=5432
```

### 3. Local Development

```bash
# Install dependencies
go mod download

# Run specific service
cd services/sso-service
go run cmd/server/main.go

```

### 4. Run with Docker Compose (Local Testing)

```bash
docker-compose up -d
```

---

## 🌐 Deployment

### Architecture Overview

```
GitHub → CI/CD → Container Registry → ArgoCD → Kubernetes
```

### Step 1: Configure GitHub Secrets

Add these secrets to your GitHub repository:

```
YC_OAUTH_TOKEN          # Yandex Cloud OAuth token
YC_CLOUD_ID             # Cloud ID
YC_FOLDER_ID            # Folder ID
YC_CONTAINER_FOLDER_ID  # Container Registry folder ID
YC_REGISTRY_ID          # Container Registry ID
YC_REGISTRY_KEY         # Service account key (JSON)
YC_K8S_CLUSTER_NAME     # Kubernetes cluster name
PAT_TOKEN               # GitHub Personal Access Token (for pushing commits)
TELEGRAM_BOT_TOKEN      # (Optional) For notifications
TELEGRAM_CHAT_ID        # (Optional) For notifications
```

### Step 2: Deploy Infrastructure

#### 2.1 Deploy PostgreSQL

```bash
kubectl apply -f deploy/k8s/stateful/postgres/
```

#### 2.2 Deploy Redis

```bash
kubectl apply -f deploy/k8s/stateful/redis/
```

#### 2.3 Deploy Kafka (Optional)

```bash
kubectl apply -f deploy/k8s/stateful/kafka/
```

### Step 3: Run Database Migrations

```bash
kubectl apply -f deploy/k8s/migrations/migration-job.yaml
kubectl logs -f job/db-migrations -n ecommerce
```

### Step 4: Install ArgoCD

```bash
# Create namespace
kubectl create namespace argocd

# Install ArgoCD
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Wait for pods to be ready
kubectl wait --for=condition=Ready pods --all -n argocd --timeout=300s

# Get admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d; echo

# Port forward to access UI
kubectl port-forward svc/argocd-server -n argocd 8080:443

# Open https://localhost:8080
# Login: admin
# Password: <from previous command>
```

### Step 5: Deploy Applications via ArgoCD

```bash
# Apply all ArgoCD Applications
kubectl apply -f deploy/k8s/argocd/applications/

# Check status
kubectl get applications -n argocd

# Watch sync progress
kubectl get applications -n argocd -w
```

### Step 6: Configure Ingress

Update Ingress hostnames in:
- `deploy/k8s/services/*/06-ingress.yaml`

```bash
# Apply Ingress resources (already applied by ArgoCD)
# Access services:
# - SSO: https://sso.your-domain.com
# - Products: https://products.your-domain.com
# - Cart: https://cart.your-domain.com
```

### Step 7: Set Up Monitoring

```bash
# Deploy Prometheus
kubectl apply -f deploy/k8s/monitoring/prometheus/

# Deploy Grafana
kubectl apply -f deploy/k8s/monitoring/grafana/

# Access Grafana
kubectl port-forward -n monitoring svc/grafana 3000:3000
# Open http://localhost:3000
# Login: admin / admin
```

---

## 🔄 GitOps Workflow

### How It Works

```
1. Developer pushes code to GitHub
   ↓
2. GitHub Actions CI/CD:
   ├─ Runs tests & linting
   ├─ Builds Docker images
   ├─ Pushes to Yandex Container Registry
   ├─ Updates kustomization.yaml with new image tags
   └─ Commits changes back to Git [skip ci]
   ↓
3. ArgoCD (every 3 minutes):
   ├─ Detects changes in Git
   ├─ Runs kustomize build
   ├─ Applies to Kubernetes
   └─ Monitors application health
   ↓
4. Kubernetes:
   └─ Rolling update with zero downtime
```

### Manual Sync

```bash
# Force sync via CLI
argocd app sync sso-service

# Sync all applications
argocd app sync --all

# Rollback to previous version
argocd app rollback sso-service <history-id>
```

### Viewing Application Status

```bash
# List all applications
argocd app list

# Get application details
argocd app get sso-service

# View sync history
argocd app history sso-service

# View application logs
argocd app logs sso-service
```

---

## 📊 Monitoring

### Prometheus Metrics

All services expose metrics on `/metrics` endpoint:

```bash
# Check metrics
kubectl port-forward -n ecommerce svc/sso-service 8080:8080
curl http://localhost:8080/metrics
```

### Key Metrics

- **HTTP Requests**: `http_requests_total`, `http_request_duration_seconds`
- **gRPC Requests**: `grpc_server_handled_total`, `grpc_server_handling_seconds`
- **Database**: `db_connections_open`, `db_query_duration_seconds`
- **Business**: `orders_created_total`, `cart_items_added_total`

### Grafana Dashboards

Import dashboards:
1. Go Processes: Dashboard ID `6671`
2. Kubernetes Cluster: Dashboard ID `7249`
3. PostgreSQL: Dashboard ID `9628`

### Health Checks

```bash
# Check service health
curl https://sso.your-domain.com/health

# Response:
# {"status":"healthy","service":"sso-service","timestamp":1234567890}
```

---

## 💻 Development

### Project Structure

```
eCommerce/
├── .github/
│   └── workflows/
│       └── go.yml              # CI/CD pipeline
├── deploy/
│   └── k8s/
│       ├── argocd/             # ArgoCD Applications
│       ├── migrations/         # Database migrations
│       ├── monitoring/         # Prometheus & Grafana
│       ├── services/           # Service manifests
│       └── stateful/           # StatefulSets (DB, Redis, Kafka)
├── migrations/                 # SQL migrations
├── nginx/                      # Local Nginx for development
│   ├── nginx.conf              # Nginx configuration
│   └── Dockerfile              # Nginx Docker image
├── proto/                      # Protocol Buffers definitions
├── services/                   # Microservices (DDD structure)
│   ├── sso-service/
│   ├── products-service/
│   ├── wallet-service/
│   ├── cart-service/
│   ├── order-service/
│   └── saga-orchestrator/
├── pkg/                        # Shared packages
│   └── logger/                 # Shared logger
├── go.mod                      # Root Go module
├── go.sum                      # Dependencies
└── docker-compose.yml          # Local development setup
```

### Service Structure (DDD - Domain-Driven Design)

```
service-name/
├── cmd/
│   └── server/
│       └── main.go                    # Entry point
├── internal/
│   ├── app/                           # Application initialization
│   │   ├── app.go                     # Main app setup
│   │   ├── grpc/                      # gRPC server setup
│   │   └── http/                      # HTTP gateway setup
│   ├── application/                   # Application layer (use cases)
│   │   └── service/                   # Business logic services
│   ├── config/                        # Configuration
│   │   └── config.go
│   ├── domain/                        # Domain layer (business rules)
│   │   ├── entity/                    # Domain entities
│   │   └── interfaces/                # Repository interfaces
│   ├── infrastructure/                # Infrastructure layer
│   │   ├── client/                    # External service clients
│   │   │   └── grpc/                  # gRPC clients
│   │   ├── db/                        # Database connections
│   │   ├── messaging/                 # Kafka/message brokers
│   │   └── repository/                # Repository implementations
│   │       └── postgres/
│   └── presentation/                  # Presentation layer
│       ├── grpc/                      # gRPC handlers
│       └── http/                      # HTTP handlers
├── .dockerignore
└── Dockerfile
```

### Local Development with Nginx

For local development, use the included Nginx reverse proxy:

```bash
# Start all services with docker-compose
docker-compose up -d

# Nginx will be available at http://localhost:80
# Routes:
# - /api/v1/auth/*     → SSO Service (8080)
# - /api/v1/products/* → Products Service (8081)
# - /api/v1/cart/*     → Cart Service (8083)
# - /api/v1/wallet/*   → Wallet Service (8082)
# - /api/v1/orders/*   → Order Service (8084)
```

### Adding a New Service

1. **Create service directory** (DDD structure):
```bash
mkdir -p services/new-service/{cmd/server,internal/{app/{grpc,http},application/service,config,domain/{entity,interfaces},infrastructure/{client/grpc,db,repository/postgres},presentation/{grpc,http}}}
```

2. **Implement service logic**

3. **Create Kubernetes manifests**:
```bash
mkdir -p deploy/k8s/services/new-service
# Create: 00-namespace.yaml, 01-secret.yaml, 02-configmap.yaml,
#         03-deployment.yaml, 04-service.yaml, 05-hpa.yaml, 06-ingress.yaml
```

4. **Create Kustomization**:
```bash
# deploy/k8s/services/new-service/kustomization.yaml
```

5. **Add to CI/CD**:
```yaml
# .github/workflows/go.yml
- name: new-service
  dockerfile: services/new-service/Dockerfile
  context: .
```

6. **Create ArgoCD Application**:
```bash
# deploy/k8s/argocd/applications/new-service.yaml
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific service tests
cd services/sso-service
go test ./...
```

### Linting

```bash
# Run golangci-lint
golangci-lint run --config=.golangci.yml

# Auto-fix issues
golangci-lint run --fix
```

---


## 🔧 Troubleshooting

### Common Issues

#### 1. ArgoCD Application OutOfSync

```bash
# Check diff
argocd app diff sso-service

# Force sync
argocd app sync sso-service --force

# Check logs
kubectl logs -n argocd deployment/argocd-application-controller
```

#### 2. Pod CrashLoopBackOff

```bash
# Check pod logs
kubectl logs -n ecommerce <pod-name>

# Check pod events
kubectl describe pod -n ecommerce <pod-name>

# Check previous logs
kubectl logs -n ecommerce <pod-name> --previous
```

#### 3. Database Connection Issues

```bash
# Test PostgreSQL connection
kubectl run -it --rm debug --image=postgres:16 --restart=Never -- \
  psql -h postgres.datastore.svc.cluster.local -U postgres -d ecommerce

# Check PostgreSQL logs
kubectl logs -n datastore statefulset/postgres
```

#### 4. Image Pull Errors

```bash
# Check imagePullSecret
kubectl get secret yc-registry-secret -n ecommerce -o yaml

# Recreate secret
kubectl create secret docker-registry yc-registry-secret \
  --docker-server=cr.yandex \
  --docker-username=json_key \
  --docker-password="$(cat key.json)" \
  -n ecommerce \
  --dry-run=client -o yaml | kubectl apply -f -
```

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Built with ❤️ using Go, Kubernetes, and GitOps**