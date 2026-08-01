package systemsettings

import (
	"context"
	"reflect"
)

type settingUpdate struct {
	key     string
	changed bool
	value   any
}

func (s *Service) UpdateAccess(ctx context.Context, input UpdateAccessInput) (AccessSettings, error) {
	ctx, flow := s.startSettingsFlow(ctx, "update", ResourceAccess, input.CurrentUserID)
	defer flow.end()
	if authorizeErr := flow.authorize(ctx); authorizeErr != nil {
		return AccessSettings{}, authorizeErr
	}

	var result AccessSettings
	txErr := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		var updateErr error
		result, updateErr = s.updateAccessInTx(txCtx, input)
		return updateErr
	})
	return result, txErr
}

func (s *Service) updateAccessInTx(
	ctx context.Context,
	input UpdateAccessInput,
) (AccessSettings, error) {
	lockErr := s.lockAccessSMTP(ctx)
	if lockErr != nil {
		return AccessSettings{}, lockErr
	}

	current, loadErr := s.EffectiveAccess(ctx)
	if loadErr != nil {
		return AccessSettings{}, loadErr
	}
	next := applyAccessPatch(current, input)
	if next.EmailVerificationRequired {
		smtpRows, smtpErr := s.settingsByKeys(ctx, smtpKeys)
		if smtpErr != nil {
			return AccessSettings{}, smtpErr
		}
		smtp, viewErr := s.smtpViewFromRows(smtpRows)
		if viewErr != nil {
			return AccessSettings{}, viewErr
		}
		if invariantErr := validateAccessSMTPInvariant(next, smtp); invariantErr != nil {
			return AccessSettings{}, invariantErr
		}
		if passwordErr := s.validateRetainedSMTPPassword(smtpRows, smtp, OptionalSecret{}); passwordErr != nil {
			return AccessSettings{}, passwordErr
		}
	}
	if reflect.DeepEqual(current, next) {
		return current, nil
	}
	if persistErr := s.persistAccess(ctx, current, next, &input.CurrentUserID); persistErr != nil {
		return AccessSettings{}, persistErr
	}
	return next, nil
}

func applyAccessPatch(current AccessSettings, input UpdateAccessInput) AccessSettings {
	next := current
	if input.AccountCreationEnabled != nil {
		next.AccountCreationEnabled = *input.AccountCreationEnabled
	}
	if input.EmailVerificationRequired != nil {
		next.EmailVerificationRequired = *input.EmailVerificationRequired
	}
	if input.ProjectCreationEnabled != nil {
		next.ProjectCreationEnabled = *input.ProjectCreationEnabled
	}
	if input.CredentialChangesEnabled != nil {
		next.CredentialChangesEnabled = *input.CredentialChangesEnabled
	}
	return next
}

func (s *Service) persistAccess(
	ctx context.Context,
	current AccessSettings,
	next AccessSettings,
	actor *string,
) error {
	return s.persistPublicUpdates(ctx, actor, []settingUpdate{
		{keyAccountCreationEnabled, current.AccountCreationEnabled != next.AccountCreationEnabled, next.AccountCreationEnabled},
		{keyEmailVerificationRequired, current.EmailVerificationRequired != next.EmailVerificationRequired, next.EmailVerificationRequired},
		{keyProjectCreationEnabled, current.ProjectCreationEnabled != next.ProjectCreationEnabled, next.ProjectCreationEnabled},
		{keyCredentialChangesEnabled, current.CredentialChangesEnabled != next.CredentialChangesEnabled, next.CredentialChangesEnabled},
	})
}

func (s *Service) UpdateSMTP(ctx context.Context, input UpdateSMTPInput) (SMTPSettings, error) {
	ctx, flow := s.startSettingsFlow(ctx, "update", ResourceSMTP, input.CurrentUserID)
	defer flow.end()
	if authorizeErr := flow.authorize(ctx); authorizeErr != nil {
		return SMTPSettings{}, authorizeErr
	}
	if secretErr := validateSecretPatch(input.Password, "password"); secretErr != nil {
		return SMTPSettings{}, secretErr
	}
	normalizeSMTPInput(&input)

	var result SMTPSettings
	txErr := s.transactor.WithinTx(ctx, func(txCtx context.Context) error {
		var updateErr error
		result, updateErr = s.updateSMTPInTx(txCtx, input)
		return updateErr
	})
	return result, txErr
}

