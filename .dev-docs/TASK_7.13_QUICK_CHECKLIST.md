# Task 7.13: Quick Testing Checklist

## Quick Reference for Interactive Testing

This is a condensed checklist for quick reference during testing. See `TASK_7.13_TESTING_PROCEDURE.md` for detailed instructions.

## Phase 1: Regression Testing (Upstream Comparison)

### Baseline Testing (upstream-comparison branch)

```bash
git checkout upstream-comparison
cd ui && npm run build && cd ..
make build
./pyrra api --prometheus-url=http://localhost:9090
```

**Quick Tests**:
- [ ] List page loads (http://localhost:9099)
- [ ] No "Burn Rate" column present
- [ ] Static SLO detail page works
- [ ] Threshold column format: [Document]
- [ ] Take screenshots for comparison

### Feature Branch Testing (add-dynamic-burn-rate branch)

```bash
# Stop API, then:
git checkout add-dynamic-burn-rate
cd ui && npm run build && cd ..
make build
./pyrra api --prometheus-url=http://localhost:9090
```

**Quick Tests**:
- [ ] List page loads with "Burn Rate" column
- [ ] Gray "Static" badges visible
- [ ] Static SLO detail page works
- [ ] Threshold column shows calculated values
- [ ] Compare with baseline screenshots

### Regression Check

- [ ] No broken functionality
- [ ] No console errors
- [ ] All original features work
- [ ] New features don't interfere

## Phase 2: Production Build Validation

### Critical Tests

**Test 1: Missing Metrics (Task 7.12.1 Fix)**
```bash
kubectl apply -f .dev/test-missing-metrics-dynamic.yaml
```
- [ ] Detail page loads (no crash)
- [ ] Tiles show "No data" (not "100%")
- [ ] Click burn rate graph button
- [ ] **NO WHITE PAGE CRASH**
- [ ] Shows static threshold fallback

**Test 2: All Indicator Types**
- [ ] Ratio indicator works
- [ ] Latency indicator works
- [ ] LatencyNative works (if available)
- [ ] BoolGauge works (if available)

**Test 3: Enhanced Tooltips**
- [ ] List page badge tooltips work
- [ ] Detail page tooltips work
- [ ] Bootstrap styling correct
- [ ] No positioning issues

**Test 4: Performance**
- [ ] Page loads < 3 seconds
- [ ] No memory leaks
- [ ] No console errors

## Quick Commands

### Build Commands
```bash
# UI build
cd ui && npm run build && cd ..

# Go build
make build

# Or single command:
go build -o pyrra .
```

### Service Commands
```bash
# Start API
./pyrra api --prometheus-url=http://localhost:9090

# Start backend (if needed)
./pyrra kubernetes --prometheus-url=http://localhost:9090

# Check services
curl http://localhost:9099/health
curl http://localhost:9090/-/healthy
```

### Test SLO Commands
```bash
# Apply dynamic SLO
kubectl apply -f .dev/test-dynamic-slo.yaml

# Apply missing metrics SLO
kubectl apply -f .dev/test-missing-metrics-dynamic.yaml

# List SLOs
kubectl get slo -n monitoring

# Delete test SLO
kubectl delete slo -n monitoring test-dynamic-apiserver
```

### Git Commands
```bash
# Switch branches
git checkout upstream-comparison
git checkout add-dynamic-burn-rate

# Check recent changes
git log --oneline -5 -- ui/src/

# Check status
git status
```

## Expected Results Summary

### Baseline (upstream-comparison)
- No "Burn Rate" column
- Standard Pyrra UI
- Threshold column shows factors (14x, 7x, 2x, 1x)

### Feature Branch (add-dynamic-burn-rate)
- "Burn Rate" column present
- Gray "Static" badges for static SLOs
- Green "Dynamic" badges for dynamic SLOs
- Threshold column shows calculated values (0.140, 0.070, etc.)
- Enhanced tooltips with traffic context

### Critical Fixes (Task 7.12.1)
- No white page crash for missing metrics
- Detail page shows "No data" (not "100%")
- Graceful error handling throughout

## Documentation Updates Required

After testing complete:

1. Update `.dev-docs/TASK_7.13_COMPREHENSIVE_UI_BUILD_TESTING.md`:
   - Fill in all test results
   - Document any issues found
   - Provide overall assessment

2. Update `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md`:
   - Add Task 7.13 completion
   - Update overall feature status

3. Update `.kiro/specs/dynamic-burn-rate-completion/tasks.md`:
   - Mark task as complete

## Success Criteria

**Minimum**:
- [ ] Regression testing complete
- [ ] No critical regressions
- [ ] Production build works
- [ ] Recent fixes verified

**Full**:
- [ ] All tests passed
- [ ] Performance acceptable
- [ ] No console errors
- [ ] Documentation complete

**Production Ready**:
- [ ] Zero critical issues
- [ ] All fixes working
- [ ] Ready for upstream

## Quick Troubleshooting

**UI build fails**: `rm -rf ui/node_modules && cd ui && npm install`  
**Go build fails**: `go clean -cache && make build`  
**API won't start**: Check Prometheus at http://localhost:9090  
**Can't access UI**: Verify API running, check port 9099  

## Time Estimates

- **Phase 1 (Regression)**: 30-45 minutes
- **Phase 2 (Production)**: 30-45 minutes
- **Documentation**: 15-30 minutes
- **Total**: 1.5-2 hours

## Notes

- Take screenshots for comparison
- Document all observations
- Note any unexpected behavior
- Check console for errors
- Test with real SLOs from cluster
