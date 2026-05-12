package assignment

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/trace"

	domaincheck "github.com/yorukot/netstamp/internal/domain/check"
	domainlabel "github.com/yorukot/netstamp/internal/domain/label"
	domainprobe "github.com/yorukot/netstamp/internal/domain/probe"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

type assignmentFlow struct {
	service   *Service
	ctx       context.Context
	span      trace.Span
	action    AssignmentEventAction
	projectID string
	probeID   string
	checkID   string
	labelID   string
}

func (s *Service) startAssignmentFlow(ctx context.Context, spanName string, action AssignmentEventAction) (context.Context, *assignmentFlow) {
	ctx, span := assignmentTracer.Start(ctx, spanName, trace.WithAttributes(
		attrAssignmentAction.String(string(action)),
	))

	return ctx, &assignmentFlow{
		service: s,
		ctx:     ctx,
		span:    span,
		action:  action,
	}
}

func (f *assignmentFlow) end() {
	f.span.End()
}

func (f *assignmentFlow) setProjectID(projectID string) {
	f.projectID = projectID
	if projectID != "" {
		f.span.SetAttributes(attrProjectID.String(projectID))
	}
}

func (f *assignmentFlow) setProbeID(probeID string) {
	f.probeID = probeID
	if probeID != "" {
		f.span.SetAttributes(attrProbeID.String(probeID))
	}
}

func (f *assignmentFlow) setCheckID(checkID string) {
	f.checkID = checkID
	if checkID != "" {
		f.span.SetAttributes(attrCheckID.String(checkID))
	}
}

func (f *assignmentFlow) setLabelID(labelID string) {
	f.labelID = labelID
	if labelID != "" {
		f.span.SetAttributes(attrLabelID.String(labelID))
	}
}

func (f *assignmentFlow) success(name AssignmentEventName) {
	f.span.SetAttributes(attrAssignmentOutcome.String(string(AssignmentOutcomeSuccess)))
	f.record(name, AssignmentOutcomeSuccess, "", nil)
}

func (f *assignmentFlow) businessFailure(name AssignmentEventName, reason AssignmentEventReason, returnErr error) error {
	f.span.SetAttributes(
		attrAssignmentOutcome.String(string(AssignmentOutcomeFailure)),
		attrAssignmentFailureReason.String(string(reason)),
	)
	f.record(name, AssignmentOutcomeFailure, reason, nil)
	return returnErr
}

func (f *assignmentFlow) technicalFailure(name AssignmentEventName, reason AssignmentEventReason, err error) error {
	f.span.SetAttributes(attrAssignmentOutcome.String(string(AssignmentOutcomeFailure)))
	recordSpanError(f.span, err, reason)
	f.record(name, AssignmentOutcomeFailure, reason, err)
	return err
}

func (f *assignmentFlow) refreshFailure(event AssignmentEventName, err error) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return f.businessFailure(event, AssignmentReasonProjectNotFound, err)
	case errors.Is(err, domainprobe.ErrProbeNotFound):
		return f.businessFailure(event, AssignmentReasonProbeNotFound, err)
	case errors.Is(err, domaincheck.ErrCheckNotFound):
		return f.businessFailure(event, AssignmentReasonCheckNotFound, err)
	case errors.Is(err, domainlabel.ErrLabelNotFound):
		return f.businessFailure(event, AssignmentReasonLabelNotFound, err)
	default:
		return f.technicalFailure(event, AssignmentReasonRefreshFailed, err)
	}
}

func (f *assignmentFlow) deleteFailure(event AssignmentEventName, err error) error {
	switch {
	case errors.Is(err, domainproject.ErrProjectNotFound):
		return f.businessFailure(event, AssignmentReasonProjectNotFound, err)
	case errors.Is(err, domainprobe.ErrProbeNotFound):
		return f.businessFailure(event, AssignmentReasonProbeNotFound, err)
	case errors.Is(err, domaincheck.ErrCheckNotFound):
		return f.businessFailure(event, AssignmentReasonCheckNotFound, err)
	default:
		return f.technicalFailure(event, AssignmentReasonDeleteFailed, err)
	}
}

func (f *assignmentFlow) record(name AssignmentEventName, outcome AssignmentEventOutcome, reason AssignmentEventReason, err error) {
	f.service.events.RecordAssignmentEvent(f.ctx, AssignmentEvent{
		Name:      name,
		Action:    f.action,
		Outcome:   outcome,
		Reason:    reason,
		ProjectID: f.projectID,
		ProbeID:   f.probeID,
		CheckID:   f.checkID,
		LabelID:   f.labelID,
		Err:       err,
	})
}
