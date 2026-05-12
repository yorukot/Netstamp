package proberuntime

import (
	"errors"
)

var (
	ErrInvalidInput      = errors.New("probe runtime input invalid")
	ErrResultConflict    = errors.New("probe result conflicts with assignment")
	ErrUnsupportedResult = errors.New("probe result type unsupported")

	errSecretVerifierMissing = errors.New("probe secret verifier is not configured")
)
