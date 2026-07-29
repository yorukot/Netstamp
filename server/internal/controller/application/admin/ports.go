package admin

import (
	"context"
)

type Repository interface {
	IsSystemAdmin(ctx context.Context, userID string) (bool, error)
	ListSystemAdmins(ctx context.Context) ([]SystemAdmin, error)
	ListManagedUsers(ctx context.Context) ([]ManagedUser, error)
	GrantSystemAdminByEmail(ctx context.Context, email string) (SystemAdmin, error)
	GrantSystemAdminByUserID(ctx context.Context, userID string) (ManagedUser, error)
	RevokeSystemAdminIfNotLast(ctx context.Context, userID string) (SystemAdminRevokeResult, error)
	CountActiveSystemAdmins(ctx context.Context) (int64, error)
	SetManagedUserDisabledAt(ctx context.Context, userID string, disabled bool) (ManagedUser, error)
	SetManagedUserPasswordHash(ctx context.Context, userID, passwordHash string) (ManagedUser, error)
	ExportData(ctx context.Context) (DataExport, error)
	ImportData(ctx context.Context, export DataExport) (DataImportResult, error)
	CreateSystemSettingAuditEvent(ctx context.Context, key, action string, updatedByUserID *string) error
}

type SessionRepository interface {
	RevokeUserSessions(ctx context.Context, userID, reason string) error
}

type APITokenRevoker interface {
	RevokeUserTokens(ctx context.Context, userID, reason string) error
}

type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
}

type ManagedPasswordRepository interface {
	ClearManagedUserPassword(ctx context.Context, userID string) (ManagedUser, error)
}

type AuthenticationMethodRepository interface {
	CountUserAuthenticationMethods(ctx context.Context, userID string) (bool, int64, error)
}
