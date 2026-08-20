package systemsettings

import (
	"bytes"
	"context"
	"errors"
	"net"
	"reflect"
	"testing"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainsystem "github.com/yorukot/netstamp/internal/domain/system"
)

const testSystemSettingsAdminID = "11111111-1111-1111-1111-111111111111"

func TestEffectiveUpdatesDefaultsEnabledAndReadsStoredValue(t *testing.T) {
	repo := newMemorySettingsRepository()
	service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, nil)

	defaults, err := service.EffectiveUpdates(context.Background())
	if err != nil {
		t.Fatalf("read default updates settings: %v", err)
	}
	if !defaults.CheckForUpdates {
		t.Fatalf("expected update checks enabled by default, got %#v", defaults)
	}

	repo.settings[keyUpdateCheckEnabled] = testPublicSetting(t, keyUpdateCheckEnabled, false)
	stored, err := service.EffectiveUpdates(context.Background())
	if err != nil {
		t.Fatalf("read stored updates settings: %v", err)
	}
	if stored.CheckForUpdates {
		t.Fatalf("expected stored update check disable, got %#v", stored)
	}
}

func TestUpdateUpdatesUsesResourceLockAndSkipsNoopWrites(t *testing.T) {
	repo := newMemorySettingsRepository()
	service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, nil)
	disabled := false

	updated, err := service.UpdateUpdates(context.Background(), UpdateUpdatesInput{
		CurrentUserID: testSystemSettingsAdminID, CheckForUpdates: &disabled,
	})
	if err != nil {
		t.Fatalf("update updates settings: %v", err)
	}
	if updated.CheckForUpdates {
		t.Fatalf("expected update checks disabled, got %#v", updated)
	}
	if !reflect.DeepEqual(repo.lockAttempts, []string{string(ResourceUpdates)}) {
		t.Fatalf("unexpected updates lock attempts: %#v", repo.lockAttempts)
	}
	if !reflect.DeepEqual(repo.upsertAttempts, []string{keyUpdateCheckEnabled}) || repo.auditAttempts != 1 {
		t.Fatalf("unexpected writes: upserts=%#v audits=%d", repo.upsertAttempts, repo.auditAttempts)
	}

	if _, err := service.UpdateUpdates(context.Background(), UpdateUpdatesInput{
		CurrentUserID: testSystemSettingsAdminID, CheckForUpdates: &disabled,
	}); err != nil {
		t.Fatalf("repeat updates settings update: %v", err)
	}
	if len(repo.upsertAttempts) != 1 || repo.auditAttempts != 1 {
		t.Fatal("no-op updates patch wrote a setting or audit event")
	}
}

func TestUpdateAccessRollsBackWritesWhenAuditFails(t *testing.T) {
	auditErr := errors.New("audit failed")
	repo := newMemorySettingsRepository()
	repo.auditErr = auditErr
	transactor := &rollbackSettingsTransactor{repo: repo}
	service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, transactor)
	accountCreationEnabled := false
	projectCreationEnabled := false

	got, err := service.UpdateAccess(context.Background(), UpdateAccessInput{
		CurrentUserID:          testSystemSettingsAdminID,
		AccountCreationEnabled: &accountCreationEnabled,
		ProjectCreationEnabled: &projectCreationEnabled,
	})
	if !errors.Is(err, auditErr) {
		t.Fatalf("expected audit error, got %v", err)
	}
	if got != (AccessSettings{}) {
		t.Fatalf("expected zero result after rollback, got %#v", got)
	}
	if len(repo.upsertAttempts) != 1 || repo.auditAttempts != 1 {
		t.Fatalf(
			"expected a write before the audit failure, got upserts=%#v audits=%d",
			repo.upsertAttempts,
			repo.auditAttempts,
		)
	}
	if len(repo.settings) != 0 || len(repo.auditEvents) != 0 {
		t.Fatalf("expected settings and audits to roll back, got settings=%#v audits=%#v", repo.settings, repo.auditEvents)
	}
	assertAccessSMTPLockOrder(t, repo.lockAttempts)
	if transactor.calls != 1 || transactor.rollbacks != 1 || transactor.commits != 0 {
		t.Fatalf("unexpected transaction outcome: %#v", transactor)
	}
}

