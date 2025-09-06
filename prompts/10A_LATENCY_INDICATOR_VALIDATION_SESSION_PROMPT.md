# Session 10: Latency Indicator Dynamic SLO Validation

## 🎯 Session Objective

**Primary Goal**: Validate dynamic burn rate functionality specifically for latency-based SLO indicators.

**Scope**: Single indicator type focus - latency indicators only
**Current Status**: Only ratio indicators tested, latency indicators unvalidated
**Session Priority**: HIGH (latency indicators are very common in production)

## 📋 Session Context

### **What We Know Works**:
- ✅ Dynamic burn rate for ratio indicators (`apiserver_request_total`)
- ✅ BurnRateThresholdDisplay component architecture
- ✅ Traffic-aware threshold calculations with real data
- ✅ UI integration with existing Pyrra patterns

### **What We Need to Validate**:
- ❓ Latency indicator dynamic expression generation
- ❓ Histogram-based traffic calculation queries  
- ❓ UI component behavior with latency SLOs
- ❓ Mathematical accuracy for histogram metrics

## 🔧 Session Setup Requirements

### **Environment Prerequisites**:
- Existing Pyrra environment with ratio dynamic SLOs working
- Access to histogram metrics (e.g., `prometheus_http_request_duration_seconds`)
- Kubernetes cluster with ability to deploy new SLOs

### **Test SLO Creation**:
Create latency-based dynamic SLO using available histogram metrics:
```yaml
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: test-latency-dynamic
  namespace: monitoring
spec:
  target: 0.95
  window: 30d
  burnRateType: dynamic
  indicator:
    latency:
      success:
        metric: prometheus_http_request_duration_seconds_bucket
        le: "0.1"  # 100ms threshold
      total:
        metric: prometheus_http_request_duration_seconds_count
```

## 📊 Test Scenarios

### **Test 10.1: Backend Rule Generation Validation**
**Objective**: Verify Pyrra generates correct PromQL for latency dynamic SLOs

**Validation Steps**:
1. **Deploy Test SLO**: Apply latency dynamic SLO to cluster
2. **Check PrometheusRule**: Examine generated rule expressions
3. **Validate Queries**: Confirm histogram aggregation patterns
4. **Compare with Static**: Check difference from static latency SLO

**Expected Behaviors**:
- PrometheusRule created without errors
- Dynamic expressions use histogram_count() or similar for traffic calculation
- Recording rules generated for latency metrics
- Alert expressions include traffic ratio calculations

**Success Criteria**:
- Generated PromQL contains histogram-based traffic calculations
- Alert expressions follow expected pattern: `(N_SLO/N_long) × E_budget_percent × (1-SLO_target)`
- No errors in Kubernetes or Pyrra logs

### **Test 10.2: Mathematical Validation**
**Objective**: Verify traffic ratio calculations work correctly with histogram data

**Validation Steps**:
1. **Extract Generated Queries**: Get actual PromQL from PrometheusRule
2. **Manual Query Testing**: Test queries in Prometheus UI
3. **Python Calculation**: Verify math using extracted values
4. **Compare Results**: UI values vs manual calculations

**Mathematical Verification**:
```bash
# Extract histogram count for SLO window (30d)
curl -s "http://localhost:9090/api/v1/query?query=sum(increase(prometheus_http_request_duration_seconds_count[30d]))"

# Extract histogram count for factor window (e.g., 1h4m for factor 14)
curl -s "http://localhost:9090/api/v1/query?query=sum(increase(prometheus_http_request_duration_seconds_count[1h4m]))"

# Python calculation verification
python -c "
n_slo = <extracted_30d_value>
n_long = <extracted_1h4m_value>
traffic_ratio = n_slo / n_long
e_budget_percent = 0.020833  # factor 14
slo_target = 0.95
threshold = traffic_ratio * e_budget_percent * (1 - slo_target)
print(f'Expected threshold: {threshold:.12f}')
"
```

**Success Criteria**:
- Histogram queries return meaningful data
- Traffic ratio calculations produce reasonable values (not 0, infinity, or NaN)
- Manual calculations match what UI component would display

