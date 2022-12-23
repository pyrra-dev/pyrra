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
							Name:     "http-increase",
							Interval: "2m30s",
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
									For:   "2m",
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
							Name:     "http",
							Interval: "30s",
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
									For:         "2m",
									Labels:      map[string]string{"severity": "critical", "job": "app", "long": "1h", "slo": "http", "short": "5m", "team": "foo"},
									Annotations: map[string]string{"description": "foo"},
								},
								{
									Alert:       "ErrorBudgetBurn",
									Expr:        intstr.FromString(`http_requests:burnrate30m{job="app",slo="http"} > (7 * (1-0.995)) and http_requests:burnrate6h{job="app",slo="http"} > (7 * (1-0.995))`),
									For:         "15m",
									Labels:      map[string]string{"severity": "critical", "job": "app", "long": "6h", "slo": "http", "short": "30m", "team": "foo"},
									Annotations: map[string]string{"description": "foo"},
								},
								{
									Alert:       "ErrorBudgetBurn",
									Expr:        intstr.FromString(`http_requests:burnrate2h{job="app",slo="http"} > (2 * (1-0.995)) and http_requests:burnrate1d{job="app",slo="http"} > (2 * (1-0.995))`),
									For:         "1h",
									Labels:      map[string]string{"severity": "warning", "job": "app", "long": "1d", "slo": "http", "short": "2h", "team": "foo"},
									Annotations: map[string]string{"description": "foo"},
								},
								{
									Alert:       "ErrorBudgetBurn",
									Expr:        intstr.FromString(`http_requests:burnrate6h{job="app",slo="http"} > (1 * (1-0.995)) and http_requests:burnrate4d{job="app",slo="http"} > (1 * (1-0.995))`),
									For:         "3h",
									Labels:      map[string]string{"severity": "warning", "job": "app", "long": "4d", "slo": "http", "short": "6h", "team": "foo"},
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
			prometheusRule, err := makePrometheusRule(tt.objective, false)
			require.NoError(t, err)
			require.Equal(t, tt.rules, prometheusRule)
		})
	}
}

func Test_makeConfigMap(t *testing.T) {
	rules := `groups:
- interval: 2m30s
  name: http-increase
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
    for: 2m
    labels:
      job: app
      severity: critical
      slo: http
      team: foo
- interval: 30s
  name: http
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
    for: 2m
    labels:
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
    for: 15m
    labels:
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
    for: 1h
    labels:
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
    for: 3h
    labels:
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
				},
				Data: map[string]string{
					"http.rules.yaml": rules,
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			configMap, err := makeConfigMap(tc.configMapName, tc.objective, false)

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
