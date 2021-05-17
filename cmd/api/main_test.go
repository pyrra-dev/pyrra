package main

import (
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func Test_relative(t *testing.T) {
	testcases := []struct {
		min float64
		max float64
		v   float64
		ex  float64
	}{
		{min: 0, max: 1, v: 0.5, ex: 0.5},
		{min: 1, max: 2, v: 1.5, ex: 0.5},
		{min: 0, max: 4, v: 0.5, ex: 0.125},
		{min: -1, max: 1, v: 0, ex: 0.5},
		{min: -1, max: 1, v: 0.5, ex: 0.75},
		{min: -2, max: -1, v: -1.5, ex: 0.5},
	}
	for _, tc := range testcases {
		require.Equal(t, tc.ex, relative(tc.min, tc.max, tc.v))
	}
}

func Test_graphLines(t *testing.T) {
	testcases := []struct {
		samples       []model.SamplePair
		start, end    time.Time
		width, height float64
		min, max      float64
		lines         []string
	}{{
		samples: []model.SamplePair{
			{Timestamp: model.TimeFromUnix(0), Value: model.SampleValue(1.0)},
			{Timestamp: model.TimeFromUnix(1), Value: model.SampleValue(0.8)},
			{Timestamp: model.TimeFromUnix(2), Value: model.SampleValue(0.6)},
			{Timestamp: model.TimeFromUnix(3), Value: model.SampleValue(0.4)},
			{Timestamp: model.TimeFromUnix(4), Value: model.SampleValue(0.2)},
			{Timestamp: model.TimeFromUnix(5), Value: model.SampleValue(0.0)},
		},
		start:  time.Unix(0, 0),
		end:    time.Unix(5, 0),
		width:  100.0,
		height: 100.0,
		min:    0.0,
		max:    1.0,
		lines: []string{
			`<path stroke="#2C9938" stroke-width="1.5" stroke-linejoin="round" stroke-linecap="round" fill="none" d="M0 0 L20 20 L40 40 L60 60 L80 80 L100 100 "/>`,
			`<path fill="#2C9938" fill-opacity="0.1" d="M0 0 L20 20 L40 40 L60 60 L80 80 L100 100 V100 H0 V0"/>`,
			`<path stroke="#e6522c" stroke-width="1" stroke-dasharray="20,5" fill="none" d="M0 100 H100"/>`,
		},
	}, {
		samples: []model.SamplePair{
			{Timestamp: model.TimeFromUnix(0), Value: model.SampleValue(1.0)},
			{Timestamp: model.TimeFromUnix(1), Value: model.SampleValue(0.6)},
			{Timestamp: model.TimeFromUnix(2), Value: model.SampleValue(0.2)},
			{Timestamp: model.TimeFromUnix(3), Value: model.SampleValue(-0.2)},
			{Timestamp: model.TimeFromUnix(4), Value: model.SampleValue(-0.6)},
			{Timestamp: model.TimeFromUnix(5), Value: model.SampleValue(-1.0)},
		},
		start:  time.Unix(0, 0),
		end:    time.Unix(5, 0),
		width:  100.0,
		height: 100.0,
		min:    -1.0,
		max:    1.0,
		lines: []string{
			`<path stroke="#2C9938" stroke-width="1.5" stroke-linejoin="round" stroke-linecap="round" fill="none" d="M0 0 L20 20 L40 40 "/>`,
			`<path fill="#2C9938" fill-opacity="0.1" d="M0 0 L20 20 L40 40 V50 H0 V0"/>`,
			`<path stroke="#e6522c" stroke-width="1.5" stroke-linejoin="round" stroke-linecap="round" fill="none" d="M60 60 L80 80 L100 100 "/>`,
			`<path fill="#e6522c" fill-opacity="0.1" d="M60 60 L80 80 L100 100 V50 H60 V60"/>`,
			`<path stroke="#e6522c" stroke-width="1" stroke-dasharray="20,5" fill="none" d="M0 50 H100"/>`,
		},
	}}
	for _, tc := range testcases {
		lines := graphLines(tc.samples, tc.start, tc.end, tc.width, tc.height, tc.min, tc.max)
		require.Equal(t, tc.lines, lines)
	}
}

func TestRoundUp(t *testing.T) {
	testcases := []struct {
		t time.Time
		d time.Duration
		e time.Time
	}{{
		t: time.Date(2020, 2, 2, 2, 2, 2, 0, time.UTC),
		d: time.Minute,
		e: time.Date(2020, 2, 2, 2, 3, 0, 0, time.UTC),
	}, {
		t: time.Date(2020, 2, 2, 2, 1, 45, 0, time.UTC),
		d: time.Minute,
		e: time.Date(2020, 2, 2, 2, 2, 0, 0, time.UTC),
	}, {
		t: time.Date(2020, 2, 2, 2, 2, 2, 0, time.UTC),
		d: 2 * time.Minute,
		e: time.Date(2020, 2, 2, 2, 4, 0, 0, time.UTC),
	}, {
		t: time.Date(2020, 2, 2, 2, 2, 2, 0, time.UTC),
		d: 5 * time.Minute,
		e: time.Date(2020, 2, 2, 2, 5, 0, 0, time.UTC),
	}}
	for _, tc := range testcases {
		require.Equal(t, tc.e, RoundUp(tc.t, tc.d))
	}
}
