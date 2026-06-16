# Performance Mode

By default, Pyrra generates recording rules that compute increases over the full SLO window, for example `increase(metric[4w])`. This is accurate but expensive. With a long window and high-cardinality metrics, Prometheus has to scan weeks of raw samples on every evaluation. In Thanos setups, this means large amounts of raw data transferred across the wire on every query.

The Wikimedia Foundation ran into this with a 12-week SLO over high-cardinality Istio metrics. Their `increase(metric[12w])` rules took up to 95 seconds to evaluate, approaching their 120-second timeout. See [#1440](https://github.com/pyrra-dev/pyrra/issues/1440) for the full context. A detailed write-up of the problem and this solution is available on the [Polar Signals blog](https://www.polarsignals.com/blog/posts/2025/12/30/optimizing-slo-performance-at-scale-how-we-solved-pyrra-s-query-performance-on-thanos).

Enable performance mode if you're seeing slow rule evaluation or high memory usage, particularly with windows longer than 7 days or high-cardinality metrics.

> **Note:** If your SLO target is above 99% (less than 1% error budget), think carefully before enabling this. The mode introduces a small inaccuracy, typically less than 1%. Whether that trade-off is acceptable depends on your system.

## How It Works

When `performanceOverAccuracy: true` is set, Pyrra generates two PrometheusRule resources instead of one.

The first rule, named `{slo-name}-short`, records short 5-minute increases at a 30-second interval together with the burnrate recording rules and alerts. The 5-minute increases are cheap for Prometheus to evaluate since they only scan 5 minutes of data at a time, and the alerts evaluate frequently against them.

The second rule, named `{slo-name}-long`, uses a subquery to sum those 5-minute values across the full window, for example `sum_over_time(metric:increase5m[4w:5m])`. Its evaluation interval is computed from the window length, typically 2 to 3 minutes, and is not user-configurable. This rule is well-suited for ThanosRuler, which handles the full-window aggregation without touching raw samples.

In Thanos setups this reduces the data transferred across the wire by around 20x, since queries consume pre-aggregated 5-minute values rather than raw samples. This figure comes from real-world measurements described in the blog post linked above.

## Configuration

Add `performanceOverAccuracy: true` to your SLO spec:

```yaml
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: pyrra-api-errors
  namespace: monitoring
  labels:
    prometheus: k8s
    role: alert-rules
spec:
  target: "99"
  window: 4w
  indicator:
    ratio:
      errors:
        metric: http_requests_total{job="pyrra",code=~"5.."}
      total:
        metric: http_requests_total{job="pyrra"}
  performanceOverAccuracy: true
```

This works for all SLI types (ratio, latency, latencyNative, boolGauge) and with both the Kubernetes operator and the filesystem mode. In filesystem mode `ruleOutput` has no effect, since there is no Prometheus Operator to route rules based on labels.

## Routing Rules to Different Instances

If you run Prometheus and ThanosRuler in the same cluster, you'll want each PrometheusRule picked up by the right instance. The [Prometheus Operator](https://prometheus-operator.dev) uses label selectors for this — check the `ruleSelector` field on your `Prometheus` and `ThanosRuler` objects to see which labels each instance watches. Then use `ruleOutput` to set matching labels on each generated rule.

```yaml
spec:
  performanceOverAccuracy: true
  ruleOutput:
    shortRulesLabels:
      prometheus: k8s
    longRulesLabels:
      prometheus: thanos-k8s
```

The short increase rules will be picked up by your in-cluster Prometheus. The long subquery rules will be picked up by ThanosRuler, which queries the pre-aggregated values through Thanos Querier.

If you omit `ruleOutput`, both rules inherit the labels from the SLO object itself. You still get the Prometheus-side performance benefit, though without the explicit routing.
