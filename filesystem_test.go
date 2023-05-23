package main

import (
	"testing"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"

	"github.com/pyrra-dev/pyrra/slo"
)

func TestMatchObjectives(t *testing.T) {
	obj1 := slo.Objective{Labels: labels.FromStrings("foo", "bar")}
	obj2 := slo.Objective{Labels: labels.FromStrings("foo", "bar", "ying", "yang")}
	obj3 := slo.Objective{Labels: labels.FromStrings("foo", "bar", "yes", "no")}
	obj4 := slo.Objective{Labels: labels.FromStrings("foo", "baz")}

	objectives := Objectives{objectives: map[string]slo.Objective{}}
	objectives.Set(obj1)
	objectives.Set(obj2)
	objectives.Set(obj3)
	objectives.Set(obj4)

	matches := objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "foo", "foo"),
	})
	require.Nil(t, matches)

	matches = objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "foo", "bar"),
	})
	require.Len(t, matches, 3)
	require.Contains(t, matches, obj1)
	require.Contains(t, matches, obj2)
	require.Contains(t, matches, obj3)

	matches = objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "foo", "baz"),
	})
	require.Len(t, matches, 1)
	require.Contains(t, matches, obj4)

	matches = objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchEqual, "foo", "bar"),
		labels.MustNewMatcher(labels.MatchEqual, "ying", "yang"),
	})
	require.Len(t, matches, 1)
	require.Contains(t, matches, obj2)

	matches = objectives.Match([]*labels.Matcher{
		labels.MustNewMatcher(labels.MatchRegexp, "foo", "ba."),
	})
	require.Len(t, matches, 4)
	require.Contains(t, matches, obj1)
	require.Contains(t, matches, obj2)
	require.Contains(t, matches, obj3)
	require.Contains(t, matches, obj4)
}
