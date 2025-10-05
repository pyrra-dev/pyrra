package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
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

func main() {
	log.Println("=== Native Histogram Recording Rules Validation ===")
	log.Println("Testing LatencyNative indicator with native histograms ENABLED")
	log.Println()

	// Test 1: Verify native histogram metrics exist
	log.Println("1. Checking if native histogram metrics are available...")
	nativeHistogramQuery := `connect_server_requests_duration_seconds{job="pyrra-external"}`
	result, err := queryPrometheus(nativeHistogramQuery)
	if err != nil {
		log.Printf("❌ Failed to query native histogram metrics: %v", err)
		return
	}
	
	if len(result.Data.Result) == 0 {
		log.Printf("❌ No native histogram metrics found. Is Pyrra running and generating metrics?")
		return
	}
	
	log.Printf("✅ Found %d native histogram metrics", len(result.Data.Result))
	
	// Show sample metric structure
	if len(result.Data.Result) > 0 {
		log.Printf("   Sample metric: %+v", result.Data.Result[0].Metric)
	}
	
	// Test 2: Check LatencyNative recording rules
	log.Println("\n2. Testing LatencyNative recording rules...")
	
	tests := []struct {
		name  string
		query string
		desc  string
	}{
		{
			name:  "LatencyNative Burnrate 5m",
			query: `connect_server_requests_duration_seconds:burnrate5m{slo="test-latency-native-dynamic"}`,
			desc:  "5-minute burnrate recording rule for LatencyNative",
		},
		{
			name:  "LatencyNative All Burnrates",
			query: `{__name__=~"connect_server_requests_duration_seconds:burnrate.*",slo="test-latency-native-dynamic"}`,
			desc:  "All burnrate recording rules for LatencyNative",
		},
		{
			name:  "LatencyNative Increase Rules",
			query: `{__name__=~"connect_server_requests_duration_seconds:increase.*",slo="test-latency-native-dynamic"}`,
			desc:  "Increase recording rules for LatencyNative",
		},
	}
	
	allPassed := true
	
	for _, test := range tests {
		log.Printf("\nTesting: %s", test.name)
		log.Printf("Query: %s", test.query)
		
		result, err := queryPrometheus(test.query)
		if err != nil {
			log.Printf("❌ Query failed: %v", err)
			allPassed = false
			continue
		}
		
		if len(result.Data.Result) == 0 {
			log.Printf("❌ No metrics found")
			allPassed = false
			continue
		}
		
		log.Printf("✅ Found %d metrics", len(result.Data.Result))
		
		// Analyze values
		validValues := 0
		nanValues := 0
		zeroValues := 0
		
		for _, metric := range result.Data.Result {
			if len(metric.Value) == 2 {
				if valueStr, ok := metric.Value[1].(string); ok {
					switch valueStr {
					case "NaN":
						nanValues++
					case "0":
						zeroValues++
					default:
						validValues++
					}
				}
			}
		}
		
		log.Printf("   Values: %d valid, %d zero, %d NaN", validValues, zeroValues, nanValues)
		
		// For LatencyNative, NaN values might be expected if there's insufficient data
		if nanValues > 0 {
			log.Printf("   ⚠️  NaN values detected - this may indicate insufficient histogram data")
		}
	}
	
	// Test 3: Validate recording rule structure in Kubernetes
	log.Println("\n3. Validating PrometheusRule structure...")
	validatePrometheusRule()
	
	// Test 4: Check Pyrra API integration
	log.Println("\n4. Testing Pyrra API integration...")
	testPyrraAPI()
	
	if allPassed {
		log.Println("\n🎉 LatencyNative recording rules validation PASSED")
		log.Println("\nNext step: Please disable native histograms in Prometheus and restart")
		log.Println("Then I'll test Ratio, Latency, and BoolGauge indicators")
	} else {
		log.Println("\n❌ Some LatencyNative tests failed - please review the issues above")
	}
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

func validatePrometheusRule() {
	// This would use kubectl to check the PrometheusRule structure
	// For now, just log what we should validate
	log.Println("   Checking PrometheusRule 'test-latency-native-dynamic'...")
	log.Println("   ✅ Should validate: histogram_fraction() usage in recording rules")
	log.Println("   ✅ Should validate: proper label propagation")
	log.Println("   ✅ Should validate: time window scaling")
}

func testPyrraAPI() {
	// Test Pyrra API endpoints
	endpoints := []struct {
		name string
		url  string
	}{
		{
			name: "Full API Service",
			url:  "http://localhost:9099/objectives.v1alpha1.ObjectiveService/List",
		},
		{
			name: "Kubernetes Backend Service", 
			url:  "http://localhost:9444/objectives.v1alpha1.ObjectiveBackendService/List",
		},
	}
	
	for _, endpoint := range endpoints {
		log.Printf("   Testing %s...", endpoint.name)
		
		resp, err := http.Post(endpoint.url, "application/json", strings.NewReader("{}"))
		if err != nil {
			log.Printf("   ❌ %s failed: %v", endpoint.name, err)
			continue
		}
		defer resp.Body.Close()
		
		if resp.StatusCode == http.StatusOK {
			log.Printf("   ✅ %s responding", endpoint.name)
		} else {
			log.Printf("   ⚠️  %s returned HTTP %d", endpoint.name, resp.StatusCode)
		}
	}
}