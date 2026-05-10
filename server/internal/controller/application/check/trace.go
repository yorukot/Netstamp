package check

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var checkTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/application/check")

const (
	attrCheckAction        = attribute.Key("check.action")
	attrCheckOutcome       = attribute.Key("check.outcome")
	attrCheckFailureReason = attribute.Key("check.failure.reason")
	attrProjectID          = attribute.Key("project.id")
	attrProjectRef         = attribute.Key("project.ref")
	attrProjectSlug        = attribute.Key("project.slug")
	attrUserID             = attribute.Key("user.id")
	attrCheckID            = attribute.Key("check.id")
)

func recordSpanError(span trace.Span, err error, reason CheckEventReason) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, string(reason))
	span.SetAttributes(attrCheckFailureReason.String(string(reason)))
}
