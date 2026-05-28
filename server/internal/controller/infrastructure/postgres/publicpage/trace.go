package pgpublicpage

import "go.opentelemetry.io/otel"

var pgpublicpageTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/publicpage")
