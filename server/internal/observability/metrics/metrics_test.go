package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProviderHandlerExposesRuntimeAndTargetMetrics(t *testing.T) {
	provider, err := NewProvider(Config{
		Env:            "test",
		ServiceName:    "netstamp-api",
		ServiceVersion: "0.1.0",
	})
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	t.Cleanup(func() {
		if err := provider.Shutdown(t.Context()); err != nil {
			t.Fatalf("shutdown provider: %v", err)
		}
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	provider.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	for _, want := range []string{
		"go_goroutines",
		"process_cpu_seconds_total",
		"target_info",
		`service_name="netstamp-api"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected metrics output to contain %q", want)
		}
	}
}
