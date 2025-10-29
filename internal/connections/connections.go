package connections

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	applicationMetricModel "k8s-monitoring-app/pkg/application_metric/model"
	applicationMetricValueModel "k8s-monitoring-app/pkg/application_metric_value/model"
)

const (
	StatusConnected = "connected"
	StatusFailed    = "failed"
	StatusTimeout   = "timeout"
)

// TestRedisConnection tests Redis connection with authentication
func TestRedisConnection(ctx context.Context, config *applicationMetricModel.Configuration) applicationMetricValueModel.MetricValue {
	start := time.Now()
	result := applicationMetricValueModel.MetricValue{}

	// Set default timeout
	timeout := 5
	if config.ConnectionTimeout > 0 {
		timeout = config.ConnectionTimeout
	}

	// Create context with timeout
	connCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Configure Redis client
	opts := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.ConnectionHost, config.ConnectionPort),
		Password:     config.ConnectionPassword,
		DB:           config.ConnectionDB,
		DialTimeout:  time.Duration(timeout) * time.Second,
		ReadTimeout:  time.Duration(timeout) * time.Second,
		WriteTimeout: time.Duration(timeout) * time.Second,
	}

	if config.ConnectionSSL {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}

	client := redis.NewClient(opts)
	defer client.Close()

	// Test connection
	pingStart := time.Now()
	pong, err := client.Ping(connCtx).Result()
	pingDuration := time.Since(pingStart)

	connectionDuration := time.Since(start)
	result.ConnectionTimeMs = connectionDuration.Milliseconds()

	if err != nil {
		// Classificar corretamente erro de timeout (contexto ou i/o timeout)
		var netErr net.Error
		if connCtx.Err() == context.DeadlineExceeded || errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			result.ConnectionStatus = StatusTimeout
			result.ConnectionError = "connection timeout"
		} else {
			result.ConnectionStatus = StatusFailed
			result.ConnectionError = err.Error()
		}
		return result
	}

	result.ConnectionStatus = StatusConnected
	result.ConnectionPingTimeMs = pingDuration.Milliseconds()
	result.ConnectionInfo = pong

	// Get Redis server info
	info, err := client.Info(connCtx, "server").Result()
	if err == nil {
		// Parse version from info (simplified - you might want to parse it better)
		result.ConnectionVersion = parseRedisVersion(info)
	}

	return result
}

// TestPostgreSQLConnection tests PostgreSQL connection with authentication
func TestPostgreSQLConnection(ctx context.Context, config *applicationMetricModel.Configuration) applicationMetricValueModel.MetricValue {
	start := time.Now()
	result := applicationMetricValueModel.MetricValue{}

	// Set default timeout
	timeout := 5
	if config.ConnectionTimeout > 0 {
		timeout = config.ConnectionTimeout
	}

	// Build connection string
	sslMode := "disable"
	if config.ConnectionSSL {
		sslMode = "require"
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		config.ConnectionHost,
		config.ConnectionPort,
		config.ConnectionUsername,
		config.ConnectionPassword,
		config.ConnectionDatabase,
		sslMode,
		timeout,
	)

	// Open connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		result.ConnectionStatus = StatusFailed
		result.ConnectionError = err.Error()
		result.ConnectionTimeMs = time.Since(start).Milliseconds()
		return result
	}
	defer db.Close()

	// Create context with timeout
	connCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Test connection
	pingStart := time.Now()
	err = db.PingContext(connCtx)
	pingDuration := time.Since(pingStart)

	connectionDuration := time.Since(start)
	result.ConnectionTimeMs = connectionDuration.Milliseconds()

	if err != nil {
		var netErr net.Error
		if connCtx.Err() == context.DeadlineExceeded || errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			result.ConnectionStatus = StatusTimeout
			result.ConnectionError = "connection timeout"
		} else {
			result.ConnectionStatus = StatusFailed
			result.ConnectionError = err.Error()
		}
		return result
	}

	result.ConnectionStatus = StatusConnected
	result.ConnectionPingTimeMs = pingDuration.Milliseconds()

	// Get PostgreSQL version
	var version string
	err = db.QueryRowContext(connCtx, "SELECT version()").Scan(&version)
	if err == nil {
		result.ConnectionVersion = version
	}

	// Get database stats
	var dbSize int64
	err = db.QueryRowContext(connCtx, "SELECT pg_database_size($1)", config.ConnectionDatabase).Scan(&dbSize)
	if err == nil {
		result.ConnectionInfo = fmt.Sprintf("Database size: %d bytes", dbSize)
	}

	return result
}

