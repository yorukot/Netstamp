package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
)

type openAPISnapshot struct {
	Servers []struct {
		URL string `json:"url"`
	} `json:"servers"`
	Components struct {
		SecuritySchemes map[string]securitySchemeSnapshot `json:"securitySchemes"`
	} `json:"components"`
	Paths map[string]pathItemSnapshot `json:"paths"`
}

type securitySchemeSnapshot struct {
	Type   string `json:"type"`
	In     string `json:"in"`
	Name   string `json:"name"`
	Scheme string `json:"scheme"`
}

type pathItemSnapshot struct {
	Get    *operationSnapshot `json:"get"`
	Post   *operationSnapshot `json:"post"`
	Patch  *operationSnapshot `json:"patch"`
	Delete *operationSnapshot `json:"delete"`
}

type operationSnapshot struct {
	OperationID string                `json:"operationId"`
	Security    []map[string][]string `json:"security"`
}

func TestNewRouterServesOpenAPIWithoutRuntimeServices(t *testing.T) {
	spec := getOpenAPI(t, Dependencies{APIVersion: "v1"})

	assertOpenAPIOperation(t, spec, http.MethodGet, "/", "getAPIStatus")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/healthz", "getHealth")
	assertOpenAPIPathAbsent(t, spec, "/livez")
	assertOpenAPIPathAbsent(t, spec, "/readyz")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/auth/register", "registerUser")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/auth/login", "loginUser")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/auth/logout", "logoutUser")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/auth/me", "getCurrentUser")
	assertOpenAPISessionCookieAuth(t, spec)
	assertOpenAPIOperation(t, spec, http.MethodPost, "/projects", "createProject")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/projects/{ref}/checks", "createProjectCheck")
	assertOpenAPIOperation(t, spec, http.MethodDelete, "/projects/{ref}/members/{user_id}", "removeProjectMember")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/projects/{ref}/probes", "listProjectProbes")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/projects/{ref}/probes", "createProjectProbe")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/projects/{ref}/probes/{probe_id}", "getProjectProbe")
	assertOpenAPIOperation(t, spec, http.MethodPatch, "/projects/{ref}/probes/{probe_id}", "updateProjectProbe")
	assertOpenAPIOperation(t, spec, http.MethodDelete, "/projects/{ref}/probes/{probe_id}", "deleteProjectProbe")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/projects/{ref}/probes/{probe_id}/secret-rotations", "rotateProjectProbeSecret")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/projects/{ref}/results/ping/series", "queryProjectPingResultSeries")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/projects/{ref}/results/traceroute/runs", "queryProjectTracerouteResultRuns")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/runtime/probes/{probe_id}/hello", "probeRuntimeHello")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/runtime/probes/{probe_id}/heartbeat", "probeRuntimeHeartbeat")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/runtime/probes/{probe_id}/assignments", "listProbeRuntimeAssignments")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/runtime/probes/{probe_id}/results", "submitProbeRuntimeResults")
	assertOpenAPIPathAbsent(t, spec, "/probes/{probe_id}/runtime/hello")
	assertOpenAPIPathAbsent(t, spec, "/probes/{probe_id}/runtime/heartbeat")
	assertOpenAPIPathAbsent(t, spec, "/probes/{probe_id}/runtime/assignments")
	assertOpenAPIPathAbsent(t, spec, "/probes/{probe_id}/runtime/results")
}

func assertOpenAPISessionCookieAuth(t *testing.T, spec openAPISnapshot) {
	t.Helper()

	if _, ok := spec.Components.SecuritySchemes["bearerAuth"]; ok {
		t.Fatal("expected bearerAuth security scheme to be absent")
	}
	scheme, ok := spec.Components.SecuritySchemes["SessionCookieAuth"]
	if !ok {
		t.Fatal("expected SessionCookieAuth security scheme")
	}
	if scheme.Type != "apiKey" || scheme.In != "cookie" || scheme.Name != "netstamp_session" {
		t.Fatalf("unexpected session cookie security scheme: %#v", scheme)
	}

	meOperation := spec.Paths["/auth/me"].Get
	if len(meOperation.Security) != 1 {
		t.Fatalf("expected auth/me security requirement, got %#v", meOperation.Security)
	}
	if _, ok := meOperation.Security[0]["SessionCookieAuth"]; !ok {
		t.Fatalf("expected auth/me to use SessionCookieAuth, got %#v", meOperation.Security)
	}
}

func TestNewRouterOpenAPIUsesRelativeServerURLWhenBackendBaseURLUnset(t *testing.T) {
	spec := getOpenAPI(t, Dependencies{APIVersion: "v1"})

	if len(spec.Servers) != 1 {
		t.Fatalf("expected one server, got %d", len(spec.Servers))
	}
	if spec.Servers[0].URL != "/api/v1" {
		t.Fatalf("expected relative server URL, got %q", spec.Servers[0].URL)
	}
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

func TestNewRouterServesMetricsOutsideVersionedAPI(t *testing.T) {
	router := NewRouter(Dependencies{
		APIVersion:     "v1",
		RequestTimeout: time.Second,
		MetricsHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if requestID := chimw.GetReqID(r.Context()); requestID != "" {
				t.Errorf("expected metrics to bypass request ID middleware, got %q", requestID)
			}
			if spanContext := trace.SpanContextFromContext(r.Context()); spanContext.IsValid() {
				t.Errorf("expected metrics to bypass otelhttp tracing, got trace ID %q", spanContext.TraceID())
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("metrics"))
		}),
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected metrics status 200, got %d", recorder.Code)
	}
	if recorder.Body.String() != "metrics" {
		t.Fatalf("expected metrics response body, got %q", recorder.Body.String())
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

func assertOpenAPIPathAbsent(t *testing.T, spec openAPISnapshot, path string) {
	t.Helper()

	if _, ok := spec.Paths[path]; ok {
		t.Fatalf("expected OpenAPI path %q not to be registered", path)
	}
}
