# K8s Monitoring App - Usage Examples

This document provides practical examples for using the K8s Monitoring App.

## Complete Workflow Example

### Step 1: Set Up Your Environment

```bash
# Port forward to access the API locally
kubectl port-forward svc/k8s-monitoring-app 8080:8080

# Set the API endpoint
export API_URL=http://localhost:8080/api/v1
```

### Step 2: Create a Project

Projects help organize your applications.

```bash
# Create a production project
curl -X POST $API_URL/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production",
    "description": "Production environment applications"
  }'

# Response:
# {
#   "id": "550e8400-e29b-41d4-a716-446655440000",
#   "name": "Production",
#   "description": "Production environment applications"
# }

# Save the project ID
export PROJECT_ID="550e8400-e29b-41d4-a716-446655440000"
```

### Step 3: Register Your Application

```bash
# Register a web application
curl -X POST $API_URL/applications \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": \"$PROJECT_ID\",
    \"name\": \"web-app\",
    \"description\": \"Main web application\",
    \"namespace\": \"production\"
  }"

# Response:
# {
#   "id": "650e8400-e29b-41d4-a716-446655440001",
#   "project_id": "550e8400-e29b-41d4-a716-446655440000",
#   "name": "web-app",
#   "description": "Main web application",
#   "namespace": "production",
#   "created_at": "2024-01-15T10:00:00Z",
#   "updated_at": "2024-01-15T10:00:00Z"
# }

# Save the application ID
export APP_ID="650e8400-e29b-41d4-a716-446655440001"
```

### Step 4: Get Available Metric Types

```bash
# List all metric types
curl $API_URL/metric-types | jq '.'

# Response:
# [
#   {
#     "id": "750e8400-e29b-41d4-a716-446655440002",
#     "name": "HealthCheck",
#     "description": "Health check of the pod",
#     "created_at": "2024-01-15T10:00:00Z",
#     "updated_at": "2024-01-15T10:00:00Z"
#   },
#   ...
# ]

# Extract specific metric type IDs
export HEALTH_CHECK_ID=$(curl -s $API_URL/metric-types | jq -r '.[] | select(.name=="HealthCheck") | .id')
export POD_STATUS_ID=$(curl -s $API_URL/metric-types | jq -r '.[] | select(.name=="PodStatus") | .id')
export MEMORY_ID=$(curl -s $API_URL/metric-types | jq -r '.[] | select(.name=="PodMemoryUsage") | .id')
export CPU_ID=$(curl -s $API_URL/metric-types | jq -r '.[] | select(.name=="PodCpuUsage") | .id')
```

### Step 5: Configure Metrics

#### Configure Health Check

```bash
curl -X POST $API_URL/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$APP_ID\",
    \"type_id\": \"$HEALTH_CHECK_ID\",
    \"configuration\": {
      \"health_check_url\": \"http://web-app.production.svc.cluster.local:8080/health\",
      \"method\": \"GET\",
      \"expected_status\": 200,
      \"timeout_seconds\": 10
    }
  }"
```

#### Configure Pod Status Monitoring

```bash
curl -X POST $API_URL/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$APP_ID\",
    \"type_id\": \"$POD_STATUS_ID\",
    \"configuration\": {
      \"pod_label_selector\": \"app=web-app\",
      \"container_name\": \"web\"
    }
  }"
```

#### Configure Memory Monitoring

```bash
curl -X POST $API_URL/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$APP_ID\",
    \"type_id\": \"$MEMORY_ID\",
    \"configuration\": {
      \"pod_label_selector\": \"app=web-app\",
      \"container_name\": \"web\"
    }
  }"
```

#### Configure CPU Monitoring

```bash
curl -X POST $API_URL/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$APP_ID\",
    \"type_id\": \"$CPU_ID\",
    \"configuration\": {
      \"pod_label_selector\": \"app=web-app\",
      \"container_name\": \"web\"
    }
  }"
```

### Step 6: View Configured Metrics

```bash
# List all metrics for the application
curl $API_URL/applications/$APP_ID/metrics | jq '.'
```

