package identity

import "time"

type PasswordResetToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"userId"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expiresAt"`
	UsedAt    *time.Time `json:"usedAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

type PasswordResetEmail struct {
	To          string
	DisplayName string
	ResetURL    string
	ExpiresAt   time.Time
}
