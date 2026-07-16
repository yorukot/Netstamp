package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

const apiTokenPrefix = "nst_pat_"

type APITokenManager struct{ hashKey []byte }

func NewAPITokenManager(hashKey string) *APITokenManager {
	return &APITokenManager{hashKey: []byte(hashKey)}
}

func (m *APITokenManager) Generate() (string, string, error) {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", "", err
	}
	raw := apiTokenPrefix + base64.RawURLEncoding.EncodeToString(random)
	return raw, raw[len(raw)-8:], nil
}

func (m *APITokenManager) Hash(rawToken string) []byte {
	mac := hmac.New(sha256.New, m.hashKey)
	_, _ = mac.Write([]byte(rawToken))
	return mac.Sum(nil)
}
