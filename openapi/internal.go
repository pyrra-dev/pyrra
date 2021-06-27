package openapi

import (
	"time"

	client "github.com/metalmatze/athene/openapi/client"
	"github.com/metalmatze/athene/slo"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
)

func InternalFromClient(o client.Objective) slo.Objective {
	var ratio *slo.RatioIndicator
	if o.HasIndicator() && o.Indicator.HasRatio() {
		ratio = &slo.RatioIndicator{
			Errors: slo.Metric{Name: o.Indicator.Ratio.Errors.GetName()},
			Total:  slo.Metric{Name: o.Indicator.Ratio.Total.GetName()},
		}
		for _, m := range o.Indicator.Ratio.Errors.GetMatchers() {
			ratio.Errors.LabelMatchers = append(ratio.Errors.LabelMatchers, &labels.Matcher{
				Type:  labels.MatchType(m.GetType()),
				Name:  m.GetName(),
				Value: m.GetValue(),
			})
		}
		for _, m := range o.Indicator.Ratio.Total.GetMatchers() {
			ratio.Total.LabelMatchers = append(ratio.Total.LabelMatchers, &labels.Matcher{
				Type:  labels.MatchType(m.GetType()),
				Name:  m.GetName(),
				Value: m.GetValue(),
			})
		}
	}

	var latency *slo.LatencyIndicator
	if o.HasIndicator() && o.Indicator.HasLatency() {
		latency = &slo.LatencyIndicator{
			Success: slo.Metric{Name: o.Indicator.Latency.Success.GetName()},
			Total:   slo.Metric{Name: o.Indicator.Latency.Total.GetName()},
		}
		for _, m := range o.Indicator.Latency.Success.GetMatchers() {
			latency.Success.LabelMatchers = append(latency.Success.LabelMatchers, &labels.Matcher{
				Type:  labels.MatchType(m.GetType()),
				Name:  m.GetName(),
				Value: m.GetValue(),
			})
		}
		for _, m := range o.Indicator.Latency.Total.GetMatchers() {
			latency.Total.LabelMatchers = append(latency.Total.LabelMatchers, &labels.Matcher{
				Type:  labels.MatchType(m.GetType()),
				Name:  m.GetName(),
				Value: m.GetValue(),
			})
		}
	}

	return slo.Objective{
		Name:        o.GetName(),
		Description: o.GetDescription(),
		Target:      o.GetTarget(),
		Window:      model.Duration(time.Duration(o.GetWindow()) * time.Millisecond),
		Config:      o.GetConfig(),
		Indicator: slo.Indicator{
			Ratio:   ratio,
			Latency: latency,
		},
	}
}
