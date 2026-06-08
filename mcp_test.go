package main

import (
	"context"
	"math"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	toon "github.com/toon-format/toon-go"
)

// TestNewMCPHandler ensures all tools register and their input schemas generate
// from the struct tags without panicking.
func TestNewMCPHandler(t *testing.T) {
	require.NotPanics(t, func() {
		h := newMCPHandler(nil, nil)
		require.NotNil(t, h)
	})
}

// TestMCPListToolsOverHTTP starts the handler and connects a real MCP client
// over HTTP to verify the transport, handshake and tool advertisement.
// tools/list does not touch the backend, so a nil objectiveServer is fine.
func TestMCPListToolsOverHTTP(t *testing.T) {
	srv := httptest.NewServer(newMCPHandler(nil, nil))
	t.Cleanup(srv.Close)

	ctx := context.Background()
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "test", Version: "v0"}, nil)
	session, err := client.Connect(ctx, &mcpsdk.StreamableClientTransport{Endpoint: srv.URL}, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })

	res, err := session.ListTools(ctx, nil)
	require.NoError(t, err)

	got := make([]string, 0, len(res.Tools))
	for _, tool := range res.Tools {
		got = append(got, tool.Name)
	}
	sort.Strings(got)

	want := []string{"get_objective", "list_objectives"}
	require.Equal(t, want, got)
}

func TestListObjectivesTOON(t *testing.T) {
	result := listObjectivesResult{Objectives: []objectiveRow{
		{
			Name: "prometheus-api-errors", Labels: "{handler=/api/v1/query,team=prometheus}",
			Window: "2w", Target: 0.99, Type: "ratio", Latency: "",
			Availability: 100, Budget: 100, AlertsFiring: 0, AlertsPending: 0, AlertsWorst: "",
		},
		{
			Name: "caddy-response-latency", Labels: "{host=demo.pyrra.dev}",
			Window: "4w", Target: 0.9, Type: "latency", Latency: "50ms",
			Availability: 32.79, Budget: -572.08, AlertsFiring: 0, AlertsPending: 0, AlertsWorst: "",
		},
	}}

	out, err := toon.Marshal(result, toon.WithLengthMarkers(true))
	require.NoError(t, err)

	s := string(out)
	t.Log("\n" + s)
	// Tabular form: a single header row with all column names, then bare rows.
	require.True(t, strings.Contains(s, "objectives[#2]"), "want length-marked array header, got:\n%s", s)
	require.True(t, strings.Contains(s, "name") && strings.Contains(s, "availability") && strings.Contains(s, "alerts_worst"))
	require.Contains(t, s, "caddy-response-latency")
}

func TestGetObjectiveGraphTOON(t *testing.T) {
	// A get_objective result with one embedded graph renders the graph nested
	// under its field with a columnar values array.
	res := getObjectiveResult{
		Name: "caddy-response-errors", Target: 0.99, Window: "4w", Type: "ratio",
		ErrorBudget: &graphData{
			Start: 1780918369, Step: 36,
			Current: 1.0,
			Values:  []float64{1.0, 1.0, 0.99, 0.99},
		},
		Duration: &graphData{
			Start: 1780918369, Step: 36,
			Series: []graphSeries{
				{Labels: `{quantile="p90"}`, Current: 0.09, Values: []float64{0.08, 0.09}},
				{Labels: `{quantile="p50"}`, Current: 0.03, Values: []float64{0.02, 0.03}},
			},
		},
	}
	out, err := toon.Marshal(res, toon.WithLengthMarkers(true))
	require.NoError(t, err)
	s := string(out)
	t.Log("\n" + s)
	// Single-series: inline current. Multi-series: per-series current.
	require.Contains(t, s, "error_budget:")
	require.Contains(t, s, "values[#4]")
	require.Contains(t, s, "series[#2]")
	require.Contains(t, s, `quantile=\"p90\"`)
	// One inline current (error_budget) + one per duration series = 3.
	require.Equal(t, 3, strings.Count(s, "current:"))
}

func TestNormalizeMaxPoints(t *testing.T) {
	require.Equal(t, 20, normalizeMaxPoints(0))  // omitted → default
	require.Equal(t, 0, normalizeMaxPoints(-1))  // negative → server default resolution
	require.Equal(t, 25, normalizeMaxPoints(25)) // explicit
}

func TestLastFinite(t *testing.T) {
	require.Equal(t, 2.0, lastFinite([]float64{1, 3, 2}))
	require.Equal(t, 0.0, lastFinite(nil))
	require.Equal(t, 3.0, lastFinite([]float64{3, math.NaN()})) // skips trailing NaN
}

func TestSelectorHelpers(t *testing.T) {
	require.Equal(t, `{__name__="my-slo"}`, selectorName("my-slo"))
	require.Equal(t, "{}", selectorFromLabels(nil))
	require.Equal(t, `{handler="/api",team="ops"}`, selectorFromLabels(map[string]string{"team": "ops", "handler": "/api"}))
	require.Equal(t, "{handler=/api,team=ops}", labelString(map[string]string{"team": "ops", "handler": "/api"}))
}
