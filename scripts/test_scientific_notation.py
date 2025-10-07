#!/usr/bin/env python3
"""
Test script to verify scientific notation formatting for very small threshold values.
This script calculates expected threshold values for high SLO targets (99.99%) 
and verifies they should be displayed in scientific notation.

IMPORTANT: Traffic ratio is N_SLO / N_alert (actual event count ratio).
For 30 day SLO window and 1 hour alert window: ratio ≈ 720 at steady traffic.
This is NOT a multiplier around 1.0!
"""

def calculate_threshold(slo_target, factor, traffic_ratio=None):
    """
    Calculate dynamic burn rate threshold.
    
    Formula: (N_SLO / N_alert) × E_budget_percent × (1 - SLO_target)
    
    traffic_ratio = N_SLO / N_alert (actual event count ratio)
    For 30 day SLO window and 1 hour alert window: ratio ≈ 720 at steady traffic
    """
    # Error budget percentage constants based on factor
    e_budget_percent = {
        14: 1/48,   # 0.020833
        7:  1/16,   # 0.0625
        2:  1/14,   # 0.071429
        1:  1/7     # 0.142857
    }
    
    # Window ratios for 30 day SLO (in hours)
    # These are the expected N_SLO / N_alert ratios at steady traffic
    window_ratios = {
        14: 720 / 1.067,    # 30d / 1h4m ≈ 675
        7:  720 / 6.433,    # 30d / 6h26m ≈ 112
        2:  720 / 25.717,   # 30d / 1d1h43m ≈ 28
        1:  720 / 102.85    # 30d / 4d6h51m ≈ 7
    }
    
    if traffic_ratio is None:
        traffic_ratio = window_ratios[factor]
    
    error_budget = 1 - slo_target
    threshold_constant = e_budget_percent[factor]
    threshold = traffic_ratio * threshold_constant * error_budget
    
    return threshold

def format_number(value):
    """Format number with scientific notation if < 0.001"""
    if abs(value) < 0.001 and value != 0:
        return f"{value:.3e}"
    return f"{value:.8f}"

def main():
    print("=" * 80)
    print("Scientific Notation Test - High SLO Target (99.99%)")
    print("=" * 80)
    print()
    
    slo_target = 0.9999
    error_budget = 1 - slo_target
    
    print(f"SLO Target: {slo_target * 100}%")
    print(f"Error Budget: {error_budget} ({error_budget * 100}%)")
    print()
    
    print("Expected Thresholds (at steady traffic with actual window ratios):")
    print("-" * 80)
    print(f"{'Factor':<10} {'E_Budget %':<15} {'N_SLO/N_alert':<15} {'Threshold':<20} {'Formatted':<20}")
    print("-" * 80)
    
    factors = [14, 7, 2, 1]
    e_budget_map = {
        14: (1/48, "2.08%"),
        7:  (1/16, "6.25%"),
        2:  (1/14, "7.14%"),
        1:  (1/7, "14.29%")
    }
    
    window_ratios = {
        14: 720 / 1.067,    # 30d / 1h4m ≈ 675
        7:  720 / 6.433,    # 30d / 6h26m ≈ 112
        2:  720 / 25.717,   # 30d / 1d1h43m ≈ 28
        1:  720 / 102.85    # 30d / 4d6h51m ≈ 7
    }
    
    for factor in factors:
        threshold = calculate_threshold(slo_target, factor)  # Uses default window ratios
        e_budget_val, e_budget_str = e_budget_map[factor]
        ratio = window_ratios[factor]
        formatted = format_number(threshold)
        
        # Check if scientific notation should be used
        uses_sci = abs(threshold) < 0.001
        notation = "✓ Scientific" if uses_sci else "  Fixed"
        
        print(f"{factor:<10} {e_budget_str:<15} {ratio:<15.1f} {threshold:<20.12f} {formatted:<20} {notation}")
    
    print()
    print("=" * 80)
    print("Test with Different Traffic Levels (Factor 14):")
    print("=" * 80)
    print()
    
    # Traffic variations relative to steady state
    # At steady traffic: N_SLO/N_alert ≈ 675 for factor 14
    # Low traffic (50%): ratio ≈ 337
    # High traffic (200%): ratio ≈ 1350
    steady_ratio = 720 / 1.067  # ≈ 675
    traffic_scenarios = [
        ("Low (50%)", steady_ratio * 0.5),
        ("Steady (100%)", steady_ratio),
        ("High (200%)", steady_ratio * 2.0),
        ("Very High (500%)", steady_ratio * 5.0)
    ]
    factor = 14  # Most sensitive alert
    
    print(f"Factor {factor} (2.08% error budget consumption, 30d SLO / 1h4m alert):")
    print("-" * 80)
    print(f"{'Traffic Level':<20} {'N_SLO/N_alert':<20} {'Threshold':<20} {'Formatted':<20}")
    print("-" * 80)
    
    for label, ratio in traffic_scenarios:
        threshold = calculate_threshold(slo_target, factor, traffic_ratio=ratio)
        formatted = format_number(threshold)
        uses_sci = abs(threshold) < 0.001
        notation = "✓ Scientific" if uses_sci else "  Fixed"
        
        print(f"{label:<20} {ratio:<20.1f} {threshold:<20.12f} {formatted:<20} {notation}")
    
    print()
    print("=" * 80)
    print("Comparison: Normal SLO Target (99.5%) vs High Target (99.99%)")
    print("=" * 80)
    print()
    
    normal_target = 0.995
    high_target = 0.9999
    factor = 14
    steady_ratio = 720 / 1.067  # ≈ 675
    
    normal_threshold = calculate_threshold(normal_target, factor, traffic_ratio=steady_ratio)
    high_threshold = calculate_threshold(high_target, factor, traffic_ratio=steady_ratio)
    
    print(f"Normal Target (99.5%):")
    print(f"  Error Budget: {1 - normal_target} ({(1 - normal_target) * 100}%)")
    print(f"  N_SLO/N_alert: {steady_ratio:.1f} (steady traffic)")
    print(f"  Threshold: {normal_threshold:.8f} ({format_number(normal_threshold)})")
    print()
    print(f"High Target (99.99%):")
    print(f"  Error Budget: {1 - high_target} ({(1 - high_target) * 100}%)")
    print(f"  N_SLO/N_alert: {steady_ratio:.1f} (steady traffic)")
    print(f"  Threshold: {high_threshold:.12f} ({format_number(high_threshold)})")
    print()
    print(f"Ratio: {normal_threshold / high_threshold:.1f}x larger for normal target")
    print()
    
    print("=" * 80)
    print("Key Insight:")
    print("=" * 80)
    print()
    print("Traffic ratio (N_SLO / N_alert) is the actual event count ratio:")
    print("  - At steady traffic for 30d/1h window: ratio ≈ 720")
    print("  - At low traffic (50%): ratio ≈ 360")
    print("  - At high traffic (200%): ratio ≈ 1440")
    print()
    print("This is NOT a multiplier around 1.0!")
    print("The ratio scales with actual traffic volume in the windows.")
    print()

if __name__ == "__main__":
    main()
