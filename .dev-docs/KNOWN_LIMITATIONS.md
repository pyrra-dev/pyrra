# Known Limitations

This document tracks known limitations in Pyrra, including both upstream issues and feature-specific constraints.

---

## Upstream Pyrra Limitations

### 1. Grouping Field Creates Multiple SLO Instances

**Status:** ⚠️ Upstream Pyrra Behavior (Unclear if Bug or Feature)  
**Severity:** Medium  
**Affects:** Both static and dynamic burn rates  
**Discovered:** Task 8.4.1 - Upstream comparison testing

#### Problem Description

When using the `grouping` field in SLO definitions, Pyrra creates multiple SLO instances (one per unique grouping label value) from a single YAML file. This behavior occurs with both regex and simple selectors, and can be confusing for users who expect one YAML to produce one SLO.

#### Example Configuration

```yaml
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: api-requests
spec:
  target: "99"
  window: 4w
  indicator:
    ratio:
      errors:
        metric: prometheus_http_requests_total{handler=~"/api.*",code=~"5.."}
      total:
        metric: prometheus_http_requests_total{handler=~"/api.*"}
      grouping:
        - handler  # ⚠️ Creates multiple SLO instances
```

#### What Happens

1. **Multiple SLOs Created:** One SLO per handler value (e.g., `/api/v1/query`, `/api/v1/query_range`, `/api/v1/notifications/live`)
2. **Recording Rules Per Group:** Rules use `sum by (handler)` to create per-handler metrics
3. **Main Page Shows All:** Each SLO instance appears as separate entry in list
4. **Detail Pages Work:** Each SLO has its own recording rules and shows correct data

#### Confusion Factor

- Users expect: One YAML file = One SLO
- Actual behavior: One YAML file = Multiple SLOs (one per grouping value)
- This can create dozens of SLOs from a single YAML file
- Unclear if this is intentional design or a limitation

#### Recommendations

**If You Want One Aggregated SLO (Recommended)**

```yaml
indicator:
  ratio:
    errors:
      metric: prometheus_http_requests_total{handler=~"/api.*",code=~"5.."}
    total:
      metric: prometheus_http_requests_total{handler=~"/api.*"}
    # ✅ No grouping field - creates single aggregated SLO
```

**Result:** Single SLO with aggregated metrics across all matching handlers.

**If You Want Per-Endpoint Tracking**

Accept that grouping creates multiple SLOs, or create separate YAML files:
- `api-query-slo.yaml` for `/api/v1/query`
- `api-labels-slo.yaml` for `/api/v1/label/__name__/values`
- etc.

**Result:** Explicit control over which endpoints get SLOs.

#### Impact on Dynamic Burn Rate Feature

✅ **No regression introduced** - behavior exists in upstream Pyrra  
✅ **Dynamic burn rates work correctly** with or without grouping  
✅ **Not a blocker** for upstream PR submission

#### When to Use Grouping

**Use Grouping When:**
- You want per-endpoint SLO tracking
- You're okay with multiple SLO instances
- You need separate alerts per endpoint

**Avoid Grouping When:**
- You want service-level aggregated metrics
- You want one YAML to produce one SLO
- You have many label values (creates many SLOs)

#### References

- **Investigation:** `.dev-docs/ISSUE_REGEX_LABEL_SELECTORS.md`
- **Upstream Comparison:** `.dev-docs/UPSTREAM_COMPARISON_REGEX_SELECTORS.md`
- **Task:** Task 8.4.1 in `.kiro/specs/dynamic-burn-rate-completion/tasks.md`

---

### 2. NaN in Burn Rate Columns When No Errors

**Status:** 🔴 Upstream Pyrra UI Bug  
**Severity:** Low (Cosmetic)  
**Affects:** All SLOs (static and dynamic) when there are no errors  
**Discovered:** Task 8.4.1 - Upstream comparison testing

#### Problem Description

