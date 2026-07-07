package auth

import (
	"context"
	"errors"
	"net/http"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func (h *Handler) requestEmailVerification(ctx context.Context, r *http.Request, input *requestEmailVerificationInput) error {
	if h.service == nil {
		return httpx.InternalServerError("email verification failed")
	}

	err := h.service.RequestEmailVerification(ctx, appauth.RequestEmailVerificationInput{
		Email:                    input.Body.Email,
		EmailVerificationBaseURL: h.resetBaseURL(r),
	})
	if err != nil {
		switch {
		case errors.Is(err, appauth.ErrInvalidInput):
			return invalidAuthInputError(err)
		case errors.Is(err, appauth.ErrEmailVerificationUnavailable):
			return httpx.ServiceUnavailable("email verification is unavailable")
		default:
			return httpx.InternalServerError("email verification failed")
		}
	}

	return nil
}

func (h *Handler) confirmEmailVerification(ctx context.Context, input *confirmEmailVerificationInput) error {
	if h.service == nil {
		return httpx.InternalServerError("email verification failed")
	}

	err := h.service.ConfirmEmailVerification(ctx, appauth.ConfirmEmailVerificationInput{
		Token: input.Body.Token,
	})
	if err != nil {
		switch {
		case errors.Is(err, appauth.ErrInvalidInput):
			return invalidAuthInputError(err)
		case errors.Is(err, appauth.ErrEmailVerificationTokenInvalid):
			return httpx.Unauthorized("invalid or expired email verification token")
		case errors.Is(err, appauth.ErrEmailVerificationUnavailable):
			return httpx.ServiceUnavailable("email verification is unavailable")
		default:
			return httpx.InternalServerError("email verification failed")
		}
	}

	return nil
}

type requestEmailVerificationInput struct {
	Body requestEmailVerificationInputBody
}

type requestEmailVerificationInputBody struct {
	Email string `json:"email"`
}

type confirmEmailVerificationInput struct {
	Body confirmEmailVerificationInputBody
}

type confirmEmailVerificationInputBody struct {
	Token string `json:"token"`
}
