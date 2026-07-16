package admin

import domainsystem "github.com/yorukot/netstamp/internal/domain/system"

type Settings struct {
	RegistrationEnabled       bool
	EmailVerificationRequired bool
	BackendBaseURL            string
	PublicWebBaseURL          string
	SMTP                      SMTPSettings
}

type SMTPSettings struct {
	Host           string
	Port           int32
	Username       string
	Password       string
	PasswordSet    bool
	From           string
	TLSMode        string
	TimeoutSeconds int32
}

type Defaults struct {
	RegistrationEnabled       bool
	EmailVerificationRequired bool
	BackendBaseURL            string
	PublicWebBaseURL          string
	SMTP                      SMTPSettings
}

type SystemAdmin = domainsystem.AdminUser

type SystemAdminRevokeResult = domainsystem.AdminRevokeResult

type ManagedUser = domainsystem.ManagedUser

type DataExport = domainsystem.DataExport

type DataImportResult = domainsystem.DataImportResult

type GetSettingsInput struct {
	CurrentUserID string
}

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

type UpdateSettingsInput struct {
	CurrentUserID             string
	RegistrationEnabled       *bool
	EmailVerificationRequired *bool
	BackendBaseURL            *string
	PublicWebBaseURL          *string
	SMTP                      UpdateSMTPSettingsInput
}

type UpdateSMTPSettingsInput struct {
	Host           *string
	Port           *int32
	Username       *string
	Password       *string
	ClearPassword  bool
	From           *string
	TLSMode        *string
	TimeoutSeconds *int32
}

type StoredSetting = domainsystem.Setting
