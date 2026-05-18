package pgtraceroute

import "go.opentelemetry.io/otel"

var pgtracerouteTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/infrastructure/postgres")
