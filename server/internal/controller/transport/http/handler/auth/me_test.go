package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestMeReturnsAuthenticatedUser(t *testing.T) {
	router := chi.NewRouter()
	verifiedAt := time.Now().UTC()
	user := identity.User{
		ID:              "11111111-1111-1111-1111-111111111111",
		Email:           "user@example.com",
		DisplayName:     "User",
		EmailVerifiedAt: &verifiedAt,
	}
	NewHandler(appauth.NewService(&staticUserRepository{user: user}, nil, nil, nil), &staticTokenVerifier{
		claims: identity.SessionClaims{
			SessionID: "session-1",
			UserID:    user.ID,
		},
	}, nil, "netstamp_session", false, true).RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", http.NoBody)
	req.Header.Set("Cookie", "netstamp_session=valid-token")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	var body meOutputBody
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Authenticated {
		t.Fatal("expected authenticated response")
	}
	if body.User.ID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("expected user id, got %q", body.User.ID)
	}
	if body.User.Email != "user@example.com" {
		t.Fatalf("expected user email, got %q", body.User.Email)
	}
	if !body.User.EmailVerified {
		t.Fatal("expected verified user")
	}
}

type staticTokenVerifier struct {
	claims identity.SessionClaims
}

func (v *staticTokenVerifier) VerifySession(context.Context, string) (identity.SessionClaims, error) {
	return v.claims, nil
}

func (v *staticTokenVerifier) CreateSession(context.Context, appauth.CreateSessionInput) (identity.CreatedSession, error) {
	return identity.CreatedSession{}, nil
}

func (v *staticTokenVerifier) CreateCSRFToken(context.Context, string) (string, error) {
	return "", nil
}

func (v *staticTokenVerifier) VerifyCSRFToken(context.Context, string, string) error {
	return nil
}

func (v *staticTokenVerifier) RevokeSession(context.Context, string, string) error {
	return nil
}

func (v *staticTokenVerifier) ListUserSessions(context.Context, string) ([]identity.AuthSession, error) {
	return nil, nil
}

func (v *staticTokenVerifier) RevokeUserSession(context.Context, string, string, string) error {
	return nil
}

func (v *staticTokenVerifier) RevokeUserSessions(context.Context, string, string) error {
	return nil
}

type staticUserRepository struct {
	user identity.User
}

func (r *staticUserRepository) CreateUser(context.Context, identity.User) (identity.User, error) {
	return identity.User{}, nil
}

func (r *staticUserRepository) GetUserByEmail(context.Context, string) (identity.User, error) {
	return r.user, nil
}

func (r *staticUserRepository) GetUserByID(context.Context, string) (identity.User, error) {
	return r.user, nil
}

func (r *staticUserRepository) UpdateUserPasswordHash(context.Context, identity.User) (identity.User, error) {
	return r.user, nil
}
