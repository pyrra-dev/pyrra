package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"sigs.k8s.io/yaml"

	"github.com/pyrra-dev/pyrra/kubernetes/api/v1alpha1"
	"github.com/pyrra-dev/pyrra/openapi"
	openapiserver "github.com/pyrra-dev/pyrra/openapi/server/go"
	"github.com/pyrra-dev/pyrra/slo"
)

var errEndpointNotImplemented = errors.New("endpoint not implement")

type Objectives struct {
	mu         sync.RWMutex
	objectives map[string]slo.Objective
}

func (os *Objectives) Set(o slo.Objective) {
	os.mu.Lock()
	os.objectives[o.Labels.String()] = o
	os.mu.Unlock()
}

func (os *Objectives) Match(ms []*labels.Matcher) []slo.Objective {
	if len(ms) == 0 {
		os.mu.RLock()
		objectives := make([]slo.Objective, 0, len(os.objectives))
		for _, o := range os.objectives {
			objectives = append(objectives, o)
		}
		os.mu.RUnlock()
		return objectives
	}

	os.mu.RLock()
	defer os.mu.RUnlock()

	var objectives []slo.Objective

Objectives:
	for _, o := range os.objectives {
		for _, m := range ms {
			v := o.Labels.Get(m.Name)
			// If there's no label with this exact name,
			// check if there are labels with the label prefix.
			if v == "" {
				v = o.Labels.Get(slo.PropagationLabelsPrefix + m.Name)
			}
			if !m.Matches(v) {
				// If no label matches then maybe the objective is grouped by this label
				var grouping bool
				for _, g := range o.Grouping() {
					if m.Name == g {
						grouping = true
					}
				}
				// If the label is not a grouping either then skip this objective
				if !grouping {
					continue Objectives
				}
			}
		}
		objectives = append(objectives, o)
	}

	return objectives
}

func cmdFilesystem(logger log.Logger, reg *prometheus.Registry, promClient api.Client, configFiles, prometheusFolder string) int {
	reconcilesTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pyrra_filesystem_reconciles_total",
		Help: "The total amount of reconciles.",
	})
	reconcilesErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pyrra_filesystem_reconciles_errors_total",
		Help: "The total amount of errors during reconciles.",
	})

	reg.MustRegister(
		reconcilesTotal,
		reconcilesErrors,
	)

	ctx, cancel := context.WithCancel(context.Background())
	objectives := &Objectives{objectives: map[string]slo.Objective{}}
	files := make(chan string, 16)
	reload := make(chan struct{}, 16)

	var gr run.Group
	{
		gr.Add(func() error {
			// Initially read all files and send them to be processed and added to the in memory store.
			filenames, err := filepath.Glob(configFiles)
			if err != nil {
				return fmt.Errorf("getting files names: %w", err)
			}
			for _, f := range filenames {
				files <- f
			}
			<-ctx.Done()
			return nil
		}, func(err error) {
			cancel()
		})
	}
	{
		dir := filepath.Dir(configFiles)
		level.Info(logger).Log("msg", "watching directory for changes", "directory", dir)

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			level.Error(logger).Log("msg", "failed to create file watcher", "err", err)
			return 1
		}

		err = watcher.Add(dir)
		if err != nil {
			level.Error(logger).Log("msg", "failed to add directory to file watcher", "directory", dir, "err", err)
			return 1
		}

		gr.Add(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case event, ok := <-watcher.Events:
					if !ok {
						continue
					}
					if event.Op&fsnotify.Write == fsnotify.Write {
						files <- event.Name
					}
				case err := <-watcher.Errors:
					level.Warn(logger).Log("msg", "encountered file watcher error", "err", err)
				}
			}
		}, func(err error) {
			_ = watcher.Close()
			cancel()
		})
	}
	{
		gr.Add(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case f := <-files:
					level.Debug(logger).Log("msg", "reading", "file", f)
					reconcilesTotal.Inc()

					bytes, err := ioutil.ReadFile(f)
					if err != nil {
						reconcilesErrors.Inc()
						return fmt.Errorf("failed to read file %q: %w", f, err)
					}

					var config v1alpha1.ServiceLevelObjective
					if err := yaml.UnmarshalStrict(bytes, &config); err != nil {
						reconcilesErrors.Inc()
						return fmt.Errorf("failed to unmarshal objective %q: %w", f, err)
					}

					objective, err := config.Internal()
					if err != nil {
						reconcilesErrors.Inc()
						return fmt.Errorf("failed to get objective: %w", err)
					}

					increases, err := objective.IncreaseRules()
					if err != nil {
						reconcilesErrors.Inc()
						return fmt.Errorf("failed to get increase rules: %w", err)
					}
					burnrates, err := objective.Burnrates()
					if err != nil {
						reconcilesErrors.Inc()
						return fmt.Errorf("failed to get burn rate rules: %w", err)
					}

					rule := monitoringv1.PrometheusRuleSpec{
						Groups: []monitoringv1.RuleGroup{increases, burnrates},
					}

					bytes, err = yaml.Marshal(rule)
					if err != nil {
						reconcilesErrors.Inc()
						return fmt.Errorf("failed to marshal recording rules: %w", err)
					}

					_, file := filepath.Split(f)
					path := filepath.Join(prometheusFolder, file)

					if err := ioutil.WriteFile(path, bytes, 0o644); err != nil {
						reconcilesErrors.Inc()
						return fmt.Errorf("failed to write file %q: %w", path, err)
					}

					objectives.Set(objective)

					reload <- struct{}{} // Trigger a Prometheus reload
				}
			}
		}, func(err error) {
			cancel()
		})
	}
	{
		// This gorountine waits for reload updates and eventually calls Prometheus' reload endpoint.
		gr.Add(func() error {
			for {
				select {
				case <-ctx.Done():
					return nil
				case <-reload:
					timeout := time.After(5 * time.Second)
					for {
						select {
						case <-reload:
							// If we receive another reload within 5s we reset the timeout to 5s again.
							timeout = time.After(5 * time.Second)
						case <-timeout:
							// Eventually we trigger a reload and then start the outer loop again
							// waiting for updates or termination.
							level.Debug(logger).Log("msg", "reloading Prometheus now")
							url := promClient.URL("/-/reload", nil)
							resp, body, err := promClient.Do(ctx, &http.Request{Method: http.MethodPost, URL: url})
							if err != nil {
								level.Warn(logger).Log("msg", "failed to reload Prometheus")
								continue
							}
							if resp.StatusCode/100 != 2 {
								level.Warn(logger).Log(
									"msg", "failed to reload Prometheus",
									"url", url,
									"status", resp.Status,
									"body", body,
								)
							}
						}
					}
				}
			}
		}, func(err error) {
			cancel()
			close(reload)
		})
	}
	{
		router := openapiserver.NewRouter(
			openapiserver.NewObjectivesApiController(&FilesystemObjectiveServer{
				objectives: objectives,
			}),
		)
		router.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

		server := http.Server{Addr: ":9444", Handler: router}

		gr.Add(func() error {
			level.Info(logger).Log("msg", "starting up HTTP API", "address", server.Addr)
			return server.ListenAndServe()
		}, func(err error) {
			_ = server.Shutdown(context.Background())
		})
	}

	if err := gr.Run(); err != nil {
		level.Error(logger).Log("msg", "failed to run", "err", err)
		return 2
	}
	return 0
}

