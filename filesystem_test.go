package main

import (
	"testing"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/pyrra-dev/pyrra/slo"
	"github.com/stretchr/testify/require"
)

func TestMatchObjectives(t *testing.T) {
	o1 := slo.Objective{Labels: labels.FromStrings("foo", "bar")}
	o2 := slo.Objective{Labels: labels.FromStrings("foo", "bar", "ying", "yang")}
	o3 := slo.Objective{Labels: labels.FromStrings("foo", "bar", "yes", "no")}
	o4 := slo.Objective{Labels: labels.FromStrings("foo", "baz")}

	objectives := Objectives{objectives: map[string]slo.Objective{}}
	objectives.Set(o1)
	objectives.Set(o2)
	objectives.Set(o3)
	objectives.Set(o4)

	matches := objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "foo", "foo"),
	})
	require.Nil(t, matches)

	matches = objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "foo", "bar"),
	})
	require.Len(t, matches, 3)
	require.Contains(t, matches, o1)
	require.Contains(t, matches, o2)
	require.Contains(t, matches, o3)

	matches = objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "foo", "baz"),
	})
	require.Len(t, matches, 1)
	require.Contains(t, matches, o4)

	matches = objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "foo", "bar"),
		labels.MustNewMatcher(labels.MatchEqual, "ying", "yang"),
	})
	require.Len(t, matches, 1)
	require.Contains(t, matches, o2)

	matches = objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchRegexp, "foo", "ba."),
	})
	require.Len(t, matches, 4)
	require.Contains(t, matches, o1)
	require.Contains(t, matches, o2)
	require.Contains(t, matches, o3)
	require.Contains(t, matches, o4)
}