### **Test 10.3: UI Component Integration**
**Objective**: Verify BurnRateThresholdDisplay works with latency dynamic SLOs

**Validation Steps**:
1. **Navigate to Detail Page**: Go to test-latency-dynamic SLO detail page
2. **Check Threshold Display**: Verify calculated values (not "Traffic-Aware")
3. **Tooltip Validation**: Confirm tooltip shows appropriate information
4. **Error Handling**: Test behavior if histogram data is insufficient

**UI Validation Checklist**:
- [ ] Dynamic SLO shows green "Dynamic" badge
- [ ] Threshold column shows calculated values instead of "Traffic-Aware"
- [ ] Values are reasonable (not 0, infinity, or error messages)
- [ ] Tooltips provide histogram-specific information
- [ ] No JavaScript errors in browser console
- [ ] Component loads within reasonable time

**Success Criteria**:
- BurnRateThresholdDisplay component renders successfully
- Shows calculated threshold values for all factors (14, 7, 2, 1)
- Values reflect histogram-based traffic patterns
- No crashes or error states

### **Test 10.4: Query Performance Assessment**
**Objective**: Ensure histogram queries perform acceptably

**Performance Testing**:
1. **Query Execution Time**: Measure histogram query performance
2. **UI Responsiveness**: Check component loading time
3. **Resource Usage**: Monitor Prometheus query load
4. **Comparison**: Compare performance with ratio indicator queries

**Performance Metrics**:
- Query execution time < 5 seconds
- UI component renders within 3 seconds
- No significant impact on Prometheus performance
- Comparable performance to ratio indicators

## 🎯 Session Success Criteria

### **Minimum Success**:
- [ ] Latency dynamic SLO deploys without errors
- [ ] PrometheusRule generates with histogram-based expressions
- [ ] Basic UI component functionality (no crashes)
- [ ] Mathematical validation shows reasonable values

### **Full Success**:
- [ ] Complete mathematical validation with histogram data
- [ ] UI displays accurate calculated thresholds
- [ ] Performance acceptable for production use
- [ ] Comprehensive error handling validation

### **Exceptional Success**:
- [ ] Performance better than or equal to ratio indicators
- [ ] Edge case handling documented (empty histograms, etc.)
- [ ] Detailed troubleshooting guide for latency indicators

## 📝 Expected Challenges

### **Likely Issues**:
1. **Query Complexity**: Histogram aggregation more complex than simple counters
2. **Metric Extraction**: `getBaseMetricSelector()` may need histogram-specific logic
3. **Performance**: Histogram queries potentially slower than ratio queries
4. **Data Availability**: Histogram metrics may have different data patterns

### **Mitigation Strategies**:
- Test with well-populated histogram metrics first
- Have backup metrics ready if primary choice has insufficient data
- Monitor query performance throughout testing
- Document any required code changes for histogram support

## 🔧 Session Methodology

### **Testing Approach**:
- **Sequential Validation**: Backend → Math → UI → Performance
- **Immediate Issue Resolution**: Fix any discovered problems before proceeding
- **Comprehensive Documentation**: Capture all findings and edge cases

### **Documentation Updates**:
- Update DYNAMIC_BURN_RATE_TESTING_SESSION.md with Test 10 results
- Note any code changes required for histogram support
- Document performance characteristics vs ratio indicators

## 📋 Session Preparation

### **Pre-Session Checklist**:
- [ ] Current environment working (ratio dynamic SLOs functional)
- [ ] Identify available histogram metrics with good data
- [ ] Prepare latency dynamic SLO YAML definition
- [ ] Ensure access to Prometheus UI for manual query testing

### **Post-Session Deliverables**:
- [ ] Test 10 results documented in testing session file
- [ ] Any required code changes identified and documented
- [ ] Performance baseline established for histogram indicators
- [ ] Clear status on latency indicator readiness (working/needs fixes/not feasible)

**Status**: 🎯 **READY FOR LATENCY INDICATOR FOCUSED VALIDATION**
