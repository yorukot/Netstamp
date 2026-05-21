package account

import "github.com/yorukot/netstamp/internal/domain/identity"

type UpdateCurrentUserInput struct {
	CurrentUserID string
	DisplayName   *string
}

type ChangeCurrentUserEmailInput struct {
	CurrentUserID string
	NewEmail      string
	Password      string
}

type ChangeCurrentUserPasswordInput struct {
	CurrentUserID   string
	CurrentPassword string
	NewPassword     string
}

type UserOutput struct {
	User identity.User
}
