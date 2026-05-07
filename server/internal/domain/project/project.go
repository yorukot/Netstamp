package project

import "time"

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
