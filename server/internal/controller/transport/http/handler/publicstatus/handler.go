package publicstatus

import (
	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	apppublic "github.com/yorukot/netstamp/internal/controller/application/publicstatus"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

type Handler struct {
	service  *apppublic.Service
	verifier appauth.TokenVerifier
}

func NewHandler(service *apppublic.Service, verifier appauth.TokenVerifier) *Handler {
	return &Handler{service: service, verifier: verifier}
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Get("/public/status-pages/{slug}", h.handleGetPublicStatusPage)

	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireAuth(h.verifier))

		r.Get("/projects/{ref}/status-pages", h.handleListPages)
		r.Post("/projects/{ref}/status-pages", h.handleCreatePage)
		r.Get("/projects/{ref}/status-pages/{page_id}", h.handleGetPage)
		r.Patch("/projects/{ref}/status-pages/{page_id}", h.handleUpdatePage)
		r.Delete("/projects/{ref}/status-pages/{page_id}", h.handleDeletePage)

		r.Post("/projects/{ref}/status-pages/{page_id}/elements", h.handleCreateElement)
		r.Patch("/projects/{ref}/status-pages/{page_id}/elements/{element_id}", h.handleUpdateElement)
		r.Delete("/projects/{ref}/status-pages/{page_id}/elements/{element_id}", h.handleDeleteElement)
	})
}
