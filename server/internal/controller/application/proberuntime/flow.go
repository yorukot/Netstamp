package proberuntime

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/trace"

	domainping "github.com/yorukot/netstamp/internal/domain/ping"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
)

type runtimeFlow struct {
	service     *Service
	ctx         context.Context
	span        trace.Span
	action      ProbeRuntimeEventAction
	probeID     string
	projectID   string
	resultCount *int
}

func (s *Service) startRuntimeFlow(ctx context.Context, spanName string, action ProbeRuntimeEventAction) (context.Context, *runtimeFlow) {
	ctx, span := runtimeTracer.Start(ctx, spanName, trace.WithAttributes(
		attrProbeRuntimeAction.String(string(action)),
	))

	return ctx, &runtimeFlow{
		service: s,
		ctx:     ctx,
		span:    span,
		action:  action,
	}
}

func (f *runtimeFlow) end() {
	f.span.End()
}

func (f *runtimeFlow) setProbeID(probeID string) {
	f.probeID = probeID
	if probeID != "" {
		f.span.SetAttributes(attrProbeID.String(probeID))
	}
}

func (f *runtimeFlow) setCredential(credential domainprobe.Credential) {
	f.setProbeID(credential.ProbeID)
	f.projectID = credential.ProjectID
	if credential.ProjectID != "" {
		f.span.SetAttributes(attrProjectID.String(credential.ProjectID))
	}
}

func (f *runtimeFlow) setResultCount(resultCount int) {
	f.resultCount = &resultCount
	f.span.SetAttributes(attrResultCount.Int(resultCount))
}

func (f *runtimeFlow) success() {
	f.span.SetAttributes(attrProbeRuntimeOutcome.String(string(ProbeRuntimeOutcomeSuccess)))
}

func (f *runtimeFlow) businessFailure(name ProbeRuntimeEventName, reason ProbeRuntimeEventReason, returnErr error) error {
	f.span.SetAttributes(
		attrProbeRuntimeOutcome.String(string(ProbeRuntimeOutcomeFailure)),
		attrProbeRuntimeFailureReason.String(string(reason)),
	)
	f.service.events.RecordProbeRuntimeEvent(f.ctx, f.runtimeEvent(name, ProbeRuntimeOutcomeFailure, reason, nil))
	return returnErr
}

func (f *runtimeFlow) technicalFailure(name ProbeRuntimeEventName, reason ProbeRuntimeEventReason, err error) error {
	f.span.SetAttributes(attrProbeRuntimeOutcome.String(string(ProbeRuntimeOutcomeFailure)))
	recordSpanError(f.span, err, reason)
	f.service.events.RecordProbeRuntimeEvent(f.ctx, f.runtimeEvent(name, ProbeRuntimeOutcomeFailure, reason, err))
	return err
}

func (f *runtimeFlow) authenticationFailure(name ProbeRuntimeEventName, err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return f.businessFailure(name, ProbeRuntimeReasonInvalidInput, err)
	case errors.Is(err, domainprobe.ErrInvalidCredential):
		return f.businessFailure(name, ProbeRuntimeReasonInvalidCredential, err)
	case errors.Is(err, domainprobe.ErrProbeNotFound):
		return f.businessFailure(name, ProbeRuntimeReasonProbeNotFound, err)
	case errors.Is(err, domainprobe.ErrProbeDisabled):
		return f.businessFailure(name, ProbeRuntimeReasonProbeDisabled, err)
	case errors.Is(err, errSecretVerifierMissing):
		return f.technicalFailure(name, ProbeRuntimeReasonSecretVerifierMissing, err)
	default:
		return f.technicalFailure(name, ProbeRuntimeReasonCredentialLookupFail, err)
	}
}

func (f *runtimeFlow) statusUpdateFailure(name ProbeRuntimeEventName, err error) error {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return f.businessFailure(name, ProbeRuntimeReasonInvalidInput, err)
	case errors.Is(err, domainprobe.ErrProbeNotFound):
		return f.businessFailure(name, ProbeRuntimeReasonProbeNotFound, err)
	default:
		return f.technicalFailure(name, ProbeRuntimeReasonStatusUpdateFailed, err)
	}
}

func (f *runtimeFlow) assignmentListFailure(name ProbeRuntimeEventName, err error) error {
	switch {
	case errors.Is(err, domainprobe.ErrProbeNotFound):
		return f.businessFailure(name, ProbeRuntimeReasonProbeNotFound, err)
	default:
		return f.technicalFailure(name, ProbeRuntimeReasonAssignmentListFailed, err)
	}
}

func (f *runtimeFlow) resultWriteFailure(err error) error {
	if errors.Is(err, domainping.ErrInvalidResult) {
		return f.businessFailure(ProbeRuntimeEventSubmitResultsFailure, ProbeRuntimeReasonInvalidResult, err)
	}

	return f.technicalFailure(ProbeRuntimeEventSubmitResultsFailure, ProbeRuntimeReasonResultWriteFailed, err)
}

func (f *runtimeFlow) runtimeEvent(name ProbeRuntimeEventName, outcome ProbeRuntimeEventOutcome, reason ProbeRuntimeEventReason, err error) ProbeRuntimeEvent {
	return ProbeRuntimeEvent{
		Name:        name,
		Action:      f.action,
		Outcome:     outcome,
		Reason:      reason,
		ProbeID:     f.probeID,
		ProjectID:   f.projectID,
		ResultCount: f.resultCount,
		Err:         err,
	}
}
