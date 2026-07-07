package admin

import appadmin "github.com/yorukot/netstamp/internal/controller/application/admin"

type settingsBody struct {
	RegistrationEnabled *bool     `json:"registrationEnabled,omitempty"`
	BackendBaseURL      *string   `json:"backendBaseUrl,omitempty"`
	PublicWebBaseURL    *string   `json:"publicWebBaseUrl,omitempty"`
	SMTP                *smtpBody `json:"smtp,omitempty"`
}

type smtpBody struct {
	Host           *string `json:"host,omitempty"`
	Port           *int32  `json:"port,omitempty"`
	Username       *string `json:"username,omitempty"`
	Password       *string `json:"password,omitempty"`
	ClearPassword  bool    `json:"clearPassword,omitempty"`
	From           *string `json:"from,omitempty"`
	TLSMode        *string `json:"tlsMode,omitempty"`
	TimeoutSeconds *int32  `json:"timeoutSeconds,omitempty"`
}

type settingsResponseBody struct {
	RegistrationEnabled bool             `json:"registrationEnabled"`
	BackendBaseURL      string           `json:"backendBaseUrl"`
	PublicWebBaseURL    string           `json:"publicWebBaseUrl"`
	SMTP                smtpResponseBody `json:"smtp"`
}

type smtpResponseBody struct {
	Host           string `json:"host"`
	Port           int32  `json:"port"`
	Username       string `json:"username"`
	PasswordSet    bool   `json:"passwordSet"`
	From           string `json:"from"`
	TLSMode        string `json:"tlsMode"`
	TimeoutSeconds int32  `json:"timeoutSeconds"`
	Configured     bool   `json:"configured"`
}

func (b settingsBody) updateInput(userID string) appadmin.UpdateSettingsInput {
	input := appadmin.UpdateSettingsInput{
		CurrentUserID:       userID,
		RegistrationEnabled: b.RegistrationEnabled,
		BackendBaseURL:      b.BackendBaseURL,
		PublicWebBaseURL:    b.PublicWebBaseURL,
	}
	if b.SMTP != nil {
		input.SMTP = appadmin.UpdateSMTPSettingsInput{
			Host:           b.SMTP.Host,
			Port:           b.SMTP.Port,
			Username:       b.SMTP.Username,
			Password:       b.SMTP.Password,
			ClearPassword:  b.SMTP.ClearPassword,
			From:           b.SMTP.From,
			TLSMode:        b.SMTP.TLSMode,
			TimeoutSeconds: b.SMTP.TimeoutSeconds,
		}
	}
	return input
}

func settingsResponse(settings appadmin.Settings) settingsResponseBody {
	return settingsResponseBody{
		RegistrationEnabled: settings.RegistrationEnabled,
		BackendBaseURL:      settings.BackendBaseURL,
		PublicWebBaseURL:    settings.PublicWebBaseURL,
		SMTP: smtpResponseBody{
			Host:           settings.SMTP.Host,
			Port:           settings.SMTP.Port,
			Username:       settings.SMTP.Username,
			PasswordSet:    settings.SMTP.PasswordSet,
			From:           settings.SMTP.From,
			TLSMode:        settings.SMTP.TLSMode,
			TimeoutSeconds: settings.SMTP.TimeoutSeconds,
			Configured:     settings.SMTP.Host != "" && settings.SMTP.From != "",
		},
	}
}
