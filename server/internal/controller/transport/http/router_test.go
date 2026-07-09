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
	Put    *operationSnapshot `json:"put"`
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
	assertOpenAPIOperation(t, spec, http.MethodPost, "/users/me/deactivation", "deactivateCurrentUser")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/admin/system-admins", "listSystemAdmins")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/admin/system-admins", "grantSystemAdmin")
	assertOpenAPIOperation(t, spec, http.MethodDelete, "/admin/system-admins/{user_id}", "revokeSystemAdmin")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/admin/users", "listManagedUsers")
	assertOpenAPIOperation(t, spec, http.MethodPatch, "/admin/users/{user_id}", "updateManagedUser")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/admin/users/{user_id}/password", "setManagedUserPassword")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/admin/data-export", "exportAdminData")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/admin/data-import", "importAdminData")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/admin/settings", "getAdminSettings")
	assertOpenAPIOperation(t, spec, http.MethodPatch, "/admin/settings", "updateAdminSettings")
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
	assertOpenAPIOperation(t, spec, http.MethodGet, "/projects/{ref}/results/latest", "queryProjectLatestResults")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/projects/{ref}/results/traceroute/runs", "queryProjectTracerouteResultRuns")
	assertOpenAPIOperation(t, spec, http.MethodGet, "/projects/{ref}/results/traceroute/insight", "queryProjectTracerouteResultInsight")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/runtime/probes/{probe_id}/hello", "probeRuntimeHello")
	assertOpenAPIOperation(t, spec, http.MethodPost, "/runtime/probes/{probe_id}/heartbeat", "probeRuntimeHeartbeat")
	assertOpenAPIOperation(t, spec, http.MethodPut, "/runtime/probes/{probe_id}/ip-family-capabilities", "updateProbeRuntimeIPFamilyCapabilities")
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
		BackendBaseURL: "https://app.netstamp.dev/",
	})

	if len(spec.Servers) != 1 {
		t.Fatalf("expected one server, got %d", len(spec.Servers))
	}
	if spec.Servers[0].URL != "https://app.netstamp.dev/api/v1" {
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
		{method: http.MethodPost, path: "/api/v1/users/me/deactivation"},
		{method: http.MethodGet, path: "/api/v1/admin/system-admins"},
		{method: http.MethodPost, path: "/api/v1/admin/system-admins"},
		{method: http.MethodDelete, path: "/api/v1/admin/system-admins/user-1"},
		{method: http.MethodGet, path: "/api/v1/admin/users"},
		{method: http.MethodPatch, path: "/api/v1/admin/users/user-1"},
		{method: http.MethodPost, path: "/api/v1/admin/users/user-1/password"},
		{method: http.MethodGet, path: "/api/v1/admin/data-export"},
		{method: http.MethodPost, path: "/api/v1/admin/data-import"},
		{method: http.MethodGet, path: "/api/v1/admin/settings"},
		{method: http.MethodPatch, path: "/api/v1/admin/settings"},
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
		{method: http.MethodGet, path: "/api/v1/projects/vector-ix/results/latest"},
		{method: http.MethodGet, path: "/api/v1/projects/vector-ix/results/traceroute/runs"},
		{method: http.MethodGet, path: "/api/v1/projects/vector-ix/results/traceroute/insight"},
		{method: http.MethodGet, path: "/api/v1/projects/vector-ix/results/traceroute/topology"},
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

func TestNewRouterDemoModeAllowsReadAndAuthSessionRoutes(t *testing.T) {
	dep := Dependencies{
		APIVersion:     "v1",
		DemoMode:       true,
		RequestTimeout: time.Second,
	}

	for _, route := range []struct {
		name   string
		method string
		path   string
		status int
	}{
		{name: "health", method: http.MethodGet, path: "/api/v1/healthz", status: http.StatusOK},
		{name: "login", method: http.MethodPost, path: "/api/v1/auth/login", status: http.StatusBadRequest},
		{name: "logout", method: http.MethodPost, path: "/api/v1/auth/logout", status: http.StatusNoContent},
	} {
		t.Run(route.name, func(t *testing.T) {
			recorder := performRouterRequest(dep, route.method, route.path)

			if recorder.Code != route.status {
				t.Fatalf("expected status %d, got %d", route.status, recorder.Code)
			}
		})
	}
}

func TestNewRouterDemoModeBlocksUnsafeRequests(t *testing.T) {
	dep := Dependencies{
		APIVersion:     "v1",
		DemoMode:       true,
		RequestTimeout: time.Second,
	}

	for _, route := range []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/api/v1/auth/register"},
		{method: http.MethodPatch, path: "/api/v1/users/me"},
		{method: http.MethodPost, path: "/api/v1/users/me/deactivation"},
		{method: http.MethodPatch, path: "/api/v1/admin/users/user-1"},
		{method: http.MethodPost, path: "/api/v1/admin/users/user-1/password"},
		{method: http.MethodPost, path: "/api/v1/admin/data-import"},
		{method: http.MethodPost, path: "/api/v1/projects"},
		{method: http.MethodDelete, path: "/api/v1/projects/vector-ix/members/user-1"},
		{method: http.MethodPost, path: "/api/v1/runtime/probes/probe-1/results"},
	} {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			recorder := performRouterRequest(dep, route.method, route.path)

			if recorder.Code != http.StatusForbidden {
				t.Fatalf("expected status 403, got %d", recorder.Code)
			}
			if body := recorder.Body.String(); !strings.Contains(body, `"detail":"demo is read-only"`) {
				t.Fatalf("expected read-only problem body, got %q", body)
			}
		})
	}
}

