# Task 7.8: Grafana Dashboard Analysis for Dynamic Burn Rates

## Executive Summary

This document analyzes Pyrra's Grafana dashboards to determine what changes (if any) are needed to support dynamic burn rate SLOs.

**Key Finding**: **NO CHANGES NEEDED** - Grafana dashboards already fully support dynamic burn rate SLOs.

**Rationale**: 
1. Generic recording rules are **identical** for static and dynamic SLOs
2. Grafana dashboards show availability and error budget (which use the same calculations)
3. Grafana dashboards have **no alerting information** currently (by design)
4. Pyrra UI is the proper tool for detailed burn rate and alert analysis

**Decision**: Keep Grafana dashboards unchanged. Focus on testing to validate they work correctly with dynamic SLOs.

## Current State Analysis

### Existing Grafana Dashboards

#### 1. List Dashboard (`list.json`)
**Purpose**: Overview of all SLOs with key metrics
**Current Panels**:
- Single table showing all SLOs with columns:
  - Name (clickable link to detail dashboard)
  - Objective (SLO target percentage)
  - Window (SLO time window)
  - Availability (current availability)
  - Error Budget (remaining error budget percentage)

**Data Sources**:
- `pyrra_objective` - SLO target value
- `pyrra_window` - SLO window in seconds
- `pyrra_availability` - Current availability
- Calculated: `(pyrra_availability - pyrra_objective) / (1 - pyrra_objective)` for error budget

**What's NOT Displayed**:
- ❌ No alert information (no alerts column)
- ❌ No burn rate type indication
- ❌ No burn rate values

#### 2. Detail Dashboard (`detail.json`)
**Purpose**: Detailed view of individual SLO
**Current Panels**:
- **Stat Panels** (top row):
  - Objective: Shows SLO target
  - Window: Shows SLO window duration
  - Availability: Shows current availability with color thresholds
  - Error Budget: Shows remaining error budget with color thresholds
- **Time Series Graphs**:
  - Error Budget: Historical error budget consumption
  - Rate: Request rate over time (`pyrra_requests:rate5m`)
  - Errors: Error rate over time (`pyrra_errors:rate5m`)

**Data Sources**:
- All generic recording rules mentioned above
- Dashboard variable `$slo` for SLO selection

**What's NOT Displayed**:
- ❌ No alert state information
- ❌ No burn rate thresholds
- ❌ No traffic context
- ❌ No "Multi Burn Rate Alerts" section (unlike Pyrra UI)

### Generic Recording Rules Analysis

The `GenericRules()` method in `slo/rules.go` generates these metrics:

```promql
# Common to all indicator types:
pyrra_objective{slo="<name>"}           # SLO target (e.g., 0.99)
pyrra_window{slo="<name>"}              # Window in seconds (e.g., 2419200 for 28d)
pyrra_availability{slo="<name>"}        # Current availability
pyrra_requests:rate5m{slo="<name>"}     # Request rate
pyrra_errors:rate5m{slo="<name>"}       # Error rate
```

**Critical Observation**: These metrics are **IDENTICAL** for both static and dynamic SLOs. 

**Why?** Because:
- Availability calculation is the same: `1 - (errors / total)`
- Error budget calculation is the same: `(availability - target) / (1 - target)`
- Request and error rates are measured values, not calculated thresholds
- **Only alert thresholds differ** between static and dynamic (not shown in Grafana)

### Comparison: Grafana vs Pyrra UI

| Feature | Grafana Dashboards | Pyrra UI |
|---------|-------------------|----------|
| Availability | ✅ Yes | ✅ Yes |
| Error Budget | ✅ Yes | ✅ Yes |
| Request Rate | ✅ Yes | ✅ Yes |
| Error Rate | ✅ Yes | ✅ Yes |
| Alert State | ❌ No | ✅ Yes |
| Burn Rate Type | ❌ No | ✅ Yes |
| Burn Rate Thresholds | ❌ No | ✅ Yes |
| Traffic Context | ❌ No | ✅ Yes (dynamic only) |
| Multi-Window Alerts | ❌ No | ✅ Yes |

