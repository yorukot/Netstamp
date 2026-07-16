package identity

import "time"

type AuthSession struct {
	ID                   string
	UserID               string
	TokenHash            []byte
	CSRFTokenHash        []byte
	UserAgent            string
	AuthenticatedAt      time.Time
	AuthenticationMethod string
	SudoEligible         bool
	IdentityID           *string
	CreatedAt            time.Time
	LastUsedAt           time.Time
	IdleExpiresAt        time.Time
	AbsoluteExpiresAt    time.Time
	RevokedAt            *time.Time
	RevokedReason        *string
}

type CreatedSession struct {
	Session         AuthSession
	RawToken        string
	RawCSRFToken    string
	ExpiresIn       int
	CookieExpiresAt time.Time
}

type SessionClaims struct {
	SessionID string
	UserID    string
}

type SudoStatus struct {
	Active    bool
	ExpiresAt time.Time
}
