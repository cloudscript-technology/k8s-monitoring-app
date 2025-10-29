-- Daily alert deduplication table to avoid repeated Slack notifications
CREATE TABLE IF NOT EXISTS alerts_sent_daily (
    id TEXT PRIMARY KEY DEFAULT (
        lower(hex(randomblob(4))) || '-' ||
        lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' ||
        substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' ||
        lower(hex(randomblob(6)))
    ),
    application_metric_id TEXT NOT NULL,
    alert_date DATE NOT NULL,
    alert_reason TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (application_metric_id) REFERENCES application_metrics(id),
    UNIQUE(application_metric_id, alert_date)
);

CREATE INDEX IF NOT EXISTS idx_alerts_sent_daily_metric_date
  ON alerts_sent_daily(application_metric_id, alert_date);

