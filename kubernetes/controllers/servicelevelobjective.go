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

package controllers

import (
	"context"
	"fmt"
	"time"

	vmv1beta1 "github.com/VictoriaMetrics/operator/api/operator/v1beta1"
	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/rulefmt"
	yamlv3 "gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/yaml"

	pyrrav1alpha1 "github.com/pyrra-dev/pyrra/kubernetes/api/v1alpha1"
	"github.com/pyrra-dev/pyrra/mimir"
	"github.com/pyrra-dev/pyrra/slo"
)

// ServiceLevelObjectiveReconciler reconciles a ServiceLevelObjective object.
type ServiceLevelObjectiveReconciler struct {
	client.Client
	MimirClient             *mimir.Client
	MimirWriteAlertingRules bool
	Logger                  kitlog.Logger
	Scheme                  *runtime.Scheme
	ConfigMapMode           bool
	VictoriaMetricsMode     bool
	GenericRules            bool
}

// +kubebuilder:rbac:groups=pyrra.dev,resources=servicelevelobjectives,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=pyrra.dev,resources=servicelevelobjectives/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=pyrra.dev,resources=servicelevelobjectives/finalizers,verbs=update
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheusrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheusrules/status,verbs=get

func (r *ServiceLevelObjectiveReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := kitlog.With(r.Logger, "reconciler", "servicelevelobjective", "namespace", req.NamespacedName)
	level.Debug(logger).Log("msg", "reconciling")

	var slo pyrrav1alpha1.ServiceLevelObjective
	if err := r.Get(ctx, req.NamespacedName, &slo); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(fmt.Errorf("getting SLO: %w", err))
	}

	if r.ConfigMapMode {
		return r.reconcileConfigMap(ctx, logger, req, slo)
	}

	if r.MimirClient != nil {
		mimirFinalizer := "mimir.servicelevelobjective.pyrra.dev/finalizer"
		if slo.ObjectMeta.DeletionTimestamp.IsZero() {
			// slo is not being deleted, add our finalizer if not already present
			if !controllerutil.ContainsFinalizer(&slo, mimirFinalizer) {
				controllerutil.AddFinalizer(&slo, mimirFinalizer)
				if err := r.Update(ctx, &slo); err != nil {
					return ctrl.Result{}, err
				}
			}
		} else {
			// slo is being deleted
			if controllerutil.ContainsFinalizer(&slo, mimirFinalizer) {
				level.Info(logger).Log("msg", "deleting mimir rule group", "name", slo.GetName())
				if err := r.deleteMimirRuleGroup(ctx, slo); err != nil {
					return ctrl.Result{}, err
				}

				// remove finalizer
				controllerutil.RemoveFinalizer(&slo, mimirFinalizer)
				if err := r.Update(ctx, &slo); err != nil {
					return ctrl.Result{}, err
				}
			}
			// Stop reconciliation as the item is being deleted
			return ctrl.Result{}, nil
		}

		return r.reconcileMimirRuleGroup(ctx, logger, slo)
	}

	if r.VictoriaMetricsMode {
		return r.reconcileVictoriaMetricsRule(ctx, logger, req, slo)
	}

	return r.reconcilePrometheusRule(ctx, logger, req, slo)
}

