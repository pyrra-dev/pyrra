# Task 8.0 - Pre-Merge Cleanup Analysis

## Overview

This document provides a comprehensive analysis of all files modified in the dynamic burn rate feature branch, categorizing them for cleanup decisions before upstream merge.

**Analysis Date:** 2025-10-13
**Branch:** add-dynamic-burn-rate
**Base:** upstream/main (pyrra-dev/pyrra)
**Total Files Modified:** 132

---

## File Categories

### Category 1: Fork-Only Files (Should NOT go to upstream)

These files are specific to our development process and should remain in the fork only:

#### .kiro/ Directory (All files)
- `.kiro/hooks/ai-behavior-reset.kiro.hook`
- `.kiro/specs/dynamic-burn-rate-completion/design.md`
- `.kiro/specs/dynamic-burn-rate-completion/requirements.md`
- `.kiro/specs/dynamic-burn-rate-completion/tasks.md`
- `.kiro/steering/ai-behavior-reminder-checklist.md`
- `.kiro/steering/ai-session-management-strategy.md`
- `.kiro/steering/dynamic-burn-rate-context.md`
- `.kiro/steering/pyrra-development-standards.md`

**Action:** Add `.kiro/` to .gitignore for upstream PR

#### .dev-docs/ Directory (All files)
All 40+ documentation files in `.dev-docs/` are development session notes and internal documentation.

**Action:** Keep in fork, exclude from upstream PR

#### prompts/ Directory (All files)
All session prompt files are development artifacts.

**Action:** Keep in fork, exclude from upstream PR

#### Development Tools (cmd/)
- `cmd/generate-test-slos/` - Test data generation
- `cmd/monitor-performance/` - Performance monitoring
- `cmd/run-synthetic-test/` - Synthetic testing
- `cmd/test-burnrate-threshold-queries/` - Query testing
- `cmd/test-health-check/` - Health check testing
- `cmd/test-query-aggregation/` - Aggregation testing
- `cmd/validate-alert-rules/` - Alert validation
- `cmd/validate-recording-rules-basic/` - Recording rule validation
- `cmd/validate-recording-rules-focused/` - Focused validation
- `cmd/validate-recording-rules-native/` - Native histogram validation
- `cmd/validate-ui-query-optimization/` - UI query optimization testing

**Action:** Keep in fork, exclude from upstream PR (these are testing/validation tools)

#### Test Scripts (scripts/)
- `scripts/production-readiness-test.sh`
- `scripts/test_scientific_notation.py`
- `scripts/validate-alert-rules.sh`
- `scripts/validate-recording-rules.sh`
- `scripts/validate_math_correctness.py`

**Action:** Keep in fork, exclude from upstream PR

#### Development Configuration
- `Dockerfile.custom` - Custom development dockerfile
- `Dockerfile.dev` - Development dockerfile

**Action:** Keep in fork, exclude from upstream PR

---

### Category 2: Core Feature Files (MUST go to upstream)

These files implement the dynamic burn rate feature:

#### Backend Implementation
- ✅ `slo/rules.go` - Core dynamic burn rate logic
- ✅ `slo/rules_test.go` - Comprehensive tests
- ✅ `slo/slo.go` - SLO types and interfaces
- ✅ `kubernetes/api/v1alpha1/servicelevelobjective_types.go` - CRD definition
- ✅ `kubernetes/api/v1alpha1/servicelevelobjective_types_test.go` - CRD tests
- ✅ `main.go` - CLI integration

#### API/Protocol
- ✅ `proto/objectives/v1alpha1/objectives.proto` - Protobuf definition
- ✅ `proto/objectives/v1alpha1/objectives.pb.go` - Generated Go code
- ✅ `proto/objectives/v1alpha1/objectives.go` - Protocol helpers

