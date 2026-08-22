package main

import (
	"context"
	"math"
	"testing"
	"time"

	connect "connectrpc.com/connect"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/go-kit/log"
	prometheusapiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"

	objectivesv1alpha1 "github.com/pyrra-dev/pyrra/proto/objectives/v1alpha1"
	"github.com/pyrra-dev/pyrra/slo"
)

func TestMatrixToValues(t *testing.T) {
	// v0 is from t 0-500 counting all up from 0 to 500
	v0 := make([]model.SamplePair, 500)
	e0 := [][]float64{
		make([]float64, 500),
		make([]float64, 500),
	}
	for i := 0; i < cap(v0); i++ {
		v0[i] = model.SamplePair{
			Timestamp: model.Time(i * 1000),
			Value:     model.SampleValue(i),
		}
		e0[0][i] = float64(i)
		e0[1][i] = float64(i)
	}

	v10 := make([]model.SamplePair, 100)
	for i := 0; i < cap(v10); i++ {
		v10[i] = model.SamplePair{
			Timestamp: model.Time(i * 1000),
			Value:     model.SampleValue(i),
		}
	}
	// offset by first 50 samples
	v11 := make([]model.SamplePair, 250)
	for i := 0; i < cap(v11); i++ {
		v11[i] = model.SamplePair{
			Timestamp: model.Time((i + 50) * 1000),
			Value:     model.SampleValue(i),
		}
	}

	e1 := [][]float64{
		make([]float64, 300), // [0-100] + [50-300]
		make([]float64, 300),
		make([]float64, 300),
	}
	for i := 0; i < 300; i++ {
		e1[0][i] = float64(i)
	}
	for i := 0; i < 100; i++ {
		e1[1][i] = float64(i)
	}
	for i := 0; i < 250; i++ {
		e1[2][50+i] = float64(i)
	}

	// Check if NaNs are returned as 0 (it's fine for errors for example to convert these).
	// Additionally, NaNs aren't possible to be marshalled to JSON. Not sure if there's a better way.
	v2 := make([]model.SamplePair, 100)
	for i := 0; i < cap(v2); i++ {
		v2[i] = model.SamplePair{
			Timestamp: model.Time(i * 1000),
			Value:     model.SampleValue(math.NaN()),
		}
	}
	e2 := [][]float64{
		make([]float64, 100),
		make([]float64, 100),
	}
	for i := 0; i < len(e2[0]); i++ {
		e2[0][i] = float64(i)
	}

	// Check NaN in multiple series
	v3 := make([]model.SamplePair, 100)
	for i := 0; i < len(v3); i++ {
		value := float64(i)
		if i%11 == 0 {
			value = math.NaN()
		}
		v3[i] = model.SamplePair{
			Timestamp: model.Time(i * 1000),
			Value:     model.SampleValue(value),
		}
	}
	e3 := [][]float64{
		make([]float64, 100), // x
		make([]float64, 100), // y[0]
		make([]float64, 100), // y[1]
	}
	for i := 0; i < len(e3[0]); i++ {
		e32value := float64(i)
		if i%11 == 0 {
			e32value = 0
		}
		e3[0][i] = float64(i)
		e3[1][i] = 0
		e3[2][i] = e32value
	}

	for _, tc := range []struct {
		name     string
		m        []*model.SampleStream
		expected [][]float64
	}{{
		name: "empty",
	}, {
		name:     "simple",
		m:        []*model.SampleStream{{Values: v0}},
		expected: e0,
	}, {
		name:     "overlapping",
		m:        []*model.SampleStream{{Values: v10}, {Values: v11}},
		expected: e1,
	}, {
		name:     "NaN",
		m:        []*model.SampleStream{{Values: v2}},
		expected: e2,
	}, {
		name:     "NaNMultiple",
		m:        []*model.SampleStream{{Values: v2}, {Values: v3}},
		expected: e3,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, matrixToValues(tc.m))
		})
	}
}

