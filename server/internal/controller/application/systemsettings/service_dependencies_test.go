package systemsettings

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

func TestServiceFailsClosedWhenStorageDependenciesAreUnavailable(t *testing.T) {
	t.Run("missing repository", func(t *testing.T) {
		service := NewService(
			nil,
			fakeSystemAdminChecker{},
			&fakeSecretCipher{},
			nil,
			Defaults{},
			"https://netstamp.example.com/api/v1/auth/external",
			passthroughSettingsTransactor{},
		)

		if _, err := service.EffectiveAccess(context.Background()); err == nil {
			t.Fatal("expected runtime access read to fail without a repository")
		}
		if _, err := service.GetAccess(context.Background(), GetAccessInput{
			CurrentUserID: testSystemSettingsAdminID,
		}); err == nil {
			t.Fatal("expected admin access read to fail without a repository")
		}
	})

	t.Run("missing transactor", func(t *testing.T) {
		service := NewService(
			newMemorySettingsRepository(),
			fakeSystemAdminChecker{},
			&fakeSecretCipher{},
			nil,
			Defaults{},
			"https://netstamp.example.com/api/v1/auth/external",
		)

		if _, err := service.GetAccess(context.Background(), GetAccessInput{
			CurrentUserID: testSystemSettingsAdminID,
		}); err == nil {
			t.Fatal("expected admin access read to fail without a transactor")
		}
	})
}

func TestEffectiveAccessRejectsStoredJSONNullInsteadOfUsingPermissiveDefaults(t *testing.T) {
	repo := newMemorySettingsRepository()
	repo.settings[keyAccountCreationEnabled] = StoredSetting{
		Key:   keyAccountCreationEnabled,
		Value: []byte(" \n null \t"),
	}
	service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, nil)

	got, err := service.EffectiveAccess(context.Background())
	if err == nil {
		t.Fatalf("expected JSON null to fail closed, got %#v", got)
	}
	if got != (AccessSettings{}) {
		t.Fatalf("expected zero access policy on invalid storage, got %#v", got)
	}
}

func TestEnabledProviderRejectsRelativeCallbackBaseURL(t *testing.T) {
	repo := newMemorySettingsRepository()
	repo.settings[keyGoogleSettings] = testPublicSetting(t, keyGoogleSettings, storedGoogleSettings{
		Enabled:     true,
		ClientID:    "google-client",
		DisplayName: "Google",
	})
	repo.settings[keyGoogleClientSecret] = testSecretSetting(keyGoogleClientSecret, "secret")
	service := NewService(
		repo,
		fakeSystemAdminChecker{},
		&fakeSecretCipher{},
		nil,
		Defaults{},
		"/api/v1/auth/external",
		passthroughSettingsTransactor{},
	)

	if _, err := service.EffectiveGoogle(context.Background()); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid callback configuration to fail closed, got %v", err)
	}
}

func TestProviderValidationRejectsMissingOrRelativeCallbackBaseURL(t *testing.T) {
	secret := "client-secret"
	enabled := true
	clientID := "client-id"
	issuerURL := "https://issuer.example.com"
	operations := []struct {
		name     string
		resource Resource
		run      func(context.Context, *Service) error
	}{
		{
			name:     "OIDC",
			resource: ResourceOIDC,
			run: func(ctx context.Context, service *Service) error {
				return service.ValidateOIDC(ctx, ValidateOIDCInput{
					CurrentUserID:    testSystemSettingsAdminID,
					ExpectedRevision: 1,
					Enabled:          &enabled,
					IssuerURL:        &issuerURL,
					ClientID:         &clientID,
					ClientSecret:     OptionalSecret{Present: true, Value: &secret},
				})
			},
		},
		{
			name:     "Google",
			resource: ResourceGoogle,
			run: func(ctx context.Context, service *Service) error {
				return service.ValidateGoogle(ctx, ValidateGoogleInput{
					CurrentUserID:    testSystemSettingsAdminID,
					ExpectedRevision: 1,
					Enabled:          &enabled,
					ClientID:         &clientID,
					ClientSecret:     OptionalSecret{Present: true, Value: &secret},
				})
			},
		},
		{
			name:     "GitHub",
			resource: ResourceGitHub,
			run: func(ctx context.Context, service *Service) error {
				return service.ValidateGitHub(ctx, ValidateGitHubInput{
					CurrentUserID:    testSystemSettingsAdminID,
					ExpectedRevision: 1,
					Enabled:          &enabled,
					ClientID:         &clientID,
					ClientSecret:     OptionalSecret{Present: true, Value: &secret},
				})
			},
		},
	}

	for _, callbackBaseURL := range []string{"", "/api/v1/auth/external"} {
		for _, operation := range operations {
			t.Run(operation.name+"/"+callbackBaseURL, func(t *testing.T) {
				repo := newMemorySettingsRepository()
				repo.revisions[string(operation.resource)] = 1
				service := NewService(
					repo,
					fakeSystemAdminChecker{},
					&fakeSecretCipher{},
					nil,
					Defaults{},
					callbackBaseURL,
					passthroughSettingsTransactor{},
				)

				err := operation.run(context.Background(), service)
				if !errors.Is(err, ErrInvalidInput) {
					t.Fatalf("expected invalid callback configuration, got %v", err)
				}
				assertSystemSettingsValidationField(t, err, "callbackUrl")
				assertNoSettingsWrites(t, repo)
			})
		}
	}
}

