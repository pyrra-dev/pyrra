package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func main() {
	prometheusURL := "http://localhost:9090"
	query := `apiserver_request:increase30d{slo="test-dynamic-apiserver"}`
	
	fmt.Println("=== Inspecting Recording Rule Series ===")
	fmt.Printf("Query: %s\n\n", query)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s&time=%d",
		prometheusURL,
		url.QueryEscape(query),
		time.Now().Unix(),
	)
	
	req, err := http.NewRequestWithContext(ctx, "GET", queryURL, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	var promResp struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(body, &promResp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("Status: %s\n", promResp.Status)
	fmt.Printf("Result Type: %s\n", promResp.Data.ResultType)
	fmt.Printf("Series Count: %d\n\n", len(promResp.Data.Result))
	
	for i, result := range promResp.Data.Result {
		fmt.Printf("Series %d:\n", i+1)
		fmt.Printf("  Labels: %v\n", result.Metric)
		fmt.Printf("  Value: %v\n\n", result.Value)
	}
}
