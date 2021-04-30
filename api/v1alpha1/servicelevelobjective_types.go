/*
Copyright 2021 Athena Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&ServiceLevelObjective{}, &ServiceLevelObjectiveList{})
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true

// ServiceLevelObjectiveList contains a list of ServiceLevelObjective
type ServiceLevelObjectiveList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceLevelObjective `json:"items"`
}

// +kubebuilder:object:root=true

// ServiceLevelObjective is the Schema for the ServiceLevelObjectives API
type ServiceLevelObjective struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceLevelObjectiveSpec   `json:"spec,omitempty"`
	Status ServiceLevelObjectiveStatus `json:"status,omitempty"`
}

// ServiceLevelObjectiveSpec defines the desired state of ServiceLevelObjective
type ServiceLevelObjectiveSpec struct {
	// Target is a string that's casted to a float64 between 0 - 100.
	// It represents the desired availability of the service in the given window.
	// float64 are not supported: https://github.com/kubernetes-sigs/controller-tools/issues/245
	Target string `json:"target"`

	// Window within which the Target is supposed to be kept. Usually something like 1d, 7d or 28d.
	Window metav1.Duration `json:"window"`

	// Latency target for the service. Combined with the Target this is used to define things like:
	// 99% of the requests need to be served within 5s.
	// +optional
	Latency string `json:"latency"`

	// ServiceLevelIndicator is the underlying data source that indicates how the service is doing.
	// This will be a Prometheus metric with specific selectors for your service.
	ServiceLevelIndicator ServiceLevelIndicator `json:"serviceLevelIndicator"`
}

// ServiceLevelIndicator defines the underlying indicator that is a Prometheus metric.
type ServiceLevelIndicator struct {
	// +optional
	HTTP *HTTPIndicator `json:"http,omitempty"`
	// +optional
	GRPC *GRPCIndicator `json:"grpc,omitempty"`

	//Custom *CustomIndicator `json:"custom"`
}

type HTTPIndicator struct {
	// Metric to use. Defaults to http_requests_total without latency or http_request_duration_seconds_bucket if latency is specified.
	// +optional
	Metric *string `json:"metric"`
	// Selectors are free form PromQL selectors for that specific service.
	// +optional
	Selectors []string `json:"selectors"`
	// ErrorSelectors are free form PromQL selectors that specify what time series should be uses as error counters.
	// Defaults to code=~"5.." selecting all HTTP 5xx responses as errors.
	// +optional
	ErrorSelectors []string `json:"errorSelectors"`
}

type GRPCIndicator struct {
	// Metric to use. Defaults to grpc_server_handled_total without latency or grpc_server_handling_seconds_bucket if latency is specified.
	// +optional
	Metric *string `json:"metric"`
	// Service is a selector to which gRPC service this indicator refers to.
	Service string `json:"service"`
	// Method is a selector to which gRPC service's method this indicator refers to.
	Method string `json:"method"`
	// Selectors are free form PromQL selectors for that specific service.
	Selectors []string `json:"selectors"`
}

type CustomIndicator struct {
	//// Type of the ServiceLevelIndicator. Can be HTTP and gRPC.
	//Type string `json:"type"`
	//// Metric to use.
	//Metric string `json:"metric"`
	//// Selectors for the metric to specify what time series are part of your service's indicator.
	//Selectors []string `json:"selectors"`
}

// Selector wraps the prometheus matchers.
type Selector string

// ServiceLevelObjectiveStatus defines the observed state of ServiceLevelObjective
type ServiceLevelObjectiveStatus struct{}
