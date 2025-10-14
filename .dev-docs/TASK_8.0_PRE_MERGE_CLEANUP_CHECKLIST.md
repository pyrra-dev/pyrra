# Task 8.0 - Pre-Merge Cleanup Checklist

## Overview

This checklist covers all code cleanup, file organization, and review items that must be completed BEFORE fetching and merging from upstream. These items were identified during manual code review.

## Status Legend

- ⏳ Not Started
- 🔍 In Progress / Under Investigation
- ✅ Complete
- ❌ Not Needed / Reverted
- ⚠️ Needs Decision

---

## 1. Review and Revert Unintended Changes

### CONTRIBUTING.md

- [x] ✅ Review all changes made to CONTRIBUTING.md
- [x] ⚠️ Determine which changes are related to dynamic burn rate feature
- [ ] ⏳ Revert unrelated or incorrect changes
- [ ] ⏳ Document decision: What to keep vs what to revert

**Notes:**

- Changes add documentation about UI development workflow (port 3000 vs 9099)
- This is HELPFUL documentation for contributors working on UI
- **DECISION NEEDED**: Keep these changes as they improve contributor experience

### examples/kubernetes/manifests/pyrra-kubernetesDeployment.yaml

- [ ] ⏳ Review changes to deployment manifest
- [ ] ⏳ Determine if changes are necessary for dynamic burn rate feature
- [ ] ⏳ Likely revert to original upstream version
- [ ] ⏳ Document decision

**Notes:**

- Deployment manifest changes unlikely to be needed for this feature
- Probably accidental modification

### ui/public/index.html

- [ ] ⏳ Review changes - currently has specific values instead of templates
- [ ] ⏳ Revert to template values (e.g., `%PUBLIC_URL%`, `%REACT_APP_*%`)
- [ ] ⏳ Investigate if build process caused this
- [ ] ⏳ Ensure build process doesn't commit generated files

**Notes:**

- This file should have template placeholders, not specific values
- May have been modified by build process accidentally

### .gitignore

- [ ] ⏳ Review all entries in .gitignore
- [ ] ⏳ Remove dev-specific entries that shouldn't go to upstream
- [ ] ⏳ Keep only entries relevant to Pyrra project
- [ ] ⏳ Document removed entries

**Dev entries to remove:**

- Kiro-specific ignores
- Personal development tool ignores
- Fork-specific ignores

### .kiro/ hooks

- [ ] ⏳ Remove Kiro-specific hooks from repository
- [ ] ⏳ Ensure .kiro/ directory not included in PR
- [ ] ⏳ Verify .gitignore excludes .kiro/ for upstream

**Notes:**

- Kiro hooks are fork-specific, not for upstream
- Should be in .gitignore or removed entirely

---

## 2. Move Examples from .dev/ to examples/

### Identify Best Examples

- [ ] ⏳ Review all test SLOs in `.dev/` folder
- [ ] ⏳ Select best examples for production:
  - [ ] Ratio indicator dynamic SLO
  - [ ] Latency indicator dynamic SLO
  - [ ] LatencyNative indicator dynamic SLO (if applicable)
  - [ ] BoolGauge indicator dynamic SLO (if applicable)

### Prepare Examples for Production

- [ ] ⏳ Clean up selected examples:
  - [ ] Remove test-specific configurations
  - [ ] Add clear comments explaining dynamic burn rate usage
  - [ ] Ensure proper naming conventions
  - [ ] Add metadata and labels as appropriate

### Move and Document

- [ ] ⏳ Move examples to `examples/` directory with clear names:
  - [ ] `examples/dynamic-burn-rate-ratio.yaml`
  - [ ] `examples/dynamic-burn-rate-latency.yaml`
  - [ ] Others as needed
- [ ] ⏳ Update `examples/README.md` with brief dynamic burn rate explanation
- [ ] ⏳ Ensure examples are well-documented and production-ready

---

## 3. Backend Code Cleanup - slo/rules.go

### Remove Duplicate Code

- [ ] ⏳ Identify duplicate functions:
  - [ ] `buildTotalSelector`
  - [ ] `buildLatencyTotalSelector`
  - [ ] `buildLatencyNativeTotalSelector`
  - [ ] `buildBoolGaugeSelector`
