package application_metric

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"k8s-monitoring-app/internal/core"
)

// Configuration stores metric-specific configuration in JSONB format
type Configuration struct {
	// For HealthCheck metrics
	HealthCheckURL string `json:"health_check_url,omitempty"`
	Method         string `json:"method,omitempty"` // GET, POST, etc.
	ExpectedStatus int    `json:"expected_status,omitempty"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`

	// For PodStatus, PodMemoryUsage, PodCpuUsage, PvcUsage, PodActiveNodes
	PodLabelSelector string `json:"pod_label_selector,omitempty"` // e.g., "app=myapp"
	ContainerName    string `json:"container_name,omitempty"`     // Optional: specific container to monitor

	// For PvcUsage
	PvcName      string `json:"pvc_name,omitempty"`
	PvcMountPath string `json:"pvc_mount_path,omitempty"` // Optional: mount path in the pod (auto-discovered if not provided)

	// For Database and Service Connection metrics (Redis, PostgreSQL, MongoDB, MySQL, Kong)
	ConnectionHost     string `json:"connection_host,omitempty"`     // Host/IP address
	ConnectionPort     int    `json:"connection_port,omitempty"`     // Port number
	ConnectionUsername string `json:"connection_username,omitempty"` // Username for authentication
	ConnectionPassword string `json:"connection_password,omitempty"` // Password for authentication
	ConnectionDatabase string `json:"connection_database,omitempty"` // Database name (for PostgreSQL, MySQL, MongoDB)
	ConnectionSSL      bool   `json:"connection_ssl,omitempty"`      // Use SSL/TLS connection
	ConnectionTimeout  int    `json:"connection_timeout,omitempty"`  // Connection timeout in seconds (default: 5)

	// For MongoDB
	ConnectionAuthSource string `json:"connection_auth_source,omitempty"` // Auth database for MongoDB (default: admin)

	// For Redis
	ConnectionDB int `json:"connection_db,omitempty"` // Redis database number (default: 0)

	// For Kong
	KongAdminURL string `json:"kong_admin_url,omitempty"` // Kong Admin API URL

	// For IngressCertificate
	IngressName      string `json:"ingress_name,omitempty"`      // Name of the Ingress resource
	IngressNamespace string `json:"ingress_namespace,omitempty"` // Namespace (if different from application namespace)
	TLSSecretName    string `json:"tls_secret_name,omitempty"`   // TLS secret name (optional, auto-discovered if not provided)
	WarningDays      int    `json:"warning_days,omitempty"`      // Days before expiration to warn (default: 30)

	// For KafkaConsumerLag
	KafkaBootstrapServers string `json:"kafka_bootstrap_servers,omitempty"` // Kafka bootstrap servers (e.g., "kafka:9092")
	KafkaConsumerGroup    string `json:"kafka_consumer_group,omitempty"`    // Consumer group name
	KafkaTopic            string `json:"kafka_topic,omitempty"`             // Topic name (optional, monitors all topics if not specified)
	KafkaSecurityProtocol string `json:"kafka_security_protocol,omitempty"` // Security protocol: PLAINTEXT, SASL_PLAINTEXT, SASL_SSL, SSL (default: PLAINTEXT)
	KafkaSaslMechanism    string `json:"kafka_sasl_mechanism,omitempty"`    // SASL mechanism: PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
	KafkaSaslUsername     string `json:"kafka_sasl_username,omitempty"`     // SASL username
	KafkaSaslPassword     string `json:"kafka_sasl_password,omitempty"`     // SASL password
	KafkaLagThreshold     int64  `json:"kafka_lag_threshold,omitempty"`     // Lag threshold for warning (default: 1000)
}

