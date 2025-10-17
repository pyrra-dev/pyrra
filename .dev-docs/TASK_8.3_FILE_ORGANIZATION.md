# Task 8.3: File Organization for PR vs Fork Separation

## Overview

This document provides a comprehensive categorization of all files in the repository, identifying which files should be included in the upstream pull request versus which should remain in the fork only. This organization ensures a clean PR focused on the core feature while preserving all development artifacts in the fork.

## Organization Principles

1. **PR Files**: Core implementation, essential tests, minimal user-facing documentation
2. **Fork Files**: Development tools, extensive testing infrastructure, internal documentation
3. **Preserve Everything**: All fork-only files will be preserved in `dev-tools-and-docs` branch
4. **Minimal Documentation**: Keep upstream docs proportional - dynamic burn rate is ONE feature among many

## File Categories

### Category 1: Core Implementation (INCLUDE IN PR)

**Backend Implementation:**

- `slo/rules.go` - Dynamic burn rate alert rule generation
- `slo/slo.go` - Core types and interfaces
- `slo/promql.go` - PromQL query generation helpers
- `slo/promql_test.go` - Unit tests for PromQL generation
- `slo/rules_test.go` - Unit tests for rule generation
- `kubernetes/api/v1alpha1/servicelevelobjective_types.go` - CRD with burnRateType field

**UI Implementation:**

- `ui/src/` - All UI components (entire directory)
  - Core components with dynamic burn rate support
  - Enhanced tooltips and traffic visualization
  - Error handling and performance optimizations
  - All React components, TypeScript files, tests

**Protobuf Definitions:**

- `proto/objectives/v1alpha1/objectives.proto` - API definitions with burnRateType
- Generated protobuf files in `ui/src/proto/`

**Build Configuration:**

- `Makefile` - Build system (review for any dev-only targets)
- `go.mod` / `go.sum` - Go dependencies
- `ui/package.json` / `ui/package-lock.json` - UI dependencies

### Category 2: Essential Documentation (INCLUDE IN PR - MINIMAL UPDATES)

**User-Facing Documentation:**

- `README.md` - Add brief dynamic burn rate section (2-3 paragraphs max)
- `examples/README.md` - Add brief explanation of dynamic examples (1-2 paragraphs)

**Example Configurations:**

- `examples/dynamic-burn-rate-ratio.yaml` - NEW: Simple ratio example
- `examples/dynamic-burn-rate-latency.yaml` - NEW: Simple latency example
- `examples/dynamic-burn-rate-latency-native.yaml` - EXISTING: Keep as is
- `examples/dynamic-burn-rate-bool-gauge.yaml` - EXISTING: Keep as is

**Optional New Documentation:**

- `docs/DYNAMIC_BURN_RATE.md` - OPTIONAL: Only if it keeps main docs clean

### Category 3: Development Tools (FORK ONLY)

**Validation Tools (`cmd/`):**

- `cmd/validate-ui-query-optimization/` - Query performance validation
- `cmd/test-burnrate-threshold-queries/` - Threshold calculation testing
- `cmd/test-query-aggregation/` - Query aggregation validation
- `cmd/validate-alert-rules/` - Alert rule validation
- `cmd/validate-recording-rules-basic/` - Basic recording rule validation
- `cmd/validate-recording-rules-focused/` - Focused recording rule validation
- `cmd/validate-recording-rules-native/` - Native histogram validation
- `cmd/run-synthetic-test/` - Synthetic metric generation for alert testing
- `cmd/monitor-performance/` - Performance monitoring
- `cmd/generate-test-slos/` - Test SLO generation
- `cmd/test-health-check/` - Health check testing

**Built Executables (FORK ONLY - TEMPORARY):**

- `*.exe` files in root directory (all validation tools)
- These are build artifacts, not source code

### Category 4: Development Documentation (FORK ONLY)

**Internal Documentation (`.dev-docs/`):**
All 40+ documents including:

