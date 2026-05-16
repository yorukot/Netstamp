package auth

import (
	"context"
	"errors"
	"net/http"

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
		SetCookie: newSessionCookie(result.AccessToken, result.ExpiresIn, h.cookieSecure),
		Body: loginOutputBody{
			User: userResponse{
				ID:          result.UserID,
				Email:       result.Email,
				DisplayName: result.DisplayName,
			},
		},
	}, nil
}

type loginInput struct {
	Body loginInputBody
}

type loginOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie" hidden:"true"`
	Body      loginOutputBody
}

type loginInputBody struct {
	Email    string `json:"email,omitempty" doc:"Email address used to sign in. It is normalized before lookup." example:"user@example.com"`
	Password string `json:"password,omitempty" writeOnly:"true" doc:"Plain-text password to verify. It is never returned by the API." example:"correct-horse-battery-staple"` //nolint:gosec // Login requests intentionally accept plaintext passwords over TLS.
}

type loginOutputBody struct {
	User userResponse `json:"user" doc:"Authenticated user."`
}
