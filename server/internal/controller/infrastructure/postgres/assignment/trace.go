package pgassignment

import "go.opentelemetry.io/otel"

var pgassignmentTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/infrastructure/postgres")
