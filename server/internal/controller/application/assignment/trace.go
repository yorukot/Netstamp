package assignment

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var assignmentTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/application/assignment")

var (
	attrAssignmentAction        = attribute.Key("assignment.action")
	attrAssignmentOutcome       = attribute.Key("assignment.outcome")
	attrAssignmentFailureReason = attribute.Key("assignment.failure.reason")
	attrProjectID               = attribute.Key("project.id")
	attrProbeID                 = attribute.Key("probe.id")
	attrCheckID                 = attribute.Key("check.id")
	attrLabelID                 = attribute.Key("label.id")
	attrErrorType               = attribute.Key("error.type")
)

func recordSpanError(span trace.Span, err error, reason AssignmentEventReason) {
	span.RecordError(err)
	markSpanTechnicalFailure(span, reason)
}

func markSpanTechnicalFailure(span trace.Span, reason AssignmentEventReason) {
	reasonValue := string(reason)
	span.SetStatus(codes.Error, reasonValue)
	span.SetAttributes(
		attrErrorType.String(reasonValue),
		attrAssignmentFailureReason.String(reasonValue),
	)
}
