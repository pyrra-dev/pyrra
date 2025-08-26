# Dynamic Burn Rate - Core Concepts and Terminology

## Overview

This document provides precise definitions and explanations of the core concepts underlying Pyrra's dynamic burn rate implementation. These concepts are fundamental to understanding how the feature works and should be referenced in all implementation discussions.

## Key Terminology

### 1. Error Rate
**Definition**: The actual percentage of requests that are failing at any given moment.
- **Example**: 0.02 = 2% of current requests are errors
- **PromQL**: `sum(rate(errors[5m])) / sum(rate(total[5m]))`
- **Note**: This is a **measured value** that changes based on actual system behavior

### 2. Error Budget
**Definition**: The allowed error rate for meeting your SLO target.
- **Formula**: `error_budget = 1 - SLO_target`
- **Example**: For 99% SLO target → error_budget = 1 - 0.99 = 0.01 (1% error rate allowed)
- **Note**: This is a **fixed value** defined by your SLO configuration

### 3. Burn Rate
**Definition**: The rate at which the error budget is being consumed relative to the SLO period.
- **Formula**: `burn_rate = error_rate / error_budget`
- **Interpretation**:
  - `burn_rate = 1.0` → Consuming error budget at exactly the rate that would exhaust it by the end of the SLO period
  - `burn_rate = 2.0` → Consuming error budget at 2x the rate (would exhaust budget in half the SLO period)
  - `burn_rate = 0.5` → Consuming error budget at half the rate (budget would last twice the SLO period)

### 4. Static vs Dynamic Burn Rate Thresholds

#### Static Burn Rate (Traditional)
**Definition**: Fixed thresholds that don't adapt to traffic patterns.
- **Alert Condition**: `burn_rate > fixed_factor * (1 - SLO_target)`
- **Problem**: Same threshold regardless of traffic volume
- **Example**: Always alert when burn_rate > 14.4 * 0.01 = 0.144

#### Dynamic Burn Rate (Adaptive)
**Definition**: Thresholds that adapt based on actual traffic patterns in different time windows.
- **Alert Condition**: `error_rate > dynamic_threshold`
- **Dynamic Threshold Formula**: `(N_SLO / N_alert) × E_budget_percent_threshold × (1 - SLO_target)`
- **Key Innovation**: The burn rate threshold itself is dynamic and traffic-aware

## The Dynamic Formula Explained

### Core Formula Components

```
dynamic_threshold = (N_SLO / N_alert) × E_budget_percent_threshold × (1 - SLO_target)
```

**Where**:
- **N_SLO** = Number of events in the SLO window (e.g., 7 days, 28 days)
- **N_alert** = Number of events in the alert window (e.g., 1 hour, 6 hours)
- **E_budget_percent_threshold** = Constant percentage of error budget consumption to alert on
- **(1 - SLO_target)** = Error budget (e.g., 0.01 for 99% SLO)

### Traffic Scaling Factor: (N_SLO / N_alert)

This ratio is the **key innovation** that makes the system traffic-aware:

#### High Traffic Scenario
- **N_SLO** = 1,000,000 events (in 7 days)
- **N_alert** = 10,000 events (in 1 hour)
- **Scaling Factor** = 1,000,000 / 10,000 = **100**
- **Result**: **Lower burn rate threshold** → requires more errors to trigger alert

#### Low Traffic Scenario  
- **N_SLO** = 1,000,000 events (in 7 days)
- **N_alert** = 100 events (in 1 hour)
- **Scaling Factor** = 1,000,000 / 100 = **10,000**
- **Result**: **Higher burn rate threshold** → requires fewer errors to trigger alert

### Error Budget Percent Thresholds (Constants)

These represent **what percentage of error budget consumption rate** we want to alert on:

| Window Period | E_budget_percent_threshold | Meaning |
|---------------|---------------------------|---------|
| 1 hour        | 1/48 (≈0.020833)        | Alert when consuming error budget at ~2.1% per hour |
| 6 hours       | 1/16 (≈0.0625)          | Alert when consuming error budget at ~6.3% per 6-hour period |
| 1 day         | 1/14 (≈0.071429)        | Alert when consuming error budget at ~7.1% per day |
| 4 days        | 1/7 (≈0.142857)         | Alert when consuming error budget at ~14.3% per 4-day period |

