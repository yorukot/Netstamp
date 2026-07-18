package pgsystem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
	pool    *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: sqlc.New(pool), pool: pool}
}

const (
	dataExportFormat         = "netstamp.admin.data.v4"
	legacyDataExportFormatV3 = "netstamp.admin.data.v3"
	legacyDataExportFormatV2 = "netstamp.admin.data.v2"
	legacyDataExportFormatV1 = "netstamp.admin.data.v1"
)

var dataExportTables = []string{
	"users",
	"password_credentials",
	"user_identities",
	"api_tokens",
	"projects",
	"project_members",
	"project_invites",
	"probes",
	"probe_credentials",
	"probe_statuses",
	"checks",
	"ping_check_configs",
	"tcp_check_configs",
	"traceroute_check_configs",
	"http_check_configs",
	"labels",
	"check_labels",
	"probe_labels",
	"probe_check_assignments",
	"ping_results",
	"tcp_results",
	"traceroute_results",
	"traceroute_result_hops",
	"traceroute_sampled_runs_1m",
	"http_results",
	"notifications",
	"alert_rules",
	"alert_notifications",
	"alert_incidents",
	"notification_outbox",
	"public_status_pages",
	"public_status_page_elements",
	"public_status_page_element_assignments",
	"assignment_refresh_jobs",
	"password_reset_tokens",
	"email_verification_tokens",
	"system_user_roles",
	"system_settings",
	"system_setting_audit_events",
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

func (r *Repository) ExportData(ctx context.Context) (domainsystem.DataExport, error) {
	tables := make(map[string][]domainsystem.RawDataRow, len(dataExportTables))
	for _, table := range dataExportTables {
		rows, err := r.exportTable(ctx, table)
		if err != nil {
			return domainsystem.DataExport{}, err
		}
		tables[table] = rows
	}
	return domainsystem.DataExport{
		Format:     dataExportFormat,
		ExportedAt: time.Now().UTC(),
		Tables:     tables,
	}, nil
}

func (r *Repository) ImportData(ctx context.Context, export domainsystem.DataExport) (domainsystem.DataImportResult, error) {
	export, err := normalizeDataImport(export)
	if err != nil {
		return domainsystem.DataImportResult{}, err
	}

	var result domainsystem.DataImportResult
	err = pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		if _, truncateErr := tx.Exec(ctx, truncateDataExportTablesSQL()); truncateErr != nil {
			return truncateErr
		}
		for _, table := range dataExportTables {
			rows := export.Tables[table]
			result.ImportedTables++
			result.ImportedRows += len(rows)
			if len(rows) == 0 {
				continue
			}
			payload, marshalErr := json.Marshal(rows)
			if marshalErr != nil {
				return domainsystem.ErrDataImportInvalid
			}
			if !json.Valid(payload) {
				return domainsystem.ErrDataImportInvalid
			}
			if _, importErr := tx.Exec(ctx, importTableSQL(table), payload); importErr != nil {
				return importErr
			}
		}
		return nil
	})
	if err != nil {
		return domainsystem.DataImportResult{}, err
	}
	result.Format = dataExportFormat
	return result, nil
}

func normalizeDataImport(export domainsystem.DataExport) (domainsystem.DataExport, error) {
	if !supportedDataImportFormat(export.Format) {
		return domainsystem.DataExport{}, domainsystem.ErrDataImportInvalid
	}
	if err := upgradeLegacyDataImport(&export); err != nil {
		return domainsystem.DataExport{}, domainsystem.ErrDataImportInvalid
	}
	if err := validateDataImportTables(export.Tables); err != nil {
		return domainsystem.DataExport{}, err
	}
	return export, nil
}

func supportedDataImportFormat(format string) bool {
	switch format {
	case dataExportFormat, legacyDataExportFormatV3, legacyDataExportFormatV2, legacyDataExportFormatV1:
		return true
	default:
		return false
	}
}

