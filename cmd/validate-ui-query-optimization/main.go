package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// PrometheusQueryResult represents a Prometheus query response
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

// QueryPerformanceMetrics tracks query execution performance
type QueryPerformanceMetrics struct {
	Query         string
	ExecutionTime time.Duration
	ResultCount   int
	UsesRecording bool
	QueryType     string
	Runs          int
	MinTime       time.Duration
	MaxTime       time.Duration
	AvgTime       time.Duration
}

func main() {
	fmt.Println("=== UI Query Optimization Validation ===")
	fmt.Println()

	prometheusURL := "http://localhost:9090"
	if len(os.Args) > 1 {
		prometheusURL = os.Args[1]
	}

	fmt.Printf("Prometheus URL: %s\n", prometheusURL)
	fmt.Println()

	// Test queries from BurnRateThresholdDisplay component
	// CORRECTED: Only test SLO window comparisons (30d in these test cases)
	// Alert windows (1h4m, 6h26m) do NOT have increase/count recording rules
	testQueries := []struct {
		name          string
		query         string
		usesRecording bool
		queryType     string
	}{
		{
			name:          "Ratio - 30d increase (Raw Metrics)",
			query:         `sum(increase(apiserver_request_total[30d]))`,
			usesRecording: false,
			queryType:     "ratio",
		},
		{
			name:          "Ratio - 30d increase (Recording Rule)",
			query:         `sum(apiserver_request:increase30d{slo="test-dynamic-apiserver"})`,
			usesRecording: true,
			queryType:     "ratio",
		},
		{
			name:          "Latency - 30d increase (Raw Metrics)",
			query:         `sum(increase(prometheus_http_request_duration_seconds_count[30d]))`,
			usesRecording: false,
			queryType:     "latency",
		},
		{
			name:          "Latency - 30d increase (Recording Rule)",
			query:         `sum(prometheus_http_request_duration_seconds:increase30d{slo="test-latency-dynamic"})`,
			usesRecording: true,
			queryType:     "latency",
		},
		{
			name:          "BoolGauge - 30d count (Raw Metrics)",
			query:         `sum(count_over_time(up{job="prometheus-k8s"}[30d]))`,
			usesRecording: false,
			queryType:     "boolGauge",
		},
		{
			name:          "BoolGauge - 30d count (Recording Rule)",
			query:         `sum(up:count30d{slo="test-bool-gauge-dynamic"})`,
			usesRecording: true,
			queryType:     "boolGauge",
		},
	}

	const numRuns = 10
	var results []QueryPerformanceMetrics

	for _, test := range testQueries {
		fmt.Printf("Testing: %s\n", test.name)
		fmt.Printf("Query: %s\n", test.query)
		fmt.Printf("Running %d iterations...\n", numRuns)

		var durations []time.Duration
		var resultCount int
		var lastError error

		for i := 0; i < numRuns; i++ {
			metrics, err := executeQuery(prometheusURL, test.query)
			if err != nil {
				lastError = err
				continue
			}
			durations = append(durations, metrics.ExecutionTime)
			resultCount = metrics.ResultCount
		}

		if len(durations) == 0 {
			fmt.Printf("  ❌ Error: %v\n", lastError)
			fmt.Println()
			continue
		}

		// Calculate statistics
		var total time.Duration
		minTime := durations[0]
		maxTime := durations[0]

		for _, d := range durations {
			total += d
			if d < minTime {
				minTime = d
			}
			if d > maxTime {
				maxTime = d
			}
		}

		avgTime := total / time.Duration(len(durations))

		metrics := QueryPerformanceMetrics{
			Query:         test.query,
			ExecutionTime: avgTime,
			ResultCount:   resultCount,
			UsesRecording: test.usesRecording,
			QueryType:     test.queryType,
			Runs:          len(durations),
			MinTime:       minTime,
			MaxTime:       maxTime,
			AvgTime:       avgTime,
		}

		results = append(results, metrics)

		fmt.Printf("  ✅ Avg Time: %v (min: %v, max: %v)\n", avgTime, minTime, maxTime)
		fmt.Printf("  📊 Result Count: %d series\n", resultCount)
		fmt.Printf("  🔧 Uses Recording Rules: %v\n", test.usesRecording)
		fmt.Printf("  🔄 Successful Runs: %d/%d\n", len(durations), numRuns)
		fmt.Println()
	}

	// Performance comparison analysis
	fmt.Println("=== Performance Comparison Analysis ===")
	fmt.Println()

	// Compare ratio queries
	ratioRaw := findMetrics(results, "ratio", false)
	ratioRecording := findMetrics(results, "ratio", true)
	if ratioRaw != nil && ratioRecording != nil {
		fmt.Println("Ratio Indicator Comparison:")
		fmt.Printf("  Raw Metrics Query:\n")
		fmt.Printf("    Avg: %v (min: %v, max: %v)\n", ratioRaw.AvgTime, ratioRaw.MinTime, ratioRaw.MaxTime)
		fmt.Printf("  Recording Rules Query:\n")
		fmt.Printf("    Avg: %v (min: %v, max: %v)\n", ratioRecording.AvgTime, ratioRecording.MinTime, ratioRecording.MaxTime)
		speedup := float64(ratioRaw.AvgTime) / float64(ratioRecording.AvgTime)
		fmt.Printf("  Speedup: %.2fx\n", speedup)
		if speedup > 1.5 {
			fmt.Println("  ✅ Recording rules provide significant performance improvement")
		} else if speedup > 1.0 {
			fmt.Println("  ⚠️  Recording rules provide modest performance improvement")
		} else {
			fmt.Println("  ❌ Recording rules do not improve performance")
		}
		fmt.Println()
	}

	// Compare latency queries
	latencyRaw := findMetrics(results, "latency", false)
	latencyRecording := findMetrics(results, "latency", true)
	if latencyRaw != nil && latencyRecording != nil {
		fmt.Println("Latency Indicator Comparison:")
		fmt.Printf("  Raw Metrics Query:\n")
		fmt.Printf("    Avg: %v (min: %v, max: %v)\n", latencyRaw.AvgTime, latencyRaw.MinTime, latencyRaw.MaxTime)
		fmt.Printf("  Recording Rules Query:\n")
		fmt.Printf("    Avg: %v (min: %v, max: %v)\n", latencyRecording.AvgTime, latencyRecording.MinTime, latencyRecording.MaxTime)
		speedup := float64(latencyRaw.AvgTime) / float64(latencyRecording.AvgTime)
		fmt.Printf("  Speedup: %.2fx\n", speedup)
		if speedup > 1.5 {
			fmt.Println("  ✅ Recording rules provide significant performance improvement")
		} else if speedup > 1.0 {
			fmt.Println("  ⚠️  Recording rules provide modest performance improvement")
		} else {
			fmt.Println("  ❌ Recording rules do not improve performance")
		}
		fmt.Println()
	}

	// Compare boolGauge queries
	boolGaugeRaw := findMetrics(results, "boolGauge", false)
	boolGaugeRecording := findMetrics(results, "boolGauge", true)
	if boolGaugeRaw != nil && boolGaugeRecording != nil {
		fmt.Println("BoolGauge Indicator Comparison:")
		fmt.Printf("  Raw Metrics Query:\n")
		fmt.Printf("    Avg: %v (min: %v, max: %v)\n", boolGaugeRaw.AvgTime, boolGaugeRaw.MinTime, boolGaugeRaw.MaxTime)
		fmt.Printf("  Recording Rules Query:\n")
		fmt.Printf("    Avg: %v (min: %v, max: %v)\n", boolGaugeRecording.AvgTime, boolGaugeRecording.MinTime, boolGaugeRecording.MaxTime)
		speedup := float64(boolGaugeRaw.AvgTime) / float64(boolGaugeRecording.AvgTime)
		fmt.Printf("  Speedup: %.2fx\n", speedup)
		if speedup > 1.5 {
			fmt.Println("  ✅ Recording rules provide significant performance improvement")
		} else if speedup > 1.0 {
			fmt.Println("  ⚠️  Recording rules provide modest performance improvement")
		} else {
			fmt.Println("  ❌ Recording rules do not improve performance")
		}
		fmt.Println()
	}

	// Recording rule availability check
	fmt.Println("=== Recording Rule Availability ===")
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
		metrics, err := executeQuery(prometheusURL, test.query)
		if err != nil {
			fmt.Printf("  ❌ %s: Not found\n", test.name)
		} else if metrics.ResultCount == 0 {
			fmt.Printf("  ⚠️  %s: Exists but no data\n", test.name)
		} else {
			fmt.Printf("  ✅ %s: %d series\n", test.name, metrics.ResultCount)
		}
	}
	fmt.Println()

	// Recommendations
	fmt.Println("=== Optimization Recommendations ===")
	fmt.Println()

	fmt.Println("Current BurnRateThresholdDisplay Implementation:")
	fmt.Println("  ❌ Uses raw metrics directly (e.g., apiserver_request_total)")
	fmt.Println("  ❌ Calculates increase()/count_over_time() inline for SLO window")
	fmt.Println("  ❌ No use of pre-computed recording rules")
	fmt.Println()

	fmt.Println("Recommended Optimization:")
	fmt.Println("  ✅ Use increase/count recording rules for SLO window")
	fmt.Println("  ✅ Keep inline increase()/count_over_time() for alert windows (no recording rules)")
	fmt.Println("  ✅ Query pattern: sum(metric:increase30d) / sum(increase(metric[1h4m]))")
	fmt.Println("  ✅ For bool-gauge: sum(metric:count30d) / sum(count_over_time(metric[1h4m]))")
	fmt.Println()

	fmt.Println("Implementation Strategy:")
	fmt.Println("  1. Use recording rule for SLO window (pre-computed)")
	fmt.Println("  2. Use inline increase()/count_over_time() for alert windows (no recording rules)")
	fmt.Println("  3. Fallback to raw metrics if recording rules unavailable")
	fmt.Println("  4. Add performance monitoring to track query execution times")
	fmt.Println()

	fmt.Println("Expected Benefits:")
	fmt.Println("  • Faster SLO window calculation (recording rule)")
	fmt.Println("  • Reduced Prometheus load for long window queries")
	fmt.Println("  • Better UI responsiveness")
	fmt.Println("  • Consistent with Pyrra's recording rule architecture")
	fmt.Println()
}

func executeQuery(prometheusURL, query string) (QueryPerformanceMetrics, error) {
	start := time.Now()

	// Build query URL
	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d",
		prometheusURL,
		url.QueryEscape(query),
		time.Now().Unix(),
	)

	// Execute query
	resp, err := http.Get(queryURL)
	if err != nil {
		return QueryPerformanceMetrics{}, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	executionTime := time.Since(start)

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return QueryPerformanceMetrics{}, fmt.Errorf("failed to read response: %w", err)
	}

	var result PrometheusQueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return QueryPerformanceMetrics{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if result.Status != "success" {
		return QueryPerformanceMetrics{}, fmt.Errorf("query failed: %s", string(body))
	}

	return QueryPerformanceMetrics{
		Query:         query,
		ExecutionTime: executionTime,
		ResultCount:   len(result.Data.Result),
	}, nil
}

func findMetrics(results []QueryPerformanceMetrics, queryType string, usesRecording bool) *QueryPerformanceMetrics {
	for _, r := range results {
		if r.QueryType == queryType && r.UsesRecording == usesRecording {
			return &r
		}
	}
	return nil
}
