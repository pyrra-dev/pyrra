# Continue Dynamic Burn Rate Implementation - Session Prompt

I'm continuing work on Pyrra's dynamic burn rate feature for traffic-aware SLO alerting. 

## Current Status
Context: #file:FEATURE_IMPLEMENTATION_SUMMARY.md

## Key Completed Work
- ✅ Ratio Indicators: Full dynamic burn rate support with optimized recording rules
- ✅ Latency Indicators: Full dynamic burn rate support with optimized recording rules  
- ✅ Performance Optimization: Both use pre-computed recording rules + dynamic threshold calculation
- ✅ Comprehensive Testing: Full test coverage for both indicator types

## Next Priority Tasks
1. **⚠️ CRITICAL FIRST**: Fix compilation error in main.go ("main redeclared in this block" at line 100)
2. **LatencyNative Indicators**: Extend buildDynamicAlertExpr() in slo/rules.go for native histograms
3. **BoolGauge Indicators**: Extend buildDynamicAlertExpr() in slo/rules.go for boolean gauges
4. Add corresponding test cases following existing TestObjective_DynamicBurnRate_* pattern
5. Validate main.go integration with dynamic burn rate features

## Implementation Pattern to Follow
Current working approach (from Ratio/Latency):
```go
case IndicatorType:
    return fmt.Sprintf(
        "(%s{%s} > ((sum(increase(%s{%s}[%s])) / sum(increase(%s{%s}[%s]))) * %f * (1-%s))) and "+
        "(%s{%s} > ((sum(increase(%s{%s}[%s])) / sum(increase(%s{%s}[%s]))) * %f * (1-%s)))",
        // Uses recording rules + dynamic threshold calculation
        o.BurnrateName(w.Short), alertMatchersString, /* dynamic threshold params */,
        o.BurnrateName(w.Long), alertMatchersString, /* dynamic threshold params */,
    )
```

## Key Files to Modify
- **⚠️ main.go** - Fix "main redeclared" compilation error at line 100 FIRST
- main_test.go - Review tests for any integration issues with dynamic burn rate
- slo/rules.go - buildDynamicAlertExpr() method (add LatencyNative & BoolGauge cases)  
- slo/rules_test.go - Add TestObjective_DynamicBurnRate_LatencyNative() and TestObjective_DynamicBurnRate_BoolGauge()

## Testing Commands
```bash
# ⚠️ CRITICAL: Fix compilation first - this will currently fail
go build .

# After fixing main.go, run dynamic burn rate tests
go test ./slo -v -run "TestObjective_DynamicBurnRate"
go test ./slo

# Validate main application integration
go test . -v -run "TestMatrixToValues|TestAlertsMatchingObjectives"
```

## Repository
- Branch: add-dynamic-burn-rate
- Owner: yairst/pyrra
- All current work committed and pushed ✅
- **⚠️ BLOCKER**: Compilation error in main.go needs immediate attention

**IMPORTANT**: Fix the "main redeclared" error in main.go before continuing with remaining indicator types. This compilation issue needs to be resolved first to ensure the application builds correctly.

Pick up where we left off extending dynamic burn rate support to the remaining indicator types.
