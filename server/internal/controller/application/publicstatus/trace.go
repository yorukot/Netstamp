package publicstatus

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var publicStatusTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/application/publicstatus")

var (
	attrPublicStatusAction        = attribute.Key("public_status.action")
	attrPublicStatusOutcome       = attribute.Key("public_status.outcome")
	attrPublicStatusFailureReason = attribute.Key("public_status.failure.reason")
	attrPublicStatusPageID        = attribute.Key("public_status.page.id")
	attrPublicStatusElementID     = attribute.Key("public_status.element.id")
	attrProjectID                 = attribute.Key("project.id")
	attrProjectRef                = attribute.Key("project.ref")
	attrProjectSlug               = attribute.Key("project.slug")
	attrUserID                    = attribute.Key("user.id")
	attrErrorType                 = attribute.Key("error.type")
)

func recordSpanError(span trace.Span, err error, reason PublicStatusReason) {
	span.RecordError(err)
	markSpanTechnicalFailure(span, reason)
}

func markSpanTechnicalFailure(span trace.Span, reason PublicStatusReason) {
	reasonValue := string(reason)
	span.SetStatus(codes.Error, reasonValue)
	span.SetAttributes(
		attrErrorType.String(reasonValue),
		attrPublicStatusFailureReason.String(reasonValue),
	)
}
