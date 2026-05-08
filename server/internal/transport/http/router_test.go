package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type openAPISnapshot struct {
	Servers []struct {
		URL string `json:"url"`
	} `json:"servers"`
	Paths map[string]pathItemSnapshot `json:"paths"`
}

type pathItemSnapshot struct {
	Get    *operationSnapshot `json:"get"`
	Post   *operationSnapshot `json:"post"`
	Patch  *operationSnapshot `json:"patch"`
	Delete *operationSnapshot `json:"delete"`
}

type operationSnapshot struct {
	OperationID string `json:"operationId"`
}

func TestNewHumaConfigUsesRelativeServerURLWhenBackendBaseURLUnset(t *testing.T) {
	config := newHumaConfig(Dependencies{APIVersion: "v1"})

	if len(config.Servers) != 1 {
		t.Fatalf("expected one server, got %d", len(config.Servers))
	}
	if config.Servers[0].URL != "/api/v1" {
		t.Fatalf("expected relative server URL, got %q", config.Servers[0].URL)
	}
}

func TestNewHumaConfigUsesBackendBaseURLServerURL(t *testing.T) {
	config := newHumaConfig(Dependencies{
		APIVersion:     "v1",
		BackendBaseURL: "https://api.netstamp.dev/",
	})

	if len(config.Servers) != 1 {
		t.Fatalf("expected one server, got %d", len(config.Servers))
	}
	if config.Servers[0].URL != "https://api.netstamp.dev/api/v1" {
		t.Fatalf("expected absolute server URL, got %q", config.Servers[0].URL)
	}
}

func TestNewRouterServesOpenAPIWithoutRuntimeServices(t *testing.T) {
	spec := getOpenAPI(t, Dependencies{APIVersion: "v1"})

	assertOpenAPIOperation(t, spec, http.MethodPost, "/auth/register", "registerUser")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/auth/login", "loginUser")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/auth/me", "getCurrentUser")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/projects", "createProject")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/projects/{ref}/checks", "createProjectCheck")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/projects/{ref}/probes", "createProjectProbe")
}

func TestNewRouterOpenAPIUsesBackendBaseURLServerURL(t *testing.T) {
	spec := getOpenAPI(t, Dependencies{
		APIVersion:     "v1",
		BackendBaseURL: "https://api.netstamp.dev/",
	})

	if len(spec.Servers) != 1 {
		t.Fatalf("expected one server, got %d", len(spec.Servers))
	}
	if spec.Servers[0].URL != "https://api.netstamp.dev/api/v1" {
		t.Fatalf("expected absolute server URL, got %q", spec.Servers[0].URL)
	}
}

func getOpenAPI(t *testing.T, dep Dependencies) openAPISnapshot {
	t.Helper()

	if dep.RequestTimeout == 0 {
		dep.RequestTimeout = time.Second
	}
	router := NewRouter(dep)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/"+dep.APIVersion+"/openapi.json", http.NoBody)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected OpenAPI status 200, got %d", recorder.Code)
	}

	var spec openAPISnapshot
	if err := json.NewDecoder(recorder.Body).Decode(&spec); err != nil {
		t.Fatalf("decode OpenAPI: %v", err)
	}
	return spec
}

func assertOpenAPIOperation(t *testing.T, spec openAPISnapshot, method, path, operationID string) {
	t.Helper()

	pathItem, ok := spec.Paths[path]
	if !ok {
		t.Fatalf("expected OpenAPI path %q to be registered", path)
	}

	var operation *operationSnapshot
	switch method {
	case http.MethodGet:
		operation = pathItem.Get
	case http.MethodPost:
		operation = pathItem.Post
	case http.MethodPatch:
		operation = pathItem.Patch
	case http.MethodDelete:
		operation = pathItem.Delete
	default:
		t.Fatalf("unsupported method %q", method)
	}

	if operation == nil {
		t.Fatalf("expected %s %s to be registered", method, path)
	}
	if operation.OperationID != operationID {
		t.Fatalf("expected %s %s operation ID %q, got %q", method, path, operationID, operation.OperationID)
	}
}
