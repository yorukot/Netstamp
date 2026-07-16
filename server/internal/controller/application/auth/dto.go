package auth

import "time"

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

type OIDCConfig struct {
	Enabled      bool
	DisplayName  string
	JITEnabled   bool
	FlowTTL      time.Duration
	AuthTimeSkew time.Duration
}

type OIDCClaims struct {
	Issuer        string
	Subject       string
	Email         string
	EmailVerified bool
	DisplayName   string
	AuthTime      time.Time
}

type StartOIDCInput struct {
	Intent    string
	SessionID string
	ReturnTo  string
}

type StartOIDCResult struct {
	AuthorizationURL string
	BrowserToken     string
	ExpiresAt        time.Time
}

type CompleteOIDCInput struct {
	Code         string
	State        string
	BrowserToken string
	UserAgent    string
}

type CompleteOIDCResult struct {
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
