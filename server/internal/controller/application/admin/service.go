package admin

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

type Service struct {
	repo     Repository
	sessions SessionRepository
	cipher   SecretCipher
	hasher   PasswordHasher
	defaults Defaults
}

func NewService(repo Repository, cipher SecretCipher, defaults Defaults, hashers ...PasswordHasher) *Service {
	if defaults.SMTP.Port == 0 {
		defaults.SMTP.Port = 587
	}
	if defaults.SMTP.TLSMode == "" {
		defaults.SMTP.TLSMode = "starttls"
	}
	if defaults.SMTP.TimeoutSeconds == 0 {
		defaults.SMTP.TimeoutSeconds = 10
	}
	defaults.SMTP.PasswordSet = defaults.SMTP.Password != ""
	var hasher PasswordHasher
	if len(hashers) > 0 {
		hasher = hashers[0]
	}
	return &Service{repo: repo, cipher: cipher, hasher: hasher, defaults: defaults}
}

func (s *Service) ConfigureSessions(repo SessionRepository) {
	s.sessions = repo
}

func (s *Service) GetSettings(ctx context.Context, input GetSettingsInput) (Settings, error) {
	if err := s.requireSystemAdmin(ctx, input.CurrentUserID); err != nil {
		return Settings{}, err
	}
	return s.EffectiveSettings(ctx)
}

func (s *Service) ListSystemAdmins(ctx context.Context, input ListSystemAdminsInput) ([]SystemAdmin, error) {
	if err := s.requireSystemAdmin(ctx, input.CurrentUserID); err != nil {
		return nil, err
	}
	return s.repo.ListSystemAdmins(ctx)
}

func (s *Service) ListManagedUsers(ctx context.Context, input ListManagedUsersInput) ([]ManagedUser, error) {
	input, err := normalizeListManagedUsersInput(input)
	if err != nil {
		return nil, err
	}
	if err := s.requireSystemAdmin(ctx, input.CurrentUserID); err != nil {
		return nil, err
	}
	return s.repo.ListManagedUsers(ctx)
}

func (s *Service) GrantSystemAdmin(ctx context.Context, input GrantSystemAdminInput) (SystemAdmin, error) {
	input, err := normalizeGrantSystemAdminInput(input)
	if err != nil {
		return SystemAdmin{}, err
	}
	if requireErr := s.requireSystemAdmin(ctx, input.CurrentUserID); requireErr != nil {
		return SystemAdmin{}, requireErr
	}

	admin, err := s.repo.GrantSystemAdminByEmail(ctx, input.Email)
	if err != nil {
		return SystemAdmin{}, err
	}
	if err := s.repo.CreateSystemSettingAuditEvent(ctx, systemAdminAuditKey(admin), auditActionGrantSystemAdmin, &input.CurrentUserID); err != nil {
		return SystemAdmin{}, err
	}
	if s.sessions != nil {
		if err := s.sessions.RevokeUserSessions(ctx, admin.ID, "system_admin_granted"); err != nil {
			return SystemAdmin{}, err
		}
	}

	return admin, nil
}

func (s *Service) UpdateManagedUser(ctx context.Context, input UpdateManagedUserInput) (ManagedUser, error) {
	input, err := normalizeUpdateManagedUserInput(input)
	if err != nil {
		return ManagedUser{}, err
	}
	if requireErr := s.requireSystemAdmin(ctx, input.CurrentUserID); requireErr != nil {
		return ManagedUser{}, requireErr
	}

	var user ManagedUser
	var loaded bool
	if input.SystemAdmin != nil {
		user, loaded, err = s.updateManagedUserAdmin(ctx, input)
		if err != nil {
			return ManagedUser{}, err
		}
	}

	if input.Disabled != nil {
		user, err = s.updateManagedUserDisabled(ctx, input, user, loaded)
		if err != nil {
			return ManagedUser{}, err
		}
		loaded = true
	}

	if !loaded {
		return s.getManagedUser(ctx, input.UserID)
	}

	return user, nil
}

func (s *Service) updateManagedUserAdmin(ctx context.Context, input UpdateManagedUserInput) (ManagedUser, bool, error) {
	if *input.SystemAdmin {
		user, err := s.grantManagedUserAdmin(ctx, input)
		return user, err == nil, err
	}

	if err := s.revokeManagedUserAdmin(ctx, input); err != nil {
		return ManagedUser{}, false, err
	}
	return ManagedUser{}, false, nil
}

