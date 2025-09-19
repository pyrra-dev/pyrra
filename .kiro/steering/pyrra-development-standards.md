---
inclusion: always
---

# Pyrra Development Standards and Context

## Project Overview

Pyrra is a Service Level Objective (SLO) management tool for Kubernetes environments using Prometheus metrics. It creates recording and alerting rules for SLOs, supports different indicator types (ratio, latency, boolean), and integrates with Prometheus Operator for rule deployment.

## Current Feature: Dynamic Burn Rate Implementation

### Core Innovation

Dynamic burn rate alerting adapts alert thresholds based on actual traffic patterns rather than using fixed static multipliers. This prevents false positives during low traffic and false negatives during high traffic periods.

**Mathematical Foundation:**

```
dynamic_threshold = (N_SLO / N_alert) × E_budget_percent_threshold × (1 - SLO_target)
```

Where:

- N_SLO = Number of events in SLO window (e.g., 28 days)
- N_alert = Number of events in alert window (e.g., 1 hour)
- E_budget_percent_threshold = Constant percentage (1/48, 1/16, 1/14, 1/7)
- (1 - SLO_target) = Error budget (e.g., 0.01 for 99% SLO)

### Implementation Status

- ✅ **Backend**: Complete for all indicator types (Ratio, Latency, LatencyNative, BoolGauge)
- ✅ **API Integration**: Complete with protobuf transmission
- ✅ **UI Foundation**: Complete with badge system and basic threshold display
- 🚧 **Comprehensive Validation**: ~30% complete (ratio and basic latency working)

## Development Standards

### Code Quality Requirements

1. **Systematic Comparison**: Always compare with working examples, never analyze in isolation
2. **Comprehensive Structure Validation**: Check naming conflicts, complete coverage, syntax validation, label consistency
3. **Question Everything**: Understand purpose of every component, explain deviations, identify gaps
4. **Test Before Success**: Syntax validation, functional testing, edge cases, performance impact
5. **Document Issues**: Categorize severity, provide context, suggest concrete fixes

### Architecture Patterns

#### UI Development Critical Workflow

Pyrra uses **two different UI serving methods**:

1. **Development UI (Port 3000)**: `npm start` - live source files with hot reload
2. **Embedded UI (Port 9099)**: `./pyrra api` - compiled files via Go embed

**CRITICAL**: UI changes require complete rebuild workflow:

```bash
# 1. Make changes in ui/src/
# 2. Test in development: npm start → http://localhost:3000
# 3. Build for production: npm run build (creates ui/build/)
# 4. Rebuild Go binary: make build (embeds ui/build/)
# 5. Restart service and test embedded UI at http://localhost:9099
```

#### Backend Implementation Patterns

- **Recording Rules**: Use optimized recording rules instead of inline calculations
- **Multi-Window Logic**: Both short and long windows use N_long for traffic scaling consistency
- **Backward Compatibility**: Static burn rate remains default, preserve existing functionality
- **Error Handling**: Graceful fallbacks and conservative defaults

### Testing Standards

#### Comprehensive Validation Requirements

1. **Indicator Type Coverage**: Test all SLO types (ratio, latency, latency_native, bool_gauge)
2. **Resilience Testing**: Missing metrics, insufficient data, mathematical edge cases
3. **Alert Firing Validation**: Prove alerts actually work with synthetic metrics
4. **UI Polish**: Enhanced tooltips, performance testing, error handling
5. **Production Readiness**: End-to-end testing, deployment validation

#### Mathematical Validation

- Cross-validate calculations against known working examples
- Test with real Prometheus data, not just synthetic scenarios
- Verify traffic scaling behavior in high/low traffic conditions
- Confirm alert sensitivity matches expected behavior

### File References for Implementation

#### Core Implementation Files

- `slo/rules.go` - Main dynamic burn rate implementation
- `slo/slo.go` - Core types and interfaces
- `kubernetes/api/v1alpha1/servicelevelobjective_types.go` - CRD definitions
- `ui/src/components/BurnRateThresholdDisplay.tsx` - UI threshold display
- `ui/src/burnrate.tsx` - Helper functions for burn rate logic

#### Documentation Files

- `.dev-docs/CORE_CONCEPTS_AND_TERMINOLOGY.md` - Mathematical definitions
- `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md` - Complete implementation status
- `.dev-docs/AI_DEVELOPMENT_QUALITY_STANDARDS.md` - Quality requirements
- `prompts/README.md` - Current session status and next priorities

#### Test Configuration

- `.dev/test-slo.yaml` - Test SLO for development
- `.dev/test-dynamic-slo.yaml` - Dynamic burn rate test configuration

## Windows Development Considerations

### Known Issues

- **CRD Generation**: `make generate` has Windows/Git Bash compatibility issues
- **Workaround**: Use `controller-gen crd paths="./kubernetes/api/v1alpha1"` directly
- **Path Globbing**: Windows doesn't handle `./...` Go-style patterns correctly
- **Build Dependencies**: May need manual installation of Go tools

### Required Tools

```bash
go install github.com/brancz/gojsontoyaml@latest
go install mvdan.cc/gofumpt@latest
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
```

## Task Execution Workflow

### Task Completion Process

1. **Pre-Completion Confirmation**: Always ask user before declaring task completion
2. **Git Status Review**: Run `git status` after completion to identify:
   - Files that need to be committed
   - Files that should be discarded
   - Ask user if uncertain about any changes
3. **Version Control**: Execute `git commit` and `git push` after user confirmation
4. **Documentation Updates**: Before completing each task, check and update:
   - Steering documents (`.kiro/steering/`) if workflow or standards change
   - Spec documents (`.kiro/specs/`) if requirements or design evolve
   - Feature documentation (`.dev-docs/`) for internal implementation details
   - Original Pyrra documentation if user-facing features are added

### Implementation Guidelines

- **Simplicity First**: Don't overcomplicate implementations
- **Minimal File Creation**: Avoid creating unnecessary files - think before implementing
- **Incremental Approach**: Build on existing patterns and infrastructure
- **Quality Gates**: Follow systematic comparison and comprehensive validation standards

### Task-Based Development Approach

The feature uses focused task-based development with specific implementation groups:

- **Task Group 1**: Latency UI completion and validation (HIGH PRIORITY)
- **Task Group 2**: Alerts table enhancement
- **Task Group 4**: Resilience testing (ALTERNATIVE PRIORITY)
- **Task Groups 3, 5-8**: Additional indicator types, alert firing, production polish

Each task should follow the quality standards checklist and update documentation with results.
