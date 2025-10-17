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

- [x] 7.11.1 Run automated performance tests

  - Run baseline performance test with current SLOs
  - Apply 50 test SLOs and run medium scale performance test
  - Apply 100 test SLOs and run large scale performance test
  - Collect and analyze performance metrics (API response time, memory usage, query load)
  - Document performance benchmarks and scaling characteristics
  - _Requirements: 5.2, 5.4_
  - _Reference: `.dev-docs/TASK_7.11_TESTING_INFRASTRUCTURE.md` for commands and tools_
  - _Deliverable: `.dev-docs/PRODUCTION_PERFORMANCE_BENCHMARKS.md` with test results_

- [x] 7.12 Manual testing - Browser compatibility and graceful degradation

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

- [x] 7.12.1 CRITICAL: Fix white page crash for dynamic SLOs with missing metrics

  - **Priority**: HIGH - Blocker for production use with missing/broken metrics
  - **Issue**: Clicking burn rate graph button for dynamic SLOs with missing metrics causes complete page crash (white screen)
  - **Root Cause**: `BurnrateGraph.tsx:284` calls `Array.from()` on undefined data when dynamic SLO has no metric data
  - **Error**: `TypeError: undefined is not iterable (cannot read property Symbol(Symbol.iterator))`
  - **Related Errors**:
    - `[BurnRateThresholdDisplay] No data returned for boolGauge indicator traffic query`
    - `POST http://localhost:9099/objectives.v1alpha1.ObjectiveService/GraphDuration 404 (Not Found)` (for latency SLOs)
  - **Implementation**:
    - Add null/undefined checks in `BurnrateGraph.tsx` before calling `Array.from()`
    - Implement graceful error handling for missing graph data
    - Display user-friendly error message instead of crashing (e.g., "No data available for this time range")
    - Ensure consistent error handling between static and dynamic SLOs
    - Test with all indicator types (ratio, latency, latencyNative, boolGauge) with missing metrics
  - **Testing**:
    - Test with dynamic SLOs that have completely missing metrics
    - Test with dynamic SLOs that have broken/non-existent metrics
    - Verify static SLOs with missing metrics continue to work (no regression)
    - Verify working dynamic SLOs are not affected
  - **Documentation**:
    - Update `.dev-docs/BROWSER_COMPATIBILITY_MATRIX.md` with fix status
    - Document error handling patterns in `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md`
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 5.2_
  - _Discovered during: Task 7.12 browser compatibility testing_

- [x] 7.13 Comprehensive UI build and deployment testing

  - **Reference Documents**:
    - `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md` - Embedded UI build workflow documentation
    - `.dev-docs/TASK_7.6_UI_REFRESH_RATE_INVESTIGATION.md` - Upstream comparison methodology
    - `.dev-docs/BROWSER_COMPATIBILITY_MATRIX.md` - Browser testing results
  - **Already Completed** ✅:
    - Embedded UI build process validated (npm run build + make build workflow documented)
    - Production UI (port 9099) tested extensively with all enhancements
    - Complete UI workflow from development to production documented and validated
    - Browser compatibility testing completed (Chrome, Firefox)
  - **Remaining Work** (Focus Areas):
    - **Systematic regression testing against upstream-comparison branch**:
      - Compare static SLO behavior between feature branch and upstream-comparison
      - Verify no unintended changes to existing Pyrra functionality
      - Document any intentional differences vs regressions
      - Test scenarios: static SLOs only, mixed static/dynamic, edge cases
    - **Final production build validation**:
      - Build UI with all recent fixes (Task 7.12.1 changes)
      - Rebuild Go binary with embedded UI
      - Test on port 9099 to verify all fixes work in production build
      - Verify missing metrics handling works in embedded UI
  - **Testing Checklist**:
    - [ ] Checkout upstream-comparison branch and start services
    - [ ] Test static SLO list page behavior (sorting, display, tooltips)
    - [ ] Test static SLO detail page (graphs, alerts, tiles)
    - [ ] Document baseline behavior
    - [ ] Switch to feature branch and compare behavior
    - [ ] Document any differences (expected vs regression)
    - [ ] Build production UI with Task 7.12.1 fixes
    - [ ] Test embedded UI (port 9099) with missing metrics scenarios
  - _Requirements: 5.2_

## Task Group 8: Upstream Integration Preparation

**Context**: Tasks 1-7 are complete with comprehensive testing and documentation in `.dev-docs/`. This task group focuses on preparing the feature for upstream contribution by organizing files, updating production documentation, and creating the pull request.

**Reference Guide**: See `.dev-docs/UPSTREAM_CONTRIBUTION_PLAN.md` for detailed file organization strategy, merge procedures, documentation guidelines, and PR template.

