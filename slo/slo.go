package slo

import (
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

const (
	// PropagationLabelsPrefix provides a way to propagate labels from the
	// ObjectMeta to the PrometheusRule.
	PropagationLabelsPrefix = "pyrra.dev/"
	defaultAlertname        = "ErrorBudgetBurn"
)

type Objective struct {
	Labels      labels.Labels
	Annotations map[string]string
	Description string
	Target      float64
	Window      model.Duration
	Config      string

	Alerting  Alerting
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

type IndicatorType int

const (
	Unknown       IndicatorType = iota
	Ratio         IndicatorType = iota
	Latency       IndicatorType = iota
	LatencyNative IndicatorType = iota
	BoolGauge     IndicatorType = iota
)

func (o Objective) IndicatorType() IndicatorType {
	if o.Indicator.Ratio != nil && o.Indicator.Ratio.Total.Name != "" {
		return Ratio
	}
	if o.Indicator.Latency != nil && o.Indicator.Latency.Total.Name != "" {
		return Latency
	}
	if o.Indicator.LatencyNative != nil && o.Indicator.LatencyNative.Total.Name != "" {
		return LatencyNative
	}
	if o.Indicator.BoolGauge != nil && o.Indicator.BoolGauge.Name != "" {
		return BoolGauge
	}
	return Unknown
}

func (o Objective) Grouping() []string {
	switch o.IndicatorType() {
	case Ratio:
		return o.Indicator.Ratio.Grouping
	case Latency:
		return o.Indicator.Latency.Grouping
	case LatencyNative:
		return o.Indicator.LatencyNative.Grouping
	case BoolGauge:
		return o.Indicator.BoolGauge.Grouping
	default:
		return nil
	}
}

func (o Objective) AlertName() string {
	if o.Alerting.Name != "" {
		return o.Alerting.Name
	}

	return defaultAlertname
}

type Indicator struct {
	Ratio         *RatioIndicator
	Latency       *LatencyIndicator
	LatencyNative *LatencyNativeIndicator
	BoolGauge     *BoolGaugeIndicator
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

type LatencyNativeIndicator struct {
	Latency  model.Duration
	Total    Metric
	Grouping []string
}

type BoolGaugeIndicator struct {
	Metric
	Grouping []string
}

type Alerting struct {
	Disabled bool
	Name     string
}

type Metric struct {
	Name          string
	LabelMatchers []*labels.Matcher
}

func (m Metric) Metric() string {
	v := parser.VectorSelector{Name: m.Name, LabelMatchers: m.LabelMatchers}
	return v.String()
}
