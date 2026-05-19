package auth

import (
	"context"
	"errors"
	"net/http"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
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
			return nil, httpx.Conflict("email already exists")
		default:
			return nil, httpx.InternalServerError("register user failed")
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
	SetCookie http.Cookie
	Body      registerOutputBody
}

type registerInputBody struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password"` //nolint:gosec // Register requests intentionally accept plaintext passwords over TLS.
}

type registerOutputBody struct {
	User userResponse `json:"user"`
}
