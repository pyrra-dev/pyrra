package controllers

import (
	"testing"
	"time"

	athenev1alpha1 "github.com/metalmatze/athene/api/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
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
				Window: metav1.Duration{Duration: 28 * 24 * time.Hour},
				ServiceLevelIndicator: athenev1alpha1.ServiceLevelIndicator{
					HTTP: &athenev1alpha1.HTTPIndicator{
						Matchers:      []string{`job="app"`},
						ErrorMatchers: []string{`status=~"5.."`},
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
					}, {
						Record: "http_requests:burnrate30m",
						Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[30m])) / sum(rate(http_requests_total{job="app"}[30m]))`),
					}, {
						Record: "http_requests:burnrate1h",
						Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[1h])) / sum(rate(http_requests_total{job="app"}[1h]))`),
					}, {
						Record: "http_requests:burnrate2h",
						Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[2h])) / sum(rate(http_requests_total{job="app"}[2h]))`),
					}, {
						Record: "http_requests:burnrate6h",
						Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[6h])) / sum(rate(http_requests_total{job="app"}[6h]))`),
					}, {
						Record: "http_requests:burnrate1d",
						Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[1d])) / sum(rate(http_requests_total{job="app"}[1d]))`),
					}, {
						Record: "http_requests:burnrate4d",
						Expr:   intstr.FromString(`sum(rate(http_requests_total{job="app",status=~"5.."}[4d])) / sum(rate(http_requests_total{job="app"}[4d]))`),
					}, {
						Alert:       "ErrorBudgetBurn",
						Expr:        intstr.FromString(`http_requests:burnrate5m > (14 * (1-0.99500)) and http_requests:burnrate1h > (14 * (1-0.99500))`),
						For:         "2m",
						Annotations: map[string]string{"severity": "critical"},
					}, {
						Alert:       "ErrorBudgetBurn",
						Expr:        intstr.FromString(`http_requests:burnrate30m > (7 * (1-0.99500)) and http_requests:burnrate6h > (7 * (1-0.99500))`),
						For:         "15m",
						Annotations: map[string]string{"severity": "critical"},
					}, {
						Alert:       "ErrorBudgetBurn",
						Expr:        intstr.FromString(`http_requests:burnrate2h > (2 * (1-0.99500)) and http_requests:burnrate1d > (2 * (1-0.99500))`),
						For:         "1h",
						Annotations: map[string]string{"severity": "warning"},
					}, {
						Alert:       "ErrorBudgetBurn",
						Expr:        intstr.FromString(`http_requests:burnrate6h > (1 * (1-0.99500)) and http_requests:burnrate4d > (1 * (1-0.99500))`),
						For:         "3h",
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
