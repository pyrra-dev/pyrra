package slo

import (
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/stretchr/testify/require"
)

func TestObjective_QueryTotal(t *testing.T) {
	testcases := []struct {
		name      string
		objective Objective
		expected  string
	}{{
		name: "http",
		objective: Objective{
			Window: model.Duration(24 * time.Hour),
			Indicator: Indicator{
				HTTP: &HTTPIndicator{},
			},
		},
		expected: `sum(increase(http_requests_total[1d]))`,
	}, {
		name: "http-custom",
		objective: Objective{
			Window: model.Duration(24 * time.Hour),
			Indicator: Indicator{HTTP: &HTTPIndicator{
				Metric:        "prometheus_http_requests_total",
				Matchers:      []*labels.Matcher{ParseMatcher(`foo="bar"`)},
				ErrorMatchers: []*labels.Matcher{ParseMatcher(`status=~"5.."`)},
			}},
		},
		expected: `sum(increase(prometheus_http_requests_total{foo="bar"}[1d]))`,
	}, {
		name: "grpc",
		objective: Objective{
			Window: model.Duration(24 * time.Hour),
			Indicator: Indicator{
				GRPC: &GRPCIndicator{
					Service: "service",
					Method:  "method",
				},
			},
		},
		expected: `sum(increase(grpc_server_handled_total{grpc_method="method",grpc_service="service"}[1d]))`,
	}, {
		name: "grpc-custom",
		objective: Objective{
			Window: model.Duration(7 * 24 * time.Hour),
			Indicator: Indicator{
				GRPC: &GRPCIndicator{
					Service:  "awesome",
					Method:   "lightspeed",
					Matchers: []*labels.Matcher{ParseMatcher(`job="app"`)},
				}},
		},
		expected: `sum(increase(grpc_server_handled_total{grpc_method="lightspeed",grpc_service="awesome",job="app"}[1w]))`,
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
		name: "http",
		objective: Objective{
			Window: model.Duration(28 * 24 * time.Hour),
			Indicator: Indicator{
				HTTP: &HTTPIndicator{},
			},
		},
		expected: `sum(increase(http_requests_total{code=~"5.."}[4w]))`,
	}, {
		name: "http-custom",
		objective: Objective{
			Window: model.Duration(28 * 24 * time.Hour),
			Indicator: Indicator{
				HTTP: &HTTPIndicator{
					Metric:        "prometheus_http_requests_total",
					Matchers:      []*labels.Matcher{ParseMatcher(`job="prometheus"`)},
					ErrorMatchers: []*labels.Matcher{ParseMatcher(`status=~"5.."`)},
				},
			},
		},
		expected: `sum(increase(prometheus_http_requests_total{job="prometheus",status=~"5.."}[4w]))`,
	}, {
		name: "grpc",
		objective: Objective{
			Window: model.Duration(24 * time.Hour),
			Indicator: Indicator{
				GRPC: &GRPCIndicator{
					Service: "service",
					Method:  "method",
				},
			},
		},
		expected: `sum(increase(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="method",grpc_service="service"}[1d]))`,
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
		name: "http",
		objective: Objective{
			Window: model.Duration(28 * 24 * time.Hour),
			Target: 0.999,
			Indicator: Indicator{
				HTTP: &HTTPIndicator{},
			},
		},
		expected: `((1 - 0.999) - (sum(increase(http_requests_total{code=~"5.."}[4w])) / sum(increase(http_requests_total[4w])))) / (1 - 0.999)`,
	}, {
		name: "http-custom",
		objective: Objective{
			Window: model.Duration(28 * 24 * time.Hour),
			Target: 0.953,
			Indicator: Indicator{
				HTTP: &HTTPIndicator{
					Metric:        "prometheus_http_requests_total",
					Matchers:      []*labels.Matcher{ParseMatcher(`job="prometheus"`)},
					ErrorMatchers: []*labels.Matcher{ParseMatcher(`status=~"5.."`)},
				},
			},
		},
		expected: `((1 - 0.953) - (sum(increase(prometheus_http_requests_total{job="prometheus",status=~"5.."}[4w])) / sum(increase(prometheus_http_requests_total{job="prometheus"}[4w])))) / (1 - 0.953)`,
	}, {
		name: "grpc",
		objective: Objective{
			Window: model.Duration(28 * 24 * time.Hour),
			Target: 0.999,
			Indicator: Indicator{
				GRPC: &GRPCIndicator{
					Service: "service",
					Method:  "method",
				},
			},
		},
		expected: `((1 - 0.999) - (sum(increase(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="method",grpc_service="service"}[4w])) / sum(increase(grpc_server_handled_total{grpc_method="method",grpc_service="service"}[4w])))) / (1 - 0.999)`,
	}, {
		name: "grpc-custom",
		objective: Objective{
			Window: model.Duration(14 * 24 * time.Hour),
			Target: 0.953,
			Indicator: Indicator{
				GRPC: &GRPCIndicator{
					Service:  "awesome",
					Method:   "lightspeed",
					Matchers: []*labels.Matcher{ParseMatcher(`job="app"`)},
				},
			},
		},
		expected: `((1 - 0.953) - (sum(increase(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="lightspeed",grpc_service="awesome",job="app"}[2w])) / sum(increase(grpc_server_handled_total{grpc_method="lightspeed",grpc_service="awesome",job="app"}[2w])))) / (1 - 0.953)`,
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.objective.QueryErrorBudget())
		})
	}
}
