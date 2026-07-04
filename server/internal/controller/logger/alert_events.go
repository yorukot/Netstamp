package logger

import (
	"context"

	"go.uber.org/zap"

	appalert "github.com/yorukot/netstamp/internal/controller/application/alert"
)

type AlertEventRecorder struct {
	root *zap.Logger
}

func NewAlertEventRecorder(root *zap.Logger) *AlertEventRecorder {
	if root == nil {
		root = zap.NewNop()
	}

	return &AlertEventRecorder{root: root}
}

func (r *AlertEventRecorder) RecordAlertEvent(ctx context.Context, event appalert.AlertEvent) {
	recordApplicationEvent(ctx, r.root, applicationEventLog{
		name:            string(event.Name),
		category:        "alert",
		action:          string(event.Action),
		outcome:         string(event.Outcome),
		reason:          string(event.Reason),
		successOutcome:  string(appalert.AlertOutcomeSuccess),
		expectedFailure: isExpectedAlertFailure(event),
		fields:          alertEventFields(event),
		err:             event.Err,
	})
}

func alertEventFields(event appalert.AlertEvent) []zap.Field {
	fields := make([]zap.Field, 0, 6)
	fields = appendStringField(fields, "user.id", event.ActorUserID)
	fields = appendStringField(fields, "project.id", event.ProjectID)
	fields = appendStringField(fields, "project.ref", event.ProjectRef)
	fields = appendStringField(fields, "project.slug", event.ProjectSlug)
	fields = appendStringField(fields, "alert.rule.id", event.RuleID)
	fields = appendStringField(fields, "alert.notification.id", event.NotificationID)

	return fields
}

func isExpectedAlertFailure(event appalert.AlertEvent) bool {
	switch event.Reason {
	case appalert.AlertReasonInvalidInput,
		appalert.AlertReasonForbidden,
		appalert.AlertReasonProjectNotFound,
		appalert.AlertReasonUserNotFound,
		appalert.AlertReasonRuleNotFound,
		appalert.AlertReasonNotificationNotFound,
		appalert.AlertReasonIncidentNotFound,
		appalert.AlertReasonNotificationTesterUnavailable:
		return true
	default:
		return false
	}
}
