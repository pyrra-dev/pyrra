package slo

import (
	"testing"
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestObjective_Burnrates(t *testing.T) {
	testcases := []struct {
		name  string
		slo   Objective
		rules monitoringv1.RuleGroup
	}{{
		name: "http-ratio", // super similar to gRPC and therefore only HTTP
		slo:  objectiveHTTPRatio(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-errors",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_requests:burnrate5m",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[5m])) / sum(rate(http_requests_total{job="thanos-receive-default"}[5m]))`,
				},
				Labels: map[string]string{"job": "thanos-receive-default", "slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate30m",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[30m])) / sum(rate(http_requests_total{job="thanos-receive-default"}[30m]))`,
				},
				Labels: map[string]string{"job": "thanos-receive-default", "slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate1h",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[1h])) / sum(rate(http_requests_total{job="thanos-receive-default"}[1h]))`,
				},
				Labels: map[string]string{"job": "thanos-receive-default", "slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate2h",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[2h])) / sum(rate(http_requests_total{job="thanos-receive-default"}[2h]))`,
				},
				Labels: map[string]string{"job": "thanos-receive-default", "slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate6h",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[6h])) / sum(rate(http_requests_total{job="thanos-receive-default"}[6h]))`,
				},
				Labels: map[string]string{"job": "thanos-receive-default", "slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate1d",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[1d])) / sum(rate(http_requests_total{job="thanos-receive-default"}[1d]))`,
				},
				Labels: map[string]string{"job": "thanos-receive-default", "slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate4d",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[4d])) / sum(rate(http_requests_total{job="thanos-receive-default"}[4d]))`,
				},
				Labels: map[string]string{"job": "thanos-receive-default", "slo": "monitoring-http-errors"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "2m",
				Expr:        intstr.FromString(`http_requests:burnrate5m{job="thanos-receive-default",slo="monitoring-http-errors"} > (14 * (1-0.99)) and http_requests:burnrate1h{job="thanos-receive-default",slo="monitoring-http-errors"} > (14 * (1-0.99))`),
				Annotations: map[string]string{"severity": "critical"},
				Labels:      map[string]string{"job": "thanos-receive-default", "long": "1h", "slo": "monitoring-http-errors", "short": "5m"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "15m",
				Expr:        intstr.FromString(`http_requests:burnrate30m{job="thanos-receive-default",slo="monitoring-http-errors"} > (7 * (1-0.99)) and http_requests:burnrate6h{job="thanos-receive-default",slo="monitoring-http-errors"} > (7 * (1-0.99))`),
				Annotations: map[string]string{"severity": "critical"},
				Labels:      map[string]string{"job": "thanos-receive-default", "long": "6h", "slo": "monitoring-http-errors", "short": "30m"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "1h",
				Expr:        intstr.FromString(`http_requests:burnrate2h{job="thanos-receive-default",slo="monitoring-http-errors"} > (2 * (1-0.99)) and http_requests:burnrate1d{job="thanos-receive-default",slo="monitoring-http-errors"} > (2 * (1-0.99))`),
				Annotations: map[string]string{"severity": "warning"},
				Labels:      map[string]string{"job": "thanos-receive-default", "long": "1d", "slo": "monitoring-http-errors", "short": "2h"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "3h",
				Expr:        intstr.FromString(`http_requests:burnrate6h{job="thanos-receive-default",slo="monitoring-http-errors"} > (1 * (1-0.99)) and http_requests:burnrate4d{job="thanos-receive-default",slo="monitoring-http-errors"} > (1 * (1-0.99))`),
				Annotations: map[string]string{"severity": "warning"},
				Labels:      map[string]string{"job": "thanos-receive-default", "long": "4d", "slo": "monitoring-http-errors", "short": "6h"},
			}},
		},
	}, {
		name: "http-ratio-grouping",
		slo:  objectiveHTTPRatioGrouping(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-errors",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_requests:burnrate5m",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[5m])) / sum by(job, handler) (rate(http_requests_total{job="thanos-receive-default"}[5m]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate30m",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[30m])) / sum by(job, handler) (rate(http_requests_total{job="thanos-receive-default"}[30m]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate1h",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[1h])) / sum by(job, handler) (rate(http_requests_total{job="thanos-receive-default"}[1h]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate2h",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[2h])) / sum by(job, handler) (rate(http_requests_total{job="thanos-receive-default"}[2h]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate6h",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[6h])) / sum by(job, handler) (rate(http_requests_total{job="thanos-receive-default"}[6h]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate1d",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[1d])) / sum by(job, handler) (rate(http_requests_total{job="thanos-receive-default"}[1d]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate4d",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[4d])) / sum by(job, handler) (rate(http_requests_total{job="thanos-receive-default"}[4d]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "2m",
				Expr:        intstr.FromString(`http_requests:burnrate5m{job="thanos-receive-default",slo="monitoring-http-errors"} > (14 * (1-0.99)) and http_requests:burnrate1h{job="thanos-receive-default",slo="monitoring-http-errors"} > (14 * (1-0.99))`),
				Annotations: map[string]string{"severity": "critical"},
				Labels:      map[string]string{"long": "1h", "slo": "monitoring-http-errors", "short": "5m"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "15m",
				Expr:        intstr.FromString(`http_requests:burnrate30m{job="thanos-receive-default",slo="monitoring-http-errors"} > (7 * (1-0.99)) and http_requests:burnrate6h{job="thanos-receive-default",slo="monitoring-http-errors"} > (7 * (1-0.99))`),
				Annotations: map[string]string{"severity": "critical"},
				Labels:      map[string]string{"long": "6h", "slo": "monitoring-http-errors", "short": "30m"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "1h",
				Expr:        intstr.FromString(`http_requests:burnrate2h{job="thanos-receive-default",slo="monitoring-http-errors"} > (2 * (1-0.99)) and http_requests:burnrate1d{job="thanos-receive-default",slo="monitoring-http-errors"} > (2 * (1-0.99))`),
				Annotations: map[string]string{"severity": "warning"},
				Labels:      map[string]string{"long": "1d", "slo": "monitoring-http-errors", "short": "2h"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "3h",
				Expr:        intstr.FromString(`http_requests:burnrate6h{job="thanos-receive-default",slo="monitoring-http-errors"} > (1 * (1-0.99)) and http_requests:burnrate4d{job="thanos-receive-default",slo="monitoring-http-errors"} > (1 * (1-0.99))`),
				Annotations: map[string]string{"severity": "warning"},
				Labels:      map[string]string{"long": "4d", "slo": "monitoring-http-errors", "short": "6h"},
			}},
		},
	}, {
		name: "http-ratio-grouping-regex",
		slo:  objectiveHTTPRatioGroupingRegex(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-errors",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_requests:burnrate5m",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_requests_total{code=~"5..",handler=~"/api.*",job="thanos-receive-default"}[5m])) / sum by(job, handler) (rate(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}[5m]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate30m",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_requests_total{code=~"5..",handler=~"/api.*",job="thanos-receive-default"}[30m])) / sum by(job, handler) (rate(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}[30m]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate1h",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_requests_total{code=~"5..",handler=~"/api.*",job="thanos-receive-default"}[1h])) / sum by(job, handler) (rate(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}[1h]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate2h",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_requests_total{code=~"5..",handler=~"/api.*",job="thanos-receive-default"}[2h])) / sum by(job, handler) (rate(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}[2h]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate6h",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_requests_total{code=~"5..",handler=~"/api.*",job="thanos-receive-default"}[6h])) / sum by(job, handler) (rate(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}[6h]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate1d",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_requests_total{code=~"5..",handler=~"/api.*",job="thanos-receive-default"}[1d])) / sum by(job, handler) (rate(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}[1d]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate4d",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_requests_total{code=~"5..",handler=~"/api.*",job="thanos-receive-default"}[4d])) / sum by(job, handler) (rate(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}[4d]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "2m",
				Expr:        intstr.FromString(`http_requests:burnrate5m{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"} > (14 * (1-0.99)) and http_requests:burnrate1h{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"} > (14 * (1-0.99))`),
				Annotations: map[string]string{"severity": "critical"},
				Labels:      map[string]string{"long": "1h", "short": "5m", "slo": "monitoring-http-errors"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "15m",
				Expr:        intstr.FromString(`http_requests:burnrate30m{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"} > (7 * (1-0.99)) and http_requests:burnrate6h{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"} > (7 * (1-0.99))`),
				Annotations: map[string]string{"severity": "critical"},
				Labels:      map[string]string{"long": "6h", "slo": "monitoring-http-errors", "short": "30m"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "1h",
				Expr:        intstr.FromString(`http_requests:burnrate2h{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"} > (2 * (1-0.99)) and http_requests:burnrate1d{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"} > (2 * (1-0.99))`),
				Annotations: map[string]string{"severity": "warning"},
				Labels:      map[string]string{"long": "1d", "slo": "monitoring-http-errors", "short": "2h"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "3h",
				Expr:        intstr.FromString(`http_requests:burnrate6h{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"} > (1 * (1-0.99)) and http_requests:burnrate4d{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"} > (1 * (1-0.99))`),
				Annotations: map[string]string{"severity": "warning"},
				Labels:      map[string]string{"long": "4d", "slo": "monitoring-http-errors", "short": "6h"},
			}},
		},
	}, {
		name: "operator-ratio",
		slo:  objectiveOperator(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-prometheus-operator-errors",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "prometheus_operator_reconcile_operations:burnrate3m",
				Expr:   intstr.FromString(`sum(rate(prometheus_operator_reconcile_errors_total[3m])) / sum(rate(prometheus_operator_reconcile_operations_total[3m]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Record: "prometheus_operator_reconcile_operations:burnrate15m",
				Expr:   intstr.FromString(`sum(rate(prometheus_operator_reconcile_errors_total[15m])) / sum(rate(prometheus_operator_reconcile_operations_total[15m]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Record: "prometheus_operator_reconcile_operations:burnrate30m",
				Expr:   intstr.FromString(`sum(rate(prometheus_operator_reconcile_errors_total[30m])) / sum(rate(prometheus_operator_reconcile_operations_total[30m]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Record: "prometheus_operator_reconcile_operations:burnrate1h",
				Expr:   intstr.FromString(`sum(rate(prometheus_operator_reconcile_errors_total[1h])) / sum(rate(prometheus_operator_reconcile_operations_total[1h]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Record: "prometheus_operator_reconcile_operations:burnrate3h",
				Expr:   intstr.FromString(`sum(rate(prometheus_operator_reconcile_errors_total[3h])) / sum(rate(prometheus_operator_reconcile_operations_total[3h]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Record: "prometheus_operator_reconcile_operations:burnrate12h",
				Expr:   intstr.FromString(`sum(rate(prometheus_operator_reconcile_errors_total[12h])) / sum(rate(prometheus_operator_reconcile_operations_total[12h]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Record: "prometheus_operator_reconcile_operations:burnrate2d",
				Expr:   intstr.FromString(`sum(rate(prometheus_operator_reconcile_errors_total[2d])) / sum(rate(prometheus_operator_reconcile_operations_total[2d]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "1m",
				Expr:        intstr.FromString(`prometheus_operator_reconcile_operations:burnrate3m{slo="monitoring-prometheus-operator-errors"} > (14 * (1-0.99)) and prometheus_operator_reconcile_operations:burnrate30m{slo="monitoring-prometheus-operator-errors"} > (14 * (1-0.99))`),
				Annotations: map[string]string{"severity": "critical"},
				Labels:      map[string]string{"long": "30m", "slo": "monitoring-prometheus-operator-errors", "short": "3m"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "8m",
				Expr:        intstr.FromString(`prometheus_operator_reconcile_operations:burnrate15m{slo="monitoring-prometheus-operator-errors"} > (7 * (1-0.99)) and prometheus_operator_reconcile_operations:burnrate3h{slo="monitoring-prometheus-operator-errors"} > (7 * (1-0.99))`),
				Annotations: map[string]string{"severity": "critical"},
				Labels:      map[string]string{"long": "3h", "slo": "monitoring-prometheus-operator-errors", "short": "15m"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "30m",
				Expr:        intstr.FromString(`prometheus_operator_reconcile_operations:burnrate1h{slo="monitoring-prometheus-operator-errors"} > (2 * (1-0.99)) and prometheus_operator_reconcile_operations:burnrate12h{slo="monitoring-prometheus-operator-errors"} > (2 * (1-0.99))`),
				Annotations: map[string]string{"severity": "warning"},
				Labels:      map[string]string{"long": "12h", "slo": "monitoring-prometheus-operator-errors", "short": "1h"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "1h30m",
				Expr:        intstr.FromString(`prometheus_operator_reconcile_operations:burnrate3h{slo="monitoring-prometheus-operator-errors"} > (1 * (1-0.99)) and prometheus_operator_reconcile_operations:burnrate2d{slo="monitoring-prometheus-operator-errors"} > (1 * (1-0.99))`),
				Annotations: map[string]string{"severity": "warning"},
				Labels:      map[string]string{"long": "2d", "slo": "monitoring-prometheus-operator-errors", "short": "3h"},
			}},
		},
	}, {
		name: "http-latency", // super similar to gRPC and therefore only HTTP
		slo:  objectiveHTTPLatency(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-latency",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_request_duration_seconds:burnrate5m",
				Expr:   intstr.FromString(`sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[5m])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[5m]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate30m",
				Expr:   intstr.FromString(`sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[30m])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[30m]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate1h",
				Expr:   intstr.FromString(`sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1h])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[1h]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate2h",
				Expr:   intstr.FromString(`sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[2h])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[2h]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate6h",
				Expr:   intstr.FromString(`sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[6h])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[6h]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate1d",
				Expr:   intstr.FromString(`sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1d])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[1d]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate4d",
				Expr:   intstr.FromString(`sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[4d])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[4d]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "2m",
				Expr:        intstr.FromString(`http_request_duration_seconds:burnrate5m{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (14 * (1-0.995)) and http_request_duration_seconds:burnrate1h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (14 * (1-0.995))`),
				Annotations: map[string]string{"severity": "critical"},
				Labels:      map[string]string{"long": "1h", "job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency", "short": "5m"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "15m",
				Expr:        intstr.FromString(`http_request_duration_seconds:burnrate30m{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (7 * (1-0.995)) and http_request_duration_seconds:burnrate6h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (7 * (1-0.995))`),
				Annotations: map[string]string{"severity": "critical"},
				Labels:      map[string]string{"long": "6h", "job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency", "short": "30m"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "1h",
				Expr:        intstr.FromString(`http_request_duration_seconds:burnrate2h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (2 * (1-0.995)) and http_request_duration_seconds:burnrate1d{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (2 * (1-0.995))`),
				Annotations: map[string]string{"severity": "warning"},
				Labels:      map[string]string{"long": "1d", "job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency", "short": "2h"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "3h",
				Expr:        intstr.FromString(`http_request_duration_seconds:burnrate6h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (1 * (1-0.995)) and http_request_duration_seconds:burnrate4d{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (1 * (1-0.995))`),
				Annotations: map[string]string{"severity": "warning"},
				Labels:      map[string]string{"long": "4d", "job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency", "short": "6h"},
			}},
		},
	}, {
		name: "http-latency-grouping",
		slo:  objectiveHTTPLatencyGrouping(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-latency",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_request_duration_seconds:burnrate5m",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[5m])) - sum by(job, handler) (rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[5m]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate30m",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[30m])) - sum by(job, handler) (rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[30m]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate1h",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1h])) - sum by(job, handler) (rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[1h]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate2h",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[2h])) - sum by(job, handler) (rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[2h]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate6h",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[6h])) - sum by(job, handler) (rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[6h]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate1d",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1d])) - sum by(job, handler) (rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[1d]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate4d",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[4d])) - sum by(job, handler) (rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[4d]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Alert:       "ErrorBudgetBurn",
				Expr:        intstr.FromString(`http_request_duration_seconds:burnrate5m{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (14 * (1-0.995)) and http_request_duration_seconds:burnrate1h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (14 * (1-0.995))`),
				For:         "2m",
				Labels:      map[string]string{"long": "1h", "short": "5m", "slo": "monitoring-http-latency"},
				Annotations: map[string]string{"severity": "critical"},
			}, {
				Alert:       "ErrorBudgetBurn",
				Expr:        intstr.FromString(`http_request_duration_seconds:burnrate30m{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (7 * (1-0.995)) and http_request_duration_seconds:burnrate6h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (7 * (1-0.995))`),
				For:         "15m",
				Labels:      map[string]string{"long": "6h", "short": "30m", "slo": "monitoring-http-latency"},
				Annotations: map[string]string{"severity": "critical"},
			}, {
				Alert:       "ErrorBudgetBurn",
				Expr:        intstr.FromString(`http_request_duration_seconds:burnrate2h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (2 * (1-0.995)) and http_request_duration_seconds:burnrate1d{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (2 * (1-0.995))`),
				For:         "1h",
				Labels:      map[string]string{"long": "1d", "short": "2h", "slo": "monitoring-http-latency"},
				Annotations: map[string]string{"severity": "warning"},
			}, {
				Alert:       "ErrorBudgetBurn",
				Expr:        intstr.FromString(`http_request_duration_seconds:burnrate6h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (1 * (1-0.995)) and http_request_duration_seconds:burnrate4d{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (1 * (1-0.995))`),
				For:         "3h",
				Labels:      map[string]string{"long": "4d", "short": "6h", "slo": "monitoring-http-latency"},
				Annotations: map[string]string{"severity": "warning"},
			}},
		},
	}, {
		name: "http-latency-grouping-regex",
		slo:  objectiveHTTPLatencyGroupingRegex(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-latency",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_request_duration_seconds:burnrate5m",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[5m])) - sum by(job, handler) (rate(http_request_duration_seconds_bucket{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1"}[5m]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate30m",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[30m])) - sum by(job, handler) (rate(http_request_duration_seconds_bucket{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1"}[30m]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate1h",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[1h])) - sum by(job, handler) (rate(http_request_duration_seconds_bucket{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1"}[1h]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate2h",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[2h])) - sum by(job, handler) (rate(http_request_duration_seconds_bucket{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1"}[2h]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate6h",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[6h])) - sum by(job, handler) (rate(http_request_duration_seconds_bucket{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1"}[6h]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate1d",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[1d])) - sum by(job, handler) (rate(http_request_duration_seconds_bucket{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1"}[1d]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate4d",
				Expr:   intstr.FromString(`sum by(job, handler) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[4d])) - sum by(job, handler) (rate(http_request_duration_seconds_bucket{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1"}[4d]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Alert:       "ErrorBudgetBurn",
				Expr:        intstr.FromString(`http_request_duration_seconds:burnrate5m{handler=~"/api.*",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (14 * (1-0.995)) and http_request_duration_seconds:burnrate1h{handler=~"/api.*",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (14 * (1-0.995))`),
				For:         "2m",
				Labels:      map[string]string{"long": "1h", "short": "5m", "slo": "monitoring-http-latency"},
				Annotations: map[string]string{"severity": "critical"},
			}, {
				Alert:       "ErrorBudgetBurn",
				Expr:        intstr.FromString(`http_request_duration_seconds:burnrate30m{handler=~"/api.*",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (7 * (1-0.995)) and http_request_duration_seconds:burnrate6h{handler=~"/api.*",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (7 * (1-0.995))`),
				For:         "15m",
				Labels:      map[string]string{"long": "6h", "short": "30m", "slo": "monitoring-http-latency"},
				Annotations: map[string]string{"severity": "critical"},
			}, {
				Alert:       "ErrorBudgetBurn",
				Expr:        intstr.FromString(`http_request_duration_seconds:burnrate2h{handler=~"/api.*",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (2 * (1-0.995)) and http_request_duration_seconds:burnrate1d{handler=~"/api.*",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (2 * (1-0.995))`),
				For:         "1h",
				Labels:      map[string]string{"long": "1d", "short": "2h", "slo": "monitoring-http-latency"},
				Annotations: map[string]string{"severity": "warning"},
			}, {
				Alert:       "ErrorBudgetBurn",
				Expr:        intstr.FromString(`http_request_duration_seconds:burnrate6h{handler=~"/api.*",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (1 * (1-0.995)) and http_request_duration_seconds:burnrate4d{handler=~"/api.*",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (1 * (1-0.995))`),
				For:         "3h",
				Labels:      map[string]string{"long": "4d", "short": "6h", "slo": "monitoring-http-latency"},
				Annotations: map[string]string{"severity": "warning"},
			}},
		},
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			group, err := tc.slo.Burnrates()
			require.NoError(t, err)
			require.Equal(t, tc.rules, group)
		})
	}
}

func Test_windows(t *testing.T) {
	ws := windows(28 * 24 * time.Hour)

	require.Equal(t, window{
		Severity: critical,
		For:      2 * time.Minute,
		Long:     1 * time.Hour,
		Short:    5 * time.Minute,
		Factor:   14,
	}, ws[0])

	require.Equal(t, window{
		Severity: critical,
		For:      15 * time.Minute,
		Long:     6 * time.Hour,
		Short:    30 * time.Minute,
		Factor:   7,
	}, ws[1])

	require.Equal(t, window{
		Severity: warning,
		For:      time.Hour,
		Long:     24 * time.Hour,
		Short:    2 * time.Hour,
		Factor:   2,
	}, ws[2])

	require.Equal(t, window{
		Severity: warning,
		For:      3 * time.Hour,
		Long:     4 * 24 * time.Hour,
		Short:    6 * time.Hour,
		Factor:   1,
	}, ws[3])
}

func TestObjective_Alerts(t *testing.T) {
	testcases := []struct {
		name   string
		slo    Objective
		alerts []MultiBurnRateAlert
	}{{
		name: "http-ratio",
		slo:  objectiveHTTPRatio(),
		alerts: []MultiBurnRateAlert{{
			Severity:   "critical",
			Short:      5 * time.Minute,
			Long:       1 * time.Hour,
			For:        2 * time.Minute,
			Factor:     14,
			QueryShort: `sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[5m])) / sum(rate(http_requests_total{job="thanos-receive-default"}[5m]))`,
			QueryLong:  `sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[1h])) / sum(rate(http_requests_total{job="thanos-receive-default"}[1h]))`,
		}, {
			Severity:   "critical",
			Short:      30 * time.Minute,
			Long:       6 * time.Hour,
			For:        15 * time.Minute,
			Factor:     7,
			QueryShort: `sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[30m])) / sum(rate(http_requests_total{job="thanos-receive-default"}[30m]))`,
			QueryLong:  `sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[6h])) / sum(rate(http_requests_total{job="thanos-receive-default"}[6h]))`,
		}, {
			Severity:   "warning",
			Short:      2 * time.Hour,
			Long:       24 * time.Hour,
			For:        time.Hour,
			Factor:     2,
			QueryShort: `sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[2h])) / sum(rate(http_requests_total{job="thanos-receive-default"}[2h]))`,
			QueryLong:  `sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[1d])) / sum(rate(http_requests_total{job="thanos-receive-default"}[1d]))`,
		}, {
			Severity:   "warning",
			Short:      6 * time.Hour,
			Long:       4 * 24 * time.Hour,
			For:        3 * time.Hour,
			Factor:     1,
			QueryShort: `sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[6h])) / sum(rate(http_requests_total{job="thanos-receive-default"}[6h]))`,
			QueryLong:  `sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[4d])) / sum(rate(http_requests_total{job="thanos-receive-default"}[4d]))`,
		}},
	}, {
		name: "http-latency",
		slo:  objectiveHTTPLatency(),
		alerts: []MultiBurnRateAlert{{
			Severity:   "critical",
			Short:      5 * time.Minute,
			Long:       1 * time.Hour,
			For:        2 * time.Minute,
			Factor:     14,
			QueryShort: `sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[5m])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[5m]))`,
			QueryLong:  `sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1h])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[1h]))`,
		}, {
			Severity:   "critical",
			Short:      30 * time.Minute,
			Long:       6 * time.Hour,
			For:        15 * time.Minute,
			Factor:     7,
			QueryShort: `sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[30m])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[30m]))`,
			QueryLong:  `sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[6h])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[6h]))`,
		}, {
			Severity:   "warning",
			Short:      2 * time.Hour,
			Long:       24 * time.Hour,
			For:        time.Hour,
			Factor:     2,
			QueryShort: `sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[2h])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[2h]))`,
			QueryLong:  `sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1d])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[1d]))`,
		}, {
			Severity:   "warning",
			Short:      6 * time.Hour,
			Long:       4 * 24 * time.Hour,
			For:        3 * time.Hour,
			Factor:     1,
			QueryShort: `sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[6h])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[6h]))`,
			QueryLong:  `sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[4d])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[4d]))`,
		}},
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			alerts, err := tc.slo.Alerts()
			require.NoError(t, err)
			require.Equal(t, tc.alerts, alerts)
		})
	}
}
