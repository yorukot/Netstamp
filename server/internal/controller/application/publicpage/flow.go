package publicpage

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/trace"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainpublicpage "github.com/yorukot/netstamp/internal/domain/publicpage"
)

type publicPageFlow struct {
	service     *Service
	ctx         context.Context
	span        trace.Span
	action      PublicPageEventAction
	actorUserID string
	projectID   string
	projectRef  string
	projectSlug string
	pageID      string
	pageSlug    string
	folderID    string
	checkID     string
	probeID     string
	checkCount  int
}

func (s *Service) startPublicPageFlow(ctx context.Context, spanName string, action PublicPageEventAction, actorUserID string) (context.Context, *publicPageFlow) {
	ctx, span := publicPageTracer.Start(ctx, spanName, trace.WithAttributes(
		attrPublicPageAction.String(string(action)),
	))
	if actorUserID != "" {
		span.SetAttributes(attrUserID.String(actorUserID))
	}

	return ctx, &publicPageFlow{
		service:     s,
		ctx:         ctx,
		span:        span,
		action:      action,
		actorUserID: actorUserID,
	}
}

func (f *publicPageFlow) end() {
	f.span.End()
}

func (f *publicPageFlow) setProjectRef(projectRef string) {
	f.projectRef = projectRef
	if projectRef != "" {
		f.span.SetAttributes(attrProjectRef.String(projectRef))
	}
}

func (f *publicPageFlow) setProject(project domainproject.Project) {
	f.projectID = project.ID
	if project.ID != "" {
		f.span.SetAttributes(attrProjectID.String(project.ID))
	}
	f.projectSlug = project.Slug
	if project.Slug != "" {
		f.span.SetAttributes(attrProjectSlug.String(project.Slug))
	}
}

func (f *publicPageFlow) setPage(page domainpublicpage.Page) {
	f.setPageID(page.ID)
	f.setPageSlug(page.Slug)
	if page.ProjectID != "" {
		f.projectID = page.ProjectID
		f.span.SetAttributes(attrProjectID.String(page.ProjectID))
	}
}

func (f *publicPageFlow) setPageID(pageID string) {
	f.pageID = pageID
	if pageID != "" {
		f.span.SetAttributes(attrPublicPageID.String(pageID))
	}
}

func (f *publicPageFlow) setPageSlug(slug string) {
	f.pageSlug = slug
	if slug != "" {
		f.span.SetAttributes(attrPublicPageSlug.String(slug))
	}
}

func (f *publicPageFlow) setFolderID(folderID string) {
	f.folderID = folderID
	if folderID != "" {
		f.span.SetAttributes(attrPublicPageFolderID.String(folderID))
	}
}

func (f *publicPageFlow) setCheckID(checkID string) {
	f.checkID = checkID
	if checkID != "" {
		f.span.SetAttributes(attrCheckID.String(checkID))
	}
}

func (f *publicPageFlow) setProbeID(probeID string) {
	f.probeID = probeID
	if probeID != "" {
		f.span.SetAttributes(attrProbeID.String(probeID))
	}
}

func (f *publicPageFlow) setCheckCount(count int) {
	f.checkCount = count
	f.span.SetAttributes(attrCheckCount.Int(count))
}

func (f *publicPageFlow) success(name PublicPageEventName) {
	f.span.SetAttributes(attrPublicPageOutcome.String(string(PublicPageOutcomeSuccess)))
	f.service.events.RecordPublicPageEvent(f.ctx, f.publicPageEvent(name, PublicPageOutcomeSuccess, "", nil))
}

func (f *publicPageFlow) businessFailure(name PublicPageEventName, reason PublicPageEventReason, returnErr error) error {
	f.span.SetAttributes(
		attrPublicPageOutcome.String(string(PublicPageOutcomeFailure)),
		attrPublicPageFailureReason.String(string(reason)),
	)
	f.service.events.RecordPublicPageEvent(f.ctx, f.publicPageEvent(name, PublicPageOutcomeFailure, reason, nil))
	return returnErr
}

func (f *publicPageFlow) technicalFailure(name PublicPageEventName, reason PublicPageEventReason, err error) error {
	f.span.SetAttributes(attrPublicPageOutcome.String(string(PublicPageOutcomeFailure)))
	recordSpanError(f.span, err, reason)
	f.service.events.RecordPublicPageEvent(f.ctx, f.publicPageEvent(name, PublicPageOutcomeFailure, reason, err))
	return err
}

func (f *publicPageFlow) projectLookupFailure(event PublicPageEventName, err error) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound), errors.Is(err, domainproject.ErrMemberNotFound):
		return f.businessFailure(event, PublicPageReasonProjectNotFound, err)
	case errors.Is(err, identity.ErrUserNotFound):
		return f.businessFailure(event, PublicPageReasonUserNotFound, err)
	default:
		return f.technicalFailure(event, PublicPageReasonProjectLookupFailed, err)
	}
}

