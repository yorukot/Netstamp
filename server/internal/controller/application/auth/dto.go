package auth

import "time"

const DefaultEmailVerificationTokenTTL = 15 * time.Minute

type RegisterInput struct {
	Email                    string
	DisplayName              string
	Password                 string
	UserAgent                string
	RequireEmailVerification bool
	EmailVerificationBaseURL string
}

type LoginInput struct {
	Email                    string
	Password                 string
	UserAgent                string
	RequireEmailVerification bool
}

type RequestPasswordResetInput struct {
	Email        string
	ResetBaseURL string
}

type ConfirmPasswordResetInput struct {
	Token       string
	NewPassword string
}

type RequestEmailVerificationInput struct {
	Email                    string
	EmailVerificationBaseURL string
}

type ConfirmEmailVerificationInput struct {
	Token string
}

type PasswordResetConfig struct {
	TokenTTL time.Duration
}

type EmailVerificationConfig struct {
	TokenTTL time.Duration
}

type ExternalAuthConfig struct {
	FlowTTL      time.Duration
	AuthTimeSkew time.Duration
}

type ExternalProviderConfig struct {
	ID          string
	DisplayName string
	JITEnabled  bool
	SudoCapable bool
}

type ExternalProviderRegistration struct {
	Config ExternalProviderConfig
	Client ExternalAuthClient
}

type ExternalProviderMethod struct {
	ID          string
	DisplayName string
	SudoCapable bool
}

type ExternalIdentityClaims struct {
	Issuer        string
	Subject       string
	Email         string
	EmailVerified bool
	DisplayName   string
	Username      string
	AvatarURL     string
	AuthTime      time.Time
}

type StartExternalAuthInput struct {
	Provider  string
	Intent    string
	SessionID string
	ReturnTo  string
}

type StartExternalAuthResult struct {
	AuthorizationURL string
	BrowserToken     string
	ExpiresAt        time.Time
}

type CompleteExternalAuthInput struct {
	Provider     string
	Code         string
	State        string
	BrowserToken string
	UserAgent    string
}

type CompleteExternalAuthResult struct {
	Intent   string
	ReturnTo string
	Access   *AuthAccessResult
}

type SudoStatusResult struct {
	Active    bool
	ExpiresAt time.Time
	Methods   []string
}

type CreateSessionInput struct {
	UserID               string
	UserAgent            string
	Now                  time.Time
	AuthenticationMethod string
	SudoEligible         bool
	IdentityID           *string
}

type SessionResult struct {
	ID                   string
	UserAgent            string
	CreatedAt            time.Time
	LastUsedAt           time.Time
	IdleExpiresAt        time.Time
	AbsoluteExpiresAt    time.Time
	AuthenticatedAt      time.Time
	AuthenticationMethod string
	IsCurrent            bool
}

type AuthAccessResult struct {
	UserID                    string
	Email                     string
	DisplayName               string
	EmailVerified             bool
	IsSystemAdmin             bool
	HasPassword               bool
	EmailVerificationRequired bool
	SessionToken              string
	ExpiresIn                 int
}
