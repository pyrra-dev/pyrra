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
	testQueries := []struct {
		name          string
		query         string
		usesRecording bool
		queryType     string
	}{
		{
			name:          "Dynamic Ratio Traffic Query (Raw Metrics)",
			query:         `sum(increase(apiserver_request_total[30d])) / sum(increase(apiserver_request_total[1h4m]))`,
			usesRecording: false,
			queryType:     "ratio",
		},
		{
			name:          "Dynamic Ratio Traffic Query (With Recording Rules)",
			query:         `sum(apiserver_request:increase30d{slo="test-dynamic-apiserver"}) / sum(apiserver_request:increase1h4m{slo="test-dynamic-apiserver"})`,
			usesRecording: true,
			queryType:     "ratio",
		},
		{
			name:          "Dynamic Latency Traffic Query (Raw Metrics)",
			query:         `sum(increase(prometheus_http_request_duration_seconds_count[30d])) / sum(increase(prometheus_http_request_duration_seconds_count[1h4m]))`,
			usesRecording: false,
			queryType:     "latency",
		},
		{
			name:          "Dynamic Latency Traffic Query (With Recording Rules)",
			query:         `sum(prometheus_http_request_duration_seconds:increase30d{slo="test-latency-dynamic"}) / sum(prometheus_http_request_duration_seconds:increase1h4m{slo="test-latency-dynamic"})`,
			usesRecording: true,
			queryType:     "latency",
		},
		{
			name:          "Dynamic LatencyNative Traffic Query (Raw Metrics)",
			query:         `sum(histogram_count(increase(http_request_duration_seconds[30d]))) / sum(histogram_count(increase(http_request_duration_seconds[1h4m])))`,
			usesRecording: false,
			queryType:     "latencyNative",
		},
		{
			name:          "Dynamic BoolGauge Traffic Query (Raw Metrics)",
			query:         `sum(count_over_time(probe_success[30d])) / sum(count_over_time(probe_success[1h4m]))`,
			usesRecording: false,
			queryType:     "boolGauge",
		},
	}

	var results []QueryPerformanceMetrics

	for _, test := range testQueries {
		fmt.Printf("Testing: %s\n", test.name)
		fmt.Printf("Query: %s\n", test.query)

		metrics, err := executeQuery(prometheusURL, test.query)
		if err != nil {
			fmt.Printf("  ❌ Error: %v\n", err)
			fmt.Println()
			continue
		}

		metrics.UsesRecording = test.usesRecording
		metrics.QueryType = test.queryType
		results = append(results, metrics)

		fmt.Printf("  ✅ Execution Time: %v\n", metrics.ExecutionTime)
		fmt.Printf("  📊 Result Count: %d series\n", metrics.ResultCount)
		fmt.Printf("  🔧 Uses Recording Rules: %v\n", metrics.UsesRecording)
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
		fmt.Printf("  Raw Metrics Query: %v\n", ratioRaw.ExecutionTime)
		fmt.Printf("  Recording Rules Query: %v\n", ratioRecording.ExecutionTime)
		speedup := float64(ratioRaw.ExecutionTime) / float64(ratioRecording.ExecutionTime)
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
		fmt.Printf("  Raw Metrics Query: %v\n", latencyRaw.ExecutionTime)
		fmt.Printf("  Recording Rules Query: %v\n", latencyRecording.ExecutionTime)
		speedup := float64(latencyRaw.ExecutionTime) / float64(latencyRecording.ExecutionTime)
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

	// Recommendations
	fmt.Println("=== Optimization Recommendations ===")
	fmt.Println()

	fmt.Println("Current BurnRateThresholdDisplay Implementation:")
	fmt.Println("  ❌ Uses raw metrics directly (e.g., apiserver_request_total)")
	fmt.Println("  ❌ Calculates increase() inline for 30d and alert windows")
	fmt.Println("  ❌ No use of pre-computed recording rules")
	fmt.Println()

	fmt.Println("Recommended Optimization:")
	fmt.Println("  ✅ Use increase recording rules (e.g., apiserver_request:increase30d)")
	fmt.Println("  ✅ Leverage pre-computed aggregations from recording rules")
	fmt.Println("  ✅ Reduce query complexity and execution time")
	fmt.Println()

	fmt.Println("Implementation Strategy:")
	fmt.Println("  1. Check if recording rules exist for the SLO")
	fmt.Println("  2. If available, use recording rules instead of raw metrics")
	fmt.Println("  3. Fallback to raw metrics if recording rules are not available")
	fmt.Println("  4. Add performance monitoring to track query execution times")
	fmt.Println()

	fmt.Println("Expected Benefits:")
	fmt.Println("  • Faster query execution (1.5x - 3x speedup)")
	fmt.Println("  • Reduced Prometheus load")
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
