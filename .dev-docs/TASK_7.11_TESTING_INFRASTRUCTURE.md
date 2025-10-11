# Task 7.11: Production Readiness Testing Infrastructure

## Overview

This document consolidates all information about the production readiness testing infrastructure created for the dynamic burn rate feature. It covers tools, documentation, generated test data, and instructions for running tests.

## Task Structure

Production readiness validation is split into three tasks:

- **Task 7.11**: Create testing infrastructure (tools, scripts, documentation) - **THIS TASK**
- **Task 7.11.1**: Run automated performance tests (baseline, 50 SLOs, 100 SLOs)
- **Task 7.12**: Manual testing (browser compatibility, graceful degradation, migration)

## Completed: Testing Infrastructure

### 1. SLO Generator Tool

**File**: `cmd/generate-test-slos/main.go`

**Features**:

- Generates configurable numbers of test SLOs
- Supports dynamic/static ratio configuration
- **Window variation**: 7d, 28d, 30d (rotates based on index % 3)
- **Target variation**: 99%, 99.5%, 99.9%, 95% (rotates based on index % 4)
- **Indicator variation**: ratio, latency (alternates based on index % 2)
- Outputs ready-to-apply Kubernetes YAML files

**Usage**:

```bash
# Build tool
go build -o generate-test-slos cmd/generate-test-slos/main.go

# Generate 50 SLOs (25 dynamic, 25 static)
./generate-test-slos -count=50 -dynamic-ratio=0.5 -namespace=monitoring -output=.dev/generated-slos-50

# Apply to Kubernetes
kubectl apply -f .dev/generated-slos-50/

# Delete when done
kubectl delete -f .dev/generated-slos-50/
```

**Example Generated SLO**:

```yaml
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: test-dynamic-slo-1
  namespace: monitoring
  labels:
    prometheus: k8s
    role: alert-rules
    pyrra.dev/team: platform
spec:
  target: "99"
  window: 7d # ← Varies: 7d, 28d, 30d
  burnRateType: dynamic
  description: "Dynamic burn rate test SLO #1 - ratio indicator, 7d window"
  indicator:
    ratio:
      errors:
        metric: apiserver_request_total{job="apiserver",code=~"5.."}
      total:
        metric: apiserver_request_total{job="apiserver"}
```

### 2. Performance Monitor Tool

**File**: `cmd/monitor-performance/main.go`

**Features**:

- Monitors API response times
- Tracks SLO counts (total, dynamic, static)
- Measures memory usage and goroutine counts
- Tests Prometheus query performance
- Generates JSON metrics for analysis
- Provides real-time console output
- Creates summary reports

**Usage**:

```bash
# Build tool
go build -o monitor-performance cmd/monitor-performance/main.go

# Monitor for 5 minutes with 10-second intervals
./monitor-performance -duration=5m -interval=10s -output=.dev-docs/performance-metrics.json

# View results
cat .dev-docs/performance-metrics.json | jq '.'

# Calculate average API response time
cat .dev-docs/performance-metrics.json | jq '[.[].api_response_time_ms] | add / length'
```

**Output Example**:

```
Performance Monitoring Started
API URL: http://localhost:9099
Prometheus URL: http://localhost:9090
Duration: 5m0s
Interval: 10s

[23:40:54] SLOs: 12 (5 dynamic, 7 static) | API: 93ms | Mem: 45.2MB | Goroutines: 15 | Prom: 26ms
[23:41:04] SLOs: 12 (5 dynamic, 7 static) | API: 87ms | Mem: 45.8MB | Goroutines: 15 | Prom: 29ms

=== Performance Summary ===
Samples: 30
SLO Count: 12 (5 dynamic, 7 static)

API Response Time:
  Average: 89ms
  Min: 75ms
  Max: 120ms

Prometheus Response Time:
  Average: 28ms

Memory Usage:
  Average: 46.5MB
  Min: 45.2MB
  Max: 48.1MB
  Growth: 2.9MB
```

### 3. Automated Test Script

**File**: `scripts/production-readiness-test.sh`

**Features**:

- Service health checks (API, Prometheus, Kubernetes)
- API response time measurement
- SLO count and distribution analysis
- Prometheus query load monitoring
- Recording and alert rules verification
- Generates comprehensive test summary report

**Usage**:

```bash
# Run automated tests
bash scripts/production-readiness-test.sh

# Review results
cat .dev-docs/production-readiness-results/test-summary.txt
```

### 4. Generated Test SLOs

**50 SLO Test Set**: `.dev/generated-slos-50/`

- 25 dynamic SLOs
- 25 static SLOs
- Windows: 7d, 28d, 30d (rotating)
- Targets: 99%, 99.5%, 99.9%, 95% (rotating)
- Indicators: ratio, latency (alternating)

**Ready to use**:

```bash
kubectl apply -f .dev/generated-slos-50/
```

## Quick Start Guide for Task 7.11.1

### Prerequisites

1. **Ensure services are running**:

   ```bash
   # Check Pyrra API
   curl http://localhost:9099

   # Check Pyrra backend
   curl -X POST -H "Content-Type: application/json" -d '{}' \
     http://localhost:9444/objectives.v1alpha1.ObjectiveBackendService/List

   # Check Prometheus
   curl http://localhost:9090/api/v1/query?query=up
   ```

2. **Build tools if not already built**:
   ```bash
   go build -o generate-test-slos cmd/generate-test-slos/main.go
   go build -o monitor-performance cmd/monitor-performance/main.go
   ```

### Step 1: Baseline Performance Test