// TestMySQLConnection tests MySQL connection with authentication
func TestMySQLConnection(ctx context.Context, config *applicationMetricModel.Configuration) applicationMetricValueModel.MetricValue {
	start := time.Now()
	result := applicationMetricValueModel.MetricValue{}

	// Set default timeout
	timeout := 5
	if config.ConnectionTimeout > 0 {
		timeout = config.ConnectionTimeout
	}

	// Build connection string
	// Format: username:password@tcp(host:port)/dbname?timeout=5s
	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?timeout=%ds",
		config.ConnectionUsername,
		config.ConnectionPassword,
		config.ConnectionHost,
		config.ConnectionPort,
		config.ConnectionDatabase,
		timeout,
	)

	if config.ConnectionSSL {
		connStr += "&tls=true"
	}

	// Open connection
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		result.ConnectionStatus = StatusFailed
		result.ConnectionError = err.Error()
		result.ConnectionTimeMs = time.Since(start).Milliseconds()
		return result
	}
	defer db.Close()

	// Create context with timeout
	connCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Test connection
	pingStart := time.Now()
	err = db.PingContext(connCtx)
	pingDuration := time.Since(pingStart)

	connectionDuration := time.Since(start)
	result.ConnectionTimeMs = connectionDuration.Milliseconds()

	if err != nil {
		var netErr net.Error
		if connCtx.Err() == context.DeadlineExceeded || errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			result.ConnectionStatus = StatusTimeout
			result.ConnectionError = "connection timeout"
		} else {
			result.ConnectionStatus = StatusFailed
			result.ConnectionError = err.Error()
		}
		return result
	}

	result.ConnectionStatus = StatusConnected
	result.ConnectionPingTimeMs = pingDuration.Milliseconds()

	// Get MySQL version
	var version string
	err = db.QueryRowContext(connCtx, "SELECT VERSION()").Scan(&version)
	if err == nil {
		result.ConnectionVersion = version
	}

	// Get database stats
	var dbSize sql.NullFloat64
	query := `
		SELECT ROUND(SUM(data_length + index_length), 0) 
		FROM information_schema.tables 
		WHERE table_schema = ?
		GROUP BY table_schema
	`
	err = db.QueryRowContext(connCtx, query, config.ConnectionDatabase).Scan(&dbSize)
	if err == nil && dbSize.Valid {
		result.ConnectionInfo = fmt.Sprintf("Database size: %.0f bytes", dbSize.Float64)
	}

	return result
}

// TestMongoDBConnection tests MongoDB connection with authentication
func TestMongoDBConnection(ctx context.Context, config *applicationMetricModel.Configuration) applicationMetricValueModel.MetricValue {
	start := time.Now()
	result := applicationMetricValueModel.MetricValue{}

	// Set default timeout
	timeout := 5
	if config.ConnectionTimeout > 0 {
		timeout = config.ConnectionTimeout
	}

	// Set default auth source
	authSource := "admin"
	if config.ConnectionAuthSource != "" {
		authSource = config.ConnectionAuthSource
	}

	// Create context with timeout
	connCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Build connection URI
	protocol := "mongodb"
	if config.ConnectionSSL {
		protocol = "mongodb+srv"
	}

	uri := fmt.Sprintf(
		"%s://%s:%s@%s:%d/%s?authSource=%s",
		protocol,
		config.ConnectionUsername,
		config.ConnectionPassword,
		config.ConnectionHost,
		config.ConnectionPort,
		config.ConnectionDatabase,
		authSource,
	)

	// Create client options
	clientOpts := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(time.Duration(timeout) * time.Second).
		SetServerSelectionTimeout(time.Duration(timeout) * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(connCtx, clientOpts)
	if err != nil {
		result.ConnectionStatus = StatusFailed
		result.ConnectionError = err.Error()
		result.ConnectionTimeMs = time.Since(start).Milliseconds()
		return result
	}
	defer func() {
		if disconnectErr := client.Disconnect(context.Background()); disconnectErr != nil {
			log.Error().Err(disconnectErr).Msg("Error disconnecting MongoDB client")
			// Log error but don't fail
		}
	}()

	// Test connection with ping
	pingStart := time.Now()
	err = client.Ping(connCtx, readpref.Primary())
	pingDuration := time.Since(pingStart)

	connectionDuration := time.Since(start)
	result.ConnectionTimeMs = connectionDuration.Milliseconds()

	if err != nil {
		var netErr net.Error
		if connCtx.Err() == context.DeadlineExceeded || errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			result.ConnectionStatus = StatusTimeout
			result.ConnectionError = "connection timeout"
		} else {
			result.ConnectionStatus = StatusFailed
			result.ConnectionError = err.Error()
		}
		return result
	}

	result.ConnectionStatus = StatusConnected
	result.ConnectionPingTimeMs = pingDuration.Milliseconds()

	// Get MongoDB server info
	var serverStatus map[string]interface{}
	db := client.Database(config.ConnectionDatabase)
	err = db.RunCommand(connCtx, map[string]interface{}{"serverStatus": 1}).Decode(&serverStatus)
	if err == nil {
		if version, ok := serverStatus["version"].(string); ok {
			result.ConnectionVersion = version
		}
		if host, ok := serverStatus["host"].(string); ok {
			result.ConnectionInfo = fmt.Sprintf("Connected to: %s", host)
		}
	}

	return result
}

