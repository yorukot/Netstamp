package scheduling

import (
	"hash/fnv"
	"log/slog"
	"sync"
	"time"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

type TaskState struct {
	AssignmentID string

	CheckVersion    string
	SelectorVersion string

	Check domaincheck.Check
	Probe domainprobe.Probe

	Phase        time.Duration
	Removed      bool
	NextDue      time.Time
	Generation   uint64
	LastPulledAt time.Time
}

type RunRequest struct {
	AssignmentID string

	CheckVersion    string
	SelectorVersion string

	Check domaincheck.Check
	Probe domainprobe.Probe

	ScheduledAt time.Time
	CreatedAt   time.Time
}

type ReconcileSummary struct {
	Added       int
	Updated     int
	Removed     int
	Unsupported int
	Active      int
}

type AssignmentStore struct {
	mu       sync.RWMutex
	probeID  string
	ttl      time.Duration
	tasks    map[string]TaskState
	lastPull time.Time
	log      *slog.Logger
}

// NewAssignmentStore creates a new AssignmentStore for the given probe ID, TTL, and logger.
// AssigmentStore is for storing and managing assignments for a probe. assignments are pull from the control plane from the pull endpoint.
func NewAssignmentStore(probeID string, ttl time.Duration, log *slog.Logger) *AssignmentStore {
	return &AssignmentStore{
		probeID: probeID,
		ttl:     ttl,
		tasks:   make(map[string]TaskState),
		log:     log,
	}
}

// Reconcile reconciles the assignment store with the given assignments, marking tasks as added, updated, or removed as necessary.
func (s *AssignmentStore) Reconcile(assignments []domainassignment.Assignment, pulledAt time.Time) ReconcileSummary {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastPull = pulledAt.UTC()
	seen := make(map[string]struct{}, len(assignments))
	summary := ReconcileSummary{}

	for _, assignment := range assignments {
		seen[assignment.ID] = struct{}{}

		// We convert the assignment to a task and check if it has changed
		nextTask, ok := s.taskFromAssignment(assignment, pulledAt)
		if !ok {
			summary.Unsupported++
			continue
		}

		// Get the assigments from the old tasks state
		current, exists := s.tasks[nextTask.AssignmentID]
		switch {
		case !exists:
			nextTask.Generation = 1
			s.tasks[nextTask.AssignmentID] = nextTask
			summary.Added++
		case current.Removed || taskChanged(current, nextTask):
			// update the generation so we can track how many times it has changed and the worker and schedular can detect stale tasks
			nextTask.Generation = current.Generation + 1
			s.tasks[nextTask.AssignmentID] = nextTask
			summary.Updated++
		default:
			// if the task exists and has not changed, mark it as pulled so the worker and scheduler can detect assignment that is too old
			current.LastPulledAt = pulledAt.UTC()
			s.tasks[current.AssignmentID] = current
		}
	}

	// remove any tasks that are no longer assigned
	for id, current := range s.tasks {
		if _, ok := seen[id]; ok || current.Removed {
			continue
		}
		current.Removed = true
		current.Generation++
		s.tasks[id] = current
		summary.Removed++
	}

	for _, task := range s.tasks {
		if !task.Removed {
			summary.Active++
		}
	}

	return summary
}

// ActiveTasks returns a list of active tasks (tasks that are not removed).
func (s *AssignmentStore) ActiveTasks() []TaskState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]TaskState, 0, len(s.tasks))
	for _, task := range s.tasks {
		if !task.Removed {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

func (s *AssignmentStore) IsFresh(now time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.lastPull.IsZero() {
		return false
	}

	return now.Sub(s.lastPull) <= s.ttl
}

func (s *AssignmentStore) CurrentForSchedule(assignmentID string, generation uint64, due time.Time) (TaskState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[assignmentID]
	if !ok || task.Removed {
		return TaskState{}, false
	}
	// Check if the task has the correct generation.
	// and if the task has the correct next due time. basically if the nextDue is not equal to the due time, which mean the task is already have next task.
	if task.Generation != generation || !task.NextDue.Equal(due) {
		return TaskState{}, false
	}

	return task, true
}

func (s *AssignmentStore) AdvanceNextDue(assignmentID string, generation uint64, nextDue time.Time) (TaskState, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[assignmentID]
	if !ok || task.Removed || task.Generation != generation {
		return TaskState{}, false
	}
	task.NextDue = nextDue.UTC()
	s.tasks[assignmentID] = task

	return task, true
}

// taskChanged returns true if the task has changed, false otherwise
func taskChanged(current, next TaskState) bool {
	return current.CheckVersion != next.CheckVersion ||
		current.SelectorVersion != next.SelectorVersion
}

// ComputePhase computes the phase for a task based on the probe ID, assignment ID, and interval.
// For example if the interval is 30s and the probe ID/assignment ID hash is 17s, the phase will be 17s.
// Assume we are at the 12:00:00 It will run at the 12:00:17 and 12:00:47 not at 12:00:00 and 12:00:30
func ComputePhase(probeID, assignmentID string, interval time.Duration) time.Duration {
	if interval <= time.Second {
		return 0
	}

	seconds := uint64(interval / time.Second)
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(probeID))
	_, _ = hash.Write([]byte(":"))
	_, _ = hash.Write([]byte(assignmentID))

	// Limit the number of seconds to 60 to long time wating
	// TODO: make this as a configurable limit
	// The phase the only is for make sure each probe not run the same check at the same time to make the server not overloaded
	// So we limit the number of seconds to 60 to avoid over waiting
	seconds = min(60, seconds)

	return time.Duration(hash.Sum64()%seconds) * time.Second
}

func ComputeNextDue(now time.Time, interval, phase time.Duration) time.Time {
	now = now.UTC()
	base := now.Truncate(interval)
	due := base.Add(phase)
	if !due.After(now) {
		due = due.Add(interval)
	}

	return due.UTC()
}

func ComputeNextFutureDue(previousDue, now time.Time, interval time.Duration) time.Time {
	next := previousDue.UTC().Add(interval)
	now = now.UTC()
	for !next.After(now) {
		next = next.Add(interval)
	}

	return next.UTC()
}

func IsTooLate(scheduledAt, now time.Time, interval time.Duration) bool {
	return now.UTC().Sub(scheduledAt.UTC()) > interval
}
