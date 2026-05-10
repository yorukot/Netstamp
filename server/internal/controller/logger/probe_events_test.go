package logger

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	appprobe "github.com/yorukot/netstamp/internal/controller/application/proberegistry"
)

func TestProbeEventRecorderLogsStructuredEvent(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	root := zap.New(core).With(
		zap.String("request_id", "req-1"),
		zap.String("client.address", "203.0.113.10"),
	)
	recorder := NewProbeEventRecorder(root)

	recorder.RecordProbeEvent(context.Background(), appprobe.ProbeEvent{
		Name:        appprobe.ProbeEventCreateFailure,
		Action:      appprobe.ProbeActionCreate,
		Outcome:     appprobe.ProbeOutcomeFailure,
		Reason:      appprobe.ProbeReasonLabelNotFound,
		ActorUserID: "user-1",
		ProjectID:   "project-1",
		ProjectRef:  "engineering",
	})

	logs := observed.All()
	if len(logs) != 1 {
		t.Fatalf("expected one log entry, got %d", len(logs))
	}

	entry := logs[0]
	if entry.Level != zapcore.WarnLevel {
		t.Fatalf("expected warn level, got %s", entry.Level)
	}
	if entry.Message != string(appprobe.ProbeEventCreateFailure) {
		t.Fatalf("expected probe event message, got %q", entry.Message)
	}

	fields := entry.ContextMap()
	assertField(t, fields, "event_name", string(appprobe.ProbeEventCreateFailure))
	assertField(t, fields, "event.category", "probe")
	assertField(t, fields, "event.action", string(appprobe.ProbeActionCreate))
	assertField(t, fields, "event.outcome", string(appprobe.ProbeOutcomeFailure))
	assertField(t, fields, "event.reason", string(appprobe.ProbeReasonLabelNotFound))
	assertField(t, fields, "user.id", "user-1")
	assertField(t, fields, "project.id", "project-1")
	assertField(t, fields, "project.ref", "engineering")
	assertField(t, fields, "request_id", "req-1")
	assertField(t, fields, "client.address", "203.0.113.10")

	for _, forbidden := range []string{"probe.secret", "probe.secret_hash", "secret", "secret_hash", "password", "access_token"} {
		if _, ok := fields[forbidden]; ok {
			t.Fatalf("forbidden field %q was logged: %#v", forbidden, fields[forbidden])
		}
	}
}

func TestProbeEventRecorderLevels(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	recorder := NewProbeEventRecorder(zap.New(core))
	technicalErr := errors.New("create probe")

	recorder.RecordProbeEvent(context.Background(), appprobe.ProbeEvent{
		Name:    appprobe.ProbeEventCreateFailure,
		Action:  appprobe.ProbeActionCreate,
		Outcome: appprobe.ProbeOutcomeFailure,
		Reason:  appprobe.ProbeReasonForbidden,
	})
	recorder.RecordProbeEvent(context.Background(), appprobe.ProbeEvent{
		Name:    appprobe.ProbeEventCreateFailure,
		Action:  appprobe.ProbeActionCreate,
		Outcome: appprobe.ProbeOutcomeFailure,
		Reason:  appprobe.ProbeReasonProbeCreateFailed,
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
	assertField(t, logs[1].ContextMap(), "error", "create probe")
}
