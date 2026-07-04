package logger

import (
	"context"

	"go.uber.org/zap"

	apppublicstatus "github.com/yorukot/netstamp/internal/controller/application/publicstatus"
)

type PublicStatusEventRecorder struct {
	root *zap.Logger
}

func NewPublicStatusEventRecorder(root *zap.Logger) *PublicStatusEventRecorder {
	if root == nil {
		root = zap.NewNop()
	}

	return &PublicStatusEventRecorder{root: root}
}

func (r *PublicStatusEventRecorder) RecordPublicStatusEvent(ctx context.Context, event apppublicstatus.PublicStatusEvent) {
	recordApplicationEvent(ctx, r.root, applicationEventLog{
		name:            string(event.Name),
		category:        "public_status",
		action:          string(event.Action),
		outcome:         string(event.Outcome),
		reason:          string(event.Reason),
		successOutcome:  string(apppublicstatus.PublicStatusOutcomeSuccess),
		expectedFailure: isExpectedPublicStatusFailure(event),
		fields:          publicStatusEventFields(event),
		err:             event.Err,
	})
}

func publicStatusEventFields(event apppublicstatus.PublicStatusEvent) []zap.Field {
	fields := make([]zap.Field, 0, 6)
	fields = appendStringField(fields, "user.id", event.ActorUserID)
	fields = appendStringField(fields, "project.id", event.ProjectID)
	fields = appendStringField(fields, "project.ref", event.ProjectRef)
	fields = appendStringField(fields, "project.slug", event.ProjectSlug)
	fields = appendStringField(fields, "public_status.page.id", event.PageID)
	fields = appendStringField(fields, "public_status.element.id", event.ElementID)

	return fields
}

func isExpectedPublicStatusFailure(event apppublicstatus.PublicStatusEvent) bool {
	switch event.Reason {
	case apppublicstatus.PublicStatusReasonInvalidInput,
		apppublicstatus.PublicStatusReasonForbidden,
		apppublicstatus.PublicStatusReasonProjectNotFound,
		apppublicstatus.PublicStatusReasonUserNotFound,
		apppublicstatus.PublicStatusReasonPageNotFound,
		apppublicstatus.PublicStatusReasonElementNotFound,
		apppublicstatus.PublicStatusReasonSlugAlreadyExists:
		return true
	default:
		return false
	}
}
