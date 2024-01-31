package slo

import "testing"

func TestObjective_AlertNameMetricAbsent(t *testing.T) {
	tests := []struct {
		name      string
		objective Objective
		want      string
	}{
		{
			name: "alert name present",
			objective: Objective{
				Alerting: Alerting{
					Name: "test-alert",
				},
			},
			want: "test-alert" + "-" + defaultAlertnameAbsent,
		},
		{
			name: "alert name absent",
			objective: Objective{
				Alerting: Alerting{},
			},
			want: defaultAlertnameAbsent,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.objective
			got := o.AlertNameMetricAbsent()
			if got != tt.want {
				t.Errorf("AlertNameMetricAbsent() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestObjective_AlertName(t *testing.T) {
	tests := []struct {
		name      string
		objective Objective
		want      string
	}{
		{
			name: "alert name present",
			objective: Objective{
				Alerting: Alerting{
					Name: "test-alert",
				},
			},
			want: "test-alert",
		},
		{
			name: "alert name absent",
			objective: Objective{
				Alerting: Alerting{},
			},
			want: defaultAlertname,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := tt.objective
			got := o.AlertName()
			if got != tt.want {
				t.Errorf("AlertName() = %v, want %v", got, tt.want)
			}
		})
	}
}
