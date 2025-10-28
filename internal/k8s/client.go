package k8s

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Client struct {
	clientset        *kubernetes.Clientset
	metricsClientset *metricsclientset.Clientset
	config           *rest.Config
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
		config:           config,
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

// NodeInfo contains detailed node status information
type NodeInfo struct {
	Name       string
	Ready      bool
	Status     string
	Conditions []NodeCondition
	Labels     map[string]string
	PodCount   int
}

// NodeCondition represents a node condition
type NodeCondition struct {
	Type    string
	Status  string
	Reason  string
	Message string
}

// GetNodesInfoForPods returns detailed information about nodes where pods are running
func (c *Client) GetNodesInfoForPods(ctx context.Context, namespace, labelSelector string) ([]NodeInfo, error) {
	// Get pods to find which nodes they're on
	pods, err := c.GetPodsByLabelSelector(ctx, namespace, labelSelector)
	if err != nil {
		return nil, err
	}

	// Count pods per node
	nodePodCount := make(map[string]int)
	nodeNames := make(map[string]bool)
	for _, pod := range pods.Items {
		if pod.Spec.NodeName != "" {
			nodeNames[pod.Spec.NodeName] = true
			nodePodCount[pod.Spec.NodeName]++
		}
	}

	// Get detailed info for each node
	nodesInfo := make([]NodeInfo, 0, len(nodeNames))
	for nodeName := range nodeNames {
		node, err := c.clientset.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
		if err != nil {
			// If we can't get node info, still add it but mark as unknown
			nodesInfo = append(nodesInfo, NodeInfo{
				Name:     nodeName,
				Ready:    false,
				Status:   "Unknown",
				PodCount: nodePodCount[nodeName],
			})
			continue
		}

		// Determine node ready status
		ready := false
		status := "Unknown"
		var conditions []NodeCondition

		for _, condition := range node.Status.Conditions {
			conditions = append(conditions, NodeCondition{
				Type:    string(condition.Type),
				Status:  string(condition.Status),
				Reason:  condition.Reason,
				Message: condition.Message,
			})

			if condition.Type == corev1.NodeReady {
				if condition.Status == corev1.ConditionTrue {
					ready = true
					status = "Ready"
				} else if condition.Status == corev1.ConditionFalse {
					status = "NotReady"
				} else {
					status = "Unknown"
				}
			}
		}

		// Check for other problematic conditions
		if ready {
			for _, condition := range node.Status.Conditions {
				if condition.Type != corev1.NodeReady && condition.Status == corev1.ConditionTrue {
					// Node has issues (MemoryPressure, DiskPressure, etc.)
					status = fmt.Sprintf("Ready (with %s)", condition.Type)
					break
				}
			}
		}

		nodesInfo = append(nodesInfo, NodeInfo{
			Name:       nodeName,
			Ready:      ready,
			Status:     status,
			Conditions: conditions,
			Labels:     node.Labels,
			PodCount:   nodePodCount[nodeName],
		})
	}

	return nodesInfo, nil
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

// PVCUsageInfo contains PVC usage information
type PVCUsageInfo struct {
	CapacityBytes  int64
	UsedBytes      int64
	AvailableBytes int64
	Percent        float64
}

// GetPVCUsageWithDiskInfo returns detailed PVC usage by executing df in a pod
func (c *Client) GetPVCUsageWithDiskInfo(ctx context.Context, namespace, pvcName, podLabelSelector, containerName, mountPath string) (*PVCUsageInfo, error) {
	// Get the PVC to retrieve capacity
	pvc, err := c.GetPVCUsage(ctx, namespace, pvcName)
	if err != nil {
		return nil, err
	}

	capacity := pvc.Status.Capacity[corev1.ResourceStorage]
	capacityBytes := capacity.Value()

	// If no mount path provided, try to discover it
	if mountPath == "" {
		discoveredPath, err := c.DiscoverPVCMountPath(ctx, namespace, pvcName, podLabelSelector)
		if err != nil {
			return nil, fmt.Errorf("failed to discover mount path: %w", err)
		}
		mountPath = discoveredPath
	}

	// Get a pod that uses this PVC
	pods, err := c.GetPodsByLabelSelector(ctx, namespace, podLabelSelector)
	if err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found with label selector: %s", podLabelSelector)
	}

	// Find a running pod
	var targetPod *corev1.Pod
	for i := range pods.Items {
		if pods.Items[i].Status.Phase == corev1.PodRunning {
			targetPod = &pods.Items[i]
			break
		}
	}

	if targetPod == nil {
		return nil, fmt.Errorf("no running pods found")
	}

	// If no container specified, use the first one
	if containerName == "" && len(targetPod.Spec.Containers) > 0 {
		containerName = targetPod.Spec.Containers[0].Name
	}

	// Execute df command to get disk usage
	// Using df -B1 to get output in bytes for accurate parsing
	command := []string{"df", "-B1", mountPath}
	output, err := c.ExecCommandInPod(ctx, namespace, targetPod.Name, containerName, command)
	if err != nil {
		return nil, fmt.Errorf("failed to execute df command: %w", err)
	}

	// Parse df output
	usedBytes, availableBytes, err := parseDfOutput(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse df output: %w", err)
	}

	var percent float64
	if capacityBytes > 0 {
		percent = float64(usedBytes) / float64(capacityBytes) * 100
	}

	return &PVCUsageInfo{
		CapacityBytes:  capacityBytes,
		UsedBytes:      usedBytes,
		AvailableBytes: availableBytes,
		Percent:        percent,
	}, nil
}

// DiscoverPVCMountPath discovers the mount path of a PVC in a pod
func (c *Client) DiscoverPVCMountPath(ctx context.Context, namespace, pvcName, podLabelSelector string) (string, error) {
	pods, err := c.GetPodsByLabelSelector(ctx, namespace, podLabelSelector)
	if err != nil {
		return "", err
	}

	for _, pod := range pods.Items {
		// Find volume that references the PVC
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == pvcName {
				// Found the volume, now find its mount path
				for _, container := range pod.Spec.Containers {
					for _, volumeMount := range container.VolumeMounts {
						if volumeMount.Name == volume.Name {
							return volumeMount.MountPath, nil
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("could not find mount path for PVC %s in pods with selector %s", pvcName, podLabelSelector)
}

// ExecCommandInPod executes a command in a pod and returns the output
func (c *Client) ExecCommandInPod(ctx context.Context, namespace, podName, containerName string, command []string) (string, error) {
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   command,
			Stdout:    true,
			Stderr:    true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("failed to create executor: %w", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		return "", fmt.Errorf("failed to execute command: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.String(), nil
}

// parseDfOutput parses the output of df command
// Expected format:
// Filesystem     1B-blocks      Used Available Use% Mounted on
// /dev/sda1   10737418240 5368709120 5368709120  50% /data
func parseDfOutput(output string) (usedBytes int64, availableBytes int64, err error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return 0, 0, fmt.Errorf("invalid df output: expected at least 2 lines, got %d", len(lines))
	}

	// Parse the data line (second line)
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return 0, 0, fmt.Errorf("invalid df output format: expected at least 4 fields, got %d", len(fields))
	}

	// Fields: [Filesystem, Blocks, Used, Available, Use%, Mounted]
	used, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse used bytes: %w", err)
	}

	available, err := strconv.ParseInt(fields[3], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse available bytes: %w", err)
	}

	return used, available, nil
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

// IngressCertificateInfo contains information about an Ingress TLS certificate
type IngressCertificateInfo struct {
	Status       string    // "valid", "expiring_soon", "expired", "not_found", "error"
	Expiration   time.Time // Certificate expiration date
	DaysToExpire int       // Days until expiration (negative if expired)
	Issuer       string    // Certificate issuer
	Subject      string    // Certificate subject/CN
	Domains      []string  // DNS names in certificate
	ErrorMessage string    // Error message if any
}

// GetIngressCertificateInfo retrieves certificate information from an Ingress resource
func (c *Client) GetIngressCertificateInfo(ctx context.Context, namespace, ingressName, tlsSecretName string, warningDays int) (*IngressCertificateInfo, error) {
	if warningDays <= 0 {
		warningDays = 30 // Default warning threshold
	}

	// Get the Ingress resource
	ingress, err := c.clientset.NetworkingV1().Ingresses(namespace).Get(ctx, ingressName, metav1.GetOptions{})
	if err != nil {
		return &IngressCertificateInfo{
			Status:       "not_found",
			ErrorMessage: fmt.Sprintf("ingress not found: %v", err),
		}, nil
	}

	// If TLS secret name not provided, get it from Ingress
	if tlsSecretName == "" {
		if len(ingress.Spec.TLS) == 0 {
			return &IngressCertificateInfo{
				Status:       "not_found",
				ErrorMessage: "no TLS configuration found in ingress",
			}, nil
		}
		tlsSecretName = ingress.Spec.TLS[0].SecretName
	}

	// Get the TLS secret
	secret, err := c.clientset.CoreV1().Secrets(namespace).Get(ctx, tlsSecretName, metav1.GetOptions{})
	if err != nil {
		return &IngressCertificateInfo{
			Status:       "not_found",
			ErrorMessage: fmt.Sprintf("TLS secret not found: %v", err),
		}, nil
	}

	// Get the certificate from the secret
	certData, ok := secret.Data["tls.crt"]
	if !ok {
		return &IngressCertificateInfo{
			Status:       "error",
			ErrorMessage: "tls.crt not found in secret",
		}, nil
	}

	// Parse the certificate
	certInfo, err := parseCertificate(certData, warningDays)
	if err != nil {
		return &IngressCertificateInfo{
			Status:       "error",
			ErrorMessage: fmt.Sprintf("failed to parse certificate: %v", err),
		}, nil
	}

	// Enrich with domains from Ingress if not present in cert
	if len(certInfo.Domains) == 0 {
		certInfo.Domains = extractDomainsFromIngress(ingress)
	}

	return certInfo, nil
}

// parseCertificate parses a PEM-encoded certificate and extracts relevant information
func parseCertificate(certData []byte, warningDays int) (*IngressCertificateInfo, error) {
	block, _ := pem.Decode(certData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	now := time.Now()
	daysToExpire := int(time.Until(cert.NotAfter).Hours() / 24)

	// Determine status
	status := "valid"
	if now.After(cert.NotAfter) {
		status = "expired"
	} else if daysToExpire <= warningDays {
		status = "expiring_soon"
	}

	// Extract domains (DNS SANs)
	domains := cert.DNSNames
	if len(domains) == 0 && cert.Subject.CommonName != "" {
		domains = []string{cert.Subject.CommonName}
	}

	return &IngressCertificateInfo{
		Status:       status,
		Expiration:   cert.NotAfter,
		DaysToExpire: daysToExpire,
		Issuer:       cert.Issuer.CommonName,
		Subject:      cert.Subject.CommonName,
		Domains:      domains,
	}, nil
}

// extractDomainsFromIngress extracts hostnames from Ingress rules
func extractDomainsFromIngress(ingress *networkingv1.Ingress) []string {
	domains := make([]string, 0)

	// From TLS hosts
	for _, tls := range ingress.Spec.TLS {
		domains = append(domains, tls.Hosts...)
	}

	// From rules
	for _, rule := range ingress.Spec.Rules {
		if rule.Host != "" {
			// Avoid duplicates
			found := false
			for _, d := range domains {
				if d == rule.Host {
					found = true
					break
				}
			}
			if !found {
				domains = append(domains, rule.Host)
			}
		}
	}

	return domains
}
