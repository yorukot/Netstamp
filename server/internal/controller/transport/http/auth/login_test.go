package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestLoginReturnsUserWithDisplayName(t *testing.T) {
	_, api := humatest.New(t)
	tokenIssuer := &handlerTokenIssuer{
		token: identity.IssuedToken{
			Value:     "access-token",
			TokenType: "Bearer",
			ExpiresIn: 3600,
		},
	}
	repo := &handlerUserRepository{
		user: identity.User{
			ID:           "11111111-1111-1111-1111-111111111111",
			Email:        "user@example.com",
			DisplayName:  "Example User",
			PasswordHash: "password-hash",
			IsActive:     true,
		},
	}
	NewHandler(newTestAuthService(repo, &handlerPasswordHasher{}, tokenIssuer), nil).RegisterRoutes(api)

	res := api.Post("/auth/login", map[string]any{
		"email":    "user@example.com",
		"password": "correct-password",
	})

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	var body loginOutputBody
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.User.ID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("expected user id, got %q", body.User.ID)
	}
	if body.User.Email != "user@example.com" {
		t.Fatalf("expected user email, got %q", body.User.Email)
	}
	if body.User.DisplayName != "Example User" {
		t.Fatalf("expected display name, got %q", body.User.DisplayName)
	}
	if body.AccessToken != "access-token" {
		t.Fatalf("expected access token, got %q", body.AccessToken)
	}
	if repo.gotEmail != "user@example.com" {
		t.Fatalf("expected lookup email, got %q", repo.gotEmail)
	}
}

func TestLoginMapsInvalidCredentialsToUnauthorized(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(newTestAuthService(
		&handlerUserRepository{getErr: appauth.ErrUserNotFound},
		&handlerPasswordHasher{},
		&handlerTokenIssuer{},
	), nil).RegisterRoutes(api)

	res := api.Post("/auth/login", map[string]any{
		"email":    "missing@example.com",
		"password": "wrong-password",
	})

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
}

func TestLoginMapsInvalidInputToUnauthorized(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerUserRepository{}
	NewHandler(newTestAuthService(
		repo,
		&handlerPasswordHasher{},
		&handlerTokenIssuer{},
	), nil).RegisterRoutes(api)

	res := api.Post("/auth/login", map[string]any{
		"email":    "   ",
		"password": "correct-password",
	})

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
	if repo.gotEmail != "" {
		t.Fatalf("expected invalid login input not to hit repository, got %q", repo.gotEmail)
	}
}

func TestLoginMapsPasswordMismatchToUnauthorized(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(newTestAuthService(
		&handlerUserRepository{
			user: identity.User{
				ID:           "11111111-1111-1111-1111-111111111111",
				Email:        "user@example.com",
				PasswordHash: "password-hash",
				IsActive:     true,
			},
		},
		&handlerPasswordHasher{compareErr: errors.New("password mismatch")},
		&handlerTokenIssuer{},
	), nil).RegisterRoutes(api)

	res := api.Post("/auth/login", map[string]any{
		"email":    "user@example.com",
		"password": "wrong-password",
	})

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
}
