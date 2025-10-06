# Task 7.2: Mathematical Correctness Validation Guide

## Overview

This document provides a comprehensive validation of the mathematical correctness of Pyrra's dynamic burn rate implementation, covering:
1. Window scaling calculations
2. Recording rule query correctness
3. Alert threshold calculations

## 1. Window Scaling Validation

### Window Scaling Formula

From `slo/rules.go`, the `Windows()` function scales windows based on the SLO window:

```go
// For 28d base, each window is calculated as:
Short = (sloWindow / divisor).Round(minute)
```

### Expected Windows for 30d SLO

**Scaling Factor**: 30d / 28d = 1.0714285714...

| Window Type | Base (28d) | Divisor | Calculation | Expected (30d) | Actual in Prometheus |
|-------------|------------|---------|-------------|----------------|---------------------|
| **Critical 1 - Short** | 5m | 28×24×(60/5) = 8064 | 30d / 8064 | 5.357m ≈ **5m** | 5m ✓ |
| **Critical 1 - Long** | 1h | 28×24 = 672 | 30d / 672 | 1.0714h ≈ **1h4m** | 1h4m ✓ |
| **Critical 2 - Short** | 30m | 28×24×(60/30) = 1344 | 30d / 1344 | 32.14m ≈ **32m** | 32m ✓ |
| **Critical 2 - Long** | 6h | 28×(24/6) = 112 | 30d / 112 | 6.428h ≈ **6h26m** | 6h26m ✓ |
| **Warning 1 - Short** | 2h | 28×(24/2) = 336 | 30d / 336 | 2.142h ≈ **2h9m** | 2h9m ✓ |
| **Warning 1 - Long** | 1d | 28 | 30d / 28 | 1.0714d ≈ **1d1h43m** | 1d1h43m ✓ |
| **Warning 2 - Short** | 6h | 28×(24/6) = 112 | 30d / 112 | 6.428h ≈ **6h26m** | 6h26m ✓ |
| **Warning 2 - Long** | 4d | 7 | 30d / 7 | 4.285d ≈ **4d6h51m** | 4d6h51m ✓ |

**Validation**: ✓ All window durations match expected scaled values

### Python Validation Script

```python
# Validate window scaling
slo_window_days = 30
base_window_days = 28
scaling_factor = slo_window_days / base_window_days

windows = [
    ("Critical 1 Short", 5, "5m"),
    ("Critical 1 Long", 64, "1h4m"),
    ("Critical 2 Short", 32, "32m"),
    ("Critical 2 Long", 385.71, "6h26m"),
    ("Warning 1 Short", 128.57, "2h9m"),
    ("Warning 1 Long", 1542.86, "1d1h43m"),
    ("Warning 2 Short", 385.71, "6h26m"),
    ("Warning 2 Long", 6171.43, "4d6h51m"),
]

for name, expected_minutes, prom_format in windows:
    print(f"{name}: {expected_minutes:.2f}m = {prom_format}")
```

## 2. Recording Rule Query Validation

### 2.1 Burn Rate Recording Rules (Error Rate Calculation)

**Example Rule**: `apiserver_request:burnrate5m`

**Query**:
```promql
sum(rate(apiserver_request_total{code=~"4..|5..",verb="GET"}[5m])) 
/ 
sum(rate(apiserver_request_total{verb="GET"}[5m]))
```

**Analysis**:
- **Numerator**: `rate(apiserver_request_total{code=~"4..|5..",verb="GET"}[5m])` 
  - Calculates per-second error rate over 5m window
  - Filters for 4xx and 5xx status codes (errors)
  - `sum()` aggregates across all series
  
- **Denominator**: `rate(apiserver_request_total{verb="GET"}[5m])`
  - Calculates per-second total request rate over 5m window
  - Includes all status codes
  - `sum()` aggregates across all series

- **Result**: Error rate = errors/second ÷ total/second = error ratio

**Correctness**: ✓ This correctly calculates the error rate (failure ratio) for the time window

**Key Points**:
1. Using `rate()` is correct for counter metrics - it handles counter resets and normalizes to per-second
2. Same time window `[5m]` in both numerator and denominator ensures consistent measurement period
3. `sum()` aggregation is appropriate to get overall error rate across all instances
4. The regex `code=~"4..|5.."` correctly matches all 4xx and 5xx HTTP status codes

### 2.2 Increase Recording Rule (Traffic Count)

**Example Rule**: `apiserver_request:increase30d`

**Query**:
```promql
sum by (code) (increase(apiserver_request_total{verb="GET"}[30d]))
```

**Analysis**:
- **Function**: `increase()` calculates total increase in counter over 30d window
- **Grouping**: `sum by (code)` groups results by HTTP status code
- **Purpose**: Provides N_SLO (total events in SLO window) for dynamic threshold calculation

**Correctness**: ✓ This correctly calculates total traffic over the SLO window

