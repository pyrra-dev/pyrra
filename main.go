package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/kong"
	"github.com/dgraph-io/ristretto"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	promconfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"

	"github.com/pyrra-dev/pyrra/openapi"
	openapiclient "github.com/pyrra-dev/pyrra/openapi/client"
	openapiserver "github.com/pyrra-dev/pyrra/openapi/server/go"
	"github.com/pyrra-dev/pyrra/slo"
)

//go:embed ui/build
var ui embed.FS

var CLI struct {
	API struct {
		PrometheusURL             *url.URL `default:"http://localhost:9090" help:"The URL to the Prometheus to query."`
		PrometheusExternalURL     *url.URL `help:"The URL for the UI to redirect users to when opening Prometheus. If empty the same as prometheus.url"`
		APIURL                    *url.URL `name:"api-url" default:"http://localhost:9444" help:"The URL to the API service like a Kubernetes Operator."`
		RoutePrefix               string   `default:"" help:"The route prefix Pyrra uses. If run behind a proxy you can change it to something like /pyrra here."`
		UIRoutePrefix             string   `default:"" help:"The route prefix Pyrra's UI uses. This is helpful for when the prefix is stripped by a proxy but still runs on /pyrra. Defaults to --route-prefix"`
		PrometheusBearerTokenPath string   `default:"" help:"Bearer token path"`
	} `cmd:"" help:"Runs Pyrra's API and UI."`
	Filesystem struct {
		ConfigFiles      string   `default:"/etc/pyrra/*.yaml" help:"The folder where Pyrra finds the config files to use."`
		PrometheusURL    *url.URL `default:"http://localhost:9090" help:"The URL to the Prometheus to query."`
		PrometheusFolder string   `default:"/etc/prometheus/pyrra/" help:"The folder where Pyrra writes the generates Prometheus rules and alerts."`
	} `cmd:"" help:"Runs Pyrra's filesystem operator and backend for the API."`
	Kubernetes struct {
		MetricsAddr   string `default:":8080" help:"The address the metric endpoint binds to."`
		ConfigMapMode bool   `default:"false" help:"If the generated recording rules should instead be saved to config maps in the default Prometheus format."`
	} `cmd:"" help:"Runs Pyrra's Kubernetes operator and backend for the API."`
}

func main() {
	ctx := kong.Parse(&CLI)

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.WithPrefix(logger, "caller", log.DefaultCaller)
	logger = log.WithPrefix(logger, "ts", log.DefaultTimestampUTC)

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewBuildInfoCollector(),
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	var prometheusURL *url.URL
	switch ctx.Command() {
	case "api":
		prometheusURL = CLI.API.PrometheusURL
	case "filesystem":
		prometheusURL = CLI.Filesystem.PrometheusURL
	}

	roundTripper, err := promconfig.NewRoundTripperFromConfig(promconfig.HTTPClientConfig{
		BearerTokenFile: CLI.API.PrometheusBearerTokenPath,
	}, "pyrra")
	if err != nil {
		level.Error(logger).Log("msg", "failed to create API client round tripper", "err", err)
		os.Exit(1)
	}

	client, err := api.NewClient(api.Config{
		Address:      prometheusURL.String(),
		RoundTripper: roundTripper,
	})
	if err != nil {
		level.Error(logger).Log("msg", "failed to create API client", "err", err)
		os.Exit(1)
	}
	// Wrap client to add extra headers for Thanos.
	client = newThanosClient(client)
	level.Info(logger).Log("msg", "using Prometheus", "url", prometheusURL.String())

	if CLI.API.PrometheusExternalURL == nil {
		CLI.API.PrometheusExternalURL = prometheusURL
	}

	var code int
	switch ctx.Command() {
	case "api":
		code = cmdAPI(logger, reg, client, CLI.API.PrometheusExternalURL, CLI.API.APIURL, CLI.API.RoutePrefix, CLI.API.UIRoutePrefix)
	case "filesystem":
		code = cmdFilesystem(logger, reg, client, CLI.Filesystem.ConfigFiles, CLI.Filesystem.PrometheusFolder)
	case "kubernetes":
		code = cmdKubernetes(logger, CLI.Kubernetes.MetricsAddr, CLI.Kubernetes.ConfigMapMode)
	}
	os.Exit(code)
}

