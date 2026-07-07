package auth

import (
	"context"
	"errors"
	"net/http"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func (h *Handler) login(ctx context.Context, input *loginInput) (*loginOutput, error) {
	result, err := h.service.Login(ctx, appauth.LoginInput{
		Email:    input.Body.Email,
		Password: input.Body.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, appauth.ErrCredentialsInvalid), errors.Is(err, appauth.ErrInvalidInput):
			return nil, httpx.Unauthorized("invalid email or password")
		default:
			return nil, httpx.InternalServerError("login failed")
		}
	}

	return &loginOutput{
		SetCookie: newSessionCookie(result.AccessToken, result.ExpiresIn, h.cookieSecure),
		Body: loginOutputBody{
			User: userResponse{
				ID:            result.UserID,
				Email:         result.Email,
				DisplayName:   result.DisplayName,
				IsSystemAdmin: result.IsSystemAdmin,
			},
		},
	}, nil
}

type loginInput struct {
	Body loginInputBody
}

type loginOutput struct {
	SetCookie http.Cookie
	Body      loginOutputBody
}

type loginInputBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginOutputBody struct {
	User userResponse `json:"user"`
}
