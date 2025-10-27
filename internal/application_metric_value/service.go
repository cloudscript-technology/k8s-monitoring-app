package application_metric_value

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"k8s-monitoring-app/internal/core"
	serverModel "k8s-monitoring-app/internal/server/model"
	model "k8s-monitoring-app/pkg/application_metric_value/model"

	"gitlab.cloudscript.com.br/general/go-instrumentation.git/log"
)

type service struct{}

func NewService() model.Service {
	return &service{}
}

func (s *service) Get(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	id := sc.Param("id")

	if len(id) == 0 {
		log.Error(ctx, errors.New("id is empty")).Msg("error getting metric value")
		return sc.String(http.StatusBadRequest, "invalid request")
	}

	metricValue, err := serverModel.ServerRepos.ApplicationMetricValue.Get(ctx, id)
	if err != nil {
		log.Error(ctx, err).Msg("error getting metric value")
		return sc.String(http.StatusNotFound, "metric value not found")
	}

	return sc.JSON(http.StatusOK, metricValue)
}

func (s *service) ListByApplicationMetric(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	applicationMetricID := sc.Param("application_metric_id")

	if len(applicationMetricID) == 0 {
		log.Error(ctx, errors.New("application_metric_id is empty")).Msg("error listing metric values")
		return sc.String(http.StatusBadRequest, "invalid request")
	}

	// Get limit from query parameter (default 100)
	limitStr := sc.QueryParam("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			if limit > 1000 {
				limit = 1000 // Max limit
			}
		}
	}

	metricValues, err := serverModel.ServerRepos.ApplicationMetricValue.ListByApplicationMetric(ctx, applicationMetricID, limit)
	if err != nil {
		log.Error(ctx, err).Msg("error listing metric values")
		return sc.String(http.StatusInternalServerError, "internal server error")
	}

	return sc.JSON(http.StatusOK, metricValues)
}

func (s *service) GetLatestByApplication(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	applicationID := sc.Param("application_id")

	if len(applicationID) == 0 {
		log.Error(ctx, errors.New("application_id is empty")).Msg("error getting latest metrics")
		return sc.String(http.StatusBadRequest, "invalid request")
	}

	// Get application details
	application, err := serverModel.ServerRepos.Application.Get(ctx, applicationID)
	if err != nil {
		log.Error(ctx, err).Msg("error getting application")
		return sc.String(http.StatusNotFound, "application not found")
	}

	// Get project details
	project, err := serverModel.ServerRepos.Project.Get(ctx, application.ProjectID)
	if err != nil {
		log.Error(ctx, err).Msg("error getting project")
		// Continue even if project not found
	}

	// Get all metrics for this application
	applicationMetrics, err := serverModel.ServerRepos.ApplicationMetric.ListByApplication(ctx, applicationID)
	if err != nil {
		log.Error(ctx, err).Msg("error listing application metrics")
		return sc.String(http.StatusInternalServerError, "internal server error")
	}

	// For each metric, get the latest value and metric type details
	type MetricWithLatestValue struct {
		// Metric Configuration
		MetricID      string          `json:"metric_id"`
		MetricTypeID  string          `json:"metric_type_id"`
		Configuration json.RawMessage `json:"configuration"`

		// Metric Type Details
		MetricTypeName        string `json:"metric_type_name"`
		MetricTypeDescription string `json:"metric_type_description"`

		// Latest Value
		LatestValue *model.ApplicationMetricValue `json:"latest_value,omitempty"`
	}

	type ResponseData struct {
		// Application Details
		ApplicationID          string `json:"application_id"`
		ApplicationName        string `json:"application_name"`
		ApplicationDescription string `json:"application_description"`
		ApplicationNamespace   string `json:"application_namespace"`

		// Project Details
		ProjectID          string `json:"project_id"`
		ProjectName        string `json:"project_name"`
		ProjectDescription string `json:"project_description"`

		// Metrics
		Metrics []MetricWithLatestValue `json:"metrics"`
	}

	response := ResponseData{
		ApplicationID:          application.ID,
		ApplicationName:        application.Name,
		ApplicationDescription: application.Description,
		ApplicationNamespace:   application.Namespace,
		ProjectID:              project.ID,
		ProjectName:            project.Name,
		ProjectDescription:     project.Description,
		Metrics:                make([]MetricWithLatestValue, 0),
	}

	for _, metric := range applicationMetrics {
		// Get metric type details
		metricType, err := serverModel.ServerRepos.MetricType.Get(ctx, metric.TypeID)
		if err != nil {
			log.Error(ctx, err).Str("metric_type_id", metric.TypeID).Msg("error getting metric type")
			continue
		}

		metricWithValue := MetricWithLatestValue{
			MetricID:              metric.ID,
			MetricTypeID:          metric.TypeID,
			Configuration:         metric.Configuration,
			MetricTypeName:        metricType.Name,
			MetricTypeDescription: metricType.Description,
		}

		// Get the latest value (limit 1)
		values, err := serverModel.ServerRepos.ApplicationMetricValue.ListByApplicationMetric(ctx, metric.ID, 1)
		if err == nil && len(values) > 0 {
			metricWithValue.LatestValue = &values[0]
		}

		response.Metrics = append(response.Metrics, metricWithValue)
	}

	return sc.JSON(http.StatusOK, response)
}
