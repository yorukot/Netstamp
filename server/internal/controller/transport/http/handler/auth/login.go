package auth

import (
	"context"
	"errors"
	"net/http"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func (h *Handler) login(ctx context.Context, input *loginInput) (*loginOutput, error) {
	emailVerificationRequired := false
	if h.settings != nil {
		settings, err := h.settings.EffectiveSettings(ctx)
		if err != nil {
			return nil, httpx.InternalServerError("login failed")
		}
		emailVerificationRequired = settings.EmailVerificationRequired
	}

	result, err := h.service.Login(ctx, appauth.LoginInput{
		Email:                    input.Body.Email,
		Password:                 input.Body.Password,
		RequireEmailVerification: emailVerificationRequired,
	})
	if err != nil {
		switch {
		case errors.Is(err, appauth.ErrCredentialsInvalid), errors.Is(err, appauth.ErrInvalidInput):
			return nil, httpx.UnauthorizedCode(httpx.CodeAuthInvalidCredentials, "invalid email or password")
		case errors.Is(err, appauth.ErrEmailVerificationRequired):
			return nil, httpx.ForbiddenCode(httpx.CodeAuthEmailVerificationRequired, "email verification is required")
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
				EmailVerified: result.EmailVerified,
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
