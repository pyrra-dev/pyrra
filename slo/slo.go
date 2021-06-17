package slo

import (
	"strings"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
)

const (
	HTTPDefaultMetric = "http_requests_total"
	GRPCDefaultMetric = "grpc_server_handled_total"
)

var (
	HTTPDefaultErrorSelector = ParseMatcher(`code=~"5.."`)
	GRPCDefaultErrorSelector = ParseMatcher(`grpc_code=~"Aborted|Unavailable|Internal|Unknown|Unimplemented|DataLoss"`)
)

type Objective struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Target      float64        `json:"target"`
	Window      model.Duration `json:"window"`

	Indicator Indicator `json:"indicator"`
}

type Indicator struct {
	HTTP *HTTPIndicator `json:"http,omitempty"`
	GRPC *GRPCIndicator `json:"grpc,omitempty"`
}

type HTTPIndicator struct {
	Metric        string            `json:"metric"`
	Matchers      []*labels.Matcher `json:"matchers"`
	ErrorMatchers []*labels.Matcher `json:"errorMatchers"`
}

func (i HTTPIndicator) AllSelectors() []*labels.Matcher {
	return append(i.Matchers, i.ErrorMatchers...)
}

type GRPCIndicator struct {
	Metric        string            `json:"metric"`
	Service       string            `json:"service"`
	Method        string            `json:"method"`
	Matchers      []*labels.Matcher `json:"matchers"`
	ErrorMatchers []*labels.Matcher `json:"errorMatchers"`
}

func (i GRPCIndicator) GRPCSelectors() []*labels.Matcher {
	return append(i.Matchers,
		[]*labels.Matcher{
			{Type: labels.MatchEqual, Name: "grpc_service", Value: i.Service},
			{Type: labels.MatchEqual, Name: "grpc_method", Value: i.Method},
		}...,
	)
}

func (i GRPCIndicator) AllSelectors() []*labels.Matcher {
	return append(i.GRPCSelectors(), i.ErrorMatchers...)
}

func ParseMatcher(s string) *labels.Matcher {
	// TODO: Find out how the parsing is usually done...
	if strings.Contains(s, "!~") {
		split := strings.Split(s, "!~")
		return &labels.Matcher{Type: labels.MatchNotRegexp, Name: split[0], Value: strings.Trim(split[1], `"`)}
	}
	if strings.Contains(s, "=~") {
		split := strings.Split(s, "=~")
		return &labels.Matcher{Type: labels.MatchRegexp, Name: split[0], Value: strings.Trim(split[1], `"`)}
	}
	if strings.Contains(s, "!=") {
		split := strings.Split(s, "!=")
		return &labels.Matcher{Type: labels.MatchNotEqual, Name: split[0], Value: strings.Trim(split[1], `"`)}
	}

	split := strings.Split(s, "=")
	return &labels.Matcher{Type: labels.MatchEqual, Name: split[0], Value: strings.Trim(split[1], `"`)}
}
