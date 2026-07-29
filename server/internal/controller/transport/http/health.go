package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

type rootBody struct {
	Message string `json:"message"`
}

type healthBody struct {
	Status string `json:"status"`
}

type PublicAccessSettings struct {
	AccountCreationEnabled   bool
	ProjectCreationEnabled   bool
	CredentialChangesEnabled bool
}

type PublicAccessSettingsProvider interface {
	PublicAccessSettings(ctx context.Context) (PublicAccessSettings, error)
}

type publicRuntimeConfigBody struct {
	DemoMode     bool                   `json:"demoMode"`
	Capabilities publicCapabilitiesBody `json:"capabilities"`
}

type publicCapabilitiesBody struct {
	AccountCreationEnabled   bool `json:"accountCreationEnabled"`
	ProjectCreationEnabled   bool `json:"projectCreationEnabled"`
	CredentialChangesEnabled bool `json:"credentialChangesEnabled"`
}

func registerSystemRoutes(api chi.Router, readinessCheck func(context.Context) error, settings PublicAccessSettingsProvider, demoMode bool) {
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
		w.Header().Set("Cache-Control", "no-store")
		if settings == nil {
			httpx.WriteProblem(w, r, httpx.InternalServerError("system settings unavailable"))
			return
		}

		access, err := settings.PublicAccessSettings(r.Context())
		if err != nil {
			httpx.WriteProblem(w, r, httpx.InternalServerError("system settings unavailable"))
			return
		}

		httpx.WriteJSON(w, http.StatusOK, publicRuntimeConfigBody{
			DemoMode: demoMode,
			Capabilities: publicCapabilitiesBody{
				AccountCreationEnabled:   access.AccountCreationEnabled && !demoMode,
				ProjectCreationEnabled:   access.ProjectCreationEnabled && !demoMode,
				CredentialChangesEnabled: access.CredentialChangesEnabled && !demoMode,
			},
		})
	})
}
