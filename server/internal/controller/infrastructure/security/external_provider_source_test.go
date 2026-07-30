package security

import (
	"context"
	"errors"
	"testing"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestDynamicExternalProviderSourceBuildsConfiguredProviders(t *testing.T) {
	settings := map[string]ExternalProviderSettings{
		identity.AuthenticationMethodOIDC: {
			Enabled: true, IssuerURL: "https://idp.example.com", ClientID: "oidc-client",
			ClientSecret: "oidc-secret", DisplayName: "Company SSO", JITEnabled: true,
		},
		identity.AuthenticationMethodGoogle: {
			Enabled: true, ClientID: "google-client", ClientSecret: "google-secret",
			DisplayName: "Google Workspace", AllowedDomains: []string{"example.com", "second.example.com"},
		},
		identity.AuthenticationMethodGitHub: {
			Enabled: true, ClientID: "github-client", ClientSecret: "github-secret",
			DisplayName: "GitHub", JITEnabled: true, AllowSignup: false,
		},
	}
	provider := &externalProviderSettingsFake{settings: settings}
	source := NewDynamicExternalProviderSource(
		provider,
		ExternalProviderCallbackURLs{
			OIDC:   "https://netstamp.example.com/oidc/callback",
			Google: "https://netstamp.example.com/google/callback",
			GitHub: "https://netstamp.example.com/github/callback",
		},
	)

	oidcRegistration, err := source.ExternalProviderRegistration(context.Background(), identity.AuthenticationMethodOIDC)
	if err != nil {
		t.Fatalf("resolve OIDC provider: %v", err)
	}
	if oidcRegistration.Config.ID != identity.AuthenticationMethodOIDC ||
		oidcRegistration.Config.DisplayName != "Company SSO" ||
		!oidcRegistration.Config.JITEnabled {
		t.Fatalf("unexpected OIDC registration: %#v", oidcRegistration.Config)
	}
	oidcClient, ok := oidcRegistration.Client.(*OIDCClient)
	if !ok ||
		oidcClient.cfg.IssuerURL != settings[identity.AuthenticationMethodOIDC].IssuerURL ||
		oidcClient.cfg.RedirectURL != source.callbackURLs.OIDC {
		t.Fatalf("unexpected OIDC client: %#v", oidcRegistration.Client)
	}

	googleRegistration, err := source.ExternalProviderRegistration(context.Background(), identity.AuthenticationMethodGoogle)
	if err != nil {
		t.Fatalf("resolve Google provider: %v", err)
	}
	googleClient, ok := googleRegistration.Client.(*OIDCClient)
	if !ok ||
		googleClient.cfg.RedirectURL != source.callbackURLs.Google ||
		len(googleClient.cfg.AllowedHostedDomains) != 2 {
		t.Fatalf("unexpected Google client: %#v", googleRegistration.Client)
	}

	githubRegistration, err := source.ExternalProviderRegistration(context.Background(), identity.AuthenticationMethodGitHub)
	if err != nil {
		t.Fatalf("resolve GitHub provider: %v", err)
	}
	githubClient, ok := githubRegistration.Client.(*GitHubOAuthClient)
	if !ok || githubClient.oauth2Config.RedirectURL != source.callbackURLs.GitHub || githubClient.allowSignup {
		t.Fatalf("unexpected GitHub client: %#v", githubRegistration.Client)
	}

	if len(provider.calls) != 3 {
		t.Fatalf("expected one settings lookup per provider, got %#v", provider.calls)
	}
}

func TestDynamicExternalProviderSourceResolvesOnlyRequestedProvider(t *testing.T) {
	expectedOIDCError := errors.New("OIDC secret cannot be decrypted")
	provider := &externalProviderSettingsFake{
		settings: map[string]ExternalProviderSettings{
			identity.AuthenticationMethodGoogle: {
				Enabled: true, ClientID: "google-client", ClientSecret: "google-secret", DisplayName: "Google",
			},
		},
		errs: map[string]error{
			identity.AuthenticationMethodOIDC: expectedOIDCError,
		},
	}
	source := NewDynamicExternalProviderSource(provider, ExternalProviderCallbackURLs{
		Google: "https://netstamp.example.com/api/v1/auth/external/google/callback",
	})

	googleRegistration, err := source.ExternalProviderRegistration(context.Background(), identity.AuthenticationMethodGoogle)
	if err != nil {
		t.Fatalf("resolve Google provider: %v", err)
	}
	if googleRegistration.Config.ID != identity.AuthenticationMethodGoogle {
		t.Fatalf("unexpected Google registration: %#v", googleRegistration.Config)
	}
	if len(provider.calls) != 1 || provider.calls[0] != identity.AuthenticationMethodGoogle {
		t.Fatalf("expected only Google settings lookup, got %#v", provider.calls)
	}

	_, err = source.ExternalProviderRegistration(context.Background(), identity.AuthenticationMethodOIDC)
	if !errors.Is(err, expectedOIDCError) {
		t.Fatalf("expected OIDC settings error, got %v", err)
	}
}

func TestDynamicExternalProviderSourceCachesEachProviderUntilItsSettingsChange(t *testing.T) {
	provider := &externalProviderSettingsFake{settings: map[string]ExternalProviderSettings{
		identity.AuthenticationMethodOIDC: {
			Enabled: true, IssuerURL: "https://idp.example.com", ClientID: "client", ClientSecret: "secret",
		},
		identity.AuthenticationMethodGoogle: {
			Enabled: true, ClientID: "google-client", ClientSecret: "google-secret",
		},
	}}
	source := NewDynamicExternalProviderSource(provider, ExternalProviderCallbackURLs{
		OIDC:   "https://netstamp.example.com/api/v1/auth/external/oidc/callback",
		Google: "https://netstamp.example.com/api/v1/auth/external/google/callback",
	})

	firstOIDC, err := source.ExternalProviderRegistration(context.Background(), identity.AuthenticationMethodOIDC)
	if err != nil {
		t.Fatalf("resolve first OIDC provider: %v", err)
	}
	firstGoogle, err := source.ExternalProviderRegistration(context.Background(), identity.AuthenticationMethodGoogle)
	if err != nil {
		t.Fatalf("resolve first Google provider: %v", err)
	}
	secondOIDC, err := source.ExternalProviderRegistration(context.Background(), identity.AuthenticationMethodOIDC)
	if err != nil {
		t.Fatalf("resolve second OIDC provider: %v", err)
	}
	if firstOIDC.Client != secondOIDC.Client {
		t.Fatal("expected unchanged OIDC settings to reuse the provider client")
	}

	updated := provider.settings[identity.AuthenticationMethodOIDC]
	updated.ClientSecret = "rotated-secret"
	provider.settings[identity.AuthenticationMethodOIDC] = updated
	thirdOIDC, err := source.ExternalProviderRegistration(context.Background(), identity.AuthenticationMethodOIDC)
	if err != nil {
		t.Fatalf("resolve changed OIDC provider: %v", err)
	}
	if secondOIDC.Client == thirdOIDC.Client {
		t.Fatal("expected changed OIDC settings to rebuild the provider client")
	}

	secondGoogle, err := source.ExternalProviderRegistration(context.Background(), identity.AuthenticationMethodGoogle)
	if err != nil {
		t.Fatalf("resolve second Google provider: %v", err)
	}
	if firstGoogle.Client != secondGoogle.Client {
		t.Fatal("expected an OIDC change not to rebuild the Google client")
	}
}

func TestDynamicExternalProviderSourceReturnsZeroRegistrationWhenDisabled(t *testing.T) {
	source := NewDynamicExternalProviderSource(
		&externalProviderSettingsFake{settings: map[string]ExternalProviderSettings{
			identity.AuthenticationMethodGitHub: {Enabled: false},
		}},
		ExternalProviderCallbackURLs{},
	)

	registration, err := source.ExternalProviderRegistration(context.Background(), identity.AuthenticationMethodGitHub)
	if err != nil {
		t.Fatalf("resolve disabled provider: %v", err)
	}
	if registration.Config.ID != "" || registration.Client != nil {
		t.Fatalf("expected zero registration, got %#v", registration)
	}
}

func TestDynamicExternalProviderSourceRejectsEnabledProviderWithoutAbsoluteHTTPCallback(t *testing.T) {
	tests := []struct {
		name        string
		callbackURL string
	}{
		{name: "missing"},
		{name: "relative path", callbackURL: "/api/v1/auth/external/google/callback"},
		{name: "missing scheme", callbackURL: "netstamp.example.com/api/v1/auth/external/google/callback"},
		{name: "unsupported scheme", callbackURL: "ftp://netstamp.example.com/google/callback"},
		{name: "credentials", callbackURL: "https://user:pass@netstamp.example.com/google/callback"},
		{name: "query", callbackURL: "https://netstamp.example.com/google/callback?next=1"},
		{name: "fragment", callbackURL: "https://netstamp.example.com/google/callback#fragment"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			provider := &externalProviderSettingsFake{settings: map[string]ExternalProviderSettings{
				identity.AuthenticationMethodGoogle: {
					Enabled: true, ClientID: "google-client", ClientSecret: "google-secret",
				},
			}}
			source := NewDynamicExternalProviderSource(provider, ExternalProviderCallbackURLs{
				Google: test.callbackURL,
			})

			registration, err := source.ExternalProviderRegistration(
				context.Background(),
				identity.AuthenticationMethodGoogle,
			)
			if err == nil {
				t.Fatal("expected invalid callback URL to make the provider unavailable")
			}
			if registration.Config.ID != "" || registration.Client != nil {
				t.Fatalf("expected zero registration, got %#v", registration)
			}
			if len(provider.calls) != 1 || provider.calls[0] != identity.AuthenticationMethodGoogle {
				t.Fatalf("expected one Google settings lookup, got %#v", provider.calls)
			}
		})
	}
}

