# Session Prompts - Dynamic Burn Rate Feature Implementation

This folder contain**When starting### **Phase### **Phase 1: Indicator Type Coverage** (Sessions 10A-C, 11)
- ✅ Ratio indicators working
- ✅ Latency indicators backend (Session 10A)
- ✅ Latency indicators UI basic (Session 10B) 
- 🚧 Latency indicators comprehensive (Session 10C)
- 🔜 Latency native & bool gauge indicators (Session 11)Indicator Type Coverage** (Sessions 10A-C, 11)
- ✅ Ratio indicators working
- ✅ Latency indicators backend (Session 10A completed)
- ✅ Latency indicators UI basic (Session 10B completed)
- 🚧 Latency indicators comprehensive (Session 10C)
- 🔜 Latency native & bool gauge indicators (Session 11)ocused session, the agent should:**
1. **Reference the comprehensive plan**: Review `10_COMPREHENSIVE_...` for full context
2. **Use the focused prompt**: Execute specific session (e.g., `10B_...` or `12_...`)
3. **Update documentation**: Update `.dev-docs/DYNAMIC_BURN_RATE_TESTING_SESSION.md` with results
4. **Plan next session**: Based on results, recommend next focused sessionsion continuation prompts for implementing the dynamic burn rate feature across multiple focused development sessions.

## 📊 Current Implementation Status (September 6, 2025)

### **✅ COMPLETED PHASES** 
- **Backend Implementation**: Complete (Sessions 1-4)
- **API Integration**: Complete (Session 5-6) 
- **UI Foundation**: Complete (Session 7-8)
- **Basic Threshold Display**: Complete (Session 9) - Ratio indicators only

### **🚧 CURRENT PHASE: Comprehensive Validation**
**Status**: Basic UI working for ratio indicators (~20% complete)
**Next Priority**: Comprehensive testing across all indicator types and scenarios

## 🎯 Active Session Prompts

### **RECOMMENDED NEXT SESSIONS**

#### `10C_LATENCY_UI_COMPLETION_AND_VALIDATION_SESSION_PROMPT.md` - **HIGH PRIORITY** 🎯
**Use this prompt to complete latency indicator comprehensive validation**  
**Focus**: Enhanced tooltips, performance assessment, error handling, production readiness  
**Scope**: Complete Session 10B gaps + comprehensive validation (2-3 hours)  
**Dependencies**: Session 10B completed (basic threshold display working)  
**Created**: September 16, 2025

#### `10B_LATENCY_UI_THRESHOLD_DISPLAY_SESSION_PROMPT.md` - COMPLETED ✅
**Status**: Session 10B completed - Basic latency threshold display working  
**Focus**: BurnRateThresholdDisplay component enhancement for histogram metric support  
**Result**: Latency SLOs show calculated values instead of "Traffic-Aware"  
**Completion**: September 14, 2025

#### `12_MISSING_METRICS_VALIDATION_SESSION_PROMPT.md` - **ALTERNATIVE PRIORITY** 🛡️
**Use this prompt for resilience testing**  
**Focus**: Missing metrics, edge cases, error handling  
**Scope**: Robustness validation (1-2 hours)  
**Dependencies**: None - can start immediately  
**Created**: September 6, 2025

#### `10_COMPREHENSIVE_DYNAMIC_BURN_RATE_VALIDATION_SESSION_PROMPT.md` - **PLANNING OVERVIEW** 📋
**Use this for session planning guidance**  
**Focus**: Multi-session roadmap and strategy  
**Scope**: Planning document, not execution prompt  
**Usage**: Reference for understanding full validation scope  

### **COMPLETED SESSION PROMPTS** ✅

#### `10A_LATENCY_INDICATOR_VALIDATION_SESSION_PROMPT.md` - COMPLETED ✅
**Status**: Session 10A completed - Backend validation and mathematical verification complete  
**Focus**: Latency indicator backend rules, recording rule architecture, error budget fix  
**Result**: -1900% error budget issue fixed, mathematical validation confirmed  
**Completion**: September 13, 2025

