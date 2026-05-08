package project

import (
	"errors"

	"github.com/yorukot/netstamp/internal/domain/identity"
	domainproject "github.com/yorukot/netstamp/internal/domain/project"
)

var (
	ErrProjectNotFound          = domainproject.ErrProjectNotFound
	ErrProjectSlugAlreadyExists = domainproject.ErrProjectSlugAlreadyExists
	ErrForbidden                = errors.New("project action forbidden")
	ErrInvalidInput             = errors.New("project input invalid")
	ErrInvalidRole              = errors.New("project member role invalid")
	ErrMemberAlreadyExists      = domainproject.ErrMemberAlreadyExists
	ErrMemberNotFound           = domainproject.ErrMemberNotFound
	ErrUserNotFound             = identity.ErrUserNotFound
	ErrLastOwner                = errors.New("project must keep an owner")
)
