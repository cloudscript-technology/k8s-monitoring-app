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
	PodPhase     string    `json:"pod_phase,omitempty"` // Running, Pending, Failed, etc.
	PodReady     bool      `json:"pod_ready,omitempty"`
	RestartCount int32     `json:"restart_count,omitempty"`
	Pods         []PodInfo `json:"pods,omitempty"` // Individual pod information
	TotalPods    int       `json:"total_pods,omitempty"`
	ReadyPods    int       `json:"ready_pods,omitempty"`

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
	ActiveNodesCount int        `json:"active_nodes_count,omitempty"`
	NodeNames        []string   `json:"node_names,omitempty"`
	Nodes            []NodeInfo `json:"nodes,omitempty"` // Detailed node information
}

// NodeInfo contains detailed information about a node
type NodeInfo struct {
	Name       string            `json:"name"`
	Ready      bool              `json:"ready"`
	Status     string            `json:"status"` // "Ready", "NotReady", "Unknown"
	Conditions []NodeCondition   `json:"conditions,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	PodCount   int               `json:"pod_count,omitempty"` // Number of application pods on this node
}

// NodeCondition represents a node condition
type NodeCondition struct {
	Type    string `json:"type"`   // Ready, MemoryPressure, DiskPressure, PIDPressure, NetworkUnavailable
	Status  string `json:"status"` // True, False, Unknown
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

// PodInfo contains detailed information about a single pod
type PodInfo struct {
	Name         string `json:"name"`
	Phase        string `json:"phase"`         // Running, Pending, Failed, Succeeded, Unknown
	Ready        bool   `json:"ready"`         // Is the pod ready
	RestartCount int32  `json:"restart_count"` // Total restarts
	NodeName     string `json:"node_name,omitempty"`
	IP           string `json:"ip,omitempty"`
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
