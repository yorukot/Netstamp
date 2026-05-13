package scheduling

import (
	"hash/fnv"
	"log/slog"
	"reflect"
	"sync"
	"time"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainping "github.com/yorukot/netstamp/internal/domain/ping"
)

type TaskState struct {
	AssignmentID    string
	ProjectID       string
	ProbeID         string
	CheckID         string
	CheckType       domaincheck.Type
	CheckVersion    string
	SelectorVersion string
	Target          string
	Interval        time.Duration
	Phase           time.Duration
	NextDue         time.Time
	PingConfig      *domainping.Config
	Enabled         bool
	Removed         bool
	Generation      uint64
	LastPulledAt    time.Time
}

type RunRequest struct {
	AssignmentID    string
	ProjectID       string
	ProbeID         string
	CheckID         string
	CheckType       domaincheck.Type
	CheckVersion    string
	SelectorVersion string
	Target          string
	ScheduledAt     time.Time
	CreatedAt       time.Time
	PingConfig      *domainping.Config
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

func NewAssignmentStore(probeID string, ttl time.Duration, log *slog.Logger) *AssignmentStore {
	return &AssignmentStore{
		probeID: probeID,
		ttl:     ttl,
		tasks:   make(map[string]TaskState),
		log:     log,
	}
}

func (s *AssignmentStore) Reconcile(assignments []domainassignment.Assignment, pulledAt time.Time) ReconcileSummary {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastPull = pulledAt.UTC()
	seen := make(map[string]struct{}, len(assignments))
	summary := ReconcileSummary{}

	for _, assignment := range assignments {
		seen[assignment.ID] = struct{}{}

		nextTask, ok := s.taskFromAssignment(assignment, pulledAt)
		if !ok {
			summary.Unsupported++
			continue
		}

		current, exists := s.tasks[nextTask.AssignmentID]
		switch {
		case !exists:
			nextTask.Generation = 1
			s.tasks[nextTask.AssignmentID] = nextTask
			summary.Added++
		case current.Removed || taskChanged(current, nextTask):
			nextTask.Generation = current.Generation + 1
			s.tasks[nextTask.AssignmentID] = nextTask
			summary.Updated++
		default:
			current.LastPulledAt = pulledAt.UTC()
			current.Enabled = true
			s.tasks[current.AssignmentID] = current
		}
	}

	for id, current := range s.tasks {
		if _, ok := seen[id]; ok || current.Removed {
			continue
		}
		current.Removed = true
		current.Enabled = false
		current.Generation++
		s.tasks[id] = current
		summary.Removed++
	}

	for _, task := range s.tasks {
		if task.Enabled && !task.Removed {
			summary.Active++
		}
	}

	return summary
}

func (s *AssignmentStore) ActiveTasks() []TaskState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]TaskState, 0, len(s.tasks))
	for _, task := range s.tasks {
		if task.Enabled && !task.Removed {
			tasks = append(tasks, cloneTask(task))
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
	if !ok || task.Removed || !task.Enabled {
		return TaskState{}, false
	}
	if task.Generation != generation || !task.NextDue.Equal(due) {
		return TaskState{}, false
	}

	return cloneTask(task), true
}

func (s *AssignmentStore) AdvanceNextDue(assignmentID string, generation uint64, nextDue time.Time) (TaskState, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[assignmentID]
	if !ok || task.Removed || !task.Enabled || task.Generation != generation {
		return TaskState{}, false
	}
	task.NextDue = nextDue.UTC()
	s.tasks[assignmentID] = task

	return cloneTask(task), true
}

func (s *AssignmentStore) taskFromAssignment(assignment domainassignment.Assignment, pulledAt time.Time) (TaskState, bool) {
	if assignment.Check == nil {
		s.log.Warn("assignment skipped without check", "assignment_id", assignment.ID, "check_id", assignment.CheckID)
		return TaskState{}, false
	}
	if assignment.Check.Type != domaincheck.TypePing {
		s.log.Warn("assignment skipped with unsupported check type", "assignment_id", assignment.ID, "check_id", assignment.CheckID, "check_type", assignment.Check.Type)
		return TaskState{}, false
	}
	if assignment.Check.PingConfig == nil {
		s.log.Warn("assignment skipped without ping config", "assignment_id", assignment.ID, "check_id", assignment.CheckID)
		return TaskState{}, false
	}

	interval := time.Duration(assignment.Check.IntervalSeconds) * time.Second
	if interval <= 0 {
		s.log.Warn("assignment skipped with invalid interval", "assignment_id", assignment.ID, "check_id", assignment.CheckID, "interval_seconds", assignment.Check.IntervalSeconds)
		return TaskState{}, false
	}

	pingConfig := *assignment.Check.PingConfig
	phase := ComputePhase(s.probeID, assignment.ID, interval)
	nextDue := ComputeNextDue(pulledAt.UTC(), interval, phase)

	return TaskState{
		AssignmentID:    assignment.ID,
		ProjectID:       assignment.ProjectID,
		ProbeID:         assignment.ProbeID,
		CheckID:         assignment.CheckID,
		CheckType:       assignment.Check.Type,
		CheckVersion:    assignment.CheckVersion,
		SelectorVersion: assignment.SelectorVersion,
		Target:          assignment.Check.Target,
		Interval:        interval,
		Phase:           phase,
		NextDue:         nextDue,
		PingConfig:      &pingConfig,
		Enabled:         true,
		LastPulledAt:    pulledAt.UTC(),
	}, true
}

func taskChanged(current, next TaskState) bool {
	return current.CheckVersion != next.CheckVersion ||
		current.SelectorVersion != next.SelectorVersion ||
		current.CheckType != next.CheckType ||
		current.Target != next.Target ||
		current.Interval != next.Interval ||
		!reflect.DeepEqual(current.PingConfig, next.PingConfig)
}

func cloneTask(task TaskState) TaskState {
	if task.PingConfig != nil {
		pingConfig := *task.PingConfig
		task.PingConfig = &pingConfig
	}
	return task
}

func (t TaskState) RunRequest(scheduledAt, createdAt time.Time) RunRequest {
	var pingConfig *domainping.Config
	if t.PingConfig != nil {
		config := *t.PingConfig
		pingConfig = &config
	}

	return RunRequest{
		AssignmentID:    t.AssignmentID,
		ProjectID:       t.ProjectID,
		ProbeID:         t.ProbeID,
		CheckID:         t.CheckID,
		CheckType:       t.CheckType,
		CheckVersion:    t.CheckVersion,
		SelectorVersion: t.SelectorVersion,
		Target:          t.Target,
		ScheduledAt:     scheduledAt.UTC(),
		CreatedAt:       createdAt.UTC(),
		PingConfig:      pingConfig,
	}
}

func ComputePhase(probeID, assignmentID string, interval time.Duration) time.Duration {
	if interval <= time.Second {
		return 0
	}

	seconds := uint64(interval / time.Second)
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(probeID))
	_, _ = hash.Write([]byte(":"))
	_, _ = hash.Write([]byte(assignmentID))

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