func TestSMTPTestUsesOneRevisionLockedSettingsSnapshot(t *testing.T) {
	repo := newMemorySettingsRepository()
	repo.revisions[string(ResourceSMTP)] = 7
	repo.settings[keySMTPHost] = testPublicSetting(t, keySMTPHost, "smtp.v1.example.com")
	repo.settings[keySMTPPort] = testPublicSetting(t, keySMTPPort, int32(465))
	repo.settings[keySMTPUsername] = testPublicSetting(t, keySMTPUsername, "mailer")
	repo.settings[keySMTPPassword] = testSecretSetting(keySMTPPassword, "password-v1")
	repo.settings[keySMTPFrom] = testPublicSetting(t, keySMTPFrom, "Netstamp <mail@example.com>")
	repo.settings[keySMTPTLSMode] = testPublicSetting(t, keySMTPTLSMode, "implicit")
	repo.settings[keySMTPTimeoutSeconds] = testPublicSetting(t, keySMTPTimeoutSeconds, int32(12))
	transactor := &rollbackSettingsTransactor{repo: repo}
	service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, transactor)
	users := &mutatingSMTPTestUsers{
		repo:         repo,
		user:         identity.User{Email: "admin@example.com"},
		nextHost:     testPublicSetting(t, keySMTPHost, "smtp.v2.example.com"),
		nextPassword: testSecretSetting(keySMTPPassword, "password-v2"),
	}
	tester := &recordingSMTPTester{}
	service.ConfigureSMTPTest(users, tester)

	err := service.TestSMTP(context.Background(), TestSMTPInput{
		CurrentUserID:    testSystemSettingsAdminID,
		ExpectedRevision: 7,
	})
	if err != nil {
		t.Fatalf("send SMTP test: %v", err)
	}

	wantSnapshot := SMTPRuntimeSettings{
		Host:           "smtp.v1.example.com",
		Port:           465,
		Username:       "mailer",
		Password:       "password-v1",
		From:           "Netstamp <mail@example.com>",
		TLSMode:        "implicit",
		TimeoutSeconds: 12,
	}
	if tester.recipient != "admin@example.com" || !reflect.DeepEqual(tester.settings, wantSnapshot) {
		t.Fatalf("unexpected SMTP test call: recipient=%q settings=%#v", tester.recipient, tester.settings)
	}
	if transactor.calls != 1 || transactor.commits != 1 {
		t.Fatalf("expected one committed snapshot transaction, got %#v", transactor)
	}
	if !reflect.DeepEqual(repo.lockAttempts, []string{string(ResourceSMTP)}) {
		t.Fatalf("unexpected revision locks: %#v", repo.lockAttempts)
	}
	if len(repo.readKeyQueries) != 1 || !reflect.DeepEqual(repo.readKeyQueries[0], smtpKeys) {
		t.Fatalf("expected one exact SMTP settings query, got %#v", repo.readKeyQueries)
	}
}

func TestSMTPTestChecksRevisionBeforeReadingMalformedSettings(t *testing.T) {
	repo := newMemorySettingsRepository()
	repo.revisions[string(ResourceSMTP)] = 2
	repo.settings[keySMTPTLSMode] = testPublicSetting(t, keySMTPTLSMode, "invalid")
	service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, nil)
	tester := &recordingSMTPTester{}
	service.ConfigureSMTPTest(&mutatingSMTPTestUsers{user: identity.User{Email: "admin@example.com"}}, tester)

	err := service.TestSMTP(context.Background(), TestSMTPInput{
		CurrentUserID:    testSystemSettingsAdminID,
		ExpectedRevision: 1,
	})
	if !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("expected stale revision error, got %v", err)
	}
	if len(repo.readKeyQueries) != 0 || tester.calls != 0 {
		t.Fatalf("stale request read or sent SMTP: reads=%#v sends=%d", repo.readKeyQueries, tester.calls)
	}
}

