package identity

import "testing"

func TestVNUserEmailNormalizesBeforeValidation(t *testing.T) {
	email, err := VNUserEmail(" USER@Example.COM ")
	if err != nil {
		t.Fatalf("expected valid email: %v", err)
	}

	if email != "user@example.com" {
		t.Fatalf("expected normalized email, got %q", email)
	}
}

func TestVNUserDisplayNameTrimsBeforeValidation(t *testing.T) {
	displayName, err := VNUserDisplayName(" User ")
	if err != nil {
		t.Fatalf("expected valid display name: %v", err)
	}

	if displayName != "User" {
		t.Fatalf("expected trimmed display name, got %q", displayName)
	}
}

func TestVNUserDisplayNameRejectsWhitespaceOnly(t *testing.T) {
	if _, err := VNUserDisplayName(" "); err == nil {
		t.Fatal("expected whitespace-only display name to be invalid")
	}
}
