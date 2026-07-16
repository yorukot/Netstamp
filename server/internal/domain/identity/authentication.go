package identity

import "time"

const (
	AuthenticationMethodPassword = "password"
	AuthenticationMethodGoogle   = "google"
	AuthenticationMethodGitHub   = "github"
	AuthenticationMethodOIDC     = "oidc"
)

type UserIdentity struct {
	ID            string
	UserID        string
	Provider      string
	Issuer        string
	Subject       string
	Email         *string
	EmailVerified bool
	DisplayName   *string
	Username      *string
	AvatarURL     *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastLoginAt   *time.Time
}

type ExternalAuthFlow struct {
	ID               string
	Provider         string
	StateHash        []byte
	BrowserTokenHash []byte
	Nonce            string
	PKCEVerifier     string
	Intent           string
	SessionID        *string
	ReturnTo         string
	CreatedAt        time.Time
	ExpiresAt        time.Time
	UsedAt           *time.Time
}
