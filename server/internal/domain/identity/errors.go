package identity

import "errors"

var (
	ErrEmailAlreadyExists             = errors.New("email already exists")
	ErrUserNotFound                   = errors.New("user not found")
	ErrSessionNotFound                = errors.New("session not found")
	ErrAPITokenNotFound               = errors.New("api token not found")
	ErrAPITokenLimitReached           = errors.New("api token limit reached")
	ErrResetTokenNotFound             = errors.New("password reset token not found")
	ErrEmailVerificationTokenNotFound = errors.New("email verification token not found")
)
