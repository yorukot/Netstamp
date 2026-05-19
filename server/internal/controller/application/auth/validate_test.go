package auth

import (
	"errors"
	"testing"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
)

func TestNormalizeRegisterInputReturnsAllFieldErrors(t *testing.T) {
	_, err := normalizeRegisterInput(RegisterInput{
		Email:       "not-an-email",
		DisplayName: "",
		Password:    "short",
	})

	assertValidationFields(t, err, []string{"email", "displayName", "password"})
}

func TestNormalizeLoginInputReturnsAllFieldErrors(t *testing.T) {
	_, err := normalizeLoginInput(LoginInput{
		Email:    "not-an-email",
		Password: "short",
	})

	assertValidationFields(t, err, []string{"email", "password"})
}

func assertValidationFields(t *testing.T, err error, wantFields []string) {
	t.Helper()

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected invalid input error, got %v", err)
	}

	fieldErrors, ok := appvalidation.FieldErrors(err)
	if !ok {
		t.Fatalf("expected field validation errors, got %v", err)
	}
	if len(fieldErrors) != len(wantFields) {
		t.Fatalf("expected %d field errors, got %d: %#v", len(wantFields), len(fieldErrors), fieldErrors)
	}
	for i, wantField := range wantFields {
		if fieldErrors[i].Field != wantField {
			t.Fatalf("expected field error %d to target %q, got %q", i, wantField, fieldErrors[i].Field)
		}
		if fieldErrors[i].Message == "" {
			t.Fatalf("expected field error %d to include a message", i)
		}
	}
}