func (s *Service) grantManagedUserAdmin(ctx context.Context, input UpdateManagedUserInput) (ManagedUser, error) {
	user, err := s.repo.GrantSystemAdminByUserID(ctx, input.UserID)
	if err != nil {
		return ManagedUser{}, err
	}
	if err := s.repo.CreateSystemSettingAuditEvent(ctx, managedUserAuditKey(user), auditActionGrantSystemAdmin, &input.CurrentUserID); err != nil {
		return ManagedUser{}, err
	}
	return user, s.revokeSessionsForUser(ctx, input.UserID, "system_admin_granted")
}

func (s *Service) revokeManagedUserAdmin(ctx context.Context, input UpdateManagedUserInput) error {
	if input.CurrentUserID == input.UserID {
		return ErrSelfSystemAdminRemoval
	}
	result, err := s.repo.RevokeSystemAdminIfNotLast(ctx, input.UserID)
	if err != nil {
		return err
	}
	if err := validateSystemAdminRevokeResult(result); err != nil {
		return err
	}
	if err := s.repo.CreateSystemSettingAuditEvent(ctx, "user:"+input.UserID, auditActionRevokeSystemAdmin, &input.CurrentUserID); err != nil {
		return err
	}
	return s.revokeSessionsForUser(ctx, input.UserID, "system_admin_revoked")
}

func (s *Service) revokeSessionsForUser(ctx context.Context, userID, reason string) error {
	if s.sessions == nil {
		return nil
	}
	return s.sessions.RevokeUserSessions(ctx, userID, reason)
}

func validateSystemAdminRevokeResult(result SystemAdminRevokeResult) error {
	if !result.TargetWasAdmin {
		return ErrSystemAdminNotFound
	}
	if !result.Revoked && result.AdminCount <= 1 {
		return ErrLastSystemAdmin
	}
	if !result.Revoked {
		return ErrSystemAdminNotFound
	}
	return nil
}

func (s *Service) updateManagedUserDisabled(ctx context.Context, input UpdateManagedUserInput, current ManagedUser, loaded bool) (ManagedUser, error) {
	if err := s.ensureManagedUserCanSetDisabled(ctx, input, current, loaded); err != nil {
		return ManagedUser{}, err
	}

	user, err := s.repo.SetManagedUserDisabledAt(ctx, input.UserID, *input.Disabled)
	if err != nil {
		return ManagedUser{}, err
	}
	action := managedUserDisabledAuditAction(*input.Disabled)
	if err := s.repo.CreateSystemSettingAuditEvent(ctx, managedUserAuditKey(user), action, &input.CurrentUserID); err != nil {
		return ManagedUser{}, err
	}
	if *input.Disabled && s.sessions != nil {
		if err := s.sessions.RevokeUserSessions(ctx, input.UserID, "account_disabled"); err != nil {
			return ManagedUser{}, err
		}
	}
	return user, nil
}

func (s *Service) ensureManagedUserCanSetDisabled(ctx context.Context, input UpdateManagedUserInput, current ManagedUser, loaded bool) error {
	if !*input.Disabled {
		return nil
	}
	if input.CurrentUserID == input.UserID {
		return ErrSelfAccountDisable
	}

	user := current
	var err error
	if !loaded {
		user, err = s.getManagedUser(ctx, input.UserID)
		if err != nil {
			return err
		}
	}
	if !user.IsSystemAdmin || user.DisabledAt != nil {
		return nil
	}

	count, err := s.repo.CountActiveSystemAdmins(ctx)
	if err != nil {
		return err
	}
	if count <= 1 {
		return ErrLastSystemAdmin
	}
	return nil
}

func managedUserDisabledAuditAction(disabled bool) string {
	if disabled {
		return auditActionDisableUser
	}
	return auditActionEnableUser
}

func (s *Service) getManagedUser(ctx context.Context, userID string) (ManagedUser, error) {
	users, err := s.repo.ListManagedUsers(ctx)
	if err != nil {
		return ManagedUser{}, err
	}
	for _, candidate := range users {
		if candidate.ID == userID {
			return candidate, nil
		}
	}
	return ManagedUser{}, identity.ErrUserNotFound
}