// TestKongConnection tests Kong API Gateway connection and health
func TestKongConnection(ctx context.Context, config *applicationMetricModel.Configuration) applicationMetricValueModel.MetricValue {
	start := time.Now()
	result := applicationMetricValueModel.MetricValue{}

	// Set default timeout
	timeout := 5
	if config.ConnectionTimeout > 0 {
		timeout = config.ConnectionTimeout
	}

	// Create context with timeout
	connCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Determine the URL to test
	url := config.KongAdminURL
	if url == "" {
		// Use standard connection if admin URL not provided
		protocol := "http"
		if config.ConnectionSSL {
			protocol = "https"
		}
		url = fmt.Sprintf("%s://%s:%d", protocol, config.ConnectionHost, config.ConnectionPort)
	}

	// Add /status endpoint for health check
	statusURL := fmt.Sprintf("%s/status", url)

	// Create HTTP client
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: config.ConnectionSSL, // Only for testing purposes
			},
		},
	}

	// Create request
	req, err := http.NewRequestWithContext(connCtx, "GET", statusURL, nil)
	if err != nil {
		result.ConnectionStatus = StatusFailed
		result.ConnectionError = err.Error()
		result.ConnectionTimeMs = time.Since(start).Milliseconds()
		return result
	}

	// Add authentication if provided
	if config.ConnectionUsername != "" && config.ConnectionPassword != "" {
		req.SetBasicAuth(config.ConnectionUsername, config.ConnectionPassword)
	}

	// Send request
	pingStart := time.Now()
	resp, err := client.Do(req)
	pingDuration := time.Since(pingStart)

	connectionDuration := time.Since(start)
	result.ConnectionTimeMs = connectionDuration.Milliseconds()

	if err != nil {
		var netErr net.Error
		if connCtx.Err() == context.DeadlineExceeded || errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
			result.ConnectionStatus = StatusTimeout
			result.ConnectionError = "connection timeout"
		} else {
			result.ConnectionStatus = StatusFailed
			result.ConnectionError = err.Error()
		}
		return result
	}
	defer resp.Body.Close()

	result.ConnectionPingTimeMs = pingDuration.Milliseconds()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.ConnectionStatus = StatusConnected
		result.ConnectionInfo = fmt.Sprintf("HTTP Status: %d", resp.StatusCode)

		// Try to get Kong version from header
		if version := resp.Header.Get("Server"); version != "" {
			result.ConnectionVersion = version
		}
	} else {
		result.ConnectionStatus = StatusFailed
		result.ConnectionError = fmt.Sprintf("HTTP Status: %d", resp.StatusCode)
	}

	return result
}

// parseRedisVersion extracts version from Redis INFO output
func parseRedisVersion(info string) string {
	// Simple parsing - look for redis_version line
	// Format: redis_version:6.2.6
	lines := []byte(info)
	for i := 0; i < len(lines); i++ {
		if i+14 < len(lines) && string(lines[i:i+14]) == "redis_version:" {
			// Find end of line
			start := i + 14
			end := start
			for end < len(lines) && lines[end] != '\n' && lines[end] != '\r' {
				end++
			}
			return string(lines[start:end])
		}
	}
	return "unknown"
}
