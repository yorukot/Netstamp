package systemsettings

import (
	"errors"
)

var (
	ErrForbidden           = errors.New("system settings access forbidden")
	ErrInvalidInput        = errors.New("invalid system settings input")
	ErrProviderUnavailable = errors.New("external provider readiness check unavailable")
	ErrSMTPTestFailed      = errors.New("SMTP test failed")
)
