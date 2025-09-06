# Session 12: Missing Metrics & Edge Case Validation

## 🎯 Session Objective

**Primary Goal**: Validate robust error handling when metrics are missing, insufficient, or edge cases occur.

**Scope**: Resilience testing for both static and dynamic SLOs
**Current Status**: Only tested with available, well-populated metrics
**Session Priority**: MEDIUM-HIGH (critical for production reliability)

## 📋 Session Context

### **What We Know Works**:
- ✅ Dynamic burn rate with available metrics (`apiserver_request_total`)
- ✅ BurnRateThresholdDisplay component with good data
- ✅ Mathematical calculations with normal traffic patterns

### **What We Need to Validate**:
- ❓ Behavior when base metrics don't exist
- ❓ Handling of metrics with no data/empty results
- ❓ Mathematical edge cases (division by zero, etc.)
- ❓ Consistent error handling between static and dynamic SLOs
- ❓ UI graceful degradation

## 🔧 Session Setup Requirements

### **Environment Prerequisites**:
- Existing Pyrra environment with working dynamic SLOs
- Ability to create SLOs with non-existent metrics
- Access to Pyrra logs and Kubernetes events

### **Test SLO Scenarios**:
We'll create several deliberately problematic SLOs to test error handling:

#### **Scenario A: Completely Non-Existent Metrics**
```yaml
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: test-missing-static
  namespace: monitoring
spec:
  target: 0.95
  window: 30d
  burnRateType: static
  indicator:
    ratio:
      errors:
        metric: completely_fictional_error_metric
      total:
        metric: completely_fictional_total_metric
---
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: test-missing-dynamic
  namespace: monitoring
spec:
  target: 0.95
  window: 30d
  burnRateType: dynamic
  indicator:
    ratio:
      errors:
        metric: completely_fictional_error_metric
      total:
        metric: completely_fictional_total_metric
```

#### **Scenario B: Real Metrics with No Data**
```yaml
# Use real metric names but with selectors that return no data
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: test-nodata-dynamic
  namespace: monitoring
spec:
  target: 0.95
  window: 30d
  burnRateType: dynamic
  indicator:
    ratio:
      errors:
        metric: prometheus_http_requests_total{code="999"}  # Non-existent code
      total:
        metric: prometheus_http_requests_total{job="nonexistent"}  # Non-existent job
```

## 📊 Test Scenarios

### **Test 12.1: Completely Missing Metrics**
**Objective**: Validate behavior when base metrics don't exist at all

**Testing Steps**:
1. **Deploy Test SLOs**: Apply both static and dynamic SLOs with fictional metrics
2. **Check Pyrra Logs**: Look for error handling in Pyrra backend
3. **Check PrometheusRule**: Verify rule generation behavior
4. **API Response**: Check GetStatus API response format
5. **UI Behavior**: Navigate to SLO detail pages

**Expected Behaviors**:
- Pyrra doesn't crash or fail completely
- PrometheusRules generate (even if queries will fail)
- API returns consistent error information
- UI shows meaningful error state (not crashes)

**Validation Checklist**:
- [ ] Pyrra backend logs show appropriate warnings/errors
- [ ] PrometheusRules created without blocking other SLOs
- [ ] API responses include error information or "no data" indicators
- [ ] UI shows consistent error state for both static and dynamic
- [ ] No JavaScript crashes in browser console

### **Test 12.2: Metrics Exist but Return No Data**
**Objective**: Test scenarios where metrics exist but selectors return empty results

**Testing Steps**:
1. **Deploy No-Data SLOs**: Use real metrics with impossible selectors
2. **Prometheus Validation**: Confirm queries return empty results
3. **Recording Rule Behavior**: Check if recording rules handle empty data
4. **UI Component Response**: Test BurnRateThresholdDisplay behavior

**Expected Results**:
- Prometheus queries execute but return no data
- Recording rules don't produce errors (may produce no output)
- UI component shows appropriate fallback ("Traffic-Aware" or error message)
- No mathematical errors (division by zero, etc.)

**Mathematical Edge Case Testing**:
```bash
# Test queries that should return no data
curl -s "http://localhost:9090/api/v1/query?query=prometheus_http_requests_total{code=\"999\"}"

# Should return: {"status":"success","data":{"resultType":"vector","result":[]}}
```

### **Test 12.3: Mathematical Edge Cases**
**Objective**: Test mathematical stability with extreme or problematic values

**Edge Case Scenarios**:
1. **Zero Traffic (N_long = 0)**:
   - Very short-lived environment
   - Recent deployment with no historical data
   - Should not cause division by zero

