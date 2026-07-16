package security

import (
	"bytes"
	"strings"
	"testing"
)

func TestAPITokenManagerGeneratesOpaquePrefixedToken(t *testing.T) {
	manager := NewAPITokenManager("test-hash-key")
	raw, hint, err := manager.Generate()
	if err != nil {
		t.Fatalf("generate API token: %v", err)
	}
	if !strings.HasPrefix(raw, apiTokenPrefix) {
		t.Fatalf("expected prefix %q, got %q", apiTokenPrefix, raw)
	}
	if len(strings.TrimPrefix(raw, apiTokenPrefix)) != 43 {
		t.Fatalf("expected 32-byte base64url payload, got token length %d", len(raw))
	}
	if len(hint) != 8 || !strings.HasSuffix(raw, hint) {
		t.Fatalf("expected eight-character suffix hint, got %q", hint)
	}
}

func TestAPITokenManagerUsesKeyedStableHashes(t *testing.T) {
	first := NewAPITokenManager("first-key")
	second := NewAPITokenManager("second-key")
	raw := "nst_pat_example"

	firstHash := first.Hash(raw)
	if len(firstHash) != 32 {
		t.Fatalf("expected SHA-256 digest, got %d bytes", len(firstHash))
	}
	if !bytes.Equal(firstHash, first.Hash(raw)) {
		t.Fatal("expected stable hash for the same key and token")
	}
	if bytes.Equal(firstHash, first.Hash(raw+"-other")) {
		t.Fatal("different token values must not share a hash")
	}
	if bytes.Equal(firstHash, second.Hash(raw)) {
		t.Fatal("different hash keys must produce different hashes")
	}
}
