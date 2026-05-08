package logger

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	applabel "github.com/yorukot/netstamp/internal/application/label"
)

func TestLabelEventRecorderLogsStructuredEvent(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	root := zap.New(core).With(
		zap.String("request_id", "req-1"),
		zap.String("client.address", "203.0.113.10"),
	)
	recorder := NewLabelEventRecorder(root)

	recorder.RecordLabelEvent(context.Background(), applabel.LabelEvent{
		Name:        applabel.LabelEventCreateSuccess,
		Action:      applabel.LabelActionCreate,
		Outcome:     applabel.LabelOutcomeSuccess,
		ActorUserID: "user-1",
		ProjectID:   "project-1",
		ProjectRef:  "engineering",
		ProjectSlug: "engineering",
		LabelID:     "label-1",
	})

	logs := observed.All()
	if len(logs) != 1 {
		t.Fatalf("expected one log entry, got %d", len(logs))
	}

	entry := logs[0]
	if entry.Level != zapcore.InfoLevel {
		t.Fatalf("expected info level, got %s", entry.Level)
	}
	if entry.Message != string(applabel.LabelEventCreateSuccess) {
		t.Fatalf("expected label event message, got %q", entry.Message)
	}

	fields := entry.ContextMap()
	assertField(t, fields, "event_name", string(applabel.LabelEventCreateSuccess))
	assertField(t, fields, "event.category", "label")
	assertField(t, fields, "event.action", string(applabel.LabelActionCreate))
	assertField(t, fields, "event.outcome", string(applabel.LabelOutcomeSuccess))
	assertField(t, fields, "user.id", "user-1")
	assertField(t, fields, "project.id", "project-1")
	assertField(t, fields, "project.ref", "engineering")
	assertField(t, fields, "project.slug", "engineering")
	assertField(t, fields, "label.id", "label-1")
	assertField(t, fields, "request_id", "req-1")
	assertField(t, fields, "client.address", "203.0.113.10")

	for _, forbidden := range []string{"label.key", "label.value", "user.email", "password", "access_token"} {
		if _, ok := fields[forbidden]; ok {
			t.Fatalf("forbidden field %q was logged: %#v", forbidden, fields[forbidden])
		}
	}
}

func TestLabelEventRecorderLevels(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	recorder := NewLabelEventRecorder(zap.New(core))
	technicalErr := errors.New("create label")

	recorder.RecordLabelEvent(context.Background(), applabel.LabelEvent{
		Name:    applabel.LabelEventCreateFailure,
		Action:  applabel.LabelActionCreate,
		Outcome: applabel.LabelOutcomeFailure,
		Reason:  applabel.LabelReasonForbidden,
	})
	recorder.RecordLabelEvent(context.Background(), applabel.LabelEvent{
		Name:    applabel.LabelEventCreateFailure,
		Action:  applabel.LabelActionCreate,
		Outcome: applabel.LabelOutcomeFailure,
		Reason:  applabel.LabelReasonLabelCreateFailed,
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
	assertField(t, logs[1].ContextMap(), "error", "create label")
}
