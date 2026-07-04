package alert

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var alertTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/application/alert")

var (
	attrAlertAction         = attribute.Key("alert.action")
	attrAlertOutcome        = attribute.Key("alert.outcome")
	attrAlertFailureReason  = attribute.Key("alert.failure.reason")
	attrAlertRuleID         = attribute.Key("alert.rule.id")
	attrAlertNotificationID = attribute.Key("alert.notification.id")
	attrAlertIncidentID     = attribute.Key("alert.incident.id")
	attrProjectID           = attribute.Key("project.id")
	attrProjectRef          = attribute.Key("project.ref")
	attrProjectSlug         = attribute.Key("project.slug")
	attrUserID              = attribute.Key("user.id")
	attrErrorType           = attribute.Key("error.type")
)

func recordSpanError(span trace.Span, err error, reason AlertReason) {
	span.RecordError(err)
	markSpanTechnicalFailure(span, reason)
}

func markSpanTechnicalFailure(span trace.Span, reason AlertReason) {
	reasonValue := string(reason)
	span.SetStatus(codes.Error, reasonValue)
	span.SetAttributes(
		attrErrorType.String(reasonValue),
		attrAlertFailureReason.String(reasonValue),
	)
}
