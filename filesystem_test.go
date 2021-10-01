package main

import (
	"testing"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/pyrra-dev/pyrra/slo"
	"github.com/stretchr/testify/require"
)

func TestMatchObjectives(t *testing.T) {
	o1 := slo.Objective{Labels: labels.FromStrings("foo", "bar")}
	o2 := slo.Objective{Labels: labels.FromStrings("foo", "bar", "ying", "yang")}
	o3 := slo.Objective{Labels: labels.FromStrings("foo", "baz")}

	objectives := Objectives{objectives: map[string]slo.Objective{}}
	objectives.Set(o1)
	objectives.Set(o2)
	objectives.Set(o3)

	matches := objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "foo", "foo"),
	})
	require.Nil(t, matches)

	matches = objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "foo", "bar"),
	})
	require.Contains(t, matches, o1)
	require.Contains(t, matches, o2)

	matches = objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "foo", "baz"),
	})
	require.Contains(t, matches, o3)

	matches = objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "foo", "bar"),
		labels.MustNewMatcher(labels.MatchEqual, "ying", "yang"),
	})
	require.Contains(t, matches, o2)

	matches = objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchRegexp, "foo", "ba."),
	})
	require.Contains(t, matches, o1)
	require.Contains(t, matches, o2)
	require.Contains(t, matches, o3)
}
