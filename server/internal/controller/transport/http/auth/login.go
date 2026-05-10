package auth

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
)

func (h *Handler) login(ctx context.Context, input *loginInput) (*loginOutput, error) {
	result, err := h.service.Login(ctx, appauth.LoginInput{
		Email:    input.Body.Email,
		Password: input.Body.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, appauth.ErrCredentialsInvalid), errors.Is(err, appauth.ErrInvalidInput):
			return nil, huma.Error401Unauthorized("invalid email or password")
		default:
			return nil, huma.Error500InternalServerError("login failed")
		}
	}

	return &loginOutput{
		Body: loginOutputBody{
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

type loginInput struct {
	Body loginInputBody
}

type loginOutput struct {
	Body loginOutputBody
}

type loginInputBody struct {
	Email    string `json:"email,omitempty" doc:"Email address used to sign in. It is normalized before lookup." example:"user@example.com"`
	Password string `json:"password,omitempty" writeOnly:"true" doc:"Plain-text password to verify. It is never returned by the API." example:"correct-horse-battery-staple"` //nolint:gosec // Login requests intentionally accept plaintext passwords over TLS.
}

type loginOutputBody struct {
	User        userResponse `json:"user" doc:"Authenticated user."`
	TokenType   string       `json:"tokenType" example:"Bearer" doc:"Token scheme to use in the Authorization header."`
	AccessToken string       `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example.signature" doc:"JWT access token issued for the authenticated user."` //nolint:gosec // Login responses intentionally return the issued access token.
	ExpiresIn   int          `json:"expiresIn" example:"43200" doc:"Access token lifetime in seconds."`
}