**Note**: Most documentation and testing work is already complete. Focus is on file organization and upstream integration.

- [ ] 8. Prepare repository for upstream contribution

- [x] 8.0 Pre-merge code cleanup and review - ✅ COMPLETE

  - **Status**: ✅ Complete - See `.dev-docs/TASK_8_CLEANUP_AND_PREPARATION.md` for comprehensive documentation
  - **Reference Documents**:

    - `.dev-docs/TASK_8_CLEANUP_AND_PREPARATION.md` - Consolidated cleanup and preparation guide (PRIMARY)
    - `.dev-docs/TASK_8.0_PRE_MERGE_CLEANUP_CHECKLIST.md` - Detailed checklist (reference)
    - `.dev-docs/TASK_8.0_CLEANUP_ANALYSIS.md` - File categorization analysis (reference)
    - `.dev-docs/TASK_8.0_CLEANUP_SUMMARY.md` - Actions completed summary (reference)

  - **Actions Completed**:

    - ✅ Reverted unintended changes (pyrra-kubernetesDeployment.yaml, ui/public/index.html, filesystem.go)
    - ✅ Removed unused code (~47 lines from slo/slo.go and CRD types)
    - ✅ Updated comment format in slo/rules.go ("originally X for Y" → "X for Y SLO period")
    - ✅ Moved ui/DYNAMIC_BURN_RATE_UI.md to .dev-docs/HISTORICAL_UI_DESIGN.md
    - ✅ Updated CONTRIBUTING.md with ui/README.md reference
    - ✅ Kept main.go native histogram changes (API server emits test metric for LatencyNative testing)
    - ✅ Kept BurnRateThresholdDisplay.spec.tsx tests (good practice)
    - ✅ Kept Toggle.tsx readOnly fix (React best practice)
    - ✅ All tests passing, code compiles successfully

  - **Architecture Understanding**:

    - ✅ Clarified that `connect_server_requests_duration_seconds` is a test metric emitted by API server only
    - ✅ Backend modes (kubernetes.go, filesystem.go) do NOT need test metric emission
    - ✅ LatencyNative feature works correctly - API server always runs in typical deployments
    - ✅ Test metric available for examples because `./pyrra api` is always running

  - **Deferred Items**:

    - ⏳ .gitignore cleanup - Will be handled before dev-tools-and-docs branch creation
    - ⏳ Move examples from .dev/ to examples/ - Moved to Task 8.2
    - ⏳ Duplicate selector functions - Kept as-is (logic differs per indicator type, consolidation not beneficial)
    - ⏳ Filesystem mode testing - Optional, can be added to Task 9.3 if desired

  - _Requirements: 6.5_

- [x] 8.1 Fetch and merge from upstream repository

  - **Reference**: See `.dev-docs/TASK_8_CLEANUP_AND_PREPARATION.md` Section "Task 8.1" for detailed steps
  - Add upstream remote if not already configured: `git remote add upstream https://github.com/pyrra-dev/pyrra.git`
  - Fetch latest upstream changes: `git fetch upstream`
  - Merge upstream/main into feature branch: `git merge upstream/main`
  - Resolve any merge conflicts that arise
  - Test that feature still works after merge:
    - `go build -o pyrra .`
    - `go test ./slo -run "TestObjective_DynamicBurnRate"`
    - `cd ui && npm run build`
  - Document any significant conflicts or changes required
  - _Requirements: 6.5_

- [x] 8.2 Move examples from .dev/ to examples/

  - **Reference**: See `.dev-docs/TASK_8_CLEANUP_AND_PREPARATION.md` Section "Task 8.2" for detailed steps
  - Review test SLOs in `.dev/` folder
  - Select best examples for production:
    - `test-dynamic-slo.yaml` → `examples/dynamic-burn-rate-ratio.yaml`
    - `test-latency-dynamic.yaml` → `examples/dynamic-burn-rate-latency.yaml`
    - Consider: LatencyNative and BoolGauge examples
  - Clean up examples:
    - Remove test-specific configurations
    - Add clear comments explaining dynamic burn rate usage
    - Ensure proper naming conventions
    - Add metadata and labels as appropriate
  - Update `examples/README.md` with dynamic burn rate explanation
  - _Requirements: 6.5_

