package label

import "errors"

var (
	ErrLabelAlreadyExists = errors.New("label already exists")
	ErrLabelNotFound      = errors.New("label not found")
	ErrInvalidInput       = errors.New("label input invalid")
)
