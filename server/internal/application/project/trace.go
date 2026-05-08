package project

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var projectTracer = otel.Tracer("github.com/yorukot/netstamp/internal/application/project")

var (
	attrProjectAction        = attribute.Key("project.action")
	attrProjectOutcome       = attribute.Key("project.outcome")
	attrProjectFailureReason = attribute.Key("project.failure.reason")
	attrErrorType            = attribute.Key("error.type")
	attrUserID               = attribute.Key("user.id")
	attrProjectID            = attribute.Key("project.id")
	attrProjectRef           = attribute.Key("project.ref")
	attrProjectSlug          = attribute.Key("project.slug")
	attrProjectMemberUserID  = attribute.Key("project.member.user.id")
	attrProjectMemberRole    = attribute.Key("project.member.role")
)

func recordSpanError(span trace.Span, err error, reason ProjectEventReason) {
	span.RecordError(err)
	markSpanTechnicalFailure(span, reason)
}

func markSpanTechnicalFailure(span trace.Span, reason ProjectEventReason) {
	reasonValue := string(reason)
	span.SetStatus(codes.Error, reasonValue)
	span.SetAttributes(
		attrErrorType.String(reasonValue),
		attrProjectFailureReason.String(reasonValue),
	)
}
