package account

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var userTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/application/user")

var (
	attrUserAction        = attribute.Key("user.action")
	attrUserOutcome       = attribute.Key("user.outcome")
	attrUserFailureReason = attribute.Key("user.failure.reason")
	attrUserID            = attribute.Key("user.id")
	attrErrorType         = attribute.Key("error.type")
)

func recordSpanError(span trace.Span, err error, reason UserEventReason) {
	span.RecordError(err)
	markSpanTechnicalFailure(span, reason)
}

func markSpanTechnicalFailure(span trace.Span, reason UserEventReason) {
	reasonValue := string(reason)
	span.SetStatus(codes.Error, reasonValue)
	span.SetAttributes(
		attrErrorType.String(reasonValue),
		attrUserFailureReason.String(reasonValue),
	)
}
