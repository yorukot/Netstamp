package probe

import (
	"errors"
	"testing"

	appvalidation "github.com/yorukot/netstamp/internal/controller/application/validation"
)

const testCurrentUserID = "11111111-1111-1111-1111-111111111111"

func TestNormalizeCreateProbeInputReturnsAllFieldErrors(t *testing.T) {
	locationName := ""
	latitude := 91.0
	longitude := 181.0

	_, err := normalizeCreateProbeInput(CreateProbeInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    "",
		Name:          "",
		LocationName:  &locationName,
		Latitude:      &latitude,
		Longitude:     &longitude,
		LabelIDs:      []string{""},
	})

	assertValidationFields(t, err, []string{"projectRef", "name", "locationName", "latitude", "longitude", "labelIds"})
}

func TestNormalizeUpdateProbeInputReturnsAllFieldErrors(t *testing.T) {
	name := ""
	locationName := ""
	latitude := 91.0
	longitude := 181.0
	labelIDs := []string{""}

	_, err := normalizeUpdateProbeInput(UpdateProbeInput{
		CurrentUserID: testCurrentUserID,
		ProjectRef:    "",
		ProbeID:       "",
		Name:          &name,
		LocationName:  &locationName,
		Latitude:      &latitude,
		Longitude:     &longitude,
		LabelIDs:      &labelIDs,
	})

	assertValidationFields(t, err, []string{"projectRef", "probeId", "name", "locationName", "latitude", "longitude", "labelIds"})
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
