package auth

import (
	"errors"
)

var (
	// InvalidInput is for the case where the auth input is invalid.
	// And we use this big group so the transport layer can understand
	// the error and return an appropriate response.
	ErrInvalidInput       = errors.New("auth input invalid")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrCredentialsInvalid = errors.New("credentials invalid")
	ErrAccessTokenInvalid = errors.New("access token invalid")
	ErrUserNotFound       = errors.New("user not found")
)
