package project

import (
	"errors"
)

var (
	ErrForbidden    = errors.New("project action forbidden")
	ErrInvalidInput = errors.New("project input invalid")
	ErrLastOwner    = errors.New("project must keep an owner")
)
