package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/oklog/run"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pyrra-dev/pyrra/kubernetes/api/v1alpha1"
	"github.com/pyrra-dev/pyrra/openapi"
	openapiserver "github.com/pyrra-dev/pyrra/openapi/server/go"
	"github.com/pyrra-dev/pyrra/slo"
	"sigs.k8s.io/yaml"
)

var objectives = map[string]slo.Objective{}

func cmdFilesystem(configFiles, prometheusFolder string) {
	var gr run.Group

	reg := prometheus.NewRegistry()

	reconcilesTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pyrra_filesystem_reconciles_total",
		Help: "The total amount of reconciles.",
	})
	reconcilesErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "pyrra_filesystem_reconciles_errors_total",
		Help: "The total amount of errors during reconciles.",
	})

	reg.MustRegister(
		prometheus.NewGoCollector(),
		reconcilesTotal,
		reconcilesErrors,
	)

	ctx, cancel := context.WithCancel(context.Background())
	files := make(chan string, 16)

	{
		gr.Add(func() error {
			// Initially read all files and send them to be processed and added to the in memory store.
			filenames, err := filepath.Glob(configFiles)
			if err != nil {
				return err
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
		log.Println("watching directory for changes", dir)

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}

		err = watcher.Add(dir)
		if err != nil {
			log.Fatal(err)
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
					log.Println("err", err)
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
					reconcilesTotal.Inc()
					log.Println("reading", f)
					bytes, err := ioutil.ReadFile(f)
					if err != nil {
						reconcilesErrors.Inc()
						return err
					}
					var config v1alpha1.ServiceLevelObjective
					if err := yaml.UnmarshalStrict(bytes, &config); err != nil {
						reconcilesErrors.Inc()
						return err
					}
					objective, err := config.Internal()
					if err != nil {
						reconcilesErrors.Inc()
						return err
					}

					burnrates, err := objective.Burnrates()
					if err != nil {
						reconcilesErrors.Inc()
						return err
					}

					rule := monitoringv1.PrometheusRuleSpec{
						Groups: []monitoringv1.RuleGroup{burnrates},
					}

					bytes, err = yaml.Marshal(rule)
					if err != nil {
						reconcilesErrors.Inc()
						return err
					}

					_, file := filepath.Split(f)
					if err := ioutil.WriteFile(filepath.Join(prometheusFolder, file), bytes, 0644); err != nil {
						reconcilesErrors.Inc()
						return err
					}

					objectives[objective.Name] = objective
				}
			}
		}, func(err error) {
			cancel()
		})
	}
	{
		router := openapiserver.NewRouter(
			openapiserver.NewObjectivesApiController(&FilesystemObjectiveServer{}),
		)
		router.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

		server := http.Server{Addr: ":9444", Handler: router}

		gr.Add(func() error {
			log.Println("Starting up HTTP API on", server.Addr)
			return server.ListenAndServe()
		}, func(err error) {
			_ = server.Shutdown(context.Background())
		})
	}

	if err := gr.Run(); err != nil {
		log.Println(err)
		return
	}
}

type FilesystemObjectiveServer struct{}

func (f FilesystemObjectiveServer) ListObjectives(ctx context.Context) (openapiserver.ImplResponse, error) {
	list := make([]openapiserver.Objective, 0, len(objectives))
	for _, objective := range objectives {
		list = append(list, openapi.ServerFromInternal(objective))
	}

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: list,
	}, nil
}

func (f FilesystemObjectiveServer) GetObjective(ctx context.Context, namespace, name string) (openapiserver.ImplResponse, error) {
	objective, ok := objectives[name]
	if !ok {
		return openapiserver.ImplResponse{Code: http.StatusNotFound}, nil
	}

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: openapi.ServerFromInternal(objective),
	}, nil
}

func (f FilesystemObjectiveServer) GetMultiBurnrateAlerts(ctx context.Context, namespace, name string) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, fmt.Errorf("endpoint not implement")
}

func (f FilesystemObjectiveServer) GetObjectiveErrorBudget(ctx context.Context, namespace, name string, i int32, i2 int32) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, fmt.Errorf("endpoint not implement")
}

func (f FilesystemObjectiveServer) GetObjectiveStatus(ctx context.Context, namespace, name string) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, fmt.Errorf("endpoint not implement")
}

func (f FilesystemObjectiveServer) GetREDRequests(ctx context.Context, namespace, name string, i int32, i2 int32) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, fmt.Errorf("endpoint not implement")
}

func (f FilesystemObjectiveServer) GetREDErrors(ctx context.Context, namespace, name string, i int32, i2 int32) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, fmt.Errorf("endpoint not implement")
}
