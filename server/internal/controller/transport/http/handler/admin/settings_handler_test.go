package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	appsettings "github.com/yorukot/netstamp/internal/controller/application/systemsettings"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
	httpmiddleware "github.com/yorukot/netstamp/internal/controller/transport/http/middleware"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

type settingsHandlerFunc func(*Handler, http.ResponseWriter, *http.Request)

func TestGetSettingsWritesPrivateNoStoreResponseWithoutETag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		handle settingsHandlerFunc
		calls  func(*recordingSettingsService) int
	}{
		{name: "access", handle: (*Handler).handleGetAccessSettings, calls: func(service *recordingSettingsService) int { return service.getAccessCalls }},
		{name: "SMTP", handle: (*Handler).handleGetSMTPSettings, calls: func(service *recordingSettingsService) int { return service.getSMTPCalls }},
		{name: "OIDC", handle: (*Handler).handleGetOIDCSettings, calls: func(service *recordingSettingsService) int { return service.getOIDCCalls }},
		{name: "Google", handle: (*Handler).handleGetGoogleSettings, calls: func(service *recordingSettingsService) int { return service.getGoogleCalls }},
		{name: "GitHub", handle: (*Handler).handleGetGitHubSettings, calls: func(service *recordingSettingsService) int { return service.getGitHubCalls }},
		{name: "updates", handle: (*Handler).handleGetUpdatesSettings, calls: func(service *recordingSettingsService) int { return service.getUpdatesCalls }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := &recordingSettingsService{}
			handler := &Handler{settings: service}
			recorder := httptest.NewRecorder()

			tt.handle(handler, recorder, authenticatedSettingsRequest(http.MethodGet, "/", ""))

			assertSettingsResponse(t, recorder)
			if got := tt.calls(service); got != 1 {
				t.Fatalf("expected one settings read, got %d", got)
			}
		})
	}
}

