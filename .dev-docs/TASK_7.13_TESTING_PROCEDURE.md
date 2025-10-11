# Task 7.13: Step-by-Step Testing Procedure

## Overview

This document provides detailed step-by-step instructions for executing Task 7.13: Comprehensive UI Build and Deployment Testing.

## Prerequisites

- Minikube cluster running with kube-prometheus stack
- Pyrra backend services available
- Both `add-dynamic-burn-rate` and `upstream-comparison` branches available locally
- Node.js and npm installed for UI builds
- Go toolchain installed for binary builds

## Phase 1: Systematic Regression Testing

### Step 1: Prepare Upstream Comparison Branch

**Objective**: Build and test original Pyrra behavior as baseline

```bash
# 1. Switch to upstream-comparison branch
git checkout upstream-comparison

# 2. Build UI
cd ui
npm install  # Only if dependencies changed
npm run build

# 3. Build Go binary with embedded UI
cd ..
make build

# 4. Verify binary created
ls -lh pyrra  # Should show recent timestamp

# 5. Start Pyrra API service
./pyrra api --prometheus-url=http://localhost:9090

# Service should start on http://localhost:9099
```

**Expected Output**:
```
level=info msg="Starting Pyrra API" version=...
level=info msg="Listening on :9099"
```

### Step 2: Test Baseline Static SLO Behavior

**Objective**: Document original Pyrra behavior for static SLOs

**Test 2.1: List Page Baseline**

1. Open browser to http://localhost:9099
2. Document observations:
   - [ ] Page loads successfully
   - [ ] SLO table displays
   - [ ] Columns present: [List columns]
   - [ ] No "Burn Rate" column (feature not implemented)
   - [ ] Tooltips work: [Describe behavior]
   - [ ] Sorting works: [Test columns]
   - [ ] Navigation works: [Click on SLO]

**Test 2.2: Detail Page Baseline**

1. Click on a static SLO (e.g., `prometheus-operator-prometheus-operator-admission`)
2. Document observations:
   - [ ] Detail page loads successfully
   - [ ] Summary tiles display: Objective, Availability, Error Budget
   - [ ] Graphs render: ErrorBudget, Requests, Errors
   - [ ] Alerts table displays
   - [ ] Threshold column shows: [Document format - factors or values?]
   - [ ] Burn rate graphs work: [Click expand button]
   - [ ] Time range controls work: [Test 1h, 12h, 1d, 4w]
   - [ ] Auto-reload toggle works: [Enable/disable]

**Test 2.3: Missing Metrics Baseline**

1. Create test SLO with non-existent metrics (if possible)
2. Document error handling:
   - [ ] List page behavior: [Document]
   - [ ] Detail page behavior: [Document]
   - [ ] Error messages: [Document]
   - [ ] Console errors: [Document]

**Screenshot Recommendations**:
- Take screenshots of list page
- Take screenshots of detail page
- Take screenshots of alerts table
- Save for comparison with feature branch

### Step 3: Prepare Feature Branch

**Objective**: Build and test feature branch with all recent fixes

```bash
# 1. Stop Pyrra API service (Ctrl+C)

# 2. Switch to feature branch
git checkout add-dynamic-burn-rate

# 3. Verify recent changes are present
git log --oneline -5 -- ui/src/

# Should show:
# - 8378126 Fix missing metrics display issues (Task 7.12.1)
# - c6f27c9 Task 7.10.4: Final validation and bug fixes
# - etc.

# 4. Build UI with all recent fixes
cd ui
npm run build

# 5. Build Go binary with embedded UI
cd ..
make build

# 6. Verify binary created with recent timestamp
ls -lh pyrra

# 7. Start Pyrra API service
./pyrra api --prometheus-url=http://localhost:9090
```

### Step 4: Test Feature Branch Static SLO Behavior

**Objective**: Compare static SLO behavior with baseline to identify regressions

**Test 4.1: List Page Comparison**

1. Open browser to http://localhost:9099
2. Compare with baseline observations:
   - [ ] Page loads successfully (same as baseline?)
   - [ ] SLO table displays (same as baseline?)
   - [ ] **NEW**: "Burn Rate" column present
   - [ ] **NEW**: Gray "Static" badges visible
   - [ ] **NEW**: Tooltips on badges work
   - [ ] Original columns unchanged: [Verify]
   - [ ] Sorting works: [Test all columns including new Burn Rate column]
   - [ ] Navigation works: [Click on SLO]

**Regression Check**:
- [ ] No layout issues or visual glitches
- [ ] No console errors
- [ ] All original functionality preserved
- [ ] New features don't interfere with existing features

**Test 4.2: Detail Page Comparison**

