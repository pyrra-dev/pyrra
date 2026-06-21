package main

import (
	"context"
	"fmt"
	"maps"
	"math"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	connect "connectrpc.com/connect"
	"github.com/go-kit/log"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	toon "github.com/toon-format/toon-go"
	"google.golang.org/protobuf/types/known/timestamppb"

	objectivesv1alpha1 "github.com/pyrra-dev/pyrra/proto/objectives/v1alpha1"
)

// mcpServer exposes the Pyrra ObjectiveService as MCP tools. It mirrors the two
// UI pages (List + Detail), decomposed into composable, pull-on-demand tools.
// Responses are encoded as TOON, not JSON, for token efficiency.
type mcpServer struct {
	objectives *objectiveServer
	logger     log.Logger
}

// newMCPHandler builds an MCP server with all tools registered and returns an
// HTTP handler. It uses non-streaming, stateless HTTP (JSON responses, no SSE).
func newMCPHandler(objectives *objectiveServer, logger log.Logger) http.Handler {
	m := &mcpServer{objectives: objectives, logger: logger}

	s := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "pyrra", Version: version}, nil)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "list_objectives",
		Description: "List all Service Level Objectives with their current availability, error budget and alert state merged in. One row per grouping instance (the overview table). Filter with labels/name.",
	}, m.listObjectives)

	mcpsdk.AddTool(s, &mcpsdk.Tool{
		Name:        "get_objective",
		Description: "Get the full detail for one objective (the whole Detail page): identity, current status tiles per grouping instance, multi-burn-rate alerts, and the error-budget / requests / errors / duration timeseries (downsampled, controlled by max_points). The raw YAML config is opt-in via include_config.",
	}, m.getObjective)

	return mcpsdk.NewStreamableHTTPHandler(
		func(*http.Request) *mcpsdk.Server { return s },
		&mcpsdk.StreamableHTTPOptions{JSONResponse: true, Stateless: true},
	)
}

type listObjectivesInput struct {
	Labels map[string]string `json:"labels,omitempty" jsonschema:"filter objectives by exact label match, e.g. {\"team\":\"prometheus\"}"`
	Name   string            `json:"name,omitempty" jsonschema:"limit to a single objective by name (still returns its instances)"`
}

type objectiveInput struct {
	Name          string            `json:"name" jsonschema:"the objective name"`
	Grouping      map[string]string `json:"grouping,omitempty" jsonschema:"restrict status, alerts and graphs to one instance of a multi-dimensional objective, e.g. {\"handler\":\"/api/v1/query\"}"`
	At            string            `json:"at,omitempty" jsonschema:"status snapshot time: 'now', a relative offset like '-1h', or RFC3339"`
	Since         string            `json:"since,omitempty" jsonschema:"graph lookback like '-1h' or '-7d'; defaults to the objective's window"`
	MaxPoints     int               `json:"max_points,omitempty" jsonschema:"max points per graph (default 20; use -1 for full resolution)"`
	IncludeConfig bool              `json:"include_config,omitempty" jsonschema:"include the raw YAML definition (a large blob; off by default)"`
}

type listObjectivesResult struct {
	Objectives []objectiveRow `toon:"objectives"`
}

type objectiveRow struct {
	Name          string  `toon:"name"`
	Labels        string  `toon:"labels"`
	Window        string  `toon:"window"`
	Target        float64 `toon:"target"`
	Type          string  `toon:"type"`
	Latency       string  `toon:"latency"`
	Availability  float64 `toon:"availability"`
	Budget        float64 `toon:"budget"`
	AlertsFiring  int     `toon:"alerts_firing"`
	AlertsPending int     `toon:"alerts_pending"`
	AlertsWorst   string  `toon:"alerts_worst"`
}

type getObjectiveResult struct {
	Name        string      `toon:"name"`
	Description string      `toon:"description"`
	Labels      string      `toon:"labels"`
	Target      float64     `toon:"target"`
	Window      string      `toon:"window"`
	Type        string      `toon:"type"`
	Latency     string      `toon:"latency,omitempty"`
	Queries     *querySet   `toon:"queries,omitempty"`
	Status      []statusRow `toon:"status"`
	Alerts      []alertRow  `toon:"alerts"`
	ErrorBudget *graphData  `toon:"error_budget,omitempty"`
	Requests    *graphData  `toon:"requests,omitempty"`
	Errors      *graphData  `toon:"errors,omitempty"`
	Duration    *graphData  `toon:"duration,omitempty"`
	Config      string      `toon:"config,omitempty"`
}

