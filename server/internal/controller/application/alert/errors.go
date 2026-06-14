package alert

import "errors"

var (
	ErrInvalidInput = errors.New("invalid alert input")
	ErrForbidden    = errors.New("alert action forbidden")
)
