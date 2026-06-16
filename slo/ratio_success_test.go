package slo

import (
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
)

// objectiveRatioSuccessTotal configures a ratio indicator with the success and
// total metrics set (errors is derived as total - success).
func objectiveRatioSuccessTotal() Objective {
	o := objectiveHTTPRatio()
	o.Indicator.Ratio = &RatioIndicator{
		Total: Metric{
			Name: "http_requests_total",
			LabelMatchers: []*labels.Matcher{
				{Type: labels.MatchEqual, Name: "job", Value: "api"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "http_requests_total"},
			},
		},
		Success: Metric{
			Name: "http_success_total",
			LabelMatchers: []*labels.Matcher{
				{Type: labels.MatchEqual, Name: "job", Value: "api"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "http_success_total"},
			},
		},
	}
	return o
}

// objectiveRatioErrorsSuccess configures a ratio indicator with the errors and
// success metrics set (total is derived as errors + success).
func objectiveRatioErrorsSuccess() Objective {
	o := objectiveHTTPRatio()
	o.Indicator.Ratio = &RatioIndicator{
		Errors: Metric{
			Name: "http_errors_total",
			LabelMatchers: []*labels.Matcher{
				{Type: labels.MatchEqual, Name: "job", Value: "api"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "http_errors_total"},
			},
		},
		Success: Metric{
			Name: "http_success_total",
			LabelMatchers: []*labels.Matcher{
				{Type: labels.MatchEqual, Name: "job", Value: "api"},
				{Type: labels.MatchEqual, Name: "__name__", Value: "http_success_total"},
			},
		},
	}
	return o
}

func TestRatioIndicator_Combo(t *testing.T) {
	require.Equal(t, RatioErrorsTotal, objectiveHTTPRatio().Indicator.Ratio.Combo())
	require.Equal(t, RatioSuccessTotal, objectiveRatioSuccessTotal().Indicator.Ratio.Combo())
	require.Equal(t, RatioErrorsSuccess, objectiveRatioErrorsSuccess().Indicator.Ratio.Combo())

	// The error+total combination keeps reporting as a ratio indicator.
	require.Equal(t, Ratio, objectiveHTTPRatio().IndicatorType())
	require.Equal(t, Ratio, objectiveRatioSuccessTotal().IndicatorType())
	require.Equal(t, Ratio, objectiveRatioErrorsSuccess().IndicatorType())
}

func TestObjective_Burnrate_SuccessCombos(t *testing.T) {
	require.Equal(t,
		`(sum(rate(http_requests_total{job="api"}[5m])) - sum(rate(http_success_total{job="api"}[5m]))) / sum(rate(http_requests_total{job="api"}[5m]))`,
		objectiveRatioSuccessTotal().Burnrate(5*time.Minute, GenerationOptions{}),
	)
	require.Equal(t,
		`sum(rate(http_errors_total{job="api"}[5m])) / (sum(rate(http_errors_total{job="api"}[5m])) + sum(rate(http_success_total{job="api"}[5m])))`,
		objectiveRatioErrorsSuccess().Burnrate(5*time.Minute, GenerationOptions{}),
	)
}

func TestObjective_BurnrateName_SuccessCombos(t *testing.T) {
	require.Equal(t, "http_requests:burnrate5m", objectiveRatioSuccessTotal().BurnrateName(5*time.Minute))
	require.Equal(t, "http_errors:burnrate5m", objectiveRatioErrorsSuccess().BurnrateName(5*time.Minute))
}

func TestObjective_QueryTotal_SuccessCombos(t *testing.T) {
	window := model.Duration(28 * 24 * time.Hour)
	require.Equal(t,
		`sum(http_requests:increase4w{job="api",slo="monitoring-http-errors"})`,
		objectiveRatioSuccessTotal().QueryTotal(window, GenerationOptions{}),
	)
	require.Equal(t,
		`sum(http_errors:increase4w{job="api",slo="monitoring-http-errors"}) + sum(http_success:increase4w{job="api",slo="monitoring-http-errors"})`,
		objectiveRatioErrorsSuccess().QueryTotal(window, GenerationOptions{}),
	)
}

func TestObjective_QueryErrors_SuccessCombos(t *testing.T) {
	window := model.Duration(28 * 24 * time.Hour)
	require.Equal(t,
		`sum(http_requests:increase4w{job="api",slo="monitoring-http-errors"}) - sum(http_success:increase4w{job="api",slo="monitoring-http-errors"})`,
		objectiveRatioSuccessTotal().QueryErrors(window, GenerationOptions{}),
	)
	require.Equal(t,
		`sum(http_errors:increase4w{job="api",slo="monitoring-http-errors"})`,
		objectiveRatioErrorsSuccess().QueryErrors(window, GenerationOptions{}),
	)
}

func TestObjective_QueryErrorBudget_SuccessCombos(t *testing.T) {
	require.Equal(t,
		`((1 - 0.99) - ((sum(http_requests:increase4w{job="api",slo="monitoring-http-errors"}) - sum(http_success:increase4w{job="api",slo="monitoring-http-errors"} or vector(0))) / sum(http_requests:increase4w{job="api",slo="monitoring-http-errors"}))) / (1 - 0.99)`,
		objectiveRatioSuccessTotal().QueryErrorBudget(GenerationOptions{}),
	)
	require.Equal(t,
		`((1 - 0.99) - (sum(http_errors:increase4w{job="api",slo="monitoring-http-errors"} or vector(0)) / (sum(http_errors:increase4w{job="api",slo="monitoring-http-errors"} or vector(0)) + sum(http_success:increase4w{job="api",slo="monitoring-http-errors"})))) / (1 - 0.99)`,
		objectiveRatioErrorsSuccess().QueryErrorBudget(GenerationOptions{}),
	)
}

