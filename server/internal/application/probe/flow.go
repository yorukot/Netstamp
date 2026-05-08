package probe

import (
	"context"

	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
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

func (f *probeFlow) End() {
	f.span.End()
}

func (f *probeFlow) SetProjectRef(projectRef string) {
	f.projectRef = projectRef
	if projectRef != "" {
		f.span.SetAttributes(attrProjectRef.String(projectRef))
	}
}

func (f *probeFlow) SetProjectID(projectID string) {
	f.projectID = projectID
	if projectID != "" {
		f.span.SetAttributes(attrProjectID.String(projectID))
	}
}

func (f *probeFlow) SetProbe(probe domainprobe.Probe) {
	f.probeID = probe.ID
	if probe.ID != "" {
		f.span.SetAttributes(attrProbeID.String(probe.ID))
	}
	if probe.ProjectID != "" {
		f.SetProjectID(probe.ProjectID)
	}
}

func (f *probeFlow) Success(name ProbeEventName) {
	f.span.SetAttributes(attrProbeOutcome.String(string(ProbeOutcomeSuccess)))
	f.service.events.RecordProbeEvent(f.ctx, f.probeEvent(name, ProbeOutcomeSuccess, "", nil))
}

func (f *probeFlow) BusinessFailure(name ProbeEventName, reason ProbeEventReason, returnErr error) error {
	f.span.SetAttributes(
		attrProbeOutcome.String(string(ProbeOutcomeFailure)),
		attrProbeFailureReason.String(string(reason)),
	)
	f.service.events.RecordProbeEvent(f.ctx, f.probeEvent(name, ProbeOutcomeFailure, reason, nil))
	return returnErr
}

func (f *probeFlow) TechnicalFailure(name ProbeEventName, reason ProbeEventReason, err error) error {
	f.span.SetAttributes(attrProbeOutcome.String(string(ProbeOutcomeFailure)))
	recordSpanError(f.span, err, reason)
	f.service.events.RecordProbeEvent(f.ctx, f.probeEvent(name, ProbeOutcomeFailure, reason, err))
	return err
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
