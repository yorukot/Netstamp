package identity

import "time"

const (
	AuthenticationMethodPassword = "password"
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
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastLoginAt   *time.Time
}

type OIDCAuthFlow struct {
	ID               string
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
