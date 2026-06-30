package auth

import (
	"net/url"
	"strings"

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

func normalizeRequestPasswordResetInput(input RequestPasswordResetInput) (RequestPasswordResetInput, error) {
	var validation appvalidation.Collector

	email, err := identity.VNUserEmail(input.Email)
	if err != nil {
		validation.AddError("email", err, input.Email)
	}
	resetBaseURL, err := normalizeResetBaseURL(input.ResetBaseURL)
	if err != nil {
		validation.AddError("resetBaseUrl", err, input.ResetBaseURL)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return RequestPasswordResetInput{}, err
	}

	return RequestPasswordResetInput{
		Email:        email,
		ResetBaseURL: resetBaseURL,
	}, nil
}

func normalizeConfirmPasswordResetInput(input ConfirmPasswordResetInput) (ConfirmPasswordResetInput, error) {
	var validation appvalidation.Collector

	token := strings.TrimSpace(input.Token)
	if token == "" {
		validation.Add("token", "must not be empty", nil)
	}
	if len(token) > 512 {
		validation.Add("token", "must be 512 characters or less", nil)
	}
	newPassword, err := identity.VNUserPassword(input.NewPassword)
	if err != nil {
		validation.AddError("newPassword", err, "")
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return ConfirmPasswordResetInput{}, err
	}

	return ConfirmPasswordResetInput{
		Token:       token,
		NewPassword: newPassword,
	}, nil
}

func normalizeResetBaseURL(value string) (string, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(value), "/")
	if trimmed == "" {
		return "", ErrInvalidInput
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" {
		return "", ErrInvalidInput
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", ErrInvalidInput
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", ErrInvalidInput
	}

	return trimmed, nil
}
