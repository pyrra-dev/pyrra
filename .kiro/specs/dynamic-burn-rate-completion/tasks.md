# Dynamic Burn Rate Feature Completion - Implementation Plan

This implementation plan breaks down the remaining work to complete the dynamic burn rate feature for Pyrra. Each task is designed to be executed independently while building upon previous work.

**Context**: The feature is ~30% complete with backend implementation finished and basic UI working for ratio indicators. Remaining work focuses on comprehensive validation, additional indicator types, and production readiness.

## Task Group 1: Latency Indicator UI Completion (HIGH PRIORITY)

- [x] 1. Enhance BurnRateThresholdDisplay component for comprehensive latency indicator support

  - Extend existing component to handle histogram metrics (`_count`, `_bucket`) extraction
  - Implement histogram-specific Prometheus query generation for traffic calculations
  - Add error handling for missing histogram data scenarios
  - _Requirements: 1.1, 2.1, 3.1_

- [x] 1.1 Implement enhanced tooltip system for latency indicators

  - Extract traffic ratio calculations from existing Prometheus queries
  - Calculate average traffic for alert window to determine above/below average status
  - Generate static threshold comparison for user context
  - Update tooltip content to show traffic context and static vs dynamic comparison
  - _Requirements: 2.1, 2.4_

- [x] 1.2 Add performance monitoring and comparison framework

  - Implement query execution time measurement for histogram vs ratio indicators
  - Add UI component render time tracking for BurnRateThresholdDisplay
  - Create performance comparison logging and reporting
  - Validate latency indicator performance within acceptable limits (2x ratio performance)
  - _Requirements: 5.1, 5.3_

- [x] 1.3 Implement comprehensive error handling for latency indicators

  - Add graceful degradation when histogram `_count` or `_bucket` metrics are missing
  - Handle insufficient histogram data points with appropriate fallback displays
  - Implement retry logic for Prometheus query timeouts
  - Add meaningful error messages in browser console for debugging
  - _Requirements: 3.1, 3.2, 3.3, 3.4_

## Task Group 2: RequestsGraph Traffic Baseline Enhancement (HIGH PRIORITY)

- [x] 2. Enhance RequestsGraph component with traffic baseline visualization for dynamic SLOs

  - Add average traffic baseline calculation using same time window as longest alert window
  - Implement horizontal dashed line overlay showing average traffic rate
  - Add legend distinguishing "Current Traffic" (solid line) vs "Average Traffic" (dashed line)
  - Include traffic ratio tooltip when hovering over current traffic points
  - _Requirements: 2.1, 2.4_

- [x] 2.1 Implement traffic context indicators in RequestsGraph

  - Calculate traffic ratio using existing BurnRateThresholdDisplay patterns
  - Add visual indicators when current traffic is significantly above/below average (>1.5x or <0.5x)
  - Implement tooltip showing traffic ratio context: "Current: 2.3x above average"
  - Ensure enhancement only applies to dynamic burn rate SLOs
  - _Requirements: 2.1, 2.4_

- [x] 2.2 Enhance burn rate type badge tooltip with traffic context

  - Extract current traffic ratio from RequestsGraph calculations
  - Update existing burn rate type badge tooltip to include traffic context
  - Show impact on alert sensitivity: "Current traffic makes alerts 2.3x more sensitive"
  - Maintain existing tooltip structure while adding dynamic context
  - _Requirements: 2.1, 2.4_

## Task Group 3: Alerts Table Column Enhancement

- [x] 3. Add new "Error Budget %" column to AlertsTable for dynamic burn rates

  - Add new table column between existing columns to show error budget percentage constants
  - For dynamic SLOs: Show error budget percentage values (2.08%, 6.25%, 7.14%, 14.29%) corresponding to (1/48, 1/16, 1/14, 1/7)
  - For static SLOs: Show factor values (14, 7, 2, 1) in the new column
  - Keep existing "Threshold" column unchanged - it shows calculated threshold values via BurnRateThresholdDisplay
  - Update table layout to accommodate the additional column without breaking responsive design
  - **COMPLETED**: New "Error Budget Consumption" column added with dynamic/static values and enhanced tooltips
  - **BONUS**: Fixed duration precision issue - now shows complete window durations (e.g., "1d1h43m" instead of "1d1h")
  - _Requirements: 2.4_

