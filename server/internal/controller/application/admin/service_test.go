package admin

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"
)

const (
	testAdminUserID  = "11111111-1111-1111-1111-111111111111"
	testTargetUserID = "22222222-2222-2222-2222-222222222222"
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

func TestGrantSystemAdminStoresAuditEvent(t *testing.T) {
	grantedAt := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	repo := &fakeAdminRepository{
		admins: map[string]bool{testAdminUserID: true},
		systemAdminByEmail: map[string]SystemAdmin{
			"operator@example.com": {
				ID:          testTargetUserID,
				Email:       "operator@example.com",
				DisplayName: "Operator",
				GrantedAt:   grantedAt,
			},
		},
	}
	svc := NewService(repo, fakeSecretCipher{}, Defaults{})

	admin, err := svc.GrantSystemAdmin(context.Background(), GrantSystemAdminInput{
		CurrentUserID: testAdminUserID,
		Email:         " OPERATOR@example.com ",
	})
	if err != nil {
		t.Fatalf("grant system admin: %v", err)
	}

	if admin.ID != testTargetUserID {
		t.Fatalf("expected granted admin %q, got %q", testTargetUserID, admin.ID)
	}
	if repo.grantedEmail != "operator@example.com" {
		t.Fatalf("expected normalized grant email, got %q", repo.grantedEmail)
	}
	if !slices.Contains(repo.auditKeys, "system_admin:"+testTargetUserID) {
		t.Fatalf("expected system admin audit key, got %#v", repo.auditKeys)
	}
	if !slices.Contains(repo.auditActions, auditActionGrantSystemAdmin) {
		t.Fatalf("expected grant audit action, got %#v", repo.auditActions)
	}
}

func TestRevokeSystemAdminRejectsSelfRemoval(t *testing.T) {
	repo := &fakeAdminRepository{admins: map[string]bool{testAdminUserID: true}}
	svc := NewService(repo, fakeSecretCipher{}, Defaults{})

	err := svc.RevokeSystemAdmin(context.Background(), RevokeSystemAdminInput{
		CurrentUserID: testAdminUserID,
		UserID:        testAdminUserID,
	})
	if !errors.Is(err, ErrSelfSystemAdminRemoval) {
		t.Fatalf("expected self-removal error, got %v", err)
	}
	if repo.revokedUserID != "" {
		t.Fatalf("expected no revoke call, got %q", repo.revokedUserID)
	}
}

func TestRevokeSystemAdminRejectsLastAdmin(t *testing.T) {
	repo := &fakeAdminRepository{
		admins:       map[string]bool{testAdminUserID: true},
		revokeResult: SystemAdminRevokeResult{AdminCount: 1, TargetWasAdmin: true, Revoked: false},
	}
	svc := NewService(repo, fakeSecretCipher{}, Defaults{})

	err := svc.RevokeSystemAdmin(context.Background(), RevokeSystemAdminInput{
		CurrentUserID: testAdminUserID,
		UserID:        testTargetUserID,
	})
	if !errors.Is(err, ErrLastSystemAdmin) {
		t.Fatalf("expected last-admin error, got %v", err)
	}
}

func TestRevokeSystemAdminStoresAuditEvent(t *testing.T) {
	repo := &fakeAdminRepository{
		admins:       map[string]bool{testAdminUserID: true},
		revokeResult: SystemAdminRevokeResult{AdminCount: 2, TargetWasAdmin: true, Revoked: true},
	}
	svc := NewService(repo, fakeSecretCipher{}, Defaults{})

	err := svc.RevokeSystemAdmin(context.Background(), RevokeSystemAdminInput{
		CurrentUserID: testAdminUserID,
		UserID:        testTargetUserID,
	})
	if err != nil {
		t.Fatalf("revoke system admin: %v", err)
	}
	if repo.revokedUserID != testTargetUserID {
		t.Fatalf("expected revoked user %q, got %q", testTargetUserID, repo.revokedUserID)
	}
	if !slices.Contains(repo.auditKeys, "system_admin:"+testTargetUserID) {
		t.Fatalf("expected system admin audit key, got %#v", repo.auditKeys)
	}
	if !slices.Contains(repo.auditActions, auditActionRevokeSystemAdmin) {
		t.Fatalf("expected revoke audit action, got %#v", repo.auditActions)
	}
}

type fakeAdminRepository struct {
	admins             map[string]bool
	settings           map[string]StoredSetting
	systemAdmins       []SystemAdmin
	managedUsers       []ManagedUser
	systemAdminByEmail map[string]SystemAdmin
	systemAdminByID    map[string]ManagedUser
	revokeResult       SystemAdminRevokeResult
	activeAdminCount   int64
	grantedEmail       string
	grantedUserID      string
	revokedUserID      string
	disabledUserID     string
	disabledValue      bool
	passwordUserID     string
	passwordHash       string
	auditActions       []string
	auditKeys          []string
}

func (r *fakeAdminRepository) IsSystemAdmin(_ context.Context, userID string) (bool, error) {
	return r.admins[userID], nil
}

func (r *fakeAdminRepository) ListSystemAdmins(context.Context) ([]SystemAdmin, error) {
	return append([]SystemAdmin(nil), r.systemAdmins...), nil
}

func (r *fakeAdminRepository) ListManagedUsers(context.Context) ([]ManagedUser, error) {
	return append([]ManagedUser(nil), r.managedUsers...), nil
}

func (r *fakeAdminRepository) GrantSystemAdminByEmail(_ context.Context, email string) (SystemAdmin, error) {
	r.grantedEmail = email
	admin, ok := r.systemAdminByEmail[email]
	if !ok {
		return SystemAdmin{}, errors.New("not found")
	}
	return admin, nil
}

func (r *fakeAdminRepository) GrantSystemAdminByUserID(_ context.Context, userID string) (ManagedUser, error) {
	r.grantedUserID = userID
	user, ok := r.systemAdminByID[userID]
	if !ok {
		return ManagedUser{}, errors.New("not found")
	}
	return user, nil
}

func (r *fakeAdminRepository) RevokeSystemAdminIfNotLast(_ context.Context, userID string) (SystemAdminRevokeResult, error) {
	r.revokedUserID = userID
	return r.revokeResult, nil
}

func (r *fakeAdminRepository) CountActiveSystemAdmins(context.Context) (int64, error) {
	if r.activeAdminCount == 0 {
		return int64(len(r.admins)), nil
	}
	return r.activeAdminCount, nil
}

func (r *fakeAdminRepository) SetManagedUserDisabledAt(_ context.Context, userID string, disabled bool) (ManagedUser, error) {
	r.disabledUserID = userID
	r.disabledValue = disabled
	for _, user := range r.managedUsers {
		if user.ID == userID {
			if disabled {
				now := time.Now().UTC()
				user.DisabledAt = &now
			} else {
				user.DisabledAt = nil
			}
			return user, nil
		}
	}
	return ManagedUser{}, errors.New("not found")
}

func (r *fakeAdminRepository) SetManagedUserPasswordHash(_ context.Context, userID, passwordHash string) (ManagedUser, error) {
	r.passwordUserID = userID
	r.passwordHash = passwordHash
	for _, user := range r.managedUsers {
		if user.ID == userID {
			return user, nil
		}
	}
	return ManagedUser{}, errors.New("not found")
}

func (r *fakeAdminRepository) ExportData(context.Context) (DataExport, error) {
	return DataExport{Format: "netstamp.admin.data.v1"}, nil
}

func (r *fakeAdminRepository) ImportData(context.Context, DataExport) (DataImportResult, error) {
	return DataImportResult{}, nil
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

func (r *fakeAdminRepository) CreateSystemSettingAuditEvent(_ context.Context, key, action string, _ *string) error {
	r.auditKeys = append(r.auditKeys, key)
	r.auditActions = append(r.auditActions, action)
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
