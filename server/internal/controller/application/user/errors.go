package account

import "errors"

var (
	ErrInvalidInput       = errors.New("user input invalid")
	ErrCredentialsInvalid = errors.New("credentials invalid")
)
