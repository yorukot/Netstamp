package auth

import "time"

type RegisterInput struct {
	Email       string
	DisplayName string
	Password    string
}

type LoginInput struct {
	Email    string
	Password string
}

type RequestPasswordResetInput struct {
	Email        string
	ResetBaseURL string
}

type ConfirmPasswordResetInput struct {
	Token       string
	NewPassword string
}

type PasswordResetConfig struct {
	TokenTTL time.Duration
}

type AuthAccessResult struct {
	UserID        string
	Email         string
	DisplayName   string
	IsSystemAdmin bool
	AccessToken   string
	ExpiresIn     int
}
