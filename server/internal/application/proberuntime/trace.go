package proberuntime

import "go.opentelemetry.io/otel"

var runtimeTracer = otel.Tracer("github.com/yorukot/netstamp/internal/application/proberuntime")
