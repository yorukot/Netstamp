package admin

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	appadmin "github.com/yorukot/netstamp/internal/controller/application/admin"
	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

type Handler struct {
	service  *appadmin.Service
	verifier appauth.TokenVerifier
}

func NewHandler(service *appadmin.Service, verifier appauth.TokenVerifier) *Handler {
	return &Handler{service: service, verifier: verifier}
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireAuth(h.verifier))

		r.Get("/admin/settings", h.handleGetSettings)
		r.Patch("/admin/settings", h.handleUpdateSettings)
	})
}

func (h *Handler) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError("admin settings service is unavailable"))
		return
	}
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	settings, err := h.service.GetSettings(r.Context(), appadmin.GetSettingsInput{CurrentUserID: userID})
	if err != nil {
		httpx.WriteProblem(w, r, mapAdminError(err, "get admin settings failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"settings": settingsResponse(settings)})
}

func (h *Handler) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError("admin settings service is unavailable"))
		return
	}
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	var body settingsBody
	if decodeErr := httpx.DecodeJSON(r, &body); decodeErr != nil {
		httpx.WriteProblem(w, r, decodeErr)
		return
	}
	settings, err := h.service.UpdateSettings(r.Context(), body.updateInput(userID))
	if err != nil {
		httpx.WriteProblem(w, r, mapAdminError(err, "update admin settings failed"))
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"settings": settingsResponse(settings)})
}

func currentUserID(r *http.Request) (string, error) {
	claims, ok := httpmiddleware.AccessTokenClaimsFromContext(r.Context())
	if !ok || claims.Subject == "" {
		return "", httpx.Unauthorized("missing auth session")
	}
	return claims.Subject, nil
}

func mapAdminError(err error, fallback string) error {
	switch {
	case errors.Is(err, appadmin.ErrForbidden):
		return httpx.Forbidden("system administrator access is required")
	case errors.Is(err, appadmin.ErrInvalidInput):
		return httpx.UnprocessableEntity("invalid admin settings")
	default:
		return httpx.InternalServerError(fallback)
	}
}