func TestNewRouterDemoModeOffKeepsUnsafeRequestsOnOriginalHandlers(t *testing.T) {
	recorder := performRouterRequest(Dependencies{
		APIVersion:     "v1",
		AuthVerifier:   staticRouterTokenVerifier{},
		RequestTimeout: time.Second,
	}, http.MethodPost, "/api/v1/projects")

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected original handler status 401, got %d", recorder.Code)
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
	webDir := writeTestWebDir(t)
	router := NewRouter(Dependencies{
		APIVersion:     "v1",
		RequestTimeout: time.Second,
		WebDir:         webDir,
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

func TestNewRouterServesFrontendSPAWhenWebDirIsConfigured(t *testing.T) {
	webDir := writeTestWebDir(t)

	for _, route := range []string{"/", "/login", "/dashboard"} {
		t.Run(route, func(t *testing.T) {
			recorder := performRouterRequest(Dependencies{
				APIVersion:     "v1",
				RequestTimeout: time.Second,
				WebDir:         webDir,
			}, http.MethodGet, route)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected frontend route status 200, got %d", recorder.Code)
			}
			if got := recorder.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
				t.Fatalf("expected frontend content type, got %q", got)
			}
			if body := recorder.Body.String(); !strings.Contains(body, "netstamp app shell") {
				t.Fatalf("expected frontend index body, got %q", body)
			}
		})
	}
}

func TestNewRouterServesFrontendAssetsWithImmutableCache(t *testing.T) {
	webDir := writeTestWebDir(t)

	recorder := performRouterRequest(Dependencies{
		APIVersion:     "v1",
		RequestTimeout: time.Second,
		WebDir:         webDir,
	}, http.MethodGet, "/assets/app.js")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected asset status 200, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Cache-Control"); got != immutableAssetCache {
		t.Fatalf("expected immutable asset cache header, got %q", got)
	}
	if body := recorder.Body.String(); body != "console.log('netstamp');" {
		t.Fatalf("expected asset body, got %q", body)
	}
}

func TestNewRouterDoesNotFallbackMissingAssetsToFrontendIndex(t *testing.T) {
	webDir := writeTestWebDir(t)

	recorder := performRouterRequest(Dependencies{
		APIVersion:     "v1",
		RequestTimeout: time.Second,
		WebDir:         webDir,
	}, http.MethodGet, "/assets/missing.js")

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected missing asset status 404, got %d", recorder.Code)
	}
	if body := recorder.Body.String(); strings.Contains(body, "netstamp app shell") {
		t.Fatalf("expected missing asset not to serve frontend index, got %q", body)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "" {
		t.Fatalf("expected missing asset not to set cache header, got %q", got)
	}
}

func TestNewRouterKeepsAPIAndHealthOutsideFrontendFallback(t *testing.T) {
	webDir := writeTestWebDir(t)

	for _, route := range []string{"/api/v1/healthz", "/healthz"} {
		t.Run(route, func(t *testing.T) {
			recorder := performRouterRequest(Dependencies{
				APIVersion:     "v1",
				RequestTimeout: time.Second,
				WebDir:         webDir,
			}, http.MethodGet, route)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected health route status 200, got %d", recorder.Code)
			}
			if got := recorder.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("expected health content type, got %q", got)
			}
			if body := recorder.Body.String(); !strings.Contains(body, `"status":"ok"`) {
				t.Fatalf("expected health body, got %q", body)
			}
		})
	}

	recorder := performRouterRequest(Dependencies{
		APIVersion:     "v1",
		RequestTimeout: time.Second,
		WebDir:         webDir,
	}, http.MethodGet, "/api/v1/missing")

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected missing API route status 404, got %d", recorder.Code)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("expected missing API content type, got %q", got)
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
		`binary_url="http://example.com/api/v1/install/${binary_filename}"`,
		`controller_url="http://example.com"`,
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

func TestNewRouterServesAgentInstallerScriptWithBackendBaseURL(t *testing.T) {
	recorder := performRouterRequest(Dependencies{
		APIVersion:     "v1",
		BackendBaseURL: "https://netstamp.example.com",
		RequestTimeout: time.Second,
	}, http.MethodGet, "/api/v1/install/agent.sh")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected install script status 200, got %d", recorder.Code)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		`binary_url="https://netstamp.example.com/api/v1/install/${binary_filename}"`,
		`controller_url="https://netstamp.example.com"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected install script body to contain %q", want)
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

func writeTestWebDir(t *testing.T) string {
	t.Helper()

	webDir := t.TempDir()
	assetDir := filepath.Join(webDir, "assets")
	if err := os.MkdirAll(assetDir, 0o755); err != nil {
		t.Fatalf("create asset dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(webDir, "index.html"), []byte("<!doctype html><title>netstamp app shell</title>"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(assetDir, "app.js"), []byte("console.log('netstamp');"), 0o644); err != nil {
		t.Fatalf("write asset: %v", err)
	}

	return webDir
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
	case http.MethodPut:
		operation = pathItem.Put
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
