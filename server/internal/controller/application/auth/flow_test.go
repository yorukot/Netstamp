package auth

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestAuthTechnicalFailureRecordsSpanError(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	defer otel.SetTracerProvider(previous)
	defer func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			t.Fatalf("shutdown tracer provider: %v", err)
		}
	}()

	service := &Service{}
	ctx, flow := service.startAuthFlow(context.Background(), "auth.test", AuthActionRegister, "user@example.com")
	err := errors.New("hash unavailable")

	flow.recordTechnicalFailure(AuthEventRegisterFailure, AuthReasonPasswordHashFailed, err)
	flow.end()

	if flushErr := provider.ForceFlush(ctx); flushErr != nil {
		t.Fatalf("force flush spans: %v", flushErr)
	}
	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected one span, got %d", len(spans))
	}
	if spans[0].Status.Code != codes.Error {
		t.Fatalf("expected error status, got %#v", spans[0].Status)
	}
	if !hasExceptionEvent(spans[0].Events) {
		t.Fatalf("expected technical failure span to include an exception event, got %#v", spans[0].Events)
	}
}

func hasExceptionEvent(events []sdktrace.Event) bool {
	for _, event := range events {
		if event.Name == "exception" {
			return true
		}
	}
	return false
}
