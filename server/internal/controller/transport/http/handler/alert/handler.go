package alert

import (
	"github.com/go-chi/chi/v5"

	appalert "github.com/yorukot/netstamp/internal/controller/application/alert"
	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
)

type Handler struct {
	service  *appalert.Service
	verifier appauth.TokenVerifier
}

func NewHandler(service *appalert.Service, verifier appauth.TokenVerifier) *Handler {
	return &Handler{service: service, verifier: verifier}
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireAuth(h.verifier))

		r.Get("/projects/{ref}/alerts/rules", h.handleListRules)
		r.Post("/projects/{ref}/alerts/rules", h.handleCreateRule)
		r.Get("/projects/{ref}/alerts/rules/{rule_id}", h.handleGetRule)
		r.Patch("/projects/{ref}/alerts/rules/{rule_id}", h.handleUpdateRule)
		r.Delete("/projects/{ref}/alerts/rules/{rule_id}", h.handleDeleteRule)

		r.Get("/projects/{ref}/alerts/incidents", h.handleListIncidents)
		r.Get("/projects/{ref}/alerts/incidents/{incident_id}", h.handleGetIncident)

		r.Get("/projects/{ref}/alerts/notifications", h.handleListNotifications)
		r.Post("/projects/{ref}/alerts/notifications", h.handleCreateNotification)
		r.Get("/projects/{ref}/alerts/notifications/{notification_id}", h.handleGetNotification)
		r.Patch("/projects/{ref}/alerts/notifications/{notification_id}", h.handleUpdateNotification)
		r.Post("/projects/{ref}/alerts/notifications/{notification_id}/test", h.handleTestNotification)
		r.Delete("/projects/{ref}/alerts/notifications/{notification_id}", h.handleDeleteNotification)
	})
}
