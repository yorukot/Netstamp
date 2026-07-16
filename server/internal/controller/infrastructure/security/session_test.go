package security

import (
	"context"
	"testing"
	"time"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestSessionManagerPersistsUserAgentAndScopesManagedSessions(t *testing.T) {
	now := time.Date(2026, time.July, 10, 9, 30, 0, 0, time.UTC)
	repo := &sessionRepositoryRecorder{
		listed: []identity.AuthSession{{ID: "22222222-2222-2222-2222-222222222222"}},
	}
	manager := NewSessionManager(repo, SessionConfig{
		HashKey:       "test-session-hash-key",
		IdleTTL:       time.Hour,
		AbsoluteTTL:   24 * time.Hour,
		TouchInterval: time.Minute,
	})
	manager.now = func() time.Time { return now }

	created, err := manager.CreateSession(context.Background(), createSessionInput("Browser/1.0", now))
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if created.Session.UserAgent != "Browser/1.0" || repo.created.UserAgent != "Browser/1.0" {
		t.Fatalf("expected persisted user agent, got created=%q stored=%q", created.Session.UserAgent, repo.created.UserAgent)
	}

	sessions, err := manager.ListUserSessions(context.Background(), "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(sessions) != 1 || repo.listUserID != "11111111-1111-1111-1111-111111111111" || !repo.listNow.Equal(now) {
		t.Fatalf("unexpected scoped list: sessions=%#v user=%q now=%s", sessions, repo.listUserID, repo.listNow)
	}

	if err := manager.RevokeUserSession(context.Background(), "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222", "user_revoked"); err != nil {
		t.Fatalf("revoke session: %v", err)
	}
	if repo.revokeUserID != "11111111-1111-1111-1111-111111111111" || repo.revokeSessionID != "22222222-2222-2222-2222-222222222222" || repo.revokeReason != "user_revoked" {
		t.Fatalf("unexpected scoped revoke: user=%q session=%q reason=%q", repo.revokeUserID, repo.revokeSessionID, repo.revokeReason)
	}
}

func TestSessionManagerSudoExpiresAfterFiveMinutes(t *testing.T) {
	authenticatedAt := time.Date(2026, time.July, 10, 9, 30, 0, 0, time.UTC)
	repo := &sessionRepositoryRecorder{active: identity.AuthSession{
		ID:              "22222222-2222-2222-2222-222222222222",
		UserID:          "11111111-1111-1111-1111-111111111111",
		AuthenticatedAt: authenticatedAt,
		SudoEligible:    true,
	}}
	manager := NewSessionManager(repo, SessionConfig{
		HashKey: "test-session-hash-key",
		SudoTTL: 5 * time.Minute,
	})

	manager.now = func() time.Time { return authenticatedAt.Add(5*time.Minute - time.Nanosecond) }
	status, err := manager.SudoStatus(context.Background(), repo.active.ID)
	if err != nil {
		t.Fatalf("get active sudo status: %v", err)
	}
	if !status.Active {
		t.Fatal("expected recent authentication to remain active before the five-minute boundary")
	}
	if want := authenticatedAt.Add(5 * time.Minute); !status.ExpiresAt.Equal(want) {
		t.Fatalf("expected sudo expiry %s, got %s", want, status.ExpiresAt)
	}

	manager.now = func() time.Time { return authenticatedAt.Add(5 * time.Minute) }
	status, err = manager.SudoStatus(context.Background(), repo.active.ID)
	if err != nil {
		t.Fatalf("get expired sudo status: %v", err)
	}
	if status.Active {
		t.Fatal("expected recent authentication to expire at the five-minute boundary")
	}
}

func TestSessionManagerElevateSessionPersistsAuthenticationMethod(t *testing.T) {
	repo := &sessionRepositoryRecorder{}
	manager := NewSessionManager(repo, SessionConfig{HashKey: "test-session-hash-key", SudoTTL: 5 * time.Minute})
	authenticatedAt := time.Date(2026, time.July, 10, 9, 30, 0, 0, time.UTC)
	identityID := "33333333-3333-3333-3333-333333333333"

	if err := manager.ElevateSession(context.Background(), "22222222-2222-2222-2222-222222222222", identity.AuthenticationMethodOIDC, &identityID, authenticatedAt); err != nil {
		t.Fatalf("elevate session: %v", err)
	}
	if repo.updatedSessionID != "22222222-2222-2222-2222-222222222222" || repo.updatedMethod != identity.AuthenticationMethodOIDC {
		t.Fatalf("unexpected authentication update: session=%q method=%q", repo.updatedSessionID, repo.updatedMethod)
	}
	if repo.updatedIdentityID == nil || *repo.updatedIdentityID != identityID || !repo.updatedAuthenticatedAt.Equal(authenticatedAt) {
		t.Fatalf("unexpected authentication metadata: identity=%v authenticatedAt=%s", repo.updatedIdentityID, repo.updatedAuthenticatedAt)
	}
}

func createSessionInput(userAgent string, now time.Time) appauth.CreateSessionInput {
	return appauth.CreateSessionInput{
		UserID:    "11111111-1111-1111-1111-111111111111",
		UserAgent: userAgent,
		Now:       now,
	}
}

type sessionRepositoryRecorder struct {
	created                identity.AuthSession
	active                 identity.AuthSession
	listed                 []identity.AuthSession
	listUserID             string
	listNow                time.Time
	revokeUserID           string
	revokeSessionID        string
	revokeReason           string
	updatedSessionID       string
	updatedAuthenticatedAt time.Time
	updatedMethod          string
	updatedIdentityID      *string
}

func (r *sessionRepositoryRecorder) CreateSession(_ context.Context, input identity.AuthSession) (identity.AuthSession, error) {
	r.created = input
	input.ID = "22222222-2222-2222-2222-222222222222"
	return input, nil
}

func (*sessionRepositoryRecorder) GetActiveSessionByTokenHash(context.Context, []byte, time.Time) (identity.AuthSession, error) {
	return identity.AuthSession{}, identity.ErrSessionNotFound
}

func (r *sessionRepositoryRecorder) GetActiveSessionByID(_ context.Context, sessionID string, _ time.Time) (identity.AuthSession, error) {
	if r.active.ID == "" || r.active.ID != sessionID {
		return identity.AuthSession{}, identity.ErrSessionNotFound
	}
	return r.active, nil
}

func (r *sessionRepositoryRecorder) UpdateSessionAuthentication(_ context.Context, sessionID string, authenticatedAt time.Time, method string, identityID *string) error {
	r.updatedSessionID = sessionID
	r.updatedAuthenticatedAt = authenticatedAt
	r.updatedMethod = method
	r.updatedIdentityID = identityID
	return nil
}

func (*sessionRepositoryRecorder) UpdateCSRFTokenHash(context.Context, string, []byte, time.Time) error {
	return nil
}

func (*sessionRepositoryRecorder) TouchSession(context.Context, string, time.Time, time.Time) error {
	return nil
}

func (*sessionRepositoryRecorder) RevokeSessionByTokenHash(context.Context, []byte, time.Time, string) error {
	return nil
}

func (r *sessionRepositoryRecorder) ListActiveSessionsForUser(_ context.Context, userID string, now time.Time) ([]identity.AuthSession, error) {
	r.listUserID = userID
	r.listNow = now
	return r.listed, nil
}

func (r *sessionRepositoryRecorder) RevokeSessionByIDForUser(_ context.Context, userID, sessionID string, _ time.Time, reason string) error {
	r.revokeUserID = userID
	r.revokeSessionID = sessionID
	r.revokeReason = reason
	return nil
}

func (*sessionRepositoryRecorder) RevokeSessionsForUser(context.Context, string, time.Time, string) error {
	return nil
}

var _ SessionRepository = (*sessionRepositoryRecorder)(nil)
