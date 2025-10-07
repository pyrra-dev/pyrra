package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// PrometheusResponse represents the Prometheus API response
type PrometheusResponse struct {
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
	prometheusURL := "http://localhost:9090"

	// Test queries to check aggregation
	testQueries := []struct {
		name           string
		query          string
		expectedSeries int
		description    string
	}{
		{
			name:           "Raw metric (should return many series)",
			query:          `apiserver_request_total{verb="GET"}`,
			expectedSeries: 74, // Expected: 74 series
			description:    "Raw apiserver_request_total metric without aggregation",
		},
		{
			name:           "Sum aggregation (should return 1 series)",
			query:          `sum(apiserver_request_total{verb="GET"})`,
			expectedSeries: 1, // Expected: 1 series
			description:    "Sum aggregation without grouping - should return single series",
		},
		{
			name:           "Increase with sum (should return 1 series)",
			query:          `sum(increase(apiserver_request_total{verb="GET"}[30d]))`,
			expectedSeries: 1, // Expected: 1 series
			description:    "Sum of increase over 30d - should return single series",
		},
		{
			name:           "Recording rule increase30d (SHOULD return multiple series for UI grouping)",
			query:          `apiserver_request:increase30d{slo="test-dynamic-apiserver"}`,
			expectedSeries: 4, // Expected: 4 series (grouped by code for RequestsGraph)
			description:    "Recording rule for SLO window - INTENTIONALLY returns multiple series for UI RequestsGraph grouping",
		},
		{
			name:           "Burn rate recording rule (should return 1 series)",
			query:          `apiserver_request:burnrate5m{slo="test-dynamic-apiserver"}`,
			expectedSeries: 1, // Expected: 1 series
			description:    "Burn rate recording rule - should return single series",
		},
		{
			name:           "UI traffic query (should return 1 series)",
			query:          `sum(increase(apiserver_request_total{verb="GET"}[30d])) / sum(increase(apiserver_request_total{verb="GET"}[1h4m]))`,
			expectedSeries: 1, // Expected: 1 series (scalar result)
			description:    "UI traffic ratio query - should return single scalar value",
		},
		{
			name:           "Alert rule query (should use sum for single series)",
			query:          `apiserver_request:burnrate5m{slo="test-dynamic-apiserver"} > scalar((sum(increase(apiserver_request_total{verb="GET"}[30d])) / sum(increase(apiserver_request_total{verb="GET"}[1h4m]))) * 0.020833 * (1-0.95))`,
			expectedSeries: 1, // Expected: 1 series
			description:    "Alert rule expression - should return single series for alert evaluation",
		},
	}

	fmt.Println("=== Query Aggregation Test ===")
	fmt.Println("Testing queries to verify they return single series results\n")

	allPassed := true

	for i, test := range testQueries {
		fmt.Printf("%d. %s\n", i+1, test.name)
		fmt.Printf("   Query: %s\n", test.query)
		fmt.Printf("   Description: %s\n", test.description)

		seriesCount, resultType, err := executeQuery(prometheusURL, test.query)
		if err != nil {
			fmt.Printf("   ❌ ERROR: %v\n\n", err)
			allPassed = false
			continue
		}

		fmt.Printf("   Result Type: %s\n", resultType)
		fmt.Printf("   Series Count: %d (expected: %d)\n", seriesCount, test.expectedSeries)

		if seriesCount == test.expectedSeries {
			fmt.Printf("   ✅ PASS\n\n")
		} else {
			fmt.Printf("   ❌ FAIL: Expected %d series, got %d\n\n", test.expectedSeries, seriesCount)
			allPassed = false
		}
	}

	if allPassed {
		fmt.Println("✅ All tests passed!")
		os.Exit(0)
	} else {
		fmt.Println("❌ Some tests failed!")
		os.Exit(1)
	}
}

func executeQuery(prometheusURL, query string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build query URL
	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d",
		prometheusURL,
		url.QueryEscape(query),
		time.Now().Unix(),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to execute query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, "", fmt.Errorf("prometheus returned status %d: %s", resp.StatusCode, string(body))
	}

	var promResp PrometheusResponse
	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil {
		return 0, "", fmt.Errorf("failed to decode response: %w", err)
	}

	if promResp.Status != "success" {
		return 0, "", fmt.Errorf("prometheus query failed: %s", promResp.Status)
	}

	return len(promResp.Data.Result), promResp.Data.ResultType, nil
}