- `AI_DEVELOPMENT_QUALITY_STANDARDS.md`
- `BROWSER_COMPATIBILITY_MATRIX.md`
- `BROWSER_COMPATIBILITY_TEST_GUIDE.md`
- `CORE_CONCEPTS_AND_TERMINOLOGY.md`
- `dynamic-burn-rate-implementation.md`
- `dynamic-burn-rate.md`
- `DYNAMIC_BURN_RATE_TESTING_SESSION.md`
- `FEATURE_IMPLEMENTATION_SUMMARY.md`
- `HISTORICAL_UI_DESIGN.md`
- `ISSUE_REGEX_LABEL_SELECTORS.md`
- `KNOWN_LIMITATIONS.md`
- `MIGRATION_GUIDE.md`
- `NATIVE_HISTOGRAM_SETUP_GUIDE.md`
- `PRODUCTION_PERFORMANCE_BENCHMARKS.md`
- `SETUP_SUMMARY.md`
- `sli-indicator-types.md`
- `TASK_*` documents (all task implementation summaries)
- `TESTING_ENVIRONMENT_REFERENCE.md`
- `UPSTREAM_COMPARISON_REGEX_SELECTORS.md`
- `UPSTREAM_CONTRIBUTION_PLAN.md`
- All JSON baseline files

### Category 5: Development Scripts (FORK ONLY)

**Validation Scripts (`scripts/`):**

- `scripts/validate_math_correctness.py`
- `scripts/test_scientific_notation.py`
- `scripts/validate-alert-rules.sh`
- `scripts/validate-recording-rules.sh`
- `scripts/production-readiness-test.sh`

### Category 6: Test Configuration (FORK ONLY)

**Development Test Files (`.dev/`):**

- All test SLO YAML files
- `minikube-env.sh`
- Python analysis scripts
- JSON response files
- Generated SLO directories

### Category 7: Project Management (FORK ONLY)

**Kiro IDE Configuration (`.kiro/`):**

- `.kiro/steering/` - Development standards and guidelines
- `.kiro/specs/` - Feature specifications and task lists
- `.kiro/hooks/` - Agent hooks configuration

**AI Development Prompts (`prompts/`):**

- All session prompt files
- `README.md` - Session status tracking

### Category 8: Standard Project Files (REVIEW CAREFULLY)

**Version Control:**

- `.gitignore` - REVIEW: Ensure no dev-only patterns added
- `.gitattributes` - REVIEW: Check for any changes

**CI/CD:**

- `.github/` - REVIEW: Check for any workflow changes
- `.goreleaser.yml` - REVIEW: Check for any changes

**Docker:**

- `Dockerfile` - REVIEW: Check for any changes
- `Dockerfile.custom` - FORK ONLY (if custom)
- `Dockerfile.dev` - FORK ONLY (development)
- `.dockerignore` - REVIEW: Ensure no dev-only patterns added

**Documentation:**

- `CONTRIBUTING.md` - REVIEW: Revert any unintended changes
- `LICENSE` - NO CHANGES (should be identical to upstream)

**Kubernetes Manifests:**

- `examples/kubernetes/` - REVIEW: Revert any unintended changes
- `examples/openshift/` - REVIEW: Revert any unintended changes
- `examples/docker-compose/` - REVIEW: Revert any unintended changes

### Category 9: Existing Examples (REVIEW)

**Keep Existing Examples:**

- `examples/caddy-response-latency.yaml`
- `examples/nginx.yaml`
- `examples/parca-grpc-queryrange-errors.yaml`
- `examples/prometheus-http.yaml`
- `examples/prometheus-operator.yaml`
- `examples/prometheus-probe-success.yaml`
- `examples/pyrra-connect-errors.yaml`
- `examples/pyrra-connect-latency.yaml`
- `examples/pyrra-filesystem-errors.yaml`
- `examples/pyrra-kubernetes-errors.yaml`
- `examples/thanos-grpc.yaml`
- `examples/thanos-http.yaml`

**Move from .dev/ to examples/:**

- `.dev/test-dynamic-slo.yaml` → `examples/dynamic-burn-rate-ratio.yaml` (rename and simplify)
- `.dev/test-latency-dynamic.yaml` → `examples/dynamic-burn-rate-latency.yaml` (rename and simplify)

### Category 10: Files Requiring Investigation

**Filesystem Mode:**

- `filesystem.go` - INVESTIGATE: Check for dynamic burn rate changes
- `filesystem_test.go` - INVESTIGATE: Check for test changes

**Kubernetes Mode:**

- `kubernetes.go` - INVESTIGATE: Check for dynamic burn rate changes
- `kubernetes_test.go` - INVESTIGATE: Check for test changes

**Prometheus Integration:**

