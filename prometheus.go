package main

import (
	"context"
	"time"

	"connectrpc.com/connect/v2"
	"github.com/go-kit/log"
	prometheusapiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	v1 "github.com/pyrra-dev/pyrra/proto/prometheus/v1"
)

type prometheusServer struct {
	logger  log.Logger
	promAPI *promCache
}

func (ps *prometheusServer) Query(ctx context.Context, req *v1.QueryRequest) (*v1.QueryResponse, error) {
	value, warnings, err := ps.promAPI.Query(ctx, req.Query, time.Unix(req.Time, 0))
	if err != nil {
		return nil, unknownError(err)
	}

	switch v := value.(type) {
	case *model.String:
		return &v1.QueryResponse{
			Warnings: warnings,
			Options: &v1.QueryResponse_String_{
				String_: &v1.String{
					Time:  v.Timestamp.Unix(),
					Value: v.Value,
				},
			},
		}, err
	case *model.Scalar:
		return &v1.QueryResponse{
			Warnings: warnings,
			Options: &v1.QueryResponse_Scalar{
				Scalar: &v1.SamplePair{
					Time:  v.Timestamp.Unix(),
					Value: float64(v.Value),
				},
			},
		}, nil
	case model.Vector:
		vector := convertVector(v)
		return &v1.QueryResponse{
			Warnings: warnings,
			Options: &v1.QueryResponse_Vector{
				Vector: vector,
			},
		}, nil
	}

	return &v1.QueryResponse{
		Warnings: warnings,
		Options:  nil,
	}, nil
}

func (ps *prometheusServer) QueryRange(ctx context.Context, req *v1.QueryRangeRequest) (*v1.QueryRangeResponse, error) {
	value, warnings, err := ps.promAPI.QueryRange(ctx, req.GetQuery(), prometheusapiv1.Range{
		Start: time.Unix(req.GetStart(), 0),
		End:   time.Unix(req.GetEnd(), 0),
		Step:  time.Duration(req.GetStep()) * time.Second,
	})
	if err != nil {
		return nil, unknownError(err)
	}

	switch v := value.(type) {
	case model.Matrix:
		return &v1.QueryRangeResponse{
			Warnings: warnings,
			Options: &v1.QueryRangeResponse_Matrix{
				Matrix: convertMatrix(v),
			},
		}, nil
	}

	return &v1.QueryRangeResponse{
		Warnings: warnings,
	}, nil
}

// unixOrZero converts a unix-seconds timestamp to time.Time, treating 0 as
// unbounded (the zero time.Time), matching Prometheus' own UI which queries
// label names/values without a time range.
func unixOrZero(t int64) time.Time {
	if t == 0 {
		return time.Time{}
	}
	return time.Unix(t, 0)
}

func (ps *prometheusServer) LabelNames(ctx context.Context, req *v1.LabelNamesRequest) (*v1.LabelNamesResponse, error) {
	names, warnings, err := ps.promAPI.LabelNames(
		ctx,
		req.GetMatchers(),
		unixOrZero(req.GetStart()),
		unixOrZero(req.GetEnd()),
	)
	if err != nil {
		return nil, unknownError(err)
	}

	// The API returns model.LabelNames, which is a []model.LabelName, while the
	// response carries plain strings.
	labelNames := make([]string, 0, len(names))
	for _, name := range names {
		labelNames = append(labelNames, string(name))
	}

	return &v1.LabelNamesResponse{
		Names:    labelNames,
		Warnings: warnings,
	}, nil
}

func (ps *prometheusServer) LabelValues(ctx context.Context, req *v1.LabelValuesRequest) (*v1.LabelValuesResponse, error) {
	values, warnings, err := ps.promAPI.LabelValues(
		ctx,
		req.GetLabel(),
		req.GetMatchers(),
		unixOrZero(req.GetStart()),
		unixOrZero(req.GetEnd()),
	)
	if err != nil {
		return nil, unknownError(err)
	}

	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, string(v))
	}

	return &v1.LabelValuesResponse{
		Values:   out,
		Warnings: warnings,
	}, nil
}

func convertVector(in model.Vector) *v1.Vector {
	samples := make([]*v1.Sample, 0, len(in))
	for _, si := range in {
		samples = append(samples,
			&v1.Sample{
				Metric: convertLabelSet(si.Metric),
				Time:   si.Timestamp.Unix(),
				Value:  float64(si.Value),
			},
		)
	}
	return &v1.Vector{Samples: samples}
}

func convertLabelSet(metric model.Metric) map[string]string {
	out := make(map[string]string, len(metric))
	for name, value := range metric {
		out[string(name)] = string(value)
	}
	return out
}

func convertMatrix(in model.Matrix) *v1.Matrix {
	samples := make([]*v1.SampleStream, in.Len())
	for i, sampleStream := range in {
		values := make([]*v1.SamplePair, len(sampleStream.Values))
		for j, pair := range sampleStream.Values {
			values[j] = &v1.SamplePair{
				Time:  pair.Timestamp.Unix(),
				Value: float64(pair.Value),
			}
		}
		samples[i] = &v1.SampleStream{
			Metric: convertLabelSet(sampleStream.Metric),
			Values: values,
		}
	}
	return &v1.Matrix{Samples: samples}
}

// unknownError keeps the Prometheus API error message on the wire, e.g. a PromQL parse error.
// connect v2 drops the message of any error that isn't a *connect.Error.
func unknownError(err error) error {
	return connect.NewError(connect.CodeUnknown, err.Error()).WithCause(err)
}
