package systemsettings

import (
	"context"

	apptx "github.com/yorukot/netstamp/internal/controller/application/tx"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainsystem "github.com/yorukot/netstamp/internal/domain/system"
)

type Repository interface {
	GetSystemSettingsByKeys(ctx context.Context, keys []string) ([]domainsystem.Setting, error)
	UpsertSystemSetting(ctx context.Context, setting domainsystem.Setting) (domainsystem.Setting, error)
	DeleteSystemSetting(ctx context.Context, key string) error
	CreateSystemSettingAuditEvent(ctx context.Context, key, action string, updatedByUserID *string) error
	GetSystemSettingRevision(ctx context.Context, resource string) (int64, error)
	LockSystemSettingRevision(ctx context.Context, resource string) (int64, error)
	BumpSystemSettingRevision(ctx context.Context, resource string) (int64, error)
}

type SystemAdminChecker interface {
	IsSystemAdmin(ctx context.Context, userID string) (bool, error)
}

type SecretCipher interface {
	Encrypt(plaintext string) (ciphertext, nonce []byte, err error)
	Decrypt(ciphertext, nonce []byte) (string, error)
}

type OIDCReadinessChecker interface {
	Check(ctx context.Context, issuerURL string) error
}

type SMTPTestUserRepository interface {
	GetUserByID(ctx context.Context, userID string) (identity.User, error)
}

type SMTPTester interface {
	SendTestEmail(ctx context.Context, recipient string, settings SMTPRuntimeSettings) error
}

type Transactor = apptx.Transactor

type StoredSetting = domainsystem.Setting
