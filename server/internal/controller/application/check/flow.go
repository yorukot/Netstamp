package check

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/trace"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	"github.com/yorukot/netstamp/internal/domain/identity"
	"github.com/yorukot/netstamp/internal/domain/label"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type checkFlow struct {
	service     *Service
	ctx         context.Context
	span        trace.Span
	action      CheckEventAction
	actorUserID string
	projectID   string
	projectRef  string
	projectSlug string
	checkID     string
}

func (s *Service) startCheckFlow(ctx context.Context, spanName string, action CheckEventAction, actorUserID string) (context.Context, *checkFlow) {
	ctx, span := checkTracer.Start(ctx, spanName, trace.WithAttributes(
		attrCheckAction.String(string(action)),
	))
	if actorUserID != "" {
		span.SetAttributes(attrUserID.String(actorUserID))
	}

	return ctx, &checkFlow{
		service:     s,
		ctx:         ctx,
		span:        span,
		action:      action,
		actorUserID: actorUserID,
	}
}

func (f *checkFlow) end() {
	f.span.End()
}

func (f *checkFlow) setProjectRef(projectRef string) {
	f.projectRef = projectRef
	if projectRef != "" {
		f.span.SetAttributes(attrProjectRef.String(projectRef))
	}
}

func (f *checkFlow) setProject(project domainproject.Project) {
	f.projectID = project.ID
	if project.ID != "" {
		f.span.SetAttributes(attrProjectID.String(project.ID))
	}
	f.projectSlug = project.Slug
	if project.Slug != "" {
		f.span.SetAttributes(attrProjectSlug.String(project.Slug))
	}
}

func (f *checkFlow) setCheckID(checkID string) {
	f.checkID = checkID
	if checkID != "" {
		f.span.SetAttributes(attrCheckID.String(checkID))
	}
}

func (f *checkFlow) success(name CheckEventName) {
	f.span.SetAttributes(attrCheckOutcome.String(string(CheckOutcomeSuccess)))
	f.service.events.RecordCheckEvent(f.ctx, f.checkEvent(name, CheckOutcomeSuccess, "", nil))
}

func (f *checkFlow) businessFailure(name CheckEventName, reason CheckEventReason, returnErr error) error {
	f.span.SetAttributes(
		attrCheckOutcome.String(string(CheckOutcomeFailure)),
		attrCheckFailureReason.String(string(reason)),
	)
	f.service.events.RecordCheckEvent(f.ctx, f.checkEvent(name, CheckOutcomeFailure, reason, nil))
	return returnErr
}

func (f *checkFlow) technicalFailure(name CheckEventName, reason CheckEventReason, err error) error {
	f.span.SetAttributes(attrCheckOutcome.String(string(CheckOutcomeFailure)))
	recordSpanError(f.span, err, reason)
	f.service.events.RecordCheckEvent(f.ctx, f.checkEvent(name, CheckOutcomeFailure, reason, err))
	return err
}

func (f *checkFlow) projectLookupFailure(event CheckEventName, err error) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound), errors.Is(err, domainproject.ErrMemberNotFound):
		return f.businessFailure(event, CheckReasonProjectNotFound, err)
	case errors.Is(err, identity.ErrUserNotFound):
		return f.businessFailure(event, CheckReasonUserNotFound, err)
	default:
		return f.technicalFailure(event, CheckReasonProjectLookupFailed, err)
	}
}

func (f *checkFlow) roleLookupFailure(event CheckEventName, err error) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound), errors.Is(err, domainproject.ErrMemberNotFound):
		return f.businessFailure(event, CheckReasonProjectNotFound, err)
	case errors.Is(err, identity.ErrUserNotFound):
		return f.businessFailure(event, CheckReasonUserNotFound, err)
	default:
		return f.technicalFailure(event, CheckReasonRoleLookupFailed, err)
	}
}

func (f *checkFlow) checkLookupFailure(event CheckEventName, err error) error {
	switch {
	case errors.Is(err, domaincheck.ErrCheckNotFound):
		return f.businessFailure(event, CheckReasonCheckNotFound, err)
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return f.businessFailure(event, CheckReasonProjectNotFound, err)
	case errors.Is(err, ErrInvalidInput), errors.Is(err, domaincheck.ErrInvalidInput), errors.Is(err, label.ErrInvalidInput):
		return f.businessFailure(event, CheckReasonInvalidInput, err)
	default:
		return f.technicalFailure(event, CheckReasonCheckLookupFailed, err)
	}
}

func (f *checkFlow) labelLookupFailure(event CheckEventName, err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput), errors.Is(err, domaincheck.ErrInvalidInput):
		return f.businessFailure(event, CheckReasonInvalidInput, err)
	case errors.Is(err, label.ErrLabelNotFound):
		return f.businessFailure(event, CheckReasonLabelNotFound, err)
	default:
		return f.technicalFailure(event, CheckReasonLabelLookupFailed, err)
	}
}

func (f *checkFlow) writeFailure(event CheckEventName, technicalReason CheckEventReason, err error) error {
	switch {
	case errors.Is(err, domaincheck.ErrCheckNotFound):
		return f.businessFailure(event, CheckReasonCheckNotFound, err)
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return f.businessFailure(event, CheckReasonProjectNotFound, err)
	case errors.Is(err, label.ErrLabelNotFound):
		return f.businessFailure(event, CheckReasonLabelNotFound, err)
	case errors.Is(err, ErrInvalidInput), errors.Is(err, domaincheck.ErrInvalidInput):
		return f.businessFailure(event, CheckReasonInvalidInput, err)
	default:
		return f.technicalFailure(event, technicalReason, err)
	}
}

func (f *checkFlow) assignmentRefreshFailure(event CheckEventName, err error) error {
	return f.technicalFailure(event, CheckReasonAssignmentRefreshFailed, err)
}

func (f *checkFlow) assignmentDeleteFailure(event CheckEventName, err error) error {
	return f.technicalFailure(event, CheckReasonAssignmentDeleteFailed, err)
}

func (f *checkFlow) checkEvent(name CheckEventName, outcome CheckEventOutcome, reason CheckEventReason, err error) CheckEvent {
	return CheckEvent{
		Name:        name,
		Action:      f.action,
		Outcome:     outcome,
		Reason:      reason,
		ActorUserID: f.actorUserID,
		ProjectID:   f.projectID,
		ProjectRef:  f.projectRef,
		ProjectSlug: f.projectSlug,
		CheckID:     f.checkID,
		Err:         err,
	}
}
