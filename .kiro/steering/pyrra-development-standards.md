---
inclusion: always
---

# Pyrra Development Standards and Context

## Project Overview

Pyrra is a Service Level Objective (SLO) management tool for Kubernetes environments using Prometheus metrics. It creates recording and alerting rules for SLOs, supports different indicator types (ratio, latency, boolean), and integrates with Prometheus Operator for rule deployment.

## Current Feature: Dynamic Burn Rate Implementation

### Core Innovation

Dynamic burn rate alerting adapts **alert thresholds** based on actual traffic patterns rather than using fixed static multipliers. This prevents false positives during low traffic and false negatives during high traffic periods.

**CRITICAL UNDERSTANDING**: Dynamic burn rates only affect alert threshold calculations. Error budget calculations remain identical between static and dynamic burn rates.

**Error Budget Formula (Same for Both Static and Dynamic)**:

```
error_budget_remaining = ((1 - SLO_target) - (1 - success/total)) / (1 - SLO_target)
```

**Dynamic Alert Threshold Formula**:

```
dynamic_threshold = (N_SLO / N_alert) × E_budget_percent_threshold × (1 - SLO_target)
```

Where:

- N_SLO = Number of events in SLO window (e.g., 28 days)
- N_alert = Number of events in alert window (e.g., 1 hour)
- E_budget_percent_threshold = Constant percentage (1/48, 1/16, 1/14, 1/7)
- (1 - SLO_target) = Error budget (e.g., 0.01 for 99% SLO)

### Key Distinction: Alert Thresholds vs Error Budget

- **Static Burn Rate**: Fixed alert thresholds (e.g., `burn_rate > 14 × (1 - SLO_target)`)
- **Dynamic Burn Rate**: Traffic-aware alert thresholds (formula above)
- **Error Budget**: Identical calculation for both approaches using success/total ratio

## Development Standards

### Code Quality Requirements

1. **Systematic Comparison**: Always compare with working examples, never analyze in isolation
2. **Comprehensive Structure Validation**: Check naming conflicts, complete coverage, syntax validation, label consistency
3. **Question Everything**: Understand purpose of every component, explain deviations, identify gaps
4. **Test Before Success**: Syntax validation, functional testing, edge cases, performance impact
5. **Document Issues**: Categorize severity, provide context, suggest concrete fixes
6. **Knowledge-Based Task Creation**: Tasks and specifications must be created only based on concrete knowledge after validation. Never formulate tasks based on assumptions about system behavior, mathematical formulas, or architectural patterns without first verifying the actual implementation.

### Architecture Patterns

#### UI Development Critical Workflow

Pyrra uses **two different UI serving methods**:

1. **Development UI (Port 3000)**: `npm start` - live source files with hot reload
2. **Embedded UI (Port 9099)**: `./pyrra api` - compiled files via Go embed

**Development UI Workflow**:

```bash
# 1. Make changes in ui/src/
# 2. Test in development: npm start → http://localhost:3000 (sufficient for most development)
# 3. Optional production validation: npm run build + make build + ./pyrra api → http://localhost:9099
```

**Note**: Port 3000 development UI is sufficient for most development work. Complete rebuild workflow only needed for final validation.

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

- **Primary development**: `cd ui && npm start` (port 3000) for live reload - sufficient for most development
- **Production UI testing**: `npm run build` + embedded UI only for final validation when needed
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

#### API Testing Commands

**Local Development API Structure**: Pyrra uses **two separate services** with different API endpoints:

1. **API Service** (`./pyrra api`):
   - **Port**: 9099
   - **Purpose**: Full-featured API server with embedded UI
   - **Endpoint**: `/objectives.v1alpha1.ObjectiveService/List`
   - **Features**: Complete ObjectiveService with List, GetStatus, GetAlerts, Graph* methods

2. **Kubernetes Backend Service** (`./pyrra kubernetes`):
   - **Port**: 9444
   - **Purpose**: Lightweight backend that connects to Kubernetes cluster
   - **Endpoint**: `/objectives.v1alpha1.ObjectiveBackendService/List`
   - **Features**: Limited ObjectiveBackendService with only List method

**Correct API Testing Commands**:

```bash
# Test Kubernetes Backend Service (port 9444)
curl -X POST -H "Content-Type: application/json" -d '{}' \
  "http://localhost:9444/objectives.v1alpha1.ObjectiveBackendService/List"

# Test Full API Service (port 9099)
curl -X POST -H "Content-Type: application/json" -d '{}' \
  "http://localhost:9099/objectives.v1alpha1.ObjectiveService/List"
```

**Protocol Details**:
- **Protocol**: Connect protocol (gRPC-Web compatible)
- **Content-Type**: `application/json`
- **Method**: POST requests to service endpoints
- **Request Body**: JSON payload (e.g., `{}` for List requests)

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