// querySet holds the raw PromQL behind an objective's status tiles, so an agent
// can re-run them at any resolution, narrow by instance, or pivot to exemplars.
type querySet struct {
	Total  string `toon:"total"`
	Errors string `toon:"errors"`
}

type statusRow struct {
	Labels       string  `toon:"labels"`
	Availability float64 `toon:"availability"`
	Errors       float64 `toon:"errors"`
	Total        float64 `toon:"total"`
	Budget       float64 `toon:"budget"`
}

type alertRow struct {
	State        string  `toon:"state"`
	Severity     string  `toon:"severity"`
	Exhaustion   string  `toon:"exhaustion"`
	Threshold    float64 `toon:"threshold"`
	ShortWindow  string  `toon:"short_window"`
	ShortCurrent float64 `toon:"short_current"`
	LongWindow   string  `toon:"long_window"`
	LongCurrent  float64 `toon:"long_current"`
	For          string  `toon:"for"`
}

// graphData is a single-series graph (error_budget): current + values inline.
// Multi-series graphs leave the top-level fields empty and populate Series.
// Prometheus does the downsampling via the query step (≈ range / max_points),
// so values is already the series at the requested resolution — no client-side
// decimation, and no min/max (that would need extra min/max_over_time queries).
type graphData struct {
	Query   string        `toon:"query,omitempty"`
	Start   int64         `toon:"start,omitempty"`
	Step    int64         `toon:"step,omitempty"`
	Current float64       `toon:"current,omitempty"`
	Values  []float64     `toon:"values,omitempty"`
	Series  []graphSeries `toon:"series,omitempty"`
}

type graphSeries struct {
	Labels  string    `toon:"labels"`
	Current float64   `toon:"current"`
	Values  []float64 `toon:"values,omitempty"`
}

func (m *mcpServer) listObjectives(ctx context.Context, _ *mcpsdk.CallToolRequest, in listObjectivesInput) (*mcpsdk.CallToolResult, any, error) {
	matchers := map[string]string{}
	maps.Copy(matchers, in.Labels)
	if in.Name != "" {
		matchers[model.MetricNameLabel] = in.Name
	}

	list, err := m.objectives.List(ctx, connect.NewRequest(&objectivesv1alpha1.ListRequest{Expr: selectorFromLabels(matchers)}))
	if err != nil {
		return nil, nil, err
	}

	// One bulk GetAlerts for all objectives; summarize per objective name.
	alertSummaries := map[string]*alertSummary{}
	if alerts, err := m.objectives.GetAlerts(ctx, connect.NewRequest(&objectivesv1alpha1.GetAlertsRequest{})); err == nil {
		for _, a := range alerts.Msg.Alerts {
			name := a.GetLabels()[model.MetricNameLabel]
			if name == "" {
				continue
			}
			s := alertSummaries[name]
			if s == nil {
				s = &alertSummary{}
				alertSummaries[name] = s
			}
			s.add(a)
		}
	}

	result := listObjectivesResult{}
	for _, o := range list.Msg.Objectives {
		name := o.GetLabels()[model.MetricNameLabel]
		objLabels := stripName(o.GetLabels())
		typ := indicatorType(o)
		latency := latencyThreshold(o)
		window := humanizeDuration(o.GetWindow().AsDuration())
		sum := alertSummaries[name]

		status, err := m.objectives.GetStatus(ctx, connect.NewRequest(&objectivesv1alpha1.GetStatusRequest{Expr: selectorName(name)}))
		if err != nil {
			// Objectives with no data error out; skip them like the UI does.
			continue
		}
		for _, st := range status.Msg.Status {
			result.Objectives = append(result.Objectives, objectiveRow{
				Name:          name,
				Labels:        labelString(mergeLabels(objLabels, st.GetLabels())),
				Window:        window,
				Target:        o.GetTarget(),
				Type:          typ,
				Latency:       latency,
				Availability:  round(st.GetAvailability().GetPercentage()*100, 3),
				Budget:        round(st.GetBudget().GetRemaining()*100, 3),
				AlertsFiring:  sum.firing(),
				AlertsPending: sum.pending(),
				AlertsWorst:   sum.worst(),
			})
		}
	}

	sort.Slice(result.Objectives, func(i, j int) bool {
		if result.Objectives[i].Name != result.Objectives[j].Name {
			return result.Objectives[i].Name < result.Objectives[j].Name
		}
		return result.Objectives[i].Labels < result.Objectives[j].Labels
	})

	return toonResult(result)
}

