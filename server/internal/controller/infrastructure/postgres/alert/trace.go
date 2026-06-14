package pgalert

import "go.opentelemetry.io/otel"

var pgalertTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/alert")
