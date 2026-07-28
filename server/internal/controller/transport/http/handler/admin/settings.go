package admin

import appadmin "github.com/yorukot/netstamp/internal/controller/application/admin"

type settingsBody struct {
	Access                  *accessBody    `json:"access,omitempty"`
	SMTP                    *smtpBody      `json:"smtp,omitempty"`
	AuthenticationProviders *providersBody `json:"authenticationProviders,omitempty"`
	Tracking                *trackingBody  `json:"tracking,omitempty"`
}

type accessBody struct {
	RegistrationEnabled       *bool `json:"registrationEnabled,omitempty"`
	EmailVerificationRequired *bool `json:"emailVerificationRequired,omitempty"`
	ProjectCreationEnabled    *bool `json:"projectCreationEnabled,omitempty"`
	CredentialChangesEnabled  *bool `json:"credentialChangesEnabled,omitempty"`
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

type providersBody struct {
	OIDC   *providerBody       `json:"oidc,omitempty"`
	Google *providerBody       `json:"google,omitempty"`
	GitHub *githubProviderBody `json:"github,omitempty"`
}
type providerBody struct {
	Enabled           *bool   `json:"enabled,omitempty"`
	IssuerURL         *string `json:"issuerUrl,omitempty"`
	ClientID          *string `json:"clientId,omitempty"`
	ClientSecret      *string `json:"clientSecret,omitempty"`
	ClearClientSecret bool    `json:"clearClientSecret,omitempty"`
	DisplayName       *string `json:"displayName,omitempty"`
	JITEnabled        *bool   `json:"jitEnabled,omitempty"`
	AllowedDomains    *string `json:"allowedDomains,omitempty"`
}
type githubProviderBody struct {
	providerBody
	AllowSignup *bool `json:"allowSignup,omitempty"`
}

type trackingBody struct {
	GoogleTagID        string `json:"googleTagId"`
	ClarityProjectID   string `json:"clarityProjectId"`
	MetaPixelID        string `json:"metaPixelId"`
	PostHogKey         string `json:"postHogKey"`
	PostHogHost        string `json:"postHogHost"`
	PlausibleDomain    string `json:"plausibleDomain"`
	PlausibleScriptURL string `json:"plausibleScriptUrl"`
	UmamiWebsiteID     string `json:"umamiWebsiteId"`
	UmamiScriptURL     string `json:"umamiScriptUrl"`
	ConsentMode        string `json:"consentMode"`
	ConsentCountries   string `json:"consentCountries"`
}

func providerInput(body *providerBody) appadmin.UpdateExternalProviderSettingsInput {
	if body == nil {
		return appadmin.UpdateExternalProviderSettingsInput{}
	}
	return appadmin.UpdateExternalProviderSettingsInput{Enabled: body.Enabled, IssuerURL: body.IssuerURL, ClientID: body.ClientID, ClientSecret: body.ClientSecret, ClearClientSecret: body.ClearClientSecret, DisplayName: body.DisplayName, JITEnabled: body.JITEnabled, AllowedDomains: body.AllowedDomains}
}

func (b settingsBody) updateInput(userID string) appadmin.UpdateSettingsInput {
	input := appadmin.UpdateSettingsInput{CurrentUserID: userID}
	if b.Access != nil {
		input.RegistrationEnabled = b.Access.RegistrationEnabled
		input.EmailVerificationRequired = b.Access.EmailVerificationRequired
		input.ProjectCreationEnabled = b.Access.ProjectCreationEnabled
		input.CredentialChangesEnabled = b.Access.CredentialChangesEnabled
	}
	if b.SMTP != nil {
		input.SMTP = appadmin.UpdateSMTPSettingsInput{Host: b.SMTP.Host, Port: b.SMTP.Port, Username: b.SMTP.Username, Password: b.SMTP.Password, ClearPassword: b.SMTP.ClearPassword, From: b.SMTP.From, TLSMode: b.SMTP.TLSMode, TimeoutSeconds: b.SMTP.TimeoutSeconds}
	}
	if b.AuthenticationProviders != nil {
		input.OIDC = providerInput(b.AuthenticationProviders.OIDC)
		input.Google = providerInput(b.AuthenticationProviders.Google)
		if b.AuthenticationProviders.GitHub != nil {
			input.GitHub = appadmin.UpdateGitHubProviderSettingsInput{UpdateExternalProviderSettingsInput: providerInput(&b.AuthenticationProviders.GitHub.providerBody), AllowSignup: b.AuthenticationProviders.GitHub.AllowSignup}
		}
	}
	if b.Tracking != nil {
		input.Tracking = &appadmin.TrackingSettings{GoogleTagID: b.Tracking.GoogleTagID, ClarityProjectID: b.Tracking.ClarityProjectID, MetaPixelID: b.Tracking.MetaPixelID, PostHogKey: b.Tracking.PostHogKey, PostHogHost: b.Tracking.PostHogHost, PlausibleDomain: b.Tracking.PlausibleDomain, PlausibleScriptURL: b.Tracking.PlausibleScriptURL, UmamiWebsiteID: b.Tracking.UmamiWebsiteID, UmamiScriptURL: b.Tracking.UmamiScriptURL, ConsentMode: b.Tracking.ConsentMode, ConsentCountries: b.Tracking.ConsentCountries}
	}
	return input
}

func providerResponse(value appadmin.ExternalProviderSettings) map[string]any {
	return map[string]any{"enabled": value.Enabled, "issuerUrl": value.IssuerURL, "clientId": value.ClientID, "clientSecretSet": value.ClientSecretSet, "displayName": value.DisplayName, "jitEnabled": value.JITEnabled, "allowedDomains": value.AllowedDomains}
}

func settingsResponse(settings appadmin.Settings) map[string]any {
	github := providerResponse(settings.GitHub.ExternalProviderSettings)
	github["allowSignup"] = settings.GitHub.AllowSignup
	return map[string]any{
		"access":                  map[string]any{"registrationEnabled": settings.RegistrationEnabled, "emailVerificationRequired": settings.EmailVerificationRequired, "projectCreationEnabled": settings.ProjectCreationEnabled, "credentialChangesEnabled": settings.CredentialChangesEnabled},
		"smtp":                    map[string]any{"host": settings.SMTP.Host, "port": settings.SMTP.Port, "username": settings.SMTP.Username, "passwordSet": settings.SMTP.PasswordSet, "from": settings.SMTP.From, "tlsMode": settings.SMTP.TLSMode, "timeoutSeconds": settings.SMTP.TimeoutSeconds, "configured": settings.SMTP.Host != "" && settings.SMTP.From != ""},
		"authenticationProviders": map[string]any{"oidc": providerResponse(settings.OIDC), "google": providerResponse(settings.Google), "github": github},
		"tracking":                trackingBody{GoogleTagID: settings.Tracking.GoogleTagID, ClarityProjectID: settings.Tracking.ClarityProjectID, MetaPixelID: settings.Tracking.MetaPixelID, PostHogKey: settings.Tracking.PostHogKey, PostHogHost: settings.Tracking.PostHogHost, PlausibleDomain: settings.Tracking.PlausibleDomain, PlausibleScriptURL: settings.Tracking.PlausibleScriptURL, UmamiWebsiteID: settings.Tracking.UmamiWebsiteID, UmamiScriptURL: settings.Tracking.UmamiScriptURL, ConsentMode: settings.Tracking.ConsentMode, ConsentCountries: settings.Tracking.ConsentCountries},
	}
}
