package account

import (
	"context"
	"errors"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

type Service struct {
	repo        Repository
	systemAdmin SystemAdminRepository
	hasher      PasswordHasher
	events      EventRecorder
}

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
	if compareErr := s.hasher.Compare(ctx, input.Password, user.PasswordHash); compareErr != nil {
		return UserOutput{}, flow.businessFailure(UserEventChangeEmailFailure, UserReasonCredentialsInvalid, ErrCredentialsInvalid)
	}

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
	if compareErr := s.hasher.Compare(ctx, input.CurrentPassword, user.PasswordHash); compareErr != nil {
		return flow.businessFailure(UserEventChangePasswordFailure, UserReasonCredentialsInvalid, ErrCredentialsInvalid)
	}

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
	flow.setUser(user)
	flow.success(UserEventChangePasswordSuccess)

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
	flow.setUser(user)
	flow.success(UserEventDeactivateSuccess)

	return nil
}
