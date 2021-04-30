package slo

import (
	"fmt"
)

// QueryTotal returns a PromQL query to get the total amount of requests served during the window.
func (o Objective) QueryTotal() string {
	if o.Indicator.HTTP != nil {
		http := o.Indicator.HTTP
		if http.Metric == "" {
			http.Metric = HTTPDefaultMetric
		}
		return fmt.Sprintf("sum(increase(%s{%s}[%s]))", o.Indicator.HTTP.Metric, o.Indicator.HTTP.Selectors.String(), o.Window)
	}
	return ""
}

// QueryErrors returns a PromQL query to get the amount of request errors during the window.
func (o Objective) QueryErrors() string {
	if o.Indicator.HTTP != nil {
		http := o.Indicator.HTTP
		if http.Metric == "" {
			http.Metric = HTTPDefaultMetric
		}
		if len(http.ErrorSelectors) == 0 {
			http.ErrorSelectors = Selectors{HTTPDefaultErrorSelector}
		}
		return fmt.Sprintf("sum(increase(%s{%s}[%s]))", http.Metric, http.AllSelectors(), o.Window)
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

		budget := fmt.Sprintf(`(100 - %.3f)/100`, o.Target)
		errors := fmt.Sprintf(`sum(increase(%s{%s}[%s]))`, http.Metric, http.AllSelectors(), o.Window)
		total := fmt.Sprintf(`sum(increase(%s{%s}[%s]))`, http.Metric, http.Selectors.String(), o.Window)
		return fmt.Sprintf(`((%s) - (%s / %s)) / (%s)`, budget, errors, total, budget)
	}
	if o.Indicator.GRPC != nil {
		grpc := o.Indicator.GRPC
		if grpc.Metric == "" {
			grpc.Metric = GRPCDefaultMetric
		}
		if len(grpc.ErrorSelectors) == 0 {
			grpc.ErrorSelectors = Selectors{GRPCDefaultErrorSelector}
		}

		budget := fmt.Sprintf(`(100 - %.3f)/100`, o.Target)
		errors := fmt.Sprintf(`sum(increase(%s{%s}[%s]))`, grpc.Metric, grpc.AllSelectors(), o.Window)
		total := fmt.Sprintf(`sum(increase(%s{%s}[%s]))`, grpc.Metric, grpc.GRPCSelectors(), o.Window)
		return fmt.Sprintf(`((%s) - (%s / %s)) / (%s)`, budget, errors, total, budget)
	}

	return ""
}
