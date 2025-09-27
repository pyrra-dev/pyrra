# Testing Environment Reference - Dynamic Burn Rate Feature

## Overview

This document provides the definitive reference for the testing environment architecture, service endpoints, and configuration details required for running dynamic burn rate alert testing. **All testing code must reference this document** to ensure correct service URLs and timing expectations.

## Service Architecture

### 1. Kubernetes Cluster
- **Type**: Minikube
- **Platform**: Windows 10 with Hyper-V
- **Minikube IP**: `172.24.13.124`
- **Access Method**: kubectl port-forward for most services

### 2. Prometheus
- **Deployment**: Kubernetes (kube-prometheus stack)
- **Namespace**: `monitoring`
- **Service**: `prometheus-k8s`
- **Port Forward**: `kubectl port-forward svc/prometheus-k8s 9090:9090 -n monitoring`
- **UI URL**: `http://localhost:9090/`
- **API URL**: `http://localhost:9090/api/v1/`
- **Evaluation Interval**: **30 seconds** (critical for alert timing calculations)
- **Status**: ✅ Running (requires port-forward)

### 3. AlertManager
- **Deployment**: Kubernetes (kube-prometheus stack)
- **Namespace**: `monitoring`
- **Service**: `alertmanager-main`
- **Port Forward**: `kubectl port-forward svc/alertmanager-main 9093:9093 -n monitoring`
- **UI URL**: `http://localhost:9093/`
- **API URL**: `http://localhost:9093/api/v2/` (v1 API deprecated as of AlertManager 0.27.0)
- **Status**: ⚠️ **Requires manual port-forward before testing**

### 4. Pyrra API Service
- **Deployment**: Local binary
- **Command**: `./pyrra api --api-url=http://localhost:9444 --prometheus-url=http://localhost:9090`
- **Port**: `9099` (embedded UI and API)
- **API URL**: `http://localhost:9099/`
- **API Endpoint**: `/objectives.v1alpha1.ObjectiveService/List`
- **Purpose**: Full-featured API server with embedded UI
- **Status**: ✅ Running (manual binary execution)

### 5. Pyrra Kubernetes Backend
- **Deployment**: Local binary
- **Command**: `./pyrra kubernetes`
- **Port**: `9444`
- **API URL**: `http://localhost:9444/`
- **API Endpoint**: `/objectives.v1alpha1.ObjectiveBackendService/List`
- **Purpose**: Kubernetes operator backend
- **Status**: ✅ Running (manual binary execution)

### 6. Push Gateway
- **Deployment**: Docker container (not in Kubernetes)
- **Network**: Minikube Docker environment
- **URL**: `http://172.24.13.124:9091`
- **Purpose**: Metric pushing for synthetic metric generation
- **Status**: ✅ Running (container)

## Alert Timing Configuration

### Prometheus Evaluation
- **Evaluation Interval**: **30 seconds**
- **Scrape Interval**: 30 seconds (default)
- **Rule Evaluation**: Every 30 seconds

### Alert State Transitions
```
Metric Condition Met → [30s evaluation] → Pending → [for duration] → Firing
```

### Expected Alert Timing
1. **Condition Detection**: 0-30s (next evaluation cycle)
2. **Pending State**: Duration depends on alert rule "for" clause
3. **Firing State**: Immediately after "for" duration expires
4. **Total Latency**: Evaluation interval + "for" duration + processing time

### Alert Rule Windows (Typical Configuration)
- **Short Window**: 5 minutes
- **Long Window**: 1 hour  
- **SLO Period**: 28 days (default)
- **"For" Duration**: Typically 2-5 minutes for burn rate alerts

## Testing Framework Configuration

### Required Service URLs
```go
// Correct URLs for testing framework
const (
    PrometheusURL    = "http://localhost:9090"
    AlertManagerURL  = "http://localhost:9093"
    PyrraAPIURL      = "http://localhost:9099"
    PyrraBackendURL  = "http://localhost:9444"
    PushGatewayURL   = "http://172.24.13.124:9091"
)
```