2. **Extreme Traffic Ratios**:
   - Very high recent traffic (ratio >> 1)
   - Very low recent traffic (ratio << 1)
   - Should produce bounded, reasonable thresholds

3. **Very Small Numbers**:
   - SLOs with very high targets (99.99%)
   - Very small window periods
   - Should handle precision appropriately

**Testing Approach**:
Create controlled scenarios or use mathematical simulation:
```python
# Test mathematical edge cases
python -c "
import math

# Test division by zero handling
def safe_threshold_calc(n_slo, n_long, e_budget_percent, slo_target):
    if n_long == 0:
        return None  # Should return meaningful fallback
    ratio = n_slo / n_long
    return ratio * e_budget_percent * (1 - slo_target)

# Test extreme values
test_cases = [
    (1000, 0, 0.020833, 0.95),      # Division by zero
    (1000000, 1, 0.020833, 0.95),   # Very high ratio
    (1, 1000000, 0.020833, 0.95),   # Very low ratio
    (100, 50, 0.020833, 0.9999),    # Very high SLO target
]

for n_slo, n_long, e_budget, target in test_cases:
    result = safe_threshold_calc(n_slo, n_long, e_budget, target)
    print(f'N_SLO: {n_slo}, N_long: {n_long}, Result: {result}')
"
```

### **Test 12.4: Insufficient Data Scenarios**
**Objective**: Test behavior with very limited metric history

**Scenarios**:
1. **Short-Lived Environment**:
   - Environment running < 1 hour
   - Long windows (4d) have no data
   - Should handle gracefully

2. **Sparse Metrics**:
   - Metrics with very infrequent data points
   - Large gaps in metric history
   - Should not break calculations

**Testing Steps**:
1. **Simulate Short Environment**: Use metrics with very limited history
2. **Check Query Results**: Verify Prometheus queries return reasonable data
3. **UI Behavior**: Test component behavior with insufficient data
4. **Error Messages**: Validate meaningful error communication

### **Test 12.5: Recovery Behavior**
**Objective**: Test system recovery when metrics become available

**Testing Approach**:
1. **Start with Missing Metrics**: Deploy SLO with non-existent metrics
2. **Create Metrics**: Generate the missing metrics (if possible)
3. **Monitor Recovery**: Check if system automatically recovers
4. **UI Updates**: Verify UI updates when data becomes available

## 🎯 Session Success Criteria

### **Minimum Success**:
- [ ] No crashes or system failures with missing metrics
- [ ] Consistent error handling between static and dynamic SLOs
- [ ] UI provides meaningful feedback for error states
- [ ] Mathematical edge cases don't break calculations

### **Full Success**:
- [ ] Comprehensive error state documentation
- [ ] Graceful degradation for all tested scenarios
- [ ] Recovery behavior validated
- [ ] Performance impact of error scenarios minimal

### **Exceptional Success**:
- [ ] Detailed error message customization for different failure modes
- [ ] Automatic retry/recovery mechanisms documented
- [ ] Performance optimizations for missing metric scenarios

## 📝 Expected Challenges

### **Likely Issues**:
1. **Inconsistent Error Handling**: Static vs dynamic may behave differently
2. **Mathematical Instability**: Division by zero or extreme values
3. **UI Error States**: Component may not handle all error cases gracefully
4. **Performance Impact**: Error scenarios may cause query timeouts

### **Mitigation Strategies**:
- Test static and dynamic SLOs in parallel for comparison
- Have mathematical fallback strategies ready
- Monitor system performance during error scenario testing
- Document all discovered edge cases for future improvement

## 🔧 Session Methodology

### **Testing Approach**:
- **Parallel Testing**: Create both static and dynamic versions of problematic SLOs
- **Incremental Complexity**: Start with simple missing metrics, progress to edge cases
- **System Monitoring**: Watch logs, performance, and overall system health
- **Recovery Testing**: Validate system can recover from error states

### **Documentation Focus**:
- **Error Patterns**: Document all discovered failure modes
- **Handling Strategies**: Note how system currently handles each error type
- **Improvement Opportunities**: Identify areas for better error handling

## 📋 Session Preparation

### **Pre-Session Checklist**:
- [ ] Current environment stable with working dynamic SLOs
- [ ] Access to Pyrra logs and Kubernetes events
- [ ] Ability to create and delete test SLOs quickly
- [ ] Backup plan to restore environment if needed

### **Safety Measures**:
- [ ] Test SLOs in separate namespace if possible
- [ ] Monitor system health throughout testing
- [ ] Have working SLO examples ready for recovery testing
- [ ] Document baseline performance before testing

**Status**: 🛡️ **READY FOR RESILIENCE AND ERROR HANDLING VALIDATION**
