package auth

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func (h *Handler) register(ctx context.Context, input *registerInput) (*registerOutput, error) {
	result, err := h.service.Register(ctx, appauth.RegisterInput{
		Email:       input.Body.Email,
		DisplayName: input.Body.DisplayName,
		Password:    input.Body.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, appauth.ErrInvalidInput):
			return nil, invalidAuthInputError(err)
		case errors.Is(err, identity.ErrEmailAlreadyExists):
			return nil, huma.Error409Conflict("email already exists")
		default:
			return nil, huma.Error500InternalServerError("register user failed")
		}
	}

	return &registerOutput{
		Body: registerOutputBody{
			User: userResponse{
				ID:          result.UserID,
				Email:       result.Email,
				DisplayName: result.DisplayName,
			},
			TokenType:   result.TokenType,
			AccessToken: result.AccessToken,
			ExpiresIn:   result.ExpiresIn,
		},
	}, nil
}

type registerInput struct {
	Body registerInputBody
}

type registerOutput struct {
	Body registerOutputBody
}

type registerInputBody struct {
	Email       string `json:"email,omitempty" doc:"Email address used to sign in." example:"user@example.com"`
	DisplayName string `json:"displayName,omitempty" doc:"Name shown in the app." example:"Jane Doe"`
	Password    string `json:"password,omitempty" writeOnly:"true" doc:"Plain-text password. It is stored only as an Argon2id hash." example:"correct-horse-battery-staple"` //nolint:gosec // Register requests intentionally accept plaintext passwords over TLS.
}

type registerOutputBody struct {
	User        userResponse `json:"user" doc:"Created user."`
	TokenType   string       `json:"tokenType" example:"Bearer" doc:"Token scheme to use in the Authorization header."`
	AccessToken string       `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example.signature" doc:"JWT access token issued for the created user."` //nolint:gosec // Register responses intentionally return the issued access token.
	ExpiresIn   int          `json:"expiresIn" example:"43200" doc:"Access token lifetime in seconds."`
}
