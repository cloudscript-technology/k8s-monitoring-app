# K8s Monitoring App - API Documentation

## Overview

This application provides a Kubernetes monitoring solution that runs inside a Kubernetes cluster and monitors applications based on user-configured metrics.

## Features

- **Project Management**: Organize applications into projects
- **Application Management**: Register applications to be monitored
- **Metric Configuration**: Configure various metric types for each application
- **Automated Monitoring**: Asynchronous collection of metrics using cron jobs
- **Metric Storage**: Historical metric data storage for analysis

## Metric Types

The following metric types are available:

1. **HealthCheck** - HTTP health check monitoring
2. **PodStatus** - Pod status and restart count
3. **PodMemoryUsage** - Memory usage and limits
4. **PodCpuUsage** - CPU usage and limits
5. **PvcUsage** - Persistent Volume Claim usage
6. **PodActiveNodes** - Active nodes where pods are running

## API Endpoints

### Projects

#### List Projects
```
GET /api/v1/projects
```

**Response:**
```json
[
  {
    "id": "uuid",
    "name": "My Project",
    "description": "Project description"
  }
]
```

#### Get Project
```
GET /api/v1/projects/:id
```

#### Create Project
```
POST /api/v1/projects
Content-Type: application/json

{
  "name": "My Project",
  "description": "Project description"
}
```

#### Update Project
```
PUT /api/v1/projects/:id
Content-Type: application/json

{
  "name": "Updated Project",
  "description": "Updated description"
}
```

#### Delete Project
```
DELETE /api/v1/projects/:id
```

---

### Applications

#### List All Applications
```
GET /api/v1/applications
```

#### List Applications by Project
```
GET /api/v1/projects/:project_id/applications
```

#### Get Application
```
GET /api/v1/applications/:id
```

#### Create Application
```
POST /api/v1/applications
Content-Type: application/json

{
  "project_id": "uuid",
  "name": "my-app",
  "description": "Application description",
  "namespace": "default"
}
```

#### Update Application
```
PUT /api/v1/applications/:id
Content-Type: application/json

{
  "name": "my-app-updated",
  "description": "Updated description",
  "namespace": "production"
}
```

#### Delete Application
```
DELETE /api/v1/applications/:id
```

---

### Metric Types

#### List Metric Types
```
GET /api/v1/metric-types
```

**Response:**
```json
[
  {
    "id": "uuid",
    "name": "HealthCheck",
    "description": "Health check of the pod",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

#### Get Metric Type
```
GET /api/v1/metric-types/:id
```

---

### Application Metrics

#### List All Application Metrics
```
GET /api/v1/application-metrics
```

#### List Metrics by Application
```
GET /api/v1/applications/:application_id/metrics
```

#### Get Application Metric
```
GET /api/v1/application-metrics/:id
```

#### Create Application Metric

##### HealthCheck Configuration
```
POST /api/v1/application-metrics
Content-Type: application/json

{
  "application_id": "uuid",
  "type_id": "uuid",
  "configuration": {
    "health_check_url": "http://my-service.default.svc.cluster.local:8080/health",
    "method": "GET",
    "expected_status": 200,
    "timeout_seconds": 10
  }
}
```

##### PodStatus Configuration
```
POST /api/v1/application-metrics
Content-Type: application/json

{
  "application_id": "uuid",
  "type_id": "uuid",
  "configuration": {
    "pod_label_selector": "app=myapp",
    "container_name": "myapp-container"
  }
}
```

##### PodMemoryUsage Configuration
```
POST /api/v1/application-metrics
Content-Type: application/json

{
  "application_id": "uuid",
  "type_id": "uuid",
  "configuration": {
    "pod_label_selector": "app=myapp",
    "container_name": "myapp-container"
  }
}
```

##### PodCpuUsage Configuration
```
POST /api/v1/application-metrics
Content-Type: application/json

{
  "application_id": "uuid",
  "type_id": "uuid",
  "configuration": {
    "pod_label_selector": "app=myapp",
    "container_name": "myapp-container"
  }
}
```

##### PvcUsage Configuration
```
POST /api/v1/application-metrics
Content-Type: application/json

{
  "application_id": "uuid",
  "type_id": "uuid",
  "configuration": {
    "pvc_name": "my-pvc"
  }
}
```

##### PodActiveNodes Configuration
```
POST /api/v1/application-metrics
Content-Type: application/json

{
  "application_id": "uuid",
  "type_id": "uuid",
  "configuration": {
    "pod_label_selector": "app=myapp"
  }
}
```

#### Update Application Metric
```
PUT /api/v1/application-metrics/:id
Content-Type: application/json

