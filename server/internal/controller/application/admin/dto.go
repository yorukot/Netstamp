package admin

import domainsystem "github.com/yorukot/netstamp/internal/domain/system"

type Settings struct {
	RegistrationEnabled       bool
	EmailVerificationRequired bool
	ProjectCreationEnabled    bool
	CredentialChangesEnabled  bool
	SMTP                      SMTPSettings
	OIDC                      ExternalProviderSettings
	Google                    ExternalProviderSettings
	GitHub                    GitHubProviderSettings
	Tracking                  TrackingSettings
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
	ProjectCreationEnabled    bool
	CredentialChangesEnabled  bool
	SMTP                      SMTPSettings
	OIDC                      ExternalProviderSettings
	Google                    ExternalProviderSettings
	GitHub                    GitHubProviderSettings
	Tracking                  TrackingSettings
}

type ExternalProviderSettings struct {
	Enabled         bool
	IssuerURL       string
	ClientID        string
	ClientSecret    string
	ClientSecretSet bool
	DisplayName     string
	JITEnabled      bool
	AllowedDomains  string
}

type GitHubProviderSettings struct {
	ExternalProviderSettings
	AllowSignup bool
}

type TrackingSettings struct {
	GoogleTagID        string
	ClarityProjectID   string
	MetaPixelID        string
	PostHogKey         string
	PostHogHost        string
	PlausibleDomain    string
	PlausibleScriptURL string
	UmamiWebsiteID     string
	UmamiScriptURL     string
	ConsentMode        string
	ConsentCountries   string
}

type SystemAdmin = domainsystem.AdminUser

type SystemAdminRevokeResult = domainsystem.AdminRevokeResult

type ManagedUser = domainsystem.ManagedUser

type DataExport = domainsystem.DataExport

type DataImportResult = domainsystem.DataImportResult

type GetSettingsInput struct {
	CurrentUserID string
}

type TestSMTPInput struct {
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
	ProjectCreationEnabled    *bool
	CredentialChangesEnabled  *bool
	SMTP                      UpdateSMTPSettingsInput
	OIDC                      UpdateExternalProviderSettingsInput
	Google                    UpdateExternalProviderSettingsInput
	GitHub                    UpdateGitHubProviderSettingsInput
	Tracking                  *TrackingSettings
}

type UpdateExternalProviderSettingsInput struct {
	Enabled           *bool
	IssuerURL         *string
	ClientID          *string
	ClientSecret      *string
	ClearClientSecret bool
	DisplayName       *string
	JITEnabled        *bool
	AllowedDomains    *string
}

type UpdateGitHubProviderSettingsInput struct {
	UpdateExternalProviderSettingsInput
	AllowSignup *bool
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
