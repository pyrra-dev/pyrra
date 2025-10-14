# Dynamic Burn Rate UI Integration - Historical Design Document

**Status:** Historical reference - Implementation has evolved beyond this document  
**Current Status:** See `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md` for up-to-date implementation details

This document describes the initial UI design and implementation plan for Pyrra's Dynamic Burn Rate feature.

## Overview

The Dynamic Burn Rate feature allows SLOs to use traffic-aware alerting thresholds that adapt based on request volume, providing more accurate alerting compared to traditional static burn rates.

## UI Components Added

### 1. Burn Rate Type Display
- **Location**: SLO List page and Detail pages
- **Purpose**: Shows whether an SLO uses static or dynamic burn rates
- **Visual Elements**:
  - Badge with color coding (Green for Dynamic, Gray for Static)
  - Icons (Eye for Dynamic, Lock for Static) 
  - Informative tooltips explaining the behavior

### 2. Enhanced SLO List Table
- **New Column**: "Burn Rate" column added between "Name" and "Window"
- **Features**:
  - Sortable column
  - Toggleable visibility via column selector
  - Hover tooltips with detailed explanations
  - Color-coded badges for quick identification

### 3. Enhanced SLO Detail Page
- **Location**: Below SLO description section
- **Content**: Shows current burn rate type with detailed tooltip
- **Responsive Design**: Adapts to different screen sizes

## Implementation Details

### Files Modified/Added:

1. **`ui/src/burnrate.tsx`** - New utility module
   - TypeScript interfaces for burn rate types
   - Mock detection logic (temporary until API integration)
   - Display configuration and helpers

2. **`ui/src/pages/List.tsx`** - Enhanced SLO list
   - Added burn rate column to table
   - Updated row interface and data population
   - Added column visibility controls

3. **`ui/src/pages/Detail.tsx`** - Enhanced detail view
   - Added burn rate information section
   - Responsive layout integration

4. **`ui/src/components/Icons.tsx`** - New icons
   - Dynamic burn rate icon (adaptive eye)
   - Static burn rate icon (lock)

5. **`examples/demo-dynamic-burnrate.yaml`** - Demo configuration
   - Example SLO configurations showcasing both types

### Mock Detection Logic (OBSOLETE)

**Note:** This mock logic was used during early development and has been replaced with actual API integration.

Early prototype used heuristic detection:

```typescript
// OBSOLETE - No longer in codebase
// Keywords that suggest dynamic behavior
const dynamicKeywords = ['dynamic', 'traffic-aware', 'adaptive', 'auto', 'smart']

// Service patterns that often benefit from dynamic burn rates
if (name.includes('latency') || name.includes('response_time')) {
  return BurnRateType.Dynamic
}

// Randomized assignment for demo purposes
if (name.includes('api') || name.includes('service')) {
  return Math.random() > 0.6 ? BurnRateType.Dynamic : BurnRateType.Static
}
```

**Current Implementation:** The UI now receives `burnRateType` directly from the API via protobuf.

## Visual Design

### Color Scheme:
- **Dynamic**: Green badges (`success` variant) - indicating intelligent/adaptive behavior
- **Static**: Gray badges (`secondary` variant) - indicating traditional/fixed behavior

### Icons:
- **Dynamic**: Eye icon - represents "intelligent observation" and adaptability
- **Static**: Lock icon - represents "fixed/locked" thresholds

### Tooltips:
- **Dynamic**: "Traffic-aware burn rate that adapts alert thresholds based on request volume for more accurate alerting"
- **Static**: "Traditional fixed burn rate thresholds - reliable but may not account for traffic variations"

## Integration Points

### Current State:
- UI components are ready and functional with mock data
- Responsive design works across different screen sizes
- Column visibility and sorting work properly

### Integration Status (COMPLETED):

1. ✅ **Protobuf Updates**: Added `Alerting` message with `burn_rate_type` field
2. ✅ **API Integration**: Updated Go conversion functions to include alerting info
3. ✅ **Replaced Mock Logic**: UI now uses real API field
4. ✅ **Testing**: End-to-end testing completed with real dynamic burn rate configurations

**See `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md` for current implementation status.**

## User Experience

### SLO List Page:
1. Users can quickly scan the "Burn Rate" column to see which SLOs use dynamic vs static alerting
2. Hover over badges for detailed explanations
3. Toggle column visibility if desired
4. Sort by burn rate type to group similar SLOs

### SLO Detail Page:
1. Clear indication of the current burn rate type
2. Educational tooltips help users understand the differences
3. Consistent visual language with the list page

## Demo Usage

To see the dynamic burn rate UI in action:

1. Start the UI server: `cd ui && npm start`
2. Visit `http://localhost:3000`
3. View the SLO list to see burn rate type indicators
4. Click on any SLO to see detailed burn rate information
5. Create SLOs with keywords like "dynamic", "latency", or "api" to see dynamic classification

## Future Enhancements

- **Filtering**: Add ability to filter SLOs by burn rate type
- **Statistics**: Show counts/percentages of dynamic vs static SLOs
- **Migration Tools**: UI helpers for converting static SLOs to dynamic
- **Performance Metrics**: Display effectiveness metrics for dynamic burn rates
- **Configuration UI**: Form-based SLO creation with burn rate type selection

## Technical Notes

- Uses React Bootstrap for consistent styling
- TypeScript interfaces ensure type safety
- Responsive design works on mobile devices
- Accessibility features included (proper ARIA labels, tooltips)
- Performance optimized (memoized calculations where appropriate)