#### UI Implementation
- ✅ `ui/src/burnrate.tsx` - Burn rate utilities
- ✅ `ui/src/components/AlertsTable.tsx` - Enhanced alerts table
- ✅ `ui/src/components/BurnRateThresholdDisplay.tsx` - Threshold display component
- ✅ `ui/src/components/Icons.tsx` - Icon additions
- ✅ `ui/src/components/graphs/BurnrateGraph.tsx` - Dynamic threshold graphs
- ✅ `ui/src/components/graphs/DurationGraph.tsx` - Duration graph enhancements
- ✅ `ui/src/components/graphs/ErrorBudgetGraph.tsx` - Error budget enhancements
- ✅ `ui/src/components/graphs/ErrorsGraph.tsx` - Errors graph enhancements
- ✅ `ui/src/components/graphs/RequestsGraph.tsx` - Traffic baseline visualization
- ✅ `ui/src/pages/Detail.tsx` - Detail page enhancements
- ✅ `ui/src/pages/List.tsx` - List page enhancements
- ✅ `ui/src/proto/objectives/v1alpha1/objectives_pb.d.ts` - TypeScript definitions
- ✅ `ui/src/proto/objectives/v1alpha1/objectives_pb.js` - Generated JS code
- ✅ `ui/src/utils/numberFormat.ts` - Scientific notation formatting
- ✅ `ui/src/utils/numberFormat.spec.ts` - Number format tests

#### Generated CRD Files
- ✅ `jsonnet/controller-gen/pyrra.dev_servicelevelobjectives.json`
- ✅ `jsonnet/controller-gen/pyrra.dev_servicelevelobjectives.yaml`

---

### Category 3: Files Needing Review/Cleanup

#### 🔍 CONTRIBUTING.md
**Changes:** Added UI development workflow documentation (port 3000 vs 9099)

**Analysis:**
- Helpful documentation for contributors
- Explains development vs production UI serving
- Not directly related to dynamic burn rate feature
- Improves contributor experience

**Recommendation:** ⚠️ **KEEP** - This is valuable contributor documentation

---

#### 🔍 examples/kubernetes/manifests/pyrra-kubernetesDeployment.yaml
**Changes:** Added resource limits and requests

```yaml
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi
```

**Analysis:**
- Not related to dynamic burn rate feature
- Good practice but changes example file
- May have been added during testing

**Recommendation:** ❌ **REVERT** - Unrelated to feature, changes example configuration

---

#### 🔍 ui/public/index.html
**Changes:** Template variables replaced with specific values

```html
<!-- Before (template) -->
<script>window.PATH_PREFIX = {{.PathPrefix}}</script>
<script>window.API_BASEPATH = {{.APIBasepath}}</script>

<!-- After (specific values) -->
<script>window.PATH_PREFIX = "/"</script>
<script>window.API_BASEPATH = undefined</script>
```

**Analysis:**
- This file should have template placeholders
- Likely modified by build process or accidental commit
- Should be reverted to template format

**Recommendation:** ❌ **REVERT** - Must use template placeholders

---

#### 🔍 .gitignore
**Changes:** Added development-specific ignores

```gitignore
# Added entries:
.vscode
pyrra-*
/.dev
.envrc
.dev-docs/*-slos.json
/generate-test-slos
/monitor-performance
```

**Analysis:**
- `.vscode` - IDE-specific, reasonable to add
- `pyrra-*` - Catches generated binaries, reasonable
- `/.dev` - Fork-specific development directory
- `.envrc` - direnv configuration, reasonable for contributors
- `.dev-docs/*-slos.json` - Fork-specific test data
- Test binaries - Fork-specific tools

**Recommendation:** ⚠️ **PARTIAL KEEP**
- Keep: `.vscode`, `pyrra-*`, `.envrc` (helpful for contributors)
- Remove: `/.dev`, `.dev-docs/*-slos.json`, test binary entries (fork-specific)

---

#### 🔍 filesystem.go
**Changes:** Added native histogram support for filesystem mode

**Analysis:**
- Adds custom duration histogram with native histogram support (`connect_server_requests_duration_seconds`)
- Adds duration interceptor for connect server requests
- **Mirrors identical changes in main.go** (API server / kubernetes mode)
- Both main.go and filesystem.go updated to emit native histogram metrics for LatencyNative indicator testing
- **CRITICAL**: Only kubernetes mode (main.go) has been tested, filesystem mode NOT tested

**Recommendation:** ✅ **KEEP** - Changes mirror main.go and are needed for LatencyNative support
- Filesystem mode should work identically to kubernetes mode (same code pattern)
- Both modes need native histogram metric emission for LatencyNative indicators
- Testing filesystem mode is optional but recommended for completeness
- If filesystem mode testing is desired, add to Task 9.3

