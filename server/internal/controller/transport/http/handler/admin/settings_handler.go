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
	GetAccess(context.Context, appsettings.GetAccessInput) (appsettings.Versioned[appsettings.AccessSettings], error)
	GetSMTP(context.Context, appsettings.GetSMTPInput) (appsettings.Versioned[appsettings.SMTPSettings], error)
	GetOIDC(context.Context, appsettings.GetOIDCInput) (appsettings.Versioned[appsettings.OIDCSettings], error)
	GetGoogle(context.Context, appsettings.GetGoogleInput) (appsettings.Versioned[appsettings.GoogleSettings], error)
	GetGitHub(context.Context, appsettings.GetGitHubInput) (appsettings.Versioned[appsettings.GitHubSettings], error)
	UpdateAccess(context.Context, appsettings.UpdateAccessInput) (appsettings.Versioned[appsettings.AccessSettings], error)
	UpdateSMTP(context.Context, appsettings.UpdateSMTPInput) (appsettings.Versioned[appsettings.SMTPSettings], error)
	UpdateOIDC(context.Context, appsettings.UpdateOIDCInput) (appsettings.Versioned[appsettings.OIDCSettings], error)
	UpdateGoogle(context.Context, appsettings.UpdateGoogleInput) (appsettings.Versioned[appsettings.GoogleSettings], error)
	UpdateGitHub(context.Context, appsettings.UpdateGitHubInput) (appsettings.Versioned[appsettings.GitHubSettings], error)
	ValidateOIDC(context.Context, appsettings.ValidateOIDCInput) error
	ValidateGoogle(context.Context, appsettings.ValidateGoogleInput) error
	ValidateGitHub(context.Context, appsettings.ValidateGitHubInput) error
	TestSMTP(context.Context, appsettings.TestSMTPInput) error
}

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
		writeSettingsProblem(w, r, appsettings.ResourceAccess, err, "get access settings failed")
		return
	}
	writeVersionedSettings(w, appsettings.ResourceAccess, result.Revision, accessSettingsResponse(result.Value))
}

func (h *Handler) handleGetSMTPSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	result, err := h.settings.GetSMTP(r.Context(), appsettings.GetSMTPInput{CurrentUserID: userID})
	if err != nil {
		writeSettingsProblem(w, r, appsettings.ResourceSMTP, err, "get SMTP settings failed")
		return
	}
	writeVersionedSettings(w, appsettings.ResourceSMTP, result.Revision, smtpSettingsResponse(result.Value))
}

func (h *Handler) handleGetOIDCSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	result, err := h.settings.GetOIDC(r.Context(), appsettings.GetOIDCInput{CurrentUserID: userID})
	if err != nil {
		writeSettingsProblem(w, r, appsettings.ResourceOIDC, err, "get OIDC settings failed")
		return
	}
	writeVersionedSettings(w, appsettings.ResourceOIDC, result.Revision, oidcSettingsResponse(result.Value))
}

func (h *Handler) handleGetGoogleSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	result, err := h.settings.GetGoogle(r.Context(), appsettings.GetGoogleInput{CurrentUserID: userID})
	if err != nil {
		writeSettingsProblem(w, r, appsettings.ResourceGoogle, err, "get Google settings failed")
		return
	}
	writeVersionedSettings(w, appsettings.ResourceGoogle, result.Revision, googleSettingsResponse(result.Value))
}

func (h *Handler) handleGetGitHubSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	result, err := h.settings.GetGitHub(r.Context(), appsettings.GetGitHubInput{CurrentUserID: userID})
	if err != nil {
		writeSettingsProblem(w, r, appsettings.ResourceGitHub, err, "get GitHub settings failed")
		return
	}
	writeVersionedSettings(w, appsettings.ResourceGitHub, result.Revision, githubSettingsResponse(result.Value))
}

func (h *Handler) handleUpdateAccessSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	revision, ok := h.accessExpectedRevision(r.Context(), w, r, userID)
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
		ExpectedRevision:          revision,
		AccountCreationEnabled:    body.AccountCreationEnabled.Value,
		EmailVerificationRequired: body.EmailVerificationRequired.Value,
		ProjectCreationEnabled:    body.ProjectCreationEnabled.Value,
		CredentialChangesEnabled:  body.CredentialChangesEnabled.Value,
	})
	if err != nil {
		writeSettingsProblem(w, r, appsettings.ResourceAccess, err, "update access settings failed")
		return
	}
	writeVersionedSettings(w, appsettings.ResourceAccess, result.Revision, accessSettingsResponse(result.Value))
}

