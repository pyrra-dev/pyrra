package controllers

import (
	"fmt"
	"testing"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	pyrrav1alpha1 "github.com/pyrra-dev/pyrra/kubernetes/api/v1alpha1"
	"github.com/pyrra-dev/pyrra/slo"
)

var (
	trueBool = true
	httpSLO  = pyrrav1alpha1.ServiceLevelObjective{
		TypeMeta: metav1.TypeMeta{
			APIVersion: pyrrav1alpha1.GroupVersion.Version,
			Kind:       "ServiceLevelObjective",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "http",
			UID:  "123",
			Labels: map[string]string{
				slo.PropagationLabelsPrefix + "team": "foo",
				"team":                               "bar",
			},
			Annotations: map[string]string{
				slo.PropagationLabelsPrefix + "description": "foo",
				"description": "bar",
			},
		},
		Spec: pyrrav1alpha1.ServiceLevelObjectiveSpec{
			Target: "99.5",
			Window: "28d",
			ServiceLevelIndicator: pyrrav1alpha1.ServiceLevelIndicator{
				Ratio: &pyrrav1alpha1.RatioIndicator{
					Errors: pyrrav1alpha1.Query{
						Metric: `http_requests_total{job="app",status=~"5.."}`,
					},
					Total: pyrrav1alpha1.Query{
						Metric: `http_requests_total{job="app"}`,
					},
				},
			},
			PartialResponseStrategy: "warn",
		},
	}
)

