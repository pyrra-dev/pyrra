# Dynamic Burn Rate - Next Session Continuation Prompt

I'm continuing work on Pyrra's dynamic burn rate feature for traffic-aware SLO alerting. 

## Current Status
Context: #file:.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md

## ✅ **COMPLETED: Backend Implementation - All Indicator Types**

### Latest Session Achievements (Aug 26, 2025)
- ✅ **All Indicator Types Complete**: Ratio, Latency, LatencyNative, and BoolGauge indicators
- ✅ **Dynamic Expression Generation**: Complete `buildDynamicAlertExpr()` implementation
- ✅ **Traffic-Aware Calculations**: Each indicator type uses appropriate Prometheus functions
- ✅ **Dynamic Window Integration**: All indicator types use dynamic windows when configured  
- ✅ **Unified Alert Building**: All types use centralized `buildAlertExpr()` method
- ✅ **Comprehensive Testing**: Full test coverage with all tests passing
- ✅ **Production Ready**: Clean build, no compilation errors, backward compatibility maintained

### Backend Implementation Summary
```go
// Dynamic expressions implemented for all types:
case Ratio:        // sum(increase(errors)) / sum(increase(total))
case Latency:      // sum(increase(slow_requests)) / sum(increase(total_requests))  
case LatencyNative:// histogram_count(sum(increase(histogram)))
case BoolGauge:    // count_over_time(boolean_gauge)
```

## 🎯 **NEXT PRIORITIES: UI & Visualization Integration**

The backend dynamic burn rate implementation is **complete and production-ready**. The next major component is integrating the feature into the user interface and visualization layers.

### Priority 1: React UI Integration
**Goal**: Update the React frontend to support dynamic burn rate configuration and display

**Key Files to Modify**:
- `ui/src/objectives.tsx` - Main SLO management interface
- `ui/src/App.tsx` - Application routing and state management
- UI components for SLO creation/editing forms
- Alert configuration components

**Tasks**:
1. Add BurnRateType selection in SLO creation forms (radio buttons: Static/Dynamic)
2. Update SLO display to show current burn rate mode
3. Add informational tooltips explaining dynamic vs static behavior
4. Validate form inputs when switching between modes
5. Update TypeScript types to include BurnRateType field

### Priority 2: Grafana Dashboard Integration
**Goal**: Update Grafana dashboards to visualize dynamic burn rate behavior

**Key Files to Check**:
- `examples/grafana/` - Dashboard JSON definitions
- Documentation for dashboard updates
- Recording rule visualization

**Tasks**:
1. Add dynamic threshold visualization panels
2. Create traffic volume correlation graphs
3. Add comparison panels (static vs dynamic behavior)
4. Update alert panel queries to handle dynamic expressions
5. Add debug panels for understanding dynamic calculations

### Priority 3: Testing & Documentation
**Goal**: Comprehensive validation and user guidance

**Tasks**:
1. End-to-end testing with UI integration
2. Update user documentation with dynamic burn rate usage
3. Create migration guide from static to dynamic
4. Performance testing with real Grafana dashboards
5. Example SLO configurations demonstrating dynamic mode

## Current Codebase Status

### ✅ **Backend (Complete)**
```bash
# All tests passing
go test ./slo -v -run "TestObjective_DynamicBurnRate"  # ✅ All 4 types
go build .  # ✅ No compilation errors
```

### 🔄 **Frontend (Next Focus)**
```bash
cd ui/
npm install
npm start  # Should work but won't show dynamic features yet
npm test   # May need updates for new functionality
```

## Technical Context

### Backend Architecture (Completed)
- **API Support**: `BurnRateType` field in SLO spec (`"static"` | `"dynamic"`)
- **Alert Generation**: Traffic-aware thresholds using recording rules + dynamic calculations  
- **Backward Compatibility**: Existing SLOs continue working unchanged
- **Performance**: Optimized expressions using pre-computed recording rules

### Frontend Architecture (To Implement)
- **React Components**: Need updates for BurnRateType selection
- **TypeScript Types**: May need BurnRateType interface additions
- **Form Validation**: Dynamic mode configuration validation
- **State Management**: SLO editing with new burn rate options

### Integration Points
- **API Communication**: Frontend ↔ Backend SLO CRUD operations
- **Grafana Integration**: Dashboard updates for dynamic visualization
- **Kubernetes CRDs**: Already support BurnRateType field (✅ Complete)

## Implementation Strategy

### Session 1: React UI Core Integration
- Focus on basic BurnRateType selection in SLO forms
- Update TypeScript interfaces and component props
- Add basic UI elements without complex visualizations

### Session 2: Advanced UI Features
- Dynamic burn rate status indicators and tooltips
- Form validation and user guidance
- SLO list/detail view updates

### Session 3: Grafana & Visualization
- Update Grafana dashboard JSON configurations
- Create dynamic threshold visualization panels
- Test with real dashboard deployments

## Repository Context
- **Branch**: add-dynamic-burn-rate
- **Owner**: yairst/pyrra  
- **Status**: Backend complete, UI integration next
- **Build Status**: ✅ All tests passing, clean compilation

## Session Focus Recommendation

I recommend starting with **Priority 1: React UI Integration** in the next session, specifically focusing on the core SLO form updates to add BurnRateType selection. This provides immediate user value and establishes the foundation for more advanced UI features.

The backend dynamic burn rate implementation is solid and production-ready, making it an excellent foundation for building the user-facing features.
