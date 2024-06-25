package main

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/alertsmanagement/armalertsmanagement"

	"github.com/pyrra-dev/pyrra/slo"
)

func writeAzManagedRule(logger log.Logger, file, subscriptionId, azureRegion, resourceGroupName, clusterName, azureMonitorWorkspace, actionGroupId string, genericRules bool) error {
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

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("failed to authenticate with Azure. Use az login: %w", err)
	}

	clientFactory, err := armalertsmanagement.NewClientFactory(subscriptionId, cred, nil)
	if err != nil {
		return fmt.Errorf("failed to Azure Managed Prometheus client: %w", err)
	}

	client := clientFactory.NewPrometheusRuleGroupsClient()

	ctx := context.Background()

	for _, ruleGroup := range rule.Groups {
		ruleCount := len(ruleGroup.Rules)
		rules := make([]*armalertsmanagement.PrometheusRule, ruleCount)
		for i, r := range ruleGroup.Rules {
			severity := int32(3)
			rules[i] = &armalertsmanagement.PrometheusRule{
				Expression:  to.Ptr(r.Expr.String()),
				Alert:       &r.Alert,
				Record:      &r.Record,
				Annotations: CopyStringMapToStringPtrMap(&r.Annotations),
				Enabled:     to.Ptr(true),
				For:         to.Ptr("PT10M"), //(*string)(r.For),
				Labels:      CopyStringMapToStringPtrMap(&r.Labels),
				Severity:    to.Ptr(severity),
				Actions:     []*armalertsmanagement.PrometheusRuleGroupAction{{ActionGroupID: &actionGroupId}},
				ResolveConfiguration: &armalertsmanagement.PrometheusRuleResolveConfiguration{
					AutoResolved:  to.Ptr(true),
					TimeToResolve: to.Ptr("PT10M"),
				},
			}
		}

		options := armalertsmanagement.PrometheusRuleGroupsClientCreateOrUpdateOptions{}
		properties := armalertsmanagement.PrometheusRuleGroupProperties{
			ClusterName: &clusterName,
			Enabled:     to.Ptr(true),
			Description: to.Ptr(ruleGroup.Name),
			Interval:    to.Ptr("PT10M"),
			Scopes:      []*string{&azureMonitorWorkspace},
			Rules:       rules,
		}
		parameters := armalertsmanagement.PrometheusRuleGroupResource{
			Location:   &azureRegion,
			Tags:       map[string]*string{},
			Properties: &properties,
		}

		result, err := client.CreateOrUpdate(ctx, resourceGroupName, ruleGroup.Name, parameters, &options)
		if err != nil {
			return fmt.Errorf("failed to create or update Azure Managed Prometheus rule group: %w", err)
		}
		logger.Log(result)
	}

	return nil
}

func CopyStringMapToStringPtrMap(input *map[string]string) map[string]*string {

	output := make(map[string]*string, len(*input))
	for key, value := range *input {
		output[key] = &value
	}

	return output
}