1. Click on same static SLO as baseline test
2. Compare with baseline observations:
   - [ ] Detail page loads successfully (same as baseline?)
   - [ ] Summary tiles display correctly (same as baseline?)
   - [ ] Graphs render correctly (same as baseline?)
   - [ ] **CHANGED**: Threshold column shows calculated values instead of factors
   - [ ] **NEW**: Burn rate type badge in header
   - [ ] **NEW**: Enhanced tooltips in alerts table
   - [ ] Burn rate graphs work: [Click expand button]
   - [ ] Time range controls work: [Test 1h, 12h, 1d, 4w]
   - [ ] Auto-reload toggle works: [Enable/disable]

**Regression Check**:
- [ ] No functionality broken
- [ ] No console errors
- [ ] All graphs render correctly
- [ ] All interactions work smoothly

**Test 4.3: Missing Metrics Comparison**

1. Test same SLO with missing metrics (if created in baseline)
2. Compare error handling:
   - [ ] List page behavior: [Compare with baseline]
   - [ ] Detail page behavior: [Compare with baseline]
   - [ ] **IMPROVED**: Should show "No data" instead of "100%"
   - [ ] **IMPROVED**: No white page crash when clicking burn rate graph
   - [ ] Error messages: [Compare with baseline]
   - [ ] Console errors: [Compare with baseline]

**Regression Check**:
- [ ] Error handling same or better than baseline
- [ ] No new crashes or errors
- [ ] Graceful degradation maintained

### Step 5: Test Mixed Static/Dynamic Environment

**Objective**: Verify static and dynamic SLOs coexist without conflicts

**Test 5.1: Create Dynamic SLO**

```bash
# Apply dynamic test SLO
kubectl apply -f .dev/test-dynamic-slo.yaml

# Wait for processing (~30 seconds)
sleep 30

# Verify SLO created
kubectl get slo -n monitoring test-dynamic-apiserver
```

**Test 5.2: Verify Mixed Environment**

