package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

type Handler struct {
	service      *appauth.Service
	verifier     appauth.TokenVerifier
	cookieSecure bool
}

func NewHandler(service *appauth.Service, verifier appauth.TokenVerifier, cookieSecure bool) *Handler {
	return &Handler{
		service:      service,
		verifier:     verifier,
		cookieSecure: cookieSecure,
	}
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Post("/auth/register", h.handleRegister)
	api.Post("/auth/login", h.handleLogin)
	api.Post("/auth/logout", h.handleLogout)
	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireAuth(h.verifier))

		r.Get("/auth/me", h.handleMe)
	})
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
