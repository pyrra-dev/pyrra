# Native Histogram Setup Guide for LatencyNative Indicators

## Overview

This guide provides step-by-step instructions for enabling native histogram support in your Prometheus setup to use LatencyNative indicators with dynamic burn rates.

## Prerequisites

- kube-prometheus stack deployed
- Access to modify kube-prometheus configuration
- Pyrra deployed in your cluster

## Step 1: Enable Native Histograms in Prometheus

### Modify main.libsonnet

In your kube-prometheus configuration, locate the `main.libsonnet` file and update the Prometheus configuration:

**Before:**
```jsonnet
prometheus: prometheus($.values.prometheus),
```

**After:**
```jsonnet
prometheus: prometheus($.values.prometheus) + {
  prometheus+: {
    spec+: {
      enableFeatures: ['native-histograms'],
    },
  },
},
```

### Regenerate Manifests

After modifying the configuration, regenerate the Kubernetes manifests in shell with admin privileges:

```bash
make generate
```

### Apply Updated Configuration

Apply the updated manifests to your cluster:

```bash
# Create the namespace and CRDs, and then wait for them to be available before creating the remaining resources
# Note that due to some CRD size we are using kubectl server-side apply feature which is generally available since kubernetes 1.22.
# If you are using previous kubernetes versions this feature may not be available and you would need to use kubectl create instead.
kubectl apply --server-side -f manifests/setup
kubectl wait \
    --for condition=Established \
    --all CustomResourceDefinition \
    --namespace=monitoring
kubectl apply -f manifests/
```

## Step 2: Update Pyrra Binary

### Rebuild Pyrra with Native Histogram Support

Ensure your Pyrra binary includes the native histogram metric exposure:

```bash
# In your Pyrra repository
make build

# Or if using Docker
docker build -t your-registry/pyrra:latest .
```

### Deploy Updated Pyrra

Update your Pyrra deployment with the new binary:

```bash
kubectl rollout restart deployment/pyrra-api -n monitoring
kubectl rollout restart deployment/pyrra-kubernetes -n monitoring
```

## Step 3: Configure Native Histogram Metrics

### Create ServiceMonitor for Pyrra

Ensure Pyrra metrics are being scraped with native histogram support:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: pyrra
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: pyrra
  endpoints:
  - port: http
    interval: 30s
    path: /metrics
    # Enable native histogram scraping
    nativeHistogramBucketLimit: 1000
```

## Step 4: Verification

### Check Prometheus Configuration

Verify that native histograms are enabled:

```bash
# Check Prometheus status
curl http://prometheus:9090/api/v1/status/config | jq '.data.yaml' | grep -A5 -B5 "native"
```

### Verify Native Histogram Metrics

Check that native histogram metrics are available:

```bash
# Query for native histogram metrics
curl 'http://prometheus:9090/api/v1/query?query=connect_server_requests_duration_seconds' | jq
```

### Test LatencyNative SLO

Create a test LatencyNative SLO to verify functionality:

```yaml
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: test-latency-native
  namespace: monitoring
spec:
  target: "99"
  window: 28d
  indicator:
    latencyNative:
      total:
        metric: connect_server_requests_duration_seconds
      latency: 200ms
  alerting:
    burnRateType: dynamic
    burnRates:
    - severity: critical
      short: 1h
      long: 5m
      factor: 14
    - severity: warning  
      short: 6h
      long: 30m
      factor: 6
```

## Step 5: Monitoring and Troubleshooting

### Common Issues

1. **Native histograms not enabled:**
   - Check Prometheus logs for feature flag errors
   - Verify `enableFeatures` configuration is applied

2. **Metrics not available:**
   - Ensure Pyrra binary is updated and running
   - Check ServiceMonitor configuration
   - Verify scrape targets in Prometheus

3. **Query failures:**
   - Confirm `histogram_count()` function is available
   - Check metric names and labels
   - Verify time ranges in queries

### Debugging Commands

```bash
# Check Prometheus targets
curl http://prometheus:9090/api/v1/targets

# Check available metrics
curl http://prometheus:9090/api/v1/label/__name__/values | grep histogram

# Test histogram_count function
curl 'http://prometheus:9090/api/v1/query?query=histogram_count(connect_server_requests_duration_seconds)'
```

## Expected Results

After successful setup, you should see:

1. **Prometheus UI:** Native histogram metrics visible in metric browser
2. **Pyrra UI:** LatencyNative SLOs showing dynamic burn rate calculations
3. **Alerts:** Dynamic burn rate alerts firing based on traffic patterns
4. **Graphs:** Request rate graphs with average traffic baselines

## Performance Considerations

- Native histograms use more memory than traditional histograms
- Monitor Prometheus resource usage after enabling
- Consider adjusting `nativeHistogramBucketLimit` based on your needs
- Native histograms provide better accuracy for percentile calculations

## Rollback Procedure

If you need to disable native histograms:

1. Revert `main.libsonnet` changes
2. Run `make generate`
3. Apply updated manifests
4. Restart Prometheus

```jsonnet
// Rollback to:
prometheus: prometheus($.values.prometheus),
```

## Support

For issues with native histogram setup:
- Check Prometheus documentation for native histogram features
- Review Pyrra logs for metric exposure issues
- Verify kube-prometheus version compatibility
- Test with simple native histogram queries first