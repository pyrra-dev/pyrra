package slo

import (
	"fmt"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"k8s.io/utils/ptr"
)

// QueryTotal returns a PromQL query to get the total amount of requests served during the window.
func (o Objective) QueryTotal(window model.Duration) string {
	expr, err := parser.ParseExpr(`sum by (grouping) (metric{} offset 1ms)`)
	if err != nil {
		return ""
	}

	var (
		metric   string
		matchers []*labels.Matcher
		grouping []string
	)
	switch o.IndicatorType() {
	case Ratio:
		metric = increaseName(o.Indicator.Ratio.Total.Name, window)
		matchers = cloneMatchers(o.Indicator.Ratio.Total.LabelMatchers)
		grouping = slices.Clone(o.Indicator.Ratio.Grouping)
	case Latency:
		metric = increaseName(o.Indicator.Latency.Total.Name, window)
		grouping = slices.Clone(o.Indicator.Latency.Grouping)
		matchers = append(
			cloneMatchers(o.Indicator.Latency.Total.LabelMatchers),
			&labels.Matcher{Type: labels.MatchEqual, Name: labels.BucketLabel, Value: ""},
		)
	case LatencyNative:
		metric = increaseName(o.Indicator.LatencyNative.Total.Name, window)
		grouping = slices.Clone(o.Indicator.LatencyNative.Grouping)
		matchers = append(
			cloneMatchers(o.Indicator.LatencyNative.Total.LabelMatchers),
			&labels.Matcher{Type: labels.MatchEqual, Name: labels.BucketLabel, Value: ""},
		)
	case BoolGauge:
		metric = countName(o.Indicator.BoolGauge.Name, window)
		matchers = cloneMatchers(o.Indicator.BoolGauge.LabelMatchers)
		grouping = slices.Clone(o.Indicator.BoolGauge.Grouping)
	default:
		return ""
	}

	matchers = append(matchers, &labels.Matcher{
		Type:  labels.MatchEqual,
		Name:  "slo",
		Value: o.Name(),
	})

	for _, m := range matchers {
		if m.Name == labels.MetricName {
			m.Value = metric
		}
	}

	objectiveReplacer{
		metric:   metric,
		matchers: matchers,
		grouping: grouping,
	}.replace(expr)

	return expr.String()
}

