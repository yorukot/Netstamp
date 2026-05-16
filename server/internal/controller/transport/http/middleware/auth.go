package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	controllerlogger "github.com/yorukot/netstamp/internal/controller/logger"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

const (
	SessionCookieName           = "netstamp_session"
	SessionCookieSecurityScheme = "sessionCookieAuth"
)

type accessTokenClaimsContextKey struct{}

func RequireAuth(verifier appauth.TokenVerifier) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		requestCtx := ctx.Context()

		if verifier == nil {
			writeHumaProblem(requestCtx, ctx, http.StatusInternalServerError, "auth verifier unavailable")
			return
		}

		token, ok := sessionCookie(ctx)
		if !ok {
			writeHumaProblem(requestCtx, ctx, http.StatusUnauthorized, "missing auth cookie")
			return
		}

		claims, err := verifier.VerifyAccessToken(requestCtx, token)
		if err != nil {
			writeHumaProblem(requestCtx, ctx, http.StatusUnauthorized, "invalid auth cookie")
			return
		}

		next(huma.WithContext(ctx, WithAccessTokenClaims(requestCtx, claims)))
	}
}

func WithAccessTokenClaims(ctx context.Context, claims identity.AccessTokenClaims) context.Context {
	return context.WithValue(ctx, accessTokenClaimsContextKey{}, claims)
}

func AccessTokenClaimsFromContext(ctx context.Context) (identity.AccessTokenClaims, bool) {
	claims, ok := ctx.Value(accessTokenClaimsContextKey{}).(identity.AccessTokenClaims)
	return claims, ok
}

func sessionCookie(ctx huma.Context) (string, bool) {
	request := http.Request{
		Header: http.Header{"Cookie": []string{ctx.Header("Cookie")}},
	}
	cookie, err := request.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		return "", false
	}

	return cookie.Value, true
}

func writeHumaProblem(requestCtx context.Context, ctx huma.Context, status int, detail string) {
	if requestID := chimw.GetReqID(requestCtx); requestID != "" {
		ctx.SetHeader("X-Request-ID", requestID)
	}
	ctx.SetHeader("Content-Type", "application/problem+json")
	ctx.SetStatus(status)

	if err := json.NewEncoder(ctx.BodyWriter()).Encode(&huma.ErrorModel{
		Status: status,
		Title:  http.StatusText(status),
		Detail: detail,
	}); err != nil {
		log := controllerlogger.FromContext(requestCtx, zap.L())
		log.Warn("failed to write auth problem response", zap.Error(err))
	}
}