- [x] 3.1 Validate BurnRateThresholdDisplay component integration with AlertsTable

  - Verify BurnRateThresholdDisplay correctly shows calculated threshold values for all indicator types
  - Test error handling for missing metrics and edge cases in alerts table context
  - Ensure consistent threshold display between ratio and latency indicators
  - Validate performance impact of real-time threshold calculations in table rows
  - _Requirements: 1.1, 1.2, 2.1_

- [x] 3.2 Enhance AlertsTable tooltip system for dynamic burn rates

  - Extract current traffic ratio from BurnRateThresholdDisplay calculations
  - Calculate average traffic for alert window comparison
  - Generate static threshold equivalent for comparison context
  - Update tooltip to show traffic context, static comparison, and formula explanation
  - **COMPLETED**: Enhanced tooltips with proper wording and DynamicBurnRateTooltip integration
  - _Requirements: 2.1, 2.4_

## Task Group 4: Additional Indicator Type Support

- [x] 4. Extend BurnRateThresholdDisplay for LatencyNative indicators

  - Add LatencyNative indicator detection logic in component
  - Implement native histogram metric extraction (`histogram_count`, `histogram_sum`)
  - Generate appropriate Prometheus queries for native histogram traffic calculations
  - Add LatencyNative-specific error handling and fallback behavior
  - _Requirements: 1.2_

- [x] 4.1 Extend BurnRateThresholdDisplay for BoolGauge indicators

  - Add BoolGauge indicator detection logic in component
  - Implement boolean gauge metric extraction and query generation
  - Use `count_over_time()` aggregation patterns for traffic calculations
  - Add BoolGauge-specific error handling and tooltip content
  - _Requirements: 1.3_

- [x] 4.2 Create comprehensive indicator type test suite

  - Write unit tests for all indicator types (Ratio, Latency, LatencyNative, BoolGauge)
  - Test metric extraction functions for each indicator type
  - Validate query generation produces correct PromQL for each type
  - Test error handling scenarios for missing metrics across all types
  - _Requirements: 1.1, 1.2, 1.3_

## Task Group 5: Resilience and Edge Case Testing

- [x] 5. Implement missing metrics handling validation

  - Create test SLOs with completely non-existent metrics
  - Validate Pyrra backend doesn't crash with fictional metrics
  - Test UI component graceful degradation with missing data
  - Ensure consistent error handling between static and dynamic SLOs
  - _Requirements: 3.1, 3.4_

- [x] 5.1 Implement mathematical edge case handling

  - Add division by zero protection in traffic ratio calculations
  - Handle extreme traffic ratios (very high/low) with bounded thresholds
  - Test precision handling for very small numbers (high SLO targets like 99.99%)
  - Implement conservative fallback calculations for edge cases
  - _Requirements: 3.4_

- [x] 5.2 Create comprehensive error recovery testing

  - Test system behavior when metrics exist but return no data
  - Validate recovery when missing metrics become available
  - Test network failure scenarios and retry mechanisms
  - Implement and test query timeout handling with appropriate fallbacks
  - _Requirements: 3.2, 3.3_

## Task Group 6: Alert Firing Validation

- [x] 6. Implement synthetic metric generation for alert testing

  - Create Prometheus client integration for generating controlled error conditions
  - Implement traffic pattern generation that exceeds calculated dynamic thresholds
  - Add metric cleanup and reset functionality for test isolation
  - Create test scenarios for both precision (no false alerts) and recall (catches real issues)
  - _Requirements: 4.1, 4.2_

- [x] 6.1 Create end-to-end alert pipeline validation

  - Test alert firing in AlertManager when dynamic thresholds are exceeded
  - Validate alert timing matches expected sensitivity levels
  - Compare dynamic vs static alert behavior with identical error conditions
  - Test alert clearing when conditions resolve
  - _Requirements: 4.3, 4.4, 4.6_

- [x] 6.2 Implement alert precision and recall testing framework

  - Create controlled scenarios where alerts should fire (recall testing)
  - Create controlled scenarios where alerts should NOT fire (precision testing)
  - Validate dynamic alerts demonstrate improved sensitivity AND specificity vs static
  - Document alert behavior characteristics and thresholds
  - _Requirements: 4.2, 4.3_

## Task Group 7: Recording Rules, Alert Rules, and Query Validation and Optimization

- [ ] 7. Validate and optimize recording rules, alert rules, and UI queries for all indicator types

