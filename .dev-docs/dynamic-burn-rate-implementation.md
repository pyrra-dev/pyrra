# Dynamic Burn Rate Implementation Documentation

## Project Context
Pyrra is a project that helps create and manage Service Level Objectives (SLOs) in Kubernetes environments using Prometheus metrics. It:
- Creates recording and alerting rules for SLOs
- Supports different types of SLOs (ratio, latency, boolean)
- Integrates with Prometheus Operator for rule deployment
- Provides both static and dynamic burn rate calculations

## Current Implementation Status

### Core Concepts
1. **Burn Rate**: The rate at which an error budget is being consumed
2. **Error Budget**: Allowed error margin within the SLO target (e.g., if target is 99%, budget is 1%)
3. **Window Periods**: Time windows for short and long-term alerting
   - Short windows: 5m, 30m, 2h, 6h
   - Long windows: 1h, 6h, 1d, 4d

### Dynamic Burn Rate Implementation
We've implemented dynamic burn rates that scale based on window size:

1. **Error Budget Percentages**:
```go
// For 1 hour window:
// Want 50% per day = 1/48 per hour
errorBudgetBurnPercent = 1.0 / 48

// For 6 hour window:
// Want 100% per 4 days = 1/16 per 6 hours
errorBudgetBurnPercent = 1.0 / 16

// For 1 day window:
// Want budget burn of 1/14 per day
errorBudgetBurnPercent = 1.0 / 14

// For 4 day window:
// Want budget burn of 1/7 per 4 days
errorBudgetBurnPercent = 1.0 / 7
```

2. **Calculation Formula**:
```
(increase[slo_window] / increase[alert_window]) * error_budget_percent
```

### Key Files
1. `slo/rules.go`:
   - Contains core implementation
   - Includes `dynamicBurnRateExpr()` and `DynamicWindows()`
   - Handles alert rule generation

2. `slo/slo.go`:
   - Defines core types and interfaces
   - Contains `Objective` struct and methods
   - Defines alerting configuration

3. `kubernetes/api/v1alpha1/servicelevelobjective_types.go`:
   - Defines Kubernetes CRD types
   - Contains API configuration options

4. `.dev/test-slo.yaml`:
   - Test SLO configuration for development
   - Used for local testing

## Implementation Progress

### Completed
1. ✅ Added `BurnRateType` field to support both static and dynamic calculations
2. ✅ Implemented `dynamicBurnRateExpr()` for burn rate calculation
3. ✅ Added proper error budget percentages in `DynamicWindows()`
4. ✅ Preserved existing window periods

### In Progress/TODO
1. Update alert expressions to use dynamic burn rates
   - Replace static factors with dynamic percentages
   - Update tests to verify new alert expressions

2. Add specific tests for dynamic burn rates:
   - Test different window periods
   - Verify error budget calculations
   - Test edge cases

3. Add documentation:
   - User guide for dynamic vs static burn rates
   - Example configurations
   - Migration guide

## Testing Strategy
1. Unit Tests:
   - Test burn rate calculations
   - Test window period calculations
   - Test alert rule generation

2. Integration Tests:
   - Test with Prometheus
   - Verify alert behavior
   - Test different SLO configurations

3. Manual Testing:
```yaml
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: test-slo
spec:
  target: 0.99
  window: 28d
  alerting:
    burnRateType: dynamic  # Enable dynamic burn rates
  indicator:
    ratio:
      errors:
        metric: http_requests_total{code=~"5.."}
      total:
        metric: http_requests_total
```

## Debugging Notes
1. Use `kubectl apply -f .dev/test-slo.yaml` to deploy test SLO
2. Check generated rules in Prometheus
3. Monitor alert behavior with different error rates

## Useful Commands
```bash
# Deploy test SLO
kubectl apply -f .dev/test-slo.yaml

# Check Prometheus rules
kubectl get prometheusrules -n monitoring

# Run unit tests
go test ./slo/...

# Build and run locally
make run
```

## References
1. [Multi-window, Multi-burn-rate Alerts](https://sre.google/workbook/alerting-on-slos/)
2. [Prometheus Operator Documentation](https://prometheus-operator.dev/)
3. [SLO Alerting Principles](https://landing.google.com/sre/workbook/chapters/alerting-on-slos/)

## Next Steps
1. Continue implementation:
   - Update alert expressions
   - Add tests
   - Update documentation
2. Review and testing
3. Create PR with comprehensive description
4. Add migration guide for users

Remember to:
- Test with different SLO targets
- Verify burn rate calculations
- Check alert behavior
- Update user documentation
