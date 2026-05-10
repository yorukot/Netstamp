package security

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
)

func TestGenerateProbeSecretReturnsURLSafePlaintextAndHash(t *testing.T) {
	reader := strings.NewReader(strings.Repeat("a", ProbeSecretByteLength))
	generator := &ProbeSecretGenerator{reader: reader}

	plaintext, hash, err := generator.GenerateProbeSecret()
	if err != nil {
		t.Fatalf("generate probe secret: %v", err)
	}
	if plaintext == "" {
		t.Fatalf("expected plaintext secret")
	}
	if strings.Contains(plaintext, "=") {
		t.Fatalf("expected unpadded base64, got %q", plaintext)
	}
	if _, err := base64.RawURLEncoding.DecodeString(plaintext); err != nil {
		t.Fatalf("expected URL-safe base64 secret, got %q: %v", plaintext, err)
	}
	if len(hash) != 64 {
		t.Fatalf("expected sha256 hex length 64, got %d", len(hash))
	}
	if hash != strings.ToLower(hash) {
		t.Fatalf("expected lowercase sha256 hex digest, got %q", hash)
	}
	if _, err := hex.DecodeString(hash); err != nil {
		t.Fatalf("expected lowercase sha256 hex digest, got %q: %v", hash, err)
	}
	if strings.Contains(hash, plaintext) {
		t.Fatalf("expected hash not to contain plaintext")
	}
}

func TestVerifyProbeSecret(t *testing.T) {
	secret := "probe-secret"
	hash := HashProbeSecret(secret)

	if !VerifyProbeSecret(secret, hash) {
		t.Fatalf("expected matching secret to verify")
	}
	if VerifyProbeSecret("different-secret", hash) {
		t.Fatalf("expected different secret to fail")
	}
}
