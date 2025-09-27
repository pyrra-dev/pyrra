# Synthetic Metric Generation for Alert Testing

This package provides comprehensive tools for testing dynamic burn rate alert firing behavior through controlled synthetic metric generation.

## Overview

The synthetic metric generation system allows you to:

1. **Generate Controlled Traffic Patterns**: Create realistic request/error patterns with configurable rates
2. **Test Alert Firing**: Validate that alerts fire correctly when thresholds are exceeded
3. **Precision and Recall Testing**: Ensure alerts don't fire when they shouldn't (precision) and do fire when they should (recall)
4. **Compare Static vs Dynamic**: Validate that dynamic burn rates provide better sensitivity than static thresholds

## Components

### SyntheticMetricGenerator

The core component that generates synthetic metrics matching Prometheus patterns:

- **Success/Error Counters**: `synthetic_test_requests_success_total`, `synthetic_test_requests_error_total`
- **Request Counters**: `synthetic_test_requests_total`
- **Latency Histograms**: `synthetic_test_request_duration_seconds`

### AlertTestRunner

Orchestrates complete test scenarios including:

- Traffic pattern execution
- Alert monitoring
- Result validation
- Precision/recall analysis

### Command Line Interface

The `cmd/alert-test/main.go` provides a complete CLI for running tests:

```bash
# List available scenarios
go run cmd/alert-test/main.go -list-scenarios

# Run all scenarios
go run cmd/alert-test/main.go

# Run specific scenario
go run cmd/alert-test/main.go -scenario="HighErrorRate_ShouldAlert"

# Use custom Prometheus URL
go run cmd/alert-test/main.go -prometheus-url="http://prometheus:9090"
```

## Predefined Test Scenarios

### 1. HighErrorRate_ShouldAlert
- **Purpose**: Validate alerts fire with high error rates
- **Traffic**: 10 req/sec, 15% error rate for 5 minutes
- **Expected**: Alerts should fire
- **Tests**: Recall (catching real issues)

### 2. LowErrorRate_ShouldNotAlert
- **Purpose**: Validate alerts don't fire with acceptable error rates
- **Traffic**: 10 req/sec, 1% error rate for 5 minutes
- **Expected**: No alerts should fire
- **Tests**: Precision (avoiding false positives)

### 3. HighTraffic_HighErrorRate
- **Purpose**: Test dynamic threshold adaptation with high traffic
- **Traffic**: 100 req/sec, 8% error rate for 3 minutes
- **Expected**: Alerts should fire (dynamic thresholds adapt)
- **Tests**: Traffic-aware sensitivity

### 4. LowTraffic_MediumErrorRate
- **Purpose**: Test increased sensitivity with low traffic
- **Traffic**: 1 req/sec, 5% error rate for 10 minutes
- **Expected**: Alerts should fire (higher sensitivity)
- **Tests**: Low-traffic sensitivity

## Validated Test Results

### ✅ Successful End-to-End Validation (September 27, 2025)

The synthetic metric generation framework has been **successfully validated** with real alert firing:

```bash
# Successful test command
go run cmd/run-synthetic-test/main.go -duration=1m -error-rate=0.25

# Results achieved:
# ✅ Synthetic metrics generated (20 req/sec, 25% error rate)
# ✅ Alert state transitions detected:
#    🟡 6 PENDING alerts → 🔥 2 FIRING alerts
# ✅ Both static and dynamic burn rate alerts triggered
# ✅ Proper alert timing analysis (pending → firing transitions)
```

**Key Achievement**: Switched to Prometheus alerts API (`/api/v1/alerts`) which properly detects:
- **inactive** → **pending** → **firing** alert lifecycle
- AlertManager API only shows active alerts, missing pending state

## Usage Examples

### Basic Traffic Generation

```go
// Create generator
promClient, _ := api.NewClient(api.Config{Address: "http://localhost:9090"})
generator := NewSyntheticMetricGenerator(promClient, "")

// Define traffic pattern
pattern := TrafficPattern{
    Service:        "my-test-service",
    Method:         "GET",
    RequestsPerSec: 10.0,
    ErrorRate:      0.1, // 10% error rate
    Duration:       5 * time.Minute,
    LatencyMs:      100,
}

// Generate traffic
ctx := context.Background()
err := generator.GenerateTrafficPattern(ctx, pattern)
```

