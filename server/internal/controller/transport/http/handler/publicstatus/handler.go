package publicstatus

import (
	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	apppublic "github.com/yorukot/netstamp/internal/controller/application/publicstatus"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

type Handler struct {
	service       *apppublic.Service
	verifier      appauth.SessionManager
	cookieName    string
	tokenVerifier httpmiddleware.APITokenVerifier
}

func NewHandler(service *apppublic.Service, verifier appauth.SessionManager, cookieName string, tokenVerifiers ...httpmiddleware.APITokenVerifier) *Handler {
	var tokenVerifier httpmiddleware.APITokenVerifier
	if len(tokenVerifiers) > 0 {
		tokenVerifier = tokenVerifiers[0]
	}
	return &Handler{service: service, verifier: verifier, cookieName: cookieName, tokenVerifier: tokenVerifier}
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Get("/public/status-pages/{slug}/summary", h.handleGetPublicStatusSummary)
	api.Get("/public/status-pages/{slug}/elements", h.handleGetPublicStatusElements)
	api.Get("/public/status-pages/{slug}/incidents", h.handleGetPublicStatusIncidents)
	api.Get("/public/status-pages/{slug}/elements/{element_id}/chart", h.handleGetPublicStatusElementChart)
	api.Get("/public/status-pages/{slug}/elements/{element_id}/daily-status", h.handleGetPublicStatusElementDailyStatus)

	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireUserAuth(h.verifier, h.tokenVerifier, h.cookieName))

		r.With(httpmiddleware.RequireScope(identity.ScopeStatusPagesWrite)).Get("/public/status-pages/{slug}/editor-context", h.handleGetPublicStatusEditorContext)
		r.With(httpmiddleware.RequireScope(identity.ScopeStatusPagesRead)).Get("/projects/{ref}/status-pages", h.handleListPages)
		r.With(httpmiddleware.RequireScope(identity.ScopeStatusPagesWrite)).Post("/projects/{ref}/status-pages", h.handleCreatePage)
		r.With(httpmiddleware.RequireScope(identity.ScopeStatusPagesRead)).Get("/projects/{ref}/status-pages/{page_id}", h.handleGetPage)
		r.With(httpmiddleware.RequireScope(identity.ScopeStatusPagesWrite)).Patch("/projects/{ref}/status-pages/{page_id}", h.handleUpdatePage)
		r.With(httpmiddleware.RequireScope(identity.ScopeStatusPagesWrite)).Delete("/projects/{ref}/status-pages/{page_id}", h.handleDeletePage)

		r.With(httpmiddleware.RequireScope(identity.ScopeStatusPagesWrite)).Post("/projects/{ref}/status-pages/{page_id}/elements", h.handleCreateElement)
		r.With(httpmiddleware.RequireScope(identity.ScopeStatusPagesWrite)).Patch("/projects/{ref}/status-pages/{page_id}/elements/{element_id}", h.handleUpdateElement)
		r.With(httpmiddleware.RequireScope(identity.ScopeStatusPagesWrite)).Delete("/projects/{ref}/status-pages/{page_id}/elements/{element_id}", h.handleDeleteElement)
	})
}
