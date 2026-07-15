package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestListSessionsMarksCurrentSession(t *testing.T) {
	now := time.Now().UTC()
	manager := &sessionRegistryTestManager{
		claims: identity.SessionClaims{SessionID: "22222222-2222-2222-2222-222222222222", UserID: "11111111-1111-1111-1111-111111111111"},
		sessions: []identity.AuthSession{
			{ID: "22222222-2222-2222-2222-222222222222", UserAgent: "Current Browser", CreatedAt: now.Add(-time.Hour), LastUsedAt: now, IdleExpiresAt: now.Add(time.Hour), AbsoluteExpiresAt: now.Add(24 * time.Hour)},
			{ID: "33333333-3333-3333-3333-333333333333", UserAgent: "Other Browser", CreatedAt: now.Add(-2 * time.Hour), LastUsedAt: now.Add(-time.Hour), IdleExpiresAt: now.Add(time.Hour), AbsoluteExpiresAt: now.Add(24 * time.Hour)},
		},
	}
	router := sessionRegistryTestRouter(manager)

	res := performSessionRegistryRequest(router, http.MethodGet, "/auth/sessions")

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	var body listSessionsOutputBody
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Sessions) != 2 || !body.Sessions[0].IsCurrent || body.Sessions[1].IsCurrent {
		t.Fatalf("expected only first session to be current, got %#v", body.Sessions)
	}
	if manager.listUserID != manager.claims.UserID {
		t.Fatalf("expected list for authenticated user %q, got %q", manager.claims.UserID, manager.listUserID)
	}
}

func TestRevokeSessionUsesAuthenticatedOwnerAndManagesCurrentCookie(t *testing.T) {
	tests := []struct {
		name              string
		sessionID         string
		expectClearCookie bool
	}{
		{name: "current session", sessionID: "22222222-2222-2222-2222-222222222222", expectClearCookie: true},
		{name: "other session", sessionID: "33333333-3333-3333-3333-333333333333", expectClearCookie: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &sessionRegistryTestManager{claims: identity.SessionClaims{
				SessionID: "22222222-2222-2222-2222-222222222222",
				UserID:    "11111111-1111-1111-1111-111111111111",
			}}
			router := sessionRegistryTestRouter(manager)

			res := performSessionRegistryRequest(router, http.MethodDelete, "/auth/sessions/"+tt.sessionID)

			if res.Code != http.StatusNoContent {
				t.Fatalf("expected status 204, got %d", res.Code)
			}
			if manager.revokeUserID != manager.claims.UserID || manager.revokeSessionID != tt.sessionID {
				t.Fatalf("unexpected revoke arguments: user=%q session=%q", manager.revokeUserID, manager.revokeSessionID)
			}
			cookies := res.Result().Cookies()
			if tt.expectClearCookie {
				if len(cookies) != 1 || cookies[0].Name != "netstamp_session" || cookies[0].MaxAge != -1 {
					t.Fatalf("expected expired current-session cookie, got %#v", cookies)
				}
			} else if len(cookies) != 0 {
				t.Fatalf("expected other-session revoke to preserve cookie, got %#v", cookies)
			}
		})
	}
}

func TestRevokeSessionReturnsNotFound(t *testing.T) {
	manager := &sessionRegistryTestManager{
		claims:    identity.SessionClaims{SessionID: "22222222-2222-2222-2222-222222222222", UserID: "11111111-1111-1111-1111-111111111111"},
		revokeErr: identity.ErrSessionNotFound,
	}
	res := performSessionRegistryRequest(sessionRegistryTestRouter(manager), http.MethodDelete, "/auth/sessions/33333333-3333-3333-3333-333333333333")

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", res.Code)
	}
	var body httpx.ProblemDetails
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != httpx.CodeAuthSessionNotFound {
		t.Fatalf("expected code %q, got %q", httpx.CodeAuthSessionNotFound, body.Code)
	}
}

func TestRevokeAllSessionsUsesAuthenticatedUserAndClearsCookie(t *testing.T) {
	manager := &sessionRegistryTestManager{claims: identity.SessionClaims{
		SessionID: "22222222-2222-2222-2222-222222222222",
		UserID:    "11111111-1111-1111-1111-111111111111",
	}}

	res := performSessionRegistryRequest(sessionRegistryTestRouter(manager), http.MethodDelete, "/auth/sessions")

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", res.Code)
	}
	if manager.revokeAllUserID != manager.claims.UserID || manager.revokeAllReason != "user_logout_all" {
		t.Fatalf("unexpected revoke-all arguments: user=%q reason=%q", manager.revokeAllUserID, manager.revokeAllReason)
	}
	cookies := res.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "netstamp_session" || cookies[0].MaxAge != -1 {
		t.Fatalf("expected expired current-session cookie, got %#v", cookies)
	}
}

func TestRevokeAllSessionsReturnsInternalError(t *testing.T) {
	manager := &sessionRegistryTestManager{
		claims:       identity.SessionClaims{SessionID: "22222222-2222-2222-2222-222222222222", UserID: "11111111-1111-1111-1111-111111111111"},
		revokeAllErr: errors.New("database unavailable"),
	}

	res := performSessionRegistryRequest(sessionRegistryTestRouter(manager), http.MethodDelete, "/auth/sessions")

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", res.Code)
	}
	if len(res.Result().Cookies()) != 0 {
		t.Fatalf("expected failed revoke-all to preserve cookie, got %#v", res.Result().Cookies())
	}
}

func sessionRegistryTestRouter(manager *sessionRegistryTestManager) http.Handler {
	router := chi.NewRouter()
	service := appauth.NewService(nil, nil, manager, nil)
	NewHandler(service, manager, nil, "netstamp_session", true, true).RegisterRoutes(router)
	return router
}

func performSessionRegistryRequest(router http.Handler, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, http.NoBody)
	req.AddCookie(&http.Cookie{Name: "netstamp_session", Value: "valid-token"})
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

type sessionRegistryTestManager struct {
	claims          identity.SessionClaims
	sessions        []identity.AuthSession
	listUserID      string
	revokeUserID    string
	revokeSessionID string
	revokeErr       error
	revokeAllUserID string
	revokeAllReason string
	revokeAllErr    error
}

func (m *sessionRegistryTestManager) CreateSession(context.Context, appauth.CreateSessionInput) (identity.CreatedSession, error) {
	return identity.CreatedSession{}, nil
}

func (m *sessionRegistryTestManager) VerifySession(context.Context, string) (identity.SessionClaims, error) {
	return m.claims, nil
}

func (m *sessionRegistryTestManager) CreateCSRFToken(context.Context, string) (string, error) {
	return "", nil
}

func (m *sessionRegistryTestManager) VerifyCSRFToken(context.Context, string, string) error {
	return nil
}

func (m *sessionRegistryTestManager) RevokeSession(context.Context, string, string) error {
	return nil
}

func (m *sessionRegistryTestManager) ListUserSessions(_ context.Context, userID string) ([]identity.AuthSession, error) {
	m.listUserID = userID
	return m.sessions, nil
}

func (m *sessionRegistryTestManager) RevokeUserSession(_ context.Context, userID, sessionID, _ string) error {
	m.revokeUserID = userID
	m.revokeSessionID = sessionID
	return m.revokeErr
}

func (m *sessionRegistryTestManager) RevokeUserSessions(_ context.Context, userID, reason string) error {
	m.revokeAllUserID = userID
	m.revokeAllReason = reason
	return m.revokeAllErr
}

var _ appauth.SessionManager = (*sessionRegistryTestManager)(nil)
