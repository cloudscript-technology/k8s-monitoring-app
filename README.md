# K8s Monitoring App

A Kubernetes-native monitoring application that runs inside your cluster and collects metrics from your applications. Configure metrics via REST API and let the application automatically collect and store them asynchronously.

## Features

- 🎯 **Easy Configuration**: Register applications and configure metrics via REST API
- 🔄 **Automatic Collection**: Metrics collected every minute via cron jobs
- 📊 **Multiple Metric Types**: Health checks, pod status, CPU, memory, PVC usage, and node tracking
- 🔌 **Database Connection Monitoring**: Test and monitor Redis, PostgreSQL, MongoDB, MySQL, and Kong connections with authentication
 - 🗄️ **Historical Data**: All metrics stored in SQLite for analysis
- 🔐 **OAuth 2.0 Authentication**: Secure Google OAuth authentication with domain restriction
- 🔒 **RBAC Ready**: Designed to work with Kubernetes security best practices
- 📈 **Scalable**: Built to monitor multiple applications and namespaces
- 🖥️ **Modern Web UI**: Real-time dashboard with HTMX and auto-refresh every 10s
- 🎨 **Beautiful Interface**: Clean design with visual indicators and progress bars

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                        │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              K8s Monitoring App                       │  │
│  │                                                        │  │
│  │  ┌────────────┐         ┌────────────────────────┐   │  │
│  │  │            │         │                        │   │  │
│  │  │  REST API  │◄────────┤  Monitoring Service   │   │  │
│  │  │            │         │  (Cron: @every 1m)    │   │  │
│  │  └─────┬──────┘         └───────────┬────────────┘   │  │
│  │        │                            │                │  │
│  │        │                            │                │  │
│  │        ▼                            ▼                │  │
│  │  ┌──────────────────────────────────────────────┐   │  │
│  │  │          SQLite Database (file)              │   │  │
│  │  │  - Projects                                  │   │  │
│  │  │  - Applications                              │   │  │
│  │  │  - Metric Types                              │   │  │
│  │  │  - Application Metrics (config)              │   │  │
│  │  │  - Application Metric Values (data)          │   │  │
│  │  └──────────────────────────────────────────────┘   │  │
│  └──────────────────────────────────────────────────────┘  │
│                           │                                 │
│                           │ Kubernetes API                  │
│                           ▼                                 │
│  ┌─────────────────────────────────────────────────────┐   │
│  │     Pods, Metrics, PVCs, Nodes                      │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Kubernetes cluster (v1.20+)
- SQLite (bundled, no external database required)
- metrics-server installed in your cluster
- Google OAuth 2.0 credentials (for authentication)

### Authentication

This application uses **Google OAuth 2.0** for secure authentication. Only users from authorized email domains can access the application.

**Setup Steps:**

