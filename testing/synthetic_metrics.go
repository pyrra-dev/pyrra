package testing

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/api"
	prometheusapiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/model"
)

// SyntheticMetricGenerator provides functionality to generate controlled error conditions
// for testing dynamic burn rate alert firing behavior
type SyntheticMetricGenerator struct {
	promClient  api.Client
	promAPI     prometheusapiv1.API
	pushGateway string
	registry    *prometheus.Registry

	// Metrics for generating synthetic data
	successCounter *prometheus.CounterVec
	errorCounter   *prometheus.CounterVec
	requestCounter *prometheus.CounterVec
	latencyHist    *prometheus.HistogramVec
}

// NewSyntheticMetricGenerator creates a new synthetic metric generator
func NewSyntheticMetricGenerator(promClient api.Client, pushGatewayURL string) *SyntheticMetricGenerator {
	registry := prometheus.NewRegistry()

	successCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "synthetic_test_requests_success_total",
			Help: "Total number of successful synthetic test requests",
		},
		[]string{"service", "method", "code"},
	)

	errorCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "synthetic_test_requests_error_total",
			Help: "Total number of failed synthetic test requests",
		},
		[]string{"service", "method", "code"},
	)

	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "synthetic_test_requests_total",
			Help: "Total number of synthetic test requests",
		},
		[]string{"service", "method"},
	)

	latencyHist := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "synthetic_test_request_duration_seconds",
			Help:    "Histogram of synthetic test request durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)

	registry.MustRegister(successCounter, errorCounter, requestCounter, latencyHist)

	return &SyntheticMetricGenerator{
		promClient:     promClient,
		promAPI:        prometheusapiv1.NewAPI(promClient),
		pushGateway:    pushGatewayURL,
		registry:       registry,
		successCounter: successCounter,
		errorCounter:   errorCounter,
		requestCounter: requestCounter,
		latencyHist:    latencyHist,
	}
}

// TrafficPattern defines a traffic generation pattern
type TrafficPattern struct {
	Service        string
	Method         string
	RequestsPerSec float64
	ErrorRate      float64 // 0.0 to 1.0
	Duration       time.Duration
	LatencyMs      float64 // Average latency in milliseconds
}

// AlertTestScenario defines a complete test scenario for alert firing validation
type AlertTestScenario struct {
	Name           string
	Description    string
	TrafficPattern TrafficPattern
	ExpectedAlert  bool // Whether we expect alerts to fire
	TestDuration   time.Duration
}

// GenerateTrafficPattern generates synthetic metrics according to the specified pattern
func (g *SyntheticMetricGenerator) GenerateTrafficPattern(ctx context.Context, pattern TrafficPattern) error {
	log.Printf("Starting traffic pattern generation: %s/%s at %.2f req/sec with %.2f%% error rate for %v",
		pattern.Service, pattern.Method, pattern.RequestsPerSec, pattern.ErrorRate*100, pattern.Duration)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	endTime := time.Now().Add(pattern.Duration)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(endTime) {
				log.Printf("Traffic pattern completed for %s/%s", pattern.Service, pattern.Method)
				return nil
			}

			// Generate requests for this second
			requestsThisSecond := int(pattern.RequestsPerSec)
			if pattern.RequestsPerSec > float64(requestsThisSecond) {
				// Handle fractional requests per second
				if time.Now().UnixNano()%1000000000 < int64(float64(1000000000)*(pattern.RequestsPerSec-float64(requestsThisSecond))) {
					requestsThisSecond++
				}
			}

			for i := 0; i < requestsThisSecond; i++ {
				g.generateSingleRequest(pattern)
			}

			// Push metrics to gateway
			if err := g.pushMetrics(pattern.Service); err != nil {
				log.Printf("Error pushing metrics: %v", err)
			}
		}
	}
}

// generateSingleRequest generates a single synthetic request
func (g *SyntheticMetricGenerator) generateSingleRequest(pattern TrafficPattern) {
	// Always count total requests
	g.requestCounter.WithLabelValues(pattern.Service, pattern.Method).Inc()

	// Determine if this request should be an error
	isError := time.Now().UnixNano()%1000 < int64(pattern.ErrorRate*1000)

	if isError {
		// Generate error
		errorCode := "500" // Internal server error
		if pattern.ErrorRate > 0.5 {
			// Mix of error codes for high error rates
			codes := []string{"500", "502", "503", "504"}
			errorCode = codes[time.Now().UnixNano()%int64(len(codes))]
		}
		g.errorCounter.WithLabelValues(pattern.Service, pattern.Method, errorCode).Inc()
	} else {
		// Generate success
		g.successCounter.WithLabelValues(pattern.Service, pattern.Method, "200").Inc()
	}

	// Record latency (with some variation)
	latencyVariation := 1.0 + (float64(time.Now().UnixNano()%200)-100)/1000.0 // ±10% variation
	latency := pattern.LatencyMs * latencyVariation / 1000.0                  // Convert to seconds
	g.latencyHist.WithLabelValues(pattern.Service, pattern.Method).Observe(latency)
}

