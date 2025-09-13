# Session 10B: Latency Indicator UI Threshold Display Enhancement

## 🎯 Session Objective

**Primary Goal**: Complete UI threshold display enhancement for latency-based dynamic SLOs to achieve full feature parity with ratio indicators.

**Scope**: UI component enhancement only - backend functionality confirmed working
**Status**: Session 10A completed backend validation, UI threshold display gap identified
**Session Priority**: HIGH (complete latency indicator validation)

## 📋 Session Context

### **What We Know Works (Session 10A Results)**:
- ✅ Latency dynamic SLO backend rule generation working correctly
- ✅ Mathematical validation completed (traffic ratios, error budget calculations)  
- ✅ Error budget display fixed (-1900% → ~97%)
- ✅ Recording rules architecture properly implemented (dual rules: total + success)
- ✅ PrometheusRule expressions parse and execute correctly

### **What Needs Completion**:
- ❌ **UI Threshold Display**: Still shows "Traffic-Aware" instead of calculated values
- ❌ **Component Architecture**: BurnRateThresholdDisplay needs histogram metric support
- ❌ **Tooltip Enhancement**: Should show latency-specific calculation details

## 🔧 Session Setup Requirements

### **Environment Prerequisites**:
- Session 10A completed (latency backend working)
- `test-latency-dynamic` SLO deployed and functional
- BurnRateThresholdDisplay component exists (from Test 9 - ratio indicators)
- UI build environment ready (`cd ui && npm run build`)

### **Technical Context**:
- **Working Latency SLO**: `test-latency-dynamic` with `prometheus_http_request_duration_seconds`
- **Current UI Issue**: Threshold column shows "Traffic-Aware" instead of calculated values
- **Expected Behavior**: Show calculated thresholds like `"0.025625 (6h26m factor 7)"`
- **Architecture Reference**: Existing ratio indicator support in BurnRateThresholdDisplay

## 📊 Specific Tasks

### **Task 1: Component Architecture Analysis**
**Objective**: Understand why BurnRateThresholdDisplay doesn't support latency indicators

**Investigation Steps**:
1. **Read Current Component**: Examine `ui/src/components/BurnRateThresholdDisplay.tsx`
2. **Identify Gaps**: Find ratio-specific logic that excludes latency indicators
3. **Check Metric Extraction**: Verify `extractTotalMetric()` and `extractErrorMetric()` functions
4. **Understand Query Generation**: Analyze how Prometheus queries are built for thresholds

**Expected Findings**:
- Component may only handle ratio indicators (`request_total` metrics)
- Histogram metric extraction (`_count`, `_bucket`) may be missing
- Query patterns may be ratio-specific (`sum by (code)` vs histogram aggregation)

### **Task 2: Extend Component for Histogram Support**
**Objective**: Add latency indicator support to BurnRateThresholdDisplay component

**Implementation Requirements**:
1. **Histogram Metric Detection**: Identify when SLO uses latency indicator
2. **Metric Extraction Enhancement**: Support `_count` and `_bucket` metric extraction
3. **Query Generation**: Generate histogram-appropriate Prometheus queries
4. **Traffic Calculation**: Implement histogram-based traffic ratio queries

**Code Pattern Example**:
```typescript
// For latency indicators, extract histogram metrics
const totalMetric = extractHistogramTotalMetric(objective); // _count metric
const errorMetric = extractHistogramErrorMetric(objective); // _bucket metric with le

// Generate histogram traffic queries
const trafficQuery = `sum(increase(${totalMetric}[${window}]))`;
```

### **Task 3: Mathematical Integration**
**Objective**: Integrate Session 10A mathematical findings into UI display

**Session 10A Results to Use**:
- **Traffic Ratio Pattern**: `N_SLO / N_long` using histogram count metrics
- **Expected Thresholds**: Factor 7 = ~0.025625 (2.56% error rate)
- **Formula**: `(N_SLO / N_long) × E_budget_percent × (1 - SLO_target)`
- **Window Mappings**: Same as ratio indicators (14: 1h4m, 7: 6h26m, etc.)

**Implementation**:
- Use histogram count queries for traffic calculation
- Apply same mathematical formula as ratio indicators
- Display results in same format as static SLOs

### **Task 4: UI Integration Testing**
**Objective**: Verify latency threshold display works correctly