1. **Create OAuth 2.0 credentials** in [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. **Configure environment variables** (see below)
3. **Users authenticate** via Google sign-in when accessing the application

**📖 Complete OAuth Setup Guide:** [docs/OAUTH_SETUP.md](docs/OAUTH_SETUP.md)

**Required Environment Variables:**
```bash
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/callback
ALLOWED_GOOGLE_DOMAINS=yourcompany.com  # Comma-separated list
```

**Security Features:**
- 🔐 Secure OAuth 2.0 flow with Google
- 🏢 Domain-restricted access (only allow specific email domains)
- 🍪 Session-based authentication with 24-hour expiry
- 🔒 HttpOnly and Secure cookies
- 🚪 Protected routes with automatic redirect to login

### Access the Web UI

Once the application is running:

```bash
# Port forward to access the UI locally
kubectl port-forward service/k8s-monitoring-app 8080:8080

# Access the dashboard
open http://localhost:8080
```

**Features:**
- ✨ Real-time metrics visualization
- 🔄 Auto-refresh every 10 seconds
- 📊 Visual indicators for health, CPU, memory, and disk
- 🎯 Organized by projects and applications

For more details about the Web UI, see [web/README.md](web/README.md).

### 1. Setup RBAC

```bash
kubectl apply -f - <<EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8s-monitoring-app
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8s-monitoring-app
rules:
  - apiGroups: [""]
    resources: ["pods", "persistentvolumeclaims", "nodes"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["pods/exec"]
    verbs: ["create"]
  - apiGroups: ["metrics.k8s.io"]
    resources: ["pods"]
    verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: k8s-monitoring-app
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: k8s-monitoring-app
subjects:
  - kind: ServiceAccount
    name: k8s-monitoring-app
    namespace: default
EOF
```

### 2. Deploy the Application

Using the provided Helm chart:

```bash
helm install k8s-monitoring-app ./chart \
  --set env[0].name=DB_PATH \
  --set env[0].value=/data/k8s_monitoring.db
```
Note: For persistence in Kubernetes, mount a PVC at `/data`.

### 3. Configure Your First Application

```bash
# Port forward to access the API
kubectl port-forward svc/k8s-monitoring-app 8080:8080

# Create a project
PROJECT_ID=$(curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{"name":"Production","description":"Production apps"}' | jq -r '.id')

# Register an application
APP_ID=$(curl -X POST http://localhost:8080/api/v1/applications \
  -H "Content-Type: application/json" \
  -d "{\"project_id\":\"$PROJECT_ID\",\"name\":\"my-app\",\"description\":\"My App\",\"namespace\":\"default\"}" | jq -r '.id')

# Get metric types
HEALTH_CHECK_TYPE=$(curl http://localhost:8080/api/v1/metric-types | jq -r '.[] | select(.name=="HealthCheck") | .id')

# Configure health check monitoring
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\":\"$APP_ID\",
    \"type_id\":\"$HEALTH_CHECK_TYPE\",
    \"configuration\":{
      \"health_check_url\":\"http://my-app.default.svc.cluster.local:8080/health\",
      \"method\":\"GET\",
      \"expected_status\":200,
      \"timeout_seconds\":10
    }
  }"
```

## Available Metric Types

### 1. HealthCheck
Performs HTTP health checks on your application endpoints.

**Configuration:**
```json
{
  "health_check_url": "http://service.namespace.svc.cluster.local/health",
  "method": "GET",
  "expected_status": 200,
  "timeout_seconds": 10
}
```

### 2. PodStatus
Monitors pod phase, readiness, and restart count.

**Configuration:**
```json
{
  "pod_label_selector": "app=myapp",
  "container_name": "main"
}
```

### 3. PodMemoryUsage
Tracks memory usage and percentage of limits.

**Configuration:**
```json
{
  "pod_label_selector": "app=myapp",
  "container_name": "main"
}
```

### 4. PodCpuUsage
Tracks CPU usage in millicores and percentage of limits.

**Configuration:**
```json
{
  "pod_label_selector": "app=myapp",
  "container_name": "main"
}
```

### 5. PvcUsage
Monitors Persistent Volume Claim capacity and usage by executing `df` inside pods.

**Configuration:**
```json
{
  "pvc_name": "my-pvc",
  "pod_label_selector": "app=myapp",
  "container_name": "main",     // Optional
  "pvc_mount_path": "/data"     // Optional: auto-discovered
}
```

**Required:** `pvc_name`, `pod_label_selector`  
**Auto-discovery:** The system automatically finds the mount path by inspecting pod volumes.

### 6. PodActiveNodes
Tracks which nodes your pods are running on.

**Configuration:**
```json
{
  "pod_label_selector": "app=myapp"
}
```

### 7. Database and Service Connection Monitoring

Monitor database and service connections with authentication support.

#### RedisConnection
Tests Redis connection with authentication.

**Configuration:**
```json
{
  "connection_host": "redis.default.svc.cluster.local",
  "connection_port": 6379,
  "connection_password": "your-password",
  "connection_db": 0,
  "connection_timeout": 5
}
```

#### PostgreSQLConnection
Tests PostgreSQL connection with authentication.

**Configuration:**
```json
{
  "connection_host": "postgres.default.svc.cluster.local",
  "connection_port": 5432,
  "connection_username": "user",
  "connection_password": "password",
  "connection_database": "mydb",
  "connection_ssl": false,
  "connection_timeout": 10
}
```

#### MongoDBConnection
Tests MongoDB connection with authentication.

**Configuration:**
```json
{
  "connection_host": "mongodb.default.svc.cluster.local",
  "connection_port": 27017,
  "connection_username": "admin",
  "connection_password": "password",
  "connection_database": "mydb",
  "connection_auth_source": "admin",
  "connection_timeout": 5
}
```

#### MySQLConnection
Tests MySQL connection with authentication.

**Configuration:**
```json
{
  "connection_host": "mysql.default.svc.cluster.local",
  "connection_port": 3306,
  "connection_username": "root",
  "connection_password": "password",
  "connection_database": "mydb",
  "connection_timeout": 5
}
```

#### KongConnection
Tests Kong API Gateway connection and health.

**Configuration:**
```json
{
  "connection_host": "kong-admin.default.svc.cluster.local",
  "connection_port": 8001,
  "kong_admin_url": "http://kong-admin.default.svc.cluster.local:8001",
  "connection_timeout": 5
}
```

**📖 For detailed documentation on connection metrics**, see:
- [docs/CONNECTION_METRICS.md](docs/CONNECTION_METRICS.md) - Complete guide with examples
- [postman/CONNECTION_METRICS_EXAMPLES.md](postman/CONNECTION_METRICS_EXAMPLES.md) - Postman examples
- [examples/connection-metrics-test.sh](examples/connection-metrics-test.sh) - Interactive testing script

## Project Structure

```
k8s-monitoring-app/
├── cmd/
│   └── main.go                    # Application entry point
├── internal/
│   ├── agent/                     # Agent management (existing)
│   ├── project/                   # Project management
│   │   ├── repository/
│   │   └── service.go
│   ├── application/               # Application management
│   │   ├── repository/
│   │   └── service.go
│   ├── metric_type/               # Metric type management
│   │   ├── repository/
│   │   └── service.go
│   ├── application_metric/        # Metric configuration
│   │   ├── repository/
│   │   └── service.go
│   ├── application_metric_value/  # Metric value storage
│   │   └── repository/
│   ├── monitoring/                # Monitoring service with cron
│   │   └── service.go
│   ├── k8s/                       # Kubernetes client wrapper
│   │   └── client.go
│   ├── core/                      # Core HTTP server
│   ├── server/                    # Server configuration
│   └── env/                       # Environment configuration
├── pkg/                           # Public models
│   ├── project/model/
│   ├── application/model/
│   ├── metric_type/model/
│   ├── application_metric/model/
│   └── application_metric_value/model/
├── database/
│   └── migrations/                # Database migrations
├── chart/                         # Helm chart
├── docs/
│   ├── API.md                     # API documentation
│   └── DEPLOYMENT.md              # Deployment guide
└── README.md
```

## Development

### Prerequisites
- Go 1.24+
- Docker (optional)
- Access to a Kubernetes cluster (can be local: Minikube, Kind, Docker Desktop)

### Local Development

The application automatically detects if it's running locally or inside a Kubernetes cluster.

**For local development**, it uses your kubeconfig in this order:
1. `KUBECONFIG` environment variable
2. `~/.kube/config` (default kubectl config)
3. In-cluster config (when running inside K8s)

#### Quick Start

1. Clone the repository:
```bash
git clone <repository-url>
cd k8s-monitoring-app
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
export DB_PATH=./data/k8s_monitoring.db
export LOG_LEVEL=debug

# Optional: specify kubeconfig explicitly
# export KUBECONFIG=/path/to/your/kubeconfig
```

5. Verify kubectl access:
```bash
kubectl get nodes
kubectl get pods --all-namespaces
```

6. Run the application:
```bash
go run cmd/main.go
```

The application will automatically use your local kubeconfig and connect to your cluster!

#### Complete Local Development Guide

For detailed instructions including:
- Using Minikube or Kind
- Hot reload setup
- Debugging tips
- Testing with different clusters
- Troubleshooting common issues

See [docs/LOCAL_DEVELOPMENT.md](docs/LOCAL_DEVELOPMENT.md)

### Running with Docker Compose

```bash
docker-compose up -d
```

This will start the monitoring application.

### Build the Application

```bash
go build -o k8s-monitoring-app cmd/main.go
```

## API Documentation

Complete API documentation is available in [docs/API.md](docs/API.md).

### Testing with Postman

A complete Postman collection is available in the `postman/` directory:

```bash
# Import these files into Postman:
postman/K8s-Monitoring-App.postman_collection.json
postman/K8s-Monitoring-App.postman_environment.json
```

The collection includes:
- All API endpoints with examples
- Pre-configured environment variables
- Test scripts that auto-save IDs
- Complete workflow for quick testing

See [postman/README.md](postman/README.md) for detailed instructions.

### Key Endpoints

#### Configuration
- `GET /health` - Health check
- `GET /api/v1/projects` - List projects
- `POST /api/v1/projects` - Create project
- `GET /api/v1/applications` - List applications
- `POST /api/v1/applications` - Register application
- `GET /api/v1/metric-types` - List available metric types
- `POST /api/v1/application-metrics` - Configure metric for application
- `GET /api/v1/applications/:id/metrics` - Get metrics for application

#### 🆕 Viewing Collected Metrics
- `GET /api/v1/applications/:id/latest-metrics` - Get latest values for all metrics
- `GET /api/v1/application-metrics/:metric_id/values?limit=100` - Get metric history
- `GET /api/v1/metric-values/:id` - Get specific metric value

See [docs/ENDPOINTS_SUMMARY.md](docs/ENDPOINTS_SUMMARY.md) for complete endpoint reference.

## Deployment

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for detailed deployment instructions.

### Helm Installation

```bash
helm install k8s-monitoring-app ./chart \
  --namespace monitoring \
  --create-namespace
```

## Configuration

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| **Database** |
| DB_PATH | SQLite database file path | No | `./data/k8s_monitoring.db` |
| **Authentication** |
| GOOGLE_CLIENT_ID | Google OAuth Client ID | Yes* | - |
| GOOGLE_CLIENT_SECRET | Google OAuth Client Secret | Yes* | - |
| GOOGLE_REDIRECT_URL | OAuth callback URL | Yes* | - |
| ALLOWED_GOOGLE_DOMAINS | Comma-separated allowed email domains | No | All domains |
| ADMIN_TOKEN | Admin token for service-to-service auth | No | - |
| **Metrics** |
| METRICS_RETENTION_DAYS | Days to keep metric history | No | 30 |
| METRICS_CLEANUP_INTERVAL | Cron expression for cleanup | No | 0 2 * * * |
| METRICS_COLLECTION_INTERVAL | Collection interval in seconds | No | 60 |
| **Alerts** |
| SLACK_ALERTS_ENABLED | Enable Slack notifications on metric failures | No | false |
| SLACK_WEBHOOK_URL | Slack Incoming Webhook URL | No | - |
| SLACK_ALERTS_DEDUP_MINUTES | Suppress repeated alerts within N minutes | No | 10 |
| **Other** |
| ENV | Environment (development/staging/production) | No | development |
| LOG_LEVEL | Logging level | No | info |
| ELASTIC_APM_SERVICE_NAME | APM service name | No | - |
| ELASTIC_APM_SERVER_URL | APM server URL | No | - |

\* Required for OAuth authentication. If not configured, authentication will be disabled.

**See also:** [docs/ENVIRONMENT_VARIABLES.md](docs/ENVIRONMENT_VARIABLES.md) for detailed configuration

## Monitoring Schedule

By default, metrics are collected every minute. To modify the schedule, update the cron expression in `internal/monitoring/service.go`:

```go
_, err := m.cron.AddFunc("@every 1m", m.collectMetrics)
```

Available cron expressions:
- `@every 30s` - Every 30 seconds
- `@every 1m` - Every minute (default)
- `@every 5m` - Every 5 minutes
- `*/2 * * * *` - Every 2 minutes

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

See LICENSE file for details.

## Troubleshooting

Having issues? Check our comprehensive troubleshooting guide:

📖 **[docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md)**

Common issues covered:
- metrics-server not found
- Health checks failing
- Pods/PVCs not found
- Database connection issues
- Performance problems

## Support

For issues and questions:
- Check [Troubleshooting Guide](docs/TROUBLESHOOTING.md)
- Create an issue in the repository
- Check existing documentation in `docs/`

## Roadmap

- [ ] Alert configuration and notification system
- [ ] Grafana dashboard templates
- [ ] Metric retention policies
- [ ] Custom metric collectors
- [ ] Multi-cluster support
- [ ] Metric export to Prometheus
- [ ] Web UI for configuration
- [ ] Metric aggregation and statistics

## Acknowledgments

- Built with [Echo](https://echo.labstack.com/) web framework
- Uses [robfig/cron](https://github.com/robfig/cron) for scheduling
- Kubernetes client-go for K8s API interaction