func (m *mcpServer) getObjective(ctx context.Context, _ *mcpsdk.CallToolRequest, in objectiveInput) (*mcpsdk.CallToolResult, any, error) {
	o, err := m.listOne(ctx, in.Name)
	if err != nil {
		return nil, nil, err
	}

	req := &objectivesv1alpha1.GetStatusRequest{
		Expr:     selectorName(in.Name),
		Grouping: selectorFromLabels(in.Grouping),
	}
	if in.At != "" {
		t, err := parseTimeArg(time.Now(), in.At)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing 'at': %w", err)
		}
		req.Time = timestamppb.New(t)
	}
	status, err := m.objectives.GetStatus(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, nil, err
	}

	res := getObjectiveResult{
		Name:        in.Name,
		Description: o.GetDescription(),
		Labels:      labelString(stripName(o.GetLabels())),
		Target:      o.GetTarget(),
		Window:      humanizeDuration(o.GetWindow().AsDuration()),
		Type:        indicatorType(o),
		Latency:     latencyThreshold(o),
		Queries:     m.statusQueries(o, in.Grouping),
	}
	for _, st := range status.Msg.Status {
		res.Status = append(res.Status, statusRow{
			Labels:       labelString(st.GetLabels()),
			Availability: round(st.GetAvailability().GetPercentage()*100, 3),
			Errors:       round(st.GetAvailability().GetErrors(), 2),
			Total:        round(st.GetAvailability().GetTotal(), 2),
			Budget:       round(st.GetBudget().GetRemaining()*100, 3),
		})
	}

	// Multi-burn-rate alerts, merged in like the detail page (all windows, with
	// current burn rates).
	alerts, err := m.objectives.GetAlerts(ctx, connect.NewRequest(&objectivesv1alpha1.GetAlertsRequest{
		Expr:     selectorName(in.Name),
		Grouping: selectorFromLabels(in.Grouping),
		Inactive: true,
		Current:  true,
	}))
	if err != nil {
		return nil, nil, err
	}
	res.Alerts = buildAlertRows(alerts.Msg.Alerts, o.GetWindow().AsDuration(), o.GetTarget())

	// Graphs (the detail page's Error Budget / Requests / Errors / Duration).
	// Prometheus does the downsampling: we set the query step to ~range/max_points
	// so it returns about max_points per series. Duration is latency-only.
	now := time.Now()
	start := now.Add(-o.GetWindow().AsDuration())
	if in.Since != "" {
		start, err = parseTimeArg(now, in.Since)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing 'since': %w", err)
		}
	}
	ctxGraph := ctx
	if mp := normalizeMaxPoints(in.MaxPoints); mp > 0 {
		step := max(now.Sub(start)/time.Duration(mp), time.Second)
		ctxGraph = contextSetGraphStep(ctx, step)
	}
	g := graphReq{
		expr:     selectorName(in.Name),
		grouping: selectorFromLabels(in.Grouping),
		start:    timestamppb.New(start),
		end:      timestamppb.New(now),
	}

	if res.ErrorBudget, err = m.fetchGraph(g, true, func(r graphReq) ([]*objectivesv1alpha1.Timeseries, error) {
		resp, err := m.objectives.GraphErrorBudget(ctxGraph, connect.NewRequest(&objectivesv1alpha1.GraphErrorBudgetRequest{Expr: r.expr, Grouping: r.grouping, Start: r.start, End: r.end}))
		if err != nil {
			return nil, err
		}
		return []*objectivesv1alpha1.Timeseries{resp.Msg.GetTimeseries()}, nil
	}); err != nil {
		return nil, nil, err
	}
	if res.Requests, err = m.fetchGraph(g, false, func(r graphReq) ([]*objectivesv1alpha1.Timeseries, error) {
		resp, err := m.objectives.GraphRate(ctxGraph, connect.NewRequest(&objectivesv1alpha1.GraphRateRequest{Expr: r.expr, Grouping: r.grouping, Start: r.start, End: r.end}))
		if err != nil {
			return nil, err
		}
		return []*objectivesv1alpha1.Timeseries{resp.Msg.GetTimeseries()}, nil
	}); err != nil {
		return nil, nil, err
	}
	if res.Errors, err = m.fetchGraph(g, false, func(r graphReq) ([]*objectivesv1alpha1.Timeseries, error) {
		resp, err := m.objectives.GraphErrors(ctxGraph, connect.NewRequest(&objectivesv1alpha1.GraphErrorsRequest{Expr: r.expr, Grouping: r.grouping, Start: r.start, End: r.end}))
		if err != nil {
			return nil, err
		}
		return []*objectivesv1alpha1.Timeseries{resp.Msg.GetTimeseries()}, nil
	}); err != nil {
		return nil, nil, err
	}
	if indicatorType(o) == "latency" || indicatorType(o) == "latency_native" {
		if res.Duration, err = m.fetchGraph(g, false, func(r graphReq) ([]*objectivesv1alpha1.Timeseries, error) {
			resp, err := m.objectives.GraphDuration(ctxGraph, connect.NewRequest(&objectivesv1alpha1.GraphDurationRequest{Expr: r.expr, Grouping: r.grouping, Start: r.start, End: r.end}))
			if err != nil {
				return nil, err
			}
			return resp.Msg.GetTimeseries(), nil
		}); err != nil {
			return nil, nil, err
		}
	}

	if in.IncludeConfig {
		res.Config = o.GetConfig()
	}

	return toonResult(res)
}

