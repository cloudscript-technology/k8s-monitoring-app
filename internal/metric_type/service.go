package metric_type

import (
	"errors"
	"net/http"

	"k8s-monitoring-app/internal/core"
	serverModel "k8s-monitoring-app/internal/server/model"
	model "k8s-monitoring-app/pkg/metric_type/model"

	"github.com/rs/zerolog/log"
)

type service struct{}

func NewService() model.Service {
	return &service{}
}

func (s *service) Get(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	id := sc.Param("id")

	if len(id) == 0 {
		log.Error().Err(errors.New("id is empty")).Msg("error getting metric type")
		return sc.String(http.StatusBadRequest, "invalid request")
	}

	metricType, err := serverModel.ServerRepos.MetricType.Get(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("error getting metric type")
		return sc.String(http.StatusNotFound, "metric type not found")
	}

	return sc.JSON(http.StatusOK, metricType)
}

func (s *service) List(sc *core.HTTPServerContext) error {
	ctx := sc.Request().Context()

	metricTypes, err := serverModel.ServerRepos.MetricType.List(ctx)
	if err != nil {
		log.Error().Err(err).Msg("error listing metric types")
		return sc.String(http.StatusInternalServerError, "internal server error")
	}

	return sc.JSON(http.StatusOK, metricTypes)
}
