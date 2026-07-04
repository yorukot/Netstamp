package publicstatus

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/trace"

	"github.com/yorukot/netstamp/internal/domain/identity"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	domainpublic "github.com/yorukot/netstamp/internal/domain/publicstatus"
)

type publicStatusFlow struct {
	service     *Service
	ctx         context.Context
	span        trace.Span
	action      PublicStatusAction
	actorUserID string
	projectID   string
	projectRef  string
	projectSlug string
	pageID      string
	elementID   string
}

func (s *Service) startPublicStatusFlow(ctx context.Context, spanName string, action PublicStatusAction, actorUserID string) (context.Context, *publicStatusFlow) {
	ctx, span := publicStatusTracer.Start(ctx, spanName, trace.WithAttributes(
		attrPublicStatusAction.String(string(action)),
	))
	if actorUserID != "" {
		span.SetAttributes(attrUserID.String(actorUserID))
	}

	return ctx, &publicStatusFlow{
		service:     s,
		ctx:         ctx,
		span:        span,
		action:      action,
		actorUserID: actorUserID,
	}
}

func (f *publicStatusFlow) end() {
	f.span.End()
}

func (f *publicStatusFlow) setProjectRef(projectRef string) {
	f.projectRef = projectRef
	if projectRef != "" {
		f.span.SetAttributes(attrProjectRef.String(projectRef))
	}
}

func (f *publicStatusFlow) setProject(project domainproject.Project) {
	f.projectID = project.ID
	if project.ID != "" {
		f.span.SetAttributes(attrProjectID.String(project.ID))
	}
	f.projectSlug = project.Slug
	if project.Slug != "" {
		f.span.SetAttributes(attrProjectSlug.String(project.Slug))
	}
}

func (f *publicStatusFlow) setPageID(pageID string) {
	f.pageID = pageID
	if pageID != "" {
		f.span.SetAttributes(attrPublicStatusPageID.String(pageID))
	}
}

func (f *publicStatusFlow) setElementID(elementID string) {
	f.elementID = elementID
	if elementID != "" {
		f.span.SetAttributes(attrPublicStatusElementID.String(elementID))
	}
}

func (f *publicStatusFlow) success() {
	f.span.SetAttributes(attrPublicStatusOutcome.String(string(PublicStatusOutcomeSuccess)))
	f.record(PublicStatusOutcomeSuccess, "", nil)
}

func (f *publicStatusFlow) businessFailure(reason PublicStatusReason, returnErr error) error {
	f.span.SetAttributes(
		attrPublicStatusOutcome.String(string(PublicStatusOutcomeFailure)),
		attrPublicStatusFailureReason.String(string(reason)),
	)
	f.record(PublicStatusOutcomeFailure, reason, nil)
	return returnErr
}

func (f *publicStatusFlow) technicalFailure(reason PublicStatusReason, err error) error {
	f.span.SetAttributes(attrPublicStatusOutcome.String(string(PublicStatusOutcomeFailure)))
	recordSpanError(f.span, err, reason)
	f.record(PublicStatusOutcomeFailure, reason, err)
	return err
}

func (f *publicStatusFlow) projectLookupFailure(err error) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound), errors.Is(err, domainproject.ErrMemberNotFound):
		return f.businessFailure(PublicStatusReasonProjectNotFound, err)
	case errors.Is(err, identity.ErrUserNotFound):
		return f.businessFailure(PublicStatusReasonUserNotFound, err)
	default:
		return f.technicalFailure(PublicStatusReasonProjectLookupFailed, err)
	}
}

func (f *publicStatusFlow) roleLookupFailure(err error) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound), errors.Is(err, domainproject.ErrMemberNotFound):
		return f.businessFailure(PublicStatusReasonProjectNotFound, err)
	case errors.Is(err, identity.ErrUserNotFound):
		return f.businessFailure(PublicStatusReasonUserNotFound, err)
	default:
		return f.technicalFailure(PublicStatusReasonRoleLookupFailed, err)
	}
}

func (f *publicStatusFlow) writeFailure(technicalReason PublicStatusReason, err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput), errors.Is(err, domainpublic.ErrInvalidInput):
		return f.businessFailure(PublicStatusReasonInvalidInput, err)
	case errors.Is(err, ErrForbidden):
		return f.businessFailure(PublicStatusReasonForbidden, err)
	case errors.Is(err, domainproject.ErrProjectNotFound), errors.Is(err, domainproject.ErrMemberNotFound):
		return f.businessFailure(PublicStatusReasonProjectNotFound, err)
	case errors.Is(err, identity.ErrUserNotFound):
		return f.businessFailure(PublicStatusReasonUserNotFound, err)
	case errors.Is(err, domainpublic.ErrPageNotFound):
		return f.businessFailure(PublicStatusReasonPageNotFound, err)
	case errors.Is(err, domainpublic.ErrElementNotFound):
		return f.businessFailure(PublicStatusReasonElementNotFound, err)
	case errors.Is(err, domainpublic.ErrSlugAlreadyExist):
		return f.businessFailure(PublicStatusReasonSlugAlreadyExists, err)
	default:
		return f.technicalFailure(technicalReason, err)
	}
}

func (f *publicStatusFlow) record(outcome PublicStatusOutcome, reason PublicStatusReason, err error) {
	if f.service.events == nil {
		return
	}
	f.service.events.RecordPublicStatusEvent(f.ctx, PublicStatusEvent{
		Name:        publicStatusEventName(f.action, outcome),
		Action:      f.action,
		Outcome:     outcome,
		Reason:      reason,
		ActorUserID: f.actorUserID,
		ProjectID:   f.projectID,
		ProjectRef:  f.projectRef,
		ProjectSlug: f.projectSlug,
		PageID:      f.pageID,
		ElementID:   f.elementID,
		Err:         err,
	})
}

func publicStatusEventName(action PublicStatusAction, outcome PublicStatusOutcome) PublicStatusEventName {
	switch action {
	case PublicStatusActionCreatePage:
		if outcome == PublicStatusOutcomeSuccess {
			return PublicStatusEventCreatePageSuccess
		}
		return PublicStatusEventCreatePageFailure
	case PublicStatusActionUpdatePage:
		if outcome == PublicStatusOutcomeSuccess {
			return PublicStatusEventUpdatePageSuccess
		}
		return PublicStatusEventUpdatePageFailure
	case PublicStatusActionDeletePage:
		if outcome == PublicStatusOutcomeSuccess {
			return PublicStatusEventDeletePageSuccess
		}
		return PublicStatusEventDeletePageFailure
	case PublicStatusActionCreateElement:
		if outcome == PublicStatusOutcomeSuccess {
			return PublicStatusEventCreateElementSuccess
		}
		return PublicStatusEventCreateElementFailure
	case PublicStatusActionUpdateElement:
		if outcome == PublicStatusOutcomeSuccess {
			return PublicStatusEventUpdateElementSuccess
		}
		return PublicStatusEventUpdateElementFailure
	case PublicStatusActionDeleteElement:
		if outcome == PublicStatusOutcomeSuccess {
			return PublicStatusEventDeleteElementSuccess
		}
		return PublicStatusEventDeleteElementFailure
	default:
		return PublicStatusEventName("public_status.unknown." + string(outcome))
	}
}
