package auth

import (
	"errors"
)

var (
	// InvalidInput is for the case where the auth input is invalid.
	// And we use this big group so the transport layer can understand
	// the error and return an appropriate response.
	ErrInvalidInput       = errors.New("auth input invalid")
	ErrCredentialsInvalid = errors.New("credentials invalid")
	ErrAccessTokenInvalid = errors.New("access token invalid")
	ErrResetTokenInvalid  = errors.New("password reset token invalid")
	ErrResetUnavailable   = errors.New("password reset unavailable")
)
