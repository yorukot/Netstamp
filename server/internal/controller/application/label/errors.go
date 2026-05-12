package label

import (
	"errors"
)

var (
	ErrProjectNotFound    = errors.New("project not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrLabelNotFound      = errors.New("label not found")
	ErrLabelAlreadyExists = errors.New("label already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrForbidden          = errors.New("label action forbidden")
)
