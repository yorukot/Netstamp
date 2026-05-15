package scheduling

import (
	"time"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

func (s *AssignmentStore) taskFromAssignment(assignment domainassignment.Assignment, pulledAt time.Time) (TaskState, bool) {
	if assignment.Check == nil {
		s.log.Warn("assignment skipped without check", "assignment_id", assignment.ID, "check_id", assignment.CheckID)
		return TaskState{}, false
	}
	if _, err := domaincheck.VNCheckType(assignment.Check.Type); err != nil {
		s.log.Warn("assignment skipped with unsupported check type", "assignment_id", assignment.ID, "check_id", assignment.CheckID, "check_type", assignment.Check.Type)
		return TaskState{}, false
	}
	// TODO: we need to modify it since basically we need to check the config base on each type
	if assignment.Check.PingConfig == nil {
		s.log.Warn("assignment skipped without ping config", "assignment_id", assignment.ID, "check_id", assignment.CheckID)
		return TaskState{}, false
	}

	interval := time.Duration(assignment.Check.IntervalSeconds) * time.Second
	if interval <= 0 {
		s.log.Warn("assignment skipped with invalid interval", "assignment_id", assignment.ID, "check_id", assignment.CheckID, "interval_seconds", assignment.Check.IntervalSeconds)
		return TaskState{}, false
	}

	phase := ComputePhase(s.probeID, assignment.ID, interval)
	nextDue := ComputeNextDue(pulledAt.UTC(), interval, phase)

	return TaskState{
		AssignmentID:    assignment.ID,

		CheckVersion:    assignment.CheckVersion,
		SelectorVersion: assignment.SelectorVersion,

		Check:           *assignment.Check,
		Probe:           *assignment.Probe,

		Phase:           phase,
		NextDue:         nextDue,
		LastPulledAt:    pulledAt.UTC(),
	}, true
}

// RunRequest returns a RunRequest for the task.
func (t TaskState) RunRequest(scheduledAt, createdAt time.Time) RunRequest {
	return RunRequest{
		AssignmentID:    t.AssignmentID,

		CheckVersion:    t.CheckVersion,
		SelectorVersion: t.SelectorVersion,

		Check:           t.Check,
		Probe:           t.Probe,
		
		ScheduledAt:     scheduledAt.UTC(),
		CreatedAt:       createdAt.UTC(),
	}
}
