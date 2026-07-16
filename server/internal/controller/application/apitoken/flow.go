package apitoken

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type flow struct {
	ctx      context.Context
	span     trace.Span
	recorder EventRecorder
	userID   string
}

func startFlow(ctx context.Context, recorder EventRecorder, operation, userID string) (context.Context, *flow) {
	ctx, span := tracer.Start(ctx, operation)
	return ctx, &flow{ctx: ctx, span: span, recorder: recorder, userID: userID}
}

func (f *flow) end() { f.span.End() }

func (f *flow) record(name, outcome, reason, tokenID string, err error) {
	if f.recorder != nil {
		f.recorder.RecordAPITokenEvent(f.ctx, Event{Name: name, Outcome: outcome, Reason: reason, UserID: f.userID, TokenID: tokenID, Err: err})
	}
}
