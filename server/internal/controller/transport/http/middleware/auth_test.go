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

func TestRequireAuthRejectsInvalidAccessToken(t *testing.T) {
	verifier := &recordingTokenVerifier{err: appauth.ErrAccessTokenInvalid}
	router := registerClaimsRoute(t, verifier)

	res := performAuthTestRequest(router, http.MethodGet, "/me", "Cookie", SessionCookieName+"=bad-token")

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
	if verifier.gotToken != "bad-token" {
		t.Fatalf("expected verifier token %q, got %q", "bad-token", verifier.gotToken)
	}
}

func TestRequireAuthStoresClaimsInContext(t *testing.T) {
	verifier := &recordingTokenVerifier{
		claims: identity.AccessTokenClaims{
			Subject: "user-1",
			Email:   "user@example.com",
		},
	}
	router := registerClaimsRoute(t, verifier)

	res := performAuthTestRequest(router, http.MethodGet, "/me", "Cookie", SessionCookieName+"=good-token")

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	if verifier.gotToken != "good-token" {
		t.Fatalf("expected verifier token %q, got %q", "good-token", verifier.gotToken)
	}
}

func registerClaimsRoute(t *testing.T, verifier appauth.TokenVerifier) http.Handler {
	t.Helper()

	router := chi.NewRouter()
	router.With(RequireAuth(verifier)).Get("/me", func(w http.ResponseWriter, r *http.Request) {
		claims, ok := AccessTokenClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, "missing claims", http.StatusInternalServerError)
			return
		}
		if claims.Subject == "" || claims.Email == "" {
			http.Error(w, "empty claims", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(claimsRouteBody{
			Subject: claims.Subject,
			Email:   claims.Email,
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
	Subject string `json:"subject"`
	Email   string `json:"email"`
}

type recordingTokenVerifier struct {
	claims   identity.AccessTokenClaims
	err      error
	gotToken string
}

func (v *recordingTokenVerifier) VerifyAccessToken(_ context.Context, value string) (identity.AccessTokenClaims, error) {
	v.gotToken = value
	if v.err != nil {
		return identity.AccessTokenClaims{}, v.err
	}
	return v.claims, nil
}
