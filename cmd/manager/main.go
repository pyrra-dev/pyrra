/*
Copyright 2021 Athene Authors.

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
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"strings"

	athenev1alpha1 "github.com/metalmatze/athene/api/v1alpha1"
	"github.com/metalmatze/athene/controllers"
	"github.com/metalmatze/athene/slo"
	"github.com/oklog/run"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
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
	_ = athenev1alpha1.AddToScheme(scheme)
	_ = monitoringv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	setupLog := ctrl.Log.WithName("setup")
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "9d76195a.metalmatze.de",
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
		mux := http.NewServeMux()
		mux.HandleFunc("/objectives", objectivesListHandler(mgr.GetClient()))
		mux.HandleFunc("/objectives/", objectiveHandler(mgr.GetClient()))
		s := http.Server{
			Addr:    ":9444",
			Handler: mux,
		}

		gr.Add(func() error {
			return s.ListenAndServe()
		}, func(err error) {
			_ = s.Shutdown(context.Background())
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

func objectiveHandler(c client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/objectives/")
		var slo athenev1alpha1.ServiceLevelObjective
		err := c.Get(r.Context(), client.ObjectKey{Namespace: "monitoring", Name: name}, &slo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		objective, err := slo.Internal()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		bytes, err := json.Marshal(objective)
		if err != nil {
			return
		}
		_, _ = w.Write(bytes)
	}
}

func objectivesListHandler(client client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var list athenev1alpha1.ServiceLevelObjectiveList
		if err := client.List(context.Background(), &list); err != nil {
			return
		}

		objectives := make([]slo.Objective, 0, len(list.Items))
		for _, s := range list.Items {
			objective, err := s.Internal()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			objectives = append(objectives, objective)
		}

		bytes, err := json.Marshal(objectives)
		if err != nil {
			return
		}
		_, _ = w.Write(bytes)
	}
}
