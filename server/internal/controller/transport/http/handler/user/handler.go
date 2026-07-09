package userhttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appuser "github.com/yorukot/netstamp/internal/controller/application/user"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

type Handler struct {
	service  *appuser.Service
	verifier appauth.TokenVerifier
}

func NewHandler(service *appuser.Service, verifier appauth.TokenVerifier) *Handler {
	return &Handler{
		service:  service,
		verifier: verifier,
	}
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireAuth(h.verifier))

		r.Patch("/users/me", h.handleUpdateCurrentUser)
		r.Post("/users/me/email-change", h.handleChangeCurrentUserEmail)
		r.Post("/users/me/password-change", h.handleChangeCurrentUserPassword)
		r.Post("/users/me/deactivation", h.handleDeactivateCurrentUser)
	})
}

func (h *Handler) handleUpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	var body updateCurrentUserInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.updateCurrentUser(r.Context(), &updateCurrentUserInput{Body: body})
	writeUserOutput(w, r, http.StatusOK, output, err)
}

func (h *Handler) handleChangeCurrentUserEmail(w http.ResponseWriter, r *http.Request) {
	var body changeCurrentUserEmailInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.changeCurrentUserEmail(r.Context(), &changeCurrentUserEmailInput{Body: body})
	writeUserOutput(w, r, http.StatusOK, output, err)
}

func (h *Handler) handleChangeCurrentUserPassword(w http.ResponseWriter, r *http.Request) {
	var body changeCurrentUserPasswordInputBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	if err := h.changeCurrentUserPassword(r.Context(), &changeCurrentUserPasswordInput{Body: body}); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) handleDeactivateCurrentUser(w http.ResponseWriter, r *http.Request) {
	if err := h.deactivateCurrentUser(r.Context()); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteNoContent(w)
}

func writeUserOutput(w http.ResponseWriter, r *http.Request, status int, output *userOutput, err error) {
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	httpx.WriteJSON(w, status, output.Body)
}
