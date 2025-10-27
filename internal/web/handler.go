package web

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"

	"k8s-monitoring-app/internal/core"
	serverModel "k8s-monitoring-app/internal/server/model"
	applicationModel "k8s-monitoring-app/pkg/application/model"
	applicationMetricValueModel "k8s-monitoring-app/pkg/application_metric_value/model"
	projectModel "k8s-monitoring-app/pkg/project/model"

	"gitlab.cloudscript.com.br/general/go-instrumentation.git/log"
)

type Handler struct {
	templates *template.Template
}

func NewHandler() (*Handler, error) {
	// Create template with custom functions
	funcMap := template.FuncMap{
		"div": func(a, b interface{}) float64 {
			aFloat, _ := toFloat64(a)
			bFloat, _ := toFloat64(b)
			if bFloat == 0 {
				return 0
			}
			return aFloat / bFloat
		},
		"add": func(a, b int) int {
			return a + b
		},
	}

	templates, err := template.New("").Funcs(funcMap).ParseGlob("web/templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		templates: templates,
	}, nil
}

// Helper function to convert interface{} to float64
func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case int32:
		return float64(val), nil
	default:
		return 0, nil
	}
}

// Dashboard renders the main dashboard page
func (h *Handler) Dashboard(sc *core.HTTPServerContext) error {
	data := map[string]interface{}{
		"Title": "Dashboard",
	}

	sc.Response().Header().Set("Content-Type", "text/html")
	sc.Response().WriteHeader(http.StatusOK)

	if err := h.templates.ExecuteTemplate(sc.Response().Writer, "layout.html", data); err != nil {
		log.Error(sc.Request().Context(), err).Msg("error executing template")
		return err
	}

	return nil
}

type ProjectWithApplications struct {
	Project      projectModel.Project
	Applications []applicationModel.Application
}

// GetProjects returns all projects with their applications
func (h *Handler) GetProjects(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	log.Info(ctx).Msg("GetProjects called")

	// Get all projects
	projects, err := serverModel.ServerRepos.Project.List(ctx)
	if err != nil {
		log.Error(ctx, err).Msg("error listing projects")
		return sc.String(http.StatusInternalServerError, "Error loading projects")
	}

	log.Info(ctx).Int("count", len(projects)).Msg("Projects retrieved")

	// Get applications for each project
	var projectsWithApps []ProjectWithApplications
	for _, project := range projects {
		apps, err := serverModel.ServerRepos.Application.ListByProject(ctx, project.ID)
		if err != nil {
			log.Error(ctx, err).Str("project_id", project.ID).Msg("error listing applications")
			continue
		}

		projectsWithApps = append(projectsWithApps, ProjectWithApplications{
			Project:      project,
			Applications: apps,
		})
	}

	// Render the projects template using pre-loaded templates
	sc.Response().Header().Set("Content-Type", "text/html")
	sc.Response().WriteHeader(http.StatusOK)

	for _, projectWithApps := range projectsWithApps {
		if err := h.templates.ExecuteTemplate(sc.Response().Writer, "project-card", projectWithApps); err != nil {
			log.Error(ctx, err).Msg("error executing template")
			return err
		}
	}

	return nil
}

type ApplicationMetricsView struct {
	ApplicationID          string
	ApplicationName        string
	ApplicationDescription string
	ApplicationNamespace   string
	MetricsByType          map[string]*MetricWithValue
}

type MetricWithValue struct {
	MetricID      string
	MetricTypeID  string
	Configuration map[string]interface{}
	LatestValue   *MetricValueParsed
}

type MetricValueParsed struct {
	ID                  string
	ApplicationMetricID string
	Value               map[string]interface{}
	CreatedAt           string
	UpdatedAt           string
}