- `prometheus.go` - INVESTIGATE: Check for dynamic burn rate changes
- `prometheus_test.go` - INVESTIGATE: Check for test changes

**Main Entry Point:**

- `main.go` - INVESTIGATE: Check for dynamic burn rate changes
- `main_test.go` - INVESTIGATE: Check for test changes

**Logging:**

- `logger.go` - INVESTIGATE: Check for any changes
- `logger_test.go` - INVESTIGATE: Check for any changes

**Code Generation:**

- `generate.go` - INVESTIGATE: Check for any changes

## Action Plan

### Step 1: Create Preservation Branch

```bash
# Create branch to preserve all development artifacts
git checkout -b dev-tools-and-docs

# This branch contains everything - no deletions
git push origin dev-tools-and-docs
```

### Step 2: Prepare PR Branch

```bash
# Switch back to feature branch
git checkout <feature-branch-name>

# Remove fork-only directories
rm -rf cmd/
rm -rf .dev-docs/
rm -rf scripts/
rm -rf .dev/
rm -rf .kiro/
rm -rf prompts/

# Remove temporary build artifacts
rm -f *.exe

# Remove custom Docker files
rm -f Dockerfile.custom
rm -f Dockerfile.dev
```

### Step 3: Add New Examples

```bash
# Copy and simplify test configurations
cp .dev/test-dynamic-slo.yaml examples/dynamic-burn-rate-ratio.yaml
cp .dev/test-latency-dynamic.yaml examples/dynamic-burn-rate-latency.yaml

# Edit files to:
# - Remove development-specific comments
# - Add clear user-facing comments
# - Simplify to essential configuration
# - Ensure they work with standard Prometheus setup
```

### Step 4: Update Documentation

**README.md:**

- Add "Dynamic Burn Rate Alerting" section (2-3 paragraphs)
- Brief explanation of traffic-aware thresholds
- Minimal configuration example
- Link to examples/

**examples/README.md:**

- Add brief dynamic burn rate explanation (1-2 paragraphs)
- List new dynamic examples
- Link to detailed docs (if created)

**Optional: docs/DYNAMIC_BURN_RATE.md:**

- Only create if it keeps main docs clean
- Focus on practical usage
- Avoid implementation details

### Step 5: Review Standard Files

```bash
# Check for unintended changes
git diff upstream/main .gitignore
git diff upstream/main .dockerignore
git diff upstream/main CONTRIBUTING.md
git diff upstream/main examples/kubernetes/
git diff upstream/main examples/openshift/
git diff upstream/main examples/docker-compose/

# Revert any unintended changes
git checkout upstream/main -- <file-with-unintended-changes>
```

### Step 6: Investigate Core Files

For each file in Category 10:

1. Review changes with `git diff upstream/main <file>`
2. Determine if changes are:
   - Essential for dynamic burn rate feature → KEEP
   - Development/testing only → REVERT
   - Unintended → REVERT
3. Document decision and rationale

### Step 7: Final Cleanup

```bash
# Run tests to ensure everything still works
make test

# Build to ensure no compilation errors
make build

# Check git status
git status

# Review all changes
git diff upstream/main

# Commit organized changes
git add .
git commit -m "chore: organize files for upstream PR"
```

## Verification Checklist

Before proceeding to PR creation:

- [ ] All development tools preserved in `dev-tools-and-docs` branch
- [ ] All internal documentation preserved in `dev-tools-and-docs` branch
- [ ] Core implementation files present in PR branch
- [ ] Essential tests present in PR branch
- [ ] User-facing documentation updated (minimal, proportional)
- [ ] New example configurations added and simplified
- [ ] Standard project files reviewed for unintended changes
- [ ] Core files investigated and decisions documented
- [ ] All tests pass
- [ ] Build succeeds
- [ ] Feature still works correctly
- [ ] No temporary files or build artifacts in PR branch
- [ ] Git history is clean and logical

## Summary

**Files for PR:** ~50-100 files (core implementation, essential tests, minimal docs)
**Files for Fork:** ~200+ files (development tools, extensive docs, test configs)

**Key Principle:** The PR should be focused, clean, and easy to review. All the extensive development work is preserved in the fork for reference and future development.

## Next Steps

After completing this file organization:

1. Proceed to Task 8.4: Create PR description with test evidence
2. Proceed to Task 9: Final validation before submission
