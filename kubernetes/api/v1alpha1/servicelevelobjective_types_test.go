package v1alpha1_test

import (
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
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
			Alerting: slo.Alerting{
				Burnrates: true,
				Absent:    true,
			},
			Indicator: slo.Indicator{
				Ratio: &slo.RatioIndicator{
					Errors: slo.Metric{
						Name: "http_requests_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "metrics-service-thanos-receive-default"},
							{Type: labels.MatchRegexp, Name: "code", Value: "5.."},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "http_requests_total"},
						},
						OriginalOffset: ptr.To(0 * time.Second),
					},
					Total: slo.Metric{
						Name: "http_requests_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "metrics-service-thanos-receive-default"},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "http_requests_total"},
						},
						OriginalOffset: ptr.To(0 * time.Second),
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
			Alerting: slo.Alerting{
				Burnrates: true,
				Absent:    true,
			},
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
						OriginalOffset: ptr.To(0 * time.Second),
					},
					Total: slo.Metric{
						Name: "grpc_server_handled_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "api"},
							{Type: labels.MatchEqual, Name: "grpc_service", Value: "conprof.WritableProfileStore"},
							{Type: labels.MatchEqual, Name: "grpc_method", Value: "Write"},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "grpc_server_handled_total"},
						},
						OriginalOffset: ptr.To(0 * time.Second),
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
  name: http-errors-with-offset
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
        metric: http_requests_total{job="metrics-service-thanos-receive-default",code=~"5.."} offset 10m
      total:
        metric: http_requests_total{job="metrics-service-thanos-receive-default"} offset 10m
