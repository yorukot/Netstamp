package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestRequireAuthRejectsMissingAuthCookie(t *testing.T) {
	verifier := &recordingTokenVerifier{}
	router := registerClaimsRoute(t, verifier)

	res := performAuthTestRequest(router, http.MethodGet, "/me")

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
	if verifier.gotToken != "" {
		t.Fatalf("expected verifier not to be called, got token %q", verifier.gotToken)
	}
	if got := res.Header().Get("WWW-Authenticate"); got != "" {
		t.Fatalf("expected empty WWW-Authenticate, got %q", got)
	}
}

func TestRequireAuthIgnoresBearerHeader(t *testing.T) {
	verifier := &recordingTokenVerifier{}
	router := registerClaimsRoute(t, verifier)

	res := performAuthTestRequest(router, http.MethodGet, "/me", "Authorization", "Bearer ignored-token")

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
	if verifier.gotToken != "" {
		t.Fatalf("expected verifier not to be called, got token %q", verifier.gotToken)
	}
}

func TestRequireAuthRejectsInvalidSession(t *testing.T) {
	verifier := &recordingTokenVerifier{err: appauth.ErrSessionInvalid}
	router := registerClaimsRoute(t, verifier)

	res := performAuthTestRequest(router, http.MethodGet, "/me", "Cookie", LocalSessionCookieName+"=bad-token")

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
	if verifier.gotToken != "bad-token" {
		t.Fatalf("expected verifier token %q, got %q", "bad-token", verifier.gotToken)
	}
}

func TestRequireAuthStoresClaimsInContext(t *testing.T) {
	verifier := &recordingTokenVerifier{
		claims: identity.SessionClaims{
			SessionID: "session-1",
			UserID:    "user-1",
		},
	}
	router := registerClaimsRoute(t, verifier)

	res := performAuthTestRequest(router, http.MethodGet, "/me", "Cookie", LocalSessionCookieName+"=good-token")

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	if verifier.gotToken != "good-token" {
		t.Fatalf("expected verifier token %q, got %q", "good-token", verifier.gotToken)
	}
}

func registerClaimsRoute(t *testing.T, verifier appauth.SessionManager) http.Handler {
	t.Helper()

	router := chi.NewRouter()
	router.With(RequireAuth(verifier, LocalSessionCookieName)).Get("/me", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := SessionClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "missing claims", http.StatusInternalServerError)
			return
		}
		if claims.SessionID == "" || claims.UserID == "" {
			http.Error(w, "empty claims", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(claimsRouteBody{
			SessionID: claims.SessionID,
			UserID:    claims.UserID,
		})
	})
	return router
}

func performAuthTestRequest(router http.Handler, method, path string, headers ...string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, http.NoBody)
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

type claimsRouteBody struct {
	SessionID string `json:"sessionId"`
	UserID    string `json:"userId"`
}

type recordingTokenVerifier struct {
	claims   identity.SessionClaims
	err      error
	gotToken string
}

func (v *recordingTokenVerifier) VerifySession(_ context.Context, value string) (identity.SessionClaims, error) {
	v.gotToken = value
	if v.err != nil {
		return identity.SessionClaims{}, v.err
	}
	return v.claims, nil
}

func (v *recordingTokenVerifier) CreateSession(context.Context, appauth.CreateSessionInput) (identity.CreatedSession, error) {
	return identity.CreatedSession{}, nil
}

func (v *recordingTokenVerifier) CreateCSRFToken(context.Context, string) (string, error) {
	return "", nil
}

func (v *recordingTokenVerifier) VerifyCSRFToken(context.Context, string, string) error {
	return nil
}

func (v *recordingTokenVerifier) RevokeSession(context.Context, string, string) error {
	return nil
}

func (v *recordingTokenVerifier) RevokeUserSessions(context.Context, string, string) error {
	return nil
}
