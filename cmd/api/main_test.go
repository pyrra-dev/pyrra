package main

import (
	"testing"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func Test_matrixToValues(t *testing.T) {
	// v0 is from t 0-500 counting all up from 0 to 500
	v0 := make([]model.SamplePair, 500)
	for i := 0; i < cap(v0); i++ {
		v0[i] = model.SamplePair{
			Timestamp: model.Time(i * 1000),
			Value:     model.SampleValue(i),
		}
	}
	e0 := make([][]float64, 500)
	for i := 0; i < cap(e0); i++ {
		e0[i] = []float64{float64(i), float64(i)}
	}

	v1 := make([]model.SamplePair, 250)
	for i := 0; i < cap(v1); i++ {
		v1[i] = model.SamplePair{
			Timestamp: model.Time((i + 100) * 1000),
			Value:     model.SampleValue(i),
		}
	}

	e1 := make([][]float64, 500)
	for i := 0; i < cap(e1); i++ {
		e1[i] = []float64{float64(i), float64(i), 0}
		if i >= 100 && i < 350 {
			e1[i][2] = float64(i - 100)
		}
	}

	for _, tc := range []struct {
		m        []*model.SampleStream
		expected [][]float64
	}{{
		m:        []*model.SampleStream{{Values: v0}},
		expected: e0,
	}, {
		m:        []*model.SampleStream{{Values: v0}, {Values: v1}},
		expected: e1,
	}} {
		t.Run("", func(t *testing.T) {
			require.Equal(t, tc.expected, matrixToValues(tc.m))
		})
	}
}

func Benchmark_matrixToValues(b *testing.B) {
	v := make([]model.SamplePair, b.N)

	for i := 0; i < b.N; i++ {
		v[i] = model.SamplePair{
			Timestamp: model.Time(i * 1000),
			Value:     model.SampleValue(i),
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	matrixToValues([]*model.SampleStream{{Values: v}})
}
