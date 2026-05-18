package scheduling

import (
	"time"

	domainassignment "github.com/yorukot/netstamp/internal/domain/assignment"
	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
)

func (s *AssignmentStore) taskFromAssignment(assignment domainassignment.Assignment, pulledAt time.Time) (TaskState, bool) {
	if assignment.Check == nil {
		s.log.Warn("assignment skipped without check", "assignment_id", assignment.ID)
		return TaskState{}, false
	}
	if !isRunnableCheckType(assignment.Check.Type) {
		s.log.Warn("assignment skipped with unsupported check type", "assignment_id", assignment.ID, "check_id", assignment.Check.ID, "check_type", assignment.Check.Type)
		return TaskState{}, false
	}
	if !hasConfigForCheckType(*assignment.Check) {
		s.log.Warn("assignment skipped without check config", "assignment_id", assignment.ID, "check_id", assignment.Check.ID, "check_type", assignment.Check.Type)
		return TaskState{}, false
	}

	interval := time.Duration(assignment.Check.IntervalSeconds) * time.Second
	if interval <= 0 {
		s.log.Warn("assignment skipped with invalid interval", "assignment_id", assignment.ID, "check_id", assignment.Check.ID, "interval_seconds", assignment.Check.IntervalSeconds)
		return TaskState{}, false
	}

	phase := ComputePhase(s.probeID, assignment.ID, interval)
	nextDue := ComputeNextDue(pulledAt.UTC(), interval, phase)

	return TaskState{
		AssignmentID: assignment.ID,

		CheckVersion:    assignment.CheckVersion,
		SelectorVersion: assignment.SelectorVersion,

		Check: *assignment.Check,

		Phase:        phase,
		NextDue:      nextDue,
		LastPulledAt: pulledAt.UTC(),
	}, true
}

func isRunnableCheckType(checkType domaincheck.Type) bool {
	return checkType == domaincheck.TypePing || checkType == domaincheck.TypeTraceroute
}

func hasConfigForCheckType(check domaincheck.Check) bool {
	switch check.Type {
	case domaincheck.TypePing:
		return check.PingConfig != nil
	case domaincheck.TypeTraceroute:
		return check.TracerouteConfig != nil
	default:
		return false
	}
}

// RunRequest returns a RunRequest for the task.
func (t TaskState) RunRequest(scheduledAt, createdAt time.Time) RunRequest {
	return RunRequest{
		AssignmentID: t.AssignmentID,

		CheckVersion:    t.CheckVersion,
		SelectorVersion: t.SelectorVersion,

		Check: t.Check,

		ScheduledAt: scheduledAt.UTC(),
		CreatedAt:   createdAt.UTC(),
	}
}