func BenchmarkMatrixToValues(b *testing.B) {
	b.Run("one", func(b *testing.B) {
		v := make([]model.SamplePair, b.N)
		for i := 0; i < b.N; i++ {
			v[i] = model.SamplePair{
				Timestamp: model.Time(i * 1000),
				Value:     model.SampleValue(i),
			}
		}

		b.ResetTimer()
		b.ReportAllocs()
		matrixToValues([]*model.SampleStream{{Values: v}})
	})

	b.Run("two", func(b *testing.B) {
		m := make([]*model.SampleStream, 2)
		for n := 0; n < 2; n++ {
			m[n] = &model.SampleStream{Values: make([]model.SamplePair, b.N)}
		}
		for i := 0; i < b.N; i++ {
			for n := 0; n < 2; n++ {
				m[n].Values[i] = model.SamplePair{
					Timestamp: model.Time(i * 1000),
					Value:     model.SampleValue(i),
				}
			}
		}

		b.ReportAllocs()
		b.ResetTimer()
		matrixToValues(m)
	})

	b.Run("five", func(b *testing.B) {
		m := make([]*model.SampleStream, 5)
		for n := 0; n < 5; n++ {
			m[n] = &model.SampleStream{Values: make([]model.SamplePair, b.N)}
		}
		for i := 0; i < b.N; i++ {
			for n := 0; n < 5; n++ {
				m[n].Values[i] = model.SamplePair{
					Timestamp: model.Time(i * 1000),
					Value:     model.SampleValue(i),
				}
			}
		}

		b.ReportAllocs()
		b.ResetTimer()
		matrixToValues(m)
	})
}

