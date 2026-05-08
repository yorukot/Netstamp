package logger

import (
	"context"

	"go.uber.org/zap"

	appproject "github.com/yorukot/netstamp/internal/application/project"
)

type ProjectEventRecorder struct {
	root *zap.Logger
}

func NewProjectEventRecorder(root *zap.Logger) *ProjectEventRecorder {
	if root == nil {
		root = zap.NewNop()
	}

	return &ProjectEventRecorder{root: root}
}

func (r *ProjectEventRecorder) RecordProjectEvent(ctx context.Context, event appproject.ProjectEvent) {
	log := FromContext(ctx, r.root)
	fields := eventFields(string(event.Name), "project", string(event.Action), string(event.Outcome), string(event.Reason))

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
	if event.TargetUserID != "" {
		fields = append(fields, zap.String("project.member.user.id", event.TargetUserID))
	}
	if event.Role != "" {
		fields = append(fields, zap.String("project.member.role", string(event.Role)))
	}
	if event.Err != nil {
		fields = append(fields, zap.Error(event.Err))
	}

	switch {
	case event.Outcome == appproject.ProjectOutcomeSuccess:
		log.Info(string(event.Name), fields...)
	case isExpectedProjectFailure(event):
		log.Warn(string(event.Name), fields...)
	default:
		log.Error(string(event.Name), fields...)
	}
}

func isExpectedProjectFailure(event appproject.ProjectEvent) bool {
	switch event.Reason {
	case appproject.ProjectReasonInvalidInput,
		appproject.ProjectReasonInvalidRole,
		appproject.ProjectReasonForbidden,
		appproject.ProjectReasonProjectNotFound,
		appproject.ProjectReasonSlugAlreadyExists,
		appproject.ProjectReasonMemberAlreadyExists,
		appproject.ProjectReasonMemberNotFound,
		appproject.ProjectReasonUserNotFound,
		appproject.ProjectReasonLastOwner:
		return true
	default:
		return false
	}
}
