package slo

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type MultiBurnRateAlert struct {
	Severity string
	Short    time.Duration
	Long     time.Duration
	For      time.Duration
	Factor   float64

	QueryShort string
	QueryLong  string
}

func (o Objective) Alerts() ([]MultiBurnRateAlert, error) {
	ws := Windows(time.Duration(o.Window))

	mbras := make([]MultiBurnRateAlert, len(ws))
	for i, w := range ws {
		queryShort, err := o.QueryBurnrate(w.Short, nil)
		if err != nil {
			return nil, err
		}
		queryLong, err := o.QueryBurnrate(w.Long, nil)
		if err != nil {
			return nil, err
		}

		mbras[i] = MultiBurnRateAlert{
			Severity:   o.alertSeverityLabel(i, w),
			Short:      w.Short,
			Long:       w.Long,
			For:        w.For,
			Factor:     w.Factor,
			QueryShort: queryShort,
			QueryLong:  queryLong,
		}
	}

	return mbras, nil
}

func (o Objective) Burnrates(opts GenerationOptions) (monitoringv1.RuleGroup, error) {
	sloName := o.Labels.Get(model.MetricNameLabel)
	externalURL := opts.ExternalURL

	ws := Windows(time.Duration(o.Window))
	burnrates := burnratesFromWindows(ws)
	rules := make([]monitoringv1.Rule, 0, len(burnrates))

	switch o.IndicatorType() {
	case Ratio:
		matchers := o.Indicator.Ratio.Total.LabelMatchers

		groupingMap := map[string]struct{}{}
		for _, g := range o.Indicator.Ratio.Grouping {
			groupingMap[g] = struct{}{}
		}

		ruleLabels := o.commonRuleLabels(sloName)
		for _, m := range matchers {
			if m.Type == labels.MatchEqual && m.Name != model.MetricNameLabel {
				ruleLabels[m.Name] = m.Value
			}
		}
		// Delete labels that are grouped as their value is part of the labels anyway
		for g := range groupingMap {
			delete(ruleLabels, g)
		}

		for _, br := range burnrates {
			rules = append(rules, monitoringv1.Rule{
				Record: o.BurnrateName(br),
				Expr:   intstr.FromString(o.Burnrate(br, opts)),
				Labels: ruleLabels,
			})
		}

		if o.Alerting.Disabled || !o.Alerting.Burnrates {
			return monitoringv1.RuleGroup{
				Name:     sloName,
				Interval: monitoringDuration("30s"), // TODO: Increase or decrease based on availability target
				Rules:    rules,
			}, nil
		}

		var alertMatchers []string
		for _, m := range matchers {
			if m.Name == model.MetricNameLabel {
				continue
			}
			if _, ok := groupingMap[m.Name]; !ok {
				if m.Type == labels.MatchRegexp || m.Type == labels.MatchNotRegexp {
					continue
				}
			}

			alertMatchers = append(alertMatchers, m.String())
		}
		alertMatchers = append(alertMatchers, fmt.Sprintf(`slo="%s"`, sloName))
		sort.Strings(alertMatchers)
		alertMatchersString := strings.Join(alertMatchers, ",")

		for i, w := range ws {
			alertLabels := o.commonRuleLabels(sloName)
			alertAnnotations := o.commonRuleAnnotations(externalURL)
			for _, m := range matchers {
				if m.Type == labels.MatchEqual && m.Name != model.MetricNameLabel {
					if _, ok := groupingMap[m.Name]; !ok { // only add labels that aren't grouped by
						alertLabels[m.Name] = m.Value
					}
				}
			}

			// Propagate useful SLO information to alerts' labels
			alertLabels["short"] = model.Duration(w.Short).String()
			alertLabels["long"] = model.Duration(w.Long).String()
			alertLabels["severity"] = o.alertSeverityLabel(i, w)
			alertLabels["exhaustion"] = o.Exhausts(w.Factor).String()

			r := monitoringv1.Rule{
				Alert: o.AlertName(),
				// TODO: Use expr replacer
				Expr: intstr.FromString(fmt.Sprintf("%s{%s} > (%.f * (1-%s)) and %s{%s} > (%.f * (1-%s))",
					o.BurnrateName(w.Short),
					alertMatchersString,
					w.Factor,
					strconv.FormatFloat(o.Target, 'f', -1, 64),
					o.BurnrateName(w.Long),
					alertMatchersString,
					w.Factor,
					strconv.FormatFloat(o.Target, 'f', -1, 64),
				)),
				For:         monitoringDuration(w.For.String()),
				Labels:      alertLabels,
				Annotations: alertAnnotations,
			}
			rules = append(rules, r)
		}
	case Latency:
		matchers := o.Indicator.Latency.Total.LabelMatchers

		groupingMap := map[string]struct{}{}
		for _, g := range o.Indicator.Latency.Grouping {
			groupingMap[g] = struct{}{}
		}

		ruleLabels := o.commonRuleLabels(sloName)
		for _, m := range matchers {
			if m.Type == labels.MatchEqual && m.Name != model.MetricNameLabel {
				ruleLabels[m.Name] = m.Value
			}
		}
		// Delete labels that are grouped as their value is part of the labels anyway
		for g := range groupingMap {
			delete(ruleLabels, g)
		}

		for _, br := range burnrates {
			rules = append(rules, monitoringv1.Rule{
				Record: o.BurnrateName(br),
				Expr:   intstr.FromString(o.Burnrate(br, opts)),
				Labels: ruleLabels,
			})
		}

		if o.Alerting.Disabled || !o.Alerting.Burnrates {
			return monitoringv1.RuleGroup{
				Name:     sloName,
				Interval: monitoringDuration("30s"), // TODO: Increase or decrease based on availability target
				Rules:    rules,
			}, nil
		}

		var alertMatchers []string
		for _, m := range matchers {
			if m.Name == model.MetricNameLabel {
				continue
			}
			if _, ok := groupingMap[m.Name]; !ok {
				if m.Type == labels.MatchRegexp || m.Type == labels.MatchNotRegexp {
					continue
				}
			}

			alertMatchers = append(alertMatchers, m.String())
		}
		alertMatchers = append(alertMatchers, fmt.Sprintf(`slo="%s"`, sloName))
		sort.Strings(alertMatchers)
		alertMatchersString := strings.Join(alertMatchers, ",")

		for i, w := range ws {
			alertLabels := o.commonRuleLabels(sloName)
			alertAnnotations := o.commonRuleAnnotations(externalURL)
			for _, m := range matchers {
				if m.Type == labels.MatchEqual && m.Name != model.MetricNameLabel {
					if _, ok := groupingMap[m.Name]; !ok { // only add labels that aren't grouped by
						alertLabels[m.Name] = m.Value
					}
				}
			}

			// Propagate useful SLO information to alerts' labels
			alertLabels["short"] = model.Duration(w.Short).String()
			alertLabels["long"] = model.Duration(w.Long).String()
			alertLabels["severity"] = o.alertSeverityLabel(i, w)
			alertLabels["exhaustion"] = o.Exhausts(w.Factor).String()

			r := monitoringv1.Rule{
				Alert: o.AlertName(),
				// TODO: Use expr replacer
				Expr: intstr.FromString(fmt.Sprintf("%s{%s} > (%.f * (1-%s)) and %s{%s} > (%.f * (1-%s))",
					o.BurnrateName(w.Short),
					alertMatchersString,
					w.Factor,
					strconv.FormatFloat(o.Target, 'f', -1, 64),
					o.BurnrateName(w.Long),
					alertMatchersString,
					w.Factor,
					strconv.FormatFloat(o.Target, 'f', -1, 64),
				)),
				For:         monitoringDuration(model.Duration(w.For).String()),
				Labels:      alertLabels,
				Annotations: alertAnnotations,
			}
			rules = append(rules, r)
		}
	case LatencyNative:
		matchers := o.Indicator.LatencyNative.Total.LabelMatchers

		groupingMap := map[string]struct{}{}
		for _, g := range o.Indicator.LatencyNative.Grouping {
			groupingMap[g] = struct{}{}
		}

		ruleLabels := o.commonRuleLabels(sloName)
		for _, m := range matchers {
			if m.Type == labels.MatchEqual && m.Name != model.MetricNameLabel {
				ruleLabels[m.Name] = m.Value
			}
		}
		// Delete labels that are grouped as their value is part of the labels anyway
		for g := range groupingMap {
			delete(ruleLabels, g)
		}

		for _, br := range burnrates {
			rules = append(rules, monitoringv1.Rule{
				Record: o.BurnrateName(br),
				Expr:   intstr.FromString(o.Burnrate(br, opts)),
				Labels: ruleLabels,
			})
		}

		if o.Alerting.Disabled || !o.Alerting.Burnrates {
			return monitoringv1.RuleGroup{
				Name:     sloName,
				Interval: monitoringDuration("30s"), // TODO: Increase or decrease based on availability target
				Rules:    rules,
			}, nil
		}

		var alertMatchers []string
		for _, m := range matchers {
			if m.Name == model.MetricNameLabel {
				continue
			}
			if _, ok := groupingMap[m.Name]; !ok {
				if m.Type == labels.MatchRegexp || m.Type == labels.MatchNotRegexp {
					continue
				}
			}

			alertMatchers = append(alertMatchers, m.String())
		}
		alertMatchers = append(alertMatchers, fmt.Sprintf(`slo="%s"`, sloName))
		sort.Strings(alertMatchers)
		alertMatchersString := strings.Join(alertMatchers, ",")

		for i, w := range ws {
			alertLabels := o.commonRuleLabels(sloName)
			alertAnnotations := o.commonRuleAnnotations(externalURL)
			for _, m := range matchers {
				if m.Type == labels.MatchEqual && m.Name != model.MetricNameLabel {
					if _, ok := groupingMap[m.Name]; !ok { // only add labels that aren't grouped by
						alertLabels[m.Name] = m.Value
					}
				}
			}

			// Propagate useful SLO information to alerts' labels
			alertLabels["short"] = model.Duration(w.Short).String()
			alertLabels["long"] = model.Duration(w.Long).String()
			alertLabels["severity"] = o.alertSeverityLabel(i, w)
			alertLabels["exhaustion"] = o.Exhausts(w.Factor).String()

			r := monitoringv1.Rule{
				Alert: o.AlertName(),
				// TODO: Use expr replacer
				Expr: intstr.FromString(fmt.Sprintf("%s{%s} > (%.f * (1-%s)) and %s{%s} > (%.f * (1-%s))",
					o.BurnrateName(w.Short),
					alertMatchersString,
					w.Factor,
					strconv.FormatFloat(o.Target, 'f', -1, 64),
					o.BurnrateName(w.Long),
					alertMatchersString,
					w.Factor,
					strconv.FormatFloat(o.Target, 'f', -1, 64),
				)),
				For:         monitoringDuration(model.Duration(w.For).String()),
				Labels:      alertLabels,
				Annotations: alertAnnotations,
			}
			rules = append(rules, r)
		}
	case BoolGauge:
		matchers := o.Indicator.BoolGauge.LabelMatchers

		groupingMap := map[string]struct{}{}
		for _, g := range o.Indicator.BoolGauge.Grouping {
			groupingMap[g] = struct{}{}
		}

		ruleLabels := o.commonRuleLabels(sloName)
		for _, m := range matchers {
			if m.Type == labels.MatchEqual && m.Name != model.MetricNameLabel {
				ruleLabels[m.Name] = m.Value
			}
		}
		// Delete labels that are grouped as their value is part of the labels anyway
		for g := range groupingMap {
			delete(ruleLabels, g)
		}

		for _, br := range burnrates {
			rules = append(rules, monitoringv1.Rule{
				Record: o.BurnrateName(br),
				Expr:   intstr.FromString(o.Burnrate(br, opts)),
				Labels: ruleLabels,
			})
		}

		if o.Alerting.Disabled || !o.Alerting.Burnrates {
			return monitoringv1.RuleGroup{
				Name:     sloName,
				Interval: monitoringDuration("30s"), // TODO: Increase or decrease based on availability target
				Rules:    rules,
			}, nil
		}

		var alertMatchers []string
		for _, m := range matchers {
			if m.Name == model.MetricNameLabel {
				continue
			}
			if _, ok := groupingMap[m.Name]; !ok {
				if m.Type == labels.MatchRegexp || m.Type == labels.MatchNotRegexp {
					continue
				}
			}

			alertMatchers = append(alertMatchers, m.String())
		}
		alertMatchers = append(alertMatchers, fmt.Sprintf(`slo="%s"`, sloName))
		sort.Strings(alertMatchers)
		alertMatchersString := strings.Join(alertMatchers, ",")

		for i, w := range ws {
			alertLabels := o.commonRuleLabels(sloName)
			alertAnnotations := o.commonRuleAnnotations(externalURL)
			for _, m := range matchers {
				if m.Type == labels.MatchEqual && m.Name != model.MetricNameLabel {
					if _, ok := groupingMap[m.Name]; !ok { // only add labels that aren't grouped by
						alertLabels[m.Name] = m.Value
					}
				}
			}

			// Propagate useful SLO information to alerts' labels
			alertLabels["short"] = model.Duration(w.Short).String()
			alertLabels["long"] = model.Duration(w.Long).String()
			alertLabels["severity"] = o.alertSeverityLabel(i, w)
			alertLabels["exhaustion"] = o.Exhausts(w.Factor).String()

			r := monitoringv1.Rule{
				Alert: o.AlertName(),
				// TODO: Use expr replacer
				Expr: intstr.FromString(fmt.Sprintf("%s{%s} > (%.f * (1-%s)) and %s{%s} > (%.f * (1-%s))",
					o.BurnrateName(w.Short),
					alertMatchersString,
					w.Factor,
					strconv.FormatFloat(o.Target, 'f', -1, 64),
					o.BurnrateName(w.Long),
					alertMatchersString,
					w.Factor,
					strconv.FormatFloat(o.Target, 'f', -1, 64),
				)),
				For:         monitoringDuration(model.Duration(w.For).String()),
				Labels:      alertLabels,
				Annotations: alertAnnotations,
			}
			rules = append(rules, r)
		}
	}

	// We only get here if alerting was not disabled
	return monitoringv1.RuleGroup{
		Name:     sloName,
		Interval: monitoringDuration("30s"), // TODO: Increase or decrease based on availability target
		Rules:    rules,
	}, nil
}