func TestAlertsMatchingObjectives(t *testing.T) {
	testcases := []struct {
		name       string
		metrics    []*model.Sample
		objectives []slo.Objective
		inactive   bool
		alerts     []*objectivesv1alpha1.Alert
	}{{
		name: "firing",
		metrics: []*model.Sample{{
			Metric: model.Metric{
				model.MetricNameLabel: "ALERTS",
				"alertname":           "ErrorBudgetBurn",
				"alertstate":          "firing",
				"job":                 "prometheus",
				"long":                "2d",
				"severity":            "warning",
				"short":               "3h",
				"slo":                 "prometheus-rule-evaluation-failures",
			},
		}},
		objectives: []slo.Objective{{
			Labels: labels.New(
				labels.Label{Name: model.MetricNameLabel, Value: "prometheus-rule-evaluation-failures"},
				labels.Label{Name: "namespace", Value: "monitoring"},
			),
			Window: model.Duration(14 * 24 * time.Hour),
		}},
		alerts: []*objectivesv1alpha1.Alert{{
			// In the UI we identify the SLO by these labels.
			Labels: map[string]string{
				model.MetricNameLabel: "prometheus-rule-evaluation-failures",
				"namespace":           "monitoring",
				"job":                 "prometheus",
			},
			Severity: "warning",
			State:    objectivesv1alpha1.Alert_firing,
			For:      durationpb.New(90 * time.Minute),
			Factor:   1,
			Short: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(3 * time.Hour),
				Current: -1,
				Query:   "",
			},
			Long: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(48 * time.Hour),
				Current: -1,
				Query:   "",
			},
		}},
	}, {
		name:    "inactive",
		metrics: []*model.Sample{},
		objectives: []slo.Objective{{
			Labels: labels.New(
				labels.Label{Name: model.MetricNameLabel, Value: "prometheus-rule-evaluation-failures"},
				labels.Label{Name: "namespace", Value: "monitoring"},
			),
			Window: model.Duration(14 * 24 * time.Hour),
		}},
		inactive: true,
		alerts: []*objectivesv1alpha1.Alert{{
			Labels: map[string]string{
				model.MetricNameLabel: "prometheus-rule-evaluation-failures",
				"namespace":           "monitoring",
				//"alertname":       "ErrorBudgetBurn",
				//"job":             "prometheus",
			},
			Severity: "critical",
			State:    objectivesv1alpha1.Alert_inactive,
			For:      durationpb.New(time.Minute),
			Factor:   14,
			Short: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(3 * time.Minute),
				Current: -1,
			},
			Long: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(30 * time.Minute),
				Current: -1,
			},
		}, {
			Labels: map[string]string{
				model.MetricNameLabel: "prometheus-rule-evaluation-failures",
				"namespace":           "monitoring",
				//"alertname":       "ErrorBudgetBurn",
				//"job":             "prometheus",
			},
			Severity: "critical",
			State:    objectivesv1alpha1.Alert_inactive,
			For:      durationpb.New(8 * time.Minute),
			Factor:   7,
			Short: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(15 * time.Minute),
				Current: -1,
			},
			Long: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(3 * time.Hour),
				Current: -1,
			},
		}, {
			Labels: map[string]string{
				model.MetricNameLabel: "prometheus-rule-evaluation-failures",
				"namespace":           "monitoring",
				//"alertname":       "ErrorBudgetBurn",
				//"job":             "prometheus",
			},
			Severity: "warning",
			State:    objectivesv1alpha1.Alert_inactive,
			For:      durationpb.New(30 * time.Minute),
			Factor:   2,
			Short: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(time.Hour),
				Current: -1,
			},
			Long: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(12 * time.Hour),
				Current: -1,
			},
		}, {
			Labels: map[string]string{
				model.MetricNameLabel: "prometheus-rule-evaluation-failures",
				"namespace":           "monitoring",
				//"alertname":       "ErrorBudgetBurn",
				//"job":             "prometheus",
			},
			Severity: "warning",
			State:    objectivesv1alpha1.Alert_inactive,
			For:      durationpb.New(90 * time.Minute),
			Factor:   1,
			Short: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(3 * time.Hour),
				Current: -1,
			},
			Long: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(48 * time.Hour),
				Current: -1,
			},
		}},
	}, {
		name: "mixed",
		metrics: []*model.Sample{{
			Metric: model.Metric{
				model.MetricNameLabel: "ALERTS",
				"alertname":           "ErrorBudgetBurn",
				"alertstate":          "firing",
				"job":                 "prometheus",
				"long":                "2d",
				"severity":            "warning",
				"short":               "3h",
				"slo":                 "prometheus-rule-evaluation-failures",
			},
		}},
		objectives: []slo.Objective{{
			Labels: labels.New(
				labels.Label{Name: model.MetricNameLabel, Value: "prometheus-rule-evaluation-failures"},
				labels.Label{Name: "namespace", Value: "monitoring"},
			),
			Window: model.Duration(14 * 24 * time.Hour),
		}},
		inactive: true,
		alerts: []*objectivesv1alpha1.Alert{{
			Labels: map[string]string{
				model.MetricNameLabel: "prometheus-rule-evaluation-failures",
				"namespace":           "monitoring",
			},
			Severity: "critical",
			State:    objectivesv1alpha1.Alert_inactive,
			For:      durationpb.New(time.Minute),
			Factor:   14,
			Short: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(3 * time.Minute),
				Current: -1,
			},
			Long: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(30 * time.Minute),
				Current: -1,
			},
		}, {
			Labels: map[string]string{
				model.MetricNameLabel: "prometheus-rule-evaluation-failures",
				"namespace":           "monitoring",
			},
			Severity: "critical",
			State:    objectivesv1alpha1.Alert_inactive,
			For:      durationpb.New(8 * time.Minute),
			Factor:   7,
			Short: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(15 * time.Minute),
				Current: -1,
			},
			Long: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(3 * time.Hour),
				Current: -1,
			},
		}, {
			Labels: map[string]string{
				model.MetricNameLabel: "prometheus-rule-evaluation-failures",
				"namespace":           "monitoring",
			},
			Severity: "warning",
			State:    objectivesv1alpha1.Alert_inactive,
			For:      durationpb.New(30 * time.Minute),
			Factor:   2,
			Short: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(time.Hour),
				Current: -1,
			},
			Long: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(12 * time.Hour),
				Current: -1,
			},
		}, {
			Labels: map[string]string{
				model.MetricNameLabel: "prometheus-rule-evaluation-failures",
				"namespace":           "monitoring",
			},
			Severity: "warning",
			State:    objectivesv1alpha1.Alert_firing,
			For:      durationpb.New(90 * time.Minute),
			Factor:   1,
			Short: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(3 * time.Hour),
				Current: -1,
			},
			Long: &objectivesv1alpha1.Burnrate{
				Window:  durationpb.New(48 * time.Hour),
				Current: -1,
			},
		}},
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.alerts, alertsMatchingObjectives(tc.metrics, tc.objectives, nil, tc.inactive))
		})
	}
}

func TestObjectiveServerPreview(t *testing.T) {
	config := `apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: prometheus-api-query
spec:
  target: '99.0'
  window: 7d
  indicator:
    ratio:
      errors:
        metric: prometheus_http_requests_total{handler=~"/api.*",code=~"5.."}
      total:
        metric: prometheus_http_requests_total{handler=~"/api.*"}
      grouping:
        - handler
`

	server := &objectiveServer{}
	resp, err := server.Preview(context.Background(), connect.NewRequest(&objectivesv1alpha1.PreviewRequest{
		Config: config,
	}))
	require.NoError(t, err)

	o := resp.Msg.Objective
	require.NotNil(t, o)
	require.Equal(t, "prometheus-api-query", o.Labels[model.MetricNameLabel])
	require.Equal(t, 0.99, o.Target)
	require.Equal(t, 7*24*time.Hour, o.Window.AsDuration())
	require.NotNil(t, o.Indicator.GetRatio())

	// The queries are filled in exactly like List does, so the UI can render the
	// detail page against real Prometheus data without persisting the SLO.
	require.NotEmpty(t, o.Queries.CountTotal)
	require.NotEmpty(t, o.Queries.CountErrors)
	require.NotEmpty(t, o.Queries.GraphErrorBudget)
	require.NotEmpty(t, o.Queries.GraphRequests)
	require.NotEmpty(t, o.Queries.GraphErrors)
}

