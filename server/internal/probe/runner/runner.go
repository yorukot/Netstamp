package runner

import (
	"context"
	"sort"
	"sync"
	"time"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	"github.com/yorukot/netstamp/internal/probe/executor"
)

type Runner struct {
	executors  *executor.Registry
	maxWorkers int
}

func New(executors *executor.Registry, maxWorkers int) *Runner {
	if maxWorkers <= 0 {
		maxWorkers = 1
	}

	return &Runner{
		executors:  executors,
		maxWorkers: maxWorkers,
	}
}

func (r *Runner) Run(ctx context.Context, assignments []domaincheck.Assignment) ([]domainprobe.Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(assignments) == 0 {
		return nil, nil
	}

	results := make([]domainprobe.Result, 0, len(assignments))
	resultCh := make(chan domainprobe.Result, len(assignments))
	sem := make(chan struct{}, r.maxWorkers)
	var wg sync.WaitGroup

	for _, assignment := range assignments {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case sem <- struct{}{}:
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			result := r.runOne(ctx, assignment)
			select {
			case resultCh <- result:
			case <-ctx.Done():
			}
		}()
	}

	wg.Wait()
	close(resultCh)

	for result := range resultCh {
		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].AssignmentID < results[j].AssignmentID
	})

	return results, ctx.Err()
}

func (r *Runner) runOne(ctx context.Context, assignment domaincheck.Assignment) domainprobe.Result {
	exec, ok := r.executors.Get(assignment.Type)
	if !ok {
		return unsupportedResult(assignment)
	}

	result, err := exec.Execute(ctx, assignment)
	if err != nil {
		if ctx.Err() != nil {
			return contextCanceledResult(assignment, err)
		}
		return executionErrorResult(assignment, err)
	}

	return result
}

func unsupportedResult(assignment domaincheck.Assignment) domainprobe.Result {
	return errorResult(assignment, "unsupported_check_type", "probe does not support assignment type")
}

func executionErrorResult(assignment domaincheck.Assignment, err error) domainprobe.Result {
	return errorResult(assignment, "execution_error", err.Error())
}

func contextCanceledResult(assignment domaincheck.Assignment, err error) domainprobe.Result {
	return errorResult(assignment, "context_canceled", err.Error())
}

func errorResult(assignment domaincheck.Assignment, code, message string) domainprobe.Result {
	now := time.Now().UTC()
	return domainprobe.Result{
		AssignmentID:    assignment.ID,
		CheckID:         assignment.CheckID,
		CheckVersion:    assignment.CheckVersion,
		SelectorVersion: assignment.SelectorVersion,
		Type:            assignment.Type,
		Ping: domainping.Result{
			StartedAt:    now,
			FinishedAt:   now,
			Status:       domainping.StatusError,
			ErrorCode:    &code,
			ErrorMessage: &message,
			Raw:          map[string]any{"assignmentType": string(assignment.Type)},
		},
	}
}