func (h *Handler) handleUpdateSMTPSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	revision, ok := h.smtpExpectedRevision(r.Context(), w, r, userID)
	if !ok {
		return
	}

	var body smtpSettingsPatchBody
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteProblem(w, r, err)
		return
	}
	result, err := h.settings.UpdateSMTP(r.Context(), appsettings.UpdateSMTPInput{
		CurrentUserID:    userID,
		ExpectedRevision: revision,
		Host:             body.Host.Value,
		Port:             body.Port.Value,
		Username:         body.Username.Value,
		Password:         optionalSecret(body.Password),
		From:             body.From.Value,
		TLSMode:          body.TLSMode.Value,
		TimeoutSeconds:   body.TimeoutSeconds.Value,
	})
	if err != nil {
		writeSettingsProblem(w, r, appsettings.ResourceSMTP, err, "update SMTP settings failed")
		return
	}
	writeVersionedSettings(w, appsettings.ResourceSMTP, result.Revision, smtpSettingsResponse(result.Value))
}

func (h *Handler) handleUpdateOIDCSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	revision, ok := h.oidcExpectedRevision(r.Context(), w, r, userID)
	if !ok {
		return
	}
	body, ok := decodeOIDCSettingsPatch(w, r)
	if !ok {
		return
	}
	result, err := h.settings.UpdateOIDC(r.Context(), oidcInput(userID, revision, body))
	if err != nil {
		writeSettingsProblem(w, r, appsettings.ResourceOIDC, err, "update OIDC settings failed")
		return
	}
	writeVersionedSettings(w, appsettings.ResourceOIDC, result.Revision, oidcSettingsResponse(result.Value))
}

func (h *Handler) handleUpdateGoogleSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	revision, ok := h.googleExpectedRevision(r.Context(), w, r, userID)
	if !ok {
		return
	}
	body, ok := decodeGoogleSettingsPatch(w, r)
	if !ok {
		return
	}
	result, err := h.settings.UpdateGoogle(r.Context(), googleInput(userID, revision, body))
	if err != nil {
		writeSettingsProblem(w, r, appsettings.ResourceGoogle, err, "update Google settings failed")
		return
	}
	writeVersionedSettings(w, appsettings.ResourceGoogle, result.Revision, googleSettingsResponse(result.Value))
}

func (h *Handler) handleUpdateGitHubSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	revision, ok := h.githubExpectedRevision(r.Context(), w, r, userID)
	if !ok {
		return
	}
	body, ok := decodeGitHubSettingsPatch(w, r)
	if !ok {
		return
	}
	result, err := h.settings.UpdateGitHub(r.Context(), githubInput(userID, revision, body))
	if err != nil {
		writeSettingsProblem(w, r, appsettings.ResourceGitHub, err, "update GitHub settings failed")
		return
	}
	writeVersionedSettings(w, appsettings.ResourceGitHub, result.Revision, githubSettingsResponse(result.Value))
}

