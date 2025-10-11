# Migration Guide: Static to Dynamic Burn Rates

## Overview

This guide provides step-by-step instructions for migrating Pyrra SLOs from static burn rates to dynamic (traffic-aware) burn rates.

**Target Audience**: SRE teams, platform engineers, Pyrra administrators

**Prerequisites**:

- Pyrra with dynamic burn rate feature installed
- kubectl access to Kubernetes cluster
- Existing static SLOs to migrate
- Basic understanding of SLO concepts

## Understanding the Difference

### Static Burn Rates

Static burn rates use **fixed multipliers** for alert thresholds:

- Factor 14: Critical alerts (2% error budget consumption)
- Factor 7: High severity alerts (6% error budget consumption)
- Factor 2: Medium severity alerts (7% error budget consumption)
- Factor 1: Low severity alerts (14% error budget consumption)

**Formula**: `error_rate > factor × (1 - SLO_target)`

**Behavior**: Thresholds remain constant regardless of traffic volume.

### Dynamic Burn Rates

Dynamic burn rates use **traffic-aware thresholds** that adapt based on actual traffic patterns:

**Formula**: `error_rate > (N_SLO / N_alert) × E_budget_percent × (1 - SLO_target)`

Where:

- `N_SLO` = Number of events in SLO window (e.g., 30 days)
- `N_alert` = Number of events in alert window (e.g., 1 hour)
- `E_budget_percent` = Error budget percentage (1/48, 1/16, 1/14, 1/7)
- `(1 - SLO_target)` = Error budget (e.g., 0.01 for 99% SLO)

**Behavior**: Thresholds scale with traffic volume, preventing false positives during low traffic and false negatives during high traffic.

### Key Benefits

1. **Reduced False Positives**: Lower thresholds during low traffic periods
2. **Improved Sensitivity**: Higher thresholds during high traffic periods
3. **Traffic-Aware Alerting**: Alerts adapt to actual service usage patterns
4. **Same Error Budget**: Error budget calculations remain identical

### Important Notes

⚠️ **Error Budget Unchanged**: Dynamic burn rates only affect alert thresholds, not error budget calculations.

⚠️ **Backward Compatible**: Static SLOs continue to work unchanged. Migration is opt-in per SLO.

⚠️ **Recording Rules**: Dynamic SLOs generate the same recording rules as static SLOs.

## Migration Steps

### Step 1: Identify SLOs to Migrate

**Criteria for Migration**:

- SLOs with variable traffic patterns (daily/weekly cycles)
- SLOs experiencing false positive alerts during low traffic
- SLOs missing issues during high traffic periods
- Production services with established baseline traffic

**Not Recommended for Migration**:

- SLOs with constant traffic patterns
- Test/development SLOs with minimal traffic
- SLOs without sufficient historical data (< 7 days)

**List Current SLOs**:

```bash
kubectl get slo -n monitoring
```

**Check SLO Configuration**:

```bash
kubectl get slo <slo-name> -n monitoring -o yaml
```

Look for `spec.burnRateType` field:

- `static` or missing = Static burn rate (default)
- `dynamic` = Dynamic burn rate

### Step 2: Backup Current Configuration

**Export SLO Configuration**:

```bash
kubectl get slo <slo-name> -n monitoring -o yaml > backup-<slo-name>.yaml
```

**Verify Backup**:

```bash
cat backup-<slo-name>.yaml
```

### Step 3: Update SLO to Dynamic Burn Rate

**Method 1: kubectl patch (Recommended)**:

```bash
kubectl patch slo <slo-name> -n monitoring --type=merge -p '{"spec":{"burnRateType":"dynamic"}}'
```

**Method 2: kubectl edit**:

```bash
kubectl edit slo <slo-name> -n monitoring
```

Add or modify the `burnRateType` field:

```yaml
spec:
  burnRateType: dynamic # Add this line
  target: "99"
  window: 30d
  # ... rest of configuration
```

**Method 3: Apply Updated YAML**:

```bash
# Edit your SLO YAML file
# Add: spec.burnRateType: dynamic

kubectl apply -f <slo-file>.yaml
```

### Step 4: Wait for Backend Processing

Pyrra backend needs time to regenerate recording and alert rules:

```bash
# Wait 30 seconds for backend to process
sleep 30
```

**Monitor Backend Logs** (optional):

```bash
# If running locally
# Check terminal running ./pyrra kubernetes

# If running in Kubernetes
kubectl logs -n monitoring deployment/pyrra-kubernetes -f
```

### Step 5: Verify Migration in Pyrra UI

**Open Pyrra UI**:

```
http://localhost:9099  # Local development
# OR
http://<pyrra-service-url>  # Production
```

**Verify SLO List Page**:

1. Find the migrated SLO in the list
2. Check the "Burn Rate" column
3. Verify badge shows **green "Dynamic"** (not gray "Static")
4. Hover over badge to see tooltip: "Dynamic Burn Rate: Adapts thresholds based on traffic patterns"

**Verify SLO Detail Page**:

