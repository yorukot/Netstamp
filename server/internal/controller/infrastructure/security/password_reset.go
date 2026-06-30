package security

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

const passwordResetTokenBytes = 32

type PasswordResetTokenManager struct{}

func NewPasswordResetTokenManager() *PasswordResetTokenManager {
	return &PasswordResetTokenManager{}
}

func (m *PasswordResetTokenManager) Generate(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	buffer := make([]byte, passwordResetTokenBytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func (m *PasswordResetTokenManager) Hash(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])
}
