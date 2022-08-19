package objectivesv1alpha1

import (
	"time"

	"github.com/pyrra-dev/pyrra/slo"
)

func FromInternal(o slo.Objective) *Objective {
	return &Objective{
		Labels:      o.Labels.Map(),
		Target:      o.Target,
		Window:      time.Duration(o.Window).Milliseconds(),
		Description: o.Description,
		Indicator:   nil,
	}
}
