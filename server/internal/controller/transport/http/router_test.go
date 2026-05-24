package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"

	"github.com/yorukot/netstamp/internal/domain/identity"
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
	Type        string `json:"type"`
	In          string `json:"in"`
	Name        string `json:"name"`
	Scheme      string `json:"scheme"`
	Description string `json:"description"`
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
	assertOpenAPIOperation(t, spec, http.MethodPatch, "/users/me", "updateCurrentUser")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/users/me/email-change", "changeCurrentUserEmail")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/users/me/password-change", "changeCurrentUserPassword")
	assertOpenAPISessionCookieAuth(t, spec)
	assertOpenAPIProbeAuth(t, spec)
	assertOpenAPIOperation(t, spec, http.MethodPost, "/projects", "createProject")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/projects/{ref}/assignments", "listProjectAssignments")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/projects/{ref}/checks", "createProjectCheck")
	assertOpenAPIOperation(t, spec, http.MethodDelete, "/projects/{ref}/members/{user_id}", "removeProjectMember")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/projects/{ref}/selector-previews", "previewProjectSelector")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/projects/{ref}/probes", "listProjectProbes")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/projects/{ref}/probes", "createProjectProbe")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/projects/{ref}/probes/{probe_id}", "getProjectProbe")
	assertOpenAPIOperation(t, spec, http.MethodPatch, "/projects/{ref}/probes/{probe_id}", "updateProjectProbe")
	assertOpenAPIOperation(t, spec, http.MethodDelete, "/projects/{ref}/probes/{probe_id}", "deleteProjectProbe")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/projects/{ref}/probes/{probe_id}/secret-rotations", "rotateProjectProbeSecret")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/projects/{ref}/results/ping/series", "queryProjectPingResultSeries")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/projects/{ref}/results/ping/insight", "queryProjectPingResultInsight")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/projects/{ref}/results/traceroute/runs", "queryProjectTracerouteResultRuns")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/projects/{ref}/measurements", "listProjectMeasurements")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/runtime/probes/{probe_id}/hello", "probeRuntimeHello")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/runtime/probes/{probe_id}/heartbeat", "probeRuntimeHeartbeat")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/runtime/probes/{probe_id}/assignments", "listProbeRuntimeAssignments")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/runtime/probes/{probe_id}/results", "submitProbeRuntimeResults")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/install/agent.sh", "getAgentInstallerScript")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/install/uninstall-agent.sh", "getAgentUninstallerScript")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/install/netstamp-agent-linux-amd64", "downloadAgentLinuxAmd64")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/install/netstamp-agent-linux-arm64", "downloadAgentLinuxArm64")
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

func assertOpenAPIProbeAuth(t *testing.T, spec openAPISnapshot) {
	t.Helper()

	scheme, ok := spec.Components.SecuritySchemes["ProbeAuth"]
	if !ok {
		t.Fatal("expected ProbeAuth security scheme")
	}
	if scheme.Type != "apiKey" || scheme.In != "header" || scheme.Name != "Authorization" {
		t.Fatalf("unexpected probe auth security scheme: %#v", scheme)
	}
	if !strings.Contains(scheme.Description, "Authorization: Probe <secret>") {
		t.Fatalf("expected probe auth description to document header value, got %q", scheme.Description)
	}

	for _, route := range []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/runtime/probes/{probe_id}/hello"},
		{method: http.MethodPost, path: "/runtime/probes/{probe_id}/heartbeat"},
		{method: http.MethodGet, path: "/runtime/probes/{probe_id}/assignments"},
		{method: http.MethodPost, path: "/runtime/probes/{probe_id}/results"},
	} {
		operation := openAPIOperationForMethod(t, spec, route.method, route.path)
		if len(operation.Security) != 1 {
			t.Fatalf("expected %s %s security requirement, got %#v", route.method, route.path, operation.Security)
		}
		if _, ok := operation.Security[0]["ProbeAuth"]; !ok {
			t.Fatalf("expected %s %s to use ProbeAuth, got %#v", route.method, route.path, operation.Security)
		}
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

func TestNewRouterServesScalarDocs(t *testing.T) {
	for _, path := range []string{"/api/v1/docs", "/api/v1/docs/"} {
		recorder := performRouterRequest(Dependencies{
			APIVersion:     "v1",
			RequestTimeout: time.Second,
		}, http.MethodGet, path)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected %s status 200, got %d", path, recorder.Code)
		}
		if got := recorder.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
			t.Fatalf("expected %s content type, got %q", path, got)
		}
		body := recorder.Body.String()
		for _, want := range []string{
			"@scalar/api-reference",
			"Scalar.createApiReference",
			`"url": "/api/v1/openapi.json"`,
			`"theme": "elysiajs"`,
			`"layout": "modern"`,
			`"showDeveloperTools": "localhost"`,
			`"title": "API #1"`,
		} {
			if !strings.Contains(body, want) {
				t.Fatalf("expected %s Scalar docs body to contain %q, got %q", path, want, body)
			}
		}
	}
}

