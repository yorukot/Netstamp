package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func TestWriteProblemWritesProblemJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	req = req.WithContext(context.WithValue(req.Context(), chimw.RequestIDKey, "request-1"))
	res := httptest.NewRecorder()

	WriteProblem(res, req, http.StatusTeapot, "short and stout")

	if res.Code != http.StatusTeapot {
		t.Fatalf("expected status 418, got %d", res.Code)
	}
	if got := res.Header().Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("expected problem content type, got %q", got)
	}
	if got := res.Header().Get("X-Request-ID"); got != "request-1" {
		t.Fatalf("expected request id header, got %q", got)
	}

	var body httpx.ProblemDetails
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode problem body: %v", err)
	}
	if body.Status != http.StatusTeapot || body.Title != http.StatusText(http.StatusTeapot) || body.Detail != "short and stout" {
		t.Fatalf("unexpected problem body: %#v", body)
	}
}

func TestWriteProblemOmitsMissingRequestID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	res := httptest.NewRecorder()

	WriteProblem(res, req, http.StatusBadRequest, "bad request")

	if got := res.Header().Get("X-Request-ID"); got != "" {
		t.Fatalf("expected no request id header, got %q", got)
	}
}

func TestWriteProblemIgnoresBodyWriteError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", http.NoBody)
	res := &failingResponseWriter{header: http.Header{}}

	WriteProblem(res, req, http.StatusInternalServerError, "failed")

	if res.status != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", res.status)
	}
	if got := res.header.Get("Content-Type"); got != "application/problem+json" {
		t.Fatalf("expected problem content type, got %q", got)
	}
}

func TestZapRecovererRecoversPanicAndLogsRequestFields(t *testing.T) {
	core, observed := observer.New(zap.DebugLevel)
	log := zap.New(core)
	handler := ZapRecoverer(log)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))
	req := httptest.NewRequest(http.MethodPost, "/panic", http.NoBody)
	req.RemoteAddr = "203.0.113.10:54321"
	req.Header.Set("User-Agent", "netstamp-test")
	req = req.WithContext(context.WithValue(req.Context(), chimw.RequestIDKey, "request-2"))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", res.Code)
	}
	if res.Body.String() != "Internal Server Error\n" {
		t.Fatalf("expected default internal server error body, got %q", res.Body.String())
	}

	logs := observed.FilterMessage("http_panic_recovered").All()
	if len(logs) != 1 {
		t.Fatalf("expected one panic log, got %d", len(logs))
	}
	fields := logs[0].ContextMap()
	assertLogField(t, fields, "request_id", "request-2")
	assertLogField(t, fields, "http.request.method", http.MethodPost)
	assertLogField(t, fields, "url.path", "/panic")
	assertLogField(t, fields, "client.address", "203.0.113.10")
	assertLogField(t, fields, "user_agent.original", "netstamp-test")
	assertLogField(t, fields, "panic", "boom")
	if _, ok := fields["stacktrace"]; !ok {
		t.Fatalf("expected stacktrace field, got %#v", fields)
	}
}

func TestZapRecovererFallsBackToHeaderRequestID(t *testing.T) {
	core, observed := observer.New(zap.DebugLevel)
	handler := ZapRecoverer(zap.New(core))(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))
	req := httptest.NewRequest(http.MethodGet, "/panic", http.NoBody)
	req.Header.Set("X-Request-ID", "header-request-id")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	logs := observed.FilterMessage("http_panic_recovered").All()
	if len(logs) != 1 {
		t.Fatalf("expected one panic log, got %d", len(logs))
	}
	assertLogField(t, logs[0].ContextMap(), "request_id", "header-request-id")
}

func TestZapRecovererHandlesNilLoggerAndPassesThroughWithoutPanic(t *testing.T) {
	handler := ZapRecoverer(nil)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/ok", http.NoBody)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", res.Code)
	}
}

func TestZapRecovererHandlesNilLoggerPanic(t *testing.T) {
	handler := ZapRecoverer(nil)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))
	req := httptest.NewRequest(http.MethodGet, "/panic", http.NoBody)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", res.Code)
	}
}

func TestSecurityHeadersSetsHeadersAndCallsNext(t *testing.T) {
	called := false
	handler := SecurityHeaders()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/secure", http.NoBody)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if !called {
		t.Fatal("expected next handler to be called")
	}
	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", res.Code)
	}
	if got := res.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("expected nosniff, got %q", got)
	}
	if got := res.Header().Get("Referrer-Policy"); got != "no-referrer" {
		t.Fatalf("expected no-referrer, got %q", got)
	}
	if got := res.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("expected DENY, got %q", got)
	}
}

func assertLogField(t *testing.T, fields map[string]any, key string, want any) {
	t.Helper()

	if got := fields[key]; got != want {
		t.Fatalf("expected log field %s=%#v, got %#v in %#v", key, want, got, fields)
	}
}

type failingResponseWriter struct {
	header http.Header
	status int
}

func (w *failingResponseWriter) Header() http.Header {
	return w.header
}

func (w *failingResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func (w *failingResponseWriter) WriteHeader(status int) {
	w.status = status
}