When an SLO has no errors in the measurement window, the alert table on the detail page shows "NaN" in the Short Burn and Long Burn columns instead of showing "0" or "0%".

#### Root Cause

**Technical Details:**
```promql
# When there are no errors, the error metric doesn't exist
sum(rate(prometheus_http_requests_total{code=~"5..",handler=~"/api.*"}[3m]))
→ Returns: [] (empty, not 0)

# Division with empty numerator returns empty
[] / sum(rate(prometheus_http_requests_total{handler=~"/api.*"}[3m]))
→ Returns: [] (empty, not 0)

# UI interprets empty as NaN
```

Prometheus doesn't return `0` for metrics that don't exist - it returns empty results. When the UI tries to display an empty result, it shows "NaN".

#### What Works Correctly

✅ **Availability/Budget Tiles:** Show correct 100% values (use different calculation)  
✅ **Main Page List:** Shows correct values (uses increase recording rules)  
✅ **Alert Firing:** Works correctly (empty < threshold = no alert)

#### What Shows NaN

❌ **Detail Page Alert Table:**
- Short Burn column: NaN
- Long Burn column: NaN
- Only when there are no errors in the time window

#### Impact

- **Functional:** None - alerts work correctly, calculations are correct
- **User Experience:** Confusing to see NaN instead of 0
- **Affects:** ALL SLOs (not specific to regex selectors or dynamic burn rates)

#### Workaround

None needed - this is cosmetic only. The feature branch likely fixes this by showing "0" or handling empty results better.

#### Fix Status

- **Upstream:** Not fixed (as of testing date)
- **Feature Branch:** Likely fixed (needs verification)
- **Recommendation:** Include fix in upstream PR

#### References

- **Upstream Comparison:** `.dev-docs/UPSTREAM_COMPARISON_REGEX_SELECTORS.md`
- **Task:** Task 8.4.1 in `.kiro/specs/dynamic-burn-rate-completion/tasks.md`

---

## Dynamic Burn Rate Feature Limitations

### 1. Requires Traffic Data for Threshold Calculation

**Status:** ⚠️ By Design  
**Severity:** Low  
**Affects:** Dynamic burn rate alerts only

#### Description

Dynamic burn rate alerts require actual traffic data to calculate adaptive thresholds. If no traffic data exists in the alert window, thresholds cannot be calculated.

#### Behavior

- **With Traffic:** Dynamic thresholds adapt based on actual request volume
- **Without Traffic:** Falls back to conservative behavior (no false alerts)
- **Low Traffic:** Thresholds scale proportionally to traffic volume

#### Workaround

Use static burn rates for SLOs with very low or intermittent traffic patterns.

---

## General Pyrra Limitations

### 1. Recording Rule Cardinality

**Status:** ⚠️ By Design  
**Severity:** Medium

#### Description

Pyrra creates recording rules for each SLO, which can increase Prometheus cardinality. Using grouping fields multiplies the number of time series.

#### Recommendation

- Limit grouping to essential labels only
- Monitor Prometheus cardinality metrics
- Use aggregated SLOs when possible (no grouping)

### 2. Alert Evaluation Interval

**Status:** ⚠️ By Design  
**Severity:** Low

#### Description

Alerts are evaluated at Prometheus's configured interval (typically 30s). Fast-burning error budget may not trigger alerts immediately.

#### Recommendation

- Use appropriate SLO windows for your use case
- Shorter windows = faster alert detection
- Balance between alert sensitivity and noise

---

## Documentation Updates

**Last Updated:** 2025-10-14  
**Next Review:** Before upstream PR submission

### Related Documents

- `.dev-docs/UPSTREAM_COMPARISON_REGEX_SELECTORS.md` - Detailed regex selector investigation
- `.dev-docs/ISSUE_REGEX_LABEL_SELECTORS.md` - Original issue discovery
- `.dev-docs/TESTING_ENVIRONMENT_REFERENCE.md` - Testing procedures
- `pyrra-development-standards.md` - Development standards and patterns
