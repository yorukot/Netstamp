package alert

import (
	"context"

	"github.com/go-chi/chi/v5"

	appalert "github.com/yorukot/netstamp/internal/controller/application/alert"
	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

type Handler struct {
	service       *appalert.Service
	verifier      appauth.SessionManager
	cookieName    string
	smtpSettings  SMTPStatusProvider
	tokenVerifier httpmiddleware.APITokenVerifier
}

type SMTPStatusProvider interface {
	SMTPConfigured(ctx context.Context) bool
}

func NewHandler(service *appalert.Service, verifier appauth.SessionManager, cookieName string, smtpSettings SMTPStatusProvider, tokenVerifiers ...httpmiddleware.APITokenVerifier) *Handler {
	var tokenVerifier httpmiddleware.APITokenVerifier
	if len(tokenVerifiers) > 0 {
		tokenVerifier = tokenVerifiers[0]
	}
	return &Handler{service: service, verifier: verifier, cookieName: cookieName, smtpSettings: smtpSettings, tokenVerifier: tokenVerifier}
}

func (h *Handler) emailSMTPConfigured(ctx context.Context) bool {
	return h.smtpSettings != nil && h.smtpSettings.SMTPConfigured(ctx)
}

func (h *Handler) RegisterRoutes(api chi.Router) {
	api.Group(func(r chi.Router) {
		r.Use(httpmiddleware.RequireUserAuth(h.verifier, h.tokenVerifier, h.cookieName))

		r.With(httpmiddleware.RequireScope(identity.ScopeAlertsRead)).Get("/projects/{ref}/alerts/rules", h.handleListRules)
		r.With(httpmiddleware.RequireScope(identity.ScopeAlertsWrite)).Post("/projects/{ref}/alerts/rules", h.handleCreateRule)
		r.With(httpmiddleware.RequireScope(identity.ScopeAlertsRead)).Get("/projects/{ref}/alerts/rules/{rule_id}", h.handleGetRule)
		r.With(httpmiddleware.RequireScope(identity.ScopeAlertsWrite)).Patch("/projects/{ref}/alerts/rules/{rule_id}", h.handleUpdateRule)
		r.With(httpmiddleware.RequireScope(identity.ScopeAlertsWrite)).Delete("/projects/{ref}/alerts/rules/{rule_id}", h.handleDeleteRule)

		r.With(httpmiddleware.RequireScope(identity.ScopeAlertsRead)).Get("/projects/{ref}/alerts/incidents", h.handleListIncidents)
		r.With(httpmiddleware.RequireScope(identity.ScopeAlertsRead)).Get("/projects/{ref}/alerts/incidents/{incident_id}", h.handleGetIncident)

		r.With(httpmiddleware.RequireScope(identity.ScopeAlertsRead)).Get("/projects/{ref}/alerts/notifications", h.handleListNotifications)
		r.With(httpmiddleware.RequireScope(identity.ScopeAlertsWrite)).Post("/projects/{ref}/alerts/notifications", h.handleCreateNotification)
		r.With(httpmiddleware.RequireScope(identity.ScopeAlertsRead)).Get("/projects/{ref}/alerts/notifications/{notification_id}", h.handleGetNotification)
		r.With(httpmiddleware.RequireScope(identity.ScopeAlertsWrite)).Patch("/projects/{ref}/alerts/notifications/{notification_id}", h.handleUpdateNotification)
		r.With(httpmiddleware.RequireScope(identity.ScopeAlertsWrite)).Post("/projects/{ref}/alerts/notifications/{notification_id}/test", h.handleTestNotification)
		r.With(httpmiddleware.RequireScope(identity.ScopeAlertsWrite)).Delete("/projects/{ref}/alerts/notifications/{notification_id}", h.handleDeleteNotification)
	})
}