func (r *ServiceLevelObjectiveReconciler) reconcilePrometheusRule(ctx context.Context, logger kitlog.Logger, req ctrl.Request, kubeObjective pyrrav1alpha1.ServiceLevelObjective) (ctrl.Result, error) {
	newRule, err := makePrometheusRule(kubeObjective, r.GenericRules)
	if err != nil {
		return ctrl.Result{}, err
	}

	var rule monitoringv1.PrometheusRule
	if err := r.Get(ctx, req.NamespacedName, &rule); err != nil {
		if errors.IsNotFound(err) {
			level.Info(logger).Log("msg", "creating prometheus rule", "namespace", rule.GetNamespace(), "name", rule.GetName())
			if err := r.Create(ctx, newRule); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, fmt.Errorf("failed to get prometheus rule: %w", err)
		}
	}

	newRule.ResourceVersion = rule.ResourceVersion

	level.Info(logger).Log("msg", "updating prometheus rule", "namespace", rule.GetNamespace(), "name", rule.GetName())
	if err := r.Update(ctx, newRule); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update prometheus rule: %w", err)
	}

	kubeObjective.Status.Type = "PrometheusRule"
	if err := r.Status().Update(ctx, &kubeObjective); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ServiceLevelObjectiveReconciler) reconcileVictoriaMetricsRule(ctx context.Context, logger kitlog.Logger, req ctrl.Request, kubeObjective pyrrav1alpha1.ServiceLevelObjective) (ctrl.Result, error) {
	newPromRule, err := makePrometheusRule(kubeObjective, r.GenericRules)
	if err != nil {
		return ctrl.Result{}, err
	}
	var rule vmv1beta1.VMRule
	newRule, err := convertPrometheusRuleToVictoriaMetricsRule(*newPromRule)
	if err != nil {
		logger.Log("msg", "failed to convert prometheus rule to victoria metrics rule", "err", err)
		return ctrl.Result{}, err
	}

	if err := r.Get(ctx, req.NamespacedName, &rule); err != nil {
		if errors.IsNotFound(err) {
			level.Info(logger).Log("msg", "creating victoria metrics rule", "namespace", rule.GetNamespace(), "name", rule.GetName())
			if err := r.Create(ctx, newRule); err != nil {
				return ctrl.Result{}, err
			}
		} else {
			return ctrl.Result{}, fmt.Errorf("failed to get victoriametrics rule: %w", err)
		}
	}
	newRule.ResourceVersion = rule.ResourceVersion

	level.Info(logger).Log("msg", "updating victoriametrics rule", "namespace", rule.GetNamespace(), "name", rule.GetName())
	if err := r.Update(ctx, newRule); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update victoriametrics rule: %w", err)
	}

	kubeObjective.Status.Type = "VictoriaMetricsRule"
	if err := r.Status().Update(ctx, &kubeObjective); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ServiceLevelObjectiveReconciler) reconcileMimirRuleGroup(ctx context.Context, logger kitlog.Logger, kubeObjective pyrrav1alpha1.ServiceLevelObjective) (ctrl.Result, error) {
	newRuleGroup, err := makeMimirRuleGroup(kubeObjective, r.GenericRules, r.MimirWriteAlertingRules)
	if err != nil {
		return ctrl.Result{}, err
	}

	level.Info(logger).Log("msg", "updating mimir rule", "name", newRuleGroup.Name)

	err = r.MimirClient.SetRuleGroup(ctx, kubeObjective.GetName(), *newRuleGroup)
	if err != nil {
		return ctrl.Result{}, err
	}

	kubeObjective.Status.Type = "MimirRule"
	if err := r.Status().Update(ctx, &kubeObjective); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ServiceLevelObjectiveReconciler) deleteMimirRuleGroup(ctx context.Context, kubeObjective pyrrav1alpha1.ServiceLevelObjective) error {
	return r.MimirClient.DeleteNamespace(ctx, kubeObjective.GetName())
}

func (r *ServiceLevelObjectiveReconciler) reconcileConfigMap(
	ctx context.Context,
	logger kitlog.Logger,
	req ctrl.Request,
	kubeObjective pyrrav1alpha1.ServiceLevelObjective,
) (ctrl.Result, error) {
	name := fmt.Sprintf("pyrra-recording-rule-%s", kubeObjective.GetName())

	newConfigMap, err := makeConfigMap(name, kubeObjective, r.GenericRules)
	if err != nil {
		return ctrl.Result{}, err
	}

	var existingConfigMap corev1.ConfigMap
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      name,
	}, &existingConfigMap); err != nil {
		if errors.IsNotFound(err) {
			level.Info(logger).Log("msg", "creating config map", "namespace", newConfigMap.GetNamespace(), "name", newConfigMap.GetName())
			if err := r.Create(ctx, newConfigMap); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to create config map: %w", err)
			}
		} else {
			return ctrl.Result{}, fmt.Errorf("failed to get config map: %w", err)
		}
	}

	newConfigMap.ResourceVersion = existingConfigMap.ResourceVersion

	level.Info(logger).Log("msg", "updating config map", "namespace", newConfigMap.GetNamespace(), "name", newConfigMap.GetName())
	if err := r.Update(ctx, newConfigMap); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update config map: %w", err)
	}

	kubeObjective.Status.Type = "ConfigMap"
	if err := r.Status().Update(ctx, &kubeObjective); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ServiceLevelObjectiveReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pyrrav1alpha1.ServiceLevelObjective{}).
		Complete(r)
}

