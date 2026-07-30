package systemsettings

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	domainsystem "github.com/yorukot/netstamp/internal/domain/system"
)

func TestEffectiveProvidersFailClosedWhenEnabled(t *testing.T) {
	corruptErr := errors.New("stored provider secret is corrupt")
	modes := []struct {
		name              string
		configure         func(*testing.T, runtimeProviderHarness, *memorySettingsRepository, *fakeSecretCipher)
		wantValidation    string
		wantError         error
		wantErrorContains string
		wantDecryptCalls  int
	}{
		{
			name: "missing secret",
			configure: func(t *testing.T, provider runtimeProviderHarness, repo *memorySettingsRepository, _ *fakeSecretCipher) {
				t.Helper()
				repo.settings[provider.publicKey] = testPublicSetting(t, provider.publicKey, provider.validPublic(true))
			},
			wantValidation: "clientSecret",
		},
		{
			name: "invalid public config",
			configure: func(t *testing.T, provider runtimeProviderHarness, repo *memorySettingsRepository, _ *fakeSecretCipher) {
				t.Helper()
				repo.settings[provider.publicKey] = testPublicSetting(t, provider.publicKey, provider.invalidPublic())
				repo.settings[provider.secretKey] = testSecretSetting(provider.secretKey, "valid-secret")
			},
			wantValidation: "clientId",
		},
		{
			name: "corrupt encrypted secret",
			configure: func(t *testing.T, provider runtimeProviderHarness, repo *memorySettingsRepository, cipher *fakeSecretCipher) {
				t.Helper()
				repo.settings[provider.publicKey] = testPublicSetting(t, provider.publicKey, provider.validPublic(true))
				repo.settings[provider.secretKey] = domainsystem.Setting{
					Key:                 provider.secretKey,
					Secret:              true,
					EncryptedValue:      []byte("corrupt"),
					EncryptedValueNonce: []byte("nonce"),
				}
				cipher.decryptErrors = map[string]error{"corrupt": corruptErr}
			},
			wantError:        corruptErr,
			wantDecryptCalls: 1,
		},
		{
			name: "empty decrypted secret",
			configure: func(t *testing.T, provider runtimeProviderHarness, repo *memorySettingsRepository, _ *fakeSecretCipher) {
				t.Helper()
				repo.settings[provider.publicKey] = testPublicSetting(t, provider.publicKey, provider.validPublic(true))
				repo.settings[provider.secretKey] = testSecretSetting(provider.secretKey, "")
			},
			wantErrorContains: "secret value is empty",
			wantDecryptCalls:  1,
		},
	}

	for _, provider := range runtimeProviderHarnesses() {
		for _, mode := range modes {
			t.Run(provider.name+"/"+mode.name, func(t *testing.T) {
				repo := newMemorySettingsRepository()
				cipher := &fakeSecretCipher{}
				mode.configure(t, provider, repo, cipher)
				service := newTestSettingsService(repo, cipher, nil, nil)

				got, err := provider.read(context.Background(), service)
				if err == nil {
					t.Fatalf("expected enabled %s provider to fail closed, got %#v", provider.name, got)
				}
				if !reflect.DeepEqual(got, provider.zeroRuntime) {
					t.Fatalf("expected zero runtime settings, got %#v", got)
				}
				if mode.wantValidation != "" {
					if !errors.Is(err, ErrInvalidInput) {
						t.Fatalf("expected invalid input, got %v", err)
					}
					assertSystemSettingsValidationField(t, err, mode.wantValidation)
				}
				if mode.wantError != nil && !errors.Is(err, mode.wantError) {
					t.Fatalf("expected wrapped error %v, got %v", mode.wantError, err)
				}
				if mode.wantErrorContains != "" && !strings.Contains(err.Error(), mode.wantErrorContains) {
					t.Fatalf("expected error containing %q, got %v", mode.wantErrorContains, err)
				}
				if len(cipher.decryptInputs) != mode.wantDecryptCalls {
					t.Fatalf("expected %d decrypt calls, got %#v", mode.wantDecryptCalls, cipher.decryptInputs)
				}
			})
		}
	}
}

