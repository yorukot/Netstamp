package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	appapitoken "github.com/yorukot/netstamp/internal/controller/application/apitoken"
	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

const (
	LocalSessionCookieName      = "netstamp_session"
	ProductionSessionCookieName = "__Host-netstamp_session"
	SessionCookieName           = LocalSessionCookieName
)

type (
	sessionClaimsContextKey struct{}
	principalContextKey     struct{}
)

type AuthKind string

const (
	AuthKindSession  AuthKind = "session"
	AuthKindAPIToken AuthKind = "api_token"
)

type Principal struct {
	Kind         AuthKind
	UserID       string
	CredentialID string
	Token        *appapitoken.Principal
}

type APITokenVerifier interface {
	Verify(ctx context.Context, rawToken string) (appapitoken.Principal, error)
}

var (
	apiTokenAuthAttempts = mustAPITokenCounter("netstamp.api_token.auth_attempts")
	apiTokenRequests     = mustAPITokenCounter("netstamp.api_token.requests")
)

func mustAPITokenCounter(name string) metric.Int64Counter {
	counter, err := otel.Meter("github.com/yorukot/netstamp/internal/controller/transport/http/middleware/auth").Int64Counter(name)
	if err != nil {
		panic(err)
	}
	return counter
}

func RequireAuth(verifier appauth.SessionManager, cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.TrimSpace(r.Header.Get("Authorization")) != "" {
				writeInvalidAPIToken(w, r)
				return
			}
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

			ctx := WithSessionClaims(r.Context(), claims)
			ctx = WithPrincipal(ctx, Principal{Kind: AuthKindSession, UserID: claims.UserID, CredentialID: claims.SessionID})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireUserAuth(sessionVerifier appauth.SessionManager, tokenVerifier APITokenVerifier, cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorization := strings.TrimSpace(r.Header.Get("Authorization"))
			if authorization != "" {
				rawToken, ok := bearerToken(authorization)
				if !ok || tokenVerifier == nil {
					recordAPITokenAuth(r.Context(), "invalid")
					writeInvalidAPIToken(w, r)
					return
				}
				tokenPrincipal, err := tokenVerifier.Verify(r.Context(), rawToken)
				if err != nil {
					recordAPITokenAuth(r.Context(), "invalid")
					writeInvalidAPIToken(w, r)
					return
				}
				principal := Principal{Kind: AuthKindAPIToken, UserID: tokenPrincipal.UserID, CredentialID: tokenPrincipal.TokenID, Token: &tokenPrincipal}
				recordAPITokenAuth(r.Context(), "success")
				trace.SpanFromContext(r.Context()).SetAttributes(attribute.String("auth.type", string(AuthKindAPIToken)), attribute.String("api_token.id", tokenPrincipal.TokenID))
				wrapped := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
				next.ServeHTTP(wrapped, r.WithContext(WithPrincipal(r.Context(), principal)))
				recordAPITokenRequest(r, wrapped.Status())
				return
			}

			if sessionVerifier == nil {
				WriteProblem(w, r, http.StatusInternalServerError, "auth verifier unavailable")
				return
			}
			rawSession, ok := sessionCookie(r, cookieName)
			if !ok {
				WriteProblemCode(w, r, http.StatusUnauthorized, httpx.CodeAuthMissingSession, "missing auth cookie")
				return
			}
			claims, err := sessionVerifier.VerifySession(r.Context(), rawSession)
			if err != nil {
				WriteProblemCode(w, r, http.StatusUnauthorized, httpx.CodeAuthInvalidSession, "invalid auth cookie")
				return
			}
			ctx := WithSessionClaims(r.Context(), claims)
			ctx = WithPrincipal(ctx, Principal{Kind: AuthKindSession, UserID: claims.UserID, CredentialID: claims.SessionID})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func recordAPITokenAuth(ctx context.Context, outcome string) {
	apiTokenAuthAttempts.Add(ctx, 1, metric.WithAttributes(attribute.String("outcome", outcome)))
}

func recordAPITokenRequest(r *http.Request, status int) {
	if status == 0 {
		status = http.StatusOK
	}
	route := "unknown"
	if routeContext := chi.RouteContext(r.Context()); routeContext != nil && routeContext.RoutePattern() != "" {
		route = routeContext.RoutePattern()
	}
	apiTokenRequests.Add(r.Context(), 1, metric.WithAttributes(
		attribute.String("http.route", route),
		attribute.String("http.request.method", r.Method),
		attribute.String("http.response.status_class", strconv.Itoa(status/100)+"xx"),
	))
}

func RequireScope(scope identity.APITokenScope) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := PrincipalFromContext(r.Context())
			if !ok {
				WriteProblem(w, r, http.StatusUnauthorized, "missing authentication")
				return
			}
			if principal.Kind == AuthKindAPIToken && (principal.Token == nil || !principal.Token.HasScope(scope)) {
				WriteProblemCode(w, r, http.StatusForbidden, httpx.CodeAuthInsufficientScope, "api token does not include the required scope")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	return principal, ok
}

func WithSessionClaims(ctx context.Context, claims identity.SessionClaims) context.Context {
	return context.WithValue(ctx, sessionClaimsContextKey{}, claims)
}

func SessionClaimsFromContext(ctx context.Context) (identity.SessionClaims, bool) {
	claims, ok := ctx.Value(sessionClaimsContextKey{}).(identity.SessionClaims)
	return claims, ok
}

func CurrentUserIDFromContext(ctx context.Context) (string, bool) {
	if principal, ok := PrincipalFromContext(ctx); ok {
		return principal.UserID, principal.UserID != ""
	}
	claims, ok := SessionClaimsFromContext(ctx)
	return claims.UserID, ok && claims.UserID != ""
}

func bearerToken(value string) (string, bool) {
	parts := strings.Fields(value)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	return parts[1], parts[1] != ""
}

func writeInvalidAPIToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="netstamp"`)
	WriteProblemCode(w, r, http.StatusUnauthorized, httpx.CodeAuthInvalidAPIToken, "invalid api token")
}

func sessionCookie(r *http.Request, cookieName string) (string, bool) {
	cookie, err := r.Cookie(cookieName)
	if err != nil || cookie.Value == "" {
		return "", false
	}

	return cookie.Value, true
}
