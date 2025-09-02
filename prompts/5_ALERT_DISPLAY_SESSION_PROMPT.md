````markdown
# Dynamic Burn Rate - Alert Display Updates Session

I'm continuing work on Pyrra's dynamic burn rate feature for traffic-aware SLO alerting.

## Current Status
Context: #file:.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md
Additional Context: #file:.dev-docs/CORE_CONCEPTS_AND_TERMINOLOGY.md

## ✅ **COMPLETED: Backend + UI Foundation + API Integration - All Ready**

### Backend Implementation Status (Complete ✅)
- ✅ **All Indicator Types Complete**: Ratio, Latency, LatencyNative, and BoolGauge indicators
- ✅ **Dynamic Expression Generation**: Complete `buildDynamicAlertExpr()` implementation
- ✅ **Traffic-Aware Calculations**: Each indicator type uses appropriate Prometheus functions
- ✅ **Production Ready**: Clean build, comprehensive testing, backward compatibility maintained

### UI Foundation Status (Complete ✅)
- ✅ **Burn Rate Display System**: Color-coded badges in List and Detail pages
- ✅ **TypeScript Infrastructure**: Complete type system with `BurnRateType` enum
- ✅ **Visual Design**: Icons, tooltips, responsive layout, accessibility features
- ✅ **User Experience**: Sortable columns, toggleable visibility, deterministic badge behavior

### API Integration Status (Complete ✅ - Aug 28, 2025)
- ✅ **Protobuf Schema**: Added `Alerting` message with `burn_rate_type` field
- ✅ **Go Conversion Functions**: Complete bidirectional conversion in `ToInternal()` and `FromInternal()`
- ✅ **TypeScript Protobuf Files**: Updated definitions and implementations for frontend
- ✅ **Real API Field Access**: Replaced mock detection with `objective.alerting?.burnRateType`
- ✅ **End-to-End Validation**: Round-trip testing confirms API integration works correctly

**Last Session Achievements**: Completed all 5 Priority 1 tasks, eliminated mock detection logic, validated API integration with comprehensive testing.

## 🎯 **NEXT PRIORITY: Alert Display Updates**

**Goal**: Update existing UI components to show dynamic burn rate information instead of static calculations

**Current Issue**: Several UI components still display static burn rate information (factors like 14, 7, 2, 1) even for dynamic SLOs that actually use adaptive thresholds based on traffic patterns.

### Priority 2: Alert Display Component Updates
**Status**: Ready to begin - API integration provides the foundation

**Key Challenge Areas**:
1. **Alert Tables**: Components showing burn rate thresholds and calculations
2. **Graph Tooltips**: Charts displaying static burn rate factors in tooltips
3. **Detail Page Information**: Alert information panels showing static calculations
4. **Error Budget Displays**: Components showing threshold calculations

**Specific Files to Investigate**:
- `ui/src/components/AlertsTable.tsx` - Alert display components
- `ui/src/components/graphs/` - Error budget and burn rate graph components  
- `ui/src/pages/Detail.tsx` - Alert information displays and panels
- `ui/src/components/ErrorBudget.tsx` - Error budget threshold displays

### Implementation Strategy

#### Phase 1: Component Discovery & Analysis
**Tasks**:
1. **Audit existing components**: Search for hardcoded static burn rate displays
2. **Identify calculation displays**: Find components showing factors (14, 7, 2, 1)
3. **Map tooltip content**: Locate static burn rate information in chart tooltips
4. **Document current behavior**: Understand what information is currently shown

**Search Strategy**:
- Look for components displaying alert thresholds or burn rate calculations
- Find references to window factors (14, 7, 2, 1) in UI code
- Identify tooltip content mentioning "burn rate" or "threshold"

#### Phase 2: Conditional Display Logic
**Tasks**:
1. **Create display helpers**: Functions to determine what information to show based on burn rate type
2. **Dynamic-aware tooltips**: Show traffic-aware threshold information for dynamic SLOs
3. **Static information preservation**: Maintain existing displays for static SLOs
4. **Visual differentiation**: Add indicators when dynamic calculations are used

**Technical Approach**:
```typescript
// Example helper function pattern
function getBurnRateDisplayInfo(objective: Objective) {
  if (objective.alerting?.burnRateType === 'dynamic') {
    return {
      type: 'Traffic-Aware Thresholds',
      description: 'Thresholds adapt based on actual request volume',
      tooltip: 'Dynamic burn rate adjusts alert thresholds based on traffic patterns'
    };
  }
  return {
    type: 'Static Thresholds', 
    description: `Fixed factors: ${window.factor}x target`,
    tooltip: `Static burn rate uses fixed multiplier: ${window.factor}`
  };
}
```

#### Phase 3: Enhanced User Experience
**Tasks**:
1. **Context-aware information**: Show appropriate details based on SLO configuration
2. **Educational tooltips**: Help users understand dynamic vs static behavior
3. **Performance indicators**: Visual cues about alert sensitivity
4. **Consistency checks**: Ensure uniform experience across all components

### Technical Context

#### Current API Integration (Completed)
The UI now has access to real burn rate type information:
```typescript
// In any component with objective data
const burnRateType = objective.alerting?.burnRateType; // 'dynamic' | 'static' | undefined
```

