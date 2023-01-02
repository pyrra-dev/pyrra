package main

import (
	"context"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/go-kit/log"
	prometheusapiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	v1 "github.com/pyrra-dev/pyrra/proto/prometheus/v1"
)

type prometheusServer struct {
	logger  log.Logger
	promAPI *promCache
}

func (ps *prometheusServer) Query(ctx context.Context, req *connect.Request[v1.QueryRequest]) (*connect.Response[v1.QueryResponse], error) {
	value, warnings, err := ps.promAPI.Query(ctx, req.Msg.Query, time.Unix(req.Msg.Time, 0))
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case *model.String:
		return connect.NewResponse(&v1.QueryResponse{
			Warnings: warnings,
			Options: &v1.QueryResponse_String_{
				String_: &v1.String{
					Time:  v.Timestamp.Unix(),
					Value: v.Value,
				},
			},
		}), err
	case *model.Scalar:
		return connect.NewResponse(&v1.QueryResponse{
			Warnings: warnings,
			Options: &v1.QueryResponse_Scalar{
				Scalar: &v1.SamplePair{
					Time:  v.Timestamp.Unix(),
					Value: float64(v.Value),
				},
			},
		}), nil
	case model.Vector:
		vector := convertVector(v)
		return connect.NewResponse(&v1.QueryResponse{
			Warnings: warnings,
			Options: &v1.QueryResponse_Vector{
				Vector: vector,
			},
		}), nil
	}

	return connect.NewResponse(&v1.QueryResponse{
		Warnings: warnings,
		Options:  nil,
	}), nil
}

func (ps *prometheusServer) QueryRange(ctx context.Context, req *connect.Request[v1.QueryRangeRequest]) (*connect.Response[v1.QueryRangeResponse], error) {
	value, warnings, err := ps.promAPI.QueryRange(ctx, req.Msg.GetQuery(), prometheusapiv1.Range{
		Start: time.Unix(req.Msg.GetStart(), 0),
		End:   time.Unix(req.Msg.GetEnd(), 0),
		Step:  time.Duration(req.Msg.GetStep()) * time.Second,
	})
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case model.Matrix:
		return connect.NewResponse(&v1.QueryRangeResponse{
			Warnings: warnings,
			Options: &v1.QueryRangeResponse_Matrix{
				Matrix: convertMatrix(v),
			},
		}), nil
	}

	return connect.NewResponse(&v1.QueryRangeResponse{
		Warnings: warnings,
	}), nil
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