func Test_makePrometheusRule(t *testing.T) {
	tests := []struct {
		name      string
		objective pyrrav1alpha1.ServiceLevelObjective
		rules     *monitoringv1.PrometheusRule
	}{
		{
			name:      "http",
			objective: httpSLO,
			rules: &monitoringv1.PrometheusRule{
				TypeMeta: metav1.TypeMeta{
					APIVersion: monitoring.GroupName + "/" + monitoringv1.Version,
					Kind:       monitoringv1.PrometheusRuleKind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "http",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: pyrrav1alpha1.GroupVersion.Version,
							Kind:       "ServiceLevelObjective",
							Name:       "http",
							UID:        "123",
							Controller: &trueBool,
						},
					},
					Labels: map[string]string{
						"pyrra.dev/team": "foo",
						"team":           "bar",
					},
				},
				Spec: monitoringv1.PrometheusRuleSpec{
					Groups: []monitoringv1.RuleGroup{
						{
							Name:                    "http-increase",
							PartialResponseStrategy: "warn",
							Interval:                monitoringDuration("2m30s"),
							Rules: []monitoringv1.Rule{
								{
									Record: "http_requests:increase4w",
									Expr:   intstr.FromString(`sum by (status) (increase(http_requests_total{job="app"}[4w]))`),
									Labels: map[string]string{
										"job":  "app",
										"slo":  "http",
										"team": "foo",
									},
								},
								{
									Alert: "SLOMetricAbsent",
									Expr:  intstr.FromString(`absent(http_requests_total{job="app"}) == 1`),
									For:   monitoringDuration("6m"),
									Annotations: map[string]string{
										"description": "foo",
									},
									Labels: map[string]string{
										"severity": "critical",
										"job":      "app",
										"slo":      "http",
										"team":     "foo",
									},
								},
							},
						},
						{
							Name:                    "http",
							PartialResponseStrategy: "warn",
							Interval:                monitoringDuration("30s"),
							Rules: []monitoringv1.Rule{
								{
									Record: "http_requests:burnrate5m",
									Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[5m])) / sum(rate(http_requests_total{job="app"}[5m]))`),
									Labels: map[string]string{"job": "app", "slo": "http", "team": "foo"},
								},
								{
									Record: "http_requests:burnrate30m",
									Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[30m])) / sum(rate(http_requests_total{job="app"}[30m]))`),
									Labels: map[string]string{"job": "app", "slo": "http", "team": "foo"},
								},
								{
									Record: "http_requests:burnrate1h",
									Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[1h])) / sum(rate(http_requests_total{job="app"}[1h]))`),
									Labels: map[string]string{"job": "app", "slo": "http", "team": "foo"},
								},
								{
									Record: "http_requests:burnrate2h",
									Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[2h])) / sum(rate(http_requests_total{job="app"}[2h]))`),
									Labels: map[string]string{"job": "app", "slo": "http", "team": "foo"},
								},
								{
									Record: "http_requests:burnrate6h",
									Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[6h])) / sum(rate(http_requests_total{job="app"}[6h]))`),
									Labels: map[string]string{"job": "app", "slo": "http", "team": "foo"},
								},
								{
									Record: "http_requests:burnrate1d",
									Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[1d])) / sum(rate(http_requests_total{job="app"}[1d]))`),
									Labels: map[string]string{"job": "app", "slo": "http", "team": "foo"},
								},
								{
									Record: "http_requests:burnrate4d",
									Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[4d])) / sum(rate(http_requests_total{job="app"}[4d]))`),
									Labels: map[string]string{"job": "app", "slo": "http", "team": "foo"},
								},
								{
									Alert:       "ErrorBudgetBurn",
									Expr:        intstr.FromString(`http_requests:burnrate5m{job="app",slo="http"} > (14 * (1-0.995)) and http_requests:burnrate1h{job="app",slo="http"} > (14 * (1-0.995))`),
									For:         monitoringDuration("2m0s"),
									Labels:      map[string]string{"severity": "critical", "job": "app", "long": "1h", "slo": "http", "short": "5m", "team": "foo", "exhaustion": "2d"},
									Annotations: map[string]string{"description": "foo"},
								},
								{
									Alert:       "ErrorBudgetBurn",
									Expr:        intstr.FromString(`http_requests:burnrate30m{job="app",slo="http"} > (7 * (1-0.995)) and http_requests:burnrate6h{job="app",slo="http"} > (7 * (1-0.995))`),
									For:         monitoringDuration("15m0s"),
									Labels:      map[string]string{"severity": "critical", "job": "app", "long": "6h", "slo": "http", "short": "30m", "team": "foo", "exhaustion": "4d"},
									Annotations: map[string]string{"description": "foo"},
								},
								{
									Alert:       "ErrorBudgetBurn",
									Expr:        intstr.FromString(`http_requests:burnrate2h{job="app",slo="http"} > (2 * (1-0.995)) and http_requests:burnrate1d{job="app",slo="http"} > (2 * (1-0.995))`),
									For:         monitoringDuration("1h0m0s"),
									Labels:      map[string]string{"severity": "warning", "job": "app", "long": "1d", "slo": "http", "short": "2h", "team": "foo", "exhaustion": "2w"},
									Annotations: map[string]string{"description": "foo"},
								},
								{
									Alert:       "ErrorBudgetBurn",
									Expr:        intstr.FromString(`http_requests:burnrate6h{job="app",slo="http"} > (1 * (1-0.995)) and http_requests:burnrate4d{job="app",slo="http"} > (1 * (1-0.995))`),
									For:         monitoringDuration("3h0m0s"),
									Labels:      map[string]string{"severity": "warning", "job": "app", "long": "4d", "slo": "http", "short": "6h", "team": "foo", "exhaustion": "4w"},
									Annotations: map[string]string{"description": "foo"},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prometheusRule, err := makePrometheusRule(tt.objective, false, false, "")
			require.NoError(t, err)
			require.Equal(t, tt.rules, prometheusRule)
		})
	}
}

func Test_makeSplitPrometheusRules(t *testing.T) {
	perfSLO := httpSLO.DeepCopy()
	perfSLO.Spec.PerformanceOverAccuracy = true
	perfSLO.Spec.RuleOutput = &pyrrav1alpha1.RuleOutput{
		ShortRulesLabels: map[string]string{"prometheus": "k8s"},
		LongRulesLabels:  map[string]string{"prometheus": "thanos-k8s"},
	}

	shortRule, longRule, err := makeSplitPrometheusRules(*perfSLO, false, false, "")
	require.NoError(t, err)

	// Short rule should have name "{name}-increase"
	require.Equal(t, "http-increase", shortRule.Name)
	// Short rule should have merged labels with ShortRulesLabels
	require.Equal(t, "k8s", shortRule.Labels["prometheus"])
	require.Equal(t, "foo", shortRule.Labels[slo.PropagationLabelsPrefix+"team"])
	// Short rule should have owner reference
	require.Len(t, shortRule.OwnerReferences, 1)
	require.Equal(t, "http", shortRule.OwnerReferences[0].Name)
	// Short rule should have 1 group with short rules
	require.Len(t, shortRule.Spec.Groups, 1)

	// Long rule should keep original name
	require.Equal(t, "http", longRule.Name)
	// Long rule should have merged labels with LongRulesLabels
	require.Equal(t, "thanos-k8s", longRule.Labels["prometheus"])
	require.Equal(t, "foo", longRule.Labels[slo.PropagationLabelsPrefix+"team"])
	// Long rule should have owner reference
	require.Len(t, longRule.OwnerReferences, 1)
	// Long rule should have 2 groups (increase + burnrates)
	require.Len(t, longRule.Spec.Groups, 2)

	// Both should have partial response strategy
	require.Equal(t, "warn", shortRule.Spec.Groups[0].PartialResponseStrategy)
	require.Equal(t, "warn", longRule.Spec.Groups[0].PartialResponseStrategy)
}

func Test_makeSplitPrometheusRulesWithoutRuleOutput(t *testing.T) {
	perfSLO := httpSLO.DeepCopy()
	perfSLO.Spec.PerformanceOverAccuracy = true
	// No RuleOutput set — both should inherit SLO labels

	shortRule, longRule, err := makeSplitPrometheusRules(*perfSLO, false, false, "")
	require.NoError(t, err)

	// Both should have the SLO's labels
	require.Equal(t, "foo", shortRule.Labels[slo.PropagationLabelsPrefix+"team"])
	require.Equal(t, "bar", shortRule.Labels["team"])
	require.Equal(t, "foo", longRule.Labels[slo.PropagationLabelsPrefix+"team"])
	require.Equal(t, "bar", longRule.Labels["team"])
}

func Test_mergeLabels(t *testing.T) {
	base := map[string]string{"a": "1", "b": "2"}
	override := map[string]string{"b": "3", "c": "4"}
	result := mergeLabels(base, override)
	require.Equal(t, map[string]string{"a": "1", "b": "3", "c": "4"}, result)

	// nil override should preserve base
	result = mergeLabels(base, nil)
	require.Equal(t, map[string]string{"a": "1", "b": "2"}, result)
}

func Test_makeConfigMap(t *testing.T) {
	rules := `groups:
- interval: 2m30s
  name: http-increase
  partial_response_strategy: warn
  rules:
  - expr: sum by (status) (increase(http_requests_total{job="app"}[4w]))
    labels:
      job: app
      slo: http
      team: foo
    record: http_requests:increase4w
  - alert: SLOMetricAbsent
    annotations:
      description: foo
    expr: absent(http_requests_total{job="app"}) == 1
    for: 6m
    labels:
      job: app
      severity: critical
      slo: http
      team: foo
- interval: 30s
  name: http
  partial_response_strategy: warn
  rules:
  - expr: sum(rate(http_requests_total{job="app",status=~"5.."}[5m])) / sum(rate(http_requests_total{job="app"}[5m]))
    labels:
      job: app
      slo: http
      team: foo
    record: http_requests:burnrate5m
  - expr: sum(rate(http_requests_total{job="app",status=~"5.."}[30m])) / sum(rate(http_requests_total{job="app"}[30m]))
    labels:
      job: app
      slo: http
      team: foo
    record: http_requests:burnrate30m
  - expr: sum(rate(http_requests_total{job="app",status=~"5.."}[1h])) / sum(rate(http_requests_total{job="app"}[1h]))
    labels:
      job: app
      slo: http
      team: foo
    record: http_requests:burnrate1h
  - expr: sum(rate(http_requests_total{job="app",status=~"5.."}[2h])) / sum(rate(http_requests_total{job="app"}[2h]))
    labels:
      job: app
      slo: http
      team: foo
    record: http_requests:burnrate2h
  - expr: sum(rate(http_requests_total{job="app",status=~"5.."}[6h])) / sum(rate(http_requests_total{job="app"}[6h]))
    labels:
      job: app
      slo: http
      team: foo
    record: http_requests:burnrate6h
  - expr: sum(rate(http_requests_total{job="app",status=~"5.."}[1d])) / sum(rate(http_requests_total{job="app"}[1d]))
    labels:
      job: app
      slo: http
      team: foo
    record: http_requests:burnrate1d
  - expr: sum(rate(http_requests_total{job="app",status=~"5.."}[4d])) / sum(rate(http_requests_total{job="app"}[4d]))
    labels:
      job: app
      slo: http
      team: foo
    record: http_requests:burnrate4d
  - alert: ErrorBudgetBurn
    annotations:
      description: foo
    expr: http_requests:burnrate5m{job="app",slo="http"} > (14 * (1-0.995)) and http_requests:burnrate1h{job="app",slo="http"}
      > (14 * (1-0.995))
    for: 2m0s
    labels:
      exhaustion: 2d
      job: app
      long: 1h
      severity: critical
      short: 5m
      slo: http
      team: foo
  - alert: ErrorBudgetBurn
    annotations:
      description: foo
    expr: http_requests:burnrate30m{job="app",slo="http"} > (7 * (1-0.995)) and http_requests:burnrate6h{job="app",slo="http"}
      > (7 * (1-0.995))
    for: 15m0s
    labels:
      exhaustion: 4d
      job: app
      long: 6h
      severity: critical
      short: 30m
      slo: http
      team: foo
  - alert: ErrorBudgetBurn
    annotations:
      description: foo
    expr: http_requests:burnrate2h{job="app",slo="http"} > (2 * (1-0.995)) and http_requests:burnrate1d{job="app",slo="http"}
      > (2 * (1-0.995))
    for: 1h0m0s
    labels:
      exhaustion: 2w
      job: app
      long: 1d
      severity: warning
      short: 2h
      slo: http
      team: foo
  - alert: ErrorBudgetBurn
    annotations:
      description: foo
    expr: http_requests:burnrate6h{job="app",slo="http"} > (1 * (1-0.995)) and http_requests:burnrate4d{job="app",slo="http"}
      > (1 * (1-0.995))
    for: 3h0m0s
    labels:
      exhaustion: 4w
      job: app
      long: 4d
      severity: warning
      short: 6h
      slo: http
      team: foo
`

	testcases := []struct {
		name          string
		configMapName string
		objective     pyrrav1alpha1.ServiceLevelObjective

		want *corev1.ConfigMap
		err  error
	}{
		{
			name: "NoInput",
			err:  fmt.Errorf("failed to get objective"),
		},
		{
			name:          "HTTP",
			configMapName: "http",
			objective:     httpSLO,
			want: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "http",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: pyrrav1alpha1.GroupVersion.Version,
							Kind:       "ServiceLevelObjective",
							Name:       "http",
							UID:        "123",
							Controller: &trueBool,
						},
					},
					Labels: map[string]string{
						slo.PropagationLabelsPrefix + "team": "foo",
						"team":                               "bar",
					},
					Annotations: map[string]string{
						slo.PropagationLabelsPrefix + "description": "foo",
						"description": "bar",
					},
				},
				Data: map[string]string{
					"http.rules.yaml": rules,
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			configMap, err := makeConfigMap(tc.configMapName, tc.objective, false, false, "")

			if tc.err != nil {
				require.Error(t, err)
				require.ErrorAs(t, err, &tc.err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.want, configMap)
		})
	}
}

func monitoringDuration(d string) *monitoringv1.Duration {
	md := monitoringv1.Duration(d)
	return &md
}
