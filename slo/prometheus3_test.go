package slo

import (
	"testing"

	"github.com/prometheus/prometheus/model/labels"
)

func TestConvertLeMatcherForPrometheus3(t *testing.T) {
	tests := []struct {
		name     string
		matcher  *labels.Matcher
		expected *labels.Matcher
	}{
		{
			name: "integer bucket",
			matcher: &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  labels.BucketLabel,
				Value: "1",
			},
			expected: &labels.Matcher{
				Type:  labels.MatchRegexp,
				Name:  labels.BucketLabel,
				Value: `1(\.0)?`,
			},
		},
		{
			name: "float bucket - unchanged",
			matcher: &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  labels.BucketLabel,
				Value: "2.5",
			},
			expected: &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  labels.BucketLabel,
				Value: "2.5",
			},
		},
		{
			name: "empty bucket - unchanged",
			matcher: &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  labels.BucketLabel,
				Value: "",
			},
			expected: &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  labels.BucketLabel,
				Value: "",
			},
		},
		{
			name: "+Inf bucket - unchanged",
			matcher: &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  labels.BucketLabel,
				Value: "+Inf",
			},
			expected: &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  labels.BucketLabel,
				Value: "+Inf",
			},
		},
		{
			name: "large integer bucket",
			matcher: &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  labels.BucketLabel,
				Value: "1000",
			},
			expected: &labels.Matcher{
				Type:  labels.MatchRegexp,
				Name:  labels.BucketLabel,
				Value: `1000(\.0)?`,
			},
		},
		{
			name: "decimal bucket - unchanged",
			matcher: &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  labels.BucketLabel,
				Value: "0.005",
			},
			expected: &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  labels.BucketLabel,
				Value: "0.005",
			},
		},
		{
			name: "zero bucket",
			matcher: &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  labels.BucketLabel,
				Value: "0",
			},
			expected: &labels.Matcher{
				Type:  labels.MatchRegexp,
				Name:  labels.BucketLabel,
				Value: `0(\.0)?`,
			},
		},
		{
			name: "non-le label unchanged",
			matcher: &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  "job",
				Value: "prometheus",
			},
			expected: &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  "job",
				Value: "prometheus",
			},
		},
		{
			name: "regex matcher unchanged",
			matcher: &labels.Matcher{
				Type:  labels.MatchRegexp,
				Name:  labels.BucketLabel,
				Value: ".*",
			},
			expected: &labels.Matcher{
				Type:  labels.MatchRegexp,
				Name:  labels.BucketLabel,
				Value: ".*",
			},
		},
		{
			name: "float with .0 suffix - unchanged",
			matcher: &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  labels.BucketLabel,
				Value: "5.0",
			},
			expected: &labels.Matcher{
				Type:  labels.MatchEqual,
				Name:  labels.BucketLabel,
				Value: "5.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertLeMatcherForPrometheus3(tt.matcher)

			if result.Type != tt.expected.Type {
				t.Errorf("Type mismatch: got %v, want %v", result.Type, tt.expected.Type)
			}
			if result.Name != tt.expected.Name {
				t.Errorf("Name mismatch: got %v, want %v", result.Name, tt.expected.Name)
			}
			if result.Value != tt.expected.Value {
				t.Errorf("Value mismatch: got %v, want %v", result.Value, tt.expected.Value)
			}
		})
	}
}

func TestApplyPrometheus3Migration(t *testing.T) {
	tests := []struct {
		name     string
		matchers []*labels.Matcher
		opts     GenerationOptions
		expected []*labels.Matcher
	}{
		{
			name: "migration disabled",
			matchers: []*labels.Matcher{
				{Type: labels.MatchEqual, Name: labels.BucketLabel, Value: "1"},
				{Type: labels.MatchEqual, Name: "job", Value: "api"},
			},
			opts: GenerationOptions{EnablePrometheus3Migration: false},
			expected: []*labels.Matcher{
				{Type: labels.MatchEqual, Name: labels.BucketLabel, Value: "1"},
				{Type: labels.MatchEqual, Name: "job", Value: "api"},
			},
		},
		{
			name: "migration enabled",
			matchers: []*labels.Matcher{
				{Type: labels.MatchEqual, Name: labels.BucketLabel, Value: "1"},
				{Type: labels.MatchEqual, Name: "job", Value: "api"},
			},
			opts: GenerationOptions{EnablePrometheus3Migration: true},
			expected: []*labels.Matcher{
				{Type: labels.MatchRegexp, Name: labels.BucketLabel, Value: `1(\.0)?`},
				{Type: labels.MatchEqual, Name: "job", Value: "api"},
			},
		},
		{
			name: "multiple le matchers with integer and float",
			matchers: []*labels.Matcher{
				{Type: labels.MatchEqual, Name: labels.BucketLabel, Value: "1"},
				{Type: labels.MatchEqual, Name: "job", Value: "api"},
				{Type: labels.MatchEqual, Name: labels.BucketLabel, Value: "0.5"},
			},
			opts: GenerationOptions{EnablePrometheus3Migration: true},
			expected: []*labels.Matcher{
				{Type: labels.MatchRegexp, Name: labels.BucketLabel, Value: `1(\.0)?`},
				{Type: labels.MatchEqual, Name: "job", Value: "api"},
				{Type: labels.MatchEqual, Name: labels.BucketLabel, Value: "0.5"},
			},
		},
		{
			name:     "empty matchers",
			matchers: []*labels.Matcher{},
			opts:     GenerationOptions{EnablePrometheus3Migration: true},
			expected: []*labels.Matcher{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyPrometheus3Migration(tt.matchers, tt.opts)

			if len(result) != len(tt.expected) {
				t.Fatalf("Length mismatch: got %d, want %d", len(result), len(tt.expected))
			}

			for i := range result {
				if result[i].Type != tt.expected[i].Type {
					t.Errorf("Matcher %d Type mismatch: got %v, want %v", i, result[i].Type, tt.expected[i].Type)
				}
				if result[i].Name != tt.expected[i].Name {
					t.Errorf("Matcher %d Name mismatch: got %v, want %v", i, result[i].Name, tt.expected[i].Name)
				}
				if result[i].Value != tt.expected[i].Value {
					t.Errorf("Matcher %d Value mismatch: got %v, want %v", i, result[i].Value, tt.expected[i].Value)
				}
			}
		})
	}
}
