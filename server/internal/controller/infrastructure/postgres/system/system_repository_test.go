package pgsystem

import (
	"encoding/json"
	"testing"

	domainsystem "github.com/yorukot/netstamp/internal/domain/system"
)

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
