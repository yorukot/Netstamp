package publicpage

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var publicPageTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/application/publicpage")

const (
	attrPublicPageAction        = attribute.Key("public_page.action")
	attrPublicPageOutcome       = attribute.Key("public_page.outcome")
	attrPublicPageFailureReason = attribute.Key("public_page.failure.reason")
	attrProjectID               = attribute.Key("project.id")
	attrProjectRef              = attribute.Key("project.ref")
	attrProjectSlug             = attribute.Key("project.slug")
	attrUserID                  = attribute.Key("user.id")
	attrPublicPageID            = attribute.Key("public_page.id")
	attrPublicPageSlug          = attribute.Key("public_page.slug")
	attrPublicPageFolderID      = attribute.Key("public_page.folder.id")
	attrCheckID                 = attribute.Key("check.id")
	attrProbeID                 = attribute.Key("probe.id")
	attrCheckCount              = attribute.Key("check.count")
)

func recordSpanError(span trace.Span, err error, reason PublicPageEventReason) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, string(reason))
	span.SetAttributes(attrPublicPageFailureReason.String(string(reason)))
}
