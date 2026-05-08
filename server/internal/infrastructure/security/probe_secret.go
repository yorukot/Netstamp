package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

const ProbeSecretByteLength = 32

type ProbeSecretGenerator struct {
	reader io.Reader
}

func NewProbeSecretGenerator() *ProbeSecretGenerator {
	return &ProbeSecretGenerator{reader: rand.Reader}
}

func (g *ProbeSecretGenerator) GenerateProbeSecret() (string, string, error) {
	reader := g.reader
	if reader == nil {
		reader = rand.Reader
	}

	secretBytes := make([]byte, ProbeSecretByteLength)
	if _, err := io.ReadFull(reader, secretBytes); err != nil {
		return "", "", fmt.Errorf("generate probe secret: %w", err)
	}

	plaintext := base64.RawURLEncoding.EncodeToString(secretBytes)
	return plaintext, HashProbeSecret(plaintext), nil
}

func HashProbeSecret(secret string) string {
	digest := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(digest[:])
}

func VerifyProbeSecret(secret, expectedHash string) bool {
	actualHash := HashProbeSecret(secret)
	return subtle.ConstantTimeCompare([]byte(actualHash), []byte(expectedHash)) == 1
}
