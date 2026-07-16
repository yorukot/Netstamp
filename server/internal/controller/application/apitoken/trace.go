package apitoken

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/application/apitoken")
