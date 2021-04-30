package slo

import (
	"fmt"
	"strings"

	"github.com/prometheus/common/model"
)

const (
	HTTPDefaultMetric        = "http_requests_total"
	HTTPDefaultErrorSelector = `code=~"5.."`
	GRPCDefaultMetric        = "grpc_server_handled_total"
	GRPCDefaultErrorSelector = `grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss"`
)

type Objective struct {
	Name   string
	Target float64
	Window model.Duration

	Indicator Indicator
}

type Indicator struct {
	HTTP *HTTPIndicator
	GRPC *GRPCIndicator
}

type HTTPIndicator struct {
	Metric         string
	Selectors      Selectors
	ErrorSelectors Selectors
}

func (i HTTPIndicator) AllSelectors() string {
	return strings.Join(append(i.Selectors, i.ErrorSelectors...), ",")
}

type GRPCIndicator struct {
	Metric         string
	Service        string
	Method         string
	Selectors      Selectors
	ErrorSelectors Selectors
}

func (i GRPCIndicator) GRPCSelectors() string {
	selectors := Selectors{
		fmt.Sprintf(`grpc_service="%s"`, i.Service),
		fmt.Sprintf(`grpc_method="%s"`, i.Method),
	}
	selectors = append(selectors, i.Selectors...)
	return strings.Join(selectors, ",")
}

func (i GRPCIndicator) AllSelectors() string {
	return strings.Join([]string{
		i.GRPCSelectors(),
		i.ErrorSelectors.String(),
	}, ",")
}

type Selectors []string

func (s Selectors) String() string {
	return strings.Join(s, ",")
}
