package executor

import (
	"context"

	"github.com/yorukot/netstamp/internal/contracts/probecontrol"
)

type Executor interface {
	Execute(context.Context, probecontrol.Assignment) (probecontrol.Result, error)
}

type Registry struct {
	executors map[probecontrol.CheckType]Executor
}

func NewRegistry() *Registry {
	return &Registry{executors: map[probecontrol.CheckType]Executor{}}
}

func (r *Registry) Register(checkType probecontrol.CheckType, executor Executor) {
	r.executors[checkType] = executor
}

func (r *Registry) Get(checkType probecontrol.CheckType) (Executor, bool) {
	executor, ok := r.executors[checkType]
	return executor, ok
}