func TestObjective_RequestRange_SuccessCombos(t *testing.T) {
	require.Equal(t,
		`sum(rate(http_requests_total{job="api"}[5m])) > 0`,
		objectiveRatioSuccessTotal().RequestRange(5*time.Minute, GenerationOptions{}),
	)
	require.Equal(t,
		`sum(rate(http_errors_total{job="api"}[5m])) + sum(rate(http_success_total{job="api"}[5m])) > 0`,
		objectiveRatioErrorsSuccess().RequestRange(5*time.Minute, GenerationOptions{}),
	)
}

func TestObjective_ErrorsRange_SuccessCombos(t *testing.T) {
	require.Equal(t,
		`(sum(rate(http_requests_total{job="api"}[5m])) - sum(rate(http_success_total{job="api"}[5m]))) / scalar(sum(rate(http_requests_total{job="api"}[5m]))) > 0`,
		objectiveRatioSuccessTotal().ErrorsRange(5*time.Minute, GenerationOptions{}),
	)
	require.Equal(t,
		`sum(rate(http_errors_total{job="api"}[5m])) / scalar(sum(rate(http_errors_total{job="api"}[5m])) + sum(rate(http_success_total{job="api"}[5m]))) > 0`,
		objectiveRatioErrorsSuccess().ErrorsRange(5*time.Minute, GenerationOptions{}),
	)
}

// genericRule returns the recorded expression for the given record name.
func genericRule(t *testing.T, o Objective, record string) string {
	t.Helper()
	group, err := o.GenericRules(GenerationOptions{})
	require.NoError(t, err)
	for _, r := range group.Rules {
		if r.Record == record {
			return r.Expr.StrVal
		}
	}
	t.Fatalf("record %q not found", record)
	return ""
}

func TestObjective_GenericRules_SuccessCombos(t *testing.T) {
	t.Run("successTotal", func(t *testing.T) {
		o := objectiveRatioSuccessTotal()
		require.Equal(t,
			`1 - (sum(http_requests:increase4w{job="api",slo="monitoring-http-errors"}) - sum(http_success:increase4w{job="api",slo="monitoring-http-errors"} or vector(0))) / sum(http_requests:increase4w{job="api",slo="monitoring-http-errors"})`,
			genericRule(t, o, "pyrra_availability"),
		)
		require.Equal(t,
			`sum(rate(http_requests_total{job="api"}[5m]))`,
			genericRule(t, o, "pyrra_requests:rate5m"),
		)
		require.Equal(t,
			`sum(rate(http_requests_total{job="api"}[5m])) - sum(rate(http_success_total{job="api"}[5m])) or vector(0)`,
			genericRule(t, o, "pyrra_errors:rate5m"),
		)
	})

	t.Run("errorsSuccess", func(t *testing.T) {
		o := objectiveRatioErrorsSuccess()
		require.Equal(t,
			`1 - sum(http_errors:increase4w{job="api",slo="monitoring-http-errors"} or vector(0)) / (sum(http_errors:increase4w{job="api",slo="monitoring-http-errors"} or vector(0)) + sum(http_success:increase4w{job="api",slo="monitoring-http-errors"}))`,
			genericRule(t, o, "pyrra_availability"),
		)
		require.Equal(t,
			`sum(rate(http_errors_total{job="api"}[5m])) + sum(rate(http_success_total{job="api"}[5m]))`,
			genericRule(t, o, "pyrra_requests:rate5m"),
		)
		require.Equal(t,
			`sum(rate(http_errors_total{job="api"}[5m])) or vector(0)`,
			genericRule(t, o, "pyrra_errors:rate5m"),
		)
	})
}

// increaseRecords returns the records produced by IncreaseRules.
func increaseRecords(t *testing.T, o Objective) map[string]string {
	t.Helper()
	group, err := o.IncreaseRules(GenerationOptions{})
	require.NoError(t, err)
	records := map[string]string{}
	for _, r := range group.Rules {
		if r.Record != "" {
			records[r.Record] = r.Expr.StrVal
		}
	}
	return records
}

func TestObjective_IncreaseRules_SuccessCombos(t *testing.T) {
	t.Run("successTotal", func(t *testing.T) {
		records := increaseRecords(t, objectiveRatioSuccessTotal())
		require.Equal(t, `sum(increase(http_requests_total{job="api"}[4w]))`, records["http_requests:increase4w"])
		require.Equal(t, `sum(increase(http_success_total{job="api"}[4w]))`, records["http_success:increase4w"])
	})

	t.Run("errorsSuccess", func(t *testing.T) {
		records := increaseRecords(t, objectiveRatioErrorsSuccess())
		require.Equal(t, `sum(increase(http_errors_total{job="api"}[4w]))`, records["http_errors:increase4w"])
		require.Equal(t, `sum(increase(http_success_total{job="api"}[4w]))`, records["http_success:increase4w"])
	})
}
