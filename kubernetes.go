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
	"syscall"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/go-kit/log"
	"github.com/oklog/run"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	vm1beta1 "github.com/VictoriaMetrics/operator/api/operator/v1beta1"

	pyrrav1alpha1 "github.com/pyrra-dev/pyrra/kubernetes/api/v1alpha1"
	"github.com/pyrra-dev/pyrra/kubernetes/controllers"
	"github.com/pyrra-dev/pyrra/mimir"
	objectivesv1alpha1 "github.com/pyrra-dev/pyrra/proto/objectives/v1alpha1"
	"github.com/pyrra-dev/pyrra/proto/objectives/v1alpha1/objectivesv1alpha1connect"
	// +kubebuilder:scaffold:imports
)

var scheme = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = pyrrav1alpha1.AddToScheme(scheme)
	_ = monitoringv1.AddToScheme(scheme)
	_ = vm1beta1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func cmdKubernetes(
	logger log.Logger,
	metricsAddr string,
	configMapMode, victoriaMetricsMode, genericRules, disableWebhooks bool,
	certFile, privateKeyFile string,
	mimirClient *mimir.Client,
	mimirWriteAlertingRules bool,
) int {
	setupLog := ctrl.Log.WithName("setup")
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	webhookServer := webhook.NewServer(webhook.Options{Port: 9443})

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		WebhookServer:    webhookServer,
		LeaderElection:   false,
		LeaderElectionID: "9d76195a.pyrra.dev",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	reconciler := &controllers.ServiceLevelObjectiveReconciler{
		Client:                  mgr.GetClient(),
		Logger:                  log.With(logger, "controllers", "ServiceLevelObjective"),
		GenericRules:            genericRules,
		ConfigMapMode:           configMapMode,
		VictoriaMetricsMode:     victoriaMetricsMode,
		MimirClient:             mimirClient,
		MimirWriteAlertingRules: mimirWriteAlertingRules,
	}
	if err = reconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ServiceLevelObjective")
		os.Exit(1)
	}
	if !disableWebhooks {
		if err = reconciler.SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "ServiceLevelObjective")
			os.Exit(1)
		}
	}
	// +kubebuilder:scaffold:builder

	var (
		gr          run.Group
		ctx, cancel = context.WithCancel(context.Background())
	)
	gr.Add(run.SignalHandler(ctx, os.Interrupt, syscall.SIGTERM))

	{
		gr.Add(func() error {
			if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
				return fmt.Errorf("unable to set up health check: %w", err)
			}
			if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
				return fmt.Errorf("unable to set up ready check: %w", err)
			}
			setupLog.Info("starting manager")
			return mgr.Start(ctx)
		}, func(_ error) {
			cancel()
		})
	}
	{
		router := http.NewServeMux()
		router.Handle(objectivesv1alpha1connect.NewObjectiveBackendServiceHandler(&KubernetesObjectiveServer{
			client: mgr.GetClient(),
		}))

		server := http.Server{
			Addr:    ":9444",
			Handler: h2c.NewHandler(router, &http2.Server{}),
		}

		gr.Add(func() error {
			if certFile != "" && privateKeyFile != "" {
				setupLog.Info("serving with TLS", "cert", certFile, "key", privateKeyFile)
				return server.ListenAndServeTLS(certFile, privateKeyFile)
			}
			return server.ListenAndServe()
		}, func(_ error) {
			shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			_ = server.Shutdown(shutdownCtx)
		})
	}

	if err := gr.Run(); err != nil {
		if _, ok := err.(run.SignalError); ok {
			setupLog.Info("terminated", "reason", err)
			return 0
		}
		setupLog.Error(err, "failed to run groups")
		return 2
	}
	return 0
}

type KubernetesClient interface {
	List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error
}

type KubernetesObjectiveServer struct {
	client KubernetesClient
}

func (s *KubernetesObjectiveServer) List(ctx context.Context, req *connect.Request[objectivesv1alpha1.ListRequest]) (*connect.Response[objectivesv1alpha1.ListResponse], error) {
	var (
		matchers         []*labels.Matcher
		nameMatcher      *labels.Matcher
		namespaceMatcher *labels.Matcher
	)

	if req.Msg.Expr != "" {
		var err error
		matchers, err = parser.ParseMetricSelector(req.Msg.Expr)
		if err != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("failed to parse expr: %w", err))
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
	if err := s.client.List(ctx, &list, &listOpts); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	objectives := make([]*objectivesv1alpha1.Objective, 0, len(list.Items))
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
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		objectives = append(objectives, objectivesv1alpha1.FromInternal(internal))
	}

	return connect.NewResponse(&objectivesv1alpha1.ListResponse{
		Objectives: objectives,
	}), nil
}
