package server

import (
	applicationRepo "k8s-monitoring-app/internal/application/repository"
	applicationMetricRepo "k8s-monitoring-app/internal/application_metric/repository"
	applicationMetricValueRepo "k8s-monitoring-app/internal/application_metric_value/repository"
	metricTypeRepo "k8s-monitoring-app/internal/metric_type/repository"
	projectRepo "k8s-monitoring-app/internal/project/repository"
	applicationModel "k8s-monitoring-app/pkg/application/model"
	applicationMetricModel "k8s-monitoring-app/pkg/application_metric/model"
	applicationMetricValueModel "k8s-monitoring-app/pkg/application_metric_value/model"
	metricTypeModel "k8s-monitoring-app/pkg/metric_type/model"
	projectModel "k8s-monitoring-app/pkg/project/model"
)

var ServerSvc *ServerServices
var ServerRepos *ServerRepositories

type ServerServices struct {
	Project                projectModel.Service
	Application            applicationModel.Service
	MetricType             metricTypeModel.Service
	ApplicationMetric      applicationMetricModel.Service
	ApplicationMetricValue applicationMetricValueModel.Service
}

type ServerRepositories struct {
	Project                projectRepo.Repository
	MetricType             metricTypeRepo.Repository
	Application            applicationRepo.Repository
	ApplicationMetric      applicationMetricRepo.Repository
	ApplicationMetricValue applicationMetricValueRepo.Repository
}
