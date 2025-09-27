package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	"github.com/pyrra-dev/pyrra/testing"
)

func main() {
	var (
		prometheusURL   = flag.String("prometheus-url", "http://localhost:9090", "Prometheus server URL")
		alertManagerURL = flag.String("alertmanager-url", "http://localhost:9093", "AlertManager server URL")
		pushGatewayURL  = flag.String("push-gateway-url", "http://172.24.13.124:9091", "Push Gateway URL")
		testDuration    = flag.Duration("duration", 5*time.Minute, "Test duration")
		errorRate       = flag.Float64("error-rate", 0.15, "Error rate (0.0 to 1.0)")
	)
	flag.Parse()

	log.Printf("Running synthetic metric test to validate alert firing")
	log.Printf("Prometheus URL: %s", *prometheusURL)
	log.Printf("AlertManager URL: %s", *alertManagerURL)
	log.Printf("Push Gateway URL: %s", *pushGatewayURL)
	log.Printf("Error Rate: %.1f%%", *errorRate*100)
	log.Printf("Test Duration: %v", *testDuration)

	// Create Prometheus client
	promClient, err := api.NewClient(api.Config{
		Address: *prometheusURL,
	})
	if err != nil {
		log.Fatalf("Error creating Prometheus client: %v", err)
	}

	// Create synthetic metric generator
	generator := testing.NewSyntheticMetricGenerator(promClient, *pushGatewayURL)

	// Create Prometheus alerts client for detecting pending and firing alerts
	alertsClient := testing.NewPrometheusAlertsClient(promClient)

	// Define synthetic traffic pattern that should trigger alerts
	pattern := testing.TrafficPattern{
		Service:        "synthetic-alert-test",
		Method:         "GET",
		RequestsPerSec: 20.0,
		ErrorRate:      *errorRate, // This should trigger dynamic burn rate alerts
		Duration:       *testDuration,
		LatencyMs:      100,
	}

	ctx, cancel := context.WithTimeout(context.Background(), *testDuration+2*time.Minute)
	defer cancel()

	log.Printf("\n=== STEP 1: BASELINE CHECK ===")
	// Check if there are any existing alerts for our synthetic service
	baselineAlerts, err := alertsClient.GetAlertsForService(ctx, pattern.Service)
	if err != nil {
		log.Printf("Baseline check failed: %v", err)
	} else {
		log.Printf("Baseline alerts for %s: %d", pattern.Service, len(baselineAlerts))
	}

	log.Printf("\n=== STEP 2: GENERATING SYNTHETIC METRICS ===")
	log.Printf("Starting synthetic traffic generation...")
	log.Printf("Service: %s", pattern.Service)
	log.Printf("Traffic: %.1f req/sec with %.1f%% error rate", pattern.RequestsPerSec, pattern.ErrorRate*100)

	// Start synthetic metric generation
	trafficDone := make(chan error, 1)
	go func() {
		trafficDone <- generator.GenerateTrafficPattern(ctx, pattern)
	}()

	// Monitor for alerts while generating traffic
	log.Printf("\n=== STEP 3: MONITORING FOR ALERTS ===")
	monitorTicker := time.NewTicker(30 * time.Second)
	defer monitorTicker.Stop()

	alertDetected := false
	monitoringEnd := time.Now().Add(*testDuration + 1*time.Minute) // Monitor a bit longer

	for time.Now().Before(monitoringEnd) {
		select {
		case <-ctx.Done():
			goto cleanup
		case err := <-trafficDone:
			if err != nil {
				log.Printf("Traffic generation completed with error: %v", err)
			} else {
				log.Printf("✅ Traffic generation completed successfully")
			}
		case <-monitorTicker.C:
			log.Printf("Checking for alerts (pending and firing states)...")

			// Get alerts related to our synthetic service from Prometheus
			serviceAlerts, err := alertsClient.GetAlertsForService(ctx, pattern.Service)
			if err != nil {
				log.Printf("Error querying Prometheus alerts: %v", err)
				continue
			}

			// Count alerts by state
			pendingCount := 0
			firingCount := 0
			inactiveCount := 0

			for _, alert := range serviceAlerts {
				switch alert.State {
				case "pending":
					pendingCount++
				case "firing":
					firingCount++
				case "inactive":
					inactiveCount++
				}
			}

			totalAlerts := len(serviceAlerts)
			if totalAlerts > 0 {
				log.Printf("🎯 SYNTHETIC ALERTS DETECTED! Total: %d (🟡 Pending: %d, 🔥 Firing: %d, ⚪ Inactive: %d)",
					totalAlerts, pendingCount, firingCount, inactiveCount)

				// Print detailed alert summary
				err = alertsClient.PrintAlertSummary(ctx, pattern.Service)
				if err != nil {
					log.Printf("Error printing alert summary: %v", err)
				}

				alertDetected = true
			} else {
				log.Printf("No synthetic alerts detected yet...")
			}

			// Also check for any burn rate alerts that might be related
			allAlerts, err := alertsClient.GetAlerts(ctx)
			if err == nil {
				burnRateCount := 0
				for _, alert := range allAlerts {
					if alertname, exists := alert.Labels["alertname"]; exists {
						if strings.Contains(strings.ToLower(alertname), "burn") ||
							strings.Contains(strings.ToLower(alertname), "budget") {
							burnRateCount++
						}
					}
				}
				if burnRateCount > 0 {
					log.Printf("ℹ️  General burn rate alerts detected: %d", burnRateCount)
				}
			}
		}
	}

cleanup:
	log.Printf("\n=== STEP 4: CLEANUP AND SUMMARY ===")

	// Cleanup synthetic metrics
	if err := generator.Cleanup(); err != nil {
		log.Printf("Warning: cleanup error: %v", err)
	}

	// Final summary
	log.Printf("\n=== SYNTHETIC ALERT TEST SUMMARY ===")
	if alertDetected {
		log.Printf("🎉 SUCCESS: Alert firing was detected during synthetic metric generation!")
		log.Printf("   This validates that the alert pipeline is working correctly.")
	} else {
		log.Printf("ℹ️  INFO: No alerts were detected during the test.")
		log.Printf("   This could mean:")
		log.Printf("   1. The error rate (%.1f%%) was not high enough to trigger alerts", *errorRate*100)
		log.Printf("   2. No SLOs are configured to monitor the synthetic service")
		log.Printf("   3. Alert evaluation intervals need more time")
		log.Printf("   4. The synthetic metrics are not being scraped by Prometheus")
	}

	log.Printf("\n=== RECOMMENDATIONS ===")
	log.Printf("To improve alert testing:")
	log.Printf("1. Ensure Push Gateway is running and accessible at %s", *pushGatewayURL)
	log.Printf("2. Configure Prometheus to scrape Push Gateway")
	log.Printf("3. Create SLOs that monitor synthetic-alert-test service")
	log.Printf("4. Try higher error rates (e.g., -error-rate=0.3 for 30%%)")
	log.Printf("5. Check Prometheus targets and rules configuration")

	log.Printf("\nSynthetic alert test completed!")
}
