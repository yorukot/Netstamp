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
	ErrIdentityNotFound               = errors.New("identity not found")
	ErrIdentityConflict               = errors.New("identity conflict")
	ErrOIDCFlowNotFound               = errors.New("oidc flow not found")
	ErrLastAuthenticationMethod       = errors.New("last authentication method")
)
