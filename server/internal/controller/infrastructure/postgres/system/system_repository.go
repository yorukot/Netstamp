package pgsystem

import (
	"context"
	"encoding/json"
	"errors"

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

func mapListSystemAdmin(row sqlc.ListSystemAdminsRow) domainsystem.AdminUser {
	return domainsystem.AdminUser{
		ID:              row.ID.String(),
		Email:           row.Email,
		DisplayName:     row.DisplayName,
		EmailVerifiedAt: row.EmailVerifiedAt,
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
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
		GrantedAt:       row.GrantedAt,
	}
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