func TestObjectiveServerPreviewInvalid(t *testing.T) {
	server := &objectiveServer{}
	_, err := server.Preview(context.Background(), connect.NewRequest(&objectivesv1alpha1.PreviewRequest{
		Config: `apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: broken
spec:
  target: not-a-number
  window: 7d
  indicator:
    ratio:
      errors:
        metric: foo_total{code=~"5.."}
      total:
        metric: foo_total
`,
	}))
	require.Error(t, err)
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

// fakePrometheus records the queries it is asked to run and replays a canned
// matrix, so the duration handlers can be tested without a Prometheus.
type fakePrometheus struct {
	queries []string
	matrix  model.Matrix
}

func (f *fakePrometheus) Query(_ context.Context, _ string, _ time.Time, _ ...prometheusapiv1.Option) (model.Value, prometheusapiv1.Warnings, error) {
	panic("not implemented")
}

func (f *fakePrometheus) QueryRange(_ context.Context, query string, _ prometheusapiv1.Range, _ ...prometheusapiv1.Option) (model.Value, prometheusapiv1.Warnings, error) {
	f.queries = append(f.queries, query)
	return f.matrix, nil, nil
}

func (f *fakePrometheus) LabelNames(_ context.Context, _ []string, _, _ time.Time, _ ...prometheusapiv1.Option) (model.LabelNames, prometheusapiv1.Warnings, error) {
	panic("not implemented")
}

func (f *fakePrometheus) LabelValues(_ context.Context, _ string, _ []string, _, _ time.Time, _ ...prometheusapiv1.Option) (model.LabelValues, prometheusapiv1.Warnings, error) {
	panic("not implemented")
}

func newDurationServer(t *testing.T, matrix model.Matrix) (*objectiveServer, *fakePrometheus) {
	t.Helper()

	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 100,
		MaxCost:     1000,
		BufferItems: 64,
	})
	require.NoError(t, err)

	prom := &fakePrometheus{matrix: matrix}
	return &objectiveServer{
		logger:  log.NewNopLogger(),
		promAPI: &promCache{api: prom, cache: cache},
	}, prom
}

func durationMatrix() model.Matrix {
	return model.Matrix{{
		Metric: model.Metric{},
		Values: []model.SamplePair{
			{Timestamp: model.Time(0), Value: 0.1},
			{Timestamp: model.Time(1000), Value: 0.2},
		},
	}}
}

const latencyConfig = `apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: caddy-response-latency
spec:
  target: '99'
  window: 4w
  indicator:
    latency:
      success:
        metric: caddy_http_response_duration_seconds_bucket{job="caddy",le="0.05"}
      total:
        metric: caddy_http_response_duration_seconds_count{job="caddy"}
`

func TestObjectiveServerPreviewGraphDuration(t *testing.T) {
	server, prom := newDurationServer(t, durationMatrix())

	resp, err := server.PreviewGraphDuration(context.Background(), connect.NewRequest(&objectivesv1alpha1.PreviewGraphDurationRequest{
		Config: latencyConfig,
	}))
	require.NoError(t, err)

	// Percentiles at or below the 99% target: 99, 95, 90, 50.
	require.Len(t, resp.Msg.Timeseries, 4)
	require.Equal(t, []string{`{quantile="p99"}`}, resp.Msg.Timeseries[0].Labels)
	require.Equal(t, []string{`{quantile="p50"}`}, resp.Msg.Timeseries[3].Labels)

	// The whole point of the preview variant: the queries read the underlying
	// histogram directly, so they work before any recording rule exists.
	require.Len(t, prom.queries, 4)
	for _, query := range prom.queries {
		require.Contains(t, query, "histogram_quantile(")
		require.Contains(t, query, "rate(")
		require.Contains(t, query, "caddy_http_response_duration_seconds_bucket")
		require.NotContains(t, query, "pyrra_")
	}
}

