package admin

import (
	"context"
	"errors"
	"slices"
	"testing"
)

func TestGetSettingsRequiresSystemAdmin(t *testing.T) {
	svc := NewService(&fakeAdminRepository{admins: map[string]bool{"user-1": false}}, fakeSecretCipher{}, Defaults{})

	_, err := svc.GetSettings(context.Background(), GetSettingsInput{CurrentUserID: "user-1"})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestUpdateSettingsStoresEncryptedPasswordAndAuditEvents(t *testing.T) {
	repo := &fakeAdminRepository{
		admins:   map[string]bool{"admin-1": true},
		settings: map[string]StoredSetting{},
	}
	svc := NewService(repo, fakeSecretCipher{}, Defaults{
		RegistrationEnabled: true,
		BackendBaseURL:      "https://api.example.com",
		PublicWebBaseURL:    "https://app.example.com",
		SMTP: SMTPSettings{
			Port:           587,
			TLSMode:        "starttls",
			TimeoutSeconds: 10,
		},
	})

	registrationEnabled := false
	backendBaseURL := "https://controller.netstamp.test"
	publicWebBaseURL := "https://console.netstamp.test"
	host := "smtp.netstamp.test"
	port := int32(465)
	username := "netstamp"
	password := "smtp-secret"
	from := "alerts@netstamp.test"
	tlsMode := "implicit"
	timeoutSeconds := int32(15)

	settings, err := svc.UpdateSettings(context.Background(), UpdateSettingsInput{
		CurrentUserID:       "admin-1",
		RegistrationEnabled: &registrationEnabled,
		BackendBaseURL:      &backendBaseURL,
		PublicWebBaseURL:    &publicWebBaseURL,
		SMTP: UpdateSMTPSettingsInput{
			Host:           &host,
			Port:           &port,
			Username:       &username,
			Password:       &password,
			From:           &from,
			TLSMode:        &tlsMode,
			TimeoutSeconds: &timeoutSeconds,
		},
	})
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}

	if settings.RegistrationEnabled {
		t.Fatal("expected registration to be disabled")
	}
	if settings.BackendBaseURL != backendBaseURL {
		t.Fatalf("expected backend base URL override, got %q", settings.BackendBaseURL)
	}
	if settings.SMTP.Password != password || !settings.SMTP.PasswordSet {
		t.Fatal("expected decrypted SMTP password in effective internal settings")
	}

	storedPassword, ok := repo.settings[keySMTPPassword]
	if !ok {
		t.Fatal("expected SMTP password setting to be stored")
	}
	if !storedPassword.Secret {
		t.Fatal("expected SMTP password to be stored as a secret")
	}
	if string(storedPassword.EncryptedValue) == password {
		t.Fatal("expected SMTP password storage to avoid plaintext")
	}
	if len(storedPassword.Value) != 0 {
		t.Fatal("expected secret setting to omit public JSON value")
	}

	for _, key := range []string{keyRegistrationEnabled, keyBackendBaseURL, keyPublicWebBaseURL, keySMTPPassword} {
		if !slices.Contains(repo.auditKeys, key) {
			t.Fatalf("expected audit event for %s, got %#v", key, repo.auditKeys)
		}
	}
}

func TestEffectiveSettingsReturnsErrorWhenSecretCannotDecrypt(t *testing.T) {
	repo := &fakeAdminRepository{
		settings: map[string]StoredSetting{
			keySMTPPassword: {
				Key:                 keySMTPPassword,
				Secret:              true,
				EncryptedValue:      []byte("invalid"),
				EncryptedValueNonce: []byte("nonce"),
			},
		},
	}
	svc := NewService(repo, fakeSecretCipher{}, Defaults{})

	_, err := svc.EffectiveSettings(context.Background())
	if err == nil {
		t.Fatal("expected decrypt error")
	}
}

func TestUpdateSettingsRequiresSMTPWhenEmailVerificationRequired(t *testing.T) {
	repo := &fakeAdminRepository{
		admins:   map[string]bool{"admin-1": true},
		settings: map[string]StoredSetting{},
	}
	svc := NewService(repo, fakeSecretCipher{}, Defaults{RegistrationEnabled: true})
	emailVerificationRequired := true

	_, err := svc.UpdateSettings(context.Background(), UpdateSettingsInput{
		CurrentUserID:             "admin-1",
		EmailVerificationRequired: &emailVerificationRequired,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}
}

type fakeAdminRepository struct {
	admins    map[string]bool
	settings  map[string]StoredSetting
	auditKeys []string
}

func (r *fakeAdminRepository) IsSystemAdmin(_ context.Context, userID string) (bool, error) {
	return r.admins[userID], nil
}

func (r *fakeAdminRepository) ListSystemSettings(context.Context) ([]StoredSetting, error) {
	settings := make([]StoredSetting, 0, len(r.settings))
	for _, setting := range r.settings {
		settings = append(settings, setting)
	}
	return settings, nil
}

func (r *fakeAdminRepository) UpsertSystemSetting(_ context.Context, setting StoredSetting) (StoredSetting, error) {
	if r.settings == nil {
		r.settings = map[string]StoredSetting{}
	}
	r.settings[setting.Key] = setting
	return setting, nil
}

func (r *fakeAdminRepository) CreateSystemSettingAuditEvent(_ context.Context, key, _ string, _ *string) error {
	r.auditKeys = append(r.auditKeys, key)
	return nil
}

type fakeSecretCipher struct{}

func (fakeSecretCipher) Encrypt(plaintext string) ([]byte, []byte, error) {
	return []byte("encrypted:" + plaintext), []byte("nonce"), nil
}

func (fakeSecretCipher) Decrypt(ciphertext, _ []byte) (string, error) {
	const prefix = "encrypted:"
	value := string(ciphertext)
	if len(value) < len(prefix) || value[:len(prefix)] != prefix {
		return "", errors.New("invalid ciphertext")
	}
	return value[len(prefix):], nil
}
