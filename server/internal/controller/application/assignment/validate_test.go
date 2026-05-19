package assignment

import (
	"errors"
	"testing"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
)

func TestNormalizeProbeTargetReturnsAllFieldErrors(t *testing.T) {
	_, _, err := normalizeProbeTarget("", "")

	assertValidationFields(t, err, []string{"projectId", "probeId"})
}

func TestNormalizeCheckTargetReturnsAllFieldErrors(t *testing.T) {
	_, _, err := normalizeCheckTarget("", "")

	assertValidationFields(t, err, []string{"projectId", "checkId"})
}

func TestNormalizeLabelTargetReturnsAllFieldErrors(t *testing.T) {
	_, _, err := normalizeLabelTarget("", "")

	assertValidationFields(t, err, []string{"projectId", "labelId"})
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
