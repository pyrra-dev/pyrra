package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/fsnotify/fsnotify"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
	connectprometheus "github.com/polarsignals/connect-go-prometheus"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"sigs.k8s.io/yaml"

	"github.com/pyrra-dev/pyrra/kubernetes/api/v1alpha1"
	objectivesv1alpha1 "github.com/pyrra-dev/pyrra/proto/objectives/v1alpha1"
	"github.com/pyrra-dev/pyrra/proto/objectives/v1alpha1/objectivesv1alpha1connect"
	"github.com/pyrra-dev/pyrra/slo"
)

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

func cmdFilesystem(logger log.Logger, reg *prometheus.Registry, promClient api.Client, configFiles, prometheusFolder string, genericRules bool) int {
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

					bytes, err := os.ReadFile(f)
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

					if genericRules {
						rules, err := objective.GenericRules()
						if err == nil {
							rule.Groups = append(rule.Groups, rules)
						} else {
							if err != slo.ErrGroupingUnsupported {
								return fmt.Errorf("failed to get generic rules: %w", err)
							}
							level.Warn(logger).Log(
								"msg", "objective with grouping unsupported with generic rules",
								"objective", objective.Name(),
							)
						}
					}

					bytes, err = yaml.Marshal(rule)
					if err != nil {
						reconcilesErrors.Inc()
						return fmt.Errorf("failed to marshal recording rules: %w", err)
					}

					_, file := filepath.Split(f)
					path := filepath.Join(prometheusFolder, file)

					if err := os.WriteFile(path, bytes, 0o644); err != nil {
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
		prometheusInterceptor := connectprometheus.NewInterceptor(reg)

		router := http.NewServeMux()
		router.Handle(objectivesv1alpha1connect.NewObjectiveBackendServiceHandler(
			&FilesystemObjectiveServer{
				objectives: objectives,
			},
			connect.WithInterceptors(prometheusInterceptor),
		))
		router.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

		server := http.Server{
			Addr:    ":9444",
			Handler: h2c.NewHandler(router, &http2.Server{}),
		}

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

func (s *FilesystemObjectiveServer) List(ctx context.Context, req *connect.Request[objectivesv1alpha1.ListRequest]) (*connect.Response[objectivesv1alpha1.ListResponse], error) {
	var matchers []*labels.Matcher
	if expr := req.Msg.Expr; expr != "" {
		var err error
		matchers, err = parser.ParseMetricSelector(expr)
		if err != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("failed to parse expr: %w", err))
		}
	}

	matchingObjectives := s.objectives.Match(matchers)
	objectives := make([]*objectivesv1alpha1.Objective, 0, len(matchingObjectives))
	for _, o := range matchingObjectives {
		objectives = append(objectives, objectivesv1alpha1.FromInternal(o))
	}

	if len(objectives) == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("no objectives found"))
	}

	return connect.NewResponse(&objectivesv1alpha1.ListResponse{
		Objectives: objectives,
	}), nil
}
