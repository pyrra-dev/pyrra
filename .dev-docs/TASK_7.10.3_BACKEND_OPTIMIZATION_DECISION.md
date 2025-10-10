# Task 7.10.3: Backend Alert Rule Query Optimization Decision

## Decision Date
2025-01-10

## Context

Following the UI query optimization in Task 7.10.2, this task evaluates whether backend alert rules should also be optimized to use recording rules for the SLO window calculation.

### Reference Documents
- `.dev-docs/TASK_7.10_VALIDATION_RESULTS.md` - Performance measurements showing 7x speedup for ratio, 2x for latency
- `.dev-docs/TASK_7.10_IMPLEMENTATION.md` - UI optimization implementation and real-world performance analysis
- `.dev-docs/TASK_7.10.1_TEST_IMPROVEMENTS.md` - Test methodology and recording rule availability

### Key Findings from UI Optimization
- **Query speedup**: 7.17x for ratio, 2.20x for latency indicators
- **UI benefit**: Minimal (~5% of 110ms total) due to network overhead
- **Primary benefit**: Prometheus load reduction, not query speed
- **Hybrid approach**: Recording rules for SLO window + inline calculations for alert windows

## Current Backend Implementation

### Alert Rule Query Pattern (slo/rules.go)

**Current Implementation** (uses raw metrics for both SLO and alert windows):

```go
// Ratio indicator example
fmt.Sprintf(
    "(%s{%s} > scalar((sum(increase(%s{%s}[%s])) / sum(increase(%s{%s}[%s]))) * %f * (1-%s))) and " +
    "(%s{%s} > scalar((sum(increase(%s{%s}[%s])) / sum(increase(%s{%s}[%s]))) * %f * (1-%s)))",
    // Short window: burnrate > dynamic threshold
    o.BurnrateName(w.Short), recordingRuleSelector,
    o.Indicator.Ratio.Total.Name, rawMetricSelector, sloWindow,  // N_SLO (raw metric)
    o.Indicator.Ratio.Total.Name, rawMetricSelector, longWindow, // N_alert (raw metric)
    eBudgetPercent, targetStr,
    // Long window: burnrate > dynamic threshold
    o.BurnrateName(w.Long), recordingRuleSelector,
    o.Indicator.Ratio.Total.Name, rawMetricSelector, sloWindow,  // N_SLO (raw metric)
    o.Indicator.Ratio.Total.Name, rawMetricSelector, longWindow, // N_alert (raw metric)
    eBudgetPercent, targetStr,
)
```

**Generated Alert Rule Example**:
```promql
(
  apiserver_request:burnrate5m{slo="test-dynamic-apiserver"} > 
  scalar((sum(increase(apiserver_request_total[30d])) / sum(increase(apiserver_request_total[1h4m]))) * 0.020833 * (1-0.95))
) and (
  apiserver_request:burnrate1h{slo="test-dynamic-apiserver"} > 
  scalar((sum(increase(apiserver_request_total[30d])) / sum(increase(apiserver_request_total[1h4m]))) * 0.020833 * (1-0.95))
)
```

### Potential Optimized Pattern

**Hybrid Approach** (recording rule for SLO window + inline for alert window):

```go
// Ratio indicator example
fmt.Sprintf(
    "(%s{%s} > scalar((sum(%s:increase%s{slo=\"%s\"}) / sum(increase(%s{%s}[%s]))) * %f * (1-%s))) and " +
    "(%s{%s} > scalar((sum(%s:increase%s{slo=\"%s\"}) / sum(increase(%s{%s}[%s]))) * %f * (1-%s)))",
    // Short window: burnrate > dynamic threshold
    o.BurnrateName(w.Short), recordingRuleSelector,
    baseMetricName, sloWindow, sloName,                          // N_SLO (recording rule)
    o.Indicator.Ratio.Total.Name, rawMetricSelector, longWindow, // N_alert (raw metric)
    eBudgetPercent, targetStr,
    // Long window: burnrate > dynamic threshold
    o.BurnrateName(w.Long), recordingRuleSelector,
    baseMetricName, sloWindow, sloName,                          // N_SLO (recording rule)
    o.Indicator.Ratio.Total.Name, rawMetricSelector, longWindow, // N_alert (raw metric)
    eBudgetPercent, targetStr,
)
```