### Complete Alert Test

```go
// Create test runner
runner := NewAlertTestRunner(promClient, "")

// Run all predefined scenarios
ctx := context.Background()
err := runner.RunAllScenarios(ctx)

// Print results
runner.PrintSummary()
runner.ValidatePrecisionAndRecall()
```

### Custom Scenario

```go
customScenario := AlertTestScenario{
    Name:        "CustomTest",
    Description: "My custom test scenario",
    TrafficPattern: TrafficPattern{
        Service:        "custom-service",
        Method:         "POST",
        RequestsPerSec: 50.0,
        ErrorRate:      0.12,
        Duration:       3 * time.Minute,
        LatencyMs:      200,
    },
    ExpectedAlert: true,
    TestDuration:  4 * time.Minute,
}

runner.AddCustomScenario(customScenario)
```

## Prerequisites

### Required Services

1. **Prometheus**: Running on localhost:9090 (or custom URL)
2. **Pyrra Backend**: Running with dynamic burn rate SLOs configured
3. **AlertManager**: For alert firing validation (optional)
4. **Push Gateway**: For metric pushing (optional)

### SLO Configuration

Ensure you have dynamic burn rate SLOs configured for the test services:

```yaml
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: synthetic-test-slo
spec:
  target: "99"
  window: 30d
  burnRateType: dynamic  # Enable dynamic burn rates
  indicator:
    ratio:
      errors:
        metric: synthetic_test_requests_error_total
      total:
        metric: synthetic_test_requests_total
```

## Metrics Generated

The synthetic metric generator creates metrics that match typical Prometheus patterns:

### Counter Metrics
```promql
# Total requests
synthetic_test_requests_total{service="test-service", method="GET"}

# Successful requests  
synthetic_test_requests_success_total{service="test-service", method="GET", code="200"}

# Failed requests
synthetic_test_requests_error_total{service="test-service", method="GET", code="500"}
```

### Histogram Metrics
```promql
# Request duration histogram
synthetic_test_request_duration_seconds{service="test-service", method="GET"}
synthetic_test_request_duration_seconds_bucket{service="test-service", method="GET", le="0.1"}
synthetic_test_request_duration_seconds_count{service="test-service", method="GET"}
synthetic_test_request_duration_seconds_sum{service="test-service", method="GET"}
```

## Validation Approach

### Precision Testing
- Generate traffic patterns that should NOT trigger alerts
- Verify no false positive alerts are fired
- Measure false positive rate

### Recall Testing  
- Generate traffic patterns that SHOULD trigger alerts
- Verify alerts are fired within expected timeframes
- Measure false negative rate

### Dynamic vs Static Comparison
- Run identical error conditions against both static and dynamic SLOs
- Compare alert sensitivity and timing
- Validate dynamic thresholds provide better precision AND recall

## Integration with CI/CD

The alert test runner can be integrated into CI/CD pipelines:

```bash
#!/bin/bash
# Run alert firing validation tests
go run cmd/alert-test/main.go -timeout=20m

if [ $? -eq 0 ]; then
    echo "✅ All alert firing tests passed"
else
    echo "❌ Alert firing tests failed"
    exit 1
fi
```

## Service Health Checking

The testing framework includes comprehensive service health checking to validate all required services are running before starting tests:

```bash
# Health check is performed automatically by default
go run cmd/precision-recall-test/main.go

# Skip health check if needed
go run cmd/precision-recall-test/main.go -skip-health-check
```

### Required Services
- **Prometheus**: `http://localhost:9090` (requires `kubectl port-forward svc/prometheus-k8s 9090:9090 -n monitoring`)
- **AlertManager**: `http://localhost:9093` (requires `kubectl port-forward svc/alertmanager-main 9093:9093 -n monitoring`)
- **Pyrra API**: `http://localhost:9099` (requires `./pyrra api --api-url=http://localhost:9444 --prometheus-url=http://localhost:9090`)
- **Pyrra Backend**: `http://localhost:9444` (requires `./pyrra kubernetes`)

### Optional Services
- **Push Gateway**: `http://172.24.13.124:9091` (Docker container in Minikube environment)