func (o Objective) BurnrateName(rate time.Duration) string {
	var metric string

	switch o.IndicatorType() {
	case Ratio:
		metric = o.Indicator.Ratio.Total.Name
	case Latency:
		metric = o.Indicator.Latency.Total.Name
	case LatencyNative:
		metric = o.Indicator.LatencyNative.Total.Name
	case BoolGauge:
		metric = o.Indicator.BoolGauge.Name
	}

	metric = strings.TrimSuffix(metric, "_total")
	metric = strings.TrimSuffix(metric, "_count")

	return fmt.Sprintf("%s:burnrate%s", metric, model.Duration(rate))
}

func (o Objective) Burnrate(timerange time.Duration, opts GenerationOptions) string {
	switch o.IndicatorType() {
	case Ratio:
		expr, err := parser.ParseExpr(`sum by (grouping) (rate(errorMetric{matchers="errors"}[1s])) / sum by (grouping) (rate(metric{matchers="total"}[1s]))`)
		if err != nil {
			return err.Error()
		}

		groupingMap := map[string]struct{}{}
		for _, s := range o.Indicator.Ratio.Grouping {
			groupingMap[s] = struct{}{}
		}

		grouping := make([]string, 0, len(groupingMap))
		for s := range groupingMap {
			grouping = append(grouping, s)
		}
		sort.Strings(grouping)

		objectiveReplacer{
			metric:        o.Indicator.Ratio.Total.Name,
			matchers:      o.Indicator.Ratio.Total.LabelMatchers,
			errorMetric:   o.Indicator.Ratio.Errors.Name,
			errorMatchers: o.Indicator.Ratio.Errors.LabelMatchers,
			grouping:      grouping,
			window:        timerange,
		}.replace(expr)

		return expr.String()
	case Latency:
		query := `
			(
				sum by (grouping) (rate(metric{matchers="total"}[1s]))
				-
				sum by (grouping) (rate(errorMetric{matchers="errors"}[1s]))
			)
			/
			sum by (grouping) (rate(metric{matchers="total"}[1s]))
`
		expr, err := parser.ParseExpr(query)
		if err != nil {
			return err.Error()
		}

		groupingMap := map[string]struct{}{}
		for _, s := range o.Indicator.Latency.Grouping {
			groupingMap[s] = struct{}{}
		}

		grouping := make([]string, 0, len(groupingMap))
		for s := range groupingMap {
			grouping = append(grouping, s)
		}
		sort.Strings(grouping)

		objectiveReplacer{
			metric:        o.Indicator.Latency.Total.Name,
			matchers:      applyPrometheus3Migration(o.Indicator.Latency.Total.LabelMatchers, opts),
			errorMetric:   o.Indicator.Latency.Success.Name,
			errorMatchers: applyPrometheus3Migration(o.Indicator.Latency.Success.LabelMatchers, opts),
			grouping:      grouping,
			window:        timerange,
		}.replace(expr)

		return expr.String()
	case LatencyNative:
		expr, err := parser.ParseExpr(`1 - histogram_fraction(0,0.696969, sum by (grouping) (rate(metric{matchers="total"}[1s])))`)
		if err != nil {
			return err.Error()
		}

		groupingMap := map[string]struct{}{}
		for _, s := range o.Indicator.LatencyNative.Grouping {
			groupingMap[s] = struct{}{}
		}

		grouping := make([]string, 0, len(groupingMap))
		for s := range groupingMap {
			grouping = append(grouping, s)
		}
		sort.Strings(grouping)

		objectiveReplacer{
			metric:   o.Indicator.LatencyNative.Total.Name,
			matchers: o.Indicator.LatencyNative.Total.LabelMatchers,
			grouping: grouping,
			target:   time.Duration(o.Indicator.LatencyNative.Latency).Seconds(),
			window:   timerange,
		}.replace(expr)

		return expr.String()
	case BoolGauge:
		query := `
			(
				sum by (grouping) (count_over_time(metric{matchers="total"}[1s]))
				-
				sum by (grouping) (sum_over_time(metric{matchers="total"}[1s]))
			)
			/
			sum by (grouping) (count_over_time(metric{matchers="total"}[1s]))
`
		expr, err := parser.ParseExpr(query)
		if err != nil {
			return err.Error()
		}

		groupingMap := map[string]struct{}{}
		for _, s := range o.Indicator.BoolGauge.Grouping {
			groupingMap[s] = struct{}{}
		}

		grouping := make([]string, 0, len(groupingMap))
		for s := range groupingMap {
			grouping = append(grouping, s)
		}
		sort.Strings(grouping)

		objectiveReplacer{
			metric:   o.Indicator.BoolGauge.Name,
			matchers: o.Indicator.BoolGauge.LabelMatchers,
			grouping: grouping,
			window:   timerange,
		}.replace(expr)

		return expr.String()
	default:
		return ""
	}
}

