package logger

import (
	"context"

	"go.uber.org/zap"

	appprobe "github.com/yorukot/netstamp/internal/controller/application/probe"
)

type ProbeEventRecorder struct {
	root *zap.Logger
}

func NewProbeEventRecorder(root *zap.Logger) *ProbeEventRecorder {
	if root == nil {
		root = zap.NewNop()
	}

	return &ProbeEventRecorder{root: root}
}

func (r *ProbeEventRecorder) RecordProbeEvent(ctx context.Context, event appprobe.ProbeEvent) {
	recordApplicationEvent(ctx, r.root, applicationEventLog{
		name:            string(event.Name),
		category:        "probe",
		action:          string(event.Action),
		outcome:         string(event.Outcome),
		reason:          string(event.Reason),
		successOutcome:  string(appprobe.ProbeOutcomeSuccess),
		expectedFailure: isExpectedProbeFailure(event),
		fields:          probeEventFields(event),
		err:             event.Err,
	})
}

func probeEventFields(event appprobe.ProbeEvent) []zap.Field {
	fields := make([]zap.Field, 0, 4)
	fields = appendStringField(fields, "user.id", event.ActorUserID)
	fields = appendStringField(fields, "project.id", event.ProjectID)
	fields = appendStringField(fields, "project.ref", event.ProjectRef)
	fields = appendStringField(fields, "probe.id", event.ProbeID)

	return fields
}

func isExpectedProbeFailure(event appprobe.ProbeEvent) bool {
	switch event.Reason {
	case appprobe.ProbeReasonInvalidInput,
		appprobe.ProbeReasonForbidden,
		appprobe.ProbeReasonProjectNotFound,
		appprobe.ProbeReasonProbeNotFound,
		appprobe.ProbeReasonLabelNotFound:
		return true
	default:
		return false
	}
}