func TestDynamicExternalProviderSourceReturnsIndependentProviderIDs(t *testing.T) {
	source := NewDynamicExternalProviderSource(&externalProviderSettingsFake{}, ExternalProviderCallbackURLs{})

	first := source.ExternalProviderIDs()
	if len(first) != 3 {
		t.Fatalf("expected three provider IDs, got %#v", first)
	}
	first[0] = "changed"

	second := source.ExternalProviderIDs()
	if second[0] != identity.AuthenticationMethodGoogle {
		t.Fatalf("expected caller mutation not to affect provider IDs, got %#v", second)
	}
}

func TestDynamicExternalProviderSourceRejectsUnknownProviderWithoutReadingSettings(t *testing.T) {
	provider := &externalProviderSettingsFake{}
	source := NewDynamicExternalProviderSource(provider, ExternalProviderCallbackURLs{})

	_, err := source.ExternalProviderRegistration(context.Background(), "unknown")
	if err == nil {
		t.Fatal("expected an unsupported provider error")
	}
	if len(provider.calls) != 0 {
		t.Fatalf("expected no settings lookup, got %#v", provider.calls)
	}
}

type externalProviderSettingsFake struct {
	settings map[string]ExternalProviderSettings
	errs     map[string]error
	calls    []string
}

func (p *externalProviderSettingsFake) ExternalProviderSettings(
	_ context.Context,
	provider string,
) (ExternalProviderSettings, error) {
	p.calls = append(p.calls, provider)
	if err := p.errs[provider]; err != nil {
		return ExternalProviderSettings{}, err
	}
	return p.settings[provider], nil
}
