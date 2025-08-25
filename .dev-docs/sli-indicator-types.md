# Service Level Indicator (SLI) Types in Pyrra

## Overview

A Service Level Indicator (SLI) is a quantitative measure that determines the level of service provided. In Pyrra, we support four types of SLIs, each designed to measure different aspects of service performance. Each type requires different metrics and calculations, but ultimately they all produce a ratio of "bad" events to total events for alerting purposes.

## Core Concept: Why Different Types?

While all SLIs eventually boil down to a ratio of "bad events" / "total events" for alerting purposes, different services expose their metrics in different formats. The four indicator types accommodate these different metric structures:

1. **Some services expose error counts directly** → Ratio indicators
2. **Some services expose latency histograms** → Latency indicators  
3. **Some services use native histograms** → LatencyNative indicators
4. **Some services expose boolean success/failure gauges** → BoolGauge indicators

## The Four Indicator Types

### 1. Ratio Indicator (`ratio`)

**Purpose**: Measures the ratio of error events to total events when you have separate error and total metrics.

**Use Cases**:
- HTTP request success rate
- gRPC call success rate  
- API endpoint error rate

**Required Metrics**:
- `errors`: Metric counting error events (e.g., HTTP 5xx responses)
- `total`: Metric counting total events (e.g., all HTTP requests)

**Example Configuration**:
```yaml
indicator:
  ratio:
    errors:
      metric: http_requests_total{code=~"5.."}
    total:
      metric: http_requests_total
```

**Burn Rate Calculation**:
```promql
sum(rate(errors_metric[window])) / sum(rate(total_metric[window]))
```

### 2. Latency Indicator (`latency`) 

**Purpose**: Measures the percentage of requests faster than a latency threshold using traditional Prometheus histograms.

**Use Cases**:
- HTTP response time SLOs
- Database query latency
- API call latency

**Required Metrics**:
- `success`: Histogram bucket metric for requests under the latency threshold (e.g., `le="0.1"`)
- `total`: Histogram count metric for total requests

**Example Configuration**:
```yaml
indicator:
  latency:
    success:
      metric: http_request_duration_seconds_bucket{le="0.1"}
    total:
      metric: http_request_duration_seconds_count
```

**Burn Rate Calculation**:
```promql
(
  sum(rate(total_metric[window]))
  -  
  sum(rate(success_metric[window]))
)
/
sum(rate(total_metric[window]))
```

**Note**: The calculation is `(total - success) / total` to get the error rate (requests slower than threshold).

### 3. LatencyNative Indicator (`latencyNative`)

**Purpose**: Measures latency SLOs using Prometheus native histograms (experimental feature).

**Use Cases**:
- Modern Prometheus setups with native histogram support
- High-cardinality latency measurements
- More precise latency calculations

**Required Metrics**:
- `total`: Native histogram metric
- `latency`: String duration threshold (e.g., "100ms")

**Example Configuration**:
```yaml
indicator:
  latencyNative:
    total:
      metric: http_request_duration_seconds
    latency: "100ms"
```

**Burn Rate Calculation**:
```promql
1 - histogram_fraction(0, threshold_seconds, sum(rate(metric[window])))
```

**Note**: Uses Prometheus `histogram_fraction()` function to calculate the percentage of requests above the latency threshold.

### 4. BoolGauge Indicator (`bool_gauge`)

**Purpose**: Measures SLOs based on boolean gauge metrics that indicate success (1) or failure (0).

**Use Cases**:
- Synthetic monitoring (probe success/failure)
- Batch job success rates
- Service health checks
- Any binary success/failure metrics

**Required Metrics**:
- Single metric that reports 1 for success, 0 for failure

**Example Configuration**:
```yaml
indicator:
  bool_gauge:
    metric: up{job="my-service"}
```

**Burn Rate Calculation**:
```promql
(
  sum(count_over_time(metric[window]))
  -
  sum(sum_over_time(metric[window]))
)
/
sum(count_over_time(metric[window]))
```

**Note**: Uses `count_over_time` for total observations and `sum_over_time` for successful observations (1s).

## Dynamic Burn Rate Support

### Current Implementation Status

- ✅ **Ratio**: Fully implemented with dynamic burn rate support
- ❌ **Latency**: Falls back to static burn rate (TODO)
- ❌ **LatencyNative**: Falls back to static burn rate (TODO)  
- ❌ **BoolGauge**: Falls back to static burn rate (TODO)

### Why Ratio First?

Ratio indicators were implemented first because:
1. They're the most common SLI type
2. They have the simplest metric structure (two separate metrics)
3. The dynamic formula is most straightforward to apply

## Implementation Details

### Metric Selection and Grouping

Each indicator type supports **grouping** to define SLOs for multiple dimensions:

```yaml
indicator:
  ratio:
    # ... metrics ...
    grouping:
      - job
      - handler
```

This creates separate SLO tracking for each combination of grouped labels.

### Label Matching

All indicator types use Prometheus label matchers for metric selection:

```yaml
errors:
  metric: http_requests_total{job="api",code=~"5.."}
```

Supported matcher types:
- `=` (equal)
- `!=` (not equal)
- `=~` (regex match)
- `!~` (regex not match)

### Recording Rules

Pyrra generates recording rules for each indicator type to pre-compute:
- Burn rates over different time windows
- Increase calculations for error budget tracking
- SLO compliance metrics

## Next Steps for Dynamic Burn Rate

### Priority Order for Extension:

1. **Latency Indicators**: Most commonly used after Ratio
2. **BoolGauge Indicators**: Simpler structure, easier to implement
3. **LatencyNative Indicators**: Experimental feature, lower priority

### Implementation Approach:

Each type will need dynamic expressions that:
1. Calculate the ratio `(N_SLO / N_alert)` using the appropriate total metric
2. Apply the error budget percentage threshold
3. Compare the actual error rate to the dynamic threshold

## References

- [Prometheus Histogram Documentation](https://prometheus.io/docs/concepts/metric_types/#histogram)
- [Native Histograms in Prometheus](https://prometheus.io/docs/concepts/metric_types/#native-histograms)
- [Error Budget is All You Need - Dynamic Burn Rate Theory](https://medium.com/@yairstark/error-budget-is-all-you-need-part-1-7f8b6b51eaa6)
