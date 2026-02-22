# Troubleshooting Guide: Common SLO Configuration Issues

Real-world problems and solutions from production Pyrra deployments.

## Issue 1: "No data points" in Error Budget Graph

### Symptoms
- SLO appears in UI but graph is empty
- Error budget shows "NaN%" or blank
- Recording rules exist in Prometheus but no data

### Root Cause
**Recording rules aren't matching any metrics** due to label mismatches.

### Diagnosis

```bash
# Check if recording rule is firing
curl -s 'http://prometheus:9090/api/v1/query?query=pyrra_slo_availability:5m' | jq

# Check if base metrics exist
curl -s 'http://prometheus:9090/api/v1/query?query=http_requests_total{job="myapp"}' | jq
```

### Solution 1: Fix Label Selectors

**Problem:** Your metric has different labels than your SLO spec.

```yaml
# ❌ Wrong - metric doesn't have 'code' label
spec:
  indicator:
    ratio:
      errors:
        metric: http_requests_total{job="myapp",code=~"5.."}
      total:
        metric: http_requests_total{job="myapp"}
```

**Check actual labels:**
```bash
curl -s 'http://prometheus:9090/api/v1/label/__name__/values' | \
  jq '.data[] | select(. == "http_requests_total")'

curl -s 'http://prometheus:9090/api/v1/query?query=http_requests_total{job="myapp"}[1m]' | \
  jq '.data.result[0].metric'
```

**Output:**
```json
{
  "__name__": "http_requests_total",
  "job": "myapp",
  "status": "500",    ← It's 'status', not 'code'!
  "method": "POST"
}
```

**Fix:**
```yaml
# ✅ Correct
spec:
  indicator:
    ratio:
      errors:
        metric: http_requests_total{job="myapp",status=~"5.."}
      total:
        metric: http_requests_total{job="myapp"}
```

### Solution 2: Verify Scrape Interval

**Problem:** Metric scrape interval doesn't match recording rule evaluation interval.

```yaml
# Prometheus config
global:
  scrape_interval: 60s  ← Scraping every 60s

# But Pyrra expects data every 5s
```

**Fix:** Adjust recording rule intervals in Pyrra config:

```yaml
# pyrra-config.yaml
spec:
  # Match your Prometheus scrape interval
  indicator:
    ratio:
      errors:
        metric: rate(http_requests_total{status=~"5.."}[2m])  ← 2x scrape interval
      total:
        metric: rate(http_requests_total[2m])
```

---

## Issue 2: Alerts Firing Too Often (False Positives)

### Symptoms
- Getting paged for minor, short-lived incidents
- Error budget burns too fast on small blips
- Alerts fire during maintenance windows

### Root Cause
**Multi-window burn rate thresholds too aggressive** for your traffic patterns.

### Diagnosis

```bash
# Check alert firing frequency
curl -s 'http://prometheus:9090/api/v1/query?query=ALERTS{alertname=~"ErrorBudgetBurn.*",alertstate="firing"}' | jq
```

### Solution 1: Tune Burn Rate Windows

Pyrra generates 4 alerts by default:

| Alert | Window | Burn Rate | Budget Consumed |
|-------|--------|-----------|-----------------|
| Critical | 1h | 14.4x | 2% in 1h |
| Critical | 6h | 6x | 5% in 6h |
| Warning | 24h | 3x | 10% in 24h |
| Warning | 72h | 1x | 10% in 72h |

**Problem:** 1-hour critical alert too sensitive for bursty traffic.

**Fix:** Create custom alert rules:

```yaml
# custom-alerts.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: pyrra-custom-alerts
spec:
  groups:
    - name: slo-custom
      interval: 30s
      rules:
        # More lenient critical alert (3h instead of 1h)
        - alert: ErrorBudgetBurn-Critical-3h
          expr: |
            pyrra_slo_error_budget_burn_rate{slo="myapp-availability"} > 4.8
            and
            increase(pyrra_slo_error_budget_consumed_ratio{slo="myapp-availability"}[3h]) > 0.05
          for: 5m  # Must be sustained for 5min
          labels:
            severity: critical
          annotations:
            summary: "High error budget burn rate"
            description: "{{ $value | humanizePercentage }} consumed in 3h"
```