{
  "configuration": {
    "health_check_url": "http://updated-url.default.svc.cluster.local/health"
  }
}
```

#### Delete Application Metric
```
DELETE /api/v1/application-metrics/:id
```

---

### Application Metric Values (Collected Data)

Metric values are collected automatically by the cron job every minute. These endpoints allow you to query the collected data.

#### Get Metric Value by ID
```
GET /api/v1/metric-values/:id
```

**Response:**
```json
{
  "id": "uuid",
  "application_metric_id": "uuid",
  "value": {
    "status": "up",
    "response_time_ms": 150,
    "status_code": 200
  },
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

#### List Values for a Specific Metric
```
GET /api/v1/application-metrics/:application_metric_id/values?limit=100
```

**Query Parameters:**
- `limit` (optional) - Number of records to return (default: 100, max: 1000)

**Response:**
```json
[
  {
    "id": "uuid",
    "application_metric_id": "uuid",
    "value": {
      "status": "up",
      "response_time_ms": 150,
      "status_code": 200
    },
    "created_at": "2024-01-15T10:30:00Z"
  },
  {
    "id": "uuid",
    "application_metric_id": "uuid",
    "value": {
      "status": "up",
      "response_time_ms": 145,
      "status_code": 200
    },
    "created_at": "2024-01-15T10:29:00Z"
  }
]
```

#### Get Latest Metrics for an Application
```
GET /api/v1/applications/:application_id/latest-metrics
```

Returns the latest collected value for each metric configured for the application, along with complete application, project, and metric type information.

**Response:**
```json
{
  "application_id": "uuid",
  "application_name": "web-app",
  "application_description": "Main web application",
  "application_namespace": "production",
  "project_id": "uuid",
  "project_name": "Production",
  "project_description": "Production environment applications",
  "metrics": [
    {
      "metric_id": "uuid",
      "metric_type_id": "uuid",
      "metric_type_name": "HealthCheck",
      "metric_type_description": "Health check of the pod",
      "configuration": {
        "health_check_url": "http://service.namespace.svc.cluster.local/health",
        "method": "GET",
        "expected_status": 200,
        "timeout_seconds": 10
      },
      "latest_value": {
        "id": "uuid",
        "application_metric_id": "uuid",
        "value": {
          "status": "up",
          "response_time_ms": 150,
          "status_code": 200,
          "error_message": ""
        },
        "created_at": "2024-01-15T10:30:00Z"
      }
    },
    {
      "metric_id": "uuid",
      "metric_type_id": "uuid",
      "metric_type_name": "PodStatus",
      "metric_type_description": "Status of the pod",
      "configuration": {
        "pod_label_selector": "app=myapp",
        "container_name": "web"
      },
      "latest_value": {
        "id": "uuid",
        "application_metric_id": "uuid",
        "value": {
          "pod_phase": "Running",
          "pod_ready": true,
          "restart_count": 0
        },
        "created_at": "2024-01-15T10:30:00Z"
      }
    }
  ]
}
```

---

## Metric Collection

Metrics are collected automatically every minute by a cron job running in the background. The collected metrics are stored in the `application_metric_values` table.

### Metric Value Structure

Each metric type stores different values:

#### HealthCheck
```json
{
  "status": "up",
  "response_time_ms": 150,
  "status_code": 200,
  "error_message": ""
}
```

#### PodStatus
```json
{
  "pod_phase": "Running",
  "pod_ready": true,
  "restart_count": 0
}
```

#### PodMemoryUsage
```json
{
  "memory_usage_bytes": 536870912,
  "memory_limit_bytes": 1073741824,
  "memory_percent": 50.0
}
```

#### PodCpuUsage
```json
{
  "cpu_usage_millicores": 250,
  "cpu_limit_millicores": 1000,
  "cpu_percent": 25.0
}
```

#### PvcUsage
```json
{
  "pvc_capacity_bytes": 10737418240,
  "pvc_used_bytes": 5368709120,
  "pvc_percent": 50.0
}
```

#### PodActiveNodes
```json
{
  "active_nodes_count": 3,
  "node_names": ["node-1", "node-2", "node-3"]
}
```

---

## Usage Example

### 1. Create a Project
```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production",
    "description": "Production applications"
  }'
```

### 2. Register an Application
```bash
curl -X POST http://localhost:8080/api/v1/applications \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "PROJECT_UUID",
    "name": "my-web-app",
    "description": "Web application",
    "namespace": "production"
  }'
```

### 3. Get Available Metric Types
```bash
curl http://localhost:8080/api/v1/metric-types
```

### 4. Configure Health Check Monitoring
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_UUID",
    "type_id": "HEALTHCHECK_TYPE_UUID",
    "configuration": {
      "health_check_url": "http://my-web-app.production.svc.cluster.local:8080/health",
      "method": "GET",
      "expected_status": 200,
      "timeout_seconds": 10
    }
  }'
```

### 5. Configure Pod Status Monitoring
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_UUID",
    "type_id": "PODSTATUS_TYPE_UUID",
    "configuration": {
      "pod_label_selector": "app=my-web-app",
      "container_name": "web"
    }
  }'
```

### 6. Configure Memory Monitoring
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_UUID",
    "type_id": "PODMEMORYUSAGE_TYPE_UUID",
    "configuration": {
      "pod_label_selector": "app=my-web-app",
      "container_name": "web"
    }
  }'
```

---

## Deployment

### Prerequisites
- Kubernetes cluster with metrics-server installed
- PostgreSQL database
- Proper RBAC permissions for the monitoring service

### Required RBAC Permissions
The service account needs the following permissions:
- `get`, `list`, `watch` on `pods`
- `get`, `list` on `pods/metrics`
- `get`, `list` on `persistentvolumeclaims`
- `get`, `list` on `nodes`

### Environment Variables
See the main README for required environment variables.

---

## Notes

- Metrics are collected every minute by default
- Historical metric data is stored indefinitely (consider implementing a retention policy)
- The `container_name` field in pod-related metrics is optional; if not specified, the first container is used
- PVC usage currently only stores capacity; actual usage requires additional storage system integration
- Health checks support various HTTP methods (GET, POST, PUT, etc.)
- All timestamps are stored in UTC

---

## Future Enhancements

- Alerting based on metric thresholds
- Grafana dashboard integration
- Metric retention policies
- Support for custom metric collectors
- Webhooks for metric events
- Metric aggregation and statistics

