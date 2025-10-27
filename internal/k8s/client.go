package k8s

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Client struct {
	clientset        *kubernetes.Clientset
	metricsClientset *metricsclientset.Clientset
}

// NewClient creates a Kubernetes client
// It tries to use the kubeconfig from KUBECONFIG env var or ~/.kube/config for local development
// Falls back to in-cluster config if neither is available (production mode)
func NewClient() (*Client, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	metricsClientset, err := metricsclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics clientset: %w", err)
	}

	return &Client{
		clientset:        clientset,
		metricsClientset: metricsClientset,
	}, nil
}

// getKubeConfig returns the kubernetes config
// Priority:
// 1. KUBECONFIG environment variable
// 2. ~/.kube/config
// 3. In-cluster config
func getKubeConfig() (*rest.Config, error) {
	// Try KUBECONFIG environment variable
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath != "" {
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err == nil {
			return config, nil
		}
		// Log but continue to next option
		fmt.Printf("Warning: Failed to load kubeconfig from KUBECONFIG env var (%s): %v\n", kubeconfigPath, err)
	}

	// Try default kubeconfig location
	if home, err := os.UserHomeDir(); err == nil {
		kubeconfigPath = filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(kubeconfigPath); err == nil {
			config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
			if err == nil {
				fmt.Printf("Using kubeconfig from: %s\n", kubeconfigPath)
				return config, nil
			}
			fmt.Printf("Warning: Failed to load kubeconfig from %s: %v\n", kubeconfigPath, err)
		}
	}

	// Fall back to in-cluster config
	fmt.Println("Using in-cluster kubernetes config")
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	return config, nil
}

// NewInClusterClient creates a Kubernetes client that works from within a pod
// Deprecated: Use NewClient() instead, which auto-detects the environment
func NewInClusterClient() (*Client, error) {
	return NewClient()
}

// GetPodsByLabelSelector returns pods matching the label selector in a namespace
func (c *Client) GetPodsByLabelSelector(ctx context.Context, namespace, labelSelector string) (*corev1.PodList, error) {
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}
	return pods, nil
}

// GetPodMetrics returns metrics for a specific pod
func (c *Client) GetPodMetrics(ctx context.Context, namespace, podName string) (*metricsv1beta1.PodMetrics, error) {
	metrics, err := c.metricsClientset.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %w", err)
	}
	return metrics, nil
}

// GetPodStatus returns the status information for a pod
func (c *Client) GetPodStatus(ctx context.Context, namespace, podName string) (*corev1.PodStatus, error) {
	pod, err := c.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}
	return &pod.Status, nil
}

// GetNodesForPods returns unique nodes where pods are running
func (c *Client) GetNodesForPods(ctx context.Context, namespace, labelSelector string) ([]string, error) {
	pods, err := c.GetPodsByLabelSelector(ctx, namespace, labelSelector)
	if err != nil {
		return nil, err
	}

	nodeMap := make(map[string]bool)
	for _, pod := range pods.Items {
		if pod.Spec.NodeName != "" {
			nodeMap[pod.Spec.NodeName] = true
		}
	}

	nodes := make([]string, 0, len(nodeMap))
	for node := range nodeMap {
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// GetPVCUsage returns PVC usage information
func (c *Client) GetPVCUsage(ctx context.Context, namespace, pvcName string) (*corev1.PersistentVolumeClaim, error) {
	pvc, err := c.clientset.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvcName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get PVC: %w", err)
	}
	return pvc, nil
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Status         string
	StatusCode     int
	ResponseTimeMs int64
	ErrorMessage   string
}

// PerformHealthCheck performs an HTTP health check on a URL
func (c *Client) PerformHealthCheck(ctx context.Context, url, method string, expectedStatus, timeoutSeconds int) HealthCheckResult {
	result := HealthCheckResult{
		Status: "down",
	}

	if timeoutSeconds <= 0 {
		timeoutSeconds = 10
	}

	if method == "" {
		method = "GET"
	}

	client := &http.Client{
		Timeout: time.Duration(timeoutSeconds) * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to create request: %v", err)
		return result
	}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	result.ResponseTimeMs = elapsed.Milliseconds()

	if err != nil {
		result.ErrorMessage = fmt.Sprintf("request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	// Read and discard body to ensure connection reuse
	_, _ = io.Copy(io.Discard, resp.Body)

	result.StatusCode = resp.StatusCode

	if expectedStatus == 0 {
		expectedStatus = 200
	}

	if resp.StatusCode == expectedStatus {
		result.Status = "up"
	} else {
		result.ErrorMessage = fmt.Sprintf("unexpected status code: got %d, expected %d", resp.StatusCode, expectedStatus)
	}

	return result
}
