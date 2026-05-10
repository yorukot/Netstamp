package logger

import (
	"context"

	"go.uber.org/zap"
)

type contextKey struct{}

func WithContext(ctx context.Context, log *zap.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, log)
}

func FromContext(ctx context.Context, fallback *zap.Logger) *zap.Logger {
	if log, ok := ctx.Value(contextKey{}).(*zap.Logger); ok && log != nil {
		return log
	}
	return fallback
}
