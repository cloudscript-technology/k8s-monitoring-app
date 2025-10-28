package application_metric

import (
	"encoding/json"
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
