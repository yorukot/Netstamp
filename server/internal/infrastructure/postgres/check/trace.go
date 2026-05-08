package pgcheck

import "go.opentelemetry.io/otel"

var pgcheckTracer = otel.Tracer("github.com/yorukot/netstamp/internal/infrastructure/postgres")
