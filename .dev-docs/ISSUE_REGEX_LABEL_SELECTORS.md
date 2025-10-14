# Issue: Regex Label Selectors Behavior Investigation

**Status:** ✅ Resolved - Not a Regression  
**Priority:** Low (Documented as Known Limitation)  
**Date Identified:** 2025-10-14  
**Date Resolved:** 2025-10-14  
**Discovered During:** Task 8.2 - Examples Migration Testing  
**Resolved During:** Task 8.4.1 - Upstream Comparison Testing

---

## Resolution Summary

**Finding:** The observed behavior is **NOT a regression** - it is existing upstream Pyrra behavior.

**Key Discoveries:**
1. ✅ Grouping field creates multiple SLO instances (upstream behavior)
2. ✅ Recording rules work correctly with regex selectors
3. ✅ Detail pages show correct data (availability/budget tiles work)
4. ❌ NaN in burn rate columns is a separate upstream UI bug (affects ALL SLOs when no errors)

**Action Taken:**
- Documented as known limitation in `.dev-docs/KNOWN_LIMITATIONS.md`
- Created comprehensive comparison report in `.dev-docs/UPSTREAM_COMPARISON_REGEX_SELECTORS.md`
- Updated examples to use simple selectors (no grouping) for clarity

**Impact:** No blocker for upstream PR submission. Dynamic burn rate feature works correctly.

---

## Original Problem Description

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

**What Happened (Initial Observation):**
1. **Multiple SLOs Created**: One SLO per specific handler matching the regex pattern
   - Example: `/api/v1/notifications/live`, `/api/v1/query`, `/api/v1/query_range`, etc.
   - Each appears as a separate SLO in the UI

2. **NaN in Alert Table**: Shows "NaN" values in "Short Burn" and "Long Burn" columns

**What Was Actually Happening (After Investigation):**
1. **Multiple SLOs**: Caused by `grouping` field, not regex selectors (upstream behavior)
2. **Recording Rules**: Work correctly - created per-group with `sum by (handler)`
3. **Detail Pages**: Work correctly - availability/budget tiles show 100% (correct when no errors)
4. **NaN Issue**: Separate upstream UI bug affecting ALL SLOs when there are no errors (not regex-specific)

### Investigation Results

**Upstream Comparison Testing (Task 8.4.1):**

✅ **Test 1: Regex + Grouping** (`test-regex-static`)
- Multiple SLOs created (one per handler) - **upstream behavior**
- Recording rules work correctly with `sum by (handler)`
- Detail pages show correct data
- NaN in burn rate columns (separate issue)

✅ **Test 2: Regex WITHOUT Grouping** (`test-regex-no-grouping`)
- Single aggregated SLO created
- Recording rules work correctly
- Detail pages show correct data
- NaN in burn rate columns (separate issue)

✅ **Test 3: Simple Selector Control** (`test-simple-control`)
- Single SLO created
- Recording rules work correctly
- Detail pages show correct data
- **NaN in burn rate columns** - proves this is NOT regex-specific

**Conclusion:** Regex selectors work correctly. The issues observed were:
1. Grouping behavior (upstream design, not a bug)
2. NaN display issue (upstream UI bug, affects all SLOs)

---

## Testing Completed ✅

### Upstream Comparison Tests
- [x] Test regex selector with static burn rate on upstream
- [x] Test regex selector with grouping on upstream
- [x] Document number of SLOs created
- [x] Document recording rule structure
- [x] Check detail page functionality
- [x] Check alert rule functionality

### Feature Branch Tests
- [x] Test same scenarios on feature branch
- [x] Compare behavior differences
- [x] Identify specific regressions (if any) - **None found**

### Results
- ✅ No regressions in feature branch
- ✅ Regex selectors work correctly in both branches
- ✅ Grouping behavior identical in both branches
- ✅ NaN issue exists in both branches (upstream UI bug)

---

## Resolution Actions Taken

1. **Documentation Updated:**
   - `.dev-docs/KNOWN_LIMITATIONS.md` - Added grouping and NaN issues
   - `.dev-docs/UPSTREAM_COMPARISON_REGEX_SELECTORS.md` - Comprehensive test report
   - This file - Updated with resolution

2. **Examples Updated:**
   - Changed examples to use simple selectors (no grouping) for clarity
   - Documented when to use grouping vs aggregated SLOs

3. **Recommendations:**
   - Use regex selectors WITHOUT grouping for aggregated SLOs
   - Use simple selectors WITH grouping for per-endpoint tracking
   - Avoid regex + grouping unless you want many SLO instances

---

## Related Files

- **Comprehensive Test Report**: `.dev-docs/UPSTREAM_COMPARISON_REGEX_SELECTORS.md`
- **Known Limitations**: `.dev-docs/KNOWN_LIMITATIONS.md`
- **Test SLOs Created**: `.dev/test-regex-static.yaml`, `.dev/test-regex-no-grouping.yaml`, `.dev/test-simple-control.yaml`
- **Task Document**: `.dev-docs/TASK_8.2_EXAMPLES_MIGRATION_SUMMARY.md`
- **Task Completion**: Task 8.4.1 in `.kiro/specs/dynamic-burn-rate-completion/tasks.md`

---

## Timeline

**Investigation Started:** 2025-10-14 (Task 8.2)  
**Testing Completed:** 2025-10-14 (Task 8.4.1)  
**Time Spent:** ~2 hours investigation + testing  
**Resolution:** Not a regression - documented as known limitation

---

## Lessons Learned

1. **Always test against upstream** before assuming regressions
2. **Separate issues can appear related** (grouping vs NaN display)
3. **Prometheus empty results ≠ zero** - important for UI handling
4. **Grouping field behavior** is subtle and needs clear documentation
5. **Regex selectors work correctly** - the issue was misunderstood initially