// QueryErrors returns a PromQL query to get the amount of request errors during the window.
func (o Objective) QueryErrors(window model.Duration) string {
	switch o.IndicatorType() {
	case Ratio:
		expr, err := parser.ParseExpr(`sum by (grouping) (metric{} offset 1ms)`)
		if err != nil {
			return ""
		}

		metric := increaseName(o.Indicator.Ratio.Errors.Name, window)
		matchers := cloneMatchers(o.Indicator.Ratio.Errors.LabelMatchers)

		for _, m := range matchers {
			if m.Name == labels.MetricName {
				m.Value = metric
				break
			}
		}
		matchers = append(matchers, &labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  "slo",
			Value: o.Name(),
		})

		objectiveReplacer{
			metric:   metric,
			matchers: matchers,
			grouping: o.Indicator.Ratio.Grouping,
		}.replace(expr)

		return expr.String()
	case Latency:
		expr, err := parser.ParseExpr(`sum by (grouping) (metric{matchers="total"} offset 1ms) - sum by (grouping) (errorMetric{matchers="errors"} offset 2ms)`)
		if err != nil {
			return ""
		}

		metric := increaseName(o.Indicator.Latency.Total.Name, window)
		matchers := cloneMatchers(o.Indicator.Latency.Total.LabelMatchers)
		for _, m := range matchers {
			if m.Name == labels.MetricName {
				m.Value = metric
				break
			}
		}
		// Add the matcher {le=""} to select the recording rule that summed up all requests
		matchers = append(matchers, &labels.Matcher{Type: labels.MatchEqual, Name: labels.BucketLabel, Value: ""})
		matchers = append(matchers, &labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  "slo",
			Value: o.Name(),
		})

		errorMetric := increaseName(o.Indicator.Latency.Success.Name, window)
		errorMatchers := cloneMatchers(o.Indicator.Latency.Success.LabelMatchers)
		for _, m := range errorMatchers {
			if m.Name == labels.MetricName {
				m.Value = errorMetric
				break
			}
		}
		errorMatchers = append(errorMatchers, &labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  "slo",
			Value: o.Name(),
		})

		objectiveReplacer{
			metric:        metric,
			matchers:      matchers,
			errorMetric:   errorMetric,
			errorMatchers: errorMatchers,
			grouping:      o.Indicator.Latency.Grouping,
		}.replace(expr)

		return expr.String()
	case LatencyNative:
		expr, err := parser.ParseExpr(`sum by (grouping) (metric{matchers="total"} offset 1ms) - sum by (grouping) (errorMetric{matchers="errors"} offset 2ms)`)
		if err != nil {
			return ""
		}

		metric := increaseName(o.Indicator.LatencyNative.Total.Name, window)
		matchers := cloneMatchers(o.Indicator.LatencyNative.Total.LabelMatchers)
		for i, m := range matchers {
			if m.Name == labels.MetricName {
				matchers[i].Value = metric
				break
			}
		}
		matchers = append(matchers, &labels.Matcher{Type: labels.MatchEqual, Name: "slo", Value: o.Name()})

		// Add the matcher {le=""} to select the recording rule that summed up all requests
		totalMatchers := append(cloneMatchers(matchers), &labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  labels.BucketLabel,
			Value: "",
		})

		errorMatchers := append(cloneMatchers(matchers), &labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  labels.BucketLabel,
			Value: fmt.Sprintf("%g", time.Duration(o.Indicator.LatencyNative.Latency).Seconds()),
		})

		objectiveReplacer{
			metric:        metric,
			matchers:      totalMatchers,
			errorMetric:   metric,
			errorMatchers: errorMatchers,
			grouping:      o.Indicator.LatencyNative.Grouping,
		}.replace(expr)

		return expr.String()
	case BoolGauge:
		expr, err := parser.ParseExpr(`sum by (grouping) (errorMetric{matchers="errors"} offset 2ms) - sum by (grouping) (metric{matchers="total"} offset 1ms)`)
		if err != nil {
			return ""
		}

		metric := sumName(o.Indicator.BoolGauge.Name, window)
		matchers := cloneMatchers(o.Indicator.BoolGauge.LabelMatchers)
		errorMetric := countName(o.Indicator.BoolGauge.Name, window)
		errorMatchers := cloneMatchers(o.Indicator.BoolGauge.LabelMatchers)

		for _, m := range matchers {
			if m.Name == labels.MetricName {
				m.Value = metric
				break
			}
		}
		matchers = append(matchers, &labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  "slo",
			Value: o.Name(),
		})

		for _, m := range errorMatchers {
			if m.Name == labels.MetricName {
				m.Value = errorMetric
				break
			}
		}
		errorMatchers = append(errorMatchers, &labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  "slo",
			Value: o.Name(),
		})

		objectiveReplacer{
			metric:        metric,
			matchers:      matchers,
			errorMetric:   errorMetric,
			errorMatchers: errorMatchers,
			grouping:      o.Indicator.BoolGauge.Grouping,
		}.replace(expr)

		return expr.String()
	default:
		return ""
	}
}

