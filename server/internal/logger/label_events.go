package logger

import (
	"context"

	"go.uber.org/zap"

	applabel "github.com/yorukot/netstamp/internal/application/label"
)

type LabelEventRecorder struct {
	root *zap.Logger
}

func NewLabelEventRecorder(root *zap.Logger) *LabelEventRecorder {
	if root == nil {
		root = zap.NewNop()
	}

	return &LabelEventRecorder{root: root}
}

func (r *LabelEventRecorder) RecordLabelEvent(ctx context.Context, event applabel.LabelEvent) {
	log := FromContext(ctx, r.root)
	fields := eventFields(string(event.Name), "label", string(event.Action), string(event.Outcome), string(event.Reason))

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
	if event.LabelID != "" {
		fields = append(fields, zap.String("label.id", event.LabelID))
	}
	if event.Err != nil {
		fields = append(fields, zap.Error(event.Err))
	}

	switch {
	case event.Outcome == applabel.LabelOutcomeSuccess:
		log.Info(string(event.Name), fields...)
	case isExpectedLabelFailure(event):
		log.Warn(string(event.Name), fields...)
	default:
		log.Error(string(event.Name), fields...)
	}
}

func isExpectedLabelFailure(event applabel.LabelEvent) bool {
	switch event.Reason {
	case applabel.LabelReasonInvalidInput,
		applabel.LabelReasonForbidden,
		applabel.LabelReasonProjectNotFound,
		applabel.LabelReasonUserNotFound,
		applabel.LabelReasonLabelNotFound,
		applabel.LabelReasonLabelAlreadyExists:
		return true
	default:
		return false
	}
}
