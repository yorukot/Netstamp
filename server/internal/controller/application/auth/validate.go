package auth

import (
	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

func normalizeRegisterInput(input RegisterInput) (RegisterInput, error) {
	var validation appvalidation.Collector

	email, err := identity.VNUserEmail(input.Email)
	if err != nil {
		validation.AddError("email", err, input.Email)
	}
	displayName, err := identity.VNUserDisplayName(input.DisplayName)
	if err != nil {
		validation.AddError("displayName", err, input.DisplayName)
	}
	password, err := identity.VNUserPassword(input.Password)
	if err != nil {
		validation.AddError("password", err, "")
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return RegisterInput{}, err
	}

	return RegisterInput{
		DisplayName: displayName,
		Password:    password,
		Email:       email,
	}, nil
}

func normalizeLoginInput(input LoginInput) (LoginInput, error) {
	var validation appvalidation.Collector

	email, err := identity.VNUserEmail(input.Email)
	if err != nil {
		validation.AddError("email", err, input.Email)
	}
	password, err := identity.VNUserPassword(input.Password)
	if err != nil {
		validation.AddError("password", err, "")
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return LoginInput{}, err
	}

	return LoginInput{
		Email:    email,
		Password: password,
	}, nil
}