func (s *Service) SetManagedUserPassword(ctx context.Context, input SetManagedUserPasswordInput) (ManagedUser, error) {
	input, err := normalizeSetManagedUserPasswordInput(input)
	if err != nil {
		return ManagedUser{}, err
	}
	if requireErr := s.requireSystemAdmin(ctx, input.CurrentUserID); requireErr != nil {
		return ManagedUser{}, requireErr
	}
	if s.hasher == nil {
		return ManagedUser{}, ErrInvalidInput
	}

	passwordHash, err := s.hasher.Hash(ctx, input.Password)
	if err != nil {
		return ManagedUser{}, err
	}
	user, err := s.repo.SetManagedUserPasswordHash(ctx, input.UserID, passwordHash)
	if err != nil {
		return ManagedUser{}, err
	}
	if auditErr := s.repo.CreateSystemSettingAuditEvent(ctx, managedUserAuditKey(user), auditActionSetPassword, &input.CurrentUserID); auditErr != nil {
		return ManagedUser{}, auditErr
	}
	if s.sessions != nil {
		if err := s.sessions.RevokeUserSessions(ctx, input.UserID, "admin_password_set"); err != nil {
			return ManagedUser{}, err
		}
	}
	return user, nil
}

func (s *Service) RevokeSystemAdmin(ctx context.Context, input RevokeSystemAdminInput) error {
	input, err := normalizeRevokeSystemAdminInput(input)
	if err != nil {
		return err
	}
	if requireErr := s.requireSystemAdmin(ctx, input.CurrentUserID); requireErr != nil {
		return requireErr
	}
	if input.CurrentUserID == input.UserID {
		return ErrSelfSystemAdminRemoval
	}

	result, err := s.repo.RevokeSystemAdminIfNotLast(ctx, input.UserID)
	if err != nil {
		return err
	}
	if !result.TargetWasAdmin {
		return ErrSystemAdminNotFound
	}
	if !result.Revoked && result.AdminCount <= 1 {
		return ErrLastSystemAdmin
	}
	if !result.Revoked {
		return ErrSystemAdminNotFound
	}

	admin := SystemAdmin{ID: input.UserID}
	if err := s.repo.CreateSystemSettingAuditEvent(ctx, systemAdminAuditKey(admin), auditActionRevokeSystemAdmin, &input.CurrentUserID); err != nil {
		return err
	}
	if s.sessions != nil {
		return s.sessions.RevokeUserSessions(ctx, input.UserID, "system_admin_revoked")
	}
	return nil
}

func (s *Service) ExportData(ctx context.Context, input ExportDataInput) (DataExport, error) {
	input, err := normalizeExportDataInput(input)
	if err != nil {
		return DataExport{}, err
	}
	if requireErr := s.requireSystemAdmin(ctx, input.CurrentUserID); requireErr != nil {
		return DataExport{}, requireErr
	}
	return s.repo.ExportData(ctx)
}

func (s *Service) ImportData(ctx context.Context, input ImportDataInput) (DataImportResult, error) {
	input, err := normalizeImportDataInput(input)
	if err != nil {
		return DataImportResult{}, err
	}
	if requireErr := s.requireSystemAdmin(ctx, input.CurrentUserID); requireErr != nil {
		return DataImportResult{}, requireErr
	}
	result, err := s.repo.ImportData(ctx, input.Export)
	if err != nil {
		return DataImportResult{}, err
	}
	if auditErr := s.repo.CreateSystemSettingAuditEvent(ctx, "data_import", auditActionDataImport, &input.CurrentUserID); auditErr != nil {
		return DataImportResult{}, auditErr
	}
	return result, nil
}

func (s *Service) UpdateSettings(ctx context.Context, input UpdateSettingsInput) (Settings, error) {
	if err := s.requireSystemAdmin(ctx, input.CurrentUserID); err != nil {
		return Settings{}, err
	}

	next, err := s.EffectiveSettings(ctx)
	if err != nil {
		return Settings{}, err
	}
	next, changed := applyUpdate(next, input)
	if err := validateSettings(next); err != nil {
		return Settings{}, err
	}

	actor := &input.CurrentUserID
	for key, value := range changed {
		setting, err := s.settingFor(key, value, actor)
		if err != nil {
			return Settings{}, err
		}
		if _, err := s.repo.UpsertSystemSetting(ctx, setting); err != nil {
			return Settings{}, err
		}
		if err := s.repo.CreateSystemSettingAuditEvent(ctx, key, auditActionUpdate, actor); err != nil {
			return Settings{}, err
		}
	}

	return s.EffectiveSettings(ctx)
}

