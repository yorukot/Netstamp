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

func TestUpdateAccessRollsBackWritesWhenRevisionBumpFails(t *testing.T) {
	bumpErr := errors.New("revision bump failed")
	repo := newMemorySettingsRepository()
	repo.revisions[string(ResourceAccess)] = 4
	repo.revisions[string(ResourceSMTP)] = 9
	repo.bumpErrors[string(ResourceAccess)] = bumpErr
	transactor := &rollbackSettingsTransactor{repo: repo}
	service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, transactor)
	accountCreationEnabled := false
	projectCreationEnabled := false

	got, err := service.UpdateAccess(context.Background(), UpdateAccessInput{
		CurrentUserID:          testSystemSettingsAdminID,
		ExpectedRevision:       4,
		AccountCreationEnabled: &accountCreationEnabled,
		ProjectCreationEnabled: &projectCreationEnabled,
	})
	if !errors.Is(err, bumpErr) {
		t.Fatalf("expected revision bump error, got %v", err)
	}
	if got != (Versioned[AccessSettings]{}) {
		t.Fatalf("expected zero result after rollback, got %#v", got)
	}
	if len(repo.upsertAttempts) != 2 || repo.auditAttempts != 2 || len(repo.bumpAttempts) != 1 {
		t.Fatalf(
			"expected two writes and audits before the late failure, got upserts=%#v audits=%d bumps=%#v",
			repo.upsertAttempts,
			repo.auditAttempts,
			repo.bumpAttempts,
		)
	}
	if len(repo.settings) != 0 || len(repo.auditEvents) != 0 {
		t.Fatalf("expected settings and audits to roll back, got settings=%#v audits=%#v", repo.settings, repo.auditEvents)
	}
	if repo.revisions[string(ResourceAccess)] != 4 || repo.revisions[string(ResourceSMTP)] != 9 {
		t.Fatalf("expected revisions to roll back, got %#v", repo.revisions)
	}
	if transactor.calls != 1 || transactor.rollbacks != 1 || transactor.commits != 0 {
		t.Fatalf("unexpected transaction outcome: %#v", transactor)
	}
}

func TestUpdateAccessEnforcesCASAndKeepsRevisionForNoop(t *testing.T) {
	repo := newMemorySettingsRepository()
	repo.revisions[string(ResourceAccess)] = 4
	repo.revisions[string(ResourceSMTP)] = 9
	transactor := &rollbackSettingsTransactor{repo: repo}
	service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, transactor)
	projectCreationEnabled := false

	updated, err := service.UpdateAccess(context.Background(), UpdateAccessInput{
		CurrentUserID:          testSystemSettingsAdminID,
		ExpectedRevision:       4,
		ProjectCreationEnabled: &projectCreationEnabled,
	})
	if err != nil {
		t.Fatalf("update access: %v", err)
	}
	if updated.Revision != 5 || updated.Value.ProjectCreationEnabled {
		t.Fatalf("unexpected first update: %#v", updated)
	}

	upsertsAfterUpdate := len(repo.upsertAttempts)
	auditsAfterUpdate := repo.auditAttempts
	bumpsAfterUpdate := len(repo.bumpAttempts)
	accountCreationEnabled := false
	_, err = service.UpdateAccess(context.Background(), UpdateAccessInput{
		CurrentUserID:          testSystemSettingsAdminID,
		ExpectedRevision:       4,
		AccountCreationEnabled: &accountCreationEnabled,
	})
	if !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("expected version conflict, got %v", err)
	}
	var conflict *VersionConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("expected typed version conflict, got %T", err)
	}
	if conflict.Resource != ResourceAccess || conflict.Expected != 4 || conflict.Current != 5 {
		t.Fatalf("unexpected conflict details: %#v", conflict)
	}
	if len(repo.upsertAttempts) != upsertsAfterUpdate ||
		repo.auditAttempts != auditsAfterUpdate ||
		len(repo.bumpAttempts) != bumpsAfterUpdate {
		t.Fatal("stale update mutated settings")
	}

	noChange, err := service.UpdateAccess(context.Background(), UpdateAccessInput{
		CurrentUserID:          testSystemSettingsAdminID,
		ExpectedRevision:       5,
		ProjectCreationEnabled: &projectCreationEnabled,
	})
	if err != nil {
		t.Fatalf("repeat access update: %v", err)
	}
	if noChange.Revision != 5 || noChange.Value.ProjectCreationEnabled {
		t.Fatalf("unexpected no-op result: %#v", noChange)
	}
	if len(repo.upsertAttempts) != upsertsAfterUpdate ||
		repo.auditAttempts != auditsAfterUpdate ||
		len(repo.bumpAttempts) != bumpsAfterUpdate {
		t.Fatal("no-op update bumped the revision or wrote an audit event")
	}
}

