package logger

import (
	"context"

	"go.uber.org/zap"

	apppublicpage "github.com/yorukot/netstamp/internal/controller/application/publicpage"
)

type PublicPageEventRecorder struct {
	root *zap.Logger
}

func NewPublicPageEventRecorder(root *zap.Logger) *PublicPageEventRecorder {
	if root == nil {
		root = zap.NewNop()
	}

	return &PublicPageEventRecorder{root: root}
}

func (r *PublicPageEventRecorder) RecordPublicPageEvent(ctx context.Context, event apppublicpage.PublicPageEvent) {
	fields := []zap.Field{}

	fields = appendStringField(fields, "user.id", event.ActorUserID)
	fields = appendStringField(fields, "project.id", event.ProjectID)
	fields = appendStringField(fields, "project.ref", event.ProjectRef)
	fields = appendStringField(fields, "project.slug", event.ProjectSlug)
	fields = appendStringField(fields, "public_page.id", event.PageID)
	fields = appendStringField(fields, "public_page.slug", event.PageSlug)
	fields = appendStringField(fields, "public_page.folder.id", event.FolderID)
	fields = appendStringField(fields, "check.id", event.CheckID)
	fields = appendStringField(fields, "probe.id", event.ProbeID)
	if event.CheckCount > 0 {
		fields = append(fields, zap.Int("check.count", event.CheckCount))
	}

	recordApplicationEvent(ctx, r.root, applicationEventLog{
		name:            string(event.Name),
		category:        "public_page",
		action:          string(event.Action),
		outcome:         string(event.Outcome),
		reason:          string(event.Reason),
		successOutcome:  string(apppublicpage.PublicPageOutcomeSuccess),
		expectedFailure: isExpectedPublicPageFailure(event),
		fields:          fields,
		err:             event.Err,
	})
}

func isExpectedPublicPageFailure(event apppublicpage.PublicPageEvent) bool {
	switch event.Reason {
	case apppublicpage.PublicPageReasonInvalidInput,
		apppublicpage.PublicPageReasonForbidden,
		apppublicpage.PublicPageReasonProjectNotFound,
		apppublicpage.PublicPageReasonUserNotFound,
		apppublicpage.PublicPageReasonPageNotFound,
		apppublicpage.PublicPageReasonFolderNotFound,
		apppublicpage.PublicPageReasonCheckNotFound,
		apppublicpage.PublicPageReasonProbeNotFound,
		apppublicpage.PublicPageReasonCheckNotPublished,
		apppublicpage.PublicPageReasonDuplicateSlug,
		apppublicpage.PublicPageReasonCheckAlreadyPublished:
		return true
	default:
		return false
	}
}