- [ ] 8.3 Organize files for PR vs fork separation

  - **Create development branch**: Create new branch (e.g., `dev-tools-and-docs`) to preserve development artifacts
  - **Identify PR files**: Determine which files should go in pull request to upstream:
    - Core feature implementation (backend, UI, CRD changes)
    - Essential documentation updates (README.md, examples/)
    - Minimal necessary test files
  - **Identify fork-only files**: Determine which files stay in fork only:
    - Development tools in `cmd/` (validate-\*, test-\*, monitor-\*, generate-\*)
    - Development documentation in `.dev-docs/`
    - Development scripts in `scripts/`
    - Test configuration files in `.dev/`
    - Steering documents in `.kiro/`
    - Prompt documents in `prompts/`
  - **Clean temporary files**: Remove any temporary or build artifacts:
    - Compiled binaries (`.exe` files in root)
    - Build caches
    - IDE-specific files not in `.gitignore`
  - **Document file organization**: Create document explaining what goes where and why
  - _Requirements: 6.5_

- [x] 8.5 Update production documentation (keep concise and proportional)




  - **Review existing READMEs**: Examine current README.md and other production docs in upstream Pyrra
  - **Keep updates minimal and proportional**: Dynamic burn rate is ONE feature among many - don't overshadow existing content
  - **Identify documentation gaps**: Determine what needs to be added vs edited:
    - Main README.md: Add brief dynamic burn rate feature section (2-3 paragraphs max)
    - examples/README.md: Add concise dynamic SLO examples with short explanations
    - New docs: Consider separate `docs/DYNAMIC_BURN_RATE.md` ONLY if it keeps main docs clean
  - **Update documentation files**: Make necessary edits to production documentation
    - Keep additions concise - users can explore details through examples and testing
    - Focus on "what" and "how to use", not extensive "why" or implementation details
    - Maintain existing documentation structure and tone
  - **Add usage examples**: Include clear but minimal examples of dynamic burn rate configuration
  - **Migration guidance**: Add brief migration notes (1-2 sentences) for users converting static to dynamic SLOs
  - **IMPORTANT**: Extensive documentation already exists in `.dev-docs/` - extract ONLY essential user-facing content for production docs
  - **Goal**: Users should understand the feature exists, how to enable it, and where to find examples - not become documentation experts
  - _Requirements: 6.1, 6.2, 6.5_

- [x] 8.4 Investigate and resolve regex label selector behavior

  - **Context**: During Task 8.2 testing, discovered that SLOs with regex label selectors (e.g., `handler=~"/api.*"`) exhibit unexpected behavior with multiple SLOs created and data inconsistencies between main page and detail page
  - **Priority**: High - Affects production use cases with regex selectors
  - **Reference**: `.dev-docs/ISSUE_REGEX_LABEL_SELECTORS.md`

- [x] 8.4.1 Upstream comparison testing

  - **Test regex selectors on upstream**: Checkout `upstream-comparison` branch and test identical SLO configuration
  - **Document upstream behavior**: Record how many SLOs are created, recording rule structure, detail page functionality
  - **Compare with feature branch**: Identify if this is a regression or existing upstream behavior
  - **Test scenarios**:
    - Regex selector with static burn rate
    - Regex selector with dynamic burn rate
    - Regex selector with grouping field
    - Simple selector as control
  - **Document findings**: Create comparison document showing upstream vs feature branch behavior
  - **Reference**: `.dev-docs/ISSUE_REGEX_LABEL_SELECTORS.md` - Investigation plan
  - _Requirements: 3.1, 3.4_

- [x] 8.4.2 Root cause analysis

  - **Code analysis**: Review `slo/rules.go`, `slo/slo.go`, and controller logic
  - **Understand grouping behavior**: How does `grouping` field affect SLO instantiation?
  - **Recording rule scoping**: Are rules per-group or aggregated?
  - **Identify mismatch**: Where does main page vs detail page calculation diverge?
  - **Dynamic burn rate impact**: Does dynamic burn rate logic exacerbate the issue?
  - **Document architecture**: Clarify intended relationship between SLO YAML and SLO instances
  - **Reference**: `.dev-docs/ISSUE_REGEX_LABEL_SELECTORS.md` - Code analysis section
  - _Requirements: 3.1, 3.4_

- [x] 8.4.3 Solution implementation

  - **Based on 8.4.1 findings**: Choose appropriate solution approach
  - **Option A - Fix regression** (if upstream works):
    - Identify regression in feature branch code
    - Fix recording rule generation or SLO instantiation
    - Ensure detail page calculations match main page
  - **Option B - Document limitation** (if upstream also breaks):
    - Add warning in documentation about regex selectors
    - Recommend using simple selectors without grouping
    - Consider upstream contribution to fix the issue
  - **Option C - Implement aggregated approach** (if design change needed):
    - One YAML = One SLO (aggregated across grouping labels)
    - Recording rules aggregate across all label values
    - Detail page shows aggregated data
  - **Test solution**: Verify fix works for all indicator types
  - **Update examples**: Ensure examples follow best practices
  - **Reference**: `.dev-docs/ISSUE_REGEX_LABEL_SELECTORS.md` - Solution design section
  - _Requirements: 3.1, 3.4, 6.5_

