package check

import (
	"errors"
)

var (
	ErrCheckNotFound   = errors.New("check not found")
	ErrInvalidInput    = errors.New("invalid input")
	ErrProjectNotFound = errors.New("project not found")
	ErrLabelNotFound   = errors.New("label not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrForbidden       = errors.New("check action forbidden")
)