**Generated Alert Rule Example**:
```promql
(
  apiserver_request:burnrate5m{slo="test-dynamic-apiserver"} > 
  scalar((sum(apiserver_request:increase30d{slo="test-dynamic-apiserver"}) / sum(increase(apiserver_request_total[1h4m]))) * 0.020833 * (1-0.95))
) and (
  apiserver_request:burnrate1h{slo="test-dynamic-apiserver"} > 
  scalar((sum(apiserver_request:increase30d{slo="test-dynamic-apiserver"}) / sum(increase(apiserver_request_total[1h4m]))) * 0.020833 * (1-0.95))
)
```

## Analysis

### Performance Characteristics

#### Alert Rule Evaluation Profile
- **Evaluation Frequency**: Every 30 seconds (configured in rule group interval)
- **Evaluation Context**: Prometheus rule engine (server-side)
- **Query Execution**: Synchronous, blocking rule evaluation
- **Performance Impact**: Affects Prometheus CPU/memory usage

#### UI Query Profile (for comparison)
- **Execution Frequency**: On-demand (user navigation)
- **Execution Context**: HTTP API request (client-initiated)
- **Query Execution**: Asynchronous, non-blocking
- **Performance Impact**: Affects user experience and API responsiveness

### Expected Performance Improvement

Based on Task 7.10.1 validation results:

| Indicator | Raw Metrics (Avg) | Recording Rules (Avg) | Speedup | Per-Alert Savings |
|-----------|-------------------|----------------------|---------|-------------------|
| **Ratio** | 48.75ms | 6.80ms | **7.17x** | ~42ms per alert |
| **Latency** | 6.34ms | 2.89ms | **2.20x** | ~3.5ms per alert |
| **BoolGauge** | 3.02ms | 3.02ms | **1.0x** | No benefit |

**Calculation for Production Impact**:
- 10 dynamic SLOs × 4 alert windows = 40 alert rules
- Each alert evaluates every 30 seconds
- Ratio indicators: 40 alerts × 42ms = 1.68 seconds saved per evaluation cycle
- Latency indicators: 40 alerts × 3.5ms = 140ms saved per evaluation cycle

**Annual Prometheus Load Reduction**:
- Ratio: 1.68s × 2 evaluations/min × 60 min × 24 hr × 365 days = ~1.77 million seconds/year
- Latency: 140ms × 2 evaluations/min × 60 min × 24 hr × 365 days = ~147,000 seconds/year

### Benefits of Optimization

#### 1. Prometheus Load Reduction (PRIMARY BENEFIT)
- **Reduced CPU usage**: 7x fewer data points scanned for ratio indicators
- **Reduced memory usage**: Smaller working set for query evaluation
- **Better scalability**: More SLOs can be managed on same Prometheus instance
- **Lower infrastructure costs**: Reduced resource requirements

#### 2. Consistent Data Source
- **UI and alerts use same data**: Both use recording rules for SLO window
- **Eliminates discrepancies**: No timing differences between UI and alert calculations
- **Easier debugging**: Single source of truth for traffic calculations

#### 3. Architectural Alignment
- **Follows Pyrra design**: Recording rules exist specifically for this purpose
- **Consistent with UI**: Same hybrid approach (recording rule + inline)
- **Best practice**: Use pre-computed metrics when available

#### 4. Alert Evaluation Speed (SECONDARY BENEFIT)
- **Faster rule evaluation**: 7x speedup for ratio, 2x for latency
- **More responsive alerting**: Slightly faster alert firing (negligible in practice)
- **Reduced evaluation backlog**: Less likely to fall behind during high load

### Risks and Considerations

