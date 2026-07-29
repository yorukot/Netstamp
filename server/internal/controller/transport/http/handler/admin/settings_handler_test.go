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

func TestGetGoogleSettingsWritesVersionedRedactedResponse(t *testing.T) {
	t.Parallel()

	service := &recordingSettingsService{
		google: appsettings.Versioned[appsettings.GoogleSettings]{
			Value: appsettings.GoogleSettings{
				Enabled: true, ClientID: "google-client", ClientSecretSet: true,
				DisplayName: "Google", JITEnabled: true,
			},
			Revision: 4,
		},
	}
	handler := &Handler{settings: service}
	recorder := httptest.NewRecorder()

	handler.handleGetGoogleSettings(recorder, authenticatedSettingsRequest(http.MethodGet, "/", ""))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	assertSettingsVersionHeaders(t, recorder, `"auth.google-4"`)

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

func TestUpdateSMTPSettingsPassesSecretClearAndRevision(t *testing.T) {
	t.Parallel()

	service := &recordingSettingsService{
		smtp: appsettings.Versioned[appsettings.SMTPSettings]{
			Value:    appsettings.SMTPSettings{Host: "smtp.example.com", Port: 2525, PasswordSet: false},
			Revision: 3,
		},
	}
	handler := &Handler{settings: service}
	request := authenticatedSettingsRequest(http.MethodPatch, "/", `{"port":2525,"password":null}`)
	request.Header.Set("If-Match", `"smtp-2"`)
	recorder := httptest.NewRecorder()

	handler.handleUpdateSMTPSettings(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	assertSettingsVersionHeaders(t, recorder, `"smtp-3"`)
	if service.updateSMTPCalls != 1 {
		t.Fatalf("expected one update, got %d", service.updateSMTPCalls)
	}
	if service.lastUpdateSMTP.CurrentUserID != "user-1" || service.lastUpdateSMTP.ExpectedRevision != 2 {
		t.Fatalf("unexpected update input: %#v", service.lastUpdateSMTP)
	}
	if service.lastUpdateSMTP.Port == nil || *service.lastUpdateSMTP.Port != 2525 {
		t.Fatalf("expected port patch, got %#v", service.lastUpdateSMTP.Port)
	}
	if !service.lastUpdateSMTP.Password.Present || service.lastUpdateSMTP.Password.Value != nil {
		t.Fatalf("expected explicit password clear, got %#v", service.lastUpdateSMTP.Password)
	}
	if service.getSMTPCalls != 0 {
		t.Fatalf("valid strong ETag should not require a pre-read, got %d reads", service.getSMTPCalls)
	}
}

func TestUpdateAccessSettingsRequiresIfMatchBeforeCallingService(t *testing.T) {
	t.Parallel()

	service := &recordingSettingsService{}
	handler := &Handler{settings: service}
	recorder := httptest.NewRecorder()

	handler.handleUpdateAccessSettings(recorder, authenticatedSettingsRequest(http.MethodPatch, "/", `{}`))

	assertProblem(t, recorder, http.StatusPreconditionRequired, httpx.CodePreconditionRequired)
	if service.updateAccessCalls != 0 || service.getAccessCalls != 0 {
		t.Fatalf("expected no settings calls, got get=%d update=%d", service.getAccessCalls, service.updateAccessCalls)
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

func TestUpdateAccessSettingsRejectsWeakIfMatchWithCurrentETag(t *testing.T) {
	t.Parallel()

	service := &recordingSettingsService{
		access: appsettings.Versioned[appsettings.AccessSettings]{Revision: 7},
	}
	handler := &Handler{settings: service}
	request := authenticatedSettingsRequest(http.MethodPatch, "/", `{}`)
	request.Header.Set("If-Match", `W/"access-7"`)
	recorder := httptest.NewRecorder()

	handler.handleUpdateAccessSettings(recorder, request)

	assertProblem(t, recorder, http.StatusPreconditionFailed, httpx.CodeSettingsVersionConflict)
	assertSettingsVersionHeaders(t, recorder, `"access-7"`)
	if service.getAccessCalls != 1 || service.updateAccessCalls != 0 {
		t.Fatalf("expected one revision read and no update, got get=%d update=%d", service.getAccessCalls, service.updateAccessCalls)
	}
}

func TestUpdateAccessSettingsWritesCurrentETagForApplicationConflict(t *testing.T) {
	t.Parallel()

	service := &recordingSettingsService{
		updateAccessErr: &appsettings.VersionConflictError{
			Resource: appsettings.ResourceAccess,
			Expected: 3,
			Current:  5,
		},
	}
	handler := &Handler{settings: service}
	request := authenticatedSettingsRequest(http.MethodPatch, "/", `{"accountCreationEnabled":false}`)
	request.Header.Set("If-Match", `"access-3"`)
	recorder := httptest.NewRecorder()

	handler.handleUpdateAccessSettings(recorder, request)

	assertProblem(t, recorder, http.StatusPreconditionFailed, httpx.CodeSettingsVersionConflict)
	assertSettingsVersionHeaders(t, recorder, `"access-5"`)
	if service.lastUpdateAccess.AccountCreationEnabled == nil || *service.lastUpdateAccess.AccountCreationEnabled {
		t.Fatalf("expected explicit false patch, got %#v", service.lastUpdateAccess.AccountCreationEnabled)
	}
}

func TestValidateGitHubSettingsPassesCandidateWithoutWritingResponseBody(t *testing.T) {
	t.Parallel()

	service := &recordingSettingsService{}
	handler := &Handler{settings: service}
	request := authenticatedSettingsRequest(http.MethodPost, "/", `{"clientSecret":null,"allowSignup":false}`)
	request.Header.Set("If-Match", `"auth.github-9"`)
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
	if service.lastValidateGitHub.ExpectedRevision != 9 {
		t.Fatalf("expected revision 9, got %d", service.lastValidateGitHub.ExpectedRevision)
	}
	if !service.lastValidateGitHub.ClientSecret.Present || service.lastValidateGitHub.ClientSecret.Value != nil {
		t.Fatalf("expected explicit client secret clear, got %#v", service.lastValidateGitHub.ClientSecret)
	}
	if service.lastValidateGitHub.AllowSignup == nil || *service.lastValidateGitHub.AllowSignup {
		t.Fatalf("expected allowSignup=false, got %#v", service.lastValidateGitHub.AllowSignup)
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
		{name: "precondition", err: appsettings.ErrPreconditionRequired, wantStatus: http.StatusPreconditionRequired, wantCode: httpx.CodePreconditionRequired},
		{name: "provider unavailable", err: appsettings.ErrProviderUnavailable, wantStatus: http.StatusServiceUnavailable, wantCode: httpx.CodeServiceUnavailable},
		{name: "SMTP failed", err: appsettings.ErrSMTPTestFailed, wantStatus: http.StatusServiceUnavailable, wantCode: httpx.CodeServiceUnavailable},
		{name: "unknown", err: errors.New("database failed"), wantStatus: http.StatusInternalServerError, wantCode: httpx.CodeInternalError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPatch, "/", http.NoBody)
			writeSettingsProblem(recorder, request, appsettings.ResourceAccess, tt.err, "settings failed")
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

func assertSettingsVersionHeaders(t *testing.T, recorder *httptest.ResponseRecorder, etag string) {
	t.Helper()
	if got := recorder.Header().Get("ETag"); got != etag {
		t.Fatalf("expected ETag %q, got %q", etag, got)
	}
	if got := recorder.Header().Get("Cache-Control"); got != settingsCacheControl {
		t.Fatalf("expected Cache-Control %q, got %q", settingsCacheControl, got)
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
	access appsettings.Versioned[appsettings.AccessSettings]
	smtp   appsettings.Versioned[appsettings.SMTPSettings]
	oidc   appsettings.Versioned[appsettings.OIDCSettings]
	google appsettings.Versioned[appsettings.GoogleSettings]
	github appsettings.Versioned[appsettings.GitHubSettings]

	getAccessErr      error
	getSMTPErr        error
	getOIDCErr        error
	getGoogleErr      error
	getGitHubErr      error
	updateAccessErr   error
	updateSMTPErr     error
	updateOIDCErr     error
	updateGoogleErr   error
	updateGitHubErr   error
	validateOIDCErr   error
	validateGoogleErr error
	validateGitHubErr error
	testSMTPErr       error

	getAccessCalls      int
	getSMTPCalls        int
	updateAccessCalls   int
	updateSMTPCalls     int
	validateGitHubCalls int

	lastUpdateAccess   appsettings.UpdateAccessInput
	lastUpdateSMTP     appsettings.UpdateSMTPInput
	lastValidateGitHub appsettings.ValidateGitHubInput
}

func (s *recordingSettingsService) GetAccess(_ context.Context, _ appsettings.GetAccessInput) (appsettings.Versioned[appsettings.AccessSettings], error) {
	s.getAccessCalls++
	return s.access, s.getAccessErr
}

func (s *recordingSettingsService) GetSMTP(_ context.Context, _ appsettings.GetSMTPInput) (appsettings.Versioned[appsettings.SMTPSettings], error) {
	s.getSMTPCalls++
	return s.smtp, s.getSMTPErr
}

func (s *recordingSettingsService) GetOIDC(_ context.Context, _ appsettings.GetOIDCInput) (appsettings.Versioned[appsettings.OIDCSettings], error) {
	return s.oidc, s.getOIDCErr
}

func (s *recordingSettingsService) GetGoogle(_ context.Context, _ appsettings.GetGoogleInput) (appsettings.Versioned[appsettings.GoogleSettings], error) {
	return s.google, s.getGoogleErr
}

func (s *recordingSettingsService) GetGitHub(_ context.Context, _ appsettings.GetGitHubInput) (appsettings.Versioned[appsettings.GitHubSettings], error) {
	return s.github, s.getGitHubErr
}

func (s *recordingSettingsService) UpdateAccess(_ context.Context, input appsettings.UpdateAccessInput) (appsettings.Versioned[appsettings.AccessSettings], error) {
	s.updateAccessCalls++
	s.lastUpdateAccess = input
	return s.access, s.updateAccessErr
}

func (s *recordingSettingsService) UpdateSMTP(_ context.Context, input appsettings.UpdateSMTPInput) (appsettings.Versioned[appsettings.SMTPSettings], error) {
	s.updateSMTPCalls++
	s.lastUpdateSMTP = input
	return s.smtp, s.updateSMTPErr
}

func (s *recordingSettingsService) UpdateOIDC(_ context.Context, _ appsettings.UpdateOIDCInput) (appsettings.Versioned[appsettings.OIDCSettings], error) {
	return s.oidc, s.updateOIDCErr
}

func (s *recordingSettingsService) UpdateGoogle(_ context.Context, _ appsettings.UpdateGoogleInput) (appsettings.Versioned[appsettings.GoogleSettings], error) {
	return s.google, s.updateGoogleErr
}

func (s *recordingSettingsService) UpdateGitHub(_ context.Context, _ appsettings.UpdateGitHubInput) (appsettings.Versioned[appsettings.GitHubSettings], error) {
	return s.github, s.updateGitHubErr
}

func (s *recordingSettingsService) ValidateOIDC(_ context.Context, _ appsettings.ValidateOIDCInput) error {
	return s.validateOIDCErr
}

func (s *recordingSettingsService) ValidateGoogle(_ context.Context, _ appsettings.ValidateGoogleInput) error {
	return s.validateGoogleErr
}

func (s *recordingSettingsService) ValidateGitHub(_ context.Context, input appsettings.ValidateGitHubInput) error {
	s.validateGitHubCalls++
	s.lastValidateGitHub = input
	return s.validateGitHubErr
}

func (s *recordingSettingsService) TestSMTP(_ context.Context, _ appsettings.TestSMTPInput) error {
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
