package admin

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	appsettings "github.com/yorukot/netstamp/internal/controller/application/systemsettings"
	"github.com/yorukot/netstamp/internal/controller/transport/http/httpx"
)

type settingsService interface {
	GetAccess(context.Context, appsettings.GetAccessInput) (appsettings.AccessSettings, error)
	GetSMTP(context.Context, appsettings.GetSMTPInput) (appsettings.SMTPSettings, error)
	GetOIDC(context.Context, appsettings.GetOIDCInput) (appsettings.OIDCSettings, error)
	GetGoogle(context.Context, appsettings.GetGoogleInput) (appsettings.GoogleSettings, error)
	GetGitHub(context.Context, appsettings.GetGitHubInput) (appsettings.GitHubSettings, error)
	UpdateAccess(context.Context, appsettings.UpdateAccessInput) (appsettings.AccessSettings, error)
	UpdateSMTP(context.Context, appsettings.UpdateSMTPInput) (appsettings.SMTPSettings, error)
	UpdateOIDC(context.Context, appsettings.UpdateOIDCInput) (appsettings.OIDCSettings, error)
	UpdateGoogle(context.Context, appsettings.UpdateGoogleInput) (appsettings.GoogleSettings, error)
	UpdateGitHub(context.Context, appsettings.UpdateGitHubInput) (appsettings.GitHubSettings, error)
	ValidateOIDC(context.Context, appsettings.ValidateOIDCInput) error
	ValidateGoogle(context.Context, appsettings.ValidateGoogleInput) error
	ValidateGitHub(context.Context, appsettings.ValidateGitHubInput) error
	TestSMTP(context.Context, appsettings.TestSMTPInput) error
}

const settingsCacheControl = "private, no-store"

func (h *Handler) registerSettingsReadRoutes(r chi.Router) {
	r.Get("/admin/settings/access", h.handleGetAccessSettings)
	r.Get("/admin/settings/smtp", h.handleGetSMTPSettings)
	r.Get("/admin/settings/authentication-providers/oidc", h.handleGetOIDCSettings)
	r.Get("/admin/settings/authentication-providers/google", h.handleGetGoogleSettings)
	r.Get("/admin/settings/authentication-providers/github", h.handleGetGitHubSettings)
}

func (h *Handler) registerSettingsSensitiveRoutes(r chi.Router) {
	r.Patch("/admin/settings/access", h.handleUpdateAccessSettings)
	r.Patch("/admin/settings/smtp", h.handleUpdateSMTPSettings)
	r.Post("/admin/settings/smtp/test", h.handleTestSMTP)
	r.Patch("/admin/settings/authentication-providers/oidc", h.handleUpdateOIDCSettings)
	r.Post("/admin/settings/authentication-providers/oidc/validate", h.handleValidateOIDCSettings)
	r.Patch("/admin/settings/authentication-providers/google", h.handleUpdateGoogleSettings)
	r.Post("/admin/settings/authentication-providers/google/validate", h.handleValidateGoogleSettings)
	r.Patch("/admin/settings/authentication-providers/github", h.handleUpdateGitHubSettings)
	r.Post("/admin/settings/authentication-providers/github/validate", h.handleValidateGitHubSettings)
}

func (h *Handler) handleGetAccessSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	result, err := h.settings.GetAccess(r.Context(), appsettings.GetAccessInput{CurrentUserID: userID})
	if err != nil {
		writeSettingsProblem(w, r, err, "get access settings failed")
		return
	}
	writeSettings(w, accessSettingsResponse(result))
}

func (h *Handler) handleGetSMTPSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	result, err := h.settings.GetSMTP(r.Context(), appsettings.GetSMTPInput{CurrentUserID: userID})
	if err != nil {
		writeSettingsProblem(w, r, err, "get SMTP settings failed")
		return
	}
	writeSettings(w, smtpSettingsResponse(result))
}

func (h *Handler) handleGetOIDCSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	result, err := h.settings.GetOIDC(r.Context(), appsettings.GetOIDCInput{CurrentUserID: userID})
	if err != nil {
		writeSettingsProblem(w, r, err, "get OIDC settings failed")
		return
	}
	writeSettings(w, oidcSettingsResponse(result))
}

func (h *Handler) handleGetGoogleSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	result, err := h.settings.GetGoogle(r.Context(), appsettings.GetGoogleInput{CurrentUserID: userID})
	if err != nil {
		writeSettingsProblem(w, r, err, "get Google settings failed")
		return
	}
	writeSettings(w, googleSettingsResponse(result))
}

func (h *Handler) handleGetGitHubSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	result, err := h.settings.GetGitHub(r.Context(), appsettings.GetGitHubInput{CurrentUserID: userID})
	if err != nil {
		writeSettingsProblem(w, r, err, "get GitHub settings failed")
		return
	}
	writeSettings(w, githubSettingsResponse(result))
}