func TestEffectiveProvidersIgnoreStaleCorruptSecretWhenDisabled(t *testing.T) {
	corruptErr := errors.New("disabled provider secret must not be decrypted")
	for _, provider := range runtimeProviderHarnesses() {
		t.Run(provider.name, func(t *testing.T) {
			repo := newMemorySettingsRepository()
			repo.settings[provider.publicKey] = testPublicSetting(t, provider.publicKey, provider.validPublic(false))
			repo.settings[provider.secretKey] = domainsystem.Setting{
				Key:                 provider.secretKey,
				Secret:              true,
				EncryptedValue:      []byte("stale-corrupt"),
				EncryptedValueNonce: []byte("nonce"),
			}
			cipher := &fakeSecretCipher{decryptErrors: map[string]error{"stale-corrupt": corruptErr}}
			service := newTestSettingsService(repo, cipher, nil, nil)

			got, err := provider.read(context.Background(), service)
			if err != nil {
				t.Fatalf("read disabled %s provider: %v", provider.name, err)
			}
			if provider.runtimeEnabled(got) {
				t.Fatalf("expected disabled runtime settings, got %#v", got)
			}
			if len(cipher.decryptInputs) != 0 {
				t.Fatalf("disabled provider decrypted stale secret: %#v", cipher.decryptInputs)
			}
		})
	}
}

func TestEffectiveSMTPRejectsMalformedStoredConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		values    map[string]any
		password  string
		wantField string
	}{
		{
			name: "blank host",
			values: map[string]any{
				keySMTPHost: "   ",
				keySMTPFrom: "alerts@example.com",
			},
			wantField: "host",
		},
		{
			name: "host with scheme",
			values: map[string]any{
				keySMTPHost: "smtp://smtp.example.com",
				keySMTPFrom: "alerts@example.com",
			},
			wantField: "host",
		},
		{
			name: "host with port",
			values: map[string]any{
				keySMTPHost: "smtp.example.com:587",
				keySMTPFrom: "alerts@example.com",
			},
			wantField: "host",
		},
		{
			name: "invalid from address",
			values: map[string]any{
				keySMTPHost: "smtp.example.com",
				keySMTPFrom: "not-an-email",
			},
			wantField: "from",
		},
		{
			name: "username without password",
			values: map[string]any{
				keySMTPHost:     "smtp.example.com",
				keySMTPFrom:     "alerts@example.com",
				keySMTPUsername: "mailer",
			},
			wantField: "username",
		},
		{
			name: "password with blank username",
			values: map[string]any{
				keySMTPHost:     "smtp.example.com",
				keySMTPFrom:     "alerts@example.com",
				keySMTPUsername: "   ",
			},
			password:  "smtp-secret",
			wantField: "username",
		},
		{
			name: "invalid TLS mode",
			values: map[string]any{
				keySMTPTLSMode: "opportunistic",
			},
			wantField: "tlsMode",
		},
		{
			name: "authenticated SMTP without TLS",
			values: map[string]any{
				keySMTPHost:     "smtp.example.com",
				keySMTPFrom:     "alerts@example.com",
				keySMTPUsername: "mailer",
				keySMTPTLSMode:  "none",
			},
			password:  "smtp-secret",
			wantField: "tlsMode",
		},
		{
			name: "invalid port",
			values: map[string]any{
				keySMTPPort: int32(0),
			},
			wantField: "port",
		},
		{
			name: "invalid timeout",
			values: map[string]any{
				keySMTPTimeoutSeconds: int32(0),
			},
			wantField: "timeoutSeconds",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newMemorySettingsRepository()
			for key, value := range test.values {
				repo.settings[key] = testPublicSetting(t, key, value)
			}
			if test.password != "" {
				repo.settings[keySMTPPassword] = testSecretSetting(keySMTPPassword, test.password)
			}
			service := newTestSettingsService(repo, &fakeSecretCipher{}, nil, nil)

			got, err := service.EffectiveSMTP(context.Background())
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected malformed stored SMTP settings to fail closed, got %#v, %v", got, err)
			}
			if got != (SMTPRuntimeSettings{}) {
				t.Fatalf("expected zero SMTP runtime settings, got %#v", got)
			}
			assertSystemSettingsValidationField(t, err, test.wantField)
		})
	}
}

