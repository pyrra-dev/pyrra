package testing

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/api"
	prometheusapiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

// ServiceHealthChecker validates that all required services are running
type ServiceHealthChecker struct {
	PrometheusURL   string
	AlertManagerURL string
	PyrraAPIURL     string
	PyrraBackendURL string
	PushGatewayURL  string
}

// NewServiceHealthChecker creates a new service health checker with default URLs
func NewServiceHealthChecker() *ServiceHealthChecker {
	return &ServiceHealthChecker{
		PrometheusURL:   "http://localhost:9090",
		AlertManagerURL: "http://localhost:9093",
		PyrraAPIURL:     "http://localhost:9099",
		PyrraBackendURL: "http://localhost:9444",
		PushGatewayURL:  "http://172.24.13.124:9091",
	}
}

// HealthCheckResult represents the result of a service health check
type HealthCheckResult struct {
	Service  string
	URL      string
	Healthy  bool
	Error    error
	Latency  time.Duration
	Required bool
}

// CheckAllServices performs health checks on all services
func (h *ServiceHealthChecker) CheckAllServices(ctx context.Context) ([]*HealthCheckResult, error) {
	results := make([]*HealthCheckResult, 0)

	// Check Prometheus (required)
	results = append(results, h.checkPrometheus(ctx))

	// Check AlertManager (required for alert testing)
	results = append(results, h.checkAlertManager(ctx))

	// Check Pyrra API (required)
	results = append(results, h.checkPyrraAPI(ctx))

	// Check Pyrra Backend (required)
	results = append(results, h.checkPyrraBackend(ctx))

	// Check Push Gateway (optional but recommended)
	results = append(results, h.checkPushGateway(ctx))

	return results, nil
}

// checkPrometheus validates Prometheus is accessible and responding
func (h *ServiceHealthChecker) checkPrometheus(ctx context.Context) *HealthCheckResult {
	start := time.Now()
	result := &HealthCheckResult{
		Service:  "Prometheus",
		URL:      h.PrometheusURL,
		Required: true,
	}

	// Create Prometheus client
	client, err := api.NewClient(api.Config{Address: h.PrometheusURL})
	if err != nil {
		result.Error = fmt.Errorf("failed to create Prometheus client: %w", err)
		result.Latency = time.Since(start)
		return result
	}

	// Test API connectivity
	promAPI := prometheusapiv1.NewAPI(client)
	_, err = promAPI.Config(ctx)
	if err != nil {
		result.Error = fmt.Errorf("failed to query Prometheus config: %w", err)
		result.Latency = time.Since(start)
		return result
	}

	result.Healthy = true
	result.Latency = time.Since(start)
	return result
}

// checkAlertManager validates AlertManager is accessible
func (h *ServiceHealthChecker) checkAlertManager(ctx context.Context) *HealthCheckResult {
	start := time.Now()
	result := &HealthCheckResult{
		Service:  "AlertManager",
		URL:      h.AlertManagerURL,
		Required: true,
	}

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Test AlertManager API v2 (v1 is deprecated as of 0.27.0)
	req, err := http.NewRequestWithContext(ctx, "GET", h.AlertManagerURL+"/api/v2/status", nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %w", err)
		result.Latency = time.Since(start)
		return result
	}

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Errorf("failed to connect to AlertManager: %w", err)
		result.Latency = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result.Error = fmt.Errorf("AlertManager returned status %d", resp.StatusCode)
		result.Latency = time.Since(start)
		return result
	}

	result.Healthy = true
	result.Latency = time.Since(start)
	return result
}