- [x] 7.1 Validate recording rules generation for all indicator types

  - Test recording rules creation for ratio, latency, latencyNative, and boolGauge indicators
  - Verify recording rules produce correct metrics for both static and dynamic SLOs
  - Validate recording rule queries use efficient aggregations and proper label handling
  - Test recording rules work correctly across different time windows and SLO targets
  - _Requirements: 5.1, 5.3_

- [x] 7.1.1 CRITICAL: Fix generic recording rules generation and UI data display

  - Investigate why generic recording rules (pyrra_availability, pyrra_requests:rate5m, pyrra_errors:rate5m) are missing for most SLOs
  - Fix UI main page showing "no data" for availability and budget columns (regression from task 6)
  - Correct detail pages showing incorrect "100%" for availability and error budget when errors exist
  - Ensure all indicator types (ratio, latency, boolGauge) generate proper generic rules for UI display
  - Validate generic rules generation works for both static and dynamic SLOs
  - Test complete UI data flow from recording rules to display components
  - _Requirements: 5.1, 5.3_

- [x] 7.2 CRITICAL: Mathematical Correctness Validation (Simple Check)

  - **Pick 2-3 recording rules and manually verify they produce correct values**
    - Use simple `python -c "..."` commands to calculate expected values using exact formulas
    - Compare calculated values with what Prometheus shows for those recording rules. Check both the time series values ​​and the time window lengths
    - Test one ratio SLO and one latency SLO to cover main indicator types
  - **Check UI calculations match expected values**
    - Verify BurnRateThresholdDisplay shows values that match manual calculations
    - Test with both high traffic (above average) and low traffic (below average) scenarios
  - **Simple testing approach**
    - Use existing test SLOs (test-dynamic-apiserver, test-latency-dynamic)
    - Guide user to check values in Prometheus UI and Pyrra UI for comparison
    - Use simple Python scripts for ground truth calculations, no complex tools
  - **Consult user before completion**: Ask if there are more components or cases to check
  - _Requirements: 5.1, 5.3_

- [x] 7.3 CRITICAL: Fix Query Aggregation (Single Series Results)

  - **Check recording rules use proper sum() aggregation**
    - Look at 2-3 recording rules and verify they use `sum()` to aggregate multi-series metrics
    - Test with test-dynamic-apiserver (base metric has 74 series)
    - **UPDATED EXPECTATION**: Burn rate recording rules should return 1 series, but increase recording rules (e.g., `apiserver_request:increase30d`) intentionally return multiple series (grouped by labels like `code`) for UI RequestsGraph component
    - Verified: Burn rate rules return 1 series, increase rules return 4 series (expected for UI grouping)
  - **Check UI queries return single series per SLO**
    - Test BurnRateThresholdDisplay queries return exactly 1 series (not 74 like raw apiserver_request_total)
    - Verify alert rules aggregate to single series per SLO per alert window
    - Verified: BurnRateThresholdDisplay uses `sum()` aggregation, alert rules use `sum()` for traffic calculation
  - **Simple testing approach**
    - Created test script `cmd/test-query-aggregation/main.go` to verify query aggregation
    - Verified with curl commands and Prometheus API
    - All 7 tests passed with correct expectations
  - **Consult user before completion**: Confirmed no additional components need checking
  - _Requirements: 5.1, 5.3_

- [x] 7.4 CRITICAL: Fix UI Number Truncation (Add Scientific Notation)

  - **Fix truncated numbers in UI components**
    - Add scientific notation to AlertsTable short burn and long burn columns when numbers are very small
    - Fix BurnRateThresholdDisplay to show scientific notation for very small thresholds
    - Add scientific notation to any graphs that show small threshold values
    - Implement simple rule: if number < 0.001, show scientific notation (e.g., 1.23e-5)
  - **Test with high SLO targets that produce small numbers**
    - Test with 99.99% SLO target that produces very small thresholds
    - Create simple test SLO with high target to verify scientific notation works
    - Guide user to check UI displays scientific notation correctly in Pyrra UI
  - **Simple testing approach**
    - Use existing SLOs and create one high-target test SLO
    - Guide user to check number display in Pyrra UI (port 3000): "Do you see scientific notation for small numbers?"
    - Use simple Python scripts to generate test values and verify formatting
  - **Consult user before completion**: Ask if there are more UI components or cases to check
  - _Requirements: 5.1, 5.3_

