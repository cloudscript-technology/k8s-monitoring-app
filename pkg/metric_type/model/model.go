package metric_type

import (
	"time"

	"k8s-monitoring-app/internal/core"
)

type MetricType struct {
	ID          string    `json:"id,omitempty"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description" validate:"required"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

type Service interface {
	Get(sc *core.HTTPServerContext) error
	List(sc *core.HTTPServerContext) error
}
