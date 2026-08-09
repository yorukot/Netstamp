package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
)

func TestResetBaseURLUsesOnlyConfiguredPublicOrigin(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/auth/password-resets", http.NoBody)
	request.Host = "attacker.example"
	request.Header.Set("X-Forwarded-Proto", "https")

	tests := []struct {
		name              string
		publicBaseURL     string
		expectedResetBase string
	}{
		{
			name:              "configured origin",
			publicBaseURL:     " https://app.netstamp.dev/ ",
			expectedResetBase: "https://app.netstamp.dev",
		},
		{
			name:              "unset origin",
			publicBaseURL:     "",
			expectedResetBase: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &Handler{publicBaseURL: tt.publicBaseURL}

			if got := handler.resetBaseURL(request); got != tt.expectedResetBase {
				t.Fatalf("expected reset base URL %q, got %q", tt.expectedResetBase, got)
			}
		})
	}
}

func TestRequestPasswordResetWithoutConfiguredPublicOriginIsUnavailable(t *testing.T) {
	handler := &Handler{service: appauth.NewService(nil, nil, nil, nil)}
	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/password-resets",
		strings.NewReader(`{"email":"user@example.com"}`),
	)
	request.Host = "attacker.example"
	request.Header.Set("X-Forwarded-Proto", "https")
	recorder := httptest.NewRecorder()

	handler.handleRequestPasswordReset(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected password reset to be unavailable, got status %d and body %q", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "attacker.example") {
		t.Fatalf("response must not contain an untrusted request host: %q", recorder.Body.String())
	}
}
