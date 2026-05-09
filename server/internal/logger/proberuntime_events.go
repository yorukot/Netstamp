package logger

import (
	"context"

	"go.uber.org/zap"

	appproberuntime "github.com/yorukot/netstamp/internal/application/proberuntime"
)

type ProbeRuntimeEventRecorder struct {
	root *zap.Logger
}

func NewProbeRuntimeEventRecorder(root *zap.Logger) *ProbeRuntimeEventRecorder {
	if root == nil {
		root = zap.NewNop()
	}

	return &ProbeRuntimeEventRecorder{root: root}
}

func (r *ProbeRuntimeEventRecorder) RecordProbeRuntimeEvent(ctx context.Context, event appproberuntime.ProbeRuntimeEvent) {
	log := FromContext(ctx, r.root)
	fields := eventFields(string(event.Name), "probe_runtime", string(event.Action), string(event.Outcome), string(event.Reason))

	if event.ProbeID != "" {
		fields = append(fields, zap.String("probe.id", event.ProbeID))
	}
	if event.ProjectID != "" {
		fields = append(fields, zap.String("project.id", event.ProjectID))
	}
	if event.ResultCount != nil {
		fields = append(fields, zap.Int("result.count", *event.ResultCount))
	}
	if event.Err != nil {
		fields = append(fields, zap.Error(event.Err))
	}

	switch {
	case event.Outcome == appproberuntime.ProbeRuntimeOutcomeSuccess:
		log.Info(string(event.Name), fields...)
	case isExpectedProbeRuntimeFailure(event):
		log.Warn(string(event.Name), fields...)
	default:
		log.Error(string(event.Name), fields...)
	}
}

func isExpectedProbeRuntimeFailure(event appproberuntime.ProbeRuntimeEvent) bool {
	switch event.Reason {
	case appproberuntime.ProbeRuntimeReasonInvalidInput,
		appproberuntime.ProbeRuntimeReasonInvalidCredential,
		appproberuntime.ProbeRuntimeReasonProbeNotFound,
		appproberuntime.ProbeRuntimeReasonProbeDisabled,
		appproberuntime.ProbeRuntimeReasonInvalidResult,
		appproberuntime.ProbeRuntimeReasonResultConflict,
		appproberuntime.ProbeRuntimeReasonUnsupportedResult:
		return true
	default:
		return false
	}
}
