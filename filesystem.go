package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/pyrra-dev/pyrra/kubernetes/api/v1alpha1"
	objectivesv1alpha1 "github.com/pyrra-dev/pyrra/proto/objectives/v1alpha1"
	"github.com/pyrra-dev/pyrra/proto/objectives/v1alpha1/objectivesv1alpha1connect"
	"github.com/pyrra-dev/pyrra/slo"
)

type SpecsTransformation struct {
	Outcome string `json:"outcome"`
	Message string `json:"message"`
}

type SpecsList struct {
	SpecsAvailable []string `json:"specsAvailable"`
	RulesGenerated []string `json:"rulesGenerated"`
}

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

func listFolderContents(folderPath string) ([]string, error) {
	var fileNames []string
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		fileNames = append(fileNames, filepath.Base(path))
		return nil
	})
	return fileNames, err
}

func listSpecsHandler(logger log.Logger, specsDir, prometheusDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		payload := &SpecsList{
			SpecsAvailable: []string{},
			RulesGenerated: []string{},
		}

		level.Info(logger).Log("msg", "listing available specs", "dir", specsDir)

		fileNames, err := listFolderContents(specsDir)
		if err != nil {
			level.Error(logger).Log("msg", "error listing available specs", "err", err)
		} else {
			payload.SpecsAvailable = append(payload.SpecsAvailable, fileNames...)
		}

		level.Info(logger).Log("msg", "listing generated rules", "dir", prometheusDir)

		fileNames, err = listFolderContents(prometheusDir)
		if err != nil {
			level.Error(logger).Log("msg", "error listing generated rules", "err", err)
		} else {
			payload.RulesGenerated = append(payload.RulesGenerated, fileNames...)
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(payload)
		if err != nil {
			level.Error(logger).Log("msg", "failed to encode payload")
		}
	}
}

func createSpecHandler(logger log.Logger, dir, prometheusFolder string, reload chan struct{}, genericRules bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodPost {
			response := SpecsTransformation{Outcome: "error", Message: "Method not allowed"}
			responseAsJSON, _ := json.Marshal(response)
			http.Error(w, string(responseAsJSON), http.StatusMethodNotAllowed)
			return
		}

		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			response := SpecsTransformation{Outcome: "error", Message: "Failed to parse form upload"}
			responseAsJSON, _ := json.Marshal(response)
			http.Error(w, string(responseAsJSON), http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("spec")
		if err != nil {
			response := SpecsTransformation{Outcome: "error", Message: "Failed to read spec field from form upload"}
			responseAsJSON, _ := json.Marshal(response)
			http.Error(w, string(responseAsJSON), http.StatusBadRequest)
			return
		}

		level.Info(logger).Log("msg", "processing SLO spec", "specFile", handler.Filename)

		ingestedSpec := dir + "/" + handler.Filename

		f, err := os.OpenFile(ingestedSpec, os.O_WRONLY|os.O_CREATE, 0o666)
		if err != nil {
			level.Error(logger).Log("msg", "failed to write spec to disk", "err", err)
			response := SpecsTransformation{Outcome: "error", Message: "There was an error writing the spec to disk"}
			responseAsJSON, _ := json.Marshal(response)
			http.Error(w, string(responseAsJSON), http.StatusInternalServerError)
			return
		}

		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			level.Error(logger).Log("msg", "failed to copy contents", "err", err)
			response := SpecsTransformation{Outcome: "error", Message: "There was an error copying file contents"}
			responseAsJSON, _ := json.Marshal(response)
			http.Error(w, string(responseAsJSON), http.StatusInternalServerError)
			return
		}

		level.Info(logger).Log("msg", "attempting to build rules from spec", "location", ingestedSpec)
		err = writeRuleFile(logger, ingestedSpec, prometheusFolder, genericRules, false)
		if err != nil {
			level.Error(logger).Log("msg", "error building rules from spec", "file", ingestedSpec, "err", err)
			response := SpecsTransformation{Outcome: "error", Message: "There was an error building rules from this spec"}
			responseAsJSON, _ := json.Marshal(response)
			http.Error(w, string(responseAsJSON), http.StatusBadRequest)

			err := os.Remove(ingestedSpec)
			if err != nil {
				level.Error(logger).Log("msg", "failed to remove spec", "file", ingestedSpec, "err", err)
			}
			return
		}

		level.Info(logger).Log("msg", "signaling Prometheus reload")
		reload <- struct{}{}

		response := SpecsTransformation{Outcome: "success", Message: "Ok"}
		responseAsJSON, _ := json.Marshal(response)
		_, err = w.Write(responseAsJSON)
		if err != nil {
			level.Error(logger).Log("msg", "failed to return http 200", "err", err)
		}
	}
}