## Example Scenarios

### Scenario 1: Monitoring a Microservices Application

```bash
# Create project
PROJECT_ID=$(curl -s -X POST $API_URL/projects \
  -H "Content-Type: application/json" \
  -d '{"name":"Microservices","description":"E-commerce microservices"}' | jq -r '.id')

# Register services
API_SERVICE_ID=$(curl -s -X POST $API_URL/applications \
  -H "Content-Type: application/json" \
  -d "{\"project_id\":\"$PROJECT_ID\",\"name\":\"api-service\",\"description\":\"API Gateway\",\"namespace\":\"production\"}" | jq -r '.id')

USER_SERVICE_ID=$(curl -s -X POST $API_URL/applications \
  -H "Content-Type: application/json" \
  -d "{\"project_id\":\"$PROJECT_ID\",\"name\":\"user-service\",\"description\":\"User Management\",\"namespace\":\"production\"}" | jq -r '.id')

ORDER_SERVICE_ID=$(curl -s -X POST $API_URL/applications \
  -H "Content-Type: application/json" \
  -d "{\"project_id\":\"$PROJECT_ID\",\"name\":\"order-service\",\"description\":\"Order Processing\",\"namespace\":\"production\"}" | jq -r '.id')

# Get metric types
HEALTH_CHECK_ID=$(curl -s $API_URL/metric-types | jq -r '.[] | select(.name=="HealthCheck") | .id')

# Configure health checks for all services
for SERVICE_ID in $API_SERVICE_ID $USER_SERVICE_ID $ORDER_SERVICE_ID; do
  SERVICE_NAME=$(curl -s $API_URL/applications/$SERVICE_ID | jq -r '.name')
  
  curl -X POST $API_URL/application-metrics \
    -H "Content-Type: application/json" \
    -d "{
      \"application_id\": \"$SERVICE_ID\",
      \"type_id\": \"$HEALTH_CHECK_ID\",
      \"configuration\": {
        \"health_check_url\": \"http://$SERVICE_NAME.production.svc.cluster.local:8080/health\",
        \"method\": \"GET\",
        \"expected_status\": 200,
        \"timeout_seconds\": 10
      }
    }"
done
```

### Scenario 2: Monitoring a Stateful Application with PVC

```bash
# Register database application
DB_APP_ID=$(curl -s -X POST $API_URL/applications \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": \"$PROJECT_ID\",
    \"name\": \"postgres-db\",
    \"description\": \"PostgreSQL Database\",
    \"namespace\": \"production\"
  }" | jq -r '.id')

# Get metric type IDs
PVC_ID=$(curl -s $API_URL/metric-types | jq -r '.[] | select(.name=="PvcUsage") | .id')
MEMORY_ID=$(curl -s $API_URL/metric-types | jq -r '.[] | select(.name=="PodMemoryUsage") | .id')
POD_STATUS_ID=$(curl -s $API_URL/metric-types | jq -r '.[] | select(.name=="PodStatus") | .id')

# Configure PVC monitoring
curl -X POST $API_URL/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$DB_APP_ID\",
    \"type_id\": \"$PVC_ID\",
    \"configuration\": {
      \"pvc_name\": \"postgres-data-pvc\",
      \"pod_label_selector\": \"app=postgres\",
      \"container_name\": \"postgres\"
    }
  }"

# Configure memory monitoring
curl -X POST $API_URL/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$DB_APP_ID\",
    \"type_id\": \"$MEMORY_ID\",
    \"configuration\": {
      \"pod_label_selector\": \"app=postgres\",
      \"container_name\": \"postgres\"
    }
  }"

# Configure pod status
curl -X POST $API_URL/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$DB_APP_ID\",
    \"type_id\": \"$POD_STATUS_ID\",
    \"configuration\": {
      \"pod_label_selector\": \"app=postgres\",
      \"container_name\": \"postgres\"
    }
  }"
```

### Scenario 3: Monitoring Node Distribution