func (o Objective) QueryErrorBudget() string {
	indicatorType := o.IndicatorType()
	switch indicatorType {
	case Ratio:
		expr, err := parser.ParseExpr(`
(
  (1 - 0.696969)
  -
  (
    sum(errorMetric{matchers="errors"} offset 2ms or vector(0))
    /
    sum(metric{matchers="total"} offset 1ms)
  )
)
/
(1 - 0.696969)
`)
		if err != nil {
			return ""
		}

		metric := increaseName(o.Indicator.Ratio.Total.Name, o.Window)
		matchers := cloneMatchers(o.Indicator.Ratio.Total.LabelMatchers)
		for _, m := range matchers {
			if m.Name == labels.MetricName {
				m.Value = metric
				break
			}
		}
		matchers = append(matchers, &labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  "slo",
			Value: o.Name(),
		})

		errorMetric := increaseName(o.Indicator.Ratio.Errors.Name, o.Window)
		errorMatchers := cloneMatchers(o.Indicator.Ratio.Errors.LabelMatchers)
		for _, m := range errorMatchers {
			if m.Name == labels.MetricName {
				m.Value = errorMetric
				break
			}
		}
		errorMatchers = append(errorMatchers, &labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  "slo",
			Value: o.Name(),
		})

		objectiveReplacer{
			metric:        metric,
			matchers:      matchers,
			errorMetric:   errorMetric,
			errorMatchers: errorMatchers,
			grouping:      o.Indicator.Ratio.Grouping,
			target:        o.Target,
		}.replace(expr)

		return expr.String()
	case Latency, LatencyNative:
		expr, err := parser.ParseExpr(`
(
  (1 - 0.696969)
  -
  (
    1 -
    sum(errorMetric{matchers="errors"} offset 2ms or vector(0))
    /
    sum(metric{matchers="total"} offset 1ms)
  )
)
/
(1 - 0.696969)
`)
		if err != nil {
			return ""
		}

		var (
			metric        string
			matchers      []*labels.Matcher
			errorMetric   string
			errorMatchers []*labels.Matcher
			grouping      []string
		)
		switch indicatorType {
		case Latency:
			metric = increaseName(o.Indicator.Latency.Total.Name, o.Window)
			matchers = cloneMatchers(o.Indicator.Latency.Total.LabelMatchers)
			errorMetric = increaseName(o.Indicator.Latency.Success.Name, o.Window)
			errorMatchers = cloneMatchers(o.Indicator.Latency.Success.LabelMatchers)
			grouping = o.Indicator.Latency.Grouping
		case LatencyNative:
			metric = increaseName(o.Indicator.LatencyNative.Total.Name, o.Window)
			matchers = cloneMatchers(o.Indicator.LatencyNative.Total.LabelMatchers)
			errorMetric = increaseName(o.Indicator.LatencyNative.Total.Name, o.Window)
			errorMatchers = cloneMatchers(o.Indicator.LatencyNative.Total.LabelMatchers)
			grouping = o.Indicator.LatencyNative.Grouping
		}

		for _, m := range matchers {
			if m.Name == labels.MetricName {
				m.Value = metric
				break
			}
		}
		for _, m := range errorMatchers {
			if m.Name == labels.MetricName {
				m.Value = errorMetric
				break
			}
		}

		// Add the matcher {le=""} to select the recording rule that summed up all requests
		matchers = append(matchers, &labels.Matcher{Type: labels.MatchEqual, Name: labels.BucketLabel, Value: ""})
		matchers = append(matchers, &labels.Matcher{Type: labels.MatchEqual, Name: "slo", Value: o.Name()})

		errorMatchers = append(errorMatchers, &labels.Matcher{Type: labels.MatchEqual, Name: "slo", Value: o.Name()})
		if indicatorType == LatencyNative {
			seconds := time.Duration(o.Indicator.LatencyNative.Latency).Seconds()
			errorMatchers = append(errorMatchers, &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  labels.BucketLabel,
				Value: fmt.Sprint(seconds),
			})
		}

		objectiveReplacer{
			metric:        metric,
			matchers:      matchers,
			errorMetric:   errorMetric,
			errorMatchers: errorMatchers,
			grouping:      grouping,
			target:        o.Target,
		}.replace(expr)

		return expr.String()
	case BoolGauge:
		expr, err := parser.ParseExpr(`
(
  (1 - 0.696969)
  -
  (
    (sum by (grouping) (metric{matchers="total"} offset 1ms) -
    sum by (grouping) (errorMetric{matchers="errors"} offset 2ms))
    /
    sum by (grouping) (metric{matchers="total"} offset 1ms)
  )
)
/
(1 - 0.696969)
`)
		if err != nil {
			return ""
		}

		metric := countName(o.Indicator.BoolGauge.Name, o.Window)
		matchers := cloneMatchers(o.Indicator.BoolGauge.LabelMatchers)
		for _, m := range matchers {
			if m.Name == labels.MetricName {
				m.Value = metric
				break
			}
		}
		matchers = append(matchers, &labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  "slo",
			Value: o.Name(),
		})

		errorMetric := sumName(o.Indicator.BoolGauge.Name, o.Window)
		errorMatchers := cloneMatchers(o.Indicator.BoolGauge.LabelMatchers)
		for _, m := range errorMatchers {
			if m.Name == labels.MetricName {
				m.Value = errorMetric
				break
			}
		}
		errorMatchers = append(errorMatchers, &labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  "slo",
			Value: o.Name(),
		})

		objectiveReplacer{
			metric:        metric,
			matchers:      matchers,
			errorMetric:   errorMetric,
			errorMatchers: errorMatchers,
			grouping:      o.Indicator.BoolGauge.Grouping,
			target:        o.Target,
		}.replace(expr)

		return expr.String()
	default:
		return ""
	}
}

