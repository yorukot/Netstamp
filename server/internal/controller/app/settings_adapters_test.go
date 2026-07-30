package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	appsystemsettings "github.com/yorukot/netstamp/internal/controller/application/systemsettings"
	domainsystem "github.com/yorukot/netstamp/internal/domain/system"
)

func TestSystemSettingsPublicAccessProviderPropagatesSettingsFailures(t *testing.T) {
	t.Parallel()

	databaseErr := errors.New("database unavailable")
	tests := []struct {
		name       string
		repository publicAccessRepositoryFake
		wantErr    error
		wantText   string
	}{
		{
			name:       "database failure",
			repository: publicAccessRepositoryFake{err: databaseErr},
			wantErr:    databaseErr,
		},
		{
			name: "invalid stored value",
			repository: publicAccessRepositoryFake{settings: []domainsystem.Setting{{
				Key:   "auth.registration_enabled",
				Value: []byte(`"enabled"`),
			}}},
			wantText: `decode system setting "auth.registration_enabled"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := appsystemsettings.NewService(&test.repository, nil, nil, nil, appsystemsettings.Defaults{}, "")
			provider := systemSettingsPublicAccessProvider{service: service}

			settings, err := provider.PublicAccessSettings(context.Background())
			if err == nil {
				t.Fatal("expected public access settings to fail closed")
			}
			if test.wantErr != nil && !errors.Is(err, test.wantErr) {
				t.Fatalf("expected error %v, got %v", test.wantErr, err)
			}
			if test.wantText != "" && !strings.Contains(err.Error(), test.wantText) {
				t.Fatalf("expected error containing %q, got %v", test.wantText, err)
			}
			if settings.AccountCreationEnabled ||
				settings.ProjectCreationEnabled ||
				settings.CredentialChangesEnabled {
				t.Fatalf("expected zero settings on failure, got %#v", settings)
			}
		})
	}
}

func TestSystemSettingsPublicAccessProviderRejectsMissingService(t *testing.T) {
	t.Parallel()

	provider := systemSettingsPublicAccessProvider{}
	settings, err := provider.PublicAccessSettings(context.Background())
	if err == nil {
		t.Fatal("expected missing system settings service to fail closed")
	}
	if settings.AccountCreationEnabled ||
		settings.ProjectCreationEnabled ||
		settings.CredentialChangesEnabled {
		t.Fatalf("expected zero settings on failure, got %#v", settings)
	}
}

type publicAccessRepositoryFake struct {
	appsystemsettings.Repository
	settings []domainsystem.Setting
	err      error
}

func (f *publicAccessRepositoryFake) GetSystemSettingsByKeys(
	context.Context,
	[]string,
) ([]domainsystem.Setting, error) {
	return f.settings, f.err
}