---

#### 🔍 ui/src/components/Toggle.tsx
**Changes:** Added `readOnly` prop to checkbox input

```tsx
// Before
<input type="checkbox" checked={checked} />

// After
<input type="checkbox" checked={checked} readOnly />
```

**Analysis:**
- Prevents React warning about controlled component without onChange
- Toggle has onClick on parent div, not onChange on input
- This is a React best practice fix
- Not directly related to dynamic burn rate but improves code quality

**Recommendation:** ✅ **KEEP** - Fixes React warning, improves code quality

---

#### 🔍 ui/DYNAMIC_BURN_RATE_UI.md
**Status:** Old documentation file

**Analysis:**
- Likely early design/planning document
- Content may be duplicated in .dev-docs/
- Historical artifact

**Recommendation:** ⚠️ **NEEDS REVIEW**
- Review content for unique information
- If valuable, move to `.dev-docs/`
- If duplicated, delete

---

### Category 4: Example Files

#### ✅ examples/latency-dynamic-burnrate.yaml
**Status:** New example file for dynamic burn rate latency indicator

**Recommendation:** ✅ **KEEP** - Good example for users

#### ✅ examples/simple-demo.yaml
**Changes:** Need to review what changed

**Recommendation:** ⚠️ **REVIEW** - Check if changes are intentional

---

### Category 5: Testing Files

#### ✅ testing/README.md
**Changes:** Updated testing documentation

**Recommendation:** ✅ **KEEP** - Improved testing docs

#### ✅ testing/prometheus_alerts.go
**Changes:** Alert testing utilities

**Recommendation:** ✅ **KEEP** - Part of testing infrastructure

#### ✅ testing/pushgateway-scrape-config.yaml
**Changes:** Pushgateway configuration

**Recommendation:** ✅ **KEEP** - Testing configuration

#### ✅ testing/service_health_check.go
**Changes:** Health check utilities

**Recommendation:** ✅ **KEEP** - Testing utilities

#### ✅ testing/synthetic-slo.yaml
**Changes:** Synthetic SLO for testing

**Recommendation:** ✅ **KEEP** - Testing configuration

#### ✅ testing/synthetic_metrics.go
**Changes:** Synthetic metric generation

**Recommendation:** ✅ **KEEP** - Testing utilities

---

#### ⚠️ ui/src/components/BurnRateThresholdDisplay.spec.tsx
**Status:** New test file for BurnRateThresholdDisplay component

**Analysis:**
- Comprehensive unit tests for new component
- Follows React testing best practices
- Upstream may or may not have UI test conventions

**Recommendation:** ⚠️ **NEEDS DECISION**
- Check if upstream has UI testing conventions
- If yes, keep and ensure tests follow conventions
- If no, consider keeping anyway as good practice

---

### Category 6: Build/Configuration Files

#### ✅ go.mod / go.sum
**Changes:** Dependency updates

**Recommendation:** ✅ **KEEP** - Required dependency changes

#### ✅ ui/README.md
**Changes:** Updated UI documentation

**Recommendation:** ⚠️ **REVIEW** - Check if changes are feature-related

---

## Code Cleanup Tasks

### Backend Cleanup (slo/rules.go)

#### Duplicate Selector Functions
Found duplicate code that should be consolidated:
- `buildTotalSelector`
- `buildLatencyTotalSelector`
- `buildLatencyNativeTotalSelector`
- `buildBoolGaugeSelector`

**Analysis:** These functions have similar logic but handle different indicator types.

**Recommendation:** ⚠️ **REVIEW CAREFULLY**
- Check if consolidation is possible without making code more complex
- If logic is significantly different, keep separate with clear comments
- If logic is similar, consolidate into generic function

#### Comment Updates
Found comments with "originally X for Y" format that should be updated to "X for Y SLO period"

**Recommendation:** ✅ **UPDATE** - Simple comment clarification

---

### Backend Cleanup (slo/slo.go)

#### Unused Function: GetRemainingErrorBudget
**Usage:** Only found in definition, not used anywhere in codebase

**Recommendation:** ❌ **REMOVE** - Unused function

