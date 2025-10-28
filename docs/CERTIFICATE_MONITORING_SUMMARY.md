# 🔒 Certificate Monitoring - Complete Implementation Summary

## ✅ Implemented Features

### 1. Database Schema
- ✅ New metric type: `IngressCertificate`
- ✅ Migration file created: `1761653454_add_ingress_certificate_metric.up.sql`

### 2. Data Models

#### Configuration Fields
```go
type Configuration struct {
    IngressName      string `json:"ingress_name"`        // Required
    IngressNamespace string `json:"ingress_namespace"`   // Optional (defaults to app namespace)
    TLSSecretName    string `json:"tls_secret_name"`     // Optional (auto-discovered)
    WarningDays      int    `json:"warning_days"`        // Optional (default: 30)
}
```

#### Metric Value Fields
```go
type MetricValue struct {
    CertificateStatus       string    `json:"certificate_status"`        // "valid", "expiring_soon", "expired", "not_found", "error"
    CertificateExpiration   time.Time `json:"certificate_expiration"`    // Expiration date
    CertificateDaysToExpire int       `json:"certificate_days_to_expire"` // Days until expiration
    CertificateIssuer       string    `json:"certificate_issuer"`        // Issuer CN
    CertificateSubject      string    `json:"certificate_subject"`       // Subject CN
    CertificateDomains      []string  `json:"certificate_domains"`       // DNS names
    CertificateError        string    `json:"certificate_error"`         // Error message
}
```

### 3. Kubernetes Client

#### New Functions
- `GetIngressCertificateInfo()`: Retrieves certificate info from Ingress
- `parseCertificate()`: Parses PEM-encoded certificates
- `extractDomainsFromIngress()`: Extracts hostnames from Ingress

#### Features
- ✅ Reads Ingress resources
- ✅ Accesses TLS secrets
- ✅ Parses x509 certificates
- ✅ Auto-discovers TLS secret if not specified
- ✅ Extracts DNS SANs and Common Name
- ✅ Calculates days until expiration
- ✅ Determines certificate status

### 4. Monitoring Service

#### Collection Logic
```go
func (m *MonitoringService) collectIngressCertificate(
    ctx context.Context,
    application *applicationModel.Application,
    config *applicationMetricModel.Configuration,
) (applicationMetricValueModel.MetricValue, error)
```

#### Status Determination
- **valid**: Certificate has more than `warning_days` until expiration
- **expiring_soon**: Certificate expires within `warning_days`
- **expired**: Certificate has already expired
- **not_found**: Ingress or TLS secret not found
- **error**: Error parsing certificate or accessing resources

### 5. Frontend Visualization

#### UI Components
```html
<div class="metric-box certificate-box">
    <div class="metric-label">🔒 certificate</div>
    <!-- Domain badges -->
    <!-- Status indicator with icon and days count -->
    <!-- Tooltip with full details -->
</div>
```

#### Status Display

| Status | Icon | Color | Example Text |
|--------|------|-------|--------------|
| valid | ✓ | Green | "Valid for 45 days" |
| expiring_soon | ⚠ | Yellow/Orange | "Expires in 10 days" |
| expired | ✗ | Red | "Expired 5 days ago" |
| not_found | ⚠ | Red | "Not found" |
| error | ✗ | Red | "Error" |

#### CSS Styles
- `.certificate-container`: Main container
- `.certificate-domains`: Domain badges
- `.domain-badge`: Individual domain styling
- `.certificate-status`: Status indicator
- `.certificate-status-valid`: Green valid state
- `.certificate-status-warning`: Yellow warning state
- `.certificate-status-expired`: Red expired state
- `.certificate-status-error`: Red error state
- `.certificate-box`: Blue gradient background

### 6. Documentation
- ✅ `INGRESS_CERTIFICATE_EXAMPLE.md`: Complete usage guide
- ✅ `CERTIFICATE_MONITORING_SUMMARY.md`: Implementation summary

## 📋 Files Modified/Created

### Created Files
1. `database/migrations/1761653454_add_ingress_certificate_metric.up.sql`
2. `database/migrations/1761653454_add_ingress_certificate_metric.down.sql`
3. `docs/INGRESS_CERTIFICATE_EXAMPLE.md`
4. `docs/CERTIFICATE_MONITORING_SUMMARY.md`

### Modified Files
1. `pkg/application_metric/model/model.go`
   - Added `IngressName`, `IngressNamespace`, `TLSSecretName`, `WarningDays` to `Configuration`

2. `pkg/application_metric_value/model/model.go`
   - Added certificate fields to `MetricValue` struct

3. `internal/k8s/client.go`
   - Added imports: `crypto/x509`, `encoding/pem`, `networkingv1`
   - Added `IngressCertificateInfo` struct
   - Added `GetIngressCertificateInfo()` function
   - Added `parseCertificate()` helper
   - Added `extractDomainsFromIngress()` helper

4. `internal/monitoring/service.go`
   - Added `case "IngressCertificate"` in `collectMetricByType()`
   - Added `collectIngressCertificate()` function

