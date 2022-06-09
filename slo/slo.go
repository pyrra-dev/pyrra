package slo

import (
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// PropagationLabelsPrefix provides a way to propagate labels from the
// ObjectMeta to the PrometheusRule.
const PropagationLabelsPrefix = "pyrra.dev/"

type Objective struct {
	Labels      labels.Labels
	Description string
	Target      float64
	Window      model.Duration
	Config      string

	Indicator Indicator
}

func (o Objective) Name() string {
	for _, l := range o.Labels {
		if l.Name == labels.MetricName {
			return l.Value
		}
	}
	return ""
}

func (o Objective) Windows() []Window {
	return Windows(time.Duration(o.Window))
}

func (o Objective) HasWindows(short, long model.Duration) (Window, bool) {
	for _, w := range Windows(time.Duration(o.Window)) {
		if w.Short == time.Duration(short) && w.Long == time.Duration(long) {
			return w, true
		}
	}

	return Window{}, false
}

func (o Objective) Grouping() []string {
	if o.Indicator.Ratio != nil {
		return o.Indicator.Ratio.Grouping
	}
	if o.Indicator.Latency != nil {
		return o.Indicator.Latency.Grouping
	}
	return nil
}

type Indicator struct {
	Ratio   *RatioIndicator
	Latency *LatencyIndicator
}

type RatioIndicator struct {
	Errors   Metric
	Total    Metric
	Grouping []string
}

type LatencyIndicator struct {
	Success  Metric
	Total    Metric
	Grouping []string
}

type Metric struct {
	Name          string
	LabelMatchers []*labels.Matcher
}

func (m Metric) Metric() string {
	v := parser.VectorSelector{Name: m.Name, LabelMatchers: m.LabelMatchers}
	return v.String()
}