func sumName(metric string, window model.Duration) string {
	return fmt.Sprintf("%s:sum%s", metric, window)
}

func countName(metric string, window model.Duration) string {
	return fmt.Sprintf("%s:count%s", metric, window)
}

func increaseName(metric string, window model.Duration) string {
	metric = strings.TrimSuffix(metric, "_total")
	metric = strings.TrimSuffix(metric, "_count")
	metric = strings.TrimSuffix(metric, "_bucket")
	return fmt.Sprintf("%s:increase%s", metric, window)
}

func (o Objective) commonRuleLabels(sloName string) map[string]string {
	ruleLabels := map[string]string{
		"slo": sloName,
	}

	o.Labels.Range(func(label labels.Label) {
		if strings.HasPrefix(label.Name, PropagationLabelsPrefix) {
			ruleLabels[strings.TrimPrefix(label.Name, PropagationLabelsPrefix)] = label.Value
		}
	})

	return ruleLabels
}

func (o Objective) commonRuleAnnotations(externalURL string) map[string]string {
	var annotations map[string]string

	// Add existing annotations from the objective
	for key, value := range o.Annotations {
		if strings.HasPrefix(key, PropagationLabelsPrefix) {
			if annotations == nil {
				annotations = make(map[string]string)
			}
			annotations[strings.TrimPrefix(key, PropagationLabelsPrefix)] = value
		}
	}

	// Add External URL if provided
	if externalURL != "" {
		sloName := o.Labels.Get(model.MetricNameLabel)
		if sloName != "" {
			if annotations == nil {
				annotations = make(map[string]string)
			}

			externalURLParsed, err := url.Parse(externalURL)
			if err != nil {
				fmt.Printf("Error parsing external URL: %v\n", err)
				// Continue without adding pyrra_url
				return annotations
			}

			params := url.Values{}
			params.Add("expr", `{__name__="`+sloName+`"}`)
			params.Add("from", "now-1h")
			params.Add("to", "now")

			externalURLParsed.Path = "/objectives"
			externalURLParsed.RawQuery = params.Encode()

			annotations["pyrra_url"] = externalURLParsed.String()

			if grouping := o.Grouping(); len(grouping) > 0 {
				groups := make([]string, 0, len(grouping))
				for _, l := range grouping {
					// Use a placeholder for the template expression to prevent url.QueryEscape from encoding it.
					// Prometheus needs {{$labels.<name>}} as-is to template the actual value at alert time.
					groups = append(groups, l+`="__TPL_`+l+`__"`)
				}
				groupingParam := url.QueryEscape("{" + strings.Join(groups, ",") + "}")
				for _, l := range grouping {
					groupingParam = strings.ReplaceAll(groupingParam, "__TPL_"+l+"__", "{{$labels."+l+"}}")
				}
				annotations["pyrra_url"] += "&grouping=" + groupingParam
			}
		}
	}

	return annotations
}

func (o Objective) countExpr() (parser.Expr, error) { // Returns a new instance of Expr with this query each time called
	return parser.ParseExpr(`sum by (grouping) (count_over_time(metric{matchers="total"}[1s]))`)
}

func (o Objective) sumExpr() (parser.Expr, error) { // Returns a new instance of Expr with this query each time called
	return parser.ParseExpr(`sum by (grouping) (sum_over_time(metric{matchers="total"}[1s]))`)
}

func (o Objective) increaseExpr() (parser.Expr, error) { // Returns a new instance of Expr with this query each time called
	return parser.ParseExpr(`sum by (grouping) (increase(metric{matchers="total"}[1s]))`)
}

func (o Objective) increaseSubqueryExpr() (parser.Expr, error) { // Returns a new instance of Expr with this query each time called
	return parser.ParseExpr(`sum by (grouping) (sum_over_time(metric{matchers="total"}[1s:2s]))`)
}

func (o Objective) absentExpr() (parser.Expr, error) {
	return parser.ParseExpr(`absent(metric{matchers="total"}) == 1`)
}

func (o Objective) increaseInterval() model.Duration {
	day := 24 * time.Hour
	window := time.Duration(o.Window)

	// TODO: Make this a function with an equation
	if window < 7*day {
		return model.Duration(30 * time.Second)
	} else if window < 14*day {
		return model.Duration(60 * time.Second)
	} else if window < 21*day {
		return model.Duration(90 * time.Second)
	} else if window < 28*day {
		return model.Duration(120 * time.Second)
	} else if window < 35*day {
		return model.Duration(150 * time.Second)
	} else if window < 42*day {
		return model.Duration(180 * time.Second)
	} else if window < 49*day {
		return model.Duration(210 * time.Second)
	}
	return model.Duration(240 * time.Second) // 8w+
}

// IncreaseRules returns a single RuleGroup with all increase rules.
func (o Objective) IncreaseRules(opts GenerationOptions) (monitoringv1.RuleGroup, error) {
	sloName := o.Labels.Get(model.MetricNameLabel)

	shortRules, longRules, err := o.splitIncreaseRulesForType(sloName, opts)
	if err != nil {
		return monitoringv1.RuleGroup{}, err
	}

	rules := append(shortRules, longRules...)

	return monitoringv1.RuleGroup{
		Name:     sloName + "-increase",
		Interval: monitoringDuration(o.increaseInterval().String()),
		Rules:    rules,
	}, nil
}

