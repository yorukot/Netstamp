package auth

import (
	"context"
	"net/http"
)

func (h *Handler) logout(ctx context.Context, input *logoutInput) (*logoutOutput, error) {
	if input.RawSessionToken != "" && h.verifier != nil {
		if err := h.verifier.RevokeSession(ctx, input.RawSessionToken, "logout"); err != nil {
			return nil, err
		}
	}
	return &logoutOutput{
		SetCookie: expiredSessionCookie(h.cookieName, h.cookieSecure),
	}, nil
}

type logoutInput struct {
	RawSessionToken string
}

type logoutOutput struct {
	SetCookie http.Cookie
}