#### `9_DYNAMIC_BURN_RATE_UI_VERIFICATION_SESSION_PROMPT.md` - COMPLETED ✅
**Status**: Session 9 completed - Basic threshold display working  
**Focus**: UI verification and real-time threshold calculations  
**Result**: BurnRateThresholdDisplay component working for ratio indicators  
**Completion**: September 6, 2025

#### `8_DYNAMIC_BURN_RATE_VALIDATION_SESSION_PROMPT.md` - COMPLETED ✅
**Status**: Sessions 1-8 completed  
**Focus**: Data validation, mathematical correctness, metric switching  
**Result**: Foundation established, real data integration working  
**Completion**: September 5, 2025

#### Legacy Prompts (Pre-September 2025)
- `ALERT_DISPLAY_SESSION_PROMPT.md` - COMPLETED ✅
- `API_INTEGRATION_SESSION_PROMPT.md` - COMPLETED ✅  
- `UI_INTEGRATION_SESSION_PROMPT.md` - COMPLETED ✅
- `BACKEND_COMPLETION_SESSION_PROMPT.md` - COMPLETED ✅

## 🔧 Session Selection Guide

### **For Immediate Next Session, Choose One**:

**Option A: Latency Comprehensive Validation** (`10C_LATENCY_UI_COMPLETION_AND_VALIDATION_SESSION_PROMPT.md`)
- **Why**: Complete latency indicator validation to production readiness
- **Scope**: Enhanced tooltips, performance assessment, error handling, comprehensive validation
- **Duration**: 2-3 hours focused development + testing
- **Risk**: Medium complexity, building on working foundation

**Option B: Resilience Testing** (`12_MISSING_METRICS_VALIDATION_SESSION_PROMPT.md`)  
- **Why**: Critical for production reliability, independent of indicator types
- **Scope**: Missing metrics, edge cases, error handling
- **Duration**: 1-2 hours focused testing  
- **Risk**: Lower complexity, easier debugging

### **Session Context for Agent**:
When starting a focused session, the agent should:
1. **Reference the comprehensive plan**: Review `10_COMPREHENSIVE_...` for full context
2. **Use the focused prompt**: Execute specific session (e.g., `10C_...` or `12_...`)
3. **Update documentation**: Update `.dev-docs/DYNAMIC_BURN_RATE_TESTING_SESSION.md` with results
4. **Plan next session**: Based on results, recommend next focused session

## 📋 Remaining Validation Roadmap

**Current Completion**: ~30% (basic UI working for ratio and latency indicators)  
**Estimated Remaining**: 3-5 focused sessions

### **Phase 1: Indicator Type Coverage** (Sessions 10A, 11)
- ✅ Ratio indicators working
- � Latency indicators (Session 10A)
- 🔜 Latency native & bool gauge indicators (Session 11)

### **Phase 2: Resilience Testing** (Sessions 12, 13)
- 🔜 Missing metrics handling (Session 12)
- � Alert firing validation (Session 13)

### **Phase 3: Production Polish** (Sessions 14, 15)
- 🔜 Enhanced tooltips & performance (Session 14)
- � Final production readiness (Session 15)

## 🎯 Context Files for Sessions

**Essential References**:
- `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md` - Complete feature overview
- `.dev-docs/DYNAMIC_BURN_RATE_TESTING_SESSION.md` - Test results and status
- `ui/src/components/BurnRateThresholdDisplay.tsx` - Current UI component

**Technical Context**:
- `.dev-docs/dynamic-burn-rate.md` - Technical specification  
- `.dev-docs/burn-rate-analysis.md` - Mathematical analysis

## 🏆 Repository Status

- **Branch**: add-dynamic-burn-rate  
- **Backend**: ✅ Complete and production-ready  
- **API Integration**: ✅ Complete with real protobuf transmission
- **UI Foundation**: ✅ Complete with badge system and basic threshold display
- **Current Focus**: 🚧 Comprehensive validation across indicator types
- **Tests**: ✅ All passing (basic functionality)
- **Build**: ✅ Clean compilation
- **Production Ready**: ⚠️ **NO** - Comprehensive testing required

**Status**: 🚧 **FOUNDATION COMPLETE - COMPREHENSIVE VALIDATION PHASE**