// SplitIncreaseRules returns two separate rule groups when PerformanceOverAccuracy is enabled:
//   - short: 5m increase recording rules and absent alerts (run on Prometheus)
//   - long: subquery rules over the full window (run on Thanos)
//
// When PerformanceOverAccuracy is false, short is empty and long contains all rules.
func (o Objective) SplitIncreaseRules(opts GenerationOptions) (short, long monitoringv1.RuleGroup, err error) {
	sloName := o.Labels.Get(model.MetricNameLabel)

	shortRules, longRules, err := o.splitIncreaseRulesForType(sloName, opts)
	if err != nil {
		return monitoringv1.RuleGroup{}, monitoringv1.RuleGroup{}, err
	}

	interval := monitoringDuration(o.increaseInterval().String())

	if len(shortRules) > 0 {
		short = monitoringv1.RuleGroup{
			Name:     sloName + "-increase",
			Interval: monitoringDuration("30s"),
			Rules:    shortRules,
		}
	}

	long = monitoringv1.RuleGroup{
		Name:     sloName + "-increase",
		Interval: interval,
		Rules:    longRules,
	}

	return short, long, nil
}

func (o Objective) splitIncreaseRulesForType(sloName string, opts GenerationOptions) (shortRules, longRules []monitoringv1.Rule, err error) {
	switch o.IndicatorType() {
	case Unknown:
		return nil, nil, nil
	case Ratio:
		return o.increaseRulesRatio(sloName)
	case Latency:
		return o.increaseRuleLatency(sloName, opts)
	case LatencyNative:
		rules, err := o.increaseRuleLatencyNative(sloName)
		return nil, rules, err
	case BoolGauge:
		return o.increaseRuleBoolGauge(sloName)
	}
	return nil, nil, nil
}

func (o Objective) increaseRulesRatio(sloName string) (shortRules, longRules []monitoringv1.Rule, err error) {
	ruleLabels := o.commonRuleLabels(sloName)
	for _, m := range o.Indicator.Ratio.Total.LabelMatchers {
		if m.Type == labels.MatchEqual && m.Name != model.MetricNameLabel {
			ruleLabels[m.Name] = m.Value
		}
	}

	groupingMap := map[string]struct{}{}
	for _, s := range o.Indicator.Ratio.Grouping {
		groupingMap[s] = struct{}{}
	}
	for _, s := range groupingLabels(
		o.Indicator.Ratio.Errors.LabelMatchers,
		o.Indicator.Ratio.Total.LabelMatchers,
	) {
		groupingMap[s] = struct{}{}
	}
	for _, m := range o.Indicator.Ratio.Total.LabelMatchers {
		if m.Type == labels.MatchRegexp || m.Type == labels.MatchNotRegexp {
			groupingMap[m.Name] = struct{}{}
		}
	}
	for g := range groupingMap {
		delete(ruleLabels, g)
	}

	grouping := make([]string, 0, len(groupingMap))
	for s := range groupingMap {
		grouping = append(grouping, s)
	}
	sort.Strings(grouping)

	expr, err := o.increaseExpr()
	if err != nil {
		return nil, nil, err
	}

	if o.PerformanceOverAccuracy {
		objectiveReplacer{
			metric:   o.Indicator.Ratio.Total.Name,
			matchers: o.Indicator.Ratio.Total.LabelMatchers,
			grouping: grouping,
			window:   5 * time.Minute,
		}.replace(expr)

		subqueryName := increaseName(o.Indicator.Ratio.Total.Name, model.Duration(5*time.Minute))

		// Short rule: increase(metric[5m])
		shortRules = append(shortRules, monitoringv1.Rule{
			Record: subqueryName,
			Expr:   intstr.FromString(expr.String()),
			Labels: ruleLabels,
		})

		// Long rule: sum_over_time(metric:increase5m[window:5m])
		subExpr, err := o.increaseSubqueryExpr()
		if err != nil {
			return nil, nil, err
		}

		subqueryLabelMatchers := o.buildSubqueryMatchers(o.Indicator.Ratio.Total.LabelMatchers, subqueryName)
		objectiveReplacer{
			metric:   subqueryName,
			matchers: subqueryLabelMatchers,
			grouping: grouping,
			window:   time.Duration(o.Window),
		}.replace(subExpr)

		longRules = append(longRules, monitoringv1.Rule{
			Record: increaseName(o.Indicator.Ratio.Total.Name, o.Window),
			Expr:   intstr.FromString(subExpr.String()),
			Labels: ruleLabels,
		})
	} else {
		objectiveReplacer{
			metric:   o.Indicator.Ratio.Total.Name,
			matchers: o.Indicator.Ratio.Total.LabelMatchers,
			grouping: grouping,
			window:   time.Duration(o.Window),
		}.replace(expr)

		longRules = append(longRules, monitoringv1.Rule{
			Record: increaseName(o.Indicator.Ratio.Total.Name, o.Window),
			Expr:   intstr.FromString(expr.String()),
			Labels: ruleLabels,
		})
	}

	alertLabels := make(map[string]string, len(ruleLabels)+1)
	for k, v := range ruleLabels {
		alertLabels[k] = v
	}
	alertLabels["severity"] = o.alertSeverityLabelAbsent()

	// Absent alerts go on short rules (they reference the raw metric)
	if o.Alerting.Absent {
		expr, err = o.absentExpr()
		if err != nil {
			return nil, nil, err
		}

		objectiveReplacer{
			metric:   o.Indicator.Ratio.Total.Name,
			matchers: o.Indicator.Ratio.Total.LabelMatchers,
		}.replace(expr)

		absentRule := monitoringv1.Rule{
			Alert:       o.AlertNameAbsent(),
			Expr:        intstr.FromString(expr.String()),
			For:         monitoringDuration(o.AbsentDuration().String()),
			Labels:      alertLabels,
			Annotations: o.commonRuleAnnotations(""),
		}
		if o.PerformanceOverAccuracy {
			shortRules = append(shortRules, absentRule)
		} else {
			longRules = append(longRules, absentRule)
		}
	}

	if o.Indicator.Ratio.Total.Name != o.Indicator.Ratio.Errors.Name {
		expr, err := o.increaseExpr()
		if err != nil {
			return nil, nil, err
		}

		if o.PerformanceOverAccuracy {
			// Short rule: increase(errors[5m])
			objectiveReplacer{
				metric:   o.Indicator.Ratio.Errors.Name,
				matchers: o.Indicator.Ratio.Errors.LabelMatchers,
				grouping: grouping,
				window:   5 * time.Minute,
			}.replace(expr)

			subqueryName := increaseName(o.Indicator.Ratio.Errors.Name, model.Duration(5*time.Minute))
			shortRules = append(shortRules, monitoringv1.Rule{
				Record: subqueryName,
				Expr:   intstr.FromString(expr.String()),
				Labels: ruleLabels,
			})

			// Long rule: sum_over_time(errors:increase5m[window:5m])
			subExpr, err := o.increaseSubqueryExpr()
			if err != nil {
				return nil, nil, err
			}

			subqueryLabelMatchers := o.buildSubqueryMatchers(o.Indicator.Ratio.Errors.LabelMatchers, subqueryName)
			objectiveReplacer{
				metric:   subqueryName,
				matchers: subqueryLabelMatchers,
				grouping: grouping,
				window:   time.Duration(o.Window),
			}.replace(subExpr)

			longRules = append(longRules, monitoringv1.Rule{
				Record: increaseName(o.Indicator.Ratio.Errors.Name, o.Window),
				Expr:   intstr.FromString(subExpr.String()),
				Labels: ruleLabels,
			})
		} else {
			objectiveReplacer{
				metric:   o.Indicator.Ratio.Errors.Name,
				matchers: o.Indicator.Ratio.Errors.LabelMatchers,
				grouping: grouping,
				window:   time.Duration(o.Window),
			}.replace(expr)

			longRules = append(longRules, monitoringv1.Rule{
				Record: increaseName(o.Indicator.Ratio.Errors.Name, o.Window),
				Expr:   intstr.FromString(expr.String()),
				Labels: ruleLabels,
			})
		}

		if o.Alerting.Absent {
			expr, err = o.absentExpr()
			if err != nil {
				return nil, nil, err
			}

			objectiveReplacer{
				metric:   o.Indicator.Ratio.Errors.Name,
				matchers: o.Indicator.Ratio.Errors.LabelMatchers,
			}.replace(expr)

			absentRule := monitoringv1.Rule{
				Alert:       o.AlertNameAbsent(),
				Expr:        intstr.FromString(expr.String()),
				For:         monitoringDuration(o.AbsentDuration().String()),
				Labels:      alertLabels,
				Annotations: o.commonRuleAnnotations(""),
			}
			if o.PerformanceOverAccuracy {
				shortRules = append(shortRules, absentRule)
			} else {
				longRules = append(longRules, absentRule)
			}
		}
	}

	return shortRules, longRules, nil
}

