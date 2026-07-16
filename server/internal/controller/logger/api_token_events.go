package logger

import (
	"context"

	"go.uber.org/zap"

	appapitoken "github.com/yorukot/netstamp/internal/controller/application/apitoken"
)

type APITokenEventRecorder struct{ root *zap.Logger }

func NewAPITokenEventRecorder(root *zap.Logger) *APITokenEventRecorder {
	if root == nil {
		root = zap.NewNop()
	}
	return &APITokenEventRecorder{root: root}
}

func (r *APITokenEventRecorder) RecordAPITokenEvent(ctx context.Context, event appapitoken.Event) {
	fields := eventFields(event.Name, "api_token", event.Name, event.Outcome, event.Reason)
	if event.UserID != "" {
		fields = append(fields, zap.String("user.id", event.UserID))
	}
	if event.TokenID != "" {
		fields = append(fields, zap.String("api_token.id", event.TokenID))
	}
	if event.Err != nil {
		fields = append(fields, zap.Error(event.Err))
	}
	log := FromContext(ctx, r.root)
	if event.Outcome == "success" {
		log.Info(event.Name, fields...)
		return
	}
	if event.Reason == "invalid_input" || event.Reason == "credentials_invalid" {
		log.Warn(event.Name, fields...)
		return
	}
	log.Error(event.Name, fields...)
}
