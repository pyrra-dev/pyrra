#!/bin/bash

# Recording Rules Validation Script
# This script validates recording rules generation for all indicator types

set -e

echo "=== Recording Rules Validation Script ==="
echo "This script will:"
echo "1. Deploy test SLOs for all indicator types"
echo "2. Wait for recording rules to be generated"
echo "3. Validate recording rules produce correct metrics"
echo "4. Test query efficiency and label handling"
echo

# Check prerequisites
echo "Checking prerequisites..."

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not available"
    exit 1
fi

# Check if Kubernetes cluster is accessible
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Kubernetes cluster is not accessible"
    exit 1
fi

# Check if Prometheus is accessible
if ! curl -s http://localhost:9090/api/v1/query?query=up > /dev/null; then
    echo "❌ Prometheus is not accessible at http://localhost:9090"
    echo "Please ensure Prometheus is running and accessible"
    exit 1
fi

echo "✅ Prerequisites check passed"
echo

# Deploy test SLOs
echo "=== Deploying Test SLOs ==="

test_slos=(
    ".dev/test-dynamic-slo.yaml"
    ".dev/test-static-slo.yaml"
    ".dev/test-latency-dynamic.yaml"
    ".dev/test-latency-static.yaml"
    ".dev/test-latency-native-dynamic.yaml"
    ".dev/test-bool-gauge-dynamic.yaml"
)

for slo_file in "${test_slos[@]}"; do
    if [ -f "$slo_file" ]; then
        echo "Deploying $slo_file..."
        kubectl apply -f "$slo_file"
    else
        echo "⚠️  Warning: $slo_file not found, skipping"
    fi
done

echo "✅ Test SLOs deployed"
echo

# Wait for Pyrra to process the SLOs and generate recording rules
echo "=== Waiting for Recording Rules Generation ==="
echo "Waiting 30 seconds for Pyrra to process SLOs and generate recording rules..."
sleep 30

# Check if recording rules were created
echo "Checking for PrometheusRule objects..."
kubectl get prometheusrule -n monitoring | grep -E "(test-dynamic|test-static|test-latency|test-bool-gauge)" || true
echo

# Wait for Prometheus to load the rules
echo "Waiting additional 30 seconds for Prometheus to load recording rules..."
sleep 30

# Build and run the validation tool
echo "=== Building Validation Tool ==="
cd cmd/validate-recording-rules
go build -o validate-recording-rules main.go
echo "✅ Validation tool built"
echo

# Run validation
echo "=== Running Recording Rules Validation ==="
./validate-recording-rules

# Check recording rule queries in Prometheus
echo
echo "=== Manual Verification Queries ==="
echo "You can manually verify recording rules in Prometheus UI (http://localhost:9090) with these queries:"
echo

echo "# Ratio indicators:"
echo "apiserver_request:burnrate5m{slo=\"test-dynamic-apiserver\"}"
echo "apiserver_request:burnrate5m{slo=\"test-static-apiserver\"}"
echo

echo "# Latency indicators:"
echo "prometheus_http_request_duration_seconds:burnrate5m{slo=\"test-latency-dynamic\"}"
echo "prometheus_http_request_duration_seconds:burnrate5m{slo=\"test-latency-static\"}"
echo

echo "# LatencyNative indicators:"
echo "connect_server_requests_duration_seconds:burnrate5m{slo=\"test-latency-native-dynamic\"}"
echo

echo "# BoolGauge indicators:"
echo "up:burnrate5m{slo=\"test-bool-gauge-dynamic\"}"
echo

echo "# Generic recording rules (all SLOs):"
echo "pyrra_availability"
echo "pyrra_requests:rate5m"
echo "pyrra_errors:rate5m"
echo

# Cleanup option
echo "=== Cleanup ==="
read -p "Do you want to remove test SLOs? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Removing test SLOs..."
    for slo_file in "${test_slos[@]}"; do
        if [ -f "$slo_file" ]; then
            kubectl delete -f "$slo_file" --ignore-not-found=true
        fi
    done
    echo "✅ Test SLOs removed"
else
    echo "Test SLOs left in place for manual inspection"
fi

echo
echo "=== Recording Rules Validation Complete ==="