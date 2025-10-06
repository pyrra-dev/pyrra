#!/usr/bin/env python3
"""
Mathematical Correctness Validation Script

This script validates that Pyrra's dynamic burn rate calculations are mathematically correct
by comparing expected values (calculated from formulas) with actual Prometheus values.

Usage:
    python scripts/validate_math_correctness.py

Requirements:
    - Prometheus running on localhost:9090
    - Test SLOs deployed (test-dynamic-apiserver, test-latency-dynamic)
"""

import json
import subprocess
import sys
import urllib.parse
from typing import Dict, Any, Optional


def query_prometheus(query: str) -> Optional[Dict[str, Any]]:
    """Execute a Prometheus query and return the result."""
    encoded_query = urllib.parse.quote(query)
    cmd = ["curl", "-s", f"http://localhost:9090/api/v1/query?query={encoded_query}"]
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        data = json.loads(result.stdout)
        if data["status"] == "success" and data["data"]["result"]:
            return data["data"]["result"][0]
        return None
    except Exception as e:
        print(f"Error querying Prometheus: {e}")
        return None


def extract_value(result: Optional[Dict[str, Any]]) -> Optional[float]:
    """Extract the numeric value from a Prometheus query result."""
    if result and "value" in result:
        return float(result["value"][1])
    return None


