package publicpage

import "errors"

var (
	ErrInvalidInput = errors.New("invalid public page input")
	ErrForbidden    = errors.New("public page action forbidden")
)
