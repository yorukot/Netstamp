package httpx

const (
	CodeRouteNotFound = "ROUTE_NOT_FOUND"
	CodeReadOnly      = "READ_ONLY"
)

const (
	CodeAuthMissingSession                = "AUTH_MISSING_SESSION"
	CodeAuthInvalidSession                = "AUTH_INVALID_SESSION"
	CodeAuthInvalidCSRF                   = "AUTH_INVALID_CSRF"
	CodeAuthInvalidCredentials            = "AUTH_INVALID_CREDENTIALS" // #nosec G101 -- public problem code, not credential material.
	CodeAuthEmailVerificationRequired     = "AUTH_EMAIL_VERIFICATION_REQUIRED"
	CodeAuthRegistrationDisabled          = "AUTH_REGISTRATION_DISABLED"
	CodeAuthEmailVerificationUnavailable  = "AUTH_EMAIL_VERIFICATION_UNAVAILABLE"
	CodeAuthPasswordResetUnavailable      = "AUTH_PASSWORD_RESET_UNAVAILABLE"
	CodeAuthEmailVerificationTokenInvalid = "AUTH_EMAIL_VERIFICATION_TOKEN_INVALID" // #nosec G101 -- public problem code, not token material.
	CodeAuthPasswordResetTokenInvalid     = "AUTH_PASSWORD_RESET_TOKEN_INVALID"
	CodeAuthSessionNotFound               = "AUTH_SESSION_NOT_FOUND"
	CodeAuthInvalidAPIToken               = "AUTH_INVALID_API_TOKEN" // #nosec G101 -- public problem code, not a credential.
	CodeAuthInsufficientScope             = "AUTH_INSUFFICIENT_SCOPE"
	CodeAuthSudoRequired                  = "AUTH_SUDO_REQUIRED"
	CodeAuthOIDCUnavailable               = "AUTH_OIDC_UNAVAILABLE"
	CodeAuthOIDCCallbackInvalid           = "AUTH_OIDC_CALLBACK_INVALID"
	CodeAuthIdentityConflict              = "AUTH_IDENTITY_CONFLICT"
	CodeAuthIdentityNotFound              = "AUTH_IDENTITY_NOT_FOUND"
	CodeAuthLastCredential                = "AUTH_LAST_CREDENTIAL"    //nolint:gosec // Public problem code, not credential material.
	CodeAPITokenNotFound                  = "API_TOKEN_NOT_FOUND"     // #nosec G101 -- public problem code, not a credential.
	CodeAPITokenLimitReached              = "API_TOKEN_LIMIT_REACHED" // #nosec G101 -- public problem code, not credential material.
)

const (
	CodeUserNotFound           = "USER_NOT_FOUND"
	CodeEmailAlreadyExists     = "EMAIL_ALREADY_EXISTS"
	CodeLastSystemAdmin        = "LAST_SYSTEM_ADMIN"
	CodeSystemAdminRequired    = "SYSTEM_ADMIN_REQUIRED"
	CodeSelfSystemAdminAction  = "SELF_SYSTEM_ADMIN_ACTION"
	CodeInvalidAdminDataImport = "INVALID_ADMIN_DATA_IMPORT"
)

const (
	CodeProjectRoleRequired        = "PROJECT_ROLE_REQUIRED"
	CodeProjectNotFound            = "PROJECT_NOT_FOUND"
	CodeProjectSlugAlreadyExists   = "PROJECT_SLUG_ALREADY_EXISTS"
	CodeProjectMemberNotFound      = "PROJECT_MEMBER_NOT_FOUND"
	CodeProjectMemberAlreadyExists = "PROJECT_MEMBER_ALREADY_EXISTS"
	CodeProjectInviteNotFound      = "PROJECT_INVITE_NOT_FOUND"
	CodeProjectInviteAlreadyExists = "PROJECT_INVITE_ALREADY_EXISTS"
	CodeProjectLastOwner           = "PROJECT_LAST_OWNER"
)

const (
	CodeLabelNotFound      = "LABEL_NOT_FOUND"
	CodeLabelAlreadyExists = "LABEL_ALREADY_EXISTS"
)

const (
	CodeCheckNotFound = "CHECK_NOT_FOUND"
)

const (
	CodeProbeNotFound               = "PROBE_NOT_FOUND"
	CodeProbeCredentialInvalid      = "PROBE_CREDENTIAL_INVALID" // #nosec G101 -- public problem code, not credential material.
	CodeProbeDisabled               = "PROBE_DISABLED"
	CodeProbeRuntimeAuthUnavailable = "PROBE_RUNTIME_AUTH_UNAVAILABLE"
)

const (
	CodeAlertRuleNotFound         = "ALERT_RULE_NOT_FOUND"
	CodeAlertIncidentNotFound     = "ALERT_INCIDENT_NOT_FOUND"
	CodeAlertNotificationNotFound = "ALERT_NOTIFICATION_NOT_FOUND"
)

const (
	CodePublicStatusPageNotFound      = "PUBLIC_STATUS_PAGE_NOT_FOUND"
	CodePublicStatusElementNotFound   = "PUBLIC_STATUS_ELEMENT_NOT_FOUND"
	CodePublicStatusSlugAlreadyExists = "PUBLIC_STATUS_SLUG_ALREADY_EXISTS"
)

const (
	CodeAgentBinaryNotFound    = "AGENT_BINARY_NOT_FOUND"
	CodeAgentBinaryUnavailable = "AGENT_BINARY_UNAVAILABLE"
)
