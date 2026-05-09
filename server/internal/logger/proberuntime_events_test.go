package logger

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	appproberuntime "github.com/yorukot/netstamp/internal/application/proberuntime"
)

func TestProbeRuntimeEventRecorderLogsStructuredEvent(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	root := zap.New(core).With(
		zap.String("request_id", "req-1"),
		zap.String("client.address", "203.0.113.10"),
	)
	recorder := NewProbeRuntimeEventRecorder(root)

	recorder.RecordProbeRuntimeEvent(context.Background(), appproberuntime.ProbeRuntimeEvent{
		Name:        appproberuntime.ProbeRuntimeEventSubmitResultsFailure,
		Action:      appproberuntime.ProbeRuntimeActionSubmitResults,
		Outcome:     appproberuntime.ProbeRuntimeOutcomeFailure,
		Reason:      appproberuntime.ProbeRuntimeReasonInvalidResult,
		ProbeID:     "probe-1",
		ProjectID:   "project-1",
		ResultCount: probeRuntimeIntPtr(1),
	})

	logs := observed.All()
	if len(logs) != 1 {
		t.Fatalf("expected one log entry, got %d", len(logs))
	}

	entry := logs[0]
	if entry.Level != zapcore.WarnLevel {
		t.Fatalf("expected warn level, got %s", entry.Level)
	}
	if entry.Message != string(appproberuntime.ProbeRuntimeEventSubmitResultsFailure) {
		t.Fatalf("expected probe runtime event message, got %q", entry.Message)
	}

	fields := entry.ContextMap()
	assertField(t, fields, "event_name", string(appproberuntime.ProbeRuntimeEventSubmitResultsFailure))
	assertField(t, fields, "event.category", "probe_runtime")
	assertField(t, fields, "event.action", string(appproberuntime.ProbeRuntimeActionSubmitResults))
	assertField(t, fields, "event.outcome", string(appproberuntime.ProbeRuntimeOutcomeFailure))
	assertField(t, fields, "event.reason", string(appproberuntime.ProbeRuntimeReasonInvalidResult))
	assertField(t, fields, "probe.id", "probe-1")
	assertField(t, fields, "project.id", "project-1")
	assertField(t, fields, "result.count", int64(1))
	assertField(t, fields, "request_id", "req-1")
	assertField(t, fields, "client.address", "203.0.113.10")

	for _, forbidden := range []string{
		"probe.secret",
		"probe.secret_hash",
		"secret",
		"secret_hash",
		"password",
		"access_token",
		"agent.version",
		"network.public_v4",
		"network.public_v6",
		"result.raw",
		"result.error_message",
	} {
		if _, ok := fields[forbidden]; ok {
			t.Fatalf("forbidden field %q was logged: %#v", forbidden, fields[forbidden])
		}
	}
}

func TestProbeRuntimeEventRecorderLevels(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	recorder := NewProbeRuntimeEventRecorder(zap.New(core))
	technicalErr := errors.New("write results")

	recorder.RecordProbeRuntimeEvent(context.Background(), appproberuntime.ProbeRuntimeEvent{
		Name:    appproberuntime.ProbeRuntimeEventHelloFailure,
		Action:  appproberuntime.ProbeRuntimeActionHello,
		Outcome: appproberuntime.ProbeRuntimeOutcomeFailure,
		Reason:  appproberuntime.ProbeRuntimeReasonInvalidCredential,
	})
	recorder.RecordProbeRuntimeEvent(context.Background(), appproberuntime.ProbeRuntimeEvent{
		Name:    appproberuntime.ProbeRuntimeEventSubmitResultsFailure,
		Action:  appproberuntime.ProbeRuntimeActionSubmitResults,
		Outcome: appproberuntime.ProbeRuntimeOutcomeFailure,
		Reason:  appproberuntime.ProbeRuntimeReasonResultWriteFailed,
		Err:     technicalErr,
	})

	logs := observed.All()
	if len(logs) != 2 {
		t.Fatalf("expected two log entries, got %d", len(logs))
	}
	if logs[0].Level != zapcore.WarnLevel {
		t.Fatalf("expected business failure to be warn, got %s", logs[0].Level)
	}
	if logs[1].Level != zapcore.ErrorLevel {
		t.Fatalf("expected technical failure to be error, got %s", logs[1].Level)
	}
	assertField(t, logs[1].ContextMap(), "error", "write results")
}

func probeRuntimeIntPtr(value int) *int {
	return &value
}