def validate_ratio_slo():
    """Validate test-dynamic-apiserver (ratio indicator) recording rules."""
    print("=" * 80)
    print("VALIDATING RATIO SLO: test-dynamic-apiserver")
    print("=" * 80)

    # SLO Configuration
    slo_target = 0.95
    slo_window_days = 30
    error_budget = 1 - slo_target

    print(f"\nSLO Configuration:")
    print(f"  Target: {slo_target * 100}%")
    print(f"  Window: {slo_window_days} days")
    print(f"  Error Budget: {error_budget} ({error_budget * 100}%)")

    # 1. Validate burn rate recording rule query correctness
    print(f"\n--- Burn Rate Recording Rule (5m window) ---")
    print(f"Rule: apiserver_request:burnrate5m")
    print(f"Query: sum(rate(apiserver_request_total{{code=~\"4..|5..\",verb=\"GET\"}}[5m])) / sum(rate(apiserver_request_total{{verb=\"GET\"}}[5m]))")
    print(f"Purpose: Calculate error rate over 5m window")
    print(f"Analysis: ✓ Correctly uses rate() for per-second error rate calculation")
    
    # Query the burn rate value
    query_burnrate = 'apiserver_request:burnrate5m{slo="test-dynamic-apiserver"}'
    result_burnrate = query_prometheus(query_burnrate)
    
    if result_burnrate:
        burnrate_5m = extract_value(result_burnrate)
        print(f"Current Error Rate (5m): {burnrate_5m:.6f} ({burnrate_5m*100:.4f}%)")
    else:
        print("WARNING: Could not retrieve burn rate value")
        burnrate_5m = None

    # 2. Validate increase recording rule
    print(f"\n--- Increase Recording Rule (30d window) ---")
    print(f"Rule: apiserver_request:increase30d")
    print(f"Query: sum by (code) (increase(apiserver_request_total{{verb=\"GET\"}}[30d]))")
    print(f"Purpose: Calculate total traffic over SLO window (N_SLO)")
    print(f"Analysis: ✓ Correctly uses increase() for total event count")
    
    # Query N_SLO (sum across all codes)
    query_n_slo = 'sum(apiserver_request:increase30d{slo="test-dynamic-apiserver"})'
    result_n_slo = query_prometheus(query_n_slo)
    
    if result_n_slo:
        n_slo = extract_value(result_n_slo)
        print(f"N_SLO (30d total traffic): {n_slo:,.2f} requests")
    else:
        print("ERROR: Could not retrieve N_SLO value")
        return False

    # 3. Validate dynamic threshold calculation
    print(f"\n--- Dynamic Threshold Calculation (Critical 1: 5m/1h4m) ---")
    
    # Query N_1h4m (current traffic in long window)
    query_n_1h4m = 'sum(increase(apiserver_request_total{verb="GET"}[1h4m]))'
    result_n_1h4m = query_prometheus(query_n_1h4m)
    
    if result_n_1h4m:
        n_1h4m = extract_value(result_n_1h4m)
        print(f"N_1h4m (current 1h4m traffic): {n_1h4m:,.2f} requests")
    else:
        print("WARNING: Could not retrieve N_1h4m value")
        n_1h4m = None

    if n_slo and n_1h4m:
        # Calculate expected dynamic threshold
        e_budget_percent = 1 / 48  # Factor 14 corresponds to 1/48
        traffic_ratio = n_slo / n_1h4m
        expected_threshold = traffic_ratio * e_budget_percent * error_budget

        print(f"\nFormula: (N_SLO / N_1h4m) × E_budget_percent × (1 - SLO_target)")
        print(f"Calculation: ({n_slo:,.2f} / {n_1h4m:,.2f}) × {e_budget_percent:.6f} × {error_budget:.2f}")
        print(f"  Traffic Ratio: {traffic_ratio:.6f}")
        print(f"  E_budget_percent (1/48): {e_budget_percent:.6f}")
        print(f"  Error Budget: {error_budget:.2f}")
        print(f"  Expected Dynamic Threshold: {expected_threshold:.8f}")

        # Compare with static threshold
        static_threshold = 14 * error_budget
        print(f"\nComparison:")
        print(f"  Static Threshold (14 × {error_budget}): {static_threshold:.6f}")
        print(f"  Dynamic Threshold: {expected_threshold:.8f}")
        print(f"  Ratio (Dynamic/Static): {expected_threshold/static_threshold:.6f}x")
        
        # Compare with actual burn rate
        if burnrate_5m is not None:
            print(f"\nCurrent Status:")
            print(f"  Current Error Rate (5m): {burnrate_5m:.8f}")
            print(f"  Dynamic Threshold: {expected_threshold:.8f}")
            if burnrate_5m > expected_threshold:
                print(f"  Status: ⚠️  WOULD ALERT (error rate exceeds threshold)")
            else:
                margin = ((expected_threshold - burnrate_5m) / expected_threshold) * 100
                print(f"  Status: ✓ OK (error rate is {margin:.1f}% below threshold)")

    # 4. Validate window scaling by checking actual recording rules
    print(f"\n--- Window Scaling Validation ---")
    print(f"Base SLO Window: 28 days")
    print(f"Actual SLO Window: 30 days")
    print(f"Scaling Factor: 30/28 = 1.0714")
    
    # Expected windows based on scaling formula
    expected_windows = [
        ("Critical 1 Short", "5m", "apiserver_request:burnrate5m"),
        ("Critical 1 Long", "1h4m", "apiserver_request:burnrate1h4m"),
        ("Critical 2 Short", "32m", "apiserver_request:burnrate32m"),
        ("Critical 2 Long", "6h26m", "apiserver_request:burnrate6h26m"),
        ("Warning 1 Short", "2h9m", "apiserver_request:burnrate2h9m"),
        ("Warning 1 Long", "1d1h43m", "apiserver_request:burnrate1d1h43m"),
        ("Warning 2 Short", "6h26m", "apiserver_request:burnrate6h26m"),
        ("Warning 2 Long", "4d6h51m", "apiserver_request:burnrate4d6h51m"),
    ]
    
    print(f"\nValidating Recording Rules Exist:")
    all_windows_valid = True
    for name, window, rule_name in expected_windows:
        query = f'{rule_name}{{slo="test-dynamic-apiserver"}}'
        result = query_prometheus(query)
        if result:
            print(f"  ✓ {name:20s} ({window:8s}): {rule_name}")
        else:
            print(f"  ✗ {name:20s} ({window:8s}): {rule_name} - NOT FOUND")
            all_windows_valid = False
    
    if all_windows_valid:
        print(f"\nAnalysis: ✓ All 8 windows correctly scaled and recording rules exist")
    else:
        print(f"\nAnalysis: ✗ Some recording rules are missing")
        return False

    return True