func TestUpdateAccessUsesResourceLocksAndSkipsNoopWrites(t *testing.T) {
	repo := newMemorySettingsRepository()
	transactor := &rollbackSettingsTransactor{repo: repo}
	service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, transactor)
	projectCreationEnabled := false

	updated, err := service.UpdateAccess(context.Background(), UpdateAccessInput{
		CurrentUserID:          testSystemSettingsAdminID,
		ProjectCreationEnabled: &projectCreationEnabled,
	})
	if err != nil {
		t.Fatalf("update access: %v", err)
	}
	if updated.ProjectCreationEnabled {
		t.Fatalf("unexpected first update: %#v", updated)
	}
	assertAccessSMTPLockOrder(t, repo.lockAttempts)

	upsertsAfterUpdate := len(repo.upsertAttempts)
	auditsAfterUpdate := repo.auditAttempts

	noChange, err := service.UpdateAccess(context.Background(), UpdateAccessInput{
		CurrentUserID:          testSystemSettingsAdminID,
		ProjectCreationEnabled: &projectCreationEnabled,
	})
	if err != nil {
		t.Fatalf("repeat access update: %v", err)
	}
	if noChange.ProjectCreationEnabled {
		t.Fatalf("unexpected no-op result: %#v", noChange)
	}
	if len(repo.upsertAttempts) != upsertsAfterUpdate ||
		repo.auditAttempts != auditsAfterUpdate {
		t.Fatal("no-op update wrote a setting or audit event")
	}
	wantLocks := []string{
		string(ResourceAccess), string(ResourceSMTP),
		string(ResourceAccess), string(ResourceSMTP),
	}
	if !reflect.DeepEqual(repo.lockAttempts, wantLocks) {
		t.Fatalf("unexpected lock order across updates: got %#v want %#v", repo.lockAttempts, wantLocks)
	}
}

func TestUpdateGooglePreservesNullableSecretPatchStates(t *testing.T) {
	replacement := "new-secret"
	workspaceDisplayName := "Google Workspace"
	tests := []struct {
		name             string
		existingSecret   string
		providerEnabled  bool
		secretPatch      OptionalSecret
		displayName      *string
		wantSecret       bool
		wantPlaintext    string
		wantSecretUpsert bool
		wantSecretDelete bool
		wantSecretAudit  string
	}{
		{
			name:             "set",
			providerEnabled:  true,
			secretPatch:      OptionalSecret{Present: true, Value: &replacement},
			wantSecret:       true,
			wantPlaintext:    replacement,
			wantSecretUpsert: true,
			wantSecretAudit:  auditActionUpdate,
		},
		{
			name:             "clear",
			existingSecret:   "old-secret",
			secretPatch:      OptionalSecret{Present: true},
			wantSecretDelete: true,
			wantSecretAudit:  auditActionClear,
		},
		{
			name:            "omit",
			existingSecret:  "old-secret",
			providerEnabled: true,
			displayName:     &workspaceDisplayName,
			wantSecret:      true,
			wantPlaintext:   "old-secret",
		},
		{
			name:        "clear already absent",
			secretPatch: OptionalSecret{Present: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newMemorySettingsRepository()
			if test.providerEnabled {
				repo.settings[keyGoogleSettings] = testPublicSetting(t, keyGoogleSettings, storedGoogleSettings{
					Enabled:     true,
					ClientID:    "google-client",
					DisplayName: "Google",
				})
			}
			if test.existingSecret != "" {
				repo.settings[keyGoogleClientSecret] = testSecretSetting(keyGoogleClientSecret, test.existingSecret)
			}
			cipher := &fakeSecretCipher{}
			service := newTestSettingsService(repo, cipher, nil, nil)

			got, err := service.UpdateGoogle(context.Background(), UpdateGoogleInput{
				CurrentUserID: testSystemSettingsAdminID,
				ClientSecret:  test.secretPatch,
				DisplayName:   test.displayName,
			})
			if err != nil {
				t.Fatalf("update Google settings: %v", err)
			}
			if got.ClientSecretSet != test.wantSecret {
				t.Fatalf("unexpected update result: %#v", got)
			}

			storedSecret, secretExists := repo.settings[keyGoogleClientSecret]
			if secretExists != test.wantSecret {
				t.Fatalf("expected secret presence %t, got setting %#v", test.wantSecret, storedSecret)
			}
			if test.wantSecret {
				if !storedSecret.Secret ||
					string(storedSecret.EncryptedValue) != "enc:"+test.wantPlaintext ||
					string(storedSecret.EncryptedValueNonce) != "nonce" {
					t.Fatalf("unexpected stored secret: %#v", storedSecret)
				}
			}
			if containsString(repo.upsertAttempts, keyGoogleClientSecret) != test.wantSecretUpsert {
				t.Fatalf("unexpected secret upsert attempts: %#v", repo.upsertAttempts)
			}
			if containsString(repo.deleteAttempts, keyGoogleClientSecret) != test.wantSecretDelete {
				t.Fatalf("unexpected secret delete attempts: %#v", repo.deleteAttempts)
			}
			if gotAuditAction(repo.auditEvents, keyGoogleClientSecret) != test.wantSecretAudit {
				t.Fatalf("unexpected secret audit events: %#v", repo.auditEvents)
			}
			if !reflect.DeepEqual(repo.lockAttempts, []string{string(ResourceGoogle)}) {
				t.Fatalf("unexpected provider locks: %#v", repo.lockAttempts)
			}

			runtime, err := service.EffectiveGoogle(context.Background())
			if err != nil {
				t.Fatalf("read runtime Google settings: %v", err)
			}
			if runtime.ClientSecret != test.wantPlaintext {
				t.Fatalf("expected runtime secret %q, got %q", test.wantPlaintext, runtime.ClientSecret)
			}
		})
	}
}

