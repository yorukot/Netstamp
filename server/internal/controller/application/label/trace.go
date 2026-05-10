package label

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var labelTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/application/label")

var (
	attrLabelAction        = attribute.Key("label.action")
	attrLabelOutcome       = attribute.Key("label.outcome")
	attrLabelFailureReason = attribute.Key("label.failure.reason")
	attrProjectID          = attribute.Key("project.id")
	attrProjectRef         = attribute.Key("project.ref")
	attrProjectSlug        = attribute.Key("project.slug")
	attrUserID             = attribute.Key("user.id")
	attrLabelID            = attribute.Key("label.id")
	attrErrorType          = attribute.Key("error.type")
)

func recordSpanError(span trace.Span, err error, reason LabelEventReason) {
	span.RecordError(err)
	markSpanTechnicalFailure(span, reason)
}

func markSpanTechnicalFailure(span trace.Span, reason LabelEventReason) {
	reasonValue := string(reason)
	span.SetStatus(codes.Error, reasonValue)
	span.SetAttributes(
		attrErrorType.String(reasonValue),
		attrLabelFailureReason.String(reasonValue),
	)
}