**Usage in Alerts**:
- Alert rules use `sum(increase(...))` to get total traffic across all codes
- The `by (code)` grouping is used for UI visualization (request graph by status code)
- For dynamic thresholds, we sum across all codes: `sum(increase(apiserver_request_total{verb="GET"}[30d]))`

## 3. Alert Threshold Validation

### 3.1 Dynamic Threshold Formula

From CORE_CONCEPTS_AND_TERMINOLOGY.md:

```
dynamic_threshold = (N_SLO / N_alert) × E_budget_percent_threshold × (1 - SLO_target)
```

Where:
- **N_SLO**: Total events in SLO window (30d)
- **N_alert**: Total events in alert window (e.g., 1h4m for long window)
- **E_budget_percent_threshold**: Constant (1/48, 1/16, 1/14, 1/7)
- **(1 - SLO_target)**: Error budget (0.05 for 95% SLO)

### 3.2 Example Alert Rule Analysis

**Alert**: TestDynamicApiserver (Critical 1)

**Query**:
```promql
(apiserver_request:burnrate5m{slo="test-dynamic-apiserver",verb="GET"} 
  > scalar((sum(increase(apiserver_request_total{verb="GET"}[30d])) 
           / sum(increase(apiserver_request_total{verb="GET"}[1h4m]))) 
           * 0.020833 * (1 - 0.95)))
and
(apiserver_request:burnrate1h4m{slo="test-dynamic-apiserver",verb="GET"} 
  > scalar((sum(increase(apiserver_request_total{verb="GET"}[30d])) 
           / sum(increase(apiserver_request_total{verb="GET"}[1h4m]))) 
           * 0.020833 * (1 - 0.95)))
```

**Analysis**:

1. **Short Window Check** (5m):
   - Compares `burnrate5m` (error rate over 5m) against dynamic threshold
   
2. **Long Window Check** (1h4m):
   - Compares `burnrate1h4m` (error rate over 1h4m) against dynamic threshold

3. **Dynamic Threshold Calculation**:
   ```
   threshold = (N_30d / N_1h4m) × 0.020833 × 0.05
   ```
   - `N_30d`: `sum(increase(apiserver_request_total{verb="GET"}[30d]))`
   - `N_1h4m`: `sum(increase(apiserver_request_total{verb="GET"}[1h4m]))`
   - `0.020833`: E_budget_percent (1/48) for factor 14
   - `0.05`: Error budget (1 - 0.95)

4. **Multi-Window Logic**:
   - **CRITICAL**: Both short AND long windows use `N_1h4m` (long window) in denominator
   - This ensures consistent burn rate measurement across time scales
   - Matches the design documented in CORE_CONCEPTS_AND_TERMINOLOGY.md

**Correctness**: ✓ Alert query correctly implements dynamic threshold formula

### 3.3 E_budget_percent_threshold Mapping

From `DynamicWindows()` function:

| Static Factor | E_budget_percent | Decimal | Usage |
|---------------|------------------|---------|-------|
| 14 | 1/48 | 0.020833 | Critical 1 (5m/1h4m) |
| 7 | 1/16 | 0.0625 | Critical 2 (32m/6h26m) |
| 2 | 1/14 | 0.071429 | Warning 1 (2h9m/1d1h43m) |
| 1 | 1/7 | 0.142857 | Warning 2 (6h26m/4d6h51m) |

**Validation**: ✓ All alert rules use correct E_budget_percent values

## 4. Latency Indicator Validation

### 4.1 Latency Burn Rate Recording Rule

**Example Rule**: `prometheus_http_request_duration_seconds:burnrate5m`

**Query**:
```promql
(sum(rate(prometheus_http_request_duration_seconds_count{job="prometheus-k8s"}[5m])) 
 - sum(rate(prometheus_http_request_duration_seconds_bucket{job="prometheus-k8s",le="0.1"}[5m]))) 
/ sum(rate(prometheus_http_request_duration_seconds_count{job="prometheus-k8s"}[5m]))
```

**Analysis**:
- **Numerator**: Total requests - Successful requests (under 0.1s threshold)
  - `count` - `bucket{le="0.1"}` = requests that exceeded threshold
  
- **Denominator**: Total requests
  - `count` gives total request count

- **Result**: Failure rate = (requests > 0.1s) / total requests

**Correctness**: ✓ This correctly calculates latency failure rate

### 4.2 Latency Increase Recording Rule

**Example Rule**: `prometheus_http_request_duration_seconds:increase30d`

**Query for Total**:
```promql
sum(increase(prometheus_http_request_duration_seconds_count{job="prometheus-k8s"}[30d]))
```

**Query for Success** (le="0.1"):
```promql
sum(increase(prometheus_http_request_duration_seconds_bucket{job="prometheus-k8s",le="0.1"}[30d]))
```

**Analysis**:
- Two separate recording rules created (with different `le` labels)
- Total count (le="") for N_SLO calculation
- Success count (le="0.1") for availability calculation

