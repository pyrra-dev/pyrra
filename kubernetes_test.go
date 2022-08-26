package main

import (
	"context"
	"testing"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	pyrrav1alpha1 "github.com/pyrra-dev/pyrra/kubernetes/api/v1alpha1"
	objectivesv1alpha1 "github.com/pyrra-dev/pyrra/proto/objectives/v1alpha1"
)

var (
	o1 = pyrrav1alpha1.ServiceLevelObjective{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "objective-one",
			Namespace: "default",
		},
		Spec: pyrrav1alpha1.ServiceLevelObjectiveSpec{Target: "99", Window: "1w"},
	}
	c1, _ = yaml.Marshal(o1)
	i1    = &objectivesv1alpha1.Objective{
		Labels: map[string]string{
			labels.MetricName: "objective-one",
			"namespace":       "default",
		},
		Target: 0.99,
		Window: durationpb.New(1 * 7 * 24 * time.Hour),
		Config: string(c1),
	}
	o2 = pyrrav1alpha1.ServiceLevelObjective{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "objective-two",
			Namespace: "monitoring",
		},
		Spec: pyrrav1alpha1.ServiceLevelObjectiveSpec{Target: "98", Window: "2w"},
	}
	c2, _ = yaml.Marshal(o2)
	i2    = &objectivesv1alpha1.Objective{
		Labels: map[string]string{
			labels.MetricName: "objective-two",
			"namespace":       "monitoring",
		},
		Target: 0.98,
		Window: durationpb.New(2 * 7 * 24 * time.Hour),
		Config: string(c2),
	}
	o3 = pyrrav1alpha1.ServiceLevelObjective{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "objective-three",
			Namespace: "default",
		},
		Spec: pyrrav1alpha1.ServiceLevelObjectiveSpec{Target: "42.123", Window: "3w"},
	}
	c3, _ = yaml.Marshal(o3)
	i3    = &objectivesv1alpha1.Objective{
		Labels: map[string]string{
			labels.MetricName: "objective-three",
			"namespace":       "default",
		},
		Target: 0.42123,
		Window: durationpb.New(3 * 7 * 24 * time.Hour),
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
	s := KubernetesObjectiveServer{client: &mockClient{}}

	testcases := []struct {
		name     string
		expr     string
		response []*objectivesv1alpha1.Objective
	}{{
		name:     "all",
		expr:     "",
		response: []*objectivesv1alpha1.Objective{i1, i2, i3},
	}, {
		name:     "nothing",
		expr:     `{__name__="bar"}`,
		response: []*objectivesv1alpha1.Objective{},
	}, {
		name:     "name",
		expr:     `{__name__="objective-two"}`,
		response: []*objectivesv1alpha1.Objective{i2},
	}, {
		name:     "nameRegex",
		expr:     `{__name__=~"objective-t.*"}`,
		response: []*objectivesv1alpha1.Objective{i2, i3},
	}, {
		name:     "namespace",
		expr:     `{namespace="default"}`,
		response: []*objectivesv1alpha1.Objective{i1, i3},
	}, {
		name:     "namespaceRegex",
		expr:     `{namespace=~"mon.*"}`,
		response: []*objectivesv1alpha1.Objective{i2},
	}}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			response, err := s.List(context.Background(), connect.NewRequest(&objectivesv1alpha1.ListRequest{
				Expr: tc.expr,
			}))
			require.NoError(t, err)
			require.Len(t, response.Msg.Objectives, len(tc.response))
			require.Equal(t, tc.response, response.Msg.Objectives)
		})
	}
}
