package security

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestDynamicExternalProviderSourceBuildsConfiguredProviders(t *testing.T) {
	settings := ExternalProvidersSettings{
		OIDC: ExternalProviderSettings{
			Enabled: true, IssuerURL: "https://idp.example.com", ClientID: "oidc-client",
			ClientSecret: "oidc-secret", DisplayName: "Company SSO", JITEnabled: true,
		},
		Google: ExternalProviderSettings{
			Enabled: true, ClientID: "google-client", ClientSecret: "google-secret",
			DisplayName: "Google Workspace", AllowedDomains: "example.com, second.example.com",
		},
		GitHub: ExternalProviderSettings{
			Enabled: true, ClientID: "github-client", ClientSecret: "github-secret",
			DisplayName: "GitHub", JITEnabled: true, AllowSignup: false,
		},
	}
	source := NewDynamicExternalProviderSource(
		&externalProviderSettingsFake{settings: settings},
		ExternalProviderCallbackURLs{
			OIDC:   "https://netstamp.example.com/oidc/callback",
			Google: "https://netstamp.example.com/google/callback",
			GitHub: "https://netstamp.example.com/github/callback",
		},
	)

	registrations, err := source.ExternalProviderRegistrations(context.Background())
	if err != nil {
		t.Fatalf("resolve external providers: %v", err)
	}
	if len(registrations) != 3 {
		t.Fatalf("expected three providers, got %d", len(registrations))
	}
	if registrations[0].Config.ID != identity.AuthenticationMethodOIDC || registrations[0].Config.DisplayName != "Company SSO" || !registrations[0].Config.JITEnabled {
		t.Fatalf("unexpected OIDC registration: %#v", registrations[0].Config)
	}
	oidcClient, ok := registrations[0].Client.(*OIDCClient)
	if !ok || oidcClient.cfg.IssuerURL != settings.OIDC.IssuerURL || oidcClient.cfg.RedirectURL != source.callbackURLs.OIDC {
		t.Fatalf("unexpected OIDC client: %#v", registrations[0].Client)
	}
	googleClient, ok := registrations[1].Client.(*OIDCClient)
	if !ok || googleClient.cfg.RedirectURL != source.callbackURLs.Google || len(googleClient.cfg.AllowedHostedDomains) != 2 {
		t.Fatalf("unexpected Google client: %#v", registrations[1].Client)
	}
	githubClient, ok := registrations[2].Client.(*GitHubOAuthClient)
	if !ok || githubClient.oauth2Config.RedirectURL != source.callbackURLs.GitHub || githubClient.allowSignup {
		t.Fatalf("unexpected GitHub client: %#v", registrations[2].Client)
	}
}

func TestDynamicExternalProviderSourceCachesUntilSettingsChange(t *testing.T) {
	provider := &externalProviderSettingsFake{settings: ExternalProvidersSettings{
		OIDC: ExternalProviderSettings{Enabled: true, IssuerURL: "https://idp.example.com", ClientID: "client", ClientSecret: "secret"},
	}}
	source := NewDynamicExternalProviderSource(provider, ExternalProviderCallbackURLs{OIDC: "https://netstamp.example.com/callback"})

	first, err := source.ExternalProviderRegistrations(context.Background())
	if err != nil {
		t.Fatalf("resolve first external providers: %v", err)
	}
	second, err := source.ExternalProviderRegistrations(context.Background())
	if err != nil {
		t.Fatalf("resolve second external providers: %v", err)
	}
	if first[0].Client != second[0].Client {
		t.Fatal("expected unchanged settings to reuse the provider client")
	}

	provider.settings.OIDC.ClientSecret = "rotated-secret"
	third, err := source.ExternalProviderRegistrations(context.Background())
	if err != nil {
		t.Fatalf("resolve changed external providers: %v", err)
	}
	if second[0].Client == third[0].Client {
		t.Fatal("expected changed settings to rebuild the provider client")
	}
}

func TestDynamicExternalProviderSourcePropagatesSettingsFailure(t *testing.T) {
	expected := errors.New("settings unavailable")
	source := NewDynamicExternalProviderSource(&externalProviderSettingsFake{err: expected}, ExternalProviderCallbackURLs{})

	_, err := source.ExternalProviderRegistrations(context.Background())
	if !errors.Is(err, expected) {
		t.Fatalf("expected settings error, got %v", err)
	}
}

type externalProviderSettingsFake struct {
	settings ExternalProvidersSettings
	err      error
}

func (p *externalProviderSettingsFake) ExternalProviderSettings(context.Context) (ExternalProvidersSettings, error) {
	return p.settings, p.err
}
