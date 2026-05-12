package probe

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/trace"

	"github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
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

func (f *probeFlow) setProbeID(probeID string) {
	f.probeID = probeID
	if probeID != "" {
		f.span.SetAttributes(attrProbeID.String(probeID))
	}
}

func (f *probeFlow) success(name ProbeEventName) {
	f.span.SetAttributes(attrProbeOutcome.String(string(ProbeOutcomeSuccess)))
	f.service.events.RecordProbeEvent(f.ctx, f.probeEvent(name, ProbeOutcomeSuccess, "", nil))
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

func (f *probeFlow) projectLookupFailure(event ProbeEventName, err error) error {
	if errors.Is(err, domainproject.ErrProjectNotFound) {
		return f.businessFailure(event, ProbeReasonProjectNotFound, err)
	}

	return f.technicalFailure(event, ProbeReasonProjectLookupFailed, err)
}

func (f *probeFlow) roleLookupFailure(event ProbeEventName, err error) error {
	if errors.Is(err, domainproject.ErrProjectNotFound) {
		return f.businessFailure(event, ProbeReasonProjectNotFound, err)
	}

	return f.technicalFailure(event, ProbeReasonRoleLookupFailed, err)
}

func (f *probeFlow) labelLookupFailure(event ProbeEventName, err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput), errors.Is(err, label.ErrInvalidInput), errors.Is(err, domainprobe.ErrInvalidInput):
		return f.businessFailure(event, ProbeReasonInvalidInput, err)
	case errors.Is(err, label.ErrLabelNotFound):
		return f.businessFailure(event, ProbeReasonLabelNotFound, err)
	default:
		return f.technicalFailure(event, ProbeReasonLabelLookupFailed, err)
	}
}

func (f *probeFlow) createFailure(err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput), errors.Is(err, label.ErrInvalidInput), errors.Is(err, domainprobe.ErrInvalidInput):
		return f.businessFailure(ProbeEventCreateFailure, ProbeReasonInvalidInput, err)
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return f.businessFailure(ProbeEventCreateFailure, ProbeReasonProjectNotFound, err)
	case errors.Is(err, label.ErrLabelNotFound):
		return f.businessFailure(ProbeEventCreateFailure, ProbeReasonLabelNotFound, err)
	default:
		return f.technicalFailure(ProbeEventCreateFailure, ProbeReasonProbeCreateFailed, err)
	}
}

func (f *probeFlow) probeLookupFailure(event ProbeEventName, err error) error {
	if errors.Is(err, domainprobe.ErrProbeNotFound) {
		return f.businessFailure(event, ProbeReasonProbeNotFound, err)
	}

	return f.technicalFailure(event, ProbeReasonProbeLookupFailed, err)
}

func (f *probeFlow) probeListFailure(err error) error {
	if errors.Is(err, domainproject.ErrProjectNotFound) {
		return f.businessFailure(ProbeEventListFailure, ProbeReasonProjectNotFound, err)
	}

	return f.technicalFailure(ProbeEventListFailure, ProbeReasonProbeListFailed, err)
}

func (f *probeFlow) updateFailure(err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput), errors.Is(err, label.ErrInvalidInput), errors.Is(err, domainprobe.ErrInvalidInput):
		return f.businessFailure(ProbeEventUpdateFailure, ProbeReasonInvalidInput, err)
	case errors.Is(err, domainprobe.ErrProbeNotFound):
		return f.businessFailure(ProbeEventUpdateFailure, ProbeReasonProbeNotFound, err)
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return f.businessFailure(ProbeEventUpdateFailure, ProbeReasonProjectNotFound, err)
	case errors.Is(err, label.ErrLabelNotFound):
		return f.businessFailure(ProbeEventUpdateFailure, ProbeReasonLabelNotFound, err)
	default:
		return f.technicalFailure(ProbeEventUpdateFailure, ProbeReasonProbeUpdateFailed, err)
	}
}

func (f *probeFlow) deleteFailure(err error) error {
	switch {
	case errors.Is(err, domainprobe.ErrProbeNotFound):
		return f.businessFailure(ProbeEventDeleteFailure, ProbeReasonProbeNotFound, err)
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return f.businessFailure(ProbeEventDeleteFailure, ProbeReasonProjectNotFound, err)
	default:
		return f.technicalFailure(ProbeEventDeleteFailure, ProbeReasonProbeDeleteFailed, err)
	}
}

func (f *probeFlow) rotateFailure(err error) error {
	switch {
	case errors.Is(err, domainprobe.ErrProbeNotFound):
		return f.businessFailure(ProbeEventSecretRotateFailure, ProbeReasonProbeNotFound, err)
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return f.businessFailure(ProbeEventSecretRotateFailure, ProbeReasonProjectNotFound, err)
	default:
		return f.technicalFailure(ProbeEventSecretRotateFailure, ProbeReasonSecretRotateFailed, err)
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
