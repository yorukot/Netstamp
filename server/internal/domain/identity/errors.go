package identity

import "errors"

var (
	ErrEmailAlreadyExists             = errors.New("email already exists")
	ErrUserNotFound                   = errors.New("user not found")
	ErrSessionNotFound                = errors.New("session not found")
	ErrResetTokenNotFound             = errors.New("password reset token not found")
	ErrEmailVerificationTokenNotFound = errors.New("email verification token not found")
)
