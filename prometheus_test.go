package main

import (
	"fmt"
	"testing"

	"github.com/prometheus/common/model"
)

// go test -bench='BenchmarkPrometheus*' -count=5 | tee BenchmarkPrometheus

func BenchmarkPrometheusConvertLabelSet(b *testing.B) {
	metric := model.Metric{}
	for i := 0; i < 99; i++ {
		metric[model.LabelName(fmt.Sprintf("foo%d", i))] = model.LabelValue(fmt.Sprintf("bar%d", i))
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		convertLabelSet(metric)
	}
}

func BenchmarkPrometheusConvertVector(b *testing.B) {
	vector := []*model.Sample{{
		Metric:    model.Metric{"foo0": "bar0", "foo1": "bar1", "foo2": "bar2"},
		Timestamp: 1669152132000,
		Value:     1,
	}, {
		Metric:    model.Metric{"foo0": "bar0", "foo1": "bar1", "foo3": "bar3"},
		Timestamp: 1669152132000,
		Value:     1,
	}, {
		Metric:    model.Metric{"foo0": "bar0", "foo1": "bar1", "foo4": "bar4"},
		Timestamp: 1669152132000,
		Value:     1,
	}}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		convertVector(vector)
	}
}

func BenchmarkPrometheusConvertMatrix(b *testing.B) {
	sp := make([]model.SamplePair, 4_000)
	for i := 0; i < 4_000; i++ {
		sp[i].Timestamp = model.Time(i)
		sp[i].Value = model.SampleValue(i)
	}
	m := model.Matrix{{Values: sp}}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		convertMatrix(m)
	}
}