#### Dynamic Burn Rate Formula (For UI Display)
When showing dynamic information, reference the actual formula:
```
Dynamic Threshold = (N_SLO / N_long) × E_budget_percent_threshold × (1 - SLO_target)
```

**Key Concepts for UI**:
- **N_SLO**: Request volume in SLO window (e.g., 7 days, 28 days)  
- **N_long**: Request volume in long alert window (for traffic scaling)
- **E_budget_percent_threshold**: Constants (1/48, 1/16, 1/14, 1/7) based on window
- **Traffic Adaptation**: Higher traffic → higher thresholds, lower traffic → lower thresholds

#### Static Burn Rate (Existing Behavior)
Static SLOs continue using fixed multipliers:
```
Static Threshold = window_factor × (1 - SLO_target)
```

### Session Tasks

#### Session Focus: Update Alert Display Components
**Primary Objectives**:
1. **Find and update static burn rate displays**: Replace hardcoded static information with dynamic-aware logic
2. **Enhance tooltips and information panels**: Provide context about alert behavior based on burn rate type  
3. **Add visual indicators**: Help users understand when dynamic thresholds are active
4. **Maintain backward compatibility**: Ensure static SLOs continue showing appropriate information

#### Recommended Workflow
1. **Discovery Phase**: Search codebase for static burn rate references
2. **Component Analysis**: Understand current information display patterns
3. **Implementation Phase**: Add conditional logic for dynamic burn rate displays
4. **Testing Phase**: Verify both static and dynamic SLOs show appropriate information
5. **Polish Phase**: Enhance user experience and add educational elements

#### Key Search Terms for Component Discovery
- "burn rate", "burnrate", "factor", "threshold"
- Numbers: "14", "7", "2", "1" (static factors)
- "alert", "window", "multiplier"

#### Success Criteria
- [ ] Static SLOs continue showing familiar factor-based information
- [ ] Dynamic SLOs show traffic-aware threshold information
- [ ] Tooltips provide educational context about alert behavior
- [ ] Visual consistency across all components
- [ ] No breaking changes to existing functionality

## Current Environment Status

### Repository State (Clean ✅)
```bash
git status  # Clean working tree after API integration commit
```

### Backend (Production Ready ✅)
```bash
go test ./slo -v   # All tests passing including dynamic burn rate
go build .         # Clean compilation
```

### Frontend (API Integration Complete ✅)  
```bash
cd ui/
npm start          # UI server with real API integration
# Visit http://localhost:3000 to see burn rate displays with real data
```

### Test Environment
```bash
kubectl get slo -n monitoring  # Shows test SLOs with real burn rate types
```

## Files Modified in Previous Session (API Integration)

### Modified Files (Aug 28, 2025):
- `proto/objectives/v1alpha1/objectives.go` - API conversion functions with alerting support
- `proto/objectives/v1alpha1/objectives.pb.go` - Generated protobuf Go code  
- `ui/src/burnrate.tsx` - Real API field usage replacing mock detection
- `ui/src/proto/objectives/v1alpha1/objectives_pb.d.ts` - TypeScript type definitions
- `ui/src/proto/objectives/v1alpha1/objectives_pb.js` - JavaScript protobuf implementation
- `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md` - Updated documentation

## Implementation Examples

### Example Component Update Pattern
```typescript
// Before: Static display only
<Tooltip content={`Burn rate factor: ${window.factor}`}>
  <span>{window.factor}x</span>
</Tooltip>

// After: Dynamic-aware display
<Tooltip content={getBurnRateTooltip(objective, window)}>
  <span>{getBurnRateDisplayText(objective, window)}</span>
  {objective.alerting?.burnRateType === 'dynamic' && <DynamicIcon />}
</Tooltip>
```

### Example Helper Functions Needed
```typescript
function getBurnRateTooltip(objective: Objective, window: Window): string {
  if (objective.alerting?.burnRateType === 'dynamic') {
    return 'Dynamic threshold adapts to traffic volume. Higher traffic = higher thresholds.';
  }
  return `Static threshold: ${window.factor}x target error rate`;
}

function getBurnRateDisplayText(objective: Objective, window: Window): string {
  if (objective.alerting?.burnRateType === 'dynamic') {
    return 'Traffic-Aware';
  }
  return `${window.factor}x`;
}
```

## Repository Context
- **Branch**: add-dynamic-burn-rate
- **Owner**: yairst/pyrra  
- **Status**: API integration complete, alert display updates ready to begin
- **Build Status**: ✅ All tests passing, clean working tree
- **Last Commit**: Priority 1 API integration completed successfully

## Session Recommendation

**Recommended Starting Point**: Begin with component discovery and analysis to understand the current state of alert display components. The API integration foundation is solid, so the focus can be entirely on improving user experience for dynamic burn rate visualization.

**Expected Outcomes**: 
- Enhanced alert display components showing appropriate information based on burn rate type
- Improved user understanding of alert behavior through better tooltips and visual indicators
- Maintained backward compatibility for existing static SLO displays
- Foundation ready for future enhancements like Grafana dashboard integration

The dynamic burn rate feature is functionally complete - this session focuses on user experience and information clarity improvements.

````
