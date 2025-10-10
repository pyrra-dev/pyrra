package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	count := flag.Int("count", 10, "Number of SLOs to generate")
	namespace := flag.String("namespace", "monitoring", "Kubernetes namespace")
	outputDir := flag.String("output", ".dev/generated-slos", "Output directory for YAML files")
	dynamicRatio := flag.Float64("dynamic-ratio", 0.5, "Ratio of dynamic SLOs (0.0-1.0)")
	flag.Parse()

	if *count <= 0 {
		fmt.Println("Error: count must be positive")
		os.Exit(1)
	}

	if *dynamicRatio < 0 || *dynamicRatio > 1 {
		fmt.Println("Error: dynamic-ratio must be between 0.0 and 1.0")
		os.Exit(1)
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	dynamicCount := int(float64(*count) * (*dynamicRatio))
	staticCount := *count - dynamicCount

	fmt.Printf("Generating %d SLOs (%d dynamic, %d static) in namespace '%s'\n", *count, dynamicCount, staticCount, *namespace)
	fmt.Printf("Output directory: %s\n", *outputDir)

	// Generate dynamic SLOs
	for i := 0; i < dynamicCount; i++ {
		sloName := fmt.Sprintf("test-dynamic-slo-%d", i+1)
		filename := filepath.Join(*outputDir, fmt.Sprintf("%s.yaml", sloName))

		if err := generateDynamicSLO(sloName, *namespace, filename, i); err != nil {
			fmt.Printf("Error generating %s: %v\n", sloName, err)
			continue
		}
		fmt.Printf("✓ Generated: %s\n", filename)
	}

	// Generate static SLOs
	for i := 0; i < staticCount; i++ {
		sloName := fmt.Sprintf("test-static-slo-%d", i+1)
		filename := filepath.Join(*outputDir, fmt.Sprintf("%s.yaml", sloName))

		if err := generateStaticSLO(sloName, *namespace, filename, i); err != nil {
			fmt.Printf("Error generating %s: %v\n", sloName, err)
			continue
		}
		fmt.Printf("✓ Generated: %s\n", filename)
	}

	fmt.Printf("\nSuccess! Generated %d SLO files\n", *count)
	fmt.Printf("\nTo apply all SLOs:\n")
	fmt.Printf("  kubectl apply -f %s/\n", *outputDir)
	fmt.Printf("\nTo delete all SLOs:\n")
	fmt.Printf("  kubectl delete -f %s/\n", *outputDir)
}

func generateDynamicSLO(name, namespace, filename string, index int) error {
	// Vary indicator types, targets, and windows for diversity
	indicatorType := []string{"ratio", "latency"}[index%2]
	target := []string{"99", "99.5", "99.9", "95"}[index%4]
	window := []string{"7d", "28d", "30d"}[index%3]

	var indicatorYAML string
	if indicatorType == "ratio" {
		indicatorYAML = `    ratio:
      errors:
        metric: apiserver_request_total{job="apiserver",code=~"5.."}
      total:
        metric: apiserver_request_total{job="apiserver"}`
	} else {
		indicatorYAML = `    latency:
      success:
        metric: prometheus_http_request_duration_seconds_bucket{handler="/api/v1/query"}
      total:
        metric: prometheus_http_request_duration_seconds_count{handler="/api/v1/query"}
      threshold: "0.1"`
	}

	yaml := fmt.Sprintf(`apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: %s
  namespace: %s
  labels:
    prometheus: k8s
    role: alert-rules
    pyrra.dev/team: platform
spec:
  target: "%s"
  window: %s
  description: "Dynamic burn rate test SLO #%d - %s indicator, %s window"
  alerting:
    burnRateType: dynamic
    burnrates: true
    absent: true
  indicator:
%s
`, name, namespace, target, window, index+1, indicatorType, window, indicatorYAML)

	return os.WriteFile(filename, []byte(yaml), 0644)
}

func generateStaticSLO(name, namespace, filename string, index int) error {
	// Vary indicator types, targets, and windows for diversity
	indicatorType := []string{"ratio", "latency"}[index%2]
	target := []string{"99", "99.5", "99.9", "95"}[index%4]
	window := []string{"7d", "28d", "30d"}[index%3]

	var indicatorYAML string
	if indicatorType == "ratio" {
		indicatorYAML = `    ratio:
      errors:
        metric: apiserver_request_total{job="apiserver",code=~"5.."}
      total:
        metric: apiserver_request_total{job="apiserver"}`
	} else {
		indicatorYAML = `    latency:
      success:
        metric: prometheus_http_request_duration_seconds_bucket{handler="/api/v1/query"}
      total:
        metric: prometheus_http_request_duration_seconds_count{handler="/api/v1/query"}
      threshold: "0.1"`
	}

	yaml := fmt.Sprintf(`apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: %s
  namespace: %s
  labels:
    prometheus: k8s
    role: alert-rules
    pyrra.dev/team: platform
spec:
  target: "%s"
  window: %s
  description: "Static burn rate test SLO #%d - %s indicator, %s window"
  alerting:
    burnRateType: static
    burnrates: true
    absent: true
  indicator:
%s
`, name, namespace, target, window, index+1, indicatorType, window, indicatorYAML)

	return os.WriteFile(filename, []byte(yaml), 0644)
}
