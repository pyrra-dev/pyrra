package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/prometheus/prometheus/model/labels"
	pyrrav1alpha1 "github.com/pyrra-dev/pyrra/kubernetes/api/v1alpha1"
	openapiserver "github.com/pyrra-dev/pyrra/openapi/server/go"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var (
	o1 = pyrrav1alpha1.ServiceLevelObjective{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "objective-one",
			Namespace: "default",
		},
		Spec: pyrrav1alpha1.ServiceLevelObjectiveSpec{Target: "99"},
	}
	c1, _ = yaml.Marshal(o1)
	i1    = openapiserver.Objective{
		Labels: map[string]string{
			labels.MetricName: "objective-one",
			"namespace":       "default",
		},
		Target: 0.99,
		Config: string(c1),
	}
	o2 = pyrrav1alpha1.ServiceLevelObjective{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "objective-two",
			Namespace: "monitoring",
		},
		Spec: pyrrav1alpha1.ServiceLevelObjectiveSpec{Target: "98"},
	}
	c2, _ = yaml.Marshal(o2)
	i2    = openapiserver.Objective{
		Labels: map[string]string{
			labels.MetricName: "objective-two",
			"namespace":       "monitoring",
		},
		Target: 0.98,
		Config: string(c2),
	}
	o3 = pyrrav1alpha1.ServiceLevelObjective{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "objective-three",
			Namespace: "default",
		},
		Spec: pyrrav1alpha1.ServiceLevelObjectiveSpec{Target: "42.123"},
	}
	c3, _ = yaml.Marshal(o3)
	i3    = openapiserver.Objective{
		Labels: map[string]string{
			labels.MetricName: "objective-three",
			"namespace":       "default",
		},
		Target: 0.42123,
		Config: string(c3),
	}
)

type mockClient struct{}

func (m *mockClient) List(_ context.Context, list client.ObjectList, opts ...client.ListOption) error {
	switch l := list.(type) {
	case *pyrrav1alpha1.ServiceLevelObjectiveList:
		l.Items = append(l.Items, o1)
		l.Items = append(l.Items, o2)
		l.Items = append(l.Items, o3)
	}
	return nil
}

func TestObjectiveServer_ListObjectives(t *testing.T) {
	s := ObjectiveServer{client: &mockClient{}}

	testcases := []struct {
		name     string
		expr     string
		response []openapiserver.Objective
	}{{
		name:     "all",
		expr:     "",
		response: []openapiserver.Objective{i1, i2, i3},
	}, {
		name:     "nothing",
		expr:     `{__name__="bar"}`,
		response: []openapiserver.Objective{},
	}, {
		name:     "name",
		expr:     `{__name__="objective-two"}`,
		response: []openapiserver.Objective{i2},
	}, {
		name:     "nameRegex",
		expr:     `{__name__=~"objective-t.*"}`,
		response: []openapiserver.Objective{i2, i3},
	}, {
		name:     "namespace",
		expr:     `{namespace="default"}`,
		response: []openapiserver.Objective{i1, i3},
	}, {
		name:     "namespaceRegex",
		expr:     `{namespace=~"mon.*"}`,
		response: []openapiserver.Objective{i2},
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			response, err := s.ListObjectives(context.Background(), tc.expr)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.Code)
			require.Len(t, response.Body, len(tc.response))
			require.Equal(t, tc.response, response.Body)
		})
	}
}
