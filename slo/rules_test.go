package slo

import (
	"testing"
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestObjective_Burnrates(t *testing.T) {
	testcases := []struct {
		name  string
		slo   Objective
		rules monitoringv1.RuleGroup
	}{{
		name: "http",
		slo: Objective{
			Name:   "http-users-errors",
			Target: 0.999,
			Window: model.Duration(28 * 24 * time.Hour),
			Indicator: Indicator{
				HTTP: &HTTPIndicator{
					Matchers: []*labels.Matcher{
						{Type: labels.MatchEqual, Name: "job", Value: "api"},
						{Type: labels.MatchEqual, Name: "handler", Value: "/users"},
					},
				},
			},
		},
		rules: monitoringv1.RuleGroup{
			Name:     "http-users-errors",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_requests:burnrate5m",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(http_requests_total{code=~"5..",handler="/users",job="api"}[5m])) / sum(rate(http_requests_total{handler="/users",job="api"}[5m]))`,
				},
				Labels: map[string]string{"handler": "/users", "job": "api", "slo": "http-users-errors"},
			}, {
				Record: "http_requests:burnrate30m",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(http_requests_total{code=~"5..",handler="/users",job="api"}[30m])) / sum(rate(http_requests_total{handler="/users",job="api"}[30m]))`,
				},
				Labels: map[string]string{"handler": "/users", "job": "api", "slo": "http-users-errors"},
			}, {
				Record: "http_requests:burnrate1h",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(http_requests_total{code=~"5..",handler="/users",job="api"}[1h])) / sum(rate(http_requests_total{handler="/users",job="api"}[1h]))`,
				},
				Labels: map[string]string{"handler": "/users", "job": "api", "slo": "http-users-errors"},
			}, {
				Record: "http_requests:burnrate2h",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(http_requests_total{code=~"5..",handler="/users",job="api"}[2h])) / sum(rate(http_requests_total{handler="/users",job="api"}[2h]))`,
				},
				Labels: map[string]string{"handler": "/users", "job": "api", "slo": "http-users-errors"},
			}, {
				Record: "http_requests:burnrate6h",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(http_requests_total{code=~"5..",handler="/users",job="api"}[6h])) / sum(rate(http_requests_total{handler="/users",job="api"}[6h]))`,
				},
				Labels: map[string]string{"handler": "/users", "job": "api", "slo": "http-users-errors"},
			}, {
				Record: "http_requests:burnrate1d",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(http_requests_total{code=~"5..",handler="/users",job="api"}[1d])) / sum(rate(http_requests_total{handler="/users",job="api"}[1d]))`,
				},
				Labels: map[string]string{"handler": "/users", "job": "api", "slo": "http-users-errors"},
			}, {
				Record: "http_requests:burnrate4d",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(http_requests_total{code=~"5..",handler="/users",job="api"}[4d])) / sum(rate(http_requests_total{handler="/users",job="api"}[4d]))`,
				},
				Labels: map[string]string{"handler": "/users", "job": "api", "slo": "http-users-errors"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "2m",
				Expr:        intstr.FromString(`http_requests:burnrate5m{handler="/users",job="api",slo="http-users-errors"} > (14 * (1-0.99900)) and http_requests:burnrate1h{handler="/users",job="api",slo="http-users-errors"} > (14 * (1-0.99900))`),
				Annotations: map[string]string{"severity": "critical"},
				Labels:      map[string]string{"handler": "/users", "job": "api", "long": "1h", "slo": "http-users-errors", "short": "5m"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "15m",
				Expr:        intstr.FromString(`http_requests:burnrate30m{handler="/users",job="api",slo="http-users-errors"} > (7 * (1-0.99900)) and http_requests:burnrate6h{handler="/users",job="api",slo="http-users-errors"} > (7 * (1-0.99900))`),
				Annotations: map[string]string{"severity": "critical"},
				Labels:      map[string]string{"handler": "/users", "job": "api", "long": "6h", "slo": "http-users-errors", "short": "30m"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "1h",
				Expr:        intstr.FromString(`http_requests:burnrate2h{handler="/users",job="api",slo="http-users-errors"} > (2 * (1-0.99900)) and http_requests:burnrate1d{handler="/users",job="api",slo="http-users-errors"} > (2 * (1-0.99900))`),
				Annotations: map[string]string{"severity": "warning"},
				Labels:      map[string]string{"handler": "/users", "job": "api", "long": "1d", "slo": "http-users-errors", "short": "2h"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "3h",
				Expr:        intstr.FromString(`http_requests:burnrate6h{handler="/users",job="api",slo="http-users-errors"} > (1 * (1-0.99900)) and http_requests:burnrate4d{handler="/users",job="api",slo="http-users-errors"} > (1 * (1-0.99900))`),
				Annotations: map[string]string{"severity": "warning"},
				Labels:      map[string]string{"handler": "/users", "job": "api", "long": "4d", "slo": "http-users-errors", "short": "6h"},
			}},
		},
		//}, {
		//slo: athenev1alpha1.ServiceLevelObjective{
		//	ObjectMeta: metav1.ObjectMeta{
		//		Name: "http-users-latency",
		//	},
		//	Spec: athenev1alpha1.ServiceLevelObjectiveSpec{
		//		Target:  "99",
		//		Latency: "1",
		//		Window:  metav1.Duration{Duration: 28 * 24 * time.Hour},
		//		ServiceLevelIndicator: athenev1alpha1.ServiceLevelIndicator{
		//			HTTP: &athenev1alpha1.HTTPIndicator{
		//				Matchers: []string{`job="api"`, `handler="/users"`},
		//			},
		//		},
		//	},
		//},
		//rules: &monitoringv1.RuleGroup{
		//	Name:     "http-users-latency-http-rules",
		//	Interval: "30s",
		//	Rules: []monitoringv1.Rule{{
		//		Record: "http_request_duration_seconds:burnrate5m",
		//		Expr: intstr.IntOrString{
		//			Type:   intstr.String,
		//			StrVal: "sum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\",code=~\"5..\"}[5m]))\n/\nsum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\"}[5m]))",
		//		},
		//	}, {
		//		Record: "http_request_duration_seconds:burnrate30m",
		//		Expr: intstr.IntOrString{
		//			Type:   intstr.String,
		//			StrVal: "sum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\",code=~\"5..\"}[30m]))\n/\nsum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\"}[30m]))",
		//		},
		//	}, {
		//		Record: "http_request_duration_seconds:burnrate1h",
		//		Expr: intstr.IntOrString{
		//			Type:   intstr.String,
		//			StrVal: "sum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\",code=~\"5..\"}[1h]))\n/\nsum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\"}[1h]))",
		//		},
		//	}, {
		//		Record: "http_request_duration_seconds:burnrate2h",
		//		Expr: intstr.IntOrString{
		//			Type:   intstr.String,
		//			StrVal: "sum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\",code=~\"5..\"}[2h]))\n/\nsum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\"}[2h]))",
		//		},
		//	}, {
		//		Record: "http_request_duration_seconds:burnrate6h",
		//		Expr: intstr.IntOrString{
		//			Type:   intstr.String,
		//			StrVal: "sum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\",code=~\"5..\"}[6h]))\n/\nsum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\"}[6h]))",
		//		},
		//	}, {
		//		Record: "http_request_duration_seconds:burnrate1d",
		//		Expr: intstr.IntOrString{
		//			Type:   intstr.String,
		//			StrVal: "sum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\",code=~\"5..\"}[1d]))\n/\nsum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\"}[1d]))",
		//		},
		//	}, {
		//		Record: "http_request_duration_seconds:burnrate4d",
		//		Expr: intstr.IntOrString{
		//			Type:   intstr.String,
		//			StrVal: "sum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\",code=~\"5..\"}[4d]))\n/\nsum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\"}[4d]))",
		//		},
		//	}, {
		//		Alert:       "ErrorBudgetBurn",
		//		For:         "2m",
		//		Expr:        intstr.FromString("http_request_duration_seconds:burnrate5m > (14 * (100-99)/100) and http_request_duration_seconds:burnrate1h > (14 * (100-99)/100)"),
		//		Annotations: map[string]string{"severity": "critical"},
		//	}, {
		//		Alert:       "ErrorBudgetBurn",
		//		For:         "15m",
		//		Expr:        intstr.FromString("http_request_duration_seconds:burnrate30m > (7 * (100-99)/100) and http_request_duration_seconds:burnrate6h > (7 * (100-99)/100)"),
		//		Annotations: map[string]string{"severity": "critical"},
		//	}, {
		//		Alert:       "ErrorBudgetBurn",
		//		For:         "1h",
		//		Expr:        intstr.FromString("http_request_duration_seconds:burnrate2h > (2 * (100-99)/100) and http_request_duration_seconds:burnrate1d > (2 * (100-99)/100)"),
		//		Annotations: map[string]string{"severity": "warning"},
		//	}, {
		//		Alert:       "ErrorBudgetBurn",
		//		For:         "3h",
		//		Expr:        intstr.FromString("http_request_duration_seconds:burnrate6h > (1 * (100-99)/100) and http_request_duration_seconds:burnrate4d > (1 * (100-99)/100)"),
		//		Annotations: map[string]string{"severity": "warning"},
		//	}},
		//},
	}, {
		name: "grpc",
		slo: Objective{
			Name:   "grpc-users-errors",
			Window: model.Duration(28 * 24 * time.Hour),
			Target: 0.999,
			Indicator: Indicator{
				GRPC: &GRPCIndicator{
					Service: "example.api",
					Method:  "Users",
				},
			},
		},
		rules: monitoringv1.RuleGroup{
			Name:     "grpc-users-errors",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "grpc_server_handled:burnrate5m",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Users",grpc_service="example.api"}[5m])) / sum(rate(grpc_server_handled_total{grpc_method="Users",grpc_service="example.api"}[5m]))`,
				},
				Labels: map[string]string{"grpc_method": "Users", "grpc_service": "example.api", "slo": "grpc-users-errors"},
			}, {
				Record: "grpc_server_handled:burnrate30m",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Users",grpc_service="example.api"}[30m])) / sum(rate(grpc_server_handled_total{grpc_method="Users",grpc_service="example.api"}[30m]))`,
				},
				Labels: map[string]string{"grpc_method": "Users", "grpc_service": "example.api", "slo": "grpc-users-errors"},
			}, {
				Record: "grpc_server_handled:burnrate1h",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Users",grpc_service="example.api"}[1h])) / sum(rate(grpc_server_handled_total{grpc_method="Users",grpc_service="example.api"}[1h]))`,
				},
				Labels: map[string]string{"grpc_method": "Users", "grpc_service": "example.api", "slo": "grpc-users-errors"},
			}, {
				Record: "grpc_server_handled:burnrate2h",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Users",grpc_service="example.api"}[2h])) / sum(rate(grpc_server_handled_total{grpc_method="Users",grpc_service="example.api"}[2h]))`,
				},
				Labels: map[string]string{"grpc_method": "Users", "grpc_service": "example.api", "slo": "grpc-users-errors"},
			}, {
				Record: "grpc_server_handled:burnrate6h",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Users",grpc_service="example.api"}[6h])) / sum(rate(grpc_server_handled_total{grpc_method="Users",grpc_service="example.api"}[6h]))`,
				},
				Labels: map[string]string{"grpc_method": "Users", "grpc_service": "example.api", "slo": "grpc-users-errors"},
			}, {
				Record: "grpc_server_handled:burnrate1d",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Users",grpc_service="example.api"}[1d])) / sum(rate(grpc_server_handled_total{grpc_method="Users",grpc_service="example.api"}[1d]))`,
				},
				Labels: map[string]string{"grpc_method": "Users", "grpc_service": "example.api", "slo": "grpc-users-errors"},
			}, {
				Record: "grpc_server_handled:burnrate4d",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: `sum(rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Users",grpc_service="example.api"}[4d])) / sum(rate(grpc_server_handled_total{grpc_method="Users",grpc_service="example.api"}[4d]))`,
				},
				Labels: map[string]string{"grpc_method": "Users", "grpc_service": "example.api", "slo": "grpc-users-errors"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "2m",
				Expr:        intstr.FromString(`grpc_server_handled:burnrate5m{grpc_method="Users",grpc_service="example.api",slo="grpc-users-errors"} > (14 * (1-0.99900)) and grpc_server_handled:burnrate1h{grpc_method="Users",grpc_service="example.api",slo="grpc-users-errors"} > (14 * (1-0.99900))`),
				Annotations: map[string]string{"severity": "critical"},
				Labels:      map[string]string{"grpc_method": "Users", "grpc_service": "example.api", "long": "1h", "slo": "grpc-users-errors", "short": "5m"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "15m",
				Expr:        intstr.FromString(`grpc_server_handled:burnrate30m{grpc_method="Users",grpc_service="example.api",slo="grpc-users-errors"} > (7 * (1-0.99900)) and grpc_server_handled:burnrate6h{grpc_method="Users",grpc_service="example.api",slo="grpc-users-errors"} > (7 * (1-0.99900))`),
				Annotations: map[string]string{"severity": "critical"},
				Labels:      map[string]string{"grpc_method": "Users", "grpc_service": "example.api", "long": "6h", "slo": "grpc-users-errors", "short": "30m"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "1h",
				Expr:        intstr.FromString(`grpc_server_handled:burnrate2h{grpc_method="Users",grpc_service="example.api",slo="grpc-users-errors"} > (2 * (1-0.99900)) and grpc_server_handled:burnrate1d{grpc_method="Users",grpc_service="example.api",slo="grpc-users-errors"} > (2 * (1-0.99900))`),
				Annotations: map[string]string{"severity": "warning"},
				Labels:      map[string]string{"grpc_method": "Users", "grpc_service": "example.api", "long": "1d", "slo": "grpc-users-errors", "short": "2h"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "3h",
				Expr:        intstr.FromString(`grpc_server_handled:burnrate6h{grpc_method="Users",grpc_service="example.api",slo="grpc-users-errors"} > (1 * (1-0.99900)) and grpc_server_handled:burnrate4d{grpc_method="Users",grpc_service="example.api",slo="grpc-users-errors"} > (1 * (1-0.99900))`),
				Annotations: map[string]string{"severity": "warning"},
				Labels:      map[string]string{"grpc_method": "Users", "grpc_service": "example.api", "long": "4d", "slo": "grpc-users-errors", "short": "6h"},
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
