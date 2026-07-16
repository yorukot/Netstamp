package apitoken

import "errors"

var (
	ErrInvalidInput       = errors.New("invalid api token input")
	ErrCredentialsInvalid = errors.New("invalid credentials")
	ErrTokenNotFound      = errors.New("api token not found")
	ErrTokenInvalid       = errors.New("invalid api token")
	ErrTokenLimitReached  = errors.New("api token limit reached")
)