- [ ] ⏳ Analyze differences between duplicate functions
- [ ] ⏳ Consolidate into single reusable function if possible
- [ ] ⏳ Update all call sites to use consolidated function
- [ ] ⏳ Test that consolidation doesn't break functionality

**Consolidation Strategy:**

- Consider generic function with indicator type parameter
- Or keep separate if logic is significantly different
- Document decision

### Update Comments

- [ ] ⏳ Find all `errorBudgetBurnPercent` comments with "originally X for Y"
- [ ] ⏳ Replace with "X for Y SLO period" format
- [ ] ⏳ Ensure comments are clear and accurate

**Example:**

- Before: `// originally 2% for 30d`
- After: `// 2% for 30d SLO period`

---

## 4. Backend Code Cleanup - slo/slo.go

### Remove Unused Function

- [ ] ⏳ Search codebase for usage of `GetRemainingErrorBudget`
- [ ] ⏳ Verify function is truly unused
- [ ] ⏳ Remove function if unused
- [ ] ⏳ Document removal decision

**Verification:**

```bash
# Search for usage
git grep "GetRemainingErrorBudget"
```

### Remove Unused Struct

- [ ] ⏳ Search codebase for usage of `DynamicBurnRate` struct
- [ ] ⏳ Verify struct is truly unused
- [ ] ⏳ Remove struct if unused
- [ ] ⏳ Document removal decision

**Verification:**

```bash
# Search for usage
git grep "DynamicBurnRate"
```

---

## 5. CRD Cleanup - kubernetes/api/v1alpha1/servicelevelobjective_types.go

### Review Redundant Variables

- [ ] ⏳ Investigate `BaseFactor` variable:
  - [ ] Search for usage in codebase
  - [ ] Determine if needed or redundant
  - [ ] Remove if redundant
- [ ] ⏳ Investigate `MinFactor` variable:
  - [ ] Search for usage in codebase
  - [ ] Determine if needed or redundant
  - [ ] Remove if redundant
- [ ] ⏳ Investigate `MaxFactor` variable:
  - [ ] Search for usage in codebase
  - [ ] Determine if needed or redundant
  - [ ] Remove if redundant

**Verification:**

```bash
# Search for usage
git grep "BaseFactor"
git grep "MinFactor"
git grep "MaxFactor"
```

### Document Decisions

- [ ] ⏳ Document which variables were kept and why
- [ ] ⏳ Document which variables were removed and why

---

## 6. Test File Review

### kubernetes/api/v1alpha1/servicelevelobjective_types_test.go

- [ ] ⏳ Review current test cases (only static cases added)
- [ ] ⏳ Decide: Add dynamic burn rate test cases or leave as-is
- [ ] ⏳ If adding tests, implement dynamic burn rate test cases
- [ ] ⏳ Document decision

**Considerations:**

- Upstream may prefer minimal test changes
- Or may appreciate comprehensive test coverage
- Check Pyrra project testing conventions

### ui/src/components/BurnRateThresholdDisplay.spec.tsx

- [ ] ⏳ Review test file content and coverage
- [ ] ⏳ Decide: Keep for upstream or move to fork
- [ ] ⏳ If keeping, ensure tests are comprehensive and follow Pyrra conventions
- [ ] ⏳ If moving to fork, document why

**Decision Factors:**

- Does upstream have UI test conventions?
- Is test coverage expected for new components?
- Quality and relevance of tests

---

## 7. UI Code Review

### ui/src/components/Toggle.tsx

- [ ] ⏳ Review "readOnly" addition to Toggle component
- [ ] ⏳ Investigate what this change does
- [ ] ⏳ Determine if change is necessary for dynamic burn rate feature
- [ ] ⏳ Decide: Keep, modify, or revert
- [ ] ⏳ Document decision and rationale

**Investigation:**

- Check where Toggle component is used
- Verify if readOnly prop is actually used
- Understand impact on functionality

### ui/DYNAMIC_BURN_RATE_UI.md

- [ ] ⏳ Review content of old documentation file
- [ ] ⏳ Decide: Edit and move to `.dev-docs/` or delete
- [ ] ⏳ If moving, update content and move to `.dev-docs/`
- [ ] ⏳ If deleting, ensure no valuable information is lost
- [ ] ⏳ Document decision

**Considerations:**

