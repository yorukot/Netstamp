package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

type Handler struct {
	service             *appauth.Service
	verifier            appauth.TokenVerifier
	cookieSecure        bool
	registrationEnabled bool
	publicWebBaseURL    string
	resetLimiter        *PasswordResetRateLimiter
}

func NewHandler(service *appauth.Service, verifier appauth.TokenVerifier, cookieSecure, registrationEnabled bool) *Handler {
	return &Handler{
		service:             service,
		verifier:            verifier,
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
	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireAuth(h.verifier))

		r.Get("/auth/me", h.handleMe)
	})
}

func (h *Handler) handleRequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var body requestPasswordResetInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	if h.resetLimiter != nil && !h.resetLimiter.Allow(r.Context(), resetLimiterClientKey(r), body.Email) {
		httpx.WriteProblem(w, r, httpx.NewError(http.StatusTooManyRequests, "too many password reset requests"))
		return
	}
	if err := h.requestPasswordReset(r.Context(), r, &requestPasswordResetInput{Body: body}); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
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
	output, err := h.register(r.Context(), &registerInput{Body: body})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	http.SetCookie(w, &output.SetCookie)
	httpx.WriteJSON(w, http.StatusCreated, output.Body)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body loginInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.login(r.Context(), &loginInput{Body: body})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	http.SetCookie(w, &output.SetCookie)
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	output, err := h.logout(r.Context(), &logoutInput{})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	http.SetCookie(w, &output.SetCookie)
	httpx.WriteNoContent(w)
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	output, err := h.me(r.Context(), &meInput{})
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output.Body)
}
