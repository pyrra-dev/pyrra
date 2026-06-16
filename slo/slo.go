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
	defaultAlertnameAbsent  = "SLOMetricAbsent"
)

type Objective struct {
	Labels      labels.Labels
	Annotations map[string]string
	Description string
	Target      float64
	Window      model.Duration
	Config      string

	PerformanceOverAccuracy bool
	RuleOutput              RuleOutput

	Alerting  Alerting
	Indicator Indicator
}

func (o Objective) Name() string {
	var name string
	o.Labels.Range(func(l labels.Label) {
		if l.Name == model.MetricNameLabel {
			name = l.Value
		}
	})
	return name
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

func (o Objective) Exhausts(factor float64) model.Duration {
	return model.Duration(time.Second * time.Duration(time.Duration(o.Window).Seconds()/factor))
}

// AbsentDuration calculates the duration when absent alerts should fire.
// The idea is as follows: Use the most critical of the multi burn rate alerts.
// For that alert to fire, both the short AND long windows have to be above the threshold.
// The long window takes the - longest - to fire.
// Assuming absence of the metric means 100% error rate,
// the time it takes to fire is the duration for the long window to go above the threshold (factor * objective).
// Finally, we add the "for" duration we add to the multi burn rate alerts.
func (o Objective) AbsentDuration() model.Duration {
	mostCritical := o.Windows()[0]
	mostCriticalThreshold := mostCritical.Factor * (1 - o.Target)
	mostCriticalDuration := time.Duration(mostCriticalThreshold*mostCritical.Long.Seconds()) * time.Second
	mostCriticalDuration += mostCritical.For
	return model.Duration(mostCriticalDuration.Round(time.Minute))
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
	if o.Indicator.Ratio != nil {
		if o.Indicator.Ratio.Total.Name != "" ||
			o.Indicator.Ratio.Errors.Name != "" ||
			o.Indicator.Ratio.Success.Name != "" {
			return Ratio
		}
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

func (o Objective) AlertNameAbsent() string {
	if o.Alerting.AbsentName != "" {
		return o.Alerting.AbsentName
	}

	return defaultAlertnameAbsent
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
	Success  Metric
	Grouping []string
}

// RatioCombo identifies which two of the total, errors, and success metrics are
// configured for a ratio indicator. Exactly two of the three are always set
// (enforced during validation), and the missing metric is derived following the
// total = errors + success relationship.
type RatioCombo int

const (
	// RatioErrorsTotal configures errors and total. The error rate is
	// errors / total. This is the original ratio behaviour.
	RatioErrorsTotal RatioCombo = iota
	// RatioSuccessTotal configures success and total. The error rate is
	// (total - success) / total, mirroring the latency indicator.
	RatioSuccessTotal
	// RatioErrorsSuccess configures errors and success. The total is derived as
	// errors + success and the error rate is errors / (errors + success).
	RatioErrorsSuccess
)

// Combo returns which two of total, errors, and success are configured.
func (r RatioIndicator) Combo() RatioCombo {
	hasTotal := r.Total.Name != ""
	hasErrors := r.Errors.Name != ""
	switch {
	case hasTotal && hasErrors:
		return RatioErrorsTotal
	case hasTotal && !hasErrors:
		return RatioSuccessTotal
	default:
		return RatioErrorsSuccess
	}
}

// PrimaryMetric returns the metric used to name recording rules and to derive
// rule labels and grouping. It prefers total, then errors, then success, so that
// the generated rule names stay stable with the original behaviour whenever a
// total metric is configured.
func (r RatioIndicator) PrimaryMetric() Metric {
	if r.Total.Name != "" {
		return r.Total
	}
	if r.Errors.Name != "" {
		return r.Errors
	}
	return r.Success
}

// SecondaryMetric returns the second configured metric, complementing
// PrimaryMetric. Together they are the two metrics that increase recording rules
// are generated for.
func (r RatioIndicator) SecondaryMetric() Metric {
	if r.Combo() == RatioErrorsTotal {
		return r.Errors
	}
	return r.Success
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
	Disabled   bool // deprecated, use Burnrates instead
	Burnrates  bool
	Absent     bool
	Name       string
	AbsentName string
	Severities AlertingSeverities
}

type AlertingSeverities struct {
	Absent       string
	FastBurn     string
	MediumBurn   string
	SlowBurn     string
	LongTermBurn string
}

type Metric struct {
	Name          string
	LabelMatchers []*labels.Matcher
}

func (m Metric) Metric() string {
	v := parser.VectorSelector{Name: m.Name, LabelMatchers: m.LabelMatchers}
	return v.String()
}

type RuleOutput struct {
	ShortRulesLabels         map[string]string
	LongRulesLabels          map[string]string
	EnableDescriptionAsLabel bool
}