### Solution 2: Add Maintenance Window Silences

**Problem:** Alerts fire during scheduled deployments.

**Create silences programmatically:**

```bash
#!/bin/bash
# silence-during-deploy.sh

ALERTMANAGER_URL="http://alertmanager:9093"
DURATION="30m"  # Silence for 30 minutes

curl -XPOST "${ALERTMANAGER_URL}/api/v2/silences" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
{
  "matchers": [
    {
      "name": "alertname",
      "value": "ErrorBudgetBurn.*",
      "isRegex": true
    },
    {
      "name": "slo",
      "value": "myapp-availability",
      "isRegex": false
    }
  ],
  "startsAt": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "endsAt": "$(date -u -d '+30 minutes' +%Y-%m-%dT%H:%M:%SZ)",
  "createdBy": "deploy-script",
  "comment": "Deployment in progress"
}
EOF
```

**Integrate into deploy pipeline:**
```yaml
# .github/workflows/deploy.yml
- name: Silence SLO alerts
  run: ./scripts/silence-during-deploy.sh
  
- name: Deploy application
  run: kubectl rollout restart deployment/myapp
  
- name: Wait for rollout
  run: kubectl rollout status deployment/myapp --timeout=10m
```

---

## Issue 3: Latency SLO Always at 100% (No Errors Recorded)

### Symptoms
- Latency SLO shows 100% availability even though you see slow requests
- Histogram buckets exist but Pyrra doesn't calculate errors

### Root Cause
**Latency threshold selector is incorrect** or histogram buckets don't align.

### Diagnosis

```bash
# Check histogram buckets
curl -s 'http://prometheus:9090/api/v1/query?query=http_request_duration_seconds_bucket{job="myapp"}' | \
  jq '.data.result[] | .metric'
```

**Output:**
```json
{
  "__name__": "http_request_duration_seconds_bucket",
  "job": "myapp",
  "le": "0.1"    ← Buckets: 0.1, 0.5, 1, 5, 10
}
```

### Solution: Match Bucket to SLO Target

**Problem:** Your SLO expects requests < 200ms, but histogram doesn't have a 0.2s bucket.

```yaml
# ❌ Wrong - no 0.2 bucket exists
spec:
  indicator:
    latency:
      success:
        metric: http_request_duration_seconds_bucket{le="0.2"}
      total:
        metric: http_request_duration_seconds_count
```

**Fix 1:** Use nearest bucket
```yaml
# ✅ Option 1: Use 0.1s bucket (stricter SLO)
spec:
  indicator:
    latency:
      success:
        metric: http_request_duration_seconds_bucket{le="0.1"}
      total:
        metric: http_request_duration_seconds_count
```

**Fix 2:** Use histogram_quantile for exact threshold
```yaml
# ✅ Option 2: Calculate 0.2s threshold exactly
spec:
  indicator:
    latencyNative:
      total:
        metric: http_request_duration_seconds
      latency: 0.2  # Pyrra calculates this automatically
```

**Fix 3:** Adjust application histogram buckets
```python
# In your application code
from prometheus_client import Histogram

http_request_duration = Histogram(
    'http_request_duration_seconds',
    'HTTP request duration',
    buckets=[0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10]  # ← Add 0.2 bucket
)
```

---

## Issue 4: SLO Not Appearing in Pyrra UI

### Symptoms
- Created SLO YAML but it doesn't show in UI
- No errors in logs
- Recording rules not created in Prometheus

### Root Cause
**Operator not watching the correct namespace** or **CRD not applied**.

### Diagnosis

```bash
# Check if SLO object exists
kubectl get servicelevelobjectives.pyrra.dev -A

# Check Pyrra operator logs
kubectl logs -n monitoring deploy/pyrra-kubernetes -f

# Verify CRD is installed
kubectl get crd servicelevelobjectives.pyrra.dev
```

### Solution 1: Fix Namespace Selector

```yaml
# pyrra-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pyrra-kubernetes
spec:
  template:
    spec:
      containers:
        - name: pyrra
          image: ghcr.io/pyrra-dev/pyrra:latest
          args:
            - kubernetes
            - --prometheus-url=http://prometheus:9090
            - --prometheus-external-url=http://prometheus.example.com
            - --namespace=monitoring  # ← Only watches 'monitoring' namespace
```

