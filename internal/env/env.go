package env

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

var (
    ENV                         string
    PORT                        int
    DB_PATH                     string
    METRICS_RETENTION_DAYS      int
    METRICS_CLEANUP_INTERVAL    string
    METRICS_COLLECTION_INTERVAL int // Collection interval in seconds (default: 60)

	// Slack Alerts Configuration
	SLACK_WEBHOOK_URL          string
	SLACK_ALERTS_ENABLED       bool
	SLACK_ALERTS_DEDUP_MINUTES int // Deduplication window in minutes (default: 10)

    // OAuth Configuration
    GOOGLE_CLIENT_ID       string
    GOOGLE_CLIENT_SECRET   string
    GOOGLE_REDIRECT_URL    string
    ALLOWED_GOOGLE_DOMAINS string // Comma-separated list of allowed domains
    ALLOWED_GOOGLE_EMAILS  string // Comma-separated list of allowed email addresses (optional)
)

func GetEnv() error {
	if os.Getenv("ENV") != "staging" && os.Getenv("ENV") != "production" {
		// Try to load .env from multiple locations
		possiblePaths := []string{
			".env",                       // Current directory
			"cmd/.env",                   // cmd subdirectory (when running from root)
			"../.env",                    // Parent directory (when running from cmd)
			filepath.Join("cmd", ".env"), // Explicit cmd path
		}

		var lastErr error
		loaded := false

		for _, path := range possiblePaths {
			err := godotenv.Load(path)
			if err == nil {
				loaded = true
				break
			}
			lastErr = err
		}

		// If .env not found in any location, it's not critical
		// Environment variables can be set externally
		if !loaded && lastErr != nil {
			// Log but don't fail - env vars might be set externally
			// return lastErr
		}
	}

	ENV = os.Getenv("ENV")
	PORT, _ = strconv.Atoi(os.Getenv("PORT"))
	if PORT == 0 {
		PORT = 8080
	}

	DB_PATH = os.Getenv("DB_PATH")
	if DB_PATH == "" {
		DB_PATH = "./data/k8s_monitoring.db"
	}

	// Slack Alerts Configuration
	SLACK_WEBHOOK_URL = os.Getenv("SLACK_WEBHOOK_URL")
	if v := os.Getenv("SLACK_ALERTS_ENABLED"); v != "" {
		// Accept true/false in various casings
		SLACK_ALERTS_ENABLED = v == "1" || v == "true" || v == "TRUE" || v == "True"
	} else {
		SLACK_ALERTS_ENABLED = false
	}

	// Slack Alerts Deduplication Window
	if v := os.Getenv("SLACK_ALERTS_DEDUP_MINUTES"); v != "" {
		if minutes, err := strconv.Atoi(v); err == nil {
			if minutes < 1 {
				SLACK_ALERTS_DEDUP_MINUTES = 1
			} else {
				SLACK_ALERTS_DEDUP_MINUTES = minutes
			}
		} else {
			SLACK_ALERTS_DEDUP_MINUTES = 10
		}
	} else {
		SLACK_ALERTS_DEDUP_MINUTES = 10
	}

    // OAuth Configuration
    GOOGLE_CLIENT_ID = os.Getenv("GOOGLE_CLIENT_ID")
    GOOGLE_CLIENT_SECRET = os.Getenv("GOOGLE_CLIENT_SECRET")
    GOOGLE_REDIRECT_URL = os.Getenv("GOOGLE_REDIRECT_URL")
    ALLOWED_GOOGLE_DOMAINS = os.Getenv("ALLOWED_GOOGLE_DOMAINS")
    ALLOWED_GOOGLE_EMAILS = os.Getenv("ALLOWED_GOOGLE_EMAILS")

	// Metrics retention configuration (default: 30 days)
	retentionDays := os.Getenv("METRICS_RETENTION_DAYS")
	if retentionDays == "" {
		METRICS_RETENTION_DAYS = 30
	} else {
		if days, err := strconv.Atoi(retentionDays); err == nil {
			METRICS_RETENTION_DAYS = days
		} else {
			METRICS_RETENTION_DAYS = 30
		}
	}

	// Metrics cleanup interval (default: daily at 2 AM)
	METRICS_CLEANUP_INTERVAL = os.Getenv("METRICS_CLEANUP_INTERVAL")
	if METRICS_CLEANUP_INTERVAL == "" {
		METRICS_CLEANUP_INTERVAL = "0 2 * * *" // Cron format: daily at 2 AM
	}

	// Metrics collection interval in seconds (default: 60)
	collectionInterval := os.Getenv("METRICS_COLLECTION_INTERVAL")
	if collectionInterval == "" {
		METRICS_COLLECTION_INTERVAL = 60
	} else {
		if seconds, err := strconv.Atoi(collectionInterval); err == nil {
			if seconds < 10 {
				// Minimum 10 seconds to avoid overloading
				METRICS_COLLECTION_INTERVAL = 10
			} else {
				METRICS_COLLECTION_INTERVAL = seconds
			}
		} else {
			METRICS_COLLECTION_INTERVAL = 60
		}
	}

	return nil
}
