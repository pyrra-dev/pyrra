package slo

import (
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func TestObjective_AbsentDuration(t *testing.T) {
	week, _ := model.ParseDuration("1w")

	require.Equal(t, model.Duration(10*time.Minute), Objective{Window: week, Target: 0.9}.AbsentDuration())
	require.Equal(t, model.Duration(5*time.Minute), Objective{Window: week, Target: 0.95}.AbsentDuration())
	require.Equal(t, model.Duration(time.Minute), Objective{Window: week, Target: 0.99}.AbsentDuration())
	require.Equal(t, model.Duration(time.Minute), Objective{Window: week, Target: 0.999}.AbsentDuration())

	month, _ := model.ParseDuration("4w")
	require.Equal(t, model.Duration(40*time.Minute), Objective{Window: month, Target: 0.9}.AbsentDuration())
	require.Equal(t, model.Duration(20*time.Minute), Objective{Window: month, Target: 0.95}.AbsentDuration())
	require.Equal(t, model.Duration(4*time.Minute), Objective{Window: month, Target: 0.99}.AbsentDuration())
	require.Equal(t, model.Duration(time.Minute), Objective{Window: month, Target: 0.999}.AbsentDuration())
}
