package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pyrra-dev/pyrra/slo"
)

func writeAzRuleFile(logger log.Logger, file, prometheusFolder string, genericRules, operatorRule bool) error {
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

	bytes, err := json.Marshal(rule)
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

		bytes, err = json.Marshal(monv1rule)
		if err != nil {
			return fmt.Errorf("failed to marshal rules: %w", err)
		}
	}

	_, fileName := filepath.Split(file)

	fileNameWithoutExtension := fileName[:len(fileName)-len(filepath.Ext(fileName))]
	fileName = fileNameWithoutExtension + ".json"

	path := filepath.Join(prometheusFolder, fileName)

	if err := os.WriteFile(path, bytes, 0o644); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}
	return nil
}