1. Refresh list page (http://localhost:9099)
2. Verify both static and dynamic SLOs present:
   - [ ] Static SLOs show gray "Static" badges
   - [ ] Dynamic SLO shows green "Dynamic" badge
   - [ ] No visual conflicts or layout issues
   - [ ] Both types sortable in Burn Rate column

3. Click on static SLO:
   - [ ] Shows "Factor" column in alerts table
   - [ ] Shows calculated threshold values
   - [ ] Burn rate type badge shows "Static"

4. Click on dynamic SLO:
   - [ ] Shows "Error Budget Consumption" column in alerts table
   - [ ] Shows traffic-aware threshold values
   - [ ] Burn rate type badge shows "Dynamic"

**Regression Check**:
- [ ] Static SLOs unaffected by presence of dynamic SLOs
- [ ] No interference between static and dynamic features
- [ ] All functionality works for both types

### Step 6: Document Regression Testing Results

**Objective**: Summarize findings and identify any regressions

Update `.dev-docs/TASK_7.13_COMPREHENSIVE_UI_BUILD_TESTING.md` with:

1. **Baseline Behavior**: Document upstream-comparison observations
2. **Feature Branch Behavior**: Document add-dynamic-burn-rate observations
3. **Intentional Differences**: List expected changes (new features)
4. **Regressions Found**: List any unintended changes or broken functionality
5. **Conclusion**: Overall assessment of regression testing

## Phase 2: Final Production Build Validation

### Step 7: Verify All Recent Fixes in Production Build

**Objective**: Confirm Task 7.12.1 fixes and all enhancements work in embedded UI

**Test 7.1: Missing Metrics Handling (Task 7.12.1 Fix)**

1. Create or use existing SLO with missing metrics:
```bash
# Example: test-missing-metrics-dynamic.yaml
kubectl apply -f .dev/test-missing-metrics-dynamic.yaml
```

2. Navigate to SLO detail page
3. Test missing metrics handling:
   - [ ] Detail page loads (no crash)
   - [ ] Availability tile shows "No data" (not "100%")
   - [ ] Error Budget tile shows "No data" (not "100%")
   - [ ] Click burn rate graph button
   - [ ] **CRITICAL**: No white page crash
   - [ ] Graph shows static threshold fallback
   - [ ] Console shows meaningful error messages
   - [ ] Can navigate away without issues

**Expected Console Messages**:
```
[BurnrateGraph] No traffic data available for dynamic threshold calculation, falling back to static threshold
[BurnRateThresholdDisplay] No data returned for ratio indicator traffic query
```

**Test 7.2: All Indicator Types**

1. Test ratio indicator:
   - [ ] Thresholds display correctly
   - [ ] Tooltips show traffic context
   - [ ] Graphs render without errors

2. Test latency indicator:
   - [ ] Thresholds display correctly
   - [ ] Histogram metrics handled correctly
   - [ ] Tooltips show traffic context
   - [ ] Graphs render without errors

3. Test latencyNative indicator (if available):
   - [ ] Thresholds display correctly
   - [ ] Native histogram metrics handled correctly
   - [ ] Tooltips show traffic context

4. Test boolGauge indicator (if available):
   - [ ] Thresholds display correctly
   - [ ] Boolean gauge metrics handled correctly
   - [ ] Tooltips show traffic context

**Test 7.3: Enhanced Tooltips**

1. List page tooltips:
   - [ ] Hover over "Static" badge
   - [ ] Tooltip appears with correct content
   - [ ] Hover over "Dynamic" badge
   - [ ] Tooltip appears with traffic-aware content

2. Detail page tooltips:
   - [ ] Hover over burn rate type badge
   - [ ] Tooltip shows enhanced content
   - [ ] Hover over threshold values in alerts table
   - [ ] Tooltips show calculation details

3. Tooltip styling:
   - [ ] Bootstrap styling applied correctly
   - [ ] Max width 400px
   - [ ] Readable font size
   - [ ] No positioning issues

**Test 7.4: Performance Validation**

1. Measure page load times:
   - [ ] List page: < 3 seconds
   - [ ] Detail page: < 3 seconds
   - [ ] Navigation: Smooth and responsive

2. Monitor memory usage:
   - [ ] Open browser dev tools
   - [ ] Check memory tab
   - [ ] Navigate between pages
   - [ ] Verify no memory leaks

3. Check console for errors:
   - [ ] No JavaScript errors
   - [ ] No React warnings
   - [ ] Only expected Prometheus query errors (for missing metrics)

### Step 8: Document Production Build Results

**Objective**: Summarize production build validation findings

Update `.dev-docs/TASK_7.13_COMPREHENSIVE_UI_BUILD_TESTING.md` with:

1. **Test Results**: Document all test outcomes
2. **Issues Found**: List any problems discovered
3. **Performance Metrics**: Document load times and memory usage
4. **Production Readiness**: Overall assessment

## Phase 3: Final Documentation Updates

### Step 9: Update Feature Implementation Summary

Update `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md`:

1. Add Task 7.13 completion status
2. Document regression testing results
3. Document production build validation results
4. Update overall feature status

### Step 10: Update Task Status

Mark task as complete in `.kiro/specs/dynamic-burn-rate-completion/tasks.md`:

```bash
# After all testing complete and documented
# Update task status to complete
```

## Success Criteria

### Minimum Success
- [ ] Regression testing completed for all 4 scenarios
- [ ] No critical regressions found in static SLO behavior
- [ ] Production build created successfully
- [ ] All recent fixes verified in production build

### Full Success
- [ ] All regression tests passed
- [ ] All production build tests passed
- [ ] Performance acceptable (< 3 second page loads)
- [ ] No console errors in normal operation
- [ ] Documentation complete and accurate

### Production Ready
- [ ] Zero critical regressions
- [ ] All Task 7.12.1 fixes working in production
- [ ] All indicator types working correctly
- [ ] Error handling robust and graceful
- [ ] Ready for upstream contribution

## Troubleshooting

### Issue: UI Build Fails

**Symptoms**: `npm run build` fails with errors

**Solutions**:
1. Check Node.js version: `node --version` (should be 14+)
2. Clear node_modules: `rm -rf node_modules && npm install`
3. Check for TypeScript errors: `npm run build` output
4. Verify all recent commits applied: `git log --oneline -5`

### Issue: Go Build Fails

**Symptoms**: `make build` fails with errors

**Solutions**:
1. Check Go version: `go version` (should be 1.19+)
2. Clean build cache: `go clean -cache`
3. Verify UI build exists: `ls ui/build/index.html`
4. Check for syntax errors: `go build -o pyrra .`

### Issue: Pyrra API Won't Start

**Symptoms**: `./pyrra api` fails or crashes

**Solutions**:
1. Check Prometheus URL: `curl http://localhost:9090/-/healthy`
2. Check port availability: `netstat -an | grep 9099`
3. Check logs for errors: Look at console output
4. Verify binary permissions: `chmod +x pyrra`

### Issue: Can't Access UI

**Symptoms**: Browser can't connect to http://localhost:9099

**Solutions**:
1. Verify Pyrra API running: Check console for "Listening on :9099"
2. Check firewall: Ensure port 9099 not blocked
3. Try different browser: Test with Chrome, Firefox
4. Check browser console: Look for network errors

## References

- **Task Definition**: `.kiro/specs/dynamic-burn-rate-completion/tasks.md` (Task 7.13)
- **Test Results**: `.dev-docs/TASK_7.13_COMPREHENSIVE_UI_BUILD_TESTING.md`
- **Feature Summary**: `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md`
- **UI Refresh Investigation**: `.dev-docs/TASK_7.6_UI_REFRESH_RATE_INVESTIGATION.md`
- **Browser Compatibility**: `.dev-docs/BROWSER_COMPATIBILITY_MATRIX.md`