func (h *Handler) handleUpdateAccessSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	var body accessSettingsPatchBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	result, err := h.settings.UpdateAccess(r.Context(), appsettings.UpdateAccessInput{
		CurrentUserID:             userID,
		AccountCreationEnabled:    body.AccountCreationEnabled.Value,
		EmailVerificationRequired: body.EmailVerificationRequired.Value,
		ProjectCreationEnabled:    body.ProjectCreationEnabled.Value,
		CredentialChangesEnabled:  body.CredentialChangesEnabled.Value,
	})
	if err != nil {
		writeSettingsProblem(w, r, err, "update access settings failed")
		return
	}
	writeSettings(w, accessSettingsResponse(result))
}

func (h *Handler) handleUpdateSMTPSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	var body smtpSettingsPatchBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	result, err := h.settings.UpdateSMTP(r.Context(), appsettings.UpdateSMTPInput{
		CurrentUserID:  userID,
		Host:           body.Host.Value,
		Port:           body.Port.Value,
		Username:       body.Username.Value,
		Password:       optionalSecret(body.Password),
		From:           body.From.Value,
		TLSMode:        body.TLSMode.Value,
		TimeoutSeconds: body.TimeoutSeconds.Value,
	})
	if err != nil {
		writeSettingsProblem(w, r, err, "update SMTP settings failed")
		return
	}
	writeSettings(w, smtpSettingsResponse(result))
}

func (h *Handler) handleUpdateOIDCSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	body, ok := decodeOIDCSettingsPatch(w, r)
	if !ok {
		return
	}
	result, err := h.settings.UpdateOIDC(r.Context(), oidcInput(userID, body))
	if err != nil {
		writeSettingsProblem(w, r, err, "update OIDC settings failed")
		return
	}
	writeSettings(w, oidcSettingsResponse(result))
}

func (h *Handler) handleUpdateGoogleSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	body, ok := decodeGoogleSettingsPatch(w, r)
	if !ok {
		return
	}
	result, err := h.settings.UpdateGoogle(r.Context(), googleInput(userID, body))
	if err != nil {
		writeSettingsProblem(w, r, err, "update Google settings failed")
		return
	}
	writeSettings(w, googleSettingsResponse(result))
}

func (h *Handler) handleUpdateGitHubSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	body, ok := decodeGitHubSettingsPatch(w, r)
	if !ok {
		return
	}
	result, err := h.settings.UpdateGitHub(r.Context(), githubInput(userID, body))
	if err != nil {
		writeSettingsProblem(w, r, err, "update GitHub settings failed")
		return
	}
	writeSettings(w, githubSettingsResponse(result))
}

func (h *Handler) handleValidateOIDCSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	body, ok := decodeOIDCSettingsPatch(w, r)
	if !ok {
		return
	}
	if err := h.settings.ValidateOIDC(r.Context(), oidcInput(userID, body)); err != nil {
		writeSettingsProblem(w, r, err, "validate OIDC settings failed")
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) handleValidateGoogleSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	body, ok := decodeGoogleSettingsPatch(w, r)
	if !ok {
		return
	}
	if err := h.settings.ValidateGoogle(r.Context(), googleInput(userID, body)); err != nil {
		writeSettingsProblem(w, r, err, "validate Google settings failed")
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) handleValidateGitHubSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	body, ok := decodeGitHubSettingsPatch(w, r)
	if !ok {
		return
	}
	if err := h.settings.ValidateGitHub(r.Context(), githubInput(userID, body)); err != nil {
		writeSettingsProblem(w, r, err, "validate GitHub settings failed")
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) handleTestSMTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	if err := h.settings.TestSMTP(r.Context(), appsettings.TestSMTPInput{CurrentUserID: userID}); err != nil {
		writeSettingsProblem(w, r, err, "test SMTP settings failed")
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) settingsRequestUser(w http.ResponseWriter, r *http.Request) (string, bool) {
	if h.settings == nil {
		httpx.WriteProblem(w, r, httpx.InternalServerError("system settings service is unavailable"))
		return "", false
	}
	userID, err := currentUserID(r)
	if err != nil {
		httpx.WriteProblem(w, r, err)
		return "", false
	}
	return userID, true
}

func writeSettings[T any](w http.ResponseWriter, settings T) {
	w.Header().Del("ETag")
	w.Header().Set("Cache-Control", settingsCacheControl)
	httpx.WriteJSON(w, http.StatusOK, settingsEnvelope[T]{Settings: settings})
}

func writeSettingsProblem(w http.ResponseWriter, r *http.Request, err error, fallback string) {
	var transportError *httpx.Error
	if errors.As(err, &transportError) {
		httpx.WriteProblem(w, r, transportError)
		return
	}

	switch {
	case errors.Is(err, appsettings.ErrForbidden):
		httpx.WriteProblem(w, r, httpx.ForbiddenCode(httpx.CodeSystemAdminRequired, "system administrator access is required"))
	case errors.Is(err, appsettings.ErrInvalidInput):
		httpx.WriteProblem(w, r, invalidSettingsInputError(err))
	case errors.Is(err, appsettings.ErrProviderUnavailable):
		httpx.WriteProblem(w, r, httpx.ServiceUnavailable("provider configuration validation is unavailable"))
	case errors.Is(err, appsettings.ErrSMTPTestFailed):
		httpx.WriteProblem(w, r, httpx.ServiceUnavailable("SMTP test delivery failed"))
	default:
		httpx.WriteProblem(w, r, httpx.InternalServerError(fallback))
	}
}