func TestObjectiveServerPreviewGraphDurationGrouping(t *testing.T) {
	config := `apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: pyrra-latency-native
spec:
  target: '99'
  window: 4w
  indicator:
    latencyNative:
      latency: 200ms
      total:
        metric: connect_server_requests_duration_seconds{job="pyrra"}
      grouping:
        - service
`

	server, prom := newDurationServer(t, durationMatrix())

	_, err := server.PreviewGraphDuration(context.Background(), connect.NewRequest(&objectivesv1alpha1.PreviewGraphDurationRequest{
		Config:   config,
		Grouping: `{service="query"}`,
	}))
	require.NoError(t, err)

	// Regression test: the grouping used to be dropped for latencyNative, so the
	// graph plotted every series instead of the selected one.
	require.NotEmpty(t, prom.queries)
	for _, query := range prom.queries {
		require.Contains(t, query, `service="query"`)
	}
}

func TestMergeGroupingMatchers(t *testing.T) {
	matchers := []*labels.Matcher{{Type: labels.MatchEqual, Name: "service", Value: "query"}}

	for _, tc := range []struct {
		name   string
		config string
		// query is a substring every generated matcher set must end up in.
		indicator func(o slo.Objective) []*labels.Matcher
	}{{
		name: "ratio",
		config: `apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: ratio
spec:
  target: '99'
  window: 4w
  indicator:
    ratio:
      errors:
        metric: foo_total{code=~"5.."}
      total:
        metric: foo_total
`,
		indicator: func(o slo.Objective) []*labels.Matcher { return o.Indicator.Ratio.Total.LabelMatchers },
	}, {
		name:      "latency",
		config:    latencyConfig,
		indicator: func(o slo.Objective) []*labels.Matcher { return o.Indicator.Latency.Total.LabelMatchers },
	}, {
		name: "latencyNative",
		config: `apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: latency-native
spec:
  target: '99'
  window: 4w
  indicator:
    latencyNative:
      latency: 200ms
      total:
        metric: foo_duration_seconds{job="pyrra"}
`,
		indicator: func(o slo.Objective) []*labels.Matcher {
			return o.Indicator.LatencyNative.Total.LabelMatchers
		},
	}, {
		name: "boolGauge",
		config: `apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: bool-gauge
spec:
  target: '99'
  window: 4w
  indicator:
    bool_gauge:
      metric: up{job="prometheus"}
`,
		indicator: func(o slo.Objective) []*labels.Matcher { return o.Indicator.BoolGauge.LabelMatchers },
	}} {
		t.Run(tc.name, func(t *testing.T) {
			objective, err := objectiveFromConfig(context.Background(), tc.config, "")
			require.NoError(t, err)

			mergeGroupingMatchers(&objective, matchers)
			require.Contains(t, tc.indicator(objective), matchers[0])
		})
	}
}

func TestObjectiveServerPreviewGraphDurationErrors(t *testing.T) {
	t.Run("invalid config", func(t *testing.T) {
		server, _ := newDurationServer(t, durationMatrix())
		_, err := server.PreviewGraphDuration(context.Background(), connect.NewRequest(&objectivesv1alpha1.PreviewGraphDurationRequest{
			Config: "not: valid: yaml:",
		}))
		require.Error(t, err)
		require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	})

	t.Run("not latency based", func(t *testing.T) {
		server, _ := newDurationServer(t, durationMatrix())
		_, err := server.PreviewGraphDuration(context.Background(), connect.NewRequest(&objectivesv1alpha1.PreviewGraphDurationRequest{
			Config: `apiVersion: pyrra.dev/v1alpha1
kind: ServiceLevelObjective
metadata:
  name: ratio
spec:
  target: '99'
  window: 4w
  indicator:
    ratio:
      errors:
        metric: foo_total{code=~"5.."}
      total:
        metric: foo_total
`,
		}))
		require.Error(t, err)
		require.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	})

	t.Run("no data", func(t *testing.T) {
		// An empty matrix for every percentile: NotFound, and notably no panic —
		// this used to hand connect a nil error.
		server, _ := newDurationServer(t, model.Matrix{})
		_, err := server.PreviewGraphDuration(context.Background(), connect.NewRequest(&objectivesv1alpha1.PreviewGraphDurationRequest{
			Config: latencyConfig,
		}))
		require.Error(t, err)
		require.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	})
}