// checkPyrraAPI validates Pyrra API service is accessible
func (h *ServiceHealthChecker) checkPyrraAPI(ctx context.Context) *HealthCheckResult {
	start := time.Now()
	result := &HealthCheckResult{
		Service:  "Pyrra API",
		URL:      h.PyrraAPIURL,
		Required: true,
	}

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Test Pyrra API connectivity (try to list objectives)
	req, err := http.NewRequestWithContext(ctx, "POST", h.PyrraAPIURL+"/objectives.v1alpha1.ObjectiveService/List", nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %w", err)
		result.Latency = time.Since(start)
		return result
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Errorf("failed to connect to Pyrra API: %w", err)
		result.Latency = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	// Accept any response that's not a connection error
	result.Healthy = true
	result.Latency = time.Since(start)
	return result
}

// checkPyrraBackend validates Pyrra Kubernetes backend is accessible
func (h *ServiceHealthChecker) checkPyrraBackend(ctx context.Context) *HealthCheckResult {
	start := time.Now()
	result := &HealthCheckResult{
		Service:  "Pyrra Backend",
		URL:      h.PyrraBackendURL,
		Required: true,
	}

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Test Pyrra Backend connectivity
	req, err := http.NewRequestWithContext(ctx, "POST", h.PyrraBackendURL+"/objectives.v1alpha1.ObjectiveBackendService/List", nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %w", err)
		result.Latency = time.Since(start)
		return result
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Errorf("failed to connect to Pyrra Backend: %w", err)
		result.Latency = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	result.Healthy = true
	result.Latency = time.Since(start)
	return result
}

// checkPushGateway validates Push Gateway is accessible
func (h *ServiceHealthChecker) checkPushGateway(ctx context.Context) *HealthCheckResult {
	start := time.Now()
	result := &HealthCheckResult{
		Service:  "Push Gateway",
		URL:      h.PushGatewayURL,
		Required: false, // Optional service
	}

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Test Push Gateway metrics endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", h.PushGatewayURL+"/metrics", nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %w", err)
		result.Latency = time.Since(start)
		return result
	}

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Errorf("failed to connect to Push Gateway: %w", err)
		result.Latency = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result.Error = fmt.Errorf("Push Gateway returned status %d", resp.StatusCode)
		result.Latency = time.Since(start)
		return result
	}

	result.Healthy = true
	result.Latency = time.Since(start)
	return result
}

// PrintHealthCheckResults prints the health check results in a formatted way
func (h *ServiceHealthChecker) PrintHealthCheckResults(results []*HealthCheckResult) {
	fmt.Println("\n=== SERVICE HEALTH CHECK RESULTS ===")

	allHealthy := true
	requiredUnhealthy := false

	for _, result := range results {
		status := "❌ FAIL"
		if result.Healthy {
			status = "✅ OK"
		} else {
			allHealthy = false
			if result.Required {
				requiredUnhealthy = true
			}
		}

		required := ""
		if result.Required {
			required = " (REQUIRED)"
		} else {
			required = " (optional)"
		}

		fmt.Printf("%s %s%s - %s (%v)\n",
			status, result.Service, required, result.URL, result.Latency.Round(time.Millisecond))

		if result.Error != nil {
			fmt.Printf("   Error: %v\n", result.Error)
		}
	}

	fmt.Println()
	if allHealthy {
		fmt.Println("🎉 All services are healthy and ready for testing!")
	} else if requiredUnhealthy {
		fmt.Println("⚠️  CRITICAL: Required services are unhealthy. Testing will likely fail.")
		fmt.Println("   Please ensure all required services are running before proceeding.")
	} else {
		fmt.Println("⚠️  Some optional services are unhealthy, but testing can proceed.")
	}
	fmt.Println()
}

// ValidateRequiredServices checks that all required services are healthy
func (h *ServiceHealthChecker) ValidateRequiredServices(ctx context.Context) error {
	results, err := h.CheckAllServices(ctx)
	if err != nil {
		return fmt.Errorf("failed to perform health checks: %w", err)
	}

	h.PrintHealthCheckResults(results)

	// Check if any required services are unhealthy
	for _, result := range results {
		if result.Required && !result.Healthy {
			return fmt.Errorf("required service %s is unhealthy: %v", result.Service, result.Error)
		}
	}

	return nil
}
