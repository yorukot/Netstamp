package pgresult

import "go.opentelemetry.io/otel"

var pgresultTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/result")