func cmdAPI(logger log.Logger, reg *prometheus.Registry, promClient api.Client, prometheusExternal, apiURL *url.URL, routePrefix, uiRoutePrefix string) int {
	build, err := fs.Sub(ui, "ui/build")
	if err != nil {
		level.Error(logger).Log("msg", "failed to read UI build files", "err", err)
		return 1
	}

	// RoutePrefix must always be at least '/'.
	routePrefix = "/" + strings.Trim(routePrefix, "/")
	if uiRoutePrefix == "" {
		uiRoutePrefix = routePrefix
	} else {
		uiRoutePrefix = "/" + strings.Trim(uiRoutePrefix, "/")
	}

	level.Info(logger).Log("msg", "UI redirect to Prometheus", "url", prometheusExternal.String())
	level.Info(logger).Log("msg", "using API at", "url", apiURL.String())
	level.Info(logger).Log("msg", "using route prefix", "prefix", routePrefix)

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		level.Error(logger).Log("msg", "failed to create cache", "err", err)
		return 1
	}
	defer cache.Close()
	promAPI := &promCache{
		api:   prometheusv1.NewAPI(promClient),
		cache: cache,
	}

	apiConfig := openapiclient.NewConfiguration()
	apiConfig.Scheme = apiURL.Scheme
	apiConfig.Host = apiURL.Host
	apiClient := openapiclient.NewAPIClient(apiConfig)

	router := openapiserver.NewRouter(
		openapiserver.NewObjectivesApiController(&ObjectivesServer{
			logger:    logger,
			promAPI:   promAPI,
			apiclient: apiClient,
		}),
	)
	router.Use(openapi.MiddlewareMetrics(reg))
	router.Use(openapi.MiddlewareLogger(logger))

	tmpl, err := template.ParseFS(build, "index.html")
	if err != nil {
		level.Error(logger).Log("msg", "failed to parse HTML template", "err", err)
		return 1
	}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{})) // TODO: Disable by default

	r.Route(routePrefix, func(r chi.Router) {
		if routePrefix != "/" {
			r.Mount("/api/v1", http.StripPrefix(routePrefix, router))
		} else {
			r.Mount("/api/v1", router)
		}

		r.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
		r.Get("/objectives", func(w http.ResponseWriter, r *http.Request) {
			err := tmpl.Execute(w, struct {
				PrometheusURL string
				PathPrefix    string
				APIBasepath   string
			}{
				PrometheusURL: prometheusExternal.String(),
				PathPrefix:    uiRoutePrefix,
				APIBasepath:   path.Join(uiRoutePrefix, "/api/v1"),
			})
			if err != nil {
				level.Warn(logger).Log("msg", "failed to populate HTML template", "err", err)
			}
		})
		r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Trim trailing slash to not care about matching e.g. /pyrra and /pyrra/
			if r.URL.Path == "/" || strings.TrimSuffix(r.URL.Path, "/") == routePrefix {
				err := tmpl.Execute(w, struct {
					PrometheusURL string
					PathPrefix    string
					APIBasepath   string
				}{
					PrometheusURL: prometheusExternal.String(),
					PathPrefix:    uiRoutePrefix,
					APIBasepath:   path.Join(uiRoutePrefix, "/api/v1"),
				})
				if err != nil {
					level.Warn(logger).Log("msg", "failed to populate HTML template", "err", err)
				}
				return
			}

			http.StripPrefix(
				routePrefix,
				http.FileServer(http.FS(build)),
			).ServeHTTP(w, r)
		}))
	})

	if routePrefix != "/" {
		// Redirect /pyrra to /pyrra/ for the UI to work properly.
		r.HandleFunc(strings.TrimSuffix(routePrefix, "/"), func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, routePrefix+"/", http.StatusPermanentRedirect)
		})
	}

	if err := http.ListenAndServe(":9099", r); err != nil {
		level.Error(logger).Log("msg", "failed to run HTTP server", "err", err)
		return 2
	}
	return 0
}

func newThanosClient(client api.Client) api.Client {
	return &thanosClient{client: client}
}

// thanosClient wraps the Prometheus Client to inject some headers to disable partial responses
// and enables querying for downsampled data.
type thanosClient struct {
	client api.Client
}

func (c *thanosClient) URL(ep string, args map[string]string) *url.URL {
	return c.client.URL(ep, args)
}

func (c *thanosClient) Do(ctx context.Context, r *http.Request) (*http.Response, []byte, error) {
	if r.Body == nil {
		return c.client.Do(ctx, r)
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("reading body: %w", err)
	}
	query, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, nil, fmt.Errorf("parsing body: %w", err)
	}

	// We don't want partial responses, especially not when calculating error budgets.
	query.Set("partial_response", "false")
	r.ContentLength += 23

	if strings.HasSuffix(r.URL.Path, "/api/v1/query_range") {
		start, err := strconv.ParseFloat(query.Get("start"), 64)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing start: %w", err)
		}
		end, err := strconv.ParseFloat(query.Get("end"), 64)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing end: %w", err)
		}

		if end-start >= 28*24*60*60 { // request 1h downsamples when range > 28d
			query.Set("max_source_resolution", "1h")
			r.ContentLength += 25
		} else if end-start >= 7*24*60*60 { // request 5m downsamples when range > 1w
			query.Set("max_source_resolution", "5m")
			r.ContentLength += 25
		}
	}

	encoded := query.Encode()
	r.Body = ioutil.NopCloser(strings.NewReader(encoded))
	return c.client.Do(ctx, r)
}

