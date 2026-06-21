package pgpublicstatus

import "go.opentelemetry.io/otel"

var pgpublicstatusTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/publicstatus")
