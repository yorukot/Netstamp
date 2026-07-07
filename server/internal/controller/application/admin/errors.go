package admin

import "errors"

var (
	ErrForbidden              = errors.New("admin access forbidden")
	ErrInvalidInput           = errors.New("invalid admin input")
	ErrLastSystemAdmin        = errors.New("system must keep an administrator")
	ErrSelfSystemAdminRemoval = errors.New("system administrator cannot remove self")
	ErrSystemAdminNotFound    = errors.New("system administrator not found")
)
