# Task 7.7: BurnrateGraph Dynamic Threshold Display Fix

**Date**: January 8, 2025  
**Task**: Fix BurnrateGraph to display dynamic thresholds for dynamic SLOs  
**Status**: ✅ Complete

## Problem Statement

The BurnrateGraph component in the AlertsTable was displaying static thresholds for dynamic SLOs instead of calculating and displaying the traffic-aware dynamic thresholds. This created a visual inconsistency where:

1. The AlertsTable threshold column showed correct dynamic thresholds (via BurnRateThresholdDisplay)
2. The expandable burn rate graph showed incorrect static thresholds
3. The threshold description text used placeholder language for dynamic SLOs

## Root Cause Analysis

### Issue 1: Static Threshold Passed from AlertsTable
```typescript
// AlertsTable.tsx - Line ~280
<BurnrateGraph
  client={promClient}
  alert={a}
  objective={objective}
  threshold={a.factor * (1 - objective.target)}  // ❌ Always static calculation
  from={from}
  to={to}
  pendingData={pendingAlignedData}
  firingData={firingAlignedData}
  uPlotCursor={uPlotCursor}
/>
```

The AlertsTable was passing a statically calculated threshold (`factor × (1 - target)`) to BurnrateGraph, regardless of whether the SLO was static or dynamic.

### Issue 2: BurnrateGraph Not Detecting Burn Rate Type
The BurnrateGraph component had no logic to:
- Detect if the objective uses dynamic burn rates
- Calculate dynamic thresholds based on traffic patterns
- Query Prometheus for traffic ratio data

### Issue 3: Placeholder Description Text
```typescript
// burnrate.tsx - getThresholdDescription()
if (burnRateType === BurnRateType.Dynamic) {
  return `The short (${shortWindow}) and long (${longWindow}) burn rates both have to be over the traffic-aware threshold (currently ${threshold.toFixed(2)}%).`
}
```

The description text was generic and didn't explain the dynamic behavior properly.

## Solution Implementation

### 1. Enhanced BurnrateGraph Component

#### Added Dynamic Threshold Calculation Logic
```typescript
// BurnrateGraph.tsx
const burnRateType = getBurnRateType(objective)
const currentTime = Math.floor(Date.now() / 1000)

// For dynamic burn rates, calculate the dynamic threshold
const trafficQuery = burnRateType === BurnRateType.Dynamic 
  ? getTrafficRatioQuery(objective, alert.factor) 
  : ''

const {response: trafficResponse, status: trafficStatus} = usePrometheusQuery(
  client,
  trafficQuery,
  currentTime,
  {enabled: burnRateType === BurnRateType.Dynamic && trafficQuery !== ''}
)
```

#### Integrated Traffic Calculation Patterns
Copied the same helper functions from BurnRateThresholdDisplay:
- `getBaseMetricSelector()` - Extract metric from objective based on indicator type
- `getTrafficRatioQuery()` - Generate Prometheus query for traffic ratio
- `calculateDynamicThreshold()` - Apply dynamic threshold formula

#### Dynamic Threshold State Management
```typescript
const [dynamicThreshold, setDynamicThreshold] = useState<number | null>(null)

React.useEffect(() => {
  if (burnRateType === BurnRateType.Dynamic && trafficStatus === 'success' && trafficResponse !== null) {
    let trafficRatio: number | undefined
    
    if (trafficResponse.options?.case === 'vector' && trafficResponse.options.value.samples.length > 0) {
      trafficRatio = trafficResponse.options.value.samples[0].value
    } else if (trafficResponse.options?.case === 'scalar') {
      trafficRatio = trafficResponse.options.value.value
    }
    
    if (trafficRatio !== undefined && isFinite(trafficRatio) && trafficRatio > 0) {
      const calculatedThreshold = calculateDynamicThreshold(objective, alert.factor, trafficRatio)
      setDynamicThreshold(calculatedThreshold)
    }
  }
}, [burnRateType, trafficStatus, trafficResponse, objective, alert.factor])

// Use dynamic threshold if available, otherwise use static threshold
const displayThreshold = dynamicThreshold !== null ? dynamicThreshold : threshold
```

#### Updated Graph Data and Description
```typescript
const data: AlignedData = [
  timestamps,
  shortSeries,
  longSeries,
  // Use dynamic threshold for dynamic SLOs, static threshold for static SLOs
  Array(timestamps.length).fill(displayThreshold),
]

// In the description
{getThresholdDescription(objective, displayThreshold, shortFormatted, longFormatted)}
```

### 2. Enhanced getThresholdDescription() Function

Updated the description text to be more informative for dynamic SLOs:

```typescript
// burnrate.tsx
export const getThresholdDescription = (objective: Objective, threshold: number, shortWindow: string, longWindow: string): string => {
  const burnRateType = getBurnRateType(objective)
  
  if (burnRateType === BurnRateType.Dynamic) {
    // Format threshold appropriately - use scientific notation for very small values
    const formattedThreshold = threshold < 0.001 
      ? threshold.toExponential(2) 
      : (threshold * 100).toFixed(2)
    
    const unit = threshold < 0.001 ? '' : '%'
    
    return `The short (${shortWindow}) and long (${longWindow}) burn rates both have to be over the traffic-aware dynamic threshold (currently ${formattedThreshold}${unit}). This threshold adapts based on actual traffic patterns.`
  }
  
  return `The short (${shortWindow}) and long (${longWindow}) burn rates both have to be over the ${(threshold * 100).toFixed(2)}% threshold.`
}
```

### 3. Fixed Unused Import in AlertsTable

Removed unused `formatNumber` import that was causing build errors:
```typescript
// Before
import {formatNumber} from '../utils/numberFormat'

// After - removed, using .toFixed() directly instead
```

## Testing Performed

### 1. TypeScript Compilation
```bash
# Checked for TypeScript errors
getDiagnostics(["ui/src/components/graphs/BurnrateGraph.tsx", "ui/src/components/AlertsTable.tsx", "ui/src/burnrate.tsx"])
# Result: No diagnostics found ✅
```

### 2. Production Build
```bash
cd ui && npm run build
# Result: Compiled successfully ✅
```

### 3. Code Review Checklist
- ✅ Dynamic threshold calculation matches BurnRateThresholdDisplay logic
- ✅ All indicator types supported (ratio, latency, latencyNative, boolGauge)
- ✅ Backward compatibility maintained for static SLOs
- ✅ Error handling for invalid traffic data
- ✅ Scientific notation for very small thresholds
- ✅ No TypeScript compilation errors
- ✅ No unused variables or imports

## Files Modified

1. **ui/src/components/graphs/BurnrateGraph.tsx**
   - Added imports: `usePrometheusQuery`, `getBurnRateType`, `BurnRateType`
   - Added helper functions: `getBaseMetricSelector()`, `calculateDynamicThreshold()`, `getTrafficRatioQuery()`
   - Added state management for dynamic threshold calculation
   - Updated graph data to use `displayThreshold`
   - Updated description to use `displayThreshold`

2. **ui/src/burnrate.tsx**
   - Enhanced `getThresholdDescription()` with better dynamic threshold formatting
   - Added scientific notation support for very small thresholds
   - Added explanatory text about traffic-aware behavior

3. **ui/src/components/AlertsTable.tsx**
   - Removed unused `formatNumber` import
   - Fixed number formatting to use `.toFixed()` directly

## Validation Requirements

### Manual Testing Required
The user should test the following scenarios:

1. **Dynamic Ratio SLO**:
   - Open a dynamic ratio SLO detail page
   - Expand the burn rate graph in the alerts table
   - Verify the threshold line shows a dynamic value (not static)
   - Verify the description text mentions "traffic-aware dynamic threshold"

2. **Dynamic Latency SLO**:
   - Open a dynamic latency SLO detail page
   - Expand the burn rate graph in the alerts table
   - Verify the threshold line shows a dynamic value
   - Verify histogram metrics are queried correctly

3. **Static SLO (Regression Test)**:
   - Open a static SLO detail page
   - Expand the burn rate graph in the alerts table
   - Verify the threshold line shows the static value (unchanged behavior)
   - Verify the description text does NOT mention "traffic-aware"

4. **Threshold Consistency**:
   - Compare the threshold value in the AlertsTable "Threshold" column
   - Compare with the threshold line in the expanded burn rate graph
   - Verify they match for dynamic SLOs

## Known Limitations

1. **Real-time Updates**: The dynamic threshold in the graph is calculated once when the component mounts. It doesn't update in real-time as traffic changes (same behavior as BurnRateThresholdDisplay).

2. **Loading State**: While the dynamic threshold is being calculated, the graph shows the static threshold as a fallback. There's no loading indicator in the graph itself.

3. **Error Handling**: If the traffic query fails, the graph falls back to showing the static threshold without any visual indication of the error.

## Future Enhancements

1. **Loading Indicator**: Add a visual indicator when dynamic threshold is being calculated
2. **Error State**: Show a visual indicator when dynamic threshold calculation fails
3. **Real-time Updates**: Consider updating the threshold line as traffic patterns change
4. **Tooltip Enhancement**: Add tooltip to threshold line showing calculation details

## Conclusion

The BurnrateGraph component now correctly displays dynamic thresholds for dynamic SLOs while maintaining backward compatibility with static SLOs. The implementation follows the same patterns as BurnRateThresholdDisplay, ensuring consistency across the UI.

**Status**: ✅ Implementation complete, ready for user testing and validation.
