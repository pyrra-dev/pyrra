package main

import (
	"context"
	"log"
	"time"

	testing "github.com/pyrra-dev/pyrra/testing"
)

func main() {
	log.Printf("Testing service health checker...")

	// Create health checker with default URLs
	healthChecker := testing.NewServiceHealthChecker()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run health checks
	results, err := healthChecker.CheckAllServices(ctx)
	if err != nil {
		log.Fatalf("Failed to perform health checks: %v", err)
	}

	// Print results
	healthChecker.PrintHealthCheckResults(results)

	// Validate required services
	if err := healthChecker.ValidateRequiredServices(ctx); err != nil {
		log.Fatalf("Required service validation failed: %v", err)
	}

	log.Printf("✅ Service health check completed successfully!")
}
