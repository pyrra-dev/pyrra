#!/bin/bash

# Production Readiness Testing Script
# This script automates testing of the dynamic burn rate feature at scale

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:9099}"
PROM_URL="${PROM_URL:-http://localhost:9090}"
NAMESPACE="${NAMESPACE:-monitoring}"
OUTPUT_DIR=".dev-docs/production-readiness-results"

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo "========================================="
echo "Production Readiness Testing"
echo "========================================="
echo "API URL: $API_URL"
echo "Prometheus URL: $PROM_URL"
echo "Namespace: $NAMESPACE"
echo "Output Directory: $OUTPUT_DIR"
echo ""

# Test 1: Service Health Check
echo -e "${YELLOW}Test 1: Service Health Check${NC}"
echo "Checking if Pyrra API is accessible..."
if curl -s -f "$API_URL" > /dev/null; then
    echo -e "${GREEN}✓ Pyrra API is accessible${NC}"
else
    echo -e "${RED}✗ Pyrra API is not accessible${NC}"
    exit 1
fi

echo "Checking if Prometheus is accessible..."
if curl -s -f "$PROM_URL/api/v1/query?query=up" > /dev/null; then
    echo -e "${GREEN}✓ Prometheus is accessible${NC}"
else
    echo -e "${RED}✗ Prometheus is not accessible${NC}"
    exit 1
fi
echo ""

# Test 2: API Response Time
echo -e "${YELLOW}Test 2: API Response Time${NC}"
echo "Measuring API response time..."
START_TIME=$(date +%s%N)
curl -s -X POST -H "Content-Type: application/json" -d '{}' \
    "$API_URL/objectives.v1alpha1.ObjectiveService/List" > "$OUTPUT_DIR/api-response.json"
END_TIME=$(date +%s%N)
RESPONSE_TIME=$(( (END_TIME - START_TIME) / 1000000 ))
echo "API Response Time: ${RESPONSE_TIME}ms"

if [ $RESPONSE_TIME -lt 3000 ]; then
    echo -e "${GREEN}✓ API response time acceptable (<3s)${NC}"
else
    echo -e "${YELLOW}⚠ API response time slow (>3s)${NC}"
fi
echo ""

# Test 3: SLO Count and Distribution
echo -e "${YELLOW}Test 3: SLO Count and Distribution${NC}"
echo "Analyzing SLO distribution..."
TOTAL_SLOS=$(jq '.objectives | length' "$OUTPUT_DIR/api-response.json")
DYNAMIC_SLOS=$(jq '[.objectives[] | select(.labels.burnRateType == "dynamic")] | length' "$OUTPUT_DIR/api-response.json")
STATIC_SLOS=$(jq '[.objectives[] | select(.labels.burnRateType == "static")] | length' "$OUTPUT_DIR/api-response.json")

echo "Total SLOs: $TOTAL_SLOS"
echo "Dynamic SLOs: $DYNAMIC_SLOS"
echo "Static SLOs: $STATIC_SLOS"

if [ $TOTAL_SLOS -gt 0 ]; then
    echo -e "${GREEN}✓ SLOs found in system${NC}"
else
    echo -e "${RED}✗ No SLOs found${NC}"
    exit 1
fi
echo ""

# Test 4: Prometheus Query Load
echo -e "${YELLOW}Test 4: Prometheus Query Load${NC}"
echo "Checking Prometheus query rate..."
QUERY_RATE=$(curl -s "$PROM_URL/api/v1/query?query=rate(prometheus_http_requests_total[5m])" | \
    jq -r '.data.result[0].value[1] // "0"')
echo "Prometheus query rate: $QUERY_RATE queries/sec"

# Convert to queries per minute
QUERIES_PER_MIN=$(echo "$QUERY_RATE * 60" | bc)
echo "Estimated queries per minute: $QUERIES_PER_MIN"

if (( $(echo "$QUERIES_PER_MIN < 100" | bc -l) )); then
    echo -e "${GREEN}✓ Prometheus query load acceptable (<100/min)${NC}"
else
    echo -e "${YELLOW}⚠ Prometheus query load high (>100/min)${NC}"
fi
echo ""

# Test 5: Memory Usage Check
echo -e "${YELLOW}Test 5: Memory Usage Check${NC}"
echo "Checking Go runtime memory usage..."
# This would require exposing metrics endpoint, skip for now
echo "Note: Memory profiling requires pprof endpoint"
echo "Run: go tool pprof http://localhost:9099/debug/pprof/heap"
echo ""

