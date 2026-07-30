package admin

import domainsystem "github.com/yorukot/netstamp/internal/domain/system"

type SystemAdmin = domainsystem.AdminUser

type SystemAdminRevokeResult = domainsystem.AdminRevokeResult

type ManagedUser = domainsystem.ManagedUser

type DataExport = domainsystem.DataExport

type DataImportResult = domainsystem.DataImportResult

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

type ExportDataInput struct {
	CurrentUserID string
}

type ImportDataInput struct {
	CurrentUserID string
	Export        DataExport
}
