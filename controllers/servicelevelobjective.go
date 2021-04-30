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

package controllers

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-logr/logr"
	athenev1alpha1 "github.com/metalmatze/athene/api/v1alpha1"
	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/common/model"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceLevelObjectiveReconciler reconciles a ServiceLevelObjective object
type ServiceLevelObjectiveReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=athene.metalmatze.de,resources=servicelevelobjectives,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=athene.metalmatze.de,resources=servicelevelobjectives/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheusrules,verbs=create
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheusrules/status,verbs=get

func (r *ServiceLevelObjectiveReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("servicelevelobjective", req.NamespacedName)
	log.Info("reconciling")

	var slo athenev1alpha1.ServiceLevelObjective
	if err := r.Get(ctx, req.NamespacedName, &slo); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	newRule, err := makePrometheusRule(slo)
	if err != nil {
		return ctrl.Result{}, err
	}

	var rule monitoringv1.PrometheusRule
	err = r.Get(ctx, req.NamespacedName, &rule)
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}
	if errors.IsNotFound(err) {
		log.Info("creating prometheus rule", "name", rule.GetName(), "namespace", rule.GetNamespace())
		if err := r.Create(ctx, newRule); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	newRule.ResourceVersion = rule.ResourceVersion

	log.Info("updating prometheus rule", "name", rule.GetName(), "namespace", rule.GetNamespace())
	if err := r.Update(ctx, newRule); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ServiceLevelObjectiveReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&athenev1alpha1.ServiceLevelObjective{}).
		Complete(r)
}

func makePrometheusRule(slo athenev1alpha1.ServiceLevelObjective) (*monitoringv1.PrometheusRule, error) {
	var groups []monitoringv1.RuleGroup
	{
		// HTTP
		rules, err := makeHTTPRules(slo)
		if err != nil {
			return nil, err
		}
		if rules != nil {
			groups = append(groups, *rules)
		}
	}
	{
		// gRPC
		rules, err := makeGRPCRules(slo)
		if err != nil {
			return nil, err
		}
		if rules != nil {
			groups = append(groups, *rules)
		}
	}

	return &monitoringv1.PrometheusRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PrometheusRuleKind,
			APIVersion: monitoring.GroupName + "/" + monitoringv1.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      slo.GetName(),
			Namespace: slo.GetNamespace(),
			Labels:    slo.GetLabels(),
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: groups,
		},
	}, nil
}

func makeHTTPRules(slo athenev1alpha1.ServiceLevelObjective) (*monitoringv1.RuleGroup, error) {
	http := slo.Spec.ServiceLevelIndicator.HTTP
	if http == nil {
		return nil, nil
	}

	if len(http.ErrorSelectors) == 0 {
		http.ErrorSelectors = []string{`code=~"5.."`}
	}

	var rules []monitoringv1.Rule
	var err error
	if slo.Spec.Latency != "" {
		rules, err = makeHTTPLatencyRules(slo)
		if err != nil {
			return nil, err
		}
	} else {
		rules, err = makeHTTPErrorRules(slo)
		if err != nil {
			return nil, err
		}
	}

	group := &monitoringv1.RuleGroup{
		Name:     fmt.Sprintf("%s-%s", slo.GetName(), "http-rules"),
		Interval: "30s", // TODO: Increase or decrease based on availability target
		Rules:    rules,
	}
	return group, nil
}

func makeHTTPErrorRules(slo athenev1alpha1.ServiceLevelObjective) ([]monitoringv1.Rule, error) {
	http := slo.Spec.ServiceLevelIndicator.HTTP

	if http.Metric == nil {
		metric := "http_requests_total"
		http.Metric = &metric
	}

	ws := windows(slo.Spec.Window.Duration)
	burnrates := burnratesFromWindows(ws)

	rules := make([]monitoringv1.Rule, 0, len(burnrates))
	for _, br := range burnrates {
		r := monitoringv1.Rule{
			Record: burnrateName(*http.Metric, br),
			Expr:   intstr.FromString(burnrate(*http.Metric, br, http.Selectors, http.ErrorSelectors)),
			//Labels: http.Selectors, // TODO: Properly parse selectors via matchers
		}
		rules = append(rules, r)
	}

	for _, w := range ws {
		r := monitoringv1.Rule{
			Alert: "ErrorBudgetBurn",
			Expr: intstr.FromString(fmt.Sprintf("%s > (%.f * (100-%s)/100) and %s > (%.f * (100-%s)/100)",
				burnrateName(*http.Metric, w.Short),
				w.Factor,
				slo.Spec.Target,
				burnrateName(*http.Metric, w.Long),
				w.Factor,
				slo.Spec.Target,
			)),
			For: model.Duration(w.For).String(),
			Annotations: map[string]string{
				"severity": string(w.Severity),
			},
		}
		rules = append(rules, r)
	}
	return rules, nil
}

