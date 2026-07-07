package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestLoginSetsSessionCookieAndReturnsUserOnly(t *testing.T) {
	router := chi.NewRouter()
	NewHandler(newAuthTestService(), nil, nil, true, true).RegisterRoutes(router)

	res := performJSONRequest(router, http.MethodPost, "/auth/login", `{"email":"user@example.com","password":"correct-horse-battery-staple"}`)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	assertSessionCookie(t, res.Result(), "issued-token", true)
	assertAuthBodyHasOnlyUser(t, res)
}

func TestRegisterSetsSessionCookieAndReturnsUserOnly(t *testing.T) {
	router := chi.NewRouter()
	NewHandler(newAuthTestService(), nil, nil, false, true).RegisterRoutes(router)

	res := performJSONRequest(router, http.MethodPost, "/auth/register", `{"email":"new@example.com","displayName":"New User","password":"correct-horse-battery-staple"}`)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", res.Code)
	}
	assertSessionCookie(t, res.Result(), "issued-token", false)
	assertAuthBodyHasOnlyUser(t, res)
}

func TestRegisterReturnsForbiddenWhenRegistrationDisabled(t *testing.T) {
	router := chi.NewRouter()
	NewHandler(newAuthTestService(), nil, nil, false, false).RegisterRoutes(router)

	res := performJSONRequest(router, http.MethodPost, "/auth/register", `{"email":"new@example.com","displayName":"New User","password":"correct-horse-battery-staple"}`)

	if res.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", res.Code)
	}

	var body httpx.ProblemDetails
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Detail != "registration is disabled" {
		t.Fatalf("expected disabled registration detail, got %q", body.Detail)
	}
}

func TestLogoutExpiresSessionCookie(t *testing.T) {
	router := chi.NewRouter()
	NewHandler(nil, nil, nil, true, true).RegisterRoutes(router)

	res := performJSONRequest(router, http.MethodPost, "/auth/logout", "")

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", res.Code)
	}
	cookie := findResponseCookie(t, res.Result(), "netstamp_session")
	if cookie.Value != "" {
		t.Fatalf("expected empty cookie value, got %q", cookie.Value)
	}
	if cookie.MaxAge != -1 {
		t.Fatalf("expected expired cookie MaxAge -1, got %d", cookie.MaxAge)
	}
	if !cookie.HttpOnly {
		t.Fatal("expected HttpOnly cookie")
	}
	if !cookie.Secure {
		t.Fatal("expected secure cookie")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite=Lax, got %v", cookie.SameSite)
	}
}

func performJSONRequest(router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	return res
}

func assertSessionCookie(t *testing.T, res *http.Response, value string, secure bool) {
	t.Helper()

	cookie := findResponseCookie(t, res, "netstamp_session")
	if cookie.Value != value {
		t.Fatalf("expected session cookie value %q, got %q", value, cookie.Value)
	}
	if cookie.Path != "/" {
		t.Fatalf("expected cookie path /, got %q", cookie.Path)
	}
	if cookie.MaxAge != 3600 {
		t.Fatalf("expected cookie MaxAge 3600, got %d", cookie.MaxAge)
	}
	if cookie.Secure != secure {
		t.Fatalf("expected cookie Secure %v, got %v", secure, cookie.Secure)
	}
	if !cookie.HttpOnly {
		t.Fatal("expected HttpOnly cookie")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite=Lax, got %v", cookie.SameSite)
	}
}

func findResponseCookie(t *testing.T, res *http.Response, name string) *http.Cookie {
	t.Helper()

	for _, cookie := range res.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("expected response cookie %q in %q", name, strings.Join(res.Header.Values("Set-Cookie"), "; "))
	return nil
}

func assertAuthBodyHasOnlyUser(t *testing.T, res *httptest.ResponseRecorder) {
	t.Helper()

	var body map[string]json.RawMessage
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if _, ok := body["user"]; !ok {
		t.Fatal("expected user in response body")
	}
	for _, field := range []string{"tokenType", "accessToken", "expiresIn"} {
		if _, ok := body[field]; ok {
			t.Fatalf("expected response body not to include %q", field)
		}
	}
}

func newAuthTestService() *appauth.Service {
	user := identity.User{
		ID:           "11111111-1111-1111-1111-111111111111",
		Email:        "user@example.com",
		DisplayName:  "User",
		PasswordHash: "hashed-password",
	}
	return appauth.NewService(&authTestUserRepository{user: user}, authTestPasswordHasher{}, authTestTokenIssuer{}, authTestEvents{})
}

type authTestUserRepository struct {
	user identity.User
}

func (r *authTestUserRepository) CreateUser(_ context.Context, input identity.User) (identity.User, error) {
	input.ID = r.user.ID
	return input, nil
}

func (r *authTestUserRepository) GetUserByEmail(context.Context, string) (identity.User, error) {
	return r.user, nil
}

func (r *authTestUserRepository) GetUserByID(context.Context, string) (identity.User, error) {
	return r.user, nil
}

func (r *authTestUserRepository) UpdateUserPasswordHash(context.Context, identity.User) (identity.User, error) {
	return r.user, nil
}

type authTestPasswordHasher struct{}

func (authTestPasswordHasher) Hash(context.Context, string) (string, error) {
	return "hashed-password", nil
}

func (authTestPasswordHasher) Compare(context.Context, string, string) error {
	return nil
}

type authTestTokenIssuer struct{}

func (authTestTokenIssuer) IssueAccessToken(context.Context, identity.AccessTokenClaims) (identity.IssuedToken, error) {
	return identity.IssuedToken{
		Value:     "issued-token",
		ExpiresIn: 3600,
	}, nil
}

type authTestEvents struct{}

func (authTestEvents) RecordAuthEvent(context.Context, appauth.AuthEvent) {}
