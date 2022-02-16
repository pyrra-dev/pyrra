package slo

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
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
	sloName := o.Labels.Get(labels.MetricName)
	ws := windows(time.Duration(o.Window))

	var (
		metric         string
		matchersString string
	)
	if o.Indicator.Ratio != nil && o.Indicator.Ratio.Total.Name != "" {
		metric = o.Indicator.Ratio.Total.Name

		// TODO: Make this a shared function with below and Burnrates func
		var alertMatchers []string
		for _, m := range o.Indicator.Ratio.Total.LabelMatchers {
			if m.Name == labels.MetricName {
				continue
			}
			alertMatchers = append(alertMatchers, m.String())
		}
		alertMatchers = append(alertMatchers, fmt.Sprintf(`slo="%s"`, sloName))
		sort.Strings(alertMatchers)
		matchersString = strings.Join(alertMatchers, ",")
	}
	if o.Indicator.Latency != nil && o.Indicator.Latency.Total.Name != "" {
		metric = o.Indicator.Latency.Total.Name

		// TODO: Make this a shared function with below and Burnrates func
		var alertMatchers []string
		for _, m := range o.Indicator.Latency.Total.LabelMatchers {
			if m.Name == labels.MetricName {
				continue
			}
			alertMatchers = append(alertMatchers, m.String())
		}
		alertMatchers = append(alertMatchers, fmt.Sprintf(`slo="%s"`, sloName))
		sort.Strings(alertMatchers)
		matchersString = strings.Join(alertMatchers, ",")
	}

	mbras := make([]MultiBurnRateAlert, len(ws))
	for i, w := range ws {
		queryShort := fmt.Sprintf("%s{%s}", burnrateName(metric, w.Short), matchersString)
		queryLong := fmt.Sprintf("%s{%s}", burnrateName(metric, w.Long), matchersString)
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

// TODO: Rename Burnrates to RecordingRules

func (o Objective) Burnrates() (monitoringv1.RuleGroup, error) {
	sloName := o.Labels.Get(labels.MetricName)

	ws := windows(time.Duration(o.Window))
	burnrates := burnratesFromWindows(ws)
	rules := make([]monitoringv1.Rule, 0, len(burnrates))

	if o.Indicator.Ratio != nil && o.Indicator.Ratio.Total.Name != "" {
		metric := o.Indicator.Ratio.Total.Name
		matchers := o.Indicator.Ratio.Total.LabelMatchers

		groupingMap := map[string]struct{}{}
		for _, g := range o.Indicator.Ratio.Grouping {
			groupingMap[g] = struct{}{}
		}

		ruleLabels := map[string]string{}
		ruleLabels["slo"] = sloName
		for _, m := range matchers {
			if m.Type == labels.MatchEqual && m.Name != labels.MetricName {
				ruleLabels[m.Name] = m.Value
			}
		}
		// Delete labels that are grouped as their value is part of the labels anyway
		for g := range groupingMap {
			delete(ruleLabels, g)
		}

		increaseRules, err := o.IncreaseRules()
		if err != nil {
			return monitoringv1.RuleGroup{}, err
		}
		for _, r := range increaseRules {
			for k, v := range ruleLabels {
				r.Labels[k] = v
			}
			rules = append(rules, r)
		}

		for _, br := range burnrates {
			rules = append(rules, monitoringv1.Rule{
				Record: burnrateName(metric, br),
				Expr:   intstr.FromString(o.Burnrate(br)),
				Labels: ruleLabels,
			})
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
			alertLabels := map[string]string{}
			for _, m := range matchers {
				if m.Type == labels.MatchEqual && m.Name != labels.MetricName {
					if _, ok := groupingMap[m.Name]; !ok { // only add labels that aren't grouped by
						alertLabels[m.Name] = m.Value
					}
				}
			}
			alertLabels["slo"] = sloName
			alertLabels["short"] = model.Duration(w.Short).String()
			alertLabels["long"] = model.Duration(w.Long).String()
			alertLabels["severity"] = string(w.Severity)

			r := monitoringv1.Rule{
				Alert: "ErrorBudgetBurn",
				// TODO: Use expr replacer
				Expr: intstr.FromString(fmt.Sprintf("%s{%s} > (%.f * (1-%s)) and %s{%s} > (%.f * (1-%s))",
					burnrateName(metric, w.Short),
					alertMatchersString,
					w.Factor,
					strconv.FormatFloat(o.Target, 'f', -1, 64),
					burnrateName(metric, w.Long),
					alertMatchersString,
					w.Factor,
					strconv.FormatFloat(o.Target, 'f', -1, 64),
				)),
				For:    model.Duration(w.For).String(),
				Labels: alertLabels,
			}
			rules = append(rules, r)
		}
	}

	if o.Indicator.Latency != nil && o.Indicator.Latency.Total.Name != "" {
		metric := o.Indicator.Latency.Total.Name
		matchers := o.Indicator.Latency.Total.LabelMatchers

		groupingMap := map[string]struct{}{}
		for _, g := range o.Indicator.Latency.Grouping {
			groupingMap[g] = struct{}{}
		}

		ruleLabels := map[string]string{}
		ruleLabels["slo"] = sloName
		for _, m := range matchers {
			if m.Type == labels.MatchEqual && m.Name != labels.MetricName {
				ruleLabels[m.Name] = m.Value
			}
		}
		// Delete labels that are grouped as their value is part of the labels anyway
		for g := range groupingMap {
			delete(ruleLabels, g)
		}

		increaseRules, err := o.IncreaseRules()
		if err != nil {
			return monitoringv1.RuleGroup{}, err
		}

		for _, r := range increaseRules {
			for k, v := range ruleLabels {
				r.Labels[k] = v
			}
			rules = append(rules, r)
		}

		for _, br := range burnrates {
			rules = append(rules, monitoringv1.Rule{
				Record: burnrateName(metric, br),
				Expr:   intstr.FromString(o.Burnrate(br)),
				Labels: ruleLabels,
			})
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
			alertLabels := map[string]string{}
			for _, m := range matchers {
				if m.Type == labels.MatchEqual && m.Name != labels.MetricName {
					if _, ok := groupingMap[m.Name]; !ok { // only add labels that aren't grouped by
						alertLabels[m.Name] = m.Value
					}
				}
			}
			alertLabels["slo"] = sloName
			alertLabels["short"] = model.Duration(w.Short).String()
			alertLabels["long"] = model.Duration(w.Long).String()
			alertLabels["severity"] = string(w.Severity)

			r := monitoringv1.Rule{
				Alert: "ErrorBudgetBurn",
				// TODO: Use expr replacer
				Expr: intstr.FromString(fmt.Sprintf("%s{%s} > (%.f * (1-%s)) and %s{%s} > (%.f * (1-%s))",
					burnrateName(metric, w.Short),
					alertMatchersString,
					w.Factor,
					strconv.FormatFloat(o.Target, 'f', -1, 64),
					burnrateName(metric, w.Long),
					alertMatchersString,
					w.Factor,
					strconv.FormatFloat(o.Target, 'f', -1, 64),
				)),
				For:    model.Duration(w.For).String(),
				Labels: alertLabels,
			}
			rules = append(rules, r)
		}
	}

	return monitoringv1.RuleGroup{
		Name:     sloName,
		Interval: "30s", // TODO: Increase or decrease based on availability target
		Rules:    rules,
	}, nil
}

func burnrateName(metric string, rate time.Duration) string {
	metric = strings.TrimSuffix(metric, "_total")
	metric = strings.TrimSuffix(metric, "_count")
	return fmt.Sprintf("%s:burnrate%s", metric, model.Duration(rate))
}

func (o Objective) Burnrate(timerange time.Duration) string {
	if o.Indicator.Ratio != nil && o.Indicator.Ratio.Total.Name != "" {
		expr, err := parser.ParseExpr(`sum by(grouping) (rate(errorMetric{matchers="errors"}[1s])) / sum by(grouping) (rate(metric{matchers="total"}[1s]))`)
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
	}
	if o.Indicator.Latency != nil && o.Indicator.Latency.Total.Name != "" {
		query := `
			(
				sum by(grouping) (rate(metric{matchers="total"}[1s]))
				-
				sum by(grouping) (rate(errorMetric{matchers="errors"}[1s]))
			)
			/
			sum by(grouping) (rate(metric{matchers="total"}[1s]))
`
		expr, err := parser.ParseExpr(query)
		if err != nil {
			return err.Error()
		}

		groupingMap := map[string]struct{}{}
		for _, s := range o.Indicator.Latency.Grouping {
			groupingMap[s] = struct{}{}
		}
		for _, m := range o.Indicator.Latency.Total.LabelMatchers {
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
			metric:        o.Indicator.Latency.Total.Name,
			matchers:      o.Indicator.Latency.Total.LabelMatchers,
			errorMetric:   o.Indicator.Latency.Success.Name,
			errorMatchers: o.Indicator.Latency.Success.LabelMatchers,
			grouping:      grouping,
			window:        timerange,
		}.replace(expr)

		return expr.String()
	}
	return ""
}

func increaseName(metric string, window model.Duration) string {
	metric = strings.TrimSuffix(metric, "_total")
	metric = strings.TrimSuffix(metric, "_count")
	metric = strings.TrimSuffix(metric, "_bucket")
	return fmt.Sprintf("%s:increase%s", metric, window)
}

func (o Objective) IncreaseRules() ([]monitoringv1.Rule, error) {
	var rules []monitoringv1.Rule

	increaseExpr := func() (parser.Expr, error) { // Returns a new instance of Expr with this query each time called
		return parser.ParseExpr(`sum by (grouping) (increase(metric{matchers="total"}[1s]))`)
	}

	if o.Indicator.Ratio != nil && o.Indicator.Ratio.Total.Name != "" {
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

		grouping := make([]string, 0, len(groupingMap))
		for s := range groupingMap {
			grouping = append(grouping, s)
		}
		sort.Strings(grouping)

		expr, err := increaseExpr()
		if err != nil {
			return nil, err
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
			Labels: map[string]string{},
		})

		if o.Indicator.Ratio.Total.Name != o.Indicator.Ratio.Errors.Name {
			expr, err := increaseExpr()
			if err != nil {
				return nil, err
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
				Labels: map[string]string{},
			})
		}
	}

	if o.Indicator.Latency != nil && o.Indicator.Latency.Total.Name != "" {
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

		grouping := make([]string, 0, len(groupingMap))
		for s := range groupingMap {
			grouping = append(grouping, s)
		}
		sort.Strings(grouping)

		expr, err := increaseExpr()
		if err != nil {
			return nil, err
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
			Labels: map[string]string{},
		})

		expr, err = increaseExpr()
		if err != nil {
			return nil, err
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

		rules = append(rules, monitoringv1.Rule{
			Record: increaseName(o.Indicator.Latency.Success.Name, o.Window),
			Expr:   intstr.FromString(expr.String()),
			Labels: map[string]string{"le": le},
		})
	}

	return rules, nil
}

type severity string

const (
	critical severity = "critical"
	warning  severity = "warning"
)

type window struct {
	Severity severity
	For      time.Duration
	Long     time.Duration
	Short    time.Duration
	Factor   float64
}

func windows(sloWindow time.Duration) []window {
	// TODO: I'm still not sure if For, Long, Short should really be based on the 28 days ratio...

	round := time.Minute // TODO: Change based on sloWindow

	// long and short rates are calculated based on the ratio for 28 days.
	return []window{{
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

func burnratesFromWindows(ws []window) []time.Duration {
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
