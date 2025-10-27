package project

import (
	"k8s-monitoring-app/internal/core"
)

type Project struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
}

type Service interface {
	Get(sc *core.HTTPServerContext) error
	Add(sc *core.HTTPServerContext) error
	List(sc *core.HTTPServerContext) error
	Update(sc *core.HTTPServerContext) error
	Delete(sc *core.HTTPServerContext) error
}