# Test 6: Kubernetes Resources
echo -e "${YELLOW}Test 6: Kubernetes Resources${NC}"
echo "Checking SLO CRDs in Kubernetes..."
if command -v kubectl &> /dev/null; then
    K8S_SLOS=$(kubectl get slo -n "$NAMESPACE" --no-headers 2>/dev/null | wc -l)
    echo "SLOs in Kubernetes: $K8S_SLOS"
    
    if [ $K8S_SLOS -eq $TOTAL_SLOS ]; then
        echo -e "${GREEN}✓ Kubernetes SLO count matches API${NC}"
    else
        echo -e "${YELLOW}⚠ Kubernetes SLO count mismatch${NC}"
    fi
else
    echo "kubectl not available, skipping Kubernetes checks"
fi
echo ""

# Test 7: Recording Rules Check
echo -e "${YELLOW}Test 7: Recording Rules Check${NC}"
echo "Checking if recording rules exist..."
RECORDING_RULES=$(curl -s "$PROM_URL/api/v1/rules" | \
    jq '[.data.groups[].rules[] | select(.type == "recording")] | length')
echo "Recording rules found: $RECORDING_RULES"

if [ $RECORDING_RULES -gt 0 ]; then
    echo -e "${GREEN}✓ Recording rules exist${NC}"
else
    echo -e "${YELLOW}⚠ No recording rules found${NC}"
fi
echo ""

# Test 8: Alert Rules Check
echo -e "${YELLOW}Test 8: Alert Rules Check${NC}"
echo "Checking if alert rules exist..."
ALERT_RULES=$(curl -s "$PROM_URL/api/v1/rules" | \
    jq '[.data.groups[].rules[] | select(.type == "alerting")] | length')
echo "Alert rules found: $ALERT_RULES"

if [ $ALERT_RULES -gt 0 ]; then
    echo -e "${GREEN}✓ Alert rules exist${NC}"
else
    echo -e "${YELLOW}⚠ No alert rules found${NC}"
fi
echo ""

# Generate Summary Report
echo -e "${YELLOW}Generating Summary Report${NC}"
cat > "$OUTPUT_DIR/test-summary.txt" << EOF
Production Readiness Test Summary
==================================
Date: $(date)
API URL: $API_URL
Prometheus URL: $PROM_URL

Test Results:
-------------
1. Service Health: PASS
2. API Response Time: ${RESPONSE_TIME}ms
3. Total SLOs: $TOTAL_SLOS (Dynamic: $DYNAMIC_SLOS, Static: $STATIC_SLOS)
4. Prometheus Query Rate: $QUERIES_PER_MIN queries/min
5. Memory Usage: See pprof for details
6. Kubernetes SLOs: $K8S_SLOS
7. Recording Rules: $RECORDING_RULES
8. Alert Rules: $ALERT_RULES

Recommendations:
----------------
EOF

# Add recommendations based on results
if [ $RESPONSE_TIME -gt 3000 ]; then
    echo "- Consider optimizing API response time (currently ${RESPONSE_TIME}ms)" >> "$OUTPUT_DIR/test-summary.txt"
fi

if (( $(echo "$QUERIES_PER_MIN > 100" | bc -l) )); then
    echo "- Monitor Prometheus query load (currently $QUERIES_PER_MIN/min)" >> "$OUTPUT_DIR/test-summary.txt"
fi

if [ $TOTAL_SLOS -lt 10 ]; then
    echo "- Test with more SLOs for better scale validation (currently $TOTAL_SLOS)" >> "$OUTPUT_DIR/test-summary.txt"
fi

echo "- Run browser compatibility tests manually" >> "$OUTPUT_DIR/test-summary.txt"
echo "- Perform load testing with 50+ SLOs" >> "$OUTPUT_DIR/test-summary.txt"
echo "- Monitor memory usage over extended period" >> "$OUTPUT_DIR/test-summary.txt"

cat "$OUTPUT_DIR/test-summary.txt"
echo ""
echo -e "${GREEN}Test summary saved to: $OUTPUT_DIR/test-summary.txt${NC}"
echo ""

# Final Status
echo "========================================="
echo -e "${GREEN}Production Readiness Testing Complete${NC}"
echo "========================================="
echo ""
echo "Next Steps:"
echo "1. Review test summary: $OUTPUT_DIR/test-summary.txt"
echo "2. Run browser compatibility tests (see BROWSER_COMPATIBILITY_TEST_GUIDE.md)"
echo "3. Generate test SLOs: go run cmd/generate-test-slos/main.go -count=50"
echo "4. Monitor performance: go run cmd/monitor-performance/main.go -duration=5m"
echo "5. Review all documentation in .dev-docs/"
