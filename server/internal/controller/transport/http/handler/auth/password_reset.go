package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func (h *Handler) requestPasswordReset(ctx context.Context, r *http.Request, input *requestPasswordResetInput) error {
	if h.service == nil {
		return httpx.InternalServerError("password reset failed")
	}

	err := h.service.RequestPasswordReset(ctx, appauth.RequestPasswordResetInput{
		Email:        input.Body.Email,
		ResetBaseURL: h.resetBaseURL(r),
	})
	if err != nil {
		switch {
		case errors.Is(err, appauth.ErrInvalidInput):
			return invalidAuthInputError(err)
		case errors.Is(err, appauth.ErrResetUnavailable):
			return httpx.ServiceUnavailableCode(httpx.CodeAuthPasswordResetUnavailable, "password reset is unavailable")
		default:
			return httpx.InternalServerError("password reset failed")
		}
	}

	return nil
}

func (h *Handler) confirmPasswordReset(ctx context.Context, input *confirmPasswordResetInput) error {
	if h.service == nil {
		return httpx.InternalServerError("password reset failed")
	}

	err := h.service.ConfirmPasswordReset(ctx, appauth.ConfirmPasswordResetInput{
		Token:       input.Body.Token,
		NewPassword: input.Body.NewPassword,
	})
	if err != nil {
		switch {
		case errors.Is(err, appauth.ErrInvalidInput):
			return invalidAuthInputError(err)
		case errors.Is(err, appauth.ErrResetTokenInvalid):
			return httpx.UnauthorizedCode(httpx.CodeAuthPasswordResetTokenInvalid, "invalid or expired password reset token")
		case errors.Is(err, appauth.ErrResetUnavailable):
			return httpx.ServiceUnavailableCode(httpx.CodeAuthPasswordResetUnavailable, "password reset is unavailable")
		default:
			return httpx.InternalServerError("password reset failed")
		}
	}

	return nil
}

func (h *Handler) resetBaseURL(r *http.Request) string {
	if h.settings != nil {
		settings, err := h.settings.EffectiveSettings(r.Context())
		if err == nil && strings.TrimSpace(settings.PublicWebBaseURL) != "" {
			return strings.TrimRight(strings.TrimSpace(settings.PublicWebBaseURL), "/")
		}
	}
	if strings.TrimSpace(h.publicWebBaseURL) != "" {
		return strings.TrimRight(strings.TrimSpace(h.publicWebBaseURL), "/")
	}

	scheme := "http"
	if forwardedProto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwardedProto == "http" || forwardedProto == "https" {
		scheme = forwardedProto
	} else if r.TLS != nil {
		scheme = "https"
	}

	return scheme + "://" + r.Host
}

type requestPasswordResetInput struct {
	Body requestPasswordResetInputBody
}

type requestPasswordResetInputBody struct {
	Email string `json:"email"`
}

type confirmPasswordResetInput struct {
	Body confirmPasswordResetInputBody
}

type confirmPasswordResetInputBody struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}