// pushMetrics pushes the current metrics to the push gateway
func (g *SyntheticMetricGenerator) pushMetrics(jobName string) error {
	if g.pushGateway == "" {
		return nil // No push gateway configured
	}

	return push.New(g.pushGateway, jobName).
		Gatherer(g.registry).
		Push()
}

// RunAlertTestScenario runs a complete alert test scenario
func (g *SyntheticMetricGenerator) RunAlertTestScenario(ctx context.Context, scenario AlertTestScenario) (*AlertTestResult, error) {
	log.Printf("Running alert test scenario: %s", scenario.Name)
	log.Printf("Description: %s", scenario.Description)
	log.Printf("Expected alert firing: %v", scenario.ExpectedAlert)

	result := &AlertTestResult{
		ScenarioName:  scenario.Name,
		StartTime:     time.Now(),
		ExpectedAlert: scenario.ExpectedAlert,
	}

	// Start traffic generation
	trafficCtx, cancelTraffic := context.WithCancel(ctx)
	trafficDone := make(chan error, 1)

	go func() {
		trafficDone <- g.GenerateTrafficPattern(trafficCtx, scenario.TrafficPattern)
	}()

	// Monitor for alerts during the test
	alertCtx, cancelAlert := context.WithCancel(ctx)
	alertDone := make(chan error, 1)

	go func() {
		alertDone <- g.monitorAlerts(alertCtx, scenario, result)
	}()

	// Wait for test duration
	testTimer := time.NewTimer(scenario.TestDuration)
	defer testTimer.Stop()

	select {
	case <-ctx.Done():
		cancelTraffic()
		cancelAlert()
		return result, ctx.Err()
	case <-testTimer.C:
		// Test duration completed
		cancelTraffic()

		// Wait for alerts to potentially fire (account for 30s evaluation + for duration)
		time.Sleep(3 * time.Minute)
		cancelAlert()

		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)

		// Wait for goroutines to finish
		<-trafficDone
		<-alertDone

		log.Printf("Alert test scenario completed: %s", scenario.Name)
		return result, nil
	}
}

// AlertTestResult contains the results of an alert test scenario
type AlertTestResult struct {
	ScenarioName  string
	StartTime     time.Time
	EndTime       time.Time
	Duration      time.Duration
	ExpectedAlert bool
	AlertsFired   []AlertEvent
	Success       bool
	ErrorMessage  string
}

// AlertEvent represents an alert firing event
type AlertEvent struct {
	AlertName string
	Labels    map[string]string
	Timestamp time.Time
	State     string // firing, pending, resolved
}

// monitorAlerts monitors for alert firing during the test scenario
// This function detects both "pending" and "active" alert states
func (g *SyntheticMetricGenerator) monitorAlerts(ctx context.Context, scenario AlertTestScenario, result *AlertTestResult) error {
	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds (less than 30s eval interval)
	defer ticker.Stop()

	// Track first pending detection for timing analysis
	var firstPendingTime *time.Time
	var firstActiveTime *time.Time

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			alerts, err := g.queryAlerts(scenario.TrafficPattern.Service)
			if err != nil {
				log.Printf("Error querying alerts: %v", err)
				continue
			}

			for _, alert := range alerts {
				// Check if this is a new alert or state change
				isNew := true
				for _, existing := range result.AlertsFired {
					if existing.AlertName == alert.AlertName &&
						existing.State == alert.State {
						isNew = false
						break
					}
				}

				if isNew {
					result.AlertsFired = append(result.AlertsFired, alert)

					// Track timing for different states
					if alert.State == "pending" && firstPendingTime == nil {
						now := time.Now()
						firstPendingTime = &now
						log.Printf("🟡 Alert PENDING: %s at %v (first pending detection)",
							alert.AlertName, alert.Timestamp)
					} else if alert.State == "active" && firstActiveTime == nil {
						now := time.Now()
						firstActiveTime = &now
						log.Printf("🔥 Alert ACTIVE: %s at %v (firing confirmed)",
							alert.AlertName, alert.Timestamp)

						// Calculate timing from pending to active
						if firstPendingTime != nil {
							pendingDuration := firstActiveTime.Sub(*firstPendingTime)
							log.Printf("⏱️  Alert timing: %v from pending to active", pendingDuration)
						}
					} else {
						log.Printf("📊 Alert detected: %s at %v (state: %s)",
							alert.AlertName, alert.Timestamp, alert.State)
					}
				}
			}
		}
	}
}