func TestUpdateGooglePreservesNullableSecretPatchStates(t *testing.T) {
	replacement := "new-secret"
	workspaceDisplayName := "Google Workspace"
	tests := []struct {
		name               string
		existingSecret     string
		providerEnabled    bool
		secretPatch        OptionalSecret
		displayName        *string
		wantSecret         bool
		wantPlaintext      string
		wantRevision       int64
		wantSecretUpsert   bool
		wantSecretDelete   bool
		wantSecretAudit    string
		wantRevisionBumped bool
	}{
		{
			name:               "set",
			providerEnabled:    true,
			secretPatch:        OptionalSecret{Present: true, Value: &replacement},
			wantSecret:         true,
			wantPlaintext:      replacement,
			wantRevision:       4,
			wantSecretUpsert:   true,
			wantSecretAudit:    auditActionUpdate,
			wantRevisionBumped: true,
		},
		{
			name:               "clear",
			existingSecret:     "old-secret",
			secretPatch:        OptionalSecret{Present: true},
			wantRevision:       4,
			wantSecretDelete:   true,
			wantSecretAudit:    auditActionClear,
			wantRevisionBumped: true,
		},
		{
			name:               "omit",
			existingSecret:     "old-secret",
			providerEnabled:    true,
			displayName:        &workspaceDisplayName,
			wantSecret:         true,
			wantPlaintext:      "old-secret",
			wantRevision:       4,
			wantRevisionBumped: true,
		},
		{
			name:         "clear already absent",
			secretPatch:  OptionalSecret{Present: true},
			wantRevision: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newMemorySettingsRepository()
			repo.revisions[string(ResourceGoogle)] = 3
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
				CurrentUserID:    testSystemSettingsAdminID,
				ExpectedRevision: 3,
				ClientSecret:     test.secretPatch,
				DisplayName:      test.displayName,
			})
			if err != nil {
				t.Fatalf("update Google settings: %v", err)
			}
			if got.Revision != test.wantRevision || got.Value.ClientSecretSet != test.wantSecret {
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
			if containsString(repo.bumpAttempts, string(ResourceGoogle)) != test.wantRevisionBumped {
				t.Fatalf("unexpected revision bump attempts: %#v", repo.bumpAttempts)
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
		get        func(context.Context, *Service) (bool, int64, error)
		getRuntime func(context.Context, *Service) error
	}{
		{
			name:      "SMTP",
			resource:  ResourceSMTP,
			secretKey: keySMTPPassword,
			get: func(ctx context.Context, service *Service) (bool, int64, error) {
				got, err := service.GetSMTP(ctx, GetSMTPInput{CurrentUserID: testSystemSettingsAdminID})
				return got.Value.PasswordSet, got.Revision, err
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
			get: func(ctx context.Context, service *Service) (bool, int64, error) {
				got, err := service.GetOIDC(ctx, GetOIDCInput{CurrentUserID: testSystemSettingsAdminID})
				return got.Value.ClientSecretSet, got.Revision, err
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
			get: func(ctx context.Context, service *Service) (bool, int64, error) {
				got, err := service.GetGoogle(ctx, GetGoogleInput{CurrentUserID: testSystemSettingsAdminID})
				return got.Value.ClientSecretSet, got.Revision, err
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
			get: func(ctx context.Context, service *Service) (bool, int64, error) {
				got, err := service.GetGitHub(ctx, GetGitHubInput{CurrentUserID: testSystemSettingsAdminID})
				return got.Value.ClientSecretSet, got.Revision, err
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
			repo.revisions[string(test.resource)] = 7
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

			secretSet, revision, err := test.get(context.Background(), service)
			if err != nil {
				t.Fatalf("get redacted settings: %v", err)
			}
			if !secretSet || revision != 7 {
				t.Fatalf("unexpected redacted result: secretSet=%t revision=%d", secretSet, revision)
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
	repo.revisions[string(ResourceGitHub)] = 7
	repo.settings[keyGitHubSettings] = testPublicSetting(t, keyGitHubSettings, storedGitHubSettings{
		ClientID:    "github-client",
		DisplayName: "GitHub",
		AllowSignup: false,
	})
	service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, nil)
	displayName := "Engineering GitHub"

	got, err := service.UpdateGitHub(context.Background(), UpdateGitHubInput{
		CurrentUserID:    testSystemSettingsAdminID,
		ExpectedRevision: 7,
		DisplayName:      &displayName,
	})
	if err != nil {
		t.Fatalf("update GitHub settings: %v", err)
	}
	if got.Revision != 8 || got.Value.AllowSignup {
		t.Fatalf("unexpected GitHub update result: %#v", got)
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

func TestAccessAndSMTPUpdatesEnforceEmailVerificationInvariant(t *testing.T) {
	t.Run("enabling verification requires configured SMTP", func(t *testing.T) {
		repo := newMemorySettingsRepository()
		repo.revisions[string(ResourceAccess)] = 1
		repo.revisions[string(ResourceSMTP)] = 2
		service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, nil)
		required := true

		_, err := service.UpdateAccess(context.Background(), UpdateAccessInput{
			CurrentUserID:             testSystemSettingsAdminID,
			ExpectedRevision:          1,
			EmailVerificationRequired: &required,
		})
		assertEmailVerificationSMTPInvariant(t, err)
		assertNoSettingsWrites(t, repo)
		assertAccessSMTPLockOrder(t, repo.lockAttempts)
	})

	t.Run("clearing SMTP is rejected while verification is required", func(t *testing.T) {
		repo := newMemorySettingsRepository()
		repo.revisions[string(ResourceAccess)] = 5
		repo.revisions[string(ResourceSMTP)] = 6
		repo.settings[keyEmailVerificationRequired] = testPublicSetting(t, keyEmailVerificationRequired, true)
		repo.settings[keySMTPHost] = testPublicSetting(t, keySMTPHost, "smtp.example.com")
		repo.settings[keySMTPFrom] = testPublicSetting(t, keySMTPFrom, "Netstamp <alerts@example.com>")
		service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, nil)
		empty := ""

		_, err := service.UpdateSMTP(context.Background(), UpdateSMTPInput{
			CurrentUserID:    testSystemSettingsAdminID,
			ExpectedRevision: 6,
			Host:             &empty,
			From:             &empty,
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
			repo.revisions[string(ResourceOIDC)] = 1
			service := newTestSettingsService(repo, &fakeSecretCipher{}, test.readiness, nil)
			enabled := true
			issuerURL := "https://issuer.example.com"
			clientID := "oidc-client"
			clientSecret := "oidc-secret"

			err := service.ValidateOIDC(context.Background(), ValidateOIDCInput{
				CurrentUserID:    testSystemSettingsAdminID,
				ExpectedRevision: 1,
				Enabled:          &enabled,
				IssuerURL:        &issuerURL,
				ClientID:         &clientID,
				ClientSecret:     OptionalSecret{Present: true, Value: &clientSecret},
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
		repo.auditAttempts != 0 ||
		len(repo.bumpAttempts) != 0 {
		t.Fatalf(
			"expected no writes, got upserts=%#v deletes=%#v audits=%d bumps=%#v",
			repo.upsertAttempts,
			repo.deleteAttempts,
			repo.auditAttempts,
			repo.bumpAttempts,
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
