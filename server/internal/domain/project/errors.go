package project

import "errors"

var (
	ErrProjectNotFound          = errors.New("project not found")
	ErrProjectSlugAlreadyExists = errors.New("project slug already exists")
	ErrMemberAlreadyExists      = errors.New("project member already exists")
	ErrMemberNotFound           = errors.New("project member not found")
	ErrInviteAlreadyExists      = errors.New("project invite already exists")
	ErrInviteNotFound           = errors.New("project invite not found")
)
