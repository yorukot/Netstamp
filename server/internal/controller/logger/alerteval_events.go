package logger

import (
	"context"

	"go.uber.org/zap"

	appalerteval "github.com/yorukot/netstamp/internal/controller/application/alerteval"
)

type AlertEvalEventRecorder struct {
	root *zap.Logger
}

func NewAlertEvalEventRecorder(root *zap.Logger) *AlertEvalEventRecorder {
	if root == nil {
		root = zap.NewNop()
	}

	return &AlertEvalEventRecorder{root: root}
}

func (r *AlertEvalEventRecorder) RecordAlertEvalEvent(ctx context.Context, event appalerteval.AlertEvalEvent) {
	recordApplicationEvent(ctx, r.root, applicationEventLog{
		name:           string(event.Name),
		category:       "alert_eval",
		action:         string(event.Action),
		outcome:        string(event.Outcome),
		reason:         string(event.Reason),
		successOutcome: string(appalerteval.AlertEvalOutcomeSuccess),
		fields:         alertEvalEventFields(event),
		err:            event.Err,
	})
}

func alertEvalEventFields(event appalerteval.AlertEvalEvent) []zap.Field {
	fields := make([]zap.Field, 0, 6)
	fields = appendStringField(fields, "project.id", event.ProjectID)
	fields = appendStringField(fields, "probe.id", event.ProbeID)
	fields = appendStringField(fields, "check.id", event.CheckID)
	fields = appendStringField(fields, "check.type", event.CheckType)
	fields = appendStringField(fields, "alert.rule.id", event.RuleID)
	fields = appendStringField(fields, "alert.incident.id", event.IncidentID)

	return fields
}
