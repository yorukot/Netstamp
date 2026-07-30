package security

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"sync"

	appauth "github.com/yorukot/netstamp/internal/controller/application/auth"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

var dynamicExternalProviderIDs = [...]string{
	identity.AuthenticationMethodGoogle,
	identity.AuthenticationMethodGitHub,
	identity.AuthenticationMethodOIDC,
}

type ExternalProviderSettings struct {
	Enabled        bool
	IssuerURL      string
	ClientID       string
	ClientSecret   string
	DisplayName    string
	JITEnabled     bool
	AllowedDomains []string
	AllowSignup    bool
}

type ExternalProviderSettingsProvider interface {
	ExternalProviderSettings(ctx context.Context, provider string) (ExternalProviderSettings, error)
}

type ExternalProviderCallbackURLs struct {
	OIDC   string
	Google string
	GitHub string
}

type cachedExternalProviderRegistration struct {
	settings     ExternalProviderSettings
	registration appauth.ExternalProviderRegistration
}

type DynamicExternalProviderSource struct {
	settingsProvider ExternalProviderSettingsProvider
	callbackURLs     ExternalProviderCallbackURLs

	mu     sync.Mutex
	cached map[string]cachedExternalProviderRegistration
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

func (s *DynamicExternalProviderSource) ExternalProviderIDs() []string {
	ids := make([]string, len(dynamicExternalProviderIDs))
	copy(ids, dynamicExternalProviderIDs[:])
	return ids
}

func (s *DynamicExternalProviderSource) ExternalProviderRegistration(
	ctx context.Context,
	provider string,
) (appauth.ExternalProviderRegistration, error) {
	if s == nil || s.settingsProvider == nil {
		return appauth.ExternalProviderRegistration{}, errors.New("external provider settings source is unavailable")
	}

	provider = strings.ToLower(strings.TrimSpace(provider))
	if !isDynamicExternalProvider(provider) {
		return appauth.ExternalProviderRegistration{}, fmt.Errorf("unsupported external provider %q", provider)
	}

	settings, err := s.settingsProvider.ExternalProviderSettings(ctx, provider)
	if err != nil {
		return appauth.ExternalProviderRegistration{}, err
	}
	if !settings.Enabled {
		s.deleteCachedRegistration(provider)
		return appauth.ExternalProviderRegistration{}, nil
	}
	settings = cloneExternalProviderSettings(settings)
	callbackURL, err := s.callbackURL(provider)
	if err != nil {
		s.deleteCachedRegistration(provider)
		return appauth.ExternalProviderRegistration{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if cached, ok := s.cached[provider]; ok && externalProviderSettingsEqual(cached.settings, settings) {
		return cached.registration, nil
	}

	registration := s.buildRegistration(provider, settings, callbackURL)
	if s.cached == nil {
		s.cached = make(map[string]cachedExternalProviderRegistration, len(dynamicExternalProviderIDs))
	}
	s.cached[provider] = cachedExternalProviderRegistration{
		settings:     settings,
		registration: registration,
	}
	return registration, nil
}

func (s *DynamicExternalProviderSource) deleteCachedRegistration(provider string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cached, provider)
}

func (s *DynamicExternalProviderSource) buildRegistration(
	provider string,
	settings ExternalProviderSettings,
	callbackURL string,
) appauth.ExternalProviderRegistration {
	config := appauth.ExternalProviderConfig{
		ID:          provider,
		DisplayName: settings.DisplayName,
		JITEnabled:  settings.JITEnabled,
		SudoCapable: true,
	}

	switch provider {
	case identity.AuthenticationMethodOIDC:
		return appauth.ExternalProviderRegistration{
			Config: config,
			Client: NewOIDCClient(OIDCClientConfig{
				IssuerURL:    settings.IssuerURL,
				ClientID:     settings.ClientID,
				ClientSecret: settings.ClientSecret,
				RedirectURL:  callbackURL,
			}),
		}
	case identity.AuthenticationMethodGoogle:
		return appauth.ExternalProviderRegistration{
			Config: config,
			Client: NewGoogleOIDCClient(GoogleOIDCClientConfig{
				ClientID:             settings.ClientID,
				ClientSecret:         settings.ClientSecret,
				RedirectURL:          callbackURL,
				AllowedHostedDomains: settings.AllowedDomains,
			}),
		}
	case identity.AuthenticationMethodGitHub:
		return appauth.ExternalProviderRegistration{
			Config: config,
			Client: NewGitHubOAuthClient(GitHubOAuthClientConfig{
				ClientID:     settings.ClientID,
				ClientSecret: settings.ClientSecret,
				RedirectURL:  callbackURL,
				AllowSignup:  settings.AllowSignup,
			}),
		}
	default:
		panic("unreachable external provider")
	}
}

func (s *DynamicExternalProviderSource) callbackURL(provider string) (string, error) {
	var raw string
	switch provider {
	case identity.AuthenticationMethodOIDC:
		raw = s.callbackURLs.OIDC
	case identity.AuthenticationMethodGoogle:
		raw = s.callbackURLs.Google
	case identity.AuthenticationMethodGitHub:
		raw = s.callbackURLs.GitHub
	default:
		return "", fmt.Errorf("unsupported external provider %q", provider)
	}

	raw = strings.TrimSpace(raw)
	parsed, err := url.Parse(raw)
	if err != nil ||
		!parsed.IsAbs() ||
		parsed.Host == "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.User != nil ||
		parsed.RawQuery != "" ||
		parsed.Fragment != "" {
		return "", fmt.Errorf("%s external provider callback URL must be an absolute HTTP URL", provider)
	}
	return raw, nil
}

func isDynamicExternalProvider(provider string) bool {
	return slices.Contains(dynamicExternalProviderIDs[:], provider)
}

func cloneExternalProviderSettings(settings ExternalProviderSettings) ExternalProviderSettings {
	settings.AllowedDomains = slices.Clone(settings.AllowedDomains)
	return settings
}

func externalProviderSettingsEqual(left, right ExternalProviderSettings) bool {
	return left.Enabled == right.Enabled &&
		left.IssuerURL == right.IssuerURL &&
		left.ClientID == right.ClientID &&
		left.ClientSecret == right.ClientSecret &&
		left.DisplayName == right.DisplayName &&
		left.JITEnabled == right.JITEnabled &&
		slices.Equal(left.AllowedDomains, right.AllowedDomains) &&
		left.AllowSignup == right.AllowSignup
}
