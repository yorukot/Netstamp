package auth

import (
	"context"
	"net/http"
)

func (h *Handler) logout(context.Context, *logoutInput) (*logoutOutput, error) {
	return &logoutOutput{
		SetCookie: expiredSessionCookie(h.cookieSecure),
	}, nil
}

type logoutInput struct{}

type logoutOutput struct {
	SetCookie http.Cookie
}