func removeSpecHandler(logger log.Logger, dir, prometheusFolder string, reload chan struct{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != http.MethodDelete {
			response := SpecsTransformation{Outcome: "error", Message: "Method not allowed"}
			responseAsJSON, _ := json.Marshal(response)
			http.Error(w, string(responseAsJSON), http.StatusMethodNotAllowed)
			return
		}

		filePath := r.URL.Query().Get("f")
		if filePath == "" {
			response := SpecsTransformation{Outcome: "error", Message: "Missing 'f' parameter in query"}
			responseAsJSON, _ := json.Marshal(response)
			http.Error(w, string(responseAsJSON), http.StatusBadRequest)
			return
		}

		level.Info(logger).Log("msg", "removing spec", "file", filePath, "dir", dir)

		cleanPath := filepath.Clean(filePath)

		err := os.Remove(dir + "/" + cleanPath)
		if err != nil {
			level.Error(logger).Log("msg", "failed to remove file", "file", filePath, "clean", cleanPath, "err", err)
			response := SpecsTransformation{Outcome: "error", Message: "Failed to remove file"}
			responseAsJSON, _ := json.Marshal(response)
			http.Error(w, string(responseAsJSON), http.StatusInternalServerError)
			return
		}

		level.Debug(logger).Log("msg", "removing generated rules", "file", filePath, "dir", prometheusFolder)
		err = os.Remove(prometheusFolder + "/" + cleanPath)
		if err != nil {
			level.Error(logger).Log("msg", "failed to remove Prometheus rules file", "file", filePath, "clean", cleanPath, "err", err)
			response := SpecsTransformation{Outcome: "error", Message: "Failed to delete rules"}
			responseAsJSON, _ := json.Marshal(response)
			http.Error(w, string(responseAsJSON), http.StatusInternalServerError)
			return
		}

		level.Info(logger).Log("msg", "signaling Prometheus reload")
		reload <- struct{}{}

		response := SpecsTransformation{Outcome: "success", Message: "Ok"}
		responseAsJSON, _ := json.Marshal(response)
		_, err = w.Write(responseAsJSON)
		if err != nil {
			level.Error(logger).Log("msg", "failed to return http 200", "err", err)
		}
	}
}

