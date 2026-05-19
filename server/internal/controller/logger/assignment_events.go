package logger

import (
	"context"

	"go.uber.org/zap"

	appassignment "github.com/yorukot/netstamp/internal/controller/application/assignment"
)

type AssignmentEventRecorder struct {
	root *zap.Logger
}

func NewAssignmentEventRecorder(root *zap.Logger) *AssignmentEventRecorder {
	if root == nil {
		root = zap.NewNop()
	}

	return &AssignmentEventRecorder{root: root}
}

func (r *AssignmentEventRecorder) RecordAssignmentEvent(ctx context.Context, event appassignment.AssignmentEvent) {
	recordApplicationEvent(ctx, r.root, applicationEventLog{
		name:            string(event.Name),
		category:        "assignment",
		action:          string(event.Action),
		outcome:         string(event.Outcome),
		reason:          string(event.Reason),
		successOutcome:  string(appassignment.AssignmentOutcomeSuccess),
		expectedFailure: isExpectedAssignmentFailure(event),
		fields:          assignmentEventFields(event),
		err:             event.Err,
	})
}

func assignmentEventFields(event appassignment.AssignmentEvent) []zap.Field {
	fields := make([]zap.Field, 0, 4)
	fields = appendStringField(fields, "project.id", event.ProjectID)
	fields = appendStringField(fields, "probe.id", event.ProbeID)
	fields = appendStringField(fields, "check.id", event.CheckID)
	fields = appendStringField(fields, "label.id", event.LabelID)

	return fields
}

func isExpectedAssignmentFailure(event appassignment.AssignmentEvent) bool {
	switch event.Reason {
	case appassignment.AssignmentReasonInvalidInput,
		appassignment.AssignmentReasonProjectNotFound,
		appassignment.AssignmentReasonProbeNotFound,
		appassignment.AssignmentReasonCheckNotFound,
		appassignment.AssignmentReasonLabelNotFound,
		appassignment.AssignmentReasonForbidden:
		return true
	default:
		return false
	}
}
