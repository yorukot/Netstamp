package auth

import "time"

type RegisterInput struct {
	Email                    string
	DisplayName              string
	Password                 string
	RequireEmailVerification bool
	EmailVerificationBaseURL string
}

type LoginInput struct {
	Email                    string
	Password                 string
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
	UserID string
	Now    time.Time
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
