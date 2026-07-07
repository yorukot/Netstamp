package admin

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

type Service struct {
	repo     Repository
	cipher   SecretCipher
	defaults Defaults
}

func NewService(repo Repository, cipher SecretCipher, defaults Defaults) *Service {
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
	return &Service{repo: repo, cipher: cipher, defaults: defaults}
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

	return admin, nil
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
	return s.repo.CreateSystemSettingAuditEvent(ctx, systemAdminAuditKey(admin), auditActionRevokeSystemAdmin, &input.CurrentUserID)
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
