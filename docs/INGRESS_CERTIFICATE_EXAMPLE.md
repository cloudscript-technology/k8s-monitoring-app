# Ingress Certificate Monitoring

## Overview

The `IngressCertificate` metric type monitors TLS certificate expiration for Kubernetes Ingress resources. This is critical for production environments to avoid certificate expiration incidents.

## Features

- ✅ Automatic certificate expiration tracking
- ✅ Configurable warning threshold (default: 30 days)
- ✅ Supports multiple domains per certificate
- ✅ Displays certificate issuer and subject
- ✅ Visual status indicators (valid, expiring soon, expired)
- ✅ Auto-discovers TLS secret from Ingress if not specified

## Configuration

### Required Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `ingress_name` | string | Name of the Ingress resource | `"my-app-ingress"` |

### Optional Fields

| Field | Type | Description | Default | Example |
|-------|------|-------------|---------|---------|
| `ingress_namespace` | string | Namespace (if different from application namespace) | Application namespace | `"production"` |
| `tls_secret_name` | string | TLS secret name | Auto-discovered from Ingress | `"my-app-tls"` |
| `warning_days` | int | Days before expiration to show warning | `30` | `15` |

## Example: Add Certificate Monitoring

### Step 1: Create a Project (if not exists)

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Apps",
    "description": "Production environment applications"
  }'
```

### Step 2: Register Your Application

```bash
curl -X POST http://localhost:8080/api/v1/applications \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "PROJECT_ID_HERE",
    "name": "my-app",
    "description": "My Application",
    "namespace": "production"
  }'
```

### Step 3: Get the IngressCertificate Metric Type ID

```bash
curl http://localhost:8080/api/v1/metric-types | jq '.[] | select(.name=="IngressCertificate")'
```

Example response:
```json
{
  "id": "abc123-...",
  "name": "IngressCertificate",
  "description": "TLS certificate expiration monitoring for Ingress resources"
}
```

### Step 4: Add Certificate Monitoring Metric

#### Basic Configuration (Auto-discover TLS secret)

```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_ID_HERE",
    "type_id": "INGRESS_CERTIFICATE_METRIC_TYPE_ID",
    "configuration": {
      "ingress_name": "my-app-ingress"
    }
  }'
```

#### Full Configuration

```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "APPLICATION_ID_HERE",
    "type_id": "INGRESS_CERTIFICATE_METRIC_TYPE_ID",
    "configuration": {
      "ingress_name": "my-app-ingress",
      "ingress_namespace": "production",
      "tls_secret_name": "my-app-tls",
      "warning_days": 15
    }
  }'
```

## Metric Values

The collected metric values include:

| Field | Type | Description | Example Values |
|-------|------|-------------|----------------|
| `certificate_status` | string | Current certificate status | `"valid"`, `"expiring_soon"`, `"expired"`, `"not_found"`, `"error"` |
| `certificate_expiration` | timestamp | Certificate expiration date | `"2025-12-31T23:59:59Z"` |
| `certificate_days_to_expire` | int | Days until expiration (negative if expired) | `45`, `10`, `-5` |
| `certificate_issuer` | string | Certificate issuer CN | `"Let's Encrypt Authority X3"` |
| `certificate_subject` | string | Certificate subject CN | `"my-app.example.com"` |
| `certificate_domains` | array | DNS names in certificate | `["my-app.example.com", "www.my-app.example.com"]` |
| `certificate_error` | string | Error message if any | `"TLS secret not found"` |

### Example Metric Value

```json
{
  "id": "value-123",
  "application_metric_id": "metric-456",
  "value": {
    "certificate_status": "expiring_soon",
    "certificate_expiration": "2025-11-15T10:30:00Z",
    "certificate_days_to_expire": 18,
    "certificate_issuer": "Let's Encrypt Authority X3",
    "certificate_subject": "my-app.example.com",
    "certificate_domains": [
      "my-app.example.com",
      "www.my-app.example.com"
    ]
  },
  "created_at": "2025-10-28T12:00:00Z"
}
```

## Status Values

### ✅ valid

Certificate is valid and has more than `warning_days` until expiration.

**Frontend Display:**
```
🔒 certificate
┌─────────────────────────────────┐
│ my-app.example.com              │
│ www.my-app.example.com          │
│ ✓ Valid for 45 days             │
└─────────────────────────────────┘
```

### ⚠ expiring_soon

Certificate will expire within `warning_days`.

**Frontend Display:**
```
🔒 certificate
┌─────────────────────────────────┐
│ my-app.example.com              │
│ ⚠ Expires in 10 days            │
└─────────────────────────────────┘
```

### ❌ expired

Certificate has already expired.

**Frontend Display:**
```
🔒 certificate
┌─────────────────────────────────┐
│ my-app.example.com              │
│ ✗ Expired 5 days ago            │
└─────────────────────────────────┘
```

### ⚠ not_found

Ingress or TLS secret not found.

**Frontend Display:**
```
🔒 certificate
┌─────────────────────────────────┐
│ my-app-ingress                  │
│ ⚠ Not found                     │
└─────────────────────────────────┘
```

### ❌ error

Error parsing certificate or accessing resources.

**Frontend Display:**
```
🔒 certificate
┌─────────────────────────────────┐
│ my-app-ingress                  │
│ ✗ Error                         │
│ tls.crt not found in secret     │
└─────────────────────────────────┘
```

## Real-World Examples

### Example 1: Standard HTTPS Application

```json
{
  "application_id": "app-123",
  "type_id": "ingress-cert-type-id",
  "configuration": {
    "ingress_name": "webapp-ingress",
    "warning_days": 30
  }
}
```

**Result:**
- Monitors: `webapp-ingress` in application's namespace
- Auto-discovers TLS secret from Ingress spec
- Warns when certificate has less than 30 days remaining

### Example 2: Multi-Domain with Custom Threshold

```json
{
  "application_id": "app-456",
  "type_id": "ingress-cert-type-id",
  "configuration": {
    "ingress_name": "api-ingress",
    "warning_days": 15
  }
}
```

**Result:**
- Monitors: `api-ingress` with 15-day warning threshold
- Tracks all domains in the certificate (e.g., `api.example.com`, `api-v2.example.com`)

### Example 3: Specific TLS Secret

```json
{
  "application_id": "app-789",
  "type_id": "ingress-cert-type-id",
  "configuration": {
    "ingress_name": "admin-ingress",
    "tls_secret_name": "admin-wildcard-cert",
    "warning_days": 45
  }
}
```

**Result:**
- Monitors specific secret: `admin-wildcard-cert`
- Useful when Ingress has multiple TLS configurations
- Warns 45 days before expiration

### Example 4: Cross-Namespace Monitoring

```json
{
  "application_id": "app-999",
  "type_id": "ingress-cert-type-id",
  "configuration": {
    "ingress_name": "shared-ingress",
    "ingress_namespace": "ingress-nginx",
    "tls_secret_name": "shared-wildcard",
    "warning_days": 60
  }
}
```

**Result:**
- Monitors Ingress in different namespace: `ingress-nginx`
- Checks specific secret: `shared-wildcard`
- Extra-long warning period: 60 days

## Kubernetes Prerequisites

### Required RBAC Permissions

The monitoring service needs permissions to read:
- Ingresses in the monitored namespaces
- Secrets in the monitored namespaces

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: k8s-monitoring-app
  namespace: production
rules:
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses"]
  verbs: ["get", "list"]
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get"]
```

