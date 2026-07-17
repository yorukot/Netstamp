package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestAuthMethodsReturnsLowerCamelProviderMetadata(t *testing.T) {
	service := appauth.NewService(nil, nil, nil, nil)
	service.ConfigureExternalAuth(nil, nil, appauth.ExternalAuthConfig{},
		appauth.ExternalProviderRegistration{
			Config: appauth.ExternalProviderConfig{ID: identity.AuthenticationMethodGitHub, DisplayName: "GitHub", SudoCapable: true},
			Client: &externalAuthClientStub{},
		},
		appauth.ExternalProviderRegistration{
			Config: appauth.ExternalProviderConfig{ID: identity.AuthenticationMethodGoogle, DisplayName: "Google", SudoCapable: true},
			Client: &externalAuthClientStub{},
		},
	)
	handler := NewHandler(service, nil, nil, "netstamp_session", false, true)
	recorder := httptest.NewRecorder()
	handler.handleAuthMethods(recorder, httptest.NewRequest(http.MethodGet, "/auth/methods", http.NoBody))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected auth methods status 200, got %d", recorder.Code)
	}
	var body struct {
		Providers []externalProviderMethodResponse `json:"providers"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode auth methods response: %v", err)
	}
	if len(body.Providers) != 2 || body.Providers[0].ID != identity.AuthenticationMethodGoogle || !body.Providers[0].SudoCapable || !body.Providers[1].SudoCapable {
		t.Fatalf("unexpected provider metadata: %#v", body.Providers)
	}
}

func TestExternalAuthErrorRedirectPreservesSensitiveReturnPath(t *testing.T) {
	handler := &Handler{publicWebBaseURL: "https://netstamp.example.com"}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/external/github/callback", http.NoBody)

	handler.redirectExternalAuthError(recorder, request, identity.AuthenticationMethodGitHub, "callback_invalid", "/settings?reauth=set-password")

	location, err := url.Parse(recorder.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	query := location.Query()
	if location.Path != "/settings" || query.Get("reauth") != "set-password" || query.Get("auth_error") != "callback_invalid" || query.Get("auth_provider") != identity.AuthenticationMethodGitHub {
		t.Fatalf("unexpected sensitive auth error redirect: %s", location.String())
	}
}

type externalAuthClientStub struct{}

func (*externalAuthClientStub) AuthorizationURL(context.Context, string, string, string, string) (string, error) {
	return "https://identity.example.com/authorize", nil
}

func (*externalAuthClientStub) Exchange(context.Context, string, string, string) (appauth.ExternalIdentityClaims, error) {
	return appauth.ExternalIdentityClaims{}, nil
}