```bash
# Register a distributed application
APP_ID=$(curl -s -X POST $API_URL/applications \
  -H "Content-Type: application/json" \
  -d "{
    \"project_id\": \"$PROJECT_ID\",
    \"name\": \"cache-cluster\",
    \"description\": \"Redis Cache Cluster\",
    \"namespace\": \"production\"
  }" | jq -r '.id')

# Get metric type
NODES_ID=$(curl -s $API_URL/metric-types | jq -r '.[] | select(.name=="PodActiveNodes") | .id')

# Configure node tracking
curl -X POST $API_URL/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"$APP_ID\",
    \"type_id\": \"$NODES_ID\",
    \"configuration\": {
      \"pod_label_selector\": \"app=redis\"
    }
  }"
```

## Querying Metric Data

### Direct Database Query Examples

Once metrics are collected, you can query them from the database:

```sql
-- Get latest health check status for all applications
SELECT 
  a.name as application_name,
  mt.name as metric_type,
  amv.value->>'status' as status,
  amv.value->>'response_time_ms' as response_time_ms,
  amv.created_at
FROM application_metric_values amv
JOIN application_metrics am ON amv.application_metric_id = am.id
JOIN applications a ON am.application_id = a.id
JOIN metric_types mt ON am.type_id = mt.id
WHERE mt.name = 'HealthCheck'
ORDER BY amv.created_at DESC
LIMIT 10;

-- Get average memory usage over last hour
SELECT 
  a.name as application_name,
  AVG((amv.value->>'memory_usage_bytes')::bigint) as avg_memory_bytes,
  AVG((amv.value->>'memory_percent')::numeric) as avg_memory_percent
FROM application_metric_values amv
JOIN application_metrics am ON amv.application_metric_id = am.id
JOIN applications a ON am.application_id = a.id
JOIN metric_types mt ON am.type_id = mt.id
WHERE mt.name = 'PodMemoryUsage'
  AND amv.created_at > NOW() - INTERVAL '1 hour'
GROUP BY a.name;

-- Get applications with high restart counts
SELECT 
  a.name as application_name,
  amv.value->>'restart_count' as restart_count,
  amv.value->>'pod_phase' as pod_phase,
  amv.created_at
FROM application_metric_values amv
JOIN application_metrics am ON amv.application_metric_id = am.id
JOIN applications a ON am.application_id = a.id
JOIN metric_types mt ON am.type_id = mt.id
WHERE mt.name = 'PodStatus'
  AND (amv.value->>'restart_count')::int > 0
ORDER BY amv.created_at DESC;
```

## Updating Metric Configuration

```bash
# Get the metric ID
METRIC_ID="your-metric-id"

# Update health check URL
curl -X PUT $API_URL/application-metrics/$METRIC_ID \
  -H "Content-Type: application/json" \
  -d '{
    "configuration": {
      "health_check_url": "http://new-service.production.svc.cluster.local:8080/healthz",
      "method": "GET",
      "expected_status": 200,
      "timeout_seconds": 5
    }
  }'
```

## Deleting Configurations

```bash
# Delete a specific metric configuration
curl -X DELETE $API_URL/application-metrics/$METRIC_ID

# Delete an application (will cascade delete all metrics)
curl -X DELETE $API_URL/applications/$APP_ID

# Delete a project (will cascade delete all applications and metrics)
curl -X DELETE $API_URL/projects/$PROJECT_ID
```

## Bulk Configuration Script

