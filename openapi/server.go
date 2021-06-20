package openapi

import (
	"time"

	openapi "github.com/metalmatze/athene/openapi/server/go"
	"github.com/metalmatze/athene/slo"
)

func FromInternal(objective slo.Objective) openapi.Objective {
	var ratio openapi.IndicatorRatio
	if objective.Indicator.Ratio != nil {
		ratio.Total.Name = objective.Indicator.Ratio.Total.Name
		for _, m := range objective.Indicator.Ratio.Total.LabelMatchers {
			ratio.Total.Matchers = append(ratio.Total.Matchers, m.String())
		}
		ratio.Total.Metric = objective.Indicator.Ratio.Total.Metric()

		ratio.Errors.Name = objective.Indicator.Ratio.Errors.Name
		for _, m := range objective.Indicator.Ratio.Errors.LabelMatchers {
			ratio.Errors.Matchers = append(ratio.Errors.Matchers, m.String())
		}
		ratio.Errors.Metric = objective.Indicator.Ratio.Errors.Metric()
	}

	var latency openapi.IndicatorLatency
	if objective.Indicator.Latency != nil {
		latency.Total.Name = objective.Indicator.Latency.Total.Name
		for _, m := range objective.Indicator.Latency.Total.LabelMatchers {
			latency.Total.Matchers = append(latency.Total.Matchers, m.String())
		}
		latency.Total.Metric = objective.Indicator.Latency.Total.Metric()

		latency.Success.Name = objective.Indicator.Latency.Success.Name
		for _, m := range objective.Indicator.Latency.Success.LabelMatchers {
			latency.Success.Matchers = append(latency.Success.Matchers, m.String())
		}
		latency.Success.Metric = objective.Indicator.Latency.Success.Metric()
	}

	return openapi.Objective{
		Name:        objective.Name,
		Description: objective.Description,
		Target:      objective.Target,
		Window:      time.Duration(objective.Window).Milliseconds(),
		Indicator: openapi.Indicator{
			Ratio:   ratio,
			Latency: latency,
		},
	}
}
