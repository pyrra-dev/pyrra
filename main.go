package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/kong"
	"github.com/cespare/xxhash/v2"
	"github.com/dgraph-io/ristretto"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	promconfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"github.com/pyrra-dev/pyrra/openapi"
	openapiclient "github.com/pyrra-dev/pyrra/openapi/client"
	openapiserver "github.com/pyrra-dev/pyrra/openapi/server/go"
)

//go:embed ui/build
var ui embed.FS

var CLI struct {
	API struct {
		PrometheusURL             *url.URL `default:"http://localhost:9090" help:"The URL to the Prometheus to query."`
		PrometheusExternalURL     *url.URL `help:"The URL for the UI to redirect users to when opening Prometheus. If empty the same as prometheus.url"`
		ApiURL                    *url.URL `default:"http://localhost:9444" help:"The URL to the API service like a Kubernetes Operator."`
		PrometheusBearerTokenPath string   `default:"" help:"Bearer token path"`
	} `cmd:"" help:"Runs Pyrra's API and UI."`
	Filesystem struct {
		ConfigFiles      string `default:"/etc/pyrra/*.yaml" help:"The folder where Pyrra finds the config files to use."`
		PrometheusFolder string `default:"/etc/prometheus/pyrra/" help:"The folder where Pyrra writes the generates Prometheus rules and alerts."`
	} `cmd:"" help:"Runs Pyrra's filesystem operator and backend for the API."`
	Kubernetes struct {
		MetricsAddr string `default:":8080" help:"The address the metric endpoint binds to."`
	} `cmd:"" help:"Runs Pyrra's Kubernetes operator and backend for the API."`
}

func main() {
	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "api":
		cmdAPI(CLI.API.PrometheusURL, CLI.API.PrometheusExternalURL, CLI.API.ApiURL, CLI.API.PrometheusBearerTokenPath)
	case "filesystem":
		cmdFilesystem(CLI.Filesystem.ConfigFiles, CLI.Filesystem.PrometheusFolder)
	case "kubernetes":
		cmdKubernetes(CLI.Kubernetes.MetricsAddr)
	}

	return
}

