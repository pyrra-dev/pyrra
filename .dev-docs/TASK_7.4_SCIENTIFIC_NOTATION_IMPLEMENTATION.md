# Task 7.4: Scientific Notation Implementation

## Overview

Implemented scientific notation formatting for very small numbers in the Pyrra UI to fix truncation issues, especially for high SLO targets (e.g., 99.99%) that produce very small threshold values.

## Implementation Summary

### 1. Created Utility Functions (`ui/src/utils/numberFormat.ts`)

**`formatNumber(value, decimalPlaces)`**
- Used for burn rate values in AlertsTable
- Rules:
  - `< 0.001`: Scientific notation (e.g., `1.23e-5`)
  - `>= 0.001`: Fixed decimal with specified places (default 3)

**`formatThreshold(value)`**
- Used for threshold values in BurnRateThresholdDisplay and graphs
- Rules:
  - `< 0.001`: Scientific notation (e.g., `1.23e-5`)
  - `>= 0.001 and < 100`: 3 decimal places (e.g., `0.123`, `99.995`)
  - `>= 100`: 2 decimal places (e.g., `123.45`)

### 2. Updated Components

**AlertsTable (`ui/src/components/AlertsTable.tsx`)**
- Updated short burn and long burn columns to use `formatNumber()`
- Very small burn rates now display in scientific notation

**BurnRateThresholdDisplay (`ui/src/components/BurnRateThresholdDisplay.tsx`)**
- Replaced inline formatting logic with `formatThreshold()`
- Consistent formatting across all threshold displays
- Added tooltips for scientific notation values

**BurnrateGraph (`ui/src/components/graphs/BurnrateGraph.tsx`)**
- Updated short burn and long burn series tooltips to use `formatNumber()`
- Updated threshold line tooltip to use `formatNumber()`
- Updated Y-axis labels to use `formatNumber()`
- Consistent formatting in graph hover values and legends

**ErrorsGraph (`ui/src/components/graphs/ErrorsGraph.tsx`)**
- Updated error rate series tooltips to use `formatNumber()`
- Scientific notation for very small error rates

**ErrorBudgetGraph (`ui/src/components/graphs/ErrorBudgetGraph.tsx`)**
- Updated error budget series tooltips to use `formatNumber()`
- Updated Y-axis labels to use `formatNumber()`
- Scientific notation for very small budget percentages

**RequestsGraph (`ui/src/components/graphs/RequestsGraph.tsx`)**
- Updated request rate series tooltips to use `formatNumber()`
- Updated baseline series tooltip to use `formatNumber()`
- Updated traffic ratio calculations to use `formatNumber()`
- Scientific notation for very small request rates

### 3. Test Files

**Unit Tests (`ui/src/utils/numberFormat.spec.ts`)**
- Comprehensive tests for both formatting functions
- Tests edge cases (NaN, Infinity, zero, negative numbers)
- Tests high SLO target scenarios (99.99%)

**Python Validation Script (`scripts/test_scientific_notation.py`)**
- Calculates expected threshold values for 99.99% SLO target
- Demonstrates when scientific notation should be used
- Shows traffic ratio calculations (N_SLO / N_alert)

**Test SLO (`dev/test-high-target-slo.yaml`)**
- 99.99% SLO target for testing scientific notation
- Dynamic burn rate type
- Uses apiserver metrics

## Key Insights from Implementation

### Traffic Ratio Understanding

**IMPORTANT**: The traffic ratio is NOT a multiplier around 1.0!

- Traffic ratio = `N_SLO / N_alert` (actual event count ratio)
- For 30 day SLO window and 1 hour alert window: ratio ≈ 720 at steady traffic
- Low traffic (50%): ratio ≈ 360
- High traffic (200%): ratio ≈ 1440

### When Scientific Notation is Used

For 99.99% SLO target with steady traffic:
- **Factor 14** (1h4m window): threshold ≈ 0.0014 → Uses 3 decimals: `0.001`
- **Factor 7** (6h26m window): threshold ≈ 0.0007 → Scientific: `7.000e-4`
- **Factor 2** (1d1h43m window): threshold ≈ 0.0002 → Scientific: `2.000e-4`
- **Factor 1** (4d6h51m window): threshold ≈ 0.0001 → Scientific: `1.000e-4`

