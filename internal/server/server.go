package server

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"k8s-monitoring-app/internal/auth"
	"k8s-monitoring-app/internal/core"
	model "k8s-monitoring-app/internal/server/model"

	applicationService "k8s-monitoring-app/internal/application"
	applicationRepositories "k8s-monitoring-app/internal/application/repository"
	applicationMetricService "k8s-monitoring-app/internal/application_metric"
	applicationMetricRepositories "k8s-monitoring-app/internal/application_metric/repository"
	applicationMetricValueService "k8s-monitoring-app/internal/application_metric_value"
	applicationMetricValueRepositories "k8s-monitoring-app/internal/application_metric_value/repository"
	metricTypeService "k8s-monitoring-app/internal/metric_type"
	metricTypeRepositories "k8s-monitoring-app/internal/metric_type/repository"
	projectService "k8s-monitoring-app/internal/project"
	projectRepositories "k8s-monitoring-app/internal/project/repository"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"gitlab.cloudscript.com.br/general/go-instrumentation.git/apmtracer"
	"gitlab.cloudscript.com.br/general/go-instrumentation.git/log"
	"gitlab.cloudscript.com.br/general/go-instrumentation.git/middleware"
	apmecho "go.elastic.co/apm/module/apmechov4/v2"
	_ "go.elastic.co/apm/module/apmsql/pq"
)

// findMigrationsPath tries to find the migrations directory from different working directories
func findMigrationsPath() string {
	possiblePaths := []string{
		"database/migrations",       // Running from project root
		"../database/migrations",    // Running from cmd directory
		"../../database/migrations", // Running from deeper directory
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return "file://" + absPath
		}
	}

	// Default fallback (running from cmd)
	return "file://../database/migrations"
}

func NewHTTPServer(config *core.ApiServiceConfiguration) (*core.HTTPServer, error) {
	// Initialize OAuth
	if err := auth.InitOAuth(); err != nil {
		log.Warn(context.Background()).Err(err).Msg("OAuth not configured - authentication will be disabled")
	}

	d, err := core.ConnectDatabase()
	if err != nil {
		return nil, err
	}

	driver, err := postgres.WithInstance(d, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	migrationsPath := findMigrationsPath()
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		return nil, err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	e := echo.New()
	s := &core.HTTPServer{
		Config:   config,
		Api:      e,
		Postgres: d,
	}

	model.ServerSvc = &model.ServerServices{
		Project:                projectService.NewService(),
		Application:            applicationService.NewService(),
		MetricType:             metricTypeService.NewService(),
		ApplicationMetric:      applicationMetricService.NewService(),
		ApplicationMetricValue: applicationMetricValueService.NewService(),
	}

	model.ServerRepos = &model.ServerRepositories{
		Project:                projectRepositories.NewRepo(d),
		Application:            applicationRepositories.NewRepo(d),
		MetricType:             metricTypeRepositories.NewRepo(d),
		ApplicationMetric:      applicationMetricRepositories.NewRepo(d),
		ApplicationMetricValue: applicationMetricValueRepositories.NewRepo(d),
	}

	e.Use(
		middleware.LogMiddlewareRequestLogger(),
		// middleware.ApmMiddlewareCaptureBody(),
		apmecho.Middleware(apmecho.WithTracer(apmtracer.Tracer)),
		auth.AuthMiddleware(), // Apply OAuth authentication middleware
	)

	bindRoutes(s)

	return s, nil
}
