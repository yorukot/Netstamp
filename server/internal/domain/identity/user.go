package identity

import (
	"errors"
	"strings"
	"time"

	"github.com/yorukot/spvalidator"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailAlreadyExists  = errors.New("email already exists")
	ErrAccessTokenInvalid  = errors.New("access token invalid")
)

type User struct {
	ID           string
	Email        string
	DisplayName  string
	PasswordHash string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func VNUserID(id string) (string, error) {
	err := spvalidator.UUID(id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func VNUserEmail(email string) (string, error) {
	err := spvalidator.Email(email)
	if err != nil {
		return "", err
	}
	return email, nil
}

func VNUserDisplayName(displayName string) (string, error) {
	err := spvalidator.Max(displayName, 64)
	if err != nil {
		return "", err
	}
	err = spvalidator.Min(displayName, 1)
	if err != nil {
		return "", err
	}

	displayName = strings.TrimSpace(displayName)

	return displayName, nil
}

func VNUserPassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	
	err := spvalidator.Min(password, 8)
	if err != nil {
		return "", err
	}
	err = spvalidator.Max(password, 128)
	if err != nil {
		return "", err
	}

	return password, nil
}
