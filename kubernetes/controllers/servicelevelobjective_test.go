package controllers

import (
	"testing"
	"time"

	athenev1alpha1 "github.com/metalmatze/athene/kubernetes/api/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Test_makePrometheusRule(t *testing.T) {
	tests := []struct {
		name      string
		objective athenev1alpha1.ServiceLevelObjective
		rules     *monitoringv1.PrometheusRule
	}{{
		name: "http",
		objective: athenev1alpha1.ServiceLevelObjective{
			ObjectMeta: metav1.ObjectMeta{Name: "http"},
			Spec: athenev1alpha1.ServiceLevelObjectiveSpec{
				Target: "99.5",
				Window: model.Duration(28 * 24 * time.Hour),
				ServiceLevelIndicator: athenev1alpha1.ServiceLevelIndicator{
					Ratio: &athenev1alpha1.RatioIndicator{
						Errors: athenev1alpha1.Query{
							Metric: `http_requests_total{job="app",status=~"5.."}`,
						},
						Total: athenev1alpha1.Query{
							Metric: `http_requests_total{job="app"}`,
						},
					},
				},
			},
		},
		rules: &monitoringv1.PrometheusRule{
			TypeMeta: metav1.TypeMeta{
				APIVersion: monitoring.GroupName + "/" + monitoringv1.Version,
				Kind:       monitoringv1.PrometheusRuleKind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "http",
			},
			Spec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{{
					Name:     "http",
					Interval: "30s",
					Rules: []monitoringv1.Rule{{
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
						Alert:       "ErrorBudgetBurn",
						Expr:        intstr.FromString(`http_requests:burnrate5m{job="app",slo="http"} > (14 * (1-0.995)) and http_requests:burnrate1h{job="app",slo="http"} > (14 * (1-0.995))`),
						For:         "2m",
						Labels:      map[string]string{"job": "app", "long": "1h", "slo": "http", "short": "5m"},
						Annotations: map[string]string{"severity": "critical"},
					}, {
						Alert:       "ErrorBudgetBurn",
						Expr:        intstr.FromString(`http_requests:burnrate30m{job="app",slo="http"} > (7 * (1-0.995)) and http_requests:burnrate6h{job="app",slo="http"} > (7 * (1-0.995))`),
						For:         "15m",
						Labels:      map[string]string{"job": "app", "long": "6h", "slo": "http", "short": "30m"},
						Annotations: map[string]string{"severity": "critical"},
					}, {
						Alert:       "ErrorBudgetBurn",
						Expr:        intstr.FromString(`http_requests:burnrate2h{job="app",slo="http"} > (2 * (1-0.995)) and http_requests:burnrate1d{job="app",slo="http"} > (2 * (1-0.995))`),
						For:         "1h",
						Labels:      map[string]string{"job": "app", "long": "1d", "slo": "http", "short": "2h"},
						Annotations: map[string]string{"severity": "warning"},
					}, {
						Alert:       "ErrorBudgetBurn",
						Expr:        intstr.FromString(`http_requests:burnrate6h{job="app",slo="http"} > (1 * (1-0.995)) and http_requests:burnrate4d{job="app",slo="http"} > (1 * (1-0.995))`),
						For:         "3h",
						Labels:      map[string]string{"job": "app", "long": "4d", "slo": "http", "short": "6h"},
						Annotations: map[string]string{"severity": "warning"},
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