// buildSubqueryMatchers creates label matchers for the subquery recording rule,
// replacing the metric name with the subquery recording rule name.
func (o Objective) buildSubqueryMatchers(original []*labels.Matcher, subqueryName string) []*labels.Matcher {
	matchers := make([]*labels.Matcher, 0, len(original))
	for _, m := range original {
		value := m.Value
		if m.Name == model.MetricNameLabel {
			value = subqueryName
		}
		matchers = append(matchers, &labels.Matcher{
			Type:  m.Type,
			Name:  m.Name,
			Value: value,
		})
	}
	return matchers
}

func (o Objective) increaseRuleLatency(sloName string, opts GenerationOptions) (shortRules, longRules []monitoringv1.Rule, err error) {
	ruleLabels := o.commonRuleLabels(sloName)
	for _, m := range o.Indicator.Latency.Total.LabelMatchers {
		if m.Type == labels.MatchEqual && m.Name != model.MetricNameLabel {
			ruleLabels[m.Name] = m.Value
		}
	}

	groupingMap := map[string]struct{}{}
	for _, s := range o.Indicator.Latency.Grouping {
		groupingMap[s] = struct{}{}
	}
	for _, s := range groupingLabels(
		o.Indicator.Latency.Success.LabelMatchers,
		o.Indicator.Latency.Total.LabelMatchers,
	) {
		groupingMap[s] = struct{}{}
	}
	for _, m := range o.Indicator.Latency.Total.LabelMatchers {
		if m.Type == labels.MatchRegexp || m.Type == labels.MatchNotRegexp {
			groupingMap[m.Name] = struct{}{}
		}
	}
	for g := range groupingMap {
		delete(ruleLabels, g)
	}

	grouping := make([]string, 0, len(groupingMap))
	for s := range groupingMap {
		grouping = append(grouping, s)
	}
	sort.Strings(grouping)

	var le string
	for _, m := range o.Indicator.Latency.Success.LabelMatchers {
		if m.Name == "le" {
			le = m.Value
			break
		}
	}
	ruleLabelsLe := map[string]string{"le": le}
	for k, v := range ruleLabels {
		ruleLabelsLe[k] = v
	}

	window := time.Duration(o.Window)
	if o.PerformanceOverAccuracy {
		window = 5 * time.Minute
	}

	// Total metric increase rule
	expr, err := o.increaseExpr()
	if err != nil {
		return nil, nil, err
	}

	objectiveReplacer{
		metric:   o.Indicator.Latency.Total.Name,
		matchers: applyPrometheus3Migration(o.Indicator.Latency.Total.LabelMatchers, opts),
		grouping: grouping,
		window:   window,
	}.replace(expr)

	totalRule := monitoringv1.Rule{
		Record: increaseName(o.Indicator.Latency.Total.Name, model.Duration(window)),
		Expr:   intstr.FromString(expr.String()),
		Labels: ruleLabels,
	}

	// Success metric increase rule
	expr, err = o.increaseExpr()
	if err != nil {
		return nil, nil, err
	}

	objectiveReplacer{
		metric:   o.Indicator.Latency.Success.Name,
		matchers: applyPrometheus3Migration(o.Indicator.Latency.Success.LabelMatchers, opts),
		grouping: grouping,
		window:   window,
	}.replace(expr)

	successRule := monitoringv1.Rule{
		Record: increaseName(o.Indicator.Latency.Success.Name, model.Duration(window)),
		Expr:   intstr.FromString(expr.String()),
		Labels: ruleLabelsLe,
	}

	if o.PerformanceOverAccuracy {
		shortRules = append(shortRules, totalRule, successRule)

		// Long rule: sum_over_time for total
		subExpr, err := o.increaseSubqueryExpr()
		if err != nil {
			return nil, nil, err
		}

		subqueryName := increaseName(o.Indicator.Latency.Total.Name, model.Duration(5*time.Minute))
		objectiveReplacer{
			metric:   subqueryName,
			matchers: o.buildSubqueryMatchers(applyPrometheus3Migration(o.Indicator.Latency.Total.LabelMatchers, opts), subqueryName),
			grouping: grouping,
			window:   time.Duration(o.Window),
		}.replace(subExpr)

		longRules = append(longRules, monitoringv1.Rule{
			Record: increaseName(o.Indicator.Latency.Total.Name, o.Window),
			Expr:   intstr.FromString(subExpr.String()),
			Labels: ruleLabels,
		})

		// Long rule: sum_over_time for success
		subExpr, err = o.increaseSubqueryExpr()
		if err != nil {
			return nil, nil, err
		}

		subqueryName = increaseName(o.Indicator.Latency.Success.Name, model.Duration(5*time.Minute))
		objectiveReplacer{
			metric:   subqueryName,
			matchers: o.buildSubqueryMatchers(applyPrometheus3Migration(o.Indicator.Latency.Success.LabelMatchers, opts), subqueryName),
			grouping: grouping,
			window:   time.Duration(o.Window),
		}.replace(subExpr)

		longRules = append(longRules, monitoringv1.Rule{
			Record: increaseName(o.Indicator.Latency.Success.Name, o.Window),
			Expr:   intstr.FromString(subExpr.String()),
			Labels: ruleLabelsLe,
		})
	} else {
		longRules = append(longRules, totalRule, successRule)
	}

	// Absent alerts go on short rules when split, long rules otherwise
	if o.Alerting.Absent {
		alertLabels := make(map[string]string, len(ruleLabels)+1)
		for k, v := range ruleLabels {
			alertLabels[k] = v
		}
		alertLabels["severity"] = o.alertSeverityLabelAbsent()

		expr, err = o.absentExpr()
		if err != nil {
			return nil, nil, err
		}

		objectiveReplacer{
			metric:   o.Indicator.Latency.Total.Name,
			matchers: applyPrometheus3Migration(o.Indicator.Latency.Total.LabelMatchers, opts),
		}.replace(expr)

		absentTotalRule := monitoringv1.Rule{
			Alert:       o.AlertNameAbsent(),
			Expr:        intstr.FromString(expr.String()),
			For:         monitoringDuration(o.AbsentDuration().String()),
			Labels:      alertLabels,
			Annotations: o.commonRuleAnnotations(""),
		}

		expr, err = o.absentExpr()
		if err != nil {
			return nil, nil, err
		}

		objectiveReplacer{
			metric:   o.Indicator.Latency.Success.Name,
			matchers: applyPrometheus3Migration(o.Indicator.Latency.Success.LabelMatchers, opts),
		}.replace(expr)

		alertLabelsLe := make(map[string]string, len(ruleLabelsLe)+1)
		for k, v := range ruleLabelsLe {
			alertLabelsLe[k] = v
		}
		alertLabelsLe["severity"] = o.alertSeverityLabelAbsent()

		absentSuccessRule := monitoringv1.Rule{
			Alert:       o.AlertNameAbsent(),
			Expr:        intstr.FromString(expr.String()),
			For:         monitoringDuration(o.AbsentDuration().String()),
			Labels:      alertLabelsLe,
			Annotations: o.commonRuleAnnotations(""),
		}

		if o.PerformanceOverAccuracy {
			shortRules = append(shortRules, absentTotalRule, absentSuccessRule)
		} else {
			longRules = append(longRules, absentTotalRule, absentSuccessRule)
		}
	}

	return shortRules, longRules, nil
}

