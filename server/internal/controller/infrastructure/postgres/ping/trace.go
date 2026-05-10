package pgping

import "go.opentelemetry.io/otel"

var pgpingTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/infrastructure/postgres")
