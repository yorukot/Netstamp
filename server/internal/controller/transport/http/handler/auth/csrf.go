package auth

import (
	"context"
	"errors"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

func (h *Handler) csrf(ctx context.Context, _ *csrfInput) (*csrfOutput, error) {
	claims, ok := httpmiddleware.SessionClaimsFromContext(ctx)
	if !ok || claims.SessionID == "" {
		return nil, httpx.UnauthorizedCode(httpx.CodeAuthMissingSession, "missing auth cookie")
	}
	if h.verifier == nil {
		return nil, httpx.InternalServerError("csrf unavailable")
	}
	token, err := h.verifier.CreateCSRFToken(ctx, claims.SessionID)
	if err != nil {
		return nil, mapCSRFError(err)
	}
	return &csrfOutput{Body: csrfOutputBody{CSRFToken: token}}, nil
}

func mapCSRFError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, appauth.ErrSessionInvalid) {
		return httpx.UnauthorizedCode(httpx.CodeAuthInvalidSession, "invalid session")
	}
	return httpx.InternalServerError("csrf unavailable")
}

type csrfInput struct{}

type csrfOutput struct {
	Body csrfOutputBody
}

type csrfOutputBody struct {
	CSRFToken string `json:"csrfToken"`
}
