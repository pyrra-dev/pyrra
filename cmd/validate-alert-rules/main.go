package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var (
		prometheusURL = flag.String("prometheus-url", "http://localhost:9090", "Prometheus server URL")
		kubeconfig    = flag.String("kubeconfig", "", "Path to kubeconfig file")
		namespace     = flag.String("namespace", "monitoring", "Kubernetes namespace")
	)
	flag.Parse()

	log.Printf("=== Alert Rules Validation Tool ===\n")
	log.Printf("Prometheus URL: %s", *prometheusURL)
	log.Printf("Namespace: %s\n", *namespace)

	// Create Kubernetes client
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Create Prometheus client
	promClient, err := api.NewClient(api.Config{
		Address: *prometheusURL,
	})
	if err != nil {
		log.Fatalf("Error creating Prometheus client: %v", err)
	}
	promAPI := v1.NewAPI(promClient)

	ctx := context.Background()

	// Test SLOs to validate
	testSLOs := []struct {
		name          string
		indicatorType string
		burnRateType  string
	}{
		{"test-dynamic-apiserver", "ratio", "dynamic"},
		{"test-latency-dynamic", "latency", "dynamic"},
		{"test-latency-native-dynamic", "latencyNative", "dynamic"},
		{"test-bool-gauge-dynamic", "boolGauge", "dynamic"},
		{"test-static-apiserver", "ratio", "static"},
	}

	totalTests := 0
	passedTests := 0
	failedTests := 0

	for _, slo := range testSLOs {
		log.Printf("\n=== Validating SLO: %s (%s, %s) ===", slo.name, slo.indicatorType, slo.burnRateType)
		
		// Test 1: Check PrometheusRule exists
		totalTests++
		log.Printf("\n[Test 1] Checking PrometheusRule exists...")
		_, err := clientset.CoreV1().RESTClient().
			Get().
			AbsPath("/apis/monitoring.coreos.com/v1").
			Namespace(*namespace).
			Resource("prometheusrules").
			Name(slo.name).
			DoRaw(ctx)
		
		if err != nil {
			log.Printf("❌ FAIL: PrometheusRule not found: %v", err)
			failedTests++
			continue
		}
		log.Printf("✅ PASS: PrometheusRule exists")
		passedTests++

		// Test 2: Check recording rules are loaded in Prometheus
		totalTests++
		log.Printf("\n[Test 2] Checking recording rules in Prometheus...")
		rules, err := promAPI.Rules(ctx)
		if err != nil {
			log.Printf("❌ FAIL: Error querying Prometheus rules: %v", err)
			failedTests++
			continue
		}

		foundRecordingRules := false
		recordingRuleCount := 0
		for _, group := range rules.Groups {
			for _, rule := range group.Rules {
				switch r := rule.(type) {
				case v1.RecordingRule:
					if strings.Contains(r.Name, slo.name) || 
					   (r.Labels != nil && string(r.Labels["slo"]) == slo.name) {
						foundRecordingRules = true
						recordingRuleCount++
					}
				}
			}
		}

		if !foundRecordingRules {
			log.Printf("❌ FAIL: No recording rules found for SLO")
			failedTests++
		} else {
			log.Printf("✅ PASS: Found %d recording rules", recordingRuleCount)
			passedTests++
		}

		// Test 3: Check alert rules are loaded in Prometheus
		totalTests++
		log.Printf("\n[Test 3] Checking alert rules in Prometheus...")
		foundAlertRules := false
		alertRuleCount := 0
		var alertExpressions []string

		for _, group := range rules.Groups {
			for _, rule := range group.Rules {
				switch r := rule.(type) {
				case v1.AlertingRule:
					if r.Labels != nil && string(r.Labels["slo"]) == slo.name {
						foundAlertRules = true
						alertRuleCount++
						alertExpressions = append(alertExpressions, r.Query)
					}
				}
			}
		}

		if !foundAlertRules {
			log.Printf("❌ FAIL: No alert rules found for SLO")
			failedTests++
		} else {
			log.Printf("✅ PASS: Found %d alert rules", alertRuleCount)
			passedTests++
		}

		// Test 4: Validate alert expressions reference recording rules (for dynamic)
		if slo.burnRateType == "dynamic" && len(alertExpressions) > 0 {
			totalTests++
			log.Printf("\n[Test 4] Validating alert expressions reference recording rules...")
			
			allUseRecordingRules := true
			for i, expr := range alertExpressions {
				// Dynamic alert expressions should reference burnrate recording rules
				if !strings.Contains(expr, ":burnrate") {
					log.Printf("❌ Alert expression %d doesn't reference recording rules: %s", i+1, expr)
					allUseRecordingRules = false
				}
			}

			if allUseRecordingRules {
				log.Printf("✅ PASS: All alert expressions reference recording rules")
				passedTests++
			} else {
				log.Printf("❌ FAIL: Some alert expressions don't reference recording rules")
				failedTests++
			}
		}

		// Test 5: Validate dynamic threshold calculation structure
		if slo.burnRateType == "dynamic" && len(alertExpressions) > 0 {
			totalTests++
			log.Printf("\n[Test 5] Validating dynamic threshold calculation structure...")
			
			allHaveScalar := true
			allHaveIncrease := true
			for i, expr := range alertExpressions {
				if !strings.Contains(expr, "scalar(") {
					log.Printf("❌ Alert expression %d missing scalar() wrapper: %s", i+1, expr)
					allHaveScalar = false
				}
				if !strings.Contains(expr, "increase(") {
					log.Printf("❌ Alert expression %d missing increase() for traffic calculation: %s", i+1, expr)
					allHaveIncrease = false
				}
			}

			if allHaveScalar && allHaveIncrease {
				log.Printf("✅ PASS: All alert expressions have correct dynamic threshold structure")
				passedTests++
			} else {
				log.Printf("❌ FAIL: Some alert expressions have incorrect structure")
				failedTests++
			}
		}

		// Test 6: Query recording rules to verify they return data
		totalTests++
		log.Printf("\n[Test 6] Querying recording rules to verify data...")
		
		// Try to query a burnrate recording rule
		query := fmt.Sprintf(`{slo="%s",__name__=~".*:burnrate.*"}`, slo.name)
		result, warnings, err := promAPI.Query(ctx, query, time.Now())
		if err != nil {
			log.Printf("❌ FAIL: Error querying recording rules: %v", err)
			failedTests++
		} else {
			if len(warnings) > 0 {
				log.Printf("⚠️  Warnings: %v", warnings)
			}
			
			resultStr := result.String()
			if strings.Contains(resultStr, "=>") {
				log.Printf("✅ PASS: Recording rules return data")
				passedTests++
			} else {
				log.Printf("❌ FAIL: Recording rules return no data")
				failedTests++
			}
		}
	}

	// Summary
	log.Printf("\n\n=== VALIDATION SUMMARY ===")
	log.Printf("Total Tests: %d", totalTests)
	log.Printf("✅ Passed: %d", passedTests)
	log.Printf("❌ Failed: %d", failedTests)
	
	if failedTests == 0 {
		log.Printf("\n🎉 All tests passed!")
	} else {
		log.Printf("\n⚠️  Some tests failed. Review the output above for details.")
	}
}
