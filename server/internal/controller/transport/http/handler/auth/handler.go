package auth

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	appadmin "github.com/yorukot/netstamp/internal/controller/application/admin"
	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

type Handler struct {
	service             *appauth.Service
	verifier            appauth.SessionManager
	settings            *appadmin.Service
	cookieName          string
	cookieSecure        bool
	registrationEnabled bool
	publicWebBaseURL    string
	resetLimiter        *PasswordResetRateLimiter
}

func NewHandler(service *appauth.Service, verifier appauth.SessionManager, settings *appadmin.Service, cookieName string, cookieSecure, registrationEnabled bool) *Handler {
	return &Handler{
		service:             service,
		verifier:            verifier,
		settings:            settings,
		cookieName:          cookieName,
		cookieSecure:        cookieSecure,
		registrationEnabled: registrationEnabled,
	}
}

func (h *Handler) ConfigurePasswordReset(publicWebBaseURL string, limiter *PasswordResetRateLimiter) *Handler {
	h.publicWebBaseURL = publicWebBaseURL
	h.resetLimiter = limiter
	return h
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Post("/auth/register", h.handleRegister)
	api.Post("/auth/login", h.handleLogin)
	api.Post("/auth/logout", h.handleLogout)
	api.Post("/auth/password-resets", h.handleRequestPasswordReset)
	api.Patch("/auth/password-resets", h.handleConfirmPasswordReset)
	api.Post("/auth/email-verifications", h.handleRequestEmailVerification)
	api.Patch("/auth/email-verifications", h.handleConfirmEmailVerification)
	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireAuth(h.verifier, h.cookieName))

		r.Get("/auth/csrf", h.handleCSRF)
		r.Get("/auth/me", h.handleMe)
		r.Get("/auth/sessions", h.handleListSessions)
		r.Delete("/auth/sessions", h.handleRevokeAllSessions)
		r.Delete("/auth/sessions/{session_id}", h.handleRevokeSession)
	})
}

func (h *Handler) handleRequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var body requestPasswordResetInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	h.handleAcceptedEmailAction(w, r, body.Email, "too many password reset requests", func(ctx context.Context) error {
		return h.requestPasswordReset(ctx, r, &requestPasswordResetInput{Body: body})
	})
}

func (h *Handler) handleConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var body confirmPasswordResetInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	if err := h.confirmPasswordReset(r.Context(), &confirmPasswordResetInput{Body: body}); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var body registerInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.register(r.Context(), r, &registerInput{Body: body})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	if output.SetCookie != nil {
		http.SetCookie(w, output.SetCookie)
	}
	httpx.WriteJSON(w, output.Status, output.Body)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body loginInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.login(r.Context(), r, &loginInput{Body: body})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	http.SetCookie(w, &output.SetCookie)
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	rawSessionToken := ""
	if cookie, err := r.Cookie(h.cookieName); err == nil {
		rawSessionToken = cookie.Value
	}
	output, err := h.logout(r.Context(), &logoutInput{RawSessionToken: rawSessionToken})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	http.SetCookie(w, &output.SetCookie)
	httpx.WriteNoContent(w)
}

func (h *Handler) handleCSRF(w http.ResponseWriter, r *http.Request) {
	output, err := h.csrf(r.Context(), &csrfInput{})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	output, err := h.me(r.Context(), &meInput{})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleListSessions(w http.ResponseWriter, r *http.Request) {
	output, err := h.listSessions(r.Context(), &listSessionsInput{})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleRevokeSession(w http.ResponseWriter, r *http.Request) {
	output, err := h.revokeSession(r.Context(), &revokeSessionInput{SessionID: httpx.Path(r, "session_id")})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	if output.SetCookie != nil {
		http.SetCookie(w, output.SetCookie)
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) handleRevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	output, err := h.revokeAllSessions(r.Context(), &revokeAllSessionsInput{})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	http.SetCookie(w, &output.SetCookie)
	httpx.WriteNoContent(w)
}

func (h *Handler) handleRequestEmailVerification(w http.ResponseWriter, r *http.Request) {
	var body requestEmailVerificationInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	h.handleAcceptedEmailAction(w, r, body.Email, "too many email verification requests", func(ctx context.Context) error {
		return h.requestEmailVerification(ctx, r, &requestEmailVerificationInput{Body: body})
	})
}

func (h *Handler) handleConfirmEmailVerification(w http.ResponseWriter, r *http.Request) {
	var body confirmEmailVerificationInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	if err := h.confirmEmailVerification(r.Context(), &confirmEmailVerificationInput{Body: body}); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) handleAcceptedEmailAction(w http.ResponseWriter, r *http.Request, email, rateLimitDetail string, action func(context.Context) error) {
	if h.resetLimiter != nil && !h.resetLimiter.Allow(r.Context(), resetLimiterClientKey(r), email) {
		httpx.WriteProblem(w, r, httpx.NewErrorCode(http.StatusTooManyRequests, httpx.CodeRateLimited, rateLimitDetail))
		return
	}
	if err := action(r.Context()); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
