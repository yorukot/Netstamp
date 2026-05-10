package httptracing_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	httptracing "github.com/yorukot/netstamp/internal/platform/observability/httptrace"
)

func TestRequestSpanName(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register?debug=true", http.NoBody)

	got := httptracing.RequestSpanName("http.server", req)
	want := "POST /api/v1/auth/register"
	if got != want {
		t.Fatalf("expected span name %q, got %q", want, got)
	}
}