- [ ] 8.6 Create pull request description and evidence

  - **Write PR description**: Create comprehensive pull request description including:
    - Feature overview and motivation (reference "Error Budget is All You Need" blog series)
    - Implementation summary (backend, API, UI changes)
    - Testing evidence (reference validation tools and results)
    - Breaking changes (if any - likely none)
    - Migration notes for existing users
  - **Compile test evidence**: Gather key testing results to include:
    - Mathematical validation results (Task 7.2 - `.dev-docs/TASK_7_2_MATH_VALIDATION_GUIDE.md`)
    - Query optimization results (Task 7.10 - `.dev-docs/TASK_7.10_COMPLETION_SUMMARY.md`)
    - UI testing results (Task 7.13 - `.dev-docs/TASK_7.13_COMPLETION_SUMMARY.md`)
    - Alert firing validation (Task 6 - `.dev-docs/TASK_6_LESSONS_LEARNED.md`)
    - Regression testing (Task 7.13 - zero regressions found)
  - **Create before/after examples**: Show clear examples of static vs dynamic behavior
  - **Document design decisions**: Explain key architectural choices
  - **Prepare for review**: Anticipate questions and prepare responses
  - _Requirements: 6.5_

## Task Group 9: Final Validation and Quality Assurance

**Context**: Most validation already completed in Tasks 1-7 (see `.dev-docs/TASK_7.13_COMPLETION_SUMMARY.md` for comprehensive regression testing results). This group focuses on final checks before PR submission.

**Note**: Task 7.13 already completed comprehensive regression testing with zero regressions found. These tasks are for final verification only.

- [ ] 9. Perform final validation checks before upstream contribution

- [ ] 9.1 Final regression verification

  - **Review Task 7.13 results**: Verify comprehensive regression testing completed successfully
  - **Spot-check key scenarios**: Quick validation of critical functionality:
    - Static SLO behavior unchanged (compare with upstream-comparison branch)
    - Dynamic SLO features working correctly
    - Mixed static/dynamic environments stable
  - **Verify test tools still work**: Run 1-2 validation tools from `cmd/` to confirm
  - **Document any new findings**: Note any issues discovered since Task 7.13
  - **Reference**: `.dev-docs/TASK_7.13_COMPLETION_SUMMARY.md` shows zero regressions found
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [ ] 9.2 Code quality and standards review

  - **Code style consistency**: Ensure code follows Pyrra project conventions
  - **Remove debug code**: Clean up any console.log, debug flags, or temporary code
  - **Comment quality**: Verify code comments are clear and helpful
  - **Test coverage**: Ensure adequate test coverage for new functionality
  - **Documentation accuracy**: Verify all code comments and docs match implementation
  - **Go formatting**: Run `gofumpt` on all Go files
  - **TypeScript/React**: Verify UI code follows existing patterns
  - _Requirements: 6.5_

- [ ] 9.3 Final production validation

  - **End-to-end smoke test**: Run complete workflow from SLO creation to alert firing
  - **Performance validation**: Verify performance meets expectations (reference Task 7.10 results)
  - **Error handling validation**: Test graceful degradation with missing metrics
  - **Cross-indicator validation**: Test all indicator types (ratio, latency, latencyNative, boolGauge)
  - **Filesystem mode validation** (if determined necessary in Task 8.0):
    - Reference Task 8.0 decision in `.dev-docs/TASK_8.0_PRE_MERGE_CLEANUP_CHECKLIST.md`
    - Test dynamic burn rate feature in filesystem mode (only kubernetes mode tested so far)
    - Verify filesystem.go changes work correctly
    - Document any filesystem-specific limitations or issues
    - If not tested, document as known limitation in PR description
  - **Reference existing validation**: Leverage comprehensive testing from Tasks 1-7
  - **Quick checklist**: Use `.dev-docs/TASK_7.13_QUICK_CHECKLIST.md` for final verification
  - _Requirements: 5.5, 6.5_

- [ ] 9.4 Prepare for upstream submission

  - **Review commit history**: Check if commits need squashing/organizing for clean history
  - **Update CHANGELOG**: Add entry for dynamic burn rate feature (if Pyrra uses changelog)
  - **Version considerations**: Note any version compatibility requirements
  - **Create PR branch**: Prepare clean branch for pull request from current feature branch
  - **Final review checklist**: Complete pre-submission checklist
  - **Backup development artifacts**: Ensure `dev-tools-and-docs` branch has all development files
  - _Requirements: 6.5_
