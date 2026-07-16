package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	appapitoken "github.com/yorukot/netstamp/internal/controller/application/apitoken"
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

func TestRequireUserAuthUsesBearerWithoutFallingBackToCookie(t *testing.T) {
	sessions := &recordingTokenVerifier{claims: identity.SessionClaims{SessionID: "session-1", UserID: "session-user"}}
	tokens := &recordingAPITokenVerifier{principal: appapitoken.Principal{
		TokenID: "token-1",
		UserID:  "token-user",
		Scopes:  []identity.APITokenScope{identity.ScopeProjectsRead},
	}}
	router := chi.NewRouter()
	router.With(RequireUserAuth(sessions, tokens, LocalSessionCookieName), RequireScope(identity.ScopeProjectsRead)).Get("/projects", func(w http.ResponseWriter, r *http.Request) {
		userID, ok := CurrentUserIDFromContext(r.Context())
		if !ok || userID != "token-user" {
			http.Error(w, "unexpected principal", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	res := performAuthTestRequest(router, http.MethodGet, "/projects",
		"Authorization", "bearer api-secret",
		"Cookie", LocalSessionCookieName+"=valid-session",
	)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", res.Code)
	}
	if tokens.gotToken != "api-secret" {
		t.Fatalf("expected API token verifier input, got %q", tokens.gotToken)
	}
	if sessions.gotToken != "" {
		t.Fatalf("Bearer authentication must not fall back to the session cookie, got %q", sessions.gotToken)
	}
}

func TestRequireUserAuthDoesNotFallBackAfterInvalidBearer(t *testing.T) {
	sessions := &recordingTokenVerifier{claims: identity.SessionClaims{SessionID: "session-1", UserID: "session-user"}}
	tokens := &recordingAPITokenVerifier{err: errors.New("invalid token")}
	router := chi.NewRouter()
	router.With(RequireUserAuth(sessions, tokens, LocalSessionCookieName)).Get("/projects", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	res := performAuthTestRequest(router, http.MethodGet, "/projects",
		"Authorization", "Bearer invalid",
		"Cookie", LocalSessionCookieName+"=valid-session",
	)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
	if got := res.Header().Get("WWW-Authenticate"); got != `Bearer realm="netstamp"` {
		t.Fatalf("unexpected WWW-Authenticate header %q", got)
	}
	if sessions.gotToken != "" {
		t.Fatalf("invalid Bearer credentials must not fall back to the session cookie, got %q", sessions.gotToken)
	}
}

func TestRequireScopeRejectsMissingTokenScopeButAllowsSession(t *testing.T) {
	t.Run("api token", func(t *testing.T) {
		tokens := &recordingAPITokenVerifier{principal: appapitoken.Principal{
			TokenID: "token-1",
			UserID:  "user-1",
			Scopes:  []identity.APITokenScope{identity.ScopeChecksRead},
		}}
		router := chi.NewRouter()
		router.With(RequireUserAuth(nil, tokens, LocalSessionCookieName), RequireScope(identity.ScopeProjectsRead)).Get("/projects", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		res := performAuthTestRequest(router, http.MethodGet, "/projects", "Authorization", "Bearer api-secret")
		if res.Code != http.StatusForbidden {
			t.Fatalf("expected status 403, got %d", res.Code)
		}
	})

	t.Run("session", func(t *testing.T) {
		sessions := &recordingTokenVerifier{claims: identity.SessionClaims{SessionID: "session-1", UserID: "user-1"}}
		router := chi.NewRouter()
		router.With(RequireUserAuth(sessions, nil, LocalSessionCookieName), RequireScope(identity.ScopeProjectsRead)).Get("/projects", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		res := performAuthTestRequest(router, http.MethodGet, "/projects", "Cookie", LocalSessionCookieName+"=valid-session")
		if res.Code != http.StatusNoContent {
			t.Fatalf("expected status 204, got %d", res.Code)
		}
	})
}

func TestBearerAuthorizationUsesCaseInsensitiveScheme(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/projects", http.NoBody)
	req.Header.Set("Authorization", "bearer api-secret")
	if !bearerAuthorization(req) {
		t.Fatal("expected lowercase Bearer scheme to bypass session CSRF validation")
	}
	req.Header.Set("Authorization", "Basic credentials")
	if bearerAuthorization(req) {
		t.Fatal("non-Bearer authorization must not bypass session CSRF validation")
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

type recordingAPITokenVerifier struct {
	principal appapitoken.Principal
	err       error
	gotToken  string
}

func (v *recordingAPITokenVerifier) Verify(_ context.Context, rawToken string) (appapitoken.Principal, error) {
	v.gotToken = rawToken
	return v.principal, v.err
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

func (v *recordingTokenVerifier) ListUserSessions(context.Context, string) ([]identity.AuthSession, error) {
	return nil, nil
}

func (v *recordingTokenVerifier) RevokeUserSession(context.Context, string, string, string) error {
	return nil
}

func (v *recordingTokenVerifier) RevokeUserSessions(context.Context, string, string) error {
	return nil
}
