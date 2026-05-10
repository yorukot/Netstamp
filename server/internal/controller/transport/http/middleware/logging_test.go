package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestZapRequestLoggerIncludesTraceFields(t *testing.T) {
	core, observed := observer.New(zap.DebugLevel)
	log := zap.New(core)

	traceID := trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	spanID := trace.SpanID{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	}))

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/health/live", nil)
	res := httptest.NewRecorder()
	handler := ZapRequestLogger(log)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	handler.ServeHTTP(res, req)

	logs := observed.All()
	if len(logs) != 1 {
		t.Fatalf("expected one log entry, got %d", len(logs))
	}

	fields := logs[0].ContextMap()
	if got := fields["trace_id"]; got != traceID.String() {
		t.Fatalf("expected trace_id %q, got %#v", traceID.String(), got)
	}
	if got := fields["span_id"]; got != spanID.String() {
		t.Fatalf("expected span_id %q, got %#v", spanID.String(), got)
	}
}