```bash
# Monitor current environment
./monitor-performance -duration=2m -interval=10s -output=.dev-docs/baseline-current-slos.json

# Run automated test script
bash scripts/production-readiness-test.sh

# Review results
cat .dev-docs/production-readiness-results/test-summary.txt
```

### Step 2: Medium Scale Test (50 SLOs)

```bash
# Apply 50 test SLOs
kubectl apply -f .dev/generated-slos-50/

# Wait for backend to process (60 seconds)
sleep 60

# Monitor performance
./monitor-performance -duration=5m -interval=10s -output=.dev-docs/medium-scale-slos.json

# Run automated tests
bash scripts/production-readiness-test.sh

# Review results
cat .dev-docs/production-readiness-results/test-summary.txt
```

### Step 3: Large Scale Test (100 SLOs)

```bash
# Generate 100 more SLOs
./generate-test-slos -count=100 -dynamic-ratio=0.5 -output=.dev/generated-slos-100

# Apply to Kubernetes
kubectl apply -f .dev/generated-slos-100/

# Wait for backend to process (120 seconds)
sleep 120

# Monitor performance
./monitor-performance -duration=10m -interval=10s -output=.dev-docs/large-scale-slos.json

# Run automated tests
bash scripts/production-readiness-test.sh

# Review results
cat .dev-docs/production-readiness-results/test-summary.txt
```

### Step 4: Create Performance Benchmarks Document

After completing steps 1-3, create `.dev-docs/PRODUCTION_PERFORMANCE_BENCHMARKS.md` with:

```markdown
# Production Performance Benchmarks

## Test Environment

- Pyrra version: [version]
- Kubernetes version: [version]
- Prometheus version: [version]
- Test date: [date]

## Baseline Metrics (Current SLOs)

| Metric                  | Value |
| ----------------------- | ----- |
| Total SLOs              | X     |
| Dynamic SLOs            | Y     |
| Static SLOs             | Z     |
| API Response Time (avg) | Xms   |
| API Response Time (max) | Xms   |
| Memory Usage (avg)      | XMB   |
| Memory Growth           | XMB   |
| Prometheus Query Rate   | X/min |

## Medium Scale Metrics (50 Additional SLOs)

| Metric                  | Value | vs Baseline |
| ----------------------- | ----- | ----------- |
| Total SLOs              | X     | +50         |
| API Response Time (avg) | Xms   | +X%         |
| Memory Usage (avg)      | XMB   | +X%         |
| Memory Growth           | XMB   | +XMB        |

## Large Scale Metrics (100 Additional SLOs)

| Metric                  | Value | vs Baseline |
| ----------------------- | ----- | ----------- |
| Total SLOs              | X     | +100        |
| API Response Time (avg) | Xms   | +X%         |
| Memory Usage (avg)      | XMB   | +X%         |
| Memory Growth           | XMB   | +XMB        |

## Scaling Characteristics

- **API Response Time**: [Linear/Sub-linear/Super-linear] scaling
- **Memory Usage**: [Linear/Sub-linear/Super-linear] scaling
- **Prometheus Query Load**: [Acceptable/High/Critical]

## Recommendations

1. [Recommendation based on results]
2. [Recommendation based on results]
3. [Recommendation based on results]

## Conclusion

[Summary of production readiness from performance perspective]
```

### Cleanup

```bash
# Delete generated test SLOs
kubectl delete -f .dev/generated-slos-50/
kubectl delete -f .dev/generated-slos-100/

# Or delete by label
kubectl delete slo -n monitoring -l pyrra.dev/team=platform
```

## Expected Results

### Baseline (Current SLOs)

- API response time: < 1000ms
- Memory usage: < 50MB
- UI loads smoothly
- No errors

### Medium Scale (50 Additional SLOs)

- API response time: < 2000ms
- Memory usage: < 100MB
- UI remains responsive
- No performance degradation

### Large Scale (100 Additional SLOs)

- API response time: < 3000ms
- Memory usage: < 200MB
- UI still functional
- May identify bottlenecks

## Troubleshooting

### Issue: Services Not Running

```bash
# Restart Pyrra backend
./pyrra kubernetes &

# Restart Pyrra API
./pyrra api &

# Wait for initialization
sleep 30
```

### Issue: API Returns 0 SLOs

This indicates the API service is not connected to the backend. Restart both services.

### Issue: kubectl Commands Fail

```bash
# Check Kubernetes connection
kubectl cluster-info

# Check namespace exists
kubectl get namespace monitoring

# Create namespace if needed
kubectl create namespace monitoring
```

## Success Criteria

### Task 7.11 (Infrastructure Creation)

- ✅ SLO generator tool created with window variation
- ✅ Performance monitor tool created
- ✅ Automated test script created
- ✅ 50 test SLOs generated
- ✅ Documentation consolidated

### Task 7.11.1 (Running Tests)

- ✅ Baseline performance test completed
- ✅ Medium scale test (50 SLOs) completed
- ✅ Large scale test (100 SLOs) completed
- ✅ Performance benchmarks document created
- ✅ Scaling characteristics documented

## References

- **Browser Testing Guide**: `.dev-docs/BROWSER_COMPATIBILITY_TEST_GUIDE.md` (for Task 7.12)
- **Task Definition**: `.kiro/specs/dynamic-burn-rate-completion/tasks.md`
- **Requirements**: `.kiro/specs/dynamic-burn-rate-completion/requirements.md` (Requirement 5)
- **Design**: `.kiro/specs/dynamic-burn-rate-completion/design.md`
