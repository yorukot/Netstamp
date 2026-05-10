package pguser

import "go.opentelemetry.io/otel"

var pguserTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/infrastructure/postgres")
