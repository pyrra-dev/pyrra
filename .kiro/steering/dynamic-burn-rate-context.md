---
inclusion: fileMatch
fileMatchPattern: '*burn*rate*'
---

# Dynamic Burn Rate Feature Context

## Feature Purpose and Innovation

Dynamic burn rate alerting represents a significant advancement in SLO monitoring by making alert thresholds traffic-aware. Unlike traditional static burn rates that use fixed multipliers regardless of traffic volume, dynamic burn rates adapt based on actual observed traffic patterns.

### Problem Solved
- **False Positives**: Static alerts fire unnecessarily during low traffic periods when small absolute error counts appear concerning but don't threaten error budget
- **False Negatives**: Static alerts miss problems during high traffic when error rates seem acceptable but rapidly consume error budget

### Mathematical Innovation
The core formula adapts the Google SRE multi-window alerting approach:
```
Traditional: burn_rate > FIXED_FACTOR × (1 - SLO_target)
Dynamic:     error_rate > (N_SLO / N_alert) × E_budget_percent × (1 - SLO_target)
```

The traffic scaling factor `(N_SLO / N_alert)` makes thresholds traffic-aware:
- **High Traffic**: Lower threshold (more errors needed to alert)
- **Low Traffic**: Higher threshold (fewer errors needed to alert)

## Implementation Architecture

### Backend Components
1. **Alert Expression Generation** (`slo/rules.go`):
   - `buildAlertExpr()` - Routes between static/dynamic modes
   - `buildDynamicAlertExpr()` - Generates traffic-aware expressions
   - Support for all indicator types: Ratio, Latency, LatencyNative, BoolGauge

2. **Window Configuration** (`slo/rules.go`):
   - `DynamicWindows()` - Maps static factors to error budget percentages
   - Maintains proportional scaling for any SLO period
   - Consistent multi-window logic using N_long for both windows

3. **CRD Extensions** (`kubernetes/api/v1alpha1/`):
   - `burnRateType` field enables dynamic mode
   - Backward compatible (static remains default)
   - Proper OpenAPI schema generation

### UI Components
1. **Badge System** (`ui/src/burnrate.tsx`):
   - Green "Dynamic" badges with eye icons
   - Gray "Static" badges with lock icons
   - Context-aware tooltips explaining behavior

2. **Threshold Display** (`ui/src/components/BurnRateThresholdDisplay.tsx`):
   - Real-time Prometheus queries for dynamic calculations
   - Static formula calculations for traditional SLOs
   - Graceful error handling and loading states

3. **Alert Integration** (`ui/src/components/AlertsTable.tsx`):
   - Enhanced tooltips showing threshold behavior
   - Traffic-aware descriptions in burn rate graphs
   - Consistent visual indicators across components

## Current Implementation Status

### Completed Components ✅
- **All Indicator Types**: Ratio, Latency, LatencyNative, BoolGauge fully implemented
- **Mathematical Correctness**: Formula validated with real Prometheus data
- **UI Integration**: Complete badge system, threshold display, tooltips
- **API Integration**: Full protobuf transmission of burn rate type
- **Testing Infrastructure**: Comprehensive test coverage for core functionality

### Validation Progress 🚧
- **Ratio Indicators**: ✅ Complete validation (UI + backend)
- **Latency Indicators**: ✅ Backend complete, 🚧 UI comprehensive validation pending
- **Other Indicators**: 🔜 LatencyNative and BoolGauge validation needed
- **Edge Cases**: 🔜 Missing metrics, insufficient data scenarios
- **Alert Firing**: 🔜 End-to-end alert validation with synthetic metrics

## Key Technical Decisions

### Multi-Window Consistency
Both short and long alert windows use the long window period (N_long) for traffic scaling. This ensures consistent burn rate measurement across time scales, matching the behavior of static burn rates.

### Recording Rule Optimization
Dynamic expressions use existing recording rules rather than inline calculations for performance:
```promql
# Uses optimized recording rules
sum(increase(pyrra:slo_errors:rate5m[1h])) / sum(increase(pyrra:slo_requests:rate5m[1h]))
```

### Error Budget Percentage Mapping
Static factors map to error budget percentages based on Google SRE recommendations:
- Factor 14 → 1/48 (50% budget consumption per day)
- Factor 7 → 1/16 (100% budget consumption per 4 days)
- Factor 2 → 1/14 (warning threshold)
- Factor 1 → 1/7 (secondary warning)

### Backward Compatibility Strategy
- Static burn rate remains default behavior
- All existing SLOs continue working unchanged
- Dynamic mode explicitly opt-in via `burnRateType: dynamic`
- No breaking changes to existing APIs or configurations

## Development Workflow Considerations

### Local Development Setup
Requires three running services:
1. `kubectl port-forward -n monitoring svc/monitoring-kube-prometheus-prometheus 9090:9090`
2. `./pyrra kubernetes` (backend on port 9444)
3. `./pyrra api --prometheus-url=http://localhost:9090` (API + UI on port 9099)

### Testing Environment
Mixed SLO environment ideal for validation:
- 14+ static SLOs from monitoring namespace
- 1+ dynamic test SLO for comparison
- Real Prometheus data with error patterns

### Windows Development Notes
- CRD generation requires workaround: `controller-gen crd paths="./kubernetes/api/v1alpha1"`
- Standard `make generate` has path globbing issues on Windows
- Manual tool installation may be required

## Next Development Priorities

1. **Session 10C**: Complete latency indicator comprehensive validation
2. **Session 12**: Missing metrics and edge case resilience testing  
3. **Session 11**: LatencyNative and BoolGauge indicator validation
4. **Sessions 13-15**: Alert firing validation, UI polish, production readiness

Each session should follow systematic comparison methodology and update comprehensive documentation with findings.