// buildAlertRows converts proto alerts into the flat table the detail page shows.
// exhaustion = window/factor, threshold = factor*(1-target).
func buildAlertRows(alerts []*objectivesv1alpha1.Alert, window time.Duration, target float64) []alertRow {
	budget := 1 - target
	rows := make([]alertRow, 0, len(alerts))
	for _, a := range alerts {
		exhaustion := time.Duration(0)
		if a.GetFactor() > 0 {
			exhaustion = time.Duration(float64(window) / a.GetFactor())
		}
		rows = append(rows, alertRow{
			State:        a.GetState().String(),
			Severity:     a.GetSeverity(),
			Exhaustion:   humanizeDuration(exhaustion),
			Threshold:    round(a.GetFactor()*budget, 4),
			ShortWindow:  humanizeDuration(a.GetShort().GetWindow().AsDuration()),
			ShortCurrent: current(a.GetShort().GetCurrent()),
			LongWindow:   humanizeDuration(a.GetLong().GetWindow().AsDuration()),
			LongCurrent:  current(a.GetLong().GetCurrent()),
			For:          humanizeDuration(a.GetFor().AsDuration()),
		})
	}
	return rows
}

// statusQueries reconstructs the instantaneous total/errors PromQL that
// GetStatus runs for this objective (with grouping merged in the same way), so
// the exact queries behind the status tiles travel with the result.
func (m *mcpServer) statusQueries(o *objectivesv1alpha1.Objective, grouping map[string]string) *querySet {
	obj := objectivesv1alpha1.ToInternal(o)

	for _, k := range slices.Sorted(maps.Keys(grouping)) {
		mt := &labels.Matcher{Type: labels.MatchEqual, Name: k, Value: grouping[k]}
		switch {
		case obj.Indicator.Ratio != nil:
			obj.Indicator.Ratio.Errors.LabelMatchers = append(obj.Indicator.Ratio.Errors.LabelMatchers, mt)
			obj.Indicator.Ratio.Total.LabelMatchers = append(obj.Indicator.Ratio.Total.LabelMatchers, mt)
		case obj.Indicator.Latency != nil:
			obj.Indicator.Latency.Success.LabelMatchers = append(obj.Indicator.Latency.Success.LabelMatchers, mt)
			obj.Indicator.Latency.Total.LabelMatchers = append(obj.Indicator.Latency.Total.LabelMatchers, mt)
		case obj.Indicator.BoolGauge != nil:
			obj.Indicator.BoolGauge.LabelMatchers = append(obj.Indicator.BoolGauge.LabelMatchers, mt)
		}
	}

	return &querySet{
		Total:  obj.QueryTotal(obj.Window, m.objectives.opts),
		Errors: obj.QueryErrors(obj.Window, m.objectives.opts),
	}
}