func TestNewRouterProtectedRoutesRequireSessionCookie(t *testing.T) {
	dep := Dependencies{
		APIVersion:     "v1",
		AuthVerifier:   staticRouterTokenVerifier{},
		RequestTimeout: time.Second,
	}

	for _, route := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/api/v1/auth/me"},
		{method: http.MethodPatch, path: "/api/v1/users/me"},
		{method: http.MethodPost, path: "/api/v1/users/me/email-change"},
		{method: http.MethodPost, path: "/api/v1/users/me/password-change"},
		{method: http.MethodPost, path: "/api/v1/projects/vector-ix/selector-previews"},
		{method: http.MethodGet, path: "/api/v1/projects/vector-ix/assignments"},
		{method: http.MethodGet, path: "/api/v1/projects/vector-ix/labels"},
		{method: http.MethodPost, path: "/api/v1/projects/vector-ix/labels"},
		{method: http.MethodPatch, path: "/api/v1/projects/vector-ix/labels/label-1"},
		{method: http.MethodDelete, path: "/api/v1/projects/vector-ix/labels/label-1"},
		{method: http.MethodGet, path: "/api/v1/projects/vector-ix/checks"},
		{method: http.MethodPost, path: "/api/v1/projects/vector-ix/checks"},
		{method: http.MethodGet, path: "/api/v1/projects/vector-ix/checks/check-1"},
		{method: http.MethodPatch, path: "/api/v1/projects/vector-ix/checks/check-1"},
		{method: http.MethodDelete, path: "/api/v1/projects/vector-ix/checks/check-1"},
		{method: http.MethodGet, path: "/api/v1/projects/vector-ix/probes"},
		{method: http.MethodPost, path: "/api/v1/projects/vector-ix/probes"},
		{method: http.MethodGet, path: "/api/v1/projects/vector-ix/probes/probe-1"},
		{method: http.MethodPatch, path: "/api/v1/projects/vector-ix/probes/probe-1"},
		{method: http.MethodDelete, path: "/api/v1/projects/vector-ix/probes/probe-1"},
		{method: http.MethodPost, path: "/api/v1/projects/vector-ix/probes/probe-1/secret-rotations"},
		{method: http.MethodGet, path: "/api/v1/projects/vector-ix/results/ping/series"},
		{method: http.MethodGet, path: "/api/v1/projects/vector-ix/results/ping/insight"},
		{method: http.MethodGet, path: "/api/v1/projects/vector-ix/results/traceroute/runs"},
		{method: http.MethodGet, path: "/api/v1/projects/vector-ix/results/traceroute/topology"},
		{method: http.MethodGet, path: "/api/v1/projects/vector-ix/measurements"},
	} {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			recorder := performRouterRequest(dep, route.method, route.path)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("expected status 401, got %d", recorder.Code)
			}
			if got := recorder.Header().Get("WWW-Authenticate"); got != "" {
				t.Fatalf("expected empty WWW-Authenticate, got %q", got)
			}
		})
	}
}