- [x] 7.5 Validate alert rules generation and end-to-end alert pipeline

  - Test alert rules creation for ratio, latency, latencyNative, and boolGauge indicators
  - Verify alert rules reference correct recording rules (not raw metrics) when available
  - Validate alert rule expressions produce correct threshold calculations
  - Test alert rules fire correctly under controlled error conditions using existing `cmd/run-synthetic-test/main.go`
  - Test complete end-to-end alert pipeline from Prometheus rules to AlertManager
  - _Requirements: 5.1, 5.3_

- [x] 7.6 Investigate and fix UI refresh rate if needed

  - Compare Detail.tsx auto-reload behavior with upstream-comparison branch
  - Check if refresh interval was modified during dynamic burn rate feature development
  - If behavior changed from original: revert to original refresh rate configuration
  - If behavior is original: document as expected behavior, no changes needed
  - Test that refresh rate is appropriate for production use
  - _Requirements: 5.2, 5.4_

- [x] 7.7 Fix BurnrateGraph to display dynamic thresholds for dynamic SLOs

  - Update BurnrateGraph component to detect dynamic burn rate type from objective
  - Integrate traffic calculation patterns from BurnRateThresholdDisplay component
  - Calculate dynamic threshold using (N_SLO / N_alert) × E_budget_percent × (1 - SLO_target) formula
  - Update threshold line in graph to show calculated dynamic threshold instead of static
  - Update getThresholdDescription() to provide meaningful description for dynamic thresholds
  - Maintain backward compatibility for static SLOs (existing behavior unchanged)
  - Test with both ratio and latency dynamic SLOs to verify correct threshold display
  - _Requirements: 2.1, 2.4_

- [x] 7.8 Design Grafana dashboard enhancements for dynamic burn rates

  - Review existing Grafana dashboard structure (list.json, detail.json)
  - Analyze current generic recording rules used by dashboards
  - Identify what dynamic burn rate information should be displayed in Grafana
  - Design dashboard panels for dynamic burn rate visualization
  - Determine if new generic recording rules are needed for dynamic SLOs
  - Document dashboard enhancement design and required changes
  - Ensure backward compatibility with static SLOs
  - **COMPLETED**: Analysis documented in `.dev-docs/TASK_7.8_GRAFANA_DASHBOARD_DESIGN.md`
  - **FINAL DECISION**: No changes needed - dashboards already support dynamic SLOs
  - **KEY FINDINGS**:
    - Generic recording rules are IDENTICAL for static and dynamic SLOs
    - Dashboards already display availability and error budget correctly for both types
    - Grafana has NO alerting information (by design) - consistent with current approach
    - Pyrra UI is the proper tool for burn rate and alert analysis
    - No backend changes required, no dashboard modifications required
  - _Requirements: 6.1, 6.2_

- [x] 7.9 Test and validate Grafana dashboard compatibility with dynamic burn rates

  - **SCOPE CHANGE**: No implementation needed - dashboards already support dynamic SLOs
  - **REFERENCE**: Follow testing plan in `.dev-docs/TASK_7.8_GRAFANA_DASHBOARD_DESIGN.md`
  - Execute Test Scenario 1: Static SLO with generic rules
  - Execute Test Scenario 2: Dynamic SLO with generic rules
  - Execute Test Scenario 3: Mixed static and dynamic SLOs
  - Execute Test Scenario 4: Backward compatibility validation
  - Complete validation checklist from design document
  - Update examples/grafana/README.md with compatibility documentation
  - Document that no dashboard changes are required
  - **KEY FINDING**: Generic rules are identical for static and dynamic SLOs
  - _Requirements: 6.1, 6.2_

- [ ] 7.10 Optimize UI component queries and performance validation

  - **Phase 1: Analysis and Validation** ✅ COMPLETE

    - Analyzed current BurnRateThresholdDisplay implementation (uses raw metrics)
    - Validated recording rules exist and provide 40x speedup for ratio indicators
    - Created validation tools (validate-ui-query-optimization, test-burnrate-threshold-queries)
    - Documented performance benchmarks and optimization strategy
    - See: `.dev-docs/TASK_7.10_UI_QUERY_OPTIMIZATION_ANALYSIS.md`

  - **Phase 2: Implementation** 🔜 NEW SUB-TASKS REQUIRED

    - Sub-task 7.10.1: Fix validation tests to use only existing recording rules (30d window)
    - Sub-task 7.10.2: Implement BurnRateThresholdDisplay optimization to use recording rules
    - Sub-task 7.10.3: Verify and optimize backend alert rule queries (if needed)
    - Sub-task 7.10.4: Test optimization with all indicator types and update documentation

  - _Requirements: 5.1, 5.3_