func TestSettingsUpdatesRejectCorruptRetainedSMTPPassword(t *testing.T) {
	corruptErr := errors.New("retained SMTP password cannot be decrypted")
	tests := []struct {
		name   string
		update func(context.Context, *Service) (any, error)
		zero   any
	}{
		{
			name: "SMTP public field update",
			update: func(ctx context.Context, service *Service) (any, error) {
				timeoutSeconds := int32(20)
				return service.UpdateSMTP(ctx, UpdateSMTPInput{
					CurrentUserID:    testSystemSettingsAdminID,
					ExpectedRevision: 6,
					TimeoutSeconds:   &timeoutSeconds,
				})
			},
			zero: Versioned[SMTPSettings]{},
		},
		{
			name: "enable email verification",
			update: func(ctx context.Context, service *Service) (any, error) {
				required := true
				return service.UpdateAccess(ctx, UpdateAccessInput{
					CurrentUserID:             testSystemSettingsAdminID,
					ExpectedRevision:          5,
					EmailVerificationRequired: &required,
				})
			},
			zero: Versioned[AccessSettings]{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newMemorySettingsRepository()
			repo.revisions[string(ResourceAccess)] = 5
			repo.revisions[string(ResourceSMTP)] = 6
			repo.settings[keySMTPHost] = testPublicSetting(t, keySMTPHost, "smtp.example.com")
			repo.settings[keySMTPUsername] = testPublicSetting(t, keySMTPUsername, "mailer")
			repo.settings[keySMTPFrom] = testPublicSetting(t, keySMTPFrom, "alerts@example.com")
			repo.settings[keySMTPPassword] = domainsystem.Setting{
				Key:                 keySMTPPassword,
				Secret:              true,
				EncryptedValue:      []byte("corrupt-smtp-password"),
				EncryptedValueNonce: []byte("nonce"),
			}
			cipher := &fakeSecretCipher{
				decryptErrors: map[string]error{"corrupt-smtp-password": corruptErr},
			}
			service := newTestSettingsService(repo, cipher, nil, nil)

			got, err := test.update(context.Background(), service)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected corrupt retained password to fail validation, got %#v, %v", got, err)
			}
			if !reflect.DeepEqual(got, test.zero) {
				t.Fatalf("expected zero update result, got %#v", got)
			}
			field := assertSystemSettingsValidationField(t, err, "password")
			if field.Message != "password must be replaced before using SMTP authentication" || field.Value != nil {
				t.Fatalf("unexpected retained-password validation: %#v", field)
			}
			if len(cipher.decryptInputs) != 1 {
				t.Fatalf("expected one retained-password check, got %#v", cipher.decryptInputs)
			}
			if repo.revisions[string(ResourceAccess)] != 5 || repo.revisions[string(ResourceSMTP)] != 6 {
				t.Fatalf("validation failure changed revisions: %#v", repo.revisions)
			}
			assertNoSettingsWrites(t, repo)
		})
	}
}

func TestStoredGoogleSettingsDecodesLegacyAllowedDomains(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		wantDomains []string
		wantErr     bool
	}{
		{
			name: "legacy comma separated string",
			payload: `{
				"Enabled": true,
				"ClientID": "google-client",
				"DisplayName": "Google Workspace",
				"JITEnabled": true,
				"AllowedDomains": " Example.COM,sub.example.com,example.com "
			}`,
			wantDomains: []string{"example.com", "sub.example.com"},
		},
		{
			name: "string array",
			payload: `{
				"enabled": true,
				"clientId": "google-client",
				"displayName": "Google Workspace",
				"jitEnabled": true,
				"allowedDomains": [" Example.COM ", "sub.example.com", "example.com"]
			}`,
			wantDomains: []string{"example.com", "sub.example.com"},
		},
		{
			name: "legacy string rejects empty segment",
			payload: `{
				"AllowedDomains": "example.com,,sub.example.com"
			}`,
			wantErr: true,
		},
		{
			name: "array rejects empty segment",
			payload: `{
				"allowedDomains": ["example.com", "", "sub.example.com"]
			}`,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got storedGoogleSettings
			err := json.Unmarshal([]byte(test.payload), &got)
			if test.wantErr {
				if err == nil || !strings.Contains(err.Error(), "must not contain empty domains") {
					t.Fatalf("expected empty-domain error, got %#v, %v", got, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("decode stored Google settings: %v", err)
			}
			if !reflect.DeepEqual(got.AllowedDomains, test.wantDomains) {
				t.Fatalf("unexpected normalized domains: got %#v want %#v", got.AllowedDomains, test.wantDomains)
			}
			if !got.Enabled ||
				got.ClientID != "google-client" ||
				got.DisplayName != "Google Workspace" ||
				!got.JITEnabled {
				t.Fatalf("Google settings fields were not preserved: %#v", got)
			}
		})
	}
}