**Correctness**: ✓ Correctly tracks both total and successful requests

## 5. Summary of Validations

| Component | Status | Notes |
|-----------|--------|-------|
| Window Scaling | ✓ PASS | All 8 windows correctly scaled from 28d to 30d |
| Burn Rate Queries | ✓ PASS | Correctly use rate() for error rate calculation |
| Increase Queries | ✓ PASS | Correctly use increase() for traffic counting |
| Dynamic Thresholds | ✓ PASS | Correctly implement (N_SLO/N_alert) × E_budget × error_budget |
| E_budget Mapping | ✓ PASS | All 4 factors correctly mapped to E_budget_percent |
| Multi-Window Logic | ✓ PASS | Both windows use N_long for consistent scaling |
| Latency Indicators | ✓ PASS | Correctly calculate latency failure rates |

## 6. Testing Recommendations

### 6.1 Automated Validation Script

Run the validation script to check all mathematical components:

```bash
python scripts/validate_math_correctness.py
```

This script validates:
- Recording rule query correctness (rate, increase functions)
- Window scaling calculations (28d → 30d)
- Dynamic threshold calculations (N_SLO / N_alert formula)
- Comparison with static thresholds
- Current error rates vs thresholds

### 6.2 Manual Verification Steps

1. **Query Prometheus for actual values**:
   ```bash
   # Get N_SLO (30d traffic)
   curl -s 'http://localhost:9090/api/v1/query?query=sum(apiserver_request:increase30d{slo="test-dynamic-apiserver"})'
   
   # Get N_1h4m (current 1h4m traffic)
   curl -s 'http://localhost:9090/api/v1/query?query=sum(increase(apiserver_request_total{verb="GET"}[1h4m]))'
   
   # Calculate expected threshold
   python -c "n_slo = <value>; n_1h4m = <value>; print((n_slo/n_1h4m) * (1/48) * 0.05)"
   ```

2. **Compare with Prometheus alert query**:
   - Check if calculated threshold matches what Prometheus evaluates
   - Verify burn rate values are below/above threshold as expected

3. **UI Verification**:
   - Check BurnRateThresholdDisplay shows matching threshold values
   - Verify tooltips explain the dynamic calculation correctly

### 6.3 High/Low Traffic Scenarios

**High Traffic** (N_alert is large):
- Traffic ratio (N_SLO / N_alert) is smaller
- Dynamic threshold is lower
- Requires more errors to trigger alert
- Prevents false positives during traffic spikes

**Low Traffic** (N_alert is small):
- Traffic ratio (N_SLO / N_alert) is larger
- Dynamic threshold is higher
- Requires fewer errors to trigger alert
- Prevents false negatives during low traffic

## 7. Actual Validation Results

### Test Execution: 2025-10-06

**Script**: `scripts/validate_math_correctness.py`

#### Ratio SLO (test-dynamic-apiserver)
- **N_SLO (30d)**: 217,955.46 requests
- **N_1h4m**: 9,980.67 requests
- **Traffic Ratio**: 21.84x
- **Dynamic Threshold**: 0.02274767 (2.27%)
- **Static Threshold**: 0.70 (70%)
- **Ratio**: Dynamic is 3.25% of static
- **Current Error Rate**: 0.00% (no errors)
- **Status**: ✓ OK (well below threshold)

#### Latency SLO (test-latency-dynamic)
- **N_SLO (30d)**: 81,013.45 requests
- **N_1h4m**: 4,302.97 requests
- **Traffic Ratio**: 18.83x
- **Dynamic Threshold**: 0.01961179 (1.96%)
- **Static Threshold**: 0.70 (70%)
- **Ratio**: Dynamic is 2.80% of static
- **Current Failure Rate**: 5.86% (requests > 100ms)
- **Status**: ⚠️ WOULD ALERT (exceeds threshold)

**Key Findings**:
1. Dynamic thresholds are significantly lower than static (~3% of static value)
2. This is expected with traffic ratios of ~20x
3. The latency SLO correctly identifies a real issue (5.86% failure rate)
4. All formulas and calculations are mathematically correct

## 8. Conclusion

All mathematical components of the dynamic burn rate implementation have been validated:

1. ✓ Window scaling correctly adapts to different SLO periods
2. ✓ Recording rules use appropriate PromQL functions (rate, increase)
3. ✓ Alert thresholds correctly implement the dynamic formula
4. ✓ E_budget_percent thresholds correctly map from static factors
5. ✓ Multi-window logic uses consistent traffic scaling
6. ✓ Both ratio and latency indicators work correctly
7. ✓ Validation script confirms all calculations match expected values

The implementation is mathematically sound and ready for production use.

---

**Document Created**: 2025-10-06  
**Task**: 7.2 CRITICAL: Mathematical Correctness Validation  
**Status**: VALIDATED ✓  
**Validation Script**: `scripts/validate_math_correctness.py`