**Testing Checklist**:
- [ ] Navigate to `test-latency-dynamic` SLO detail page
- [ ] Verify threshold column shows calculated values (not "Traffic-Aware")
- [ ] Confirm values are reasonable (0.001-0.1 range for typical thresholds)
- [ ] Check tooltip displays histogram-specific information
- [ ] Verify no JavaScript errors in browser console
- [ ] Compare with ratio indicator display for consistency

**Success Criteria**:
- Latency SLO shows numerical threshold values like ratio indicators
- Values reflect histogram-based traffic patterns from Session 10A
- Component renders without errors or performance issues

## 🎯 Session Success Criteria

### **Minimum Success**:
- [ ] Component architecture analysis complete - gaps identified
- [ ] Basic histogram metric support added to BurnRateThresholdDisplay
- [ ] No crashes or errors when viewing latency dynamic SLO detail page

### **Full Success**:
- [ ] Latency dynamic SLO shows calculated threshold values (not "Traffic-Aware")
- [ ] Values match mathematical expectations from Session 10A validation
- [ ] UI experience equivalent to ratio indicator threshold display
- [ ] Tooltip provides histogram-specific calculation details

### **Exceptional Success**:
- [ ] Performance equivalent to ratio indicator queries
- [ ] Enhanced tooltip showing traffic patterns (e.g., "8.2x recent traffic")
- [ ] Comprehensive error handling for edge cases (missing histogram data)

## 📝 Implementation Strategy

### **Approach 1: Extend Existing Component (Recommended)**
**Rationale**: BurnRateThresholdDisplay already works for ratio indicators
**Method**: Add histogram detection and query generation logic
**Advantage**: Maintains architectural consistency, leverages existing patterns

### **Approach 2: Create Histogram-Specific Component**
**Rationale**: Different enough to warrant separate implementation
**Method**: Create LatencyThresholdDisplay component
**Risk**: Code duplication, maintenance overhead

### **Recommended Pattern**:
```typescript
// Inside BurnRateThresholdDisplay
const isLatencyIndicator = objective.indicator?.latency !== undefined;
if (isLatencyIndicator) {
  // Use histogram-specific logic
  const histogramQueries = buildHistogramTrafficQueries(objective, factors);
} else {
  // Use existing ratio logic
  const ratioQueries = buildRatioTrafficQueries(objective, factors);
}
```

## 🔧 Technical Notes

### **Session 10A Validated Architecture**:
- **Recording Rules**: Both total (`le=""`) and success (`le="0.1"`) rules exist
- **Traffic Queries**: Use `prometheus_http_request_duration_seconds_count` for total requests
- **Success Calculation**: Use `prometheus_http_request_duration_seconds_bucket{le="0.1"}` for success
- **Formula Confirmed**: Same mathematical pattern as ratio indicators

### **Key Difference from Ratio Indicators**:
- **Ratio**: Single metric with multiple label values (`code="200"`, `code="500"`)
- **Latency**: Multiple metrics for total vs success (`_count` vs `_bucket{le="X"}`)
- **Query Pattern**: Histogram aggregation instead of label-based grouping

### **Performance Considerations**:
- Histogram queries may be slightly slower than ratio queries
- Use existing recording rules when possible (leverage Session 10A architecture)
- Consider caching threshold calculations (same pattern as ratio indicators)

## 📋 Session Deliverables

### **Code Changes**:
- [ ] Enhanced `BurnRateThresholdDisplay.tsx` with histogram support
- [ ] Updated metric extraction functions for latency indicators
- [ ] Histogram-specific query generation logic
- [ ] UI build and validation (`npm run build`)

### **Testing Results**:
- [ ] Screenshot or description of working latency threshold display
- [ ] Performance comparison with ratio indicator queries
- [ ] Edge case handling validation (missing data, etc.)

### **Documentation Updates**:
- [ ] Update `DYNAMIC_BURN_RATE_TESTING_SESSION.md` with Session 10B results
- [ ] Document any new architectural patterns or lessons learned
- [ ] Note completion of latency indicator validation

## 🚀 Session Completion Criteria

**Session 10B Complete When**:
1. Latency dynamic SLO threshold display shows calculated values
2. UI experience matches ratio indicator threshold display quality
3. No functional regressions in existing ratio indicator behavior
4. Documentation updated with session results

**Next Session Preparation**:
- Session 10A + 10B complete → Consider Session 11 (other indicator types)
- Or Session 12 (resilience testing) if comprehensive latency validation desired
- Update `prompts/README.md` with completion status

**Status**: 🎯 **READY FOR LATENCY UI THRESHOLD DISPLAY COMPLETION**