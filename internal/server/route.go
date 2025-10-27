package server

import (
	"context"

	"k8s-monitoring-app/internal/core"
	model "k8s-monitoring-app/internal/server/model"

	"gitlab.cloudscript.com.br/general/go-instrumentation.git/log"
)

func bindRoutes(s *core.HTTPServer) {
	log.Info(context.Background()).Msg("Binding routes")

	s.Api.GET("/health", s.WrapHandler(s.Health))

	apiV1 := s.Api.Group("/api/v1")

	// Project routes
	apiV1.GET("/projects", s.WrapHandler(model.ServerSvc.Project.List))
	apiV1.GET("/projects/:id", s.WrapHandler(model.ServerSvc.Project.Get))
	apiV1.POST("/projects", s.WrapHandler(model.ServerSvc.Project.Add))
	apiV1.PUT("/projects/:id", s.WrapHandler(model.ServerSvc.Project.Update))
	apiV1.DELETE("/projects/:id", s.WrapHandler(model.ServerSvc.Project.Delete))

	// Application routes
	apiV1.GET("/applications", s.WrapHandler(model.ServerSvc.Application.List))
	apiV1.GET("/applications/:id", s.WrapHandler(model.ServerSvc.Application.Get))
	apiV1.GET("/projects/:project_id/applications", s.WrapHandler(model.ServerSvc.Application.ListByProject))
	apiV1.POST("/applications", s.WrapHandler(model.ServerSvc.Application.Add))
	apiV1.PUT("/applications/:id", s.WrapHandler(model.ServerSvc.Application.Update))
	apiV1.DELETE("/applications/:id", s.WrapHandler(model.ServerSvc.Application.Delete))

	// Metric Type routes
	apiV1.GET("/metric-types", s.WrapHandler(model.ServerSvc.MetricType.List))
	apiV1.GET("/metric-types/:id", s.WrapHandler(model.ServerSvc.MetricType.Get))

	// Application Metric routes
	apiV1.GET("/application-metrics", s.WrapHandler(model.ServerSvc.ApplicationMetric.List))
	apiV1.GET("/application-metrics/:id", s.WrapHandler(model.ServerSvc.ApplicationMetric.Get))
	apiV1.GET("/applications/:application_id/metrics", s.WrapHandler(model.ServerSvc.ApplicationMetric.ListByApplication))
	apiV1.POST("/application-metrics", s.WrapHandler(model.ServerSvc.ApplicationMetric.Add))
	apiV1.PUT("/application-metrics/:id", s.WrapHandler(model.ServerSvc.ApplicationMetric.Update))
	apiV1.DELETE("/application-metrics/:id", s.WrapHandler(model.ServerSvc.ApplicationMetric.Delete))

	// Application Metric Value routes (read-only - values are collected by cron)
	apiV1.GET("/metric-values/:id", s.WrapHandler(model.ServerSvc.ApplicationMetricValue.Get))
	apiV1.GET("/application-metrics/:application_metric_id/values", s.WrapHandler(model.ServerSvc.ApplicationMetricValue.ListByApplicationMetric))
	apiV1.GET("/applications/:application_id/latest-metrics", s.WrapHandler(model.ServerSvc.ApplicationMetricValue.GetLatestByApplication))
}