func (o Objective) increaseRuleLatencyNative(sloName string) ([]monitoringv1.Rule, error) {
	var rules []monitoringv1.Rule

	ruleLabels := o.commonRuleLabels(sloName)
	for _, m := range o.Indicator.LatencyNative.Total.LabelMatchers {
		if m.Type == labels.MatchEqual && m.Name != model.MetricNameLabel {
			ruleLabels[m.Name] = m.Value
		}
	}

	expr, err := parser.ParseExpr(`histogram_count(sum by (grouping) (increase(metric{matchers="total"}[1s])))`)
	if err != nil {
		return rules, err
	}

	objectiveReplacer{
		metric:   o.Indicator.LatencyNative.Total.Name,
		matchers: slices.Clone(o.Indicator.LatencyNative.Total.LabelMatchers),
		grouping: slices.Clone(o.Indicator.LatencyNative.Grouping),
		window:   time.Duration(o.Window),
	}.replace(expr)

	rules = append(rules, monitoringv1.Rule{
		Record: increaseName(o.Indicator.LatencyNative.Total.Name, o.Window),
		Expr:   intstr.FromString(expr.String()),
		Labels: ruleLabels,
	})

	expr, err = parser.ParseExpr(`histogram_fraction(0, 0.696969, sum by (grouping) (increase(metric{matchers="total"}[1s]))) * histogram_count(sum by (grouping) (increase(metric{matchers="total"}[1s])))`)
	if err != nil {
		return rules, err
	}

	latencySeconds := time.Duration(o.Indicator.LatencyNative.Latency).Seconds()
	objectiveReplacer{
		metric:   o.Indicator.LatencyNative.Total.Name,
		matchers: slices.Clone(o.Indicator.LatencyNative.Total.LabelMatchers),
		grouping: slices.Clone(o.Indicator.LatencyNative.Grouping),
		window:   time.Duration(o.Window),
		target:   latencySeconds,
	}.replace(expr)

	ruleLabels = maps.Clone(ruleLabels)
	ruleLabels["le"] = fmt.Sprintf("%g", latencySeconds)

	rules = append(rules, monitoringv1.Rule{
		Record: increaseName(o.Indicator.LatencyNative.Total.Name, o.Window),
		Expr:   intstr.FromString(expr.String()),
		Labels: ruleLabels,
	})

	return rules, nil
}

func (o Objective) increaseRuleBoolGauge(sloName string) (shortRules, longRules []monitoringv1.Rule, err error) {
	ruleLabels := o.commonRuleLabels(sloName)
	for _, m := range o.Indicator.BoolGauge.LabelMatchers {
		if m.Type == labels.MatchEqual && m.Name != model.MetricNameLabel {
			ruleLabels[m.Name] = m.Value
		}
	}

	groupingMap := map[string]struct{}{}
	for _, s := range o.Indicator.BoolGauge.Grouping {
		groupingMap[s] = struct{}{}
	}
	for _, s := range o.Indicator.BoolGauge.LabelMatchers {
		groupingMap[s.Name] = struct{}{}
	}
	for _, m := range o.Indicator.BoolGauge.LabelMatchers {
		if m.Type == labels.MatchRegexp || m.Type == labels.MatchNotRegexp {
			groupingMap[m.Name] = struct{}{}
		}
	}
	// Delete labels that are grouped, as their value is part of the recording rule anyway
	for g := range groupingMap {
		delete(ruleLabels, g)
	}

	grouping := make([]string, 0, len(groupingMap))
	for s := range groupingMap {
		grouping = append(grouping, s)
	}
	sort.Strings(grouping)

	count, err := o.countExpr()
	if err != nil {
		return nil, nil, err
	}

	sum, err := o.sumExpr()
	if err != nil {
		return nil, nil, err
	}

	if o.PerformanceOverAccuracy {
		// Short rules: 5m window count_over_time + sum_over_time
		objectiveReplacer{
			metric:   o.Indicator.BoolGauge.Name,
			matchers: o.Indicator.BoolGauge.LabelMatchers,
			grouping: grouping,
			window:   5 * time.Minute,
		}.replace(count)

		objectiveReplacer{
			metric:   o.Indicator.BoolGauge.Name,
			matchers: o.Indicator.BoolGauge.LabelMatchers,
			grouping: grouping,
			window:   5 * time.Minute,
		}.replace(sum)

		countSubqueryName := countName(o.Indicator.BoolGauge.Name, model.Duration(5*time.Minute))
		sumSubqueryName := sumName(o.Indicator.BoolGauge.Name, model.Duration(5*time.Minute))

		shortRules = append(shortRules, monitoringv1.Rule{
			Record: countSubqueryName,
			Expr:   intstr.FromString(count.String()),
			Labels: ruleLabels,
		})

		shortRules = append(shortRules, monitoringv1.Rule{
			Record: sumSubqueryName,
			Expr:   intstr.FromString(sum.String()),
			Labels: ruleLabels,
		})

		// Long rules: sum_over_time subqueries over the 5m recording rules
		countSubExpr, err := o.increaseSubqueryExpr()
		if err != nil {
			return nil, nil, err
		}
		objectiveReplacer{
			metric:   countSubqueryName,
			matchers: o.buildSubqueryMatchers(o.Indicator.BoolGauge.LabelMatchers, countSubqueryName),
			grouping: grouping,
			window:   time.Duration(o.Window),
		}.replace(countSubExpr)

		longRules = append(longRules, monitoringv1.Rule{
			Record: countName(o.Indicator.BoolGauge.Name, o.Window),
			Expr:   intstr.FromString(countSubExpr.String()),
			Labels: ruleLabels,
		})

		sumSubExpr, err := o.increaseSubqueryExpr()
		if err != nil {
			return nil, nil, err
		}
		objectiveReplacer{
			metric:   sumSubqueryName,
			matchers: o.buildSubqueryMatchers(o.Indicator.BoolGauge.LabelMatchers, sumSubqueryName),
			grouping: grouping,
			window:   time.Duration(o.Window),
		}.replace(sumSubExpr)

		longRules = append(longRules, monitoringv1.Rule{
			Record: sumName(o.Indicator.BoolGauge.Name, o.Window),
			Expr:   intstr.FromString(sumSubExpr.String()),
			Labels: ruleLabels,
		})
	} else {
		objectiveReplacer{
			metric:   o.Indicator.BoolGauge.Name,
			matchers: o.Indicator.BoolGauge.LabelMatchers,
			grouping: grouping,
			window:   time.Duration(o.Window),
		}.replace(count)

		objectiveReplacer{
			metric:   o.Indicator.BoolGauge.Name,
			matchers: o.Indicator.BoolGauge.LabelMatchers,
			grouping: grouping,
			window:   time.Duration(o.Window),
		}.replace(sum)

		longRules = append(longRules, monitoringv1.Rule{
			Record: countName(o.Indicator.BoolGauge.Name, o.Window),
			Expr:   intstr.FromString(count.String()),
			Labels: ruleLabels,
		})

		longRules = append(longRules, monitoringv1.Rule{
			Record: sumName(o.Indicator.BoolGauge.Name, o.Window),
			Expr:   intstr.FromString(sum.String()),
			Labels: ruleLabels,
		})
	}

	if o.Alerting.Absent {
		expr, err := o.absentExpr()
		if err != nil {
			return nil, nil, err
		}

		objectiveReplacer{
			metric:   o.Indicator.BoolGauge.Name,
			matchers: o.Indicator.BoolGauge.LabelMatchers,
		}.replace(expr)

		alertLabels := make(map[string]string, len(ruleLabels)+1)
		for k, v := range ruleLabels {
			alertLabels[k] = v
		}
		// Add severity label for alerts
		alertLabels["severity"] = o.alertSeverityLabelAbsent()

		absentRule := monitoringv1.Rule{
			Alert:       o.AlertNameAbsent(),
			Expr:        intstr.FromString(expr.String()),
			For:         monitoringDuration(o.AbsentDuration().String()),
			Labels:      alertLabels,
			Annotations: o.commonRuleAnnotations(""),
		}
		if o.PerformanceOverAccuracy {
			shortRules = append(shortRules, absentRule)
		} else {
			longRules = append(longRules, absentRule)
		}
	}

	return shortRules, longRules, nil
}

