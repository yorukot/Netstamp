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

type CreateSessionInput struct {
	UserID    string
	UserAgent string
	Now       time.Time
}

type SessionResult struct {
	ID                string
	UserAgent         string
	CreatedAt         time.Time
	LastUsedAt        time.Time
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
	IsCurrent         bool
}

type AuthAccessResult struct {
	UserID                    string
	Email                     string
	DisplayName               string
	EmailVerified             bool
	IsSystemAdmin             bool
	EmailVerificationRequired bool
	SessionToken              string
	ExpiresIn                 int
}