func (f *publicPageFlow) roleLookupFailure(event PublicPageEventName, err error) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound), errors.Is(err, domainproject.ErrMemberNotFound):
		return f.businessFailure(event, PublicPageReasonProjectNotFound, err)
	case errors.Is(err, identity.ErrUserNotFound):
		return f.businessFailure(event, PublicPageReasonUserNotFound, err)
	default:
		return f.technicalFailure(event, PublicPageReasonRoleLookupFailed, err)
	}
}

func (f *publicPageFlow) pageLookupFailure(event PublicPageEventName, err error) error {
	switch {
	case errors.Is(err, domainpublicpage.ErrPageNotFound):
		return f.businessFailure(event, PublicPageReasonPageNotFound, err)
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return f.businessFailure(event, PublicPageReasonProjectNotFound, err)
	case errors.Is(err, ErrInvalidInput), errors.Is(err, domainpublicpage.ErrInvalidInput):
		return f.businessFailure(event, PublicPageReasonInvalidInput, err)
	default:
		return f.technicalFailure(event, PublicPageReasonPageLookupFailed, err)
	}
}

func (f *publicPageFlow) folderLookupFailure(event PublicPageEventName, err error) error {
	switch {
	case errors.Is(err, domainpublicpage.ErrFolderNotFound):
		return f.businessFailure(event, PublicPageReasonFolderNotFound, err)
	case errors.Is(err, domainpublicpage.ErrPageNotFound):
		return f.businessFailure(event, PublicPageReasonPageNotFound, err)
	case errors.Is(err, ErrInvalidInput), errors.Is(err, domainpublicpage.ErrInvalidInput):
		return f.businessFailure(event, PublicPageReasonInvalidInput, err)
	default:
		return f.technicalFailure(event, PublicPageReasonFolderListFailed, err)
	}
}

func (f *publicPageFlow) writeFailure(event PublicPageEventName, technicalReason PublicPageEventReason, err error) error {
	switch {
	case errors.Is(err, domainpublicpage.ErrDuplicateSlug):
		return f.businessFailure(event, PublicPageReasonDuplicateSlug, err)
	case errors.Is(err, domainpublicpage.ErrCheckAlreadyPublished):
		return f.businessFailure(event, PublicPageReasonCheckAlreadyPublished, err)
	case errors.Is(err, domainpublicpage.ErrCheckNotPublished):
		return f.businessFailure(event, PublicPageReasonCheckNotPublished, err)
	case errors.Is(err, domainpublicpage.ErrPageNotFound):
		return f.businessFailure(event, PublicPageReasonPageNotFound, err)
	case errors.Is(err, domainpublicpage.ErrFolderNotFound):
		return f.businessFailure(event, PublicPageReasonFolderNotFound, err)
	case errors.Is(err, domaincheck.ErrCheckNotFound):
		return f.businessFailure(event, PublicPageReasonCheckNotFound, err)
	case errors.Is(err, domainprobe.ErrProbeNotFound):
		return f.businessFailure(event, PublicPageReasonProbeNotFound, err)
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return f.businessFailure(event, PublicPageReasonProjectNotFound, err)
	case errors.Is(err, ErrInvalidInput), errors.Is(err, domainpublicpage.ErrInvalidInput), errors.Is(err, domaincheck.ErrInvalidInput), errors.Is(err, domainprobe.ErrInvalidInput):
		return f.businessFailure(event, PublicPageReasonInvalidInput, err)
	default:
		return f.technicalFailure(event, technicalReason, err)
	}
}

func (f *publicPageFlow) publicPairLookupFailure(err error) error {
	switch {
	case errors.Is(err, domainpublicpage.ErrCheckNotPublished):
		return f.businessFailure(PublicPageEventPingInsightFailure, PublicPageReasonCheckNotPublished, err)
	case errors.Is(err, domainpublicpage.ErrPageNotFound):
		return f.businessFailure(PublicPageEventPingInsightFailure, PublicPageReasonPageNotFound, err)
	case errors.Is(err, domaincheck.ErrCheckNotFound):
		return f.businessFailure(PublicPageEventPingInsightFailure, PublicPageReasonCheckNotFound, err)
	case errors.Is(err, domainprobe.ErrProbeNotFound):
		return f.businessFailure(PublicPageEventPingInsightFailure, PublicPageReasonProbeNotFound, err)
	default:
		return f.technicalFailure(PublicPageEventPingInsightFailure, PublicPageReasonPublicPairLookupFailed, err)
	}
}

func (f *publicPageFlow) publicPageEvent(name PublicPageEventName, outcome PublicPageEventOutcome, reason PublicPageEventReason, err error) PublicPageEvent {
	return PublicPageEvent{
		Name:        name,
		Action:      f.action,
		Outcome:     outcome,
		Reason:      reason,
		ActorUserID: f.actorUserID,
		ProjectID:   f.projectID,
		ProjectRef:  f.projectRef,
		ProjectSlug: f.projectSlug,
		PageID:      f.pageID,
		PageSlug:    f.pageSlug,
		FolderID:    f.folderID,
		CheckID:     f.checkID,
		ProbeID:     f.probeID,
		CheckCount:  f.checkCount,
		Err:         err,
	}
}