type severity string

const (
	critical severity = "critical"
	warning  severity = "warning"
)

type Window struct {
	Severity severity
	For      time.Duration
	Long     time.Duration
	Short    time.Duration
	Factor   float64
}

func Windows(sloWindow time.Duration) []Window {
	// TODO: I'm still not sure if For, Long, Short should really be based on the 28 days ratio...

	round := time.Minute // TODO: Change based on sloWindow

	// long and short rates are calculated based on the ratio for 28 days.
	return []Window{{
		Severity: critical,
		For:      (sloWindow / (28 * 24 * (60 / 2))).Round(round), // 2m for 28d - half short
		Long:     (sloWindow / (28 * 24)).Round(round),            // 1h for 28d
		Short:    (sloWindow / (28 * 24 * (60 / 5))).Round(round), // 5m for 28d
		Factor:   14,                                              // error budget burn: 50% within a day
	}, {
		Severity: critical,
		For:      (sloWindow / (28 * 24 * (60 / 15))).Round(round), // 15m for 28d - half short
		Long:     (sloWindow / (28 * (24 / 6))).Round(round),       // 6h for 28d
		Short:    (sloWindow / (28 * 24 * (60 / 30))).Round(round), // 30m for 28d
		Factor:   7,                                                // error budget burn: 20% within a day / 100% within 5 days
	}, {
		Severity: warning,
		For:      (sloWindow / (28 * 24)).Round(round),       // 1h for 28d - half short
		Long:     (sloWindow / 28).Round(round),              // 1d for 28d
		Short:    (sloWindow / (28 * (24 / 2))).Round(round), // 2h for 28d
		Factor:   2,                                          // error budget burn: 10% within a day / 100% within 10 days
	}, {
		Severity: warning,
		For:      (sloWindow / (28 * (24 / 3))).Round(round), // 3h for 28d - half short
		Long:     (sloWindow / 7).Round(round),               // 4d for 28d
		Short:    (sloWindow / (28 * (24 / 6))).Round(round), // 6h for 28d
		Factor:   1,                                          // error budget burn: 100% until the end of sloWindow
	}}
}

func burnratesFromWindows(ws []Window) []time.Duration {
	dedup := map[time.Duration]bool{}
	for _, w := range ws {
		dedup[w.Long] = true
		dedup[w.Short] = true
	}
	burnrates := make([]time.Duration, 0, len(dedup))
	for duration := range dedup {
		burnrates = append(burnrates, duration)
	}

	sort.Slice(burnrates, func(i, j int) bool {
		return burnrates[i].Nanoseconds() < burnrates[j].Nanoseconds()
	})

	return burnrates
}

var ErrGroupingUnsupported = errors.New("objective with grouping not supported in generic rules")

