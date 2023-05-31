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
	Unknown   IndicatorType = iota
	Ratio     IndicatorType = iota
	Latency   IndicatorType = iota
	BoolGauge IndicatorType = iota
)

func (o Objective) IndicatorType() IndicatorType {
	if o.Indicator.Ratio != nil && o.Indicator.Ratio.Total.Name != "" {
		return Ratio
	}
	if o.Indicator.Latency != nil && o.Indicator.Latency.Total.Name != "" {
		return Latency
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

// AbsentDuration returns the duration Pyrra tells Prometheus to wait before firing an alert.
// This duration is based on the objective's window and target and
// returns the duration to burn the passed percentage (1-100) of the error budget assuming absence of errors is a 100% error rate.
// Learn more here: https://sre.google/workbook/alerting-on-slos/#burn_rates_and_time_to_complete_budget_ex
func (o Objective) AbsentDuration(percent float64) model.Duration {
	sec := time.Duration(o.Window).Seconds()
	budget := 1 - o.Target

	// how many seconds does it take to burn one percent of the error budget?
	percentBurnSeconds := (sec / 100) * budget * percent

	percentBurnDuration := time.Second * time.Duration(percentBurnSeconds)
	rounded := percentBurnDuration.Round(time.Second)

	// We truncate the seconds behind the minutes.
	// In the units tests there were durations like 4m2s, and we simply ignore those 2s.
	if rounded > 2*time.Minute {
		rounded = rounded.Truncate(time.Minute)
	}

	// It most likely doesn't make sense to wait less than a minute with Prometheus scrape intervals.
	// If 1 minute is too long for your use case, please open an issue.
	if rounded < time.Minute {
		return model.Duration(time.Minute)
	}

	return model.Duration(rounded)
}

type Indicator struct {
	Ratio     *RatioIndicator
	Latency   *LatencyIndicator
	BoolGauge *BoolGaugeIndicator
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
