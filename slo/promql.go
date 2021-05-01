package slo

import (
	"fmt"

	"github.com/prometheus/common/model"
)

// QueryTotal returns a PromQL query to get the total amount of requests served during the window.
func (o Objective) QueryTotal(window model.Duration) string {
	if o.Indicator.HTTP != nil {
		http := o.Indicator.HTTP
		if http.Metric == "" {
			http.Metric = HTTPDefaultMetric
		}
		return fmt.Sprintf("sum(increase(%s{%s}[%s]))", o.Indicator.HTTP.Metric, o.Indicator.HTTP.Selectors.String(), window)
	}
	return ""
}

// QueryErrors returns a PromQL query to get the amount of request errors during the window.
func (o Objective) QueryErrors(window model.Duration) string {
	if o.Indicator.HTTP != nil {
		http := o.Indicator.HTTP
		if http.Metric == "" {
			http.Metric = HTTPDefaultMetric
		}
		if len(http.ErrorSelectors) == 0 {
			http.ErrorSelectors = Selectors{HTTPDefaultErrorSelector}
		}
		return fmt.Sprintf("sum(increase(%s{%s}[%s]))", http.Metric, http.AllSelectors(), window)
	}
	return ""
}

func (o Objective) QueryErrorBudget() string {
	if o.Indicator.HTTP != nil {
		http := o.Indicator.HTTP
		if http.Metric == "" {
			http.Metric = HTTPDefaultMetric
		}
		if len(http.ErrorSelectors) == 0 {
			http.ErrorSelectors = Selectors{HTTPDefaultErrorSelector}
		}

		budget := fmt.Sprintf(`(1 - %.3f)`, o.Target)
		errors := fmt.Sprintf(`sum(increase(%s{%s}[%s]))`, http.Metric, http.AllSelectors(), o.Window)
		total := fmt.Sprintf(`sum(increase(%s{%s}[%s]))`, http.Metric, http.Selectors.String(), o.Window)
		return fmt.Sprintf(`(%s - (%s / %s)) / %s`, budget, errors, total, budget)
	}
	if o.Indicator.GRPC != nil {
		grpc := o.Indicator.GRPC
		if grpc.Metric == "" {
			grpc.Metric = GRPCDefaultMetric
		}
		if len(grpc.ErrorSelectors) == 0 {
			grpc.ErrorSelectors = Selectors{GRPCDefaultErrorSelector}
		}

		budget := fmt.Sprintf(`(1 - %.3f)`, o.Target)
		errors := fmt.Sprintf(`sum(increase(%s{%s}[%s]))`, grpc.Metric, grpc.AllSelectors(), o.Window)
		total := fmt.Sprintf(`sum(increase(%s{%s}[%s]))`, grpc.Metric, grpc.GRPCSelectors(), o.Window)
		return fmt.Sprintf(`(%s - (%s / %s)) / %s`, budget, errors, total, budget)
	}

	return ""
}
