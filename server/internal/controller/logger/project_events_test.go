package logger

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	appproject "github.com/yorukot/netstamp/internal/controller/application/project"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

func TestProjectEventRecorderLogsStructuredEvent(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	root := zap.New(core).With(
		zap.String("request_id", "req-1"),
		zap.String("client.address", "203.0.113.10"),
	)
	recorder := NewProjectEventRecorder(root)

	recorder.RecordProjectEvent(context.Background(), appproject.ProjectEvent{
		Name:         appproject.ProjectEventAddMemberSuccess,
		Action:       appproject.ProjectActionAddMember,
		Outcome:      appproject.ProjectOutcomeSuccess,
		ActorUserID:  "actor-user",
		ProjectID:    "project-1",
		ProjectRef:   "engineering",
		ProjectSlug:  "engineering",
		TargetUserID: "target-user",
		Role:         domainproject.RoleViewer,
	})

	logs := observed.All()
	if len(logs) != 1 {
		t.Fatalf("expected one log entry, got %d", len(logs))
	}

	entry := logs[0]
	if entry.Level != zapcore.InfoLevel {
		t.Fatalf("expected info level, got %s", entry.Level)
	}
	if entry.Message != string(appproject.ProjectEventAddMemberSuccess) {
		t.Fatalf("expected project event message, got %q", entry.Message)
	}

	fields := entry.ContextMap()
	assertField(t, fields, "event_name", string(appproject.ProjectEventAddMemberSuccess))
	assertField(t, fields, "event.category", "project")
	assertField(t, fields, "event.action", string(appproject.ProjectActionAddMember))
	assertField(t, fields, "event.outcome", string(appproject.ProjectOutcomeSuccess))
	assertField(t, fields, "user.id", "actor-user")
	assertField(t, fields, "project.id", "project-1")
	assertField(t, fields, "project.ref", "engineering")
	assertField(t, fields, "project.slug", "engineering")
	assertField(t, fields, "project.member.user.id", "target-user")
	assertField(t, fields, "project.member.role", string(domainproject.RoleViewer))
	assertField(t, fields, "request_id", "req-1")
	assertField(t, fields, "client.address", "203.0.113.10")

	for _, forbidden := range []string{"project.member.email", "user.email", "password", "access_token"} {
		if _, ok := fields[forbidden]; ok {
			t.Fatalf("forbidden field %q was logged: %#v", forbidden, fields[forbidden])
		}
	}
}

func TestProjectEventRecorderLevels(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	recorder := NewProjectEventRecorder(zap.New(core))
	technicalErr := errors.New("update member role")

	recorder.RecordProjectEvent(context.Background(), appproject.ProjectEvent{
		Name:    appproject.ProjectEventUpdateMemberRoleFailure,
		Action:  appproject.ProjectActionUpdateMemberRole,
		Outcome: appproject.ProjectOutcomeFailure,
		Reason:  appproject.ProjectReasonForbidden,
	})
	recorder.RecordProjectEvent(context.Background(), appproject.ProjectEvent{
		Name:    appproject.ProjectEventUpdateMemberRoleFailure,
		Action:  appproject.ProjectActionUpdateMemberRole,
		Outcome: appproject.ProjectOutcomeFailure,
		Reason:  appproject.ProjectReasonMemberRoleUpdateFailed,
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
	assertField(t, logs[1].ContextMap(), "error", "update member role")
}