func TestRedactedGetsSurviveCorruptSecrets(t *testing.T) {
	corruptErr := errors.New("secret cannot be decrypted")
	tests := []struct {
		name       string
		resource   Resource
		secretKey  string
		get        func(context.Context, *Service) (bool, error)
		getRuntime func(context.Context, *Service) error
	}{
		{
			name:      "SMTP",
			resource:  ResourceSMTP,
			secretKey: keySMTPPassword,
			get: func(ctx context.Context, service *Service) (bool, error) {
				got, err := service.GetSMTP(ctx, GetSMTPInput{CurrentUserID: testSystemSettingsAdminID})
				return got.PasswordSet, err
			},
			getRuntime: func(ctx context.Context, service *Service) error {
				_, err := service.EffectiveSMTP(ctx)
				return err
			},
		},
		{
			name:      "OIDC",
			resource:  ResourceOIDC,
			secretKey: keyOIDCClientSecret,
			get: func(ctx context.Context, service *Service) (bool, error) {
				got, err := service.GetOIDC(ctx, GetOIDCInput{CurrentUserID: testSystemSettingsAdminID})
				return got.ClientSecretSet, err
			},
			getRuntime: func(ctx context.Context, service *Service) error {
				_, err := service.EffectiveOIDC(ctx)
				return err
			},
		},
		{
			name:      "Google",
			resource:  ResourceGoogle,
			secretKey: keyGoogleClientSecret,
			get: func(ctx context.Context, service *Service) (bool, error) {
				got, err := service.GetGoogle(ctx, GetGoogleInput{CurrentUserID: testSystemSettingsAdminID})
				return got.ClientSecretSet, err
			},
			getRuntime: func(ctx context.Context, service *Service) error {
				_, err := service.EffectiveGoogle(ctx)
				return err
			},
		},
		{
			name:      "GitHub",
			resource:  ResourceGitHub,
			secretKey: keyGitHubClientSecret,
			get: func(ctx context.Context, service *Service) (bool, error) {
				got, err := service.GetGitHub(ctx, GetGitHubInput{CurrentUserID: testSystemSettingsAdminID})
				return got.ClientSecretSet, err
			},
			getRuntime: func(ctx context.Context, service *Service) error {
				_, err := service.EffectiveGitHub(ctx)
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newMemorySettingsRepository()
			repo.settings[test.secretKey] = domainsystem.Setting{
				Key:                 test.secretKey,
				Secret:              true,
				EncryptedValue:      []byte("corrupt"),
				EncryptedValueNonce: []byte("nonce"),
			}
			switch test.resource {
			case ResourceOIDC:
				repo.settings[keyOIDCSettings] = testPublicSetting(t, keyOIDCSettings, storedOIDCSettings{
					Enabled:     true,
					IssuerURL:   "https://issuer.example.com",
					ClientID:    "oidc-client",
					DisplayName: "Single sign-on",
				})
			case ResourceGoogle:
				repo.settings[keyGoogleSettings] = testPublicSetting(t, keyGoogleSettings, storedGoogleSettings{
					Enabled:     true,
					ClientID:    "google-client",
					DisplayName: "Google",
				})
			case ResourceGitHub:
				repo.settings[keyGitHubSettings] = testPublicSetting(t, keyGitHubSettings, storedGitHubSettings{
					Enabled:     true,
					ClientID:    "github-client",
					DisplayName: "GitHub",
				})
			}
			cipher := &fakeSecretCipher{decryptErrors: map[string]error{"corrupt": corruptErr}}
			service := newTestSettingsService(repo, cipher, nil, nil)

			secretSet, err := test.get(context.Background(), service)
			if err != nil {
				t.Fatalf("get redacted settings: %v", err)
			}
			if !secretSet {
				t.Fatal("expected redacted settings to report that the secret is set")
			}
			if len(repo.lockAttempts) != 0 {
				t.Fatalf("redacted GET acquired write locks: %#v", repo.lockAttempts)
			}
			if len(cipher.decryptInputs) != 0 {
				t.Fatalf("redacted GET attempted to decrypt the secret: %#v", cipher.decryptInputs)
			}

			if err := test.getRuntime(context.Background(), service); !errors.Is(err, corruptErr) {
				t.Fatalf("expected runtime read to fail closed on corrupt secret, got %v", err)
			}
			if len(cipher.decryptInputs) != 1 {
				t.Fatalf("expected one runtime decryption attempt, got %#v", cipher.decryptInputs)
			}
		})
	}
}