func (r *ServiceLevelObjectiveReconciler) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&pyrrav1alpha1.ServiceLevelObjective{}).
		Complete()
}

func makeConfigMap(name string, kubeObjective pyrrav1alpha1.ServiceLevelObjective, genericRules bool) (*corev1.ConfigMap, error) {
	objective, err := kubeObjective.Internal()
	if err != nil {
		return nil, fmt.Errorf("failed to get objective: %w", err)
	}

	increases, err := objective.IncreaseRules()
	if err != nil {
		return nil, fmt.Errorf("failed to get increase rules: %w", err)
	}
	burnrates, err := objective.Burnrates()
	if err != nil {
		return nil, fmt.Errorf("failed to get burn rate rules: %w", err)
	}

	rule := monitoringv1.PrometheusRuleSpec{
		Groups: []monitoringv1.RuleGroup{increases, burnrates},
	}

	if genericRules {
		rules, err := objective.GenericRules()
		if err != nil {
			if err != slo.ErrGroupingUnsupported {
				return nil, fmt.Errorf("failed to get generic rules: %w", err)
			}
			// ignore these rules
		} else {
			rule.Groups = append(rule.Groups, rules)
		}
	}

	for i := range rule.Groups {
		rule.Groups[i].PartialResponseStrategy = kubeObjective.Spec.PartialResponseStrategy
	}

	bytes, err := yaml.Marshal(rule)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recording rule: %w", err)
	}

	data := map[string]string{
		fmt.Sprintf("%s.rules.yaml", name): string(bytes),
	}

	isController := true
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: kubeObjective.GetNamespace(),
			Labels:    kubeObjective.GetLabels(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: kubeObjective.APIVersion,
					Kind:       kubeObjective.Kind,
					Name:       kubeObjective.Name,
					UID:        kubeObjective.UID,
					Controller: &isController,
				},
			},
		},
		Data: data,
	}, nil
}

func makeMimirRuleGroup(kubeObjective pyrrav1alpha1.ServiceLevelObjective, genericRules, writeAlertingRules bool) (*rulefmt.RuleGroup, error) {
	objective, err := kubeObjective.Internal()
	if err != nil {
		return nil, fmt.Errorf("failed to get objective: %w", err)
	}

	increases, err := objective.IncreaseRules()
	if err != nil {
		return nil, fmt.Errorf("failed to get increase rules: %w", err)
	}
	increasesMimirRules := prometheusRulesToMimirRules(increases.Rules, writeAlertingRules)

	burnrates, err := objective.Burnrates()
	if err != nil {
		return nil, fmt.Errorf("failed to get burn rate rules: %w", err)
	}
	burnratesMimirRules := prometheusRulesToMimirRules(burnrates.Rules, writeAlertingRules)

	genericMimirRules := []rulefmt.RuleNode{}
	if genericRules {
		rules, err := objective.GenericRules()
		if err != nil {
			if err != slo.ErrGroupingUnsupported {
				return nil, fmt.Errorf("failed to get generic rules: %w", err)
			}
		} else {
			genericMimirRules = append(genericMimirRules, prometheusRulesToMimirRules(rules.Rules, writeAlertingRules)...)
		}
	}

	combinedRules := make([]rulefmt.RuleNode, len(increasesMimirRules)+len(burnratesMimirRules)+len(genericMimirRules))
	i := 0
	for _, r := range increasesMimirRules {
		combinedRules[i] = r
		i++
	}
	for _, r := range burnratesMimirRules {
		combinedRules[i] = r
		i++
	}
	for _, r := range genericMimirRules {
		combinedRules[i] = r
		i++
	}

	return &rulefmt.RuleGroup{
		Name:     kubeObjective.GetName(),
		Interval: model.Duration(time.Second * 30),
		Rules:    combinedRules,
	}, nil
}

func prometheusRuleToMimirRuleNode(promRule monitoringv1.Rule) rulefmt.RuleNode {
	if promRule.Alert != "" {
		forVal := time.Minute * 5
		if promRule.For != nil {
			// TODO
			forVal, _ = time.ParseDuration(string(*promRule.For))
		}
		return rulefmt.RuleNode{
			Expr:        yamlStringNode(promRule.Expr.String()),
			Alert:       yamlStringNode(promRule.Alert),
			For:         model.Duration(forVal),
			Labels:      promRule.Labels,
			Annotations: promRule.Annotations,
		}
	}
	return rulefmt.RuleNode{
		Expr:   yamlStringNode(promRule.Expr.String()),
		Record: yamlStringNode(promRule.Record),
		Labels: promRule.Labels,
	}
}

