package pgsystem

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres"
	"github.com/yorukot/netstamp/internal/controller/infrastructure/postgres/sqlc"
	"github.com/yorukot/netstamp/internal/domain/identity"
	domainsystem "github.com/yorukot/netstamp/internal/domain/system"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: sqlc.New(pool)}
}

func (r *Repository) IsSystemAdmin(ctx context.Context, userIDValue string) (bool, error) {
	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return false, err
	}
	return postgres.Queries(ctx, r.queries).IsSystemAdmin(ctx, userID)
}

func (r *Repository) ListSystemAdmins(ctx context.Context) ([]domainsystem.AdminUser, error) {
	rows, err := postgres.Queries(ctx, r.queries).ListSystemAdmins(ctx)
	if err != nil {
		return nil, err
	}
	admins := make([]domainsystem.AdminUser, 0, len(rows))
	for _, row := range rows {
		admins = append(admins, mapListSystemAdmin(row))
	}
	return admins, nil
}

func (r *Repository) ListManagedUsers(ctx context.Context) ([]domainsystem.ManagedUser, error) {
	rows, err := postgres.Queries(ctx, r.queries).ListManagedUsers(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]domainsystem.ManagedUser, 0, len(rows))
	for _, row := range rows {
		users = append(users, mapManagedUser(row))
	}
	return users, nil
}

func (r *Repository) GrantSystemAdminByEmail(ctx context.Context, email string) (domainsystem.AdminUser, error) {
	row, err := postgres.Queries(ctx, r.queries).GrantSystemAdminByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainsystem.AdminUser{}, identity.ErrUserNotFound
		}
		return domainsystem.AdminUser{}, err
	}
	return mapGrantedSystemAdmin(row), nil
}

func (r *Repository) GrantSystemAdminByUserID(ctx context.Context, userIDValue string) (domainsystem.ManagedUser, error) {
	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return domainsystem.ManagedUser{}, err
	}
	row, err := postgres.Queries(ctx, r.queries).GrantSystemAdminByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainsystem.ManagedUser{}, identity.ErrUserNotFound
		}
		return domainsystem.ManagedUser{}, err
	}
	return mapGrantedManagedUser(row), nil
}

func (r *Repository) RevokeSystemAdminIfNotLast(ctx context.Context, userIDValue string) (domainsystem.AdminRevokeResult, error) {
	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return domainsystem.AdminRevokeResult{}, err
	}
	row, err := postgres.Queries(ctx, r.queries).RevokeSystemAdminIfNotLast(ctx, userID)
	if err != nil {
		return domainsystem.AdminRevokeResult{}, err
	}
	return domainsystem.AdminRevokeResult{
		AdminCount:     row.AdminCount,
		TargetWasAdmin: row.TargetWasAdmin,
		Revoked:        row.Revoked,
	}, nil
}

func (r *Repository) CountActiveSystemAdmins(ctx context.Context) (int64, error) {
	return postgres.Queries(ctx, r.queries).CountActiveSystemAdmins(ctx)
}

func (r *Repository) SetManagedUserDisabledAt(ctx context.Context, userIDValue string, disabled bool) (domainsystem.ManagedUser, error) {
	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return domainsystem.ManagedUser{}, err
	}
	var disabledAt *time.Time
	if disabled {
		now := time.Now().UTC()
		disabledAt = &now
	}
	row, err := postgres.Queries(ctx, r.queries).SetManagedUserDisabledAt(ctx, sqlc.SetManagedUserDisabledAtParams{
		ID:         userID,
		DisabledAt: disabledAt,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainsystem.ManagedUser{}, identity.ErrUserNotFound
		}
		return domainsystem.ManagedUser{}, err
	}
	return mapDisabledManagedUser(row), nil
}

func (r *Repository) SetManagedUserPasswordHash(ctx context.Context, userIDValue, passwordHash string) (domainsystem.ManagedUser, error) {
	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return domainsystem.ManagedUser{}, err
	}
	row, err := postgres.Queries(ctx, r.queries).SetManagedUserPasswordHash(ctx, sqlc.SetManagedUserPasswordHashParams{
		UserID:       userID,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainsystem.ManagedUser{}, identity.ErrUserNotFound
		}
		return domainsystem.ManagedUser{}, err
	}
	return mapPasswordManagedUser(row), nil
}

