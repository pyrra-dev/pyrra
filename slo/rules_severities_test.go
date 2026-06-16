package slo

import (
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
)

func TestObjective_AlertSeverityLabel(t *testing.T) {
	windows := Windows(28 * 24 * time.Hour)

	t.Run("default severities", func(t *testing.T) {
		o := Objective{
			Alerting: Alerting{
				Severities: AlertingSeverities{},
			},
		}

		// Default severities based on window definitions
		require.Equal(t, "critical", o.alertSeverityLabel(0, windows[0]))
		require.Equal(t, "critical", o.alertSeverityLabel(1, windows[1]))
		require.Equal(t, "warning", o.alertSeverityLabel(2, windows[2]))
		require.Equal(t, "warning", o.alertSeverityLabel(3, windows[3]))
	})

	t.Run("custom severities all set", func(t *testing.T) {
		o := Objective{
			Alerting: Alerting{
				Severities: AlertingSeverities{
					FastBurn:     "page",
					MediumBurn:   "page",
					SlowBurn:     "info",
					LongTermBurn: "info",
				},
			},
		}

		require.Equal(t, "page", o.alertSeverityLabel(0, windows[0]))
		require.Equal(t, "page", o.alertSeverityLabel(1, windows[1]))
		require.Equal(t, "info", o.alertSeverityLabel(2, windows[2]))
		require.Equal(t, "info", o.alertSeverityLabel(3, windows[3]))
	})

	t.Run("custom severities partial", func(t *testing.T) {
		o := Objective{
			Alerting: Alerting{
				Severities: AlertingSeverities{
					FastBurn: "page",
					SlowBurn: "info",
					// MediumBurn and LongTermBurn not set, should use defaults
				},
			},
		}

		require.Equal(t, "page", o.alertSeverityLabel(0, windows[0]))
		require.Equal(t, "critical", o.alertSeverityLabel(1, windows[1])) // falls back to default
		require.Equal(t, "info", o.alertSeverityLabel(2, windows[2]))
		require.Equal(t, "warning", o.alertSeverityLabel(3, windows[3])) // falls back to default
	})
}

func TestObjective_AlertSeverityLabelAbsent(t *testing.T) {
	t.Run("default absent severity", func(t *testing.T) {
		o := Objective{
			Alerting: Alerting{
				Severities: AlertingSeverities{},
			},
		}

		require.Equal(t, "critical", o.alertSeverityLabelAbsent())
	})

	t.Run("custom absent severity", func(t *testing.T) {
		o := Objective{
			Alerting: Alerting{
				Severities: AlertingSeverities{
					Absent: "warning",
				},
			},
		}

		require.Equal(t, "warning", o.alertSeverityLabelAbsent())
	})

	t.Run("empty string absent severity falls back to default", func(t *testing.T) {
		o := Objective{
			Alerting: Alerting{
				Severities: AlertingSeverities{
					Absent: "",
				},
			},
		}

		require.Equal(t, "critical", o.alertSeverityLabelAbsent())
	})
}