5. `web/templates/application-metrics.html`
   - Added certificate metric section
   - Added domain badges display
   - Added status indicators with icons
   - Added tooltips with full details

6. `web/static/css/style.css`
   - Added `.certificate-*` styles
   - Added color schemes for different statuses
   - Added domain badge styling

## 🚀 How to Use

### 1. Run Database Migration

The migration will run automatically on application start.

### 2. Add Certificate Monitoring

```bash
# Get the IngressCertificate metric type ID
CERT_TYPE_ID=$(curl -s http://localhost:8080/api/v1/metric-types | \
  jq -r '.[] | select(.name=="IngressCertificate") | .id')

# Add certificate monitoring to your application
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d "{
    \"application_id\": \"YOUR_APP_ID\",
    \"type_id\": \"$CERT_TYPE_ID\",
    \"configuration\": {
      \"ingress_name\": \"my-app-ingress\",
      \"warning_days\": 30
    }
  }"
```

### 3. View in UI

Navigate to `http://localhost:8080` and you'll see:

```
┌─────────────────────────────────────────┐
│ My Application                           │
│ namespace: production                    │
├─────────────────────────────────────────┤
│ 🔒 certificate                          │
│ ┌─────────────────────────────────┐   │
│ │ my-app.example.com              │   │
│ │ www.my-app.example.com          │   │
│ │ ✓ Valid for 45 days             │   │
│ └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

## 🔧 Configuration Examples

### Minimal Configuration
```json
{
  "ingress_name": "my-ingress"
}
```
- Uses application namespace
- Auto-discovers TLS secret
- 30-day warning threshold

### Full Configuration
```json
{
  "ingress_name": "my-ingress",
  "ingress_namespace": "ingress-nginx",
  "tls_secret_name": "my-tls",
  "warning_days": 15
}
```
- Specific namespace
- Specific TLS secret
- Custom warning threshold

## 🎯 Use Cases

### 1. Production HTTPS Application
Monitor your main application's certificate:
```json
{
  "ingress_name": "webapp-prod",
  "warning_days": 30
}
```

### 2. API with Multiple Domains
Track certificates with multiple SANs:
```json
{
  "ingress_name": "api-ingress",
  "warning_days": 45
}
```

### 3. Shared/Wildcard Certificate
Monitor shared certificates across namespaces:
```json
{
  "ingress_name": "shared-ingress",
  "ingress_namespace": "ingress-nginx",
  "tls_secret_name": "wildcard-cert",
  "warning_days": 60
}
```

## 🔐 Required RBAC Permissions

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: k8s-monitoring-app
rules:
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses"]
  verbs: ["get", "list"]
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get"]
```

## ⚠️ Important Notes

### Security
- The monitoring app reads TLS secrets to parse certificates
- Ensure proper RBAC permissions are configured
- Secrets are only read, never modified

### Limitations
- Only monitors certificates already in Kubernetes secrets
- Does not validate certificate chains
- Does not check OCSP or CRL
- Assumes certificates are PEM-encoded

### Best Practices
1. **Set appropriate warning thresholds**: 30-60 days for production
2. **Monitor all public-facing Ingresses**: Don't miss any certificates
3. **Check regularly**: Default 60-second interval is good
4. **Use with cert-manager**: Track auto-renewal status
5. **Test in staging first**: Verify RBAC and configurations

## 🧪 Testing

### Test with a Real Ingress

1. Create a test Ingress with TLS:
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: test-ingress
  namespace: default
spec:
  tls:
  - hosts:
    - test.example.com
    secretName: test-tls
  rules:
  - host: test.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: test-svc
            port:
              number: 80
```

2. Add certificate monitoring:
```bash
curl -X POST http://localhost:8080/api/v1/application-metrics \
  -H "Content-Type: application/json" \
  -d '{
    "application_id": "YOUR_APP_ID",
    "type_id": "CERT_METRIC_TYPE_ID",
    "configuration": {
      "ingress_name": "test-ingress"
    }
  }'
```

3. Wait 60 seconds for collection
4. Check the UI or API:
```bash
curl http://localhost:8080/api/v1/applications/YOUR_APP_ID/latest-metrics | jq
```

## 📊 Monitoring Recommendations

| Environment | Warning Days | Check Interval | Retention Days |
|-------------|-------------|----------------|----------------|
| Development | 7-15 | 5 minutes | 7 |
| Staging | 15-30 | 1 minute | 30 |
| Production | 30-60 | 1 minute | 90 |
| Critical | 60-90 | 1 minute | 180 |

## 🎉 Summary

This implementation provides:
- ✅ **Automatic certificate monitoring** for all your Ingress resources
- ✅ **Early warning system** to prevent certificate expiration incidents
- ✅ **Beautiful visual indicators** in the Web UI
- ✅ **Complete API access** for automation and alerting
- ✅ **Flexible configuration** for various scenarios
- ✅ **Production-ready** with proper error handling

**The monitoring app will now track your certificates and alert you before they expire!** 🚀

