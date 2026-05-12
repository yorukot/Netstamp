package check

import (
	"errors"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrForbidden    = errors.New("check action forbidden")
)
