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
		name: "http-ratio",
		slo:  objectiveHTTPRatio(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-errors",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_requests:burnrate5m",
				Expr:   intstr.FromString(`sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[5m])) / sum(rate(http_requests_total{job="thanos-receive-default"}[5m]))`),
				Labels: map[string]string{"job": "thanos-receive-default", "slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate30m",
				Expr:   intstr.FromString(`sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[30m])) / sum(rate(http_requests_total{job="thanos-receive-default"}[30m]))`),
				Labels: map[string]string{"job": "thanos-receive-default", "slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate1h",
				Expr:   intstr.FromString(`sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[1h])) / sum(rate(http_requests_total{job="thanos-receive-default"}[1h]))`),
				Labels: map[string]string{"job": "thanos-receive-default", "slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate2h",
				Expr:   intstr.FromString(`sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[2h])) / sum(rate(http_requests_total{job="thanos-receive-default"}[2h]))`),
				Labels: map[string]string{"job": "thanos-receive-default", "slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate6h",
				Expr:   intstr.FromString(`sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[6h])) / sum(rate(http_requests_total{job="thanos-receive-default"}[6h]))`),
				Labels: map[string]string{"job": "thanos-receive-default", "slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate1d",
				Expr:   intstr.FromString(`sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[1d])) / sum(rate(http_requests_total{job="thanos-receive-default"}[1d]))`),
				Labels: map[string]string{"job": "thanos-receive-default", "slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate4d",
				Expr:   intstr.FromString(`sum(rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[4d])) / sum(rate(http_requests_total{job="thanos-receive-default"}[4d]))`),
				Labels: map[string]string{"job": "thanos-receive-default", "slo": "monitoring-http-errors"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "2m",
				Expr:   intstr.FromString(`http_requests:burnrate5m{job="thanos-receive-default",slo="monitoring-http-errors"} > (14 * (1-0.99)) and http_requests:burnrate1h{job="thanos-receive-default",slo="monitoring-http-errors"} > (14 * (1-0.99))`),
				Labels: map[string]string{"severity": "critical", "job": "thanos-receive-default", "long": "1h", "slo": "monitoring-http-errors", "short": "5m"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "15m",
				Expr:   intstr.FromString(`http_requests:burnrate30m{job="thanos-receive-default",slo="monitoring-http-errors"} > (7 * (1-0.99)) and http_requests:burnrate6h{job="thanos-receive-default",slo="monitoring-http-errors"} > (7 * (1-0.99))`),
				Labels: map[string]string{"severity": "critical", "job": "thanos-receive-default", "long": "6h", "slo": "monitoring-http-errors", "short": "30m"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "1h",
				Expr:   intstr.FromString(`http_requests:burnrate2h{job="thanos-receive-default",slo="monitoring-http-errors"} > (2 * (1-0.99)) and http_requests:burnrate1d{job="thanos-receive-default",slo="monitoring-http-errors"} > (2 * (1-0.99))`),
				Labels: map[string]string{"severity": "warning", "job": "thanos-receive-default", "long": "1d", "slo": "monitoring-http-errors", "short": "2h"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "3h",
				Expr:   intstr.FromString(`http_requests:burnrate6h{job="thanos-receive-default",slo="monitoring-http-errors"} > (1 * (1-0.99)) and http_requests:burnrate4d{job="thanos-receive-default",slo="monitoring-http-errors"} > (1 * (1-0.99))`),
				Labels: map[string]string{"severity": "warning", "job": "thanos-receive-default", "long": "4d", "slo": "monitoring-http-errors", "short": "6h"},
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
				Expr:   intstr.FromString(`sum by (handler, job) (rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[5m])) / sum by (handler, job) (rate(http_requests_total{job="thanos-receive-default"}[5m]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate30m",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[30m])) / sum by (handler, job) (rate(http_requests_total{job="thanos-receive-default"}[30m]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate1h",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[1h])) / sum by (handler, job) (rate(http_requests_total{job="thanos-receive-default"}[1h]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate2h",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[2h])) / sum by (handler, job) (rate(http_requests_total{job="thanos-receive-default"}[2h]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate6h",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[6h])) / sum by (handler, job) (rate(http_requests_total{job="thanos-receive-default"}[6h]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate1d",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[1d])) / sum by (handler, job) (rate(http_requests_total{job="thanos-receive-default"}[1d]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate4d",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[4d])) / sum by (handler, job) (rate(http_requests_total{job="thanos-receive-default"}[4d]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "2m",
				Expr:   intstr.FromString(`http_requests:burnrate5m{job="thanos-receive-default",slo="monitoring-http-errors"} > (14 * (1-0.99)) and http_requests:burnrate1h{job="thanos-receive-default",slo="monitoring-http-errors"} > (14 * (1-0.99))`),
				Labels: map[string]string{"severity": "critical", "long": "1h", "slo": "monitoring-http-errors", "short": "5m"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "15m",
				Expr:   intstr.FromString(`http_requests:burnrate30m{job="thanos-receive-default",slo="monitoring-http-errors"} > (7 * (1-0.99)) and http_requests:burnrate6h{job="thanos-receive-default",slo="monitoring-http-errors"} > (7 * (1-0.99))`),
				Labels: map[string]string{"severity": "critical", "long": "6h", "slo": "monitoring-http-errors", "short": "30m"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "1h",
				Expr:   intstr.FromString(`http_requests:burnrate2h{job="thanos-receive-default",slo="monitoring-http-errors"} > (2 * (1-0.99)) and http_requests:burnrate1d{job="thanos-receive-default",slo="monitoring-http-errors"} > (2 * (1-0.99))`),
				Labels: map[string]string{"severity": "warning", "long": "1d", "slo": "monitoring-http-errors", "short": "2h"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "3h",
				Expr:   intstr.FromString(`http_requests:burnrate6h{job="thanos-receive-default",slo="monitoring-http-errors"} > (1 * (1-0.99)) and http_requests:burnrate4d{job="thanos-receive-default",slo="monitoring-http-errors"} > (1 * (1-0.99))`),
				Labels: map[string]string{"severity": "warning", "long": "4d", "slo": "monitoring-http-errors", "short": "6h"},
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
				Expr:   intstr.FromString(`sum by (handler, job) (rate(http_requests_total{code=~"5..",handler=~"/api.*",job="thanos-receive-default"}[5m])) / sum by (handler, job) (rate(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}[5m]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate30m",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(http_requests_total{code=~"5..",handler=~"/api.*",job="thanos-receive-default"}[30m])) / sum by (handler, job) (rate(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}[30m]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate1h",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(http_requests_total{code=~"5..",handler=~"/api.*",job="thanos-receive-default"}[1h])) / sum by (handler, job) (rate(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}[1h]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate2h",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(http_requests_total{code=~"5..",handler=~"/api.*",job="thanos-receive-default"}[2h])) / sum by (handler, job) (rate(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}[2h]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate6h",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(http_requests_total{code=~"5..",handler=~"/api.*",job="thanos-receive-default"}[6h])) / sum by (handler, job) (rate(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}[6h]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate1d",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(http_requests_total{code=~"5..",handler=~"/api.*",job="thanos-receive-default"}[1d])) / sum by (handler, job) (rate(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}[1d]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "http_requests:burnrate4d",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(http_requests_total{code=~"5..",handler=~"/api.*",job="thanos-receive-default"}[4d])) / sum by (handler, job) (rate(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}[4d]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "2m",
				Expr:   intstr.FromString(`http_requests:burnrate5m{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"} > (14 * (1-0.99)) and http_requests:burnrate1h{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"} > (14 * (1-0.99))`),
				Labels: map[string]string{"severity": "critical", "long": "1h", "short": "5m", "slo": "monitoring-http-errors"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "15m",
				Expr:   intstr.FromString(`http_requests:burnrate30m{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"} > (7 * (1-0.99)) and http_requests:burnrate6h{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"} > (7 * (1-0.99))`),
				Labels: map[string]string{"severity": "critical", "long": "6h", "slo": "monitoring-http-errors", "short": "30m"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "1h",
				Expr:   intstr.FromString(`http_requests:burnrate2h{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"} > (2 * (1-0.99)) and http_requests:burnrate1d{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"} > (2 * (1-0.99))`),
				Labels: map[string]string{"severity": "warning", "long": "1d", "slo": "monitoring-http-errors", "short": "2h"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "3h",
				Expr:   intstr.FromString(`http_requests:burnrate6h{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"} > (1 * (1-0.99)) and http_requests:burnrate4d{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"} > (1 * (1-0.99))`),
				Labels: map[string]string{"severity": "warning", "long": "4d", "slo": "monitoring-http-errors", "short": "6h"},
			}},
		},
	}, {
		name: "grpc-errors",
		slo:  objectiveGRPCRatio(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-grpc-errors",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "grpc_server_handled:burnrate5m",
				Expr:   intstr.FromString(`sum(rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[5m])) / sum(rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[5m]))`),
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "job": "api", "slo": "monitoring-grpc-errors"},
			}, {
				Record: "grpc_server_handled:burnrate30m",
				Expr:   intstr.FromString(`sum(rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[30m])) / sum(rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[30m]))`),
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "job": "api", "slo": "monitoring-grpc-errors"},
			}, {
				Record: "grpc_server_handled:burnrate1h",
				Expr:   intstr.FromString(`sum(rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1h])) / sum(rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1h]))`),
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "job": "api", "slo": "monitoring-grpc-errors"},
			}, {
				Record: "grpc_server_handled:burnrate2h",
				Expr:   intstr.FromString(`sum(rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[2h])) / sum(rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[2h]))`),
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "job": "api", "slo": "monitoring-grpc-errors"},
			}, {
				Record: "grpc_server_handled:burnrate6h",
				Expr:   intstr.FromString(`sum(rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[6h])) / sum(rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[6h]))`),
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "job": "api", "slo": "monitoring-grpc-errors"},
			}, {
				Record: "grpc_server_handled:burnrate1d",
				Expr:   intstr.FromString(`sum(rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1d])) / sum(rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1d]))`),
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "job": "api", "slo": "monitoring-grpc-errors"},
			}, {
				Record: "grpc_server_handled:burnrate4d",
				Expr:   intstr.FromString(`sum(rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[4d])) / sum(rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[4d]))`),
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "job": "api", "slo": "monitoring-grpc-errors"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`grpc_server_handled:burnrate5m{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} > (14 * (1-0.999)) and grpc_server_handled:burnrate1h{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} > (14 * (1-0.999))`),
				For:    "2m",
				Labels: map[string]string{"severity": "critical", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "job": "api", "slo": "monitoring-grpc-errors", "short": "5m", "long": "1h"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`grpc_server_handled:burnrate30m{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} > (7 * (1-0.999)) and grpc_server_handled:burnrate6h{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} > (7 * (1-0.999))`),
				For:    "15m",
				Labels: map[string]string{"severity": "critical", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "job": "api", "slo": "monitoring-grpc-errors", "short": "30m", "long": "6h"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`grpc_server_handled:burnrate2h{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} > (2 * (1-0.999)) and grpc_server_handled:burnrate1d{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} > (2 * (1-0.999))`),
				For:    "1h",
				Labels: map[string]string{"severity": "warning", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "job": "api", "slo": "monitoring-grpc-errors", "short": "2h", "long": "1d"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`grpc_server_handled:burnrate6h{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} > (1 * (1-0.999)) and grpc_server_handled:burnrate4d{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} > (1 * (1-0.999))`),
				For:    "3h",
				Labels: map[string]string{"severity": "warning", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "job": "api", "slo": "monitoring-grpc-errors", "short": "6h", "long": "4d"},
			}},
		},
	}, {
		name: "grpc-errors-grouping",
		slo:  objectiveGRPCRatioGrouping(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-grpc-errors",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "grpc_server_handled:burnrate5m",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[5m])) / sum by (handler, job) (rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[5m]))`),
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "slo": "monitoring-grpc-errors"},
			}, {
				Record: "grpc_server_handled:burnrate30m",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[30m])) / sum by (handler, job) (rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[30m]))`),
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "slo": "monitoring-grpc-errors"},
			}, {
				Record: "grpc_server_handled:burnrate1h",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1h])) / sum by (handler, job) (rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1h]))`),
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "slo": "monitoring-grpc-errors"},
			}, {
				Record: "grpc_server_handled:burnrate2h",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[2h])) / sum by (handler, job) (rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[2h]))`),
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "slo": "monitoring-grpc-errors"},
			}, {
				Record: "grpc_server_handled:burnrate6h",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[6h])) / sum by (handler, job) (rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[6h]))`),
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "slo": "monitoring-grpc-errors"},
			}, {
				Record: "grpc_server_handled:burnrate1d",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1d])) / sum by (handler, job) (rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1d]))`),
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "slo": "monitoring-grpc-errors"},
			}, {
				Record: "grpc_server_handled:burnrate4d",
				Expr:   intstr.FromString(`sum by (handler, job) (rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[4d])) / sum by (handler, job) (rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[4d]))`),
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "slo": "monitoring-grpc-errors"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`grpc_server_handled:burnrate5m{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} > (14 * (1-0.999)) and grpc_server_handled:burnrate1h{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} > (14 * (1-0.999))`),
				For:    "2m",
				Labels: map[string]string{"severity": "critical", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "slo": "monitoring-grpc-errors", "short": "5m", "long": "1h"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`grpc_server_handled:burnrate30m{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} > (7 * (1-0.999)) and grpc_server_handled:burnrate6h{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} > (7 * (1-0.999))`),
				For:    "15m",
				Labels: map[string]string{"severity": "critical", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "slo": "monitoring-grpc-errors", "short": "30m", "long": "6h"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`grpc_server_handled:burnrate2h{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} > (2 * (1-0.999)) and grpc_server_handled:burnrate1d{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} > (2 * (1-0.999))`),
				For:    "1h",
				Labels: map[string]string{"severity": "warning", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "slo": "monitoring-grpc-errors", "short": "2h", "long": "1d"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`grpc_server_handled:burnrate6h{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} > (1 * (1-0.999)) and grpc_server_handled:burnrate4d{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} > (1 * (1-0.999))`),
				For:    "3h",
				Labels: map[string]string{"severity": "warning", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "slo": "monitoring-grpc-errors", "short": "6h", "long": "4d"},
			}},
		},
	}, {
		name: "http-latency",
		slo:  objectiveHTTPLatency(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-latency",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_request_duration_seconds:burnrate5m",
				Expr:   intstr.FromString(`(sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[5m])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[5m]))) / sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[5m]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate30m",
				Expr:   intstr.FromString(`(sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[30m])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[30m]))) / sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[30m]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate1h",
				Expr:   intstr.FromString(`(sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1h])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[1h]))) / sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1h]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate2h",
				Expr:   intstr.FromString(`(sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[2h])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[2h]))) / sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[2h]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate6h",
				Expr:   intstr.FromString(`(sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[6h])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[6h]))) / sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[6h]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate1d",
				Expr:   intstr.FromString(`(sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1d])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[1d]))) / sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1d]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate4d",
				Expr:   intstr.FromString(`(sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[4d])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[4d]))) / sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[4d]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "2m",
				Expr:   intstr.FromString(`http_request_duration_seconds:burnrate5m{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (14 * (1-0.995)) and http_request_duration_seconds:burnrate1h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (14 * (1-0.995))`),
				Labels: map[string]string{"severity": "critical", "long": "1h", "job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency", "short": "5m"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "15m",
				Expr:   intstr.FromString(`http_request_duration_seconds:burnrate30m{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (7 * (1-0.995)) and http_request_duration_seconds:burnrate6h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (7 * (1-0.995))`),
				Labels: map[string]string{"severity": "critical", "long": "6h", "job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency", "short": "30m"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "1h",
				Expr:   intstr.FromString(`http_request_duration_seconds:burnrate2h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (2 * (1-0.995)) and http_request_duration_seconds:burnrate1d{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (2 * (1-0.995))`),
				Labels: map[string]string{"severity": "warning", "long": "1d", "job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency", "short": "2h"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "3h",
				Expr:   intstr.FromString(`http_request_duration_seconds:burnrate6h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (1 * (1-0.995)) and http_request_duration_seconds:burnrate4d{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (1 * (1-0.995))`),
				Labels: map[string]string{"severity": "warning", "long": "4d", "job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency", "short": "6h"},
			}},
		},
	}, {
		name: "http-latency-native",
		slo:  objectiveHTTPNativeLatency(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-latency",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_request_duration_seconds:burnrate5m",
				Expr:   intstr.FromString(`1 - histogram_fraction(0, 1, rate(http_request_duration_seconds{code=~"2..",job="metrics-service-thanos-receive-default"}[5m]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate30m",
				Expr:   intstr.FromString(`1 - histogram_fraction(0, 1, rate(http_request_duration_seconds{code=~"2..",job="metrics-service-thanos-receive-default"}[30m]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate1h",
				Expr:   intstr.FromString(`1 - histogram_fraction(0, 1, rate(http_request_duration_seconds{code=~"2..",job="metrics-service-thanos-receive-default"}[1h]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate2h",
				Expr:   intstr.FromString(`1 - histogram_fraction(0, 1, rate(http_request_duration_seconds{code=~"2..",job="metrics-service-thanos-receive-default"}[2h]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate6h",
				Expr:   intstr.FromString(`1 - histogram_fraction(0, 1, rate(http_request_duration_seconds{code=~"2..",job="metrics-service-thanos-receive-default"}[6h]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate1d",
				Expr:   intstr.FromString(`1 - histogram_fraction(0, 1, rate(http_request_duration_seconds{code=~"2..",job="metrics-service-thanos-receive-default"}[1d]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate4d",
				Expr:   intstr.FromString(`1 - histogram_fraction(0, 1, rate(http_request_duration_seconds{code=~"2..",job="metrics-service-thanos-receive-default"}[4d]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "2m",
				Expr:   intstr.FromString(`http_request_duration_seconds:burnrate5m{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (14 * (1-0.995)) and http_request_duration_seconds:burnrate1h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (14 * (1-0.995))`),
				Labels: map[string]string{"severity": "critical", "long": "1h", "job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency", "short": "5m"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "15m",
				Expr:   intstr.FromString(`http_request_duration_seconds:burnrate30m{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (7 * (1-0.995)) and http_request_duration_seconds:burnrate6h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (7 * (1-0.995))`),
				Labels: map[string]string{"severity": "critical", "long": "6h", "job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency", "short": "30m"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "1h",
				Expr:   intstr.FromString(`http_request_duration_seconds:burnrate2h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (2 * (1-0.995)) and http_request_duration_seconds:burnrate1d{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (2 * (1-0.995))`),
				Labels: map[string]string{"severity": "warning", "long": "1d", "job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency", "short": "2h"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "3h",
				Expr:   intstr.FromString(`http_request_duration_seconds:burnrate6h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (1 * (1-0.995)) and http_request_duration_seconds:burnrate4d{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (1 * (1-0.995))`),
				Labels: map[string]string{"severity": "warning", "long": "4d", "job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency", "short": "6h"},
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
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[5m])) - sum by (handler, job) (rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[5m]))) / sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[5m]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate30m",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[30m])) - sum by (handler, job) (rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[30m]))) / sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[30m]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate1h",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1h])) - sum by (handler, job) (rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[1h]))) / sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1h]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate2h",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[2h])) - sum by (handler, job) (rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[2h]))) / sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[2h]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate6h",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[6h])) - sum by (handler, job) (rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[6h]))) / sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[6h]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate1d",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1d])) - sum by (handler, job) (rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[1d]))) / sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1d]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate4d",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[4d])) - sum by (handler, job) (rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[4d]))) / sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[4d]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "2m",
				Expr:   intstr.FromString(`http_request_duration_seconds:burnrate5m{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (14 * (1-0.995)) and http_request_duration_seconds:burnrate1h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (14 * (1-0.995))`),
				Labels: map[string]string{"severity": "critical", "long": "1h", "slo": "monitoring-http-latency", "short": "5m"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "15m",
				Expr:   intstr.FromString(`http_request_duration_seconds:burnrate30m{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (7 * (1-0.995)) and http_request_duration_seconds:burnrate6h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (7 * (1-0.995))`),
				Labels: map[string]string{"severity": "critical", "long": "6h", "slo": "monitoring-http-latency", "short": "30m"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "1h",
				Expr:   intstr.FromString(`http_request_duration_seconds:burnrate2h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (2 * (1-0.995)) and http_request_duration_seconds:burnrate1d{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (2 * (1-0.995))`),
				Labels: map[string]string{"severity": "warning", "long": "1d", "slo": "monitoring-http-latency", "short": "2h"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "3h",
				Expr:   intstr.FromString(`http_request_duration_seconds:burnrate6h{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (1 * (1-0.995)) and http_request_duration_seconds:burnrate4d{job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (1 * (1-0.995))`),
				Labels: map[string]string{"severity": "warning", "long": "4d", "slo": "monitoring-http-latency", "short": "6h"},
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
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[5m])) - sum by (handler, job) (rate(http_request_duration_seconds_bucket{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1"}[5m]))) / sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[5m]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate30m",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[30m])) - sum by (handler, job) (rate(http_request_duration_seconds_bucket{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1"}[30m]))) / sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[30m]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate1h",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[1h])) - sum by (handler, job) (rate(http_request_duration_seconds_bucket{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1"}[1h]))) / sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[1h]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate2h",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[2h])) - sum by (handler, job) (rate(http_request_duration_seconds_bucket{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1"}[2h]))) / sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[2h]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate6h",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[6h])) - sum by (handler, job) (rate(http_request_duration_seconds_bucket{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1"}[6h]))) / sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[6h]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate1d",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[1d])) - sum by (handler, job) (rate(http_request_duration_seconds_bucket{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1"}[1d]))) / sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[1d]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:burnrate4d",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[4d])) - sum by (handler, job) (rate(http_request_duration_seconds_bucket{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1"}[4d]))) / sum by (handler, job) (rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[4d]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`http_request_duration_seconds:burnrate5m{handler=~"/api.*",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (14 * (1-0.995)) and http_request_duration_seconds:burnrate1h{handler=~"/api.*",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (14 * (1-0.995))`),
				For:    "2m",
				Labels: map[string]string{"severity": "critical", "long": "1h", "short": "5m", "slo": "monitoring-http-latency"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`http_request_duration_seconds:burnrate30m{handler=~"/api.*",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (7 * (1-0.995)) and http_request_duration_seconds:burnrate6h{handler=~"/api.*",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (7 * (1-0.995))`),
				For:    "15m",
				Labels: map[string]string{"severity": "critical", "long": "6h", "short": "30m", "slo": "monitoring-http-latency"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`http_request_duration_seconds:burnrate2h{handler=~"/api.*",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (2 * (1-0.995)) and http_request_duration_seconds:burnrate1d{handler=~"/api.*",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (2 * (1-0.995))`),
				For:    "1h",
				Labels: map[string]string{"severity": "warning", "long": "1d", "short": "2h", "slo": "monitoring-http-latency"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`http_request_duration_seconds:burnrate6h{handler=~"/api.*",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (1 * (1-0.995)) and http_request_duration_seconds:burnrate4d{handler=~"/api.*",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"} > (1 * (1-0.995))`),
				For:    "3h",
				Labels: map[string]string{"severity": "warning", "long": "4d", "short": "6h", "slo": "monitoring-http-latency"},
			}},
		},
	}, {
		name: "grpc-latency",
		slo:  objectiveGRPCLatency(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-grpc-latency",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "grpc_server_handling_seconds:burnrate1m",
				Expr:   intstr.FromString(`(sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1m])) - sum(rate(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[1m]))) / sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1m]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "job": "api", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Record: "grpc_server_handling_seconds:burnrate8m",
				Expr:   intstr.FromString(`(sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[8m])) - sum(rate(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[8m]))) / sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[8m]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "job": "api", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Record: "grpc_server_handling_seconds:burnrate15m",
				Expr:   intstr.FromString(`(sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[15m])) - sum(rate(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[15m]))) / sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[15m]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "job": "api", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Record: "grpc_server_handling_seconds:burnrate30m",
				Expr:   intstr.FromString(`(sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[30m])) - sum(rate(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[30m]))) / sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[30m]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "job": "api", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Record: "grpc_server_handling_seconds:burnrate1h30m",
				Expr:   intstr.FromString(`(sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1h30m])) - sum(rate(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[1h30m]))) / sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1h30m]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "job": "api", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Record: "grpc_server_handling_seconds:burnrate6h",
				Expr:   intstr.FromString(`(sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[6h])) - sum(rate(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[6h]))) / sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[6h]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "job": "api", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Record: "grpc_server_handling_seconds:burnrate1d",
				Expr:   intstr.FromString(`(sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1d])) - sum(rate(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[1d]))) / sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1d]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "job": "api", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`grpc_server_handling_seconds:burnrate1m{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"} > (14 * (1-0.995)) and grpc_server_handling_seconds:burnrate15m{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"} > (14 * (1-0.995))`),
				For:    "1m",
				Labels: map[string]string{"severity": "critical", "long": "15m", "short": "1m", "slo": "monitoring-grpc-latency", "job": "api", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`grpc_server_handling_seconds:burnrate8m{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"} > (7 * (1-0.995)) and grpc_server_handling_seconds:burnrate1h30m{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"} > (7 * (1-0.995))`),
				For:    "4m",
				Labels: map[string]string{"severity": "critical", "long": "1h30m", "short": "8m", "slo": "monitoring-grpc-latency", "job": "api", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`grpc_server_handling_seconds:burnrate30m{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"} > (2 * (1-0.995)) and grpc_server_handling_seconds:burnrate6h{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"} > (2 * (1-0.995))`),
				For:    "15m",
				Labels: map[string]string{"severity": "warning", "long": "6h", "short": "30m", "slo": "monitoring-grpc-latency", "job": "api", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`grpc_server_handling_seconds:burnrate1h30m{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"} > (1 * (1-0.995)) and grpc_server_handling_seconds:burnrate1d{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"} > (1 * (1-0.995))`),
				For:    "45m",
				Labels: map[string]string{"severity": "warning", "long": "1d", "short": "1h30m", "slo": "monitoring-grpc-latency", "job": "api", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}},
		},
	}, {
		name: "grpc-latency-grouping",
		slo:  objectiveGRPCLatencyGrouping(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-grpc-latency",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "grpc_server_handling_seconds:burnrate1m",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1m])) - sum by (handler, job) (rate(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[1m]))) / sum by (handler, job) (rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1m]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Record: "grpc_server_handling_seconds:burnrate8m",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[8m])) - sum by (handler, job) (rate(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[8m]))) / sum by (handler, job) (rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[8m]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Record: "grpc_server_handling_seconds:burnrate15m",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[15m])) - sum by (handler, job) (rate(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[15m]))) / sum by (handler, job) (rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[15m]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Record: "grpc_server_handling_seconds:burnrate30m",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[30m])) - sum by (handler, job) (rate(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[30m]))) / sum by (handler, job) (rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[30m]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Record: "grpc_server_handling_seconds:burnrate1h30m",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1h30m])) - sum by (handler, job) (rate(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[1h30m]))) / sum by (handler, job) (rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1h30m]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Record: "grpc_server_handling_seconds:burnrate6h",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[6h])) - sum by (handler, job) (rate(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[6h]))) / sum by (handler, job) (rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[6h]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Record: "grpc_server_handling_seconds:burnrate1d",
				Expr:   intstr.FromString(`(sum by (handler, job) (rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1d])) - sum by (handler, job) (rate(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[1d]))) / sum by (handler, job) (rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1d]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`grpc_server_handling_seconds:burnrate1m{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"} > (14 * (1-0.995)) and grpc_server_handling_seconds:burnrate15m{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"} > (14 * (1-0.995))`),
				For:    "1m",
				Labels: map[string]string{"severity": "critical", "long": "15m", "short": "1m", "slo": "monitoring-grpc-latency", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`grpc_server_handling_seconds:burnrate8m{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"} > (7 * (1-0.995)) and grpc_server_handling_seconds:burnrate1h30m{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"} > (7 * (1-0.995))`),
				For:    "4m",
				Labels: map[string]string{"severity": "critical", "long": "1h30m", "short": "8m", "slo": "monitoring-grpc-latency", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`grpc_server_handling_seconds:burnrate30m{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"} > (2 * (1-0.995)) and grpc_server_handling_seconds:burnrate6h{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"} > (2 * (1-0.995))`),
				For:    "15m",
				Labels: map[string]string{"severity": "warning", "long": "6h", "short": "30m", "slo": "monitoring-grpc-latency", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`grpc_server_handling_seconds:burnrate1h30m{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"} > (1 * (1-0.995)) and grpc_server_handling_seconds:burnrate1d{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"} > (1 * (1-0.995))`),
				For:    "45m",
				Labels: map[string]string{"severity": "warning", "long": "1d", "short": "1h30m", "slo": "monitoring-grpc-latency", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
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
				Alert:  "ErrorBudgetBurn",
				For:    "1m",
				Expr:   intstr.FromString(`prometheus_operator_reconcile_operations:burnrate3m{slo="monitoring-prometheus-operator-errors"} > (14 * (1-0.99)) and prometheus_operator_reconcile_operations:burnrate30m{slo="monitoring-prometheus-operator-errors"} > (14 * (1-0.99))`),
				Labels: map[string]string{"severity": "critical", "long": "30m", "slo": "monitoring-prometheus-operator-errors", "short": "3m"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "8m",
				Expr:   intstr.FromString(`prometheus_operator_reconcile_operations:burnrate15m{slo="monitoring-prometheus-operator-errors"} > (7 * (1-0.99)) and prometheus_operator_reconcile_operations:burnrate3h{slo="monitoring-prometheus-operator-errors"} > (7 * (1-0.99))`),
				Labels: map[string]string{"severity": "critical", "long": "3h", "slo": "monitoring-prometheus-operator-errors", "short": "15m"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "30m",
				Expr:   intstr.FromString(`prometheus_operator_reconcile_operations:burnrate1h{slo="monitoring-prometheus-operator-errors"} > (2 * (1-0.99)) and prometheus_operator_reconcile_operations:burnrate12h{slo="monitoring-prometheus-operator-errors"} > (2 * (1-0.99))`),
				Labels: map[string]string{"severity": "warning", "long": "12h", "slo": "monitoring-prometheus-operator-errors", "short": "1h"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "1h30m",
				Expr:   intstr.FromString(`prometheus_operator_reconcile_operations:burnrate3h{slo="monitoring-prometheus-operator-errors"} > (1 * (1-0.99)) and prometheus_operator_reconcile_operations:burnrate2d{slo="monitoring-prometheus-operator-errors"} > (1 * (1-0.99))`),
				Labels: map[string]string{"severity": "warning", "long": "2d", "slo": "monitoring-prometheus-operator-errors", "short": "3h"},
			}},
		},
	}, {
		name: "operator-ratio-grouping",
		slo:  objectiveOperatorGrouping(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-prometheus-operator-errors",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "prometheus_operator_reconcile_operations:burnrate3m",
				Expr:   intstr.FromString(`sum by (namespace) (rate(prometheus_operator_reconcile_errors_total[3m])) / sum by (namespace) (rate(prometheus_operator_reconcile_operations_total[3m]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Record: "prometheus_operator_reconcile_operations:burnrate15m",
				Expr:   intstr.FromString(`sum by (namespace) (rate(prometheus_operator_reconcile_errors_total[15m])) / sum by (namespace) (rate(prometheus_operator_reconcile_operations_total[15m]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Record: "prometheus_operator_reconcile_operations:burnrate30m",
				Expr:   intstr.FromString(`sum by (namespace) (rate(prometheus_operator_reconcile_errors_total[30m])) / sum by (namespace) (rate(prometheus_operator_reconcile_operations_total[30m]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Record: "prometheus_operator_reconcile_operations:burnrate1h",
				Expr:   intstr.FromString(`sum by (namespace) (rate(prometheus_operator_reconcile_errors_total[1h])) / sum by (namespace) (rate(prometheus_operator_reconcile_operations_total[1h]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Record: "prometheus_operator_reconcile_operations:burnrate3h",
				Expr:   intstr.FromString(`sum by (namespace) (rate(prometheus_operator_reconcile_errors_total[3h])) / sum by (namespace) (rate(prometheus_operator_reconcile_operations_total[3h]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Record: "prometheus_operator_reconcile_operations:burnrate12h",
				Expr:   intstr.FromString(`sum by (namespace) (rate(prometheus_operator_reconcile_errors_total[12h])) / sum by (namespace) (rate(prometheus_operator_reconcile_operations_total[12h]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Record: "prometheus_operator_reconcile_operations:burnrate2d",
				Expr:   intstr.FromString(`sum by (namespace) (rate(prometheus_operator_reconcile_errors_total[2d])) / sum by (namespace) (rate(prometheus_operator_reconcile_operations_total[2d]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "1m",
				Expr:   intstr.FromString(`prometheus_operator_reconcile_operations:burnrate3m{slo="monitoring-prometheus-operator-errors"} > (14 * (1-0.99)) and prometheus_operator_reconcile_operations:burnrate30m{slo="monitoring-prometheus-operator-errors"} > (14 * (1-0.99))`),
				Labels: map[string]string{"severity": "critical", "long": "30m", "slo": "monitoring-prometheus-operator-errors", "short": "3m"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "8m",
				Expr:   intstr.FromString(`prometheus_operator_reconcile_operations:burnrate15m{slo="monitoring-prometheus-operator-errors"} > (7 * (1-0.99)) and prometheus_operator_reconcile_operations:burnrate3h{slo="monitoring-prometheus-operator-errors"} > (7 * (1-0.99))`),
				Labels: map[string]string{"severity": "critical", "long": "3h", "slo": "monitoring-prometheus-operator-errors", "short": "15m"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "30m",
				Expr:   intstr.FromString(`prometheus_operator_reconcile_operations:burnrate1h{slo="monitoring-prometheus-operator-errors"} > (2 * (1-0.99)) and prometheus_operator_reconcile_operations:burnrate12h{slo="monitoring-prometheus-operator-errors"} > (2 * (1-0.99))`),
				Labels: map[string]string{"severity": "warning", "long": "12h", "slo": "monitoring-prometheus-operator-errors", "short": "1h"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "1h30m",
				Expr:   intstr.FromString(`prometheus_operator_reconcile_operations:burnrate3h{slo="monitoring-prometheus-operator-errors"} > (1 * (1-0.99)) and prometheus_operator_reconcile_operations:burnrate2d{slo="monitoring-prometheus-operator-errors"} > (1 * (1-0.99))`),
				Labels: map[string]string{"severity": "warning", "long": "2d", "slo": "monitoring-prometheus-operator-errors", "short": "3h"},
			}},
		},
	}, {
		name: "apiserver-write-response-errors",
		slo:  objectiveAPIServerRatio(),
		rules: monitoringv1.RuleGroup{
			Name:     "apiserver-write-response-errors",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "apiserver_request:burnrate3m",
				Expr:   intstr.FromString(`sum by (verb) (rate(apiserver_request_total{code=~"5..",job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[3m])) / sum by (verb) (rate(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[3m]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-write-response-errors"},
			}, {
				Record: "apiserver_request:burnrate15m",
				Expr:   intstr.FromString(`sum by (verb) (rate(apiserver_request_total{code=~"5..",job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[15m])) / sum by (verb) (rate(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[15m]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-write-response-errors"},
			}, {
				Record: "apiserver_request:burnrate30m",
				Expr:   intstr.FromString(`sum by (verb) (rate(apiserver_request_total{code=~"5..",job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[30m])) / sum by (verb) (rate(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[30m]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-write-response-errors"},
			}, {
				Record: "apiserver_request:burnrate1h",
				Expr:   intstr.FromString(`sum by (verb) (rate(apiserver_request_total{code=~"5..",job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[1h])) / sum by (verb) (rate(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[1h]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-write-response-errors"},
			}, {
				Record: "apiserver_request:burnrate3h",
				Expr:   intstr.FromString(`sum by (verb) (rate(apiserver_request_total{code=~"5..",job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[3h])) / sum by (verb) (rate(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[3h]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-write-response-errors"},
			}, {
				Record: "apiserver_request:burnrate12h",
				Expr:   intstr.FromString(`sum by (verb) (rate(apiserver_request_total{code=~"5..",job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[12h])) / sum by (verb) (rate(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[12h]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-write-response-errors"},
			}, {
				Record: "apiserver_request:burnrate2d",
				Expr:   intstr.FromString(`sum by (verb) (rate(apiserver_request_total{code=~"5..",job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[2d])) / sum by (verb) (rate(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[2d]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-write-response-errors"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`apiserver_request:burnrate3m{job="apiserver",slo="apiserver-write-response-errors",verb=~"POST|PUT|PATCH|DELETE"} > (14 * (1-0.99)) and apiserver_request:burnrate30m{job="apiserver",slo="apiserver-write-response-errors",verb=~"POST|PUT|PATCH|DELETE"} > (14 * (1-0.99))`),
				For:    "1m",
				Labels: map[string]string{"severity": "critical", "long": "30m", "short": "3m", "job": "apiserver", "slo": "apiserver-write-response-errors"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`apiserver_request:burnrate15m{job="apiserver",slo="apiserver-write-response-errors",verb=~"POST|PUT|PATCH|DELETE"} > (7 * (1-0.99)) and apiserver_request:burnrate3h{job="apiserver",slo="apiserver-write-response-errors",verb=~"POST|PUT|PATCH|DELETE"} > (7 * (1-0.99))`),
				For:    "8m",
				Labels: map[string]string{"severity": "critical", "long": "3h", "short": "15m", "job": "apiserver", "slo": "apiserver-write-response-errors"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`apiserver_request:burnrate1h{job="apiserver",slo="apiserver-write-response-errors",verb=~"POST|PUT|PATCH|DELETE"} > (2 * (1-0.99)) and apiserver_request:burnrate12h{job="apiserver",slo="apiserver-write-response-errors",verb=~"POST|PUT|PATCH|DELETE"} > (2 * (1-0.99))`),
				For:    "30m",
				Labels: map[string]string{"severity": "warning", "long": "12h", "short": "1h", "job": "apiserver", "slo": "apiserver-write-response-errors"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`apiserver_request:burnrate3h{job="apiserver",slo="apiserver-write-response-errors",verb=~"POST|PUT|PATCH|DELETE"} > (1 * (1-0.99)) and apiserver_request:burnrate2d{job="apiserver",slo="apiserver-write-response-errors",verb=~"POST|PUT|PATCH|DELETE"} > (1 * (1-0.99))`),
				For:    "1h30m",
				Labels: map[string]string{"severity": "warning", "long": "2d", "short": "3h", "job": "apiserver", "slo": "apiserver-write-response-errors"},
			}},
		},
	}, {
		name: "apiserver-write-response-errors-alerting-disabled",
		slo:  objectiveAPIServerRatioAlertingDisabled(),
		rules: monitoringv1.RuleGroup{
			Name:     "apiserver-write-response-errors",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "apiserver_request:burnrate3m",
				Expr:   intstr.FromString(`sum by (verb) (rate(apiserver_request_total{code=~"5..",job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[3m])) / sum by (verb) (rate(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[3m]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-write-response-errors"},
			}, {
				Record: "apiserver_request:burnrate15m",
				Expr:   intstr.FromString(`sum by (verb) (rate(apiserver_request_total{code=~"5..",job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[15m])) / sum by (verb) (rate(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[15m]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-write-response-errors"},
			}, {
				Record: "apiserver_request:burnrate30m",
				Expr:   intstr.FromString(`sum by (verb) (rate(apiserver_request_total{code=~"5..",job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[30m])) / sum by (verb) (rate(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[30m]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-write-response-errors"},
			}, {
				Record: "apiserver_request:burnrate1h",
				Expr:   intstr.FromString(`sum by (verb) (rate(apiserver_request_total{code=~"5..",job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[1h])) / sum by (verb) (rate(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[1h]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-write-response-errors"},
			}, {
				Record: "apiserver_request:burnrate3h",
				Expr:   intstr.FromString(`sum by (verb) (rate(apiserver_request_total{code=~"5..",job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[3h])) / sum by (verb) (rate(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[3h]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-write-response-errors"},
			}, {
				Record: "apiserver_request:burnrate12h",
				Expr:   intstr.FromString(`sum by (verb) (rate(apiserver_request_total{code=~"5..",job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[12h])) / sum by (verb) (rate(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[12h]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-write-response-errors"},
			}, {
				Record: "apiserver_request:burnrate2d",
				Expr:   intstr.FromString(`sum by (verb) (rate(apiserver_request_total{code=~"5..",job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[2d])) / sum by (verb) (rate(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[2d]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-write-response-errors"},
			}},
		},
	}, {
		name: "apiserver-read-resource-latency",
		slo:  objectiveAPIServerLatency(),
		rules: monitoringv1.RuleGroup{
			Name:     "apiserver-read-resource-latency",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "apiserver_request_duration_seconds:burnrate3m",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[3m])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[3m]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[3m]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate15m",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[15m])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[15m]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[15m]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate30m",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[30m])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[30m]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[30m]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate1h",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[1h])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[1h]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[1h]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate3h",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[3h])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[3h]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[3h]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate12h",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[12h])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[12h]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[12h]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate2d",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[2d])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[2d]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[2d]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`apiserver_request_duration_seconds:burnrate3m{job="apiserver",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"} > (14 * (1-0.99)) and apiserver_request_duration_seconds:burnrate30m{job="apiserver",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"} > (14 * (1-0.99))`),
				For:    "1m",
				Labels: map[string]string{"severity": "critical", "long": "30m", "short": "3m", "job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`apiserver_request_duration_seconds:burnrate15m{job="apiserver",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"} > (7 * (1-0.99)) and apiserver_request_duration_seconds:burnrate3h{job="apiserver",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"} > (7 * (1-0.99))`),
				For:    "8m",
				Labels: map[string]string{"severity": "critical", "long": "3h", "short": "15m", "job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`apiserver_request_duration_seconds:burnrate1h{job="apiserver",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"} > (2 * (1-0.99)) and apiserver_request_duration_seconds:burnrate12h{job="apiserver",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"} > (2 * (1-0.99))`),
				For:    "30m",
				Labels: map[string]string{"severity": "warning", "long": "12h", "short": "1h", "job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Alert:  "ErrorBudgetBurn",
				Expr:   intstr.FromString(`apiserver_request_duration_seconds:burnrate3h{job="apiserver",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"} > (1 * (1-0.99)) and apiserver_request_duration_seconds:burnrate2d{job="apiserver",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"} > (1 * (1-0.99))`),
				For:    "1h30m",
				Labels: map[string]string{"severity": "warning", "long": "2d", "short": "3h", "job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}},
		},
	}, {
		name: "apiserver-read-resource-latency-alerting-disabled",
		slo:  objectiveAPIServerLatencyAlertingDisabled(),
		rules: monitoringv1.RuleGroup{
			Name:     "apiserver-read-resource-latency",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "apiserver_request_duration_seconds:burnrate3m",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[3m])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[3m]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[3m]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate15m",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[15m])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[15m]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[15m]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate30m",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[30m])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[30m]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[30m]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate1h",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[1h])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[1h]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[1h]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate3h",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[3h])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[3h]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[3h]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate12h",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[12h])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[12h]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[12h]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate2d",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[2d])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[2d]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[2d]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}},
		},
	}, {
		name: "apiserver-read-resource-latency-custom-alertname",
		slo:  objectiveAPIServerLatencyCustomAlertname(),
		rules: monitoringv1.RuleGroup{
			Name:     "apiserver-read-resource-latency",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "apiserver_request_duration_seconds:burnrate3m",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[3m])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[3m]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[3m]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate15m",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[15m])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[15m]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[15m]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate30m",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[30m])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[30m]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[30m]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate1h",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[1h])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[1h]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[1h]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate3h",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[3h])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[3h]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[3h]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate12h",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[12h])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[12h]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[12h]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:burnrate2d",
				Expr:   intstr.FromString(`(sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[2d])) - sum by (resource, verb) (rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[2d]))) / sum by (resource, verb) (rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[2d]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Alert:  "APIServerLatencyErrorBudgetBurn",
				Expr:   intstr.FromString(`apiserver_request_duration_seconds:burnrate3m{job="apiserver",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"} > (14 * (1-0.99)) and apiserver_request_duration_seconds:burnrate30m{job="apiserver",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"} > (14 * (1-0.99))`),
				For:    "1m",
				Labels: map[string]string{"severity": "critical", "long": "30m", "short": "3m", "job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Alert:  "APIServerLatencyErrorBudgetBurn",
				Expr:   intstr.FromString(`apiserver_request_duration_seconds:burnrate15m{job="apiserver",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"} > (7 * (1-0.99)) and apiserver_request_duration_seconds:burnrate3h{job="apiserver",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"} > (7 * (1-0.99))`),
				For:    "8m",
				Labels: map[string]string{"severity": "critical", "long": "3h", "short": "15m", "job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Alert:  "APIServerLatencyErrorBudgetBurn",
				Expr:   intstr.FromString(`apiserver_request_duration_seconds:burnrate1h{job="apiserver",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"} > (2 * (1-0.99)) and apiserver_request_duration_seconds:burnrate12h{job="apiserver",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"} > (2 * (1-0.99))`),
				For:    "30m",
				Labels: map[string]string{"severity": "warning", "long": "12h", "short": "1h", "job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Alert:  "APIServerLatencyErrorBudgetBurn",
				Expr:   intstr.FromString(`apiserver_request_duration_seconds:burnrate3h{job="apiserver",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"} > (1 * (1-0.99)) and apiserver_request_duration_seconds:burnrate2d{job="apiserver",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"} > (1 * (1-0.99))`),
				For:    "1h30m",
				Labels: map[string]string{"severity": "warning", "long": "2d", "short": "3h", "job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}},
		},
	}, {
		name: "prometheus-up-targets",
		slo:  objectiveUpTargets(),
		rules: monitoringv1.RuleGroup{
			Name:     "up-targets",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "up:burnrate5m",
				Expr:   intstr.FromString(`(sum(count_over_time(up[5m])) - sum(sum_over_time(up[5m]))) / sum(count_over_time(up[5m]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Record: "up:burnrate30m",
				Expr:   intstr.FromString(`(sum(count_over_time(up[30m])) - sum(sum_over_time(up[30m]))) / sum(count_over_time(up[30m]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Record: "up:burnrate1h",
				Expr:   intstr.FromString(`(sum(count_over_time(up[1h])) - sum(sum_over_time(up[1h]))) / sum(count_over_time(up[1h]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Record: "up:burnrate2h",
				Expr:   intstr.FromString(`(sum(count_over_time(up[2h])) - sum(sum_over_time(up[2h]))) / sum(count_over_time(up[2h]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Record: "up:burnrate6h",
				Expr:   intstr.FromString(`(sum(count_over_time(up[6h])) - sum(sum_over_time(up[6h]))) / sum(count_over_time(up[6h]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Record: "up:burnrate1d",
				Expr:   intstr.FromString(`(sum(count_over_time(up[1d])) - sum(sum_over_time(up[1d]))) / sum(count_over_time(up[1d]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Record: "up:burnrate4d",
				Expr:   intstr.FromString(`(sum(count_over_time(up[4d])) - sum(sum_over_time(up[4d]))) / sum(count_over_time(up[4d]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "2m",
				Expr:   intstr.FromString(`up:burnrate5m{slo="up-targets"} > (14 * (1-0.99)) and up:burnrate1h{slo="up-targets"} > (14 * (1-0.99))`),
				Labels: map[string]string{"severity": "critical", "long": "1h", "short": "5m", "slo": "up-targets"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "15m",
				Expr:   intstr.FromString(`up:burnrate30m{slo="up-targets"} > (7 * (1-0.99)) and up:burnrate6h{slo="up-targets"} > (7 * (1-0.99))`),
				Labels: map[string]string{"severity": "critical", "long": "6h", "slo": "up-targets", "short": "30m"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "1h",
				Expr:   intstr.FromString(`up:burnrate2h{slo="up-targets"} > (2 * (1-0.99)) and up:burnrate1d{slo="up-targets"} > (2 * (1-0.99))`),
				Labels: map[string]string{"severity": "warning", "long": "1d", "slo": "up-targets", "short": "2h"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "3h",
				Expr:   intstr.FromString(`up:burnrate6h{slo="up-targets"} > (1 * (1-0.99)) and up:burnrate4d{slo="up-targets"} > (1 * (1-0.99))`),
				Labels: map[string]string{"severity": "warning", "long": "4d", "slo": "up-targets", "short": "6h"},
			}},
		},
	}, {
		name: "prometheus-up-targets-grouping-regex",
		slo:  objectiveUpTargetsGroupingRegex(),
		rules: monitoringv1.RuleGroup{
			Name:     "up-targets",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "up:burnrate5m",
				Expr:   intstr.FromString(`(sum by (instance, job) (count_over_time(up{instance!~"(127.0.0.1|localhost).*"}[5m])) - sum by (instance, job) (sum_over_time(up{instance!~"(127.0.0.1|localhost).*"}[5m]))) / sum by (instance, job) (count_over_time(up{instance!~"(127.0.0.1|localhost).*"}[5m]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Record: "up:burnrate30m",
				Expr:   intstr.FromString(`(sum by (instance, job) (count_over_time(up{instance!~"(127.0.0.1|localhost).*"}[30m])) - sum by (instance, job) (sum_over_time(up{instance!~"(127.0.0.1|localhost).*"}[30m]))) / sum by (instance, job) (count_over_time(up{instance!~"(127.0.0.1|localhost).*"}[30m]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Record: "up:burnrate1h",
				Expr:   intstr.FromString(`(sum by (instance, job) (count_over_time(up{instance!~"(127.0.0.1|localhost).*"}[1h])) - sum by (instance, job) (sum_over_time(up{instance!~"(127.0.0.1|localhost).*"}[1h]))) / sum by (instance, job) (count_over_time(up{instance!~"(127.0.0.1|localhost).*"}[1h]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Record: "up:burnrate2h",
				Expr:   intstr.FromString(`(sum by (instance, job) (count_over_time(up{instance!~"(127.0.0.1|localhost).*"}[2h])) - sum by (instance, job) (sum_over_time(up{instance!~"(127.0.0.1|localhost).*"}[2h]))) / sum by (instance, job) (count_over_time(up{instance!~"(127.0.0.1|localhost).*"}[2h]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Record: "up:burnrate6h",
				Expr:   intstr.FromString(`(sum by (instance, job) (count_over_time(up{instance!~"(127.0.0.1|localhost).*"}[6h])) - sum by (instance, job) (sum_over_time(up{instance!~"(127.0.0.1|localhost).*"}[6h]))) / sum by (instance, job) (count_over_time(up{instance!~"(127.0.0.1|localhost).*"}[6h]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Record: "up:burnrate1d",
				Expr:   intstr.FromString(`(sum by (instance, job) (count_over_time(up{instance!~"(127.0.0.1|localhost).*"}[1d])) - sum by (instance, job) (sum_over_time(up{instance!~"(127.0.0.1|localhost).*"}[1d]))) / sum by (instance, job) (count_over_time(up{instance!~"(127.0.0.1|localhost).*"}[1d]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Record: "up:burnrate4d",
				Expr:   intstr.FromString(`(sum by (instance, job) (count_over_time(up{instance!~"(127.0.0.1|localhost).*"}[4d])) - sum by (instance, job) (sum_over_time(up{instance!~"(127.0.0.1|localhost).*"}[4d]))) / sum by (instance, job) (count_over_time(up{instance!~"(127.0.0.1|localhost).*"}[4d]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "2m",
				Expr:   intstr.FromString(`up:burnrate5m{instance!~"(127.0.0.1|localhost).*",slo="up-targets"} > (14 * (1-0.99)) and up:burnrate1h{instance!~"(127.0.0.1|localhost).*",slo="up-targets"} > (14 * (1-0.99))`),
				Labels: map[string]string{"severity": "critical", "long": "1h", "short": "5m", "slo": "up-targets"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "15m",
				Expr:   intstr.FromString(`up:burnrate30m{instance!~"(127.0.0.1|localhost).*",slo="up-targets"} > (7 * (1-0.99)) and up:burnrate6h{instance!~"(127.0.0.1|localhost).*",slo="up-targets"} > (7 * (1-0.99))`),
				Labels: map[string]string{"severity": "critical", "long": "6h", "slo": "up-targets", "short": "30m"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "1h",
				Expr:   intstr.FromString(`up:burnrate2h{instance!~"(127.0.0.1|localhost).*",slo="up-targets"} > (2 * (1-0.99)) and up:burnrate1d{instance!~"(127.0.0.1|localhost).*",slo="up-targets"} > (2 * (1-0.99))`),
				Labels: map[string]string{"severity": "warning", "long": "1d", "slo": "up-targets", "short": "2h"},
			}, {
				Alert:  "ErrorBudgetBurn",
				For:    "3h",
				Expr:   intstr.FromString(`up:burnrate6h{instance!~"(127.0.0.1|localhost).*",slo="up-targets"} > (1 * (1-0.99)) and up:burnrate4d{instance!~"(127.0.0.1|localhost).*",slo="up-targets"} > (1 * (1-0.99))`),
				Labels: map[string]string{"severity": "warning", "long": "4d", "slo": "up-targets", "short": "6h"},
			}},
		},
	}}

	require.Len(t, testcases, 20)

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			group, err := tc.slo.Burnrates()
			require.NoError(t, err)
			require.Equal(t, tc.rules, group)
		})
	}
}

func TestObjective_BurnrateNames(t *testing.T) {
	testcases := []struct {
		name     string
		slo      Objective
		burnrate string
	}{{
		name:     "http-ratio",
		slo:      objectiveHTTPRatio(),
		burnrate: "http_requests:burnrate5m",
	}, {
		name:     "http-ratio-grouping",
		slo:      objectiveHTTPRatioGrouping(),
		burnrate: "http_requests:burnrate5m",
	}, {
		name:     "http-ratio-grouping-regex",
		slo:      objectiveHTTPRatioGroupingRegex(),
		burnrate: "http_requests:burnrate5m",
	}, {
		name:     "grpc-errors",
		slo:      objectiveGRPCRatio(),
		burnrate: "grpc_server_handled:burnrate5m",
	}, {
		name:     "grpc-errors-grouping",
		slo:      objectiveGRPCRatioGrouping(),
		burnrate: "grpc_server_handled:burnrate5m",
	}, {
		name:     "http-latency",
		slo:      objectiveHTTPLatency(),
		burnrate: "http_request_duration_seconds:burnrate5m",
	}, {
		name:     "http-latency-native",
		slo:      objectiveHTTPNativeLatency(),
		burnrate: "http_request_duration_seconds:burnrate5m",
	}, {
		name:     "http-latency-grouping",
		slo:      objectiveHTTPLatencyGrouping(),
		burnrate: "http_request_duration_seconds:burnrate5m",
	}, {
		name:     "http-latency-grouping-regex",
		slo:      objectiveHTTPLatencyGroupingRegex(),
		burnrate: "http_request_duration_seconds:burnrate5m",
	}, {
		name:     "grpc-latency",
		slo:      objectiveGRPCLatency(),
		burnrate: "grpc_server_handling_seconds:burnrate5m",
	}, {
		name:     "grpc-latency-grouping",
		slo:      objectiveGRPCLatencyGrouping(),
		burnrate: "grpc_server_handling_seconds:burnrate5m",
	}, {
		name:     "operator-ratio",
		slo:      objectiveOperator(),
		burnrate: "prometheus_operator_reconcile_operations:burnrate5m",
	}, {
		name:     "operator-ratio-grouping",
		slo:      objectiveOperatorGrouping(),
		burnrate: "prometheus_operator_reconcile_operations:burnrate5m",
	}, {
		name:     "apiserver-write-response-errors",
		slo:      objectiveAPIServerRatio(),
		burnrate: "apiserver_request:burnrate5m",
	}, {
		name:     "apiserver-read-resource-latency",
		slo:      objectiveAPIServerLatency(),
		burnrate: "apiserver_request_duration_seconds:burnrate5m",
	}, {
		name:     "apiserver-read-resource-latency",
		slo:      objectiveUpTargets(),
		burnrate: "up:burnrate5m",
	}, {
		name:     "apiserver-read-resource-latency",
		slo:      objectiveUpTargetsGroupingRegex(),
		burnrate: "up:burnrate5m",
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.burnrate, tc.slo.BurnrateName(5*time.Minute))
		})
	}
}

func TestObjective_IncreaseRules(t *testing.T) {
	testcases := []struct {
		name  string
		slo   Objective
		rules monitoringv1.RuleGroup
	}{{
		name: "http-ratio",
		slo:  objectiveHTTPRatio(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-errors-increase",
			Interval: "2m30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_requests:increase4w",
				Expr:   intstr.FromString(`sum by (code) (increase(http_requests_total{job="thanos-receive-default"}[4w]))`),
				Labels: map[string]string{"job": "thanos-receive-default", "slo": "monitoring-http-errors"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(http_requests_total{job="thanos-receive-default"}) == 1`),
				For:    "2m",
				Labels: map[string]string{"job": "thanos-receive-default", "slo": "monitoring-http-errors", "severity": "critical"},
			}},
		},
	}, {
		name: "http-ratio-grouping",
		slo:  objectiveHTTPRatioGrouping(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-errors-increase",
			Interval: "2m30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_requests:increase4w",
				Expr:   intstr.FromString(`sum by (code, handler, job) (increase(http_requests_total{job="thanos-receive-default"}[4w]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(http_requests_total{job="thanos-receive-default"}) == 1`),
				For:    "2m",
				Labels: map[string]string{"slo": "monitoring-http-errors", "severity": "critical"},
			}},
		},
	}, {
		name: "http-ratio-grouping-regex",
		slo:  objectiveHTTPRatioGroupingRegex(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-errors-increase",
			Interval: "2m30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_requests:increase4w",
				Expr:   intstr.FromString(`sum by (code, handler, job) (increase(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}[4w]))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}) == 1`),
				For:    "2m",
				Labels: map[string]string{"slo": "monitoring-http-errors", "severity": "critical"},
			}},
		},
	}, {
		name: "grpc-errors",
		slo:  objectiveGRPCRatio(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-grpc-errors-increase",
			Interval: "2m30s",
			Rules: []monitoringv1.Rule{{
				Record: "grpc_server_handled:increase4w",
				Expr:   intstr.FromString(`sum by (grpc_code) (increase(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[4w]))`),
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "job": "api", "slo": "monitoring-grpc-errors"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}) == 1`),
				For:    "2m",
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "job": "api", "slo": "monitoring-grpc-errors", "severity": "critical"},
			}},
		},
	}, {
		name: "grpc-errors-grouping",
		slo:  objectiveGRPCRatioGrouping(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-grpc-errors-increase",
			Interval: "2m30s",
			Rules: []monitoringv1.Rule{{
				Record: "grpc_server_handled:increase4w",
				Expr:   intstr.FromString(`sum by (grpc_code, handler, job) (increase(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[4w]))`),
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "slo": "monitoring-grpc-errors"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}) == 1`),
				For:    "2m",
				Labels: map[string]string{"grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "slo": "monitoring-grpc-errors", "severity": "critical"},
			}},
		},
	}, {
		name: "http-latency",
		slo:  objectiveHTTPLatency(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-latency-increase",
			Interval: "2m30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_request_duration_seconds:increase4w",
				Expr:   intstr.FromString(`sum by (code) (increase(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[4w]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:increase4w",
				Expr:   intstr.FromString(`sum by (code) (increase(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[4w]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency", "le": "1"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}) == 1`),
				For:    "2m",
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency", "severity": "critical"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}) == 1`),
				For:    "2m",
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency", "le": "1", "severity": "critical"},
			}},
		},
	}, {
		name: "http-latency-native",
		slo:  objectiveHTTPNativeLatency(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-latency-increase",
			Interval: "2m30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_request_duration_seconds:increase4w",
				Expr:   intstr.FromString(`histogram_count(increase(http_request_duration_seconds{code=~"2..",job="metrics-service-thanos-receive-default"}[4w]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:increase4w",
				Expr:   intstr.FromString(`histogram_fraction(0, 1, increase(http_request_duration_seconds{code=~"2..",job="metrics-service-thanos-receive-default"}[4w])) * histogram_count(increase(http_request_duration_seconds{code=~"2..",job="metrics-service-thanos-receive-default"}[4w]))`),
				Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency", "le": "1"},
				//}, {
				//	Alert:  "SLOMetricAbsent",
				//	Expr:   intstr.FromString(`absent(http_request_duration_seconds{code=~"2..",job="metrics-service-thanos-receive-default"}) == 1`),
				//	For:    "2m",
				//	Labels: map[string]string{"job": "metrics-service-thanos-receive-default", "slo": "monitoring-http-latency", "severity": "critical"},
			}},
		},
	}, {
		name: "http-latency-grouping",
		slo:  objectiveHTTPLatencyGrouping(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-latency-increase",
			Interval: "2m30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_request_duration_seconds:increase4w",
				Expr:   intstr.FromString(`sum by (code, handler, job) (increase(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[4w]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:increase4w",
				Expr:   intstr.FromString(`sum by (code, handler, job) (increase(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[4w]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency", "le": "1"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}) == 1`),
				For:    "2m",
				Labels: map[string]string{"slo": "monitoring-http-latency", "severity": "critical"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}) == 1`),
				For:    "2m",
				Labels: map[string]string{"slo": "monitoring-http-latency", "le": "1", "severity": "critical"},
			}},
		},
	}, {
		name: "http-latency-grouping-regex",
		slo:  objectiveHTTPLatencyGroupingRegex(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-latency-increase",
			Interval: "2m30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_request_duration_seconds:increase4w",
				Expr:   intstr.FromString(`sum by (code, handler, job) (increase(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[4w]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "http_request_duration_seconds:increase4w",
				Expr:   intstr.FromString(`sum by (code, handler, job) (increase(http_request_duration_seconds_bucket{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1"}[4w]))`),
				Labels: map[string]string{"slo": "monitoring-http-latency", "le": "1"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}) == 1`),
				For:    "2m",
				Labels: map[string]string{"slo": "monitoring-http-latency", "severity": "critical"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(http_request_duration_seconds_bucket{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1"}) == 1`),
				For:    "2m",
				Labels: map[string]string{"slo": "monitoring-http-latency", "le": "1", "severity": "critical"},
			}},
		},
	}, {
		name: "grpc-latency",
		slo:  objectiveGRPCLatency(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-grpc-latency-increase",
			Interval: "1m",
			Rules: []monitoringv1.Rule{{
				Record: "grpc_server_handling_seconds:increase1w",
				Expr:   intstr.FromString(`sum(increase(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1w]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "job": "api", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Record: "grpc_server_handling_seconds:increase1w",
				Expr:   intstr.FromString(`sum(increase(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[1w]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "job": "api", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "le": "0.6"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}) == 1`),
				For:    "1m",
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "job": "api", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "severity": "critical"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}) == 1`),
				For:    "1m",
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "job": "api", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "le": "0.6", "severity": "critical"},
			}},
		},
	}, {
		name: "grpc-latency-grouping",
		slo:  objectiveGRPCLatencyGrouping(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-grpc-latency-increase",
			Interval: "1m",
			Rules: []monitoringv1.Rule{{
				Record: "grpc_server_handling_seconds:increase1w",
				Expr:   intstr.FromString(`sum by (handler, job) (increase(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1w]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore"},
			}, {
				Record: "grpc_server_handling_seconds:increase1w",
				Expr:   intstr.FromString(`sum by (handler, job) (increase(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[1w]))`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "le": "0.6"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}) == 1`),
				For:    "1m",
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "severity": "critical"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}) == 1`),
				For:    "1m",
				Labels: map[string]string{"slo": "monitoring-grpc-latency", "grpc_method": "Write", "grpc_service": "conprof.WritableProfileStore", "le": "0.6", "severity": "critical"},
			}},
		},
	}, {
		name: "operator-ratio",
		slo:  objectiveOperator(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-prometheus-operator-errors-increase",
			Interval: "1m30s",
			Rules: []monitoringv1.Rule{{
				Record: "prometheus_operator_reconcile_operations:increase2w",
				Expr:   intstr.FromString(`sum(increase(prometheus_operator_reconcile_operations_total[2w]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(prometheus_operator_reconcile_operations_total) == 1`),
				For:    "1m",
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors", "severity": "critical"},
			}, {
				Record: "prometheus_operator_reconcile_errors:increase2w",
				Expr:   intstr.FromString(`sum(increase(prometheus_operator_reconcile_errors_total[2w]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(prometheus_operator_reconcile_errors_total) == 1`),
				For:    "1m",
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors", "severity": "critical"},
			}},
		},
	}, {
		name: "operator-ratio-grouping",
		slo:  objectiveOperatorGrouping(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-prometheus-operator-errors-increase",
			Interval: "1m30s",
			Rules: []monitoringv1.Rule{{
				Record: "prometheus_operator_reconcile_operations:increase2w",
				Expr:   intstr.FromString(`sum by (namespace) (increase(prometheus_operator_reconcile_operations_total[2w]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(prometheus_operator_reconcile_operations_total) == 1`),
				For:    "1m",
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors", "severity": "critical"},
			}, {
				Record: "prometheus_operator_reconcile_errors:increase2w",
				Expr:   intstr.FromString(`sum by (namespace) (increase(prometheus_operator_reconcile_errors_total[2w]))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(prometheus_operator_reconcile_errors_total) == 1`),
				For:    "1m",
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors", "severity": "critical"},
			}},
		},
	}, {
		name: "apiserver-write-response-errors",
		slo:  objectiveAPIServerRatio(),
		rules: monitoringv1.RuleGroup{
			Name:     "apiserver-write-response-errors-increase",
			Interval: "1m30s",
			Rules: []monitoringv1.Rule{{
				Record: "apiserver_request:increase2w",
				Expr:   intstr.FromString(`sum by (code, verb) (increase(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[2w]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-write-response-errors"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}) == 1`),
				For:    "1m",
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-write-response-errors", "severity": "critical"},
			}},
		},
	}, {
		name: "apiserver-read-resource-latency",
		slo:  objectiveAPIServerLatency(),
		rules: monitoringv1.RuleGroup{
			Name:     "apiserver-read-resource-latency-increase",
			Interval: "1m30s",
			Rules: []monitoringv1.Rule{{
				Record: "apiserver_request_duration_seconds:increase2w",
				Expr:   intstr.FromString(`sum by (resource, verb) (increase(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[2w]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency"},
			}, {
				Record: "apiserver_request_duration_seconds:increase2w",
				Expr:   intstr.FromString(`sum by (resource, verb) (increase(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[2w]))`),
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency", "le": "0.1"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}) == 1`),
				For:    "1m",
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency", "severity": "critical"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}) == 1`),
				For:    "1m",
				Labels: map[string]string{"job": "apiserver", "slo": "apiserver-read-resource-latency", "le": "0.1", "severity": "critical"},
			}},
		},
	}, {
		name: "prometheus-up-targets",
		slo:  objectiveUpTargets(),
		rules: monitoringv1.RuleGroup{
			Name:     "up-targets-increase",
			Interval: "2m30s",
			Rules: []monitoringv1.Rule{{
				Record: "up:count4w",
				Expr:   intstr.FromString(`sum(count_over_time(up[4w]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Record: "up:sum4w",
				Expr:   intstr.FromString(`sum(sum_over_time(up[4w]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(up) == 1`),
				For:    "2m",
				Labels: map[string]string{"severity": "critical", "slo": "up-targets"},
			}},
		},
	}, {
		name: "prometheus-up-targets-grouping-regex",
		slo:  objectiveUpTargetsGroupingRegex(),
		rules: monitoringv1.RuleGroup{
			Name:     "up-targets-increase",
			Interval: "2m30s",
			Rules: []monitoringv1.Rule{{
				Record: "up:count4w",
				Expr:   intstr.FromString(`sum by (instance, job) (count_over_time(up{instance!~"(127.0.0.1|localhost).*"}[4w]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Record: "up:sum4w",
				Expr:   intstr.FromString(`sum by (instance, job) (sum_over_time(up{instance!~"(127.0.0.1|localhost).*"}[4w]))`),
				Labels: map[string]string{"slo": "up-targets"},
			}, {
				Alert:  "SLOMetricAbsent",
				Expr:   intstr.FromString(`absent(up{instance!~"(127.0.0.1|localhost).*"}) == 1`),
				For:    "2m",
				Labels: map[string]string{"severity": "critical", "slo": "up-targets"},
			}},
		},
	}}

	require.Len(t, testcases, 17)

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			group, err := tc.slo.IncreaseRules()
			require.NoError(t, err)
			require.Equal(t, tc.rules, group)
		})
	}
}

func Test_windows(t *testing.T) {
	ws := Windows(28 * 24 * time.Hour)

	require.Equal(t, Window{
		Severity: critical,
		For:      2 * time.Minute,
		Long:     1 * time.Hour,
		Short:    5 * time.Minute,
		Factor:   14,
	}, ws[0])

	require.Equal(t, Window{
		Severity: critical,
		For:      15 * time.Minute,
		Long:     6 * time.Hour,
		Short:    30 * time.Minute,
		Factor:   7,
	}, ws[1])

	require.Equal(t, Window{
		Severity: warning,
		For:      time.Hour,
		Long:     24 * time.Hour,
		Short:    2 * time.Hour,
		Factor:   2,
	}, ws[2])

	require.Equal(t, Window{
		Severity: warning,
		For:      3 * time.Hour,
		Long:     4 * 24 * time.Hour,
		Short:    6 * time.Hour,
		Factor:   1,
	}, ws[3])
}

func TestObjective_GrafanaRules(t *testing.T) {
	testcases := []struct {
		name  string
		slo   Objective
		rules monitoringv1.RuleGroup
		err   error
	}{{
		name: "http-ratio",
		slo:  objectiveHTTPRatio(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-errors-generic",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "pyrra_objective",
				Expr:   intstr.FromString(`0.99`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "pyrra_window",
				Expr:   intstr.FromInt(int((28 * 24 * time.Hour).Seconds())),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "pyrra_availability",
				Expr:   intstr.FromString(`1 - sum(http_requests:increase4w{code=~"5..",job="thanos-receive-default",slo="monitoring-http-errors"} or vector(0)) / sum(http_requests:increase4w{job="thanos-receive-default",slo="monitoring-http-errors"})`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "pyrra_requests_total",
				Expr:   intstr.FromString(`sum(http_requests_total{job="thanos-receive-default"})`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}, {
				Record: "pyrra_errors_total",
				Expr:   intstr.FromString(`sum(http_requests_total{code=~"5..",job="thanos-receive-default"} or vector(0))`),
				Labels: map[string]string{"slo": "monitoring-http-errors"},
			}},
		},
	}, {
		name: "http-ratio-grouping",
		slo:  objectiveHTTPRatioGrouping(),
		err:  ErrGroupingUnsupported,
	}, {
		name: "http-ratio-grouping-regex",
		slo:  objectiveHTTPRatioGroupingRegex(),
		err:  ErrGroupingUnsupported,
	}, {
		name: "grpc-errors",
		slo:  objectiveGRPCRatio(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-grpc-errors-generic",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "pyrra_objective",
				Expr:   intstr.FromString(`0.999`),
				Labels: map[string]string{"slo": "monitoring-grpc-errors"},
			}, {
				Record: "pyrra_window",
				Expr:   intstr.FromInt(int((28 * 24 * time.Hour).Seconds())),
				Labels: map[string]string{"slo": "monitoring-grpc-errors"},
			}, {
				Record: "pyrra_availability",
				Expr:   intstr.FromString(`1 - sum(grpc_server_handled:increase4w{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} or vector(0)) / sum(grpc_server_handled:increase4w{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"})`),
				Labels: map[string]string{"slo": "monitoring-grpc-errors"},
			}, {
				Record: "pyrra_requests_total",
				Expr:   intstr.FromString(`sum(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"})`),
				Labels: map[string]string{"slo": "monitoring-grpc-errors"},
			}, {
				Record: "pyrra_errors_total",
				Expr:   intstr.FromString(`sum(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"} or vector(0))`),
				Labels: map[string]string{"slo": "monitoring-grpc-errors"},
			}},
		},
	}, {
		name: "grpc-errors-grouping",
		slo:  objectiveGRPCRatioGrouping(),
		err:  ErrGroupingUnsupported,
	}, {
		name: "http-latency",
		slo:  objectiveHTTPLatency(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-http-latency-generic",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "pyrra_objective",
				Expr:   intstr.FromString("0.995"),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "pyrra_window",
				Expr:   intstr.FromInt(2419200),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "pyrra_availability",
				Expr:   intstr.FromString(`sum(http_request_duration_seconds:increase4w{code=~"2..",job="metrics-service-thanos-receive-default",le="1",slo="monitoring-http-latency"} or vector(0)) / sum(http_request_duration_seconds:increase4w{code=~"2..",job="metrics-service-thanos-receive-default",le="",slo="monitoring-http-latency"})`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "pyrra_requests_total",
				Expr:   intstr.FromString(`sum(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"})`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}, {
				Record: "pyrra_errors_total",
				Expr:   intstr.FromString(`sum(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}) - sum(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"})`),
				Labels: map[string]string{"slo": "monitoring-http-latency"},
			}},
		},
	}, {
		name: "http-latency-grouping",
		slo:  objectiveHTTPLatencyGrouping(),
		err:  ErrGroupingUnsupported,
	}, {
		name: "http-latency-grouping-regex",
		slo:  objectiveHTTPLatencyGroupingRegex(),
		err:  ErrGroupingUnsupported,
	}, {
		name: "grpc-latency",
		slo:  objectiveGRPCLatency(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-grpc-latency-generic",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "pyrra_objective",
				Expr:   intstr.FromString("0.995"),
				Labels: map[string]string{"slo": "monitoring-grpc-latency"},
			}, {
				Record: "pyrra_window",
				Expr:   intstr.FromInt(604800),
				Labels: map[string]string{"slo": "monitoring-grpc-latency"},
			}, {
				Record: "pyrra_availability",
				Expr:   intstr.FromString(`sum(grpc_server_handling_seconds:increase1w{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6",slo="monitoring-grpc-latency"} or vector(0)) / sum(grpc_server_handling_seconds:increase1w{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="",slo="monitoring-grpc-latency"})`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency"},
			}, {
				Record: "pyrra_requests_total",
				Expr:   intstr.FromString(`sum(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"})`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency"},
			}, {
				Record: "pyrra_errors_total",
				Expr:   intstr.FromString(`sum(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}) - sum(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"})`),
				Labels: map[string]string{"slo": "monitoring-grpc-latency"},
			}},
		},
	}, {
		name: "grpc-latency-grouping",
		slo:  objectiveGRPCLatencyGrouping(),
		err:  ErrGroupingUnsupported,
	}, {
		name: "operator-ratio",
		slo:  objectiveOperator(),
		rules: monitoringv1.RuleGroup{
			Name:     "monitoring-prometheus-operator-errors-generic",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "pyrra_objective",
				Expr:   intstr.FromString("0.99"),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Record: "pyrra_window",
				Expr:   intstr.FromInt(1209600),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Record: "pyrra_availability",
				Expr:   intstr.FromString(`1 - sum(prometheus_operator_reconcile_errors:increase2w{slo="monitoring-prometheus-operator-errors"} or vector(0)) / sum(prometheus_operator_reconcile_operations:increase2w{slo="monitoring-prometheus-operator-errors"})`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Record: "pyrra_requests_total",
				Expr:   intstr.FromString(`sum(prometheus_operator_reconcile_operations_total)`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}, {
				Record: "pyrra_errors_total",
				Expr:   intstr.FromString(`sum(prometheus_operator_reconcile_errors_total or vector(0))`),
				Labels: map[string]string{"slo": "monitoring-prometheus-operator-errors"},
			}},
		},
	}, {
		name: "operator-ratio-grouping",
		slo:  objectiveOperatorGrouping(),
		err:  ErrGroupingUnsupported,
	}, {
		name: "apiserver-write-response-errors",
		slo:  objectiveAPIServerRatio(),
		rules: monitoringv1.RuleGroup{
			Name:     "apiserver-write-response-errors-generic",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "pyrra_objective",
				Expr:   intstr.FromString(`0.99`),
				Labels: map[string]string{"slo": "apiserver-write-response-errors"},
			}, {
				Record: "pyrra_window",
				Expr:   intstr.FromInt(int((14 * 24 * time.Hour).Seconds())),
				Labels: map[string]string{"slo": "apiserver-write-response-errors"},
			}, {
				Record: "pyrra_availability",
				Expr:   intstr.FromString(`1 - sum(apiserver_request:increase2w{code=~"5..",job="apiserver",slo="apiserver-write-response-errors",verb=~"POST|PUT|PATCH|DELETE"} or vector(0)) / sum(apiserver_request:increase2w{job="apiserver",slo="apiserver-write-response-errors",verb=~"POST|PUT|PATCH|DELETE"})`),
				Labels: map[string]string{"slo": "apiserver-write-response-errors"},
			}, {
				Record: "pyrra_requests_total",
				Expr:   intstr.FromString(`sum(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"})`),
				Labels: map[string]string{"slo": "apiserver-write-response-errors"},
			}, {
				Record: "pyrra_errors_total",
				Expr:   intstr.FromString(`sum(apiserver_request_total{code=~"5..",job="apiserver",verb=~"POST|PUT|PATCH|DELETE"} or vector(0))`),
				Labels: map[string]string{"slo": "apiserver-write-response-errors"},
			}},
		},
	}, {
		name: "apiserver-read-resource-latency",
		slo:  objectiveAPIServerLatency(),
		err:  ErrGroupingUnsupported,
	}}

	require.Len(t, testcases, 14)

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			group, err := tc.slo.GenericRules()
			if tc.err != nil {
				require.Error(t, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.rules, group)
			}
		})
	}
}
