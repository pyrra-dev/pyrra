package slo

import (
	"fmt"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

// QueryTotal returns a PromQL query to get the total amount of requests served during the window.
func (o Objective) QueryTotal(window model.Duration) string {
	expr, err := parser.ParseExpr(`sum(increase(metric{}[1s]))`)
	if err != nil {
		return ""
	}

	var metric string
	var matchers []*labels.Matcher

	if o.Indicator.HTTP != nil {
		http := o.Indicator.HTTP
		if http.Metric == "" {
			http.Metric = HTTPDefaultMetric
		}
		metric = http.Metric
		matchers = http.Matchers
	}
	if o.Indicator.GRPC != nil {
		grpc := o.Indicator.GRPC
		if grpc.Metric == "" {
			grpc.Metric = GRPCDefaultMetric
		}
		metric = grpc.Metric
		matchers = grpc.GRPCSelectors()
	}

	metricMatcher := &labels.Matcher{Type: labels.MatchEqual, Name: "__name__", Value: metric}
	matchers = append([]*labels.Matcher{metricMatcher}, matchers...)

	objectiveReplacer{
		metric:   metric,
		matchers: matchers,
		window:   time.Duration(window),
	}.replace(expr)

	return expr.String()
}

// QueryErrors returns a PromQL query to get the amount of request errors during the window.
func (o Objective) QueryErrors(window model.Duration) string {
	expr, err := parser.ParseExpr(`sum(increase(metric{}[1s]))`)
	if err != nil {
		return err.Error()
	}

	var metric string
	var matchers []*labels.Matcher

	if o.Indicator.HTTP != nil {
		http := o.Indicator.HTTP

		metric = http.Metric
		if metric == "" {
			metric = HTTPDefaultMetric
		}

		if len(http.ErrorMatchers) == 0 {
			http.ErrorMatchers = []*labels.Matcher{HTTPDefaultErrorSelector}
		}
		matchers = http.AllSelectors()
	}
	if o.Indicator.GRPC != nil {
		grpc := o.Indicator.GRPC

		metric = grpc.Metric
		if metric == "" {
			metric = GRPCDefaultMetric
		}

		if len(grpc.ErrorMatchers) == 0 {
			grpc.ErrorMatchers = []*labels.Matcher{GRPCDefaultErrorSelector}
		}
		matchers = grpc.AllSelectors()
	}

	metricMatcher := &labels.Matcher{Type: labels.MatchEqual, Name: "__name__", Value: metric}
	matchers = append([]*labels.Matcher{metricMatcher}, matchers...)

	objectiveReplacer{
		metric:   metric,
		matchers: matchers,
		window:   time.Duration(window),
	}.replace(expr)

	return expr.String()
}

type objectiveReplacer struct {
	metric        string
	matchers      []*labels.Matcher
	errorMatchers []*labels.Matcher
	grouping      []string
	window        time.Duration
	target        float64
}

func (o Objective) QueryErrorBudget() string {
	expr, err := parser.ParseExpr(`((1 - 0.696969) - (sum(increase(metric{matchers="errors"}[1s]) or vector(0)) / sum(increase(metric{matchers="total"}[1s])))) / (1 - 0.696969)`)
	if err != nil {
		return ""
	}

	var metric string
	var matchers []*labels.Matcher
	var errorMatchers []*labels.Matcher

	if o.Indicator.HTTP != nil {
		if o.Indicator.HTTP.Metric == "" {
			o.Indicator.HTTP.Metric = HTTPDefaultMetric
		}
		if len(o.Indicator.HTTP.ErrorMatchers) == 0 {
			o.Indicator.HTTP.ErrorMatchers = []*labels.Matcher{HTTPDefaultErrorSelector}
		}

		metric = o.Indicator.HTTP.Metric
		matchers = o.Indicator.HTTP.Matchers
		errorMatchers = o.Indicator.HTTP.AllSelectors()
	}
	if o.Indicator.GRPC != nil {
		if o.Indicator.GRPC.Metric == "" {
			o.Indicator.GRPC.Metric = GRPCDefaultMetric
		}
		if len(o.Indicator.GRPC.ErrorMatchers) == 0 {
			o.Indicator.GRPC.ErrorMatchers = []*labels.Matcher{GRPCDefaultErrorSelector}
		}

		metric = o.Indicator.GRPC.Metric
		matchers = o.Indicator.GRPC.GRPCSelectors()
		errorMatchers = o.Indicator.GRPC.AllSelectors()
	}

	metricMatcher := &labels.Matcher{Type: labels.MatchEqual, Name: "__name__", Value: metric}
	matchers = append([]*labels.Matcher{metricMatcher}, matchers...)
	errorMatchers = append([]*labels.Matcher{metricMatcher}, errorMatchers...)

	objectiveReplacer{
		metric:        metric,
		matchers:      matchers,
		errorMatchers: errorMatchers,
		window:        time.Duration(o.Window),
		target:        o.Target,
	}.replace(expr)

	return expr.String()
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
		n.Name = r.metric
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
		if n.Val == 0.696969 {
			n.Val = r.target
		}
	default:
		panic(fmt.Sprintf("no support for type %T", n))
	}
}

