package slo

import (
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/stretchr/testify/require"
)

var (
	objectiveHTTP = Objective{
		Window: model.Duration(28 * 24 * time.Hour),
		Target: 0.999,
		Indicator: Indicator{
			HTTP: &HTTPIndicator{},
		},
	}
	objectiveHTTPCustom = Objective{
		Window: model.Duration(12 * 24 * time.Hour),
		Target: 0.953,
		Indicator: Indicator{HTTP: &HTTPIndicator{
			Metric:        "prometheus_http_requests_total",
			Matchers:      []*labels.Matcher{ParseMatcher(`foo="bar"`)},
			ErrorMatchers: []*labels.Matcher{ParseMatcher(`status=~"5.."`)},
		}},
	}
	objectiveGRPC = Objective{
		Window: model.Duration(28 * 24 * time.Hour),
		Target: 0.999,
		Indicator: Indicator{
			GRPC: &GRPCIndicator{
				Service: "service",
				Method:  "method",
			},
		},
	}
	objectiveGRPCCustom = Objective{
		Window: model.Duration(7 * 24 * time.Hour),
		Target: 0.953,
		Indicator: Indicator{
			GRPC: &GRPCIndicator{
				Service:  "awesome",
				Method:   "lightspeed",
				Matchers: []*labels.Matcher{ParseMatcher(`job="app"`)},
			}},
	}
)

func TestObjective_QueryTotal(t *testing.T) {
	testcases := []struct {
		name      string
		objective Objective
		expected  string
	}{{
		name:      "http",
		objective: objectiveHTTP,
		expected:  `sum(increase(http_requests_total[4w]))`,
	}, {
		name:      "http-custom",
		objective: objectiveHTTPCustom,
		expected:  `sum(increase(prometheus_http_requests_total{foo="bar"}[12d]))`,
	}, {
		name:      "grpc",
		objective: objectiveGRPC,
		expected:  `sum(increase(grpc_server_handled_total{grpc_method="method",grpc_service="service"}[4w]))`,
	}, {
		name:      "grpc-custom",
		objective: objectiveGRPCCustom,
		expected:  `sum(increase(grpc_server_handled_total{grpc_method="lightspeed",grpc_service="awesome",job="app"}[1w]))`,
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
		name:      "http",
		objective: objectiveHTTP,
		expected:  `sum(increase(http_requests_total{code=~"5.."}[4w]))`,
	}, {
		name:      "http-custom",
		objective: objectiveHTTPCustom,
		expected:  `sum(increase(prometheus_http_requests_total{foo="bar",status=~"5.."}[12d]))`,
	}, {
		name:      "grpc",
		objective: objectiveGRPC,
		expected:  `sum(increase(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="method",grpc_service="service"}[4w]))`,
	}, {
		name:      "grpc-custom",
		objective: objectiveGRPCCustom,
		expected:  `sum(increase(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="lightspeed",grpc_service="awesome",job="app"}[1w]))`,
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
		name:      "http",
		objective: objectiveHTTP,
		expected:  `((1 - 0.999) - (sum(increase(http_requests_total{code=~"5.."}[4w]) or vector(0)) / sum(increase(http_requests_total[4w])))) / (1 - 0.999)`,
	}, {
		name: "http-matchers",
		objective: Objective{
			Window: model.Duration(28 * 24 * time.Hour),
			Target: 0.953,
			Indicator: Indicator{
				HTTP: &HTTPIndicator{
					Metric:   "prometheus_http_requests_total",
					Matchers: []*labels.Matcher{ParseMatcher(`job="prometheus"`)},
				},
			},
		},
		expected: `((1 - 0.953) - (sum(increase(prometheus_http_requests_total{code=~"5..",job="prometheus"}[4w]) or vector(0)) / sum(increase(prometheus_http_requests_total{job="prometheus"}[4w])))) / (1 - 0.953)`,
	}, {
		name:      "http-custom",
		objective: objectiveHTTPCustom,
		expected:  `((1 - 0.953) - (sum(increase(prometheus_http_requests_total{foo="bar",status=~"5.."}[12d]) or vector(0)) / sum(increase(prometheus_http_requests_total{foo="bar"}[12d])))) / (1 - 0.953)`,
	}, {
		name:      "grpc",
		objective: objectiveGRPC,
		expected:  `((1 - 0.999) - (sum(increase(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="method",grpc_service="service"}[4w]) or vector(0)) / sum(increase(grpc_server_handled_total{grpc_method="method",grpc_service="service"}[4w])))) / (1 - 0.999)`,
	}, {
		name:      "grpc-custom",
		objective: objectiveGRPCCustom,
		expected:  `((1 - 0.953) - (sum(increase(grpc_server_handled_total{grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss",grpc_method="lightspeed",grpc_service="awesome",job="app"}[1w]) or vector(0)) / sum(increase(grpc_server_handled_total{grpc_method="lightspeed",grpc_service="awesome",job="app"}[1w])))) / (1 - 0.953)`,
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
		name:      "http",
		objective: objectiveHTTP,
		timerange: 6 * time.Hour,
		expected:  `sum by(code) (rate(http_requests_total[6h]))`,
	}, {
		name:      "http-custom",
		objective: objectiveHTTPCustom,
		timerange: 6 * time.Hour,
		expected:  `sum by(status) (rate(prometheus_http_requests_total{foo="bar"}[6h]))`,
	}, {
		name:      "grpc",
		objective: objectiveGRPC,
		timerange: 24 * time.Hour,
		expected:  `sum by(grpc_code) (rate(grpc_server_handled_total{grpc_method="method",grpc_service="service"}[1d]))`,
	}, {
		name:      "grpc-custom",
		objective: objectiveGRPCCustom,
		timerange: 13 * time.Hour,
		expected:  `sum by(grpc_code) (rate(grpc_server_handled_total{grpc_method="lightspeed",grpc_service="awesome",job="app"}[13h]))`,
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.objective.RequestRange(tc.timerange))
		})
	}
}
