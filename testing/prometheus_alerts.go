package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
)

// PrometheusAlertsClient provides access to Prometheus alerts API
type PrometheusAlertsClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewPrometheusAlertsClient creates a new Prometheus alerts client
func NewPrometheusAlertsClient(promClient api.Client) *PrometheusAlertsClient {
	config := promClient.URL("", nil)
	return &PrometheusAlertsClient{
		baseURL: strings.TrimSuffix(config.String(), "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// PrometheusAlert represents an alert from Prometheus alerts API
type PrometheusAlert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	State       string            `json:"state"`       // inactive, pending, firing
	ActiveAt    time.Time         `json:"activeAt"`
	Value       string            `json:"value"`
}

// PrometheusAlertsResponse represents the response from Prometheus alerts API
type PrometheusAlertsResponse struct {
	Status string `json:"status"`
	Data   struct {
		Alerts []PrometheusAlert `json:"alerts"`
	} `json:"data"`
}

// GetAlerts retrieves all alerts from Prometheus
func (c *PrometheusAlertsClient) GetAlerts(ctx context.Context) ([]PrometheusAlert, error) {
	url := fmt.Sprintf("%s/api/v1/alerts", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Prometheus API error: %d - %s", resp.StatusCode, string(body))
	}

	var response PrometheusAlertsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if response.Status != "success" {
		return nil, fmt.Errorf("Prometheus API returned status: %s", response.Status)
	}

	return response.Data.Alerts, nil
}

// GetAlertsForService retrieves alerts related to a specific service
func (c *PrometheusAlertsClient) GetAlertsForService(ctx context.Context, serviceName string) ([]PrometheusAlert, error) {
	allAlerts, err := c.GetAlerts(ctx)
	if err != nil {
		return nil, err
	}

	var serviceAlerts []PrometheusAlert
	for _, alert := range allAlerts {
		// Check if alert is related to the service
		if service, exists := alert.Labels["service"]; exists && service == serviceName {
			serviceAlerts = append(serviceAlerts, alert)
		}
		// Also check SLO name for synthetic alerts
		if slo, exists := alert.Labels["slo"]; exists && strings.Contains(slo, serviceName) {
			serviceAlerts = append(serviceAlerts, alert)
		}
		// Check alertname for synthetic alerts
		if alertname, exists := alert.Labels["alertname"]; exists && strings.Contains(strings.ToLower(alertname), strings.ToLower(serviceName)) {
			serviceAlerts = append(serviceAlerts, alert)
		}
	}

	return serviceAlerts, nil
}

// GetAlertsByState retrieves alerts filtered by state (inactive, pending, firing)
func (c *PrometheusAlertsClient) GetAlertsByState(ctx context.Context, state string) ([]PrometheusAlert, error) {
	allAlerts, err := c.GetAlerts(ctx)
	if err != nil {
		return nil, err
	}

	var filteredAlerts []PrometheusAlert
	for _, alert := range allAlerts {
		if alert.State == state {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}

	return filteredAlerts, nil
}

// PrintAlertSummary prints a summary of alerts by state
func (c *PrometheusAlertsClient) PrintAlertSummary(ctx context.Context, serviceName string) error {
	alerts, err := c.GetAlertsForService(ctx, serviceName)
	if err != nil {
		return err
	}

	if len(alerts) == 0 {
		fmt.Printf("No alerts found for service: %s\n", serviceName)
		return nil
	}

	// Count by state
	stateCount := make(map[string]int)
	for _, alert := range alerts {
		stateCount[alert.State]++
	}

	fmt.Printf("\n=== ALERT SUMMARY FOR %s ===\n", serviceName)
	for state, count := range stateCount {
		var emoji string
		switch state {
		case "firing":
			emoji = "🔥"
		case "pending":
			emoji = "🟡"
		case "inactive":
			emoji = "⚪"
		default:
			emoji = "❓"
		}
		fmt.Printf("%s %s: %d alerts\n", emoji, strings.ToUpper(state), count)
	}

	fmt.Printf("\nDetailed alerts:\n")
	for _, alert := range alerts {
		var emoji string
		switch alert.State {
		case "firing":
			emoji = "🔥"
		case "pending":
			emoji = "🟡"
		case "inactive":
			emoji = "⚪"
		default:
			emoji = "❓"
		}
		
		alertname := alert.Labels["alertname"]
		slo := alert.Labels["slo"]
		fmt.Printf("  %s %s [%s] - SLO: %s - Active: %v\n", 
			emoji, alertname, alert.State, slo, alert.ActiveAt.Format("15:04:05"))
	}
	fmt.Printf("=====================================\n")

	return nil
}