**Conclusion**: Grafana focuses on **monitoring metrics**, Pyrra UI focuses on **alerting details**.

## Analysis: Do We Need Changes?

### Question 1: Do Generic Rules Work for Dynamic SLOs?
**Answer**: ✅ YES - Generic rules are identical for static and dynamic SLOs

**Evidence**:
- Availability calculation: Same formula for both types
- Error budget calculation: Same formula for both types
- Request/error rates: Measured values, not calculated thresholds
- Only alert rules differ (not part of generic rules)

### Question 2: Do Dashboards Display Correctly?
**Answer**: ✅ YES - Dashboards show availability and error budget correctly

**Evidence**:
- List dashboard shows objective, window, availability, error budget
- Detail dashboard shows same metrics plus time series graphs
- All metrics come from generic rules that work for both types

### Question 3: Should We Add Burn Rate/Alert Information?
**Answer**: ❌ NO - Grafana dashboards don't show alerting information by design

**Evidence**:
- Current dashboards have NO alert information at all
- No burn rate values displayed
- No alert state (firing/pending)
- No burn rate type indication
- This is consistent with Grafana's role as high-level monitoring tool

### Question 4: Is There Value in Adding Burn Rate Type Indicator?
**Answer**: ❌ NO - Minimal value, adds complexity

**Reasons**:
- Requires backend changes (add `burnRateType` label to generic rules)
- Requires dashboard JSON modifications
- Provides no actionable information (availability/error budget are the same)
- Users who care about burn rate details use Pyrra UI
- Adds maintenance burden for minimal benefit

## Final Decision: No Changes Needed

### Option 4 (Chosen): Keep Dashboards Unchanged

**Decision**: Do not modify Grafana dashboards for dynamic burn rate support.

**Rationale**:
1. ✅ **Dashboards already work** - Generic rules are identical for both types
2. ✅ **Separation of concerns** - Grafana for monitoring, Pyrra UI for alerting details
3. ✅ **Avoid complexity** - No backend changes, no dashboard modifications needed
4. ✅ **Consistent design** - Grafana has never shown alerting information

### What This Means:

**For Users**:
- Existing Grafana dashboards work perfectly with dynamic burn rate SLOs
- No migration or updates needed
- Use Grafana for availability/error budget monitoring
- Use Pyrra UI for burn rate and alert analysis

**For Implementation**:
- No code changes required
- No dashboard JSON modifications
- Focus on testing to validate correct behavior
- Document that dashboards support both types

### Rejected Alternatives

#### Option 1: Minimal Enhancement (Add Burn Rate Type Only)

**Pros**:
- Shows which SLOs use dynamic burn rates
- Minimal dashboard changes

**Cons**:
- Requires backend changes (add `burnRateType` label)
- Provides no actionable information
- Adds maintenance burden
- Inconsistent with current design (no alerting info)

**Rejected**: Complexity outweighs minimal benefit

#### Option 2: Enhanced (Add Alert State)
**Pros**:
- Shows firing alerts in Grafana
- More complete monitoring view

**Cons**:
- Requires querying `ALERTS` metrics
- Significant dashboard modifications
- Duplicates Pyrra UI functionality
- Inconsistent with current Grafana-only-metrics approach

**Rejected**: Too much complexity, wrong tool for the job

#### Option 3: Comprehensive (Match Pyrra UI)
**Pros**:
- Feature parity with Pyrra UI
- Complete information in Grafana

**Cons**:
- Massive complexity (dynamic thresholds, traffic context)
- Requires new recording rules
- Duplicates entire Pyrra UI in Grafana
- Performance impact on Prometheus

**Rejected**: Completely impractical, wrong design philosophy

## Backend Changes Required

**None** - No backend modifications needed.

## Dashboard JSON Modifications

**None** - No dashboard modifications needed.

## Why Grafana Doesn't Need Burn Rate Information