#### Unused Struct: DynamicBurnRate (in slo/slo.go)
**Usage:** 
- Defined in `slo/slo.go`
- Also defined in `kubernetes/api/v1alpha1/servicelevelobjective_types.go` (CRD)
- NOT used in actual code
- Test functions named `TestObjective_DynamicBurnRate*` but they test the feature, not the struct

**Recommendation:** ❌ **REMOVE from slo/slo.go** - Unused struct (keep CRD version)

---

### CRD Cleanup (kubernetes/api/v1alpha1/servicelevelobjective_types.go)

#### DynamicBurnRate Struct Fields
Found fields in CRD:
- `BaseFactor` - Used in variable name `baseFactors` in rules.go but not the struct field
- `MinFactor` - Not found in usage search
- `MaxFactor` - Not found in usage search

**Analysis:** The `DynamicBurnRate` struct in CRD appears to be unused. The actual implementation doesn't use these fields.

**Recommendation:** ⚠️ **NEEDS CAREFUL REVIEW**
- Verify if CRD struct is actually used
- Check if this was planned future functionality
- If unused, remove entire struct from CRD
- If removing, regenerate CRD YAML/JSON files

---

### Test File Review

#### kubernetes/api/v1alpha1/servicelevelobjective_types_test.go
**Status:** Only static test cases added

**Recommendation:** ⚠️ **CONSIDER** adding dynamic burn rate test cases for completeness

#### ui/src/components/BurnRateThresholdDisplay.spec.tsx
**Status:** Comprehensive UI component tests

**Recommendation:** ⚠️ **DECISION NEEDED** - Keep for upstream or fork-only?

---

## Summary of Actions Needed

### Immediate Actions (Before Upstream PR)

1. ❌ **REVERT** `examples/kubernetes/manifests/pyrra-kubernetesDeployment.yaml` - Remove resource limits
2. ❌ **REVERT** `ui/public/index.html` - Restore template placeholders
3. ⚠️ **PARTIAL CLEANUP** `.gitignore` - Keep useful entries, remove fork-specific
4. ❌ **REMOVE** `slo/slo.go::GetRemainingErrorBudget()` - Unused function
5. ❌ **REMOVE** `slo/slo.go::DynamicBurnRate` struct - Unused (keep CRD version if used)
6. ⚠️ **REVIEW** `slo/rules.go` - Consolidate duplicate selector functions if possible
7. ✅ **UPDATE** `slo/rules.go` - Fix comment format ("originally" → "for")
8. ⚠️ **REVIEW** `kubernetes/api/v1alpha1/servicelevelobjective_types.go` - DynamicBurnRate struct usage
9. ⚠️ **REVIEW** `ui/DYNAMIC_BURN_RATE_UI.md` - Move to .dev-docs or delete
10. ⚠️ **DECISION** `filesystem.go` - Test filesystem mode or document limitation

### Files to Exclude from Upstream PR

- All `.kiro/` files
- All `.dev-docs/` files
- All `prompts/` files
- All `cmd/` testing tools
- All `scripts/` testing scripts
- `Dockerfile.custom`
- `Dockerfile.dev`

### Files Requiring Decisions

1. `CONTRIBUTING.md` - Keep helpful documentation?
2. `filesystem.go` - Test or document as limitation?
3. `ui/DYNAMIC_BURN_RATE_UI.md` - Move or delete?
4. `ui/src/components/BurnRateThresholdDisplay.spec.tsx` - Include in upstream?
5. `kubernetes/api/v1alpha1/servicelevelobjective_types.go` - Remove DynamicBurnRate struct?

---

## Next Steps

1. **User Decisions Required** - Get user input on decision points above
2. **Execute Cleanup** - Perform reverts and removals
3. **Test After Cleanup** - Ensure feature still works after cleanup
4. **Update Checklist** - Mark completed items in TASK_8.0_PRE_MERGE_CLEANUP_CHECKLIST.md
5. **Create Summary** - Document all cleanup actions taken
6. **Proceed to Task 8.1** - Fetch and merge from upstream

---

## References

- **Checklist:** `.dev-docs/TASK_8.0_PRE_MERGE_CLEANUP_CHECKLIST.md`
- **Contribution Plan:** `.dev-docs/UPSTREAM_CONTRIBUTION_PLAN.md`
- **Feature Summary:** `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md`
