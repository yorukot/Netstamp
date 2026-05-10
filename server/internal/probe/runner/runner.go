package runner

import (
	"context"

	"github.com/yorukot/netstamp/internal/contracts/probecontrol"
	"github.com/yorukot/netstamp/internal/probe/executor"
)

type Runner struct {
	executors *executor.Registry
}

func New(executors *executor.Registry) *Runner {
	return &Runner{executors: executors}
}

func (r *Runner) Run(ctx context.Context, assignments []probecontrol.Assignment) ([]probecontrol.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	_ = r.executors
	results := make([]probecontrol.Result, 0, len(assignments))
	return results, nil
}