func TestEmailVerificationRejectsSMTPAuthenticationWithoutTLS(t *testing.T) {
	repo := newMemorySettingsRepository()
	repo.revisions[string(ResourceAccess)] = 3
	repo.revisions[string(ResourceSMTP)] = 5
	repo.settings[keySMTPHost] = testPublicSetting(t, keySMTPHost, "smtp.example.com")
	repo.settings[keySMTPUsername] = testPublicSetting(t, keySMTPUsername, "mailer")
	repo.settings[keySMTPPassword] = testSecretSetting(keySMTPPassword, "password")
	repo.settings[keySMTPFrom] = testPublicSetting(t, keySMTPFrom, "Netstamp <mail@example.com>")
	repo.settings[keySMTPTLSMode] = testPublicSetting(t, keySMTPTLSMode, "none")
	service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, nil)
	required := true

	_, err := service.UpdateAccess(context.Background(), UpdateAccessInput{
		CurrentUserID:             testSystemSettingsAdminID,
		ExpectedRevision:          3,
		EmailVerificationRequired: &required,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected SMTP TLS validation error, got %v", err)
	}
	assertNoSettingsWrites(t, repo)
}

func TestExplicitNullClearsMalformedSecretRows(t *testing.T) {
	t.Run("SMTP", func(t *testing.T) {
		repo := newMemorySettingsRepository()
		repo.revisions[string(ResourceAccess)] = 1
		repo.revisions[string(ResourceSMTP)] = 4
		repo.settings[keySMTPPassword] = StoredSetting{
			Key:   keySMTPPassword,
			Value: []byte(`"not-a-secret-row"`),
		}
		service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, nil)

		got, err := service.UpdateSMTP(context.Background(), UpdateSMTPInput{
			CurrentUserID:    testSystemSettingsAdminID,
			ExpectedRevision: 4,
			Password:         OptionalSecret{Present: true},
		})
		if err != nil {
			t.Fatalf("clear malformed SMTP password row: %v", err)
		}
		if got.Revision != 5 || got.Value.PasswordSet {
			t.Fatalf("unexpected SMTP clear result: %#v", got)
		}
		assertMalformedSecretCleared(t, repo, keySMTPPassword, ResourceSMTP)
	})

	t.Run("provider", func(t *testing.T) {
		repo := newMemorySettingsRepository()
		repo.revisions[string(ResourceGoogle)] = 6
		repo.settings[keyGoogleClientSecret] = StoredSetting{
			Key:   keyGoogleClientSecret,
			Value: []byte(`"not-a-secret-row"`),
		}
		service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, nil)

		got, err := service.UpdateGoogle(context.Background(), UpdateGoogleInput{
			CurrentUserID:    testSystemSettingsAdminID,
			ExpectedRevision: 6,
			ClientSecret:     OptionalSecret{Present: true},
		})
		if err != nil {
			t.Fatalf("clear malformed Google client secret row: %v", err)
		}
		if got.Revision != 7 || got.Value.ClientSecretSet {
			t.Fatalf("unexpected provider clear result: %#v", got)
		}
		assertMalformedSecretCleared(t, repo, keyGoogleClientSecret, ResourceGoogle)
	})
}

func assertMalformedSecretCleared(
	t *testing.T,
	repo *memorySettingsRepository,
	key string,
	resource Resource,
) {
	t.Helper()
	if _, exists := repo.settings[key]; exists {
		t.Fatalf("malformed secret row %q still exists", key)
	}
	if !containsString(repo.deleteAttempts, key) ||
		gotAuditAction(repo.auditEvents, key) != auditActionClear ||
		!containsString(repo.bumpAttempts, string(resource)) {
		t.Fatalf(
			"expected delete, clear audit, and revision bump; deletes=%#v audits=%#v bumps=%#v",
			repo.deleteAttempts,
			repo.auditEvents,
			repo.bumpAttempts,
		)
	}
}

type mutatingSMTPTestUsers struct {
	repo         *memorySettingsRepository
	user         identity.User
	nextHost     StoredSetting
	nextPassword StoredSetting
}

func (f *mutatingSMTPTestUsers) GetUserByID(context.Context, string) (identity.User, error) {
	if f.repo != nil {
		f.repo.settings[keySMTPHost] = f.nextHost
		f.repo.settings[keySMTPPassword] = f.nextPassword
	}
	return f.user, nil
}

type recordingSMTPTester struct {
	recipient string
	settings  SMTPRuntimeSettings
	calls     int
}

func (f *recordingSMTPTester) SendTestEmail(
	_ context.Context,
	recipient string,
	settings SMTPRuntimeSettings,
) error {
	f.calls++
	f.recipient = recipient
	f.settings = settings
	return nil
}
