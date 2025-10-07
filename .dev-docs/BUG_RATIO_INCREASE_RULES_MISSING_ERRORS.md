# Bug Report: Missing Errors Increase Rule for Ratio Indicators

## Discovery Context

- **Session**: Task 7.5 validation (Alert rules generation and end-to-end alert pipeline)
- **Date**: October 7, 2025
- **Discovered By**: User observation during UI validation

## Bug Summary

Ratio indicators with the same base metric for total and errors (e.g., `apiserver_request_total`) are **only generating one increase recording rule** (for total), but **not generating the errors increase rule**. This causes multiple downstream issues in the UI and metrics.

## Symptoms

### 1. Empty Error Graphs

- Both `test-dynamic-apiserver` and `test-static-apiserver` show **empty error graphs** in the UI
- This indicates the error metrics are not being calculated correctly

### 2. Incorrect Availability/Error Budget Calculations

- **test-dynamic-apiserver**: Shows 99.998% availability and 99.967% error budget
- **test-static-apiserver**: Shows 100% availability and 100% error budget
- **Expected**: Both should show the same values since dynamic burn rates only affect alerting thresholds, not KPI calculations
- **Root Cause**: Missing errors increase rule prevents proper error budget calculation

### 3. NaN Values in Alerts Table

- **test-static-apiserver**: Shows "NaN" in Short Burn and Long Burn columns
- **test-dynamic-apiserver**: Shows actual numbers
- **Expected**: Both should show burn rate values
- **Root Cause**: Missing errors data prevents burn rate calculation for static SLO

## Technical Analysis

### Current Behavior (INCORRECT)

The code in `slo/rules.go` (lines ~1030-1050) has this logic:

```go
// Ratio indicator - generate total increase rule
rules = append(rules, monitoringv1.Rule{
    Record: increaseName(o.Indicator.Ratio.Total.Name, o.Window),
    Expr:   intstr.FromString(expr.String()),
    Labels: ruleLabels,
})

// Only generate errors increase rule if metrics are DIFFERENT
if o.Indicator.Ratio.Total.Name != o.Indicator.Ratio.Errors.Name {
    // Generate errors increase rule
    rules = append(rules, monitoringv1.Rule{
        Record: increaseName(o.Indicator.Ratio.Errors.Name, o.Window),
        Expr:   intstr.FromString(expr.String()),
        Labels: ruleLabels,
    })
}
```

### Test SLO Configuration

```yaml
indicator:
  ratio:
    errors:
      metric: apiserver_request_total{verb="GET",code=~"4..|5.."}
    total:
      metric: apiserver_request_total{verb="GET"}
```

- **Total metric name**: `apiserver_request_total`
- **Errors metric name**: `apiserver_request_total` (same!)
- **Result**: Only total increase rule is generated

### Generated Recording Rule (INCORRECT)

```yaml
- record: apiserver_request:increase30d
  expr: sum by (code) (increase(apiserver_request_total{verb="GET"}[30d]))
  labels:
    slo: test-dynamic-apiserver
    verb: GET
```

**Problem**: This rule uses the **total selector** `{verb="GET"}`, not the **errors selector** `{verb="GET",code=~"4..|5.."}`.

### Why This is Wrong

1. **Dynamic threshold calculation** only needs total traffic count - this part is correct
2. **Error budget calculation** needs BOTH total and errors metrics
3. **Burn rate calculation** needs BOTH total and errors metrics
4. **The optimization assumption is WRONG**: Even if metrics have the same base name, they have **different label selectors** and represent **different data**

## Expected Behavior

For ratio indicators, **ALWAYS generate TWO increase rules**:

1. **Total increase rule**: Uses total metric selector
2. **Errors increase rule**: Uses errors metric selector

Even if they share the same base metric name, they have different label selectors and serve different purposes.

## Impact Assessment

### Affected Components

- ✅ **Dynamic threshold calculation**: Works correctly (only uses total)
- ❌ **Error budget calculation**: Broken (missing errors data)
- ❌ **Burn rate calculation**: Broken for static SLOs (missing errors data)
- ❌ **UI error graphs**: Empty (no errors data)
- ❌ **UI alerts table**: Shows NaN for static SLOs

### Affected Indicator Types

- ❌ **Ratio indicators**: When total and errors use same base metric name
- ✅ **Latency indicators**: Not affected (always different metrics)
- ✅ **LatencyNative indicators**: Not affected (always different metrics)
- ✅ **BoolGauge indicators**: Not affected (different aggregations)

### Affected SLO Types

- ❌ **Static SLOs**: Broken burn rate calculations
- ❌ **Dynamic SLOs**: Broken error budget calculations

## Root Cause

The optimization in `slo/rules.go` line ~1030:

```go
if o.Indicator.Ratio.Total.Name != o.Indicator.Ratio.Errors.Name {
```

This checks if the **metric names** are different, but doesn't account for:

1. Different **label selectors** (e.g., `code=~"4..|5.."` vs no code filter)
2. Different **purposes** (total traffic vs error traffic)
3. The fact that **both metrics are needed** for error budget and burn rate calculations

