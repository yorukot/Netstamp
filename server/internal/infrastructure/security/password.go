package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2idVersion = argon2.Version
	saltLength      = 16
	keyLength       = 32
)

var (
	ErrPasswordMismatch = errors.New("password mismatch")
	ErrInvalidHash      = errors.New("invalid password hash")
)

type Argon2idConfig struct {
	MemoryKiB   int
	Iterations  int
	Parallelism int
}

type Argon2idPasswordHasher struct {
	cfg Argon2idConfig
}

func NewArgon2idPasswordHasher(cfg Argon2idConfig) *Argon2idPasswordHasher {
	return &Argon2idPasswordHasher{cfg: cfg}
}

func (h *Argon2idPasswordHasher) Hash(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		uint32(h.cfg.Iterations),
		uint32(h.cfg.MemoryKiB),
		uint8(h.cfg.Parallelism),
		keyLength,
	)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2idVersion,
		h.cfg.MemoryKiB,
		h.cfg.Iterations,
		h.cfg.Parallelism,
		encodedSalt,
		encodedHash,
	), nil
}

func (h *Argon2idPasswordHasher) Compare(password, encoded string) error {
	params, salt, expectedHash, err := decodeArgon2idHash(encoded)
	if err != nil {
		return err
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		uint32(params.Iterations),
		uint32(params.MemoryKiB),
		uint8(params.Parallelism),
		uint32(len(expectedHash)),
	)

	if subtle.ConstantTimeCompare(actualHash, expectedHash) != 1 {
		return ErrPasswordMismatch
	}

	return nil
}

func decodeArgon2idHash(encoded string) (Argon2idConfig, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return Argon2idConfig{}, nil, nil, ErrInvalidHash
	}

	// parts[0] is empty because the string starts with "$".
	if parts[1] != "argon2id" {
		return Argon2idConfig{}, nil, nil, ErrInvalidHash
	}

	versionPart := parts[2]
	if !strings.HasPrefix(versionPart, "v=") {
		return Argon2idConfig{}, nil, nil, ErrInvalidHash
	}

	version, err := strconv.Atoi(strings.TrimPrefix(versionPart, "v="))
	if err != nil {
		return Argon2idConfig{}, nil, nil, ErrInvalidHash
	}

	if version != argon2.Version {
		return Argon2idConfig{}, nil, nil, ErrInvalidHash
	}

	cfg, err := decodeArgon2idParams(parts[3])
	if err != nil {
		return Argon2idConfig{}, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return Argon2idConfig{}, nil, nil, ErrInvalidHash
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return Argon2idConfig{}, nil, nil, ErrInvalidHash
	}

	return cfg, salt, hash, nil
}

func decodeArgon2idParams(encoded string) (Argon2idConfig, error) {
	values := strings.Split(encoded, ",")
	if len(values) != 3 {
		return Argon2idConfig{}, ErrInvalidHash
	}

	cfg := Argon2idConfig{}

	for _, value := range values {
		keyValue := strings.SplitN(value, "=", 2)
		if len(keyValue) != 2 {
			return Argon2idConfig{}, ErrInvalidHash
		}

		key := keyValue[0]
		rawValue := keyValue[1]

		parsedValue, err := strconv.Atoi(rawValue)
		if err != nil {
			return Argon2idConfig{}, ErrInvalidHash
		}

		switch key {
		case "m":
			cfg.MemoryKiB = parsedValue
		case "t":
			cfg.Iterations = parsedValue
		case "p":
			cfg.Parallelism = parsedValue
		default:
			return Argon2idConfig{}, ErrInvalidHash
		}
	}

	if cfg.MemoryKiB <= 0 || cfg.Iterations <= 0 || cfg.Parallelism <= 0 {
		return Argon2idConfig{}, ErrInvalidHash
	}

	return cfg, nil
}
