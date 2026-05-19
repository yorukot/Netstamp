package label

import (
	"errors"
	"testing"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
)

const (
	testCurrentUserID = "11111111-1111-1111-1111-111111111111"
	testLabelID       = "22222222-2222-2222-2222-222222222222"
)

func TestNormalizeCreateLabelInputPreservesCurrentUserID(t *testing.T) {
	input, err := normalizeCreateLabelInput(CreateLabelInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    " project ",
		Key:           " region ",
		Value:         " tokyo ",
	})
	if err != nil {
		t.Fatalf("expected valid input: %v", err)
	}

	if input.CurrentUserID != testCurrentUserID {
		t.Fatalf("expected current user ID to be preserved, got %q", input.CurrentUserID)
	}
}

func TestNormalizeCreateLabelInputReturnsAllFieldErrors(t *testing.T) {
	_, err := normalizeCreateLabelInput(CreateLabelInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    "",
		Key:           "",
		Value:         "",
	})

	assertValidationFields(t, err, []string{"projectRef", "key", "value"})
}

func TestNormalizeUpdateLabelInputPreservesCurrentUserID(t *testing.T) {
	key := " region "
	input, err := normalizeUpdateLabelInput(UpdateLabelInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    " project ",
		LabelID:       testLabelID,
		Key:           &key,
	})
	if err != nil {
		t.Fatalf("expected valid input: %v", err)
	}

	if input.CurrentUserID != testCurrentUserID {
		t.Fatalf("expected current user ID to be preserved, got %q", input.CurrentUserID)
	}
}

func TestNormalizeUpdateLabelInputReturnsAllFieldErrors(t *testing.T) {
	key := ""
	value := ""
	_, err := normalizeUpdateLabelInput(UpdateLabelInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    "",
		LabelID:       "",
		Key:           &key,
		Value:         &value,
	})

	assertValidationFields(t, err, []string{"projectRef", "labelId", "key", "value"})
}

func TestNormalizeDeleteLabelInputReturnsAllFieldErrors(t *testing.T) {
	_, err := normalizeDeleteLabelInput(DeleteLabelInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    "",
		LabelID:       "",
	})

	assertValidationFields(t, err, []string{"projectRef", "labelId"})
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