func cmdFilesystem(logger log.Logger, reg *prometheus.Registry, promClient api.Client, configFiles, prometheusFolder string, specsAPI, genericRules bool) int {
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
		}, func(_ error) {
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
		}, func(_ error) {
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
					// We only care about watching for files with a valid yaml extension
					if filepath.Ext(f) != ".yaml" && filepath.Ext(f) != ".yml" {
						level.Warn(logger).Log("msg", "ignoring non YAML file", "file", f)
						continue
					}

					level.Debug(logger).Log("msg", "processing", "file", f)
					reconcilesTotal.Inc()

					err := writeRuleFile(logger, f, prometheusFolder, genericRules, false)
					if err != nil {
						reconcilesErrors.Inc()
						level.Error(logger).Log("msg", "error creating rule file", "file", f, "err", err)
					}

					_, objective, err := objectiveFromFile(f)
					if err != nil {
						reconcilesErrors.Inc()
						level.Error(logger).Log("msg", "failed to get objective from file", "file", f, "err", err)
					}
					objectives.Set(objective)

					reload <- struct{}{} // Trigger a Prometheus reload
				}
			}
		}, func(_ error) {
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
		}, func(_ error) {
			cancel()
			close(reload)
		})
	}
	{
		prometheusInterceptor := connectprometheus.NewInterceptor(reg)
		dir := filepath.Dir(configFiles)

		router := http.NewServeMux()
		router.Handle(objectivesv1alpha1connect.NewObjectiveBackendServiceHandler(
			&FilesystemObjectiveServer{
				objectives: objectives,
			},
			connect.WithInterceptors(prometheusInterceptor),
		))
		router.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

		if specsAPI {
			level.Info(logger).Log("msg", "specs API endpoints are enabled")
			router.HandleFunc("/specs/remove", removeSpecHandler(logger, dir, prometheusFolder, reload))
			router.HandleFunc("/specs/create", createSpecHandler(logger, dir, prometheusFolder, reload, genericRules))
			router.HandleFunc("/specs/list", listSpecsHandler(logger, dir, prometheusFolder))
		} else {
			level.Info(logger).Log("msg", "specs API endpoints are disabled")
		}

		server := http.Server{
			Addr:    ":9444",
			Handler: h2c.NewHandler(router, &http2.Server{}),
		}

		gr.Add(func() error {
			level.Info(logger).Log("msg", "starting up HTTP API", "address", server.Addr)
			return server.ListenAndServe()
		}, func(_ error) {
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

func (s *FilesystemObjectiveServer) List(_ context.Context, req *connect.Request[objectivesv1alpha1.ListRequest]) (*connect.Response[objectivesv1alpha1.ListResponse], error) {
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

func writeRuleFile(logger log.Logger, file, prometheusFolder string, genericRules, operatorRule bool) error {
	kubeObjective, objective, err := objectiveFromFile(file)
	if err != nil {
		return fmt.Errorf("failed to get objective: %w", err)
	}

	warn, err := kubeObjective.ValidateCreate()
	if len(warn) > 0 {
		for _, w := range warn {
			level.Warn(logger).Log(
				"msg", "validation warning",
				"file", file,
				"warning", w,
			)
		}
	}
	if err != nil {
		return fmt.Errorf("invalid objective: %s - %w", file, err)
	}

	increases, err := objective.IncreaseRules()
	if err != nil {
		return fmt.Errorf("failed to get increase rules: %w", err)
	}

	burnrates, err := objective.Burnrates()
	if err != nil {
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

	bytes, err := yaml.Marshal(rule)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	if operatorRule {
		monv1rule := &monitoringv1.PrometheusRule{
			TypeMeta: metav1.TypeMeta{
				Kind:       monitoringv1.PrometheusRuleKind,
				APIVersion: monitoring.GroupName + "/" + monitoringv1.Version,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      kubeObjective.GetName(),
				Namespace: kubeObjective.GetNamespace(),
				Labels:    kubeObjective.GetLabels(),
			},
			Spec: rule,
		}

		bytes, err = yaml.Marshal(monv1rule)
		if err != nil {
			return fmt.Errorf("failed to marshal rules: %w", err)
		}
	}

	_, f := filepath.Split(file)
	path := filepath.Join(prometheusFolder, f)

	if err := os.WriteFile(path, bytes, 0o644); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}
	return nil
}

func objectiveFromFile(file string) (v1alpha1.ServiceLevelObjective, slo.Objective, error) {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return v1alpha1.ServiceLevelObjective{}, slo.Objective{}, fmt.Errorf("failed to read file %q: %w", file, err)
	}

	var config v1alpha1.ServiceLevelObjective
	if err := yaml.UnmarshalStrict(bytes, &config); err != nil {
		return v1alpha1.ServiceLevelObjective{}, slo.Objective{}, fmt.Errorf("failed to unmarshal objective %q: %w", file, err)
	}

	objective, err := config.Internal()
	if err != nil {
		return v1alpha1.ServiceLevelObjective{}, slo.Objective{}, fmt.Errorf("failed to get objective %q: %w", file, err)
	}

	return config, objective, nil
}
