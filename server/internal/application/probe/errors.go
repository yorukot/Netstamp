package probe

import "errors"

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrLabelNotFound   = errors.New("probe label not found")
	ErrInvalidInput    = errors.New("probe input invalid")
)
