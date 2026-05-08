package network

import (
	"errors"
	"testing"
)

func TestParseIPFamily(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want IPFamily
	}{
		{name: "inet", raw: "inet", want: IPFamilyInet},
		{name: "inet6", raw: "inet6", want: IPFamilyInet6},
		{name: "trim", raw: " inet ", want: IPFamilyInet},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIPFamily(tt.raw)
			if err != nil {
				t.Fatalf("parse ip family: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestParseIPFamilyRejectsUnknownValue(t *testing.T) {
	_, err := ParseIPFamily("ipv10")
	if !errors.Is(err, ErrInvalidIPFamily) {
		t.Fatalf("expected invalid ip family, got %v", err)
	}
}

func TestParseOptionalIPFamily(t *testing.T) {
	got, err := ParseOptionalIPFamily(nil)
	if err != nil {
		t.Fatalf("parse omitted ip family: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}

	raw := "inet6"
	got, err = ParseOptionalIPFamily(&raw)
	if err != nil {
		t.Fatalf("parse ip family: %v", err)
	}
	if got == nil || *got != IPFamilyInet6 {
		t.Fatalf("expected inet6, got %#v", got)
	}
}
