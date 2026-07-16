package account

import "github.com/yorukot/netstamp/internal/domain/identity"

type UpdateCurrentUserInput struct {
	CurrentUserID string
	DisplayName   *string
}

type ChangeCurrentUserEmailInput struct {
	CurrentUserID string
	NewEmail      string
}

type ChangeCurrentUserPasswordInput struct {
	CurrentUserID    string
	CurrentSessionID string
	NewPassword      string
}

type AuthenticationMethodsOutput struct {
	HasPassword bool
	Identities  []identity.UserIdentity
}

type DeactivateCurrentUserInput struct {
	CurrentUserID string
}

type UserOutput struct {
	User identity.User
}