#### 1. Implementation Complexity
- **Moderate complexity**: Need to extract base metric name and construct recording rule query
- **Multiple indicator types**: Must handle ratio, latency, latencyNative, boolGauge
- **Testing required**: Must validate alert rules still fire correctly

#### 2. Recording Rule Dependency
- **Requires recording rules**: Alert rules depend on recording rules being available
- **Startup timing**: Recording rules must be evaluated before alert rules
- **Fallback not possible**: Alert rules can't fall back to raw metrics (unlike UI)

#### 3. Backward Compatibility
- **No impact on static SLOs**: Optimization only affects dynamic burn rate alerts
- **No API changes**: Alert rule structure remains the same
- **No CRD changes**: SLO configuration unchanged

#### 4. Maintenance Burden
- **Additional code paths**: More complexity in buildDynamicAlertExpr()
- **Testing overhead**: Must test both raw and optimized query patterns
- **Documentation**: Need to document optimization strategy

### Comparison with UI Optimization

| Aspect | UI Optimization | Backend Optimization |
|--------|----------------|---------------------|
| **Primary Benefit** | Prometheus load reduction | Prometheus load reduction |
| **Secondary Benefit** | Query speed (minimal UI impact) | Alert evaluation speed |
| **Execution Frequency** | On-demand (user navigation) | Every 30s (continuous) |
| **Fallback Strategy** | Can fall back to raw metrics | No fallback possible |
| **Implementation Risk** | Low (UI can degrade gracefully) | Medium (alerts must work) |
| **Testing Complexity** | Low (visual validation) | High (alert firing validation) |
| **Maintenance Burden** | Low (single component) | Medium (multiple indicator types) |

## Decision

### **RECOMMENDATION: IMPLEMENT OPTIMIZATION**

**Rationale**:

1. **Significant Prometheus Load Reduction**: The primary benefit of reducing Prometheus CPU/memory usage is substantial, especially at scale. With 40 alert rules evaluating every 30 seconds, the cumulative savings are significant.

2. **Consistent with UI Implementation**: The UI already uses this optimization pattern. Applying the same pattern to backend alert rules ensures consistency and eliminates potential discrepancies.

3. **Architectural Alignment**: Pyrra generates recording rules specifically for this purpose. Using them in alert rules aligns with the system's design intent.

4. **Proven Pattern**: The hybrid approach (recording rule for SLO window + inline for alert window) is validated and working in the UI.

5. **Acceptable Risk**: While implementation complexity is moderate, the benefits outweigh the risks. Comprehensive testing can mitigate the risks.

### Implementation Strategy

#### Phase 1: Core Implementation
1. **Add helper function** to extract base metric name (similar to UI implementation)
2. **Update buildDynamicAlertExpr()** to use hybrid query pattern
3. **Implement for ratio indicators first** (highest benefit: 7x speedup)
4. **Implement for latency indicators** (good benefit: 2x speedup)
5. **Skip boolGauge optimization** (no benefit, already fast)

#### Phase 2: Testing and Validation
1. **Unit tests**: Verify query generation produces correct PromQL
2. **Integration tests**: Deploy test SLOs and verify alert rules are created correctly
3. **Alert firing tests**: Use existing `run-synthetic-test` tool to verify alerts fire
4. **Performance validation**: Measure Prometheus CPU/memory usage before/after

#### Phase 3: Documentation
1. **Update design document**: Document optimization strategy
2. **Update steering document**: Add backend optimization to development standards
3. **Create implementation guide**: Document query patterns for future reference

### Implementation Priority

**Priority**: **MEDIUM-HIGH**

**Timing**: Implement after Task 7.10.4 (final UI validation) is complete

**Estimated Effort**: 2-3 hours implementation + 2-3 hours testing

### Success Criteria

✅ **Implementation Complete**:
- [ ] Helper function added to extract base metric name
- [ ] buildDynamicAlertExpr() updated for ratio indicators
- [ ] buildDynamicAlertExpr() updated for latency indicators
- [ ] Unit tests pass for all indicator types
- [ ] Integration tests verify alert rules are created correctly

