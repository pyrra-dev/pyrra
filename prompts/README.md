# Session Prompts - Dynamic Burn Rate Feature

This folder contains session continuation prompts for implementing the dynamic burn rate feature across multiple focused development sessions.

## Current Session Prompts

### `UI_INTEGRATION_SESSION_PROMPT.md` - **ACTIVE**
**Use this prompt for the next session**  
**Focus**: UI Integration (React frontend updates)  
**Status**: Backend implementation complete, ready for frontend work  
**Updated**: August 26, 2025

### `BACKEND_COMPLETION_SESSION_PROMPT.md` - COMPLETED
**Status**: Used for backend implementation completion  
**Focus**: LatencyNative and BoolGauge indicator support  
**Completion**: All indicator types now implemented ✅

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
- **Frontend**: 🎯 Next development focus
- **Tests**: ✅ All passing
- **Build**: ✅ Clean compilation
