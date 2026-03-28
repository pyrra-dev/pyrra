package slo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIncreaseRules_Prometheus3Migration(t *testing.T) {
	objective := objectiveHTTPLatency()

	// Test with migration disabled
	t.Run("migration_disabled", func(t *testing.T) {
		opts := GenerationOptions{EnablePrometheus3Migration: false}
		ruleGroup, err := objective.IncreaseRules(opts)
		require.NoError(t, err)

		// Find the rule with le label
		var foundRule bool
		for _, rule := range ruleGroup.Rules {
			if rule.Labels["le"] == "1" {
				foundRule = true
				// The expression should contain le="1" (not regex)
				require.Contains(t, rule.Expr.String(), `le="1"`)
				require.NotContains(t, rule.Expr.String(), `le=~`)
			}
		}
		require.True(t, foundRule, "Should find a rule with le label")
	})

	// Test with migration enabled
	t.Run("migration_enabled", func(t *testing.T) {
		opts := GenerationOptions{EnablePrometheus3Migration: true}
		ruleGroup, err := objective.IncreaseRules(opts)
		require.NoError(t, err)

		// Find the rule with le label
		var foundRule bool
		for _, rule := range ruleGroup.Rules {
			if rule.Labels["le"] == "1" {
				foundRule = true
				// The expression should contain le=~ with regex (backslashes are escaped in String())
				exprStr := rule.Expr.String()
				require.Contains(t, exprStr, `le=~"1(\\.0)?"`)
				require.NotContains(t, exprStr, `le="1"`)
			}
		}
		require.True(t, foundRule, "Should find a rule with le label")
	})
}

func TestGenericRules_Prometheus3Migration(t *testing.T) {
	objective := objectiveHTTPLatency()

	// Test with migration disabled
	t.Run("migration_disabled", func(t *testing.T) {
		opts := GenerationOptions{EnablePrometheus3Migration: false}
		ruleGroup, err := objective.GenericRules(opts)
		require.NoError(t, err)

		// Find the availability rule
		var foundRule bool
		for _, rule := range ruleGroup.Rules {
			if rule.Record == "pyrra_availability" {
				foundRule = true
				// Should contain le="1" and le="" (not regex)
				require.Contains(t, rule.Expr.String(), `le="1"`)
				require.Contains(t, rule.Expr.String(), `le=""`)
			}
		}
		require.True(t, foundRule, "Should find pyrra_availability rule")
	})

	// Test with migration enabled
	t.Run("migration_enabled", func(t *testing.T) {
		opts := GenerationOptions{EnablePrometheus3Migration: true}
		ruleGroup, err := objective.GenericRules(opts)
		require.NoError(t, err)

		// Find the availability rule
		var foundRule bool
		for _, rule := range ruleGroup.Rules {
			if rule.Record == "pyrra_availability" {
				foundRule = true
				// Should contain le=~ with regex for integer bucket (backslashes are escaped in String())
				exprStr := rule.Expr.String()
				require.Contains(t, exprStr, `le=~"1(\\.0)?"`)
			}
		}
		require.True(t, foundRule, "Should find pyrra_availability rule")
	})
}
