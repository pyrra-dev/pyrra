package slo

import (
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/stretchr/testify/require"
)

var (
	objectiveHTTPRatio = func() Objective {
		return Objective{
			Labels: labels.FromStrings(labels.MetricName, "monitoring-http-errors"),
			Target: 0.99,
			Window: model.Duration(28 * 24 * time.Hour),
			Indicator: Indicator{
				Ratio: &RatioIndicator{
					Errors: Metric{
						Name: "http_requests_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "thanos-receive-default"},
							{Type: labels.MatchRegexp, Name: "code", Value: "5.."},
							{Type: labels.MatchEqual, Name: "__name__", Value: "http_requests_total"},
						},
					},
					Total: Metric{
						Name: "http_requests_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "thanos-receive-default"},
							{Type: labels.MatchEqual, Name: "__name__", Value: "http_requests_total"},
						},
					},
				},
			},
		}
	}
	objectiveHTTPRatioGrouping = func() Objective {
		o := objectiveHTTPRatio()
		o.Indicator.Ratio.Grouping = []string{"job", "handler"}
		return o
	}
	objectiveGRPCRatio = func() Objective {
		return Objective{
			Labels:      labels.FromStrings(labels.MetricName, "monitoring-grpc-errors"),
			Description: "",
			Target:      0.999,
			Window:      model.Duration(28 * 24 * time.Hour),
			Indicator: Indicator{
				Ratio: &RatioIndicator{
					Errors: Metric{
						Name: "grpc_server_handled_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "api"},
							{Type: labels.MatchEqual, Name: "grpc_service", Value: "conprof.WritableProfileStore"},
							{Type: labels.MatchEqual, Name: "grpc_method", Value: "Write"},
							{Type: labels.MatchRegexp, Name: "grpc_code", Value: "Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss"},
							{Type: labels.MatchEqual, Name: "__name__", Value: "grpc_server_handled_total"},
						},
					},
					Total: Metric{
						Name: "grpc_server_handled_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "api"},
							{Type: labels.MatchEqual, Name: "grpc_service", Value: "conprof.WritableProfileStore"},
							{Type: labels.MatchEqual, Name: "grpc_method", Value: "Write"},
							{Type: labels.MatchEqual, Name: "__name__", Value: "grpc_server_handled_total"},
						},
					},
				},
			},
		}
	}
	objectiveGRPCRatioGrouping = func() Objective {
		o := objectiveGRPCRatio()
		o.Indicator.Ratio.Grouping = []string{"job", "handler"}
		return o
	}
	objectiveHTTPLatency = func() Objective {
		return Objective{
			Labels: labels.FromStrings(labels.MetricName, "monitoring-http-latency"),
			Target: 0.995,
			Window: model.Duration(28 * 24 * time.Hour),
			Indicator: Indicator{
				Latency: &LatencyIndicator{
					Success: Metric{
						Name: "http_request_duration_seconds_bucket",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "metrics-service-thanos-receive-default"},
							{Type: labels.MatchRegexp, Name: "code", Value: "2.."},
							{Type: labels.MatchEqual, Name: "le", Value: "1"},
							{Type: labels.MatchEqual, Name: "__name__", Value: "http_request_duration_seconds_bucket"},
						},
					},
					Total: Metric{
						Name: "http_request_duration_seconds_count",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "metrics-service-thanos-receive-default"},
							{Type: labels.MatchRegexp, Name: "code", Value: "2.."},
							{Type: labels.MatchEqual, Name: "__name__", Value: "http_request_duration_seconds_count"},
						},
					},
				},
			},
		}
	}
	objectiveHTTPLatencyGrouping = func() Objective {
		o := objectiveHTTPLatency()
		o.Indicator.Latency.Grouping = []string{"job", "handler"}
		return o
	}
	objectiveGRPCLatency = func() Objective {
		return Objective{
			Labels: labels.FromStrings(labels.MetricName, "monitoring-grpc-latency"),
			Target: 0.995,
			Window: model.Duration(7 * 24 * time.Hour),
			Indicator: Indicator{
				Latency: &LatencyIndicator{
					Success: Metric{
						Name: "grpc_server_handling_seconds_bucket",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "api"},
							{Type: labels.MatchEqual, Name: "grpc_service", Value: "conprof.WritableProfileStore"},
							{Type: labels.MatchEqual, Name: "grpc_method", Value: "Write"},
							{Type: labels.MatchEqual, Name: "le", Value: "0.6"},
							{Type: labels.MatchEqual, Name: "__name__", Value: "grpc_server_handling_seconds_bucket"},
						},
					},
					Total: Metric{
						Name: "grpc_server_handling_seconds_count",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "api"},
							{Type: labels.MatchEqual, Name: "grpc_service", Value: "conprof.WritableProfileStore"},
							{Type: labels.MatchEqual, Name: "grpc_method", Value: "Write"},
							{Type: labels.MatchEqual, Name: "__name__", Value: "grpc_server_handling_seconds_count"},
						},
					},
				},
			},
		}
	}
	objectiveGRPCLatencyGrouping = func() Objective {
		o := objectiveGRPCLatency()
		o.Indicator.Latency.Grouping = []string{"job", "handler"}
		return o
	}
	objectiveOperator = func() Objective {
		return Objective{
			Labels: labels.FromStrings(labels.MetricName, "monitoring-prometheus-operator-errors"),
			Target: 0.99,
			Window: model.Duration(14 * 24 * time.Hour),
			Indicator: Indicator{
				Ratio: &RatioIndicator{
					Errors: Metric{
						Name: "prometheus_operator_reconcile_errors_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "__name__", Value: "prometheus_operator_reconcile_errors_total"},
						},
					},
					Total: Metric{
						Name: "prometheus_operator_reconcile_operations_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "__name__", Value: "prometheus_operator_reconcile_operations_total"},
						},
					},
				},
			},
		}
	}
	objectiveOperatorGrouping = func() Objective {
		o := objectiveOperator()
		o.Indicator.Ratio.Grouping = []string{"namespace"}
		return o
	}
)

