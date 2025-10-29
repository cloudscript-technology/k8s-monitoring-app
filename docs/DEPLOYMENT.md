# K8s Monitoring App - Deployment Guide

## Prerequisites

1. **Kubernetes Cluster**: A running Kubernetes cluster (v1.20+)
2. **Metrics Server**: Installed and running in the cluster
3. **SQLite Storage**: Optional PersistentVolumeClaim for persisting the SQLite DB file
4. **kubectl**: Configured to access your cluster
5. **Helm** (optional): For easier deployment using the provided chart

## Required Kubernetes Permissions

The monitoring application requires specific RBAC permissions to interact with the Kubernetes API.

### Create Service Account and RBAC

```yaml
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
  # Pod access
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["get", "list", "watch"]
  
  # Pod metrics access
  - apiGroups: ["metrics.k8s.io"]
    resources: ["pods"]
    verbs: ["get", "list"]
  
  # PVC access
  - apiGroups: [""]
    resources: ["persistentvolumeclaims"]
    verbs: ["get", "list"]
  
  # Pod exec access (for PVC usage monitoring with df)
  - apiGroups: [""]
    resources: ["pods/exec"]
    verbs: ["create"]
  
  # Node access
  - apiGroups: [""]
    resources: ["nodes"]
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
```

Apply the RBAC configuration:
```bash
kubectl apply -f rbac.yaml
```

## Environment Variables

The application requires the following environment variables:

```bash
# Database Configuration (SQLite)
DB_PATH=/data/k8s_monitoring.db

# APM Configuration (optional)
ELASTIC_APM_SERVICE_NAME=k8s-monitoring-app
ELASTIC_APM_SERVER_URL=http://apm-server:8200
ELASTIC_APM_ENVIRONMENT=production

# Logging
LOG_LEVEL=info
```

## Deployment Options

### Option 1: Using Kubernetes Manifests

#### 1. Create ConfigMap for Environment Variables

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8s-monitoring-app-config
  namespace: default
data:
  DB_PATH: "/data/k8s_monitoring.db"
  LOG_LEVEL: "info"
```

#### 2. Create PersistentVolumeClaim (for SQLite persistence)

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: k8s-monitoring-app-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
```

#### 3. Create Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-monitoring-app
  namespace: default
  labels:
    app: k8s-monitoring-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: k8s-monitoring-app
  template:
    metadata:
      labels:
        app: k8s-monitoring-app
    spec:
      serviceAccountName: k8s-monitoring-app
      containers:
      - name: k8s-monitoring-app
        image: your-registry/k8s-monitoring-app:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
          name: http
        envFrom:
        - configMapRef:
            name: k8s-monitoring-app-config
        volumeMounts:
        - name: data
          mountPath: /data
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: k8s-monitoring-app-pvc
```

#### 4. Create Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: k8s-monitoring-app
  namespace: default
spec:
  selector:
    app: k8s-monitoring-app
  ports:
  - port: 8080
    targetPort: 8080
    name: http
  type: ClusterIP
```

#### 5. Create Ingress (optional)

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: k8s-monitoring-app
  namespace: default
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
  - hosts:
    - monitoring.example.com
    secretName: k8s-monitoring-app-tls
  rules:
  - host: monitoring.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: k8s-monitoring-app
            port:
              number: 8080
```

Apply all manifests:
```bash
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f ingress.yaml
```

### Option 2: Using Helm

The application includes a Helm chart in the `chart/` directory.

#### 1. Update values.yaml

Edit `chart/values.yaml` to configure your deployment:

```yaml
replicaCount: 1

image:
  repository: your-registry/k8s-monitoring-app
  pullPolicy: Always
  tag: "latest"

serviceAccount:
  create: true
  name: k8s-monitoring-app

service:
  type: ClusterIP
  port: 8080

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: monitoring.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: k8s-monitoring-app-tls
      hosts:
        - monitoring.example.com

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 256Mi

env:
  - name: DB_PATH
    value: "/data/k8s_monitoring.db"

volumeMounts:
  - name: data
    mountPath: /data

volumes:
  - name: data
    persistentVolumeClaim:
      claimName: k8s-monitoring-app-pvc
```

#### 2. Install with Helm

```bash
helm install k8s-monitoring-app ./chart \
  --namespace default \
  --create-namespace