### Alert Detection Timeouts
```go
// Recommended timeouts based on 30s evaluation interval
const (
    MinAlertLatency     = 30 * time.Second   // One evaluation cycle
    TypicalAlertLatency = 2 * time.Minute    // Evaluation + "for" duration
    MaxAlertLatency     = 5 * time.Minute    // Conservative maximum
    AlertCheckInterval  = 5 * time.Second    // How often to check alert status
)
```

## Service Dependencies and Startup Order

### Required Services for Full Testing
1. **Minikube cluster** (must be running)
2. **Prometheus** (port-forward required)
3. **AlertManager** (port-forward required)
4. **Pyrra Kubernetes backend** (`./pyrra kubernetes`)
5. **Pyrra API service** (`./pyrra api --api-url=http://localhost:9444 --prometheus-url=http://localhost:9090`)
6. **Push Gateway** (Docker container)

### Service Health Checks
```bash
# Verify all services are accessible
curl -s http://localhost:9090/api/v1/status/config > /dev/null && echo "✅ Prometheus OK" || echo "❌ Prometheus FAIL"
curl -s http://localhost:9093/api/v2/status > /dev/null && echo "✅ AlertManager OK" || echo "❌ AlertManager FAIL"
curl -s http://localhost:9099/api/v1/status > /dev/null && echo "✅ Pyrra API OK" || echo "❌ Pyrra API FAIL"
curl -s http://localhost:9444/api/v1/status > /dev/null && echo "✅ Pyrra Backend OK" || echo "❌ Pyrra Backend FAIL"
curl -s http://172.24.13.124:9091/metrics > /dev/null && echo "✅ Push Gateway OK" || echo "❌ Push Gateway FAIL"

# Or use the automated health checker
go run cmd/test-health-check/main.go
```

## SLO Configuration for Testing

### Available Test SLOs
**Note**: Specific SLO names and configurations need to be confirmed with user before testing.

### Synthetic Metric SLO Template
```yaml
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: synthetic-test-slo
  namespace: monitoring
spec:
  target: "99"
  window: 28d
  burnRateType: dynamic  # or static for comparison
  indicator:
    ratio:
      errors:
        metric: synthetic_test_requests_error_total
      total:
        metric: synthetic_test_requests_total
```

## Common Testing Issues and Solutions

### 1. Connection Refused Errors
**Symptom**: `dial tcp [::1]:9093: connectex: No connection could be made`
**Cause**: Service not running or port-forward not active
**Solution**: Verify service is running and port-forward is active

### 2. Alert Not Firing
**Possible Causes**:
- Prometheus not evaluating rules (check rule status)
- Metrics not reaching Prometheus (check targets)
- Alert rule syntax errors (check Prometheus logs)
- Insufficient "for" duration elapsed

### 3. Metric Push Failures
**Symptom**: Push Gateway connection errors
**Cause**: Incorrect Minikube IP or Push Gateway not running
**Solution**: Verify Minikube IP with `minikube ip` and container status

### 4. Timing Issues
**Symptom**: Tests timeout waiting for alerts
**Cause**: Insufficient timeout or unrealistic expectations
**Solution**: Use timeouts >= 5 minutes for alert firing tests

## Pre-Test Checklist

Before running any alert testing:

- [ ] Minikube cluster is running
- [ ] Prometheus port-forward is active (`kubectl port-forward svc/prometheus-k8s 9090:9090 -n monitoring`)
- [ ] AlertManager port-forward is active (`kubectl port-forward svc/alertmanager-main 9093:9093 -n monitoring`)
- [ ] Pyrra Kubernetes backend is running (`./pyrra kubernetes`)
- [ ] Pyrra API service is running (`./pyrra api --api-url=http://localhost:9444 --prometheus-url=http://localhost:9090`)
- [ ] Push Gateway container is running
- [ ] Test SLOs are deployed and configured
- [ ] All service health checks pass

## Testing Framework Updates Required

Based on this reference, the following updates are needed in the testing framework:

1. **Update all hardcoded URLs** to match this configuration
2. **Adjust alert detection timeouts** to account for 30s evaluation interval
3. **Add service health checks** before starting tests
4. **Update Push Gateway URL** to use Minikube IP
5. **Add pre-test validation** to ensure all services are running

---

**Document Status**: ✅ **ACTIVE REFERENCE** - All testing code must use these configurations
**Last Updated**: September 27, 2025
**Next Update**: After any service configuration changes