- [x] 7.10.1 Fix validation tests and establish correct baseline

  - **Fix test queries**: Use only 30d window comparisons (has data)
    - Compare: `apiserver_request:increase30d` vs `increase(apiserver_request_total[30d])`
    - Remove tests for alert windows (1h4m, 6h26m) that have no data yet
  - **Update validation tools**:
    - Edit `cmd/validate-ui-query-optimization/main.go` to use only 30d window
    - Edit `cmd/test-burnrate-threshold-queries/main.go` to use only 30d window
    - Rebuild tools: `go build -o validate-ui-query-optimization.exe ./cmd/validate-ui-query-optimization`
  - **Run corrected tests** with Prometheus and document actual performance
    - Execute: `./validate-ui-query-optimization.exe`
    - Execute: `./test-burnrate-threshold-queries.exe`
  - **Update analysis documents** with correct test results:
    - Fix `.dev-docs/TASK_7.10_VALIDATION_RESULTS.md` with real measurements
    - Update `.dev-docs/TASK_7.10_UI_QUERY_OPTIMIZATION_ANALYSIS.md` with validated findings
  - **Document recording rule availability**: Which exist, which have data, which need time
  - _Requirements: 5.1, 5.3_

- [x] 7.10.2 Implement BurnRateThresholdDisplay optimization

  - **Prerequisite**: Task 7.10.1 complete ✅ - Validation shows 7.17x speedup for ratio, 2.20x for latency
  - **Reference documents**:
    - `.dev-docs/TASK_7.10_VALIDATION_RESULTS.md` - Performance measurements and findings
    - `.dev-docs/TASK_7.10_UI_QUERY_OPTIMIZATION_ANALYSIS.md` - Implementation strategy
    - `.dev-docs/TASK_7.10.1_TEST_IMPROVEMENTS.md` - Test methodology and results
  - **Key findings from validation**:
    - Ratio indicators: 48.75ms → 6.80ms (7.17x speedup) ✅ HIGH PRIORITY
    - Latency indicators: 6.34ms → 2.89ms (2.20x speedup) ✅ IMPLEMENT
    - BoolGauge indicators: 3.02ms (already fast) ❌ SKIP OPTIMIZATION
    - Only SLO window has recording rules (not alert windows)
    - Hybrid approach required: recording rule for SLO window + inline for alert windows
  - **Implementation** in `ui/src/components/BurnRateThresholdDisplay.tsx`:
    - Add `getTrafficRatioQueryOptimized()` function for hybrid query generation
    - Add `getBaseMetricName()` helper (strip \_total, \_count, \_bucket suffixes)
    - Update `getTrafficRatioQuery()` to use recording rules for SLO window, inline for alert windows
    - Add fallback to raw metrics if recording rules unavailable
  - **Correct query pattern** (hybrid approach):
    - SLO window: `sum({metric}:increase{sloWindow}{slo="..."})` (use recording rule)
    - Alert window: `sum(increase({metric}[{alertWindow}]))` (use inline calculation)
    - Combined: `sum({metric}:increase30d{slo="..."}) / sum(increase({metric}[1h4m]))`
  - **Implementation priority**:
    - ✅ Implement for ratio indicators (7x speedup)
    - ✅ Implement for latency indicators (2x speedup)
    - ❌ Skip boolGauge optimization (already fast, no benefit)
  - **Test with indicator types**: ratio, latency (skip boolGauge per validation findings)
  - **Verify performance improvement** using `validate-ui-query-optimization.exe` from 7.10.1
  - **Update documentation**: Document optimization in `.dev-docs/TASK_7.10_IMPLEMENTATION.md`
  - _Requirements: 5.1, 5.3_