func TestStoredGoogleSettingsRejectsExplicitNullAllowedDomains(t *testing.T) {
	var got storedGoogleSettings
	err := json.Unmarshal([]byte(`{"enabled":true,"allowedDomains":null}`), &got)
	if err == nil || !strings.Contains(err.Error(), "allowedDomains must not be null") {
		t.Fatalf("expected explicit null allowlist to be rejected, got %#v, %v", got, err)
	}
}

func TestEnabledProviderUpdateRejectsOmittedCorruptRetainedSecret(t *testing.T) {
	corruptErr := errors.New("retained secret cannot be decrypted")
	for _, provider := range runtimeProviderHarnesses() {
		t.Run(provider.name, func(t *testing.T) {
			repo := newMemorySettingsRepository()
			repo.revisions[string(provider.resource)] = 5
			repo.settings[provider.publicKey] = testPublicSetting(t, provider.publicKey, provider.validPublic(true))
			repo.settings[provider.secretKey] = domainsystem.Setting{
				Key:                 provider.secretKey,
				Secret:              true,
				EncryptedValue:      []byte("corrupt-retained"),
				EncryptedValueNonce: []byte("nonce"),
			}
			cipher := &fakeSecretCipher{decryptErrors: map[string]error{"corrupt-retained": corruptErr}}
			service := newTestSettingsService(repo, cipher, nil, nil)

			got, err := provider.updateDisplayName(context.Background(), service, 5)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected corrupt retained secret to require replacement, got %#v, %v", got, err)
			}
			if !reflect.DeepEqual(got, provider.zeroVersioned) {
				t.Fatalf("expected zero update result, got %#v", got)
			}
			fields := assertSystemSettingsValidationField(t, err, "clientSecret")
			if fields.Message != "clientSecret must be replaced before enabling this provider" || fields.Value != nil {
				t.Fatalf("unexpected retained-secret validation: %#v", fields)
			}
			if len(cipher.decryptInputs) != 1 {
				t.Fatalf("expected one retained-secret check, got %#v", cipher.decryptInputs)
			}
			assertNoSettingsWrites(t, repo)
		})
	}
}

func TestEffectiveGoogleReadsPublicAndSecretInOneQuery(t *testing.T) {
	repo := newMemorySettingsRepository()
	repo.settings[keyGoogleSettings] = testPublicSetting(t, keyGoogleSettings, storedGoogleSettings{
		Enabled:        true,
		ClientID:       "google-client",
		DisplayName:    "Google Workspace",
		JITEnabled:     true,
		AllowedDomains: []string{"example.com"},
	})
	repo.settings[keyGoogleClientSecret] = testSecretSetting(keyGoogleClientSecret, "google-secret")
	cipher := &fakeSecretCipher{}
	service := newTestSettingsService(repo, cipher, nil, nil)

	got, err := service.EffectiveGoogle(context.Background())
	if err != nil {
		t.Fatalf("read Google runtime settings: %v", err)
	}
	if !got.Enabled ||
		got.ClientID != "google-client" ||
		got.ClientSecret != "google-secret" ||
		got.DisplayName != "Google Workspace" {
		t.Fatalf("unexpected Google runtime settings: %#v", got)
	}
	if len(repo.readKeyQueries) != 1 || !reflect.DeepEqual(repo.readKeyQueries[0], googleKeys) {
		t.Fatalf("expected one public+secret query %#v, got %#v", googleKeys, repo.readKeyQueries)
	}
	if !reflect.DeepEqual(cipher.decryptInputs, []string{"enc:google-secret"}) {
		t.Fatalf("unexpected secret decrypt calls: %#v", cipher.decryptInputs)
	}
}

type runtimeProviderHarness struct {
	name              string
	resource          Resource
	publicKey         string
	secretKey         string
	validPublic       func(bool) any
	invalidPublic     func() any
	read              func(context.Context, *Service) (any, error)
	runtimeEnabled    func(any) bool
	updateDisplayName func(context.Context, *Service, int64) (any, error)
	zeroRuntime       any
	zeroVersioned     any
}

