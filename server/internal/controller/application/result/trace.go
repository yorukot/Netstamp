package result

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var resultTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/application/result")

const (
	attrProjectRef = attribute.Key("project.ref")
	attrProjectID  = attribute.Key("project.id")
	attrProbeID    = attribute.Key("probe.id")
	attrCheckID    = attribute.Key("check.id")
	attrMetric     = attribute.Key("result.metric")
)
