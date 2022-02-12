package slo

import (
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
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
	objectiveHTTPRatioGroupingRegex = func() Objective {
		matcher := &labels.Matcher{
			Type:  labels.MatchRegexp,
			Name:  "handler",
			Value: "/api.*",
		}
		o := objectiveHTTPRatioGrouping()
		o.Indicator.Ratio.Total.LabelMatchers = append(o.Indicator.Ratio.Total.LabelMatchers, matcher)
		o.Indicator.Ratio.Errors.LabelMatchers = append(o.Indicator.Ratio.Errors.LabelMatchers, matcher)
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
	objectiveHTTPLatencyGroupingRegex = func() Objective {
		matcher := &labels.Matcher{
			Type:  labels.MatchRegexp,
			Name:  "handler",
			Value: "/api.*",
		}
		o := objectiveHTTPLatencyGrouping()
		o.Indicator.Latency.Grouping = []string{"job", "handler"}
		o.Indicator.Latency.Success.LabelMatchers = append(o.Indicator.Latency.Success.LabelMatchers, matcher)
		o.Indicator.Latency.Total.LabelMatchers = append(o.Indicator.Latency.Total.LabelMatchers, matcher)
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
	objectiveAPIServerRatio = func() Objective {
		return Objective{
			Labels: labels.FromStrings(labels.MetricName, "apiserver-write-response-errors"),
			Target: 99,
			Window: model.Duration(14 * 24 * time.Hour),
			Indicator: Indicator{
				Ratio: &RatioIndicator{
					Errors: Metric{
						Name: "apiserver_request_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "apiserver"},
							{Type: labels.MatchRegexp, Name: "verb", Value: "POST|PUT|PATCH|DELETE"},
							{Type: labels.MatchRegexp, Name: "code", Value: "5.."},
						},
					},
					Total: Metric{
						Name: "apiserver_request_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "apiserver"},
							{Type: labels.MatchRegexp, Name: "verb", Value: "POST|PUT|PATCH|DELETE"},
						},
					},
				},
			},
		}
	}
	objectiveAPIServerLatency = func() Objective {
		return Objective{
			Labels: labels.FromStrings(labels.MetricName, "apiserver-read-resource-latency"),
			Target: 99,
			Window: model.Duration(14 * 24 * time.Hour),
			Indicator: Indicator{
				Latency: &LatencyIndicator{
					Success: Metric{
						Name: "apiserver_request_duration_seconds_bucket",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "apiserver"},
							{Type: labels.MatchRegexp, Name: "verb", Value: "LIST|GET"},
							{Type: labels.MatchRegexp, Name: "resource", Value: "resource|"},
							{Type: labels.MatchEqual, Name: "le", Value: "0.1"},
						},
					},
					Total: Metric{
						Name: "apiserver_request_duration_seconds_count",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "apiserver"},
							{Type: labels.MatchRegexp, Name: "verb", Value: "LIST|GET"},
							{Type: labels.MatchRegexp, Name: "resource", Value: "resource|"},
						},
					},
				},
			},
		}
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
		expected:  `sum(http_requests:increase4w{job="thanos-receive-default",slo="monitoring-http-errors"})`,
	}, {
		name:      "http-ratio-grouping",
		objective: objectiveHTTPRatioGrouping(),
		expected:  `sum by(job, handler) (http_requests:increase4w{job="thanos-receive-default",slo="monitoring-http-errors"})`,
	}, {
		name:      "http-ratio-grouping-regex",
		objective: objectiveHTTPRatioGroupingRegex(),
		expected:  `sum by(job, handler) (http_requests:increase4w{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"})`,
	}, {
		name:      "grpc-ratio",
		objective: objectiveGRPCRatio(),
		expected:  `sum(grpc_server_handled:increase4w{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"})`,
	}, {
		name:      "grpc-ratio-grouping",
		objective: objectiveGRPCRatioGrouping(),
		expected:  `sum by(job, handler) (grpc_server_handled:increase4w{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"})`,
	}, {
		name:      "http-latency",
		objective: objectiveHTTPLatency(),
		expected:  `sum(http_request_duration_seconds:increase4w{code=~"2..",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"})`,
	}, {
		name:      "http-latency-grouping",
		objective: objectiveHTTPLatencyGrouping(),
		expected:  `sum by(job, handler) (http_request_duration_seconds:increase4w{code=~"2..",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"})`,
	}, {
		name:      "http-latency-grouping-regex",
		objective: objectiveHTTPLatencyGroupingRegex(),
		expected:  `sum by(job, handler) (http_request_duration_seconds:increase4w{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",slo="monitoring-http-latency"})`,
	}, {
		name:      "grpc-latency",
		objective: objectiveGRPCLatency(),
		expected:  `sum(grpc_server_handling_seconds:increase1w{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"})`,
	}, {
		name:      "grpc-latency-grouping",
		objective: objectiveGRPCLatencyGrouping(),
		expected:  `sum by(job, handler) (grpc_server_handling_seconds:increase1w{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-latency"})`,
	}, {
		name:      "operator-ratio",
		objective: objectiveOperator(),
		expected:  `sum(prometheus_operator_reconcile_operations:increase2w{slo="monitoring-prometheus-operator-errors"})`,
	}, {
		name:      "operator-ratio-grouping",
		objective: objectiveOperatorGrouping(),
		expected:  `sum by(namespace) (prometheus_operator_reconcile_operations:increase2w{slo="monitoring-prometheus-operator-errors"})`,
	}, {
		name:      "apiserver-write-response-errors",
		objective: objectiveAPIServerRatio(),
		expected:  `sum(apiserver_request:increase2w{job="apiserver",slo="apiserver-write-response-errors",verb=~"POST|PUT|PATCH|DELETE"})`,
	}, {
		name:      "apiserver-read-resource-latency",
		objective: objectiveAPIServerLatency(),
		expected:  `sum(apiserver_request_duration_seconds:increase2w{job="apiserver",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"})`,
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
		expected:  `sum(http_requests:increase4w{code=~"5..",job="thanos-receive-default",slo="monitoring-http-errors"})`,
	}, {
		name:      "http-ratio-grouping",
		objective: objectiveHTTPRatioGrouping(),
		expected:  `sum by(job, handler) (http_requests:increase4w{code=~"5..",job="thanos-receive-default",slo="monitoring-http-errors"})`,
	}, {
		name:      "http-ratio-grouping-regex",
		objective: objectiveHTTPRatioGroupingRegex(),
		expected:  `sum by(job, handler) (http_requests:increase4w{code=~"5..",handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"})`,
	}, {
		name:      "grpc-ratio",
		objective: objectiveGRPCRatio(),
		expected:  `sum(grpc_server_handled:increase4w{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"})`,
	}, {
		name:      "grpc-ratio-grouping",
		objective: objectiveGRPCRatioGrouping(),
		expected:  `sum by(job, handler) (grpc_server_handled:increase4w{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"})`,
	}, {
		name:      "http-latency",
		objective: objectiveHTTPLatency(),
		expected:  `sum(http_request_duration_seconds:increase4w{code=~"2..",job="metrics-service-thanos-receive-default",le="",slo="monitoring-http-latency"}) - sum(http_request_duration_seconds:increase4w{code=~"2..",job="metrics-service-thanos-receive-default",le="1",slo="monitoring-http-latency"})`,
	}, {
		name:      "http-latency-grouping",
		objective: objectiveHTTPLatencyGrouping(),
		expected:  `sum by(job, handler) (http_request_duration_seconds:increase4w{code=~"2..",job="metrics-service-thanos-receive-default",le="",slo="monitoring-http-latency"}) - sum by(job, handler) (http_request_duration_seconds:increase4w{code=~"2..",job="metrics-service-thanos-receive-default",le="1",slo="monitoring-http-latency"})`,
	}, {
		name:      "http-latency-grouping-regex",
		objective: objectiveHTTPLatencyGroupingRegex(),
		expected:  `sum by(job, handler) (http_request_duration_seconds:increase4w{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="",slo="monitoring-http-latency"}) - sum by(job, handler) (http_request_duration_seconds:increase4w{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1",slo="monitoring-http-latency"})`,
	}, {
		name:      "grpc-latency",
		objective: objectiveGRPCLatency(),
		expected:  `sum(grpc_server_handling_seconds:increase1w{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="",slo="monitoring-grpc-latency"}) - sum(grpc_server_handling_seconds:increase1w{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6",slo="monitoring-grpc-latency"})`,
	}, {
		name:      "grpc-latency-grouping",
		objective: objectiveGRPCLatencyGrouping(),
		expected:  `sum by(job, handler) (grpc_server_handling_seconds:increase1w{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="",slo="monitoring-grpc-latency"}) - sum by(job, handler) (grpc_server_handling_seconds:increase1w{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6",slo="monitoring-grpc-latency"})`,
	}, {
		name:      "operator-ratio",
		objective: objectiveOperator(),
		expected:  `sum(prometheus_operator_reconcile_errors:increase2w{slo="monitoring-prometheus-operator-errors"})`,
	}, {
		name:      "operator-ratio-grouping",
		objective: objectiveOperatorGrouping(),
		expected:  `sum by(namespace) (prometheus_operator_reconcile_errors:increase2w{slo="monitoring-prometheus-operator-errors"})`,
	}, {
		name:      "apiserver-write-response-errors",
		objective: objectiveAPIServerRatio(),
		expected:  `sum(apiserver_request:increase2w{code=~"5..",job="apiserver",slo="apiserver-write-response-errors",verb=~"POST|PUT|PATCH|DELETE"})`,
	}, {
		name:      "apiserver-read-resource-latency",
		objective: objectiveAPIServerLatency(),
		expected:  `sum(apiserver_request_duration_seconds:increase2w{job="apiserver",le="",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"}) - sum(apiserver_request_duration_seconds:increase2w{job="apiserver",le="0.1",resource=~"resource|",slo="apiserver-read-resource-latency",verb=~"LIST|GET"})`,
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
		expected:  `((1 - 0.99) - (sum(http_requests:increase4w{code=~"5..",job="thanos-receive-default",slo="monitoring-http-errors"} or vector(0)) / sum(http_requests:increase4w{job="thanos-receive-default",slo="monitoring-http-errors"}))) / (1 - 0.99)`,
	}, {
		name:      "http-ratio-grouping",
		objective: objectiveHTTPRatioGrouping(),
		expected:  `((1 - 0.99) - (sum(http_requests:increase4w{code=~"5..",job="thanos-receive-default",slo="monitoring-http-errors"} or vector(0)) / sum(http_requests:increase4w{job="thanos-receive-default",slo="monitoring-http-errors"}))) / (1 - 0.99)`,
	}, {
		name:      "http-ratio-grouping-regex",
		objective: objectiveHTTPRatioGroupingRegex(),
		expected:  `((1 - 0.99) - (sum(http_requests:increase4w{code=~"5..",handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"} or vector(0)) / sum(http_requests:increase4w{handler=~"/api.*",job="thanos-receive-default",slo="monitoring-http-errors"}))) / (1 - 0.99)`,
	}, {
		name:      "grpc-ratio",
		objective: objectiveGRPCRatio(),
		expected:  `((1 - 0.999) - (sum(grpc_server_handled:increase4w{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} or vector(0)) / sum(grpc_server_handled:increase4w{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"}))) / (1 - 0.999)`,
	}, {
		name:      "grpc-ratio-grouping",
		objective: objectiveGRPCRatioGrouping(),
		expected:  `((1 - 0.999) - (sum(grpc_server_handled:increase4w{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"} or vector(0)) / sum(grpc_server_handled:increase4w{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",slo="monitoring-grpc-errors"}))) / (1 - 0.999)`,
	}, {
		name:      "http-latency",
		objective: objectiveHTTPLatency(),
		expected:  `((1 - 0.995) - (1 - sum(http_request_duration_seconds:increase4w{code=~"2..",job="metrics-service-thanos-receive-default",le="1",slo="monitoring-http-latency"} or vector(0)) / sum(http_request_duration_seconds:increase4w{code=~"2..",job="metrics-service-thanos-receive-default",le="",slo="monitoring-http-latency"}))) / (1 - 0.995)`,
	}, {
		name:      "http-latency-grouping",
		objective: objectiveHTTPLatencyGrouping(),
		expected:  `((1 - 0.995) - (1 - sum(http_request_duration_seconds:increase4w{code=~"2..",job="metrics-service-thanos-receive-default",le="1",slo="monitoring-http-latency"} or vector(0)) / sum(http_request_duration_seconds:increase4w{code=~"2..",job="metrics-service-thanos-receive-default",le="",slo="monitoring-http-latency"}))) / (1 - 0.995)`,
	}, {
		name:      "http-latency-grouping-regex",
		objective: objectiveHTTPLatencyGroupingRegex(),
		expected:  `((1 - 0.995) - (1 - sum(http_request_duration_seconds:increase4w{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1",slo="monitoring-http-latency"} or vector(0)) / sum(http_request_duration_seconds:increase4w{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="",slo="monitoring-http-latency"}))) / (1 - 0.995)`,
	}, {
		name:      "grpc-latency",
		objective: objectiveGRPCLatency(),
		expected:  `((1 - 0.995) - (1 - sum(grpc_server_handling_seconds:increase1w{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6",slo="monitoring-grpc-latency"} or vector(0)) / sum(grpc_server_handling_seconds:increase1w{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="",slo="monitoring-grpc-latency"}))) / (1 - 0.995)`,
	}, {
		name:      "grpc-latency-regex",
		objective: objectiveGRPCLatencyGrouping(),
		expected:  `((1 - 0.995) - (1 - sum(grpc_server_handling_seconds:increase1w{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6",slo="monitoring-grpc-latency"} or vector(0)) / sum(grpc_server_handling_seconds:increase1w{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="",slo="monitoring-grpc-latency"}))) / (1 - 0.995)`,
	}, {
		name:      "operator-ratio",
		objective: objectiveOperator(),
		expected:  `((1 - 0.99) - (sum(prometheus_operator_reconcile_errors:increase2w{slo="monitoring-prometheus-operator-errors"} or vector(0)) / sum(prometheus_operator_reconcile_operations:increase2w{slo="monitoring-prometheus-operator-errors"}))) / (1 - 0.99)`,
	}, {
		name:      "operator-ratio-grouping",
		objective: objectiveOperatorGrouping(),
		expected:  `((1 - 0.99) - (sum(prometheus_operator_reconcile_errors:increase2w{slo="monitoring-prometheus-operator-errors"} or vector(0)) / sum(prometheus_operator_reconcile_operations:increase2w{slo="monitoring-prometheus-operator-errors"}))) / (1 - 0.99)`,
	}, {
		name:      "apiserver-write-response-errors",
		objective: objectiveAPIServerRatio(),
		expected:  `((1 - 99) - (sum(apiserver_request:increase2w{code=~"5..",job="apiserver",slo="apiserver-write-response-errors",verb=~"POST|PUT|PATCH|DELETE"} or vector(0)) / sum(apiserver_request:increase2w{job="apiserver",slo="apiserver-write-response-errors",verb=~"POST|PUT|PATCH|DELETE"}))) / (1 - 99)`,
	}, {
		name:      "apiserver-read-resource-latency",
		objective: objectiveAPIServerRatio(),
		expected:  `((1 - 99) - (sum(apiserver_request:increase2w{code=~"5..",job="apiserver",slo="apiserver-write-response-errors",verb=~"POST|PUT|PATCH|DELETE"} or vector(0)) / sum(apiserver_request:increase2w{job="apiserver",slo="apiserver-write-response-errors",verb=~"POST|PUT|PATCH|DELETE"}))) / (1 - 99)`,
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
		name:      "http-ratio-grouping",
		objective: objectiveHTTPRatioGrouping(),
		timerange: 6 * time.Hour,
		expected:  `sum by(code) (rate(http_requests_total{job="thanos-receive-default"}[6h])) > 0`,
	}, {
		name:      "http-ratio-grouping-regex",
		objective: objectiveHTTPRatioGroupingRegex(),
		timerange: 6 * time.Hour,
		expected:  `sum by(code) (rate(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}[6h])) > 0`,
	}, {
		name:      "grpc-ratio",
		objective: objectiveGRPCRatio(),
		timerange: 6 * time.Hour,
		expected:  `sum by(grpc_code) (rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[6h])) > 0`,
	}, {
		name:      "grpc-ratio-grouping",
		objective: objectiveGRPCRatioGrouping(),
		timerange: 6 * time.Hour,
		expected:  `sum by(grpc_code) (rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[6h])) > 0`,
	}, {
		name:      "http-latency",
		objective: objectiveHTTPLatency(),
		timerange: 2 * time.Hour,
		expected:  `sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[2h]))`,
	}, {
		name:      "http-latency-grouping",
		objective: objectiveHTTPLatencyGrouping(),
		timerange: 2 * time.Hour,
		expected:  `sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[2h]))`,
	}, {
		name:      "http-latency-grouping-regex",
		objective: objectiveHTTPLatencyGroupingRegex(),
		timerange: 2 * time.Hour,
		expected:  `sum(rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[2h]))`,
	}, {
		name:      "grpc-latency",
		objective: objectiveGRPCLatency(),
		timerange: 3 * time.Hour,
		expected:  `sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[3h]))`,
	}, {
		name:      "grpc-latency-grouping",
		objective: objectiveGRPCLatencyGrouping(),
		timerange: 3 * time.Hour,
		expected:  `sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[3h]))`,
	}, {
		name:      "operator-ratio",
		objective: objectiveOperator(),
		timerange: 5 * time.Minute,
		expected:  `sum(rate(prometheus_operator_reconcile_operations_total[5m])) > 0`,
	}, {
		name:      "operator-ratio-grouping",
		objective: objectiveOperatorGrouping(),
		timerange: 5 * time.Minute,
		expected:  `sum(rate(prometheus_operator_reconcile_operations_total[5m])) > 0`,
	}, {
		name:      "apiserver-write-response-errors",
		objective: objectiveAPIServerRatio(),
		timerange: 2 * time.Hour,
		expected:  `sum by(code) (rate(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[2h])) > 0`,
	}, {
		name:      "apiserver-read-resource-latency",
		objective: objectiveAPIServerLatency(),
		timerange: 2 * time.Hour,
		expected:  `sum(rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[2h]))`,
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
		expected:  `sum by(code) (rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[6h])) / scalar(sum(rate(http_requests_total{job="thanos-receive-default"}[6h]))) > 0`,
	}, {
		name:      "http-ratio-grouping",
		objective: objectiveHTTPRatioGrouping(),
		timerange: 6 * time.Hour,
		expected:  `sum by(code) (rate(http_requests_total{code=~"5..",job="thanos-receive-default"}[6h])) / scalar(sum(rate(http_requests_total{job="thanos-receive-default"}[6h]))) > 0`,
	}, {
		name:      "http-ratio-grouping-regex",
		objective: objectiveHTTPRatioGroupingRegex(),
		timerange: 6 * time.Hour,
		expected:  `sum by(code) (rate(http_requests_total{code=~"5..",handler=~"/api.*",job="thanos-receive-default"}[6h])) / scalar(sum(rate(http_requests_total{handler=~"/api.*",job="thanos-receive-default"}[6h]))) > 0`,
	}, {
		name:      "grpc-ratio",
		objective: objectiveGRPCRatio(),
		timerange: 6 * time.Hour,
		expected:  `sum by(grpc_code) (rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[6h])) / scalar(sum(rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[6h]))) > 0`,
	}, {
		name:      "grpc-ratio-grouping",
		objective: objectiveGRPCRatioGrouping(),
		timerange: 6 * time.Hour,
		expected:  `sum by(grpc_code) (rate(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[6h])) / scalar(sum(rate(grpc_server_handled_total{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[6h]))) > 0`,
	}, {
		name:      "http-latency",
		objective: objectiveHTTPLatency(),
		timerange: time.Hour,
		expected:  `(sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1h])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[1h]))) / sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1h]))`,
	}, {
		name:      "http-latency-grouping",
		objective: objectiveHTTPLatencyGrouping(),
		timerange: time.Hour,
		expected:  `(sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1h])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",job="metrics-service-thanos-receive-default",le="1"}[1h]))) / sum(rate(http_request_duration_seconds_count{code=~"2..",job="metrics-service-thanos-receive-default"}[1h]))`,
	}, {
		name:      "http-latency-grouping-regex",
		objective: objectiveHTTPLatencyGroupingRegex(),
		timerange: time.Hour,
		expected:  `(sum(rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[1h])) - sum(rate(http_request_duration_seconds_bucket{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default",le="1"}[1h]))) / sum(rate(http_request_duration_seconds_count{code=~"2..",handler=~"/api.*",job="metrics-service-thanos-receive-default"}[1h]))`,
	}, {
		name:      "grpc-latency",
		objective: objectiveGRPCLatency(),
		timerange: time.Hour,
		expected:  `(sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1h])) - sum(rate(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[1h]))) / sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1h]))`,
	}, {
		name:      "grpc-latency-grouping",
		objective: objectiveGRPCLatencyGrouping(),
		timerange: time.Hour,
		expected:  `(sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1h])) - sum(rate(grpc_server_handling_seconds_bucket{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api",le="0.6"}[1h]))) / sum(rate(grpc_server_handling_seconds_count{grpc_method="Write",grpc_service="conprof.WritableProfileStore",job="api"}[1h]))`,
	}, {
		name:      "operator-ratio",
		objective: objectiveOperator(),
		timerange: 5 * time.Minute,
		expected:  `sum(rate(prometheus_operator_reconcile_errors_total[5m])) / scalar(sum(rate(prometheus_operator_reconcile_operations_total[5m]))) > 0`,
	}, {
		name:      "operator-ratio-grouping",
		objective: objectiveOperatorGrouping(),
		timerange: 5 * time.Minute,
		expected:  `sum(rate(prometheus_operator_reconcile_errors_total[5m])) / scalar(sum(rate(prometheus_operator_reconcile_operations_total[5m]))) > 0`,
	}, {
		name:      "apiserver-write-response-errors",
		objective: objectiveAPIServerRatio(),
		timerange: 2 * time.Hour,
		expected:  `sum by(code) (rate(apiserver_request_total{code=~"5..",job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[2h])) / scalar(sum(rate(apiserver_request_total{job="apiserver",verb=~"POST|PUT|PATCH|DELETE"}[2h]))) > 0`,
	}, {
		name:      "apiserver-read-resource-latency",
		objective: objectiveAPIServerLatency(),
		timerange: 2 * time.Hour,
		expected:  `(sum(rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[2h])) - sum(rate(apiserver_request_duration_seconds_bucket{job="apiserver",le="0.1",resource=~"resource|",verb=~"LIST|GET"}[2h]))) / sum(rate(apiserver_request_duration_seconds_count{job="apiserver",resource=~"resource|",verb=~"LIST|GET"}[2h]))`,
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.objective.ErrorsRange(tc.timerange))
		})
	}
}