1. **MANDATORY Pre-Development Type Research**: ALWAYS understand types and APIs before writing code:

   **A. Check Actual Type Definitions First:**

   - **Read .d.ts files** for protobuf messages, React components, and external libraries
   - **Never assume API structure** - always verify actual type definitions
   - **Example**: Before using `Objective`, read `ui/src/proto/objectives/v1alpha1/objectives_pb.d.ts`
   - **Check constructor signatures** and required vs optional properties

   **B. Follow Existing Patterns:**

   - **Always examine working examples** before writing new code
   - **Look at existing test files** to see how types are used correctly
   - **Copy proven patterns** rather than inventing new approaches
   - **Example**: Check `ui/src/components/BurnRateThresholdDisplay.spec.tsx` before writing new tests

   **C. Understand API Surface Before Implementation:**

   - **Read component interfaces** and prop types
   - **Check import statements** in existing files to understand correct imports
   - **Verify mock patterns** used in existing tests
   - **Test small pieces first** rather than writing large files

2. **Code Quality Validation**: ALWAYS check for compilation and syntax errors before proceeding:

   - **Kiro IDE Problems Tab**: Check the Problems tab in Kiro IDE for TypeScript errors, syntax issues, and other compilation problems
   - **Ask User for Problem Report**: If unable to resolve compilation errors, ask user to check Problems tab and report specific errors with file names, line numbers, and error messages
   - **Fix Before Proceeding**: All compilation errors must be resolved before moving to testing phase
   - **Example Request**: "Please check the Problems tab in Kiro IDE and let me know if there are any TypeScript errors or compilation issues in the files I modified"

3. **Implementation Testing**: ALWAYS test the changes made within the task before asking for approval:

   - **Steering documents** (`.kiro/steering/`) if workflow or standards change
   - **Spec documents** (`.kiro/specs/`) if requirements or design evolve
   - **Feature documentation** (`.dev-docs/`) for internal implementation details
   - **Original Pyrra documentation** if user-facing features are added
   - **NEVER skip this step** - documentation updates are required before approval

4. **MANDATORY Documentation Updates**: BEFORE asking for approval, ALWAYS update relevant documentation:

   - **Steering documents** (`.kiro/steering/`) if workflow or standards change
   - **Spec documents** (`.kiro/specs/`) if requirements or design evolve
   - **Feature documentation** (`.dev-docs/`) for internal implementation details
   - **Original Pyrra documentation** if user-facing features are added
   - **NEVER skip this step** - documentation updates are required before approval

5. **Pre-Completion Confirmation**: MANDATORY - Only after successful type research, code validation, testing AND documentation updates, ask user for explicit approval before declaring task completion

   - **NEVER declare task completion without user approval**
   - **ALL sub-tasks must be completed and tested**
   - **Ask explicitly**: "Are you ready for me to mark this task as complete?"
   - **Wait for explicit approval** (e.g., "yes", "approved", "complete it")
   - **If user identifies missing work, continue implementation**

6. **Git Status Review**: Run `git status` after completion to identify:

   - Files that need to be committed
   - Files that should be discarded
   - Ask user if uncertain about any changes

7. **Version Control**: Execute `git commit` and `git push` after user confirmation

### IDE Integration and Problem Resolution

#### Kiro IDE Problems Tab Workflow

**MANDATORY**: Always check and resolve IDE problems before task completion:

1. **Problem Detection**: Ask user to check Problems tab in Kiro IDE after making code changes
2. **Problem Reporting**: Request specific error details including:
   - File path and name
   - Line and column numbers
   - Error codes (e.g., TS1128, TS2345)
   - Complete error messages
3. **Problem Resolution**: Fix all TypeScript errors, syntax issues, and compilation problems
4. **Verification**: Confirm with user that Problems tab is clear before proceeding

**Example Problem Report Format**:

```json
[
  {
    "resource": "/path/to/file.tsx",
    "owner": "typescript",
    "code": "1128",
    "severity": 8,
    "message": "Declaration or statement expected.",
    "startLineNumber": 850,
    "startColumn": 1
  }
]
```

#### Common Problem Categories

- **TypeScript Errors**: Type mismatches, missing imports, interface violations
- **Syntax Errors**: Missing brackets, semicolons, malformed statements
- **Import/Export Issues**: Incorrect module references, circular dependencies
- **Test File Problems**: Mock configuration, type definitions, test structure

#### TypeScript Test File Development Standards

**MANDATORY Pre-Writing Checklist for .spec.tsx files:**

1. **Type Definition Research (REQUIRED):**

   ```bash
   # Always check these before writing tests:
   - Read component .d.ts files for protobuf messages
   - Check existing .spec.tsx files for proven patterns
   - Verify import statements and mock structures
   - Understand constructor signatures and required properties
   ```

