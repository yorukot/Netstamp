package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestRegisterReturnsCreatedUserWithDisplayName(t *testing.T) {
	_, api := humatest.New(t)
	repo := &handlerUserRepository{}
	tokenIssuer := &handlerTokenIssuer{
		token: identity.IssuedToken{
			Value:     "access-token",
			TokenType: "Bearer",
			ExpiresIn: 3600,
		},
	}
	NewHandler(newTestAuthService(repo, &handlerPasswordHasher{}, tokenIssuer), nil).RegisterRoutes(api)

	res := api.Post("/auth/register", map[string]any{
		"email":       " User@Example.COM ",
		"displayName": "  Example User  ",
		"password":    "correct-password",
	})

	if res.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", res.Code)
	}

	var body registerOutputBody
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.User.ID != "created-user" {
		t.Fatalf("expected user id, got %q", body.User.ID)
	}
	if body.User.Email != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", body.User.Email)
	}
	if body.User.DisplayName == nil || *body.User.DisplayName != "Example User" {
		t.Fatalf("expected display name, got %#v", body.User.DisplayName)
	}
	if body.AccessToken != "access-token" {
		t.Fatalf("expected access token, got %q", body.AccessToken)
	}
	if body.TokenType != "Bearer" {
		t.Fatalf("expected token type, got %q", body.TokenType)
	}
	if body.ExpiresIn != 3600 {
		t.Fatalf("expected expiry, got %d", body.ExpiresIn)
	}
	if repo.gotCreateInput.DisplayName != "Example User" {
		t.Fatalf("expected display name in create input, got %q", repo.gotCreateInput.DisplayName)
	}
	if tokenIssuer.gotInput.DisplayName == nil || *tokenIssuer.gotInput.DisplayName != "Example User" {
		t.Fatalf("expected display name in token input, got %#v", tokenIssuer.gotInput.DisplayName)
	}
}

func TestRegisterRejectsMissingDisplayName(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(newTestAuthService(&handlerUserRepository{}, &handlerPasswordHasher{}, &handlerTokenIssuer{}), nil).RegisterRoutes(api)

	res := api.Post("/auth/register", map[string]any{
		"email":    "user@example.com",
		"password": "correct-password",
	})

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", res.Code)
	}
	assertRegisterHumaErrorDetail(t, res, "body.displayName", "must not be blank")
}

func TestRegisterReturnsServiceFieldErrorsForSemanticValidation(t *testing.T) {
	tests := []struct {
		name        string
		body        map[string]any
		wantField   string
		wantMessage string
	}{
		{
			name: "invalid email",
			body: map[string]any{
				"email":       "not-an-email",
				"displayName": "Example User",
				"password":    "correct-password",
			},
			wantField:   "body.email",
			wantMessage: "must be a valid email address",
		},
		{
			name: "blank display name",
			body: map[string]any{
				"email":       "user@example.com",
				"displayName": "   ",
				"password":    "correct-password",
			},
			wantField:   "body.displayName",
			wantMessage: "must not be blank",
		},
		{
			name: "short password",
			body: map[string]any{
				"email":       "user@example.com",
				"displayName": "Example User",
				"password":    "short",
			},
			wantField:   "body.password",
			wantMessage: "must be at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, api := humatest.New(t)
			NewHandler(newTestAuthService(&handlerUserRepository{}, &handlerPasswordHasher{}, &handlerTokenIssuer{}), nil).RegisterRoutes(api)

			res := api.Post("/auth/register", tt.body)
			if res.Code != http.StatusUnprocessableEntity {
				t.Fatalf("expected status 422, got %d", res.Code)
			}
			assertRegisterHumaErrorDetail(t, res, tt.wantField, tt.wantMessage)
		})
	}
}

func TestRegisterMapsDuplicateEmailToConflict(t *testing.T) {
	_, api := humatest.New(t)
	NewHandler(newTestAuthService(
		&handlerUserRepository{createErr: appauth.ErrEmailAlreadyExists},
		&handlerPasswordHasher{},
		&handlerTokenIssuer{},
	), nil).RegisterRoutes(api)

	res := api.Post("/auth/register", map[string]any{
		"email":       "user@example.com",
		"displayName": "Example User",
		"password":    "correct-password",
	})

	if res.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", res.Code)
	}
}

func assertRegisterHumaErrorDetail(t *testing.T, res *httptest.ResponseRecorder, wantLocation, wantMessage string) {
	t.Helper()

	var body huma.ErrorModel
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	for _, detail := range body.Errors {
		if detail.Location == wantLocation && detail.Message == wantMessage {
			return
		}
	}

	t.Fatalf("expected error detail %q/%q, got %#v", wantLocation, wantMessage, body.Errors)
}

func newTestAuthService(repo appauth.UserRepository, hasher appauth.PasswordHasher, tokens appauth.TokenIssuer) *appauth.Service {
	return appauth.NewService(repo, hasher, tokens, handlerSecurityEventRecorder{})
}

type handlerUserRepository struct {
	user           identity.User
	createdUser    identity.User
	getErr         error
	createErr      error
	gotEmail       string
	gotCreateInput identity.CreateUserInput
}

func (r *handlerUserRepository) CreateUser(_ context.Context, input identity.CreateUserInput) (identity.User, error) {
	r.gotCreateInput = input
	if r.createErr != nil {
		return identity.User{}, r.createErr
	}
	if r.createdUser.ID != "" {
		return r.createdUser, nil
	}
	return identity.User{
		ID:           "created-user",
		Email:        input.Email,
		DisplayName:  &input.DisplayName,
		PasswordHash: input.PasswordHash,
		IsActive:     true,
	}, nil
}

func (r *handlerUserRepository) GetUserByEmail(_ context.Context, email string) (identity.User, error) {
	r.gotEmail = email
	if r.getErr != nil {
		return identity.User{}, r.getErr
	}
	return r.user, nil
}

type handlerPasswordHasher struct {
	hashErr    error
	compareErr error
}

func (h *handlerPasswordHasher) Hash(password string) (string, error) {
	if h.hashErr != nil {
		return "", h.hashErr
	}
	return "hashed:" + password, nil
}

func (h *handlerPasswordHasher) Compare(_, _ string) error {
	return h.compareErr
}

type handlerTokenIssuer struct {
	token    identity.IssuedToken
	err      error
	gotInput identity.AccessTokenInput
}

func (i *handlerTokenIssuer) IssueAccessToken(_ context.Context, input identity.AccessTokenInput) (identity.IssuedToken, error) {
	i.gotInput = input
	if i.err != nil {
		return identity.IssuedToken{}, i.err
	}
	return i.token, nil
}

type handlerSecurityEventRecorder struct{}

func (handlerSecurityEventRecorder) RecordAuthEvent(context.Context, appauth.AuthEvent) {}