// GetApplicationMetrics returns metrics for a specific application
func (h *Handler) GetApplicationMetrics(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	applicationID := sc.Param("id")
	if applicationID == "" {
		return sc.String(http.StatusBadRequest, "Application ID required")
	}

	// Get application details
	application, err := serverModel.ServerRepos.Application.Get(ctx, applicationID)
	if err != nil {
		log.Error(ctx, err).Str("application_id", applicationID).Msg("error getting application")
		return sc.String(http.StatusNotFound, "Application not found")
	}

	// Get all metrics for this application
	applicationMetrics, err := serverModel.ServerRepos.ApplicationMetric.ListByApplication(ctx, applicationID)
	if err != nil {
		log.Error(ctx, err).Str("application_id", applicationID).Msg("error listing application metrics")
		return sc.String(http.StatusInternalServerError, "Error loading metrics")
	}

	// Build metrics map organized by type
	metricsByType := make(map[string]*MetricWithValue)

	for _, metric := range applicationMetrics {
		// Get metric type details
		metricType, err := serverModel.ServerRepos.MetricType.Get(ctx, metric.TypeID)
		if err != nil {
			log.Error(ctx, err).Str("metric_type_id", metric.TypeID).Msg("error getting metric type")
			continue
		}

		metricWithValue := &MetricWithValue{
			MetricID:     metric.ID,
			MetricTypeID: metric.TypeID,
		}

		// Parse Configuration from JSON to map
		var config map[string]interface{}
		if err := json.Unmarshal(metric.Configuration, &config); err == nil {
			metricWithValue.Configuration = config
		}

		// Get the latest value (limit 1)
		values, err := serverModel.ServerRepos.ApplicationMetricValue.ListByApplicationMetric(ctx, metric.ID, 1)
		if err == nil && len(values) > 0 {
			// Parse the Value from JSON to map
			var valueMap map[string]interface{}
			if err := json.Unmarshal(values[0].Value, &valueMap); err == nil {
				metricWithValue.LatestValue = &MetricValueParsed{
					ID:                  values[0].ID,
					ApplicationMetricID: values[0].ApplicationMetricID,
					Value:               valueMap,
					CreatedAt:           values[0].CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
					UpdatedAt:           values[0].UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
				}
			}
		}

		metricsByType[metricType.Name] = metricWithValue
	}

	data := ApplicationMetricsView{
		ApplicationID:          application.ID,
		ApplicationName:        application.Name,
		ApplicationDescription: application.Description,
		ApplicationNamespace:   application.Namespace,
		MetricsByType:          metricsByType,
	}

	log.Info(ctx).
		Str("app_name", data.ApplicationName).
		Int("metrics_count", len(metricsByType)).
		Msg("Rendering application metrics")

	// Debug: log metric types
	for metricTypeName, metric := range metricsByType {
		hasValue := metric.LatestValue != nil
		log.Info(ctx).
			Str("metric_type", metricTypeName).
			Bool("has_value", hasValue).
			Msg("Metric in map")
	}

	// Render the application metrics template using pre-loaded templates with custom functions
	sc.Response().Header().Set("Content-Type", "text/html")
	sc.Response().WriteHeader(http.StatusOK)

	if err := h.templates.ExecuteTemplate(sc.Response().Writer, "application-metrics", data); err != nil {
		log.Error(ctx, err).
			Str("app_id", applicationID).
			Str("app_name", data.ApplicationName).
			Msg("error executing template - check template syntax")
		return err
	}

	return nil
}

// Helper function to get metric value by type
func getMetricValueByType(ctx context.Context, applicationID, metricTypeName string) *applicationMetricValueModel.ApplicationMetricValue {
	// Get all metrics for the application
	applicationMetrics, err := serverModel.ServerRepos.ApplicationMetric.ListByApplication(ctx, applicationID)
	if err != nil {
		return nil
	}

	// Find the metric with the specified type
	for _, metric := range applicationMetrics {
		metricType, err := serverModel.ServerRepos.MetricType.Get(ctx, metric.TypeID)
		if err != nil {
			continue
		}

		if metricType.Name == metricTypeName {
			// Get the latest value
			values, err := serverModel.ServerRepos.ApplicationMetricValue.ListByApplicationMetric(ctx, metric.ID, 1)
			if err == nil && len(values) > 0 {
				return &values[0]
			}
		}
	}

	return nil
}
