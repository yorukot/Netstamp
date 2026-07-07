package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

type SecretCipher struct {
	gcm cipher.AEAD
}

func NewSecretCipher(key string) (*SecretCipher, error) {
	if key == "" {
		return nil, errors.New("secret cipher key is required")
	}
	sum := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &SecretCipher{gcm: gcm}, nil
}

func (c *SecretCipher) Encrypt(plaintext string) ([]byte, []byte, error) {
	if c == nil || c.gcm == nil {
		return nil, nil, errors.New("secret cipher is unavailable")
	}
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	ciphertext := c.gcm.Seal(nil, nonce, []byte(plaintext), nil)
	return ciphertext, nonce, nil
}

func (c *SecretCipher) Decrypt(ciphertext, nonce []byte) (string, error) {
	if c == nil || c.gcm == nil {
		return "", errors.New("secret cipher is unavailable")
	}
	plaintext, err := c.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
