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
	service    *appuser.Service
	verifier   appauth.SessionManager
	cookieName string
	sudo       *appauth.Service
}

func NewHandler(service *appuser.Service, verifier appauth.SessionManager, cookieName string, sudoServices ...*appauth.Service) *Handler {
	handler := &Handler{
		service:    service,
		verifier:   verifier,
		cookieName: cookieName,
	}
	if len(sudoServices) > 0 {
		handler.sudo = sudoServices[0]
	}
	return handler
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireAuth(h.verifier, h.cookieName))

		r.Patch("/users/me", h.handleUpdateCurrentUser)
		r.Get("/users/me/authentication-methods", h.handleAuthenticationMethods)
		if h.sudo != nil {
			r.With(httpmiddleware.RequirePasswordChangeAuthorization(h.sudo)).Put("/users/me/password", h.handleChangeCurrentUserPassword)
			r.Group(func(sensitive chi.Router) {
				sensitive.Use(httpmiddleware.RequireSudo(h.sudo))
				sensitive.Post("/users/me/email-change", h.handleChangeCurrentUserEmail)
				sensitive.Delete("/users/me/password", h.handleRemoveCurrentUserPassword)
				sensitive.Delete("/users/me/identities/{identity_id}", h.handleRemoveCurrentUserIdentity)
				sensitive.Post("/users/me/deactivation", h.handleDeactivateCurrentUser)
			})
		} else {
			r.Post("/users/me/email-change", h.handleChangeCurrentUserEmail)
			r.Post("/users/me/password-change", h.handleChangeCurrentUserPassword)
			r.Post("/users/me/deactivation", h.handleDeactivateCurrentUser)
		}
	})
}

func (h *Handler) handleAuthenticationMethods(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r.Context())
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	output, err := h.service.ListAuthenticationMethods(r.Context(), userID)
	if err != nil {
		httpx.WriteProblem(w, r, mapUserError(err, "list authentication methods failed"))
		return
	}
	identities := make([]map[string]any, 0, len(output.Identities))
	for _, item := range output.Identities {
		identities = append(identities, map[string]any{
			"id": item.ID, "provider": item.Provider, "issuer": item.Issuer, "email": item.Email,
			"emailVerified": item.EmailVerified, "displayName": item.DisplayName, "username": item.Username,
			"avatarUrl": item.AvatarURL, "createdAt": item.CreatedAt, "lastLoginAt": item.LastLoginAt,
		})
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"hasPassword": output.HasPassword, "identities": identities})
}

func (h *Handler) handleRemoveCurrentUserPassword(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r.Context())
	if err == nil {
		err = h.service.RemoveCurrentUserPassword(r.Context(), userID, currentSessionID(r.Context()))
	}
	if err != nil {
		httpx.WriteProblem(w, r, mapUserError(err, "remove password failed"))
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) handleRemoveCurrentUserIdentity(w http.ResponseWriter, r *http.Request) {
	userID, err := currentUserID(r.Context())
	if err == nil {
		err = h.service.RemoveCurrentUserIdentity(r.Context(), userID, currentSessionID(r.Context()), httpx.Path(r, "identity_id"))
	}
	if err != nil {
		httpx.WriteProblem(w, r, mapUserError(err, "remove identity failed"))
		return
	}
	httpx.WriteNoContent(w)
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