func TestGetGoogleSettingsWritesRedactedResponse(t *testing.T) {
	t.Parallel()

	service := &recordingSettingsService{
		google: appsettings.GoogleSettings{
			Enabled: true, ClientID: "google-client", ClientSecretSet: true,
			DisplayName: "Google", JITEnabled: true,
		},
	}
	handler := &Handler{settings: service}
	recorder := httptest.NewRecorder()

	handler.handleGetGoogleSettings(recorder, authenticatedSettingsRequest(http.MethodGet, "/", ""))

	assertSettingsResponse(t, recorder)
	var response struct {
		Settings map[string]any `json:"settings"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := response.Settings["clientSecret"]; ok {
		t.Fatal("response exposed clientSecret")
	}
	if got, ok := response.Settings["clientSecretSet"].(bool); !ok || !got {
		t.Fatalf("expected clientSecretSet=true, got %#v", response.Settings["clientSecretSet"])
	}
	if _, ok := response.Settings["callbackUrl"]; ok {
		t.Fatal("expected nil callbackUrl to be omitted")
	}
	domains, ok := response.Settings["allowedDomains"].([]any)
	if !ok || len(domains) != 0 {
		t.Fatalf("expected allowedDomains=[], got %#v", response.Settings["allowedDomains"])
	}
}

func TestUpdateSettingsAcceptsMissingAndIgnoredIfMatch(t *testing.T) {
	t.Parallel()

	requests := []struct {
		name          string
		ifMatch       string
		legacyIfMatch bool
	}{
		{name: "missing"},
		{name: "legacy", legacyIfMatch: true},
		{name: "malformed", ifMatch: "not-an-entity-tag"},
	}
	tests := []struct {
		name          string
		body          string
		legacyIfMatch string
		handle        settingsHandlerFunc
		calls         func(*recordingSettingsService) int
	}{
		{
			name: "access", body: `{"accountCreationEnabled":false}`, legacyIfMatch: `"access-99"`,
			handle: (*Handler).handleUpdateAccessSettings,
			calls:  func(service *recordingSettingsService) int { return service.updateAccessCalls },
		},
		{
			name: "SMTP", body: `{"port":2525}`, legacyIfMatch: `"smtp-99"`,
			handle: (*Handler).handleUpdateSMTPSettings,
			calls:  func(service *recordingSettingsService) int { return service.updateSMTPCalls },
		},
		{
			name: "OIDC", body: `{"enabled":false}`, legacyIfMatch: `"auth.oidc-99"`,
			handle: (*Handler).handleUpdateOIDCSettings,
			calls:  func(service *recordingSettingsService) int { return service.updateOIDCCalls },
		},
		{
			name: "Google", body: `{"enabled":false}`, legacyIfMatch: `"auth.google-99"`,
			handle: (*Handler).handleUpdateGoogleSettings,
			calls:  func(service *recordingSettingsService) int { return service.updateGoogleCalls },
		},
		{
			name: "GitHub", body: `{"enabled":false}`, legacyIfMatch: `"auth.github-99"`,
			handle: (*Handler).handleUpdateGitHubSettings,
			calls:  func(service *recordingSettingsService) int { return service.updateGitHubCalls },
		},
		{
			name: "updates", body: `{"checkForUpdates":false}`, legacyIfMatch: `"updates-99"`,
			handle: (*Handler).handleUpdateUpdatesSettings,
			calls:  func(service *recordingSettingsService) int { return service.updateUpdatesCalls },
		},
	}

	for _, tt := range tests {
		for _, requestCase := range requests {
			t.Run(tt.name+"/"+requestCase.name, func(t *testing.T) {
				t.Parallel()

				service := &recordingSettingsService{}
				handler := &Handler{settings: service}
				request := authenticatedSettingsRequest(http.MethodPatch, "/", tt.body)
				switch {
				case requestCase.legacyIfMatch:
					request.Header.Set("If-Match", tt.legacyIfMatch)
				case requestCase.ifMatch != "":
					request.Header.Set("If-Match", requestCase.ifMatch)
				}
				recorder := httptest.NewRecorder()

				tt.handle(handler, recorder, request)

				assertSettingsResponse(t, recorder)
				if got := tt.calls(service); got != 1 {
					t.Fatalf("expected one settings update, got %d", got)
				}
			})
		}
	}
}

func TestUpdateSMTPSettingsPassesSecretClear(t *testing.T) {
	t.Parallel()

	service := &recordingSettingsService{
		smtp: appsettings.SMTPSettings{Host: "smtp.example.com", Port: 2525, PasswordSet: false},
	}
	handler := &Handler{settings: service}
	request := authenticatedSettingsRequest(http.MethodPatch, "/", `{"port":2525,"password":null}`)
	recorder := httptest.NewRecorder()

	handler.handleUpdateSMTPSettings(recorder, request)

	assertSettingsResponse(t, recorder)
	if service.updateSMTPCalls != 1 {
		t.Fatalf("expected one update, got %d", service.updateSMTPCalls)
	}
	if service.lastUpdateSMTP.CurrentUserID != "user-1" {
		t.Fatalf("unexpected update input: %#v", service.lastUpdateSMTP)
	}
	if service.lastUpdateSMTP.Port == nil || *service.lastUpdateSMTP.Port != 2525 {
		t.Fatalf("expected port patch, got %#v", service.lastUpdateSMTP.Port)
	}
	if !service.lastUpdateSMTP.Password.Present || service.lastUpdateSMTP.Password.Value != nil {
		t.Fatalf("expected explicit password clear, got %#v", service.lastUpdateSMTP.Password)
	}
}

func TestSettingsActionsAcceptMissingAndIgnoredIfMatch(t *testing.T) {
	t.Parallel()

	requests := []struct {
		name          string
		ifMatch       string
		legacyIfMatch bool
	}{
		{name: "missing"},
		{name: "legacy", legacyIfMatch: true},
		{name: "malformed", ifMatch: "not-an-entity-tag"},
	}
	tests := []struct {
		name          string
		body          string
		legacyIfMatch string
		handle        settingsHandlerFunc
		calls         func(*recordingSettingsService) int
	}{
		{
			name: "OIDC validate", body: `{}`, legacyIfMatch: `"auth.oidc-99"`,
			handle: (*Handler).handleValidateOIDCSettings,
			calls:  func(service *recordingSettingsService) int { return service.validateOIDCCalls },
		},
		{
			name: "Google validate", body: `{}`, legacyIfMatch: `"auth.google-99"`,
			handle: (*Handler).handleValidateGoogleSettings,
			calls:  func(service *recordingSettingsService) int { return service.validateGoogleCalls },
		},
		{
			name: "GitHub validate", body: `{}`, legacyIfMatch: `"auth.github-99"`,
			handle: (*Handler).handleValidateGitHubSettings,
			calls:  func(service *recordingSettingsService) int { return service.validateGitHubCalls },
		},
		{
			name: "SMTP test", legacyIfMatch: `"smtp-99"`,
			handle: (*Handler).handleTestSMTP,
			calls:  func(service *recordingSettingsService) int { return service.testSMTPCalls },
		},
	}

	for _, tt := range tests {
		for _, requestCase := range requests {
			t.Run(tt.name+"/"+requestCase.name, func(t *testing.T) {
				t.Parallel()

				service := &recordingSettingsService{}
				handler := &Handler{settings: service}
				request := authenticatedSettingsRequest(http.MethodPost, "/", tt.body)
				switch {
				case requestCase.legacyIfMatch:
					request.Header.Set("If-Match", tt.legacyIfMatch)
				case requestCase.ifMatch != "":
					request.Header.Set("If-Match", requestCase.ifMatch)
				}
				recorder := httptest.NewRecorder()

				tt.handle(handler, recorder, request)

				if recorder.Code != http.StatusNoContent {
					t.Fatalf("expected status 204, got %d: %s", recorder.Code, recorder.Body.String())
				}
				if recorder.Body.Len() != 0 {
					t.Fatalf("expected empty response body, got %q", recorder.Body.String())
				}
				assertNoETag(t, recorder)
				if got := tt.calls(service); got != 1 {
					t.Fatalf("expected one settings action, got %d", got)
				}
			})
		}
	}
}

func TestValidateGitHubSettingsPassesCandidateWithoutWritingResponseBody(t *testing.T) {
	t.Parallel()

	service := &recordingSettingsService{}
	handler := &Handler{settings: service}
	request := authenticatedSettingsRequest(http.MethodPost, "/", `{"clientSecret":null,"allowSignup":false}`)
	recorder := httptest.NewRecorder()

	handler.handleValidateGitHubSettings(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("expected empty response body, got %q", recorder.Body.String())
	}
	if service.validateGitHubCalls != 1 {
		t.Fatalf("expected one validation call, got %d", service.validateGitHubCalls)
	}
	if !service.lastValidateGitHub.ClientSecret.Present || service.lastValidateGitHub.ClientSecret.Value != nil {
		t.Fatalf("expected explicit client secret clear, got %#v", service.lastValidateGitHub.ClientSecret)
	}
	if service.lastValidateGitHub.AllowSignup == nil || *service.lastValidateGitHub.AllowSignup {
		t.Fatalf("expected allowSignup=false, got %#v", service.lastValidateGitHub.AllowSignup)
	}
}

func TestAdminSensitiveSettingsRouteFailsClosedWithoutSudoService(t *testing.T) {
	t.Parallel()

	service := &recordingSettingsService{}
	handler := NewHandler(nil, staticAdminSessionManager{}, httpmiddleware.LocalSessionCookieName).ConfigureSettings(service)
	router := chi.NewRouter()
	handler.RegisterRoutes(router)

	request := httptest.NewRequest(http.MethodPatch, "/admin/settings/access", strings.NewReader(`{}`))
	request.AddCookie(&http.Cookie{Name: httpmiddleware.LocalSessionCookieName, Value: "session-token"})
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	assertProblem(t, recorder, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable)
	if service.updateAccessCalls != 0 {
		t.Fatalf("expected fail-closed middleware to stop the handler, got %d updates", service.updateAccessCalls)
	}
}

func TestWriteSettingsProblemMapsApplicationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "forbidden", err: appsettings.ErrForbidden, wantStatus: http.StatusForbidden, wantCode: httpx.CodeSystemAdminRequired},
		{name: "validation", err: appsettings.ErrInvalidInput, wantStatus: http.StatusUnprocessableEntity, wantCode: httpx.CodeValidationFailed},
		{name: "provider unavailable", err: appsettings.ErrProviderUnavailable, wantStatus: http.StatusServiceUnavailable, wantCode: httpx.CodeServiceUnavailable},
		{name: "SMTP failed", err: appsettings.ErrSMTPTestFailed, wantStatus: http.StatusServiceUnavailable, wantCode: httpx.CodeServiceUnavailable},
		{name: "unknown", err: errors.New("database failed"), wantStatus: http.StatusInternalServerError, wantCode: httpx.CodeInternalError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPatch, "/", http.NoBody)
			writeSettingsProblem(recorder, request, tt.err, "settings failed")
			assertProblem(t, recorder, tt.wantStatus, tt.wantCode)
		})
	}
}

func authenticatedSettingsRequest(method, target, body string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	ctx := httpmiddleware.WithPrincipal(request.Context(), httpmiddleware.Principal{
		Kind:   httpmiddleware.AuthKindSession,
		UserID: "user-1",
	})
	return request.WithContext(ctx)
}

func assertSettingsResponse(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Cache-Control"); got != settingsCacheControl {
		t.Fatalf("expected Cache-Control %q, got %q", settingsCacheControl, got)
	}
	assertNoETag(t, recorder)

	var response map[string]json.RawMessage
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode settings envelope: %v", err)
	}
	if _, ok := response["settings"]; !ok {
		t.Fatalf("expected settings response envelope, got %s", recorder.Body.String())
	}
}

func assertNoETag(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()
	if values := recorder.Header().Values("ETag"); len(values) != 0 {
		t.Fatalf("expected no ETag header, got %q", values)
	}
}

func assertProblem(t *testing.T, recorder *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if recorder.Code != status {
		t.Fatalf("expected status %d, got %d: %s", status, recorder.Code, recorder.Body.String())
	}
	var problem httpx.ProblemDetails
	if err := json.Unmarshal(recorder.Body.Bytes(), &problem); err != nil {
		t.Fatalf("decode problem: %v", err)
	}
	if problem.Code != code {
		t.Fatalf("expected problem code %q, got %q", code, problem.Code)
	}
}

type recordingSettingsService struct {
	access  appsettings.AccessSettings
	smtp    appsettings.SMTPSettings
	oidc    appsettings.OIDCSettings
	google  appsettings.GoogleSettings
	github  appsettings.GitHubSettings
	updates appsettings.UpdatesSettings

	getAccessErr      error
	getSMTPErr        error
	getOIDCErr        error
	getGoogleErr      error
	getGitHubErr      error
	getUpdatesErr     error
	updateAccessErr   error
	updateSMTPErr     error
	updateOIDCErr     error
	updateGoogleErr   error
	updateGitHubErr   error
	updateUpdatesErr  error
	validateOIDCErr   error
	validateGoogleErr error
	validateGitHubErr error
	testSMTPErr       error

	getAccessCalls      int
	getSMTPCalls        int
	getOIDCCalls        int
	getGoogleCalls      int
	getGitHubCalls      int
	getUpdatesCalls     int
	updateAccessCalls   int
	updateSMTPCalls     int
	updateOIDCCalls     int
	updateGoogleCalls   int
	updateGitHubCalls   int
	updateUpdatesCalls  int
	validateOIDCCalls   int
	validateGoogleCalls int
	validateGitHubCalls int
	testSMTPCalls       int

	lastUpdateAccess   appsettings.UpdateAccessInput
	lastUpdateSMTP     appsettings.UpdateSMTPInput
	lastValidateGitHub appsettings.ValidateGitHubInput
}

func (s *recordingSettingsService) GetAccess(_ context.Context, _ appsettings.GetAccessInput) (appsettings.AccessSettings, error) {
	s.getAccessCalls++
	return s.access, s.getAccessErr
}

func (s *recordingSettingsService) GetSMTP(_ context.Context, _ appsettings.GetSMTPInput) (appsettings.SMTPSettings, error) {
	s.getSMTPCalls++
	return s.smtp, s.getSMTPErr
}

func (s *recordingSettingsService) GetOIDC(_ context.Context, _ appsettings.GetOIDCInput) (appsettings.OIDCSettings, error) {
	s.getOIDCCalls++
	return s.oidc, s.getOIDCErr
}

func (s *recordingSettingsService) GetGoogle(_ context.Context, _ appsettings.GetGoogleInput) (appsettings.GoogleSettings, error) {
	s.getGoogleCalls++
	return s.google, s.getGoogleErr
}

func (s *recordingSettingsService) GetGitHub(_ context.Context, _ appsettings.GetGitHubInput) (appsettings.GitHubSettings, error) {
	s.getGitHubCalls++
	return s.github, s.getGitHubErr
}

func (s *recordingSettingsService) GetUpdates(_ context.Context, _ appsettings.GetUpdatesInput) (appsettings.UpdatesSettings, error) {
	s.getUpdatesCalls++
	return s.updates, s.getUpdatesErr
}

func (s *recordingSettingsService) UpdateAccess(_ context.Context, input appsettings.UpdateAccessInput) (appsettings.AccessSettings, error) {
	s.updateAccessCalls++
	s.lastUpdateAccess = input
	return s.access, s.updateAccessErr
}

func (s *recordingSettingsService) UpdateSMTP(_ context.Context, input appsettings.UpdateSMTPInput) (appsettings.SMTPSettings, error) {
	s.updateSMTPCalls++
	s.lastUpdateSMTP = input
	return s.smtp, s.updateSMTPErr
}

func (s *recordingSettingsService) UpdateOIDC(_ context.Context, _ appsettings.UpdateOIDCInput) (appsettings.OIDCSettings, error) {
	s.updateOIDCCalls++
	return s.oidc, s.updateOIDCErr
}

func (s *recordingSettingsService) UpdateGoogle(_ context.Context, _ appsettings.UpdateGoogleInput) (appsettings.GoogleSettings, error) {
	s.updateGoogleCalls++
	return s.google, s.updateGoogleErr
}

func (s *recordingSettingsService) UpdateGitHub(_ context.Context, _ appsettings.UpdateGitHubInput) (appsettings.GitHubSettings, error) {
	s.updateGitHubCalls++
	return s.github, s.updateGitHubErr
}

func (s *recordingSettingsService) UpdateUpdates(_ context.Context, _ appsettings.UpdateUpdatesInput) (appsettings.UpdatesSettings, error) {
	s.updateUpdatesCalls++
	return s.updates, s.updateUpdatesErr
}

func (s *recordingSettingsService) ValidateOIDC(_ context.Context, _ appsettings.ValidateOIDCInput) error {
	s.validateOIDCCalls++
	return s.validateOIDCErr
}

func (s *recordingSettingsService) ValidateGoogle(_ context.Context, _ appsettings.ValidateGoogleInput) error {
	s.validateGoogleCalls++
	return s.validateGoogleErr
}

func (s *recordingSettingsService) ValidateGitHub(_ context.Context, input appsettings.ValidateGitHubInput) error {
	s.validateGitHubCalls++
	s.lastValidateGitHub = input
	return s.validateGitHubErr
}

func (s *recordingSettingsService) TestSMTP(_ context.Context, _ appsettings.TestSMTPInput) error {
	s.testSMTPCalls++
	return s.testSMTPErr
}

type staticAdminSessionManager struct{}

func (staticAdminSessionManager) CreateSession(context.Context, appauth.CreateSessionInput) (identity.CreatedSession, error) {
	return identity.CreatedSession{}, nil
}

func (staticAdminSessionManager) VerifySession(context.Context, string) (identity.SessionClaims, error) {
	return identity.SessionClaims{SessionID: "session-1", UserID: "user-1"}, nil
}

func (staticAdminSessionManager) CreateCSRFToken(context.Context, string) (string, error) {
	return "", nil
}

func (staticAdminSessionManager) VerifyCSRFToken(context.Context, string, string) error {
	return nil
}

func (staticAdminSessionManager) RevokeSession(context.Context, string, string) error {
	return nil
}

func (staticAdminSessionManager) ListUserSessions(context.Context, string) ([]identity.AuthSession, error) {
	return nil, nil
}

func (staticAdminSessionManager) RevokeUserSession(context.Context, string, string, string) error {
	return nil
}

func (staticAdminSessionManager) RevokeUserSessions(context.Context, string, string) error {
	return nil
}