func (m *mcpServer) listOne(ctx context.Context, name string) (*objectivesv1alpha1.Objective, error) {
	resp, err := m.objectives.List(ctx, connect.NewRequest(&objectivesv1alpha1.ListRequest{Expr: selectorName(name)}))
	if err != nil {
		return nil, err
	}
	if len(resp.Msg.Objectives) == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("objective %q not found", name))
	}
	return resp.Msg.Objectives[0], nil
}

// graphReq carries the resolved request parameters into a graph fetch closure.
type graphReq struct {
	expr     string
	grouping string
	start    *timestamppb.Timestamp
	end      *timestamppb.Timestamp
}

// fetchGraph runs a Graph* RPC closure and builds a graphData. An empty result
// (NotFound — e.g. a healthy SLO with no errors) yields nil, so the field is
// omitted rather than surfaced as an error.
func (m *mcpServer) fetchGraph(r graphReq, single bool, fetch func(graphReq) ([]*objectivesv1alpha1.Timeseries, error)) (*graphData, error) {
	tss, err := fetch(r)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return nil, nil
		}
		return nil, err
	}
	g := &graphData{}
	if len(tss) > 0 {
		g.Query = tss[0].GetQuery()
	}
	if single {
		if len(tss) > 0 {
			applySingleSeries(g, tss[0])
		}
	} else {
		for _, ts := range tss {
			applyMultiSeries(g, ts)
		}
	}
	if g.Values == nil && g.Series == nil {
		return nil, nil
	}
	return g, nil
}

// defaultMaxPoints keeps the bundled graphs small by default; callers bump it
// (or pass -1 for full resolution = the server's native ~1000-point step).
const defaultMaxPoints = 20

// normalizeMaxPoints maps the input to a points target driving the query step:
// 0 (omitted) → default, negative → 0 (no override, server default resolution).
func normalizeMaxPoints(mp int) int {
	switch {
	case mp == 0:
		return defaultMaxPoints
	case mp < 0:
		return 0
	default:
		return mp
	}
}

// timeseriesGrid splits a Pyrra Timeseries into its timestamp axis and data
// series. matrixToValues packs timestamps (unix seconds) as Series[0]; the
// actual data is Series[1:], aligned with Labels[0:].
func timeseriesGrid(ts *objectivesv1alpha1.Timeseries) (timestamps []float64, data []*objectivesv1alpha1.Series, labels []string) {
	if ts == nil || len(ts.GetSeries()) == 0 {
		return nil, nil, nil
	}
	return ts.GetSeries()[0].GetValues(), ts.GetSeries()[1:], ts.GetLabels()
}

// grid returns the start (unix seconds) and per-point step from the timestamp axis.
func grid(timestamps []float64) (start, baseStep int64) {
	if len(timestamps) == 0 {
		return 0, 0
	}
	start = int64(timestamps[0])
	if len(timestamps) >= 2 {
		baseStep = int64(timestamps[1] - timestamps[0])
	}
	return start, baseStep
}

// Prometheus already returns the series at the requested step, so we pass the
// values through (rounded) — no client-side downsampling.
func applySingleSeries(res *graphData, ts *objectivesv1alpha1.Timeseries) {
	timestamps, data, _ := timeseriesGrid(ts)
	if len(data) == 0 {
		return
	}
	vals := data[0].GetValues()
	res.Start, res.Step = grid(timestamps)
	res.Current = lastFinite(vals)
	res.Values = roundAll(vals)
}

func applyMultiSeries(res *graphData, ts *objectivesv1alpha1.Timeseries) {
	timestamps, data, labels := timeseriesGrid(ts)
	if len(data) == 0 {
		return
	}
	res.Start, res.Step = grid(timestamps)
	for i, s := range data {
		vals := s.GetValues()
		label := ""
		if i < len(labels) {
			label = labels[i]
		}
		res.Series = append(res.Series, graphSeries{
			Labels: label, Current: lastFinite(vals), Values: roundAll(vals),
		})
	}
}

func toonResult(v any) (*mcpsdk.CallToolResult, any, error) {
	b, err := toon.Marshal(v, toon.WithLengthMarkers(true))
	if err != nil {
		return nil, nil, fmt.Errorf("encoding TOON: %w", err)
	}
	return &mcpsdk.CallToolResult{Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: string(b)}}}, nil, nil
}

