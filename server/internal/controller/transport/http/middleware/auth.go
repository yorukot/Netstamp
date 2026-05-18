package middleware

import (
	"context"
	"net/http"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

const SessionCookieName = "netstamp_session"

type accessTokenClaimsContextKey struct{}

func RequireAuth(verifier appauth.TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if verifier == nil {
				WriteProblem(w, r, http.StatusInternalServerError, "auth verifier unavailable")
				return
			}

			token, ok := sessionCookie(r)
			if !ok {
				WriteProblem(w, r, http.StatusUnauthorized, "missing auth cookie")
				return
			}

			claims, err := verifier.VerifyAccessToken(r.Context(), token)
			if err != nil {
				WriteProblem(w, r, http.StatusUnauthorized, "invalid auth cookie")
				return
			}

			next.ServeHTTP(w, r.WithContext(WithAccessTokenClaims(r.Context(), claims)))
		})
	}
}

func WithAccessTokenClaims(ctx context.Context, claims identity.AccessTokenClaims) context.Context {
	return context.WithValue(ctx, accessTokenClaimsContextKey{}, claims)
}

func AccessTokenClaimsFromContext(ctx context.Context) (identity.AccessTokenClaims, bool) {
	claims, ok := ctx.Value(accessTokenClaimsContextKey{}).(identity.AccessTokenClaims)
	return claims, ok
}

func sessionCookie(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return "", false
	}

	return cookie.Value, true
}