2. **Proven Pattern Copying (REQUIRED):**

   ```typescript
   // GOOD: Copy from working test file
   import { ConnectError } from '@connectrpc/connect'
   error: { message: 'Error text' } as ConnectError

   // BAD: Assume constructor exists
   error: new ConnectError('Error text', 'CODE')
   ```

3. **Protobuf Message Construction (REQUIRED):**

   ```typescript
   // GOOD: Check .d.ts file first, use direct properties
   new Objective({
     target: 0.99,
     labels: { __name__: 'test' },
     indicator: { value: { ratio: { ... } } }
   })

   // BAD: Assume nested spec property exists
   new Objective({
     spec: { target: '99', indicator: { ... } }
   })
   ```

4. **Mock Pattern Verification (REQUIRED):**

   ```typescript
   // GOOD: Copy exact mock pattern from working tests
   jest.mock("../prometheus", () => ({
     usePrometheusQuery: jest.fn(),
   }));

   // BAD: Invent new mock patterns
   ```

**Test File Error Prevention Rules:**

- **NEVER write test files without reading existing working examples first**
- **ALWAYS verify type definitions in .d.ts files before using types**
- **ALWAYS ask user to check Problems tab immediately after creating test files**
- **ALWAYS fix ALL TypeScript errors before proceeding with test execution**

### Implementation Guidelines

- **Simplicity First**: Don't overcomplicate implementations
- **Minimal File Creation**: Avoid creating unnecessary files - think before implementing
- **Incremental Approach**: Build on existing patterns and infrastructure
- **Quality Gates**: Follow systematic comparison and comprehensive validation standards
- **IDE-First Development**: Use Kiro IDE Problems tab as primary quality gate

### Task-Based Development Approach

The feature uses focused task-based development with specific implementation groups:

- **Task Group 1**: Latency UI completion and validation (HIGH PRIORITY)
- **Task Group 2**: Alerts table enhancement
- **Task Group 4**: Resilience testing (ALTERNATIVE PRIORITY)
- **Task Groups 3, 5-8**: Additional indicator types, alert firing, production polish

Each task should follow the quality standards checklist and update documentation with results.

## Task 6 Development Lessons Learned

### Over-Engineering Anti-Pattern

**Issue**: Created multiple CLI tools without validating each component works individually.

**Example**: Built `validate-alerts`, `alert-test`, `test-alert-firing`, `precision-recall-test` tools in parallel, but only `run-synthetic-test` actually worked.

**Root Cause**: Violated "Test Before Success" principle by building theoretical frameworks instead of validating working solutions.

**Solution**: 
- Build ONE working tool first
- Enhance it incrementally with testing at each step
- Don't create parallel implementations until the first one is proven

### Requirements Misinterpretation Pattern

**Issue**: Interpreted "synthetic metric generation for alert testing" as needing multiple specialized tools.

**Reality**: Task needed ONE tool that generates synthetic metrics AND validates alerts fire.

**Lesson**: Focus on actual requirements, not theoretical completeness. Ask clarifying questions before building complex solutions.

### Working Solution Abandonment Anti-Pattern

**Issue**: Instead of enhancing working `run-synthetic-test` tool, created new parallel tools.

**Impact**: Wasted effort on broken implementations while working solution remained incomplete.

**Solution**: Always build on proven success rather than starting over. Enhance working solutions incrementally.

### Key Development Principles Reinforced

1. **Simplicity First**: Make it work, then make it better
2. **Test Continuously**: Validate each component as you build it  
3. **Build on Success**: Enhance working solutions rather than creating parallel ones
4. **Understand Requirements**: Focus on actual needs, not theoretical completeness
5. **Follow Steering Guidelines**: The development standards exist to prevent these exact issues
## Tes
ting Environment Reference

For comprehensive testing environment setup and configuration details, see:
- **`.dev-docs/TESTING_ENVIRONMENT_REFERENCE.md`** - Complete service architecture, URLs, ports, and timing expectations for alert testing

This document provides the definitive reference for:
- Service endpoints and port configurations
- Alert timing expectations (30s Prometheus evaluation interval)
- Required service dependencies and startup order
- Pre-test validation checklist
- Common troubleshooting scenarios

All testing code must reference this document to ensure correct service URLs and timing expectations.
## Addi
tional Development Documentation

### Task 6 Development Lessons
- **`.dev-docs/TASK_6_LESSONS_LEARNED.md`** - Lessons learned from task 6 development process issues

### Testing Environment Reference  
- **`.dev-docs/TESTING_ENVIRONMENT_REFERENCE.md`** - Complete testing environment setup and configuration reference