- [x] 7.10.3 Review backend alert rule query optimization

  - **Reference documents**:
    - `.dev-docs/TASK_7.10_VALIDATION_RESULTS.md` - Performance measurements showing 7x speedup for ratio, 2x for latency
    - `.dev-docs/TASK_7.10_IMPLEMENTATION.md` - UI optimization implementation and real-world performance analysis
    - `.dev-docs/TASK_7.10.1_TEST_IMPROVEMENTS.md` - Test methodology and recording rule availability
  - **Context from 7.10.2**: UI optimization implemented with hybrid approach (recording rules for SLO window)
  - **Key finding**: Query speedup (7x) provides minimal UI benefit (~5% of 110ms total) due to network overhead
  - **Primary benefit**: Prometheus load reduction, not query speed
  - Check if alert rules in slo/rules.go use raw metrics in dynamic threshold calculation
  - Current: `scalar((sum(increase(metric[30d])) / sum(increase(metric[1h4m]))) * threshold)`
  - **IMPORTANT**: Alert windows (1h4m, 6h26m) do NOT have increase recording rules
  - Potential optimization: `scalar((sum(metric:increase30d{slo="..."}) / sum(increase(metric[1h4m]))) * threshold)`
  - Only SLO window can use recording rule, alert windows must use inline increase()
  - Evaluate if optimization is needed or if current approach is acceptable
  - Consider: Alert rules evaluated every 30s (different performance profile than UI on-demand queries)
  - Consider: Main benefit would be Prometheus load reduction, not alert evaluation speed
  - Document decision and rationale in `.dev-docs/TASK_7.10.3_BACKEND_OPTIMIZATION_DECISION.md`
  - _Requirements: 5.1, 5.3_

- [x] 7.10.4 Final validation and documentation cleanup

  - **Prerequisites**: All 7.10 sub-tasks complete ✅
    - Task 7.10.1: Test improvements and validation methodology ✅
    - Task 7.10.2: UI optimization implementation ✅
    - Task 7.10.3: Backend optimization decision ✅
    - Task 7.10.5: Backend optimization implementation ✅
  - **Reference documents** (completed during 7.10 series):
    - `.dev-docs/TASK_7.10_VALIDATION_RESULTS.md` - Performance measurements from 7.10.1
    - `.dev-docs/TASK_7.10.1_TEST_IMPROVEMENTS.md` - Test methodology and tools
    - `.dev-docs/TASK_7.10_UI_QUERY_OPTIMIZATION_ANALYSIS.md` - Original analysis
    - `.dev-docs/TASK_7.10_IMPLEMENTATION.md` - UI implementation details and real-world analysis from 7.10.2
    - `.dev-docs/TASK_7.10.3_BACKEND_OPTIMIZATION_DECISION.md` - Backend decision rationale from 7.10.3
    - `.dev-docs/TASK_7.10.5_BACKEND_IMPLEMENTATION.md` - Backend implementation details from 7.10.5
    - `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md` - Overall feature status (updated throughout)
  - **Implementation summary**:
    - ✅ UI optimization: BurnRateThresholdDisplay uses recording rules for SLO window
    - ✅ Backend optimization: Alert rules use recording rules for SLO window
    - ✅ Hybrid approach: Recording rules for SLO window + inline for alert windows
    - ✅ Bug fix: Dynamic SLO window support (fixes synthetic SLOs with 1d window)
    - ✅ Performance: 7.17x speedup for ratio, 2.20x for latency indicators
    - ✅ Primary benefit: Prometheus CPU/memory load reduction at scale
  - **Validation status**: ✅ Core validation completed
    - Ratio indicators: 7.17x query speedup validated (48.75ms → 6.80ms)
    - Latency indicators: 2.20x query speedup validated (6.34ms → 2.89ms)
    - Real-world UI performance: ~110ms total (network overhead dominates, 5% improvement)
    - Backend alert rules: Optimized queries generated correctly
    - Synthetic SLOs: Threshold display fixed (dynamic window support)
  - **Final validation tasks**:
    - **Test all indicator types** with optimized implementation:
      - Ratio indicators: Verify threshold display and alert rules
      - Latency indicators: Verify threshold display and alert rules
      - BoolGauge indicators: Verify no optimization applied (as designed)
      - LatencyNative indicators: Verify fallback to raw metrics (as designed)
    - **Verify alert firing** with optimized backend queries:
      - Use `cmd/run-synthetic-test/main.go` to test synthetic SLOs
      - Confirm alerts fire correctly with recording rule queries
      - Validate alert timing matches expected behavior
    - **Performance validation** (optional re-validation):
      - Run `validate-ui-query-optimization.exe` to confirm speedups
      - Measure Prometheus CPU/memory usage before/after (if possible)
      - Verify no performance regressions
    - **Cross-validation**:
      - Verify UI and backend use same recording rules (consistency)
      - Confirm no duplicate calculations between UI and alerts
      - Check that threshold values match between UI tooltip and alert rules
  - **Documentation consolidation**:
    - Review all `.dev-docs/TASK_7.10*.md` documents for completeness
    - Update `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md` with final status
    - Consider creating `.dev-docs/TASK_7.10_FINAL_SUMMARY.md` if needed
    - Update `.kiro/steering/pyrra-development-standards.md` with optimization patterns
    - Document lessons learned and best practices
  - **Code cleanup** (if needed):
    - Remove any debug logging or temporary code
    - Ensure consistent code style across UI and backend
    - Add final code comments for future maintainers
  - **Success criteria**:
    - All indicator types display thresholds correctly (including synthetic SLOs)
    - Alert rules use optimized queries and fire correctly
    - Documentation is complete and accurate
    - No regressions in functionality or performance
    - Code is production-ready
  - _Requirements: 5.1, 5.3_