type prometheusAPI interface {
	// Query performs a query for the given time.
	Query(ctx context.Context, query string, ts time.Time) (model.Value, prometheusv1.Warnings, error)
	// QueryRange performs a query for the given range.
	QueryRange(ctx context.Context, query string, r prometheusv1.Range) (model.Value, prometheusv1.Warnings, error)
}

func RoundUp(t time.Time, d time.Duration) time.Time {
	n := t.Round(d)
	if n.Before(t) {
		return n.Add(d)
	}
	return n
}

type promCache struct {
	api   prometheusAPI
	cache *ristretto.Cache
}

type promCacheKeyType string

const promCacheKey promCacheKeyType = "promCache"

func contextSetPromCache(ctx context.Context, t time.Duration) context.Context {
	return context.WithValue(ctx, promCacheKey, t)
}

func contextGetPromCache(ctx context.Context) time.Duration {
	t, ok := ctx.Value(promCacheKey).(time.Duration)
	if ok {
		return t
	}
	return 0
}

func (p *promCache) Query(ctx context.Context, query string, ts time.Time) (model.Value, prometheusv1.Warnings, error) {
	if value, exists := p.cache.Get(query); exists {
		return value.(model.Value), nil, nil
	}

	start := time.Now()
	value, warnings, err := p.api.Query(ctx, query, ts)
	duration := time.Since(start)
	if err != nil {
		return nil, warnings, fmt.Errorf("prometheus query: %w", err)
	}
	if len(warnings) > 0 {
		return value, warnings, nil
	}

	cacheDuration := contextGetPromCache(ctx)
	if cacheDuration > 0 {
		if v, ok := value.(model.Vector); ok {
			if len(v) > 0 {
				_ = p.cache.SetWithTTL(query, value, duration.Milliseconds(), cacheDuration)
			}
		}
	}

	return value, warnings, nil
}

func (p *promCache) QueryRange(ctx context.Context, query string, r prometheusv1.Range) (model.Value, prometheusv1.Warnings, error) {
	// Get the full time range of this query from start to end.
	// We round by 10s to adjust for small imperfections to increase cache hits.
	timeRange := r.End.Sub(r.Start).Round(10 * time.Second)
	cacheKey := fmt.Sprintf("%d;%s", timeRange.Milliseconds(), query)

	if value, exists := p.cache.Get(cacheKey); exists {
		return value.(model.Value), nil, nil
	}

	start := time.Now()
	value, warnings, err := p.api.QueryRange(ctx, query, r)
	duration := time.Since(start)
	if err != nil {
		return nil, warnings, fmt.Errorf("prometheus query range: %w", err)
	}
	if len(warnings) > 0 {
		return value, warnings, nil
	}

	cacheDuration := contextGetPromCache(ctx)
	if cacheDuration > 0 {
		if m, ok := value.(model.Matrix); ok {
			if len(m) > 0 {
				_ = p.cache.SetWithTTL(cacheKey, value, duration.Milliseconds(), cacheDuration)
			}
		}
	}

	return value, warnings, nil
}

type ObjectivesServer struct {
	logger    log.Logger
	promAPI   *promCache
	apiclient *openapiclient.APIClient
}

func (o *ObjectivesServer) ListObjectives(ctx context.Context, query string) (openapiserver.ImplResponse, error) {
	if query != "" {
		// We'll parse the query matchers already to make sure it's valid before passing on to the backend.
		if _, err := parser.ParseMetricSelector(query); err != nil {
			return openapiserver.ImplResponse{Code: http.StatusBadRequest}, err
		}
	}

	objectives, _, err := o.apiclient.ObjectivesApi.ListObjectives(ctx).Expr(query).Execute()
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	apiObjectives := make([]openapiserver.Objective, len(objectives))
	for i, objective := range objectives {
		apiObjectives[i] = openapi.ServerFromClient(objective)
	}

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: apiObjectives,
	}, nil
}

