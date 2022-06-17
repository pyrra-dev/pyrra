/*
Copyright 2021 Pyrra Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"net/http"
	"os"

	"github.com/go-kit/log"
	"github.com/oklog/run"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	pyrrav1alpha1 "github.com/pyrra-dev/pyrra/kubernetes/api/v1alpha1"
	"github.com/pyrra-dev/pyrra/kubernetes/controllers"
	"github.com/pyrra-dev/pyrra/openapi"
	openapiserver "github.com/pyrra-dev/pyrra/openapi/server/go"
	// +kubebuilder:scaffold:imports
)

var scheme = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = pyrrav1alpha1.AddToScheme(scheme)
	_ = monitoringv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func cmdKubernetes(logger log.Logger, metricsAddr string, configMapMode bool) int {
	setupLog := ctrl.Log.WithName("setup")
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     false,
		LeaderElectionID:   "9d76195a.metalmatze.de",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.ServiceLevelObjectiveReconciler{
		Client:        mgr.GetClient(),
		Logger:        log.With(logger, "controllers", "ServiceLevelObjective"),
		Scheme:        mgr.GetScheme(),
		ConfigMapMode: configMapMode,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ServiceLevelObjective")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	var gr run.Group
	{
		router := openapiserver.NewRouter(
			openapiserver.NewObjectivesApiController(&ObjectiveServer{client: mgr.GetClient()}),
		)

		server := http.Server{Addr: ":9444", Handler: router}

		gr.Add(func() error {
			return server.ListenAndServe()
		}, func(err error) {
			_ = server.Shutdown(context.Background())
		})
	}
	{
		gr.Add(func() error {
			setupLog.Info("starting manager")
			return mgr.Start(ctrl.SetupSignalHandler())
		}, func(err error) {})
	}

	if err := gr.Run(); err != nil {
		setupLog.Error(err, "failed to run groups")
	}
	return 0
}

type KubernetesClient interface {
	List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error
}

type ObjectiveServer struct {
	client KubernetesClient
}

func (o *ObjectiveServer) ListObjectives(ctx context.Context, expr string) (openapiserver.ImplResponse, error) {
	var (
		matchers         []*labels.Matcher
		nameMatcher      *labels.Matcher
		namespaceMatcher *labels.Matcher
	)

	if expr != "" {
		var err error
		matchers, err = parser.ParseMetricSelector(expr)
		if err != nil {
			return openapiserver.ImplResponse{Code: http.StatusBadRequest}, err
		}
		for _, m := range matchers {
			if m.Name == labels.MetricName {
				nameMatcher = m
			}
			if m.Name == "namespace" {
				namespaceMatcher = m
			}
		}
	}

	listOpts := client.ListOptions{}
	for _, m := range matchers {
		if m.Name == "namespace" && m.Type == labels.MatchEqual {
			listOpts.Namespace = m.Value
			break
		}
	}

	var list pyrrav1alpha1.ServiceLevelObjectiveList
	if err := o.client.List(ctx, &list, &listOpts); err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	objectives := make([]openapiserver.Objective, 0, len(list.Items))
	for _, s := range list.Items {
		if nameMatcher != nil {
			if !nameMatcher.Matches(s.GetName()) {
				continue
			}
		}
		if namespaceMatcher != nil {
			if !namespaceMatcher.Matches(s.GetNamespace()) {
				continue
			}
		}

		internal, err := s.Internal()
		if err != nil {
			return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
		}
		objectives = append(objectives, openapi.ServerFromInternal(internal))
	}

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: objectives,
	}, nil
}

func (o *ObjectiveServer) GetMultiBurnrateAlerts(ctx context.Context, expr, grouping string, inactive, current bool) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, errEndpointNotImplemented
}

func (o *ObjectiveServer) GetObjectiveErrorBudget(ctx context.Context, expr, grouping string, start, end int32) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, errEndpointNotImplemented
}

func (o *ObjectiveServer) GetObjectiveStatus(ctx context.Context, expr, grouping string) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, errEndpointNotImplemented
}

func (o *ObjectiveServer) GetREDErrors(ctx context.Context, expr, grouping string, start, end int32) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, errEndpointNotImplemented
}

func (o *ObjectiveServer) GetREDRequests(ctx context.Context, expr, grouping string, start, end int32) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, errEndpointNotImplemented
}
