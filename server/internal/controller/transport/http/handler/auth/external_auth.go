package auth

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

const localExternalAuthFlowCookieName = "netstamp_external_auth_flow"

func (h *Handler) handleAuthMethods(w http.ResponseWriter, r *http.Request) {
	registrationEnabled := h.registrationEnabled
	if h.settings != nil {
		if settings, err := h.settings.EffectiveSettings(r.Context()); err == nil {
			registrationEnabled = settings.RegistrationEnabled
		}
	}
	providers := h.service.ExternalProviderMethods()
	providerResponses := make([]externalProviderMethodResponse, 0, len(providers))
	oidc := map[string]any{"enabled": false, "displayName": "Single sign-on"}
	for _, provider := range providers {
		providerResponses = append(providerResponses, externalProviderMethodResponse{
			ID: provider.ID, DisplayName: provider.DisplayName, SudoCapable: provider.SudoCapable,
		})
		if provider.ID == identity.AuthenticationMethodOIDC {
			oidc = map[string]any{"enabled": true, "displayName": provider.DisplayName}
			break
		}
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"password":  map[string]any{"enabled": true, "registrationEnabled": registrationEnabled},
		"providers": providerResponses,
		"oidc":      oidc,
	})
}

type externalProviderMethodResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	SudoCapable bool   `json:"sudoCapable"`
}

func (h *Handler) handleExternalAuthStart(w http.ResponseWriter, r *http.Request) {
	h.handleExternalAuthStartForProvider(w, r, chi.URLParam(r, "provider"))
}

func (h *Handler) handleOIDCStart(w http.ResponseWriter, r *http.Request) {
	h.handleExternalAuthStartForProvider(w, r, identity.AuthenticationMethodOIDC)
}

func (h *Handler) handleExternalAuthStartForProvider(w http.ResponseWriter, r *http.Request, provider string) {
	intent := r.URL.Query().Get("intent")
	if intent == "" {
		intent = appauth.ExternalAuthIntentLogin
	}
	sessionID := ""
	if intent != appauth.ExternalAuthIntentLogin {
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
	result, err := h.service.StartExternalAuth(r.Context(), appauth.StartExternalAuthInput{
		Provider: provider, Intent: intent, SessionID: sessionID, ReturnTo: r.URL.Query().Get("returnTo"),
	})
	if err != nil {
		switch {
		case errors.Is(err, appauth.ErrSudoRequired):
			httpx.WriteProblem(w, r, httpx.ForbiddenCode(httpx.CodeAuthSudoRequired, "recent authentication required"))
		case errors.Is(err, appauth.ErrExternalAuthSudoUnsupported):
			httpx.WriteProblem(w, r, httpx.BadRequest("this provider cannot be used for recent authentication"))
		case errors.Is(err, appauth.ErrExternalAuthUnavailable):
			httpx.WriteProblem(w, r, httpx.ServiceUnavailableCode(httpx.CodeAuthOIDCUnavailable, "external authentication is unavailable"))
		default:
			httpx.WriteProblem(w, r, httpx.BadRequest("invalid external authentication request"))
		}
		return
	}
	flowCookie := newExternalAuthFlowCookie(h.externalAuthFlowCookieName(), result.BrowserToken, result.ExpiresAt, h.cookieSecure)
	http.SetCookie(w, &flowCookie)
	http.Redirect(w, r, result.AuthorizationURL, http.StatusFound)
}

func (h *Handler) handleExternalAuthCallback(w http.ResponseWriter, r *http.Request) {
	h.handleExternalAuthCallbackForProvider(w, r, chi.URLParam(r, "provider"))
}

func (h *Handler) handleOIDCCallback(w http.ResponseWriter, r *http.Request) {
	h.handleExternalAuthCallbackForProvider(w, r, identity.AuthenticationMethodOIDC)
}

func (h *Handler) handleExternalAuthCallbackForProvider(w http.ResponseWriter, r *http.Request, provider string) {
	flowCookie, err := r.Cookie(h.externalAuthFlowCookieName())
	if err != nil {
		h.redirectExternalAuthError(w, r, provider, "callback_invalid", "")
		return
	}
	expiredFlowCookie := expiredExternalAuthFlowCookie(h.externalAuthFlowCookieName(), h.cookieSecure)
	http.SetCookie(w, &expiredFlowCookie)
	result, err := h.service.CompleteExternalAuth(r.Context(), appauth.CompleteExternalAuthInput{
		Provider: provider, Code: r.URL.Query().Get("code"), State: r.URL.Query().Get("state"),
		BrowserToken: flowCookie.Value, UserAgent: r.UserAgent(),
	})
	if err != nil {
		code := "callback_invalid"
		switch {
		case errors.Is(err, appauth.ErrIdentityConflict):
			code = "identity_conflict"
		case errors.Is(err, appauth.ErrJITProvisioningDisabled):
			code = "jit_disabled"
		case errors.Is(err, appauth.ErrAccountDisabled):
			code = "account_disabled"
		case errors.Is(err, appauth.ErrSudoRequired):
			code = "sudo_expired"
		}
		returnTo := ""
		if result.Intent != appauth.ExternalAuthIntentLogin {
			returnTo = result.ReturnTo
		}
		h.redirectExternalAuthError(w, r, provider, code, returnTo)
		return
	}
	if result.Access != nil {
		cookie := newSessionCookie(h.cookieName, result.Access.SessionToken, result.Access.ExpiresIn, h.cookieSecure)
		http.SetCookie(w, &cookie)
	}
	http.Redirect(w, r, h.webRedirect(result.ReturnTo), http.StatusFound)
}

func (h *Handler) redirectExternalAuthError(w http.ResponseWriter, r *http.Request, provider, code, returnTo string) {
	path := normalizeExternalAuthErrorReturnTo(returnTo)
	parsed, err := url.Parse(path)
	if err != nil {
		parsed = &url.URL{Path: "/login"}
	}
	values := parsed.Query()
	values.Set("auth_error", code)
	values.Set("auth_provider", provider)
	parsed.RawQuery = values.Encode()
	http.Redirect(w, r, h.webRedirect(parsed.String()), http.StatusFound)
}

func normalizeExternalAuthErrorReturnTo(returnTo string) string {
	if returnTo == "" {
		return "/login"
	}
	parsed, err := url.Parse(returnTo)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || !strings.HasPrefix(parsed.Path, "/") || strings.HasPrefix(parsed.Path, "//") || strings.ContainsAny(parsed.Path, "\\\r\n") {
		return "/login"
	}
	return parsed.String()
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

func (h *Handler) externalAuthFlowCookieName() string {
	if h.cookieSecure {
		return "__Host-netstamp_external_auth_flow"
	}
	return localExternalAuthFlowCookieName
}
