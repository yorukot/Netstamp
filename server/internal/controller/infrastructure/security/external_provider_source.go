package security

import (
	"context"
	"errors"
	"strings"
	"sync"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

type ExternalProviderSettings struct {
	Enabled        bool
	IssuerURL      string
	ClientID       string
	ClientSecret   string
	DisplayName    string
	JITEnabled     bool
	AllowedDomains string
	AllowSignup    bool
}

type ExternalProvidersSettings struct {
	OIDC   ExternalProviderSettings
	Google ExternalProviderSettings
	GitHub ExternalProviderSettings
}

type ExternalProviderSettingsProvider interface {
	ExternalProviderSettings(ctx context.Context) (ExternalProvidersSettings, error)
}

type ExternalProviderCallbackURLs struct {
	OIDC   string
	Google string
	GitHub string
}

type DynamicExternalProviderSource struct {
	settingsProvider ExternalProviderSettingsProvider
	callbackURLs     ExternalProviderCallbackURLs

	mu                  sync.Mutex
	cached              bool
	cachedSettings      ExternalProvidersSettings
	cachedRegistrations []appauth.ExternalProviderRegistration
}

func NewDynamicExternalProviderSource(
	settingsProvider ExternalProviderSettingsProvider,
	callbackURLs ExternalProviderCallbackURLs,
) *DynamicExternalProviderSource {
	return &DynamicExternalProviderSource{
		settingsProvider: settingsProvider,
		callbackURLs:     callbackURLs,
	}
}

func (s *DynamicExternalProviderSource) ExternalProviderRegistrations(ctx context.Context) ([]appauth.ExternalProviderRegistration, error) {
	if s == nil || s.settingsProvider == nil {
		return nil, errors.New("external provider settings source is unavailable")
	}
	settings, err := s.settingsProvider.ExternalProviderSettings(ctx)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.cached || s.cachedSettings != settings {
		s.cachedSettings = settings
		s.cachedRegistrations = s.buildRegistrations(settings)
		s.cached = true
	}

	registrations := make([]appauth.ExternalProviderRegistration, len(s.cachedRegistrations))
	copy(registrations, s.cachedRegistrations)
	return registrations, nil
}

func (s *DynamicExternalProviderSource) buildRegistrations(settings ExternalProvidersSettings) []appauth.ExternalProviderRegistration {
	registrations := make([]appauth.ExternalProviderRegistration, 0, 3)
	if settings.OIDC.Enabled {
		registrations = append(registrations, appauth.ExternalProviderRegistration{
			Config: appauth.ExternalProviderConfig{
				ID:          identity.AuthenticationMethodOIDC,
				DisplayName: settings.OIDC.DisplayName,
				JITEnabled:  settings.OIDC.JITEnabled,
				SudoCapable: true,
			},
			Client: NewOIDCClient(OIDCClientConfig{
				IssuerURL:    settings.OIDC.IssuerURL,
				ClientID:     settings.OIDC.ClientID,
				ClientSecret: settings.OIDC.ClientSecret,
				RedirectURL:  s.callbackURLs.OIDC,
			}),
		})
	}
	if settings.Google.Enabled {
		registrations = append(registrations, appauth.ExternalProviderRegistration{
			Config: appauth.ExternalProviderConfig{
				ID:          identity.AuthenticationMethodGoogle,
				DisplayName: settings.Google.DisplayName,
				JITEnabled:  settings.Google.JITEnabled,
				SudoCapable: true,
			},
			Client: NewGoogleOIDCClient(GoogleOIDCClientConfig{
				ClientID:             settings.Google.ClientID,
				ClientSecret:         settings.Google.ClientSecret,
				RedirectURL:          s.callbackURLs.Google,
				AllowedHostedDomains: strings.Split(settings.Google.AllowedDomains, ","),
			}),
		})
	}
	if settings.GitHub.Enabled {
		registrations = append(registrations, appauth.ExternalProviderRegistration{
			Config: appauth.ExternalProviderConfig{
				ID:          identity.AuthenticationMethodGitHub,
				DisplayName: settings.GitHub.DisplayName,
				JITEnabled:  settings.GitHub.JITEnabled,
				SudoCapable: true,
			},
			Client: NewGitHubOAuthClient(GitHubOAuthClientConfig{
				ClientID:     settings.GitHub.ClientID,
				ClientSecret: settings.GitHub.ClientSecret,
				RedirectURL:  s.callbackURLs.GitHub,
				AllowSignup:  settings.GitHub.AllowSignup,
			}),
		})
	}
	return registrations
}