- [x] 7.10.5 Implement backend alert rule query optimization

  - **Prerequisite**: Task 7.10.3 complete ✅ - Decision documented to implement optimization
  - **Reference documents**:
    - `.dev-docs/TASK_7.10.3_BACKEND_OPTIMIZATION_DECISION.md` - Decision rationale and implementation strategy
    - `.dev-docs/TASK_7.10_VALIDATION_RESULTS.md` - Performance measurements (7x ratio, 2x latency)
    - `.dev-docs/TASK_7.10_IMPLEMENTATION.md` - UI implementation pattern to follow
  - **Priority**: MEDIUM-HIGH (Prometheus load reduction at scale)
  - **Expected benefits**:
    - Ratio indicators: 7.17x speedup (48.75ms → 6.80ms) = ~42ms saved per alert
    - Latency indicators: 2.20x speedup (6.34ms → 2.89ms) = ~3.5ms saved per alert
    - Production impact: ~1.77 million seconds/year saved for ratio indicators at scale
    - Primary benefit: Prometheus CPU/memory load reduction (not alert speed)
  - **Implementation in `slo/rules.go`**:
    - Add `getBaseMetricName()` helper function (similar to UI implementation)
    - Update `buildDynamicAlertExpr()` for ratio indicators to use hybrid approach
    - Update `buildDynamicAlertExpr()` for latency indicators to use hybrid approach
    - Skip boolGauge optimization (already fast, no benefit)
    - Use hybrid pattern: recording rule for SLO window + inline for alert windows
  - **Correct query pattern** (hybrid approach):
    - Current: `scalar((sum(increase(metric[30d])) / sum(increase(metric[1h4m]))) * threshold)`
    - Optimized: `scalar((sum(metric:increase30d{slo="..."}) / sum(increase(metric[1h4m]))) * threshold)`
    - Only SLO window uses recording rule, alert windows use inline increase()
  - **Testing requirements**:
    - Unit tests: Verify query generation produces correct PromQL
    - Integration tests: Deploy test SLOs and verify alert rules created correctly
    - Alert firing tests: Use `cmd/run-synthetic-test/main.go` to verify alerts fire
    - Performance validation: Measure Prometheus CPU/memory usage before/after
  - **Documentation**:
    - Document implementation in `.dev-docs/TASK_7.10.5_BACKEND_IMPLEMENTATION.md`
    - Update `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md` with completion status
    - Update steering documents with backend optimization patterns
  - **Success criteria**:
    - Alert rules use recording rules for SLO window calculation
    - Alert rules still fire correctly (validated with synthetic tests)
    - Prometheus load reduced (measured)
    - No regressions in alert behavior
  - _Requirements: 5.1, 5.3_

- [x] 7.11 Create production readiness testing infrastructure

  - Create SLO generator tool with window variation (7d, 28d, 30d)
  - Create performance monitoring tool
  - Create automated test scripts
  - Generate 50 test SLOs for medium scale testing
  - Create consolidated testing documentation
  - _Requirements: 5.2, 5.4_
  - _Reference: `.dev-docs/TASK_7.11_TESTING_INFRASTRUCTURE.md` for complete tool documentation_