func (s *Service) EffectiveSettings(ctx context.Context) (Settings, error) {
	settings := Settings{
		RegistrationEnabled:       s.defaults.RegistrationEnabled,
		EmailVerificationRequired: s.defaults.EmailVerificationRequired,
		BackendBaseURL:            s.defaults.BackendBaseURL,
		PublicWebBaseURL:          s.defaults.PublicWebBaseURL,
		SMTP:                      s.defaults.SMTP,
	}
	values, err := s.storedSettings(ctx)
	if err != nil {
		return Settings{}, err
	}

	applyBool(values, keyRegistrationEnabled, &settings.RegistrationEnabled)
	applyBool(values, keyEmailVerificationRequired, &settings.EmailVerificationRequired)
	applyString(values, keyBackendBaseURL, &settings.BackendBaseURL)
	applyString(values, keyPublicWebBaseURL, &settings.PublicWebBaseURL)
	applyString(values, keySMTPHost, &settings.SMTP.Host)
	applyInt32(values, keySMTPPort, &settings.SMTP.Port)
	applyString(values, keySMTPUsername, &settings.SMTP.Username)
	if err := applySecretString(values, keySMTPPassword, &settings.SMTP.Password, s.cipher); err != nil {
		return Settings{}, err
	}
	applyString(values, keySMTPFrom, &settings.SMTP.From)
	applyString(values, keySMTPTLSMode, &settings.SMTP.TLSMode)
	applyInt32(values, keySMTPTimeoutSeconds, &settings.SMTP.TimeoutSeconds)
	settings.SMTP.PasswordSet = settings.SMTP.Password != ""

	if settings.SMTP.Port == 0 {
		settings.SMTP.Port = 587
	}
	if settings.SMTP.TLSMode == "" {
		settings.SMTP.TLSMode = "starttls"
	}
	if settings.SMTP.TimeoutSeconds == 0 {
		settings.SMTP.TimeoutSeconds = 10
	}

	return settings, nil
}

func (s *Service) EffectiveSMTP(ctx context.Context) (SMTPSettings, error) {
	settings, err := s.EffectiveSettings(ctx)
	if err != nil {
		return SMTPSettings{}, err
	}
	return settings.SMTP, nil
}

func (s *Service) BackendBaseURL(ctx context.Context) (string, error) {
	settings, err := s.EffectiveSettings(ctx)
	if err != nil {
		return "", err
	}
	return settings.BackendBaseURL, nil
}

func (s *Service) SMTPConfigured(ctx context.Context) bool {
	smtpSettings, err := s.EffectiveSMTP(ctx)
	if err != nil {
		return false
	}
	return smtpSettings.Host != "" && smtpSettings.From != ""
}

func (s *Service) IsSystemAdmin(ctx context.Context, userID string) (bool, error) {
	if s == nil || s.repo == nil {
		return false, nil
	}
	return s.repo.IsSystemAdmin(ctx, userID)
}

