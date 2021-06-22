package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/dgraph-io/ristretto"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/metalmatze/athene/openapi"
	openapiclient "github.com/metalmatze/athene/openapi/client"
	openapiserver "github.com/metalmatze/athene/openapi/server/go"
	"github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

//go:embed ui/build
var ui embed.FS

func main() {
	build, err := fs.Sub(ui, "ui/build")
	if err != nil {
		log.Fatal(err)
	}

	prometheusURL := flag.String("prometheus.url", "http://localhost:9090", "The URL to the Prometheus to query.")
	apiURL := flag.String("api.url", "http://localhost:9444", "The URL to the API service like a Kubernetes Operator.")
	flag.Parse()

	log.Println("Using Prometheus at", *prometheusURL)
	log.Println("Using API at", *apiURL)

	client, err := api.NewClient(api.Config{Address: *prometheusURL})
	if err != nil {
		log.Fatal(err)
	}
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
		api:   prometheusv1.NewAPI(client),
		cache: cache,
	}

	parsedAPIURL, err := url.Parse(*apiURL)
	if err != nil {
		log.Fatal(err)
		return
	}

	apiConfig := openapiclient.NewConfiguration()
	apiConfig.Scheme = parsedAPIURL.Scheme
	apiConfig.Host = parsedAPIURL.Host
	apiClient := openapiclient.NewAPIClient(apiConfig)

	router := openapiserver.NewRouter(
		openapiserver.NewObjectivesApiController(&ObjectivesServer{
			promAPI:   promAPI,
			apiclient: apiClient,
		}),
	)

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{})) // TODO: Disable by default
	r.Mount("/api/v1", router)
	r.Get("/objectives/{name}", func(w http.ResponseWriter, r *http.Request) {
		file, err := build.Open("index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = io.Copy(w, file)
	})
	r.Handle("/*", http.FileServer(http.FS(build)))

	if err := http.ListenAndServe(":9099", r); err != nil {
		log.Fatal(err)
	}
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
	api   prometheusv1.API
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

func (o *ObjectivesServer) GetObjective(ctx context.Context, name string) (openapiserver.ImplResponse, error) {
	objective, _, err := o.apiclient.ObjectivesApi.GetObjective(ctx, name).Execute()
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	return openapiserver.ImplResponse{
		Code: http.StatusCreated,
		Body: openapi.ServerFromClient(objective),
	}, nil
}

func (o *ObjectivesServer) GetObjectiveStatus(ctx context.Context, name string) (openapiserver.ImplResponse, error) {
	clientObjective, _, err := o.apiclient.ObjectivesApi.GetObjective(ctx, name).Execute()
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

func (o *ObjectivesServer) GetObjectiveErrorBudget(ctx context.Context, name string, startTimestamp int32, endTimestamp int32) (openapiserver.ImplResponse, error) {
	clientObjective, _, err := o.apiclient.ObjectivesApi.GetObjective(ctx, name).Execute()
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

	pairs := make([]openapiserver.ErrorBudgetPair, len(matrix[0].Values))

	for _, m := range matrix {
		for i, pair := range m.Values {
			pairs[i] = openapiserver.ErrorBudgetPair{T: pair.Timestamp.Unix(), V: float64(pair.Value)}
		}
	}

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: openapiserver.ErrorBudget{Pair: pairs},
	}, nil
}

const (
	alertstateInactive = "inactive"
	alertstatePending  = "pending"
	alertstateFiring   = "firing"
)

func (o *ObjectivesServer) GetMultiBurnrateAlerts(ctx context.Context, name string) (openapiserver.ImplResponse, error) {
	clientObjective, _, err := o.apiclient.ObjectivesApi.GetObjective(ctx, name).Execute()
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

func (o *ObjectivesServer) GetREDRequests(ctx context.Context, name string, startTimestamp int32, endTimestamp int32) (openapiserver.ImplResponse, error) {
	clientObjective, _, err := o.apiclient.ObjectivesApi.GetObjective(ctx, name).Execute()
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

	labels := make([]string, len(matrix))
	values := make([][]float64, len(matrix[0].Values))

	for i, m := range matrix {
		labels[i] = model.LabelSet(m.Metric).String()

		for j, pair := range m.Values {
			if i == 0 {
				values[j] = make([]float64, len(matrix)+1) // +1 because the first value is the timestamp
				values[j][0] = float64(pair.Timestamp.Unix())
				values[j][1] = float64(pair.Value)
			}
			values[j][i+1] = float64(pair.Value) // i+1 because the first value is the timestamp
		}
	}

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: openapiserver.QueryRange{
			Labels: labels,
			Values: values,
		},
	}, nil
}

func (o *ObjectivesServer) GetREDErrors(ctx context.Context, name string, startTimestamp int32, endTimestamp int32) (openapiserver.ImplResponse, error) {
	clientObjective, _, err := o.apiclient.ObjectivesApi.GetObjective(ctx, name).Execute()
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
	values := make([][]float64, valueLength)

	pairLength := len(matrix) + 1 // +1 because the first value is the timestamp

	for i, m := range matrix {
		labels[i] = model.LabelSet(m.Metric).String()

		for j, pair := range m.Values {
			if cap(values[j]) == 0 {
				values[j] = make([]float64, pairLength)
			}
			values[j][0] = float64(pair.Timestamp.Unix())
			values[j][1] = float64(pair.Value)
			values[j][i+1] = float64(pair.Value) // i+1 because the first value is the timestamp
		}
	}

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: openapiserver.QueryRange{
			Labels: labels,
			Values: values,
		},
	}, nil
}
