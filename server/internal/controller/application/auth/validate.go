package auth

import (
	"net/url"
	"strings"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
	"github.com/yorukot/netstamp/internal/domain/identity"
)

const maxSessionUserAgentRunes = 512

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

	emailVerificationBaseURL := strings.TrimRight(strings.TrimSpace(input.EmailVerificationBaseURL), "/")
	if input.RequireEmailVerification {
		normalizedBaseURL, err := normalizeResetBaseURL(input.EmailVerificationBaseURL)
		if err != nil {
			validation.AddError("emailVerificationBaseUrl", err, input.EmailVerificationBaseURL)
		}
		if err := validation.Err(ErrInvalidInput); err != nil {
			return RegisterInput{}, err
		}
		emailVerificationBaseURL = normalizedBaseURL
	}

	return RegisterInput{
		DisplayName:              displayName,
		Password:                 password,
		Email:                    email,
		UserAgent:                normalizeSessionUserAgent(input.UserAgent),
		RequireEmailVerification: input.RequireEmailVerification,
		EmailVerificationBaseURL: emailVerificationBaseURL,
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
		Email:                    email,
		Password:                 password,
		UserAgent:                normalizeSessionUserAgent(input.UserAgent),
		RequireEmailVerification: input.RequireEmailVerification,
	}, nil
}

func normalizeSessionUserAgent(value string) string {
	value = strings.TrimSpace(value)
	runes := []rune(value)
	if len(runes) > maxSessionUserAgentRunes {
		return string(runes[:maxSessionUserAgentRunes])
	}
	return value
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

func normalizeRequestEmailVerificationInput(input RequestEmailVerificationInput) (RequestEmailVerificationInput, error) {
	var validation appvalidation.Collector

	email, err := identity.VNUserEmail(input.Email)
	if err != nil {
		validation.AddError("email", err, input.Email)
	}
	baseURL, err := normalizeResetBaseURL(input.EmailVerificationBaseURL)
	if err != nil {
		validation.AddError("emailVerificationBaseUrl", err, input.EmailVerificationBaseURL)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return RequestEmailVerificationInput{}, err
	}

	return RequestEmailVerificationInput{
		Email:                    email,
		EmailVerificationBaseURL: baseURL,
	}, nil
}

func normalizeConfirmEmailVerificationInput(input ConfirmEmailVerificationInput) (ConfirmEmailVerificationInput, error) {
	var validation appvalidation.Collector

	token := strings.TrimSpace(input.Token)
	if token == "" {
		validation.Add("token", "must not be empty", nil)
	}
	if len(token) > 512 {
		validation.Add("token", "must be 512 characters or less", nil)
	}
	if err := validation.Err(ErrInvalidInput); err != nil {
		return ConfirmEmailVerificationInput{}, err
	}

	return ConfirmEmailVerificationInput{Token: token}, nil
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