func decodeOIDCSettingsPatch(w http.ResponseWriter, r *http.Request) (oidcProviderSettingsPatchBody, bool) {
	var body oidcProviderSettingsPatchBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return oidcProviderSettingsPatchBody{}, false
	}
	return body, true
}

func decodeGoogleSettingsPatch(w http.ResponseWriter, r *http.Request) (googleProviderSettingsPatchBody, bool) {
	var body googleProviderSettingsPatchBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return googleProviderSettingsPatchBody{}, false
	}
	return body, true
}

func decodeGitHubSettingsPatch(w http.ResponseWriter, r *http.Request) (githubProviderSettingsPatchBody, bool) {
	var body githubProviderSettingsPatchBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return githubProviderSettingsPatchBody{}, false
	}
	return body, true
}

func oidcInput(userID string, body oidcProviderSettingsPatchBody) appsettings.UpdateOIDCInput {
	return appsettings.UpdateOIDCInput{
		CurrentUserID: userID, Enabled: body.Enabled.Value,
		IssuerURL: body.IssuerURL.Value, ClientID: body.ClientID.Value, ClientSecret: optionalSecret(body.ClientSecret),
		DisplayName: body.DisplayName.Value, JITEnabled: body.JITEnabled.Value,
	}
}

func googleInput(userID string, body googleProviderSettingsPatchBody) appsettings.UpdateGoogleInput {
	return appsettings.UpdateGoogleInput{
		CurrentUserID: userID, Enabled: body.Enabled.Value,
		ClientID: body.ClientID.Value, ClientSecret: optionalSecret(body.ClientSecret),
		DisplayName: body.DisplayName.Value, JITEnabled: body.JITEnabled.Value, AllowedDomains: body.AllowedDomains.Value,
	}
}

func githubInput(userID string, body githubProviderSettingsPatchBody) appsettings.UpdateGitHubInput {
	return appsettings.UpdateGitHubInput{
		CurrentUserID: userID, Enabled: body.Enabled.Value,
		ClientID: body.ClientID.Value, ClientSecret: optionalSecret(body.ClientSecret),
		DisplayName: body.DisplayName.Value, JITEnabled: body.JITEnabled.Value, AllowSignup: body.AllowSignup.Value,
	}
}

func optionalSecret(value optionalNullableString) appsettings.OptionalSecret {
	return appsettings.OptionalSecret{Present: value.Present, Value: value.Value}
}

func accessSettingsResponse(value appsettings.AccessSettings) accessSettingsResponseBody {
	return accessSettingsResponseBody{
		AccountCreationEnabled: value.AccountCreationEnabled, EmailVerificationRequired: value.EmailVerificationRequired,
		ProjectCreationEnabled: value.ProjectCreationEnabled, CredentialChangesEnabled: value.CredentialChangesEnabled,
	}
}

func smtpSettingsResponse(value appsettings.SMTPSettings) smtpSettingsResponseBody {
	return smtpSettingsResponseBody{
		Host: value.Host, Port: value.Port, Username: value.Username, PasswordSet: value.PasswordSet,
		From: value.From, TLSMode: value.TLSMode, TimeoutSeconds: value.TimeoutSeconds, Configured: value.Configured,
	}
}

func providerSettingsResponse(
	enabled bool,
	clientID string,
	clientSecretSet bool,
	displayName string,
	jitEnabled bool,
	callbackURL *string,
) providerSettingsResponseBody {
	return providerSettingsResponseBody{
		Enabled: enabled, ClientID: clientID, ClientSecretSet: clientSecretSet,
		DisplayName: displayName, JITEnabled: jitEnabled, CallbackURL: callbackURL,
	}
}

func oidcSettingsResponse(value appsettings.OIDCSettings) oidcProviderSettingsResponseBody {
	return oidcProviderSettingsResponseBody{
		providerSettingsResponseBody: providerSettingsResponse(
			value.Enabled, value.ClientID, value.ClientSecretSet, value.DisplayName, value.JITEnabled, value.CallbackURL,
		),
		IssuerURL: value.IssuerURL,
	}
}

func googleSettingsResponse(value appsettings.GoogleSettings) googleProviderSettingsResponseBody {
	domains := make([]string, len(value.AllowedDomains))
	copy(domains, value.AllowedDomains)
	return googleProviderSettingsResponseBody{
		providerSettingsResponseBody: providerSettingsResponse(
			value.Enabled, value.ClientID, value.ClientSecretSet, value.DisplayName, value.JITEnabled, value.CallbackURL,
		),
		AllowedDomains: domains,
	}
}

func githubSettingsResponse(value appsettings.GitHubSettings) githubProviderSettingsResponseBody {
	return githubProviderSettingsResponseBody{
		providerSettingsResponseBody: providerSettingsResponse(
			value.Enabled, value.ClientID, value.ClientSecretSet, value.DisplayName, value.JITEnabled, value.CallbackURL,
		),
		AllowSignup: value.AllowSignup,
	}
}
