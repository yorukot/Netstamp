package pgsystem

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	domainsystem "github.com/yorukot/netstamp/internal/domain/system"
)

func TestImportDataLocksSystemSettingsResourcesInCanonicalOrder(t *testing.T) {
	locker := &recordingSystemSettingsResourceLocker{}

	if err := lockSystemSettingsResources(context.Background(), locker); err != nil {
		t.Fatalf("lock system settings resources: %v", err)
	}
	if !slices.Equal(locker.resources, []string{
		"access",
		"smtp",
		"auth.oidc",
		"auth.google",
		"auth.github",
	}) {
		t.Fatalf("unexpected lock order: %#v", locker.resources)
	}
}

func TestImportDataStopsWhenSystemSettingsResourceLockFails(t *testing.T) {
	expected := errors.New("lock failed")
	locker := &recordingSystemSettingsResourceLocker{
		failResource: "auth.oidc",
		err:          expected,
	}

	err := lockSystemSettingsResources(context.Background(), locker)
	if !errors.Is(err, expected) {
		t.Fatalf("expected lock failure, got %v", err)
	}
	if !strings.Contains(err.Error(), `"auth.oidc"`) {
		t.Fatalf("expected failed resource in error, got %v", err)
	}
	if !slices.Equal(locker.resources, []string{"access", "smtp", "auth.oidc"}) {
		t.Fatalf("unexpected lock attempts: %#v", locker.resources)
	}
}

func TestLockSystemSettingsResourceRequiresTransaction(t *testing.T) {
	repo := &Repository{}

	err := repo.LockSystemSettingsResource(context.Background(), "access")
	if err == nil {
		t.Fatal("expected lock outside transaction to fail")
	}
	if got, want := err.Error(), "lock system settings resource outside transaction"; got != want {
		t.Fatalf("unexpected error: got %q, want %q", got, want)
	}
}

type recordingSystemSettingsResourceLocker struct {
	resources    []string
	failResource string
	err          error
}

func (l *recordingSystemSettingsResourceLocker) LockSystemSettingsResource(
	_ context.Context,
	resource string,
) error {
	l.resources = append(l.resources, resource)
	if resource == l.failResource {
		return l.err
	}
	return nil
}

func TestNormalizeDataImportUpgradesLegacyPasswordCredentials(t *testing.T) {
	tables := emptyDataExportTables()
	delete(tables, "password_credentials")
	delete(tables, "user_identities")
	tables["users"] = []domainsystem.RawDataRow{json.RawMessage(`{
		"id":"11111111-1111-1111-1111-111111111111",
		"email":"person@example.com",
		"display_name":"Person",
		"password_hash":"argon2id-hash",
		"created_at":"2026-07-16T10:00:00Z",
		"updated_at":"2026-07-16T10:00:00Z"
	}`)}

	normalized, err := normalizeDataImport(domainsystem.DataExport{Format: legacyDataExportFormatV2, Tables: tables})
	if err != nil {
		t.Fatalf("normalize legacy export: %v", err)
	}
	if normalized.Format != dataExportFormat {
		t.Fatalf("expected format %q, got %q", dataExportFormat, normalized.Format)
	}
	if len(normalized.Tables["password_credentials"]) != 1 || len(normalized.Tables["user_identities"]) != 0 {
		t.Fatalf("unexpected upgraded authentication tables: credentials=%d identities=%d", len(normalized.Tables["password_credentials"]), len(normalized.Tables["user_identities"]))
	}
	var user map[string]json.RawMessage
	if err := json.Unmarshal(normalized.Tables["users"][0], &user); err != nil {
		t.Fatalf("decode upgraded user: %v", err)
	}
	if _, exists := user["password_hash"]; exists {
		t.Fatal("legacy password hash should be removed from the users row")
	}
	var credential struct {
		UserID       string `json:"user_id"`
		PasswordHash string `json:"password_hash"`
	}
	if err := json.Unmarshal(normalized.Tables["password_credentials"][0], &credential); err != nil {
		t.Fatalf("decode upgraded credential: %v", err)
	}
	if credential.UserID != "11111111-1111-1111-1111-111111111111" || credential.PasswordHash != "argon2id-hash" {
		t.Fatalf("unexpected credential: %#v", credential)
	}
}

func TestNormalizeDataImportAcceptsPasswordlessCurrentExport(t *testing.T) {
	tables := emptyDataExportTables()
	tables["users"] = []domainsystem.RawDataRow{json.RawMessage(`{
		"id":"11111111-1111-1111-1111-111111111111",
		"email":"sso@example.com",
		"display_name":"SSO User",
		"created_at":"2026-07-16T10:00:00Z",
		"updated_at":"2026-07-16T10:00:00Z"
	}`)}

	if _, err := normalizeDataImport(domainsystem.DataExport{Format: dataExportFormat, Tables: tables}); err != nil {
		t.Fatalf("normalize passwordless current export: %v", err)
	}
}

func TestNormalizeDataImportUpgradesV3WithHTTPResultTables(t *testing.T) {
	tables := emptyDataExportTables()
	delete(tables, "http_check_configs")
	delete(tables, "http_results")

	normalized, err := normalizeDataImport(domainsystem.DataExport{Format: legacyDataExportFormatV3, Tables: tables})
	if err != nil {
		t.Fatalf("normalize v3 export: %v", err)
	}
	if normalized.Format != dataExportFormat {
		t.Fatalf("expected format %q, got %q", dataExportFormat, normalized.Format)
	}
	if len(normalized.Tables["http_check_configs"]) != 0 || len(normalized.Tables["http_results"]) != 0 {
		t.Fatalf("expected empty HTTP tables in upgraded export: %#v", normalized.Tables)
	}
}

func emptyDataExportTables() map[string][]domainsystem.RawDataRow {
	tables := make(map[string][]domainsystem.RawDataRow, len(dataExportTables))
	for _, table := range dataExportTables {
		tables[table] = []domainsystem.RawDataRow{}
	}
	return tables
}
