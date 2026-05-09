package validation

import (
	"errors"
	"testing"
)

var errInvalid = errors.New("invalid")

func TestRequiredStringTrimsAndValidates(t *testing.T) {
	got, err := RequiredString(errInvalid, "name", "  Example  ", 10)
	if err != nil {
		t.Fatalf("required string: %v", err)
	}
	if got != "Example" {
		t.Fatalf("expected trimmed string, got %q", got)
	}

	_, err = RequiredString(errInvalid, "name", "   ", 10)
	if !errors.Is(err, errInvalid) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	assertFieldError(t, err, "name", "must not be blank")

	_, err = RequiredString(errInvalid, "name", "abcdef", 5)
	if !errors.Is(err, errInvalid) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	assertFieldError(t, err, "name", "must be at most 5 characters")
}

func TestRequiredStringRange(t *testing.T) {
	got, err := RequiredStringRange(errInvalid, "password", "  correct-password  ", 8, 128)
	if err != nil {
		t.Fatalf("required string range: %v", err)
	}
	if got != "correct-password" {
		t.Fatalf("expected trimmed string, got %q", got)
	}

	_, err = RequiredStringRange(errInvalid, "password", "short", 8, 128)
	if !errors.Is(err, errInvalid) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	assertFieldError(t, err, "password", "must be at least 8 characters")
}

func TestOptionalString(t *testing.T) {
	got, err := OptionalString(errInvalid, "description", nil, 10)
	if err != nil {
		t.Fatalf("optional string: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil optional string, got %#v", got)
	}

	value := "  Tokyo  "
	got, err = OptionalString(errInvalid, "description", &value, 10)
	if err != nil {
		t.Fatalf("optional string: %v", err)
	}
	if got == nil || *got != "Tokyo" {
		t.Fatalf("expected trimmed optional string, got %#v", got)
	}
}

func TestPositiveInt32(t *testing.T) {
	got, err := PositiveInt32(errInvalid, "intervalSeconds", 30)
	if err != nil {
		t.Fatalf("positive int32: %v", err)
	}
	if got != 30 {
		t.Fatalf("expected value 30, got %d", got)
	}

	_, err = PositiveInt32(errInvalid, "intervalSeconds", 0)
	if !errors.Is(err, errInvalid) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	assertFieldError(t, err, "intervalSeconds", "must be greater than 0")
}

func TestOptionalInt32Range(t *testing.T) {
	value := int32(10)
	got, err := OptionalInt32Range(errInvalid, "packetSizeBytes", &value, 0, 20)
	if err != nil {
		t.Fatalf("int32 range: %v", err)
	}
	if got == nil || *got != 10 {
		t.Fatalf("expected value 10, got %#v", got)
	}

	tooLarge := int32(21)
	_, err = OptionalInt32Range(errInvalid, "packetSizeBytes", &tooLarge, 0, 20)
	if !errors.Is(err, errInvalid) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	assertFieldError(t, err, "packetSizeBytes", "must be between 0 and 20")
}

func TestCanonicalUUIDSet(t *testing.T) {
	got, err := CanonicalUUIDSet(errInvalid, "labelIds", []string{
		" 33333333-3333-3333-3333-333333333333 ",
		"33333333-3333-3333-3333-333333333333",
		"44444444-4444-4444-4444-444444444444",
	})
	if err != nil {
		t.Fatalf("canonical uuid set: %v", err)
	}

	want := []string{
		"33333333-3333-3333-3333-333333333333",
		"44444444-4444-4444-4444-444444444444",
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d ids, got %#v", len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected id %d to be %q, got %q", i, want[i], got[i])
		}
	}

	_, err = CanonicalUUIDSet(errInvalid, "labelIds", []string{"not-a-uuid"})
	if !errors.Is(err, errInvalid) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	assertFieldError(t, err, "labelIds", "must contain valid UUIDs")
}

func TestCanonicalUUID(t *testing.T) {
	got, err := CanonicalUUID(errInvalid, "userId", " 33333333-3333-3333-3333-333333333333 ")
	if err != nil {
		t.Fatalf("canonical uuid: %v", err)
	}
	if got != "33333333-3333-3333-3333-333333333333" {
		t.Fatalf("expected canonical uuid, got %q", got)
	}

	_, err = CanonicalUUID(errInvalid, "userId", "not-a-uuid")
	if !errors.Is(err, errInvalid) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	assertFieldError(t, err, "userId", "must be a valid UUID")
}

func assertFieldError(t *testing.T, err error, wantField, wantMessage string) {
	t.Helper()

	fields, ok := FieldErrors(err)
	if !ok {
		t.Fatalf("expected field errors, got %v", err)
	}
	for _, field := range fields {
		if field.Field == wantField && field.Message == wantMessage {
			return
		}
	}

	t.Fatalf("expected field error %q/%q, got %#v", wantField, wantMessage, fields)
}
