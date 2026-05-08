package label

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var labelTracer = otel.Tracer("github.com/yorukot/netstamp/internal/application/label")

var (
	attrLabelAction = attribute.Key("label.action")
	attrProjectID   = attribute.Key("project.id")
	attrProjectRef  = attribute.Key("project.ref")
	attrUserID      = attribute.Key("user.id")
	attrLabelID     = attribute.Key("label.id")
	attrErrorType   = attribute.Key("error.type")
)

func recordSpanError(span trace.Span, err error, reason string) {
	span.RecordError(err)
	span.SetStatus(codes.Error, reason)
	span.SetAttributes(attrErrorType.String(reason))
}
