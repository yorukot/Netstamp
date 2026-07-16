package system

import (
	"encoding/json"
	"time"
)

type AdminUser struct {
	ID              string
	Email           string
	DisplayName     string
	EmailVerifiedAt *time.Time
	DisabledAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	GrantedAt       time.Time
}

type ManagedUser struct {
	ID              string
	Email           string
	DisplayName     string
	EmailVerifiedAt *time.Time
	DisabledAt      *time.Time
	IsSystemAdmin   bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	GrantedAt       *time.Time
	HasPassword     bool
}

type AdminRevokeResult struct {
	AdminCount     int64
	TargetWasAdmin bool
	Revoked        bool
}

type DataExport struct {
	Format     string
	ExportedAt time.Time
	Tables     map[string][]RawDataRow
}

type DataImportResult struct {
	Format         string
	ImportedTables int
	ImportedRows   int
}

type RawDataRow = json.RawMessage
