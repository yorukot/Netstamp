package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestSystemConfigReturnsPublicAccessCapabilitiesWithoutTracking(t *testing.T) {
	router := chi.NewRouter()
	registerSystemRoutes(router, nil, publicAccessSettingsFake{
		settings: PublicAccessSettings{
			AccountCreationEnabled:   true,
			ProjectCreationEnabled:   false,
			CredentialChangesEnabled: true,
		},
	}, false)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/system/config", http.NoBody))

	if response.Code != http.StatusOK {
		t.Fatalf("expected system config status 200, got %d", response.Code)
	}
	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("expected no-store cache policy, got %q", got)
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode system config: %v", err)
	}
	if _, exists := body["tracking"]; exists {
		t.Fatal("system config must not expose tracking settings")
	}
	capabilities, ok := body["capabilities"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected capabilities payload: %#v", body["capabilities"])
	}
	if capabilities["accountCreationEnabled"] != true ||
		capabilities["projectCreationEnabled"] != false ||
		capabilities["credentialChangesEnabled"] != true {
		t.Fatalf("unexpected capabilities: %#v", capabilities)
	}
}

func TestSystemConfigFailsClosedWhenAccessSettingsUnavailable(t *testing.T) {
	tests := []struct {
		name     string
		provider PublicAccessSettingsProvider
	}{
		{name: "missing provider"},
		{name: "settings read failure", provider: publicAccessSettingsFake{err: errors.New("database unavailable")}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router := chi.NewRouter()
			registerSystemRoutes(router, nil, test.provider, false)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/system/config", http.NoBody))

			if response.Code != http.StatusInternalServerError {
				t.Fatalf("expected system config status 500, got %d", response.Code)
			}
			if got := response.Header().Get("Content-Type"); got != "application/problem+json" {
				t.Fatalf("expected problem response content type, got %q", got)
			}
			var problem map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil {
				t.Fatalf("decode system config problem: %v", err)
			}
			if problem["code"] != "INTERNAL_ERROR" {
				t.Fatalf("expected INTERNAL_ERROR problem, got %#v", problem)
			}
			if _, exists := problem["capabilities"]; exists {
				t.Fatalf("failure response must not contain capabilities: %#v", problem)
			}
		})
	}
}

func TestSystemConfigDisablesCapabilitiesInDemoMode(t *testing.T) {
	router := chi.NewRouter()
	registerSystemRoutes(router, nil, publicAccessSettingsFake{
		settings: PublicAccessSettings{
			AccountCreationEnabled:   true,
			ProjectCreationEnabled:   true,
			CredentialChangesEnabled: true,
		},
	}, true)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/system/config", http.NoBody))

	var body publicRuntimeConfigBody
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode system config: %v", err)
	}
	if body.Capabilities.AccountCreationEnabled ||
		body.Capabilities.ProjectCreationEnabled ||
		body.Capabilities.CredentialChangesEnabled {
		t.Fatalf("expected demo mode to disable all mutable capabilities: %#v", body.Capabilities)
	}
}

type publicAccessSettingsFake struct {
	settings PublicAccessSettings
	err      error
}

func (f publicAccessSettingsFake) PublicAccessSettings(context.Context) (PublicAccessSettings, error) {
	return f.settings, f.err
}
