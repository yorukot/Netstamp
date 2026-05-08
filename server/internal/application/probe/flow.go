package probe

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/trace"
)

type probeFlow struct {
	service     *Service
	ctx         context.Context
	span        trace.Span
	action      ProbeEventAction
	actorUserID string
	projectID   string
	projectRef  string
	probeID     string
}

func (s *Service) startProbeFlow(ctx context.Context, spanName string, action ProbeEventAction, actorUserID string) (context.Context, *probeFlow) {
	ctx, span := probeTracer.Start(ctx, spanName, trace.WithAttributes(
		attrProbeAction.String(string(action)),
	))
	if actorUserID != "" {
		span.SetAttributes(attrUserID.String(actorUserID))
	}

	return ctx, &probeFlow{
		service:     s,
		ctx:         ctx,
		span:        span,
		action:      action,
		actorUserID: actorUserID,
	}
}

func (f *probeFlow) end() {
	f.span.End()
}

func (f *probeFlow) setProjectRef(projectRef string) {
	f.projectRef = projectRef
	if projectRef != "" {
		f.span.SetAttributes(attrProjectRef.String(projectRef))
	}
}

func (f *probeFlow) setProjectID(projectID string) {
	f.projectID = projectID
	if projectID != "" {
		f.span.SetAttributes(attrProjectID.String(projectID))
	}
}

func (f *probeFlow) businessFailure(name ProbeEventName, reason ProbeEventReason, returnErr error) error {
	f.span.SetAttributes(
		attrProbeOutcome.String(string(ProbeOutcomeFailure)),
		attrProbeFailureReason.String(string(reason)),
	)
	f.service.events.RecordProbeEvent(f.ctx, f.probeEvent(name, ProbeOutcomeFailure, reason, nil))
	return returnErr
}

func (f *probeFlow) technicalFailure(name ProbeEventName, reason ProbeEventReason, err error) error {
	f.span.SetAttributes(attrProbeOutcome.String(string(ProbeOutcomeFailure)))
	recordSpanError(f.span, err, reason)
	f.service.events.RecordProbeEvent(f.ctx, f.probeEvent(name, ProbeOutcomeFailure, reason, err))
	return err
}

func (f *probeFlow) projectLookupFailure(err error) error {
	if errors.Is(err, ErrProjectNotFound) {
		return f.businessFailure(ProbeEventCreateFailure, ProbeReasonProjectNotFound, err)
	}

	return f.technicalFailure(ProbeEventCreateFailure, ProbeReasonProjectLookupFailed, err)
}

func (f *probeFlow) labelLookupFailure(err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return f.businessFailure(ProbeEventCreateFailure, ProbeReasonInvalidInput, err)
	case errors.Is(err, ErrLabelNotFound):
		return f.businessFailure(ProbeEventCreateFailure, ProbeReasonLabelNotFound, err)
	default:
		return f.technicalFailure(ProbeEventCreateFailure, ProbeReasonLabelLookupFailed, err)
	}
}

func (f *probeFlow) createFailure(err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return f.businessFailure(ProbeEventCreateFailure, ProbeReasonInvalidInput, err)
	case errors.Is(err, ErrProjectNotFound):
		return f.businessFailure(ProbeEventCreateFailure, ProbeReasonProjectNotFound, err)
	case errors.Is(err, ErrLabelNotFound):
		return f.businessFailure(ProbeEventCreateFailure, ProbeReasonLabelNotFound, err)
	default:
		return f.technicalFailure(ProbeEventCreateFailure, ProbeReasonProbeCreateFailed, err)
	}
}

func (f *probeFlow) probeEvent(name ProbeEventName, outcome ProbeEventOutcome, reason ProbeEventReason, err error) ProbeEvent {
	return ProbeEvent{
		Name:        name,
		Action:      f.action,
		Outcome:     outcome,
		Reason:      reason,
		ActorUserID: f.actorUserID,
		ProjectID:   f.projectID,
		ProjectRef:  f.projectRef,
		ProbeID:     f.probeID,
		Err:         err,
	}
}
