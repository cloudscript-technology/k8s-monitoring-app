package monitoring

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"k8s-monitoring-app/internal/k8s"
	serverModel "k8s-monitoring-app/internal/server/model"
	applicationModel "k8s-monitoring-app/pkg/application/model"
	applicationMetricModel "k8s-monitoring-app/pkg/application_metric/model"
	applicationMetricValueModel "k8s-monitoring-app/pkg/application_metric_value/model"
	metricTypeModel "k8s-monitoring-app/pkg/metric_type/model"

	"github.com/robfig/cron/v3"
	"gitlab.cloudscript.com.br/general/go-instrumentation.git/log"
	corev1 "k8s.io/api/core/v1"
)

type MonitoringService struct {
	cron      *cron.Cron
	k8sClient *k8s.Client
	db        *sql.DB
}

func NewMonitoringService(db *sql.DB) (*MonitoringService, error) {
	k8sClient, err := k8s.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	c := cron.New()

	return &MonitoringService{
		cron:      c,
		k8sClient: k8sClient,
		db:        db,
	}, nil
}

func (m *MonitoringService) Start() error {
	// Schedule the monitoring job to run every minute
	_, err := m.cron.AddFunc("@every 1m", m.collectMetrics)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	m.cron.Start()
	log.Info(context.Background()).Msg("Monitoring service started")

	return nil
}

func (m *MonitoringService) Stop() {
	ctx := m.cron.Stop()
	<-ctx.Done()
	log.Info(context.Background()).Msg("Monitoring service stopped")
}

func (m *MonitoringService) collectMetrics() {
	ctx := context.Background()
	log.Info(ctx).Msg("Starting metric collection")

	// Get all application metrics
	applicationMetrics, err := serverModel.ServerRepos.ApplicationMetric.List(ctx)
	if err != nil {
		log.Error(ctx, err).Msg("failed to list application metrics")
		return
	}

	for _, appMetric := range applicationMetrics {
		// Get the application details
		application, err := serverModel.ServerRepos.Application.Get(ctx, appMetric.ApplicationID)
		if err != nil {
			log.Error(ctx, err).Str("application_id", appMetric.ApplicationID).Msg("failed to get application")
			continue
		}

		// Get the metric type
		metricType, err := serverModel.ServerRepos.MetricType.Get(ctx, appMetric.TypeID)
		if err != nil {
			log.Error(ctx, err).Str("metric_type_id", appMetric.TypeID).Msg("failed to get metric type")
			continue
		}

		// Collect the metric based on type
		if err := m.collectMetricByType(ctx, &application, &metricType, &appMetric); err != nil {
			log.Error(ctx, err).
				Str("application", application.Name).
				Str("metric_type", metricType.Name).
				Msg("failed to collect metric")
		}
	}

	log.Info(ctx).Msg("Metric collection completed")
}

func (m *MonitoringService) collectMetricByType(
	ctx context.Context,
	application *applicationModel.Application,
	metricType *metricTypeModel.MetricType,
	appMetric *applicationMetricModel.ApplicationMetric,
) error {
	// Parse configuration
	var config applicationMetricModel.Configuration
	if err := json.Unmarshal(appMetric.Configuration, &config); err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	var metricValue applicationMetricValueModel.MetricValue
	var err error

	switch metricType.Name {
	case "HealthCheck":
		metricValue, err = m.collectHealthCheck(ctx, &config)
	case "PodStatus":
		metricValue, err = m.collectPodStatus(ctx, application, &config)
	case "PodMemoryUsage":
		metricValue, err = m.collectPodMemoryUsage(ctx, application, &config)
	case "PodCpuUsage":
		metricValue, err = m.collectPodCpuUsage(ctx, application, &config)
	case "PvcUsage":
		metricValue, err = m.collectPvcUsage(ctx, application, &config)
	case "PodActiveNodes":
		metricValue, err = m.collectPodActiveNodes(ctx, application, &config)
	default:
		return fmt.Errorf("unknown metric type: %s", metricType.Name)
	}

	if err != nil {
		return err
	}

	// Store the metric value
	return m.storeMetricValue(ctx, appMetric.ID, metricValue)
}