func (s *Service) updateSMTPInTx(
	ctx context.Context,
	input UpdateSMTPInput,
) (SMTPSettings, error) {
	lockErr := s.lockAccessSMTP(ctx)
	if lockErr != nil {
		return SMTPSettings{}, lockErr
	}

	rows, loadErr := s.settingsByKeys(ctx, smtpKeys)
	if loadErr != nil {
		return SMTPSettings{}, loadErr
	}
	current, viewErr := s.smtpViewFromRows(rows)
	if viewErr != nil {
		return SMTPSettings{}, viewErr
	}
	next := applySMTPPatch(current, input)
	if validationErr := validateSMTP(next); validationErr != nil {
		return SMTPSettings{}, validationErr
	}
	if passwordErr := s.validateRetainedSMTPPassword(rows, next, input.Password); passwordErr != nil {
		return SMTPSettings{}, passwordErr
	}
	access, accessErr := s.EffectiveAccess(ctx)
	if accessErr != nil {
		return SMTPSettings{}, accessErr
	}
	if invariantErr := validateAccessSMTPInvariant(access, next); invariantErr != nil {
		return SMTPSettings{}, invariantErr
	}
	if smtpPatchIsNoop(current, next, input.Password) {
		return current, nil
	}
	if persistErr := s.persistSMTP(ctx, current, next, input.Password, &input.CurrentUserID); persistErr != nil {
		return SMTPSettings{}, persistErr
	}
	return next, nil
}

func (s *Service) validateRetainedSMTPPassword(
	rows map[string]StoredSetting,
	settings SMTPSettings,
	password OptionalSecret,
) error {
	if !settings.PasswordSet || (password.Present && password.Value != nil) {
		return nil
	}
	value, decryptErr := s.decryptSecret(rows, keySMTPPassword)
	if decryptErr != nil || value == "" {
		return invalidField("password", "password must be replaced before using SMTP authentication", nil)
	}
	return nil
}

func normalizeSMTPInput(input *UpdateSMTPInput) {
	input.Host = normalizeStringPointer(input.Host)
	input.Username = normalizeStringPointer(input.Username)
	input.From = normalizeStringPointer(input.From)
	input.TLSMode = normalizeStringPointer(input.TLSMode)
}

func applySMTPPatch(current SMTPSettings, input UpdateSMTPInput) SMTPSettings {
	next := current
	if input.Host != nil {
		next.Host = *input.Host
	}
	if input.Port != nil {
		next.Port = *input.Port
	}
	if input.Username != nil {
		next.Username = *input.Username
	}
	if input.Password.Present {
		next.PasswordSet = input.Password.Value != nil
	}
	if input.From != nil {
		next.From = *input.From
	}
	if input.TLSMode != nil {
		next.TLSMode = *input.TLSMode
	}
	if input.TimeoutSeconds != nil {
		next.TimeoutSeconds = *input.TimeoutSeconds
	}
	next.Configured = smtpDeliveryConfigured(next)
	return next
}

func smtpPatchIsNoop(current, next SMTPSettings, password OptionalSecret) bool {
	publicEqual := current.Host == next.Host &&
		current.Port == next.Port &&
		current.Username == next.Username &&
		current.From == next.From &&
		current.TLSMode == next.TLSMode &&
		current.TimeoutSeconds == next.TimeoutSeconds
	return publicEqual && !secretReplacement(password, current.PasswordSet)
}

func (s *Service) persistSMTP(
	ctx context.Context,
	current SMTPSettings,
	next SMTPSettings,
	password OptionalSecret,
	actor *string,
) error {
	if persistErr := s.persistPublicUpdates(ctx, actor, []settingUpdate{
		{keySMTPHost, current.Host != next.Host, next.Host},
		{keySMTPPort, current.Port != next.Port, next.Port},
		{keySMTPUsername, current.Username != next.Username, next.Username},
		{keySMTPFrom, current.From != next.From, next.From},
		{keySMTPTLSMode, current.TLSMode != next.TLSMode, next.TLSMode},
		{keySMTPTimeoutSeconds, current.TimeoutSeconds != next.TimeoutSeconds, next.TimeoutSeconds},
	}); persistErr != nil {
		return persistErr
	}
	return s.persistProviderSecret(ctx, keySMTPPassword, current.PasswordSet, password, actor)
}

func (s *Service) persistPublicUpdates(ctx context.Context, actor *string, updates []settingUpdate) error {
	for _, update := range updates {
		if !update.changed {
			continue
		}
		if persistErr := s.upsertPublic(ctx, update.key, update.value, actor); persistErr != nil {
			return persistErr
		}
	}
	return nil
}

func (s *Service) lockAccessSMTP(ctx context.Context) error {
	if accessErr := s.repo.LockSystemSettingsResource(ctx, string(ResourceAccess)); accessErr != nil {
		return accessErr
	}
	return s.repo.LockSystemSettingsResource(ctx, string(ResourceSMTP))
}
