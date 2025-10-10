package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type PrometheusQueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func main() {
	fmt.Println("=== BurnRateThresholdDisplay Query Validation ===")
	fmt.Println()

	prometheusURL := "http://localhost:9090"

	// Test current implementation queries (raw metrics)
	// CORRECTED: Only test SLO window (30d in these test cases)
	// Alert windows do NOT have increase/count recording rules
	// Run multiple times and average for accurate performance measurement
	const numRuns = 10

	fmt.Println("1. Testing CURRENT implementation (raw metrics - SLO window):")
	fmt.Printf("   Running %d iterations per query...\n", numRuns)
	fmt.Println()

	currentQueries := []struct {
		name  string
		query string
	}{
		{
			name:  "Ratio - 30d increase",
			query: "sum(increase(apiserver_request_total[30d]))",
		},
		{
			name:  "Latency - 30d increase",
			query: "sum(increase(prometheus_http_request_duration_seconds_count[30d]))",
		},
		{
			name:  "BoolGauge - 30d count",
			query: "sum(count_over_time(up{job=\"prometheus-k8s\"}[30d]))",
		},
	}

	for _, test := range currentQueries {
		fmt.Printf("  %s\n", test.name)

		var durations []time.Duration
		var resultCount int
		var sampleValue string

		for i := 0; i < numRuns; i++ {
			result, duration, err := executeQuery(prometheusURL, test.query)
			if err != nil {
				continue
			}
			durations = append(durations, duration)
			resultCount = len(result.Data.Result)
			if len(result.Data.Result) > 0 && sampleValue == "" {
				sampleValue = fmt.Sprintf("%v", result.Data.Result[0].Value[1])
			}
		}

		if len(durations) == 0 {
			fmt.Printf("    ❌ All queries failed\n")
		} else {
			avgDuration := calculateAverage(durations)
			minDuration := calculateMin(durations)
			maxDuration := calculateMax(durations)

			fmt.Printf("    ✅ Avg Duration: %v (min: %v, max: %v)\n", avgDuration, minDuration, maxDuration)
			fmt.Printf("    📊 Results: %d series\n", resultCount)
			if sampleValue != "" {
				fmt.Printf("    📈 Sample Value: %v\n", sampleValue)
			}
			fmt.Printf("    🔄 Successful Runs: %d/%d\n", len(durations), numRuns)
		}
		fmt.Println()
	}

	// Test optimized implementation queries (recording rules)
	fmt.Println("2. Testing OPTIMIZED implementation (recording rules - SLO window):")
	fmt.Printf("   Running %d iterations per query...\n", numRuns)
	fmt.Println()

	optimizedQueries := []struct {
		name  string
		query string
	}{
		{
			name:  "Ratio - 30d increase",
			query: "sum(apiserver_request:increase30d{slo=\"test-dynamic-apiserver\"})",
		},
		{
			name:  "Latency - 30d increase",
			query: "sum(prometheus_http_request_duration_seconds:increase30d{slo=\"test-latency-dynamic\"})",
		},
		{
			name:  "BoolGauge - 30d count",
			query: "sum(up:count30d{slo=\"test-bool-gauge-dynamic\"})",
		},
	}

	for _, test := range optimizedQueries {
		fmt.Printf("  %s\n", test.name)

		var durations []time.Duration
		var resultCount int
		var sampleValue string

		for i := 0; i < numRuns; i++ {
			result, duration, err := executeQuery(prometheusURL, test.query)
			if err != nil {
				continue
			}
			durations = append(durations, duration)
			resultCount = len(result.Data.Result)
			if len(result.Data.Result) > 0 && sampleValue == "" {
				sampleValue = fmt.Sprintf("%v", result.Data.Result[0].Value[1])
			}
		}

		if len(durations) == 0 {
			fmt.Printf("    ❌ All queries failed\n")
		} else {
			avgDuration := calculateAverage(durations)
			minDuration := calculateMin(durations)
			maxDuration := calculateMax(durations)

			fmt.Printf("    ✅ Avg Duration: %v (min: %v, max: %v)\n", avgDuration, minDuration, maxDuration)
			fmt.Printf("    📊 Results: %d series\n", resultCount)
			if sampleValue != "" {
				fmt.Printf("    📈 Sample Value: %v\n", sampleValue)
			}
			fmt.Printf("    🔄 Successful Runs: %d/%d\n", len(durations), numRuns)
		}
		fmt.Println()
	}

	// Check recording rule availability
	fmt.Println("3. Recording Rule Availability Analysis:")
	fmt.Println()

	recordingRules := []struct {
		name  string
		query string
	}{
		{name: "Ratio - 30d window", query: "apiserver_request:increase30d{slo=\"test-dynamic-apiserver\"}"},
		{name: "Latency - 30d window", query: "prometheus_http_request_duration_seconds:increase30d{slo=\"test-latency-dynamic\"}"},
		{name: "BoolGauge - 30d window", query: "up:count30d{slo=\"test-bool-gauge-dynamic\"}"},
	}

	for _, test := range recordingRules {
		result, _, err := executeQuery(prometheusURL, test.query)
		if err != nil {
			fmt.Printf("  ❌ %s: Not found\n", test.name)
		} else if len(result.Data.Result) == 0 {
			fmt.Printf("  ⚠️  %s: Exists but no data\n", test.name)
		} else {
			fmt.Printf("  ✅ %s: %d series\n", test.name, len(result.Data.Result))
		}
	}
	fmt.Println()

	fmt.Println("  IMPORTANT: Only SLO window has increase/count recording rules!")
	fmt.Println("  • Alert windows (1h4m, 6h26m, etc.) do NOT have increase/count recording rules")
	fmt.Println("  • Optimization applies to SLO window calculation only")
	fmt.Println("  • Alert windows must use inline increase()/count_over_time() calculations")
	fmt.Println()

	// Performance comparison
	fmt.Println("4. Performance Comparison Summary:")
	fmt.Println()
	fmt.Println("  Current Implementation (Raw Metrics):")
	fmt.Println("    • Calculates increase()/count_over_time() for both SLO and alert windows")
	fmt.Println("    • SLO window calculation is expensive (scans long time range)")
	fmt.Println()
	fmt.Println("  Optimized Implementation (Hybrid Approach):")
	fmt.Println("    • Use recording rule for SLO window (pre-computed)")
	fmt.Println("    • Use inline increase()/count_over_time() for alert windows (no recording rules)")
	fmt.Println("    • Query pattern: sum(metric:increase30d) / sum(increase(metric[1h4m]))")
	fmt.Println("    • For bool-gauge: sum(metric:count30d) / sum(count_over_time(metric[1h4m]))")
	fmt.Println()
	fmt.Println("  Expected Benefits:")
	fmt.Println("    • Faster SLO window calculation")
	fmt.Println("    • Reduced Prometheus load for long window queries")
	fmt.Println("    • Better UI responsiveness")
	fmt.Println()
	fmt.Println("  Recommendation: ✅ Implement hybrid optimization approach")
	fmt.Println()
}

func executeQuery(prometheusURL, query string) (PrometheusQueryResult, time.Duration, error) {
	start := time.Now()

	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d",
		prometheusURL,
		url.QueryEscape(query),
		time.Now().Unix(),
	)

	resp, err := http.Get(queryURL)
	if err != nil {
		return PrometheusQueryResult{}, 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return PrometheusQueryResult{}, duration, fmt.Errorf("failed to read response: %w", err)
	}

	var result PrometheusQueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return PrometheusQueryResult{}, duration, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if result.Status != "success" {
		return PrometheusQueryResult{}, duration, fmt.Errorf("query failed: %s", string(body))
	}

	return result, duration, nil
}

func calculateAverage(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total / time.Duration(len(durations))
}

func calculateMin(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	min := durations[0]
	for _, d := range durations {
		if d < min {
			min = d
		}
	}
	return min
}

func calculateMax(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	max := durations[0]
	for _, d := range durations {
		if d > max {
			max = d
		}
	}
	return max
}
