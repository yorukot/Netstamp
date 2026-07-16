package account

import (
	"context"
	"errors"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

type Service struct {
	repo        Repository
	systemAdmin SystemAdminRepository
	sessions    SessionRepository
	apiTokens   APITokenRevoker
	hasher      PasswordHasher
	events      EventRecorder
	authMethods AuthenticationRepository
}

func (s *Service) ConfigureAuthenticationMethods(repo AuthenticationRepository) { s.authMethods = repo }

func NewService(repo Repository, hasher PasswordHasher, events EventRecorder) *Service {
	return &Service{
		repo:   repo,
		hasher: hasher,
		events: events,
	}
}

func (s *Service) ConfigureSystemAdmin(repo SystemAdminRepository) {
	s.systemAdmin = repo
}

func (s *Service) ConfigureSessions(repo SessionRepository) {
	s.sessions = repo
}

func (s *Service) ConfigureAPITokens(revoker APITokenRevoker) { s.apiTokens = revoker }

func (s *Service) UpdateCurrentUser(ctx context.Context, input UpdateCurrentUserInput) (UserOutput, error) {
	ctx, flow := s.startUserFlow(ctx, "user.profile.update", UserActionUpdateProfile, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeUpdateCurrentUserInput(input)
	if err != nil {
		return UserOutput{}, flow.businessFailure(UserEventUpdateProfileFailure, UserReasonInvalidInput, err)
	}

	user, err := s.repo.UpdateUserDisplayName(ctx, identity.User{
		ID:          input.CurrentUserID,
		DisplayName: *input.DisplayName,
	})
	if err != nil {
		return UserOutput{}, flow.updateFailure(UserEventUpdateProfileFailure, err)
	}
	flow.setUser(user)
	flow.success(UserEventUpdateProfileSuccess)

	return UserOutput{User: user}, nil
}

func (s *Service) ChangeCurrentUserEmail(ctx context.Context, input ChangeCurrentUserEmailInput) (UserOutput, error) {
	ctx, flow := s.startUserFlow(ctx, "user.email.change", UserActionChangeEmail, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeChangeCurrentUserEmailInput(input)
	if err != nil {
		return UserOutput{}, flow.businessFailure(UserEventChangeEmailFailure, UserReasonInvalidInput, err)
	}

	user, err := s.repo.GetUserByID(ctx, input.CurrentUserID)
	if err != nil {
		return UserOutput{}, flow.lookupFailure(UserEventChangeEmailFailure, err)
	}
	flow.setUser(user)
	user, err = s.repo.UpdateUserEmail(ctx, identity.User{
		ID:    input.CurrentUserID,
		Email: input.NewEmail,
	})
	if errors.Is(err, identity.ErrEmailAlreadyExists) {
		return UserOutput{}, flow.businessFailure(UserEventChangeEmailFailure, UserReasonEmailAlreadyExists, err)
	}
	if err != nil {
		return UserOutput{}, flow.updateFailure(UserEventChangeEmailFailure, err)
	}
	flow.setUser(user)
	flow.success(UserEventChangeEmailSuccess)

	return UserOutput{User: user}, nil
}

func (s *Service) ChangeCurrentUserPassword(ctx context.Context, input ChangeCurrentUserPasswordInput) error {
	ctx, flow := s.startUserFlow(ctx, "user.password.change", UserActionChangePassword, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeChangeCurrentUserPasswordInput(input)
	if err != nil {
		return flow.businessFailure(UserEventChangePasswordFailure, UserReasonInvalidInput, err)
	}

	user, err := s.repo.GetUserByID(ctx, input.CurrentUserID)
	if err != nil {
		return flow.lookupFailure(UserEventChangePasswordFailure, err)
	}
	flow.setUser(user)
	passwordHash, err := s.hasher.Hash(ctx, input.NewPassword)
	if err != nil {
		return flow.technicalFailure(UserEventChangePasswordFailure, UserReasonPasswordHashFailed, err)
	}

	user, err = s.repo.UpdateUserPasswordHash(ctx, identity.User{
		ID:           input.CurrentUserID,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return flow.updateFailure(UserEventChangePasswordFailure, err)
	}
	if s.sessions != nil {
		if err := s.sessions.RevokeUserSessionsExcept(ctx, input.CurrentUserID, input.CurrentSessionID, "password_change"); err != nil {
			return flow.updateFailure(UserEventChangePasswordFailure, err)
		}
	}
	if s.apiTokens != nil {
		if err := s.apiTokens.RevokeUserTokens(ctx, input.CurrentUserID, "password_change"); err != nil {
			return flow.updateFailure(UserEventChangePasswordFailure, err)
		}
	}
	flow.setUser(user)
	flow.success(UserEventChangePasswordSuccess)

	return nil
}

func (s *Service) ListAuthenticationMethods(ctx context.Context, userID string) (AuthenticationMethodsOutput, error) {
	if s.authMethods == nil || userID == "" {
		return AuthenticationMethodsOutput{}, ErrInvalidInput
	}
	hasPassword, _, err := s.authMethods.CountUserAuthenticationMethods(ctx, userID)
	if err != nil {
		return AuthenticationMethodsOutput{}, err
	}
	identities, err := s.authMethods.ListUserIdentities(ctx, userID)
	if err != nil {
		return AuthenticationMethodsOutput{}, err
	}
	return AuthenticationMethodsOutput{HasPassword: hasPassword, Identities: identities}, nil
}

func (s *Service) RemoveCurrentUserPassword(ctx context.Context, userID, sessionID string) error {
	if s.authMethods == nil || userID == "" {
		return ErrInvalidInput
	}
	_, identityCount, err := s.authMethods.CountUserAuthenticationMethods(ctx, userID)
	if err != nil {
		return err
	}
	if identityCount == 0 {
		return ErrLastCredential
	}
	if _, err := s.authMethods.DeleteUserPasswordCredential(ctx, userID); err != nil {
		return err
	}
	if s.sessions != nil {
		if err := s.sessions.RevokeUserSessionsExcept(ctx, userID, sessionID, "password_removed"); err != nil {
			return err
		}
	}
	if s.apiTokens != nil {
		return s.apiTokens.RevokeUserTokens(ctx, userID, "password_removed")
	}
	return nil
}

func (s *Service) RemoveCurrentUserIdentity(ctx context.Context, userID, sessionID, identityID string) error {
	if s.authMethods == nil || userID == "" || sessionID == "" || identityID == "" {
		return ErrInvalidInput
	}
	hasPassword, identityCount, err := s.authMethods.CountUserAuthenticationMethods(ctx, userID)
	if err != nil {
		return err
	}
	if !hasPassword && identityCount <= 1 {
		return ErrLastCredential
	}
	if err := s.authMethods.DeleteUserIdentity(ctx, userID, identityID); err != nil {
		if errors.Is(err, identity.ErrIdentityNotFound) {
			return ErrIdentityNotFound
		}
		return err
	}
	if s.sessions != nil {
		if err := s.sessions.RevokeUserSessionsExcept(ctx, userID, sessionID, "identity_removed"); err != nil {
			return err
		}
	}
	if s.apiTokens != nil {
		return s.apiTokens.RevokeUserTokens(ctx, userID, "identity_removed")
	}
	return nil
}

func (s *Service) DeactivateCurrentUser(ctx context.Context, input DeactivateCurrentUserInput) error {
	ctx, flow := s.startUserFlow(ctx, "user.deactivate", UserActionDeactivate, input.CurrentUserID)
	defer flow.end()

	input, err := normalizeDeactivateCurrentUserInput(input)
	if err != nil {
		return flow.businessFailure(UserEventDeactivateFailure, UserReasonInvalidInput, err)
	}

	current, err := s.repo.GetUserByID(ctx, input.CurrentUserID)
	if err != nil {
		return flow.lookupFailure(UserEventDeactivateFailure, err)
	}
	flow.setUser(current)
	if current.IsSystemAdmin && current.DisabledAt == nil && s.systemAdmin != nil {
		activeAdmins, countErr := s.systemAdmin.CountActiveSystemAdmins(ctx)
		if countErr != nil {
			return flow.technicalFailure(UserEventDeactivateFailure, UserReasonUserLookupFailed, countErr)
		}
		if activeAdmins <= 1 {
			return flow.businessFailure(UserEventDeactivateFailure, UserReasonInvalidInput, ErrLastSystemAdmin)
		}
	}

	user, err := s.repo.DisableUser(ctx, input.CurrentUserID)
	if err != nil {
		return flow.updateFailure(UserEventDeactivateFailure, err)
	}
	if s.sessions != nil {
		if err := s.sessions.RevokeUserSessions(ctx, input.CurrentUserID, "account_deactivated"); err != nil {
			return flow.updateFailure(UserEventDeactivateFailure, err)
		}
	}
	if s.apiTokens != nil {
		if err := s.apiTokens.RevokeUserTokens(ctx, input.CurrentUserID, "account_deactivated"); err != nil {
			return flow.updateFailure(UserEventDeactivateFailure, err)
		}
	}
	flow.setUser(user)
	flow.success(UserEventDeactivateSuccess)

	return nil
}
