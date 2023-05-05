package objectivesv1alpha1

import (
	"strings"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/util/strutil"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/pyrra-dev/pyrra/slo"
)

func ToInternal(o *Objective) slo.Objective {
	var ratio *slo.RatioIndicator
	var latency *slo.LatencyIndicator
	var latencyNative *slo.LatencyNativeIndicator
	var boolGauge *slo.BoolGaugeIndicator

	if o.Indicator != nil {
		if r := o.Indicator.GetRatio(); r != nil {
			ratio = &slo.RatioIndicator{
				Errors:   slo.Metric{Name: r.Errors.GetName()},
				Total:    slo.Metric{Name: r.Total.GetName()},
				Grouping: r.GetGrouping(),
			}
			for _, m := range r.Errors.GetMatchers() {
				ratio.Errors.LabelMatchers = append(ratio.Errors.LabelMatchers, &labels.Matcher{
					Type:  labels.MatchType(m.GetType()),
					Name:  m.GetName(),
					Value: m.GetValue(),
				})
			}
			for _, m := range r.Total.GetMatchers() {
				ratio.Total.LabelMatchers = append(ratio.Total.LabelMatchers, &labels.Matcher{
					Type:  labels.MatchType(m.GetType()),
					Name:  m.GetName(),
					Value: m.GetValue(),
				})
			}
		}

		if l := o.Indicator.GetLatency(); l != nil {
			latency = &slo.LatencyIndicator{
				Success:  slo.Metric{Name: l.Success.GetName()},
				Total:    slo.Metric{Name: l.Total.GetName()},
				Grouping: l.GetGrouping(),
			}
			for _, m := range l.Success.GetMatchers() {
				latency.Success.LabelMatchers = append(latency.Success.LabelMatchers, &labels.Matcher{
					Type:  labels.MatchType(m.GetType()),
					Name:  m.GetName(),
					Value: m.GetValue(),
				})
			}
			for _, m := range l.Total.GetMatchers() {
				latency.Total.LabelMatchers = append(latency.Total.LabelMatchers, &labels.Matcher{
					Type:  labels.MatchType(m.GetType()),
					Name:  m.GetName(),
					Value: m.GetValue(),
				})
			}
		}
		if l := o.Indicator.GetLatencyNative(); l != nil {
			latency, err := model.ParseDuration(l.Latency)
			if err != nil {
				return slo.Objective{}
			}
			latencyNative = &slo.LatencyNativeIndicator{
				Total:    slo.Metric{Name: l.Total.GetName()},
				Grouping: l.GetGrouping(),
				Latency:  latency,
			}
			for _, m := range l.Total.GetMatchers() {
				latencyNative.Total.LabelMatchers = append(latencyNative.Total.LabelMatchers, &labels.Matcher{
					Type:  labels.MatchType(m.GetType()),
					Name:  m.GetName(),
					Value: m.GetValue(),
				})
			}
		}

		if b := o.Indicator.GetBoolGauge(); b != nil {
			boolGauge = &slo.BoolGaugeIndicator{
				Metric:   slo.Metric{Name: b.BoolGauge.GetName()},
				Grouping: b.GetGrouping(),
			}
			for _, m := range b.BoolGauge.GetMatchers() {
				boolGauge.LabelMatchers = append(boolGauge.LabelMatchers, &labels.Matcher{
					Type:  labels.MatchType(m.GetType()),
					Name:  m.GetName(),
					Value: m.GetValue(),
				})
			}
		}
	}

	ls := make([]labels.Label, 0, len(o.Labels))
	for name, value := range o.Labels {
		ls = append(ls, labels.Label{Name: name, Value: value})
	}

	return slo.Objective{
		Labels:      ls,
		Description: o.Description,
		Target:      o.Target,
		Window:      model.Duration(o.Window.AsDuration()),
		Config:      o.Config,
		Alerting:    slo.Alerting{}, // TODO
		Indicator: slo.Indicator{
			Ratio:         ratio,
			Latency:       latency,
			LatencyNative: latencyNative,
			BoolGauge:     boolGauge,
		},
	}
}

