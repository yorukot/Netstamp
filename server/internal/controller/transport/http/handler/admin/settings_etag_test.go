package admin

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

func TestParseSettingsIfMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		header     []string
		want       int64
		wantStatus int
		wantErr    error
	}{
		{name: "valid", header: []string{`"auth.oidc-42"`}, want: 42},
		{name: "missing", wantStatus: http.StatusPreconditionRequired},
		{name: "empty", header: []string{" "}, wantStatus: http.StatusPreconditionRequired},
		{name: "unquoted", header: []string{"auth.oidc-42"}, wantStatus: http.StatusBadRequest},
		{name: "multiple header values", header: []string{`"auth.oidc-1"`, `"auth.oidc-2"`}, wantStatus: http.StatusBadRequest},
		{name: "multiple header values starting blank", header: []string{"", `"auth.oidc-2"`}, wantStatus: http.StatusBadRequest},
		{name: "entity tag list", header: []string{`"auth.oidc-1", "auth.oidc-2"`}, wantStatus: http.StatusBadRequest},
		{name: "invalid revision", header: []string{`"auth.oidc-nope"`}, wantStatus: http.StatusBadRequest},
		{name: "zero revision", header: []string{`"auth.oidc-0"`}, wantStatus: http.StatusBadRequest},
		{name: "positive sign", header: []string{`"auth.oidc-+1"`}, wantStatus: http.StatusBadRequest},
		{name: "leading zero", header: []string{`"auth.oidc-01"`}, wantStatus: http.StatusBadRequest},
		{name: "revision overflow", header: []string{`"auth.oidc-999999999999999999999"`}, wantStatus: http.StatusBadRequest},
		{name: "weak", header: []string{`W/"auth.oidc-42"`}, wantErr: errSettingsVersionConflict},
		{name: "wildcard", header: []string{"*"}, wantErr: errSettingsVersionConflict},
		{name: "different resource", header: []string{`"auth.google-42"`}, wantErr: errSettingsVersionConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodPatch, "/", http.NoBody)
			for _, value := range tt.header {
				request.Header.Add("If-Match", value)
			}

			got, err := parseSettingsIfMatch(request, "auth.oidc")
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if tt.wantStatus != 0 {
				var httpErr *httpx.Error
				if !errors.As(err, &httpErr) {
					t.Fatalf("expected HTTP error, got %v", err)
				}
				if httpErr.Status != tt.wantStatus {
					t.Fatalf("expected status %d, got %d", tt.wantStatus, httpErr.Status)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse If-Match: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected revision %d, got %d", tt.want, got)
			}
		})
	}
}

func TestSetVersionedSettingsHeaders(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	setVersionedSettingsHeaders(recorder, "smtp", 7)

	if got := recorder.Header().Get("ETag"); got != `"smtp-7"` {
		t.Fatalf("expected SMTP ETag, got %q", got)
	}
	if got := recorder.Header().Get("Cache-Control"); got != settingsCacheControl {
		t.Fatalf("expected private no-store, got %q", got)
	}
}

func TestResolveSettingsIfMatchReturnsCurrentVersionForRejectedStrongRequirement(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodPatch, "/", http.NoBody)
	request.Header.Set("If-Match", `W/"smtp-3"`)

	revision, err := resolveSettingsIfMatch(request, "smtp", func() (int64, error) {
		return 8, nil
	})
	if revision != 0 {
		t.Fatalf("expected no accepted revision, got %d", revision)
	}

	var conflict *settingsVersionConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("expected request settings conflict, got %v", err)
	}
	if conflict.resource != "smtp" || conflict.currentRevision != 8 {
		t.Fatalf("unexpected conflict: %#v", conflict)
	}
}
