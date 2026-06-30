package identity

import "errors"

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrResetTokenNotFound = errors.New("password reset token not found")
)
