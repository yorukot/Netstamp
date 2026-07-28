package admin

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
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

func TestTestSMTPRequiresSystemAdmin(t *testing.T) {
	repo := &fakeAdminRepository{admins: map[string]bool{testAdminUserID: false}}
	users := &fakeSMTPTestUserRepository{user: identity.User{ID: testAdminUserID, Email: "admin@example.com"}}
	tester := &fakeSMTPTester{}
	svc := NewService(repo, fakeSecretCipher{}, Defaults{})
	svc.ConfigureSMTPTest(users, tester)

	err := svc.TestSMTP(context.Background(), TestSMTPInput{CurrentUserID: testAdminUserID})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden error, got %v", err)
	}
	if users.calls != 0 || tester.calls != 0 {
		t.Fatalf("expected authorization failure before SMTP work, lookups=%d sends=%d", users.calls, tester.calls)
	}
}

func TestTestSMTPSendsToCurrentAdministrator(t *testing.T) {
	repo := &fakeAdminRepository{admins: map[string]bool{testAdminUserID: true}}
	users := &fakeSMTPTestUserRepository{user: identity.User{ID: testAdminUserID, Email: "admin@example.com"}}
	tester := &fakeSMTPTester{}
	svc := NewService(repo, fakeSecretCipher{}, Defaults{})
	svc.ConfigureSMTPTest(users, tester)

	err := svc.TestSMTP(context.Background(), TestSMTPInput{CurrentUserID: testAdminUserID})
	if err != nil {
		t.Fatalf("test SMTP: %v", err)
	}
	if users.userID != testAdminUserID || tester.recipient != users.user.Email {
		t.Fatalf("unexpected SMTP test target: user=%q recipient=%q", users.userID, tester.recipient)
	}
}

func TestTestSMTPMapsDeliveryFailure(t *testing.T) {
	repo := &fakeAdminRepository{admins: map[string]bool{testAdminUserID: true}}
	users := &fakeSMTPTestUserRepository{user: identity.User{ID: testAdminUserID, Email: "admin@example.com"}}
	tester := &fakeSMTPTester{err: errors.New("delivery failed")}
	svc := NewService(repo, fakeSecretCipher{}, Defaults{})
	svc.ConfigureSMTPTest(users, tester)

	err := svc.TestSMTP(context.Background(), TestSMTPInput{CurrentUserID: testAdminUserID})
	if !errors.Is(err, ErrSMTPTestFailed) {
		t.Fatalf("expected SMTP test failure, got %v", err)
	}
}

