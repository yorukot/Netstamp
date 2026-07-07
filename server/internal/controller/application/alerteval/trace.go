package alerteval

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var alertEvalTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/application/alerteval")

var (
	attrAlertEvalAction        = attribute.Key("alert_eval.action")
	attrAlertEvalOutcome       = attribute.Key("alert_eval.outcome")
	attrAlertEvalFailureReason = attribute.Key("alert_eval.failure.reason")
	attrProjectID              = attribute.Key("project.id")
	attrProbeID                = attribute.Key("probe.id")
	attrCheckID                = attribute.Key("check.id")
	attrCheckType              = attribute.Key("check.type")
	attrAlertRuleID            = attribute.Key("alert.rule.id")
	attrAlertIncidentID        = attribute.Key("alert.incident.id")
	attrErrorType              = attribute.Key("error.type")
)

func recordSpanError(span trace.Span, err error, reason AlertEvalReason) {
	span.RecordError(err)
	markSpanTechnicalFailure(span, reason)
}

func markSpanTechnicalFailure(span trace.Span, reason AlertEvalReason) {
	reasonValue := string(reason)
	span.SetStatus(codes.Error, reasonValue)
	span.SetAttributes(
		attrErrorType.String(reasonValue),
		attrAlertEvalFailureReason.String(reasonValue),
	)
}
