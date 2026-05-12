package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	controllerlogger "github.com/yorukot/netstamp/internal/controller/logger"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

type accessTokenClaimsContextKey struct{}

func RequireAuth(verifier appauth.TokenVerifier) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		requestCtx := ctx.Context()

		if verifier == nil {
			writeHumaProblem(requestCtx, ctx, http.StatusInternalServerError, "auth verifier unavailable")
			return
		}

		token, ok := bearerToken(ctx.Header("Authorization"))
		if !ok {
			writeHumaProblem(requestCtx, ctx, http.StatusUnauthorized, "missing bearer token")
			return
		}

		claims, err := verifier.VerifyAccessToken(requestCtx, token)
		if err != nil {
			writeHumaProblem(requestCtx, ctx, http.StatusUnauthorized, "invalid access token")
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

func bearerToken(value string) (string, bool) {
	parts := strings.Fields(value)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	return parts[1], true
}

func writeHumaProblem(requestCtx context.Context, ctx huma.Context, status int, detail string) {
	if requestID := chimw.GetReqID(requestCtx); requestID != "" {
		ctx.SetHeader("X-Request-ID", requestID)
	}
	if status == http.StatusUnauthorized {
		ctx.SetHeader("WWW-Authenticate", "Bearer")
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