```

#### 3. Upgrade with Helm

```bash
helm upgrade k8s-monitoring-app ./chart \
  --namespace default
```

## Database Setup

### SQLite Initialization

The application uses SQLite with automatic migrations. No external database server is required. The database file will be created at the path configured in `DB_PATH` (e.g., `/data/k8s_monitoring.db`).

### Migration Path

Migrations are located in `database/migrations/` and will be applied automatically on startup.

## Verify Deployment

### Check Pods
```bash
kubectl get pods -l app=k8s-monitoring-app
```

### Check Logs
```bash
kubectl logs -l app=k8s-monitoring-app -f
```

### Test Health Endpoint
```bash
kubectl port-forward svc/k8s-monitoring-app 8080:8080
curl http://localhost:8080/health
```

### Check Monitoring Job
Look for log entries indicating metric collection:
```bash
kubectl logs -l app=k8s-monitoring-app | grep "metric collection"
```

## Metrics Server Verification

Ensure metrics-server is running:
```bash
kubectl get deployment metrics-server -n kube-system
```

Test metrics availability:
```bash
kubectl top pods
```

If metrics-server is not installed:
```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

## Troubleshooting

### Pod Cannot Access Kubernetes API

**Symptom**: Errors like "failed to create in-cluster config"

**Solution**: 
- Verify ServiceAccount is correctly configured
- Check RBAC permissions
- Ensure the pod is using the correct ServiceAccount

### Metrics Not Being Collected

**Symptom**: No data in application_metric_values table

**Solution**:
- Check if metrics-server is running
- Verify RBAC permissions for metrics.k8s.io
- Check cron job logs
- Verify application metrics are configured correctly

### Database Connection Issues

**Symptom**: "failed to open SQLite database"

**Solution**:
- Verify `DB_PATH` is set correctly
- Ensure the pod has write permissions to the mounted path
- Confirm the PVC is bound and mounted
- Check that migrations run without errors in logs

### Memory/CPU Metrics Not Available

**Symptom**: "failed to get pod metrics"

**Solution**:
- Verify metrics-server is running
- Check if pods have resource requests/limits defined
- Wait a few minutes for metrics to be available

## Security Considerations

1. **SQLite File Access**: Limit access to the mounted volume via RBAC and PVC policies
2. **RBAC**: Use the minimal required permissions
3. **Network Policies**: Restrict access to the monitoring app
4. **TLS**: Enable TLS for ingress endpoints
5. **Service Account**: Use a dedicated ServiceAccount with limited scope

## Scaling

The application currently runs as a single replica due to the cron job nature. For high availability:

1. Consider using leader election for cron jobs
2. Scale the API server separately from the monitoring worker
3. Use a distributed cron solution like Kubernetes CronJobs

## Monitoring the Monitor

### Prometheus Metrics

Consider exposing metrics about the monitoring application itself:
- Number of metrics collected
- Collection duration
- Error rates
- Database connection pool stats

### Alerting

Set up alerts for:
- Application not responding to health checks
- Database connection failures
- Metric collection failures
- High error rates

## Backup and Recovery

### Database Backups
```bash
# Backup SQLite DB file from pod
POD=$(kubectl get pods -l app=k8s-monitoring-app -o jsonpath='{.items[0].metadata.name}')
kubectl cp default/$POD:/data/k8s_monitoring.db ./backup/k8s_monitoring.db

# Restore SQLite DB file to pod
kubectl cp ./backup/k8s_monitoring.db default/$POD:/data/k8s_monitoring.db
```

### Configuration Backup
```bash
# Export all resources
kubectl get all,cm,secret,ingress -l app=k8s-monitoring-app -o yaml > backup.yaml
```

## Cleanup

### Remove Application
```bash
# Using kubectl
kubectl delete -f deployment.yaml
kubectl delete -f service.yaml
kubectl delete -f ingress.yaml
kubectl delete -f configmap.yaml
kubectl delete -f secret.yaml
kubectl delete -f rbac.yaml

# Using Helm
helm uninstall k8s-monitoring-app --namespace default
```

### Remove Database Data
```bash
# Delete PVC to remove persisted data (careful!)
kubectl delete pvc k8s-monitoring-app-pvc -n default
```