func TestObjective_QueryTotal(t *testing.T) {
	testcases := []struct {
		name      string
		objective Objective
		expected  string
	}{{
		name:      "http-ratio",
		objective: objectiveHTTPRatio(),
		expected:  `sum(increase(http_requests_total{job="thanos-receive-default"}[4w]))`,
	}, {
		name:      "http-ratio-grouping",
		objective: objectiveHTTPRatioGrouping(),
		expected:  `sum by(job, handler) (increase(http_requests_total{job="thanos-receive-default"}[4w]))`,
	}, {
		name:      "grpc-ratio",
		objective: objectiveGRPCRatio(),
		expected:  `sum(increase(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[4w]))`,
	}, {
		name:      "grpc-ratio-grouping",
		objective: objectiveGRPCRatioGrouping(),
		expected:  `sum by(job, handler) (increase(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[4w]))`,
	}, {
		name:      "http-latency",
		objective: objectiveHTTPLatency(),
		expected:  `sum(increase(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[4w]))`,
	}, {
		name:      "http-latency-grouping",
		objective: objectiveHTTPLatencyGrouping(),
		expected:  `sum by(job, handler) (increase(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[4w]))`,
	}, {
		name:      "grpc-latency",
		objective: objectiveGRPCLatency(),
		expected:  `sum(increase(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1w]))`,
	}, {
		name:      "grpc-latency-grouping",
		objective: objectiveGRPCLatencyGrouping(),
		expected:  `sum by(job, handler) (increase(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1w]))`,
	}, {
		name:      "operator-ratio",
		objective: objectiveOperator(),
		expected:  `sum(increase(prometheus_operator_reconcile_operations_total[2w]))`,
	}, {
		name:      "operator-ratio-grouping",
		objective: objectiveOperatorGrouping(),
		expected:  `sum by(namespace) (increase(prometheus_operator_reconcile_operations_total[2w]))`,
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.objective.QueryTotal(tc.objective.Window))
		})
	}
}

func TestObjective_QueryErrors(t *testing.T) {
	testcases := []struct {
		name      string
		objective Objective
		expected  string
	}{{
		name:      "http-ratio",
		objective: objectiveHTTPRatio(),
		expected:  `sum(increase(http_requests_total{code=~"5..",job="thanos-receive-default"}[4w]))`,
	}, {
		name:      "http-ratio-grouping",
		objective: objectiveHTTPRatioGrouping(),
		expected:  `sum by(job, handler) (increase(http_requests_total{code=~"5..",job="thanos-receive-default"}[4w]))`,
	}, {
		name:      "grpc-ratio",
		objective: objectiveGRPCRatio(),
		expected:  `sum(increase(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[4w]))`,
	}, {
		name:      "grpc-ratio-grouping",
		objective: objectiveGRPCRatioGrouping(),
		expected:  `sum by(job, handler) (increase(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[4w]))`,
	}, {
		name:      "http-latency",
		objective: objectiveHTTPLatency(),
		expected:  `sum(increase(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[4w])) - sum(increase(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[4w]))`,
	}, {
		name:      "http-latency-grouping",
		objective: objectiveHTTPLatencyGrouping(),
		expected:  `sum by(job, handler) (increase(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[4w])) - sum by(job, handler) (increase(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[4w]))`,
	}, {
		name:      "grpc-latency",
		objective: objectiveGRPCLatency(),
		expected:  `sum(increase(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1w])) - sum(increase(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[1w]))`,
	}, {
		name:      "grpc-latency-grouping",
		objective: objectiveGRPCLatencyGrouping(),
		expected:  `sum by(job, handler) (increase(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1w])) - sum by(job, handler) (increase(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[1w]))`,
	}, {
		name:      "operator-ratio",
		objective: objectiveOperator(),
		expected:  `sum(increase(prometheus_operator_reconcile_errors_total[2w]))`,
	}, {
		name:      "operator-ratio-grouping",
		objective: objectiveOperatorGrouping(),
		expected:  `sum by(namespace) (increase(prometheus_operator_reconcile_errors_total[2w]))`,
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.objective.QueryErrors(tc.objective.Window))
		})
	}
}

