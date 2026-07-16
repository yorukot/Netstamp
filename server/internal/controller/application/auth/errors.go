package auth

import (
	"errors"
)

var (
	// InvalidInput is for the case where the auth input is invalid.
	// And we use this big group so the transport layer can understand
	// the error and return an appropriate response.
	ErrInvalidInput                  = errors.New("auth input invalid")
	ErrCredentialsInvalid            = errors.New("credentials invalid")
	ErrSessionInvalid                = errors.New("session invalid")
	ErrAccountDisabled               = errors.New("account disabled")
	ErrResetTokenInvalid             = errors.New("password reset token invalid")
	ErrResetUnavailable              = errors.New("password reset unavailable")
	ErrEmailVerificationRequired     = errors.New("email verification required")
	ErrEmailVerificationTokenInvalid = errors.New("email verification token invalid")
	ErrEmailVerificationUnavailable  = errors.New("email verification unavailable")
	ErrSudoRequired                  = errors.New("recent authentication required")
	ErrExternalAuthUnavailable       = errors.New("external authentication unavailable")
	ErrExternalAuthCallbackInvalid   = errors.New("external authentication callback invalid")
	ErrExternalAuthSudoUnsupported   = errors.New("external authentication does not support sudo")
	ErrOIDCUnavailable               = ErrExternalAuthUnavailable
	ErrOIDCCallbackInvalid           = ErrExternalAuthCallbackInvalid
	ErrIdentityConflict              = errors.New("identity conflict")
	ErrIdentityNotFound              = errors.New("identity not found")
	ErrJITProvisioningDisabled       = errors.New("external authentication jit provisioning disabled")
)
