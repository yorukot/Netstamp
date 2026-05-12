package proberegistry

import (
	"errors"
)

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrLabelNotFound   = errors.New("label not found")
	ErrProbeNotFound   = errors.New("probe not found")
	ErrInvalidInput    = errors.New("invalid input")
	ErrForbidden       = errors.New("probe action forbidden")
)