func (s *Service) requireSystemAdmin(ctx context.Context, userID string) error {
	ok, err := s.IsSystemAdmin(ctx, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}

func (s *Service) storedSettings(ctx context.Context) (map[string]StoredSetting, error) {
	if s.repo == nil {
		return map[string]StoredSetting{}, nil
	}
	rows, err := s.repo.ListSystemSettings(ctx)
	if err != nil {
		return nil, err
	}
	values := make(map[string]StoredSetting, len(rows))
	for _, row := range rows {
		values[row.Key] = row
	}
	return values, nil
}

func applyUpdate(settings Settings, input UpdateSettingsInput) (Settings, map[string]any) {
	changed := make(map[string]any)
	if input.RegistrationEnabled != nil {
		settings.RegistrationEnabled = *input.RegistrationEnabled
		changed[keyRegistrationEnabled] = settings.RegistrationEnabled
	}
	if input.EmailVerificationRequired != nil {
		settings.EmailVerificationRequired = *input.EmailVerificationRequired
		changed[keyEmailVerificationRequired] = settings.EmailVerificationRequired
	}
	if value := trimStringPointer(input.BackendBaseURL); value != nil {
		settings.BackendBaseURL = *value
		changed[keyBackendBaseURL] = settings.BackendBaseURL
	}
	if value := trimStringPointer(input.PublicWebBaseURL); value != nil {
		settings.PublicWebBaseURL = *value
		changed[keyPublicWebBaseURL] = settings.PublicWebBaseURL
	}

	if value := trimStringPointer(input.SMTP.Host); value != nil {
		settings.SMTP.Host = *value
		changed[keySMTPHost] = settings.SMTP.Host
	}
	if input.SMTP.Port != nil {
		settings.SMTP.Port = *input.SMTP.Port
		changed[keySMTPPort] = settings.SMTP.Port
	}
	if value := trimStringPointer(input.SMTP.Username); value != nil {
		settings.SMTP.Username = *value
		changed[keySMTPUsername] = settings.SMTP.Username
	}
	if input.SMTP.ClearPassword {
		settings.SMTP.Password = ""
		settings.SMTP.PasswordSet = false
		changed[keySMTPPassword] = settings.SMTP.Password
	} else if value := input.SMTP.Password; value != nil {
		settings.SMTP.Password = *value
		settings.SMTP.PasswordSet = *value != ""
		changed[keySMTPPassword] = settings.SMTP.Password
	}
	if value := trimStringPointer(input.SMTP.From); value != nil {
		settings.SMTP.From = *value
		changed[keySMTPFrom] = settings.SMTP.From
	}
	if value := trimStringPointer(input.SMTP.TLSMode); value != nil {
		settings.SMTP.TLSMode = *value
		changed[keySMTPTLSMode] = settings.SMTP.TLSMode
	}
	if input.SMTP.TimeoutSeconds != nil {
		settings.SMTP.TimeoutSeconds = *input.SMTP.TimeoutSeconds
		changed[keySMTPTimeoutSeconds] = settings.SMTP.TimeoutSeconds
	}
	return settings, changed
}

func (s *Service) settingFor(key string, value any, actor *string) (StoredSetting, error) {
	setting := StoredSetting{Key: key, UpdatedByUserID: actor}
	if key == keySMTPPassword {
		if s.cipher == nil {
			return StoredSetting{}, errors.New("secret cipher is unavailable")
		}
		password, ok := value.(string)
		if !ok {
			return StoredSetting{}, errors.New("secret setting value must be a string")
		}
		ciphertext, nonce, err := s.cipher.Encrypt(password)
		if err != nil {
			return StoredSetting{}, err
		}
		setting.Secret = true
		setting.EncryptedValue = ciphertext
		setting.EncryptedValueNonce = nonce
		return setting, nil
	}
	setting.Value = jsonValue(value)
	return setting, nil
}

func applyBool(values map[string]StoredSetting, key string, target *bool) {
	setting, ok := values[key]
	if !ok || setting.Secret || len(setting.Value) == 0 {
		return
	}
	var value bool
	if err := json.Unmarshal(setting.Value, &value); err == nil {
		*target = value
	}
}

func applyString(values map[string]StoredSetting, key string, target *string) {
	setting, ok := values[key]
	if !ok || setting.Secret || len(setting.Value) == 0 {
		return
	}
	var value string
	if err := json.Unmarshal(setting.Value, &value); err == nil {
		*target = value
	}
}

func applyInt32(values map[string]StoredSetting, key string, target *int32) {
	setting, ok := values[key]
	if !ok || setting.Secret || len(setting.Value) == 0 {
		return
	}
	var value int32
	if err := json.Unmarshal(setting.Value, &value); err == nil {
		*target = value
	}
}

func applySecretString(values map[string]StoredSetting, key string, target *string, cipher SecretCipher) error {
	setting, ok := values[key]
	if !ok || !setting.Secret {
		return nil
	}
	if cipher == nil || len(setting.EncryptedValue) == 0 || len(setting.EncryptedValueNonce) == 0 {
		return errors.New("secret setting cannot be decrypted")
	}
	value, err := cipher.Decrypt(setting.EncryptedValue, setting.EncryptedValueNonce)
	if err != nil {
		return err
	}
	*target = value
	return nil
}

func DurationSeconds(value time.Duration) int32 {
	if value <= 0 {
		return 0
	}
	return int32(value.Seconds())
}
