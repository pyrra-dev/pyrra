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

#### Development Environment Setup

The testing environment consists of:

- **Kubernetes Cluster**: Minikube with Hyper-V on Windows 10
- **Monitoring Stack**: kube-prometheus (jsonnet-based) providing Prometheus, Grafana, AlertManager
- **Pyrra Services**: Local binaries (`./pyrra api` on port 9099, `./pyrra kubernetes` on port 9444)
- **Development UI**: `npm start` in ui/ folder (port 3000) for live development
- **Test SLOs**: Both static and dynamic SLOs from examples/ and .dev/ folders

#### Testing Approach Methodology

**Two-Tier Testing Strategy:**

1. **Terminal-Based Tests**: AI performs direct commands for faster iteration

   - Prometheus API queries via curl
   - Kubernetes API validation
   - Mathematical calculations and cross-validation
   - Performance measurements and logging analysis

2. **UI-Based Tests**: Interactive human-guided validation
   - Human operator performs queries in Prometheus UI (port 9090)
   - Human operator tests functionality in Pyrra UI (port 3000 for development)
   - AI guides specific test scenarios and interprets feedback
   - Visual validation of tooltips, error states, and user experience

**Development Workflow:**

- Primary development uses `cd ui && npm start` (port 3000) for live reload
- Production UI testing (`npm run build` + embedded UI) only for final validation
- Multiple terminals for different services (API, backend, UI) managed by human operator

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
- Use terminal commands for mathematical verification (no LLM calculations)

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

**Static SLOs:**

- `.dev/test-slo.yaml` - Basic static SLO for development
- `.dev/test-static-slo.yaml` - Static burn rate test configuration
- `examples/` - Original Pyrra project examples (mostly static)

**Dynamic SLOs:**

- `.dev/test-dynamic-slo.yaml` - Dynamic burn rate ratio indicator
- `.dev/test-latency-dynamic.yaml` - Dynamic burn rate latency indicator

**Environment Setup:**

- `.dev/minikube-env.sh` - Docker environment variables for Minikube integration
- `.envrc` - direnv configuration for automatic environment loading

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

1. **Implementation Testing**: ALWAYS test the changes made within the task before asking for approval:

   - **Interactive Testing Approach**: Guide user through small, focused test steps rather than long test lists
   - Test in development UI first: Use existing `npm start` on port 3000 for immediate feedback
   - Guide one small test step at a time: "Please do X and tell me what you see"
   - Wait for user feedback before proceeding to next test step
   - Check browser console for errors or warnings
   - Test edge cases and error scenarios when applicable
   - Validate performance impact if performance-related changes were made
   - Only if development testing passes, then optionally test embedded UI: `npm run build && make build && ./pyrra api`

2. **MANDATORY Documentation Updates**: BEFORE asking for approval, ALWAYS update relevant documentation:

   - **Steering documents** (`.kiro/steering/`) if workflow or standards change
   - **Spec documents** (`.kiro/specs/`) if requirements or design evolve
   - **Feature documentation** (`.dev-docs/`) for internal implementation details
   - **Original Pyrra documentation** if user-facing features are added
   - **NEVER skip this step** - documentation updates are required before approval

3. **Pre-Completion Confirmation**: Only after successful testing AND documentation updates, ask user before declaring task completion

4. **Git Status Review**: Run `git status` after completion to identify:

   - Files that need to be committed
   - Files that should be discarded
   - Ask user if uncertain about any changes

5. **Version Control**: Execute `git commit` and `git push` after user confirmation

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