```bash
#!/bin/bash

API_URL="http://localhost:8080/api/v1"

# Create project
PROJECT_ID=$(curl -s -X POST $API_URL/projects \
  -H "Content-Type: application/json" \
  -d '{"name":"Staging","description":"Staging environment"}' | jq -r '.id')

# Define applications
declare -A APPS=(
  ["frontend"]="production"
  ["backend"]="production"
  ["worker"]="production"
  ["scheduler"]="production"
)

# Get metric types
HEALTH_CHECK_ID=$(curl -s $API_URL/metric-types | jq -r '.[] | select(.name=="HealthCheck") | .id')
POD_STATUS_ID=$(curl -s $API_URL/metric-types | jq -r '.[] | select(.name=="PodStatus") | .id')
MEMORY_ID=$(curl -s $API_URL/metric-types | jq -r '.[] | select(.name=="PodMemoryUsage") | .id')
CPU_ID=$(curl -s $API_URL/metric-types | jq -r '.[] | select(.name=="PodCpuUsage") | .id')

# Register and configure each application
for APP_NAME in "${!APPS[@]}"; do
  NAMESPACE="${APPS[$APP_NAME]}"
  
  echo "Configuring $APP_NAME in $NAMESPACE..."
  
  # Register application
  APP_ID=$(curl -s -X POST $API_URL/applications \
    -H "Content-Type: application/json" \
    -d "{
      \"project_id\": \"$PROJECT_ID\",
      \"name\": \"$APP_NAME\",
      \"description\": \"$APP_NAME service\",
      \"namespace\": \"$NAMESPACE\"
    }" | jq -r '.id')
  
  # Configure health check
  curl -s -X POST $API_URL/application-metrics \
    -H "Content-Type: application/json" \
    -d "{
      \"application_id\": \"$APP_ID\",
      \"type_id\": \"$HEALTH_CHECK_ID\",
      \"configuration\": {
        \"health_check_url\": \"http://$APP_NAME.$NAMESPACE.svc.cluster.local:8080/health\",
        \"method\": \"GET\",
        \"expected_status\": 200,
        \"timeout_seconds\": 10
      }
    }" > /dev/null
  
  # Configure pod status
  curl -s -X POST $API_URL/application-metrics \
    -H "Content-Type: application/json" \
    -d "{
      \"application_id\": \"$APP_ID\",
      \"type_id\": \"$POD_STATUS_ID\",
      \"configuration\": {
        \"pod_label_selector\": \"app=$APP_NAME\"
      }
    }" > /dev/null
  
  # Configure memory monitoring
  curl -s -X POST $API_URL/application-metrics \
    -H "Content-Type: application/json" \
    -d "{
      \"application_id\": \"$APP_ID\",
      \"type_id\": \"$MEMORY_ID\",
      \"configuration\": {
        \"pod_label_selector\": \"app=$APP_NAME\"
      }
    }" > /dev/null
  
  # Configure CPU monitoring
  curl -s -X POST $API_URL/application-metrics \
    -H "Content-Type: application/json" \
    -d "{
      \"application_id\": \"$APP_ID\",
      \"type_id\": \"$CPU_ID\",
      \"configuration\": {
        \"pod_label_selector\": \"app=$APP_NAME\"
      }
    }" > /dev/null
  
  echo "✓ $APP_NAME configured"
done

echo "All applications configured successfully!"
```

## Testing Metric Collection

```bash
# Watch the monitoring service logs
kubectl logs -f -l app=k8s-monitoring-app | grep "metric collection"

# Check if metrics are being stored
kubectl exec -it postgres-pod -- psql -U monitoring k8s_monitoring -c \
  "SELECT COUNT(*) FROM application_metric_values WHERE created_at > NOW() - INTERVAL '5 minutes';"
```

## Troubleshooting Examples

### Check if Application is Properly Registered

```bash
# List all applications
curl $API_URL/applications | jq '.[] | {name: .name, namespace: .namespace, id: .id}'

# Get specific application details
curl $API_URL/applications/$APP_ID | jq '.'
```

### Verify Metric Configuration

```bash
# List metrics for an application
curl $API_URL/applications/$APP_ID/metrics | jq '.[] | {type_id: .type_id, configuration: .configuration}'
```

### Check Latest Collected Metrics

```sql
-- Connect to database
kubectl exec -it postgres-pod -- psql -U monitoring k8s_monitoring

-- Check latest metrics
SELECT 
  a.name,
  mt.name as metric_type,
  amv.value,
  amv.created_at
FROM application_metric_values amv
JOIN application_metrics am ON amv.application_metric_id = am.id
JOIN applications a ON am.application_id = a.id
JOIN metric_types mt ON am.type_id = mt.id
ORDER BY amv.created_at DESC
LIMIT 10;
```

