package ping

import (
	"context"

	"github.com/yorukot/netstamp/internal/contracts/probecontrol"
)

type Executor struct{}

func New() *Executor {
	return &Executor{}
}

func (e *Executor) Execute(ctx context.Context, assignment probecontrol.Assignment) (probecontrol.Result, error) {
	if err := ctx.Err(); err != nil {
		return probecontrol.Result{}, err
	}
	return probecontrol.Result{AssignmentID: assignment.ID, CheckID: assignment.CheckID, CheckVersion: assignment.CheckVersion, Type: assignment.Type}, nil
}
