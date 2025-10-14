# Issue: Regex Label Selectors Behavior Investigation

**Status:** 🔍 Investigation Required  
**Priority:** High  
**Date Identified:** 2025-10-14  
**Discovered During:** Task 8.2 - Examples Migration Testing

---

## Problem Description

When using regex label selectors (e.g., `handler=~"/api.*"`) in SLO definitions, Pyrra exhibits unexpected behavior that creates inconsistencies between the main page and detail page displays.

### Observed Behavior

**Test Case:** SLO with regex selector
```yaml
indicator:
  ratio:
    errors:
      metric: prometheus_http_requests_total{handler=~"/api.*",code=~"5.."}
    total:
      metric: prometheus_http_requests_total{handler=~"/api.*"}
    grouping:
      - handler
```

**What Happened:**
1. **Multiple SLOs Created**: One SLO per specific handler matching the regex pattern
   - Example: `/api/v1/notifications/live`, `/api/v1/query`, `/api/v1/query_range`, etc.
   - Each appears as a separate SLO in the UI

2. **Aggregated Recording Rules**: Recording rules created with aggregated values
   - Rules use `handler=~"/api.*"` (the regex pattern)
   - Not specific to individual handlers

3. **Data Inconsistency**:
   - **Detail Page**: Shows "No data" for availability and Error Budget tiles
   - **Detail Page**: Shows "NaN" values in "Short Burn" and "Long Burn" columns in alert tables
   - **Main Page**: Shows actual values in availability and budget columns
   - **Root Cause**: Recording rules aggregate across all handlers, but detail page expects per-handler data

### Key Questions

1. **Upstream Behavior**: Does upstream Pyrra exhibit the same behavior with regex selectors?
   - Test with `upstream-comparison` branch
   - Use same SLO configuration with regex selector
   - Document whether it breaks or works correctly

2. **Expected Behavior**: What should happen with regex selectors?
   - **Option A**: Create multiple SLOs (one per matching label value) with per-SLO recording rules
   - **Option B**: Create single aggregated SLO with aggregated recording rules
   - **Current**: Hybrid approach (multiple SLOs + aggregated rules) = broken

3. **Design Philosophy**: What is the intended relationship between SLO YAML and SLO instances?
   - **One-to-One**: One YAML file = One SLO (aggregated)
   - **One-to-Many**: One YAML file = Multiple SLOs (per grouping label value)

### Impact

**User Experience:**
- Confusing to see multiple SLOs from one YAML definition
- "No data" in detail pages makes SLOs appear broken
- Inconsistent data between main page and detail page

**Functional Impact:**
- Detail page unusable for regex-based SLOs
- Alert rules may not fire correctly (NaN thresholds)
- Error budget calculations inconsistent

---

## Investigation Plan

### Phase 1: Upstream Comparison Testing

**Objective**: Determine if this is a regression or existing upstream behavior

**Steps:**
1. Checkout `upstream-comparison` branch
2. Create test SLO with regex selector:
   ```yaml
   indicator:
     ratio:
       errors:
         metric: prometheus_http_requests_total{handler=~"/api.*",code=~"5.."}
       total:
         metric: prometheus_http_requests_total{handler=~"/api.*"}
       grouping:
         - handler
   ```
3. Apply to Kubernetes cluster
4. Observe behavior:
   - How many SLOs appear in UI?
   - What do recording rules look like?
   - Does detail page show data?
   - Are alert rules functional?
5. Document findings

**Expected Outcomes:**
- **If upstream breaks**: This is an existing Pyrra limitation, not our regression
- **If upstream works**: We introduced a regression that needs fixing

### Phase 2: Code Analysis

**Objective**: Understand how grouping and label selectors interact

**Files to Review:**
- `slo/rules.go` - Recording rule generation logic
- `slo/slo.go` - SLO object handling
- `kubernetes/controllers/` - How SLOs are processed from CRDs

**Key Questions:**
- How does `grouping` field affect SLO instantiation?
- How are recording rules scoped (per-group vs aggregated)?
- Where is the mismatch between main page and detail page calculations?

### Phase 3: Dynamic Burn Rate Specific Analysis

**Objective**: Determine if dynamic burn rate logic exacerbates the issue

**Focus Areas:**
- Traffic calculation logic in dynamic burn rate alerts
- How `N_alert` and `N_SLO` are calculated with grouping
- Whether static burn rates have the same issue

**Test Cases:**
1. Regex selector + static burn rate
2. Regex selector + dynamic burn rate
3. Simple selector + dynamic burn rate (control)

### Phase 4: Solution Design

**Based on investigation findings, choose approach:**

#### Option A: Fix to Match Upstream (if upstream works)
- Identify regression in our code
- Fix recording rule generation or SLO instantiation
- Ensure detail page calculations match main page

#### Option B: Document Limitation (if upstream also breaks)
- Add warning in documentation about regex selectors
- Recommend using simple selectors without grouping
- Consider upstream contribution to fix the issue

#### Option C: Implement Aggregated SLO Approach
- One YAML = One SLO (aggregated across grouping labels)
- Recording rules aggregate across all label values
- Detail page shows aggregated data
- Simpler, more predictable behavior

---

## Recommended Approach (Pending Investigation)

**Personal Opinion**: One SLO YAML should define one SLO (aggregated)

**Rationale:**
- More predictable behavior
- Easier to understand and manage
- Aligns with "Service Level Objective" concept (service-level, not endpoint-level)
- Users can create multiple YAML files if they want per-endpoint SLOs

**However**: Must align with upstream Pyrra behavior and philosophy

---

## Testing Checklist

### Upstream Comparison Tests
- [ ] Test regex selector with static burn rate on upstream
- [ ] Test regex selector with grouping on upstream
- [ ] Document number of SLOs created
- [ ] Document recording rule structure
- [ ] Check detail page functionality
- [ ] Check alert rule functionality

### Feature Branch Tests
- [ ] Test same scenarios on feature branch
- [ ] Compare behavior differences
- [ ] Identify specific regressions (if any)

### Edge Cases
- [ ] Multiple grouping labels with regex
- [ ] Regex matching zero labels
- [ ] Regex matching hundreds of labels
- [ ] Mixed regex and exact match selectors

---

## Success Criteria

1. **Understanding Achieved**: Clear documentation of upstream behavior
2. **Root Cause Identified**: Know exactly where the mismatch occurs
3. **Solution Decided**: Clear path forward (fix, document, or redesign)
4. **Tests Pass**: All SLO examples work correctly in both main and detail pages
5. **Documentation Updated**: Clear guidance for users on label selectors

---

## Related Files

- **Test SLO**: `.dev/test-dynamic-slo.yaml` (original test with regex)
- **Example SLO**: `examples/dynamic-burn-rate-ratio.yaml` (updated to simple selector)
- **Task Document**: `.dev-docs/TASK_8.2_EXAMPLES_MIGRATION_SUMMARY.md`
- **Rules Generation**: `slo/rules.go`

---

## Timeline

**Estimated Effort**: 2-4 hours investigation + 2-8 hours implementation (depending on findings)

**Suggested Scheduling**: 
- Can be done as part of Task 9 (upstream integration preparation)
- Or as separate investigation task before finalizing examples
- Should be resolved before upstream PR submission

---

## Notes

This issue highlights the importance of testing with realistic, production-like configurations. The regex selector pattern is common in production environments where users want to track multiple endpoints with a single SLO definition.

The inconsistency between main page and detail page suggests a fundamental architectural question about how Pyrra handles SLO instantiation and data aggregation.
