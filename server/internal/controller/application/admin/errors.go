package admin

import (
	"errors"

	domainsystem "github.com/yorukot/netstamp/internal/domain/system"
)

var (
	ErrForbidden              = errors.New("admin access forbidden")
	ErrInvalidInput           = errors.New("invalid admin input")
	ErrLastSystemAdmin        = errors.New("system must keep an administrator")
	ErrSelfSystemAdminRemoval = errors.New("system administrator cannot remove self")
	ErrSelfAccountDisable     = errors.New("system administrator cannot disable self from admin settings")
	ErrSystemAdminNotFound    = errors.New("system administrator not found")
	ErrDataImportInvalid      = domainsystem.ErrDataImportInvalid
)
