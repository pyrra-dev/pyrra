package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

// PrometheusQueryResult represents the structure of Prometheus query responses
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

// RecordingRuleValidation represents a recording rule validation test
type RecordingRuleValidation struct {
	SLOName       string
	IndicatorType string
	BurnRateType  string
	RuleName      string
	ExpectedQuery string
	Description   string
}

// ValidationResult represents the result of a validation test
type ValidationResult struct {
	Test        RecordingRuleValidation
	Success     bool
	Error       string
	ActualQuery string
	MetricCount int
	QueryTime   time.Duration
}

func main() {
	log.Println("=== Recording Rules Validation for All Indicator Types ===")
	
	// Define test cases for all indicator types and burn rate types
	testCases := []RecordingRuleValidation{
		// Ratio indicators
		{
			SLOName:       "test-dynamic-apiserver",
			IndicatorType: "ratio",
			BurnRateType:  "dynamic",
			RuleName:      "apiserver_request:burnrate5m",
			Description:   "Ratio indicator with dynamic burn rate - 5m burnrate rule",
		},
		{
			SLOName:       "test-static-apiserver", 
			IndicatorType: "ratio",
			BurnRateType:  "static",
			RuleName:      "apiserver_request:burnrate5m",
			Description:   "Ratio indicator with static burn rate - 5m burnrate rule",
		},
		// Latency indicators
		{
			SLOName:       "test-latency-dynamic",
			IndicatorType: "latency",
			BurnRateType:  "dynamic", 
			RuleName:      "prometheus_http_request_duration_seconds:burnrate5m",
			Description:   "Latency indicator with dynamic burn rate - 5m burnrate rule",
		},
		{
			SLOName:       "test-latency-static",
			IndicatorType: "latency",
			BurnRateType:  "static",
			RuleName:      "prometheus_http_request_duration_seconds:burnrate5m", 
			Description:   "Latency indicator with static burn rate - 5m burnrate rule",
		},
		// LatencyNative indicators
		{
			SLOName:       "test-latency-native-dynamic",
			IndicatorType: "latencyNative",
			BurnRateType:  "dynamic",
			RuleName:      "connect_server_requests_duration_seconds:burnrate5m",
			Description:   "LatencyNative indicator with dynamic burn rate - 5m burnrate rule",
		},
		// BoolGauge indicators  
		{
			SLOName:       "test-bool-gauge-dynamic",
			IndicatorType: "boolGauge",
			BurnRateType:  "dynamic",
			RuleName:      "up:burnrate5m",
			Description:   "BoolGauge indicator with dynamic burn rate - 5m burnrate rule",
		},
	}

	// Run validation tests
	results := make([]ValidationResult, 0, len(testCases))
	
	for _, test := range testCases {
		log.Printf("Testing: %s (%s %s)", test.Description, test.IndicatorType, test.BurnRateType)
		result := validateRecordingRule(test)
		results = append(results, result)
		
		if result.Success {
			log.Printf("✅ PASS: %s - Found %d metrics in %v", test.RuleName, result.MetricCount, result.QueryTime)
		} else {
			log.Printf("❌ FAIL: %s - %s", test.RuleName, result.Error)
		}
	}
	
	// Print summary
	printValidationSummary(results)
	
	// Additional validation tests
	log.Println("\n=== Additional Validation Tests ===")
	validateQueryEfficiency()
	validateLabelHandling()
	validateTimeWindows()
}

func validateRecordingRule(test RecordingRuleValidation) ValidationResult {
	result := ValidationResult{
		Test:    test,
		Success: false,
	}
	
	// Query Prometheus for the recording rule
	query := fmt.Sprintf(`%s{slo="%s"}`, test.RuleName, test.SLOName)
	
	startTime := time.Now()
	prometheusResult, err := queryPrometheus(query)
	result.QueryTime = time.Since(startTime)
	
	if err != nil {
		result.Error = fmt.Sprintf("Failed to query Prometheus: %v", err)
		return result
	}
	
	if prometheusResult.Status != "success" {
		result.Error = fmt.Sprintf("Prometheus query failed: %s", prometheusResult.Status)
		return result
	}
	
	result.MetricCount = len(prometheusResult.Data.Result)
	
	if result.MetricCount == 0 {
		result.Error = "No metrics found for recording rule"
		return result
	}
	
	// Validate that the recording rule has proper labels
	for _, metric := range prometheusResult.Data.Result {
		// Check for required labels
		if sloLabel, exists := metric.Metric["slo"]; !exists || sloLabel != test.SLOName {
			result.Error = fmt.Sprintf("Missing or incorrect 'slo' label: expected %s, got %s", test.SLOName, sloLabel)
			return result
		}
		
		// Validate metric value is numeric and reasonable
		if len(metric.Value) != 2 {
			result.Error = "Invalid metric value format"
			return result
		}
		
		// Check that the value is a valid number (not NaN or Inf)
		valueStr, ok := metric.Value[1].(string)
		if !ok {
			result.Error = "Metric value is not a string"
			return result
		}
		
		if valueStr == "NaN" || valueStr == "+Inf" || valueStr == "-Inf" {
			result.Error = fmt.Sprintf("Invalid metric value: %s", valueStr)
			return result
		}
	}
	
	result.Success = true
	return result
}

func queryPrometheus(query string) (*PrometheusQueryResult, error) {
	prometheusURL := "http://localhost:9090"
	
	// URL encode the query
	encodedQuery := url.QueryEscape(query)
	fullURL := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, encodedQuery)
	
	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response: %v", err)
	}
	
	var result PrometheusQueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("Failed to parse JSON: %v", err)
	}
	
	return &result, nil
}

