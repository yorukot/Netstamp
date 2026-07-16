package auth

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

const localOIDCFlowCookieName = "netstamp_oidc_flow"

func (h *Handler) handleAuthMethods(w http.ResponseWriter, r *http.Request) {
	registrationEnabled := h.registrationEnabled
	if h.settings != nil {
		if settings, err := h.settings.EffectiveSettings(r.Context()); err == nil {
			registrationEnabled = settings.RegistrationEnabled
		}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"password": map[string]any{"enabled": true, "registrationEnabled": registrationEnabled},
		"oidc":     map[string]any{"enabled": h.oidcEnabled, "displayName": h.oidcDisplayName},
	})
}

func (h *Handler) handleOIDCStart(w http.ResponseWriter, r *http.Request) {
	intent := r.URL.Query().Get("intent")
	if intent == "" {
		intent = appauth.OIDCIntentLogin
	}
	sessionID := ""
	if intent != appauth.OIDCIntentLogin {
		cookie, err := r.Cookie(h.cookieName)
		if err != nil || cookie.Value == "" {
			httpx.WriteProblem(w, r, httpx.UnauthorizedCode(httpx.CodeAuthMissingSession, "authentication required"))
			return
		}
		claims, err := h.verifier.VerifySession(r.Context(), cookie.Value)
		if err != nil {
			httpx.WriteProblem(w, r, httpx.UnauthorizedCode(httpx.CodeAuthInvalidSession, "invalid session"))
			return
		}
		sessionID = claims.SessionID
	}
	result, err := h.service.StartOIDC(r.Context(), appauth.StartOIDCInput{Intent: intent, SessionID: sessionID, ReturnTo: r.URL.Query().Get("returnTo")})
	if err != nil {
		switch {
		case errors.Is(err, appauth.ErrSudoRequired):
			httpx.WriteProblem(w, r, httpx.ForbiddenCode(httpx.CodeAuthSudoRequired, "recent authentication required"))
		case errors.Is(err, appauth.ErrOIDCUnavailable):
			httpx.WriteProblem(w, r, httpx.ServiceUnavailableCode(httpx.CodeAuthOIDCUnavailable, "single sign-on is unavailable"))
		default:
			httpx.WriteProblem(w, r, httpx.BadRequest("invalid OIDC request"))
		}
		return
	}
	flowCookie := newOIDCFlowCookie(h.oidcFlowCookieName(), result.BrowserToken, result.ExpiresAt, h.cookieSecure)
	http.SetCookie(w, &flowCookie)
	http.Redirect(w, r, result.AuthorizationURL, http.StatusFound)
}

func (h *Handler) handleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	flowCookie, err := r.Cookie(h.oidcFlowCookieName())
	if err != nil {
		h.redirectOIDCError(w, r, "callback_invalid")
		return
	}
	expiredFlowCookie := expiredOIDCFlowCookie(h.oidcFlowCookieName(), h.cookieSecure)
	http.SetCookie(w, &expiredFlowCookie)
	result, err := h.service.CompleteOIDC(r.Context(), appauth.CompleteOIDCInput{Code: r.URL.Query().Get("code"), State: r.URL.Query().Get("state"), BrowserToken: flowCookie.Value, UserAgent: r.UserAgent()})
	if err != nil {
		code := "callback_invalid"
		switch {
		case errors.Is(err, appauth.ErrIdentityConflict):
			code = "identity_conflict"
		case errors.Is(err, appauth.ErrJITProvisioningDisabled):
			code = "jit_disabled"
		case errors.Is(err, appauth.ErrAccountDisabled):
			code = "account_disabled"
		}
		h.redirectOIDCError(w, r, code)
		return
	}
	if result.Access != nil {
		cookie := newSessionCookie(h.cookieName, result.Access.SessionToken, result.Access.ExpiresIn, h.cookieSecure)
		http.SetCookie(w, &cookie)
	}
	http.Redirect(w, r, h.webRedirect(result.ReturnTo), http.StatusFound)
}

func (h *Handler) redirectOIDCError(w http.ResponseWriter, r *http.Request, code string) {
	values := url.Values{}
	values.Set("oidc_error", code)
	http.Redirect(w, r, h.webRedirect("/login?"+values.Encode()), http.StatusFound)
}

func (h *Handler) webRedirect(path string) string {
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.ContainsAny(path, "\\\r\n") {
		path = "/"
	}
	base := strings.TrimRight(h.publicWebBaseURL, "/")
	if base == "" {
		return path
	}
	return base + path
}

func (h *Handler) oidcFlowCookieName() string {
	if h.cookieSecure {
		return "__Host-netstamp_oidc_flow"
	}
	return localOIDCFlowCookieName
}
