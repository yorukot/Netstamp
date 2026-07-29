package systemsettings

import (
	"errors"
	"fmt"
)

var (
	ErrForbidden            = errors.New("system settings access forbidden")
	ErrInvalidInput         = errors.New("invalid system settings input")
	ErrPreconditionRequired = errors.New("system settings precondition required")
	ErrVersionConflict      = errors.New("system settings version conflict")
	ErrProviderUnavailable  = errors.New("external provider readiness check unavailable")
	ErrSMTPTestFailed       = errors.New("SMTP test failed")
)

type VersionConflictError struct {
	Resource Resource
	Expected int64
	Current  int64
}

func (e *VersionConflictError) Error() string {
	return fmt.Sprintf("%s revision conflict: expected %d, current %d", e.Resource, e.Expected, e.Current)
}

func (e *VersionConflictError) Unwrap() error {
	return ErrVersionConflict
}