func cmdAPI(prometheusURL, prometheusExternal, apiURL *url.URL, prometheusBearerTokenPath string) {
	build, err := fs.Sub(ui, "ui/build")
	if err != nil {
		log.Fatal(err)
	}

	if prometheusExternal == nil {
		prometheusExternal = prometheusURL
	}

	log.Println("Using Prometheus at", prometheusURL.String())
	log.Println("Using external Prometheus at", prometheusExternal.String())
	log.Println("Using API at", apiURL.String())

	reg := prometheus.NewRegistry()

	config := api.Config{Address: prometheusURL.String()}
	if len(prometheusBearerTokenPath) > 0 {
		config.RoundTripper = promconfig.NewAuthorizationCredentialsFileRoundTripper("Bearer", prometheusBearerTokenPath, api.DefaultRoundTripper)
	}

	client, err := api.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	thanosClient := newThanosClient(client)

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cache.Close()
	promAPI := &promCache{
		api:   prometheusv1.NewAPI(thanosClient),
		cache: cache,
	}

	apiConfig := openapiclient.NewConfiguration()
	apiConfig.Scheme = apiURL.Scheme
	apiConfig.Host = apiURL.Host
	apiClient := openapiclient.NewAPIClient(apiConfig)

	router := openapiserver.NewRouter(
		openapiserver.NewObjectivesApiController(&ObjectivesServer{
			promAPI:   promAPI,
			apiclient: apiClient,
		}),
	)
	router.Use(openapi.MiddlewareMetrics(reg))

	tmpl, err := template.ParseFS(build, "index.html")
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{})) // TODO: Disable by default
	r.Mount("/api/v1", router)
	r.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	r.Get("/objectives/{namespace}/{name}", func(w http.ResponseWriter, r *http.Request) {
		if err := tmpl.Execute(w, struct {
			PrometheusURL string
		}{
			PrometheusURL: prometheusExternal.String(),
		}); err != nil {
			log.Println(err)
			return
		}
	})
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			if err := tmpl.Execute(w, struct {
				PrometheusURL string
			}{
				PrometheusURL: prometheusExternal.String(),
			}); err != nil {
				log.Println(err)
			}
			return
		}

		http.FileServer(http.FS(build)).ServeHTTP(w, r)
	}))

	if err := http.ListenAndServe(":9099", r); err != nil {
		log.Fatal(err)
	}
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
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}
	query, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, nil, err
	}

	// We don't want partial responses, especially not when calculating error budgets.
	query.Set("partial_response", "false")
	r.ContentLength += 23

	if strings.HasSuffix(r.URL.Path, "/api/v1/query_range") {
		start, err := strconv.ParseFloat(query.Get("start"), 64)
		if err != nil {
			return nil, nil, err
		}
		end, err := strconv.ParseFloat(query.Get("end"), 64)
		if err != nil {
			return nil, nil, err
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

func (p *promCache) Query(ctx context.Context, query string, ts time.Time) (model.Value, prometheusv1.Warnings, error) {
	xxh := xxhash.New()
	_, _ = xxh.WriteString(query)
	hash := xxh.Sum64()

	if value, exists := p.cache.Get(hash); exists {
		return value.(model.Value), nil, nil
	}

	value, _, err := p.api.Query(ctx, query, ts)
	if err != nil {
		return nil, nil, err
	}

	// TODO might need to pass cache duration via ctx?
	_ = p.cache.SetWithTTL(hash, value, 10, 5*time.Minute)

	return value, nil, nil
}

func (p *promCache) QueryRange(ctx context.Context, query string, r prometheusv1.Range) (model.Value, prometheusv1.Warnings, error) {
	xxh := xxhash.New()
	_, _ = xxh.WriteString(query)
	_, _ = xxh.WriteString(r.Start.String())
	_, _ = xxh.WriteString(r.End.String())
	hash := xxh.Sum64()

	if value, exists := p.cache.Get(hash); exists {
		return value.(model.Value), nil, nil
	}

	value, _, err := p.api.QueryRange(ctx, query, r)
	if err != nil {
		return nil, nil, err
	}

	// TODO might need to pass cache duration via ctx?
	_ = p.cache.SetWithTTL(hash, value, 100, 10*time.Minute)

	return value, nil, nil
}

type ObjectivesServer struct {
	promAPI   *promCache
	apiclient *openapiclient.APIClient
}

func (o *ObjectivesServer) ListObjectives(ctx context.Context) (openapiserver.ImplResponse, error) {
	objectives, _, err := o.apiclient.ObjectivesApi.ListObjectives(ctx).Execute()
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

func (o *ObjectivesServer) GetObjective(ctx context.Context, namespace, name string) (openapiserver.ImplResponse, error) {
	objective, _, err := o.apiclient.ObjectivesApi.GetObjective(ctx, namespace, name).Execute()
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	return openapiserver.ImplResponse{
		Code: http.StatusCreated,
		Body: openapi.ServerFromClient(objective),
	}, nil
}

func (o *ObjectivesServer) GetObjectiveStatus(ctx context.Context, namespace, name string) (openapiserver.ImplResponse, error) {
	clientObjective, _, err := o.apiclient.ObjectivesApi.GetObjective(ctx, namespace, name).Execute()
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	objective := openapi.InternalFromClient(clientObjective)

	ts := RoundUp(time.Now().UTC(), 5*time.Minute)

	status := openapiserver.ObjectiveStatus{
		Availability: openapiserver.ObjectiveStatusAvailability{},
		Budget:       openapiserver.ObjectiveStatusBudget{},
	}

	queryTotal := objective.QueryTotal(objective.Window)
	log.Println(queryTotal)
	value, _, err := o.promAPI.Query(ctx, queryTotal, ts)
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	vec := value.(model.Vector)
	for _, v := range vec {
		status.Availability.Total = float64(v.Value)
	}

	if status.Availability.Total == 0 {
		return openapiserver.ImplResponse{Code: http.StatusNotFound}, nil
	}

	queryErrors := objective.QueryErrors(objective.Window)
	log.Println(queryErrors)
	value, _, err = o.promAPI.Query(ctx, queryErrors, ts)
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	for _, v := range value.(model.Vector) {
		status.Availability.Errors = float64(v.Value)
	}

	status.Availability.Percentage = 1 - (status.Availability.Errors / status.Availability.Total)

	status.Budget.Total = 1 - objective.Target
	status.Budget.Remaining = (status.Budget.Total - (status.Availability.Errors / status.Availability.Total)) / status.Budget.Total
	status.Budget.Max = status.Budget.Total * status.Availability.Total

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: status,
	}, nil
}

func (o *ObjectivesServer) GetObjectiveErrorBudget(ctx context.Context, namespace, name string, startTimestamp int32, endTimestamp int32) (openapiserver.ImplResponse, error) {
	clientObjective, _, err := o.apiclient.ObjectivesApi.GetObjective(ctx, namespace, name).Execute()
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	objective := openapi.InternalFromClient(clientObjective)

	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now

	if startTimestamp != 0 && endTimestamp != 0 {
		start = time.Unix(int64(startTimestamp), 0)
		end = time.Unix(int64(endTimestamp), 0)
	}

	step := end.Sub(start) / 1000

	query := objective.QueryErrorBudget()
	log.Println(query)
	value, _, err := o.promAPI.QueryRange(ctx, query, prometheusv1.Range{
		Start: start,
		End:   end,
		Step:  step,
	})
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	matrix, ok := value.(model.Matrix)
	if !ok {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("no matrix returned")
	}

	if len(matrix) != 1 {
		return openapiserver.ImplResponse{Code: http.StatusNotFound}, fmt.Errorf("no data")
	}

	if len(matrix) == 0 {
		return openapiserver.ImplResponse{Code: http.StatusNotFound}, fmt.Errorf("no data")
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

func (o *ObjectivesServer) GetMultiBurnrateAlerts(ctx context.Context, namespace, name string) (openapiserver.ImplResponse, error) {
	clientObjective, _, err := o.apiclient.ObjectivesApi.GetObjective(ctx, namespace, name).Execute()
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	objective := openapi.InternalFromClient(clientObjective)

	baseAlerts, err := objective.Alerts()
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	var alerts []openapiserver.MultiBurnrateAlert

	for _, ba := range baseAlerts {
		short := &openapiserver.Burnrate{
			Window:  ba.Short.Milliseconds(),
			Current: -1,
			Query:   ba.QueryShort,
		}
		long := &openapiserver.Burnrate{
			Window:  ba.Long.Milliseconds(),
			Current: -1,
			Query:   ba.QueryLong,
		}

		var wg sync.WaitGroup
		wg.Add(3)

		go func(b *openapiserver.Burnrate) {
			defer wg.Done()

			value, _, err := o.promAPI.Query(ctx, b.Query, time.Now())
			if err != nil {
				log.Println(err)
				return
			}
			vec, ok := value.(model.Vector)
			if !ok {
				log.Println("no vector")
				return
			}
			if vec.Len() != 1 {
				return
			}
			b.Current = float64(vec[0].Value)
		}(short)

		go func(b *openapiserver.Burnrate) {
			defer wg.Done()

			value, _, err := o.promAPI.Query(ctx, b.Query, time.Now())
			if err != nil {
				log.Println(err)
				return
			}
			vec, ok := value.(model.Vector)
			if !ok {
				log.Println("no vector")
				return
			}
			if vec.Len() != 1 {
				return
			}
			b.Current = float64(vec[0].Value)
		}(long)

		alertstate := alertstateInactive

		go func(name string, short, long int64) {
			defer wg.Done()

			s := model.Duration(time.Duration(short) * time.Millisecond)
			l := model.Duration(time.Duration(long) * time.Millisecond)

			query := fmt.Sprintf(`ALERTS{slo="%s",short="%s",long="%s"}`, name, s, l)
			value, _, err := o.promAPI.Query(ctx, query, time.Now())
			if err != nil {
				log.Println(err)
				return
			}
			vec, ok := value.(model.Vector)
			if !ok {
				log.Println("no vector")
				return
			}
			if vec.Len() != 1 {
				return
			}
			sample := vec[0]

			if sample.Value != 1 {
				log.Println("alert is not pending or firing")
				return
			}

			ls := model.LabelSet(sample.Metric)
			as := ls["alertstate"]
			if as == alertstatePending {
				alertstate = alertstatePending
			}
			if as == alertstateFiring {
				alertstate = alertstateFiring
			}
		}(objective.Name, short.Window, long.Window)

		wg.Wait()

		alerts = append(alerts, openapiserver.MultiBurnrateAlert{
			Severity: ba.Severity,
			For:      ba.For.Milliseconds(),
			Factor:   ba.Factor,
			Short:    *short,
			Long:     *long,
			State:    alertstate,
		})
	}

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: alerts,
	}, nil
}

func (o *ObjectivesServer) GetREDRequests(ctx context.Context, namespace, name string, startTimestamp int32, endTimestamp int32) (openapiserver.ImplResponse, error) {
	clientObjective, _, err := o.apiclient.ObjectivesApi.GetObjective(ctx, namespace, name).Execute()
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	objective := openapi.InternalFromClient(clientObjective)

	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now

	if startTimestamp != 0 && endTimestamp != 0 {
		start = time.Unix(int64(startTimestamp), 0)
		end = time.Unix(int64(endTimestamp), 0)
	}
	step := end.Sub(start) / 1000

	diff := end.Sub(start)
	timeRange := 5 * time.Minute
	if diff >= 28*24*time.Hour {
		timeRange = 3 * time.Hour
	} else if diff >= 7*24*time.Hour {
		timeRange = time.Hour
	} else if diff >= 24*time.Hour {
		timeRange = 30 * time.Minute
	} else if diff >= 12*time.Hour {
		timeRange = 15 * time.Minute
	}
	query := objective.RequestRange(timeRange)
	log.Println(query)

	value, _, err := o.promAPI.QueryRange(ctx, query, prometheusv1.Range{
		Start: start,
		End:   end,
		Step:  step,
	})
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	if value.Type() != model.ValMatrix {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("returned data is not a matrix")
	}

	matrix, ok := value.(model.Matrix)
	if !ok {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("no matrix returned")
	}

	if len(matrix) == 0 {
		return openapiserver.ImplResponse{Code: http.StatusNotFound}, fmt.Errorf("no data")
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

func (o *ObjectivesServer) GetREDErrors(ctx context.Context, namespace, name string, startTimestamp int32, endTimestamp int32) (openapiserver.ImplResponse, error) {
	clientObjective, _, err := o.apiclient.ObjectivesApi.GetObjective(ctx, namespace, name).Execute()
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	objective := openapi.InternalFromClient(clientObjective)

	now := time.Now()
	start := now.Add(-1 * time.Hour)
	end := now

	if startTimestamp != 0 && endTimestamp != 0 {
		start = time.Unix(int64(startTimestamp), 0)
		end = time.Unix(int64(endTimestamp), 0)
	}
	step := end.Sub(start) / 1000

	diff := end.Sub(start)
	timeRange := 5 * time.Minute
	if diff >= 28*24*time.Hour {
		timeRange = 3 * time.Hour
	} else if diff >= 7*24*time.Hour {
		timeRange = time.Hour
	} else if diff >= 24*time.Hour {
		timeRange = 30 * time.Minute
	} else if diff >= 12*time.Hour {
		timeRange = 15 * time.Minute
	}

	query := objective.ErrorsRange(timeRange)
	log.Println(query)

	value, _, err := o.promAPI.QueryRange(ctx, query, prometheusv1.Range{
		Start: start,
		End:   end,
		Step:  step,
	})
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	if value.Type() != model.ValMatrix {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("returned data is not a matrix")
	}

	matrix, ok := value.(model.Matrix)
	if !ok {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("no matrix returned")
	}

	if len(matrix) == 0 {
		return openapiserver.ImplResponse{Code: http.StatusNotFound}, fmt.Errorf("no data")
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

func matrixToValues(m model.Matrix) [][]float64 {
	valuesMap := make(map[int64][]float64, len(m[0].Values))

	for i, stream := range m {
		for _, pair := range stream.Values {
			if cap(valuesMap[pair.Timestamp.Unix()]) == 0 {
				valuesMap[pair.Timestamp.Unix()] = make([]float64, len(m))
			}
			valuesMap[pair.Timestamp.Unix()][i] = float64(pair.Value)
		}
	}

	values := make([][]float64, 0, len(m[0].Values))
	for t, vs := range valuesMap {
		values = append(values, append([]float64{float64(t)}, vs...))
	}

	// TODO: Maybe there's a way to do it without a map and sort
	// as this sort is super CPU intensive
	sort.Slice(values, func(i, j int) bool {
		return values[i][0] < values[j][0]
	})

	return values
}
