package auth

import (
	"context"
	"errors"
	"net/http"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func (h *Handler) register(ctx context.Context, r *http.Request, input *registerInput) (*registerOutput, error) {
	registrationEnabled := h.registrationEnabled
	emailVerificationRequired := false
	if h.settings != nil {
		settings, err := h.settings.EffectiveSettings(ctx)
		if err != nil {
			return nil, httpx.InternalServerError("register user failed")
		}
		registrationEnabled = settings.RegistrationEnabled
		emailVerificationRequired = settings.EmailVerificationRequired
	}
	if !registrationEnabled {
		return nil, httpx.ForbiddenCode(httpx.CodeAuthRegistrationDisabled, "registration is disabled")
	}

	result, err := h.service.Register(ctx, appauth.RegisterInput{
		Email:                    input.Body.Email,
		DisplayName:              input.Body.DisplayName,
		Password:                 input.Body.Password,
		RequireEmailVerification: emailVerificationRequired,
		EmailVerificationBaseURL: h.resetBaseURL(r),
	})
	if err != nil {
		switch {
		case errors.Is(err, appauth.ErrInvalidInput):
			return nil, invalidAuthInputError(err)
		case errors.Is(err, identity.ErrEmailAlreadyExists):
			return nil, httpx.ConflictCode(httpx.CodeEmailAlreadyExists, "email already exists")
		case errors.Is(err, appauth.ErrEmailVerificationUnavailable):
			return nil, httpx.ServiceUnavailableCode(httpx.CodeAuthEmailVerificationUnavailable, "email verification is unavailable")
		default:
			return nil, httpx.InternalServerError("register user failed")
		}
	}

	if result.EmailVerificationRequired {
		return &registerOutput{
			Status: http.StatusAccepted,
			Body: registerEmailVerificationRequiredOutputBody{
				EmailVerificationRequired: true,
			},
		}, nil
	}

	return &registerOutput{
		Status:    http.StatusCreated,
		SetCookie: sessionCookiePtr(newSessionCookie(h.cookieName, result.SessionToken, result.ExpiresIn, h.cookieSecure)),
		Body: registerOutputBody{
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

func sessionCookiePtr(cookie http.Cookie) *http.Cookie {
	return &cookie
}

type registerInput struct {
	Body registerInputBody
}

type registerOutput struct {
	Status    int
	SetCookie *http.Cookie
	Body      any
}

type registerInputBody struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password"`
}

type registerOutputBody struct {
	User userResponse `json:"user"`
}

type registerEmailVerificationRequiredOutputBody struct {
	EmailVerificationRequired bool `json:"emailVerificationRequired"`
}
