package account

import "errors"

var (
	ErrInvalidInput       = errors.New("user input invalid")
	ErrCredentialsInvalid = errors.New("credentials invalid")
	ErrLastSystemAdmin    = errors.New("system must keep an administrator")
	ErrLastCredential     = errors.New("account must keep an authentication method")
	ErrIdentityNotFound   = errors.New("identity not found")
)
