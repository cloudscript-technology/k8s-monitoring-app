package server

import (
	"context"

	"k8s-monitoring-app/internal/auth"
	"k8s-monitoring-app/internal/core"
	model "k8s-monitoring-app/internal/server/model"
	"k8s-monitoring-app/internal/web"

	"gitlab.cloudscript.com.br/general/go-instrumentation.git/log"
)

func bindRoutes(s *core.HTTPServer) {
	log.Info(context.Background()).Msg("Binding routes")

	// Web UI Handler
	webHandler, err := web.NewHandler()
	if err != nil {
		log.Error(context.Background(), err).Msg("Failed to initialize web handler")
	} else {
		log.Info(context.Background()).Msg("Web handler initialized successfully")
	}

	// Health check (no auth required)
	s.Api.GET("/health", s.WrapHandler(s.Health))

	// Auth routes (no auth required)
	authGroup := s.Api.Group("/auth")
	if webHandler != nil {
		authGroup.GET("/login", s.WrapHandler(webHandler.RenderLogin))
		authGroup.GET("/error", s.WrapHandler(webHandler.RenderAuthError))
	}
	authGroup.GET("/google", auth.HandleLogin)
	authGroup.GET("/callback", auth.HandleCallback)
	authGroup.GET("/logout", auth.HandleLogout)

	// Web UI routes
	if webHandler != nil {
		log.Info(context.Background()).Msg("Binding web UI routes")
		s.Api.GET("/", s.WrapHandler(webHandler.Dashboard))
		s.Api.Static("/static", "web/static")

		// Registration pages
		s.Api.GET("/cadastros/projetos", s.WrapHandler(webHandler.RenderCadastroProjects))
		s.Api.GET("/cadastros/aplicacoes", s.WrapHandler(webHandler.RenderCadastroApplications))
		s.Api.GET("/cadastros/metricas", s.WrapHandler(webHandler.RenderCadastroMetrics))

		// HTMX partial endpoints
		apiUI := s.Api.Group("/api/ui")
		apiUI.GET("/projects", s.WrapHandler(webHandler.GetProjects))
		apiUI.GET("/applications/:id/metrics", s.WrapHandler(webHandler.GetApplicationMetrics))
		apiUI.GET("/projects-list", s.WrapHandler(webHandler.GetProjectsList))
		apiUI.GET("/applications-list", s.WrapHandler(webHandler.GetApplicationsList))
		apiUI.GET("/metrics-list", s.WrapHandler(webHandler.GetMetricsList))
		apiUI.GET("/projects-options", s.WrapHandler(webHandler.GetProjectsOptions))
		apiUI.GET("/applications-options", s.WrapHandler(webHandler.GetApplicationsOptions))
		apiUI.GET("/metric-types-options", s.WrapHandler(webHandler.GetMetricTypesOptions))
		apiUI.GET("/metric-configuration-fields/:id", s.WrapHandler(webHandler.GetMetricConfigurationFields))
		
		// DELETE endpoints for UI
		apiUI.DELETE("/metrics/:id", s.WrapHandler(webHandler.DeleteMetric))
		apiUI.DELETE("/applications/:id", s.WrapHandler(webHandler.DeleteApplication))
		apiUI.DELETE("/projects/:id", s.WrapHandler(webHandler.DeleteProject))
	} else {
		log.Warn(context.Background()).Msg("Web handler is nil, skipping web UI routes")
	}

	// REST API routes
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