func TestEffectiveProviderReadsAreFaultIsolatedAndLive(t *testing.T) {
	corruptErr := errors.New("OIDC secret cannot be decrypted")
	repo := newMemorySettingsRepository()
	repo.settings[keyOIDCClientSecret] = domainsystem.Setting{
		Key:                 keyOIDCClientSecret,
		Secret:              true,
		EncryptedValue:      []byte("corrupt-oidc"),
		EncryptedValueNonce: []byte("nonce"),
	}
	repo.settings[keyOIDCSettings] = testPublicSetting(t, keyOIDCSettings, storedOIDCSettings{
		Enabled:     true,
		IssuerURL:   "https://issuer.example.com",
		ClientID:    "oidc-client",
		DisplayName: "Single sign-on",
	})
	repo.settings[keyGoogleSettings] = testPublicSetting(t, keyGoogleSettings, storedGoogleSettings{
		Enabled:        true,
		ClientID:       "google-v1",
		DisplayName:    "Google Workspace",
		JITEnabled:     true,
		AllowedDomains: []string{"example.com"},
	})
	repo.settings[keyGoogleClientSecret] = testSecretSetting(keyGoogleClientSecret, "google-v1")
	cipher := &fakeSecretCipher{decryptErrors: map[string]error{"corrupt-oidc": corruptErr}}
	service := newTestSettingsService(repo, cipher, nil, nil)

	first, err := service.EffectiveGoogle(context.Background())
	if err != nil {
		t.Fatalf("read Google runtime settings while OIDC is corrupt: %v", err)
	}
	wantFirst := GoogleRuntimeSettings{
		Enabled:        true,
		ClientID:       "google-v1",
		ClientSecret:   "google-v1",
		DisplayName:    "Google Workspace",
		JITEnabled:     true,
		AllowedDomains: []string{"example.com"},
	}
	if !reflect.DeepEqual(first, wantFirst) {
		t.Fatalf("unexpected Google runtime settings: got %#v want %#v", first, wantFirst)
	}
	for _, query := range repo.readKeyQueries {
		if containsString(query, keyOIDCSettings) ||
			containsString(query, keyOIDCClientSecret) ||
			containsString(query, keyGitHubSettings) ||
			containsString(query, keyGitHubClientSecret) {
			t.Fatalf("Google runtime read touched another provider: %#v", repo.readKeyQueries)
		}
	}
	if !reflect.DeepEqual(cipher.decryptInputs, []string{"enc:google-v1"}) {
		t.Fatalf("unexpected decryption inputs: %#v", cipher.decryptInputs)
	}

	repo.settings[keyGoogleSettings] = testPublicSetting(t, keyGoogleSettings, storedGoogleSettings{
		Enabled:     true,
		ClientID:    "google-v2",
		DisplayName: "Google",
	})
	repo.settings[keyGoogleClientSecret] = testSecretSetting(keyGoogleClientSecret, "google-v2")
	second, err := service.EffectiveGoogle(context.Background())
	if err != nil {
		t.Fatalf("read updated Google runtime settings: %v", err)
	}
	if second.ClientID != "google-v2" || second.ClientSecret != "google-v2" {
		t.Fatalf("runtime settings were stale after repository update: %#v", second)
	}

	if _, err := service.EffectiveOIDC(context.Background()); !errors.Is(err, corruptErr) {
		t.Fatalf("expected only OIDC runtime read to fail, got %v", err)
	}
}