1. Click on the SLO name to open detail page
2. Check burn rate type badge at top (should be green "Dynamic")
3. Scroll to "Multi Burn Rate Alerts" table
4. Check "Threshold" column:
   - Should show calculated values (e.g., "0.00123")
   - OR show "Traffic-Aware" if calculation pending
   - Should NOT show static values (e.g., "0.700")
5. Hover over threshold values to see tooltip with traffic context

### Step 6: Verify Alert Rules in Prometheus

**Check Alert Rules**:

```bash
# Open Prometheus UI
http://localhost:9090  # Local development
# OR
http://<prometheus-url>  # Production

# Navigate to: Status > Rules
# Find your SLO's alert rules
# Example: <slo-name>-short-burn or <slo-name>-long-burn
```

**Verify Alert Expression**:
Dynamic alert rules should include traffic ratio calculation:

```promql
# Example dynamic alert expression
(
  sum(rate(metric[1h4m]))
  >
  (
    (sum(increase(metric[30d])) / sum(increase(metric[1h4m])))
    * 0.020833  # Error budget percentage
    * (1 - 0.99)  # SLO target
  )
)
```

### Step 7: Monitor Alert Behavior

**Initial Monitoring Period**: 24-48 hours

**What to Watch**:

1. **Alert Firing**: Do alerts fire appropriately?
2. **False Positives**: Reduced compared to static?
3. **False Negatives**: Are real issues still caught?
4. **Threshold Adaptation**: Do thresholds change with traffic?

**Check AlertManager**:

```
http://localhost:9093  # Local development
# OR
http://<alertmanager-url>  # Production
```

**Compare with Static Behavior** (if you have parallel static SLO):

- Same error conditions should trigger alerts
- Dynamic alerts should be more traffic-aware
- Alert timing may differ based on traffic patterns

### Step 8: Document Migration

**Record Migration Details**:

- SLO name and namespace
- Migration date and time
- Previous behavior (static thresholds)
- New behavior (dynamic thresholds)
- Any issues encountered
- Alert behavior changes observed

## Rollback Procedure

If migration causes issues, you can rollback to static burn rates:

### Quick Rollback

**Method 1: kubectl patch**:

```bash
kubectl patch slo <slo-name> -n monitoring --type=merge -p '{"spec":{"burnRateType":"static"}}'
```

**Method 2: Restore from Backup**:

```bash
kubectl apply -f backup-<slo-name>.yaml
```

### Verify Rollback

1. Wait 30 seconds for backend processing
2. Check Pyrra UI - badge should be gray "Static"
3. Check thresholds return to static values
4. Verify alert rules revert to static expressions

### Rollback Considerations

- Rollback is immediate (after backend processing)
- No data loss - error budget calculations unchanged
- Alert rules regenerate automatically
- Recording rules remain the same

## Batch Migration

For migrating multiple SLOs:

### Create Migration Script

```bash
#!/bin/bash
# migrate-to-dynamic.sh

SLOS=(
  "slo-1"
  "slo-2"
  "slo-3"
)

NAMESPACE="monitoring"

for slo in "${SLOS[@]}"; do
  echo "Migrating $slo..."

  # Backup
  kubectl get slo $slo -n $NAMESPACE -o yaml > backup-$slo.yaml

  # Migrate
  kubectl patch slo $slo -n $NAMESPACE --type=merge -p '{"spec":{"burnRateType":"dynamic"}}'

  # Wait
  sleep 5
done

echo "Waiting for backend processing..."
sleep 30

echo "Migration complete. Verify in Pyrra UI."
```

### Execute Batch Migration

```bash
chmod +x migrate-to-dynamic.sh
./migrate-to-dynamic.sh
```

### Verify Batch Migration

```bash
# Check all SLOs
kubectl get slo -n monitoring -o custom-columns=NAME:.metadata.name,BURN_RATE_TYPE:.spec.burnRateType
```

## Troubleshooting

### Issue 1: Badge Still Shows "Static" After Migration

**Symptoms**:

- kubectl shows `burnRateType: dynamic`
- Pyrra UI still shows gray "Static" badge

**Causes**:

- Backend hasn't processed change yet
- Browser cache showing old data
- API not reflecting latest state

**Solutions**:

1. Wait 30-60 seconds for backend processing
2. Hard refresh browser (Ctrl+Shift+R)
3. Check backend logs for errors
4. Verify API response includes `burnRateType: dynamic`

### Issue 2: Thresholds Show "Traffic-Aware" Instead of Values

**Symptoms**:

- Badge shows "Dynamic" correctly
- Threshold column shows "Traffic-Aware" text
- No calculated values displayed

**Causes**:

- Insufficient traffic data for calculation
- Prometheus query failing
- Metric history too short

**Solutions**:

1. Wait for more traffic data to accumulate
2. Check Prometheus has metric history (30d for 30d SLO window)
3. Check browser console for query errors
4. Verify base metrics exist and have data

### Issue 3: Alert Rules Not Updated

**Symptoms**:

- UI shows dynamic burn rate
- Prometheus alert rules still use static expressions

**Causes**:

- Backend hasn't regenerated rules yet
- PrometheusRule object not updated
- Prometheus hasn't reloaded rules

**Solutions**:

1. Wait 60 seconds for backend processing
2. Check PrometheusRule object:
   ```bash
   kubectl get prometheusrule -n monitoring <slo-name> -o yaml
   ```
3. Force Prometheus reload (if needed)
4. Check backend logs for rule generation errors

### Issue 4: Migration Causes Alert Storm

**Symptoms**:

- Alerts firing immediately after migration
- More alerts than with static burn rates

**Causes**:

- Current traffic significantly above average
- Dynamic thresholds more sensitive than static
- Actual error rate exceeds new thresholds

**Solutions**:

1. Check if alerts are legitimate (real issues)
2. Review traffic patterns - is current traffic abnormal?
3. Consider if static thresholds were too lenient
4. If false positives, rollback and investigate
5. May need to adjust SLO target or window

### Issue 5: No Alerts Firing After Migration

**Symptoms**:

- Alerts stopped firing after migration
- Known issues not triggering alerts

**Causes**:

- Current traffic significantly below average
- Dynamic thresholds less sensitive than static
- Calculation error in threshold

**Solutions**:

1. Check current traffic vs average traffic
2. Verify thresholds are calculating correctly
3. Check if error rate actually below new thresholds
4. Consider if static thresholds were too aggressive
5. If missing real issues, rollback and investigate

## Best Practices

### 1. Gradual Migration

- **Start Small**: Migrate 1-2 SLOs first
- **Monitor Closely**: Watch behavior for 24-48 hours
- **Learn Patterns**: Understand how dynamic thresholds behave
- **Expand Gradually**: Migrate more SLOs after validation

### 2. Parallel Testing

- **Create Duplicate SLO**: One static, one dynamic
- **Compare Behavior**: Monitor both for same service
- **Validate Improvement**: Confirm dynamic is better
- **Switch Over**: Migrate production SLO after validation

### 3. Documentation

- **Record Migrations**: Keep log of what was migrated when
- **Document Issues**: Note any problems encountered
- **Share Learnings**: Help team understand dynamic behavior
- **Update Runbooks**: Include dynamic burn rate context

### 4. Monitoring

- **Alert Behavior**: Track alert frequency and accuracy
- **False Positive Rate**: Should decrease with dynamic
- **False Negative Rate**: Should not increase
- **Threshold Adaptation**: Verify thresholds change with traffic

### 5. Team Communication

- **Announce Migrations**: Let team know SLOs are changing
- **Explain Behavior**: Help team understand dynamic thresholds
- **Provide Context**: Share this migration guide
- **Gather Feedback**: Ask team about alert behavior changes

## FAQ

### Q: Will migration affect my error budget?

**A**: No. Error budget calculations are identical for static and dynamic burn rates. Only alert thresholds change.

### Q: Can I migrate back to static?

**A**: Yes. Rollback is simple and immediate (see Rollback Procedure section).

### Q: Do I need to migrate all SLOs?

**A**: No. Migration is opt-in per SLO. Static and dynamic SLOs coexist without issues.

### Q: Will my Grafana dashboards still work?

**A**: Yes. Grafana dashboards use generic recording rules which are identical for static and dynamic SLOs.

### Q: How long does migration take?

**A**: The kubectl command is instant. Backend processing takes ~30 seconds. Full validation takes 24-48 hours.

### Q: What if I have custom alert rules?

**A**: Custom alert rules are not affected. Only Pyrra-generated alert rules change.

### Q: Can I migrate SLOs with custom windows?

**A**: Yes. Dynamic burn rates work with any SLO window (7d, 30d, 90d, etc.).

### Q: What about SLOs with multiple indicators?

**A**: Each SLO has one indicator type. Migration works for all indicator types (ratio, latency, latency_native, bool_gauge).

### Q: Will migration cause downtime?

**A**: No. Migration is seamless with no service interruption.

### Q: How do I know if migration was successful?

**A**: Check Pyrra UI for green "Dynamic" badge and calculated threshold values. Verify alert rules in Prometheus.

## Additional Resources

- **Feature Documentation**: `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md`
- **Core Concepts**: `.dev-docs/CORE_CONCEPTS_AND_TERMINOLOGY.md`
- **Testing Guide**: `.dev-docs/TASK_7.12_MANUAL_TESTING_GUIDE.md`
- **Browser Compatibility**: `.dev-docs/BROWSER_COMPATIBILITY_MATRIX.md`
- **Requirements**: `.kiro/specs/dynamic-burn-rate-completion/requirements.md`
- **Design**: `.kiro/specs/dynamic-burn-rate-completion/design.md`

## Support

For issues or questions:

1. Check this migration guide
2. Review troubleshooting section
3. Check `.dev-docs/` documentation
4. Review Pyrra logs for errors
5. Consult with platform team

## Changelog

- **January 11, 2025**: Initial migration guide created
- **January 11, 2025**: Validated during Task 7.12 manual testing - migration and rollback procedures confirmed working