✅ **Validation Complete**:
- [ ] Alert firing tests pass (using run-synthetic-test)
- [ ] Prometheus CPU/memory usage reduced (measured)
- [ ] No regressions in alert behavior
- [ ] Documentation updated

## Alternative Considered: Do Not Optimize

### Rationale for Rejection

**Why not optimize?**
- Current implementation works correctly
- Alert rules already use recording rules for burn rate calculation
- Only the dynamic threshold calculation uses raw metrics
- Implementation adds complexity

**Why this was rejected**:
- Prometheus load reduction is significant at scale
- Consistency with UI implementation is valuable
- Architectural alignment with Pyrra's design
- Benefits clearly outweigh the moderate implementation complexity

## Implementation Code Sketch

### Helper Function (similar to UI)

```go
// getBaseMetricName strips common suffixes to match recording rule naming
func getBaseMetricName(metricName string) string {
    metricName = strings.TrimSuffix(metricName, "_total")
    metricName = strings.TrimSuffix(metricName, "_count")
    metricName = strings.TrimSuffix(metricName, "_bucket")
    return metricName
}
```

### Updated buildDynamicAlertExpr() for Ratio Indicators

```go
case Ratio:
    recordingRuleSelector := alertMatchersString
    rawMetricSelector := o.buildTotalSelector(alertMatchersString)
    baseMetricName := getBaseMetricName(o.Indicator.Ratio.Total.Name)
    sloName := o.Labels.Get(labels.MetricName)

    return fmt.Sprintf(
        "("+
            "%s{%s} > "+
            "scalar((sum(%s:increase%s{slo=\"%s\"}) / sum(increase(%s{%s}[%s]))) * %f * (1-%s))"+
            ") and ("+
            "%s{%s} > "+
            "scalar((sum(%s:increase%s{slo=\"%s\"}) / sum(increase(%s{%s}[%s]))) * %f * (1-%s))"+
            ")",
        // Short window: use recording rule > dynamic threshold
        o.BurnrateName(w.Short), recordingRuleSelector,
        // Short window dynamic threshold: recording rule for SLO window
        baseMetricName, sloWindow, sloName,
        // Short window dynamic threshold: inline for alert window
        o.Indicator.Ratio.Total.Name, rawMetricSelector, longWindow,
        eBudgetPercent, targetStr,
        // Long window: use recording rule > dynamic threshold
        o.BurnrateName(w.Long), recordingRuleSelector,
        // Long window dynamic threshold: recording rule for SLO window
        baseMetricName, sloWindow, sloName,
        // Long window dynamic threshold: inline for alert window
        o.Indicator.Ratio.Total.Name, rawMetricSelector, longWindow,
        eBudgetPercent, targetStr,
    )
```

## Next Steps

1. **Complete Task 7.10.4**: Finish UI validation and documentation cleanup
2. **Create Task 7.10.5**: Implement backend alert rule optimization (new sub-task)
3. **Update tasks.md**: Add Task 7.10.5 to implementation plan
4. **Consult user**: Confirm priority and timing for backend optimization

## References

- **Performance Validation**: `.dev-docs/TASK_7.10_VALIDATION_RESULTS.md`
- **UI Implementation**: `.dev-docs/TASK_7.10_IMPLEMENTATION.md`
- **Test Methodology**: `.dev-docs/TASK_7.10.1_TEST_IMPROVEMENTS.md`
- **Backend Code**: `slo/rules.go` (buildDynamicAlertExpr function)
- **Requirements**: Task 7.10.3 in `.kiro/specs/dynamic-burn-rate-completion/tasks.md`

## Conclusion

**Backend alert rule optimization is recommended** based on:
- Significant Prometheus load reduction (primary benefit)
- Consistency with UI implementation
- Architectural alignment with Pyrra's design
- Proven hybrid approach pattern
- Acceptable implementation complexity

The optimization should be implemented as a follow-up task (7.10.5) after completing UI validation (7.10.4).
