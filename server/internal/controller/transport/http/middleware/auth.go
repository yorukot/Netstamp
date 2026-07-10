package middleware

import (
	"context"
	"net/http"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

const (
	LocalSessionCookieName      = "netstamp_session"
	ProductionSessionCookieName = "__Host-netstamp_session"
	SessionCookieName           = LocalSessionCookieName
)

type sessionClaimsContextKey struct{}

func RequireAuth(verifier appauth.SessionManager, cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if verifier == nil {
				WriteProblem(w, r, http.StatusInternalServerError, "auth verifier unavailable")
				return
			}

			token, ok := sessionCookie(r, cookieName)
			if !ok {
				WriteProblemCode(w, r, http.StatusUnauthorized, httpx.CodeAuthMissingSession, "missing auth cookie")
				return
			}

			claims, err := verifier.VerifySession(r.Context(), token)
			if err != nil {
				WriteProblemCode(w, r, http.StatusUnauthorized, httpx.CodeAuthInvalidSession, "invalid auth cookie")
				return
			}

			next.ServeHTTP(w, r.WithContext(WithSessionClaims(r.Context(), claims)))
		})
	}
}

func WithSessionClaims(ctx context.Context, claims identity.SessionClaims) context.Context {
	return context.WithValue(ctx, sessionClaimsContextKey{}, claims)
}

func SessionClaimsFromContext(ctx context.Context) (identity.SessionClaims, bool) {
	claims, ok := ctx.Value(sessionClaimsContextKey{}).(identity.SessionClaims)
	return claims, ok
}

func CurrentUserIDFromContext(ctx context.Context) (string, bool) {
	claims, ok := SessionClaimsFromContext(ctx)
	return claims.UserID, ok && claims.UserID != ""
}

func sessionCookie(r *http.Request, cookieName string) (string, bool) {
	cookie, err := r.Cookie(cookieName)
	if err != nil || cookie.Value == "" {
		return "", false
	}

	return cookie.Value, true
}
