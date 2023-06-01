package slo

import (
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func TestObjective_AbsentDuration(t *testing.T) {
	week, _ := model.ParseDuration("1w")
	month, _ := model.ParseDuration("4w")

	// 1w 1%
	require.Equal(t, model.Duration(10*time.Minute), Objective{Window: week, Target: 0.9}.AbsentDuration(1.0))
	require.Equal(t, model.Duration(5*time.Minute), Objective{Window: week, Target: 0.95}.AbsentDuration(1.0))
	require.Equal(t, model.Duration(time.Minute), Objective{Window: week, Target: 0.99}.AbsentDuration(1.0))
	require.Equal(t, model.Duration(time.Minute), Objective{Window: week, Target: 0.999}.AbsentDuration(1.0))
	// 1w 5%
	require.Equal(t, model.Duration(50*time.Minute), Objective{Window: week, Target: 0.9}.AbsentDuration(5.0))
	require.Equal(t, model.Duration(25*time.Minute), Objective{Window: week, Target: 0.95}.AbsentDuration(5.0))
	require.Equal(t, model.Duration(5*time.Minute), Objective{Window: week, Target: 0.99}.AbsentDuration(5.0))
	require.Equal(t, model.Duration(time.Minute), Objective{Window: week, Target: 0.999}.AbsentDuration(5.0))
	// 4w 1%
	require.Equal(t, model.Duration(40*time.Minute), Objective{Window: month, Target: 0.9}.AbsentDuration(1.0))
	require.Equal(t, model.Duration(20*time.Minute), Objective{Window: month, Target: 0.95}.AbsentDuration(1.0))
	require.Equal(t, model.Duration(4*time.Minute), Objective{Window: month, Target: 0.99}.AbsentDuration(1.0))
	require.Equal(t, model.Duration(time.Minute), Objective{Window: month, Target: 0.999}.AbsentDuration(1.0))
	// 4w 5%
	require.Equal(t, model.Duration(3*time.Hour+21*time.Minute), Objective{Window: month, Target: 0.9}.AbsentDuration(5.0))
	require.Equal(t, model.Duration(time.Hour+40*time.Minute), Objective{Window: month, Target: 0.95}.AbsentDuration(5.0))
	require.Equal(t, model.Duration(20*time.Minute), Objective{Window: month, Target: 0.99}.AbsentDuration(5.0))
	require.Equal(t, model.Duration(2*time.Minute), Objective{Window: month, Target: 0.999}.AbsentDuration(5.0))
}
