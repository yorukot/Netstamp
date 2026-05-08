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
	ErrPasswordMismatch      = errors.New("password mismatch")
	ErrInvalidHash           = errors.New("invalid password hash")
	ErrInvalidArgon2idConfig = errors.New("invalid argon2id config")
)

type Argon2idConfig struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
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
	if h.cfg.MemoryKiB == 0 || h.cfg.Iterations == 0 || h.cfg.Parallelism == 0 {
		return "", ErrInvalidArgon2idConfig
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.cfg.Iterations,
		h.cfg.MemoryKiB,
		h.cfg.Parallelism,
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
	if len(expectedHash) != keyLength {
		return ErrInvalidHash
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.MemoryKiB,
		params.Parallelism,
		keyLength,
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

		switch key {
		case "m":
			memoryKiB, err := parseArgon2idUint32Param(rawValue)
			if err != nil {
				return Argon2idConfig{}, err
			}
			cfg.MemoryKiB = memoryKiB
		case "t":
			iterations, err := parseArgon2idUint32Param(rawValue)
			if err != nil {
				return Argon2idConfig{}, err
			}
			cfg.Iterations = iterations
		case "p":
			parallelism, err := parseArgon2idUint8Param(rawValue)
			if err != nil {
				return Argon2idConfig{}, err
			}
			cfg.Parallelism = parallelism
		default:
			return Argon2idConfig{}, ErrInvalidHash
		}
	}

	if cfg.MemoryKiB == 0 || cfg.Iterations == 0 || cfg.Parallelism == 0 {
		return Argon2idConfig{}, ErrInvalidHash
	}

	return cfg, nil
}

func parseArgon2idUint32Param(value string) (uint32, error) {
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil || parsed == 0 {
		return 0, ErrInvalidHash
	}

	return uint32(parsed), nil
}

func parseArgon2idUint8Param(value string) (uint8, error) {
	parsed, err := strconv.ParseUint(value, 10, 8)
	if err != nil || parsed == 0 {
		return 0, ErrInvalidHash
	}

	return uint8(parsed), nil
}
