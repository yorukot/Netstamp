package security

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestArgon2idPasswordHasherHashesAndComparesPassword(t *testing.T) {
	ctx := context.Background()
	hasher := NewArgon2idPasswordHasher(testArgon2idConfig())

	encoded, err := hasher.Hash(ctx, "correct horse battery staple")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if !strings.HasPrefix(encoded, "$argon2id$v=19$m=8,t=1,p=1$") {
		t.Fatalf("unexpected encoded hash prefix: %q", encoded)
	}

	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		t.Fatalf("expected six hash parts, got %d in %q", len(parts), encoded)
	}
	if _, err := base64.RawStdEncoding.DecodeString(parts[4]); err != nil {
		t.Fatalf("expected raw base64 salt, got %q: %v", parts[4], err)
	}
	if _, err := base64.RawStdEncoding.DecodeString(parts[5]); err != nil {
		t.Fatalf("expected raw base64 hash, got %q: %v", parts[5], err)
	}

	if err := hasher.Compare(ctx, "correct horse battery staple", encoded); err != nil {
		t.Fatalf("compare password: %v", err)
	}
}

func TestArgon2idPasswordHasherRejectsMismatch(t *testing.T) {
	ctx := context.Background()
	hasher := NewArgon2idPasswordHasher(testArgon2idConfig())
	encoded, err := hasher.Hash(ctx, "correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	err = hasher.Compare(ctx, "wrong-password", encoded)
	if !errors.Is(err, ErrPasswordMismatch) {
		t.Fatalf("expected password mismatch, got %v", err)
	}
}

func TestArgon2idPasswordHasherRejectsInvalidConfig(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name string
		cfg  Argon2idConfig
	}{
		{name: "memory", cfg: Argon2idConfig{Iterations: 1, Parallelism: 1}},
		{name: "iterations", cfg: Argon2idConfig{MemoryKiB: 8, Parallelism: 1}},
		{name: "parallelism", cfg: Argon2idConfig{MemoryKiB: 8, Iterations: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewArgon2idPasswordHasher(tt.cfg).Hash(ctx, "password")
			if !errors.Is(err, ErrInvalidArgon2idConfig) {
				t.Fatalf("expected invalid config, got %v", err)
			}
		})
	}
}

func TestArgon2idPasswordHasherReturnsRandomReaderError(t *testing.T) {
	ctx := context.Background()
	originalReadRandom := readRandom
	t.Cleanup(func() {
		readRandom = originalReadRandom
	})
	readRandom = func([]byte) (int, error) {
		return 0, io.ErrUnexpectedEOF
	}

	_, err := NewArgon2idPasswordHasher(testArgon2idConfig()).Hash(ctx, "password")
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected random reader error, got %v", err)
	}
}

func TestArgon2idPasswordHasherRejectsInvalidHashes(t *testing.T) {
	ctx := context.Background()
	hasher := NewArgon2idPasswordHasher(testArgon2idConfig())
	validEncoded, err := hasher.Hash(ctx, "password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	parts := strings.Split(validEncoded, "$")

	tests := []struct {
		name    string
		encoded string
	}{
		{name: "wrong part count", encoded: "argon2id"},
		{name: "wrong algorithm", encoded: "$bcrypt$v=19$m=8,t=1,p=1$c2FsdA$hash"},
		{name: "missing version prefix", encoded: "$argon2id$19$m=8,t=1,p=1$c2FsdA$hash"},
		{name: "bad version", encoded: "$argon2id$v=bad$m=8,t=1,p=1$c2FsdA$hash"},
		{name: "unsupported version", encoded: "$argon2id$v=18$m=8,t=1,p=1$c2FsdA$hash"},
		{name: "bad params", encoded: "$argon2id$v=19$m=8,t=1$c2FsdA$hash"},
		{name: "bad salt base64", encoded: "$argon2id$v=19$m=8,t=1,p=1$***$hash"},
		{name: "bad hash base64", encoded: "$argon2id$v=19$m=8,t=1,p=1$c2FsdA$***"},
		{name: "short hash", encoded: strings.Join([]string{parts[0], parts[1], parts[2], parts[3], parts[4], "c2hvcnQ"}, "$")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hasher.Compare(ctx, "password", tt.encoded)
			if !errors.Is(err, ErrInvalidHash) {
				t.Fatalf("expected invalid hash, got %v", err)
			}
		})
	}
}

func TestDecodeArgon2idParamsRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		params string
	}{
		{name: "missing value", params: "m=8,t=1,p"},
		{name: "unknown key", params: "m=8,t=1,x=1"},
		{name: "bad memory", params: "m=bad,t=1,p=1"},
		{name: "zero memory", params: "m=0,t=1,p=1"},
		{name: "memory overflow", params: "m=4294967296,t=1,p=1"},
		{name: "bad iterations", params: "m=8,t=bad,p=1"},
		{name: "zero iterations", params: "m=8,t=0,p=1"},
		{name: "bad parallelism", params: "m=8,t=1,p=bad"},
		{name: "zero parallelism", params: "m=8,t=1,p=0"},
		{name: "parallelism overflow", params: "m=8,t=1,p=256"},
		{name: "missing parallelism after duplicate key", params: "m=8,m=9,t=1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodeArgon2idParams(tt.params)
			if !errors.Is(err, ErrInvalidHash) {
				t.Fatalf("expected invalid hash, got %v", err)
			}
		})
	}
}

func testArgon2idConfig() Argon2idConfig {
	return Argon2idConfig{
		MemoryKiB:   8,
		Iterations:  1,
		Parallelism: 1,
	}
}