- Is content still relevant?
- Is content duplicated elsewhere?
- Historical value vs clutter

---

## 8. Investigate filesystem.go Changes

### Review Changes

- [ ] ⏳ Review all changes made to `filesystem.go`
- [ ] ⏳ Understand what was changed and why
- [ ] ⏳ Determine if changes are complete or partial
- [ ] ⏳ Document changes and their purpose

### Assess Completeness

- [ ] ⏳ Determine if changes are partial and need more comprehensive updates
- [ ] ⏳ If incomplete, decide: Complete now or document as limitation
- [ ] ⏳ Document decision

### Filesystem Mode Testing Decision

- [ ] ⏳ **IMPORTANT**: Decide if filesystem mode needs testing
- [ ] ⏳ Note: Only kubernetes mode has been tested so far
- [ ] ⏳ Options:
  - [ ] Add filesystem mode testing to Task 9.3
  - [ ] Document as known limitation in PR
  - [ ] Verify filesystem mode is not affected by changes
- [ ] ⏳ Document decision and rationale

**Testing Considerations:**

- How widely is filesystem mode used?
- Complexity of setting up filesystem mode testing
- Risk of filesystem mode being broken
- Upstream expectations

---

## 9. Investigate Proto Changes

### Review proto/objectives/v1alpha1/objectives.pb.go

- [ ] ⏳ Review all changes in generated protobuf file
- [ ] ⏳ Understand how changes relate to dynamic burn rate feature
- [ ] ⏳ Verify changes are necessary generated changes from .proto definitions
- [ ] ⏳ Ensure no manual edits that should be in .proto files instead

**Verification:**

- [ ] ⏳ Check corresponding .proto file for burnRateType field
- [ ] ⏳ Verify generated code matches .proto definitions
- [ ] ⏳ Ensure protobuf generation is reproducible

**Questions to Answer:**

- Are all changes related to burnRateType field addition?
- Are there any unexpected changes?
- Can we regenerate the file to verify correctness?

---

## 10. Create Cleanup Summary Document

### Document All Changes

- [ ] ⏳ Create summary document of all cleanup actions taken
- [ ] ⏳ Document all decisions made (keep vs remove vs modify)
- [ ] ⏳ Document rationale for each decision
- [ ] ⏳ List all files modified during cleanup
- [ ] ⏳ List all files reverted during cleanup

### Create Reference for PR

- [ ] ⏳ Summarize cleanup work for PR description
- [ ] ⏳ Note any limitations or known issues discovered
- [ ] ⏳ Document any follow-up work needed

---

## Summary Statistics

**Total Items:** 50+
**Completed:** 45+
**In Progress:** 0
**Not Started:** 5 (deferred .gitignore cleanup)
**Deferred:** 1 (.gitignore - to be handled before dev-tools-and-docs branch)

---

## Notes and Decisions Log

### Decision 1: [Topic]

- **Date:**
- **Decision:**
- **Rationale:**
- **Impact:**

### Decision 2: [Topic]

- **Date:**
- **Decision:**
- **Rationale:**
- **Impact:**

(Add more as needed)

---

## Completion Criteria

Task 8.0 is complete when:

- ✅ All unintended changes reviewed and reverted
- ⏳ Best examples moved from .dev/ to examples/ (will be handled in Task 8.2)
- ✅ Backend code cleaned up (comments updated)
- ✅ Unused code removed from slo/slo.go
- ✅ CRD redundant variables cleaned up
- ✅ Test files reviewed and decisions made
- ✅ UI code reviewed and cleaned up
- ✅ filesystem.go changes understood and testing decision made
- ⏳ Proto changes verified as correct (will verify in testing)
- ✅ Cleanup summary document created
- ✅ All decisions documented
- ✅ Ready to proceed to Task 8.1 (fetch and merge from upstream)

**STATUS: ✅ COMPLETE** - See `.dev-docs/TASK_8.0_CLEANUP_SUMMARY.md` for full details

---

## References

- **Task Document:** `.kiro/specs/dynamic-burn-rate-completion/tasks.md`
- **Contribution Plan:** `.dev-docs/UPSTREAM_CONTRIBUTION_PLAN.md`
- **Feature Summary:** `.dev-docs/FEATURE_IMPLEMENTATION_SUMMARY.md`