func printValidationSummary(results []ValidationResult) {
	log.Println("\n=== Validation Summary ===")
	
	passed := 0
	failed := 0
	
	// Group results by indicator type
	byIndicator := make(map[string][]ValidationResult)
	for _, result := range results {
		indicator := result.Test.IndicatorType
		byIndicator[indicator] = append(byIndicator[indicator], result)
	}
	
	// Sort indicator types for consistent output
	indicators := make([]string, 0, len(byIndicator))
	for indicator := range byIndicator {
		indicators = append(indicators, indicator)
	}
	sort.Strings(indicators)
	
	for _, indicator := range indicators {
		log.Printf("\n%s Indicators:", strings.Title(indicator))
		for _, result := range byIndicator[indicator] {
			status := "❌ FAIL"
			if result.Success {
				status = "✅ PASS"
				passed++
			} else {
				failed++
			}
			
			log.Printf("  %s %s (%s): %s", 
				status, 
				result.Test.BurnRateType,
				result.Test.SLOName,
				result.Test.RuleName)
			
			if !result.Success {
				log.Printf("    Error: %s", result.Error)
			} else {
				log.Printf("    Metrics: %d, Query Time: %v", result.MetricCount, result.QueryTime)
			}
		}
	}
	
	log.Printf("\n=== Final Results ===")
	log.Printf("Total Tests: %d", len(results))
	log.Printf("Passed: %d", passed)
	log.Printf("Failed: %d", failed)
	
	if failed > 0 {
		log.Printf("❌ VALIDATION FAILED - %d tests failed", failed)
		os.Exit(1)
	} else {
		log.Printf("✅ ALL TESTS PASSED")
	}
}

func validateQueryEfficiency() {
	log.Println("Testing query efficiency...")
	
	// Test that recording rules use efficient aggregations
	testQueries := []struct {
		name  string
		query string
		maxTime time.Duration
	}{
		{
			name:  "Burnrate recording rule query",
			query: `apiserver_request:burnrate5m{slo="test-dynamic-apiserver"}`,
			maxTime: 100 * time.Millisecond,
		},
		{
			name:  "Availability recording rule query", 
			query: `pyrra_availability{slo="test-dynamic-apiserver"}`,
			maxTime: 100 * time.Millisecond,
		},
		{
			name:  "Request rate recording rule query",
			query: `pyrra_requests:rate5m{slo="test-dynamic-apiserver"}`,
			maxTime: 100 * time.Millisecond,
		},
	}
	
	for _, test := range testQueries {
		startTime := time.Now()
		result, err := queryPrometheus(test.query)
		queryTime := time.Since(startTime)
		
		if err != nil {
			log.Printf("❌ %s: Query failed - %v", test.name, err)
			continue
		}
		
		if queryTime > test.maxTime {
			log.Printf("⚠️  %s: Query took %v (expected < %v)", test.name, queryTime, test.maxTime)
		} else {
			log.Printf("✅ %s: Query completed in %v with %d results", test.name, queryTime, len(result.Data.Result))
		}
	}
}

func validateLabelHandling() {
	log.Println("Testing label handling...")
	
	// Test that recording rules have proper label propagation
	query := `{__name__=~".*:burnrate.*",slo="test-dynamic-apiserver"}`
	
	result, err := queryPrometheus(query)
	if err != nil {
		log.Printf("❌ Label validation query failed: %v", err)
		return
	}
	
	if len(result.Data.Result) == 0 {
		log.Printf("❌ No burnrate metrics found for label validation")
		return
	}
	
	// Check that all burnrate metrics have consistent labels
	expectedLabels := []string{"slo"}
	labelCounts := make(map[string]int)
	
	for _, metric := range result.Data.Result {
		for label := range metric.Metric {
			labelCounts[label]++
		}
	}
	
	log.Printf("Label distribution across %d burnrate metrics:", len(result.Data.Result))
	for label, count := range labelCounts {
		log.Printf("  %s: %d metrics", label, count)
	}
	
	// Verify required labels are present
	for _, requiredLabel := range expectedLabels {
		if count, exists := labelCounts[requiredLabel]; !exists {
			log.Printf("❌ Required label '%s' missing from recording rules", requiredLabel)
		} else if count != len(result.Data.Result) {
			log.Printf("⚠️  Label '%s' not present on all metrics (%d/%d)", requiredLabel, count, len(result.Data.Result))
		} else {
			log.Printf("✅ Label '%s' present on all recording rules", requiredLabel)
		}
	}
}

func validateTimeWindows() {
	log.Println("Testing time window scaling...")
	
	// Test different SLO window sizes and verify recording rules scale appropriately
	testCases := []struct {
		sloName string
		window  string
		expectedRules []string
	}{
		{
			sloName: "test-dynamic-apiserver",
			window:  "30d",
			expectedRules: []string{
				"apiserver_request:burnrate5m",
				"apiserver_request:burnrate30m", 
				"apiserver_request:burnrate2h",
				"apiserver_request:burnrate6h26m",
			},
		},
	}
	
	for _, test := range testCases {
		log.Printf("Testing SLO %s with %s window:", test.sloName, test.window)
		
		for _, ruleName := range test.expectedRules {
			query := fmt.Sprintf(`%s{slo="%s"}`, ruleName, test.sloName)
			result, err := queryPrometheus(query)
			
			if err != nil {
				log.Printf("❌ %s: Query failed - %v", ruleName, err)
				continue
			}
			
			if len(result.Data.Result) == 0 {
				log.Printf("❌ %s: No metrics found", ruleName)
			} else {
				log.Printf("✅ %s: Found %d metrics", ruleName, len(result.Data.Result))
			}
		}
	}
}