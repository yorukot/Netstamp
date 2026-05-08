package logger

import (
	"context"

	"go.uber.org/zap"

	appcheck "github.com/yorukot/netstamp/internal/application/check"
)

type CheckEventRecorder struct {
	root *zap.Logger
}

func NewCheckEventRecorder(root *zap.Logger) *CheckEventRecorder {
	if root == nil {
		root = zap.NewNop()
	}

	return &CheckEventRecorder{root: root}
}

func (r *CheckEventRecorder) RecordCheckEvent(ctx context.Context, event appcheck.CheckEvent) {
	log := FromContext(ctx, r.root)
	fields := []zap.Field{
		zap.String("event_name", string(event.Name)),
		zap.String("event.category", "check"),
		zap.String("event.action", string(event.Action)),
		zap.String("event.outcome", string(event.Outcome)),
	}

	if event.Reason != "" {
		fields = append(fields, zap.String("event.reason", string(event.Reason)))
	}
	if event.ActorUserID != "" {
		fields = append(fields, zap.String("user.id", event.ActorUserID))
	}
	if event.ProjectID != "" {
		fields = append(fields, zap.String("project.id", event.ProjectID))
	}
	if event.ProjectRef != "" {
		fields = append(fields, zap.String("project.ref", event.ProjectRef))
	}
	if event.ProjectSlug != "" {
		fields = append(fields, zap.String("project.slug", event.ProjectSlug))
	}
	if event.CheckID != "" {
		fields = append(fields, zap.String("check.id", event.CheckID))
	}
	if event.Err != nil {
		fields = append(fields, zap.Error(event.Err))
	}

	switch {
	case event.Outcome == appcheck.CheckOutcomeSuccess:
		log.Info(string(event.Name), fields...)
	case isExpectedCheckFailure(event):
		log.Warn(string(event.Name), fields...)
	default:
		log.Error(string(event.Name), fields...)
	}
}

func isExpectedCheckFailure(event appcheck.CheckEvent) bool {
	switch event.Reason {
	case appcheck.CheckReasonInvalidInput,
		appcheck.CheckReasonForbidden,
		appcheck.CheckReasonProjectNotFound,
		appcheck.CheckReasonUserNotFound,
		appcheck.CheckReasonCheckNotFound,
		appcheck.CheckReasonLabelNotFound:
		return true
	default:
		return false
	}
}
