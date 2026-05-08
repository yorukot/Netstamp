package probe

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var probeTracer = otel.Tracer("github.com/yorukot/netstamp/internal/application/probe")

var (
	attrProbeAction        = attribute.Key("probe.action")
	attrProbeOutcome       = attribute.Key("probe.outcome")
	attrProbeFailureReason = attribute.Key("probe.failure.reason")
	attrErrorType          = attribute.Key("error.type")
	attrUserID             = attribute.Key("user.id")
	attrProjectID          = attribute.Key("project.id")
	attrProjectRef         = attribute.Key("project.ref")
)

func recordSpanError(span trace.Span, err error, reason ProbeEventReason) {
	span.RecordError(err)
	markSpanTechnicalFailure(span, reason)
}

func markSpanTechnicalFailure(span trace.Span, reason ProbeEventReason) {
	reasonValue := string(reason)
	span.SetStatus(codes.Error, reasonValue)
	span.SetAttributes(
		attrErrorType.String(reasonValue),
		attrProbeFailureReason.String(reasonValue),
	)
}
