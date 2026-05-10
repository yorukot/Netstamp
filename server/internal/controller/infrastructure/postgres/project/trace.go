package pgproject

import "go.opentelemetry.io/otel"

var pgprojectTracer = otel.Tracer("github.com/yorukot/netstamp/internal/controller/infrastructure/postgres")
