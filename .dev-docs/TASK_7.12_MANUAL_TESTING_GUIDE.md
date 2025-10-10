# Task 7.12: Manual Testing Guide - Browser Compatibility and Graceful Degradation

## ⚠️ IMPORTANT: This is an INTERACTIVE Task

**This task requires human interaction**. You will need to:
- Open browsers and visually inspect UI behavior
- Test interactive features (tooltips, sorting, navigation)
- Observe error handling and recovery
- Verify migration behavior in UI
- Document observations and issues

**Estimated Time**: 2-3 hours

## Prerequisites

1. **Services Running**:
   ```bash
   # Pyrra API on port 9099
   ./pyrra api &
   
   # Pyrra Backend on port 9444
   ./pyrra kubernetes &
   
   # Prometheus on port 9090 (should already be running)
   ```

2. **Test SLOs Deployed**:
   - Mix of static and dynamic SLOs
   - At least 10-15 SLOs for meaningful testing

3. **Browsers Available**:
   - Chrome (HIGH priority)
   - Firefox (HIGH priority)
   - Edge (MEDIUM priority - optional)

## Test URLs

- **Pyrra UI**: http://localhost:9099
- **Prometheus UI**: http://localhost:9090

## Part 1: Browser Compatibility Testing

### Reference Document

**See**: `.dev-docs/BROWSER_COMPATIBILITY_TEST_GUIDE.md` for detailed test scenarios

### Quick Test Checklist

#### Chrome Testing (HIGH Priority)

1. **Visual Rendering**:
   - [ ] Open http://localhost:9099
   - [ ] Verify burn rate column displays
   - [ ] Check badge colors (green=dynamic, gray=static)
   - [ ] Check icons (eye=dynamic, lock=static)
   - [ ] Verify table layout is correct

2. **Interactive Tooltips**:
   - [ ] Hover over dynamic SLO badge
   - [ ] Verify tooltip shows traffic-aware description
   - [ ] Hover over static SLO badge
   - [ ] Verify tooltip shows static description
   - [ ] Check tooltips disappear on mouse out

3. **Column Sorting**:
   - [ ] Click "Burn Rate" column header
   - [ ] Verify table re-sorts
   - [ ] Click again to reverse sort
   - [ ] Check sort arrow indicator appears

4. **Column Visibility**:
   - [ ] Click "Columns" dropdown
   - [ ] Uncheck "Burn Rate" checkbox
   - [ ] Verify column hides
   - [ ] Check "Burn Rate" again
   - [ ] Verify column shows

5. **Navigation**:
   - [ ] Click on a dynamic SLO name
   - [ ] Verify detail page loads
   - [ ] Check burn rate badge on detail page
   - [ ] Navigate back to list
   - [ ] Verify state preserved

6. **API Communication**:
   - [ ] Open DevTools (F12)
   - [ ] Go to Network tab
   - [ ] Refresh page
   - [ ] Verify API request succeeds (200 OK)
   - [ ] Check console for errors (should be none)

7. **Performance**:
   - [ ] Open DevTools Performance tab
   - [ ] Record while navigating
   - [ ] Check page loads within 3 seconds
   - [ ] Verify smooth scrolling

8. **Error Handling**:
   - [ ] Stop Pyrra API service
   - [ ] Refresh browser
   - [ ] Verify error message displays
   - [ ] Restart API
   - [ ] Verify recovery works

**Document Results**: Note any issues, take screenshots if needed

#### Firefox Testing (HIGH Priority)

Repeat all Chrome tests in Firefox:
- [ ] Visual Rendering
- [ ] Interactive Tooltips
- [ ] Column Sorting
- [ ] Column Visibility
- [ ] Navigation
- [ ] API Communication
- [ ] Performance
- [ ] Error Handling

**Document Results**: Note any differences from Chrome

#### Edge Testing (MEDIUM Priority - Optional)

If time permits, repeat tests in Edge:
- [ ] Visual Rendering
- [ ] Interactive Tooltips
- [ ] Column Sorting
- [ ] Navigation

**Document Results**: Note any differences from Chrome/Firefox

### Create Browser Compatibility Matrix

After testing, create `.dev-docs/BROWSER_COMPATIBILITY_MATRIX.md` with:

```markdown
# Browser Compatibility Matrix

## Tested Browsers

| Browser | Version | Platform | Status | Notes |
|---------|---------|----------|--------|-------|
| Chrome | [version] | Windows | ✅ PASS | [any notes] |
| Firefox | [version] | Windows | ✅ PASS | [any notes] |
| Edge | [version] | Windows | ✅ PASS | [any notes] |

## Known Issues

### Issue 1: [Title]
- **Browser**: [browser name]
- **Description**: [what's wrong]
- **Workaround**: [how to fix or work around]
- **Severity**: [Critical/High/Medium/Low]

## Recommendations

- **Recommended Browsers**: Chrome, Firefox
- **Minimum Browser Versions**: Chrome 90+, Firefox 88+
- **Known Limitations**: [any limitations]
```

## Part 2: Graceful Degradation Testing

### Test 1: Network Throttling

**Objective**: Verify UI handles slow network gracefully

**Steps**:
1. Open http://localhost:9099 in Chrome
2. Open DevTools (F12)
3. Go to Network tab
4. Select "Slow 3G" from throttling dropdown
5. Navigate through UI:
   - [ ] Refresh SLO list page
   - [ ] Click on SLO detail page
   - [ ] Navigate back to list
6. Observe behavior:
   - [ ] Loading states display
   - [ ] UI remains functional
   - [ ] No crashes or blank pages
   - [ ] Appropriate timeouts

**Document**: How does UI behave with slow network? Are loading states clear?

### Test 2: API Failure Simulation

**Objective**: Verify UI handles API failures gracefully

**Steps**:
1. Open http://localhost:9099
2. Verify UI loads correctly
3. Stop Pyrra API service:
   ```bash
   # Find and kill pyrra api process
   # Or press Ctrl+C in terminal running ./pyrra api
   ```
4. Refresh browser
5. Observe behavior:
   - [ ] Error message displays
   - [ ] No blank pages
   - [ ] Error message is user-friendly
   - [ ] No JavaScript errors in console
6. Restart Pyrra API:
   ```bash
   ./pyrra api &
   ```
7. Refresh browser
8. Verify recovery:
   - [ ] UI loads correctly
   - [ ] SLOs display
   - [ ] No lingering errors

**Document**: What error messages appear? Is recovery smooth?

### Test 3: Prometheus Unavailability

**Objective**: Verify threshold calculations handle Prometheus failures

**Steps**:
1. Open http://localhost:9099
2. Navigate to a dynamic SLO detail page
3. Note threshold values displayed
4. Stop Prometheus:
   ```bash
   # Method depends on how Prometheus is running
   # If in Kubernetes: kubectl scale deployment prometheus --replicas=0
   # If local: stop the process
   ```
5. Refresh SLO detail page
6. Observe behavior:
   - [ ] Page still loads
   - [ ] Thresholds show fallback ("Traffic-Aware" or error)
   - [ ] Error messages in console (expected)
   - [ ] No crashes
7. Restart Prometheus
8. Refresh page
9. Verify recovery:
   - [ ] Thresholds calculate correctly
   - [ ] No lingering errors

**Document**: How does UI handle missing Prometheus data?

### Test 4: Missing Metrics (Already Tested in Task 5)

**Note**: This was already validated in Task 5 with test SLOs using non-existent metrics. No additional testing needed.

**Reference**: See `.dev-docs/TASK_5_MISSING_METRICS_VALIDATION.md` (if exists) or Task 5 results

## Part 3: Migration Testing

### Test 1: Static to Dynamic Migration

**Objective**: Verify migration process and UI behavior changes

**Steps**:

1. **Create Test Static SLO**:
   ```bash
   kubectl apply -f - <<EOF
   apiVersion: pyrra.dev/v1alpha1
   kind: ServiceLevelObjective
   metadata:
     name: test-migration
     namespace: monitoring
   spec:
     target: "99"
     window: 30d
     burnRateType: static
     indicator:
       ratio:
         errors:
           metric: apiserver_request_total{code=~"5.."}
         total:
           metric: apiserver_request_total
   EOF
   ```

2. **Verify Static Behavior in UI**:
   - [ ] Open http://localhost:9099
   - [ ] Find "test-migration" SLO in list
   - [ ] Verify gray "Static" badge
   - [ ] Click to open detail page
   - [ ] Verify static thresholds displayed (e.g., "0.700, 0.350, 0.100, 0.050")
   - [ ] Note threshold values for comparison

3. **Migrate to Dynamic**:
   ```bash
   kubectl patch slo test-migration -n monitoring --type=merge -p '{"spec":{"burnRateType":"dynamic"}}'
   ```

4. **Wait for Backend Processing**:
   ```bash
   sleep 30
   ```

