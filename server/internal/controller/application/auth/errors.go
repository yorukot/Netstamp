package auth

import (
	"errors"

	"github.com/yorukot/netstamp/internal/domain/identity"
)

var (
	ErrEmailAlreadyExists  = identity.ErrEmailAlreadyExists
	ErrInvalidInput        = errors.New("auth input invalid")
	ErrDisplayNameRequired = errors.New("display name required")
	ErrDisplayNameTooLong  = errors.New("display name too long")
	ErrCredentialsInvalid  = errors.New("credentials invalid")
	ErrUserInactive        = errors.New("user inactive")
	ErrAccessTokenInvalid  = identity.ErrAccessTokenInvalid
)
