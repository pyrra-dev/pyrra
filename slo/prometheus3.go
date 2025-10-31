package slo

import (
	"strconv"
	"strings"

	"github.com/prometheus/prometheus/model/labels"
)

// GenerationOptions contains configuration for rule generation.
type GenerationOptions struct {
	// EnablePrometheus3Migration enables generation of rules compatible with both
	// Prometheus 2 and 3. This transforms le label matchers from exact matches
	// (e.g., le="1") to regex matches (e.g., le=~"1(\.0)?") to handle the
	// Prometheus 3 change where classic histogram bucket labels are normalized to floats.
	// This only affects classic histograms (Latency indicators); native histograms
	// (LatencyNative indicators) are not affected by this change.
	EnablePrometheus3Migration bool
}

// convertLeMatcherForPrometheus3 converts a le (bucket label) matcher from an exact
// match to a regex match that works with both Prometheus 2 and 3.
//
// Prometheus 3 normalizes classic histogram bucket labels to floats, so le="1" becomes le="1.0".
// This function only converts integer values (values without a decimal point):
//   - le="0" -> le=~"0(\.0)?"
//   - le="1" -> le=~"1(\.0)?"
//   - le="1000" -> le=~"1000(\.0)?"
//
// Values that already contain decimals, empty strings, or +Inf are left unchanged:
//   - le="0.5" -> le="0.5" (unchanged)
//   - le="5.0" -> le="5.0" (unchanged)
//   - le="" -> le="" (unchanged)
//   - le="+Inf" -> le="+Inf" (unchanged)
func convertLeMatcherForPrometheus3(matcher *labels.Matcher) *labels.Matcher {
	// Only convert exact matches on the le (bucket) label
	if matcher.Name != labels.BucketLabel || matcher.Type != labels.MatchEqual {
		return matcher
	}

	value := matcher.Value

	// Don't touch empty values (used for +Inf bucket) or +Inf
	if value == "" || value == "+Inf" {
		return matcher
	}

	// If the value already contains a decimal point, it's already in float format
	// and doesn't need transformation (e.g., "0.5", "5.0")
	if strings.Contains(value, ".") {
		return matcher
	}

	// Try to parse as integer to ensure it's a valid integer value
	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		// Integer value like "1" -> "1(\.0)?"
		return &labels.Matcher{
			Type:  labels.MatchRegexp,
			Name:  labels.BucketLabel,
			Value: value + `(\.0)?`,
		}
	}

	// For unparseable values, return unchanged
	return matcher
}

// applyPrometheus3Migration applies the Prometheus 3 migration to a slice of matchers.
// It clones the matchers and converts any le label matchers to regex form.
func applyPrometheus3Migration(matchers []*labels.Matcher, opts GenerationOptions) []*labels.Matcher {
	if !opts.EnablePrometheus3Migration {
		return matchers
	}

	result := make([]*labels.Matcher, len(matchers))
	for i, m := range matchers {
		result[i] = convertLeMatcherForPrometheus3(m)
	}
	return result
}
