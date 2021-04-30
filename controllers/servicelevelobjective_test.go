package controllers

import (
	"testing"
	"time"

	athenav1alpha1 "github.com/metalmatze/athena/api/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func Test_makeHTTPRules(t *testing.T) {
	testcases := []struct {
		slo   athenav1alpha1.ServiceLevelObjective
		rules *monitoringv1.RuleGroup
	}{{
		slo: athenav1alpha1.ServiceLevelObjective{
			ObjectMeta: metav1.ObjectMeta{
				Name: "http-users-errors",
			},
			Spec: athenav1alpha1.ServiceLevelObjectiveSpec{
				Target: "99.9",
				Window: metav1.Duration{Duration: 28 * 24 * time.Hour},
				ServiceLevelIndicator: athenav1alpha1.ServiceLevelIndicator{
					HTTP: &athenav1alpha1.HTTPIndicator{
						Selectors: []string{`job="api"`, `handler="/users"`},
					},
				},
			},
		},
		rules: &monitoringv1.RuleGroup{
			Name:     "http-users-errors-http-rules",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_requests:burnrate5m",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "sum(rate(http_requests_total{job=\"api\",handler=\"/users\",code=~\"5..\"}[5m]))\n/\nsum(rate(http_requests_total{job=\"api\",handler=\"/users\"}[5m]))",
				},
			}, {
				Record: "http_requests:burnrate30m",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "sum(rate(http_requests_total{job=\"api\",handler=\"/users\",code=~\"5..\"}[30m]))\n/\nsum(rate(http_requests_total{job=\"api\",handler=\"/users\"}[30m]))",
				},
			}, {
				Record: "http_requests:burnrate1h",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "sum(rate(http_requests_total{job=\"api\",handler=\"/users\",code=~\"5..\"}[1h]))\n/\nsum(rate(http_requests_total{job=\"api\",handler=\"/users\"}[1h]))",
				},
			}, {
				Record: "http_requests:burnrate2h",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "sum(rate(http_requests_total{job=\"api\",handler=\"/users\",code=~\"5..\"}[2h]))\n/\nsum(rate(http_requests_total{job=\"api\",handler=\"/users\"}[2h]))",
				},
			}, {
				Record: "http_requests:burnrate6h",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "sum(rate(http_requests_total{job=\"api\",handler=\"/users\",code=~\"5..\"}[6h]))\n/\nsum(rate(http_requests_total{job=\"api\",handler=\"/users\"}[6h]))",
				},
			}, {
				Record: "http_requests:burnrate1d",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "sum(rate(http_requests_total{job=\"api\",handler=\"/users\",code=~\"5..\"}[1d]))\n/\nsum(rate(http_requests_total{job=\"api\",handler=\"/users\"}[1d]))",
				},
			}, {
				Record: "http_requests:burnrate4d",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "sum(rate(http_requests_total{job=\"api\",handler=\"/users\",code=~\"5..\"}[4d]))\n/\nsum(rate(http_requests_total{job=\"api\",handler=\"/users\"}[4d]))",
				},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "2m",
				Expr:        intstr.FromString("http_requests:burnrate5m > (14 * (100-99.9)/100) and http_requests:burnrate1h > (14 * (100-99.9)/100)"),
				Annotations: map[string]string{"severity": "critical"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "15m",
				Expr:        intstr.FromString("http_requests:burnrate30m > (7 * (100-99.9)/100) and http_requests:burnrate6h > (7 * (100-99.9)/100)"),
				Annotations: map[string]string{"severity": "critical"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "1h",
				Expr:        intstr.FromString("http_requests:burnrate2h > (2 * (100-99.9)/100) and http_requests:burnrate1d > (2 * (100-99.9)/100)"),
				Annotations: map[string]string{"severity": "warning"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "3h",
				Expr:        intstr.FromString("http_requests:burnrate6h > (1 * (100-99.9)/100) and http_requests:burnrate4d > (1 * (100-99.9)/100)"),
				Annotations: map[string]string{"severity": "warning"},
			}},
		},
	}, {
		slo: athenav1alpha1.ServiceLevelObjective{
			ObjectMeta: metav1.ObjectMeta{
				Name: "http-users-latency",
			},
			Spec: athenav1alpha1.ServiceLevelObjectiveSpec{
				Target:  "99",
				Latency: "1",
				Window:  metav1.Duration{Duration: 28 * 24 * time.Hour},
				ServiceLevelIndicator: athenav1alpha1.ServiceLevelIndicator{
					HTTP: &athenav1alpha1.HTTPIndicator{
						Selectors: []string{`job="api"`, `handler="/users"`},
					},
				},
			},
		},
		rules: &monitoringv1.RuleGroup{
			Name:     "http-users-latency-http-rules",
			Interval: "30s",
			Rules: []monitoringv1.Rule{{
				Record: "http_request_duration_seconds:burnrate5m",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "1-(sum(rate(http_request_duration_seconds_bucket{namespace=\"default\",job=\"fooapp\",le=\"1\",code!~\"5..\"}[5m]))/sum(rate(http_request_duration_seconds_count{namespace=\"default\",job=\"fooapp\"}[5m])))",
				},
			}, {
				Record: "http_request_duration_seconds:burnrate30m",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "sum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\",code=~\"5..\"}[30m]))\n/\nsum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\"}[30m]))",
				},
			}, {
				Record: "http_request_duration_seconds:burnrate1h",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "sum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\",code=~\"5..\"}[1h]))\n/\nsum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\"}[1h]))",
				},
			}, {
				Record: "http_request_duration_seconds:burnrate2h",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "sum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\",code=~\"5..\"}[2h]))\n/\nsum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\"}[2h]))",
				},
			}, {
				Record: "http_request_duration_seconds:burnrate6h",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "sum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\",code=~\"5..\"}[6h]))\n/\nsum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\"}[6h]))",
				},
			}, {
				Record: "http_request_duration_seconds:burnrate1d",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "sum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\",code=~\"5..\"}[1d]))\n/\nsum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\"}[1d]))",
				},
			}, {
				Record: "http_request_duration_seconds:burnrate4d",
				Expr: intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "sum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\",code=~\"5..\"}[4d]))\n/\nsum(rate(http_request_duration_seconds_bucket{job=\"api\",handler=\"/users\"}[4d]))",
				},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "2m",
				Expr:        intstr.FromString("http_request_duration_seconds:burnrate5m > (14 * (100-99)/100) and http_request_duration_seconds:burnrate1h > (14 * (100-99)/100)"),
				Annotations: map[string]string{"severity": "critical"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "15m",
				Expr:        intstr.FromString("http_request_duration_seconds:burnrate30m > (7 * (100-99)/100) and http_request_duration_seconds:burnrate6h > (7 * (100-99)/100)"),
				Annotations: map[string]string{"severity": "critical"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "1h",
				Expr:        intstr.FromString("http_request_duration_seconds:burnrate2h > (2 * (100-99)/100) and http_request_duration_seconds:burnrate1d > (2 * (100-99)/100)"),
				Annotations: map[string]string{"severity": "warning"},
			}, {
				Alert:       "ErrorBudgetBurn",
				For:         "3h",
				Expr:        intstr.FromString("http_request_duration_seconds:burnrate6h > (1 * (100-99)/100) and http_request_duration_seconds:burnrate4d > (1 * (100-99)/100)"),
				Annotations: map[string]string{"severity": "warning"},
			}},
		},
	}}

	for _, tc := range testcases {
		group, err := makeHTTPRules(tc.slo)
		require.NoError(t, err)
		require.Equal(t, tc.rules, group)
	}
}