func upgradeLegacyDataImport(export *domainsystem.DataExport) error {
	if export.Format == dataExportFormat {
		return nil
	}
	if export.Tables == nil {
		return domainsystem.ErrDataImportInvalid
	}
	if export.Format == legacyDataExportFormatV1 {
		if _, exists := export.Tables["api_tokens"]; !exists {
			export.Tables["api_tokens"] = []domainsystem.RawDataRow{}
		}
	}
	if export.Format == legacyDataExportFormatV1 || export.Format == legacyDataExportFormatV2 {
		if err := upgradeLegacyUserCredentials(export); err != nil {
			return err
		}
	}
	for _, table := range []string{"http_check_configs", "http_results"} {
		if _, exists := export.Tables[table]; !exists {
			export.Tables[table] = []domainsystem.RawDataRow{}
		}
	}
	export.Format = dataExportFormat
	return nil
}

func validateDataImportTables(tables map[string][]domainsystem.RawDataRow) error {
	allowed := make(map[string]struct{}, len(dataExportTables))
	for _, table := range dataExportTables {
		allowed[table] = struct{}{}
		if _, ok := tables[table]; !ok {
			return domainsystem.ErrDataImportInvalid
		}
	}
	for table := range tables {
		if _, ok := allowed[table]; !ok {
			return domainsystem.ErrDataImportInvalid
		}
	}
	return nil
}

func upgradeLegacyUserCredentials(export *domainsystem.DataExport) error {
	users, ok := export.Tables["users"]
	if !ok {
		return domainsystem.ErrDataImportInvalid
	}
	credentials := make([]domainsystem.RawDataRow, 0, len(users))
	upgradedUsers := make([]domainsystem.RawDataRow, 0, len(users))
	for _, raw := range users {
		var row map[string]json.RawMessage
		if err := json.Unmarshal(raw, &row); err != nil {
			return domainsystem.ErrDataImportInvalid
		}
		id, idOK := row["id"]
		passwordHash, hashOK := row["password_hash"]
		createdAt, createdOK := row["created_at"]
		updatedAt, updatedOK := row["updated_at"]
		if !idOK || !hashOK || !createdOK || !updatedOK {
			return domainsystem.ErrDataImportInvalid
		}
		credential, err := json.Marshal(map[string]json.RawMessage{
			"user_id": id, "password_hash": passwordHash, "created_at": createdAt, "updated_at": updatedAt,
		})
		if err != nil {
			return domainsystem.ErrDataImportInvalid
		}
		delete(row, "password_hash")
		upgraded, err := json.Marshal(row)
		if err != nil {
			return domainsystem.ErrDataImportInvalid
		}
		credentials = append(credentials, credential)
		upgradedUsers = append(upgradedUsers, upgraded)
	}
	export.Tables["users"] = upgradedUsers
	export.Tables["password_credentials"] = credentials
	export.Tables["user_identities"] = []domainsystem.RawDataRow{}
	export.Format = dataExportFormat
	return nil
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

func (r *Repository) exportTable(ctx context.Context, table string) ([]domainsystem.RawDataRow, error) {
	var raw []byte
	query := fmt.Sprintf(
		"SELECT COALESCE(jsonb_agg(to_jsonb(t)), '[]'::jsonb) FROM (SELECT * FROM %s) AS t",
		quoteTable(table),
	)
	if err := r.pool.QueryRow(ctx, query).Scan(&raw); err != nil {
		return nil, err
	}

	var rows []json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}
	out := make([]domainsystem.RawDataRow, 0, len(rows))
	for _, row := range rows {
		if !json.Valid(row) {
			return nil, domainsystem.ErrDataImportInvalid
		}
		out = append(out, domainsystem.RawDataRow(append([]byte(nil), row...)))
	}
	return out, nil
}

func truncateDataExportTablesSQL() string {
	quoted := make([]string, 0, len(dataExportTables))
	for _, table := range dataExportTables {
		quoted = append(quoted, quoteTable(table))
	}
	return "TRUNCATE TABLE " + strings.Join(quoted, ", ") + " RESTART IDENTITY CASCADE"
}

func importTableSQL(table string) string {
	quoted := quoteTable(table)
	return fmt.Sprintf("INSERT INTO %s SELECT * FROM jsonb_populate_recordset(NULL::%s, $1::jsonb)", quoted, quoted)
}

func quoteTable(table string) string {
	return pgx.Identifier{"public", table}.Sanitize()
}
