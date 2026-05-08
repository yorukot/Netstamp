package normalize

import (
	"errors"
	"testing"
)

var errInvalid = errors.New("invalid")

func TestEmailTrimsAndLowercases(t *testing.T) {
	got := Email(" User@Example.COM ")
	if got != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", got)
	}
}

func TestRequiredString(t *testing.T) {
	got, err := RequiredString("  Example  ", errInvalid)
	if err != nil {
		t.Fatalf("required string: %v", err)
	}
	if got != "Example" {
		t.Fatalf("expected trimmed string, got %q", got)
	}

	_, err = RequiredString("   ", errInvalid)
	if !errors.Is(err, errInvalid) {
		t.Fatalf("expected invalid error, got %v", err)
	}
}

func TestOptionalString(t *testing.T) {
	got, err := OptionalString(nil, errInvalid)
	if err != nil {
		t.Fatalf("optional string: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil optional string, got %#v", got)
	}

	value := "  Tokyo  "
	got, err = OptionalString(&value, errInvalid)
	if err != nil {
		t.Fatalf("optional string: %v", err)
	}
	if got == nil || *got != "Tokyo" {
		t.Fatalf("expected trimmed optional string, got %#v", got)
	}

	empty := "   "
	_, err = OptionalString(&empty, errInvalid)
	if !errors.Is(err, errInvalid) {
		t.Fatalf("expected invalid error, got %v", err)
	}
}

func TestProjectSlug(t *testing.T) {
	got, err := ProjectSlug("  platform-project  ", errInvalid)
	if err != nil {
		t.Fatalf("project slug: %v", err)
	}
	if got != "platform-project" {
		t.Fatalf("expected normalized slug, got %q", got)
	}

	for _, value := range []string{"", "   ", "Platform_Project", "project slug"} {
		t.Run(value, func(t *testing.T) {
			_, err := ProjectSlug(value, errInvalid)
			if !errors.Is(err, errInvalid) {
				t.Fatalf("expected invalid error, got %v", err)
			}
		})
	}
}

func TestIsProjectSlug(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: "engineering", want: true},
		{value: "platform-1", want: true},
		{value: "Platform", want: false},
		{value: "project_slug", want: false},
		{value: " project ", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := IsProjectSlug(tt.value); got != tt.want {
				t.Fatalf("expected %t, got %t", tt.want, got)
			}
		})
	}
}

func TestCanonicalUUIDSet(t *testing.T) {
	got, err := CanonicalUUIDSet([]string{
		" 33333333-3333-3333-3333-333333333333 ",
		"33333333-3333-3333-3333-333333333333",
		"44444444-4444-4444-4444-444444444444",
	}, errInvalid)
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

	_, err = CanonicalUUIDSet([]string{"not-a-uuid"}, errInvalid)
	if !errors.Is(err, errInvalid) {
		t.Fatalf("expected invalid error, got %v", err)
	}
}