// selectorFromLabels builds a Prometheus label matcher like {k="v",...}.
func selectorFromLabels(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%s=%q", k, v))
	}
	sort.Strings(parts)
	return "{" + strings.Join(parts, ",") + "}"
}

func selectorName(name string) string {
	return fmt.Sprintf("{%s=%q}", model.MetricNameLabel, name)
}

func stripName(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		if k == model.MetricNameLabel {
			continue
		}
		out[k] = v
	}
	return out
}

func mergeLabels(a, b map[string]string) map[string]string {
	out := make(map[string]string, len(a)+len(b))
	maps.Copy(out, a)
	maps.Copy(out, b)
	return out
}

// labelString renders a label map as a compact, sorted {k=v,...} string.
func labelString(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, k+"="+v)
	}
	sort.Strings(parts)
	return "{" + strings.Join(parts, ",") + "}"
}

func indicatorType(o *objectivesv1alpha1.Objective) string {
	ind := o.GetIndicator()
	switch {
	case ind.GetRatio() != nil:
		return "ratio"
	case ind.GetLatency() != nil:
		return "latency"
	case ind.GetLatencyNative() != nil:
		return "latency_native"
	case ind.GetBoolGauge() != nil:
		return "bool"
	}
	return ""
}

func latencyThreshold(o *objectivesv1alpha1.Objective) string {
	ind := o.GetIndicator()
	if l := ind.GetLatency(); l != nil {
		for _, mt := range l.GetSuccess().GetMatchers() {
			if mt.GetName() == "le" {
				if f, err := strconv.ParseFloat(mt.GetValue(), 64); err == nil && f > 0 {
					return humanizeDuration(time.Duration(f * float64(time.Second)))
				}
			}
		}
	}
	if ln := ind.GetLatencyNative(); ln != nil {
		return ln.GetLatency()
	}
	return ""
}

func humanizeDuration(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	return model.Duration(d).String()
}

func parseTimeArg(now time.Time, s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "now" {
		return now, nil
	}
	if s[0] == '-' || s[0] == '+' {
		d, err := model.ParseDuration(strings.TrimLeft(s, "+-"))
		if err != nil {
			return time.Time{}, err
		}
		if s[0] == '-' {
			return now.Add(-time.Duration(d)), nil
		}
		return now.Add(time.Duration(d)), nil
	}
	return time.Parse(time.RFC3339, s)
}

func round(f float64, decimals int) float64 {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0
	}
	p := math.Pow(10, float64(decimals))
	return math.Round(f*p) / p
}

// current normalizes a burn-rate current value; the API uses -1 to signal NaN.
func current(f float64) float64 {
	if f < 0 || math.IsNaN(f) {
		return 0
	}
	return round(f, 4)
}

// lastFinite returns the last non-NaN/Inf value of a series, rounded.
func lastFinite(vals []float64) float64 {
	cur := 0.0
	for _, v := range vals {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		cur = v
	}
	return round(cur, 4)
}

// roundAll rounds every value for token efficiency (NaN/Inf → 0).
func roundAll(vals []float64) []float64 {
	out := make([]float64, len(vals))
	for i, v := range vals {
		out[i] = round(v, 4)
	}
	return out
}

// alertSummary aggregates active alerts for one objective.
type alertSummary struct {
	firingN, pendingN    int
	hasCritical, hasWarn bool
}

func (s *alertSummary) add(a *objectivesv1alpha1.Alert) {
	if s == nil {
		return
	}
	switch a.GetState() {
	case objectivesv1alpha1.Alert_firing:
		s.firingN++
	case objectivesv1alpha1.Alert_pending:
		s.pendingN++
	}
	switch a.GetSeverity() {
	case "critical":
		s.hasCritical = true
	case "warning":
		s.hasWarn = true
	}
}

func (s *alertSummary) firing() int {
	if s == nil {
		return 0
	}
	return s.firingN
}

func (s *alertSummary) pending() int {
	if s == nil {
		return 0
	}
	return s.pendingN
}

func (s *alertSummary) worst() string {
	if s == nil {
		return ""
	}
	if s.hasCritical {
		return "critical"
	}
	if s.hasWarn {
		return "warning"
	}
	return ""
}