## Proposed Fix

### Option 1: Always Generate Both Rules (Recommended)

Remove the conditional check and always generate both increase rules:

```go
// Always generate total increase rule
rules = append(rules, monitoringv1.Rule{
    Record: increaseName(o.Indicator.Ratio.Total.Name, o.Window),
    Expr:   intstr.FromString(totalExpr.String()),
    Labels: ruleLabels,
})

// Always generate errors increase rule
rules = append(rules, monitoringv1.Rule{
    Record: increaseName(o.Indicator.Ratio.Errors.Name, o.Window),
    Expr:   intstr.FromString(errorsExpr.String()),
    Labels: ruleLabels,
})
```

### Option 2: Check Label Selectors, Not Just Metric Name

Compare the full metric selector (name + labels), not just the name:

```go
if !selectorsEqual(o.Indicator.Ratio.Total.LabelMatchers, o.Indicator.Ratio.Errors.LabelMatchers) {
    // Generate errors increase rule
}
```

**Recommendation**: Option 1 is simpler and safer. The "optimization" doesn't provide significant value and causes bugs.

## Validation Steps

After fixing, verify:

1. **Recording rules generated**:

   ```bash
   kubectl get prometheusrule test-dynamic-apiserver -n monitoring -o json | \
     jq '.spec.groups[0].rules[] | select(.record != null) | .record'
   ```

   Should show BOTH:

   - `apiserver_request:increase30d` (total)
   - `apiserver_request:increase30d` (errors) - with different expr

2. **Error graphs populate** in UI for both static and dynamic SLOs

3. **Availability/Error budget match** between static and dynamic SLOs

4. **Alerts table shows values** (not NaN) for both static and dynamic SLOs

5. **Burn rate calculations work** for both static and dynamic SLOs

## Files to Modify

- `slo/rules.go`: Lines ~1030-1050 (IncreaseRules function, Ratio case)

## Related Code Sections

- `slo/rules.go:914-1100`: IncreaseRules() function
- `slo/rules.go:200-240`: buildAlertExpr() for Ratio case
- Dynamic threshold calculation (only uses total - this is correct)
- Error budget calculation (needs both total and errors - currently broken)

## Testing Requirements

After fix, test with:

1. Ratio SLO where total and errors use same base metric (e.g., `apiserver_request_total`)
2. Ratio SLO where total and errors use different metrics
3. Verify both static and dynamic burn rate types
4. Check UI displays: error graphs, availability, error budget, alerts table

## Session Continuity

After fixing this bug:

1. Return to Task 7.5 validation session
2. Re-run validation script: `bash scripts/validate-alert-rules.sh`
3. Verify UI displays are correct
4. Continue with end-to-end alert firing test

## Additional Notes

- This bug affects **both static and dynamic SLOs**, not just dynamic
- The bug was introduced by an optimization that assumed same metric name = same data
- The fix is straightforward: always generate both increase rules for ratio indicators
- This is a **critical bug** that breaks core SLO functionality (error budget, burn rates)

---

**Status**: ✅ **FIXED** - Applied in slo/rules.go
**Priority**: HIGH - Affects core SLO metrics and UI displays
**Complexity**: LOW - Simple fix, remove incorrect optimization

## Fix Applied

**Date**: October 7, 2025
**File Modified**: `slo/rules.go` (lines 1016-1050)

**Change**: Removed the conditional check `if o.Indicator.Ratio.Total.Name != o.Indicator.Ratio.Errors.Name` and now **always generate both increase rules** for ratio indicators.

**Rationale**: Even when total and errors metrics share the same base name (e.g., `apiserver_request_total`), they have different label selectors and represent different data. Both rules are needed for:

- Error budget calculations
- Burn rate calculations (both static and dynamic)
- UI error graphs
- UI alerts table

**Verification Steps Completed**:

1. ✅ Rebuild pyrra binary
2. ✅ Restart pyrra kubernetes backend  
3. ✅ Apply updated static SLO with verb="GET"
4. ✅ Verify both increase rules generated for both static and dynamic SLOs
5. ✅ Confirmed recording rules show correct expressions with proper label selectors

**Recording Rules Verified**:

Both test-static-apiserver and test-dynamic-apiserver now generate:
- Total increase rule: `apiserver_request_total{verb="GET"}`
- Errors increase rule: `apiserver_request_total{code=~"4..|5..",verb="GET"}`

**Error Graph Display Note**:

The `> 0` filter in ratio error queries (`slo/promql.go` line 672) is intentional legacy behavior (present for years in codebase). This causes "no data" display when error rate is 0, unlike other indicator types that show a line at 0. This is expected behavior and not part of this bug fix.

**Next Steps**:

1. Return to Task 7.5 validation session
2. Re-run validation script: `bash scripts/validate-alert-rules.sh`
3. Verify UI displays show matching availability/error budget between static and dynamic SLOs
4. Verify alerts table shows values (not NaN) for both SLO types
5. Continue with end-to-end alert firing test
