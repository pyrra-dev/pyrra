package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

type PerformanceMetrics struct {
	Timestamp          time.Time     `json:"timestamp"`
	APIResponseTime    time.Duration `json:"api_response_time_ms"`
	SLOCount           int           `json:"slo_count"`
	DynamicSLOCount    int           `json:"dynamic_slo_count"`
	StaticSLOCount     int           `json:"static_slo_count"`
	MemoryUsageMB      float64       `json:"memory_usage_mb"`
	GoroutineCount     int           `json:"goroutine_count"`
	PrometheusQueryOK  bool          `json:"prometheus_query_ok"`
	PrometheusRespTime time.Duration `json:"prometheus_response_time_ms"`
}

type ObjectiveList struct {
	Objectives []struct {
		Alerting struct {
			BurnRateType string `json:"burnRateType"`
		} `json:"alerting"`
	} `json:"objectives"`
}

func main() {
	apiURL := flag.String("api-url", "http://localhost:9099", "Pyrra API URL")
	promURL := flag.String("prom-url", "http://localhost:9090", "Prometheus URL")
	duration := flag.Duration("duration", 5*time.Minute, "Monitoring duration")
	interval := flag.Duration("interval", 10*time.Second, "Sampling interval")
	outputFile := flag.String("output", ".dev-docs/performance-metrics.json", "Output file for metrics")
	flag.Parse()

	fmt.Printf("Performance Monitoring Started\n")
	fmt.Printf("API URL: %s\n", *apiURL)
	fmt.Printf("Prometheus URL: %s\n", *promURL)
	fmt.Printf("Duration: %v\n", *duration)
	fmt.Printf("Interval: %v\n", *interval)
	fmt.Printf("Output: %s\n\n", *outputFile)

	var metrics []PerformanceMetrics
	startTime := time.Now()
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	// Initial sample
	if m, err := collectMetrics(*apiURL, *promURL); err == nil {
		metrics = append(metrics, m)
		printMetrics(m)
	} else {
		fmt.Printf("Error collecting initial metrics: %v\n", err)
	}

	// Periodic sampling
	for {
		select {
		case <-ticker.C:
			if time.Since(startTime) >= *duration {
				goto done
			}

			if m, err := collectMetrics(*apiURL, *promURL); err == nil {
				metrics = append(metrics, m)
				printMetrics(m)
			} else {
				fmt.Printf("Error collecting metrics: %v\n", err)
			}
		}
	}

done:
	// Save metrics to file
	if err := saveMetrics(metrics, *outputFile); err != nil {
		fmt.Printf("Error saving metrics: %v\n", err)
		os.Exit(1)
	}

	// Print summary
	printSummary(metrics)
	fmt.Printf("\nMetrics saved to: %s\n", *outputFile)
}

func collectMetrics(apiURL, promURL string) (PerformanceMetrics, error) {
	m := PerformanceMetrics{
		Timestamp: time.Now(),
	}

	// Collect memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.MemoryUsageMB = float64(memStats.Alloc) / 1024 / 1024
	m.GoroutineCount = runtime.NumGoroutine()

	// Test API response time
	apiStart := time.Now()
	resp, err := http.Post(apiURL+"/objectives.v1alpha1.ObjectiveService/List", "application/json", strings.NewReader("{}"))
	if err != nil {
		return m, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()
	m.APIResponseTime = time.Since(apiStart)

	// Parse SLO counts
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return m, fmt.Errorf("failed to read API response: %w", err)
	}

	var objList ObjectiveList
	if err := json.Unmarshal(body, &objList); err != nil {
		return m, fmt.Errorf("failed to parse API response: %w", err)
	}

	m.SLOCount = len(objList.Objectives)
	for _, obj := range objList.Objectives {
		if obj.Alerting.BurnRateType == "dynamic" {
			m.DynamicSLOCount++
		} else {
			m.StaticSLOCount++
		}
	}

	// Test Prometheus response time
	promStart := time.Now()
	promResp, err := http.Get(promURL + "/api/v1/query?query=up")
	if err == nil {
		promResp.Body.Close()
		m.PrometheusQueryOK = true
		m.PrometheusRespTime = time.Since(promStart)
	} else {
		m.PrometheusQueryOK = false
	}

	return m, nil
}

func printMetrics(m PerformanceMetrics) {
	fmt.Printf("[%s] SLOs: %d (%d dynamic, %d static) | API: %dms | Mem: %.1fMB | Goroutines: %d | Prom: ",
		m.Timestamp.Format("15:04:05"),
		m.SLOCount,
		m.DynamicSLOCount,
		m.StaticSLOCount,
		m.APIResponseTime.Milliseconds(),
		m.MemoryUsageMB,
		m.GoroutineCount,
	)
	if m.PrometheusQueryOK {
		fmt.Printf("%dms\n", m.PrometheusRespTime.Milliseconds())
	} else {
		fmt.Printf("FAIL\n")
	}
}

func printSummary(metrics []PerformanceMetrics) {
	if len(metrics) == 0 {
		return
	}

	var totalAPITime, totalPromTime time.Duration
	var totalMem float64
	var minAPITime, maxAPITime time.Duration = time.Hour, 0
	var minMem, maxMem float64 = 1e9, 0

	for _, m := range metrics {
		totalAPITime += m.APIResponseTime
		totalPromTime += m.PrometheusRespTime
		totalMem += m.MemoryUsageMB

		if m.APIResponseTime < minAPITime {
			minAPITime = m.APIResponseTime
		}
		if m.APIResponseTime > maxAPITime {
			maxAPITime = m.APIResponseTime
		}
		if m.MemoryUsageMB < minMem {
			minMem = m.MemoryUsageMB
		}
		if m.MemoryUsageMB > maxMem {
			maxMem = m.MemoryUsageMB
		}
	}

	count := len(metrics)
	avgAPITime := totalAPITime / time.Duration(count)
	avgPromTime := totalPromTime / time.Duration(count)
	avgMem := totalMem / float64(count)

	fmt.Printf("\n=== Performance Summary ===\n")
	fmt.Printf("Samples: %d\n", count)
	fmt.Printf("SLO Count: %d (%d dynamic, %d static)\n",
		metrics[len(metrics)-1].SLOCount,
		metrics[len(metrics)-1].DynamicSLOCount,
		metrics[len(metrics)-1].StaticSLOCount,
	)
	fmt.Printf("\nAPI Response Time:\n")
	fmt.Printf("  Average: %dms\n", avgAPITime.Milliseconds())
	fmt.Printf("  Min: %dms\n", minAPITime.Milliseconds())
	fmt.Printf("  Max: %dms\n", maxAPITime.Milliseconds())
	fmt.Printf("\nPrometheus Response Time:\n")
	fmt.Printf("  Average: %dms\n", avgPromTime.Milliseconds())
	fmt.Printf("\nMemory Usage:\n")
	fmt.Printf("  Average: %.1fMB\n", avgMem)
	fmt.Printf("  Min: %.1fMB\n", minMem)
	fmt.Printf("  Max: %.1fMB\n", maxMem)
	fmt.Printf("  Growth: %.1fMB\n", maxMem-minMem)
}

func saveMetrics(metrics []PerformanceMetrics, filename string) error {
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