### Example Ingress with TLS

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-app-ingress
  namespace: production
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - my-app.example.com
    - www.my-app.example.com
    secretName: my-app-tls
  rules:
  - host: my-app.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: my-app
            port:
              number: 80
```

## Integration with cert-manager

This metric works perfectly with cert-manager automated certificates:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: my-app-cert
  namespace: production
spec:
  secretName: my-app-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
  - my-app.example.com
  - www.my-app.example.com
```

The monitoring app will:
1. Track the certificate in `my-app-tls` secret
2. Warn you before expiration (even though cert-manager auto-renews)
3. Alert if renewal fails

## Best Practices

### Warning Thresholds

| Environment | Recommended Warning Days | Reasoning |
|-------------|-------------------------|-----------|
| Development | 7-15 days | Certificates change frequently |
| Staging | 15-30 days | Match production setup |
| Production | 30-60 days | Extra time to respond to issues |
| Critical Apps | 60-90 days | Maximum advance warning |

### Monitoring Strategy

1. **One metric per Ingress**: Create separate metrics for each Ingress
2. **Check frequently**: Default 60-second collection interval is good
3. **Set appropriate warnings**: Consider your cert renewal process
4. **Monitor shared certificates**: Don't forget wildcard or shared certs

### Troubleshooting

#### Certificate not found

```json
{
  "certificate_status": "not_found",
  "certificate_error": "ingress not found: ingresses.networking.k8s.io \"my-app\" not found"
}
```

**Solutions:**
- Check Ingress name spelling
- Verify namespace is correct
- Ensure RBAC permissions are set

#### TLS secret not found

```json
{
  "certificate_status": "not_found",
  "certificate_error": "TLS secret not found: secrets \"my-tls\" not found"
}
```

**Solutions:**
- Check if Ingress has TLS configuration
- Verify secret name in Ingress spec
- Check if cert-manager created the secret
- Ensure RBAC allows reading secrets

#### Parse error

```json
{
  "certificate_status": "error",
  "certificate_error": "failed to parse certificate: x509: malformed certificate"
}
```

**Solutions:**
- Check if secret contains valid certificate data
- Verify `tls.crt` key exists in secret
- Ensure certificate is PEM-encoded

## API Endpoints

### Get Latest Certificate Status

```bash
curl http://localhost:8080/api/v1/applications/{app_id}/latest-metrics
```

### Get Certificate History

```bash
curl http://localhost:8080/api/v1/application-metrics/{metric_id}/values?limit=10
```

### View in Web UI

Navigate to: `http://localhost:8080`

The certificate status will be displayed in the application card with:
- 🔒 Icon for certificate section
- Color-coded status (green/yellow/red)
- Domain names
- Days to expiration
- Tooltip with full details (issuer, expiration date)

## See Also

- [API Documentation](API.md)
- [Metric Types](METRIC_TYPES.md)
- [Local Development](LOCAL_DEVELOPMENT.md)
- [Kubernetes Integration](KUBERNETES.md)