## Alert Timing Expectations

### Prometheus Configuration
- **Evaluation Interval**: 30 seconds
- **Alert Detection Latency**: 30s (evaluation) + "for" duration + processing time
- **Recommended Test Timeouts**: Minimum 5 minutes for alert firing tests

### Alert State Transitions
```
Condition Met → [30s evaluation] → Pending → [for duration] → Firing
```

## Troubleshooting

### Common Issues

1. **Service Health Check Failures**:
   - Verify all required services are running
   - Check port-forward commands are active for Kubernetes services
   - Confirm Pyrra binaries are running with correct flags

2. **No Alerts Detected**: 
   - Verify SLO configuration includes `burnRateType: dynamic`
   - Check Prometheus rules are generated correctly
   - Ensure AlertManager is configured and accessible
   - Allow sufficient time for 30s evaluation interval + "for" duration

3. **Metrics Not Appearing**:
   - Verify Push Gateway is accessible at correct Minikube IP
   - Check Prometheus is scraping the correct endpoints
   - Confirm metric names match SLO indicator configuration

4. **Test Timeouts**:
   - Use minimum 5-8 minute timeouts for alert firing tests
   - Account for 30s Prometheus evaluation interval
   - Allow extra time for low-traffic scenarios
   - Check Prometheus query performance

### Debug Mode

Enable debug logging by setting log level:

```go
log.SetLevel(log.DebugLevel)
```

This will show detailed information about:
- Metric generation rates
- Prometheus query results
- Alert detection events
- Traffic pattern execution

## Performance Considerations

- **Metric Cardinality**: Each service/method combination creates separate metric series
- **Query Load**: Alert monitoring queries run every 5 seconds during tests
- **Memory Usage**: Metrics are held in memory during test execution
- **Cleanup**: Always call `Cleanup()` to reset metrics between tests

## Precision and Recall Testing

The framework includes comprehensive precision and recall testing capabilities:

### Precision Testing (False Positive Detection)
Tests scenarios that should NOT trigger alerts to validate precision:

```bash
# Run precision tests only
go run cmd/precision-recall-test/main.go -test-type=precision

# Example scenarios:
# - Very low error rates (0.5%)
# - Normal operating conditions (1% error rate)
# - High traffic with low errors (dynamic adaptation test)
# - Short error spikes that resolve quickly
```

### Recall Testing (True Positive Detection)
Tests scenarios that SHOULD trigger alerts to validate recall:

```bash
# Run recall tests only
go run cmd/precision-recall-test/main.go -test-type=recall

# Example scenarios:
# - Sustained high error rates (25%)
# - Moderate errors with low traffic (dynamic sensitivity test)
# - Critical error rates (40%)
# - Gradual error rate increases
```

### Comprehensive Analysis
Run both precision and recall tests with detailed metrics:

```bash
# Run complete precision/recall analysis
go run cmd/precision-recall-test/main.go -test-type=both

# Custom AlertManager and alert names
go run cmd/precision-recall-test/main.go \
  -alertmanager-url="http://alertmanager:9093" \
  -static-alert="ErrorBudgetBurn" \
  -dynamic-alert="DynamicErrorBudgetBurn"
```

### Metrics Calculated
- **Precision**: Accuracy of fired alerts (TP / (TP + FP))
- **Recall**: Coverage of real issues (TP / (TP + FN))
- **Accuracy**: Overall correctness ((TP + TN) / Total)
- **F1 Score**: Harmonic mean of precision and recall
- **Alert Latency**: Average time from issue start to alert firing
- **Improvement Metrics**: Dynamic vs static performance comparison

## Future Enhancements

Potential improvements for the synthetic metric generation system:

1. **Realistic Traffic Patterns**: Add support for traffic spikes, gradual increases, etc.
2. **Multiple Indicator Types**: Support latency and boolean gauge indicators
3. **Distributed Testing**: Run tests across multiple Prometheus instances
4. **Historical Data**: Generate historical metrics for longer-term testing
5. **Custom Alert Rules**: Test custom alert rule configurations
6. **Machine Learning Integration**: Use ML to generate more realistic traffic patterns
7. **Chaos Engineering**: Integration with chaos engineering tools for realistic failure scenarios