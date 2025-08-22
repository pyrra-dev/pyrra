# Current Burn Rate Implementation Analysis

## Overview

The current implementation uses the traditional multi-window, multi-burn-rate approach as described in the [Google SRE Workbook](https://sre.google/workbook/implementing-slos/).

## Key Components

### 1. Alert Structure
```go
type MultiBurnRateAlert struct {
    Severity string
    Short    time.Duration
    Long     time.Duration
    For      time.Duration
    Factor   float64
    QueryShort string
    QueryLong  string
}
```

### 2. Alert Rule Generation
The core alerting logic uses this expression pattern:
```
<burn_rate_short>{labels} > (factor * (1-target)) 
  and 
<burn_rate_long>{labels} > (factor * (1-target))
```

Where:
- `factor` is a static burn rate multiplier
- `target` is the SLO target (e.g., 0.99 for 99% availability)
- Short and long windows are used together to reduce alert flappiness

### 3. Implementation Location
- Main implementation: `slo/rules.go`
- Key methods:
  - `Objective.Burnrates()`: Generates the burn rate recording and alerting rules
  - `Objective.Burnrate()`: Generates the PromQL for calculating burn rates
  - `Windows()`: Defines the static burn rate windows and factors

## Differences from Dynamic Burn Rate

The current implementation differs from our dynamic burn rate approach in several key ways:

1. **Static vs Dynamic Factors**
   - Current: Uses predefined static factors (e.g., 14.4, 6, 3)
   - Dynamic: Factor calculated from ratio of events in different windows

2. **Window Pairs**
   - Current: Uses pairs of windows (short+long) to reduce noise
   - Dynamic: Uses single windows with dynamic thresholds

3. **Threshold Calculation**
   - Current: `threshold = factor * (1-target)`
   - Dynamic: `threshold = (N_slo/N_alert) * error_budget`

## Next Steps

To implement dynamic burn rate alerting, we need to:

1. Extend the SLO spec to support dynamic burn rate configuration
2. Modify the alert rule generation to support dynamic thresholds
3. Update the PromQL generation to calculate event ratios
4. Ensure backward compatibility with existing static burn rates

_Last updated: 2025-08-22_
