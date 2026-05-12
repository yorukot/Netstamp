package logger

import (
	"context"

	"go.uber.org/zap"
)

type applicationEventLog struct {
	name            string
	category        string
	action          string
	outcome         string
	reason          string
	successOutcome  string
	expectedFailure bool
	fields          []zap.Field
	err             error
}

func recordApplicationEvent(ctx context.Context, root *zap.Logger, event applicationEventLog) {
	log := FromContext(ctx, root)
	fields := eventFields(event.name, event.category, event.action, event.outcome, event.reason)
	fields = append(fields, event.fields...)
	if event.err != nil {
		fields = append(fields, zap.Error(event.err))
	}

	switch {
	case event.outcome == event.successOutcome:
		log.Info(event.name, fields...)
	case event.expectedFailure:
		log.Warn(event.name, fields...)
	default:
		log.Error(event.name, fields...)
	}
}

func appendStringField(fields []zap.Field, key, value string) []zap.Field {
	if value == "" {
		return fields
	}

	return append(fields, zap.String(key, value))
}
