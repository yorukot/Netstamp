package auth

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var authTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/application/auth")

var (
	attrAuthAction        = attribute.Key("auth.action")
	attrAuthOutcome       = attribute.Key("auth.outcome")
	attrAuthFailureReason = attribute.Key("auth.failure.reason")
	attrErrorType         = attribute.Key("error.type")
	attrUserID            = attribute.Key("user.id")
)

func recordSpanError(span trace.Span, err error, reason AuthEventReason) {
	span.RecordError(err)
	markSpanTechnicalFailure(span, reason)
}

func markSpanTechnicalFailure(span trace.Span, reason AuthEventReason) {
	reasonValue := string(reason)
	span.SetStatus(codes.Error, reasonValue)
	span.SetAttributes(
		attrErrorType.String(reasonValue),
		attrAuthFailureReason.String(reasonValue),
	)
}