func (m *MonitoringService) collectHealthCheck(ctx context.Context, config *applicationMetricModel.Configuration) (applicationMetricValueModel.MetricValue, error) {
	result := m.k8sClient.PerformHealthCheck(
		ctx,
		config.HealthCheckURL,
		config.Method,
		config.ExpectedStatus,
		config.TimeoutSeconds,
	)

	return applicationMetricValueModel.MetricValue{
		Status:         result.Status,
		ResponseTimeMs: result.ResponseTimeMs,
		StatusCode:     result.StatusCode,
		ErrorMessage:   result.ErrorMessage,
	}, nil
}

func (m *MonitoringService) collectPodStatus(
	ctx context.Context,
	application *applicationModel.Application,
	config *applicationMetricModel.Configuration,
) (applicationMetricValueModel.MetricValue, error) {
	pods, err := m.k8sClient.GetPodsByLabelSelector(ctx, application.Namespace, config.PodLabelSelector)
	if err != nil {
		return applicationMetricValueModel.MetricValue{}, err
	}

	if len(pods.Items) == 0 {
		return applicationMetricValueModel.MetricValue{
			PodPhase:     "NotFound",
			PodReady:     false,
			RestartCount: 0,
		}, nil
	}

	// Use the first pod or aggregate if multiple
	pod := pods.Items[0]

	// Find the specific container if specified
	var restartCount int32
	var containerReady bool

	if config.ContainerName != "" {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Name == config.ContainerName {
				restartCount = containerStatus.RestartCount
				containerReady = containerStatus.Ready
				break
			}
		}
	} else {
		// Use first container
		if len(pod.Status.ContainerStatuses) > 0 {
			restartCount = pod.Status.ContainerStatuses[0].RestartCount
			containerReady = pod.Status.ContainerStatuses[0].Ready
		}
	}

	return applicationMetricValueModel.MetricValue{
		PodPhase:     string(pod.Status.Phase),
		PodReady:     containerReady,
		RestartCount: restartCount,
	}, nil
}

func (m *MonitoringService) collectPodMemoryUsage(
	ctx context.Context,
	application *applicationModel.Application,
	config *applicationMetricModel.Configuration,
) (applicationMetricValueModel.MetricValue, error) {
	pods, err := m.k8sClient.GetPodsByLabelSelector(ctx, application.Namespace, config.PodLabelSelector)
	if err != nil {
		return applicationMetricValueModel.MetricValue{}, err
	}

	if len(pods.Items) == 0 {
		return applicationMetricValueModel.MetricValue{
			MemoryUsageBytes: 0,
			MemoryLimitBytes: 0,
			MemoryPercent:    0,
		}, nil
	}

	pod := pods.Items[0]

	// Get pod metrics
	podMetrics, err := m.k8sClient.GetPodMetrics(ctx, application.Namespace, pod.Name)
	if err != nil {
		return applicationMetricValueModel.MetricValue{}, err
	}

	var memoryUsage int64
	var memoryLimit int64

	// Find the specific container
	for i, container := range podMetrics.Containers {
		if config.ContainerName == "" || container.Name == config.ContainerName {
			memoryUsage = container.Usage.Memory().Value()

			// Get memory limit from pod spec
			if i < len(pod.Spec.Containers) {
				if limit, ok := pod.Spec.Containers[i].Resources.Limits[corev1.ResourceMemory]; ok {
					memoryLimit = limit.Value()
				}
			}
			break
		}
	}

	var memoryPercent float64
	if memoryLimit > 0 {
		memoryPercent = float64(memoryUsage) / float64(memoryLimit) * 100
	}

	return applicationMetricValueModel.MetricValue{
		MemoryUsageBytes: memoryUsage,
		MemoryLimitBytes: memoryLimit,
		MemoryPercent:    memoryPercent,
	}, nil
}

