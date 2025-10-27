package application

import (
	"time"

	"k8s-monitoring-app/internal/core"
)

type Application struct {
	ID          string    `json:"id,omitempty"`
	ProjectID   string    `json:"project_id" validate:"required"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description" validate:"required"`
	Namespace   string    `json:"namespace" validate:"required"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

type Service interface {
	Get(sc *core.HTTPServerContext) error
	Add(sc *core.HTTPServerContext) error
	List(sc *core.HTTPServerContext) error
	ListByProject(sc *core.HTTPServerContext) error
	Update(sc *core.HTTPServerContext) error
	Delete(sc *core.HTTPServerContext) error
}
