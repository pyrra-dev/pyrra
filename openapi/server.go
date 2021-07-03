package openapi

import (
	"time"

	client "github.com/metalmatze/athene/openapi/client"
	server "github.com/metalmatze/athene/openapi/server/go"
	"github.com/metalmatze/athene/slo"
)

func ServerFromInternal(objective slo.Objective) server.Objective {
	var ratio server.IndicatorRatio
	if objective.Indicator.Ratio != nil {
		ratio.Total.Name = objective.Indicator.Ratio.Total.Name
		for _, m := range objective.Indicator.Ratio.Total.LabelMatchers {
			ratio.Total.Matchers = append(ratio.Total.Matchers, server.QueryMatchers{
				Name:  m.Name,
				Value: m.Value,
				Type:  int32(m.Type),
			})
		}
		ratio.Total.Metric = objective.Indicator.Ratio.Total.Metric()

		ratio.Errors.Name = objective.Indicator.Ratio.Errors.Name
		for _, m := range objective.Indicator.Ratio.Errors.LabelMatchers {
			ratio.Errors.Matchers = append(ratio.Errors.Matchers, server.QueryMatchers{
				Name:  m.Name,
				Value: m.Value,
				Type:  int32(m.Type),
			})
		}
		ratio.Errors.Metric = objective.Indicator.Ratio.Errors.Metric()
	}

	var latency server.IndicatorLatency
	if objective.Indicator.Latency != nil {
		latency.Total.Name = objective.Indicator.Latency.Total.Name
		for _, m := range objective.Indicator.Latency.Total.LabelMatchers {
			latency.Total.Matchers = append(latency.Total.Matchers, server.QueryMatchers{
				Name:  m.Name,
				Value: m.Value,
				Type:  int32(m.Type),
			})
		}
		latency.Total.Metric = objective.Indicator.Latency.Total.Metric()

		latency.Success.Name = objective.Indicator.Latency.Success.Name
		for _, m := range objective.Indicator.Latency.Success.LabelMatchers {
			latency.Success.Matchers = append(latency.Success.Matchers, server.QueryMatchers{
				Name:  m.Name,
				Value: m.Value,
				Type:  int32(m.Type),
			})
		}
		latency.Success.Metric = objective.Indicator.Latency.Success.Metric()
	}

	return server.Objective{
		Name:        objective.Name,
		Namespace:   objective.Namespace,
		Description: objective.Description,
		Target:      objective.Target,
		Window:      time.Duration(objective.Window).Milliseconds(),
		Config:      objective.Config,
		Indicator: server.Indicator{
			Ratio:   ratio,
			Latency: latency,
		},
	}
}

func ServerFromClient(o client.Objective) server.Objective {
	var ratio server.IndicatorRatio
	if o.HasIndicator() && o.Indicator.HasRatio() {
		ratio.Total.Name = o.Indicator.Ratio.Total.GetName()
		ratio.Total.Metric = o.Indicator.Ratio.Total.GetMetric()
		ratio.Errors.Name = o.Indicator.Ratio.Errors.GetName()
		ratio.Errors.Metric = o.Indicator.Ratio.Errors.GetMetric()

		for _, m := range o.Indicator.Ratio.Total.GetMatchers() {
			ratio.Total.Matchers = append(ratio.Total.Matchers, server.QueryMatchers{
				Name:  m.GetName(),
				Value: m.GetValue(),
				Type:  m.GetType(),
			})
		}
		for _, m := range o.Indicator.Ratio.Errors.GetMatchers() {
			ratio.Errors.Matchers = append(ratio.Errors.Matchers, server.QueryMatchers{
				Name:  m.GetName(),
				Value: m.GetValue(),
				Type:  m.GetType(),
			})
		}
	}
	var latency server.IndicatorLatency
	if o.HasIndicator() && o.Indicator.HasLatency() {
		latency.Total.Name = o.Indicator.Latency.Total.GetName()
		latency.Total.Metric = o.Indicator.Latency.Total.GetMetric()
		latency.Success.Name = o.Indicator.Latency.Success.GetName()
		latency.Success.Metric = o.Indicator.Latency.Success.GetMetric()

		for _, m := range o.Indicator.Latency.Total.GetMatchers() {
			latency.Total.Matchers = append(latency.Total.Matchers, server.QueryMatchers{
				Name:  m.GetName(),
				Value: m.GetValue(),
				Type:  m.GetType(),
			})
		}
		for _, m := range o.Indicator.Latency.Success.GetMatchers() {
			latency.Success.Matchers = append(latency.Success.Matchers, server.QueryMatchers{
				Name:  m.GetName(),
				Value: m.GetValue(),
				Type:  m.GetType(),
			})
		}
	}

	return server.Objective{
		Name:        o.GetName(),
		Namespace:   o.GetNamespace(),
		Description: o.GetDescription(),
		Target:      o.GetTarget(),
		Window:      o.GetWindow(),
		Config:      o.GetConfig(),
		Indicator: server.Indicator{
			Ratio:   ratio,
			Latency: latency,
		},
	}
}
