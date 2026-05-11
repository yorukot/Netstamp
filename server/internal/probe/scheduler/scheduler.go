package scheduler

import (
	"context"
	"sync"
	"time"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

const defaultInterval = 30 * time.Second

type Scheduler struct {
	mu          sync.Mutex
	assignments map[string]assignmentState
}

type assignmentState struct {
	assignment domaincheck.Assignment
	nextRun    time.Time
	running    bool
}

func New() *Scheduler {
	return &Scheduler{assignments: map[string]assignmentState{}}
}

func (s *Scheduler) Replace(assignments []domaincheck.Assignment, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := make(map[string]assignmentState, len(assignments))
	for _, assignment := range assignments {
		if assignment.ID == "" {
			continue
		}
		current, ok := s.assignments[assignment.ID]
		if !ok {
			next[assignment.ID] = assignmentState{
				assignment: assignment,
				nextRun:    now,
			}
			continue
		}
		current.assignment = assignment
		next[assignment.ID] = current
	}
	s.assignments = next
}

func (s *Scheduler) Due(ctx context.Context, now time.Time) ([]domaincheck.Assignment, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	due := make([]domaincheck.Assignment, 0)
	for id, state := range s.assignments {
		if state.running || state.nextRun.After(now) {
			continue
		}
		state.running = true
		s.assignments[id] = state
		due = append(due, state.assignment)
	}

	return due, nil
}

func (s *Scheduler) Complete(assignments []domaincheck.Assignment, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, assignment := range assignments {
		state, ok := s.assignments[assignment.ID]
		if !ok {
			continue
		}
		state.running = false
		state.nextRun = now.Add(interval(assignment.IntervalSeconds))
		s.assignments[assignment.ID] = state
	}
}

func interval(seconds int32) time.Duration {
	if seconds <= 0 {
		return defaultInterval
	}

	return time.Duration(seconds) * time.Second
}
