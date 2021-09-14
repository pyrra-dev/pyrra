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
	"fmt"
	"net/http"
	"os"

	"github.com/oklog/run"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	pyrrav1alpha1 "github.com/pyrra-dev/pyrra/kubernetes/api/v1alpha1"
	"github.com/pyrra-dev/pyrra/kubernetes/controllers"
	"github.com/pyrra-dev/pyrra/openapi"
	openapiserver "github.com/pyrra-dev/pyrra/openapi/server/go"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = pyrrav1alpha1.AddToScheme(scheme)
	_ = monitoringv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func cmdKubernetes(metricsAddr string, namespace string) {
	setupLog := ctrl.Log.WithName("setup")
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     false,
		LeaderElectionID:   "9d76195a.metalmatze.de",
		Namespace:          namespace,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.ServiceLevelObjectiveReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("ServiceLevelObjective"),
		Scheme: mgr.GetScheme(),
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
		}, func(err error) {
		})
	}

	if err := gr.Run(); err != nil {
		setupLog.Error(err, "failed to run groups")
		return
	}
}

type ObjectiveServer struct {
	client client.Client
}

func (o *ObjectiveServer) ListObjectives(ctx context.Context) (openapiserver.ImplResponse, error) {
	var list pyrrav1alpha1.ServiceLevelObjectiveList
	if err := o.client.List(context.Background(), &list); err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	objectives := make([]openapiserver.Objective, 0, len(list.Items))
	for _, s := range list.Items {
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

func (o *ObjectiveServer) GetObjective(ctx context.Context, namespace, name string) (openapiserver.ImplResponse, error) {
	var slo pyrrav1alpha1.ServiceLevelObjective
	err := o.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &slo)
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	objective, err := slo.Internal()
	if err != nil {
		return openapiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}

	return openapiserver.ImplResponse{
		Code: http.StatusOK,
		Body: openapi.ServerFromInternal(objective),
	}, nil
}

func (o *ObjectiveServer) GetMultiBurnrateAlerts(ctx context.Context, namespace, name string) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, fmt.Errorf("endpoint not implement")
}

func (o *ObjectiveServer) GetObjectiveErrorBudget(ctx context.Context, namespace, name string, i int32, i2 int32) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, fmt.Errorf("endpoint not implement")
}

func (o *ObjectiveServer) GetObjectiveStatus(ctx context.Context, namespace, name string) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, fmt.Errorf("endpoint not implement")
}

func (o *ObjectiveServer) GetREDErrors(ctx context.Context, namespace, name string, i int32, i2 int32) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, fmt.Errorf("endpoint not implement")
}

func (o *ObjectiveServer) GetREDRequests(ctx context.Context, namespace, name string, i int32, i2 int32) (openapiserver.ImplResponse, error) {
	return openapiserver.ImplResponse{}, fmt.Errorf("endpoint not implement")
}