func TestObjective_BurnratesWithCustomSeverities(t *testing.T) {
	t.Run("ratio with custom severities", func(t *testing.T) {
		o := Objective{
			Labels: labels.FromStrings(model.MetricNameLabel, "test-slo"),
			Target: 0.99,
			Window: model.Duration(28 * 24 * time.Hour),
			Alerting: Alerting{
				Burnrates: true,
				Absent:    false,
				Severities: AlertingSeverities{
					FastBurn:     "page",
					MediumBurn:   "high",
					SlowBurn:     "medium",
					LongTermBurn: "low",
				},
			},
			Indicator: Indicator{
				Ratio: &RatioIndicator{
					Errors: Metric{
						Name: "http_requests_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "test"},
							{Type: labels.MatchRegexp, Name: "code", Value: "5.."},
							{Type: labels.MatchEqual, Name: "__name__", Value: "http_requests_total"},
						},
					},
					Total: Metric{
						Name: "http_requests_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "test"},
							{Type: labels.MatchEqual, Name: "__name__", Value: "http_requests_total"},
						},
					},
				},
			},
		}

		ruleGroup, err := o.Burnrates(GenerationOptions{})
		require.NoError(t, err)

		// Find alert rules (skip recording rules)
		var alertRules []struct {
			severity string
			short    string
		}
		for _, rule := range ruleGroup.Rules {
			if rule.Alert != "" {
				alertRules = append(alertRules, struct {
					severity string
					short    string
				}{
					severity: rule.Labels["severity"],
					short:    rule.Labels["short"],
				})
			}
		}

		require.Len(t, alertRules, 4, "should have 4 multi-burn-rate alerts")
		require.Equal(t, "page", alertRules[0].severity, "fast burn should use custom severity")
		require.Equal(t, "high", alertRules[1].severity, "medium burn should use custom severity")
		require.Equal(t, "medium", alertRules[2].severity, "slow burn should use custom severity")
		require.Equal(t, "low", alertRules[3].severity, "long-term burn should use custom severity")
	})

	t.Run("latency with custom severities", func(t *testing.T) {
		o := Objective{
			Labels: labels.FromStrings(model.MetricNameLabel, "test-latency-slo"),
			Target: 0.995,
			Window: model.Duration(28 * 24 * time.Hour),
			Alerting: Alerting{
				Burnrates: true,
				Absent:    false,
				Severities: AlertingSeverities{
					FastBurn:     "pagerduty",
					MediumBurn:   "slack-urgent",
					SlowBurn:     "slack",
					LongTermBurn: "email",
				},
			},
			Indicator: Indicator{
				Latency: &LatencyIndicator{
					Success: Metric{
						Name: "http_request_duration_seconds_bucket",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "test"},
							{Type: labels.MatchEqual, Name: "le", Value: "0.5"},
							{Type: labels.MatchEqual, Name: "__name__", Value: "http_request_duration_seconds_bucket"},
						},
					},
					Total: Metric{
						Name: "http_request_duration_seconds_count",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "test"},
							{Type: labels.MatchEqual, Name: "__name__", Value: "http_request_duration_seconds_count"},
						},
					},
				},
			},
		}

		ruleGroup, err := o.Burnrates(GenerationOptions{})
		require.NoError(t, err)

		var alertRules []struct {
			severity string
		}
		for _, rule := range ruleGroup.Rules {
			if rule.Alert != "" {
				alertRules = append(alertRules, struct {
					severity string
				}{
					severity: rule.Labels["severity"],
				})
			}
		}

		require.Len(t, alertRules, 4, "should have 4 multi-burn-rate alerts")
		require.Equal(t, "pagerduty", alertRules[0].severity)
		require.Equal(t, "slack-urgent", alertRules[1].severity)
		require.Equal(t, "slack", alertRules[2].severity)
		require.Equal(t, "email", alertRules[3].severity)
	})
}

func TestObjective_IncreaseRulesWithCustomAbsentSeverity(t *testing.T) {
	t.Run("ratio with custom absent severity", func(t *testing.T) {
		o := Objective{
			Labels: labels.FromStrings(model.MetricNameLabel, "test-slo"),
			Target: 0.99,
			Window: model.Duration(28 * 24 * time.Hour),
			Alerting: Alerting{
				Absent: true,
				Severities: AlertingSeverities{
					Absent: "warning",
				},
			},
			Indicator: Indicator{
				Ratio: &RatioIndicator{
					Errors: Metric{
						Name: "http_requests_errors_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "__name__", Value: "http_requests_errors_total"},
						},
					},
					Total: Metric{
						Name: "http_requests_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "__name__", Value: "http_requests_total"},
						},
					},
				},
			},
		}

		ruleGroup, err := o.IncreaseRules(GenerationOptions{})
		require.NoError(t, err)

		// Find absent alert rules
		var absentAlerts []struct {
			severity string
		}
		for _, rule := range ruleGroup.Rules {
			if rule.Alert == "SLOMetricAbsent" {
				absentAlerts = append(absentAlerts, struct {
					severity string
				}{
					severity: rule.Labels["severity"],
				})
			}
		}

		require.NotEmpty(t, absentAlerts, "should have absent alerts")
		for _, alert := range absentAlerts {
			require.Equal(t, "warning", alert.severity, "absent alert should use custom severity")
		}
	})

	t.Run("latency with custom absent severity", func(t *testing.T) {
		o := Objective{
			Labels: labels.FromStrings(model.MetricNameLabel, "test-latency-slo"),
			Target: 0.995,
			Window: model.Duration(28 * 24 * time.Hour),
			Alerting: Alerting{
				Absent: true,
				Severities: AlertingSeverities{
					Absent: "info",
				},
			},
			Indicator: Indicator{
				Latency: &LatencyIndicator{
					Success: Metric{
						Name: "http_request_duration_seconds_bucket",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "test"},
							{Type: labels.MatchEqual, Name: "le", Value: "0.5"},
							{Type: labels.MatchEqual, Name: "__name__", Value: "http_request_duration_seconds_bucket"},
						},
					},
					Total: Metric{
						Name: "http_request_duration_seconds_count",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "test"},
							{Type: labels.MatchEqual, Name: "__name__", Value: "http_request_duration_seconds_count"},
						},
					},
				},
			},
		}

		ruleGroup, err := o.IncreaseRules(GenerationOptions{})
		require.NoError(t, err)

		var absentAlerts []struct {
			severity string
		}
		for _, rule := range ruleGroup.Rules {
			if rule.Alert == "SLOMetricAbsent" {
				absentAlerts = append(absentAlerts, struct {
					severity string
				}{
					severity: rule.Labels["severity"],
				})
			}
		}

		require.NotEmpty(t, absentAlerts, "should have absent alerts")
		for _, alert := range absentAlerts {
			require.Equal(t, "info", alert.severity, "absent alert should use custom severity")
		}
	})

	t.Run("boolGauge with custom absent severity", func(t *testing.T) {
		o := Objective{
			Labels: labels.FromStrings(model.MetricNameLabel, "test-boolgauge-slo"),
			Target: 0.99,
			Window: model.Duration(28 * 24 * time.Hour),
			Alerting: Alerting{
				Absent: true,
				Severities: AlertingSeverities{
					Absent: "page",
				},
			},
			Indicator: Indicator{
				BoolGauge: &BoolGaugeIndicator{
					Metric: Metric{
						Name: "up",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "test"},
							{Type: labels.MatchEqual, Name: "__name__", Value: "up"},
						},
					},
				},
			},
		}

		ruleGroup, err := o.IncreaseRules(GenerationOptions{})
		require.NoError(t, err)

		var absentAlerts []struct {
			severity string
		}
		for _, rule := range ruleGroup.Rules {
			if rule.Alert == "SLOMetricAbsent" {
				absentAlerts = append(absentAlerts, struct {
					severity string
				}{
					severity: rule.Labels["severity"],
				})
			}
		}

		require.Len(t, absentAlerts, 1, "should have one absent alert")
		require.Equal(t, "page", absentAlerts[0].severity, "absent alert should use custom severity")
	})
}

