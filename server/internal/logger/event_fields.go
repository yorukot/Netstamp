package logger

import "go.uber.org/zap"

func eventFields(name, category, action, outcome, reason string) []zap.Field {
	fields := []zap.Field{
		zap.String("event_name", name),
		zap.String("event.category", category),
		zap.String("event.action", action),
		zap.String("event.outcome", outcome),
	}

	if reason != "" {
		fields = append(fields, zap.String("event.reason", reason))
	}

	return fields
}