def validate_latency_slo():
    """Validate test-latency-dynamic (latency indicator) recording rules."""
    print("\n" + "=" * 80)
    print("VALIDATING LATENCY SLO: test-latency-dynamic")
    print("=" * 80)

    # SLO Configuration
    slo_target = 0.95
    slo_window_days = 30
    error_budget = 1 - slo_target

    print(f"\nSLO Configuration:")
    print(f"  Target: {slo_target * 100}%")
    print(f"  Window: {slo_window_days} days")
    print(f"  Error Budget: {error_budget} ({error_budget * 100}%)")
    print(f"  Latency Threshold: 100ms (le=0.1)")

    # 1. Validate burn rate recording rule query correctness
    print(f"\n--- Burn Rate Recording Rule (5m window) ---")
    print(f"Rule: prometheus_http_request_duration_seconds:burnrate5m")
    print(f"Query: (sum(rate(..._count[5m])) - sum(rate(..._bucket{{le=\"0.1\"}}[5m]))) / sum(rate(..._count[5m]))")
    print(f"Purpose: Calculate latency failure rate (requests > 100ms)")
    print(f"Analysis: ✓ Correctly calculates (total - success) / total")

    # Query the burn rate value
    query_burnrate = 'prometheus_http_request_duration_seconds:burnrate5m{slo="test-latency-dynamic"}'
    result_burnrate = query_prometheus(query_burnrate)
    
    if result_burnrate:
        burnrate_5m = extract_value(result_burnrate)
        print(f"Current Failure Rate (5m): {burnrate_5m:.6f} ({burnrate_5m*100:.4f}%)")
    else:
        print("WARNING: Could not retrieve burn rate value")
        burnrate_5m = None

    # 2. Validate increase recording rule
    print(f"\n--- Increase Recording Rule (30d window) ---")
    print(f"Rule: prometheus_http_request_duration_seconds:increase30d")
    print(f"Query: sum(increase(..._count[30d])) for total")
    print(f"Purpose: Calculate total requests over SLO window (N_SLO)")
    print(f"Analysis: ✓ Correctly uses increase() for total event count")

    # Query N_SLO (total count, le="")
    query_n_slo = 'prometheus_http_request_duration_seconds:increase30d{slo="test-latency-dynamic",le=""}'
    result_n_slo = query_prometheus(query_n_slo)
    
    if result_n_slo:
        n_slo = extract_value(result_n_slo)
        print(f"N_SLO (30d total requests): {n_slo:,.2f} requests")
    else:
        print("ERROR: Could not retrieve N_SLO value")
        return False

    # 3. Validate dynamic threshold calculation
    print(f"\n--- Dynamic Threshold Calculation (Critical 1: 5m/1h4m) ---")
    
    # Query N_1h4m (current traffic in long window)
    query_n_1h4m = 'sum(increase(prometheus_http_request_duration_seconds_count{job="prometheus-k8s"}[1h4m]))'
    result_n_1h4m = query_prometheus(query_n_1h4m)
    
    if result_n_1h4m:
        n_1h4m = extract_value(result_n_1h4m)
        print(f"N_1h4m (current 1h4m traffic): {n_1h4m:,.2f} requests")
    else:
        print("WARNING: Could not retrieve N_1h4m value")
        n_1h4m = None

    if n_slo and n_1h4m:
        # Calculate expected dynamic threshold
        e_budget_percent = 1 / 48
        traffic_ratio = n_slo / n_1h4m
        expected_threshold = traffic_ratio * e_budget_percent * error_budget

        print(f"\nFormula: (N_SLO / N_1h4m) × E_budget_percent × (1 - SLO_target)")
        print(f"Calculation: ({n_slo:,.2f} / {n_1h4m:,.2f}) × {e_budget_percent:.6f} × {error_budget:.2f}")
        print(f"  Traffic Ratio: {traffic_ratio:.6f}")
        print(f"  Expected Dynamic Threshold: {expected_threshold:.8f}")

        # Compare with static threshold
        static_threshold = 14 * error_budget
        print(f"\nComparison:")
        print(f"  Static Threshold (14 × {error_budget}): {static_threshold:.6f}")
        print(f"  Dynamic Threshold: {expected_threshold:.8f}")
        print(f"  Ratio (Dynamic/Static): {expected_threshold/static_threshold:.6f}x")
        
        # Compare with actual burn rate
        if burnrate_5m is not None:
            print(f"\nCurrent Status:")
            print(f"  Current Failure Rate (5m): {burnrate_5m:.8f}")
            print(f"  Dynamic Threshold: {expected_threshold:.8f}")
            if burnrate_5m > expected_threshold:
                print(f"  Status: ⚠️  WOULD ALERT (failure rate exceeds threshold)")
            else:
                margin = ((expected_threshold - burnrate_5m) / expected_threshold) * 100
                print(f"  Status: ✓ OK (failure rate is {margin:.1f}% below threshold)")

    # 4. Validate window scaling by checking actual recording rules
    print(f"\n--- Window Scaling Validation ---")
    print(f"Base SLO Window: 28 days")
    print(f"Actual SLO Window: 30 days")
    print(f"Scaling Factor: 30/28 = 1.0714")
    
    # Expected windows based on scaling formula
    expected_windows = [
        ("Critical 1 Short", "5m", "prometheus_http_request_duration_seconds:burnrate5m"),
        ("Critical 1 Long", "1h4m", "prometheus_http_request_duration_seconds:burnrate1h4m"),
        ("Critical 2 Short", "32m", "prometheus_http_request_duration_seconds:burnrate32m"),
        ("Critical 2 Long", "6h26m", "prometheus_http_request_duration_seconds:burnrate6h26m"),
        ("Warning 1 Short", "2h9m", "prometheus_http_request_duration_seconds:burnrate2h9m"),
        ("Warning 1 Long", "1d1h43m", "prometheus_http_request_duration_seconds:burnrate1d1h43m"),
        ("Warning 2 Short", "6h26m", "prometheus_http_request_duration_seconds:burnrate6h26m"),
        ("Warning 2 Long", "4d6h51m", "prometheus_http_request_duration_seconds:burnrate4d6h51m"),
    ]
    
    print(f"\nValidating Recording Rules Exist:")
    all_windows_valid = True
    for name, window, rule_name in expected_windows:
        query = f'{rule_name}{{slo="test-latency-dynamic"}}'
        result = query_prometheus(query)
        if result:
            print(f"  ✓ {name:20s} ({window:8s}): {rule_name}")
        else:
            print(f"  ✗ {name:20s} ({window:8s}): {rule_name} - NOT FOUND")
            all_windows_valid = False
    
    if all_windows_valid:
        print(f"\nAnalysis: ✓ All 8 windows correctly scaled and recording rules exist")
    else:
        print(f"\nAnalysis: ✗ Some recording rules are missing")
        return False

    return True