func (o *ObjectivesServer) GetObjectiveStatus(ctx context.Context, expr, grouping string) (openapiserver.ImplResponse, error) {
	clientObjectives, _, err := o.apiclient.ObjectivesApi.ListObjectives(ctx).Expr(expr).Execute()
	if err != nil {
		var apiErr openapiclient.GenericOpenAPIError
		if errors.As(err, &apiErr) {
			if strings.HasPrefix(apiErr.Error(), strconv.Itoa(http.StatusNotFound)) {
				return openapiserver.ImplResponse{Code: http.StatusNotFound}, apiErr
			}
		}
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	if len(clientObjectives) != 1 {
		return openapiserver.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("expr matches more than one SLO, it matches: %d", len(clientObjectives))
	}

	objective := openapi.InternalFromClient(clientObjectives[0])

	// Merge grouping into objective's query
	if grouping != "" {
		groupingMatchers, err := parser.ParseMetricSelector(grouping)
		if err != nil {
			return openapiserver.ImplResponse{}, err
		}
		if objective.Indicator.Ratio != nil {
			for _, m := range groupingMatchers {
				objective.Indicator.Ratio.Errors.LabelMatchers = append(objective.Indicator.Ratio.Errors.LabelMatchers, m)
				objective.Indicator.Ratio.Total.LabelMatchers = append(objective.Indicator.Ratio.Total.LabelMatchers, m)
			}
		}
		if objective.Indicator.Latency != nil {
			for _, m := range groupingMatchers {
				objective.Indicator.Latency.Success.LabelMatchers = append(objective.Indicator.Latency.Success.LabelMatchers, m)
				objective.Indicator.Latency.Total.LabelMatchers = append(objective.Indicator.Latency.Total.LabelMatchers, m)
			}
		}
	}

	ts := RoundUp(time.Now().UTC(), 5*time.Minute)

	queryTotal := objective.QueryTotal(objective.Window)
	level.Debug(o.logger).Log("msg", "sending query total", "query", queryTotal)
	value, _, err := o.promAPI.Query(contextSetPromCache(ctx, 15*time.Second), queryTotal, ts)
	if err != nil {
		level.Warn(o.logger).Log("msg", "failed to query total", "query", queryTotal, "err", err)
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	statuses := map[model.Fingerprint]*openapiserver.ObjectiveStatus{}

	for _, v := range value.(model.Vector) {
		labels := make(map[string]string)
		for k, v := range v.Metric {
			labels[string(k)] = string(v)
		}

		statuses[v.Metric.Fingerprint()] = &openapiserver.ObjectiveStatus{
			Labels: labels,
			Availability: openapiserver.ObjectiveStatusAvailability{
				Percentage: 1,
				Total:      float64(v.Value),
			},
		}
	}

	queryErrors := objective.QueryErrors(objective.Window)
	level.Debug(o.logger).Log("msg", "sending query errors", "query", queryErrors)
	value, _, err = o.promAPI.Query(contextSetPromCache(ctx, 15*time.Second), queryErrors, ts)
	if err != nil {
		level.Warn(o.logger).Log("msg", "failed to query errors", "query", queryErrors, "err", err)
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	for _, v := range value.(model.Vector) {
		s := statuses[v.Metric.Fingerprint()]
		s.Availability.Errors = float64(v.Value)
		s.Availability.Percentage = 1 - (s.Availability.Errors / s.Availability.Total)
	}

	statusSlice := make([]openapiserver.ObjectiveStatus, 0, len(statuses))

	for _, s := range statuses {
		s.Budget.Total = 1 - objective.Target
		s.Budget.Remaining = (s.Budget.Total - (s.Availability.Errors / s.Availability.Total)) / s.Budget.Total
		s.Budget.Max = s.Budget.Total * s.Availability.Total

		// If this objective has no requests, we'll skip showing it too
		if s.Availability.Total == 0 {
			continue
		}

		if math.IsNaN(s.Availability.Percentage) {
			s.Availability.Percentage = 1
		}
		if math.IsNaN(s.Budget.Remaining) {
			s.Budget.Remaining = 1
		}

		statusSlice = append(statusSlice, *s)
	}

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: statusSlice,
	}, nil
}

func (o *ObjectivesServer) GetObjectiveErrorBudget(ctx context.Context, expr, grouping string, startTimestamp, endTimestamp int32) (openapiserver.ImplResponse, error) {
	clientObjectives, _, err := o.apiclient.ObjectivesApi.ListObjectives(ctx).Expr(expr).Execute()
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	if len(clientObjectives) != 1 {
		return openapiserver.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("expr matches more than one SLO, it matches: %d", len(clientObjectives))
	}
	objective := openapi.InternalFromClient(clientObjectives[0])

	// Merge grouping into objective's query
	if grouping != "" {
		groupingMatchers, err := parser.ParseMetricSelector(grouping)
		if err != nil {
			return openapiserver.ImplResponse{}, err
		}
		if objective.Indicator.Ratio != nil {
			groupings := map[string]struct{}{}
			for _, g := range objective.Indicator.Ratio.Grouping {
				groupings[g] = struct{}{}
			}

			for _, m := range groupingMatchers {
				objective.Indicator.Ratio.Errors.LabelMatchers = append(objective.Indicator.Ratio.Errors.LabelMatchers, m)
				objective.Indicator.Ratio.Total.LabelMatchers = append(objective.Indicator.Ratio.Total.LabelMatchers, m)
				delete(groupings, m.Name)
			}

			objective.Indicator.Ratio.Grouping = []string{}
			for g := range groupings {
				objective.Indicator.Ratio.Grouping = append(objective.Indicator.Ratio.Grouping, g)
			}
		}
		if objective.Indicator.Latency != nil {
			groupings := map[string]struct{}{}
			for _, g := range objective.Indicator.Ratio.Grouping {
				groupings[g] = struct{}{}
			}

			for _, m := range groupingMatchers {
				objective.Indicator.Latency.Success.LabelMatchers = append(objective.Indicator.Latency.Success.LabelMatchers, m)
				objective.Indicator.Latency.Total.LabelMatchers = append(objective.Indicator.Latency.Total.LabelMatchers, m)
				delete(groupings, m.Name)
			}

			objective.Indicator.Ratio.Grouping = []string{}
			for g := range groupings {
				objective.Indicator.Ratio.Grouping = append(objective.Indicator.Ratio.Grouping, g)
			}
		}
	}

	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now

	if startTimestamp != 0 && endTimestamp != 0 {
		start = time.Unix(int64(startTimestamp), 0)
		end = time.Unix(int64(endTimestamp), 0)
	}

	step := end.Sub(start) / 1000

	query := objective.QueryErrorBudget()
	level.Debug(o.logger).Log("msg", "sending query error budget", "query", query, "step", step)
	value, _, err := o.promAPI.QueryRange(contextSetPromCache(ctx, 15*time.Second), query, prometheusv1.Range{
		Start: start,
		End:   end,
		Step:  step,
	})
	if err != nil {
		level.Warn(o.logger).Log("msg", "failed to query error budget", "query", query, "err", err)
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	matrix, ok := value.(model.Matrix)
	if !ok {
		err := fmt.Errorf("no matrix returned")
		level.Debug(o.logger).Log("msg", "returned data wasn't of type matrix", "query", query, "err", err)
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	if len(matrix) == 0 {
		level.Debug(o.logger).Log("msg", "returned no data", "query", query)
		return openapiserver.ImplResponse{Code: http.StatusNotFound, Body: struct{}{}}, nil
	}

	valueLength := 0
	for _, m := range matrix {
		if len(m.Values) > valueLength {
			valueLength = len(m.Values)
		}
	}

	values := matrixToValues(matrix)

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: openapiserver.QueryRange{
			Query:  query,
			Labels: nil,
			Values: values,
		},
	}, nil
}

const (
	alertstateInactive = "inactive"
	alertstatePending  = "pending"
	alertstateFiring   = "firing"
)

func (o *ObjectivesServer) GetMultiBurnrateAlerts(ctx context.Context, expr, grouping string, inactive bool) (openapiserver.ImplResponse, error) {
	clientObjectives, _, err := o.apiclient.ObjectivesApi.ListObjectives(ctx).Expr(expr).Execute()
	if err != nil {
		var apiErr openapiclient.GenericOpenAPIError
		if errors.As(err, &apiErr) {
			if strings.HasPrefix(apiErr.Error(), strconv.Itoa(http.StatusNotFound)) {
				return openapiserver.ImplResponse{Code: http.StatusNotFound}, apiErr
			}
		}
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	objectives := make([]slo.Objective, 0, len(clientObjectives))
	for _, o := range clientObjectives {
		objectives = append(objectives, openapi.InternalFromClient(o))
	}

	// Match alerts that at least have one character for the slo name.
	queryAlerts := `ALERTS{slo=~".+"}`

	if grouping != "" && grouping != "{}" {
		// If grouping exists we merge those matchers directly into the queryAlerts query.
		groupingMatchers, err := parser.ParseMetricSelector(grouping)
		if err != nil {
			return openapiserver.ImplResponse{}, fmt.Errorf("failed parsing grouping matchers: %w", err)
		}

		expr, err := parser.ParseExpr(queryAlerts)
		if err != nil {
			return openapiserver.ImplResponse{}, fmt.Errorf("failed parsing alerts metric: %w", err)
		}

		vec := expr.(*parser.VectorSelector)
		for _, m := range groupingMatchers {
			if m.Name == labels.MetricName || m.Name == "slo" { // adding some safeguards that shouldn't be allowed.
				continue
			}
			vec.LabelMatchers = append(vec.LabelMatchers, m)
		}

		queryAlerts = vec.String()
	}

	level.Debug(o.logger).Log("msg", "sending query for alerts", "query", queryAlerts)
	value, _, err := o.promAPI.Query(contextSetPromCache(ctx, 5*time.Second), queryAlerts, time.Now())
	if err != nil {
		level.Warn(o.logger).Log("msg", "failed to query alerts", "query", queryAlerts, "err", err)
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	vector, ok := value.(model.Vector)
	if !ok {
		err := fmt.Errorf("no vector returned")
		level.Debug(o.logger).Log("msg", "returned data wasn't of type vector", "query", queryAlerts, "err", err)
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	alerts := alertsMatchingObjectives(vector, objectives, inactive)

	if true {
		for _, objective := range objectives {
			mtx := &sync.Mutex{}
			windowsMap := map[time.Duration]float64{}
			for _, w := range objective.Windows() {
				windowsMap[w.Short] = 0
				windowsMap[w.Long] = 0
			}

			// TODO: Make concurrent with worker pool
			for w := range windowsMap {
				// TODO: Returns incorrect burnrate query for latency (groups by code)
				value, _, err := o.promAPI.Query(ctx, objective.Burnrate(w), time.Now())
				if err != nil {
					return openapiserver.ImplResponse{}, err
				}
				vec, ok := value.(model.Vector)
				if !ok {
					return openapiserver.ImplResponse{}, fmt.Errorf("expected vector value from Prometheus")
				}
				if vec.Len() != 1 {
					return openapiserver.ImplResponse{}, fmt.Errorf("expected vector with one value from Prometheus")
				}
				mtx.Lock()
				windowsMap[w] = float64(vec[0].Value)
				mtx.Unlock()
			}

			// Match objectives to alerts to update response
		Alerts:
			for i, alert := range alerts {
				for k, v := range alert.Labels {
					if objective.Labels.Get(k) != v {
						continue Alerts
					}
				}
				short := time.Duration(alert.Short.Window) * time.Millisecond
				long := time.Duration(alert.Long.Window) * time.Millisecond
				alerts[i].Short.Current = windowsMap[short]
				alerts[i].Long.Current = windowsMap[long]
			}
		}
	}

	return openapiserver.ImplResponse{Code: http.StatusOK, Body: alerts}, nil
}

// alertsMatchingObjectives loops through all alerts trying to match objectives based on their labels.
// All labels of an objective need to be equal if they exist on the ALERTS metric.
// Therefore, only a subset on labels are taken into account
// which gives the ALERTS metric the opportunity to include more custom labels.
func alertsMatchingObjectives(metrics model.Vector, objectives []slo.Objective, inactive bool) []openapiserver.MultiBurnrateAlert {
	alerts := make([]openapiserver.MultiBurnrateAlert, 0, len(metrics))

	if inactive {
		for _, o := range objectives {
			if len(o.Labels) == 0 {
				continue
			}
			lset := map[string]string{}
			for _, l := range o.Labels {
				lset[l.Name] = l.Value
			}
			for _, w := range o.Windows() {
				alerts = append(alerts, openapiserver.MultiBurnrateAlert{
					Labels:   lset,
					Severity: string(w.Severity),
					For:      w.For.Milliseconds(),
					Factor:   w.Factor,
					Short: openapiserver.Burnrate{
						Window:  w.Short.Milliseconds(),
						Current: -1,
						Query:   o.Burnrate(w.Short),
					},
					Long: openapiserver.Burnrate{
						Window:  w.Long.Milliseconds(),
						Current: -1,
						Query:   o.Burnrate(w.Long),
					},
					State: alertstateInactive,
				})
			}
		}
	}

	// For each alert iterate over all given objectives to find matching ones to return.
	// If this gets out of hand as it's O(alerts*objectives) we should probably use hashing to find matches instead.
	for _, sample := range metrics {
	Objectives:
		for _, o := range objectives {
			if len(o.Labels) == 0 {
				continue
			}

			lset := map[string]string{}
			for _, l := range o.Labels {
				// check if each label of the objective is present in the alert.
				// If it's present make sure the values match
				name := l.Name
				if name == labels.MetricName {
					// The __name__ is called slo in the ALERTS metrics.
					name = "slo"
				}
				value, found := sample.Metric[model.LabelName(name)]
				if found {
					if string(value) != l.Value {
						continue Objectives
					}
					lset[l.Name] = l.Value
				}
			}

			short, err := model.ParseDuration(string(sample.Metric[model.LabelName("short")]))
			if err != nil {
				// TODO: Return the error and not just skip?
				continue
			}
			long, err := model.ParseDuration(string(sample.Metric[model.LabelName("long")]))
			if err != nil {
				// TODO: Return the error and not just skip?
				continue
			}

			window, found := o.HasWindows(short, long)
			if !found {
				// It could be that the labels match, however the long and short burn rate windows don't.
				// If that's the case we can't say for sure it's the same objective since it window may be different.
				continue
			}

			// Add potentially missing labels from objective to alerts' labelset
			for _, l := range o.Labels {
				lset[l.Name] = l.Value
			}

			// Add potentially missing labels from metric to alerts' labelset.
			// Excluding a couple ones that are part of the struct itself.
			// This is mostly important for listing the same objectives with grouping by labels.
			for n, v := range sample.Metric {
				name := string(n)
				value := string(v)
				if name == "alertname" ||
					name == "alertstate" ||
					name == "long" ||
					name == "severity" ||
					name == "short" ||
					name == "slo" ||
					name == labels.MetricName {
					continue
				}
				lset[name] = value
			}

			if !inactive { // If we don't include inactive we can simply append
				alerts = append(alerts, openapiserver.MultiBurnrateAlert{
					Labels:   lset,
					Severity: string(sample.Metric["severity"]),
					State:    string(sample.Metric["alertstate"]),
					For:      window.For.Milliseconds(),
					Factor:   window.Factor,
					Short: openapiserver.Burnrate{
						Window:  time.Duration(short).Milliseconds(),
						Current: -1,
						Query:   o.Burnrate(time.Duration(short)),
					},
					Long: openapiserver.Burnrate{
						Window:  time.Duration(long).Milliseconds(),
						Current: -1,
						Query:   o.Burnrate(time.Duration(long)),
					},
				})
			} else {
			alerts:
				for i, alert := range alerts {
					for n, v := range alert.Labels {
						if lset[n] != v {
							continue alerts
						}
					}
					// only continues here if all labels match
					if alert.Short.Window != time.Duration(short).Milliseconds() {
						continue alerts
					}
					if alert.Long.Window != time.Duration(long).Milliseconds() {
						continue alerts
					}
					// only update state for exact short and long burn rate match
					alerts[i].State = string(sample.Metric["alertstate"])
				}
			}
		}
	}

	return alerts
}

func (o *ObjectivesServer) GetREDRequests(ctx context.Context, expr, grouping string, startTimestamp, endTimestamp int32) (openapiserver.ImplResponse, error) {
	clientObjectives, _, err := o.apiclient.ObjectivesApi.ListObjectives(ctx).Expr(expr).Execute()
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	if len(clientObjectives) != 1 {
		return openapiserver.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("expr matches not exactly one SLO")
	}
	objective := openapi.InternalFromClient(clientObjectives[0])

	// Merge grouping into objective's query
	if grouping != "" {
		groupingMatchers, err := parser.ParseMetricSelector(grouping)
		if err != nil {
			return openapiserver.ImplResponse{}, err
		}
		if objective.Indicator.Ratio != nil {
			for _, m := range groupingMatchers {
				objective.Indicator.Ratio.Errors.LabelMatchers = append(objective.Indicator.Ratio.Errors.LabelMatchers, m)
				objective.Indicator.Ratio.Total.LabelMatchers = append(objective.Indicator.Ratio.Total.LabelMatchers, m)
			}
		}
		if objective.Indicator.Latency != nil {
			for _, m := range groupingMatchers {
				objective.Indicator.Latency.Success.LabelMatchers = append(objective.Indicator.Latency.Success.LabelMatchers, m)
				objective.Indicator.Latency.Total.LabelMatchers = append(objective.Indicator.Latency.Total.LabelMatchers, m)
			}
		}
	}

	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now

	if startTimestamp != 0 && endTimestamp != 0 {
		start = time.Unix(int64(startTimestamp), 0)
		end = time.Unix(int64(endTimestamp), 0)
	}
	step := end.Sub(start) / 1000

	timeRange := rangeInterval(start, end)
	cacheDuration := rangeCache(start, end)

	query := objective.RequestRange(timeRange)
	level.Debug(o.logger).Log("msg", "running range request", "query", query)

	value, _, err := o.promAPI.QueryRange(contextSetPromCache(ctx, cacheDuration), query, prometheusv1.Range{
		Start: start,
		End:   end,
		Step:  step,
	})
	if err != nil {
		level.Warn(o.logger).Log("msg", "failed to run range request", "query", query, "err", err)
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	if value.Type() != model.ValMatrix {
		err := fmt.Errorf("returned data is not a matrix")
		level.Warn(o.logger).Log("query", query, "err", err)
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	matrix, ok := value.(model.Matrix)
	if !ok {
		err := fmt.Errorf("no matrix returned")
		level.Warn(o.logger).Log("query", query, "err", err)
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	if len(matrix) == 0 {
		level.Debug(o.logger).Log("msg", "no data returned", "query", query)
		return openapiserver.ImplResponse{Code: http.StatusNotFound, Body: struct{}{}}, nil
	}

	valueLength := 0
	for _, m := range matrix {
		if len(m.Values) > valueLength {
			valueLength = len(m.Values)
		}
	}

	labels := make([]string, len(matrix))
	for i, stream := range matrix {
		labels[i] = model.LabelSet(stream.Metric).String()
	}

	values := matrixToValues(matrix)

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: openapiserver.QueryRange{
			Query:  query,
			Labels: labels,
			Values: values,
		},
	}, nil
}

func (o *ObjectivesServer) GetREDErrors(ctx context.Context, expr, grouping string, startTimestamp, endTimestamp int32) (openapiserver.ImplResponse, error) {
	clientObjectives, _, err := o.apiclient.ObjectivesApi.ListObjectives(ctx).Expr(expr).Execute()
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	if len(clientObjectives) != 1 {
		return openapiserver.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("expr matches not exactly one SLO")
	}
	objective := openapi.InternalFromClient(clientObjectives[0])

	// Merge grouping into objective's query
	if grouping != "" {
		groupingMatchers, err := parser.ParseMetricSelector(grouping)
		if err != nil {
			return openapiserver.ImplResponse{}, err
		}
		if objective.Indicator.Ratio != nil {
			for _, m := range groupingMatchers {
				objective.Indicator.Ratio.Errors.LabelMatchers = append(objective.Indicator.Ratio.Errors.LabelMatchers, m)
				objective.Indicator.Ratio.Total.LabelMatchers = append(objective.Indicator.Ratio.Total.LabelMatchers, m)
			}
		}
		if objective.Indicator.Latency != nil {
			for _, m := range groupingMatchers {
				objective.Indicator.Latency.Success.LabelMatchers = append(objective.Indicator.Latency.Success.LabelMatchers, m)
				objective.Indicator.Latency.Total.LabelMatchers = append(objective.Indicator.Latency.Total.LabelMatchers, m)
			}
		}
	}

	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now

	if startTimestamp != 0 && endTimestamp != 0 {
		start = time.Unix(int64(startTimestamp), 0)
		end = time.Unix(int64(endTimestamp), 0)
	}
	step := end.Sub(start) / 1000

	timeRange := rangeInterval(start, end)
	cacheDuration := rangeCache(start, end)

	query := objective.ErrorsRange(timeRange)
	level.Debug(o.logger).Log("msg", "running error request", "query", query)

	value, _, err := o.promAPI.QueryRange(contextSetPromCache(ctx, cacheDuration), query, prometheusv1.Range{
		Start: start,
		End:   end,
		Step:  step,
	})
	if err != nil {
		level.Warn(o.logger).Log("msg", "failed to run range error request", "query", query, "err", err)
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	if value.Type() != model.ValMatrix {
		err := fmt.Errorf("returned data is not a matrix")
		level.Warn(o.logger).Log("query", query, "err", err)
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	matrix, ok := value.(model.Matrix)
	if !ok {
		err := fmt.Errorf("no matrix returned")
		level.Warn(o.logger).Log("query", query, "err", err)
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	if len(matrix) == 0 {
		level.Debug(o.logger).Log("msg", "no data returned", "query", query)
		return openapiserver.ImplResponse{Code: http.StatusNotFound, Body: struct{}{}}, nil
	}

	valueLength := 0
	for _, m := range matrix {
		if len(m.Values) > valueLength {
			valueLength = len(m.Values)
		}
	}

	labels := make([]string, len(matrix))
	for i, stream := range matrix {
		labels[i] = model.LabelSet(stream.Metric).String()
	}

	values := matrixToValues(matrix)

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: openapiserver.QueryRange{
			Query:  query,
			Labels: labels,
			Values: values,
		},
	}, nil
}

const (
	hours12 = 12 * time.Hour
	day     = 24 * time.Hour
	week    = 7 * day
	month   = 4 * week
)

func rangeInterval(start, end time.Time) time.Duration {
	diff := end.Sub(start)
	d := 5 * time.Minute
	// TODO: Refactor for early returns instead
	if diff >= month {
		d = 3 * time.Hour
	} else if diff >= week {
		d = time.Hour
	} else if diff >= day {
		d = 30 * time.Minute
	} else if diff >= hours12 {
		d = 15 * time.Minute
	}
	return d
}

func rangeCache(start, end time.Time) time.Duration {
	diff := end.Sub(start)
	d := 15 * time.Second
	// TODO: Refactor for early returns instead
	if diff >= month {
		d = 5 * time.Minute
	} else if diff >= week {
		d = 3 * time.Minute
	} else if diff >= day {
		d = 90 * time.Second
	} else if diff >= hours12 {
		d = 45 * time.Second
	}
	return d
}

func matrixToValues(m model.Matrix) [][]float64 {
	series := len(m)
	if series == 0 {
		return nil
	}

	if series == 1 {
		vs := make([][]float64, len(m)+1) // +1 for timestamps
		for i, stream := range m {
			vs[0] = make([]float64, len(stream.Values))
			vs[i+1] = make([]float64, len(stream.Values))

			for j, pair := range stream.Values {
				vs[0][j] = float64(pair.Timestamp / 1000)
				if !math.IsNaN(float64(pair.Value)) {
					vs[i+1][j] = float64(pair.Value)
				}
			}
		}
		return vs
	}

	pairs := make(map[int64][]float64, len(m[0].Values))
	for i, stream := range m {
		for _, pair := range stream.Values {
			t := int64(pair.Timestamp / 1000)
			if _, ok := pairs[t]; !ok {
				pairs[t] = make([]float64, series)
			}
			if !math.IsNaN(float64(pair.Value)) {
				pairs[t][i] = float64(pair.Value)
			}
		}
	}

	vs := make(values, series+1)
	for i := 0; i < series+1; i++ {
		vs[i] = make([]float64, len(pairs))
	}
	var i int
	for t, fs := range pairs {
		vs[0][i] = float64(t)
		for j, f := range fs {
			vs[j+1][i] = f
		}
		i++
	}

	sort.Sort(vs)

	return vs
}

type values [][]float64

func (v values) Len() int {
	return len(v[0])
}

func (v values) Less(i, j int) bool {
	return v[0][i] < v[0][j]
}

// Swap iterates over all []float64 and consistently swaps them.
func (v values) Swap(i, j int) {
	for n := range v {
		v[n][i], v[n][j] = v[n][j], v[n][i]
	}
}