With low traffic (50% of steady), even Factor 14 uses scientific notation.

## Testing Instructions

### 1. Run Unit Tests

```bash
cd ui
npm test -- numberFormat.spec.ts --watchAll=false
```

All 11 tests should pass.

### 2. Run Python Validation Script

```bash
python scripts/test_scientific_notation.py
```

This shows expected threshold values and when scientific notation should be used.

### 3. Test in Development UI

**Start the development UI:**
```bash
cd ui
npm start
```

**Navigate to:** http://localhost:3000

**Test Cases:**

a. **Normal SLO (99.5% target)**
   - Check AlertsTable: burn rates should show 3 decimal places
   - Check threshold column: should show normal decimal notation
   - Example: `0.010` or `0.005`

b. **High Target SLO (99.99% target)**
   - Apply `.dev/test-high-target-slo.yaml` to cluster
   - Check AlertsTable: look for scientific notation in burn rate columns
   - Check threshold column: should show scientific notation for factors 7, 2, 1
   - Example: `7.029e-4` or `2.000e-4`

c. **Hover over thresholds**
   - Tooltips should explain scientific notation values
   - Example: "Very small threshold: 7.029e-4"

d. **Check BurnrateGraph**
   - Expand an alert row to see the graph
   - Hover over short burn, long burn, and threshold lines
   - Should show formatted values with scientific notation if < 0.001
   - Check Y-axis labels for scientific notation

e. **Check other graphs (Errors, Error Budget, Requests)**
   - Navigate to detail page for an SLO
   - Hover over graph lines
   - Check tooltips and legends for scientific notation
   - Verify Y-axis labels use scientific notation when appropriate

### 4. Visual Verification Checklist

**In Pyrra UI (port 3000), verify:**

- [ ] AlertsTable short burn column shows scientific notation for very small values
- [ ] AlertsTable long burn column shows scientific notation for very small values
- [ ] Threshold column shows scientific notation for high SLO targets
- [ ] Tooltips explain scientific notation values
- [ ] BurnrateGraph tooltips (short, long, threshold) show scientific notation
- [ ] BurnrateGraph Y-axis labels show scientific notation
- [ ] ErrorsGraph tooltips show scientific notation for small error rates
- [ ] ErrorBudgetGraph tooltips and Y-axis show scientific notation
- [ ] RequestsGraph tooltips show scientific notation for small request rates
- [ ] No truncated numbers (e.g., no `0.000` when value is actually `0.0001`)
- [ ] Consistent formatting across all components

## Files Modified

1. `ui/src/utils/numberFormat.ts` - New utility functions
2. `ui/src/utils/numberFormat.spec.ts` - New unit tests
3. `ui/src/components/AlertsTable.tsx` - Updated burn rate formatting
4. `ui/src/components/BurnRateThresholdDisplay.tsx` - Updated threshold formatting
5. `ui/src/components/graphs/BurnrateGraph.tsx` - Updated all series tooltips and axis labels
6. `ui/src/components/graphs/ErrorsGraph.tsx` - Updated error rate tooltips
7. `ui/src/components/graphs/ErrorBudgetGraph.tsx` - Updated budget tooltips and axis labels
8. `ui/src/components/graphs/RequestsGraph.tsx` - Updated request rate tooltips
9. `scripts/test_scientific_notation.py` - New validation script
10. `.dev/test-high-target-slo.yaml` - New test SLO configuration

## Next Steps

After visual verification in the UI:

1. Test with actual high-target SLOs in your cluster
2. Verify scientific notation appears correctly in all contexts
3. Check that tooltips are helpful and informative
4. Consider if any other UI components need similar formatting

## Requirements Satisfied

- ✅ **Requirement 5.1**: Enhanced UI displays with scientific notation
- ✅ **Requirement 5.3**: Improved number formatting for very small values
- ✅ Fixed truncated numbers in AlertsTable (short burn, long burn columns)
- ✅ Fixed truncated numbers in BurnRateThresholdDisplay
- ✅ Fixed truncated numbers in BurnrateGraph tooltips
- ✅ Implemented simple rule: if number < 0.001, show scientific notation
- ✅ Created test SLO with high target (99.99%)
- ✅ Created validation scripts and unit tests