func Test_makeGRPCRules(t *testing.T) {
	slo := athenav1alpha1.ServiceLevelObjective{
		ObjectMeta: metav1.ObjectMeta{
			Name: "grpc-users-errors",
		},
		Spec: athenav1alpha1.ServiceLevelObjectiveSpec{
			Target: "99.9",
			Window: metav1.Duration{Duration: 28 * 24 * time.Hour},
			ServiceLevelIndicator: athenav1alpha1.ServiceLevelIndicator{
				GRPC: &athenav1alpha1.GRPCIndicator{
					Service: "example.api",
					Method:  "Users",
				},
			},
		},
	}

	expected := &monitoringv1.RuleGroup{
		Name:     "grpc-users-errors-grpc-rules",
		Interval: "30s",
		Rules: []monitoringv1.Rule{{
			Record: "grpc_server_handled:burnrate5m",
			Expr: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "sum(rate(grpc_server_handled_total{grpc_method=\"Users\",grpc_service=\"example.api\"}[5m]))\n/\nsum(rate(grpc_server_handled_total{grpc_method=\"Users\",grpc_service=\"example.api\"}[5m]))",
			},
		}, {
			Record: "grpc_server_handled:burnrate30m",
			Expr: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "sum(rate(grpc_server_handled_total{grpc_method=\"Users\",grpc_service=\"example.api\"}[30m]))\n/\nsum(rate(grpc_server_handled_total{grpc_method=\"Users\",grpc_service=\"example.api\"}[30m]))",
			},
		}, {
			Record: "grpc_server_handled:burnrate1h",
			Expr: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "sum(rate(grpc_server_handled_total{grpc_method=\"Users\",grpc_service=\"example.api\"}[1h]))\n/\nsum(rate(grpc_server_handled_total{grpc_method=\"Users\",grpc_service=\"example.api\"}[1h]))",
			},
		}, {
			Record: "grpc_server_handled:burnrate2h",
			Expr: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "sum(rate(grpc_server_handled_total{grpc_method=\"Users\",grpc_service=\"example.api\"}[2h]))\n/\nsum(rate(grpc_server_handled_total{grpc_method=\"Users\",grpc_service=\"example.api\"}[2h]))",
			},
		}, {
			Record: "grpc_server_handled:burnrate6h",
			Expr: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "sum(rate(grpc_server_handled_total{grpc_method=\"Users\",grpc_service=\"example.api\"}[6h]))\n/\nsum(rate(grpc_server_handled_total{grpc_method=\"Users\",grpc_service=\"example.api\"}[6h]))",
			},
		}, {
			Record: "grpc_server_handled:burnrate1d",
			Expr: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "sum(rate(grpc_server_handled_total{grpc_method=\"Users\",grpc_service=\"example.api\"}[1d]))\n/\nsum(rate(grpc_server_handled_total{grpc_method=\"Users\",grpc_service=\"example.api\"}[1d]))",
			},
		}, {
			Record: "grpc_server_handled:burnrate4d",
			Expr: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "sum(rate(grpc_server_handled_total{grpc_method=\"Users\",grpc_service=\"example.api\"}[4d]))\n/\nsum(rate(grpc_server_handled_total{grpc_method=\"Users\",grpc_service=\"example.api\"}[4d]))",
			},
		}},
	}

	group, err := makeGRPCRules(slo)

	require.NoError(t, err)
	require.Equal(t, expected, group)
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

	//tw := tabwriter.NewWriter(os.Stderr, 0, 10, 3, ' ', tabwriter.TabIndent)
	//_, _ = fmt.Fprintln(tw, "Severity\tFor\tLong\tShort\tFactor")
	//for _, w := range windows(28 * 24 * time.Hour) {
	//	_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%.f\n", w.Severity, model.Duration(w.For), model.Duration(w.Long), model.Duration(w.Short), w.Factor)
	//}
	//_ = tw.Flush()
}
