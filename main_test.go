package main

import (
	"math"
	"testing"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func TestMatrixToValues(t *testing.T) {
	// v0 is from t 0-500 counting all up from 0 to 500
	v0 := make([]model.SamplePair, 500)
	e0 := [][]float64{
		make([]float64, 500),
		make([]float64, 500),
	}
	for i := 0; i < cap(v0); i++ {
		v0[i] = model.SamplePair{
			Timestamp: model.Time(i * 1000),
			Value:     model.SampleValue(i),
		}
		e0[0][i] = float64(i)
		e0[1][i] = float64(i)
	}

	v10 := make([]model.SamplePair, 100)
	for i := 0; i < cap(v10); i++ {
		v10[i] = model.SamplePair{
			Timestamp: model.Time(i * 1000),
			Value:     model.SampleValue(i),
		}
	}
	// offset by first 50 samples
	v11 := make([]model.SamplePair, 250)
	for i := 0; i < cap(v11); i++ {
		v11[i] = model.SamplePair{
			Timestamp: model.Time((i + 50) * 1000),
			Value:     model.SampleValue(i),
		}
	}

	e1 := [][]float64{
		make([]float64, 300), // [0-100] + [50-300]
		make([]float64, 300),
		make([]float64, 300),
	}
	for i := 0; i < 300; i++ {
		e1[0][i] = float64(i)
	}
	for i := 0; i < 100; i++ {
		e1[1][i] = float64(i)
	}
	for i := 0; i < 250; i++ {
		e1[2][50+i] = float64(i)
	}

	// Check if NaNs are returned as 0 (it's fine for errors for example to convert these).
	// Additionally, NaNs aren't possible to be marshalled to JSON. Not sure if there's a better way.
	v2 := make([]model.SamplePair, 100)
	for i := 0; i < cap(v2); i++ {
		v2[i] = model.SamplePair{
			Timestamp: model.Time(i * 1000),
			Value:     model.SampleValue(math.NaN()),
		}
	}
	e2 := [][]float64{
		make([]float64, 100),
		make([]float64, 100),
	}
	for i := 0; i < len(e2[0]); i++ {
		e2[0][i] = float64(i)
	}

	// Check NaN in multiple series
	v3 := make([]model.SamplePair, 100)
	for i := 0; i < len(v3); i++ {
		value := float64(i)
		if i%11 == 0 {
			value = math.NaN()
		}
		v3[i] = model.SamplePair{
			Timestamp: model.Time(i * 1000),
			Value:     model.SampleValue(value),
		}
	}
	e3 := [][]float64{
		make([]float64, 100), // x
		make([]float64, 100), // y[0]
		make([]float64, 100), // y[1]
	}
	for i := 0; i < len(e3[0]); i++ {
		e32value := float64(i)
		if i%11 == 0 {
			e32value = 0
		}
		e3[0][i] = float64(i)
		e3[1][i] = 0
		e3[2][i] = e32value
	}

	for _, tc := range []struct {
		name     string
		m        []*model.SampleStream
		expected [][]float64
	}{{
		name: "empty",
	}, {
		name:     "simple",
		m:        []*model.SampleStream{{Values: v0}},
		expected: e0,
	}, {
		name:     "overlapping",
		m:        []*model.SampleStream{{Values: v10}, {Values: v11}},
		expected: e1,
	}, {
		name:     "NaN",
		m:        []*model.SampleStream{{Values: v2}},
		expected: e2,
	}, {
		name:     "NaNMultiple",
		m:        []*model.SampleStream{{Values: v2}, {Values: v3}},
		expected: e3,
	}} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, matrixToValues(tc.m))
		})
	}
}

func BenchmarkMatrixToValues(b *testing.B) {
	b.Run("one", func(b *testing.B) {
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
	})

	b.Run("two", func(b *testing.B) {
		m := make([]*model.SampleStream, 2)
		for n := 0; n < 2; n++ {
			m[n] = &model.SampleStream{Values: make([]model.SamplePair, b.N)}
		}
		for i := 0; i < b.N; i++ {
			for n := 0; n < 2; n++ {
				m[n].Values[i] = model.SamplePair{
					Timestamp: model.Time(i * 1000),
					Value:     model.SampleValue(i),
				}
			}
		}

		b.ReportAllocs()
		b.ResetTimer()
		matrixToValues(m)
	})

	b.Run("five", func(b *testing.B) {
		m := make([]*model.SampleStream, 5)
		for n := 0; n < 5; n++ {
			m[n] = &model.SampleStream{Values: make([]model.SamplePair, b.N)}
		}
		for i := 0; i < b.N; i++ {
			for n := 0; n < 5; n++ {
				m[n].Values[i] = model.SamplePair{
					Timestamp: model.Time(i * 1000),
					Value:     model.SampleValue(i),
				}
			}
		}

		b.ReportAllocs()
		b.ResetTimer()
		matrixToValues(m)
	})
}
