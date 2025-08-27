# Dynamic Burn Rate - API Integration & Alert Display Session

I'm continuing work on Pyrra's dynamic burn rate feature for traffic-aware SLO alerting.

## Current Status
Context: #file:.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md
Additional Context: #file:.dev-docs/CORE_CONCEPTS_AND_TERMINOLOGY.md

## ✅ **COMPLETED: Backend + UI Foundation - All Ready**

### Backend Implementation Status (Complete ✅)
- ✅ **All Indicator Types Complete**: Ratio, Latency, LatencyNative, and BoolGauge indicators
- ✅ **Dynamic Expression Generation**: Complete `buildDynamicAlertExpr()` implementation
- ✅ **Traffic-Aware Calculations**: Each indicator type uses appropriate Prometheus functions
- ✅ **Production Ready**: Clean build, comprehensive testing, backward compatibility maintained

### UI Integration Status (Foundation Complete ✅)
- ✅ **Burn Rate Display System**: Color-coded badges in List and Detail pages
- ✅ **TypeScript Infrastructure**: Complete type system with `BurnRateType` enum
- ✅ **Visual Design**: Icons, tooltips, responsive layout, accessibility features
- ✅ **User Experience**: Sortable columns, toggleable visibility, deterministic badge behavior
- ✅ **Mock Detection Logic**: Working heuristics until API integration

## 🎯 **NEXT PRIORITIES: Complete API Integration & Alert Display**

The backend implementation is complete and the UI foundation is established. The final components needed are:

### Priority 1: Protobuf & API Integration  
**Goal**: Eliminate mock detection and connect UI to actual backend `burnRateType` field

**Current Issue**: The `burnRateType` field exists in the Go structs but isn't transmitted via the protobuf API to the frontend.

**Key Files to Modify**:
- `proto/objectives/v1alpha1/objectives.proto` - Add `Alerting` message with `burn_rate_type` field
- `proto/objectives/v1alpha1/objectives.go` - Update `ToInternal()` and `FromInternal()` conversion functions
- UI protobuf regeneration via `buf generate`

**Tasks**:
1. Add `Alerting` message to protobuf definition (partially done - needs completion)
2. Update Go conversion functions to include alerting information in API responses
3. Regenerate protobuf TypeScript files for frontend
4. Replace mock detection logic with actual API field
5. Test end-to-end API integration

### Priority 2: Alert Display Updates
**Goal**: Update existing UI components to show dynamic burn rate information instead of static calculations

**Current Issue**: Detail page shows static burn rate tooltips (factors like 14, 7, 2, 1) even for dynamic SLOs that actually use adaptive thresholds.

**Key Files to Check**:
- `ui/src/components/AlertsTable.tsx` - Alert display components
- `ui/src/components/graphs/` - Error budget and burn rate graphs  
- `ui/src/pages/Detail.tsx` - Alert information displays

**Tasks**:
1. Identify UI components displaying static burn rate calculations
2. Add conditional logic to show appropriate information based on burn rate type
3. Create dynamic-specific tooltips and displays
4. Add visual indicators when dynamic mode is active
5. Test with real dynamic SLO configurations

### Priority 3: Testing & Polish
**Goal**: Comprehensive validation and user experience refinement

**Tasks**:
1. End-to-end testing with actual dynamic SLO configurations
2. Verify dynamic threshold calculations display correctly
3. Test API integration with various SLO types
4. User experience polish and edge case handling
5. Documentation updates for complete feature

## Technical Context

### Protobuf Integration Challenge
The backend has the `BurnRateType` field in `slo.Alerting` struct, but the protobuf API doesn't expose it:

```go
// In proto/objectives/v1alpha1/objectives.go ToInternal() - Line 95
Alerting: slo.Alerting{}, // TODO - Currently empty!
```

The `FromInternal()` function also doesn't include alerting information in responses.

### Required Protobuf Changes
```proto
message Alerting {
  bool burnrates = 1;
  bool absent = 2; 
  string name = 3;
  string absent_name = 4;
  string burn_rate_type = 5;  // Key field needed
}

message Objective {
  // ... existing fields
  Alerting alerting = 8;  // Add alerting field
}
```

### UI Mock Detection (Temporary)
Current UI uses heuristics in `ui/src/burnrate.tsx`:
```typescript
// Temporary until API provides real burnRateType field
const hasDynamicKeywords = dynamicKeywords.some(keyword => 
  searchText.includes(keyword)
)
```

This should be replaced with: `objective.alerting?.burnRateType`

## Current Environment Status

### Backend (Production Ready)
```bash
go test ./slo -v   # ✅ All tests passing
go build .         # ✅ Clean compilation
```

### Frontend (Foundation Ready)  
```bash
cd ui/
npm start          # ✅ UI server running with burn rate display
# Visit http://localhost:3000 to see current implementation
```

### Kubernetes Integration
```bash
kubectl get slo -n monitoring  # Shows test SLOs with burn rate badges in UI
```

## Implementation Strategy

### Session 1: Protobuf & API Integration
- Complete protobuf `Alerting` message definition
- Update Go conversion functions (`ToInternal`, `FromInternal`)
- Regenerate TypeScript protobuf files
- Update UI to use real API field

### Session 2: Alert Display Components  
- Audit existing UI components showing static burn rate info
- Implement dynamic-aware display logic
- Add appropriate visual indicators and tooltips

### Session 3: Testing & Validation
- End-to-end testing with real dynamic configurations
- User experience validation and polish
- Documentation updates

## Files Modified in Previous Session

### New Files Created:
- `ui/src/burnrate.tsx` - Burn rate type system and utilities
- `ui/DYNAMIC_BURN_RATE_UI.md` - UI implementation documentation
- `examples/simple-demo.yaml` - Working demo SLO configurations

### Files Modified:
- `ui/src/pages/List.tsx` - Added burn rate column with badges
- `ui/src/pages/Detail.tsx` - Added burn rate information display  
- `ui/src/components/Icons.tsx` - Added dynamic/static icons
- `proto/objectives/v1alpha1/objectives.proto` - Partial protobuf updates (needs completion)

### Files to Clean Up:
- Check for any temporary files that should be removed

## Repository Context
- **Branch**: add-dynamic-burn-rate
- **Owner**: yairst/pyrra  
- **Status**: Backend complete, UI foundation complete, API integration next
- **Build Status**: ✅ All tests passing, UI working with mock data

## Session Focus Recommendation

I recommend starting with **Priority 1: Protobuf & API Integration** in the next session. This will:

1. **Eliminate Technical Debt**: Replace mock detection logic with proper API integration
2. **Enable Real Testing**: Allow testing with actual dynamic SLO configurations  
3. **Foundation for Polish**: Set up the infrastructure needed for Priority 2 improvements

The backend dynamic burn rate implementation is solid and production-ready. The UI foundation provides excellent user experience. The missing piece is the API connection between them.

Once API integration is complete, we can focus on updating the existing alert display components to show dynamic burn rate calculations instead of the legacy static displays.
