package label

import (
	"errors"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrForbidden    = errors.New("label action forbidden")
)
