#!/bin/bash

# Alert Rules Validation Script
# Validates alert rules generation for all indicator types

echo "=== Alert Rules Validation ==="
echo ""

# Test SLOs to validate
declare -a test_slos=(
    "test-dynamic-apiserver:ratio:dynamic"
    "test-latency-dynamic:latency:dynamic"
    "test-latency-native-dynamic:latencyNative:dynamic"
    "test-bool-gauge-dynamic:boolGauge:dynamic"
    "test-static-apiserver:ratio:static"
)

total_tests=0
passed_tests=0
failed_tests=0

for slo_info in "${test_slos[@]}"; do
    IFS=':' read -r slo_name indicator_type burn_rate_type <<< "$slo_info"
    
    echo "=== Validating SLO: $slo_name ($indicator_type, $burn_rate_type) ==="
    echo ""
    
    # Test 1: Check PrometheusRule exists
    ((total_tests++))
    echo "[Test 1] Checking PrometheusRule exists..."
    if kubectl get prometheusrule "$slo_name" -n monitoring &>/dev/null; then
        echo "✅ PASS: PrometheusRule exists"
        ((passed_tests++))
    else
        echo "❌ FAIL: PrometheusRule not found"
        ((failed_tests++))
        continue
    fi
    
    # Test 2: Check recording rules exist
    ((total_tests++))
    echo ""
    echo "[Test 2] Checking recording rules in PrometheusRule..."
    recording_rule_count=$(kubectl get prometheusrule "$slo_name" -n monitoring -o json | jq '[.spec.groups[].rules[] | select(.record != null)] | length')
    if [ "$recording_rule_count" -gt 0 ]; then
        echo "✅ PASS: Found $recording_rule_count recording rules"
        ((passed_tests++))
    else
        echo "❌ FAIL: No recording rules found"
        ((failed_tests++))
    fi
    
    # Test 3: Check alert rules exist
    ((total_tests++))
    echo ""
    echo "[Test 3] Checking alert rules in PrometheusRule..."
    alert_rule_count=$(kubectl get prometheusrule "$slo_name" -n monitoring -o json | jq '[.spec.groups[].rules[] | select(.alert != null)] | length')
    if [ "$alert_rule_count" -gt 0 ]; then
        echo "✅ PASS: Found $alert_rule_count alert rules"
        ((passed_tests++))
    else
        echo "❌ FAIL: No alert rules found"
        ((failed_tests++))
    fi
    
    # Test 4: For dynamic SLOs, verify alert expressions reference recording rules
    if [ "$burn_rate_type" = "dynamic" ]; then
        ((total_tests++))
        echo ""
        echo "[Test 4] Validating alert expressions reference recording rules..."
        
        # Extract alert expressions (excluding SLOMetricAbsent alerts) and check if they contain :burnrate
        alert_exprs=$(kubectl get prometheusrule "$slo_name" -n monitoring -o json | jq -r '.spec.groups[].rules[] | select(.alert != null and .alert != "SLOMetricAbsent") | .expr')
        
        all_use_recording_rules=true
        while IFS= read -r expr; do
            if [[ ! "$expr" =~ ":burnrate" ]]; then
                echo "❌ Alert expression doesn't reference recording rules: $expr"
                all_use_recording_rules=false
            fi
        done <<< "$alert_exprs"
        
        if [ "$all_use_recording_rules" = true ]; then
            echo "✅ PASS: All alert expressions reference recording rules"
            ((passed_tests++))
        else
            echo "❌ FAIL: Some alert expressions don't reference recording rules"
            ((failed_tests++))
        fi
        
        # Test 5: Validate dynamic threshold calculation structure
        ((total_tests++))
        echo ""
        echo "[Test 5] Validating dynamic threshold calculation structure..."
        
        all_have_scalar=true
        all_have_traffic_calc=true
        while IFS= read -r expr; do
            if [[ ! "$expr" =~ "scalar(" ]]; then
                echo "❌ Alert expression missing scalar() wrapper"
                all_have_scalar=false
            fi
            # Check for traffic calculation - can be increase() or count_over_time() depending on indicator type
            if [[ ! "$expr" =~ "increase(" ]] && [[ ! "$expr" =~ "count_over_time(" ]]; then
                echo "❌ Alert expression missing traffic calculation (increase or count_over_time)"
                all_have_traffic_calc=false
            fi
        done <<< "$alert_exprs"
        
        if [ "$all_have_scalar" = true ] && [ "$all_have_traffic_calc" = true ]; then
            echo "✅ PASS: All alert expressions have correct dynamic threshold structure"
            ((passed_tests++))
        else
            echo "❌ FAIL: Some alert expressions have incorrect structure"
            ((failed_tests++))
        fi
    fi
    
    echo ""
    echo "---"
    echo ""
done

# Summary
echo ""
echo "=== VALIDATION SUMMARY ==="
echo "Total Tests: $total_tests"
echo "✅ Passed: $passed_tests"
echo "❌ Failed: $failed_tests"
echo ""

if [ $failed_tests -eq 0 ]; then
    echo "🎉 All tests passed!"
    exit 0
else
    echo "⚠️  Some tests failed. Review the output above for details."
    exit 1
fi