func runtimeProviderHarnesses() []runtimeProviderHarness {
	return []runtimeProviderHarness{
		{
			name:      "OIDC",
			resource:  ResourceOIDC,
			publicKey: keyOIDCSettings,
			secretKey: keyOIDCClientSecret,
			validPublic: func(enabled bool) any {
				return storedOIDCSettings{
					Enabled:     enabled,
					IssuerURL:   "https://issuer.example.com",
					ClientID:    "oidc-client",
					DisplayName: "Single sign-on",
				}
			},
			invalidPublic: func() any {
				return storedOIDCSettings{
					Enabled:     true,
					IssuerURL:   "https://issuer.example.com",
					DisplayName: "Single sign-on",
				}
			},
			read: func(ctx context.Context, service *Service) (any, error) {
				return service.EffectiveOIDC(ctx)
			},
			runtimeEnabled: func(value any) bool {
				return value.(OIDCRuntimeSettings).Enabled
			},
			updateDisplayName: func(ctx context.Context, service *Service, revision int64) (any, error) {
				displayName := "Updated single sign-on"
				return service.UpdateOIDC(ctx, UpdateOIDCInput{
					CurrentUserID:    testSystemSettingsAdminID,
					ExpectedRevision: revision,
					DisplayName:      &displayName,
				})
			},
			zeroRuntime:   OIDCRuntimeSettings{},
			zeroVersioned: Versioned[OIDCSettings]{},
		},
		{
			name:      "Google",
			resource:  ResourceGoogle,
			publicKey: keyGoogleSettings,
			secretKey: keyGoogleClientSecret,
			validPublic: func(enabled bool) any {
				return storedGoogleSettings{
					Enabled:        enabled,
					ClientID:       "google-client",
					DisplayName:    "Google",
					AllowedDomains: []string{"example.com"},
				}
			},
			invalidPublic: func() any {
				return storedGoogleSettings{
					Enabled:        true,
					DisplayName:    "Google",
					AllowedDomains: []string{"example.com"},
				}
			},
			read: func(ctx context.Context, service *Service) (any, error) {
				return service.EffectiveGoogle(ctx)
			},
			runtimeEnabled: func(value any) bool {
				return value.(GoogleRuntimeSettings).Enabled
			},
			updateDisplayName: func(ctx context.Context, service *Service, revision int64) (any, error) {
				displayName := "Updated Google"
				return service.UpdateGoogle(ctx, UpdateGoogleInput{
					CurrentUserID:    testSystemSettingsAdminID,
					ExpectedRevision: revision,
					DisplayName:      &displayName,
				})
			},
			zeroRuntime:   GoogleRuntimeSettings{},
			zeroVersioned: Versioned[GoogleSettings]{},
		},
		{
			name:      "GitHub",
			resource:  ResourceGitHub,
			publicKey: keyGitHubSettings,
			secretKey: keyGitHubClientSecret,
			validPublic: func(enabled bool) any {
				return storedGitHubSettings{
					Enabled:     enabled,
					ClientID:    "github-client",
					DisplayName: "GitHub",
					AllowSignup: false,
				}
			},
			invalidPublic: func() any {
				return storedGitHubSettings{
					Enabled:     true,
					DisplayName: "GitHub",
					AllowSignup: false,
				}
			},
			read: func(ctx context.Context, service *Service) (any, error) {
				return service.EffectiveGitHub(ctx)
			},
			runtimeEnabled: func(value any) bool {
				return value.(GitHubRuntimeSettings).Enabled
			},
			updateDisplayName: func(ctx context.Context, service *Service, revision int64) (any, error) {
				displayName := "Updated GitHub"
				return service.UpdateGitHub(ctx, UpdateGitHubInput{
					CurrentUserID:    testSystemSettingsAdminID,
					ExpectedRevision: revision,
					DisplayName:      &displayName,
				})
			},
			zeroRuntime:   GitHubRuntimeSettings{},
			zeroVersioned: Versioned[GitHubSettings]{},
		},
	}
}

func assertSystemSettingsValidationField(
	t *testing.T,
	err error,
	field string,
) appvalidation.FieldError {
	t.Helper()
	fields, ok := appvalidation.FieldErrors(err)
	if !ok {
		t.Fatalf("expected field validation error, got %v", err)
	}
	for _, candidate := range fields {
		if candidate.Field == field {
			return candidate
		}
	}
	t.Fatalf("expected validation field %q, got %#v", field, fields)
	return appvalidation.FieldError{}
}
