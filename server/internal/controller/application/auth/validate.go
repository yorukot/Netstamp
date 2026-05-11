package auth

import (
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func normalizeRegisterInput(input RegisterInput) (RegisterInput, error) {
	email, err := identity.VNUserEmail(input.Email)
	if err != nil {
		return RegisterInput{}, invalidAuthField("email", err.Error(), input.Email)
	}
	displayName, err := identity.VNUserDisplayName(input.DisplayName)
	if err != nil {
		return RegisterInput{}, invalidAuthField("displayName", err.Error(), input.DisplayName)
	}
	password, err := identity.VNUserPassword(input.Password)
	if err != nil {
		return RegisterInput{}, invalidAuthField("password", err.Error(), input.Password)
	}

	return RegisterInput{
		DisplayName: displayName,
		Password:    password,
		Email:       email,
	}, nil
}

func normalizeLoginInput(input LoginInput) (LoginInput, error) {
	email, err := identity.VNUserEmail(input.Email)
	if err != nil {
		return LoginInput{}, invalidAuthField("email", err.Error(), input.Email)
	}
	password, err := identity.VNUserPassword(input.Password)
	if err != nil {
		return LoginInput{}, invalidAuthField("password", err.Error(), input.Password)
	}

	return LoginInput{
		Email:    email,
		Password: password,
	}, nil
}

func invalidAuthField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}