func TestObjective_Alerts(t *testing.T) {
	t.Run("default severities in Alerts()", func(t *testing.T) {
		o := Objective{
			Labels: labels.FromStrings(model.MetricNameLabel, "test-slo"),
			Target: 0.99,
			Window: model.Duration(28 * 24 * time.Hour),
			Alerting: Alerting{
				Severities: AlertingSeverities{},
			},
			Indicator: Indicator{
				Ratio: &RatioIndicator{
					Errors: Metric{
						Name: "http_requests_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "test"},
							{Type: labels.MatchRegexp, Name: "code", Value: "5.."},
							{Type: labels.MatchEqual, Name: "__name__", Value: "http_requests_total"},
						},
					},
					Total: Metric{
						Name: "http_requests_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "test"},
							{Type: labels.MatchEqual, Name: "__name__", Value: "http_requests_total"},
						},
					},
				},
			},
		}

		alerts, err := o.Alerts()
		require.NoError(t, err)
		require.Len(t, alerts, 4, "should have 4 multi-burn-rate alerts")

		require.Equal(t, "critical", alerts[0].Severity)
		require.Equal(t, "critical", alerts[1].Severity)
		require.Equal(t, "warning", alerts[2].Severity)
		require.Equal(t, "warning", alerts[3].Severity)
	})

	t.Run("custom severities in Alerts()", func(t *testing.T) {
		o := Objective{
			Labels: labels.FromStrings(model.MetricNameLabel, "test-slo"),
			Target: 0.99,
			Window: model.Duration(28 * 24 * time.Hour),
			Alerting: Alerting{
				Severities: AlertingSeverities{
					FastBurn:     "urgent",
					MediumBurn:   "high",
					SlowBurn:     "medium",
					LongTermBurn: "low",
				},
			},
			Indicator: Indicator{
				Ratio: &RatioIndicator{
					Errors: Metric{
						Name: "http_requests_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "test"},
							{Type: labels.MatchRegexp, Name: "code", Value: "5.."},
							{Type: labels.MatchEqual, Name: "__name__", Value: "http_requests_total"},
						},
					},
					Total: Metric{
						Name: "http_requests_total",
						LabelMatchers: []*labels.Matcher{
							{Type: labels.MatchEqual, Name: "job", Value: "test"},
							{Type: labels.MatchEqual, Name: "__name__", Value: "http_requests_total"},
						},
					},
				},
			},
		}

		alerts, err := o.Alerts()
		require.NoError(t, err)
		require.Len(t, alerts, 4, "should have 4 multi-burn-rate alerts")

		require.Equal(t, "urgent", alerts[0].Severity)
		require.Equal(t, "high", alerts[1].Severity)
		require.Equal(t, "medium", alerts[2].Severity)
		require.Equal(t, "low", alerts[3].Severity)
	})
}
