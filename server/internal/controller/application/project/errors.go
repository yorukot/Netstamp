package project

import (
	"errors"
)

var (
	ErrProjectNotFound          = errors.New("project not found")
	ErrProjectSlugAlreadyExists = errors.New("project slug already exists")
	ErrForbidden                = errors.New("project action forbidden")
	ErrInvalidInput             = errors.New("project input invalid")
	ErrMemberAlreadyExists      = errors.New("project member already exists")
	ErrMemberNotFound           = errors.New("project member not found")
	ErrUserNotFound             = errors.New("user not found")
	ErrLastOwner                = errors.New("project must keep an owner")
)