type FilesystemObjectiveServer struct {
	objectives *Objectives
}

func (f FilesystemObjectiveServer) ListObjectives(ctx context.Context, query string) (openapiserver.ImplResponse, error) {
	var matchers []*labels.Matcher
	if query != "" {
		var err error
		matchers, err = parser.ParseMetricSelector(query)
		if err != nil {
			return openapiserver.ImplResponse{Code: http.StatusBadRequest}, err
		}
	}

	objectives := f.objectives.Match(matchers)
	list := make([]openapiserver.Objective, 0, len(objectives))
	for _, o := range objectives {
		list = append(list, openapi.ServerFromInternal(o))
	}

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: list,
	}, nil
}

func (f FilesystemObjectiveServer) GetObjective(ctx context.Context, expr string) (openapiserver.ImplResponse, error) {
	// TODO: Parse expr to match properly

	f.objectives.mu.RLock()
	objective, ok := f.objectives.objectives[expr]
	f.objectives.mu.RUnlock()
	if !ok {
		return openapiserver.ImplResponse{Code: http.StatusNotFound}, nil
	}

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: openapi.ServerFromInternal(objective),
	}, nil
}

func (f FilesystemObjectiveServer) GetMultiBurnrateAlerts(ctx context.Context, expr, grouping string, inactive, current bool) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, errEndpointNotImplemented
}

func (f FilesystemObjectiveServer) GetObjectiveErrorBudget(ctx context.Context, expr, grouping string, i, i2 int32) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, errEndpointNotImplemented
}

func (f FilesystemObjectiveServer) GetObjectiveStatus(ctx context.Context, expr, grouping string) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, errEndpointNotImplemented
}

func (f FilesystemObjectiveServer) GetREDRequests(ctx context.Context, expr, grouping string, i, i2 int32) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, errEndpointNotImplemented
}

func (f FilesystemObjectiveServer) GetREDErrors(ctx context.Context, expr, grouping string, i, i2 int32) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, errEndpointNotImplemented
}
