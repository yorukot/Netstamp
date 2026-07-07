package system

import "time"

type AdminUser struct {
	ID              string
	Email           string
	DisplayName     string
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	GrantedAt       time.Time
}

type AdminRevokeResult struct {
	AdminCount     int64
	TargetWasAdmin bool
	Revoked        bool
}