func TestUpdateGitHubPreservesStoredAllowSignupFalse(t *testing.T) {
	repo := newMemorySettingsRepository()
	repo.settings[keyGitHubSettings] = testPublicSetting(t, keyGitHubSettings, storedGitHubSettings{
		ClientID:    "github-client",
		DisplayName: "GitHub",
		AllowSignup: false,
	})
	service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, nil)
	displayName := "Engineering GitHub"

	got, err := service.UpdateGitHub(context.Background(), UpdateGitHubInput{
		CurrentUserID: testSystemSettingsAdminID,
		DisplayName:   &displayName,
	})
	if err != nil {
		t.Fatalf("update GitHub settings: %v", err)
	}
	if got.AllowSignup {
		t.Fatalf("unexpected GitHub update result: %#v", got)
	}
	if !reflect.DeepEqual(repo.lockAttempts, []string{string(ResourceGitHub)}) {
		t.Fatalf("unexpected provider locks: %#v", repo.lockAttempts)
	}

	stored := repo.settings[keyGitHubSettings]
	if !bytes.Contains(stored.Value, []byte(`"allowSignup":false`)) {
		t.Fatalf("expected explicit allowSignup=false in stored JSON, got %s", stored.Value)
	}
	var decoded storedGitHubSettings
	if decodeErr := decodeTestSetting(stored, &decoded); decodeErr != nil {
		t.Fatalf("decode stored GitHub settings: %v", decodeErr)
	}
	if decoded.AllowSignup {
		t.Fatalf("stored GitHub setting changed allowSignup to true: %#v", decoded)
	}

	runtime, err := service.EffectiveGitHub(context.Background())
	if err != nil {
		t.Fatalf("read GitHub runtime settings: %v", err)
	}
	if runtime.AllowSignup {
		t.Fatalf("runtime GitHub setting changed allowSignup to true: %#v", runtime)
	}
}

