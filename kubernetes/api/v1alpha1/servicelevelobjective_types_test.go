package v1alpha1_test

import (
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/pyrra-dev/pyrra/kubernetes/api/v1alpha1"
	"github.com/pyrra-dev/pyrra/slo"
)

var examples = []struct {
	config    string
	objective slo.Objective
}{
	{
		config: `
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: http-errors
  namespace: monitoring
  labels:
    prometheus: k8s
    role: alert-rules
    pyrra.dev/team: foo
  annotations:
    pyrra.dev/description: "foo"
spec:
  target: 99
  window: 1w
  indicator:
    ratio:
      errors:
        metric: http_requests_total{job="metrics-service-thanos-receive-default",code=~"5.."}
      total:
        metric: http_requests_total{job="metrics-service-thanos-receive-default"}
`,
		objective: slo.Objective{
			Labels: labels.FromStrings(
				labels.MetricName, "http-errors",
				"namespace", "monitoring",
				"pyrra.dev/team", "foo",
			),
			Annotations: map[string]string{"pyrra.dev/description": "foo"},
			Description: "",
			Target:      0.99,
			Window:      model.Duration(7 * 24 * time.Hour),
			Indicator: slo.Indicator{
				Ratio: &slo.RatioIndicator{
					Errors: slo.Metric{
						Name: "http_requests_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "metrics-service-thanos-receive-default"},
							{Type: labels.MatchRegexp, Name: "code", Value: "5.."},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "http_requests_total"},
						},
					},
					Total: slo.Metric{
						Name: "http_requests_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "metrics-service-thanos-receive-default"},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "http_requests_total"},
						},
					},
				},
			},
		},
	},
	{
		config: `
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
 name: grpc-errors
 namespace: monitoring
 labels:
   prometheus: k8s
   role: alert-rules
spec:
 target: 99.9
 window: 1w
 indicator:
   ratio:
     errors:
       metric: grpc_server_handled_total{job="api",grpc_service="conprof.WritableProfileStore",grpc_method="Write",grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss"}
     total:
       metric: grpc_server_handled_total{job="api",grpc_service="conprof.WritableProfileStore",grpc_method="Write"}
`,
		objective: slo.Objective{
			Labels: labels.FromStrings(
				labels.MetricName, "grpc-errors",
				"namespace", "monitoring",
			),
			Description: "",
			Target:      0.9990000000000001, // TODO fix this? maybe not /100?
			Window:      model.Duration(7 * 24 * time.Hour),
			Indicator: slo.Indicator{
				Ratio: &slo.RatioIndicator{
					Errors: slo.Metric{
						Name: "grpc_server_handled_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "api"},
							{Type: labels.MatchEqual, Name: "grpc_service", Value: "conprof.WritableProfileStore"},
							{Type: labels.MatchEqual, Name: "grpc_method", Value: "Write"},
							{Type: labels.MatchRegexp, Name: "grpc_code", Value: "Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss"},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "grpc_server_handled_total"},
						},
					},
					Total: slo.Metric{
						Name: "grpc_server_handled_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "api"},
							{Type: labels.MatchEqual, Name: "grpc_service", Value: "conprof.WritableProfileStore"},
							{Type: labels.MatchEqual, Name: "grpc_method", Value: "Write"},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "grpc_server_handled_total"},
						},
					},
				},
			},
		},
	},
	{
		config: `
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
 name: http-latency
 namespace: monitoring
 labels:
   prometheus: k8s
   role: alert-rules
spec:
 target: 99.5
 window: 4w
 indicator:
   latency:
     success:
       metric: http_request_duration_seconds_bucket{job="metrics-service-thanos-receive-default",code=~"2..",le="1"}
     total:
       metric: http_request_duration_seconds_count{job="metrics-service-thanos-receive-default",code=~"2.."}
`,
		objective: slo.Objective{
			Labels: labels.FromStrings(
				labels.MetricName, "http-latency",
				"namespace", "monitoring",
			),
			Target: 0.995,
			Window: model.Duration(28 * 24 * time.Hour),
			Indicator: slo.Indicator{
				Latency: &slo.LatencyIndicator{
					Success: slo.Metric{
						Name: "http_request_duration_seconds_bucket",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "metrics-service-thanos-receive-default"},
							{Type: labels.MatchRegexp, Name: "code", Value: "2.."},
							{Type: labels.MatchEqual, Name: "le", Value: "1"},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "http_request_duration_seconds_bucket"},
						},
					},
					Total: slo.Metric{
						Name: "http_request_duration_seconds_count",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "metrics-service-thanos-receive-default"},
							{Type: labels.MatchRegexp, Name: "code", Value: "2.."},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "http_request_duration_seconds_count"},
						},
					},
				},
			},
		},
	},
	{
		config: `
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
 name: grpc-latency
 namespace: monitoring
 labels:
   prometheus: k8s
   role: alert-rules
spec:
 target: 99.5
 window: 1w
 indicator:
   latency:
     success:
       metric: grpc_server_handling_seconds_bucket{job="api",grpc_service="conprof.WritableProfileStore",grpc_method="Write",le="0.6"}
     total:
       metric: grpc_server_handling_seconds_count{job="api",grpc_service="conprof.WritableProfileStore",grpc_method="Write"}
`,
		objective: slo.Objective{
			Labels: labels.FromStrings(
				labels.MetricName, "grpc-latency",
				"namespace", "monitoring",
			),
			Target: 0.995,
			Window: model.Duration(7 * 24 * time.Hour),
			Indicator: slo.Indicator{
				Latency: &slo.LatencyIndicator{
					Success: slo.Metric{
						Name: "grpc_server_handling_seconds_bucket",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "api"},
							{Type: labels.MatchEqual, Name: "grpc_service", Value: "conprof.WritableProfileStore"},
							{Type: labels.MatchEqual, Name: "grpc_method", Value: "Write"},
							{Type: labels.MatchEqual, Name: "le", Value: "0.6"},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "grpc_server_handling_seconds_bucket"},
						},
					},
					Total: slo.Metric{
						Name: "grpc_server_handling_seconds_count",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "api"},
							{Type: labels.MatchEqual, Name: "grpc_service", Value: "conprof.WritableProfileStore"},
							{Type: labels.MatchEqual, Name: "grpc_method", Value: "Write"},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "grpc_server_handling_seconds_count"},
						},
					},
				},
			},
		},
	},
	{
		config: `
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
 name: http-errorslatency
 namespace: monitoring
 labels:
   prometheus: k8s
   role: alert-rules
spec:
 target: 99
 window: 4w
 indicator:
   latency:
     success:
       metric: nginx_ingress_controller_request_duration_seconds_bucket{ingress="lastfm",path="/",status!~"5..",le="0.5"}
     total:
       metric: nginx_ingress_controller_request_duration_seconds_count{ingress="lastfm",path="/"}
`,
		objective: slo.Objective{
			Labels: labels.FromStrings(
				labels.MetricName, "http-errorslatency",
				"namespace", "monitoring",
			),
			Target: 0.99,
			Window: model.Duration(28 * 24 * time.Hour),
			Indicator: slo.Indicator{
				Latency: &slo.LatencyIndicator{
					Success: slo.Metric{
						Name: "nginx_ingress_controller_request_duration_seconds_bucket",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "ingress", Value: "lastfm"},
							{Type: labels.MatchEqual, Name: "path", Value: "/"},
							{Type: labels.MatchNotRegexp, Name: "status", Value: "5.."},
							{Type: labels.MatchEqual, Name: "le", Value: "0.5"},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "nginx_ingress_controller_request_duration_seconds_bucket"},
						},
					},
					Total: slo.Metric{
						Name: "nginx_ingress_controller_request_duration_seconds_count",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "ingress", Value: "lastfm"},
							{Type: labels.MatchEqual, Name: "path", Value: "/"},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "nginx_ingress_controller_request_duration_seconds_count"},
						},
					},
				},
			},
		},
	},
	{
		config: `
apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
 name: prometheus-operator-errors
 namespace: monitoring
 labels:
   prometheus: k8s
   role: alert-rules
spec:
 target: 99
 window: 2w
 indicator:
   ratio:
     errors:
       metric: prometheus_operator_reconcile_errors_total
     total:
       metric: prometheus_operator_reconcile_operations_total
`,
		objective: slo.Objective{
			Labels: labels.FromStrings(
				labels.MetricName, "prometheus-operator-errors",
				"namespace", "monitoring",
			),
			Target: 0.99,
			Window: model.Duration(14 * 24 * time.Hour),
			Indicator: slo.Indicator{
				Ratio: &slo.RatioIndicator{
					Errors: slo.Metric{
						Name: "prometheus_operator_reconcile_errors_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "prometheus_operator_reconcile_errors_total"},
						},
					},
					Total: slo.Metric{
						Name: "prometheus_operator_reconcile_operations_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "prometheus_operator_reconcile_operations_total"},
						},
					},
				},
			},
		},
	},
}

func TestServiceLevelObjective_Internal(t *testing.T) {
	for _, example := range examples {
		t.Run(example.objective.Labels.Get(labels.MetricName), func(t *testing.T) {
			objective := v1alpha1.ServiceLevelObjective{}
			err := yaml.UnmarshalStrict([]byte(example.config), &objective)
			require.NoError(t, err)

			internal, err := objective.Internal()
			internal.Config = "" // Ignore the embedded config for now
			require.NoError(t, err)
			require.Equal(t, example.objective, internal)
		})
	}
}
