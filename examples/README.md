# Pyrra Examples

This directory contains example Service Level Objective (SLO) configurations for Pyrra.

## Service-Specific Examples

- `prometheus-http.yaml` - Prometheus HTTP request success rate
- `prometheus-operator.yaml` - Prometheus Operator availability
- `prometheus-probe-success.yaml` - Blackbox Exporter probe success
- `pyrra-connect-errors.yaml` - Pyrra API error rate
- `pyrra-connect-latency.yaml` - Pyrra API latency (native histogram)
- `pyrra-filesystem-errors.yaml` - Pyrra filesystem backend errors
- `pyrra-kubernetes-errors.yaml` - Pyrra Kubernetes backend errors
- `nginx.yaml` - NGINX request success rate and latency
- `caddy-response-latency.yaml` - Caddy server response latency
- `parca-grpc-queryrange-errors.yaml` - Parca gRPC query errors
- `thanos-grpc.yaml` - Thanos gRPC request success
- `thanos-http.yaml` - Thanos HTTP request success

## Dynamic Burn Rate Examples

Dynamic burn rates adapt alert thresholds based on actual traffic patterns, preventing false positives during low traffic and maintaining sensitivity during high traffic.

**When to use:**
- Services with variable traffic patterns (business hours vs nights/weekends)
- When you want alerts that adapt to actual service load

**Configuration:**
```yaml
alerting:
  burnRateType: dynamic  # or "static" for traditional behavior
```

### Examples by Indicator Type

- `dynamic-burn-rate-ratio.yaml` - Ratio indicator (API server requests)
- `dynamic-burn-rate-latency.yaml` - Latency indicator (Prometheus HTTP latency)
- `dynamic-burn-rate-latency-native.yaml` - LatencyNative indicator (Pyrra API latency)
- `dynamic-burn-rate-bool-gauge.yaml` - BoolGauge indicator (Prometheus availability)

## Indicator Types

Pyrra supports four indicator types:

1. **Ratio** - Success/failure ratio (e.g., HTTP 5xx errors vs total requests)
2. **Latency** - Request latency using traditional histograms
3. **LatencyNative** - Request latency using native histograms (requires Prometheus 2.40+)
4. **BoolGauge** - Binary states (1=up/success, 0=down/failure)

## Deployment Examples

- `kubernetes/` - Kubernetes manifests and CRD examples
- `docker-compose/` - Docker Compose setup with Prometheus
- `grafana/` - Grafana dashboard JSON files
- `mimir/` - Grafana Mimir integration
- `openshift/` - OpenShift-specific deployment

## Getting Started

1. Choose an example that matches your use case
2. Update metric names to match your service
3. Adjust target percentage and time window
4. Deploy the SLO:
   - **Kubernetes:** `kubectl apply -f your-slo.yaml`
   - **Filesystem:** Place YAML in configured directory
5. Verify in Pyrra UI (default: http://localhost:9099)

## Additional Resources

- Main documentation: See project README.md
- Grafana dashboards: See `grafana/README.md`
- Kubernetes deployment: See `kubernetes/README.md`