func (r *Repository) ClearManagedUserPassword(ctx context.Context, userIDValue string) (domainsystem.ManagedUser, error) {
	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return domainsystem.ManagedUser{}, err
	}
	row, err := postgres.Queries(ctx, r.queries).ClearManagedUserPassword(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domainsystem.ManagedUser{}, identity.ErrUserNotFound
	}
	if err != nil {
		return domainsystem.ManagedUser{}, err
	}
	return mapClearPasswordManagedUser(row), nil
}

func (r *Repository) GrantFirstSystemAdminIfNone(ctx context.Context, userIDValue string) (bool, error) {
	userID, err := postgres.ParseUUID(userIDValue, identity.ErrUserNotFound)
	if err != nil {
		return false, err
	}
	return postgres.Queries(ctx, r.queries).GrantFirstSystemAdminIfNone(ctx, userID)
}

func (r *Repository) ListSystemSettings(ctx context.Context) ([]domainsystem.Setting, error) {
	rows, err := postgres.Queries(ctx, r.queries).ListSystemSettings(ctx)
	if err != nil {
		return nil, err
	}
	settings := make([]domainsystem.Setting, 0, len(rows))
	for _, row := range rows {
		settings = append(settings, mapSetting(row))
	}
	return settings, nil
}

func (r *Repository) GetSystemSettingsByKeys(ctx context.Context, keys []string) ([]domainsystem.Setting, error) {
	rows, err := postgres.Queries(ctx, r.queries).GetSystemSettingsByKeys(ctx, keys)
	if err != nil {
		return nil, err
	}
	settings := make([]domainsystem.Setting, 0, len(rows))
	for _, row := range rows {
		settings = append(settings, mapSetting(row))
	}
	return settings, nil
}

func (r *Repository) UpsertSystemSetting(ctx context.Context, setting domainsystem.Setting) (domainsystem.Setting, error) {
	var updatedByUserID *uuid.UUID
	if setting.UpdatedByUserID != nil {
		parsed, err := postgres.ParseUUID(*setting.UpdatedByUserID, identity.ErrUserNotFound)
		if err != nil {
			return domainsystem.Setting{}, err
		}
		updatedByUserID = &parsed
	}

	row, err := postgres.Queries(ctx, r.queries).UpsertSystemSetting(ctx, sqlc.UpsertSystemSettingParams{
		Key:                 setting.Key,
		Value:               []byte(setting.Value),
		EncryptedValue:      setting.EncryptedValue,
		EncryptedValueNonce: setting.EncryptedValueNonce,
		Secret:              setting.Secret,
		UpdatedByUserID:     updatedByUserID,
	})
	if err != nil {
		return domainsystem.Setting{}, err
	}
	return mapSetting(row), nil
}

func (r *Repository) DeleteSystemSetting(ctx context.Context, key string) error {
	return postgres.Queries(ctx, r.queries).DeleteSystemSetting(ctx, key)
}

func (r *Repository) CreateSystemSettingAuditEvent(ctx context.Context, key, action string, updatedByUserIDValue *string) error {
	var updatedByUserID *uuid.UUID
	if updatedByUserIDValue != nil {
		parsed, err := postgres.ParseUUID(*updatedByUserIDValue, identity.ErrUserNotFound)
		if err != nil {
			return err
		}
		updatedByUserID = &parsed
	}
	return postgres.Queries(ctx, r.queries).CreateSystemSettingAuditEvent(ctx, sqlc.CreateSystemSettingAuditEventParams{
		Key:             key,
		Action:          action,
		UpdatedByUserID: updatedByUserID,
	})
}

func (r *Repository) LockSystemSettingsResource(ctx context.Context, resource string) error {
	if _, ok := postgres.TxFromContext(ctx); !ok {
		return errors.New("lock system settings resource outside transaction")
	}
	return postgres.Queries(ctx, r.queries).LockSystemSettingsResource(ctx, resource)
}

