package pgtcp

import "go.opentelemetry.io/otel"

var pgtcpTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/infrastructure/postgres")
