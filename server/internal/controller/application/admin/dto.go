package admin

import (
	"time"

	domainsystem "github.com/yorukot/netstamp/internal/domain/system"
)

type SystemAdmin = domainsystem.AdminUser

type SystemAdminRevokeResult = domainsystem.AdminRevokeResult

type ManagedUser = domainsystem.ManagedUser

type ListSystemAdminsInput struct {
	CurrentUserID string
}

type GrantSystemAdminInput struct {
	CurrentUserID string
	Email         string
}

type RevokeSystemAdminInput struct {
	CurrentUserID string
	UserID        string
}

type ListManagedUsersInput struct {
	CurrentUserID string
}

type UpdateStatusInput struct {
	CurrentUserID string
}

type UpdateStatus struct {
	CurrentVersion  string
	LatestVersion   *string
	UpdateAvailable bool
	ReleaseURL      *string
	PublishedAt     *time.Time
	LastCheckedAt   *time.Time
	CheckError      *string
}

type UpdateManagedUserInput struct {
	CurrentUserID string
	UserID        string
	Disabled      *bool
	SystemAdmin   *bool
}

type SetManagedUserPasswordInput struct {
	CurrentUserID string
	UserID        string
	Password      string
}

type ClearManagedUserPasswordInput struct {
	CurrentUserID string
	UserID        string
}
