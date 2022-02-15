package controllers

import (
	"testing"
	"time"

	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	pyrrav1alpha1 "github.com/pyrra-dev/pyrra/kubernetes/api/v1alpha1"
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
		},
		Spec: pyrrav1alpha1.ServiceLevelObjectiveSpec{
			Target: "99.5",
			Window: model.Duration(28 * 24 * time.Hour),
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
	}{{
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
			},
			Spec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{{
					Name:     "http",
					Interval: "30s",
					Rules: []monitoringv1.Rule{{
						Record: "http_requests:increase4w",
						Expr:   intstr.FromString(`sum by(status) (increase(http_requests_total{job="app"}[4w]))`),
						Labels: map[string]string{"job": "app", "slo": "http"},
					}, {
						Record: "http_requests:burnrate5m",
						Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[5m])) / sum(rate(http_requests_total{job="app"}[5m]))`),
						Labels: map[string]string{"job": "app", "slo": "http"},
					}, {
						Record: "http_requests:burnrate30m",
						Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[30m])) / sum(rate(http_requests_total{job="app"}[30m]))`),
						Labels: map[string]string{"job": "app", "slo": "http"},
					}, {
						Record: "http_requests:burnrate1h",
						Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[1h])) / sum(rate(http_requests_total{job="app"}[1h]))`),
						Labels: map[string]string{"job": "app", "slo": "http"},
					}, {
						Record: "http_requests:burnrate2h",
						Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[2h])) / sum(rate(http_requests_total{job="app"}[2h]))`),
						Labels: map[string]string{"job": "app", "slo": "http"},
					}, {
						Record: "http_requests:burnrate6h",
						Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[6h])) / sum(rate(http_requests_total{job="app"}[6h]))`),
						Labels: map[string]string{"job": "app", "slo": "http"},
					}, {
						Record: "http_requests:burnrate1d",
						Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[1d])) / sum(rate(http_requests_total{job="app"}[1d]))`),
						Labels: map[string]string{"job": "app", "slo": "http"},
					}, {
						Record: "http_requests:burnrate4d",
						Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[4d])) / sum(rate(http_requests_total{job="app"}[4d]))`),
						Labels: map[string]string{"job": "app", "slo": "http"},
					}, {
						Alert:  "ErrorBudgetBurn",
						Expr:   intstr.FromString(`http_requests:burnrate5m{job="app",slo="http"} > (14 * (1-0.995)) and http_requests:burnrate1h{job="app",slo="http"} > (14 * (1-0.995))`),
						For:    "2m",
						Labels: map[string]string{"severity": "critical", "job": "app", "long": "1h", "slo": "http", "short": "5m"},
					}, {
						Alert:  "ErrorBudgetBurn",
						Expr:   intstr.FromString(`http_requests:burnrate30m{job="app",slo="http"} > (7 * (1-0.995)) and http_requests:burnrate6h{job="app",slo="http"} > (7 * (1-0.995))`),
						For:    "15m",
						Labels: map[string]string{"severity": "critical", "job": "app", "long": "6h", "slo": "http", "short": "30m"},
					}, {
						Alert:  "ErrorBudgetBurn",
						Expr:   intstr.FromString(`http_requests:burnrate2h{job="app",slo="http"} > (2 * (1-0.995)) and http_requests:burnrate1d{job="app",slo="http"} > (2 * (1-0.995))`),
						For:    "1h",
						Labels: map[string]string{"severity": "warning", "job": "app", "long": "1d", "slo": "http", "short": "2h"},
					}, {
						Alert:  "ErrorBudgetBurn",
						Expr:   intstr.FromString(`http_requests:burnrate6h{job="app",slo="http"} > (1 * (1-0.995)) and http_requests:burnrate4d{job="app",slo="http"} > (1 * (1-0.995))`),
						For:    "3h",
						Labels: map[string]string{"severity": "warning", "job": "app", "long": "4d", "slo": "http", "short": "6h"},
					}},
				}},
			},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prometheusRule, err := makePrometheusRule(tt.objective)
			require.NoError(t, err)
			require.Equal(t, tt.rules, prometheusRule)
		})
	}
}

func Test_makeConfigMap(t *testing.T) {
	yamlData := `groups:
- interval: 30s
  name: http
  rules:
  - expr: sum by(status) (increase(http_requests_total{job="app"}[4w]))
    labels:
      job: app
      slo: http
    record: http_requests:increase4w
  - expr: sum(rate(http_requests_total{job="app",status=~"5.."}[5m])) / sum(rate(http_requests_total{job="app"}[5m]))
    labels:
      job: app
      slo: http
    record: http_requests:burnrate5m
  - expr: sum(rate(http_requests_total{job="app",status=~"5.."}[30m])) / sum(rate(http_requests_total{job="app"}[30m]))
    labels:
      job: app
      slo: http
    record: http_requests:burnrate30m
  - expr: sum(rate(http_requests_total{job="app",status=~"5.."}[1h])) / sum(rate(http_requests_total{job="app"}[1h]))
    labels:
      job: app
      slo: http
    record: http_requests:burnrate1h
  - expr: sum(rate(http_requests_total{job="app",status=~"5.."}[2h])) / sum(rate(http_requests_total{job="app"}[2h]))
    labels:
      job: app
      slo: http
    record: http_requests:burnrate2h
  - expr: sum(rate(http_requests_total{job="app",status=~"5.."}[6h])) / sum(rate(http_requests_total{job="app"}[6h]))
    labels:
      job: app
      slo: http
    record: http_requests:burnrate6h
  - expr: sum(rate(http_requests_total{job="app",status=~"5.."}[1d])) / sum(rate(http_requests_total{job="app"}[1d]))
    labels:
      job: app
      slo: http
    record: http_requests:burnrate1d
  - expr: sum(rate(http_requests_total{job="app",status=~"5.."}[4d])) / sum(rate(http_requests_total{job="app"}[4d]))
    labels:
      job: app
      slo: http
    record: http_requests:burnrate4d
  - alert: ErrorBudgetBurn
    expr: http_requests:burnrate5m{job="app",slo="http"} > (14 * (1-0.995)) and http_requests:burnrate1h{job="app",slo="http"}
      > (14 * (1-0.995))
    for: 2m
    labels:
      job: app
      long: 1h
      severity: critical
      short: 5m
      slo: http
  - alert: ErrorBudgetBurn
    expr: http_requests:burnrate30m{job="app",slo="http"} > (7 * (1-0.995)) and http_requests:burnrate6h{job="app",slo="http"}
      > (7 * (1-0.995))
    for: 15m
    labels:
      job: app
      long: 6h
      severity: critical
      short: 30m
      slo: http
  - alert: ErrorBudgetBurn
    expr: http_requests:burnrate2h{job="app",slo="http"} > (2 * (1-0.995)) and http_requests:burnrate1d{job="app",slo="http"}
      > (2 * (1-0.995))
    for: 1h
    labels:
      job: app
      long: 1d
      severity: warning
      short: 2h
      slo: http
  - alert: ErrorBudgetBurn
    expr: http_requests:burnrate6h{job="app",slo="http"} > (1 * (1-0.995)) and http_requests:burnrate4d{job="app",slo="http"}
      > (1 * (1-0.995))
    for: 3h
    labels:
      job: app
      long: 4d
      severity: warning
      short: 6h
      slo: http
`

	tests := []struct {
		name          string
		configMapName string
		objective     pyrrav1alpha1.ServiceLevelObjective

		wantErrMsg string
		want       *corev1.ConfigMap
	}{
		{
			name:       "empty input yields error",
			wantErrMsg: "getting SLO",
		},
		{
			name:          "http",
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
				},
				Data: map[string]string{
					"http.rules.yaml": yamlData,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configMap, err := makeConfigMap(tt.configMapName, tt.objective)

			if tt.wantErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.want, configMap)
		})
	}
}