func (o Objective) QueryBurnrate(timerange time.Duration, groupingMatchers []*labels.Matcher) (string, error) {
	metric := ""
	matchers := map[string]*labels.Matcher{}

	switch o.IndicatorType() {
	case Ratio:
		metric = o.BurnrateName(timerange)
		for _, m := range o.Indicator.Ratio.Total.LabelMatchers {
			matchers[m.Name] = &labels.Matcher{ // Copy labels by value to avoid race
				Type:  m.Type,
				Name:  m.Name,
				Value: m.Value,
			}
		}
	case Latency:
		metric = o.BurnrateName(timerange)
		for _, m := range o.Indicator.Latency.Total.LabelMatchers {
			matchers[m.Name] = &labels.Matcher{ // Copy labels by value to avoid race
				Type:  m.Type,
				Name:  m.Name,
				Value: m.Value,
			}
		}
	case LatencyNative:
		metric = o.BurnrateName(timerange)
		for _, m := range o.Indicator.LatencyNative.Total.LabelMatchers {
			matchers[m.Name] = &labels.Matcher{
				Type:  m.Type,
				Name:  m.Name,
				Value: m.Value,
			}
		}
	case BoolGauge:
		metric = o.BurnrateName(timerange)
		for _, m := range o.Indicator.BoolGauge.LabelMatchers {
			matchers[m.Name] = &labels.Matcher{ // Copy labels by value to avoid race
				Type:  m.Type,
				Name:  m.Name,
				Value: m.Value,
			}
		}
	}

	if metric == "" {
		return "", fmt.Errorf("objective misses indicator")
	}

	expr, err := parser.ParseExpr(`metric{} offset 1ms`)
	if err != nil {
		return "", err
	}

	for i, m := range matchers {
		if m.Name == labels.MetricName {
			matchers[i].Value = metric
		}
	}

	for _, m := range groupingMatchers {
		if m.Type != labels.MatchEqual {
			return "", fmt.Errorf("grouping matcher has to be MatchEqual not %s", m.Type.String())
		}

		matchers[m.Name] = &labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  m.Name,
			Value: m.Value,
		}
	}

	matchers["slo"] = &labels.Matcher{
		Type:  labels.MatchEqual,
		Name:  "slo",
		Value: o.Name(),
	}

	matchersSlice := make([]*labels.Matcher, 0, len(matchers))
	for _, m := range matchers {
		matchersSlice = append(matchersSlice, m)
	}

	objectiveReplacer{
		metric:   metric,
		matchers: matchersSlice,
	}.replace(expr)

	return expr.String(), nil
}

type objectiveReplacer struct {
	metric        string
	matchers      []*labels.Matcher
	offset        *time.Duration
	errorMetric   string
	errorMatchers []*labels.Matcher
	errorOffset   *time.Duration
	grouping      []string
	window        time.Duration
	target        float64
	percentile    float64
}

func (r objectiveReplacer) replace(node parser.Node) {
	switch n := node.(type) {
	case *parser.AggregateExpr:
		if len(n.Grouping) > 0 {
			n.Grouping = r.grouping
		}
		r.replace(n.Expr)
	case *parser.Call:
		r.replace(n.Args)
	case parser.Expressions:
		for _, expr := range n {
			r.replace(expr)
		}
	case *parser.MatrixSelector:
		n.Range = r.window
		r.replace(n.VectorSelector)
	case *parser.VectorSelector:
		if n.Name == "errorMetric" {
			n.Name = r.errorMetric
		} else {
			n.Name = r.metric
		}
		if n.OriginalOffset == 2*time.Millisecond {
			// 2ms is the placeholder for the errorOffset
			n.OriginalOffset = ptr.Deref(r.errorOffset, 0)
		} else if n.OriginalOffset == 1*time.Millisecond {
			// 1ms is the placeholder for the offset
			n.OriginalOffset = ptr.Deref(r.offset, 0)
		} else {
			n.OriginalOffset = 0
		}
		if len(n.LabelMatchers) > 1 {
			for _, m := range n.LabelMatchers {
				if m.Name == "matchers" {
					if m.Value == "errors" {
						n.LabelMatchers = r.errorMatchers
					} else {
						n.LabelMatchers = r.matchers
					}
				}
			}
		} else {
			n.LabelMatchers = r.matchers
		}
	case *parser.BinaryExpr:
		r.replace(n.LHS)
		r.replace(n.RHS)
	case *parser.ParenExpr:
		r.replace(n.Expr)
	case *parser.NumberLiteral:
		switch n.Val {
		case 0.696969:
			n.Val = r.target
		case 0.420:
			n.Val = r.percentile
		case 86400:
			n.Val = r.window.Seconds()
		default:
		}
	default:
		panic(fmt.Sprintf("no support for type %T", n))
	}
}

