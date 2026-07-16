package middleware

import (
	"context"
	"errors"
	"net/http"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

type SudoService interface {
	RequireSudo(context.Context, string) error
}

type PasswordChangeAuthorizationService interface {
	AuthorizePasswordChange(context.Context, string, string) error
}

func RequireSudo(service SudoService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := SessionClaimsFromContext(r.Context())
			if !ok || claims.SessionID == "" {
				httpx.WriteProblem(w, r, httpx.UnauthorizedCode(httpx.CodeAuthMissingSession, "authentication required"))
				return
			}
			if err := service.RequireSudo(r.Context(), claims.SessionID); err != nil {
				if errors.Is(err, appauth.ErrSudoRequired) {
					httpx.WriteProblem(w, r, httpx.ForbiddenCode(httpx.CodeAuthSudoRequired, "recent authentication required"))
					return
				}
				httpx.WriteProblem(w, r, httpx.UnauthorizedCode(httpx.CodeAuthInvalidSession, "invalid session"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequirePasswordChangeAuthorization(service PasswordChangeAuthorizationService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := SessionClaimsFromContext(r.Context())
			if !ok || claims.UserID == "" || claims.SessionID == "" {
				httpx.WriteProblem(w, r, httpx.UnauthorizedCode(httpx.CodeAuthMissingSession, "authentication required"))
				return
			}
			if err := service.AuthorizePasswordChange(r.Context(), claims.UserID, claims.SessionID); err != nil {
				if errors.Is(err, appauth.ErrSudoRequired) {
					httpx.WriteProblem(w, r, httpx.ForbiddenCode(httpx.CodeAuthSudoRequired, "recent authentication required"))
					return
				}
				httpx.WriteProblem(w, r, httpx.UnauthorizedCode(httpx.CodeAuthInvalidSession, "invalid session"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
