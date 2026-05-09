package auth

import (
	"errors"
	"net/mail"
	"unicode/utf8"

	appvalidation "github.com/yorukot/netstamp/internal/application/validation"
	"github.com/yorukot/netstamp/internal/normalize"
)

const (
	maxEmailRunes            = 254
	maxDisplayNameRunes      = 100
	minRegisterPasswordRunes = 8
	maxRegisterPasswordRunes = 128
)

type normalizedRegisterInput struct {
	email       string
	displayName string
	password    string
}

type normalizedLoginInput struct {
	email    string
	password string
}

func normalizeRegisterInput(input RegisterInput) (normalizedRegisterInput, error) {
	email, err := normalizeRegisterEmail(input.Email)
	if err != nil {
		return normalizedRegisterInput{}, err
	}
	displayName, err := normalizeRegisterDisplayName(input.DisplayName)
	if err != nil {
		return normalizedRegisterInput{}, err
	}
	password, err := validateRegisterPassword(input.Password)
	if err != nil {
		return normalizedRegisterInput{}, err
	}

	return normalizedRegisterInput{
		email:       email,
		displayName: displayName,
		password:    password,
	}, nil
}

func normalizeRegisterEmail(value string) (string, error) {
	normalized := normalize.Email(value)
	if normalized == "" {
		return "", invalidAuthField("email", "must not be blank", value)
	}
	if utf8.RuneCountInString(normalized) > maxEmailRunes {
		return "", invalidAuthField("email", "must be at most 254 characters", value)
	}
	parsed, err := mail.ParseAddress(normalized)
	if err != nil || parsed.Address != normalized {
		return "", invalidAuthField("email", "must be a valid email address", value)
	}

	return normalized, nil
}

func normalizeRegisterDisplayName(value string) (string, error) {
	normalized, err := appvalidation.RequiredString(errors.Join(ErrInvalidInput, ErrDisplayNameRequired), "displayName", value, 0)
	if err != nil {
		return "", err
	}
	if utf8.RuneCountInString(normalized) > maxDisplayNameRunes {
		return "", appvalidation.New(
			errors.Join(ErrInvalidInput, ErrDisplayNameTooLong),
			"displayName",
			"must be at most 100 characters",
			value,
		)
	}

	return normalized, nil
}

func validateRegisterPassword(value string) (string, error) {
	if _, err := appvalidation.RequiredStringRange(ErrInvalidInput, "password", value, minRegisterPasswordRunes, 0); err != nil {
		return "", err
	}
	if utf8.RuneCountInString(value) > maxRegisterPasswordRunes {
		return "", invalidAuthField("password", "must be at most 128 characters", value)
	}

	return value, nil
}

func normalizeLoginInput(input LoginInput) (normalizedLoginInput, error) {
	email, err := normalizeLoginEmail(input.Email)
	if err != nil {
		return normalizedLoginInput{}, err
	}
	password, err := validateLoginPassword(input.Password)
	if err != nil {
		return normalizedLoginInput{}, err
	}

	return normalizedLoginInput{
		email:    email,
		password: password,
	}, nil
}

func normalizeLoginEmail(value string) (string, error) {
	normalized := normalize.Email(value)
	if normalized == "" {
		return "", invalidAuthField("email", "must not be blank", value)
	}
	if utf8.RuneCountInString(normalized) > maxEmailRunes {
		return "", invalidAuthField("email", "must be at most 254 characters", value)
	}

	return normalized, nil
}

func validateLoginPassword(value string) (string, error) {
	if _, err := appvalidation.RequiredString(ErrInvalidInput, "password", value, maxRegisterPasswordRunes); err != nil {
		return "", err
	}

	return value, nil
}

func invalidAuthField(field, message string, value any) error {
	return appvalidation.New(ErrInvalidInput, field, message, value)
}

func registerValidationReason(err error) AuthEventReason {
	if errors.Is(err, ErrDisplayNameRequired) || errors.Is(err, ErrDisplayNameTooLong) {
		return AuthReasonDisplayNameInvalid
	}

	return AuthReasonInvalidInput
}