func FromInternal(o slo.Objective) *Objective {
	var ratio *Ratio
	if r := o.Indicator.Ratio; r != nil {
		ratio = &Ratio{
			Grouping: o.Grouping(),
			Errors: &Query{
				Name:   r.Errors.Name,
				Metric: r.Errors.Metric(),
			},
			Total: &Query{
				Name:   r.Total.Name,
				Metric: r.Total.Metric(),
			},
		}
		for _, m := range r.Total.LabelMatchers {
			ratio.Total.Matchers = append(ratio.Total.Matchers, &LabelMatcher{
				Type:  LabelMatcher_Type(m.Type),
				Name:  m.Name,
				Value: m.Value,
			})
		}
		for _, m := range r.Errors.LabelMatchers {
			ratio.Errors.Matchers = append(ratio.Errors.Matchers, &LabelMatcher{
				Type:  LabelMatcher_Type(m.Type),
				Name:  m.Name,
				Value: m.Value,
			})
		}
	}

	var latency *Latency
	if l := o.Indicator.Latency; l != nil {
		latency = &Latency{
			Grouping: o.Grouping(),
			Success: &Query{
				Name:   l.Success.Name,
				Metric: l.Success.Metric(),
			},
			Total: &Query{
				Name:   l.Total.Name,
				Metric: l.Total.Metric(),
			},
		}
		for _, m := range l.Total.LabelMatchers {
			latency.Total.Matchers = append(latency.Total.Matchers, &LabelMatcher{
				Type:  LabelMatcher_Type(m.Type),
				Name:  m.Name,
				Value: m.Value,
			})
		}
		for _, m := range l.Success.LabelMatchers {
			latency.Success.Matchers = append(latency.Success.Matchers, &LabelMatcher{
				Type:  LabelMatcher_Type(m.Type),
				Name:  m.Name,
				Value: m.Value,
			})
		}
	}
	var latencyNative *LatencyNative
	if l := o.Indicator.LatencyNative; l != nil {
		latencyNative = &LatencyNative{
			Grouping: o.Grouping(),
			Total: &Query{
				Name:   l.Total.Name,
				Metric: l.Total.Metric(),
			},
			Latency: l.Latency.String(),
		}
		for _, m := range l.Total.LabelMatchers {
			latencyNative.Total.Matchers = append(latencyNative.Total.Matchers, &LabelMatcher{
				Type:  LabelMatcher_Type(m.Type),
				Name:  m.Name,
				Value: m.Value,
			})
		}
	}

	var boolGauge *BoolGauge
	if b := o.Indicator.BoolGauge; b != nil {
		boolGauge = &BoolGauge{
			Grouping: o.Grouping(),
			BoolGauge: &Query{
				Name:   b.Metric.Name,
				Metric: b.Metric.Metric(),
			},
		}
		for _, m := range b.Metric.LabelMatchers {
			boolGauge.BoolGauge.Matchers = append(boolGauge.BoolGauge.Matchers, &LabelMatcher{
				Type:  LabelMatcher_Type(m.Type),
				Name:  m.Name,
				Value: m.Value,
			})
		}
	}

	lset := make(map[string]string, o.Labels.Len())
	for n, v := range o.Labels.Map() {
		name := strings.TrimPrefix(n, slo.PropagationLabelsPrefix)
		name = strutil.SanitizeLabelName(name)
		lset[name] = v
	}

	objective := &Objective{
		Labels:      lset,
		Target:      o.Target,
		Window:      durationpb.New(time.Duration(o.Window)),
		Description: o.Description,
		Config:      o.Config,
	}
	if ratio != nil {
		objective.Indicator = &Indicator{
			Options: &Indicator_Ratio{ratio},
		}
	}
	if latency != nil {
		objective.Indicator = &Indicator{
			Options: &Indicator_Latency{latency},
		}
	}
	if latencyNative != nil {
		objective.Indicator = &Indicator{
			Options: &Indicator_LatencyNative{latencyNative},
		}
	}
	if boolGauge != nil {
		objective.Indicator = &Indicator{
			Options: &Indicator_BoolGauge{boolGauge},
		}
	}
	return objective
}
