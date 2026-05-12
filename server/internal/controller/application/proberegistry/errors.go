package proberegistry

import (
	"errors"
)

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrForbidden    = errors.New("probe action forbidden")
)
