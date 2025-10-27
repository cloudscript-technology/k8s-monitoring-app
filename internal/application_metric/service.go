package application_metric

import (
	"errors"
	"net/http"

	"k8s-monitoring-app/internal/core"
	serverModel "k8s-monitoring-app/internal/server/model"
	model "k8s-monitoring-app/pkg/application_metric/model"

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
		log.Error(ctx, errors.New("id is empty")).Msg("error getting application metric")
		return sc.String(http.StatusBadRequest, "invalid request")
	}

	applicationMetric, err := serverModel.ServerRepos.ApplicationMetric.Get(ctx, id)
	if err != nil {
		log.Error(ctx, err).Msg("error getting application metric")
		return sc.String(http.StatusNotFound, "application metric not found")
	}

	return sc.JSON(http.StatusOK, applicationMetric)
}

func (s *service) List(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	applicationMetrics, err := serverModel.ServerRepos.ApplicationMetric.List(ctx)
	if err != nil {
		log.Error(ctx, err).Msg("error listing application metrics")
		return sc.String(http.StatusInternalServerError, "internal server error")
	}

	return sc.JSON(http.StatusOK, applicationMetrics)
}

func (s *service) ListByApplication(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	applicationID := sc.Param("application_id")

	if len(applicationID) == 0 {
		log.Error(ctx, errors.New("application_id is empty")).Msg("error listing application metrics by application")
		return sc.String(http.StatusBadRequest, "invalid request")
	}

	applicationMetrics, err := serverModel.ServerRepos.ApplicationMetric.ListByApplication(ctx, applicationID)
	if err != nil {
		log.Error(ctx, err).Msg("error listing application metrics by application")
		return sc.String(http.StatusInternalServerError, "internal server error")
	}

	return sc.JSON(http.StatusOK, applicationMetrics)
}

func (s *service) Add(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	applicationMetric := model.ApplicationMetric{}
	if err := sc.Bind(&applicationMetric); err != nil {
		log.Error(ctx, err).Msg("error binding application metric")
		return sc.String(http.StatusBadRequest, "invalid request body")
	}

	// Validate that the application exists
	_, err := serverModel.ServerRepos.Application.Get(ctx, applicationMetric.ApplicationID)
	if err != nil {
		log.Error(ctx, err).Msg("error getting application")
		return sc.String(http.StatusBadRequest, "application not found")
	}

	// Validate that the metric type exists
	_, err = serverModel.ServerRepos.MetricType.Get(ctx, applicationMetric.TypeID)
	if err != nil {
		log.Error(ctx, err).Msg("error getting metric type")
		return sc.String(http.StatusBadRequest, "metric type not found")
	}

	if err := serverModel.ServerRepos.ApplicationMetric.Add(ctx, &applicationMetric); err != nil {
		log.Error(ctx, err).Msg("error add application metric")
		return sc.String(http.StatusInternalServerError, "internal server error")
	}

	return sc.JSON(http.StatusCreated, applicationMetric)
}

func (s *service) Update(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()
	id := sc.Param("id")

	// First get the existing application metric to check it exists
	_, err := serverModel.ServerRepos.ApplicationMetric.Get(ctx, id)
	if err != nil {
		log.Error(ctx, err).Msg("error getting application metric")
		return sc.String(http.StatusNotFound, "application metric not found")
	}

	applicationMetric := model.ApplicationMetric{}
	if err := sc.Bind(&applicationMetric); err != nil {
		log.Error(ctx, err).Msg("error binding application metric")
		return sc.String(http.StatusBadRequest, "Invalid Request")
	}
	applicationMetric.ID = id

	// Validate that the metric type exists if it's being changed
	if applicationMetric.TypeID != "" {
		_, err := serverModel.ServerRepos.MetricType.Get(ctx, applicationMetric.TypeID)
		if err != nil {
			log.Error(ctx, err).Msg("error getting metric type")
			return sc.String(http.StatusBadRequest, "metric type not found")
		}
	}

	if err := serverModel.ServerRepos.ApplicationMetric.Update(ctx, &applicationMetric); err != nil {
		log.Error(ctx, err).Msg("error updating application metric")
		return sc.String(http.StatusInternalServerError, "Internal Server Error")
	}

	return sc.JSON(http.StatusOK, applicationMetric)
}

func (s *service) Delete(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()
	id := sc.Param("id")

	if len(id) == 0 {
		log.Error(ctx, errors.New("id is empty")).Msg("error deleting application metric")
		return sc.String(http.StatusBadRequest, "Invalid Request")
	}

	err := serverModel.ServerRepos.ApplicationMetric.Delete(ctx, id)
	if err != nil {
		log.Error(ctx, err).Msg("error deleting application metric")
		return sc.String(http.StatusInternalServerError, "Internal Server Error")
	}

	return sc.JSON(http.StatusOK, map[string]bool{"success": true})
}