**Key Point**: These are **predefined constants**, not calculated values. They remain the same regardless of SLO configuration.

## Alert Behavior and False Positives/Negatives

### Alert States Definition
- **"Positive"** = Alert firing (we have an alert)
- **"Negative"** = No alert (alert not firing)
- **"False Positive"** = Alert fires when it shouldn't (unnecessary alert)
- **"False Negative"** = Alert doesn't fire when it should (missed alert)

### How Dynamic Burn Rate Prevents Issues

#### Low Traffic Periods
- **Dynamic Behavior**: Higher burn rate threshold
- **Prevention**: **False Positives** - Small absolute numbers of errors during low traffic might look concerning but aren't actually consuming error budget at a dangerous rate
- **Example**: 2 errors out of 10 requests (20% error rate) might trigger static alerts, but if this represents tiny absolute traffic, it's not actually threatening the error budget

#### High Traffic Periods
- **Dynamic Behavior**: Lower burn rate threshold  
- **Prevention**: **False Negatives** - Ensures we catch problems during high traffic when we need more absolute errors to trigger, but the error budget consumption rate is still dangerous
- **Example**: 1000 errors out of 100,000 requests (1% error rate) might not trigger static alerts, but during high traffic this could rapidly exhaust the error budget

## Mathematical Relationship Summary

The key insight is that **the burn rate threshold itself is dynamic**:

```
Static Approach:   burn_rate > FIXED_THRESHOLD
Dynamic Approach:  burn_rate > ADAPTIVE_THRESHOLD_BASED_ON_TRAFFIC
```

This creates **traffic-aware scaling** where:
1. The system calculates what constitutes a "dangerous" error budget consumption rate
2. This calculation adapts based on current traffic patterns
3. The resulting thresholds prevent both false positives (low traffic) and false negatives (high traffic)

## Implementation Notes

### PromQL Pattern
The implementation compares the actual error rate against the dynamically calculated threshold:

```promql
error_rate > ((N_SLO / N_alert) × E_budget_percent_threshold × (1 - SLO_target))
```

Where:
- `error_rate` is calculated using recording rules for efficiency
- `N_SLO / N_alert` is calculated inline using `increase()` functions
- The result is a traffic-adaptive alert threshold

### Multi-Window Logic (Critical Detail)
**Important**: Both short and long windows use **N_long** (long window period) in the denominator for traffic scaling:
- Short window alert: `error_rate_short > (N_SLO / N_long) × E_budget_percent × (1 - SLO_target)`
- Long window alert: `error_rate_long > (N_SLO / N_long) × E_budget_percent × (1 - SLO_target)`

This ensures **consistent burn rate measurement** across different time scales, similar to how static burn rates use the same factor for both windows.

### Window.Factor Dual Purpose Design
The `Window.Factor` field serves different semantic purposes based on burn rate type:

**Static Mode**: `Factor = Static burn rate multiplier`
- Example values: 14, 7, 2, 1
- Usage: `burnrate > factor × (1 - SLO_target)`

**Dynamic Mode**: `Factor = E_budget_percent_threshold` 
- Example values: 1/48, 1/16, 1/14, 1/7
- Usage: `error_rate > ((N_SLO / N_alert) × factor × (1 - SLO_target))`

### Window Period Scaling
Window periods are **automatically scaled** based on the configured SLO period via the `Windows(sloWindow)` function:
- **Base calculation**: `(sloWindow / 28days) × base_period`
- **Example**: For 7-day SLO, 1-hour window becomes 15 minutes
- **E_budget_percent_thresholds remain constant** regardless of SLO period scaling

---

**Document Purpose**: Provide authoritative definitions for all dynamic burn rate discussions and implementations.
**Last Updated**: August 26, 2025
**Audience**: Development team, code reviewers, future maintainers
