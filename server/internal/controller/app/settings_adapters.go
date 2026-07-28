package app

import (
	"context"
	"time"

	appadmin "github.com/yorukot/netstamp/internal/controller/application/admin"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/notify"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/security"
)

type adminSMTPProvider struct {
	service *appadmin.Service
}

func (p adminSMTPProvider) SMTPConfig(ctx context.Context) (notify.SMTPConfig, error) {
	settings, err := p.service.EffectiveSMTP(ctx)
	if err != nil {
		return notify.SMTPConfig{}, err
	}
	return notify.SMTPConfig{
		Host:     settings.Host,
		Port:     settings.Port,
		Username: settings.Username,
		Password: settings.Password,
		From:     settings.From,
		TLSMode:  settings.TLSMode,
		Timeout:  time.Duration(settings.TimeoutSeconds) * time.Second,
	}, nil
}

type adminExternalProviderSettingsProvider struct {
	service *appadmin.Service
}

func (p adminExternalProviderSettingsProvider) ExternalProviderSettings(ctx context.Context) (security.ExternalProvidersSettings, error) {
	settings, err := p.service.EffectiveSettings(ctx)
	if err != nil {
		return security.ExternalProvidersSettings{}, err
	}
	return security.ExternalProvidersSettings{
		OIDC: security.ExternalProviderSettings{
			Enabled:      settings.OIDC.Enabled,
			IssuerURL:    settings.OIDC.IssuerURL,
			ClientID:     settings.OIDC.ClientID,
			ClientSecret: settings.OIDC.ClientSecret,
			DisplayName:  settings.OIDC.DisplayName,
			JITEnabled:   settings.OIDC.JITEnabled,
		},
		Google: security.ExternalProviderSettings{
			Enabled:        settings.Google.Enabled,
			ClientID:       settings.Google.ClientID,
			ClientSecret:   settings.Google.ClientSecret,
			DisplayName:    settings.Google.DisplayName,
			JITEnabled:     settings.Google.JITEnabled,
			AllowedDomains: settings.Google.AllowedDomains,
		},
		GitHub: security.ExternalProviderSettings{
			Enabled:      settings.GitHub.Enabled,
			ClientID:     settings.GitHub.ClientID,
			ClientSecret: settings.GitHub.ClientSecret,
			DisplayName:  settings.GitHub.DisplayName,
			JITEnabled:   settings.GitHub.JITEnabled,
			AllowSignup:  settings.GitHub.AllowSignup,
		},
	}, nil
}
