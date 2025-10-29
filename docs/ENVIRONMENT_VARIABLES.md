# Environment Variables

## Database Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DB_PATH` | SQLite database file path | `./data/k8s_monitoring.db` | No |

## Application Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `ENV` | Environment name (development, staging, production) | `development` | No |
| `ADMIN_TOKEN` | Admin authentication token | - | No |

## Slack Alerts

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SLACK_ALERTS_ENABLED` | Enable Slack notifications for metric failures | `false` | No |
| `SLACK_WEBHOOK_URL` | Slack Incoming Webhook URL | - | No |
| `SLACK_ALERTS_DEDUP_MINUTES` | Suppress repeated alerts within this window (minutes) | `10` | No |

When `SLACK_ALERTS_ENABLED` is `true` and `SLACK_WEBHOOK_URL` is set, the monitoring service will send a Slack message when it detects failures in metrics like `HealthCheck` (status down), `PodStatus` (not ready/degraded/pending/not found), `IngressCertificate` (expired/error/expiring soon), `KafkaConsumerLag` (critical/error), and connection metrics (status failed/timeout).

Example:
```bash
export SLACK_ALERTS_ENABLED=true
export SLACK_WEBHOOK_URL=https://hooks.slack.com/services/XXX/YYY/ZZZ
export SLACK_ALERTS_DEDUP_MINUTES=10  # Avoid repeats within 10 minutes
```

## Metrics Retention Configuration

### METRICS_RETENTION_DAYS

**Description**: Number of days to keep metrics history in the database. Older metrics will be automatically deleted.

**Default**: `30` (30 days)

**Examples**:
- `1` - Keep only last 24 hours
- `7` - Keep last week
- `30` - Keep last month (default)
- `90` - Keep last 3 months
- `365` - Keep last year

**Usage**:
```bash
export METRICS_RETENTION_DAYS=7  # Keep only 7 days of history
```

### METRICS_CLEANUP_INTERVAL

**Description**: Cron schedule for running the automatic cleanup job. Uses standard cron syntax.

**Default**: `"0 2 * * *"` (Daily at 2 AM)

**Cron Format**:
```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6) (Sunday to Saturday)
│ │ │ │ │
│ │ │ │ │
* * * * *
```

**Examples**:

| Schedule | Description |
|----------|-------------|
| `"0 2 * * *"` | Daily at 2:00 AM (default) |
| `"0 */6 * * *"` | Every 6 hours |
| `"0 0 * * 0"` | Weekly on Sunday at midnight |
| `"30 3 * * *"` | Daily at 3:30 AM |
| `"0 0 */7 * *"` | Every 7 days at midnight |
| `"@daily"` | Once per day at midnight |
| `"@weekly"` | Once per week at midnight on Sunday |
| `"@monthly"` | Once per month at midnight on first day |

**Usage**:
```bash
export METRICS_CLEANUP_INTERVAL="0 */6 * * *"  # Run every 6 hours
```

## Kubernetes Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `KUBECONFIG` | Path to kubeconfig file | `~/.kube/config` | No |

**Note**: If not set, the application will try these locations in order:
1. `KUBECONFIG` environment variable
2. `~/.kube/config` file
3. In-cluster configuration (when running inside Kubernetes)

## Complete Example

### For Development (Keep 1 day, cleanup every 6 hours)

```bash
export ENV=development
export DB_PATH=./data/k8s_monitoring.db
export METRICS_RETENTION_DAYS=1
export METRICS_CLEANUP_INTERVAL="0 */6 * * *"
export KUBECONFIG=~/.kube/config
```

### For Production (Keep 90 days, cleanup daily)

```bash
export ENV=production
export DB_PATH=/var/lib/k8s-monitoring-app/k8s_monitoring.db
export METRICS_RETENTION_DAYS=90
export METRICS_CLEANUP_INTERVAL="0 2 * * *"
```

### For Testing (Keep 7 days, cleanup weekly)

```bash
export ENV=staging
export DB_PATH=./data/k8s_monitoring_staging.db
export METRICS_RETENTION_DAYS=7
export METRICS_CLEANUP_INTERVAL="0 0 * * 0"  # Weekly on Sunday
```

## How Cleanup Works

### Automatic Cleanup Process

1. **Scheduled Execution**: The cleanup job runs automatically based on `METRICS_CLEANUP_INTERVAL`
2. **Retention Check**: Calculates cutoff date: `NOW - METRICS_RETENTION_DAYS`
3. **Deletion**: Removes all `application_metric_values` records older than cutoff date
4. **Logging**: Records how many records were deleted

### Example Logs

```json
{
  "level": "info",
  "retention_days": 30,
  "cutoff_date": "2024-10-01T00:00:00Z",
  "message": "Starting metrics cleanup"
}

{
  "level": "info",
  "deleted_records": 15420,
  "retention_days": 30,
  "message": "Metrics cleanup completed"
}
```

### Manual Cleanup

If you need to manually trigger cleanup, restart the application at the scheduled time or modify the cron schedule to run immediately:

```bash
# Set to run every minute (for testing)
export METRICS_CLEANUP_INTERVAL="* * * * *"
```

## Best Practices

### Storage Considerations

Estimate your storage needs:

```
Records per minute = Number of applications × Number of metrics per app
Records per day = Records per minute × 1440 (minutes in a day)
Records for retention period = Records per day × METRICS_RETENTION_DAYS
```

**Example**:
- 10 applications
- 5 metrics per application
- 30 days retention

```
10 apps × 5 metrics = 50 records/minute
50 × 1440 = 72,000 records/day
72,000 × 30 = 2,160,000 records total
```

### Recommended Settings

| Use Case | Retention Days | Cleanup Interval | Reasoning |
|----------|----------------|------------------|-----------|
| Development | 1-7 | Every 6 hours | Fast iteration, limited storage |
| Testing/Staging | 7-14 | Daily | Enough for debugging |
| Production Monitoring | 30-90 | Daily at 2 AM | Balance between history and storage |
| Compliance/Audit | 365+ | Weekly | Long-term data retention |

### Performance Tips

1. **Index Optimization**: The cleanup uses `created_at` field - ensure it's indexed
2. **Off-Peak Hours**: Schedule cleanup during low-traffic hours (e.g., 2-4 AM)
3. **Storage Monitoring**: Monitor database size and adjust retention as needed
4. **Gradual Increase**: Start with shorter retention and increase if needed

## Troubleshooting

### Cleanup not running

**Check logs**:
```bash
grep "cleanup" /path/to/logs | grep -E "(Starting|completed)"
```

**Verify cron schedule**:
```bash
# The application logs will show the schedule on startup
grep "Monitoring service started" /path/to/logs
```

### Too much data being deleted

**Increase retention**:
```bash
export METRICS_RETENTION_DAYS=90  # Instead of 30
```

### Not enough cleanup

**Decrease retention or increase frequency**:
```bash
export METRICS_RETENTION_DAYS=7
export METRICS_CLEANUP_INTERVAL="0 */12 * * *"  # Twice daily
```

## See Also

- [Local Development Guide](LOCAL_DEVELOPMENT.md)
- [Deployment Guide](DEPLOYMENT.md)
- [API Documentation](API.md)
