package admin

import "context"

type Repository interface {
	IsSystemAdmin(ctx context.Context, userID string) (bool, error)
	ListSystemAdmins(ctx context.Context) ([]SystemAdmin, error)
	GrantSystemAdminByEmail(ctx context.Context, email string) (SystemAdmin, error)
	RevokeSystemAdminIfNotLast(ctx context.Context, userID string) (SystemAdminRevokeResult, error)
	ListSystemSettings(ctx context.Context) ([]StoredSetting, error)
	UpsertSystemSetting(ctx context.Context, setting StoredSetting) (StoredSetting, error)
	CreateSystemSettingAuditEvent(ctx context.Context, key, action string, updatedByUserID *string) error
}

type SecretCipher interface {
	Encrypt(plaintext string) (ciphertext, nonce []byte, err error)
	Decrypt(ciphertext, nonce []byte) (string, error)
}