func (h *Handler) handleValidateOIDCSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	revision, ok := h.oidcExpectedRevision(r.Context(), w, r, userID)
	if !ok {
		return
	}
	body, ok := decodeOIDCSettingsPatch(w, r)
	if !ok {
		return
	}
	if err := h.settings.ValidateOIDC(r.Context(), oidcInput(userID, revision, body)); err != nil {
		writeSettingsProblem(w, r, appsettings.ResourceOIDC, err, "validate OIDC settings failed")
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) handleValidateGoogleSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	revision, ok := h.googleExpectedRevision(r.Context(), w, r, userID)
	if !ok {
		return
	}
	body, ok := decodeGoogleSettingsPatch(w, r)
	if !ok {
		return
	}
	if err := h.settings.ValidateGoogle(r.Context(), googleInput(userID, revision, body)); err != nil {
		writeSettingsProblem(w, r, appsettings.ResourceGoogle, err, "validate Google settings failed")
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) handleValidateGitHubSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	revision, ok := h.githubExpectedRevision(r.Context(), w, r, userID)
	if !ok {
		return
	}
	body, ok := decodeGitHubSettingsPatch(w, r)
	if !ok {
		return
	}
	if err := h.settings.ValidateGitHub(r.Context(), githubInput(userID, revision, body)); err != nil {
		writeSettingsProblem(w, r, appsettings.ResourceGitHub, err, "validate GitHub settings failed")
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) handleTestSMTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.settingsRequestUser(w, r)
	if !ok {
		return
	}
	revision, ok := h.smtpExpectedRevision(r.Context(), w, r, userID)
	if !ok {
		return
	}
	if err := h.settings.TestSMTP(r.Context(), appsettings.TestSMTPInput{CurrentUserID: userID, ExpectedRevision: revision}); err != nil {
		writeSettingsProblem(w, r, appsettings.ResourceSMTP, err, "test SMTP settings failed")
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

func writeVersionedSettings[T any](w http.ResponseWriter, resource appsettings.Resource, revision int64, settings T) {
	setVersionedSettingsHeaders(w, string(resource), revision)
	httpx.WriteJSON(w, http.StatusOK, settingsEnvelope[T]{Settings: settings})
}

func writeSettingsProblem(w http.ResponseWriter, r *http.Request, resource appsettings.Resource, err error, fallback string) {
	var requestConflict *settingsVersionConflictError
	if errors.As(err, &requestConflict) {
		writeSettingsVersionConflict(w, r, string(resource), requestConflict.currentRevision)
		return
	}

	var versionConflict *appsettings.VersionConflictError
	if errors.As(err, &versionConflict) {
		writeSettingsVersionConflict(w, r, string(resource), versionConflict.Current)
		return
	}

	var transportError *httpx.Error
	if errors.As(err, &transportError) {
		httpx.WriteProblem(w, r, transportError)
		return
	}

	switch {
	case errors.Is(err, appsettings.ErrForbidden):
		httpx.WriteProblem(w, r, httpx.ForbiddenCode(httpx.CodeSystemAdminRequired, "system administrator access is required"))
	case errors.Is(err, appsettings.ErrPreconditionRequired):
		httpx.WriteProblem(w, r, httpx.NewErrorCode(http.StatusPreconditionRequired, httpx.CodePreconditionRequired, "If-Match header is required"))
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

func writeSettingsVersionConflict(w http.ResponseWriter, r *http.Request, resource string, revision int64) {
	setVersionedSettingsHeaders(w, resource, revision)
	httpx.WriteProblem(w, r, httpx.NewErrorCode(http.StatusPreconditionFailed, httpx.CodeSettingsVersionConflict, "settings have changed; refresh and retry"))
}

func (h *Handler) accessExpectedRevision(ctx context.Context, w http.ResponseWriter, r *http.Request, userID string) (int64, bool) {
	return h.expectedSettingsRevision(w, r, appsettings.ResourceAccess, func() (int64, error) {
		result, err := h.settings.GetAccess(ctx, appsettings.GetAccessInput{CurrentUserID: userID})
		return result.Revision, err
	})
}

func (h *Handler) smtpExpectedRevision(ctx context.Context, w http.ResponseWriter, r *http.Request, userID string) (int64, bool) {
	return h.expectedSettingsRevision(w, r, appsettings.ResourceSMTP, func() (int64, error) {
		result, err := h.settings.GetSMTP(ctx, appsettings.GetSMTPInput{CurrentUserID: userID})
		return result.Revision, err
	})
}

func (h *Handler) oidcExpectedRevision(ctx context.Context, w http.ResponseWriter, r *http.Request, userID string) (int64, bool) {
	return h.expectedSettingsRevision(w, r, appsettings.ResourceOIDC, func() (int64, error) {
		result, err := h.settings.GetOIDC(ctx, appsettings.GetOIDCInput{CurrentUserID: userID})
		return result.Revision, err
	})
}

func (h *Handler) googleExpectedRevision(ctx context.Context, w http.ResponseWriter, r *http.Request, userID string) (int64, bool) {
	return h.expectedSettingsRevision(w, r, appsettings.ResourceGoogle, func() (int64, error) {
		result, err := h.settings.GetGoogle(ctx, appsettings.GetGoogleInput{CurrentUserID: userID})
		return result.Revision, err
	})
}

func (h *Handler) githubExpectedRevision(ctx context.Context, w http.ResponseWriter, r *http.Request, userID string) (int64, bool) {
	return h.expectedSettingsRevision(w, r, appsettings.ResourceGitHub, func() (int64, error) {
		result, err := h.settings.GetGitHub(ctx, appsettings.GetGitHubInput{CurrentUserID: userID})
		return result.Revision, err
	})
}

func (h *Handler) expectedSettingsRevision(
	w http.ResponseWriter,
	r *http.Request,
	resource appsettings.Resource,
	currentRevision func() (int64, error),
) (int64, bool) {
	revision, err := resolveSettingsIfMatch(r, string(resource), currentRevision)
	if err != nil {
		writeSettingsProblem(w, r, resource, err, "read current settings version failed")
		return 0, false
	}
	return revision, true
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

func oidcInput(userID string, revision int64, body oidcProviderSettingsPatchBody) appsettings.UpdateOIDCInput {
	return appsettings.UpdateOIDCInput{
		CurrentUserID: userID, ExpectedRevision: revision, Enabled: body.Enabled.Value,
		IssuerURL: body.IssuerURL.Value, ClientID: body.ClientID.Value, ClientSecret: optionalSecret(body.ClientSecret),
		DisplayName: body.DisplayName.Value, JITEnabled: body.JITEnabled.Value,
	}
}

func googleInput(userID string, revision int64, body googleProviderSettingsPatchBody) appsettings.UpdateGoogleInput {
	return appsettings.UpdateGoogleInput{
		CurrentUserID: userID, ExpectedRevision: revision, Enabled: body.Enabled.Value,
		ClientID: body.ClientID.Value, ClientSecret: optionalSecret(body.ClientSecret),
		DisplayName: body.DisplayName.Value, JITEnabled: body.JITEnabled.Value, AllowedDomains: body.AllowedDomains.Value,
	}
}

func githubInput(userID string, revision int64, body githubProviderSettingsPatchBody) appsettings.UpdateGitHubInput {
	return appsettings.UpdateGitHubInput{
		CurrentUserID: userID, ExpectedRevision: revision, Enabled: body.Enabled.Value,
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
