package slo

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
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
	ws := windows(time.Duration(o.Window))

	mbras := make([]MultiBurnRateAlert, len(ws))
	for i, w := range ws {
		mbras[i] = MultiBurnRateAlert{
			Severity:   string(w.Severity),
			Short:      w.Short,
			Long:       w.Long,
			For:        w.For,
			Factor:     w.Factor,
			QueryShort: o.Burnrate(w.Short),
			QueryLong:  o.Burnrate(w.Long),
		}
	}

	return mbras, nil
}

func (o Objective) Burnrates() (monitoringv1.RuleGroup, error) {
	ws := windows(time.Duration(o.Window))
	burnrates := burnratesFromWindows(ws)
	rules := make([]monitoringv1.Rule, 0, len(burnrates))

	if o.Indicator.Ratio != nil && o.Indicator.Ratio.Total.Name != "" {
		metric := o.Indicator.Ratio.Total.Name
		matcher := o.Indicator.Ratio.Total.LabelMatchers

		var matcherLabels = map[string]string{}
		for _, m := range matcher {
			if m.Type == labels.MatchEqual || m.Type == labels.MatchRegexp {
				if m.Name != "__name__" {
					matcherLabels[m.Name] = m.Value
				}
			}
		}
		matcherLabels["slo"] = o.Labels.Get(labels.MetricName)

		matcherLabelsString := func(ml map[string]string) string {
			var s []string
			for n, v := range ml {
				s = append(s, fmt.Sprintf(`%s="%s"`, n, v))
			}
			sort.Slice(s, func(i, j int) bool {
				return s[i] < s[j]
			})
			return strings.Join(s, ",")
		}(matcherLabels)

		for _, br := range burnrates {
			rules = append(rules, monitoringv1.Rule{
				Record: burnrateName(metric, br),
				Expr:   intstr.FromString(o.Burnrate(br)),
				Labels: matcherLabels,
			})
		}

		for _, w := range ws {
			alertLabels := map[string]string{}
			for n, v := range matcherLabels {
				alertLabels[n] = v
			}
			alertLabels["short"] = model.Duration(w.Short).String()
			alertLabels["long"] = model.Duration(w.Long).String()

			r := monitoringv1.Rule{
				Alert: "ErrorBudgetBurn",
				// TODO: Use expr replacer
				Expr: intstr.FromString(fmt.Sprintf("%s{%s} > (%.f * (1-%s)) and %s{%s} > (%.f * (1-%s))",
					burnrateName(metric, w.Short),
					matcherLabelsString,
					w.Factor,
					strconv.FormatFloat(o.Target, 'f', -1, 64),
					burnrateName(metric, w.Long),
					matcherLabelsString,
					w.Factor,
					strconv.FormatFloat(o.Target, 'f', -1, 64),
				)),
				For: model.Duration(w.For).String(),
				Annotations: map[string]string{
					"severity": string(w.Severity),
				},
				Labels: alertLabels,
			}
			rules = append(rules, r)
		}
	}
	if o.Indicator.Latency != nil && o.Indicator.Latency.Total.Name != "" {
		var matcherLabels = map[string]string{}
		for _, m := range o.Indicator.Latency.Total.LabelMatchers {
			if m.Type == labels.MatchEqual || m.Type == labels.MatchRegexp {
				if m.Name != "__name__" {
					matcherLabels[m.Name] = m.Value
				}
			}
		}
		matcherLabels["slo"] = o.Labels.Get(labels.MetricName)

		matcherLabelsString := func(ml map[string]string) string {
			var s []string
			for n, v := range ml {
				s = append(s, fmt.Sprintf(`%s="%s"`, n, v))
			}
			sort.Slice(s, func(i, j int) bool {
				return s[i] < s[j]
			})
			return strings.Join(s, ",")
		}(matcherLabels)

		for _, br := range burnrates {
			rules = append(rules, monitoringv1.Rule{
				Record: burnrateName(o.Indicator.Latency.Total.Name, br),
				Expr:   intstr.FromString(o.Burnrate(br)),
				Labels: matcherLabels,
			})
		}

		for _, w := range ws {
			alertLabels := map[string]string{}
			for n, v := range matcherLabels {
				alertLabels[n] = v
			}
			alertLabels["short"] = model.Duration(w.Short).String()
			alertLabels["long"] = model.Duration(w.Long).String()

			r := monitoringv1.Rule{
				Alert: "ErrorBudgetBurn",
				// TODO: Use expr replacer
				Expr: intstr.FromString(fmt.Sprintf("%s{%s} > (%.f * (1-%s)) and %s{%s} > (%.f * (1-%s))",
					burnrateName(o.Indicator.Latency.Total.Name, w.Short),
					matcherLabelsString,
					w.Factor,
					strconv.FormatFloat(o.Target, 'f', -1, 64),
					burnrateName(o.Indicator.Latency.Total.Name, w.Long),
					matcherLabelsString,
					w.Factor,
					strconv.FormatFloat(o.Target, 'f', -1, 64),
				)),
				For: model.Duration(w.For).String(),
				Annotations: map[string]string{
					"severity": string(w.Severity),
				},
				Labels: alertLabels,
			}
			rules = append(rules, r)
		}
	}

	return monitoringv1.RuleGroup{
		Name:     o.Labels.Get(labels.MetricName),
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

		objectiveReplacer{
			metric:        o.Indicator.Ratio.Total.Name,
			matchers:      o.Indicator.Ratio.Total.LabelMatchers,
			errorMetric:   o.Indicator.Ratio.Errors.Name,
			errorMatchers: o.Indicator.Ratio.Errors.LabelMatchers,
			grouping:      o.Indicator.Ratio.Grouping,
			window:        timerange,
		}.replace(expr)

		return expr.String()
	}
	if o.Indicator.Latency != nil && o.Indicator.Latency.Total.Name != "" {
		expr, err := parser.ParseExpr(`sum(rate(metric{matchers="total"}[1s])) -  sum(rate(errorMetric{matchers="errors"}[1s]))`)
		if err != nil {
			return err.Error()
		}

		objectiveReplacer{
			metric:        o.Indicator.Latency.Total.Name,
			matchers:      o.Indicator.Latency.Total.LabelMatchers,
			errorMetric:   o.Indicator.Latency.Success.Name,
			errorMatchers: o.Indicator.Latency.Success.LabelMatchers,
			window:        timerange,
		}.replace(expr)

		return expr.String()
	}
	return ""
}

type severity string

const critical severity = "critical"
const warning severity = "warning"

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
