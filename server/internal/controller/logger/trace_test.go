package logger

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestTraceFields(t *testing.T) {
	traceID := trace.TraceID{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10}
	spanID := trace.SpanID{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	}))

	fields := TraceFields(ctx)
	if len(fields) != 2 {
		t.Fatalf("expected two trace fields, got %d", len(fields))
	}
	if fields[0].Key != "trace_id" || fields[0].String != traceID.String() {
		t.Fatalf("unexpected trace_id field: %#v", fields[0])
	}
	if fields[1].Key != "span_id" || fields[1].String != spanID.String() {
		t.Fatalf("unexpected span_id field: %#v", fields[1])
	}
}

func TestTraceFieldsReturnsNilForInvalidContext(t *testing.T) {
	if fields := TraceFields(context.Background()); fields != nil {
		t.Fatalf("expected no trace fields, got %#v", fields)
	}
}