func makeHTTPLatencyRules(slo athenev1alpha1.ServiceLevelObjective) ([]monitoringv1.Rule, error) {
	const bucketSuffix = "_bucket"

	http := slo.Spec.ServiceLevelIndicator.HTTP

	if http.Metric == nil {
		metric := "http_request_duration_seconds_bucket"
		http.Metric = &metric
	}
	// Latency burnrates only work on histograms.
	if !strings.HasSuffix(*http.Metric, bucketSuffix) {
		return nil, fmt.Errorf("expected _bucket suffix on metric for latency SLO")
	}

	ws := windows(slo.Spec.Window.Duration)
	burnrates := burnratesFromWindows(ws)

	rules := make([]monitoringv1.Rule, 0, len(burnrates))
	for _, br := range burnrates {
		r := monitoringv1.Rule{
			Record: burnrateName(strings.TrimSuffix(*http.Metric, bucketSuffix), br),
			Expr:   intstr.FromString(burnrate(*http.Metric, br, http.Selectors, http.ErrorSelectors)),
			//Labels: http.Selectors, // TODO: Properly parse selectors via matchers
		}
		rules = append(rules, r)
	}
	for _, w := range ws {
		r := monitoringv1.Rule{
			Alert: "ErrorBudgetBurn",
			Expr: intstr.FromString(fmt.Sprintf("%s > (%.f * (100-%s)/100) and %s > (%.f * (100-%s)/100)",
				burnrateName(strings.TrimSuffix(*http.Metric, bucketSuffix), w.Short),
				w.Factor,
				slo.Spec.Target,
				burnrateName(strings.TrimSuffix(*http.Metric, bucketSuffix), w.Long),
				w.Factor,
				slo.Spec.Target,
			)),
			For: model.Duration(w.For).String(),
			Annotations: map[string]string{
				"severity": string(w.Severity),
			},
		}
		rules = append(rules, r)
	}

	return rules, nil
}

func makeGRPCRules(slo athenev1alpha1.ServiceLevelObjective) (*monitoringv1.RuleGroup, error) {
	grpc := slo.Spec.ServiceLevelIndicator.GRPC
	if grpc == nil {
		return nil, nil
	}

	if grpc.Metric == nil {
		metric := "grpc_server_handled_total"
		grpc.Metric = &metric
	}

	ws := windows(slo.Spec.Window.Duration)
	burnrates := burnratesFromWindows(ws)

	rules := make([]monitoringv1.Rule, 0, len(burnrates))
	for _, br := range burnrates {
		selectors := append(grpc.Selectors, []string{
			fmt.Sprintf(`grpc_method="%s"`, grpc.Method),
			fmt.Sprintf(`grpc_service="%s"`, grpc.Service),
		}...)
		r := monitoringv1.Rule{
			Record: burnrateName(*grpc.Metric, br),
			Expr:   intstr.FromString(burnrate(*grpc.Metric, br, selectors, nil)),
			//Labels: nil,
		}
		rules = append(rules, r)
	}

	return &monitoringv1.RuleGroup{
		Name:     fmt.Sprintf("%s-%s", slo.GetName(), "grpc-rules"),
		Interval: "30s", // TODO: Increase or decrease based on availability target
		Rules:    rules,
	}, nil
}

func burnrateName(metric string, rate time.Duration) string {
	metric = strings.TrimSuffix(metric, "_total")
	return fmt.Sprintf("%s:burnrate%s", metric, model.Duration(rate))
}

func burnrate(metric string, rate time.Duration, selectors []string, errorSelectors []string) string {
	errors := fmt.Sprintf("sum(rate(%s{%s}[%s]))", metric, strings.Join(append(selectors, errorSelectors...), ","), model.Duration(rate))
	total := fmt.Sprintf("sum(rate(%s{%s}[%s]))", metric, strings.Join(selectors, ","), model.Duration(rate))
	return fmt.Sprintf("%s\n/\n%s", errors, total)
}

type severity string

const critical severity = "critical"
const warning severity = "warning"

type window struct {
	Severity severity
	For      time.Duration
	Long     time.Duration
	Short    time.Duration
	Factor   float64
}

func windows(sloWindow time.Duration) []window {
	// TODO: I'm still not sure if For, Long, Short should really be based on the 28 days ratio...

	round := time.Minute // TODO: Change based on sloWindow

	// long and short rates are calculated based on the ratio for 28 days.
	return []window{{
		Severity: critical,
		For:      (sloWindow / (28 * 24 * (60 / 2))).Round(round), // 2m for 28d - half short
		Long:     (sloWindow / (28 * 24)).Round(round),            // 1h for 28d
		Short:    (sloWindow / (28 * 24 * (60 / 5))).Round(round), // 5m for 28d
		Factor:   14,                                              // error budget burn: 50% within a day
	}, {
		Severity: critical,
		For:      (sloWindow / (28 * 24 * (60 / 15))).Round(round), // 15m for 28d - half short
		Long:     (sloWindow / (28 * (24 / 6))).Round(round),       // 6h for 28d
		Short:    (sloWindow / (28 * 24 * (60 / 30))).Round(round), // 30m for 28d
		Factor:   7,                                                // error budget burn: 20% within a day / 100% within 5 days
	}, {
		Severity: warning,
		For:      (sloWindow / (28 * 24)).Round(round),       // 1h for 28d - half short
		Long:     (sloWindow / 28).Round(round),              // 1d for 28d
		Short:    (sloWindow / (28 * (24 / 2))).Round(round), // 2h for 28d
		Factor:   2,                                          // error budget burn: 10% within a day / 100% within 10 days
	}, {
		Severity: warning,
		For:      (sloWindow / (28 * (24 / 3))).Round(round), // 3h for 28d - half short
		Long:     (sloWindow / 7).Round(round),               // 4d for 28d
		Short:    (sloWindow / (28 * (24 / 6))).Round(round), // 6h for 28d
		Factor:   1,                                          // error budget burn: 100% until the end of sloWindow
	}}
}

func burnratesFromWindows(ws []window) []time.Duration {
	dedup := map[time.Duration]bool{}
	for _, w := range ws {
		dedup[w.Long] = true
		dedup[w.Short] = true
	}
	burnrates := make([]time.Duration, 0, len(dedup))
	for duration := range dedup {
		burnrates = append(burnrates, duration)
	}

	sort.Slice(burnrates, func(i, j int) bool {
		return burnrates[i].Nanoseconds() < burnrates[j].Nanoseconds()
	})

	return burnrates
}