func TestUpdateOIDCHoldsResourceLockThroughReadinessAndPersistence(t *testing.T) {
	repo := newMemorySettingsRepository()
	transactor := &rollbackSettingsTransactor{repo: repo}
	checkedInTransaction := false
	var locksAtReadiness []string
	var writesAtReadiness []string
	readiness := &fakeOIDCReadinessChecker{
		onCheck: func(context.Context, string) {
			checkedInTransaction = transactor.active
			locksAtReadiness = append([]string(nil), repo.lockAttempts...)
			writesAtReadiness = append([]string(nil), repo.upsertAttempts...)
		},
	}
	service := newTestSettingsService(repo, &fakeSecretCipher{}, readiness, transactor)
	enabled := true
	issuerURL := "https://issuer.example.com"
	clientID := "oidc-client"
	clientSecret := "oidc-secret"

	got, err := service.UpdateOIDC(context.Background(), UpdateOIDCInput{
		CurrentUserID: testSystemSettingsAdminID,
		Enabled:       &enabled,
		IssuerURL:     &issuerURL,
		ClientID:      &clientID,
		ClientSecret:  OptionalSecret{Present: true, Value: &clientSecret},
	})
	if err != nil {
		t.Fatalf("update OIDC settings: %v", err)
	}
	if !got.Enabled || got.IssuerURL != issuerURL || got.ClientID != clientID || !got.ClientSecretSet {
		t.Fatalf("unexpected OIDC update result: %#v", got)
	}
	if !checkedInTransaction {
		t.Fatal("OIDC readiness ran outside the update transaction")
	}
	wantLocks := []string{string(ResourceOIDC)}
	if !reflect.DeepEqual(locksAtReadiness, wantLocks) || !reflect.DeepEqual(repo.lockAttempts, wantLocks) {
		t.Fatalf("OIDC readiness did not run under its resource lock: at readiness=%#v final=%#v", locksAtReadiness, repo.lockAttempts)
	}
	if len(writesAtReadiness) != 0 {
		t.Fatalf("OIDC settings were persisted before readiness succeeded: %#v", writesAtReadiness)
	}
	if transactor.calls != 1 || transactor.commits != 1 || transactor.rollbacks != 0 {
		t.Fatalf("unexpected OIDC transaction outcome: %#v", transactor)
	}
	if !containsString(repo.upsertAttempts, keyOIDCSettings) || !containsString(repo.upsertAttempts, keyOIDCClientSecret) {
		t.Fatalf("OIDC settings were not persisted after readiness: %#v", repo.upsertAttempts)
	}
}

func TestAccessAndSMTPUpdatesEnforceEmailVerificationInvariant(t *testing.T) {
	t.Run("enabling verification requires configured SMTP", func(t *testing.T) {
		repo := newMemorySettingsRepository()
		service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, nil)
		required := true

		_, err := service.UpdateAccess(context.Background(), UpdateAccessInput{
			CurrentUserID:             testSystemSettingsAdminID,
			EmailVerificationRequired: &required,
		})
		assertEmailVerificationSMTPInvariant(t, err)
		assertNoSettingsWrites(t, repo)
		assertAccessSMTPLockOrder(t, repo.lockAttempts)
	})

	t.Run("clearing SMTP is rejected while verification is required", func(t *testing.T) {
		repo := newMemorySettingsRepository()
		repo.settings[keyEmailVerificationRequired] = testPublicSetting(t, keyEmailVerificationRequired, true)
		repo.settings[keySMTPHost] = testPublicSetting(t, keySMTPHost, "smtp.example.com")
		repo.settings[keySMTPFrom] = testPublicSetting(t, keySMTPFrom, "Netstamp <alerts@example.com>")
		service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, nil)
		empty := ""

		_, err := service.UpdateSMTP(context.Background(), UpdateSMTPInput{
			CurrentUserID: testSystemSettingsAdminID,
			Host:          &empty,
			From:          &empty,
		})
		assertEmailVerificationSMTPInvariant(t, err)
		assertNoSettingsWrites(t, repo)
		assertAccessSMTPLockOrder(t, repo.lockAttempts)
	})
}

