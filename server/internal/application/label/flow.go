package label

import (
	"context"
	"errors"

	domainproject "github.com/yorukot/netstamp/internal/domain/project"
	"go.opentelemetry.io/otel/trace"
)

type labelFlow struct {
	service     *Service
	ctx         context.Context
	span        trace.Span
	action      LabelEventAction
	actorUserID string
	projectID   string
	projectRef  string
	projectSlug string
	labelID     string
}

func (s *Service) startLabelFlow(ctx context.Context, spanName string, action LabelEventAction, actorUserID string) (context.Context, *labelFlow) {
	ctx, span := labelTracer.Start(ctx, spanName, trace.WithAttributes(
		attrLabelAction.String(string(action)),
	))
	if actorUserID != "" {
		span.SetAttributes(attrUserID.String(actorUserID))
	}

	return ctx, &labelFlow{
		service:     s,
		ctx:         ctx,
		span:        span,
		action:      action,
		actorUserID: actorUserID,
	}
}

func (f *labelFlow) end() {
	f.span.End()
}

func (f *labelFlow) setProjectRef(projectRef string) {
	f.projectRef = projectRef
	if projectRef != "" {
		f.span.SetAttributes(attrProjectRef.String(projectRef))
	}
}

func (f *labelFlow) setProject(project domainproject.Project) {
	f.projectID = project.ID
	if project.ID != "" {
		f.span.SetAttributes(attrProjectID.String(project.ID))
	}
	f.projectSlug = project.Slug
	if project.Slug != "" {
		f.span.SetAttributes(attrProjectSlug.String(project.Slug))
	}
}

func (f *labelFlow) setProjectID(projectID string) {
	f.projectID = projectID
	if projectID != "" {
		f.span.SetAttributes(attrProjectID.String(projectID))
	}
}

func (f *labelFlow) setLabelID(labelID string) {
	f.labelID = labelID
	if labelID != "" {
		f.span.SetAttributes(attrLabelID.String(labelID))
	}
}

func (f *labelFlow) success(name LabelEventName) {
	f.span.SetAttributes(attrLabelOutcome.String(string(LabelOutcomeSuccess)))
	f.service.events.RecordLabelEvent(f.ctx, f.labelEvent(name, LabelOutcomeSuccess, "", nil))
}

func (f *labelFlow) businessFailure(name LabelEventName, reason LabelEventReason, returnErr error) error {
	f.span.SetAttributes(
		attrLabelOutcome.String(string(LabelOutcomeFailure)),
		attrLabelFailureReason.String(string(reason)),
	)
	f.service.events.RecordLabelEvent(f.ctx, f.labelEvent(name, LabelOutcomeFailure, reason, nil))
	return returnErr
}

func (f *labelFlow) technicalFailure(name LabelEventName, reason LabelEventReason, err error) error {
	f.span.SetAttributes(attrLabelOutcome.String(string(LabelOutcomeFailure)))
	recordSpanError(f.span, err, reason)
	f.service.events.RecordLabelEvent(f.ctx, f.labelEvent(name, LabelOutcomeFailure, reason, err))
	return err
}

func (f *labelFlow) lookupFailure(event LabelEventName, err error) error {
	if errors.Is(err, ErrLabelNotFound) {
		return f.businessFailure(event, LabelReasonLabelNotFound, err)
	}
	if errors.Is(err, ErrProjectNotFound) {
		return f.businessFailure(event, LabelReasonProjectNotFound, err)
	}

	return f.technicalFailure(event, LabelReasonLabelLookupFailed, err)
}

func (f *labelFlow) writeFailure(event LabelEventName, technicalReason LabelEventReason, err error) error {
	switch {
	case errors.Is(err, ErrLabelNotFound):
		return f.businessFailure(event, LabelReasonLabelNotFound, err)
	case errors.Is(err, ErrLabelAlreadyExists):
		return f.businessFailure(event, LabelReasonLabelAlreadyExists, err)
	case errors.Is(err, ErrInvalidInput):
		return f.businessFailure(event, LabelReasonInvalidInput, err)
	case errors.Is(err, ErrProjectNotFound):
		return f.businessFailure(event, LabelReasonProjectNotFound, err)
	default:
		return f.technicalFailure(event, technicalReason, err)
	}
}

func (f *labelFlow) resolveFailure(err error) error {
	switch {
	case errors.Is(err, ErrLabelNotFound):
		return f.businessFailure(LabelEventResolveFailure, LabelReasonLabelNotFound, err)
	case errors.Is(err, ErrInvalidInput):
		return f.businessFailure(LabelEventResolveFailure, LabelReasonInvalidInput, err)
	case errors.Is(err, ErrProjectNotFound):
		return f.businessFailure(LabelEventResolveFailure, LabelReasonProjectNotFound, err)
	default:
		return f.technicalFailure(LabelEventResolveFailure, LabelReasonLabelResolveFailed, err)
	}
}

func (f *labelFlow) labelEvent(name LabelEventName, outcome LabelEventOutcome, reason LabelEventReason, err error) LabelEvent {
	return LabelEvent{
		Name:        name,
		Action:      f.action,
		Outcome:     outcome,
		Reason:      reason,
		ActorUserID: f.actorUserID,
		ProjectID:   f.projectID,
		ProjectRef:  f.projectRef,
		ProjectSlug: f.projectSlug,
		LabelID:     f.labelID,
		Err:         err,
	}
}