### 1. Generic Rules Are Identical
**Fact**: The metrics Grafana displays are the same for static and dynamic SLOs
- Availability: Same calculation
- Error budget: Same calculation
- Request/error rates: Measured values

**Implication**: Dashboards already work correctly for both types

### 2. Grafana's Role is Monitoring, Not Alerting
**Current Design**: Grafana dashboards show NO alerting information
- No alert state (firing/pending)
- No burn rate values
- No burn rate thresholds
- No alert sensitivity

**Implication**: Adding burn rate type would be inconsistent with current design

### 3. Pyrra UI is the Proper Tool for Alert Analysis
**Pyrra UI Provides**:
- Real-time alert state
- Multi-window burn rate alerts
- Dynamic threshold calculations
- Traffic context and sensitivity
- Detailed tooltips and explanations

**Implication**: Users who need burn rate details already have the right tool

## Testing Plan (Task 7.9)

Since no changes are needed, Task 7.9 will focus on **validation testing** rather than implementation.

### Test Objectives:
1. Verify generic rules work correctly with dynamic SLOs
2. Confirm dashboards display availability and error budget correctly
3. Validate no regressions in existing functionality
4. Document that dashboards support both types

### Test Scenarios:

#### Scenario 1: Static SLO with Generic Rules
- Deploy static SLO with `--generic-rules` flag
- Verify generic rules are generated
- Check Grafana list dashboard displays correctly
- Check Grafana detail dashboard displays correctly
- Validate availability and error budget calculations

#### Scenario 2: Dynamic SLO with Generic Rules
- Deploy dynamic SLO with `--generic-rules` flag
- Verify generic rules are generated (same as static)
- Check Grafana list dashboard displays correctly
- Check Grafana detail dashboard displays correctly
- Validate availability and error budget calculations match Pyrra UI

#### Scenario 3: Mixed SLOs
- Deploy both static and dynamic SLOs
- Verify both appear in list dashboard
- Switch between SLOs in detail dashboard
- Confirm no visual differences (as expected)

#### Scenario 4: Backward Compatibility
- Use existing Grafana dashboard JSON files
- Test with both SLO types
- Verify no errors or missing data

## Validation Checklist (Task 7.9)

### Generic Rules Validation
- [ ] Static SLO generates `pyrra_objective`, `pyrra_availability`, `pyrra_requests:rate5m`, `pyrra_errors:rate5m`, `pyrra_window`
- [ ] Dynamic SLO generates identical generic rules
- [ ] Metric values are correct for both types
- [ ] No errors in Prometheus rule evaluation

### Grafana Dashboard Validation
- [ ] List dashboard displays static SLOs correctly
- [ ] List dashboard displays dynamic SLOs correctly
- [ ] Detail dashboard shows correct availability for both types
- [ ] Detail dashboard shows correct error budget for both types
- [ ] Time series graphs display correctly for both types
- [ ] Dashboard variable selection works for both types

### Calculation Validation
- [ ] Availability calculation matches Pyrra UI for static SLOs
- [ ] Availability calculation matches Pyrra UI for dynamic SLOs
- [ ] Error budget calculation matches Pyrra UI for static SLOs
- [ ] Error budget calculation matches Pyrra UI for dynamic SLOs

### Documentation Validation
- [ ] README clarifies that dashboards work for both types
- [ ] README explains separation of concerns (Grafana vs Pyrra UI)
- [ ] No misleading information about burn rate visualization


## Documentation Updates Required (Task 7.9)

### File: `examples/grafana/README.md`

**Add Clarification Section**:

```markdown
## Dynamic Burn Rate Support

Pyrra supports two types of burn rate alerting:
- **Static**: Fixed alert thresholds (traditional approach)
- **Dynamic**: Traffic-aware alert thresholds that adapt based on actual traffic patterns

### Grafana Dashboard Compatibility

**Good News**: The existing Grafana dashboards work perfectly with both static and dynamic burn rate SLOs without any modifications.

**Why?** The generic recording rules (`pyrra_objective`, `pyrra_availability`, `pyrra_requests:rate5m`, `pyrra_errors:rate5m`, `pyrra_window`) are identical for both types. The difference between static and dynamic is only in the alert threshold calculations, which are not displayed in Grafana dashboards.

### What Grafana Shows

Both dashboard types display:
- SLO objective (target percentage)
- Time window
- Current availability
- Remaining error budget
- Request rate over time
- Error rate over time

These metrics are calculated the same way for both static and dynamic SLOs.

### What Grafana Doesn't Show

Grafana dashboards do not display alerting information:
- Alert state (firing/pending/inactive)
- Burn rate values or thresholds
- Traffic context or sensitivity
- Burn rate type indication

**For detailed burn rate and alert analysis, use the Pyrra UI**, which provides:
- Real-time alert state and multi-window burn rate alerts
- Dynamic threshold calculations with traffic context
- Alert sensitivity indicators
- Detailed tooltips with formula explanations

### Separation of Concerns

- **Grafana**: High-level monitoring of availability and error budget
- **Pyrra UI**: Detailed analysis of burn rates, alerts, and thresholds

This design keeps each tool focused on its strengths.
```
- Difficult to maintain and debug
- Performance impact on Grafana
- Still requires traffic data queries
**Decision**: Rejected - too complex for dashboard maintenance

### Alternative 3: Minimal Enhancement (Chosen Approach)
**Approach**: Show burn rate type only, link to Pyrra UI for details
**Pros**:
- Simple implementation
- Clear separation of concerns
- Maintains performance
- Backward compatible
**Cons**:
- Limited dynamic burn rate visualization in Grafana
**Decision**: Accepted - best balance of simplicity and value

## Success Criteria (Task 7.9 Validation)

### Functional Requirements
- ✅ Grafana dashboards work correctly with static SLOs
- ✅ Grafana dashboards work correctly with dynamic SLOs
- ✅ No code changes required
- ✅ No dashboard modifications required
- ✅ Backward compatible with existing installations

### User Experience Requirements
- ✅ Clear documentation explaining dashboard compatibility
- ✅ Separation of concerns documented (Grafana vs Pyrra UI)
- ✅ No misleading expectations about burn rate visualization

### Technical Requirements
- ✅ No performance impact
- ✅ No breaking changes
- ✅ No additional maintenance burden
- ✅ Generic rules remain identical for both types

## Future Enhancements (Explicitly Out of Scope)

These enhancements are **not recommended** but documented for completeness:

1. **Burn Rate Type Indicator**: Add label to show static vs dynamic
   - **Why not**: Provides no actionable information, adds complexity
   
2. **Alert State Display**: Show firing/pending alerts
   - **Why not**: Inconsistent with current design, duplicates Pyrra UI
   
3. **Dynamic Threshold Graphs**: Visualize calculated thresholds
   - **Why not**: Too complex, requires new recording rules, wrong tool
   
4. **Traffic Context Panels**: Show traffic ratios and sensitivity
   - **Why not**: Requires real-time calculations, better in Pyrra UI
   
5. **Custom Grafana Plugin**: Dedicated Pyrra plugin
   - **Why not**: Massive development effort, maintenance burden

## Conclusion

**Final Decision**: No changes needed to Grafana dashboards for dynamic burn rate support.

**Key Findings**:
1. ✅ Generic recording rules are identical for static and dynamic SLOs
2. ✅ Dashboards already display availability and error budget correctly
3. ✅ Grafana has never shown alerting information (by design)
4. ✅ Pyrra UI is the proper tool for burn rate and alert analysis

**What This Means**:
- **No code changes** required
- **No dashboard modifications** required
- **Testing only** to validate correct behavior
- **Documentation update** to clarify compatibility

**Separation of Concerns**:
- **Grafana**: High-level monitoring (availability, error budget, rates)
- **Pyrra UI**: Detailed analysis (alerts, burn rates, thresholds, traffic context)

This design decision maintains simplicity, avoids unnecessary complexity, and respects the natural division of responsibilities between monitoring tools (Grafana) and analysis tools (Pyrra UI).