// queryAlerts queries Prometheus for active alerts related to the test service
func (g *SyntheticMetricGenerator) queryAlerts(serviceName string) ([]AlertEvent, error) {
	// Query for alerts related to our synthetic service
	query := fmt.Sprintf(`ALERTS{service="%s"}`, serviceName)

	result, warnings, err := g.promAPI.Query(context.Background(), query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("error querying alerts: %w", err)
	}

	if len(warnings) > 0 {
		log.Printf("Query warnings: %v", warnings)
	}

	var alerts []AlertEvent

	switch result.Type() {
	case model.ValVector:
		vector := result.(model.Vector)
		for _, sample := range vector {
			alert := AlertEvent{
				AlertName: string(sample.Metric["alertname"]),
				Labels:    make(map[string]string),
				Timestamp: sample.Timestamp.Time(),
				State:     string(sample.Metric["alertstate"]),
			}

			// Copy all labels
			for name, value := range sample.Metric {
				alert.Labels[string(name)] = string(value)
			}

			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

// ValidateAlertBehavior validates that the alert behavior matches expectations
// Considers both "pending" and "active" states as alert detection
func (r *AlertTestResult) ValidateAlertBehavior() {
	// Count alerts in both pending and active states
	alertsDetected := len(r.AlertsFired) > 0

	// Count different alert states for detailed reporting
	pendingCount := 0
	activeCount := 0
	for _, alert := range r.AlertsFired {
		switch alert.State {
		case "pending":
			pendingCount++
		case "active":
			activeCount++
		}
	}

	if r.ExpectedAlert && alertsDetected {
		r.Success = true
		log.Printf("✅ Test PASSED: Expected alerts and alerts detected (%d pending, %d active)",
			pendingCount, activeCount)
	} else if !r.ExpectedAlert && !alertsDetected {
		r.Success = true
		log.Printf("✅ Test PASSED: Expected no alerts and no alerts detected")
	} else if r.ExpectedAlert && !alertsDetected {
		r.Success = false
		r.ErrorMessage = "Expected alerts to fire but none were detected (neither pending nor active)"
		log.Printf("❌ Test FAILED: Expected alerts but none detected")
	} else {
		r.Success = false
		r.ErrorMessage = fmt.Sprintf("Expected no alerts but %d alerts detected (%d pending, %d active)",
			len(r.AlertsFired), pendingCount, activeCount)
		log.Printf("❌ Test FAILED: Expected no alerts but %d detected (%d pending, %d active)",
			len(r.AlertsFired), pendingCount, activeCount)
	}
}

// Cleanup removes synthetic metrics and resets counters
func (g *SyntheticMetricGenerator) Cleanup() error {
	log.Printf("Cleaning up synthetic metrics...")

	// Reset all metrics
	g.successCounter.Reset()
	g.errorCounter.Reset()
	g.requestCounter.Reset()
	g.latencyHist.Reset()

	// If using push gateway, we could delete the job
	// This would require additional push gateway API calls

	return nil
}

// PredefinedScenarios returns a set of predefined test scenarios for common cases
func PredefinedScenarios() []AlertTestScenario {
	return []AlertTestScenario{
		{
			Name:        "HighErrorRate_ShouldAlert",
			Description: "High error rate that should trigger dynamic burn rate alerts",
			TrafficPattern: TrafficPattern{
				Service:        "synthetic-test-high-error",
				Method:         "GET",
				RequestsPerSec: 10.0,
				ErrorRate:      0.15, // 15% error rate - should trigger alerts
				Duration:       5 * time.Minute,
				LatencyMs:      100,
			},
			ExpectedAlert: true,
			TestDuration:  8 * time.Minute, // Allow time for 30s evaluation + for duration
		},
		{
			Name:        "LowErrorRate_ShouldNotAlert",
			Description: "Low error rate that should not trigger alerts",
			TrafficPattern: TrafficPattern{
				Service:        "synthetic-test-low-error",
				Method:         "GET",
				RequestsPerSec: 10.0,
				ErrorRate:      0.01, // 1% error rate - should not trigger alerts
				Duration:       5 * time.Minute,
				LatencyMs:      100,
			},
			ExpectedAlert: false,
			TestDuration:  8 * time.Minute, // Allow time to confirm no false alerts
		},
		{
			Name:        "HighTraffic_HighErrorRate",
			Description: "High traffic with high error rate - dynamic thresholds should adapt",
			TrafficPattern: TrafficPattern{
				Service:        "synthetic-test-high-traffic",
				Method:         "GET",
				RequestsPerSec: 100.0, // 10x higher traffic
				ErrorRate:      0.08,  // 8% error rate
				Duration:       3 * time.Minute,
				LatencyMs:      50,
			},
			ExpectedAlert: true,
			TestDuration:  6 * time.Minute, // Allow time for 30s evaluation + for duration
		},
		{
			Name:        "LowTraffic_MediumErrorRate",
			Description: "Low traffic with medium error rate - should be more sensitive",
			TrafficPattern: TrafficPattern{
				Service:        "synthetic-test-low-traffic",
				Method:         "GET",
				RequestsPerSec: 1.0,  // Very low traffic
				ErrorRate:      0.05, // 5% error rate
				Duration:       10 * time.Minute,
				LatencyMs:      200,
			},
			ExpectedAlert: true,
			TestDuration:  15 * time.Minute, // Allow time for low traffic scenario
		},
	}
}