**Fix:** Watch all namespaces
```yaml
args:
  - kubernetes
  - --prometheus-url=http://prometheus:9090
  - --prometheus-external-url=http://prometheus.example.com
  # Remove --namespace flag to watch all namespaces
```

Or use label selector:
```yaml
args:
  - kubernetes
  - --prometheus-url=http://prometheus:9090
  - --label-selector=pyrra.dev/enabled=true  # Only watch labeled SLOs
```

### Solution 2: Ensure PrometheusRule Creation

**Problem:** SLO exists but recording rules aren't created.

```bash
# Check if PrometheusRule was created
kubectl get prometheusrule -n monitoring -l pyrra.dev/service-level-objective=myapp-availability
```

**Fix:** Check Pyrra RBAC permissions
```yaml
# pyrra-rbac.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: pyrra-kubernetes
rules:
  - apiGroups: ["monitoring.coreos.com"]
    resources: ["prometheusrules"]
    verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]  # ← Ensure all verbs
  - apiGroups: ["pyrra.dev"]
    resources: ["servicelevelobjectives"]
    verbs: ["get", "list", "watch"]
```

Apply fix:
```bash
kubectl apply -f pyrra-rbac.yaml
kubectl rollout restart -n monitoring deploy/pyrra-kubernetes
```

---

## Issue 5: Error Budget Calculation Wrong

### Symptoms
- Error budget percentage doesn't match expected value
- Budget increases when errors occur (should decrease!)
- Remaining budget is negative

### Root Cause
**Ratio calculation inverted** or **time window misconfigured**.

### Diagnosis

```bash
# Check raw recording rule values
curl -s 'http://prometheus:9090/api/v1/query?query=pyrra_slo_error_budget_remaining_ratio{slo="myapp"}' | jq

# Check availability calculation
curl -s 'http://prometheus:9090/api/v1/query?query=pyrra_slo_availability:5m{slo="myapp"}' | jq
```

### Solution 1: Fix Inverted Ratio

**Problem:** Errors and total are swapped.

```yaml
# ❌ Wrong - total and errors inverted
spec:
  indicator:
    ratio:
      errors:
        metric: http_requests_total{status="200"}  # ← This is success, not error!
      total:
        metric: http_requests_total
```

**Fix:**
```yaml
# ✅ Correct
spec:
  indicator:
    ratio:
      errors:
        metric: http_requests_total{status=~"5.."}  # ← Errors only
      total:
        metric: http_requests_total  # ← All requests
```

### Solution 2: Validate Window Configuration

**Problem:** Window doesn't match retention.

```yaml
# Prometheus retention: 7 days
# But SLO window is 30 days
spec:
  target: "99.9"
  window: 30d  # ❌ Prometheus can't look back this far!
```

**Fix:**
```yaml
spec:
  target: "99.9"
  window: 7d  # ✅ Match Prometheus retention
```

Or increase Prometheus retention:
```yaml
# prometheus.yaml
storage:
  tsdb:
    retention.time: 30d  # Match SLO window
```

---

## Issue 6: High Cardinality Causing Performance Issues

### Symptoms
- Prometheus queries timing out
- Pyrra UI slow to load
- High memory usage on Prometheus

### Root Cause
**Too many unique label combinations** in grouping clause.

### Diagnosis

```bash
# Check cardinality of your metrics
curl -s 'http://prometheus:9090/api/v1/query?query=count(http_requests_total)by(job,instance,method,status,path)' | jq
```

**Output:**
```json
{
  "data": {
    "result": [
      {
        "metric": {},
        "value": [1677649200, "125000"]  ← 125k unique combinations!
      }
    ]
  }
}
```

### Solution: Reduce Grouping Dimensions

**Problem:** Grouping by high-cardinality labels like `path` or `user_id`.

```yaml
# ❌ Wrong - path has thousands of unique values
spec:
  indicator:
    ratio:
      errors:
        metric: http_requests_total{status=~"5.."}
      total:
        metric: http_requests_total
      grouping:
        - job
        - instance
        - method
        - path  # ← /api/users/1, /api/users/2, /api/users/3...
```

