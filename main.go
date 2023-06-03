package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	"github.com/bufbuild/connect-go"
	"github.com/dgraph-io/ristretto"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
	connectprometheus "github.com/polarsignals/connect-go-prometheus"
	"github.com/prometheus/client_golang/api"
	prometheusapiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	promconfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/types/known/durationpb"

	objectivesv1alpha1 "github.com/pyrra-dev/pyrra/proto/objectives/v1alpha1"
	"github.com/pyrra-dev/pyrra/proto/objectives/v1alpha1/objectivesv1alpha1connect"
	"github.com/pyrra-dev/pyrra/proto/prometheus/v1/prometheusv1connect"
	"github.com/pyrra-dev/pyrra/slo"
)

//go:embed ui/build
var ui embed.FS

var CLI struct {
	API struct {
		PrometheusURL               *url.URL          `default:"http://localhost:9090" help:"The URL to the Prometheus to query."`
		PrometheusExternalURL       *url.URL          `help:"The URL for the UI to redirect users to when opening Prometheus. If empty the same as prometheus.url"`
		APIURL                      *url.URL          `name:"api-url" default:"http://localhost:9444" help:"The URL to the API service like a Kubernetes Operator."`
		RoutePrefix                 string            `default:"" help:"The route prefix Pyrra uses. If run behind a proxy you can change it to something like /pyrra here."`
		UIRoutePrefix               string            `default:"" help:"The route prefix Pyrra's UI uses. This is helpful for when the prefix is stripped by a proxy but still runs on /pyrra. Defaults to --route-prefix"`
		PrometheusBearerTokenPath   string            `default:"" help:"Bearer token path"`
		PrometheusBasicAuthUsername string            `default:"" help:"The HTTP basic authentication username"`
		PrometheusBasicAuthPassword promconfig.Secret `default:"" help:"The HTTP basic authentication password"`
	} `cmd:"" help:"Runs Pyrra's API and UI."`
	Filesystem struct {
		ConfigFiles      string   `default:"/etc/pyrra/*.yaml" help:"The folder where Pyrra finds the config files to use. Any non yaml files will be ignored."`
		PrometheusURL    *url.URL `default:"http://localhost:9090" help:"The URL to the Prometheus to query."`
		PrometheusFolder string   `default:"/etc/prometheus/pyrra/" help:"The folder where Pyrra writes the generates Prometheus rules and alerts."`
		GenericRules     bool     `default:"false" help:"Enabled generic recording rules generation to make it easier for tools like Grafana."`
	} `cmd:"" help:"Runs Pyrra's filesystem operator and backend for the API."`
	Kubernetes struct {
		MetricsAddr   string `default:":8080" help:"The address the metric endpoint binds to."`
		ConfigMapMode bool   `default:"false" help:"If the generated recording rules should instead be saved to config maps in the default Prometheus format."`
		GenericRules  bool   `default:"false" help:"Enabled generic recording rules generation to make it easier for tools like Grafana."`
	} `cmd:"" help:"Runs Pyrra's Kubernetes operator and backend for the API."`
	Generate struct {
		ConfigFiles      string `default:"/etc/pyrra/*.yaml" help:"The folder where Pyrra finds the config files to use."`
		PrometheusFolder string `default:"/etc/prometheus/pyrra/" help:"The folder where Pyrra writes the generated Prometheus rules and alerts."`
		GenericRules     bool   `default:"false" help:"Enabled generic recording rules generation to make it easier for tools like Grafana."`
		OperatorRule     bool   `default:"false" help:"Generate rule files as prometheus-operator PrometheusRule: https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.PrometheusRule."`
	} `cmd:"" help:"Read SLO config files and rewrites them as Prometheus rules and alerts."`
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
	default:
		prometheusURL, _ = url.Parse("http://localhost:9090")
	}

	roundTripper, err := promconfig.NewRoundTripperFromConfig(promconfig.HTTPClientConfig{
		BasicAuth: &promconfig.BasicAuth{
			Username: CLI.API.PrometheusBasicAuthUsername,
			Password: CLI.API.PrometheusBasicAuthPassword,
		},
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
		code = cmdAPI(
			logger,
			reg,
			client,
			CLI.API.PrometheusExternalURL,
			CLI.API.APIURL,
			CLI.API.RoutePrefix,
			CLI.API.UIRoutePrefix,
		)
	case "filesystem":
		code = cmdFilesystem(
			logger,
			reg,
			client,
			CLI.Filesystem.ConfigFiles,
			CLI.Filesystem.PrometheusFolder,
			CLI.Filesystem.GenericRules,
		)
	case "kubernetes":
		code = cmdKubernetes(
			logger,
			CLI.Kubernetes.MetricsAddr,
			CLI.Kubernetes.ConfigMapMode,
			CLI.Kubernetes.GenericRules,
		)
	case "generate":
		code = cmdGenerate(
			logger,
			CLI.Generate.ConfigFiles,
			CLI.Generate.PrometheusFolder,
			CLI.Generate.GenericRules,
			CLI.Generate.OperatorRule,
		)
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
		api: &promLogger{
			api:    prometheusapiv1.NewAPI(promClient),
			logger: logger,
		},
		cache: cache,
	}

	tmpl, err := template.ParseFS(build, "index.html")
	if err != nil {
		level.Error(logger).Log("msg", "failed to parse HTML template", "err", err)
		return 1
	}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedHeaders: []string{
			"Content-Type",
			"Connect-Protocol-Version",
		},
	})) // TODO: Disable by default

	prometheusInterceptor := connectprometheus.NewInterceptor(reg)

	r.Route(routePrefix, func(r chi.Router) {
		objectiveService := &objectiveServer{
			logger:  log.WithPrefix(logger, "service", "objective"),
			promAPI: promAPI,
			client: newBackendClientCache(
				objectivesv1alpha1connect.NewObjectiveBackendServiceClient(
					http.DefaultClient,
					apiURL.String(),
					connect.WithInterceptors(prometheusInterceptor),
				),
			),
		}

		objectivePath, objectiveHandler := objectivesv1alpha1connect.NewObjectiveServiceHandler(
			objectiveService,
			connect.WithInterceptors(prometheusInterceptor),
		)

		prometheusService := &prometheusServer{
			logger:  log.WithPrefix(logger, "service", "prometheus"),
			promAPI: promAPI,
		}
		prometheusPath, prometheusHandler := prometheusv1connect.NewPrometheusServiceHandler(prometheusService)

		if routePrefix != "/" {
			r.Mount(objectivePath, http.StripPrefix(routePrefix, objectiveHandler))
			r.Mount(prometheusPath, http.StripPrefix(routePrefix, prometheusHandler))
		} else {
			r.Mount(objectivePath, objectiveHandler)
			r.Mount(prometheusPath, prometheusHandler)
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
				APIBasepath:   uiRoutePrefix,
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
					APIBasepath:   uiRoutePrefix,
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

	var (
		gr  run.Group
		ctx = context.Background()
	)
	gr.Add(run.SignalHandler(ctx, os.Interrupt, syscall.SIGTERM))

	{
		httpServer := &http.Server{
			Addr:    ":9099",
			Handler: h2c.NewHandler(r, &http2.Server{}),
		}
		gr.Add(
			func() error {
				return httpServer.ListenAndServe()
			},
			func(error) {
				shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()
				_ = httpServer.Shutdown(shutdownCtx)
			},
		)
	}
	if err := gr.Run(); err != nil {
		if _, ok := err.(run.SignalError); ok {
			level.Info(logger).Log("msg", "terminated HTTP server", "reason", err)
			return 0
		}
		level.Error(logger).Log("msg", "failed to run HTTP server", "err", err)
		return 2
	}
	return 0
}

func newBackendClientCache(client objectivesv1alpha1connect.ObjectiveBackendServiceClient) objectivesv1alpha1connect.ObjectiveBackendServiceClient {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 100,
		MaxCost:     10 * 1000, // 10 seconds
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	return &backendClientCache{client: client, cache: cache}
}

type backendClientCache struct {
	client objectivesv1alpha1connect.ObjectiveBackendServiceClient
	cache  *ristretto.Cache
}

// List calls the backend service and caches the result for 10 seconds if the request is successful.
func (b *backendClientCache) List(ctx context.Context, req *connect.Request[objectivesv1alpha1.ListRequest]) (*connect.Response[objectivesv1alpha1.ListResponse], error) {
	key := req.Msg.Expr + req.Msg.Grouping

	list, found := b.cache.Get(key)
	if found {
		return connect.NewResponse(list.(*objectivesv1alpha1.ListResponse)), nil
	}

	start := time.Now()
	resp, err := b.client.List(ctx, req)
	if err != nil {
		return nil, err
	}

	_ = b.cache.SetWithTTL(key, resp.Msg, time.Since(start).Milliseconds(), 10*time.Second)

	return resp, nil
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
	body, err := io.ReadAll(r.Body)
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
	r.Body = io.NopCloser(strings.NewReader(encoded))
	return c.client.Do(ctx, r)
}

type prometheusAPI interface {
	// Query performs a query for the given time.
	Query(ctx context.Context, query string, ts time.Time, opts ...prometheusapiv1.Option) (model.Value, prometheusapiv1.Warnings, error)
	// QueryRange performs a query for the given range.
	QueryRange(ctx context.Context, query string, r prometheusapiv1.Range, opts ...prometheusapiv1.Option) (model.Value, prometheusapiv1.Warnings, error)
}

type promLogger struct {
	api    prometheusAPI
	logger log.Logger
}

func (l *promLogger) Query(ctx context.Context, query string, ts time.Time, opts ...prometheusapiv1.Option) (model.Value, prometheusapiv1.Warnings, error) {
	level.Debug(l.logger).Log(
		"msg", "running instant query",
		"query", query,
		"ts", ts,
	)
	return l.api.Query(ctx, query, ts, opts...)
}

func (l *promLogger) QueryRange(ctx context.Context, query string, r prometheusapiv1.Range, opts ...prometheusapiv1.Option) (model.Value, prometheusapiv1.Warnings, error) {
	level.Debug(l.logger).Log(
		"msg", "running range query",
		"query", query,
		"start", r.Start,
		"end", r.End,
	)
	return l.api.QueryRange(ctx, query, r, opts...)
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

func (p *promCache) Query(ctx context.Context, query string, ts time.Time) (model.Value, prometheusapiv1.Warnings, error) {
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

func (p *promCache) QueryRange(ctx context.Context, query string, r prometheusapiv1.Range) (model.Value, prometheusapiv1.Warnings, error) {
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

type objectiveServer struct {
	logger  log.Logger
	promAPI *promCache
	client  objectivesv1alpha1connect.ObjectiveBackendServiceClient
}

func (s *objectiveServer) getObjective(ctx context.Context, expr string) (slo.Objective, error) {
	resp, err := s.client.List(ctx, connect.NewRequest(&objectivesv1alpha1.ListRequest{
		Expr: expr,
	}))
	if err != nil {
		return slo.Objective{}, err
	}

	if len(resp.Msg.Objectives) != 1 {
		return slo.Objective{}, connect.NewError(connect.CodeAborted, fmt.Errorf("expr matches more than one SLO, it matches: %d", len(resp.Msg.Objectives)))
	}

	return objectivesv1alpha1.ToInternal(resp.Msg.Objectives[0]), nil
}

func (s *objectiveServer) List(ctx context.Context, req *connect.Request[objectivesv1alpha1.ListRequest]) (*connect.Response[objectivesv1alpha1.ListResponse], error) {
	if expr := req.Msg.Expr; expr != "" {
		if _, err := parser.ParseMetricSelector(expr); err != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("failed to parse expr: %w", err))
		}
	}

	resp, err := s.client.List(ctx, connect.NewRequest(&objectivesv1alpha1.ListRequest{
		Expr: req.Msg.Expr,
	}))
	if err != nil {
		return nil, err
	}

	groupingMatchers := map[string]*labels.Matcher{}
	if req.Msg.Grouping != "" {
		ms, err := parser.ParseMetricSelector(req.Msg.Grouping)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		for _, m := range ms {
			groupingMatchers[m.Name] = m
		}
	}

	for _, o := range resp.Msg.Objectives {
		oi := objectivesv1alpha1.ToInternal(o)

		// If specific grouping was selected we need to merge the label matchers for the queries.
		if len(groupingMatchers) > 0 {
			switch oi.IndicatorType() {
			case slo.Ratio:
				groupingMatchersErrors := make(map[string]*labels.Matcher, len(groupingMatchers))
				groupingMatchersTotal := make(map[string]*labels.Matcher, len(groupingMatchers))
				for _, matcher := range groupingMatchers {
					// We need to copy the matchers to avoid modifying the original later on.
					groupingMatchersErrors[matcher.Name] = &labels.Matcher{Type: matcher.Type, Name: matcher.Name, Value: matcher.Value}
					groupingMatchersTotal[matcher.Name] = &labels.Matcher{Type: matcher.Type, Name: matcher.Name, Value: matcher.Value}
				}

				for _, m := range oi.Indicator.Ratio.Errors.LabelMatchers {
					if rm, replace := groupingMatchers[m.Name]; replace {
						m.Type = rm.Type
						m.Value = rm.Value
						delete(groupingMatchersErrors, m.Name)
					}
				}
				for _, m := range oi.Indicator.Ratio.Total.LabelMatchers {
					if rm, replace := groupingMatchers[m.Name]; replace {
						m.Type = rm.Type
						m.Value = rm.Value
						delete(groupingMatchersTotal, m.Name)
					}
				}

				// Now add the remaining matchers we didn't find before.
				for _, m := range groupingMatchersErrors {
					oi.Indicator.Ratio.Errors.LabelMatchers = append(oi.Indicator.Ratio.Errors.LabelMatchers, m)
				}
				for _, m := range groupingMatchersTotal {
					oi.Indicator.Ratio.Total.LabelMatchers = append(oi.Indicator.Ratio.Total.LabelMatchers, m)
				}
			case slo.Latency:
				groupingMatchersSuccess := make(map[string]*labels.Matcher, len(groupingMatchers))
				groupingMatchersTotal := make(map[string]*labels.Matcher, len(groupingMatchers))
				for _, matcher := range groupingMatchers {
					// We need to copy the matchers to avoid modifying the original later on.
					groupingMatchersSuccess[matcher.Name] = &labels.Matcher{Type: matcher.Type, Name: matcher.Name, Value: matcher.Value}
					groupingMatchersTotal[matcher.Name] = &labels.Matcher{Type: matcher.Type, Name: matcher.Name, Value: matcher.Value}
				}

				for _, m := range oi.Indicator.Latency.Success.LabelMatchers {
					if rm, replace := groupingMatchers[m.Name]; replace {
						m.Type = rm.Type
						m.Value = rm.Value
						delete(groupingMatchersSuccess, m.Name)
					}
				}
				for _, m := range oi.Indicator.Latency.Total.LabelMatchers {
					if rm, replace := groupingMatchers[m.Name]; replace {
						m.Type = rm.Type
						m.Value = rm.Value
						delete(groupingMatchersTotal, m.Name)
					}
				}

				// Now add the remaining matchers we didn't find before.
				for _, m := range groupingMatchersSuccess {
					oi.Indicator.Latency.Success.LabelMatchers = append(oi.Indicator.Latency.Success.LabelMatchers, m)
				}
				for _, m := range groupingMatchersTotal {
					oi.Indicator.Latency.Total.LabelMatchers = append(oi.Indicator.Latency.Total.LabelMatchers, m)
				}
			case slo.LatencyNative:
				for _, m := range oi.Indicator.LatencyNative.Total.LabelMatchers {
					if rm, replace := groupingMatchers[m.Name]; replace {
						m.Type = rm.Type
						m.Value = rm.Value
						delete(groupingMatchers, m.Name)
					}
				}
				for _, m := range groupingMatchers {
					oi.Indicator.LatencyNative.Total.LabelMatchers = append(oi.Indicator.LatencyNative.Total.LabelMatchers, m)
				}
			case slo.BoolGauge:
				for _, m := range oi.Indicator.BoolGauge.LabelMatchers {
					if rm, replace := groupingMatchers[m.Name]; replace {
						m.Type = rm.Type
						m.Value = rm.Value
						delete(groupingMatchers, m.Name)
					}
				}
				for _, m := range groupingMatchers {
					oi.Indicator.BoolGauge.LabelMatchers = append(oi.Indicator.BoolGauge.LabelMatchers, m)
				}
			}
		}

		o.Queries = &objectivesv1alpha1.Queries{
			CountTotal:       oi.QueryTotal(oi.Window),
			CountErrors:      oi.QueryErrors(oi.Window),
			GraphErrorBudget: oi.QueryErrorBudget(),
			GraphRequests:    oi.RequestRange(time.Second),
			GraphErrors:      oi.ErrorsRange(time.Second),
		}
	}

	return connect.NewResponse(&objectivesv1alpha1.ListResponse{
		Objectives: resp.Msg.Objectives,
	}), nil
}

func (s *objectiveServer) GetStatus(ctx context.Context, req *connect.Request[objectivesv1alpha1.GetStatusRequest]) (*connect.Response[objectivesv1alpha1.GetStatusResponse], error) {
	objective, err := s.getObjective(ctx, req.Msg.Expr)
	if err != nil {
		return nil, err
	}

	// Merge grouping into objective's query
	if req.Msg.Grouping != "" {
		groupingMatchers, err := parser.ParseMetricSelector(req.Msg.Grouping)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
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
		if objective.Indicator.BoolGauge != nil {
			objective.Indicator.BoolGauge.LabelMatchers = append(objective.Indicator.BoolGauge.LabelMatchers, groupingMatchers...)
		}
	}

	ts := time.Now()
	if req.Msg.Time != nil {
		ts = req.Msg.Time.AsTime()
	}

	queryTotal := objective.QueryTotal(objective.Window)
	value, _, err := s.promAPI.Query(contextSetPromCache(ctx, 15*time.Second), queryTotal, ts)
	if err != nil {
		level.Warn(s.logger).Log("msg", "failed to query total", "query", queryTotal, "err", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	statuses := map[model.Fingerprint]*objectivesv1alpha1.ObjectiveStatus{}

	for _, v := range value.(model.Vector) {
		ls := make(map[string]string)
		for k, v := range v.Metric {
			ls[string(k)] = string(v)
		}

		statuses[v.Metric.Fingerprint()] = &objectivesv1alpha1.ObjectiveStatus{
			Labels: ls,
			Availability: &objectivesv1alpha1.Availability{
				Percentage: 1,
				Total:      float64(v.Value),
			},
		}
	}

	queryErrors := objective.QueryErrors(objective.Window)
	value, _, err = s.promAPI.Query(contextSetPromCache(ctx, 15*time.Second), queryErrors, ts)
	if err != nil {
		level.Warn(s.logger).Log("msg", "failed to query errors", "query", queryErrors, "err", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	for _, v := range value.(model.Vector) {
		if s, exists := statuses[v.Metric.Fingerprint()]; exists {
			s.Availability.Errors = float64(v.Value)
			s.Availability.Percentage = 1 - (s.Availability.Errors / s.Availability.Total)
		} else {
			objectiveLabels := make(map[string]string)
			for k, v := range v.Metric {
				objectiveLabels[string(k)] = string(v)
			}

			statuses[v.Metric.Fingerprint()] = &objectivesv1alpha1.ObjectiveStatus{
				Labels: objectiveLabels,
				Availability: &objectivesv1alpha1.Availability{
					Percentage: 1 - (s.Availability.Errors / s.Availability.Total),
					Total:      float64(v.Value),
				},
			}
		}
	}

	statusSlice := make([]*objectivesv1alpha1.ObjectiveStatus, 0, len(statuses))
	for _, s := range statuses {
		s.Budget = &objectivesv1alpha1.Budget{}
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

		statusSlice = append(statusSlice, s)
	}

	return connect.NewResponse(&objectivesv1alpha1.GetStatusResponse{
		Status: statusSlice,
	}), nil
}

func (s *objectiveServer) GraphErrorBudget(ctx context.Context, req *connect.Request[objectivesv1alpha1.GraphErrorBudgetRequest]) (*connect.Response[objectivesv1alpha1.GraphErrorBudgetResponse], error) {
	objective, err := s.getObjective(ctx, req.Msg.Expr)
	if err != nil {
		return nil, err
	}

	if req.Msg.Grouping != "" && req.Msg.Grouping != "{}" {
		groupingMatchers, err := parser.ParseMetricSelector(req.Msg.Grouping)
		if err != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("failed parsing alerts metric: %w", err))
		}

		if ratio := objective.Indicator.Ratio; ratio != nil {
			groupings := map[string]struct{}{}
			for _, g := range ratio.Grouping {
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

			objective.Indicator.Latency.Grouping = []string{}
			for g := range groupings {
				objective.Indicator.Latency.Grouping = append(objective.Indicator.Latency.Grouping, g)
			}
		}
		if objective.Indicator.BoolGauge != nil {
			groupings := map[string]struct{}{}
			for _, g := range objective.Indicator.BoolGauge.Grouping {
				groupings[g] = struct{}{}
			}

			for _, m := range groupingMatchers {
				objective.Indicator.BoolGauge.LabelMatchers = append(objective.Indicator.BoolGauge.LabelMatchers, m)
				delete(groupings, m.Name)
			}

			objective.Indicator.BoolGauge.Grouping = []string{}
			for g := range groupings {
				objective.Indicator.BoolGauge.Grouping = append(objective.Indicator.BoolGauge.Grouping, g)
			}
		}
	}
	if objective.Indicator.LatencyNative != nil && objective.Indicator.Ratio.Total.Name != "" {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unimplemented"))
	}

	end := time.Now()
	start := end.Add(-1 * time.Hour)

	if !req.Msg.Start.AsTime().IsZero() && !req.Msg.End.AsTime().IsZero() {
		start = req.Msg.Start.AsTime()
		end = req.Msg.End.AsTime()
	}
	step := end.Sub(start) / 1000

	query := objective.QueryErrorBudget()
	value, _, err := s.promAPI.QueryRange(contextSetPromCache(ctx, 15*time.Second), query, prometheusapiv1.Range{
		Start: start,
		End:   end,
		Step:  step,
	})
	if err != nil {
		level.Warn(s.logger).Log("msg", "failed to query error budget", "query", query, "err", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	matrix, ok := value.(model.Matrix)
	if !ok {
		err := fmt.Errorf("no matrix returned")
		level.Debug(s.logger).Log("msg", "returned data wasn't of type matrix", "query", query, "err", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if len(matrix) == 0 {
		level.Debug(s.logger).Log("msg", "returned no data", "query", query)
		return nil, connect.NewError(connect.CodeNotFound, nil)
	}

	valueLength := 0
	for _, m := range matrix {
		if len(m.Values) > valueLength {
			valueLength = len(m.Values)
		}
	}

	values := matrixToValues(matrix)

	// TODO: Return Samples from above function
	series := make([]*objectivesv1alpha1.Series, 0, len(values))
	for _, float64s := range values {
		series = append(series, &objectivesv1alpha1.Series{Values: float64s})
	}

	return connect.NewResponse(&objectivesv1alpha1.GraphErrorBudgetResponse{
		Timeseries: &objectivesv1alpha1.Timeseries{
			Query:  query,
			Series: series,
		},
	}), nil
}

func (s *objectiveServer) GetAlerts(ctx context.Context, req *connect.Request[objectivesv1alpha1.GetAlertsRequest]) (*connect.Response[objectivesv1alpha1.GetAlertsResponse], error) {
	resp, err := s.client.List(ctx, connect.NewRequest(&objectivesv1alpha1.ListRequest{
		Expr: req.Msg.Expr,
	}))
	if err != nil {
		return nil, err
	}

	objectives := make([]slo.Objective, 0, len(resp.Msg.Objectives))
	for _, o := range resp.Msg.Objectives {
		objectives = append(objectives, objectivesv1alpha1.ToInternal(o))
	}

	// Match alerts that at least have one character for the slo name.
	queryAlerts := `ALERTS{slo=~".+"}`

	var groupingMatchers []*labels.Matcher

	if req.Msg.Grouping != "" && req.Msg.Grouping != "{}" {
		expr, err := parser.ParseExpr(queryAlerts)
		if err != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("failed parsing alerts metric: %w", err))
		}

		// If grouping exists we merge those matchers directly into the queryAlerts query.
		groupingMatchers, err = parser.ParseMetricSelector(req.Msg.Grouping)
		if err != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("failed parsing grouping matchers: %w", err))
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

	value, _, err := s.promAPI.Query(contextSetPromCache(ctx, 5*time.Second), queryAlerts, time.Now())
	if err != nil {
		level.Warn(s.logger).Log("msg", "failed to query alerts", "query", queryAlerts, "err", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	vector, ok := value.(model.Vector)
	if !ok {
		err := fmt.Errorf("no vector returned")
		level.Debug(s.logger).Log("msg", "returned data wasn't of type vector", "query", queryAlerts, "err", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	alerts := alertsMatchingObjectives(vector, objectives, groupingMatchers, req.Msg.Inactive)

	if req.Msg.Current {
		for _, objective := range objectives {
			mtx := &sync.Mutex{}
			windowsMap := map[time.Duration]float64{}
			for _, w := range objective.Windows() {
				windowsMap[w.Short] = -1
				windowsMap[w.Long] = -1
			}

			var wg sync.WaitGroup
			for w := range windowsMap {
				wg.Add(1)
				go func(w time.Duration) {
					defer wg.Done()

					query, err := objective.QueryBurnrate(w, groupingMatchers)
					if err != nil {
						level.Warn(s.logger).Log("msg", "failed to prepare current burn rate query", "err", err)
						return
					}
					value, _, err := s.promAPI.Query(contextSetPromCache(ctx, instantCache(w)), query, time.Now())
					if err != nil {
						level.Warn(s.logger).Log("msg", "failed to query current burn rate", "query", query, "err", err)
						return
					}
					vec, ok := value.(model.Vector)
					if !ok {
						level.Warn(s.logger).Log("msg", "failed to query current burn rate", "query", query, "err", "expected vector value from Prometheus")
						return
					}
					if vec.Len() == 0 {
						return
					}
					if vec.Len() != 1 {
						level.Warn(s.logger).Log("msg", "failed to query current burn rate", "query", query, "err", "expected vector with one value from Prometheus")
						return
					}

					current := float64(vec[0].Value)
					if math.IsNaN(current) {
						// ignore current values if NaN and return the -1 indicating NaN
						return
					}

					mtx.Lock()
					windowsMap[w] = current
					mtx.Unlock()
				}(w)
			}

			wg.Wait()

			// Match objectives to alerts to update response
		Alerts:
			for i, alert := range alerts {
				for k, v := range alert.Labels {
					if objective.Labels.Get(k) != v {
						continue Alerts
					}
				}
				short := alert.Short.Window
				alerts[i].Short.Window = short
				alerts[i].Short.Current = windowsMap[short.AsDuration()]
				long := alert.Long.Window
				alerts[i].Long.Window = long
				alerts[i].Long.Current = windowsMap[long.AsDuration()]
			}
		}
	}

	return connect.NewResponse(&objectivesv1alpha1.GetAlertsResponse{Alerts: alerts}), nil
}

// alertsMatchingObjectives loops through all alerts trying to match objectives based on their labels.
// All labels of an objective need to be equal if they exist on the ALERTS metric.
// Therefore, only a subset on labels are taken into account
// which gives the ALERTS metric the opportunity to include more custom labels.
func alertsMatchingObjectives(metrics model.Vector, objectives []slo.Objective, grouping []*labels.Matcher, inactive bool) []*objectivesv1alpha1.Alert {
	alerts := make([]*objectivesv1alpha1.Alert, 0, len(metrics))

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
				queryShort, _ := o.QueryBurnrate(w.Short, grouping)
				queryLong, _ := o.QueryBurnrate(w.Long, grouping)

				alerts = append(alerts, &objectivesv1alpha1.Alert{
					Labels:   lset,
					Severity: string(w.Severity),
					For:      durationpb.New(w.For),
					Factor:   w.Factor,
					Short: &objectivesv1alpha1.Burnrate{
						Window:  durationpb.New(w.Short),
						Current: -1,
						Query:   queryShort,
					},
					Long: &objectivesv1alpha1.Burnrate{
						Window:  durationpb.New(w.Long),
						Current: -1,
						Query:   queryLong,
					},
					State: objectivesv1alpha1.Alert_inactive,
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
				alerts = append(alerts, &objectivesv1alpha1.Alert{
					Labels:   lset,
					Severity: string(sample.Metric["severity"]),
					State: objectivesv1alpha1.Alert_State( // Convert string to Alert_State enum
						objectivesv1alpha1.Alert_State_value[strings.ToLower(
							string(sample.Metric["alertstate"]),
						)],
					),
					For:    durationpb.New(window.For),
					Factor: window.Factor,
					Short: &objectivesv1alpha1.Burnrate{
						Window:  durationpb.New(time.Duration(short)),
						Current: -1,
						Query:   o.Burnrate(time.Duration(short)),
					},
					Long: &objectivesv1alpha1.Burnrate{
						Window:  durationpb.New(time.Duration(long)),
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
					if alert.Short.Window.AsDuration() != time.Duration(short) {
						continue alerts
					}
					if alert.Long.Window.AsDuration() != time.Duration(long) {
						continue alerts
					}
					// only update state for exact short and long burn rate match
					alerts[i].State = objectivesv1alpha1.Alert_State( // Convert string to Alert_State enum
						objectivesv1alpha1.Alert_State_value[strings.ToLower(
							string(sample.Metric["alertstate"]),
						)],
					)
				}
			}
		}
	}

	return alerts
}

func (s *objectiveServer) GraphRate(ctx context.Context, req *connect.Request[objectivesv1alpha1.GraphRateRequest]) (*connect.Response[objectivesv1alpha1.GraphRateResponse], error) {
	objective, err := s.getObjective(ctx, req.Msg.Expr)
	if err != nil {
		return nil, err
	}

	// Merge grouping into objective's query
	if req.Msg.Grouping != "" {
		groupingMatchers, err := parser.ParseMetricSelector(req.Msg.Grouping)
		if err != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("failed to parse expr: %w", err))
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
		if objective.Indicator.BoolGauge != nil {
			objective.Indicator.BoolGauge.LabelMatchers = append(objective.Indicator.BoolGauge.LabelMatchers, groupingMatchers...)
		}
	}

	end := time.Now()
	start := end.Add(-1 * time.Hour)

	if !req.Msg.Start.AsTime().IsZero() && !req.Msg.End.AsTime().IsZero() {
		start = req.Msg.Start.AsTime()
		end = req.Msg.End.AsTime()
	}
	step := end.Sub(start) / 1000

	timeRange := rangeInterval(start, end)
	cacheDuration := rangeCache(start, end)

	query := objective.RequestRange(timeRange)

	value, _, err := s.promAPI.QueryRange(contextSetPromCache(ctx, cacheDuration), query, prometheusapiv1.Range{
		Start: start,
		End:   end,
		Step:  step,
	})
	if err != nil {
		level.Warn(s.logger).Log("msg", "failed to run range request", "query", query, "err", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if value.Type() != model.ValMatrix {
		err := fmt.Errorf("returned data is not a matrix")
		level.Warn(s.logger).Log("query", query, "err", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	matrix, ok := value.(model.Matrix)
	if !ok {
		err := fmt.Errorf("no matrix returned")
		level.Warn(s.logger).Log("query", query, "err", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if len(matrix) == 0 {
		level.Debug(s.logger).Log("msg", "no data returned", "query", query)
		return nil, connect.NewError(connect.CodeNotFound, err)
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

	// TODO: Return Samples from above function
	series := make([]*objectivesv1alpha1.Series, 0, len(values))
	for _, float64s := range values {
		series = append(series, &objectivesv1alpha1.Series{Values: float64s})
	}

	return connect.NewResponse(&objectivesv1alpha1.GraphRateResponse{
		Timeseries: &objectivesv1alpha1.Timeseries{
			Labels: labels,
			Query:  query,
			Series: series,
		},
	}), nil
}

func (s *objectiveServer) GraphErrors(ctx context.Context, req *connect.Request[objectivesv1alpha1.GraphErrorsRequest]) (*connect.Response[objectivesv1alpha1.GraphErrorsResponse], error) {
	objective, err := s.getObjective(ctx, req.Msg.Expr)
	if err != nil {
		return nil, err
	}

	// Merge grouping into objective's query
	if req.Msg.Grouping != "" {
		groupingMatchers, err := parser.ParseMetricSelector(req.Msg.Grouping)
		if err != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("failed to parse expr: %w", err))
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
		if objective.Indicator.BoolGauge != nil {
			objective.Indicator.BoolGauge.LabelMatchers = append(objective.Indicator.BoolGauge.LabelMatchers, groupingMatchers...)
		}
	}

	end := time.Now()
	start := end.Add(-1 * time.Hour)

	if !req.Msg.Start.AsTime().IsZero() && !req.Msg.End.AsTime().IsZero() {
		start = req.Msg.Start.AsTime()
		end = req.Msg.End.AsTime()
	}
	step := end.Sub(start) / 1000

	timeRange := rangeInterval(start, end)
	cacheDuration := rangeCache(start, end)

	query := objective.ErrorsRange(timeRange)
	value, _, err := s.promAPI.QueryRange(contextSetPromCache(ctx, cacheDuration), query, prometheusapiv1.Range{
		Start: start,
		End:   end,
		Step:  step,
	})
	if err != nil {
		level.Warn(s.logger).Log("msg", "failed to run range error request", "query", query, "err", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if value.Type() != model.ValMatrix {
		err := fmt.Errorf("returned data is not a matrix")
		level.Warn(s.logger).Log("query", query, "err", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	matrix, ok := value.(model.Matrix)
	if !ok {
		err := fmt.Errorf("no matrix returned")
		level.Warn(s.logger).Log("query", query, "err", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if len(matrix) == 0 {
		level.Debug(s.logger).Log("msg", "no data returned", "query", query)
		return nil, connect.NewError(connect.CodeNotFound, err)
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

	// TODO: Return Samples from above function
	series := make([]*objectivesv1alpha1.Series, 0, len(values))
	for _, float64s := range values {
		series = append(series, &objectivesv1alpha1.Series{Values: float64s})
	}

	return connect.NewResponse(&objectivesv1alpha1.GraphErrorsResponse{
		Timeseries: &objectivesv1alpha1.Timeseries{
			Labels: labels,
			Query:  query,
			Series: series,
		},
	}), nil
}

var percentiles = []float64{0.999, 0.99, 0.95, 0.9, 0.5}

func (s *objectiveServer) GraphDuration(ctx context.Context, req *connect.Request[objectivesv1alpha1.GraphDurationRequest]) (*connect.Response[objectivesv1alpha1.GraphDurationResponse], error) {
	objective, err := s.getObjective(ctx, req.Msg.Expr)
	if err != nil {
		return nil, err
	}

	// Merge grouping into objective's query
	if req.Msg.Grouping != "" {
		groupingMatchers, err := parser.ParseMetricSelector(req.Msg.Grouping)
		if err != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("failed to parse expr: %w", err))
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
		if objective.Indicator.BoolGauge != nil {
			objective.Indicator.BoolGauge.LabelMatchers = append(objective.Indicator.BoolGauge.LabelMatchers, groupingMatchers...)
		}
	}

	end := time.Now()
	start := end.Add(-1 * time.Hour)

	if !req.Msg.Start.AsTime().IsZero() && !req.Msg.End.AsTime().IsZero() {
		start = req.Msg.Start.AsTime()
		end = req.Msg.End.AsTime()
	}
	step := end.Sub(start) / 1000

	timeRange := rangeInterval(start, end)
	cacheDuration := rangeCache(start, end)

	timeseries := make([]*objectivesv1alpha1.Timeseries, 0, len(percentiles))

	objectivePercentiles := percentiles
	contains := false
	for _, p := range percentiles {
		if p == objective.Target {
			contains = true
		}
	}
	if !contains {
		objectivePercentiles = append(objectivePercentiles, objective.Target)
	}

	// sort in descending order
	sort.Slice(objectivePercentiles, func(i, j int) bool {
		return objectivePercentiles[i] > objectivePercentiles[j]
	})

	for _, percentile := range objectivePercentiles {
		if objective.Target >= percentile {
			query := objective.DurationRange(timeRange, percentile)
			value, _, err := s.promAPI.QueryRange(contextSetPromCache(ctx, cacheDuration), query, prometheusapiv1.Range{
				Start: start,
				End:   end,
				Step:  step,
			})
			if err != nil {
				level.Warn(s.logger).Log("msg", "failed to run range error request", "query", query, "err", err)
				return nil, connect.NewError(connect.CodeInternal, err)
			}

			if value.Type() != model.ValMatrix {
				err := fmt.Errorf("returned data is not a matrix")
				level.Warn(s.logger).Log("query", query, "err", err)
				return nil, connect.NewError(connect.CodeInternal, err)
			}

			matrix, ok := value.(model.Matrix)
			if !ok {
				err := fmt.Errorf("no matrix returned")
				level.Warn(s.logger).Log("query", query, "err", err)
				return nil, connect.NewError(connect.CodeInternal, err)
			}

			if len(matrix) == 0 {
				level.Debug(s.logger).Log("msg", "no data returned", "query", query)
				return nil, connect.NewError(connect.CodeNotFound, err)
			}

			valueLength := 0
			for _, m := range matrix {
				if len(m.Values) > valueLength {
					valueLength = len(m.Values)
				}
			}

			values := matrixToValues(matrix)

			series := make([]*objectivesv1alpha1.Series, 0, len(values))
			for _, float64s := range values {
				series = append(series, &objectivesv1alpha1.Series{Values: float64s})
			}

			timeseries = append(timeseries,
				&objectivesv1alpha1.Timeseries{
					Labels: []string{fmt.Sprintf(`{quantile="p%.f"}`, 100*percentile)}, // TODO: Nicer format float
					Query:  query,
					Series: series,
				},
			)
		}
	}

	return connect.NewResponse(&objectivesv1alpha1.GraphDurationResponse{
		Timeseries: timeseries,
	}), nil
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
	return instantCache(end.Sub(start))
}

func instantCache(duration time.Duration) time.Duration {
	d := 15 * time.Second
	// TODO: Refactor for early returns instead
	if duration >= month {
		d = 5 * time.Minute
	} else if duration >= week {
		d = 3 * time.Minute
	} else if duration >= day {
		d = 90 * time.Second
	} else if duration >= hours12 {
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
