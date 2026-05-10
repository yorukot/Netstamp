package logger

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
)

func TestAuthEventRecorderLogsPseudonymizedAuthEvent(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	root := zap.New(core).With(
		zap.String("request_id", "req-1"),
		zap.String("client.address", "203.0.113.10"),
	)
	recorder := NewAuthEventRecorder(root, "test-pseudonym-key")

	recorder.RecordAuthEvent(context.Background(), appauth.AuthEvent{
		Name:    appauth.AuthEventLoginSuccess,
		Action:  appauth.AuthActionLogin,
		Outcome: appauth.AuthOutcomeSuccess,
		UserID:  "user-1",
		Email:   " User@Example.COM ",
	})

	logs := observed.All()
	if len(logs) != 1 {
		t.Fatalf("expected one log entry, got %d", len(logs))
	}

	entry := logs[0]
	if entry.Level != zapcore.InfoLevel {
		t.Fatalf("expected info level, got %s", entry.Level)
	}
	if entry.Message != string(appauth.AuthEventLoginSuccess) {
		t.Fatalf("expected auth event message, got %q", entry.Message)
	}

	fields := entry.ContextMap()
	assertField(t, fields, "event_name", string(appauth.AuthEventLoginSuccess))
	assertField(t, fields, "event.category", "auth")
	assertField(t, fields, "event.action", string(appauth.AuthActionLogin))
	assertField(t, fields, "event.outcome", string(appauth.AuthOutcomeSuccess))
	assertField(t, fields, "user.id", "user-1")
	assertField(t, fields, "request_id", "req-1")
	assertField(t, fields, "client.address", "203.0.113.10")

	emailHash, ok := fields["user.email_hash"].(string)
	if !ok || emailHash == "" {
		t.Fatalf("expected user.email_hash field, got %#v", fields["user.email_hash"])
	}
	if emailHash == "user@example.com" || emailHash == " User@Example.COM " {
		t.Fatalf("email hash leaked raw email: %q", emailHash)
	}
	for _, forbidden := range []string{"user.email", "password", "password_hash", "access_token"} {
		if _, ok := fields[forbidden]; ok {
			t.Fatalf("forbidden field %q was logged: %#v", forbidden, fields[forbidden])
		}
	}
}

func TestAuthEventRecorderLevels(t *testing.T) {
	core, observed := observer.New(zapcore.DebugLevel)
	recorder := NewAuthEventRecorder(zap.New(core), "test-pseudonym-key")

	recorder.RecordAuthEvent(context.Background(), appauth.AuthEvent{
		Name:    appauth.AuthEventLoginFailure,
		Action:  appauth.AuthActionLogin,
		Outcome: appauth.AuthOutcomeFailure,
		Reason:  appauth.AuthReasonCredentialsInvalid,
		Email:   "user@example.com",
	})
	recorder.RecordAuthEvent(context.Background(), appauth.AuthEvent{
		Name:    appauth.AuthEventLoginFailure,
		Action:  appauth.AuthActionLogin,
		Outcome: appauth.AuthOutcomeFailure,
		Reason:  appauth.AuthReasonUserInactive,
		Email:   "user@example.com",
	})
	recorder.RecordAuthEvent(context.Background(), appauth.AuthEvent{
		Name:    appauth.AuthEventTokenIssueFailure,
		Action:  appauth.AuthActionLogin,
		Outcome: appauth.AuthOutcomeFailure,
		Reason:  appauth.AuthReasonAccessTokenIssueFail,
		Email:   "user@example.com",
		Err:     errors.New("sign token"),
	})

	logs := observed.All()
	if len(logs) != 3 {
		t.Fatalf("expected three log entries, got %d", len(logs))
	}
	if logs[0].Level != zapcore.WarnLevel {
		t.Fatalf("expected credentials failure to be warn, got %s", logs[0].Level)
	}
	if logs[1].Level != zapcore.WarnLevel {
		t.Fatalf("expected inactive user failure to be warn, got %s", logs[1].Level)
	}
	if logs[2].Level != zapcore.ErrorLevel {
		t.Fatalf("expected token issue failure to be error, got %s", logs[2].Level)
	}
	assertField(t, logs[2].ContextMap(), "error", "sign token")
}

func assertField(t *testing.T, fields map[string]any, key string, want any) {
	t.Helper()

	if got, ok := fields[key]; !ok || got != want {
		t.Fatalf("expected field %q=%#v, got %#v", key, want, got)
	}
}
