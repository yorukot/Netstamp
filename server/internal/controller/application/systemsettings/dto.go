package systemsettings

import "time"

type Resource string

const (
	ResourceAccess  Resource = "access"
	ResourceSMTP    Resource = "smtp"
	ResourceOIDC    Resource = "auth.oidc"
	ResourceGoogle  Resource = "auth.google"
	ResourceGitHub  Resource = "auth.github"
	ResourceUpdates Resource = "updates"
)

// OptionalSecret preserves the three JSON PATCH states used by secret fields:
// omitted (Present=false), null (Present=true, Value=nil), and replacement.
type OptionalSecret struct {
	Present bool
	Value   *string
}

type Defaults struct {
	Access AccessSettings
	SMTP   SMTPSettings
	OIDC   OIDCSettings
	Google GoogleSettings
	GitHub GitHubSettings
}

type AccessSettings struct {
	AccountCreationEnabled    bool
	EmailVerificationRequired bool
	ProjectCreationEnabled    bool
	CredentialChangesEnabled  bool
}

type UpdatesSettings struct {
	CheckForUpdates bool
}

type SMTPSettings struct {
	Host           string
	Port           int32
	Username       string
	PasswordSet    bool
	From           string
	TLSMode        string
	TimeoutSeconds int32
	Configured     bool
}

type SMTPRuntimeSettings struct {
	Host           string
	Port           int32
	Username       string
	Password       string
	From           string
	TLSMode        string
	TimeoutSeconds int32
}

type OIDCSettings struct {
	Enabled         bool
	IssuerURL       string
	ClientID        string
	ClientSecretSet bool
	DisplayName     string
	JITEnabled      bool
	CallbackURL     *string
}

type GoogleSettings struct {
	Enabled         bool
	ClientID        string
	ClientSecretSet bool
	DisplayName     string
	JITEnabled      bool
	AllowedDomains  []string
	CallbackURL     *string
}

type GitHubSettings struct {
	Enabled         bool
	ClientID        string
	ClientSecretSet bool
	DisplayName     string
	JITEnabled      bool
	AllowSignup     bool
	CallbackURL     *string
}

type OIDCRuntimeSettings struct {
	Enabled      bool
	IssuerURL    string
	ClientID     string
	ClientSecret string
	DisplayName  string
	JITEnabled   bool
}

type GoogleRuntimeSettings struct {
	Enabled        bool
	ClientID       string
	ClientSecret   string
	DisplayName    string
	JITEnabled     bool
	AllowedDomains []string
}

type GitHubRuntimeSettings struct {
	Enabled      bool
	ClientID     string
	ClientSecret string
	DisplayName  string
	JITEnabled   bool
	AllowSignup  bool
}

type GetAccessInput struct {
	CurrentUserID string
}

type GetSMTPInput struct {
	CurrentUserID string
}

type GetOIDCInput struct {
	CurrentUserID string
}

type GetGoogleInput struct {
	CurrentUserID string
}

type GetGitHubInput struct {
	CurrentUserID string
}

type GetUpdatesInput struct {
	CurrentUserID string
}

type UpdateAccessInput struct {
	CurrentUserID             string
	AccountCreationEnabled    *bool
	EmailVerificationRequired *bool
	ProjectCreationEnabled    *bool
	CredentialChangesEnabled  *bool
}

type UpdateSMTPInput struct {
	CurrentUserID  string
	Host           *string
	Port           *int32
	Username       *string
	Password       OptionalSecret
	From           *string
	TLSMode        *string
	TimeoutSeconds *int32
}

type UpdateOIDCInput struct {
	CurrentUserID string
	Enabled       *bool
	IssuerURL     *string
	ClientID      *string
	ClientSecret  OptionalSecret
	DisplayName   *string
	JITEnabled    *bool
}

type UpdateGoogleInput struct {
	CurrentUserID  string
	Enabled        *bool
	ClientID       *string
	ClientSecret   OptionalSecret
	DisplayName    *string
	JITEnabled     *bool
	AllowedDomains *[]string
}

type UpdateGitHubInput struct {
	CurrentUserID string
	Enabled       *bool
	ClientID      *string
	ClientSecret  OptionalSecret
	DisplayName   *string
	JITEnabled    *bool
	AllowSignup   *bool
}

type UpdateUpdatesInput struct {
	CurrentUserID   string
	CheckForUpdates *bool
}

type (
	ValidateOIDCInput   = UpdateOIDCInput
	ValidateGoogleInput = UpdateGoogleInput
	ValidateGitHubInput = UpdateGitHubInput
)

type TestSMTPInput struct {
	CurrentUserID string
}

type SMTPTestUser struct {
	Email      string
	DisabledAt *time.Time
}