func mapListSystemAdmin(row sqlc.ListSystemAdminsRow) domainsystem.AdminUser {
	return domainsystem.AdminUser{
		ID:              row.ID.String(),
		Email:           row.Email,
		DisplayName:     row.DisplayName,
		EmailVerifiedAt: row.EmailVerifiedAt,
		DisabledAt:      row.DisabledAt,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		GrantedAt:       row.GrantedAt,
	}
}

func mapGrantedSystemAdmin(row sqlc.GrantSystemAdminByEmailRow) domainsystem.AdminUser {
	return domainsystem.AdminUser{
		ID:              row.ID.String(),
		Email:           row.Email,
		DisplayName:     row.DisplayName,
		EmailVerifiedAt: row.EmailVerifiedAt,
		DisabledAt:      row.DisabledAt,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		GrantedAt:       row.GrantedAt,
	}
}

func mapManagedUser(row sqlc.ListManagedUsersRow) domainsystem.ManagedUser {
	return domainsystem.ManagedUser{
		ID:              row.ID.String(),
		Email:           row.Email,
		DisplayName:     row.DisplayName,
		EmailVerifiedAt: row.EmailVerifiedAt,
		DisabledAt:      row.DisabledAt,
		IsSystemAdmin:   row.IsSystemAdmin,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		GrantedAt:       row.GrantedAt,
		HasPassword:     row.HasPassword,
	}
}

func mapGrantedManagedUser(row sqlc.GrantSystemAdminByUserIDRow) domainsystem.ManagedUser {
	return domainsystem.ManagedUser{
		ID:              row.ID.String(),
		Email:           row.Email,
		DisplayName:     row.DisplayName,
		EmailVerifiedAt: row.EmailVerifiedAt,
		DisabledAt:      row.DisabledAt,
		IsSystemAdmin:   row.IsSystemAdmin,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		GrantedAt:       &row.GrantedAt,
		HasPassword:     row.HasPassword,
	}
}

func mapDisabledManagedUser(row sqlc.SetManagedUserDisabledAtRow) domainsystem.ManagedUser {
	return domainsystem.ManagedUser{
		ID:              row.ID.String(),
		Email:           row.Email,
		DisplayName:     row.DisplayName,
		EmailVerifiedAt: row.EmailVerifiedAt,
		DisabledAt:      row.DisabledAt,
		IsSystemAdmin:   row.IsSystemAdmin,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		GrantedAt:       row.GrantedAt,
		HasPassword:     row.HasPassword,
	}
}

func mapPasswordManagedUser(row sqlc.SetManagedUserPasswordHashRow) domainsystem.ManagedUser {
	return domainsystem.ManagedUser{
		ID:              row.ID.String(),
		Email:           row.Email,
		DisplayName:     row.DisplayName,
		EmailVerifiedAt: row.EmailVerifiedAt,
		DisabledAt:      row.DisabledAt,
		IsSystemAdmin:   row.IsSystemAdmin,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		GrantedAt:       row.GrantedAt,
		HasPassword:     row.HasPassword,
	}
}

func mapClearPasswordManagedUser(row sqlc.ClearManagedUserPasswordRow) domainsystem.ManagedUser {
	return domainsystem.ManagedUser{ID: row.ID.String(), Email: row.Email, DisplayName: row.DisplayName, EmailVerifiedAt: row.EmailVerifiedAt, DisabledAt: row.DisabledAt, IsSystemAdmin: row.IsSystemAdmin, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, GrantedAt: row.GrantedAt, HasPassword: row.HasPassword}
}

func mapSetting(row sqlc.SystemSetting) domainsystem.Setting {
	var value json.RawMessage
	if len(row.Value) > 0 {
		value = append(json.RawMessage(nil), row.Value...)
	}

	var updatedByUserID *string
	if row.UpdatedByUserID != nil {
		value := row.UpdatedByUserID.String()
		updatedByUserID = &value
	}

	return domainsystem.Setting{
		Key:                 row.Key,
		Value:               value,
		EncryptedValue:      append([]byte(nil), row.EncryptedValue...),
		EncryptedValueNonce: append([]byte(nil), row.EncryptedValueNonce...),
		Secret:              row.Secret,
		UpdatedByUserID:     updatedByUserID,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}
}
