package objectivesv1alpha1

import (
	"testing"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"

	"github.com/pyrra-dev/pyrra/slo"
)

// TestRatioSuccessRoundTrip ensures the ratio success metric survives the
// conversion from the internal model to the proto representation and back.
func TestRatioSuccessRoundTrip(t *testing.T) {
	for name, ratio := range map[string]*slo.RatioIndicator{
		"successTotal": {
			Total: slo.Metric{
				Name: "http_requests_total",
				LabelMatchers: []*labels.Matcher{
					{Type: labels.MatchEqual, Name: labels.MetricName, Value: "http_requests_total"},
					{Type: labels.MatchEqual, Name: "job", Value: "api"},
				},
			},
			Success: slo.Metric{
				Name: "http_success_total",
				LabelMatchers: []*labels.Matcher{
					{Type: labels.MatchEqual, Name: labels.MetricName, Value: "http_success_total"},
					{Type: labels.MatchEqual, Name: "job", Value: "api"},
				},
			},
		},
		"errorsSuccess": {
			Errors: slo.Metric{
				Name: "http_errors_total",
				LabelMatchers: []*labels.Matcher{
					{Type: labels.MatchEqual, Name: labels.MetricName, Value: "http_errors_total"},
					{Type: labels.MatchEqual, Name: "job", Value: "api"},
				},
			},
			Success: slo.Metric{
				Name: "http_success_total",
				LabelMatchers: []*labels.Matcher{
					{Type: labels.MatchEqual, Name: labels.MetricName, Value: "http_success_total"},
					{Type: labels.MatchEqual, Name: "job", Value: "api"},
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			in := slo.Objective{Indicator: slo.Indicator{Ratio: ratio}}

			pb := FromInternal(in)
			require.NotNil(t, pb.Indicator.GetRatio())
			require.Equal(t, ratio.Success.Name, pb.Indicator.GetRatio().GetSuccess().GetName())

			out := ToInternal(pb)
			require.Equal(t, ratio.Success.Name, out.Indicator.Ratio.Success.Name)
			require.Equal(t, ratio.Success.LabelMatchers, out.Indicator.Ratio.Success.LabelMatchers)
			require.Equal(t, ratio.Errors.Name, out.Indicator.Ratio.Errors.Name)
			require.Equal(t, ratio.Total.Name, out.Indicator.Ratio.Total.Name)
		})
	}
}
