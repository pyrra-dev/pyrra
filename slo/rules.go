package slo

import (
	"errors"
	"fmt"
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
			Severity:   string(w.Severity),
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

func (o Objective) Burnrates() (monitoringv1.RuleGroup, error) {
	sloName := o.Labels.Get(labels.MetricName)

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
			if m.Type == labels.MatchEqual && m.Name != labels.MetricName {
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
				Expr:   intstr.FromString(o.Burnrate(br)),
				Labels: ruleLabels,
			})
		}

		if o.Alerting.Disabled {
			return monitoringv1.RuleGroup{
				Name:     sloName,
				Interval: "30s", // TODO: Increase or decrease based on availability target
				Rules:    rules,
			}, nil
		}

		var alertMatchers []string
		for _, m := range matchers {
			if m.Name == labels.MetricName {
				continue
			}
			alertMatchers = append(alertMatchers, m.String())
		}
		alertMatchers = append(alertMatchers, fmt.Sprintf(`slo="%s"`, sloName))
		sort.Strings(alertMatchers)
		alertMatchersString := strings.Join(alertMatchers, ",")

		for _, w := range ws {
			alertLabels := o.commonRuleLabels(sloName)
			alertAnnotations := o.commonRuleAnnotations()
			for _, m := range matchers {
				if m.Type == labels.MatchEqual && m.Name != labels.MetricName {
					if _, ok := groupingMap[m.Name]; !ok { // only add labels that aren't grouped by
						alertLabels[m.Name] = m.Value
					}
				}
			}
			alertLabels["short"] = model.Duration(w.Short).String()
			alertLabels["long"] = model.Duration(w.Long).String()
			alertLabels["severity"] = string(w.Severity)

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
				For:         monitoringv1.Duration(model.Duration(w.For).String()),
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
			if m.Type == labels.MatchEqual && m.Name != labels.MetricName {
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
				Expr:   intstr.FromString(o.Burnrate(br)),
				Labels: ruleLabels,
			})
		}

		if o.Alerting.Disabled {
			return monitoringv1.RuleGroup{
				Name:     sloName,
				Interval: "30s", // TODO: Increase or decrease based on availability target
				Rules:    rules,
			}, nil
		}

		var alertMatchers []string
		for _, m := range matchers {
			if m.Name == labels.MetricName {
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

		for _, w := range ws {
			alertLabels := o.commonRuleLabels(sloName)
			alertAnnotations := o.commonRuleAnnotations()
			for _, m := range matchers {
				if m.Type == labels.MatchEqual && m.Name != labels.MetricName {
					if _, ok := groupingMap[m.Name]; !ok { // only add labels that aren't grouped by
						alertLabels[m.Name] = m.Value
					}
				}
			}
			alertLabels["short"] = model.Duration(w.Short).String()
			alertLabels["long"] = model.Duration(w.Long).String()
			alertLabels["severity"] = string(w.Severity)

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
				For:         monitoringv1.Duration(model.Duration(w.For).String()),
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
			if m.Type == labels.MatchEqual && m.Name != labels.MetricName {
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
				Expr:   intstr.FromString(o.Burnrate(br)),
				Labels: ruleLabels,
			})
		}

		if o.Alerting.Disabled {
			return monitoringv1.RuleGroup{
				Name:     sloName,
				Interval: "30s", // TODO: Increase or decrease based on availability target
				Rules:    rules,
			}, nil
		}

		var alertMatchers []string
		for _, m := range matchers {
			if m.Name == labels.MetricName {
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

		for _, w := range ws {
			alertLabels := o.commonRuleLabels(sloName)
			alertAnnotations := o.commonRuleAnnotations()
			for _, m := range matchers {
				if m.Type == labels.MatchEqual && m.Name != labels.MetricName {
					if _, ok := groupingMap[m.Name]; !ok { // only add labels that aren't grouped by
						alertLabels[m.Name] = m.Value
					}
				}
			}
			alertLabels["short"] = model.Duration(w.Short).String()
			alertLabels["long"] = model.Duration(w.Long).String()
			alertLabels["severity"] = string(w.Severity)

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
				For:         monitoringv1.Duration(model.Duration(w.For).String()),
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
			if m.Type == labels.MatchEqual && m.Name != labels.MetricName {
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
				Expr:   intstr.FromString(o.Burnrate(br)),
				Labels: ruleLabels,
			})
		}

		if o.Alerting.Disabled {
			return monitoringv1.RuleGroup{
				Name:     sloName,
				Interval: "30s", // TODO: Increase or decrease based on availability target
				Rules:    rules,
			}, nil
		}

		var alertMatchers []string
		for _, m := range matchers {
			if m.Name == labels.MetricName {
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

		for _, w := range ws {
			alertLabels := o.commonRuleLabels(sloName)
			alertAnnotations := o.commonRuleAnnotations()
			for _, m := range matchers {
				if m.Type == labels.MatchEqual && m.Name != labels.MetricName {
					if _, ok := groupingMap[m.Name]; !ok { // only add labels that aren't grouped by
						alertLabels[m.Name] = m.Value
					}
				}
			}
			alertLabels["short"] = model.Duration(w.Short).String()
			alertLabels["long"] = model.Duration(w.Long).String()
			alertLabels["severity"] = string(w.Severity)

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
				For:         monitoringv1.Duration(model.Duration(w.For).String()),
				Labels:      alertLabels,
				Annotations: alertAnnotations,
			}
			rules = append(rules, r)
		}
	}

	// We only get here if alerting was not disabled
	return monitoringv1.RuleGroup{
		Name:     sloName,
		Interval: "30s", // TODO: Increase or decrease based on availability target
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

func (o Objective) Burnrate(timerange time.Duration) string {
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
		for _, m := range o.Indicator.Ratio.Total.LabelMatchers {
			if m.Type == labels.MatchRegexp || m.Type == labels.MatchNotRegexp {
				groupingMap[m.Name] = struct{}{}
			}
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
			matchers:      o.Indicator.Latency.Total.LabelMatchers,
			errorMetric:   o.Indicator.Latency.Success.Name,
			errorMatchers: o.Indicator.Latency.Success.LabelMatchers,
			grouping:      grouping,
			window:        timerange,
		}.replace(expr)

		return expr.String()
	case LatencyNative:
		expr, err := parser.ParseExpr(`1 - histogram_fraction(0,0.696969, rate(metric{matchers="total"}[1s]))`)
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

	for _, label := range o.Labels {
		if strings.HasPrefix(label.Name, PropagationLabelsPrefix) {
			ruleLabels[strings.TrimPrefix(label.Name, PropagationLabelsPrefix)] = label.Value
		}
	}

	return ruleLabels
}

func (o Objective) commonRuleAnnotations() map[string]string {
	var annotations map[string]string
	if len(o.Annotations) > 0 {
		annotations = make(map[string]string)
		for key, value := range o.Annotations {
			if strings.HasPrefix(key, PropagationLabelsPrefix) {
				annotations[strings.TrimPrefix(key, PropagationLabelsPrefix)] = value
			}
		}
	}

	return annotations
}

func (o Objective) IncreaseRules() (monitoringv1.RuleGroup, error) {
	sloName := o.Labels.Get(labels.MetricName)

	countExpr := func() (parser.Expr, error) { // Returns a new instance of Expr with this query each time called
		return parser.ParseExpr(`sum by (grouping) (count_over_time(metric{matchers="total"}[1s]))`)
	}

	sumExpr := func() (parser.Expr, error) { // Returns a new instance of Expr with this query each time called
		return parser.ParseExpr(`sum by (grouping) (sum_over_time(metric{matchers="total"}[1s]))`)
	}

	increaseExpr := func() (parser.Expr, error) { // Returns a new instance of Expr with this query each time called
		return parser.ParseExpr(`sum by (grouping) (increase(metric{matchers="total"}[1s]))`)
	}

	absentExpr := func() (parser.Expr, error) {
		return parser.ParseExpr(`absent(metric{matchers="total"}) == 1`)
	}

	var rules []monitoringv1.Rule

	switch o.IndicatorType() {
	case Ratio:
		ruleLabels := o.commonRuleLabels(sloName)
		for _, m := range o.Indicator.Ratio.Total.LabelMatchers {
			if m.Type == labels.MatchEqual && m.Name != labels.MetricName {
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
		// Delete labels that are grouped, as their value is part of the recording rule anyway
		for g := range groupingMap {
			delete(ruleLabels, g)
		}

		grouping := make([]string, 0, len(groupingMap))
		for s := range groupingMap {
			grouping = append(grouping, s)
		}
		sort.Strings(grouping)

		expr, err := increaseExpr()
		if err != nil {
			return monitoringv1.RuleGroup{}, err
		}

		objectiveReplacer{
			metric:   o.Indicator.Ratio.Total.Name,
			matchers: o.Indicator.Ratio.Total.LabelMatchers,
			grouping: grouping,
			window:   time.Duration(o.Window),
		}.replace(expr)

		rules = append(rules, monitoringv1.Rule{
			Record: increaseName(o.Indicator.Ratio.Total.Name, o.Window),
			Expr:   intstr.FromString(expr.String()),
			Labels: ruleLabels,
		})

		expr, err = absentExpr()
		if err != nil {
			return monitoringv1.RuleGroup{}, err
		}

		objectiveReplacer{
			metric:   o.Indicator.Ratio.Total.Name,
			matchers: o.Indicator.Ratio.Total.LabelMatchers,
		}.replace(expr)

		alertLabels := make(map[string]string, len(ruleLabels)+1)
		for k, v := range ruleLabels {
			alertLabels[k] = v
		}
		// Add severity label for alerts
		alertLabels["severity"] = string(critical)

		rules = append(rules, monitoringv1.Rule{
			Alert: "SLOMetricAbsent",
			Expr:  intstr.FromString(expr.String()),
			For: monitoringv1.Duration(model.Duration(
				(time.Duration(o.Window) / (28 * 24 * (60 / 2))).Round(time.Minute),
			).String()),
			Labels:      alertLabels,
			Annotations: o.commonRuleAnnotations(),
		})

		if o.Indicator.Ratio.Total.Name != o.Indicator.Ratio.Errors.Name {
			expr, err := increaseExpr()
			if err != nil {
				return monitoringv1.RuleGroup{}, err
			}

			objectiveReplacer{
				metric:   o.Indicator.Ratio.Errors.Name,
				matchers: o.Indicator.Ratio.Errors.LabelMatchers,
				grouping: grouping,
				window:   time.Duration(o.Window),
			}.replace(expr)

			rules = append(rules, monitoringv1.Rule{
				Record: increaseName(o.Indicator.Ratio.Errors.Name, o.Window),
				Expr:   intstr.FromString(expr.String()),
				Labels: ruleLabels,
			})

			expr, err = absentExpr()
			if err != nil {
				return monitoringv1.RuleGroup{}, err
			}

			objectiveReplacer{
				metric:   o.Indicator.Ratio.Errors.Name,
				matchers: o.Indicator.Ratio.Errors.LabelMatchers,
			}.replace(expr)

			rules = append(rules, monitoringv1.Rule{
				Alert: "SLOMetricAbsent",
				Expr:  intstr.FromString(expr.String()),
				For: monitoringv1.Duration(model.Duration(
					(time.Duration(o.Window) / (28 * 24 * (60 / 2))).Round(time.Minute),
				).String()),
				Labels:      alertLabels,
				Annotations: o.commonRuleAnnotations(),
			})
		}
	case Latency:
		ruleLabels := o.commonRuleLabels(sloName)
		for _, m := range o.Indicator.Latency.Total.LabelMatchers {
			if m.Type == labels.MatchEqual && m.Name != labels.MetricName {
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
		// Delete labels that are grouped, as their value is part of the recording rule anyway
		for g := range groupingMap {
			delete(ruleLabels, g)
		}

		grouping := make([]string, 0, len(groupingMap))
		for s := range groupingMap {
			grouping = append(grouping, s)
		}
		sort.Strings(grouping)

		expr, err := increaseExpr()
		if err != nil {
			return monitoringv1.RuleGroup{}, err
		}

		objectiveReplacer{
			metric:   o.Indicator.Latency.Total.Name,
			matchers: o.Indicator.Latency.Total.LabelMatchers,
			grouping: grouping,
			window:   time.Duration(o.Window),
		}.replace(expr)

		rules = append(rules, monitoringv1.Rule{
			Record: increaseName(o.Indicator.Latency.Total.Name, o.Window),
			Expr:   intstr.FromString(expr.String()),
			Labels: ruleLabels,
		})

		expr, err = increaseExpr()
		if err != nil {
			return monitoringv1.RuleGroup{}, err
		}

		objectiveReplacer{
			metric:   o.Indicator.Latency.Success.Name,
			matchers: o.Indicator.Latency.Success.LabelMatchers,
			grouping: grouping,
			window:   time.Duration(o.Window),
		}.replace(expr)

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

		rules = append(rules, monitoringv1.Rule{
			Record: increaseName(o.Indicator.Latency.Success.Name, o.Window),
			Expr:   intstr.FromString(expr.String()),
			Labels: ruleLabelsLe,
		})

		expr, err = absentExpr()
		if err != nil {
			return monitoringv1.RuleGroup{}, err
		}

		objectiveReplacer{
			metric:   o.Indicator.Latency.Total.Name,
			matchers: o.Indicator.Latency.Total.LabelMatchers,
		}.replace(expr)

		alertLabels := make(map[string]string, len(ruleLabels)+1)
		for k, v := range ruleLabels {
			alertLabels[k] = v
		}
		// Add severity label for alerts
		alertLabels["severity"] = string(critical)

		rules = append(rules, monitoringv1.Rule{
			Alert: "SLOMetricAbsent",
			Expr:  intstr.FromString(expr.String()),
			For: monitoringv1.Duration(model.Duration(
				(time.Duration(o.Window) / (28 * 24 * (60 / 2))).Round(time.Minute),
			).String()),
			Labels:      alertLabels,
			Annotations: o.commonRuleAnnotations(),
		})

		expr, err = absentExpr()
		if err != nil {
			return monitoringv1.RuleGroup{}, err
		}

		objectiveReplacer{
			metric:   o.Indicator.Latency.Success.Name,
			matchers: o.Indicator.Latency.Success.LabelMatchers,
		}.replace(expr)

		alertLabelsLe := make(map[string]string, len(ruleLabelsLe)+1)
		for k, v := range ruleLabelsLe {
			alertLabelsLe[k] = v
		}
		// Add severity label for alerts
		alertLabelsLe["severity"] = string(critical)

		rules = append(rules, monitoringv1.Rule{
			Alert: "SLOMetricAbsent",
			Expr:  intstr.FromString(expr.String()),
			For: monitoringv1.Duration(model.Duration(
				(time.Duration(o.Window) / (28 * 24 * (60 / 2))).Round(time.Minute),
			).String()),
			Labels:      alertLabelsLe,
			Annotations: o.commonRuleAnnotations(),
		})
	case LatencyNative:
		ruleLabels := o.commonRuleLabels(sloName)
		for _, m := range o.Indicator.LatencyNative.Total.LabelMatchers {
			if m.Type == labels.MatchEqual && m.Name != labels.MetricName {
				ruleLabels[m.Name] = m.Value
			}
		}

		expr, err := parser.ParseExpr(`histogram_count(increase(metric{matchers="total"}[1s]))`)
		if err != nil {
			return monitoringv1.RuleGroup{}, err
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

		expr, err = parser.ParseExpr(`histogram_fraction(0, 0.696969, increase(metric{matchers="total"}[1s])) * histogram_count(increase(metric{matchers="total"}[1s]))`)
		if err != nil {
			return monitoringv1.RuleGroup{}, err
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
	case BoolGauge:
		ruleLabels := o.commonRuleLabels(sloName)
		for _, m := range o.Indicator.BoolGauge.LabelMatchers {
			if m.Type == labels.MatchEqual && m.Name != labels.MetricName {
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

		expr, err := countExpr()
		if err != nil {
			return monitoringv1.RuleGroup{}, err
		}

		sum, err := sumExpr()
		if err != nil {
			return monitoringv1.RuleGroup{}, err
		}

		objectiveReplacer{
			metric:   o.Indicator.BoolGauge.Name,
			matchers: o.Indicator.BoolGauge.LabelMatchers,
			grouping: grouping,
			window:   time.Duration(o.Window),
		}.replace(expr)

		objectiveReplacer{
			metric:   o.Indicator.BoolGauge.Name,
			matchers: o.Indicator.BoolGauge.LabelMatchers,
			grouping: grouping,
			window:   time.Duration(o.Window),
		}.replace(sum)

		rules = append(rules, monitoringv1.Rule{
			Record: countName(o.Indicator.BoolGauge.Name, o.Window),
			Expr:   intstr.FromString(expr.String()),
			Labels: ruleLabels,
		})

		rules = append(rules, monitoringv1.Rule{
			Record: sumName(o.Indicator.BoolGauge.Name, o.Window),
			Expr:   intstr.FromString(sum.String()),
			Labels: ruleLabels,
		})

		expr, err = countExpr()
		if err != nil {
			return monitoringv1.RuleGroup{}, err
		}

		objectiveReplacer{
			metric:   o.Indicator.BoolGauge.Name,
			matchers: o.Indicator.BoolGauge.LabelMatchers,
			grouping: grouping,
			window:   time.Duration(o.Window),
		}.replace(expr)

		expr, err = absentExpr()
		if err != nil {
			return monitoringv1.RuleGroup{}, err
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
		alertLabels["severity"] = string(critical)

		rules = append(rules, monitoringv1.Rule{
			Alert: "SLOMetricAbsent",
			Expr:  intstr.FromString(expr.String()),
			For: monitoringv1.Duration(model.Duration(
				(time.Duration(o.Window) / (28 * 24 * (60 / 2))).Round(time.Minute),
			).String()),
			Labels:      alertLabels,
			Annotations: o.commonRuleAnnotations(),
		})
	}

	day := 24 * time.Hour

	var interval model.Duration
	window := time.Duration(o.Window)

	// TODO: Make this a function with an equation
	if window < 7*day {
		interval = model.Duration(30 * time.Second)
	} else if window < 14*day {
		interval = model.Duration(60 * time.Second)
	} else if window < 21*day {
		interval = model.Duration(90 * time.Second)
	} else if window < 28*day {
		interval = model.Duration(120 * time.Second)
	} else if window < 35*day {
		interval = model.Duration(150 * time.Second)
	} else if window < 42*day {
		interval = model.Duration(180 * time.Second)
	} else if window < 49*day {
		interval = model.Duration(210 * time.Second)
	} else { // 8w
		interval = model.Duration(240 * time.Second)
	}

	return monitoringv1.RuleGroup{
		Name:     sloName + "-increase",
		Interval: monitoringv1.Duration(interval.String()),
		Rules:    rules,
	}, nil
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

func (o Objective) GenericRules() (monitoringv1.RuleGroup, error) {
	sloName := o.Labels.Get(labels.MetricName)
	var rules []monitoringv1.Rule

	switch o.IndicatorType() {
	case Ratio:
		if len(o.Indicator.Ratio.Grouping) > 0 {
			return monitoringv1.RuleGroup{}, ErrGroupingUnsupported
		}

		ruleLabels := o.commonRuleLabels(sloName)

		rules = append(rules, monitoringv1.Rule{
			Record: "pyrra_objective",
			Expr:   intstr.FromString(strconv.FormatFloat(o.Target, 'f', -1, 64)),
			Labels: ruleLabels,
		})
		rules = append(rules, monitoringv1.Rule{
			Record: "pyrra_window",
			Expr:   intstr.FromInt(int(time.Duration(o.Window).Seconds())),
			Labels: ruleLabels,
		})

		availability, err := parser.ParseExpr(`1 - sum(errorMetric{matchers="errors"} or vector(0)) / sum(metric{matchers="total"})`)
		if err != nil {
			return monitoringv1.RuleGroup{}, err
		}

		totalIncreaseName := increaseName(o.Indicator.Ratio.Total.Name, o.Window)

		// Copy the list of matchers to modify them
		totalMatchers := make([]*labels.Matcher, 0, len(o.Indicator.Ratio.Total.LabelMatchers))
		for _, m := range o.Indicator.Ratio.Total.LabelMatchers {
			value := m.Value
			if m.Name == labels.MetricName {
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
			if m.Name == labels.MetricName {
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

		rate, err := parser.ParseExpr(`sum(metric{matchers="total"})`)
		if err != nil {
			return monitoringv1.RuleGroup{}, err
		}

		objectiveReplacer{
			metric:   o.Indicator.Ratio.Total.Name,
			matchers: o.Indicator.Ratio.Total.LabelMatchers,
		}.replace(rate)

		rules = append(rules, monitoringv1.Rule{
			Record: "pyrra_requests_total",
			Expr:   intstr.FromString(rate.String()),
			Labels: ruleLabels,
		})

		errorsExpr := func() (parser.Expr, error) { // Returns a new instance of Expr with this query each time called
			return parser.ParseExpr(`sum(metric{matchers="total"} or vector(0))`)
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
			Record: "pyrra_errors_total",
			Expr:   intstr.FromString(errorsParsedExpr.String()),
			Labels: ruleLabels,
		})
	case Latency:
		if len(o.Indicator.Latency.Grouping) > 0 {
			return monitoringv1.RuleGroup{}, ErrGroupingUnsupported
		}

		ruleLabels := o.commonRuleLabels(sloName)

		rules = append(rules, monitoringv1.Rule{
			Record: "pyrra_objective",
			Expr:   intstr.FromString(strconv.FormatFloat(o.Target, 'f', -1, 64)),
			Labels: ruleLabels,
		})
		rules = append(rules, monitoringv1.Rule{
			Record: "pyrra_window",
			Expr:   intstr.FromInt(int(time.Duration(o.Window).Seconds())),
			Labels: ruleLabels,
		})
		// availability
		{
			expr, err := parser.ParseExpr(`sum(errorMetric{matchers="errors"} or vector(0)) / sum(metric{matchers="total"})`)
			if err != nil {
				return monitoringv1.RuleGroup{}, err
			}

			metric := increaseName(o.Indicator.Latency.Total.Name, o.Window)
			matchers := o.Indicator.Latency.Total.LabelMatchers
			for _, m := range matchers {
				if m.Name == labels.MetricName {
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

			errorMetric := increaseName(o.Indicator.Latency.Success.Name, o.Window)
			errorMatchers := o.Indicator.Latency.Success.LabelMatchers
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
			rate, err := parser.ParseExpr(`sum(metric{matchers="total"})`)
			if err != nil {
				return monitoringv1.RuleGroup{}, err
			}

			metric := o.Indicator.Latency.Total.Name
			matchers := o.Indicator.Latency.Total.LabelMatchers
			for _, m := range matchers {
				if m.Name == labels.MetricName {
					m.Value = metric
					break
				}
			}
			objectiveReplacer{
				metric:   metric,
				matchers: matchers,
			}.replace(rate)

			rules = append(rules, monitoringv1.Rule{
				Record: "pyrra_requests_total",
				Expr:   intstr.FromString(rate.String()),
				Labels: ruleLabels,
			})
		}
		// errors
		{
			errorsExpr, err := parser.ParseExpr(`sum(metric{matchers="total"}) - sum(errorMetric{matchers="errors"})`)
			if err != nil {
				return monitoringv1.RuleGroup{}, err
			}

			metric := o.Indicator.Latency.Total.Name
			matchers := o.Indicator.Latency.Total.LabelMatchers
			for _, m := range matchers {
				if m.Name == labels.MetricName {
					m.Value = metric
					break
				}
			}

			errorMetric := o.Indicator.Latency.Success.Name
			errorMatchers := o.Indicator.Latency.Success.LabelMatchers
			for _, m := range errorMatchers {
				if m.Name == labels.MetricName {
					m.Value = errorMetric
					break
				}
			}

			objectiveReplacer{
				metric:        metric,
				matchers:      matchers,
				errorMetric:   errorMetric,
				errorMatchers: errorMatchers,
			}.replace(errorsExpr)

			rules = append(rules, monitoringv1.Rule{
				Record: "pyrra_errors_total",
				Expr:   intstr.FromString(errorsExpr.String()),
				Labels: ruleLabels,
			})
		}
	}

	return monitoringv1.RuleGroup{
		Name:     sloName + "-generic",
		Interval: "30s",
		Rules:    rules,
	}, nil
}
