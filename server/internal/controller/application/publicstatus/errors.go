package publicstatus

import "errors"

var (
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidInput = errors.New("public status page input invalid")
)