func TestNewRouterRuntimeRoutesRequireProbeCredential(t *testing.T) {
	dep := Dependencies{
		APIVersion:     "v1",
		RequestTimeout: time.Second,
	}

	for _, route := range []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/api/v1/runtime/probes/probe-1/hello"},
		{method: http.MethodPost, path: "/api/v1/runtime/probes/probe-1/heartbeat"},
		{method: http.MethodGet, path: "/api/v1/runtime/probes/probe-1/assignments"},
		{method: http.MethodPost, path: "/api/v1/runtime/probes/probe-1/results"},
	} {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			recorder := performRouterRequest(dep, route.method, route.path)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("expected status 401, got %d", recorder.Code)
			}
			if got := recorder.Header().Get("WWW-Authenticate"); got != "Probe" {
				t.Fatalf("expected Probe WWW-Authenticate, got %q", got)
			}
		})
	}
}

func TestNewRouterPublicRoutesBypassAuthGroups(t *testing.T) {
	for _, route := range []struct {
		method string
		path   string
		status int
	}{
		{method: http.MethodGet, path: "/api/v1/healthz", status: http.StatusOK},
		{method: http.MethodGet, path: "/api/v1/openapi.json", status: http.StatusOK},
		{method: http.MethodGet, path: "/api/v1/docs", status: http.StatusOK},
		{method: http.MethodPost, path: "/api/v1/auth/logout", status: http.StatusNoContent},
	} {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			recorder := performRouterRequest(Dependencies{
				APIVersion:     "v1",
				RequestTimeout: time.Second,
			}, route.method, route.path)

			if recorder.Code != route.status {
				t.Fatalf("expected status %d, got %d", route.status, recorder.Code)
			}
		})
	}
}

func TestNewRouterWritesPlainTextNotFoundForBrowsers(t *testing.T) {
	recorder := performRouterRequestWithHeaders(Dependencies{
		APIVersion:     "v1",
		RequestTimeout: time.Second,
	}, http.MethodGet, "/api/v1/missing", map[string]string{
		"Accept": "text/html",
	})

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected missing route status 404, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Fatalf("expected plain text not found content type, got %q", got)
	}
	if body := recorder.Body.String(); body != "404 page not found\n" {
		t.Fatalf("expected plain text not found body, got %q", body)
	}
}

func TestNewRouterWritesProblemNotFoundForAPIClients(t *testing.T) {
	recorder := performRouterRequest(Dependencies{
		APIVersion:     "v1",
		RequestTimeout: time.Second,
	}, http.MethodGet, "/api/v1/missing")

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected missing route status 404, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("expected problem content type, got %q", got)
	}
	if body := recorder.Body.String(); !strings.Contains(body, `"detail":"route not found"`) {
		t.Fatalf("expected problem not found body, got %q", body)
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

func TestNewRouterServesAgentInstallerScript(t *testing.T) {
	recorder := performRouterRequest(Dependencies{
		APIVersion:     "v1",
		RequestTimeout: time.Second,
	}, http.MethodGet, "/api/v1/install/agent.sh")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected install script status 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/x-shellscript; charset=utf-8" {
		t.Fatalf("expected install script content type, got %q", got)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		"#!/bin/sh",
		"binary_filename=netstamp-agent-linux-amd64",
		"binary_filename=netstamp-agent-linux-arm64",
		`binary_url="https://netstamp.dev/api/v1/install/${binary_filename}"`,
		`controller_url="https://netstamp.dev"`,
		"sudo netstamp-agent service install --url",
		"linux/amd64 and linux/arm64",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected install script body to contain %q", want)
		}
	}
	for _, notWant := range []string{
		"__NETSTAMP_AGENT_BINARY_URL__",
		"__NETSTAMP_CONTROLLER_URL__",
		"NETSTAMP_PROBE_SECRET",
		"systemctl enable --now",
		"AmbientCapabilities=CAP_NET_RAW",
	} {
		if strings.Contains(body, notWant) {
			t.Fatalf("expected thin install script body not to contain %q", notWant)
		}
	}
}

