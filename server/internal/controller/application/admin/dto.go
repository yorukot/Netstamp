package admin

import domainsystem "github.com/yorukot/netstamp/internal/domain/system"

type Settings struct {
	RegistrationEnabled bool
	BackendBaseURL      string
	PublicWebBaseURL    string
	SMTP                SMTPSettings
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
	RegistrationEnabled bool
	BackendBaseURL      string
	PublicWebBaseURL    string
	SMTP                SMTPSettings
}

type GetSettingsInput struct {
	CurrentUserID string
}

type UpdateSettingsInput struct {
	CurrentUserID       string
	RegistrationEnabled *bool
	BackendBaseURL      *string
	PublicWebBaseURL    *string
	SMTP                UpdateSMTPSettingsInput
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
