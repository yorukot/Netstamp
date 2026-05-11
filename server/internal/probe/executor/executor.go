package executor

import (
	"context"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

type Executor interface {
	Execute(context.Context, domaincheck.Assignment) (domainprobe.Result, error)
}

type Registry struct {
	executors map[domaincheck.Type]Executor
}

func NewRegistry() *Registry {
	return &Registry{executors: map[domaincheck.Type]Executor{}}
}

func (r *Registry) Register(checkType domaincheck.Type, executor Executor) {
	r.executors[checkType] = executor
}

func (r *Registry) Get(checkType domaincheck.Type) (Executor, bool) {
	executor, ok := r.executors[checkType]
	return executor, ok
}
