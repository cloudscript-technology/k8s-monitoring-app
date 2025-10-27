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
	PvcName string `json:"pvc_name,omitempty"`
}

type ApplicationMetric struct {
	ID            string          `json:"id,omitempty"`
	ApplicationID string          `json:"application_id" validate:"required"`
	TypeID        string          `json:"type_id" validate:"required"`
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
