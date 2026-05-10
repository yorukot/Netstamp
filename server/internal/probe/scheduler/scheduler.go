package scheduler

import (
	"context"

	"github.com/yorukot/netstamp/internal/contracts/probecontrol"
)

type Scheduler struct{}

func New() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) Next(ctx context.Context, assignments probecontrol.AssignmentSet) ([]probecontrol.Assignment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return assignments.Assignments, nil
}
