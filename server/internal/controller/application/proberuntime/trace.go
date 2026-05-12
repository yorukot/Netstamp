package proberuntime

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var runtimeTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/application/proberuntime")

const (
	attrProbeRuntimeAction        = attribute.Key("probe_runtime.action")
	attrProbeRuntimeOutcome       = attribute.Key("probe_runtime.outcome")
	attrProbeRuntimeFailureReason = attribute.Key("probe_runtime.failure.reason")
	attrProbeID                   = attribute.Key("probe.id")
	attrProjectID                 = attribute.Key("project.id")
)

func recordSpanError(span trace.Span, err error, reason ProbeRuntimeEventReason) {
	if err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, string(reason))
	span.SetAttributes(attrProbeRuntimeFailureReason.String(string(reason)))
}