func TestObjective_QueryErrorBudget(t *testing.T) {
	testcases := []struct {
		name      string
		objective Objective
		expected  string
	}{{
		name:      "http-ratio",
		objective: objectiveHTTPRatio(),
		expected:  `((1 - 0.99) - (sum(increase(http_requests_total{code=~"5..",job="thanos-receive-default"}[4w]) or vector(0)) / sum(increase(http_requests_total{job="thanos-receive-default"}[4w])))) / (1 - 0.99)`,
	}, {
		name:      "http-ratio-grouping",
		objective: objectiveHTTPRatioGrouping(),
		expected:  `((1 - 0.99) - (sum by(job, handler) (increase(http_requests_total{code=~"5..",job="thanos-receive-default"}[4w]) or vector(0)) / sum by(job, handler) (increase(http_requests_total{job="thanos-receive-default"}[4w])))) / (1 - 0.99)`,
	}, {
		name:      "grpc-ratio",
		objective: objectiveGRPCRatio(),
		expected:  `((1 - 0.999) - (sum(increase(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[4w]) or vector(0)) / sum(increase(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[4w])))) / (1 - 0.999)`,
	}, {
		name:      "grpc-ratio-grouping",
		objective: objectiveGRPCRatioGrouping(),
		expected:  `((1 - 0.999) - (sum by(job, handler) (increase(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[4w]) or vector(0)) / sum by(job, handler) (increase(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[4w])))) / (1 - 0.999)`,
	}, {
		name:      "http-latency",
		objective: objectiveHTTPLatency(),
		expected:  `((1 - 0.995) - (1 - sum(increase(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[4w]) or vector(0)) / sum(increase(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[4w])))) / (1 - 0.995)`,
	}, {
		name:      "http-latency-grouping",
		objective: objectiveHTTPLatencyGrouping(),
		expected:  `((1 - 0.995) - (1 - sum(increase(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[4w]) or vector(0)) / sum(increase(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[4w])))) / (1 - 0.995)`,
	}, {
		name:      "grpc-latency",
		objective: objectiveGRPCLatency(),
		expected:  `((1 - 0.995) - (1 - sum(increase(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[1w]) or vector(0)) / sum(increase(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1w])))) / (1 - 0.995)`,
	}, {
		name:      "operator-ratio",
		objective: objectiveOperator(),
		expected:  `((1 - 0.99) - (sum(increase(prometheus_operator_reconcile_errors_total[2w]) or vector(0)) / sum(increase(prometheus_operator_reconcile_operations_total[2w])))) / (1 - 0.99)`,
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.objective.QueryErrorBudget())
		})
	}
}

func TestObjective_RequestRange(t *testing.T) {
	testcases := []struct {
		name      string
		objective Objective
		timerange time.Duration
		expected  string
	}{{
		name:      "http-ratio",
		objective: objectiveHTTPRatio(),
		timerange: 6 * time.Hour,
		expected:  `sum by(code) (rate(http_requests_total{job="thanos-receive-default"}[6h])) > 0`,
	}, {
		name:      "grpc-ratio",
		objective: objectiveGRPCRatio(),
		timerange: 6 * time.Hour,
		expected:  `sum by(grpc_code) (rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[6h])) > 0`,
	}, {
		name:      "operator-ratio",
		objective: objectiveOperator(),
		timerange: 5 * time.Minute,
		expected:  `sum(rate(prometheus_operator_reconcile_operations_total[5m])) > 0`,
	}, {
		name:      "http-latency",
		objective: objectiveHTTPLatency(),
		timerange: 2 * time.Hour,
		expected:  `sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[2h]))`,
	}, {
		name:      "grpc-latency",
		objective: objectiveGRPCLatency(),
		timerange: 3 * time.Hour,
		expected:  `sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[3h]))`,
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.objective.RequestRange(tc.timerange))
		})
	}
}

func TestObjective_ErrorsRange(t *testing.T) {
	testcases := []struct {
		name      string
		objective Objective
		timerange time.Duration
		expected  string
	}{{
		name:      "http-ratio",
		objective: objectiveHTTPRatio(),
		timerange: 6 * time.Hour,
		expected:  `sum by(code) (rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[6h])) / scalar(sum(rate(http_requests_total{job="thanos-receive-default"}[6h])))`,
	}, {
		name:      "grpc-ratio",
		objective: objectiveGRPCRatio(),
		timerange: 6 * time.Hour,
		expected:  `sum by(grpc_code) (rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[6h])) / scalar(sum(rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[6h])))`,
	}, {
		name:      "operator-ratio",
		objective: objectiveOperator(),
		timerange: 5 * time.Minute,
		expected:  `sum(rate(prometheus_operator_reconcile_errors_total[5m])) / scalar(sum(rate(prometheus_operator_reconcile_operations_total[5m])))`,
	}, {
		name:      "http-latency",
		objective: objectiveHTTPLatency(),
		timerange: time.Hour,
		expected:  `sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1h])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[1h]))`,
	}, {
		name:      "grpc-latency",
		objective: objectiveGRPCLatency(),
		timerange: time.Hour,
		expected:  `sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1h])) - sum(rate(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[1h]))`,
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.objective.ErrorsRange(tc.timerange))
		})
	}
}