`,
		objective: slo.Objective{
			Labels: labels.FromStrings(
				labels.MetricName, "http-errors-with-offset",
				"namespace", "monitoring",
				"pyrra.dev/team", "foo",
			),
			Annotations: map[string]string{"pyrra.dev/description": "foo"},
			Description: "",
			Target:      0.99,
			Window:      model.Duration(7 * 24 * time.Hour),
			Alerting: slo.Alerting{
				Burnrates: true,
				Absent:    true,
			},
			Indicator: slo.Indicator{
				Ratio: &slo.RatioIndicator{
					Errors: slo.Metric{
						Name: "http_requests_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "metrics-service-thanos-receive-default"},
							{Type: labels.MatchRegexp, Name: "code", Value: "5.."},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "http_requests_total"},
						},
						OriginalOffset: ptr.To(10 * time.Minute),
					},
					Total: slo.Metric{
						Name: "http_requests_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "metrics-service-thanos-receive-default"},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "http_requests_total"},
						},
						OriginalOffset: ptr.To(10 * time.Minute),
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
			Alerting: slo.Alerting{
				Burnrates: true,
				Absent:    true,
			},
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
						OriginalOffset: ptr.To(0 * time.Second),
					},
					Total: slo.Metric{
						Name: "http_request_duration_seconds_count",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "metrics-service-thanos-receive-default"},
							{Type: labels.MatchRegexp, Name: "code", Value: "2.."},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "http_request_duration_seconds_count"},
						},
						OriginalOffset: ptr.To(0 * time.Second),
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
			Alerting: slo.Alerting{
				Burnrates: true,
				Absent:    true,
			},
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
						OriginalOffset: ptr.To(0 * time.Second),
					},
					Total: slo.Metric{
						Name: "grpc_server_handling_seconds_count",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "api"},
							{Type: labels.MatchEqual, Name: "grpc_service", Value: "conprof.WritableProfileStore"},
							{Type: labels.MatchEqual, Name: "grpc_method", Value: "Write"},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "grpc_server_handling_seconds_count"},
						},
						OriginalOffset: ptr.To(0 * time.Second),
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
			Alerting: slo.Alerting{
				Burnrates: true,
				Absent:    true,
			},
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
						OriginalOffset: ptr.To(0 * time.Second),
					},
					Total: slo.Metric{
						Name: "nginx_ingress_controller_request_duration_seconds_count",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "ingress", Value: "lastfm"},
							{Type: labels.MatchEqual, Name: "path", Value: "/"},
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "nginx_ingress_controller_request_duration_seconds_count"},
						},
						OriginalOffset: ptr.To(0 * time.Second),
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
			Alerting: slo.Alerting{
				Burnrates: true,
				Absent:    true,
			},
			Indicator: slo.Indicator{
				Ratio: &slo.RatioIndicator{
					Errors: slo.Metric{
						Name: "prometheus_operator_reconcile_errors_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "prometheus_operator_reconcile_errors_total"},
						},
						OriginalOffset: ptr.To(0 * time.Second),
					},
					Total: slo.Metric{
						Name: "prometheus_operator_reconcile_operations_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: labels.MetricName, Value: "prometheus_operator_reconcile_operations_total"},
						},
						OriginalOffset: ptr.To(0 * time.Second),
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

func TestServiceLevelObjective_Validate(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		empty := &v1alpha1.ServiceLevelObjective{}
		warn, err := empty.ValidateCreate()
		require.EqualError(t, err, "name must be set")
		require.Nil(t, warn)

		empty.Name = "name"
		warn, err = empty.ValidateCreate()
		require.EqualError(t, err, "target must be set")
		require.Equal(t, "namespace must be set", warn[0])

		empty.Namespace = "namespace"

		empty.Spec.Target = "-99"
		warn, err = empty.ValidateCreate()
		require.EqualError(t, err, "target must be between 0 and 100")
		require.Nil(t, warn)

		empty.Spec.Target = "9999"
		warn, err = empty.ValidateCreate()
		require.EqualError(t, err, "target must be between 0 and 100")
		require.Nil(t, warn)

		empty.Spec.Target = "0.9134"
		warn, err = empty.ValidateCreate()
		require.Equal(t, "target is from 0-100 (91.34), not 0-1 (0.9134)", warn[0])
		require.EqualError(t, err, "window must be set")

		empty.Spec.Target = "99"

		empty.Spec.Window = "2t"
		warn, err = empty.ValidateCreate()
		require.Nil(t, warn)
		require.EqualError(t, err, `unknown unit "t" in duration "2t"`)

		empty.Spec.Window = "2w"
		warn, err = empty.ValidateCreate()
		require.Nil(t, warn)
		require.EqualError(t, err, "one of ratio, latency, latencyNative or bool_gauge must be set")
	})

	t.Run("ratio", func(t *testing.T) {
		ratio := func() *v1alpha1.ServiceLevelObjective {
			return &v1alpha1.ServiceLevelObjective{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: v1alpha1.ServiceLevelObjectiveSpec{
					Target: "99",
					Window: "2w",
					ServiceLevelIndicator: v1alpha1.ServiceLevelIndicator{
						Ratio: &v1alpha1.RatioIndicator{
							Errors:   v1alpha1.Query{Metric: `errors{foo="bar"} offset 10m`},
							Total:    v1alpha1.Query{Metric: `total{foo="bar"} offset 10m`},
							Grouping: nil,
						},
					},
				},
			}
		}

		warn, err := ratio().ValidateCreate()
		require.NoError(t, err)
		require.Nil(t, warn)

		t.Run("empty", func(t *testing.T) {
			ratio := ratio()
			ratio.Spec.ServiceLevelIndicator.Ratio.Errors.Metric = ""
			ratio.Spec.ServiceLevelIndicator.Ratio.Total.Metric = ""
			warn, err := ratio.ValidateCreate()
			require.EqualError(t, err, "ratio total metric must be set")
			require.Nil(t, warn)

			ratio.Spec.ServiceLevelIndicator.Ratio.Errors.Metric = ""
			ratio.Spec.ServiceLevelIndicator.Ratio.Total.Metric = "foo"
			warn, err = ratio.ValidateCreate()
			require.EqualError(t, err, "ratio errors metric must be set")
			require.Nil(t, warn)

			ratio.Spec.ServiceLevelIndicator.Ratio.Errors.Metric = "foo"
			ratio.Spec.ServiceLevelIndicator.Ratio.Total.Metric = ""
			warn, err = ratio.ValidateCreate()
			require.EqualError(t, err, "ratio total metric must be set")
			require.Nil(t, warn)
		})

		t.Run("equal", func(t *testing.T) {
			ratio := ratio()
			ratio.Spec.ServiceLevelIndicator.Ratio.Errors.Metric = "foo"
			ratio.Spec.ServiceLevelIndicator.Ratio.Total.Metric = "foo"
			warn, err := ratio.ValidateCreate()
			require.Equal(t, "ratio errors metric should be different from ratio total metric", warn[0])
			require.NoError(t, err)
		})

		t.Run("invalidMetric", func(t *testing.T) {
			ratio := ratio()
			ratio.Spec.ServiceLevelIndicator.Ratio.Total.Metric = "foo{"
			warn, err := ratio.ValidateCreate()
			require.EqualError(t, err, "failed to parse ratio total metric: 1:5: parse error: unexpected end of input inside braces")
			require.Nil(t, warn)

			ratio.Spec.ServiceLevelIndicator.Ratio.Total.Metric = "foo}"
			warn, err = ratio.ValidateCreate()
			require.EqualError(t, err, "failed to parse ratio total metric: 1:4: parse error: unexpected character: '}'")
			require.Nil(t, warn)

			ratio.Spec.ServiceLevelIndicator.Ratio.Total.Metric = "$$$"
			warn, err = ratio.ValidateCreate()
			require.EqualError(t, err, "failed to parse ratio total metric: 1:1: parse error: unexpected character: '$'")
			require.Nil(t, warn)

			ratio.Spec.ServiceLevelIndicator.Ratio.Total.Metric = `foo{foo="bar'}`
			warn, err = ratio.ValidateCreate()
			require.EqualError(t, err, "failed to parse ratio total metric: 1:9: parse error: unterminated quoted string")
			require.Nil(t, warn)
		})
	})

	t.Run("latency", func(t *testing.T) {
		latency := func() *v1alpha1.ServiceLevelObjective {
			return &v1alpha1.ServiceLevelObjective{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: v1alpha1.ServiceLevelObjectiveSpec{
					Target: "99",
					Window: "2w",
					ServiceLevelIndicator: v1alpha1.ServiceLevelIndicator{
						Latency: &v1alpha1.LatencyIndicator{
							Success:  v1alpha1.Query{Metric: `foo_bucket{foo="bar",le="1"}`},
							Total:    v1alpha1.Query{Metric: `foo_count{foo="bar"}`},
							Grouping: nil,
						},
					},
				},
			}
		}

		warn, err := latency().ValidateCreate()
		require.NoError(t, err)
		require.Nil(t, warn)

		t.Run("empty", func(t *testing.T) {
			latency := latency()
			latency.Spec.ServiceLevelIndicator.Latency.Success.Metric = ""
			latency.Spec.ServiceLevelIndicator.Latency.Total.Metric = ""
			warn, err := latency.ValidateCreate()
			require.EqualError(t, err, "latency total metric must be set")
			require.Nil(t, warn)

			latency.Spec.ServiceLevelIndicator.Latency.Success.Metric = ""
			latency.Spec.ServiceLevelIndicator.Latency.Total.Metric = "foo"
			warn, err = latency.ValidateCreate()
			require.EqualError(t, err, "latency success metric must be set")
			require.Nil(t, warn)

			latency.Spec.ServiceLevelIndicator.Latency.Success.Metric = "foo"
			latency.Spec.ServiceLevelIndicator.Latency.Total.Metric = ""
			warn, err = latency.ValidateCreate()
			require.EqualError(t, err, "latency total metric must be set")
			require.Nil(t, warn)
		})

		t.Run("equal", func(t *testing.T) {
			latency := latency()
			latency.Spec.ServiceLevelIndicator.Latency.Success.Metric = "foo"
			latency.Spec.ServiceLevelIndicator.Latency.Total.Metric = "foo"
			warn, err := latency.ValidateCreate()
			require.NotNil(t, warn)
			require.Equal(t, "latency success metric should be different from latency total metric", warn[0])
			require.Error(t, err)
		})

		t.Run("warnings", func(t *testing.T) {
			latency := latency()
			latency.Spec.ServiceLevelIndicator.Latency.Success.Metric = `foo{le="1"}`
			latency.Spec.ServiceLevelIndicator.Latency.Total.Metric = "bar"
			warn, err = latency.ValidateCreate()
			require.NoError(t, err)
			require.Equal(t,
				admission.Warnings{
					"latency total metric should usually be a histogram count",
					"latency success metric should usually be a histogram bucket",
				},
				warn)
		})

		t.Run("invalidMetric", func(t *testing.T) {
			latency := latency()
			latency.Spec.ServiceLevelIndicator.Latency.Total.Metric = "foo{"
			warn, err := latency.ValidateCreate()
			require.EqualError(t, err, "failed to parse latency total metric: 1:5: parse error: unexpected end of input inside braces")
			require.Nil(t, warn)

			latency.Spec.ServiceLevelIndicator.Latency.Total.Metric = "foo}"
			warn, err = latency.ValidateCreate()
			require.EqualError(t, err, "failed to parse latency total metric: 1:4: parse error: unexpected character: '}'")
			require.Nil(t, warn)

			latency.Spec.ServiceLevelIndicator.Latency.Total.Metric = "$$$"
			warn, err = latency.ValidateCreate()
			require.EqualError(t, err, "failed to parse latency total metric: 1:1: parse error: unexpected character: '$'")
			require.Nil(t, warn)

			latency.Spec.ServiceLevelIndicator.Latency.Total.Metric = `foo{foo="bar'}`
			warn, err = latency.ValidateCreate()
			require.EqualError(t, err, "failed to parse latency total metric: 1:9: parse error: unterminated quoted string")
			require.Nil(t, warn)

			latency.Spec.ServiceLevelIndicator.Latency.Total.Metric = `foo{foo="bar"}`
			latency.Spec.ServiceLevelIndicator.Latency.Success.Metric = `foo{foo="baz"}`
			_, err = latency.ValidateCreate()
			require.EqualError(t, err, "latency success metric must contain a le label matcher")

			latency.Spec.ServiceLevelIndicator.Latency.Success.Metric = `foo{le="foo"}`
			_, err = latency.ValidateCreate()
			require.EqualError(t, err, `latency success metric must contain a le label matcher with a float value: strconv.ParseFloat: parsing "foo": invalid syntax`)

			latency.Spec.ServiceLevelIndicator.Latency.Success.Metric = `foo{le="1.0"} or vector(0)`
			_, err = latency.ValidateCreate()
			require.EqualError(t, err, `latency success metric must be a vector selector, but got *parser.BinaryExpr`)
			latency.Spec.ServiceLevelIndicator.Latency.Total.Metric = `foo{le="1.0"} or vector(0)`
			_, err = latency.ValidateCreate()
			require.EqualError(t, err, `latency total metric must be a vector selector, but got *parser.BinaryExpr`)
		})
	})

	t.Run("latencyNative", func(t *testing.T) {
		latencyNative := func() *v1alpha1.ServiceLevelObjective {
			return &v1alpha1.ServiceLevelObjective{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: v1alpha1.ServiceLevelObjectiveSpec{
					Target: "99",
					Window: "2w",
					ServiceLevelIndicator: v1alpha1.ServiceLevelIndicator{
						LatencyNative: &v1alpha1.NativeLatencyIndicator{
							Latency:  "1s",
							Total:    v1alpha1.Query{Metric: `foo{foo="bar"}`},
							Grouping: nil,
						},
					},
				},
			}
		}

		warn, err := latencyNative().ValidateCreate()
		require.NoError(t, err)
		require.Nil(t, warn)

		t.Run("empty", func(t *testing.T) {
			ln := latencyNative()
			ln.Spec.ServiceLevelIndicator.LatencyNative.Total.Metric = ""
			warn, err := ln.ValidateCreate()
			require.EqualError(t, err, "latencyNative total metric must be set")
			require.Nil(t, warn)

			ln = latencyNative()
			ln.Spec.ServiceLevelIndicator.LatencyNative.Latency = ""
			warn, err = ln.ValidateCreate()
			require.EqualError(t, err, "latencyNative latency objective must be set")
			require.Nil(t, warn)
		})

		t.Run("invalidLatency", func(t *testing.T) {
			ln := latencyNative()
			ln.Spec.ServiceLevelIndicator.LatencyNative.Latency = "foo"
			warn, err := ln.ValidateCreate()
			require.EqualError(t, err, `latencyNative latency objective must be a valid duration: not a valid duration string: "foo"`)
			require.Nil(t, warn)
		})

		t.Run("invalidMetric", func(t *testing.T) {
			ln := latencyNative()
			ln.Spec.ServiceLevelIndicator.LatencyNative.Total.Metric = "foo{"
			warn, err := ln.ValidateCreate()
			require.EqualError(t, err, "failed to parse latencyNative total metric: 1:5: parse error: unexpected end of input inside braces")
			require.Nil(t, warn)

			ln.Spec.ServiceLevelIndicator.LatencyNative.Total.Metric = "foo}"
			warn, err = ln.ValidateCreate()
			require.EqualError(t, err, "failed to parse latencyNative total metric: 1:4: parse error: unexpected character: '}'")
			require.Nil(t, warn)

			ln.Spec.ServiceLevelIndicator.LatencyNative.Total.Metric = "$$$"
			warn, err = ln.ValidateCreate()
			require.EqualError(t, err, "failed to parse latencyNative total metric: 1:1: parse error: unexpected character: '$'")
			require.Nil(t, warn)

			ln.Spec.ServiceLevelIndicator.LatencyNative.Total.Metric = `foo{foo="bar'}`
			warn, err = ln.ValidateCreate()
			require.EqualError(t, err, "failed to parse latencyNative total metric: 1:9: parse error: unterminated quoted string")
			require.Nil(t, warn)
		})
	})

	t.Run("boolGauge", func(t *testing.T) {
		boolGauge := func() *v1alpha1.ServiceLevelObjective {
			return &v1alpha1.ServiceLevelObjective{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: v1alpha1.ServiceLevelObjectiveSpec{
					Target: "99",
					Window: "2w",
					ServiceLevelIndicator: v1alpha1.ServiceLevelIndicator{
						BoolGauge: &v1alpha1.BoolGaugeIndicator{
							Query: v1alpha1.Query{
								Metric: `foo{foo="bar"}`,
							},
						},
					},
				},
			}
		}

		warn, err := boolGauge().ValidateCreate()
		require.NoError(t, err)
		require.Nil(t, warn)

		t.Run("empty", func(t *testing.T) {
			bg := boolGauge()
			bg.Spec.ServiceLevelIndicator.BoolGauge.Query.Metric = ""
			warn, err := bg.ValidateCreate()
			require.EqualError(t, err, "boolGauge metric must be set")
			require.Nil(t, warn)
		})

		t.Run("invalidMetric", func(t *testing.T) {
			bg := boolGauge()
			bg.Spec.ServiceLevelIndicator.BoolGauge.Query.Metric = "foo{"
			warn, err := bg.ValidateCreate()
			require.EqualError(t, err, "failed to parse boolGauge metric: 1:5: parse error: unexpected end of input inside braces")
			require.Nil(t, warn)

			bg.Spec.ServiceLevelIndicator.BoolGauge.Query.Metric = "foo}"
			warn, err = bg.ValidateCreate()
			require.EqualError(t, err, "failed to parse boolGauge metric: 1:4: parse error: unexpected character: '}'")
			require.Nil(t, warn)

			bg.Spec.ServiceLevelIndicator.BoolGauge.Query.Metric = "$$$"
			warn, err = bg.ValidateCreate()
			require.EqualError(t, err, "failed to parse boolGauge metric: 1:1: parse error: unexpected character: '$'")
			require.Nil(t, warn)

			bg.Spec.ServiceLevelIndicator.BoolGauge.Query.Metric = `foo{foo="bar'}`
			warn, err = bg.ValidateCreate()
			require.EqualError(t, err, "failed to parse boolGauge metric: 1:9: parse error: unterminated quoted string")
			require.Nil(t, warn)
		})
	})
}
