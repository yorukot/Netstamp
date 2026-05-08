package logger

import (
	"context"

	"go.uber.org/zap"

	appprobe "github.com/yorukot/netstamp/internal/application/probe"
)

type ProbeEventRecorder struct {
	root *zap.Logger
}

func NewProbeEventRecorder(root *zap.Logger) *ProbeEventRecorder {
	if root == nil {
		root = zap.NewNop()
	}

	return &ProbeEventRecorder{root: root}
}

func (r *ProbeEventRecorder) RecordProbeEvent(ctx context.Context, event appprobe.ProbeEvent) {
	log := FromContext(ctx, r.root)
	fields := []zap.Field{
		zap.String("event_name", string(event.Name)),
		zap.String("event.category", "probe"),
		zap.String("event.action", string(event.Action)),
		zap.String("event.outcome", string(event.Outcome)),
	}

	if event.Reason != "" {
		fields = append(fields, zap.String("event.reason", string(event.Reason)))
	}
	if event.ActorUserID != "" {
		fields = append(fields, zap.String("user.id", event.ActorUserID))
	}
	if event.ProjectID != "" {
		fields = append(fields, zap.String("project.id", event.ProjectID))
	}
	if event.ProjectRef != "" {
		fields = append(fields, zap.String("project.ref", event.ProjectRef))
	}
	if event.ProbeID != "" {
		fields = append(fields, zap.String("probe.id", event.ProbeID))
	}
	if event.Err != nil {
		fields = append(fields, zap.Error(event.Err))
	}

	switch {
	case event.Outcome == appprobe.ProbeOutcomeSuccess:
		log.Info(string(event.Name), fields...)
	case isExpectedProbeFailure(event):
		log.Warn(string(event.Name), fields...)
	default:
		log.Error(string(event.Name), fields...)
	}
}

func isExpectedProbeFailure(event appprobe.ProbeEvent) bool {
	switch event.Reason {
	case appprobe.ProbeReasonInvalidInput,
		appprobe.ProbeReasonProjectNotFound,
		appprobe.ProbeReasonLabelNotFound:
		return true
	default:
		return false
	}
}
