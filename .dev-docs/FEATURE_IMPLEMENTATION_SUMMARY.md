# Dynamic Burn Rate Feature - Complete Implementation Summary

## Overview

The dynamic burn rate feature introduces adaptive alerting to Pyrra that adjusts alert thresholds based on actual traffic patterns rather than using fixed static multipliers. This implementation is based on the method described in the "Error Budget is All You Need" blog series.

## Core Concept & Formula

### Dynamic Burn Rate Calculation
```
dynamic_burn_rate = (N_SLO / N_alert) × E_budget_percent_threshold
```

Where:
- **N_SLO** = Number of events in the SLO window (e.g., 28 days)
- **N_alert** = Number of events in the alerting window (e.g., 1 hour, 6 hours)
- **E_budget_percent_threshold** = Constant percentage of error budget consumption to alert on

### Alert Threshold Calculation
```
alert_threshold = dynamic_burn_rate × (1 - SLO_target)
```

The PromQL comparison becomes:
```
error_rate > alert_threshold
```

### Error Budget Percent Thresholds (Constants)
These are **predefined constant values**, not calculated:

| Window Period | E_budget_percent_threshold | Reasoning |
|---------------|---------------------------|-----------|
| 1 hour        | 1/48                     | 50% budget consumption per day |
| 6 hours       | 1/16                     | 100% budget consumption per 4 days |
| 1 day         | 1/14                     | Balanced daily threshold |
| 4 days        | 1/7                      | Long-term budget management |

## Implementation Status

### ✅ Completed Components

#### 1. API & Type System (Complete)
- **`BurnRateType` field** added to `Alerting` struct in `slo/slo.go`
- **Kubernetes CRD support** in `servicelevelobjective_types.go`
- **Backward compatibility** with default "static" behavior
- **Type safety** with proper JSON marshaling

#### 2. Core Algorithm Infrastructure (Complete)
- **`DynamicWindows()` method**: Assigns predefined E_budget_percent_threshold constants to window periods
- **`dynamicBurnRateExpr()` method**: Generates PromQL expressions for dynamic calculations
- **Window period integration**: Uses existing window structure with dynamic factors

#### 3. Development Environment (Complete)
- **Minikube setup** with Prometheus, Grafana, and kube-prometheus-stack
- **Build pipeline** functional with all tests passing
- **Test configuration** available in `.dev/test-slo.yaml`
- **Documentation** comprehensive in `.dev-docs/` folder

### 🔧 Current Implementation Gap

**Critical Issue**: The dynamic burn rate calculation logic exists but **is not integrated into actual alert rule generation**.

**Root Cause**: The `Alerts()` method correctly calculates dynamic factors, but the generated alert rules still use `QueryBurnrate()` which applies traditional static burn rate logic instead of the dynamic threshold calculation.

## Next Steps - Prioritized Implementation Plan

### Priority 1: Core Alert Logic Integration 🚨
**Goal**: Make dynamic burn rate calculations actually work in alert rules

#### 1.1 Fix Alert Rule Generation Logic
- **File**: `slo/rules.go`
- **Task**: Modify alert rule generation to use dynamic expressions when `BurnRateType = "dynamic"`
- **Specific Changes**:
  - Update the `Alerts()` method to generate different PromQL for dynamic vs static
  - Implement proper threshold calculation: `dynamic_burn_rate × (1 - SLO_target)`
  - Ensure PromQL compares `error_rate` to the calculated threshold

#### 1.2 Update PromQL Generation
- **Task**: Modify PromQL templates to incorporate `(N_SLO / N_alert) × E_budget_percent_threshold` formula
- **Integration**: Connect `dynamicBurnRateExpr()` output to actual alert conditions
- **Validation**: Ensure generated PromQL is syntactically correct and logically sound

#### 1.3 Test Dynamic vs Static Behavior
- **Create test cases** comparing static vs dynamic alert generation
- **Validate PromQL output** for both modes
- **Verify threshold calculations** are mathematically correct

### Priority 2: Testing & Validation 🧪
**Goal**: Ensure reliability and correctness

#### 2.1 Unit Tests
- **Dynamic threshold calculation tests**
- **E_budget_percent_threshold constant validation**
- **PromQL generation tests** for both static and dynamic modes
- **Edge case handling** (zero traffic, high traffic spikes)

#### 2.2 Integration Tests
- **End-to-end SLO creation** with dynamic burn rates
- **Alert rule deployment** to Kubernetes/Prometheus
- **Alert firing verification** under different traffic conditions

#### 2.3 Traffic Pattern Testing
- **Low traffic scenarios** (validate dynamic adaptation)
- **High traffic scenarios** (ensure threshold scaling)
- **Traffic spike handling** (alert responsiveness)

### Priority 3: User Experience Enhancement 🎨
**Goal**: Make dynamic burn rates accessible and understandable

