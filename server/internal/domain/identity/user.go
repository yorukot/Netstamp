package identity

import (
	"errors"
	"time"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrAccessTokenInvalid = errors.New("access token invalid")
)

type CreateUserInput struct {
	Email        string
	DisplayName  string
	PasswordHash string
}

type User struct {
	ID           string
	Email        string
	DisplayName  *string
	PasswordHash string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
