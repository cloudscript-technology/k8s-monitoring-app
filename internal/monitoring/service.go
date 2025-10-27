package monitoring

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"k8s-monitoring-app/internal/connections"
	"k8s-monitoring-app/internal/env"
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
	// Get collection interval from environment (default: 60 seconds)
	collectionInterval := env.METRICS_COLLECTION_INTERVAL
	if collectionInterval <= 0 {
		collectionInterval = 60
	}

	// Schedule the monitoring job using the configured interval
	cronSpec := fmt.Sprintf("@every %ds", collectionInterval)
	_, err := m.cron.AddFunc(cronSpec, m.collectMetrics)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	// Schedule the cleanup job to run based on configuration
	cleanupInterval := env.METRICS_CLEANUP_INTERVAL
	if cleanupInterval == "" {
		cleanupInterval = "0 2 * * *" // Default: daily at 2 AM
	}

	_, err = m.cron.AddFunc(cleanupInterval, m.cleanupOldMetrics)
	if err != nil {
		return fmt.Errorf("failed to add cleanup cron job: %w", err)
	}

	m.cron.Start()
	log.Info(context.Background()).
		Int("collection_interval_seconds", collectionInterval).
		Int("retention_days", env.METRICS_RETENTION_DAYS).
		Str("cleanup_interval", cleanupInterval).
		Msg("Monitoring service started")

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
	case "RedisConnection":
		metricValue = m.collectRedisConnection(ctx, &config)
	case "PostgreSQLConnection":
		metricValue = m.collectPostgreSQLConnection(ctx, &config)
	case "MongoDBConnection":
		metricValue = m.collectMongoDBConnection(ctx, &config)
	case "MySQLConnection":
		metricValue = m.collectMySQLConnection(ctx, &config)
	case "KongConnection":
		metricValue = m.collectKongConnection(ctx, &config)
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
			TotalPods:    0,
			ReadyPods:    0,
			Pods:         []applicationMetricValueModel.PodInfo{},
		}, nil
	}

	// Aggregate status from all pods
	totalPods := len(pods.Items)
	runningPods := 0
	readyPods := 0
	totalRestarts := int32(0)
	hasFailedPods := false
	hasPendingPods := false

	// Collect individual pod information
	podInfos := make([]applicationMetricValueModel.PodInfo, 0, len(pods.Items))

	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			runningPods++
		} else if pod.Status.Phase == corev1.PodFailed {
			hasFailedPods = true
		} else if pod.Status.Phase == corev1.PodPending {
			hasPendingPods = true
		}

		// Check container status for this specific pod
		podReady := false
		podRestartCount := int32(0)

		if config.ContainerName != "" {
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.Name == config.ContainerName {
					podReady = containerStatus.Ready
					podRestartCount = containerStatus.RestartCount
					if containerStatus.Ready {
						readyPods++
					}
					totalRestarts += containerStatus.RestartCount
					break
				}
			}
		} else {
			// Use first container
			if len(pod.Status.ContainerStatuses) > 0 {
				podReady = pod.Status.ContainerStatuses[0].Ready
				podRestartCount = pod.Status.ContainerStatuses[0].RestartCount
				if pod.Status.ContainerStatuses[0].Ready {
					readyPods++
				}
				totalRestarts += pod.Status.ContainerStatuses[0].RestartCount
			}
		}

		// Add individual pod info
		podInfos = append(podInfos, applicationMetricValueModel.PodInfo{
			Name:         pod.Name,
			Phase:        string(pod.Status.Phase),
			Ready:        podReady,
			RestartCount: podRestartCount,
			NodeName:     pod.Spec.NodeName,
			IP:           pod.Status.PodIP,
		})
	}

	// Determine overall phase
	overallPhase := "Running"
	overallReady := true

	if hasFailedPods {
		overallPhase = "Degraded" // Some pods failed
		overallReady = false
	} else if readyPods < totalPods {
		overallPhase = "Running"
		overallReady = false // Not all pods ready
	} else if hasPendingPods {
		overallPhase = "Pending"
		overallReady = false
	}

	return applicationMetricValueModel.MetricValue{
		PodPhase:     overallPhase,
		PodReady:     overallReady,
		RestartCount: totalRestarts,
		TotalPods:    totalPods,
		ReadyPods:    readyPods,
		Pods:         podInfos,
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
	// Get PVC usage with disk info by executing df in the pod
	usageInfo, err := m.k8sClient.GetPVCUsageWithDiskInfo(
		ctx,
		application.Namespace,
		config.PvcName,
		config.PodLabelSelector,
		config.ContainerName,
		config.PvcMountPath,
	)
	if err != nil {
		return applicationMetricValueModel.MetricValue{}, fmt.Errorf("failed to get PVC usage: %w", err)
	}

	return applicationMetricValueModel.MetricValue{
		PvcCapacityBytes: usageInfo.CapacityBytes,
		PvcUsedBytes:     usageInfo.UsedBytes,
		PvcPercent:       usageInfo.Percent,
	}, nil
}

