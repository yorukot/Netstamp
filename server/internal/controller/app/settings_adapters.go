package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	appadmin "github.com/yorukot/netstamp/internal/controller/application/admin"
	appsystemsettings "github.com/yorukot/netstamp/internal/controller/application/systemsettings"
	appupdatecheck "github.com/yorukot/netstamp/internal/controller/application/updatecheck"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/notify"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/security"
	httpserver "github.com/yorukot/netstamp/internal/controller/transport/http"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

type updateStatusProvider struct {
	cache *appupdatecheck.Cache
}

func (p updateStatusProvider) ReadUpdateStatus() appadmin.UpdateStatus {
	status := p.cache.Snapshot()
	return appadmin.UpdateStatus{
		CurrentVersion: status.CurrentVersion, LatestVersion: optionalString(status.LatestVersion),
		UpdateAvailable: status.UpdateAvailable, ReleaseURL: optionalString(status.ReleaseURL),
		PublishedAt: optionalTime(status.PublishedAt), LastCheckedAt: optionalTime(status.LastCheckedAt),
		CheckError: optionalString(status.CheckError),
	}
}

func optionalString(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func optionalTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	value = value.UTC()
	return &value
}

type systemSettingsSMTPProvider struct {
	service *appsystemsettings.Service
}

func (p systemSettingsSMTPProvider) SMTPConfig(ctx context.Context) (notify.SMTPConfig, error) {
	settings, err := p.service.EffectiveSMTP(ctx)
	if err != nil {
		return notify.SMTPConfig{}, err
	}
	return notifySMTPConfig(settings), nil
}

func notifySMTPConfig(settings appsystemsettings.SMTPRuntimeSettings) notify.SMTPConfig {
	return notify.SMTPConfig{
		Host:     settings.Host,
		Port:     settings.Port,
		Username: settings.Username,
		Password: settings.Password,
		From:     settings.From,
		TLSMode:  settings.TLSMode,
		Timeout:  time.Duration(settings.TimeoutSeconds) * time.Second,
	}
}

func (p systemSettingsSMTPProvider) SMTPConfigured(ctx context.Context) bool {
	settings, err := p.service.EffectiveSMTP(ctx)
	return err == nil && strings.TrimSpace(settings.Host) != "" && strings.TrimSpace(settings.From) != ""
}

type systemSettingsSMTPTester struct{}

func (systemSettingsSMTPTester) SendTestEmail(
	ctx context.Context,
	recipient string,
	settings appsystemsettings.SMTPRuntimeSettings,
) error {
	provider := staticSMTPConfigProvider{config: notifySMTPConfig(settings)}
	tester := notify.NewSMTPTester(notify.NewDynamicSMTPSender(provider))
	return tester.SendTestEmail(ctx, recipient)
}

type staticSMTPConfigProvider struct {
	config notify.SMTPConfig
}

func (p staticSMTPConfigProvider) SMTPConfig(context.Context) (notify.SMTPConfig, error) {
	return p.config, nil
}

type systemSettingsPublicAccessProvider struct {
	service *appsystemsettings.Service
}

func (p systemSettingsPublicAccessProvider) PublicAccessSettings(ctx context.Context) (httpserver.PublicAccessSettings, error) {
	if p.service == nil {
		return httpserver.PublicAccessSettings{}, errors.New("system settings service is unavailable")
	}
	settings, err := p.service.EffectiveAccess(ctx)
	if err != nil {
		return httpserver.PublicAccessSettings{}, err
	}
	return httpserver.PublicAccessSettings{
		AccountCreationEnabled:   settings.AccountCreationEnabled,
		ProjectCreationEnabled:   settings.ProjectCreationEnabled,
		CredentialChangesEnabled: settings.CredentialChangesEnabled,
	}, nil
}

type systemSettingsExternalProvider struct {
	service *appsystemsettings.Service
}

func (p systemSettingsExternalProvider) ExternalProviderSettings(
	ctx context.Context,
	provider string,
) (security.ExternalProviderSettings, error) {
	switch provider {
	case identity.AuthenticationMethodOIDC:
		settings, err := p.service.EffectiveOIDC(ctx)
		if err != nil {
			return security.ExternalProviderSettings{}, err
		}
		return security.ExternalProviderSettings{
			Enabled:      settings.Enabled,
			IssuerURL:    settings.IssuerURL,
			ClientID:     settings.ClientID,
			ClientSecret: settings.ClientSecret,
			DisplayName:  settings.DisplayName,
			JITEnabled:   settings.JITEnabled,
		}, nil
	case identity.AuthenticationMethodGoogle:
		settings, err := p.service.EffectiveGoogle(ctx)
		if err != nil {
			return security.ExternalProviderSettings{}, err
		}
		return security.ExternalProviderSettings{
			Enabled:        settings.Enabled,
			ClientID:       settings.ClientID,
			ClientSecret:   settings.ClientSecret,
			DisplayName:    settings.DisplayName,
			JITEnabled:     settings.JITEnabled,
			AllowedDomains: settings.AllowedDomains,
		}, nil
	case identity.AuthenticationMethodGitHub:
		settings, err := p.service.EffectiveGitHub(ctx)
		if err != nil {
			return security.ExternalProviderSettings{}, err
		}
		return security.ExternalProviderSettings{
			Enabled:      settings.Enabled,
			ClientID:     settings.ClientID,
			ClientSecret: settings.ClientSecret,
			DisplayName:  settings.DisplayName,
			JITEnabled:   settings.JITEnabled,
			AllowSignup:  settings.AllowSignup,
		}, nil
	default:
		return security.ExternalProviderSettings{}, fmt.Errorf("unsupported external provider %q", provider)
	}
}
