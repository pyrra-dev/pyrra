package slo

import (
	"fmt"
	"sort"
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

	var metric string
	var matcher []*labels.Matcher
	var errorMatcher []*labels.Matcher

	if o.Indicator.HTTP != nil {
		if o.Indicator.HTTP.Metric == "" {
			o.Indicator.HTTP.Metric = HTTPDefaultMetric
		}
		if len(o.Indicator.HTTP.ErrorMatchers) == 0 {
			o.Indicator.HTTP.ErrorMatchers = []*labels.Matcher{HTTPDefaultErrorSelector}
		}

		metric = o.Indicator.HTTP.Metric
		matcher = o.Indicator.HTTP.Matchers
		errorMatcher = o.Indicator.HTTP.AllSelectors()
	}
	if o.Indicator.GRPC != nil {
		if o.Indicator.GRPC.Metric == "" {
			o.Indicator.GRPC.Metric = GRPCDefaultMetric
		}
		if len(o.Indicator.GRPC.ErrorMatchers) == 0 {
			o.Indicator.GRPC.ErrorMatchers = []*labels.Matcher{GRPCDefaultErrorSelector}
		}

		metric = o.Indicator.GRPC.Metric
		matcher = o.Indicator.GRPC.GRPCSelectors()
		errorMatcher = o.Indicator.GRPC.AllSelectors()
	}

	mbras := make([]MultiBurnRateAlert, len(ws))
	for i, w := range ws {
		queryShort, err := Burnrate(metric, w.Short, matcher, errorMatcher)
		if err != nil {
			return nil, err
		}
		queryLong, err := Burnrate(metric, w.Long, matcher, errorMatcher)
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
	ws := windows(time.Duration(o.Window))
	burnrates := burnratesFromWindows(ws)
	rules := make([]monitoringv1.Rule, 0, len(burnrates))

	var metric string
	var matcher []*labels.Matcher
	var errorMatcher []*labels.Matcher

	if o.Indicator.HTTP != nil {
		if o.Indicator.HTTP.Metric == "" {
			o.Indicator.HTTP.Metric = HTTPDefaultMetric
		}
		if len(o.Indicator.HTTP.ErrorMatchers) == 0 {
			o.Indicator.HTTP.ErrorMatchers = []*labels.Matcher{HTTPDefaultErrorSelector}
		}

		metric = o.Indicator.HTTP.Metric
		matcher = o.Indicator.HTTP.Matchers
		errorMatcher = o.Indicator.HTTP.AllSelectors()
	}
	if o.Indicator.GRPC != nil {
		if o.Indicator.GRPC.Metric == "" {
			o.Indicator.GRPC.Metric = GRPCDefaultMetric
		}
		if len(o.Indicator.GRPC.ErrorMatchers) == 0 {
			o.Indicator.GRPC.ErrorMatchers = []*labels.Matcher{GRPCDefaultErrorSelector}
		}

		metric = o.Indicator.GRPC.Metric
		matcher = o.Indicator.GRPC.GRPCSelectors()
		errorMatcher = o.Indicator.GRPC.AllSelectors()
	}

	var matcherLabels = map[string]string{}
	for _, m := range matcher {
		if m.Type == labels.MatchEqual || m.Type == labels.MatchRegexp {
			matcherLabels[m.Name] = m.Value
		}
	}
	matcherLabels["slo"] = o.Name

	// Create the label matcher string without the slo label
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
		expr, err := Burnrate(metric, br, matcher, errorMatcher)
		if err != nil {
			return monitoringv1.RuleGroup{}, err
		}

		rules = append(rules, monitoringv1.Rule{
			Record: burnrateName(metric, br),
			Expr:   intstr.FromString(expr),
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
			Expr: intstr.FromString(fmt.Sprintf("%s{%s} > (%.f * (1-%.5f)) and %s{%s} > (%.f * (1-%.5f))",
				burnrateName(metric, w.Short),
				matcherLabelsString,
				w.Factor,
				o.Target,
				burnrateName(metric, w.Long),
				matcherLabelsString,
				w.Factor,
				o.Target,
			)),
			For: model.Duration(w.For).String(),
			Annotations: map[string]string{
				"severity": string(w.Severity),
			},
			Labels: alertLabels,
		}
		rules = append(rules, r)
	}

	return monitoringv1.RuleGroup{
		Name:     o.Name,
		Interval: "30s", // TODO: Increase or decrease based on availability target
		Rules:    rules,
	}, nil
}

func burnrateName(metric string, rate time.Duration) string {
	metric = strings.TrimSuffix(metric, "_total")
	return fmt.Sprintf("%s:burnrate%s", metric, model.Duration(rate))
}

func Burnrate(metric string, rate time.Duration, matchers []*labels.Matcher, errorMatchers []*labels.Matcher) (string, error) {
	expr, err := parser.ParseExpr(`sum(rate(metric{matchers="errors"}[1s])) / sum(rate(metric{matchers="total"}[1s]))`)
	if err != nil {
		return "", err
	}

	metricMatcher := &labels.Matcher{Type: labels.MatchEqual, Name: "__name__", Value: metric}
	matchers = append([]*labels.Matcher{metricMatcher}, matchers...)
	errorMatchers = append([]*labels.Matcher{metricMatcher}, errorMatchers...)

	objectiveReplacer{
		metric:        metric,
		matchers:      matchers,
		errorMatchers: errorMatchers,
		window:        rate,
	}.replace(expr)

	return expr.String(), nil
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
