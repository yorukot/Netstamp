package pgprobe

import "go.opentelemetry.io/otel"

var pgprobeTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/infrastructure/postgres")
