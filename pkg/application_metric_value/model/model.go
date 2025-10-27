package application_metric_value

import (
	"encoding/json"
	"time"

	"k8s-monitoring-app/internal/core"
)

// MetricValue stores the actual metric data in JSONB format
type MetricValue struct {
	// For HealthCheck
	Status         string `json:"status,omitempty"` // "up" or "down"
	ResponseTimeMs int64  `json:"response_time_ms,omitempty"`
	StatusCode     int    `json:"status_code,omitempty"`
	ErrorMessage   string `json:"error_message,omitempty"`

	// For PodStatus
	PodPhase     string `json:"pod_phase,omitempty"` // Running, Pending, Failed, etc.
	PodReady     bool   `json:"pod_ready,omitempty"`
	RestartCount int32  `json:"restart_count,omitempty"`

	// For PodMemoryUsage
	MemoryUsageBytes int64   `json:"memory_usage_bytes,omitempty"`
	MemoryLimitBytes int64   `json:"memory_limit_bytes,omitempty"`
	MemoryPercent    float64 `json:"memory_percent,omitempty"`

	// For PodCpuUsage
	CpuUsageMillicores int64   `json:"cpu_usage_millicores,omitempty"`
	CpuLimitMillicores int64   `json:"cpu_limit_millicores,omitempty"`
	CpuPercent         float64 `json:"cpu_percent,omitempty"`

	// For PvcUsage
	PvcCapacityBytes int64   `json:"pvc_capacity_bytes,omitempty"`
	PvcUsedBytes     int64   `json:"pvc_used_bytes,omitempty"`
	PvcPercent       float64 `json:"pvc_percent,omitempty"`

	// For PodActiveNodes
	ActiveNodesCount int      `json:"active_nodes_count,omitempty"`
	NodeNames        []string `json:"node_names,omitempty"`
}

type ApplicationMetricValue struct {
	ID                  string          `json:"id,omitempty"`
	ApplicationMetricID string          `json:"application_metric_id"`
	Value               json.RawMessage `json:"value"`
	CreatedAt           time.Time       `json:"created_at,omitempty"`
	UpdatedAt           time.Time       `json:"updated_at,omitempty"`
}

type Service interface {
	Get(sc *core.HTTPServerContext) error
	ListByApplicationMetric(sc *core.HTTPServerContext) error
	GetLatestByApplication(sc *core.HTTPServerContext) error
}