- [ ] 7.11.1 Run automated performance tests

  - Run baseline performance test with current SLOs
  - Apply 50 test SLOs and run medium scale performance test
  - Apply 100 test SLOs and run large scale performance test
  - Collect and analyze performance metrics (API response time, memory usage, query load)
  - Document performance benchmarks and scaling characteristics
  - _Requirements: 5.2, 5.4_
  - _Reference: `.dev-docs/TASK_7.11_TESTING_INFRASTRUCTURE.md` for commands and tools_
  - _Deliverable: `.dev-docs/PRODUCTION_PERFORMANCE_BENCHMARKS.md` with test results_

- [ ] 7.12 Manual testing - Browser compatibility and graceful degradation

  - **Interactive Testing Required**: Test in Chrome, Firefox, and Edge browsers
  - **Follow Guide**: `.dev-docs/TASK_7.12_MANUAL_TESTING_GUIDE.md` (complete step-by-step instructions)
  - Test graceful degradation: network throttling, API failures, Prometheus unavailability
  - Test migration from static to dynamic SLOs with UI verification
  - Test rollback procedures
  - Create browser compatibility matrix document (`.dev-docs/BROWSER_COMPATIBILITY_MATRIX.md`)
  - Create migration guide document (`.dev-docs/MIGRATION_GUIDE.md`)
  - _Requirements: 5.2, 5.4_
  - _Reference: `.dev-docs/BROWSER_COMPATIBILITY_TEST_GUIDE.md` for detailed test scenarios_
  - _Note: This task requires human interaction for visual validation and browser testing_

- [ ] 7.13 Comprehensive UI build and deployment testing
  - Validate embedded UI build process (npm run build + make build)
  - Test production UI (port 9099) shows all enhancements correctly
  - Verify no regressions in existing static SLO functionality using upstream comparison branch
  - Test complete UI workflow from development to production deployment
  - _Requirements: 5.2_

## Task Group 8: Documentation and Migration Support

- [ ] 8. Create comprehensive documentation and migration support

- [ ] 8.1 Create troubleshooting and debugging documentation

  - Document common issues and resolution steps for dynamic burn rate setup
  - Create debugging guide for missing metrics and edge case scenarios
  - Document performance tuning guidelines and optimization strategies
  - Create migration guide for converting static to dynamic SLOs
  - _Requirements: 6.1, 6.2, 6.4_

- [ ] 8.2 Implement deployment validation testing
  - Test complete installation procedures from documentation
  - Validate deployment in production-like environments
  - Test migration procedures with real static SLO conversions
  - Create performance baseline documentation for different scales
  - _Requirements: 6.3, 6.5_

## Task Group 9: Comprehensive Regression Testing

- [ ] 9. Create full feature regression test suite with upstream comparison

- [ ] 9.1 Implement upstream comparison regression testing

  - Compare behavior between feature branch and upstream-comparison branch
  - Validate existing static SLO functionality remains identical to original Pyrra
  - Test that dynamic burn rate feature doesn't break core Pyrra functionality
  - Create automated comparison tests for UI components and backend behavior
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [ ] 9.2 Validate mathematical accuracy across all indicator types

  - Test all indicator types with both static and dynamic burn rates
  - Validate mathematical accuracy across all scenarios with real Prometheus data
  - Cross-validate calculations against known working examples
  - Test edge cases and boundary conditions for all indicator types
  - **GROUND TRUTH REGRESSION**: Ensure mathematical validation framework from Task 7.2 passes for all scenarios
  - **UNIQUENESS REGRESSION**: Verify query uniqueness validation from Task 7.3 continues to pass across all SLOs
  - **PRECISION REGRESSION**: Test scientific notation and precision handling from Task 7.4 works consistently
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [ ] 9.3 Comprehensive UI consistency and user experience testing

  - Test UI consistency and user experience across all indicator types
  - Validate tooltips, error handling, and performance displays work correctly
  - Test responsive design and cross-browser compatibility
  - Ensure consistent behavior between development and production UI builds
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [ ] 9.4 Production readiness validation

  - Conduct end-to-end testing in production-like environments
  - Validate performance characteristics meet production requirements
  - Test feature stability over extended periods (multi-day testing)
  - Create final production deployment checklist and validation procedures
  - _Requirements: 5.5, 6.5_

- [ ] 9.5 Finalize upstream contribution preparation
  - Ensure all code changes follow Pyrra project standards and conventions
  - Create comprehensive test evidence documentation for pull request
  - Validate feature works across different Kubernetes and Prometheus versions using upstream comparison
  - Prepare feature documentation for upstream integration
  - _Requirements: 6.5_