func (m *MonitoringService) collectPodCpuUsage(
	ctx context.Context,
	application *applicationModel.Application,
	config *applicationMetricModel.Configuration,
) (applicationMetricValueModel.MetricValue, error) {
	pods, err := m.k8sClient.GetPodsByLabelSelector(ctx, application.Namespace, config.PodLabelSelector)
	if err != nil {
		return applicationMetricValueModel.MetricValue{}, err
	}

	if len(pods.Items) == 0 {
		return applicationMetricValueModel.MetricValue{
			CpuUsageMillicores: 0,
			CpuLimitMillicores: 0,
			CpuPercent:         0,
		}, nil
	}

	pod := pods.Items[0]

	// Get pod metrics
	podMetrics, err := m.k8sClient.GetPodMetrics(ctx, application.Namespace, pod.Name)
	if err != nil {
		return applicationMetricValueModel.MetricValue{}, err
	}

	var cpuUsage int64
	var cpuLimit int64

	// Find the specific container
	for i, container := range podMetrics.Containers {
		if config.ContainerName == "" || container.Name == config.ContainerName {
			cpuUsage = container.Usage.Cpu().MilliValue()

			// Get CPU limit from pod spec
			if i < len(pod.Spec.Containers) {
				if limit, ok := pod.Spec.Containers[i].Resources.Limits[corev1.ResourceCPU]; ok {
					cpuLimit = limit.MilliValue()
				}
			}
			break
		}
	}

	var cpuPercent float64
	if cpuLimit > 0 {
		cpuPercent = float64(cpuUsage) / float64(cpuLimit) * 100
	}

	return applicationMetricValueModel.MetricValue{
		CpuUsageMillicores: cpuUsage,
		CpuLimitMillicores: cpuLimit,
		CpuPercent:         cpuPercent,
	}, nil
}

func (m *MonitoringService) collectPvcUsage(
	ctx context.Context,
	application *applicationModel.Application,
	config *applicationMetricModel.Configuration,
) (applicationMetricValueModel.MetricValue, error) {
	pvc, err := m.k8sClient.GetPVCUsage(ctx, application.Namespace, config.PvcName)
	if err != nil {
		return applicationMetricValueModel.MetricValue{}, err
	}

	capacity := pvc.Status.Capacity[corev1.ResourceStorage]
	capacityBytes := capacity.Value()

	// Note: Kubernetes doesn't directly provide used space for PVCs
	// This would typically require additional metrics from the storage system
	// For now, we'll just store the capacity

	return applicationMetricValueModel.MetricValue{
		PvcCapacityBytes: capacityBytes,
		PvcUsedBytes:     0, // Would need storage system integration
		PvcPercent:       0,
	}, nil
}

func (m *MonitoringService) collectPodActiveNodes(
	ctx context.Context,
	application *applicationModel.Application,
	config *applicationMetricModel.Configuration,
) (applicationMetricValueModel.MetricValue, error) {
	nodes, err := m.k8sClient.GetNodesForPods(ctx, application.Namespace, config.PodLabelSelector)
	if err != nil {
		return applicationMetricValueModel.MetricValue{}, err
	}

	return applicationMetricValueModel.MetricValue{
		ActiveNodesCount: len(nodes),
		NodeNames:        nodes,
	}, nil
}

func (m *MonitoringService) storeMetricValue(
	ctx context.Context,
	applicationMetricID string,
	metricValue applicationMetricValueModel.MetricValue,
) error {
	valueJSON, err := json.Marshal(metricValue)
	if err != nil {
		return fmt.Errorf("failed to marshal metric value: %w", err)
	}

	metricValueRecord := applicationMetricValueModel.ApplicationMetricValue{
		ApplicationMetricID: applicationMetricID,
		Value:               valueJSON,
	}

	if err := serverModel.ServerRepos.ApplicationMetricValue.Add(ctx, &metricValueRecord); err != nil {
		return fmt.Errorf("failed to store metric value: %w", err)
	}

	return nil
}