func (o Objective) RequestRange(timerange time.Duration) string {
	switch o.IndicatorType() {
	case Ratio:
		expr, err := parser.ParseExpr(`sum by (group) (rate(metric{}[1s] offset 1ms)) > 0`)
		if err != nil {
			return err.Error()
		}

		matchers := cloneMatchers(o.Indicator.Ratio.Total.LabelMatchers)
		for i, m := range matchers {
			if m.Name == labels.MetricName {
				matchers[i].Value = o.Indicator.Ratio.Total.Name
			}
		}

		objectiveReplacer{
			metric:   o.Indicator.Ratio.Total.Name,
			matchers: matchers,
			offset:   o.Indicator.Ratio.Total.OriginalOffset,
			grouping: groupingLabels(
				o.Indicator.Ratio.Errors.LabelMatchers,
				matchers,
			),
			window: timerange,
			target: o.Target,
		}.replace(expr)

		return expr.String()
	case Latency:
		expr, err := parser.ParseExpr(`sum(rate(metric{}[1s] offset 1ms))`)
		if err != nil {
			return err.Error()
		}

		objectiveReplacer{
			metric:        o.Indicator.Latency.Total.Name,
			matchers:      o.Indicator.Latency.Total.LabelMatchers,
			offset:        o.Indicator.Latency.Total.OriginalOffset,
			errorMetric:   o.Indicator.Latency.Success.Name,
			errorMatchers: o.Indicator.Latency.Success.LabelMatchers,
			errorOffset:   o.Indicator.Latency.Success.OriginalOffset,
			window:        timerange,
		}.replace(expr)

		return expr.String()
	case LatencyNative:
		expr, err := parser.ParseExpr(`sum(histogram_count(rate(metric{}[1s] offset 1ms)))`)
		if err != nil {
			return err.Error()
		}

		objectiveReplacer{
			metric:   o.Indicator.LatencyNative.Total.Name,
			matchers: o.Indicator.LatencyNative.Total.LabelMatchers,
			offset:   o.Indicator.LatencyNative.Total.OriginalOffset,
			window:   timerange,
		}.replace(expr)

		return expr.String()
	case BoolGauge:
		expr, err := parser.ParseExpr(`sum by(group) (count_over_time(metric{matchers="total"}[1s] offset 1ms)) / 86400`)
		if err != nil {
			return err.Error()
		}

		matchers := cloneMatchers(o.Indicator.BoolGauge.LabelMatchers)
		objectiveReplacer{
			metric:   o.Indicator.BoolGauge.Name,
			matchers: matchers,
			offset:   o.Indicator.BoolGauge.OriginalOffset,
			grouping: o.Grouping(),
			window:   timerange,
		}.replace(expr)

		return expr.String()
	default:
		return ""
	}
}