func TestUpdateSettingsStoresEncryptedPasswordAndAuditEvents(t *testing.T) {
	repo := &fakeAdminRepository{
		admins:   map[string]bool{"admin-1": true},
		settings: map[string]StoredSetting{},
	}
	svc := NewService(repo, fakeSecretCipher{}, Defaults{
		RegistrationEnabled:      true,
		ProjectCreationEnabled:   true,
		CredentialChangesEnabled: true,
		SMTP: SMTPSettings{
			Port:           587,
			TLSMode:        "starttls",
			TimeoutSeconds: 10,
		},
	})

	registrationEnabled := false
	projectCreationEnabled := false
	credentialChangesEnabled := false
	host := "smtp.netstamp.test"
	port := int32(465)
	username := "netstamp"
	password := "smtp-secret"
	from := "alerts@netstamp.test"
	tlsMode := "implicit"
	timeoutSeconds := int32(15)

	settings, err := svc.UpdateSettings(context.Background(), UpdateSettingsInput{
		CurrentUserID:            "admin-1",
		RegistrationEnabled:      &registrationEnabled,
		ProjectCreationEnabled:   &projectCreationEnabled,
		CredentialChangesEnabled: &credentialChangesEnabled,
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
	if settings.ProjectCreationEnabled || settings.CredentialChangesEnabled {
		t.Fatal("expected instance access policies to be disabled")
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

	for _, key := range []string{keyRegistrationEnabled, keyProjectCreationEnabled, keyCredentialChangesEnabled, keySMTPPassword} {
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

func TestUpdateSettingsStoresProviderSecretEncrypted(t *testing.T) {
	repo := &fakeAdminRepository{admins: map[string]bool{"admin-1": true}, settings: map[string]StoredSetting{}}
	svc := NewService(repo, fakeSecretCipher{}, Defaults{})
	enabled := true
	issuerURL := "https://identity.example.com"
	clientID := "netstamp"
	clientSecret := "provider-secret"
	displayName := "Company SSO"

	settings, err := svc.UpdateSettings(context.Background(), UpdateSettingsInput{CurrentUserID: "admin-1", OIDC: UpdateExternalProviderSettingsInput{Enabled: &enabled, IssuerURL: &issuerURL, ClientID: &clientID, ClientSecret: &clientSecret, DisplayName: &displayName}})
	if err != nil {
		t.Fatalf("update provider settings: %v", err)
	}
	if !settings.OIDC.Enabled || settings.OIDC.ClientSecret != clientSecret || !settings.OIDC.ClientSecretSet {
		t.Fatalf("unexpected effective OIDC settings: %#v", settings.OIDC)
	}
	stored := repo.settings[keyOIDCClientSecret]
	if !stored.Secret || string(stored.EncryptedValue) == clientSecret || len(stored.Value) != 0 {
		t.Fatalf("expected encrypted provider secret, got %#v", stored)
	}
	if strings.Contains(string(repo.settings[keyOIDCSettings].Value), clientSecret) {
		t.Fatal("provider public settings must not contain the client secret")
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

func TestClearManagedUserPasswordRequiresAnotherAuthenticationMethod(t *testing.T) {
	repo := &fakeAdminRepository{admins: map[string]bool{testAdminUserID: true}}
	svc := NewService(repo, fakeSecretCipher{}, Defaults{})
	svc.ConfigureAuthenticationMethods(fakeAuthenticationMethodRepository{identityCount: 0})

	_, err := svc.ClearManagedUserPassword(context.Background(), ClearManagedUserPasswordInput{CurrentUserID: testAdminUserID, UserID: testTargetUserID})
	if !errors.Is(err, identity.ErrLastAuthenticationMethod) {
		t.Fatalf("expected last authentication method error, got %v", err)
	}
	if repo.clearedPasswordUserID != "" {
		t.Fatalf("expected password to remain, clear called for %q", repo.clearedPasswordUserID)
	}
}

func TestClearManagedUserPasswordAuditsRemoval(t *testing.T) {
	repo := &fakeAdminRepository{
		admins: map[string]bool{testAdminUserID: true},
		managedUsers: []ManagedUser{{
			ID: testTargetUserID, Email: "person@example.com", DisplayName: "Person", HasPassword: true,
		}},
	}
	svc := NewService(repo, fakeSecretCipher{}, Defaults{})
	svc.ConfigureAuthenticationMethods(fakeAuthenticationMethodRepository{identityCount: 1})

	user, err := svc.ClearManagedUserPassword(context.Background(), ClearManagedUserPasswordInput{CurrentUserID: testAdminUserID, UserID: testTargetUserID})
	if err != nil {
		t.Fatalf("clear managed user password: %v", err)
	}
	if user.HasPassword || repo.clearedPasswordUserID != testTargetUserID {
		t.Fatalf("unexpected cleared user: %#v", user)
	}
	if !slices.Contains(repo.auditActions, auditActionClearPassword) {
		t.Fatalf("expected password removal audit event, got %#v", repo.auditActions)
	}
}

type fakeAdminRepository struct {
	admins                map[string]bool
	settings              map[string]StoredSetting
	systemAdmins          []SystemAdmin
	managedUsers          []ManagedUser
	systemAdminByEmail    map[string]SystemAdmin
	systemAdminByID       map[string]ManagedUser
	revokeResult          SystemAdminRevokeResult
	activeAdminCount      int64
	grantedEmail          string
	grantedUserID         string
	revokedUserID         string
	disabledUserID        string
	disabledValue         bool
	passwordUserID        string
	clearedPasswordUserID string
	passwordHash          string
	auditActions          []string
	auditKeys             []string
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

func (r *fakeAdminRepository) ClearManagedUserPassword(_ context.Context, userID string) (ManagedUser, error) {
	r.clearedPasswordUserID = userID
	for _, user := range r.managedUsers {
		if user.ID == userID {
			user.HasPassword = false
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

type fakeSMTPTestUserRepository struct {
	user   identity.User
	err    error
	userID string
	calls  int
}

func (r *fakeSMTPTestUserRepository) GetUserByID(_ context.Context, userID string) (identity.User, error) {
	r.calls++
	r.userID = userID
	return r.user, r.err
}

type fakeSMTPTester struct {
	recipient string
	err       error
	calls     int
}

func (t *fakeSMTPTester) SendTestEmail(_ context.Context, recipient string) error {
	t.calls++
	t.recipient = recipient
	return t.err
}

type fakeAuthenticationMethodRepository struct {
	hasPassword   bool
	identityCount int64
}

func (r fakeAuthenticationMethodRepository) CountUserAuthenticationMethods(context.Context, string) (bool, int64, error) {
	return r.hasPassword, r.identityCount, nil
}

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