**Fix:** Use route templates instead
```yaml
# ✅ Better - route has ~20 unique values
spec:
  indicator:
    ratio:
      errors:
        metric: http_requests_total{status=~"5.."}
      total:
        metric: http_requests_total
      grouping:
        - job
        - method
        - route  # ← /api/users/{id}, /api/orders/{id}
```

**Application-side fix** (relabel paths to routes):

```python
# Flask example
from prometheus_client import Counter
from flask import Flask, request

app = Flask(__name__)

http_requests = Counter(
    'http_requests_total',
    'HTTP requests',
    ['method', 'route', 'status']  # Use route, not path
)

@app.after_request
def track_metrics(response):
    http_requests.labels(
        method=request.method,
        route=request.url_rule.rule if request.url_rule else 'unknown',  # ← Route template
        status=response.status_code
    ).inc()
    return response
```

---

## Issue 7: Filesystem Mode Not Reading SLOs

### Symptoms
- Using `pyrra filesystem` mode
- YAML files exist but SLOs not appearing
- No errors in logs

### Root Cause
**File watcher not triggering** or **invalid YAML format**.

### Diagnosis

```bash
# Check Pyrra logs
docker logs pyrra

# Validate YAML syntax
yamllint /etc/pyrra/*.yaml

# Check file permissions
ls -la /etc/pyrra/
```

### Solution 1: Fix File Permissions

```bash
# Ensure Pyrra can read files
chmod 644 /etc/pyrra/*.yaml
chown -R pyrra:pyrra /etc/pyrra/
```

### Solution 2: Force Reload

```bash
# Restart Pyrra to reload files
docker restart pyrra

# Or send SIGHUP signal
docker exec pyrra kill -HUP 1
```

### Solution 3: Validate SLO YAML Structure

**Problem:** Missing required fields.

```yaml
# ❌ Invalid - missing 'window' and 'target'
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: myapp
spec:
  indicator:
    ratio:
      errors:
        metric: errors_total
      total:
        metric: requests_total
```

**Fix:**
```yaml
# ✅ Valid
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: myapp
  namespace: default
spec:
  target: "99"       # ← Required
  window: 7d         # ← Required
  indicator:
    ratio:
      errors:
        metric: errors_total
      total:
        metric: requests_total
```

---

## Quick Reference: Common Queries

### Check SLO Health
```promql
# Current availability (should be near target)
pyrra_slo_availability:5m{slo="myapp"}

# Remaining error budget (should be > 0)
pyrra_slo_error_budget_remaining_ratio{slo="myapp"}

# Burn rate (should be < 1.0 normally)
pyrra_slo_error_budget_burn_rate{slo="myapp"}
```

### Debug Recording Rules
```bash
# List all Pyrra recording rules
curl -s http://prometheus:9090/api/v1/rules | \
  jq '.data.groups[] | select(.name | contains("pyrra"))'

# Check rule evaluation time
curl -s http://prometheus:9090/api/v1/rules | \
  jq '.data.groups[] | select(.name | contains("pyrra")) | .lastEvaluation'
```

### Validate Metric Existence
```bash
# Check if base metric exists
curl -s 'http://prometheus:9090/api/v1/query?query=up{job="myapp"}' | jq

# Check recording rule output
curl -s 'http://prometheus:9090/api/v1/query?query=pyrra_slo_availability:5m{slo="myapp"}' | jq
```

---

## Architecture Diagram: How Pyrra Processes SLOs

```
┌─────────────────────┐
│  SLO Definition     │
│  (Kubernetes CRD    │
│   or YAML file)     │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Pyrra Operator     │
│  (watches SLOs)     │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Generate           │
│  PrometheusRule     │
│  (4 recording rules)│
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Prometheus         │
│  (evaluates rules)  │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Recording Rules    │
│  Output TSDB        │
│  - availability:5m  │
│  - burn_rate        │
│  - budget_remaining │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Pyrra API          │
│  (queries Prom)     │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│  Pyrra UI           │
│  (displays graphs)  │
└─────────────────────┘
```

---

**Still stuck?** Open an issue on [GitHub](https://github.com/pyrra-dev/pyrra/issues) or join the [discussions](https://github.com/pyrra-dev/pyrra/discussions)!
