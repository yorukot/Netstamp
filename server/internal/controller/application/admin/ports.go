package admin

import "context"

type Repository interface {
	IsSystemAdmin(ctx context.Context, userID string) (bool, error)
	ListSystemSettings(ctx context.Context) ([]StoredSetting, error)
	UpsertSystemSetting(ctx context.Context, setting StoredSetting) (StoredSetting, error)
	CreateSystemSettingAuditEvent(ctx context.Context, key, action string, updatedByUserID *string) error
}

type SecretCipher interface {
	Encrypt(plaintext string) (ciphertext, nonce []byte, err error)
	Decrypt(ciphertext, nonce []byte) (string, error)
}