// UnmarshalJSON provides lenient parsing for specific fields while keeping the overall schema strict.
// It accepts both numbers and numeric strings for tolerant integer fields
// and both booleans and boolean strings for tolerant boolean fields.
func (c *Configuration) UnmarshalJSON(data []byte) error {
	// Unmarshal into raw map to avoid early type errors on tolerant fields
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	var cfg Configuration

	// Helpers
	parseInt := func(v json.RawMessage) (int, error) {
		var i int
		if err := json.Unmarshal(v, &i); err == nil {
			return i, nil
		}
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			ii, convErr := strconv.Atoi(s)
			if convErr != nil {
				return 0, fmt.Errorf("invalid numeric string: %s", s)
			}
			return ii, nil
		}
		return 0, fmt.Errorf("invalid integer value")
	}
	parseInt64 := func(v json.RawMessage) (int64, error) {
		var i64 int64
		if err := json.Unmarshal(v, &i64); err == nil {
			return i64, nil
		}
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			ii, convErr := strconv.ParseInt(s, 10, 64)
			if convErr != nil {
				return 0, fmt.Errorf("invalid numeric string: %s", s)
			}
			return ii, nil
		}
		return 0, fmt.Errorf("invalid integer value")
	}
	parseBool := func(v json.RawMessage) (bool, error) {
		var b bool
		if err := json.Unmarshal(v, &b); err == nil {
			return b, nil
		}
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			bb, convErr := strconv.ParseBool(s)
			if convErr != nil {
				return false, fmt.Errorf("invalid boolean string: %s", s)
			}
			return bb, nil
		}
		return false, fmt.Errorf("invalid boolean value")
	}

	// Strings
	_ = json.Unmarshal(m["health_check_url"], &cfg.HealthCheckURL)
	_ = json.Unmarshal(m["method"], &cfg.Method)
	_ = json.Unmarshal(m["pod_label_selector"], &cfg.PodLabelSelector)
	_ = json.Unmarshal(m["container_name"], &cfg.ContainerName)
	_ = json.Unmarshal(m["pvc_name"], &cfg.PvcName)
	_ = json.Unmarshal(m["pvc_mount_path"], &cfg.PvcMountPath)
	_ = json.Unmarshal(m["connection_host"], &cfg.ConnectionHost)
	_ = json.Unmarshal(m["connection_username"], &cfg.ConnectionUsername)
	_ = json.Unmarshal(m["connection_password"], &cfg.ConnectionPassword)
	_ = json.Unmarshal(m["connection_database"], &cfg.ConnectionDatabase)
	_ = json.Unmarshal(m["connection_auth_source"], &cfg.ConnectionAuthSource)
	_ = json.Unmarshal(m["kong_admin_url"], &cfg.KongAdminURL)
	_ = json.Unmarshal(m["ingress_name"], &cfg.IngressName)
	_ = json.Unmarshal(m["ingress_namespace"], &cfg.IngressNamespace)
	_ = json.Unmarshal(m["tls_secret_name"], &cfg.TLSSecretName)
	_ = json.Unmarshal(m["kafka_bootstrap_servers"], &cfg.KafkaBootstrapServers)
	_ = json.Unmarshal(m["kafka_consumer_group"], &cfg.KafkaConsumerGroup)
	_ = json.Unmarshal(m["kafka_topic"], &cfg.KafkaTopic)
	_ = json.Unmarshal(m["kafka_security_protocol"], &cfg.KafkaSecurityProtocol)
	_ = json.Unmarshal(m["kafka_sasl_mechanism"], &cfg.KafkaSaslMechanism)
	_ = json.Unmarshal(m["kafka_sasl_username"], &cfg.KafkaSaslUsername)
	_ = json.Unmarshal(m["kafka_sasl_password"], &cfg.KafkaSaslPassword)

	// Ints and bools (tolerant parsing for common misconfigurations)
	if v, ok := m["timeout_seconds"]; ok && len(v) > 0 && string(v) != "null" {
		if i, err := parseInt(v); err == nil {
			cfg.TimeoutSeconds = i
		} else {
			return fmt.Errorf("invalid timeout_seconds: %w", err)
		}
	}
	if v, ok := m["connection_port"]; ok && len(v) > 0 && string(v) != "null" {
		if i, err := parseInt(v); err == nil {
			cfg.ConnectionPort = i
		} else {
			return fmt.Errorf("invalid connection_port: %w", err)
		}
	}
	if v, ok := m["connection_timeout"]; ok && len(v) > 0 && string(v) != "null" {
		if i, err := parseInt(v); err == nil {
			cfg.ConnectionTimeout = i
		} else {
			return fmt.Errorf("invalid connection_timeout: %w", err)
		}
	}
	if v, ok := m["connection_db"]; ok && len(v) > 0 && string(v) != "null" {
		if i, err := parseInt(v); err == nil {
			cfg.ConnectionDB = i
		} else {
			return fmt.Errorf("invalid connection_db: %w", err)
		}
	}
	if v, ok := m["connection_ssl"]; ok && len(v) > 0 && string(v) != "null" {
		if b, err := parseBool(v); err == nil {
			cfg.ConnectionSSL = b
		} else {
			return fmt.Errorf("invalid connection_ssl: %w", err)
		}
	}

	// Special tolerant fields
	if v, ok := m["expected_status"]; ok && len(v) > 0 && string(v) != "null" {
		if i, err := parseInt(v); err == nil {
			cfg.ExpectedStatus = i
		} else {
			return fmt.Errorf("invalid expected_status: %w", err)
		}
	}
	if v, ok := m["warning_days"]; ok && len(v) > 0 && string(v) != "null" {
		if i, err := parseInt(v); err == nil {
			cfg.WarningDays = i
		} else {
			return fmt.Errorf("invalid warning_days: %w", err)
		}
	}
	if v, ok := m["kafka_lag_threshold"]; ok && len(v) > 0 && string(v) != "null" {
		if i64, err := parseInt64(v); err == nil {
			cfg.KafkaLagThreshold = i64
		} else {
			return fmt.Errorf("invalid kafka_lag_threshold: %w", err)
		}
	}

	*c = cfg
	return nil
}

type ApplicationMetric struct {
	ID            string          `json:"id,omitempty"`
	ApplicationID string          `json:"application_id" validate:"required"`
	TypeID        string          `json:"metric_type_id" validate:"required"`
	Configuration json.RawMessage `json:"configuration" validate:"required"`
	CreatedAt     time.Time       `json:"created_at,omitempty"`
	UpdatedAt     time.Time       `json:"updated_at,omitempty"`
}

type Service interface {
	Get(sc *core.HTTPServerContext) error
	Add(sc *core.HTTPServerContext) error
	List(sc *core.HTTPServerContext) error
	ListByApplication(sc *core.HTTPServerContext) error
	Update(sc *core.HTTPServerContext) error
	Delete(sc *core.HTTPServerContext) error
}