func TestValidateOIDCClassifiesReadinessFailures(t *testing.T) {
	semanticErr := errors.New("discovery issuer mismatch")
	networkErr := &net.DNSError{Err: "temporary name resolution failure", IsTemporary: true}
	tests := []struct {
		name             string
		readiness        OIDCReadinessChecker
		cause            error
		want             error
		wantFieldError   bool
		wantCauseWrapped bool
	}{
		{
			name:           "invalid discovery metadata is validation failure",
			readiness:      &fakeOIDCReadinessChecker{err: &fakeOIDCMetadataError{cause: semanticErr}},
			want:           ErrInvalidInput,
			wantFieldError: true,
		},
		{
			name:             "deadline is provider outage",
			readiness:        &fakeOIDCReadinessChecker{err: context.DeadlineExceeded},
			cause:            context.DeadlineExceeded,
			want:             ErrProviderUnavailable,
			wantCauseWrapped: true,
		},
		{
			name:             "network failure is provider outage",
			readiness:        &fakeOIDCReadinessChecker{err: networkErr},
			cause:            networkErr,
			want:             ErrProviderUnavailable,
			wantCauseWrapped: true,
		},
		{
			name: "missing checker is provider outage",
			want: ErrProviderUnavailable,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newMemorySettingsRepository()
			service := newTestSettingsService(repo, &fakeSecretCipher{}, test.readiness, nil)
			enabled := true
			issuerURL := "https://issuer.example.com"
			clientID := "oidc-client"
			clientSecret := "oidc-secret"

			err := service.ValidateOIDC(context.Background(), ValidateOIDCInput{
				CurrentUserID: testSystemSettingsAdminID,
				Enabled:       &enabled,
				IssuerURL:     &issuerURL,
				ClientID:      &clientID,
				ClientSecret:  OptionalSecret{Present: true, Value: &clientSecret},
			})
			if !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
			if errors.Is(test.want, ErrInvalidInput) && errors.Is(err, ErrProviderUnavailable) {
				t.Fatalf("semantic readiness error was classified as provider unavailable: %v", err)
			}
			if errors.Is(test.want, ErrProviderUnavailable) && errors.Is(err, ErrInvalidInput) {
				t.Fatalf("provider outage was classified as invalid input: %v", err)
			}
			if test.wantCauseWrapped && !errors.Is(err, test.cause) {
				t.Fatalf("expected readiness cause %v to be preserved, got %v", test.cause, err)
			}
			if test.wantFieldError {
				fields, ok := appvalidation.FieldErrors(err)
				if !ok || len(fields) != 1 {
					t.Fatalf("expected one issuerUrl validation error, got %#v", fields)
				}
				if fields[0].Field != "issuerUrl" ||
					fields[0].Message != "issuer discovery metadata is invalid or does not match issuerUrl" ||
					fields[0].Value != issuerURL {
					t.Fatalf("unexpected issuerUrl validation error: %#v", fields[0])
				}
			}
			if len(repo.lockAttempts) != 0 {
				t.Fatalf("provider validation acquired write locks: %#v", repo.lockAttempts)
			}
			assertNoSettingsWrites(t, repo)
		})
	}
}

func newTestSettingsService(
	repo Repository,
	cipher SecretCipher,
	readiness OIDCReadinessChecker,
	transactor Transactor,
) *Service {
	const callbackBaseURL = "https://netstamp.example.com/api/v1/auth/external"
	if transactor == nil {
		transactor = passthroughSettingsTransactor{}
	}
	return NewService(
		repo,
		fakeSystemAdminChecker{},
		cipher,
		readiness,
		Defaults{},
		callbackBaseURL,
		transactor,
	)
}

func assertEmailVerificationSMTPInvariant(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
	fields, ok := appvalidation.FieldErrors(err)
	if !ok || len(fields) != 1 {
		t.Fatalf("expected one validation field, got %#v", fields)
	}
	if fields[0].Field != "emailVerificationRequired" ||
		fields[0].Message != "email verification requires configured SMTP host and from address" ||
		fields[0].Value != true {
		t.Fatalf("unexpected invariant validation: %#v", fields[0])
	}
}

func assertNoSettingsWrites(t *testing.T, repo *memorySettingsRepository) {
	t.Helper()
	if len(repo.upsertAttempts) != 0 ||
		len(repo.deleteAttempts) != 0 ||
		repo.auditAttempts != 0 {
		t.Fatalf(
			"expected no writes, got upserts=%#v deletes=%#v audits=%d",
			repo.upsertAttempts,
			repo.deleteAttempts,
			repo.auditAttempts,
		)
	}
}

func assertAccessSMTPLockOrder(t *testing.T, got []string) {
	t.Helper()
	want := []string{string(ResourceAccess), string(ResourceSMTP)}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected access/SMTP lock order %#v, got %#v", want, got)
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func gotAuditAction(events []settingsAuditEvent, key string) string {
	for _, event := range events {
		if event.key == key {
			return event.action
		}
	}
	return ""
}
