package system

import "time"

type AdminUser struct {
	ID              string
	Email           string
	DisplayName     string
	EmailVerifiedAt *time.Time
	DisabledAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	GrantedAt       time.Time
}

type ManagedUser struct {
	ID              string
	Email           string
	DisplayName     string
	EmailVerifiedAt *time.Time
	DisabledAt      *time.Time
	IsSystemAdmin   bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	GrantedAt       *time.Time
	HasPassword     bool
}

type AdminRevokeResult struct {
	AdminCount     int64
	TargetWasAdmin bool
	Revoked        bool
}
