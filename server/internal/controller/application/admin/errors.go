package admin

import "errors"

var (
	ErrForbidden    = errors.New("admin access forbidden")
	ErrInvalidInput = errors.New("invalid admin settings input")
)
