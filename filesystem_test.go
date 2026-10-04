package main

import (
	"net/url"
	"testing"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"

	"github.com/pyrra-dev/pyrra/slo"
)

func TestReloadEndpoint(t *testing.T) {
	promClient, err := api.NewClient(api.Config{Address: "http://localhost:9090"})
	require.NoError(t, err)

	// Without an override the default Prometheus /-/reload endpoint is used.
	require.Equal(t, "http://localhost:9090/-/reload", reloadEndpoint(promClient, nil).String())

	// An empty override (kong's default for an unset *url.URL flag) is ignored.
	require.Equal(t, "http://localhost:9090/-/reload", reloadEndpoint(promClient, &url.URL{}).String())

	// A configured override takes precedence over the Prometheus URL.
	override := &url.URL{Scheme: "http", Host: "localhost:17902", Path: "/rule/-/reload"}
	require.Equal(t, "http://localhost:17902/rule/-/reload", reloadEndpoint(promClient, override).String())
}

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