5. **Verify Dynamic Behavior in UI**:
   - [ ] Refresh http://localhost:9099
   - [ ] Find "test-migration" SLO in list
   - [ ] Verify green "Dynamic" badge
   - [ ] Click to open detail page
   - [ ] Verify dynamic thresholds or "Traffic-Aware" displayed
   - [ ] Check if thresholds differ from static values
   - [ ] Verify tooltip shows traffic-aware description

6. **Test Rollback**:
   ```bash
   kubectl patch slo test-migration -n monitoring --type=merge -p '{"spec":{"burnRateType":"static"}}'
   sleep 30
   ```

7. **Verify Rollback in UI**:
   - [ ] Refresh http://localhost:9099
   - [ ] Verify gray "Static" badge
   - [ ] Check thresholds match original static values

8. **Cleanup**:
   ```bash
   kubectl delete slo test-migration -n monitoring
   ```

**Document**:
- Was migration seamless?
- Did UI update correctly?
- Were there any errors or delays?
- Did rollback work correctly?

### Test 2: Backward Compatibility

**Objective**: Verify existing static SLOs unaffected

**Steps**:
1. **Before Migration**:
   - [ ] Note current static SLO count
   - [ ] Note behavior of existing static SLOs
   - [ ] Take screenshots if helpful

2. **After Adding Dynamic SLOs**:
   - [ ] Verify static SLO count unchanged
   - [ ] Verify static SLOs still show gray badges
   - [ ] Verify static thresholds unchanged
   - [ ] Verify no errors in static SLO behavior

**Document**: Are static SLOs completely unaffected by dynamic feature?

## Part 4: Create Migration Guide

After completing migration testing, create `.dev-docs/MIGRATION_GUIDE.md`:

```markdown
# Migration Guide: Static to Dynamic Burn Rates

## Overview

This guide provides step-by-step instructions for migrating SLOs from static to dynamic burn rates.

## Prerequisites

- Pyrra with dynamic burn rate feature installed
- kubectl access to Kubernetes cluster
- Existing static SLOs to migrate

## Migration Steps

### Step 1: Identify SLOs to Migrate

[Instructions for identifying which SLOs to migrate]

### Step 2: Update SLO Configuration

[kubectl patch command with example]

### Step 3: Verify Migration

[How to verify in UI and Prometheus]

### Step 4: Monitor Behavior

[What to watch for after migration]

## Rollback Procedure

[How to rollback if needed]

## Troubleshooting

### Issue 1: [Common issue]
**Solution**: [How to fix]

## Best Practices

- [Best practice 1]
- [Best practice 2]
```

## Success Criteria

### Minimum Success
- ✅ Chrome testing completed (all 8 scenarios)
- ✅ Firefox testing completed (all 8 scenarios)
- ✅ Graceful degradation validated (3 scenarios)
- ✅ Migration testing completed
- ✅ Browser compatibility matrix created
- ✅ Migration guide created

### Full Success
- ✅ All minimum success criteria
- ✅ Edge testing completed
- ✅ Comprehensive documentation with screenshots
- ✅ All issues documented with workarounds

## Deliverables

After completing this task, you should have:

1. **Browser Compatibility Matrix** (`.dev-docs/BROWSER_COMPATIBILITY_MATRIX.md`)
   - Tested browsers and versions
   - Test results for each browser
   - Known issues and workarounds

2. **Migration Guide** (`.dev-docs/MIGRATION_GUIDE.md`)
   - Step-by-step migration instructions
   - Rollback procedures
   - Troubleshooting guide
   - Best practices

3. **Test Results Documentation**
   - Notes on graceful degradation behavior
   - Screenshots of any issues
   - Performance observations

## Tips for Effective Testing

1. **Take Screenshots**: Visual evidence helps document issues
2. **Use DevTools**: Console and Network tabs reveal hidden issues
3. **Test Systematically**: Complete one browser fully before moving to next
4. **Document Everything**: Even small observations can be valuable
5. **Test Edge Cases**: Try unusual interactions (rapid clicking, etc.)

## References

- **Detailed Browser Tests**: `.dev-docs/BROWSER_COMPATIBILITY_TEST_GUIDE.md`
- **Testing Infrastructure**: `.dev-docs/TASK_7.11_TESTING_INFRASTRUCTURE.md`
- **Task Definition**: `.kiro/specs/dynamic-burn-rate-completion/tasks.md` (Task 7.12)
- **Requirements**: `.kiro/specs/dynamic-burn-rate-completion/requirements.md` (Requirement 5.2, 5.4)

## Next Steps After Completion

1. Update `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md` with Task 7.12 completion
2. Mark Task 7.12 as complete in tasks.md
3. Review all production readiness documentation
4. Prepare for upstream contribution if all tasks complete
