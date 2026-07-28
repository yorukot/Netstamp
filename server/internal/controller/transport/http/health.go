package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	appadmin "github.com/yorukot/netstamp/internal/controller/application/admin"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

type rootBody struct {
	Message string `json:"message"`
}

type healthBody struct {
	Status string `json:"status"`
}

func registerSystemRoutes(api chi.Router, readinessCheck func(context.Context) error, settings *appadmin.Service, demoMode bool) {
	api.Get("/", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, rootBody{
			Message: "Netstamp API is running",
		})
	})

	api.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if readinessCheck == nil {
			httpx.WriteJSON(w, http.StatusOK, healthBody{
				Status: "ok",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := readinessCheck(ctx); err != nil {
			httpx.WriteProblem(w, r, httpx.ServiceUnavailable("readiness check failed"))
			return
		}

		httpx.WriteJSON(w, http.StatusOK, healthBody{
			Status: "ok",
		})
	})

	api.Get("/system/config", func(w http.ResponseWriter, r *http.Request) {
		config := appadmin.Settings{RegistrationEnabled: true, ProjectCreationEnabled: true, CredentialChangesEnabled: true}
		if settings != nil {
			if effective, err := settings.EffectiveSettings(r.Context()); err == nil {
				config = effective
			}
		}
		registrationEnabled := config.RegistrationEnabled && !demoMode
		projectCreationEnabled := config.ProjectCreationEnabled && !demoMode
		credentialChangesEnabled := config.CredentialChangesEnabled && !demoMode
		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"demoMode": demoMode,
			"capabilities": map[string]bool{
				"registrationEnabled":      registrationEnabled,
				"projectCreationEnabled":   projectCreationEnabled,
				"credentialChangesEnabled": credentialChangesEnabled,
			},
			"tracking": map[string]string{
				"googleTagId": config.Tracking.GoogleTagID, "clarityProjectId": config.Tracking.ClarityProjectID,
				"metaPixelId": config.Tracking.MetaPixelID, "postHogKey": config.Tracking.PostHogKey,
				"postHogHost": config.Tracking.PostHogHost, "plausibleDomain": config.Tracking.PlausibleDomain,
				"plausibleScriptUrl": config.Tracking.PlausibleScriptURL, "umamiWebsiteId": config.Tracking.UmamiWebsiteID,
				"umamiScriptUrl": config.Tracking.UmamiScriptURL, "consentMode": config.Tracking.ConsentMode,
				"consentCountries": config.Tracking.ConsentCountries,
			},
		})
	})
}
