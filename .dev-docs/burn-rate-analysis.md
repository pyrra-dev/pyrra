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
   - Both approaches use pairs of windows (short+long) to reduce noise and alert flappiness
   - The long window is used both for factor calculation and burn rate verification
   - The SAME dynamic factor (calculated from the long window) is applied to both windows

3. **Factor (Burn Rate) Calculation**
   - Current: Uses static predefined factors
   - Dynamic: Factor is calculated as `(N_slo/N_long) * error_budget`
   - This SAME factor is used for both windows in the alert expression:
     ```
     burn_rate_short > factor AND burn_rate_long > factor
     ```
   - Using the same factor ensures consistency in how we measure error budget burn rate across different time windows

## Next Steps

To implement dynamic burn rate alerting, we need to:

1. Extend the SLO spec to support dynamic burn rate configuration
2. Modify the alert rule generation to support dynamic thresholds
3. Update the PromQL generation to calculate event ratios
4. Ensure backward compatibility with existing static burn rates

_Last updated: 2025-08-22_
