package admin

import (
	"context"
	"errors"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

type Service struct {
	repo           Repository
	sessions       SessionRepository
	apiTokens      APITokenRevoker
	hasher         PasswordHasher
	authMethods    AuthenticationMethodRepository
	updateStatus   UpdateStatusReader
	updateSettings UpdateSettingsReader
}

func (s *Service) ConfigureAuthenticationMethods(repo AuthenticationMethodRepository) {
	s.authMethods = repo
}

func NewService(repo Repository, hashers ...PasswordHasher) *Service {
	var hasher PasswordHasher
	if len(hashers) > 0 {
		hasher = hashers[0]
	}
	return &Service{repo: repo, hasher: hasher}
}

func (s *Service) ConfigureSessions(repo SessionRepository) {
	s.sessions = repo
}

func (s *Service) ConfigureAPITokens(revoker APITokenRevoker) { s.apiTokens = revoker }

func (s *Service) ConfigureUpdateStatus(reader UpdateStatusReader) { s.updateStatus = reader }

func (s *Service) ConfigureUpdateSettings(reader UpdateSettingsReader) { s.updateSettings = reader }

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

func (s *Service) GetUpdateStatus(ctx context.Context, input UpdateStatusInput) (UpdateStatus, error) {
	if err := s.requireSystemAdmin(ctx, input.CurrentUserID); err != nil {
		return UpdateStatus{}, err
	}
	if s.updateSettings == nil {
		return UpdateStatus{}, errors.New("update settings reader is unavailable")
	}
	enabled, err := s.updateSettings.UpdateCheckEnabled(ctx)
	if err != nil {
		return UpdateStatus{}, err
	}
	if s.updateStatus == nil {
		return UpdateStatus{}, errors.New("update status reader is unavailable")
	}
	status := s.updateStatus.ReadUpdateStatus()
	if !enabled {
		return UpdateStatus{CurrentVersion: status.CurrentVersion}, nil
	}
	return status, nil
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
	if *input.Disabled && s.apiTokens != nil {
		if err := s.apiTokens.RevokeUserTokens(ctx, input.UserID, "account_disabled"); err != nil {
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
	if s.apiTokens != nil {
		if err := s.apiTokens.RevokeUserTokens(ctx, input.UserID, "admin_password_set"); err != nil {
			return ManagedUser{}, err
		}
	}
	return user, nil
}

func (s *Service) ClearManagedUserPassword(ctx context.Context, input ClearManagedUserPasswordInput) (ManagedUser, error) {
	input, err := normalizeClearManagedUserPasswordInput(input)
	if err != nil || s.authMethods == nil {
		return ManagedUser{}, ErrInvalidInput
	}
	if requireErr := s.requireSystemAdmin(ctx, input.CurrentUserID); requireErr != nil {
		return ManagedUser{}, requireErr
	}
	_, identityCount, err := s.authMethods.CountUserAuthenticationMethods(ctx, input.UserID)
	if err != nil {
		return ManagedUser{}, err
	}
	if identityCount == 0 {
		return ManagedUser{}, identity.ErrLastAuthenticationMethod
	}
	repo, ok := s.repo.(ManagedPasswordRepository)
	if !ok {
		return ManagedUser{}, ErrInvalidInput
	}
	user, err := repo.ClearManagedUserPassword(ctx, input.UserID)
	if err != nil {
		return ManagedUser{}, err
	}
	if auditErr := s.repo.CreateSystemSettingAuditEvent(ctx, managedUserAuditKey(user), auditActionClearPassword, &input.CurrentUserID); auditErr != nil {
		return ManagedUser{}, auditErr
	}
	if s.sessions != nil {
		if err := s.sessions.RevokeUserSessions(ctx, input.UserID, "admin_password_cleared"); err != nil {
			return ManagedUser{}, err
		}
	}
	if s.apiTokens != nil {
		if err := s.apiTokens.RevokeUserTokens(ctx, input.UserID, "admin_password_cleared"); err != nil {
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
