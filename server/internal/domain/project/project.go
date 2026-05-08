package project

import (
	"errors"
	"time"
)

var (
	ErrProjectNotFound          = errors.New("project not found")
	ErrProjectSlugAlreadyExists = errors.New("project slug already exists")
	ErrMemberAlreadyExists      = errors.New("project member already exists")
	ErrMemberNotFound           = errors.New("project member not found")
)

type Role string

const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

type Project struct {
	ID              string
	Name            string
	Slug            string
	CreatedByUserID string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type Member struct {
	ID        string
	ProjectID string
	UserID    string
	Email     string
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateProjectStorageInput struct {
	Name            string
	Slug            string
	CreatedByUserID string
}

type UpdateProjectStorageInput struct {
	ProjectID string
	Name      string
	Slug      string
}

type AddMemberStorageInput struct {
	ProjectID string
	UserID    string
	Role      Role
}

type UpdateMemberRoleStorageInput struct {
	ProjectID string
	UserID    string
	Role      Role
}