func (m *MonitoringService) collectPodActiveNodes(
	ctx context.Context,
	application *applicationModel.Application,
	config *applicationMetricModel.Configuration,
) (applicationMetricValueModel.MetricValue, error) {
	// Get detailed node information
	nodesInfo, err := m.k8sClient.GetNodesInfoForPods(ctx, application.Namespace, config.PodLabelSelector)
	if err != nil {
		return applicationMetricValueModel.MetricValue{}, err
	}

	// Extract node names for backward compatibility
	nodeNames := make([]string, 0, len(nodesInfo))

	// Convert k8s.NodeInfo to model.NodeInfo
	nodes := make([]applicationMetricValueModel.NodeInfo, 0, len(nodesInfo))
	for _, nodeInfo := range nodesInfo {
		nodeNames = append(nodeNames, nodeInfo.Name)

		// Convert conditions
		conditions := make([]applicationMetricValueModel.NodeCondition, 0, len(nodeInfo.Conditions))
		for _, cond := range nodeInfo.Conditions {
			conditions = append(conditions, applicationMetricValueModel.NodeCondition{
				Type:    cond.Type,
				Status:  cond.Status,
				Reason:  cond.Reason,
				Message: cond.Message,
			})
		}

		nodes = append(nodes, applicationMetricValueModel.NodeInfo{
			Name:       nodeInfo.Name,
			Ready:      nodeInfo.Ready,
			Status:     nodeInfo.Status,
			Conditions: conditions,
			Labels:     nodeInfo.Labels,
			PodCount:   nodeInfo.PodCount,
		})
	}

	return applicationMetricValueModel.MetricValue{
		ActiveNodesCount: len(nodes),
		NodeNames:        nodeNames,
		Nodes:            nodes,
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

// Connection metric collection methods

func (m *MonitoringService) collectRedisConnection(
	ctx context.Context,
	config *applicationMetricModel.Configuration,
) applicationMetricValueModel.MetricValue {
	return connections.TestRedisConnection(ctx, config)
}

func (m *MonitoringService) collectPostgreSQLConnection(
	ctx context.Context,
	config *applicationMetricModel.Configuration,
) applicationMetricValueModel.MetricValue {
	return connections.TestPostgreSQLConnection(ctx, config)
}

func (m *MonitoringService) collectMongoDBConnection(
	ctx context.Context,
	config *applicationMetricModel.Configuration,
) applicationMetricValueModel.MetricValue {
	return connections.TestMongoDBConnection(ctx, config)
}

func (m *MonitoringService) collectMySQLConnection(
	ctx context.Context,
	config *applicationMetricModel.Configuration,
) applicationMetricValueModel.MetricValue {
	return connections.TestMySQLConnection(ctx, config)
}

func (m *MonitoringService) collectKongConnection(
	ctx context.Context,
	config *applicationMetricModel.Configuration,
) applicationMetricValueModel.MetricValue {
	return connections.TestKongConnection(ctx, config)
}

// cleanupOldMetrics removes metric values older than the configured retention period
func (m *MonitoringService) cleanupOldMetrics() {
	ctx := context.Background()
	retentionDays := env.METRICS_RETENTION_DAYS
	if retentionDays <= 0 {
		retentionDays = 30 // Default to 30 days if not configured
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	log.Info(ctx).
		Int("retention_days", retentionDays).
		Time("cutoff_date", cutoffDate).
		Msg("Starting metrics cleanup")

	// Delete old metric values
	query := `DELETE FROM application_metric_values WHERE created_at < $1`
	result, err := m.db.ExecContext(ctx, query, cutoffDate)
	if err != nil {
		log.Error(ctx, err).Msg("Failed to cleanup old metrics")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	log.Info(ctx).
		Int64("deleted_records", rowsAffected).
		Int("retention_days", retentionDays).
		Msg("Metrics cleanup completed")
}
