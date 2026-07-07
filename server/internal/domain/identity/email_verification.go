package identity

import "time"

type EmailVerificationToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"userId"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expiresAt"`
	UsedAt    *time.Time `json:"usedAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

type EmailVerificationEmail struct {
	To              string
	DisplayName     string
	VerificationURL string
	ExpiresAt       time.Time
}