func (o Objective) RequestRange(timerange time.Duration) string {
	//expr, err := parser.ParseExpr(`sum by(group) (rate(metric{}[1s])) > 0`) // TODO
	expr, err := parser.ParseExpr(`sum by(group) (rate(metric{}[1s]))`)
	if err != nil {
		return err.Error()
	}

	var metric string
	var matchers []*labels.Matcher
	var grouping []string

	if o.Indicator.HTTP != nil {
		if o.Indicator.HTTP.Metric == "" {
			o.Indicator.HTTP.Metric = HTTPDefaultMetric
		}
		if len(o.Indicator.HTTP.ErrorMatchers) == 0 {
			o.Indicator.HTTP.ErrorMatchers = []*labels.Matcher{HTTPDefaultErrorSelector}
		}

		metric = o.Indicator.HTTP.Metric
		matchers = o.Indicator.HTTP.Matchers
		for _, m := range o.Indicator.HTTP.ErrorMatchers {
			grouping = append(grouping, m.Name)
		}
	}
	if o.Indicator.GRPC != nil {
		if o.Indicator.GRPC.Metric == "" {
			o.Indicator.GRPC.Metric = GRPCDefaultMetric
		}
		if len(o.Indicator.GRPC.ErrorMatchers) == 0 {
			o.Indicator.GRPC.ErrorMatchers = []*labels.Matcher{GRPCDefaultErrorSelector}
		}

		metric = o.Indicator.GRPC.Metric
		matchers = o.Indicator.GRPC.GRPCSelectors()
		for _, m := range o.Indicator.GRPC.ErrorMatchers {
			grouping = append(grouping, m.Name)
		}
	}

	metricMatcher := &labels.Matcher{Type: labels.MatchEqual, Name: "__name__", Value: metric}
	matchers = append([]*labels.Matcher{metricMatcher}, matchers...)

	objectiveReplacer{
		metric:   metric,
		matchers: matchers,
		grouping: grouping,
		window:   timerange,
		target:   o.Target,
	}.replace(expr)

	return expr.String()
}

func (o Objective) ErrorsRange(timerange time.Duration) string {
	expr, err := parser.ParseExpr(`sum by(group) (rate(metric{matchers="errors"}[1s])) / scalar(sum(rate(metric{matchers="total"}[1s])))`)
	if err != nil {
		return err.Error()
	}

	var (
		metric        string
		matchers      []*labels.Matcher
		errorMatchers []*labels.Matcher
		grouping      []string
	)

	if o.Indicator.HTTP != nil {
		if o.Indicator.HTTP.Metric == "" {
			o.Indicator.HTTP.Metric = HTTPDefaultMetric
		}
		if len(o.Indicator.HTTP.ErrorMatchers) == 0 {
			o.Indicator.HTTP.ErrorMatchers = []*labels.Matcher{HTTPDefaultErrorSelector}
		}

		metric = o.Indicator.HTTP.Metric
		matchers = o.Indicator.HTTP.Matchers
		errorMatchers = o.Indicator.HTTP.AllSelectors()
		for _, m := range o.Indicator.HTTP.ErrorMatchers {
			grouping = append(grouping, m.Name)
		}
	}
	if o.Indicator.GRPC != nil {
		if o.Indicator.GRPC.Metric == "" {
			o.Indicator.GRPC.Metric = GRPCDefaultMetric
		}
		if len(o.Indicator.GRPC.ErrorMatchers) == 0 {
			o.Indicator.GRPC.ErrorMatchers = []*labels.Matcher{GRPCDefaultErrorSelector}
		}

		metric = o.Indicator.GRPC.Metric
		matchers = o.Indicator.GRPC.GRPCSelectors()
		errorMatchers = o.Indicator.GRPC.AllSelectors()
		for _, m := range o.Indicator.GRPC.ErrorMatchers {
			grouping = append(grouping, m.Name)
		}
	}

	metricMatcher := &labels.Matcher{Type: labels.MatchEqual, Name: "__name__", Value: metric}
	matchers = append([]*labels.Matcher{metricMatcher}, matchers...)

	objectiveReplacer{
		metric:        metric,
		matchers:      matchers,
		errorMatchers: errorMatchers,
		grouping:      grouping,
		window:        timerange,
	}.replace(expr)

	return expr.String()
}
