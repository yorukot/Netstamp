package auth

import (
	"context"
	"errors"
	"net/http"

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
		SetCookie: newSessionCookie(result.AccessToken, result.ExpiresIn, h.cookieSecure),
		Body: registerOutputBody{
			User: userResponse{
				ID:          result.UserID,
				Email:       result.Email,
				DisplayName: result.DisplayName,
			},
		},
	}, nil
}

type registerInput struct {
	Body registerInputBody
}

type registerOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie" hidden:"true"`
	Body      registerOutputBody
}

type registerInputBody struct {
	Email       string `json:"email" format:"email" doc:"Email address used to sign in." example:"user@example.com"`
	DisplayName string `json:"displayName" minLength:"1" maxLength:"64" doc:"Name shown in the app." example:"Jane Doe"`
	Password    string `json:"password" minLength:"8" maxLength:"128" writeOnly:"true" doc:"Plain-text password. It is stored only as an Argon2id hash." example:"correct-horse-battery-staple"` //nolint:gosec // Register requests intentionally accept plaintext passwords over TLS.
}

type registerOutputBody struct {
	User userResponse `json:"user" doc:"Created user."`
}
