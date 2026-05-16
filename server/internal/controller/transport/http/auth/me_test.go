package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestMeReturnsAuthenticatedUser(t *testing.T) {
	_, api := humatest.New(t)
	user := identity.User{
		ID:          "11111111-1111-1111-1111-111111111111",
		Email:       "user@example.com",
		DisplayName: "User",
	}
	NewHandler(appauth.NewService(&staticUserRepository{user: user}, nil, nil, nil), &staticTokenVerifier{
		claims: identity.AccessTokenClaims{
			Subject: user.ID,
			Email:   user.Email,
		},
	}, false).RegisterRoutes(api)

	res := api.Get("/auth/me", "Cookie: netstamp_session=valid-token")

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
}

type staticTokenVerifier struct {
	claims identity.AccessTokenClaims
}

func (v *staticTokenVerifier) VerifyAccessToken(context.Context, string) (identity.AccessTokenClaims, error) {
	return v.claims, nil
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
