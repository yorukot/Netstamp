package systemsettings

const (
	keyAccountCreationEnabled    = "auth.registration_enabled"
	keyEmailVerificationRequired = "auth.email_verification_required"
	keyProjectCreationEnabled    = "access.project_creation_enabled"
	keyCredentialChangesEnabled  = "access.credential_changes_enabled"
	keyUpdateCheckEnabled        = "updates.check_enabled"

	keySMTPHost           = "smtp.host"
	keySMTPPort           = "smtp.port"
	keySMTPUsername       = "smtp.username"
	keySMTPPassword       = "smtp.password"
	keySMTPFrom           = "smtp.from"
	keySMTPTLSMode        = "smtp.tls_mode"
	keySMTPTimeoutSeconds = "smtp.timeout_seconds"

	keyOIDCSettings       = "auth.provider.oidc"
	keyOIDCClientSecret   = "auth.provider.oidc.client_secret" //nolint:gosec // Setting key, not credential material.
	keyGoogleSettings     = "auth.provider.google"
	keyGoogleClientSecret = "auth.provider.google.client_secret" //nolint:gosec // Setting key, not credential material.
	keyGitHubSettings     = "auth.provider.github"
	keyGitHubClientSecret = "auth.provider.github.client_secret" //nolint:gosec // Setting key, not credential material.

	auditActionUpdate = "update"
	auditActionClear  = "clear"
)

var (
	accessKeys = []string{
		keyAccountCreationEnabled,
		keyEmailVerificationRequired,
		keyProjectCreationEnabled,
		keyCredentialChangesEnabled,
	}
	smtpKeys = []string{
		keySMTPHost,
		keySMTPPort,
		keySMTPUsername,
		keySMTPPassword,
		keySMTPFrom,
		keySMTPTLSMode,
		keySMTPTimeoutSeconds,
	}
	oidcKeys    = []string{keyOIDCSettings, keyOIDCClientSecret}
	googleKeys  = []string{keyGoogleSettings, keyGoogleClientSecret}
	gitHubKeys  = []string{keyGitHubSettings, keyGitHubClientSecret}
	updatesKeys = []string{keyUpdateCheckEnabled}
)
