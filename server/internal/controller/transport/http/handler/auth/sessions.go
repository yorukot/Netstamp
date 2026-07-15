package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func (h *Handler) listSessions(ctx context.Context, _ *listSessionsInput) (*listSessionsOutput, error) {
	claims, ok := httpmiddleware.SessionClaimsFromContext(ctx)
	if !ok || claims.UserID == "" || claims.SessionID == "" {
		return nil, httpx.UnauthorizedCode(httpx.CodeAuthMissingSession, "missing auth cookie")
	}

	sessions, err := h.service.ListSessions(ctx, claims.UserID, claims.SessionID)
	if err != nil {
		return nil, httpx.InternalServerError("failed to list sessions")
	}

	items := make([]sessionOutputBody, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, sessionOutputBody{
			ID:                session.ID,
			UserAgent:         session.UserAgent,
			CreatedAt:         session.CreatedAt,
			LastUsedAt:        session.LastUsedAt,
			IdleExpiresAt:     session.IdleExpiresAt,
			AbsoluteExpiresAt: session.AbsoluteExpiresAt,
			IsCurrent:         session.IsCurrent,
		})
	}

	return &listSessionsOutput{Body: listSessionsOutputBody{Sessions: items}}, nil
}

func (h *Handler) revokeSession(ctx context.Context, input *revokeSessionInput) (*revokeSessionOutput, error) {
	claims, ok := httpmiddleware.SessionClaimsFromContext(ctx)
	if !ok || claims.UserID == "" || claims.SessionID == "" {
		return nil, httpx.UnauthorizedCode(httpx.CodeAuthMissingSession, "missing auth cookie")
	}

	if err := h.service.RevokeSessionByID(ctx, claims.UserID, input.SessionID); err != nil {
		if errors.Is(err, identity.ErrSessionNotFound) {
			return nil, httpx.NotFoundCode(httpx.CodeAuthSessionNotFound, "session not found")
		}
		return nil, httpx.InternalServerError("failed to revoke session")
	}

	output := &revokeSessionOutput{}
	if input.SessionID == claims.SessionID {
		cookie := expiredSessionCookie(h.cookieName, h.cookieSecure)
		output.SetCookie = &cookie
	}
	return output, nil
}

func (h *Handler) revokeAllSessions(ctx context.Context, _ *revokeAllSessionsInput) (*revokeAllSessionsOutput, error) {
	claims, ok := httpmiddleware.SessionClaimsFromContext(ctx)
	if !ok || claims.UserID == "" || claims.SessionID == "" {
		return nil, httpx.UnauthorizedCode(httpx.CodeAuthMissingSession, "missing auth cookie")
	}

	if err := h.service.RevokeAllSessions(ctx, claims.UserID); err != nil {
		return nil, httpx.InternalServerError("failed to revoke sessions")
	}

	return &revokeAllSessionsOutput{SetCookie: expiredSessionCookie(h.cookieName, h.cookieSecure)}, nil
}

type listSessionsInput struct{}

type listSessionsOutput struct {
	Body listSessionsOutputBody
}

type listSessionsOutputBody struct {
	Sessions []sessionOutputBody `json:"sessions"`
}

type sessionOutputBody struct {
	ID                string    `json:"id"`
	UserAgent         string    `json:"userAgent"`
	CreatedAt         time.Time `json:"createdAt"`
	LastUsedAt        time.Time `json:"lastUsedAt"`
	IdleExpiresAt     time.Time `json:"idleExpiresAt"`
	AbsoluteExpiresAt time.Time `json:"absoluteExpiresAt"`
	IsCurrent         bool      `json:"isCurrent"`
}

type revokeSessionInput struct {
	SessionID string
}

type revokeSessionOutput struct {
	SetCookie *http.Cookie
}

type revokeAllSessionsInput struct{}

type revokeAllSessionsOutput struct {
	SetCookie http.Cookie
}