func (o Objective) ErrorsRange(timerange time.Duration) string {
	switch o.IndicatorType() {
	case Ratio:
		expr, err := parser.ParseExpr(`sum by (group) (rate(errorMetric{matchers="errors"}[1s] offset 2ms)) / scalar(sum(rate(metric{matchers="total"}[1s] offset 1ms))) > 0`)
		if err != nil {
			return err.Error()
		}

		matchers := cloneMatchers(o.Indicator.Ratio.Total.LabelMatchers)
		for i, m := range matchers {
			if m.Name == labels.MetricName {
				matchers[i].Value = o.Indicator.Ratio.Total.Name
			}
		}

		errorMatchers := cloneMatchers(o.Indicator.Ratio.Errors.LabelMatchers)
		for i, m := range errorMatchers {
			if m.Name == labels.MetricName {
				errorMatchers[i].Value = o.Indicator.Ratio.Errors.Name
			}
		}

		objectiveReplacer{
			metric:        o.Indicator.Ratio.Total.Name,
			matchers:      matchers,
			offset:        o.Indicator.Ratio.Total.OriginalOffset,
			errorMetric:   o.Indicator.Ratio.Errors.Name,
			errorMatchers: errorMatchers,
			errorOffset:   o.Indicator.Ratio.Errors.OriginalOffset,
			grouping: groupingLabels(
				errorMatchers,
				matchers,
			),
			window: timerange,
		}.replace(expr)

		return expr.String()
	case Latency:
		expr, err := parser.ParseExpr(`(sum(rate(metric{matchers="total"}[1s] offset 1ms)) -  sum(rate(errorMetric{matchers="errors"}[1s] offset 2ms))) / sum(rate(metric{matchers="total"}[1s] offset 1ms))`)
		if err != nil {
			return err.Error()
		}

		objectiveReplacer{
			metric:        o.Indicator.Latency.Total.Name,
			matchers:      o.Indicator.Latency.Total.LabelMatchers,
			offset:        o.Indicator.Latency.Total.OriginalOffset,
			errorMetric:   o.Indicator.Latency.Success.Name,
			errorMatchers: o.Indicator.Latency.Success.LabelMatchers,
			errorOffset:   o.Indicator.Latency.Success.OriginalOffset,
			window:        timerange,
		}.replace(expr)

		return expr.String()
	case LatencyNative:
		expr, err := parser.ParseExpr(`1 - sum(histogram_fraction(0,0.696969, rate(metric{matchers="total"}[1s] offset 1ms)))`)
		if err != nil {
			return err.Error()
		}

		objectiveReplacer{
			metric:   o.Indicator.LatencyNative.Total.Name,
			matchers: o.Indicator.LatencyNative.Total.LabelMatchers,
			offset:   o.Indicator.LatencyNative.Total.OriginalOffset,
			window:   timerange,
			target:   time.Duration(o.Indicator.LatencyNative.Latency).Seconds(),
		}.replace(expr)

		return expr.String()
	case BoolGauge:
		expr, err := parser.ParseExpr(`100 * sum by (group) ((count_over_time(metric{matchers="total"}[1s] offset 1ms) - sum_over_time(metric{matchers="total"}[1s] offset 1ms))) / sum by(group) (count_over_time(metric{matchers="total"}[1s] offset 1ms))`)
		if err != nil {
			return err.Error()
		}

		objectiveReplacer{
			metric:   o.Indicator.BoolGauge.Name,
			matchers: o.Indicator.BoolGauge.LabelMatchers,
			offset:   o.Indicator.BoolGauge.OriginalOffset,
			window:   timerange,
		}.replace(expr)

		return expr.String()
	default:
		return ""
	}
}

func (o Objective) DurationRange(timerange time.Duration, percentile float64) string {
	switch o.IndicatorType() {
	case Latency:
		expr, err := parser.ParseExpr(`histogram_quantile(0.420, sum by (le) (rate(errorMetric{matchers="errors"}[1s] offset 2ms)))`)
		if err != nil {
			return err.Error()
		}

		matchers := make([]*labels.Matcher, 0, len(o.Indicator.Latency.Success.LabelMatchers))
		for _, m := range o.Indicator.Latency.Success.LabelMatchers {
			if m.Name != "le" {
				matchers = append(matchers, m)
			}
		}

		objectiveReplacer{
			errorMetric:   o.Indicator.Latency.Success.Name,
			errorMatchers: matchers,
			errorOffset:   o.Indicator.Latency.Success.OriginalOffset,
			window:        timerange,
			grouping:      []string{labels.BucketLabel},
			percentile:    percentile,
		}.replace(expr)

		return expr.String()
	case LatencyNative:
		expr, err := parser.ParseExpr(`histogram_quantile(0.420, sum(rate(metric{matchers="total"}[1s] offset 1ms)))`)
		if err != nil {
			return err.Error()
		}

		objectiveReplacer{
			metric:     o.Indicator.LatencyNative.Total.Name,
			matchers:   o.Indicator.LatencyNative.Total.LabelMatchers,
			offset:     o.Indicator.LatencyNative.Total.OriginalOffset,
			grouping:   o.Indicator.LatencyNative.Grouping,
			window:     timerange,
			percentile: percentile,
		}.replace(expr)

		return expr.String()
	default:
		return ""
	}
}

func groupingLabels(errorMatchers, totalMatchers []*labels.Matcher) []string {
	groupingLabels := map[string]struct{}{}
	for _, m := range errorMatchers {
		groupingLabels[m.Name] = struct{}{}
	}
	for _, m := range totalMatchers {
		delete(groupingLabels, m.Name)
	}

	// This deletes the le label as grouping by it should usually not be wanted,
	// and we have to remove it for the latency SLOs.
	delete(groupingLabels, labels.BucketLabel)

	return maps.Keys(groupingLabels)
}

func cloneMatchers(matchers []*labels.Matcher) []*labels.Matcher {
	r := make([]*labels.Matcher, len(matchers))
	for i, matcher := range matchers {
		m := *matcher
		r[i] = &m
	}
	return r
}
