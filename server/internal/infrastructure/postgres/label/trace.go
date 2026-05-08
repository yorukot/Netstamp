package pglabel

import "go.opentelemetry.io/otel"

var pglabelTracer = otel.Tracer("github.com/yorukot/netstamp/internal/infrastructure/postgres")