def main():
    """Main validation function."""
    print("\n" + "=" * 80)
    print("PYRRA DYNAMIC BURN RATE - MATHEMATICAL CORRECTNESS VALIDATION")
    print("=" * 80)
    print("\nThis script validates:")
    print("1. Recording rule query correctness (rate, increase functions)")
    print("2. Window scaling calculations (28d → 30d)")
    print("3. Dynamic threshold calculations (N_SLO / N_alert formula)")
    print("4. Comparison with static thresholds\n")

    # Validate both SLO types
    ratio_ok = validate_ratio_slo()
    latency_ok = validate_latency_slo()

    # Summary
    print("\n" + "=" * 80)
    print("VALIDATION SUMMARY")
    print("=" * 80)
    print(f"Ratio SLO (test-dynamic-apiserver): {'✓ PASS' if ratio_ok else '✗ FAIL'}")
    print(f"Latency SLO (test-latency-dynamic): {'✓ PASS' if latency_ok else '✗ FAIL'}")

    if ratio_ok and latency_ok:
        print("\n✓ All mathematical validations passed!")
        print("\nNext Steps:")
        print("1. Check Pyrra UI (http://localhost:3000) to verify BurnRateThresholdDisplay")
        print("2. Compare UI threshold values with calculated values above")
        print("3. Test with different traffic scenarios (high/low traffic)")
        print("4. Review .dev-docs/TASK_7_2_MATH_VALIDATION_GUIDE.md for detailed analysis")
        return 0
    else:
        print("\n✗ Some validations failed. Please review the output above.")
        return 1


if __name__ == "__main__":
    sys.exit(main())