#### 3.1 UI Components
- **`BurnrateGraph.tsx`**: Visualize dynamic vs static thresholds
- **`AlertsTable.tsx`**: Display burn rate type and dynamic factors
- **Configuration controls**: Allow users to select dynamic vs static mode
- **Tooltips and help text**: Explain dynamic burn rate concepts

#### 3.2 Documentation
- **User guide**: When to use dynamic vs static burn rates
- **Migration examples**: Converting existing SLOs
- **Best practices**: Traffic volume considerations
- **Troubleshooting**: Common issues and solutions

### Priority 4: Observability & Operations 📊
**Goal**: Provide operational visibility into dynamic behavior

#### 4.1 Grafana Dashboard Updates
- **Dynamic threshold visualization** in burn rate panels
- **Traffic volume correlation** with alert thresholds
- **Comparison dashboards** showing static vs dynamic behavior
- **Debug panels** for understanding dynamic calculations

#### 4.2 Recording Rules & Efficiency
- **Pre-compute traffic ratios** for efficiency
- **Recording rules** for complex dynamic calculations
- **Performance optimization** for high-cardinality metrics

#### 4.3 Monitoring & Alerting
- **Meta-alerts** for dynamic burn rate calculation failures
- **Threshold behavior monitoring** 
- **Performance metrics** for dynamic vs static overhead

## Technical Architecture

### Current File Structure
```
slo/
├── slo.go              # Core types including Alerting.BurnRateType
├── rules.go            # Alert generation logic (needs integration)
└── rules_test.go       # Test cases (needs dynamic tests)

kubernetes/api/v1alpha1/
└── servicelevelobjective_types.go  # CRD with BurnRateType field

.dev-docs/              # Implementation documentation
├── dynamic-burn-rate.md
├── burn-rate-analysis.md
├── dynamic-burn-rate-implementation.md
└── FEATURE_IMPLEMENTATION_SUMMARY.md  # This file
```

### Key Methods Status
| Method | Status | Purpose |
|--------|--------|---------|
| `DynamicWindows()` | ✅ Complete | Assigns E_budget_percent_threshold constants |
| `dynamicBurnRateExpr()` | ✅ Complete | Generates dynamic PromQL expressions |
| `Alerts()` | 🔧 Needs Integration | Alert rule generation (not using dynamic yet) |
| `QueryBurnrate()` | 🔧 Needs Update | PromQL generation (still static only) |

## Implementation Notes

### Formula Implementation Details
```go
// In DynamicWindows() - these are CONSTANTS, not calculations
var errorBudgetBurnPercent float64
switch {
case w.Long == time.Hour:
    errorBudgetBurnPercent = 1.0 / 48  // E_budget_percent_threshold
case w.Long == 6*time.Hour:
    errorBudgetBurnPercent = 1.0 / 16  // E_budget_percent_threshold
case w.Long == 24*time.Hour:
    errorBudgetBurnPercent = 1.0 / 14  // E_budget_percent_threshold
case w.Long == 4*24*time.Hour:
    errorBudgetBurnPercent = 1.0 / 7   // E_budget_percent_threshold
}
```

### PromQL Integration Pattern
```promql
# Current (static): burn_rate > static_factor * (1 - slo_target)
# Target (dynamic): error_rate > ((N_slo / N_alert) * E_budget_percent_threshold) * (1 - slo_target)
```

## Success Criteria

### Functional Requirements
- [x] API supports both "static" and "dynamic" burn rate types
- [ ] Dynamic SLOs generate mathematically correct alert thresholds
- [ ] Alert firing behavior adapts to traffic volume changes
- [ ] Backward compatibility maintained for existing static SLOs

### Performance Requirements
- [ ] Dynamic calculations don't significantly impact rule evaluation time
- [ ] PromQL queries remain efficient at scale
- [ ] Memory usage remains reasonable for high-cardinality metrics

### User Experience Requirements
- [ ] Clear documentation explaining dynamic vs static trade-offs
- [ ] UI clearly indicates which mode is active
- [ ] Migration path from static to dynamic is straightforward
- [ ] Troubleshooting information is accessible

## Risk Assessment

### Technical Risks
- **PromQL Complexity**: Dynamic queries may be more resource-intensive
- **Edge Cases**: Zero or very low traffic scenarios need special handling
- **Backward Compatibility**: Changes must not break existing deployments

### Mitigation Strategies
- **Comprehensive Testing**: Cover edge cases and performance scenarios
- **Feature Flags**: Allow gradual rollout and easy rollback
- **Documentation**: Clear guidance on when to use each mode
- **Monitoring**: Track performance impact of dynamic calculations

---

**Status**: Implementation foundation complete, core integration in progress
**Last Updated**: August 24, 2025
**Next Review**: After Priority 1 completion