func yamlStringNode(val string) yamlv3.Node {
	n := yamlv3.Node{}
	n.SetString(val)
	return n
}

func prometheusRulesToMimirRules(promRules []monitoringv1.Rule, writeAlertingRules bool) []rulefmt.RuleNode {
	rules := []rulefmt.RuleNode{}
	for _, r := range promRules {
		if r.Alert != "" && !writeAlertingRules {
			continue
		}
		rules = append(rules, prometheusRuleToMimirRuleNode(r))
	}

	return rules
}

func makePrometheusRule(kubeObjective pyrrav1alpha1.ServiceLevelObjective, genericRules bool) (*monitoringv1.PrometheusRule, error) {
	objective, err := kubeObjective.Internal()
	if err != nil {
		return nil, fmt.Errorf("failed to get objective: %w", err)
	}

	increases, err := objective.IncreaseRules()
	if err != nil {
		return nil, fmt.Errorf("failed to get increase rules: %w", err)
	}
	burnrates, err := objective.Burnrates()
	if err != nil {
		return nil, fmt.Errorf("failed to get burn rate rules: %w", err)
	}

	rule := monitoringv1.PrometheusRuleSpec{
		Groups: []monitoringv1.RuleGroup{increases, burnrates},
	}

	if genericRules {
		rules, err := objective.GenericRules()
		if err != nil {
			if err != slo.ErrGroupingUnsupported {
				return nil, fmt.Errorf("failed to get generic rules: %w", err)
			}
			// ignore these rules
		} else {
			rule.Groups = append(rule.Groups, rules)
		}
	}

	for i := range rule.Groups {
		rule.Groups[i].PartialResponseStrategy = kubeObjective.Spec.PartialResponseStrategy
	}

	isController := true
	return &monitoringv1.PrometheusRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       monitoringv1.PrometheusRuleKind,
			APIVersion: monitoring.GroupName + "/" + monitoringv1.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      kubeObjective.GetName(),
			Namespace: kubeObjective.GetNamespace(),
			Labels:    kubeObjective.GetLabels(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: kubeObjective.APIVersion,
					Kind:       kubeObjective.Kind,
					Name:       kubeObjective.Name,
					UID:        kubeObjective.UID,
					Controller: &isController,
				},
			},
		},
		Spec: rule,
	}, nil
}

// convert a PrometheusRule to a VictoriaMetricsRule
// heavily inspired by the VictoriaMetrics operator convert code
// https://github.com/VictoriaMetrics/operator/blob/master/internal/controller/operator/converter/apis.go
func convertPrometheusRuleToVictoriaMetricsRule(promRule monitoringv1.PrometheusRule) (*vmv1beta1.VMRule, error) {
	ruleGroups := make([]vmv1beta1.RuleGroup, 0, len(promRule.Spec.Groups))
	for _, promGroup := range promRule.Spec.Groups {
		ruleItems := make([]vmv1beta1.Rule, 0, len(promGroup.Rules))
		for _, promRuleItem := range promGroup.Rules {
			trule := vmv1beta1.Rule{
				Labels:      promRuleItem.Labels,
				Annotations: promRuleItem.Annotations,
				Expr:        promRuleItem.Expr.String(),
				Record:      promRuleItem.Record,
				Alert:       promRuleItem.Alert,
			}
			if promRuleItem.For != nil {
				trule.For = string(*promRuleItem.For)
			}
			ruleItems = append(ruleItems, trule)
		}

		tgroup := vmv1beta1.RuleGroup{
			Name:  promGroup.Name,
			Rules: ruleItems,
		}
		if promGroup.Interval != nil {
			tgroup.Interval = string(*promGroup.Interval)
		}
		ruleGroups = append(ruleGroups, tgroup)
	}
	vm := &vmv1beta1.VMRule{
		ObjectMeta: promRule.ObjectMeta,
		Spec: vmv1beta1.VMRuleSpec{
			Groups: ruleGroups,
		},
	}
	return vm, nil
}