func (o Objective) GenericRules(opts GenerationOptions) (monitoringv1.RuleGroup, error) {
	sloName := o.Labels.Get(model.MetricNameLabel)
	var rules []monitoringv1.Rule

	ruleLabels := o.commonRuleLabels(sloName)

	if o.RuleOutput.EnableDescriptionAsLabel {
		ruleLabels["description"] = o.Description
	}

	rules = append(rules, monitoringv1.Rule{
		Record: "pyrra_objective",
		Expr:   intstr.FromString(strconv.FormatFloat(o.Target, 'f', -1, 64)),
		Labels: ruleLabels,
	})
	rules = append(rules, monitoringv1.Rule{
		Record: "pyrra_window",
		Expr:   intstr.FromString(strconv.FormatInt(int64(time.Duration(o.Window).Seconds()), 10)),
		Labels: ruleLabels,
	})

	switch o.IndicatorType() {
	case Ratio:
		if len(o.Indicator.Ratio.Grouping) > 0 {
			return monitoringv1.RuleGroup{}, ErrGroupingUnsupported
		}

		availability, err := parser.ParseExpr(`1 - sum(errorMetric{matchers="errors"} or vector(0)) / sum(metric{matchers="total"})`)
		if err != nil {
			return monitoringv1.RuleGroup{}, err
		}

		totalIncreaseName := increaseName(o.Indicator.Ratio.Total.Name, o.Window)

		// Copy the list of matchers to modify them
		totalMatchers := make([]*labels.Matcher, 0, len(o.Indicator.Ratio.Total.LabelMatchers))
		for _, m := range o.Indicator.Ratio.Total.LabelMatchers {
			value := m.Value
			if m.Name == model.MetricNameLabel {
				value = totalIncreaseName
			}
			totalMatchers = append(totalMatchers, &labels.Matcher{
				Type:  m.Type,
				Name:  m.Name,
				Value: value,
			})
		}

		totalMatchers = append(totalMatchers, &labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  "slo",
			Value: o.Name(),
		})

		errorsIncreaseName := increaseName(o.Indicator.Ratio.Errors.Name, o.Window)

		errorMatchers := make([]*labels.Matcher, 0, len(o.Indicator.Ratio.Errors.LabelMatchers))
		for _, m := range o.Indicator.Ratio.Errors.LabelMatchers {
			value := m.Value
			if m.Name == model.MetricNameLabel {
				value = errorsIncreaseName
			}
			errorMatchers = append(errorMatchers, &labels.Matcher{
				Type:  m.Type,
				Name:  m.Name,
				Value: value,
			})
		}

		errorMatchers = append(errorMatchers, &labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  "slo",
			Value: o.Name(),
		})

		objectiveReplacer{
			metric:        totalIncreaseName,
			matchers:      totalMatchers,
			errorMetric:   errorsIncreaseName,
			errorMatchers: errorMatchers,
		}.replace(availability)

		rules = append(rules, monitoringv1.Rule{
			Record: "pyrra_availability",
			Expr:   intstr.FromString(availability.String()),
			Labels: ruleLabels,
		})

		rate, err := parser.ParseExpr(`sum(rate(metric{matchers="total"}[5m]))`)
		if err != nil {
			return monitoringv1.RuleGroup{}, err
		}

		objectiveReplacer{
			metric:   o.Indicator.Ratio.Total.Name,
			matchers: o.Indicator.Ratio.Total.LabelMatchers,
		}.replace(rate)

		rules = append(rules, monitoringv1.Rule{
			Record: "pyrra_requests:rate5m",
			Expr:   intstr.FromString(rate.String()),
			Labels: ruleLabels,
		})

		errorsExpr := func() (parser.Expr, error) { // Returns a new instance of Expr with this query each time called
			return parser.ParseExpr(`sum(rate(metric{matchers="total"}[5m])) or vector(0)`)
		}
		errorsParsedExpr, err := errorsExpr()
		if err != nil {
			return monitoringv1.RuleGroup{}, err
		}

		objectiveReplacer{
			metric:   o.Indicator.Ratio.Errors.Name,
			matchers: o.Indicator.Ratio.Errors.LabelMatchers,
		}.replace(errorsParsedExpr)

		rules = append(rules, monitoringv1.Rule{
			Record: "pyrra_errors:rate5m",
			Expr:   intstr.FromString(errorsParsedExpr.String()),
			Labels: ruleLabels,
		})
	case Latency:
		if len(o.Indicator.Latency.Grouping) > 0 {
			return monitoringv1.RuleGroup{}, ErrGroupingUnsupported
		}

		// availability
		{
			expr, err := parser.ParseExpr(`sum(errorMetric{matchers="errors"} or vector(0)) / sum(metric{matchers="total"})`)
			if err != nil {
				return monitoringv1.RuleGroup{}, err
			}

			metric := increaseName(o.Indicator.Latency.Total.Name, o.Window)
			matchers := o.Indicator.Latency.Total.LabelMatchers
			for _, m := range matchers {
				if m.Name == model.MetricNameLabel {
					m.Value = metric
					break
				}
			}
			matchers = append(matchers, &labels.Matcher{Type: labels.MatchEqual, Name: "le", Value: ""})
			matchers = append(matchers, &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  "slo",
				Value: o.Name(),
			})
			// Apply Prometheus 3 migration to matchers
			matchers = applyPrometheus3Migration(matchers, opts)

			errorMetric := increaseName(o.Indicator.Latency.Success.Name, o.Window)
			errorMatchers := o.Indicator.Latency.Success.LabelMatchers
			for _, m := range errorMatchers {
				if m.Name == model.MetricNameLabel {
					m.Value = errorMetric
					break
				}
			}
			errorMatchers = append(errorMatchers, &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  "slo",
				Value: o.Name(),
			})
			// Apply Prometheus 3 migration to errorMatchers
			errorMatchers = applyPrometheus3Migration(errorMatchers, opts)

			objectiveReplacer{
				metric:        metric,
				matchers:      matchers,
				errorMetric:   errorMetric,
				errorMatchers: errorMatchers,
				window:        time.Duration(o.Window),
			}.replace(expr)

			rules = append(rules, monitoringv1.Rule{
				Record: "pyrra_availability",
				Expr:   intstr.FromString(expr.String()),
				Labels: ruleLabels,
			})
		}
		// rate
		{
			rate, err := parser.ParseExpr(`sum(rate(metric{matchers="total"}[5m]))`)
			if err != nil {
				return monitoringv1.RuleGroup{}, err
			}

			metric := o.Indicator.Latency.Total.Name
			matchers := o.Indicator.Latency.Total.LabelMatchers
			for _, m := range matchers {
				if m.Name == model.MetricNameLabel {
					m.Value = metric
					break
				}
			}
			objectiveReplacer{
				metric:   metric,
				matchers: applyPrometheus3Migration(matchers, opts),
			}.replace(rate)

			rules = append(rules, monitoringv1.Rule{
				Record: "pyrra_requests:rate5m",
				Expr:   intstr.FromString(rate.String()),
				Labels: ruleLabels,
			})
		}
		// errors
		{
			errorsExpr, err := parser.ParseExpr(`sum(rate(metric{matchers="total"}[5m])) - sum(rate(errorMetric{matchers="errors"}[5m]))`)
			if err != nil {
				return monitoringv1.RuleGroup{}, err
			}

			metric := o.Indicator.Latency.Total.Name
			matchers := o.Indicator.Latency.Total.LabelMatchers
			for _, m := range matchers {
				if m.Name == model.MetricNameLabel {
					m.Value = metric
					break
				}
			}

			errorMetric := o.Indicator.Latency.Success.Name
			errorMatchers := o.Indicator.Latency.Success.LabelMatchers
			for _, m := range errorMatchers {
				if m.Name == model.MetricNameLabel {
					m.Value = errorMetric
					break
				}
			}

			objectiveReplacer{
				metric:        metric,
				matchers:      applyPrometheus3Migration(matchers, opts),
				errorMetric:   errorMetric,
				errorMatchers: applyPrometheus3Migration(errorMatchers, opts),
			}.replace(errorsExpr)

			rules = append(rules, monitoringv1.Rule{
				Record: "pyrra_errors:rate5m",
				Expr:   intstr.FromString(errorsExpr.String()),
				Labels: ruleLabels,
			})
		}
	case LatencyNative:
		latencySeconds := time.Duration(o.Indicator.LatencyNative.Latency).Seconds()

		// availability
		{
			expr, err := parser.ParseExpr(`sum(metric{matchers="errors"} or vector(0)) / sum(metric{matchers="total"})`)
			if err != nil {
				return monitoringv1.RuleGroup{}, err
			}

			metric := increaseName(o.Indicator.LatencyNative.Total.Name, o.Window)
			matchers := o.Indicator.LatencyNative.Total.LabelMatchers
			for _, m := range matchers {
				if m.Name == model.MetricNameLabel {
					m.Value = metric
					break
				}
			}
			matchers = append(matchers, &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  "slo",
				Value: o.Name(),
			})

			errorMatchers := slices.Clone(matchers)
			errorMatchers = append(errorMatchers, &labels.Matcher{Type: labels.MatchEqual, Name: "le", Value: fmt.Sprintf("%g", latencySeconds)})
			matchers = append(matchers, &labels.Matcher{Type: labels.MatchEqual, Name: "le", Value: ""})

			objectiveReplacer{
				metric:        metric,
				matchers:      matchers,
				errorMatchers: errorMatchers,
				window:        time.Duration(o.Window),
			}.replace(expr)

			rules = append(rules, monitoringv1.Rule{
				Record: "pyrra_availability",
				Expr:   intstr.FromString(expr.String()),
				Labels: ruleLabels,
			})
		}

	case BoolGauge:
		if len(o.Indicator.BoolGauge.Grouping) > 0 {
			return monitoringv1.RuleGroup{}, ErrGroupingUnsupported
		}

		totalMetric := countName(o.Indicator.BoolGauge.Name, o.Window)
		totalMatchers := cloneMatchers(o.Indicator.BoolGauge.LabelMatchers)
		for _, m := range totalMatchers {
			if m.Name == model.MetricNameLabel {
				m.Value = totalMetric
				break
			}
		}
		totalMatchers = append(totalMatchers, &labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  "slo",
			Value: o.Name(),
		})

		successMetric := sumName(o.Indicator.BoolGauge.Name, o.Window)
		successMatchers := cloneMatchers(o.Indicator.BoolGauge.LabelMatchers)
		for _, m := range successMatchers {
			if m.Name == model.MetricNameLabel {
				m.Value = successMetric
				break
			}
		}
		successMatchers = append(successMatchers, &labels.Matcher{
			Type:  labels.MatchEqual,
			Name:  "slo",
			Value: o.Name(),
		})

		// availability
		{
			expr, err := parser.ParseExpr(`sum(errorMetric{matchers="errors"}) / sum(metric{matchers="total"})`)
			if err != nil {
				return monitoringv1.RuleGroup{}, err
			}

			objectiveReplacer{
				metric:        totalMetric,
				matchers:      totalMatchers,
				errorMetric:   successMetric,
				errorMatchers: successMatchers,
			}.replace(expr)

			rules = append(rules, monitoringv1.Rule{
				Record: "pyrra_availability",
				Expr:   intstr.FromString(expr.String()),
				Labels: ruleLabels,
			})
		}

		// rate
		{
			rate, err := parser.ParseExpr(`sum(metric{matchers="total"})`)
			if err != nil {
				return monitoringv1.RuleGroup{}, err
			}

			objectiveReplacer{
				metric:   totalMetric,
				matchers: totalMatchers,
			}.replace(rate)

			rules = append(rules, monitoringv1.Rule{
				Record: "pyrra_requests:rate5m",
				Expr:   intstr.FromString(rate.String()),
				Labels: ruleLabels,
			})
		}

		// errors
		{
			rate, err := parser.ParseExpr(`sum(metric{matchers="total"}) - sum(errorMetric{matchers="errors"})`)
			if err != nil {
				return monitoringv1.RuleGroup{}, err
			}

			objectiveReplacer{
				metric:        totalMetric,
				matchers:      totalMatchers,
				errorMetric:   successMetric,
				errorMatchers: successMatchers,
			}.replace(rate)

			rules = append(rules, monitoringv1.Rule{
				Record: "pyrra_errors:rate5m",
				Expr:   intstr.FromString(rate.String()),
				Labels: ruleLabels,
			})
		}
	}

	return monitoringv1.RuleGroup{
		Name:     sloName + "-generic",
		Interval: monitoringDuration("30s"),
		Rules:    rules,
	}, nil
}

func monitoringDuration(d string) *monitoringv1.Duration {
	md := monitoringv1.Duration(d)
	return &md
}

// alertSeverityLabel returns the severity label for the given window index.
// If the severity is not set, it returns the severity from the default window.
func (o Objective) alertSeverityLabel(windowIndex int, w Window) string {
	var v string
	switch windowIndex {
	case 0:
		v = o.Alerting.Severities.FastBurn
	case 1:
		v = o.Alerting.Severities.MediumBurn
	case 2:
		v = o.Alerting.Severities.SlowBurn
	case 3:
		v = o.Alerting.Severities.LongTermBurn
	}
	if v != "" {
		return v
	}

	return string(w.Severity)
}

func (o Objective) alertSeverityLabelAbsent() string {
	if o.Alerting.Severities.Absent != "" {
		return o.Alerting.Severities.Absent
	}
	return string(critical)
}
