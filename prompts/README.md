# Session Prompts - Dynamic Burn## Session Development Strategy

The dynamic burn rate feature is being developed across multiple focused sessions:

1. ✅ **Sessions 1-3**: Core backend implementation (Ratio, Latency indicators)
2. ✅ **Session 4**: Backend completion (LatencyNative, BoolGauge indicators) → `BACKEND_COMPLETION_SESSION_PROMPT.md`
3. ✅ **Session 5**: React UI foundation (burn rate display system) → `UI_INTEGRATION_SESSION_PROMPT.md`
4. ✅ **Session 6**: Protobuf & API integration → `API_INTEGRATION_SESSION_PROMPT.md` 
5. 🎯 **Session 7**: Alert display component updates → `ALERT_DISPLAY_SESSION_PROMPT.md`
6. 🔜 **Session 8**: Advanced UI features (enhanced tooltips, educational content)
7. 🔜 **Session 9**: Grafana dashboard integration

## Prompt Usage Guide

- **🎯 For Next Session**: Use `ALERT_DISPLAY_SESSION_PROMPT.md`
- **📚 For Reference**: Check completed session prompts to understand implementation history
- **🔍 For Code Review**: Use `CODE_REVIEW_SESSION_PROMPT.md` template
This folder contains session continuation prompts for implementing the dynamic burn rate feature across multiple focused development sessions.

## Current Session Prompts

### `ALERT_DISPLAY_SESSION_PROMPT.md` - **ACTIVE** 🎯
**Use this prompt for the next session**  
**Focus**: Alert Display Updates (Dynamic burn rate UI components)  
**Status**: API integration complete, alert display enhancement ready  
**Created**: August 28, 2025

### `API_INTEGRATION_SESSION_PROMPT.md` - COMPLETED ✅
**Status**: Completed - Protobuf & API integration finished  
**Focus**: Real API field access, eliminated mock detection  
**Completion**: All 5 Priority 1 tasks completed successfully (Aug 28, 2025)

### `UI_INTEGRATION_SESSION_PROMPT.md` - COMPLETED ✅
**Status**: Completed - UI foundation and burn rate display system  
**Focus**: React frontend burn rate badges, icons, and user experience  
**Completion**: Complete UI foundation with mock detection logic

### `BACKEND_COMPLETION_SESSION_PROMPT.md` - COMPLETED ✅
**Status**: Used for backend implementation completion  
**Focus**: LatencyNative and BoolGauge indicator support  
**Completion**: All indicator types now implemented

### `CODE_REVIEW_SESSION_PROMPT.md` - REFERENCE
**Status**: Code review template  
**Focus**: Validation and quality assurance  
**Usage**: Reference for thorough implementation review

## Session Development Strategy

The dynamic burn rate feature is being developed across multiple focused sessions:

1. ✅ **Sessions 1-3**: Core backend implementation (Ratio, Latency indicators)
2. ✅ **Session 4**: Backend completion (LatencyNative, BoolGauge indicators) → `BACKEND_COMPLETION_SESSION_PROMPT.md`
3. 🎯 **Session 5**: React UI integration (BurnRateType selection) → `UI_INTEGRATION_SESSION_PROMPT.md`
4. 🔜 **Session 6**: Advanced UI features (status indicators, tooltips)
5. 🔜 **Session 7**: Grafana dashboard integration

## Prompt Usage Guide

- **🎯 For Next Session**: Use `UI_INTEGRATION_SESSION_PROMPT.md`
- **📚 For Reference**: Check `BACKEND_COMPLETION_SESSION_PROMPT.md` to understand what was implemented
- **🔍 For Code Review**: Use `CODE_REVIEW_SESSION_PROMPT.md` template

## Context Files

Key context files to reference in sessions:
- `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md` - Complete feature overview
- `.dev-docs/dynamic-burn-rate.md` - Technical specification
- `.dev-docs/burn-rate-analysis.md` - Mathematical analysis

## Repository Status

- **Branch**: add-dynamic-burn-rate  
- **Backend**: ✅ Complete and production-ready
- **API Integration**: ✅ Complete - Real protobuf field transmission
- **UI Foundation**: ✅ Complete - Burn rate display system with badges and icons
- **Next Focus**: 🎯 Alert display component updates  
- **Tests**: ✅ All passing
- **Build**: ✅ Clean compilation
