package identity

import (
	"strings"
	"time"

	"github.com/yorukot/spvalidator"
)

type User struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	DisplayName     string     `json:"displayName"`
	PasswordHash    string     `json:"-"`
	HasPassword     bool       `json:"hasPassword"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt,omitempty"`
	DisabledAt      *time.Time `json:"disabledAt,omitempty"`
	IsSystemAdmin   bool       `json:"isSystemAdmin"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func VNUserID(id string) (string, error) {
	err := spvalidator.UUID(id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func VNUserEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	err := spvalidator.Email(email)
	if err != nil {
		return "", err
	}
	return email, nil
}

func VNUserDisplayName(displayName string) (string, error) {
	displayName = strings.TrimSpace(displayName)

	err := spvalidator.Max(displayName, 64)
	if err != nil {
		return "", err
	}
	err = spvalidator.Min(displayName, 1)
	if err != nil {
		return "", err
	}

	return displayName, nil
}

func VNUserPassword(password string) (string, error) {
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