func TestNewRouterServesAgentUninstallerScript(t *testing.T) {
	recorder := performRouterRequest(Dependencies{
		APIVersion:     "v1",
		RequestTimeout: time.Second,
	}, http.MethodGet, "/api/v1/install/uninstall-agent.sh")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected uninstall script status 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/x-shellscript; charset=utf-8" {
		t.Fatalf("expected uninstall script content type, got %q", got)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		"#!/bin/sh",
		"--purge",
		"exec \"$agent_path\" service uninstall \"$@\"",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected uninstall script body to contain %q", want)
		}
	}
	if strings.Contains(body, "systemctl disable --now") {
		t.Fatalf("expected uninstall script wrapper not to contain systemctl implementation")
	}
}

func TestNewRouterServesAgentBinary(t *testing.T) {
	for _, filename := range []string{
		"netstamp-agent-linux-amd64",
		"netstamp-agent-linux-arm64",
	} {
		t.Run(filename, func(t *testing.T) {
			agentDir := t.TempDir()
			agentPath := filepath.Join(agentDir, filename)
			if err := os.WriteFile(agentPath, []byte(filename+"-binary"), 0o755); err != nil {
				t.Fatalf("write test agent binary: %v", err)
			}

			recorder := performRouterRequest(Dependencies{
				APIVersion:     "v1",
				RequestTimeout: time.Second,
				AgentBinaryDir: agentDir,
			}, http.MethodGet, "/api/v1/install/"+filename)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected agent binary status 200, got %d", recorder.Code)
			}
			if got := recorder.Header().Get("Content-Type"); got != "application/octet-stream" {
				t.Fatalf("expected agent binary content type, got %q", got)
			}
			if got := recorder.Header().Get("Content-Disposition"); got != `attachment; filename="`+filename+`"` {
				t.Fatalf("expected agent binary content disposition, got %q", got)
			}
			if recorder.Body.String() != filename+"-binary" {
				t.Fatalf("expected agent binary body, got %q", recorder.Body.String())
			}
		})
	}
}

func TestNewRouterReportsMissingAgentBinary(t *testing.T) {
	for _, filename := range []string{
		"netstamp-agent-linux-amd64",
		"netstamp-agent-linux-arm64",
	} {
		t.Run(filename, func(t *testing.T) {
			recorder := performRouterRequest(Dependencies{
				APIVersion:     "v1",
				RequestTimeout: time.Second,
				AgentBinaryDir: t.TempDir(),
			}, http.MethodGet, "/api/v1/install/"+filename)

			if recorder.Code != http.StatusNotFound {
				t.Fatalf("expected missing agent binary status 404, got %d", recorder.Code)
			}
			if body := recorder.Body.String(); !strings.Contains(body, `"detail":"agent binary not found"`) {
				t.Fatalf("expected missing agent binary problem body, got %q", body)
			}
		})
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

func performRouterRequest(dep Dependencies, method, path string) *httptest.ResponseRecorder {
	return performRouterRequestWithHeaders(dep, method, path, nil)
}

func performRouterRequestWithHeaders(dep Dependencies, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	router := NewRouter(dep)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(method, path, http.NoBody)
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	router.ServeHTTP(recorder, request)
	return recorder
}

type staticRouterTokenVerifier struct{}

func (staticRouterTokenVerifier) VerifyAccessToken(context.Context, string) (identity.AccessTokenClaims, error) {
	return identity.AccessTokenClaims{
		Subject: "user-1",
		Email:   "user@example.com",
	}, nil
}

func assertOpenAPIOperation(t *testing.T, spec openAPISnapshot, method, path, operationID string) {
	t.Helper()

	operation := openAPIOperationForMethod(t, spec, method, path)
	if operation.OperationID != operationID {
		t.Fatalf("expected %s %s operation ID %q, got %q", method, path, operationID, operation.OperationID)
	}
}

func openAPIOperationForMethod(t *testing.T, spec openAPISnapshot, method, path string) *operationSnapshot {
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
	return operation
}

func assertOpenAPIPathAbsent(t *testing.T, spec openAPISnapshot, path string) {
	t.Helper()

	if _, ok := spec.Paths[path]; ok {
		t.Fatalf("expected OpenAPI path %q not to be registered", path)
	}
}
