//go:build integration

package e2e

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
)

const (
	addSystemSettingRevisionsMigration  int64 = 202607290001
	dropSystemSettingRevisionsMigration int64 = 202607300001
)

func TestSystemSettingRevisionsMigrationRoundTrip(t *testing.T) {
	suite := newAPISuite(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, openErr := sql.Open("pgx", suite.pool.Config().ConnString())
	if openErr != nil {
		t.Fatalf("open migration round-trip database: %v", openErr)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close migration round-trip database: %v", err)
		}
	})
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping migration round-trip database: %v", err)
	}
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}

	dir := migrationsDir(t)
	assertSystemSettingRevisionsTableExists(ctx, t, db, false)
	assertSystemSettingRevisionsRollbackSequenceExists(ctx, t, db, true)

	// Restore the pre-removal schema and give it a representative revision
	// history. The following Up must retain the old maximum for a safe rollback.
	if err := goose.DownToContext(ctx, db, dir, addSystemSettingRevisionsMigration); err != nil {
		t.Fatalf("prepare pre-removal system setting revisions schema: %v", err)
	}
	assertSystemSettingRevisionsTableExists(ctx, t, db, true)
	assertSystemSettingRevisionsRollbackSequenceExists(ctx, t, db, false)

	var initialMaxRevision int64
	if err := db.QueryRowContext(ctx, `SELECT max(revision) FROM system_setting_revisions`).Scan(&initialMaxRevision); err != nil {
		t.Fatalf("read initial maximum system setting revision: %v", err)
	}
	priorMaxRevision := initialMaxRevision + 100
	result, execErr := db.ExecContext(
		ctx,
		`UPDATE system_setting_revisions SET revision = $1 WHERE resource = 'access'`,
		priorMaxRevision,
	)
	if execErr != nil {
		t.Fatalf("seed pre-removal system setting revision: %v", execErr)
	}
	if rows, err := result.RowsAffected(); err != nil {
		t.Fatalf("read seeded system setting revision row count: %v", err)
	} else if rows != 1 {
		t.Fatalf("seeded system setting revision rows = %d, want 1", rows)
	}

	if err := goose.UpToContext(ctx, db, dir, dropSystemSettingRevisionsMigration); err != nil {
		t.Fatalf("apply system setting revisions removal migration: %v", err)
	}
	assertSystemSettingRevisionsTableExists(ctx, t, db, false)
	assertSystemSettingRevisionsRollbackSequenceExists(ctx, t, db, true)

	if err := goose.DownToContext(ctx, db, dir, addSystemSettingRevisionsMigration); err != nil {
		t.Fatalf("roll back system setting revisions removal migration: %v", err)
	}
	assertSystemSettingRevisionsTableExists(ctx, t, db, true)
	assertSystemSettingRevisionsRollbackSequenceExists(ctx, t, db, false)
	assertSystemSettingRevisionsSchema(ctx, t, db)
	assertCanonicalSystemSettingRevisions(ctx, t, db, priorMaxRevision)

	if err := goose.UpToContext(ctx, db, dir, dropSystemSettingRevisionsMigration); err != nil {
		t.Fatalf("reapply system setting revisions removal migration: %v", err)
	}
	assertSystemSettingRevisionsTableExists(ctx, t, db, false)
	assertSystemSettingRevisionsRollbackSequenceExists(ctx, t, db, true)
}

func assertSystemSettingRevisionsTableExists(
	ctx context.Context,
	t *testing.T,
	db *sql.DB,
	want bool,
) {
	t.Helper()

	var exists bool
	if err := db.QueryRowContext(
		ctx,
		`SELECT to_regclass('public.system_setting_revisions') IS NOT NULL`,
	).Scan(&exists); err != nil {
		t.Fatalf("check system_setting_revisions existence: %v", err)
	}
	if exists != want {
		t.Fatalf("system_setting_revisions existence = %t, want %t", exists, want)
	}
}

func assertSystemSettingRevisionsRollbackSequenceExists(
	ctx context.Context,
	t *testing.T,
	db *sql.DB,
	want bool,
) {
	t.Helper()

	var exists bool
	if err := db.QueryRowContext(
		ctx,
		`SELECT to_regclass('public.system_setting_revisions_rollback_epoch') IS NOT NULL`,
	).Scan(&exists); err != nil {
		t.Fatalf("check system_setting_revisions_rollback_epoch existence: %v", err)
	}
	if exists != want {
		t.Fatalf("system_setting_revisions_rollback_epoch existence = %t, want %t", exists, want)
	}
}

func assertSystemSettingRevisionsSchema(ctx context.Context, t *testing.T, db *sql.DB) {
	t.Helper()

	type column struct {
		name       string
		dataType   string
		notNull    bool
		hasDefault bool
	}
	wantColumns := []column{
		{name: "resource", dataType: "text", notNull: true, hasDefault: false},
		{name: "revision", dataType: "bigint", notNull: true, hasDefault: true},
	}

	rows, queryErr := db.QueryContext(ctx, `
		SELECT column_name, data_type, is_nullable = 'NO', column_default IS NOT NULL
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND table_name = 'system_setting_revisions'
		ORDER BY ordinal_position
	`)
	if queryErr != nil {
		t.Fatalf("query system_setting_revisions columns: %v", queryErr)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Errorf("close system_setting_revisions column rows: %v", err)
		}
	}()

	var gotColumns []column
	for rows.Next() {
		var got column
		if err := rows.Scan(&got.name, &got.dataType, &got.notNull, &got.hasDefault); err != nil {
			t.Fatalf("scan system_setting_revisions column: %v", err)
		}
		gotColumns = append(gotColumns, got)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate system_setting_revisions columns: %v", err)
	}
	if len(gotColumns) != len(wantColumns) {
		t.Fatalf("system_setting_revisions columns = %#v, want %#v", gotColumns, wantColumns)
	}
	for index := range wantColumns {
		if gotColumns[index] != wantColumns[index] {
			t.Errorf("system_setting_revisions column %d = %#v, want %#v", index, gotColumns[index], wantColumns[index])
		}
	}

	wantConstraints := map[string]string{
		"system_setting_revisions_pkey":               "primarykey(resource)",
		"system_setting_revisions_resource_not_empty": "check(length(btrim(resource))>0)",
		"system_setting_revisions_revision_positive":  "check(revision>0)",
	}
	constraintRows, constraintQueryErr := db.QueryContext(ctx, `
		SELECT conname, pg_get_constraintdef(oid, true)
		FROM pg_constraint
		WHERE conrelid = 'public.system_setting_revisions'::regclass
		ORDER BY conname
	`)
	if constraintQueryErr != nil {
		t.Fatalf("query system_setting_revisions constraints: %v", constraintQueryErr)
	}
	defer func() {
		if err := constraintRows.Close(); err != nil {
			t.Errorf("close system_setting_revisions constraint rows: %v", err)
		}
	}()

	gotConstraints := make(map[string]string)
	for constraintRows.Next() {
		var name string
		var definition string
		if err := constraintRows.Scan(&name, &definition); err != nil {
			t.Fatalf("scan system_setting_revisions constraint: %v", err)
		}
		gotConstraints[name] = normalizeConstraintDefinition(definition)
	}
	if err := constraintRows.Err(); err != nil {
		t.Fatalf("iterate system_setting_revisions constraints: %v", err)
	}
	if len(gotConstraints) != len(wantConstraints) {
		t.Fatalf("system_setting_revisions constraints = %#v, want %#v", gotConstraints, wantConstraints)
	}
	for name, wantDefinition := range wantConstraints {
		if gotDefinition := gotConstraints[name]; gotDefinition != wantDefinition {
			t.Errorf("system_setting_revisions constraint %q = %q, want %q", name, gotDefinition, wantDefinition)
		}
	}

	var defaultRevision int64
	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO system_setting_revisions (resource) VALUES ('test.default') RETURNING revision`,
	).Scan(&defaultRevision); err != nil {
		t.Fatalf("insert system setting revision with default: %v", err)
	}
	if defaultRevision != 1 {
		t.Errorf("default system setting revision = %d, want 1", defaultRevision)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM system_setting_revisions WHERE resource = 'test.default'`); err != nil {
		t.Fatalf("delete default system setting revision test row: %v", err)
	}
}

func assertCanonicalSystemSettingRevisions(
	ctx context.Context,
	t *testing.T,
	db *sql.DB,
	priorMaxRevision int64,
) {
	t.Helper()

	wantResources := []string{"access", "auth.github", "auth.google", "auth.oidc", "smtp"}
	rows, queryErr := db.QueryContext(ctx, `
		SELECT resource, revision
		FROM system_setting_revisions
		ORDER BY resource
	`)
	if queryErr != nil {
		t.Fatalf("query canonical system setting revisions: %v", queryErr)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.Errorf("close canonical system setting revision rows: %v", err)
		}
	}()

	var gotResources []string
	var rollbackRevision int64
	for rows.Next() {
		var resource string
		var revision int64
		if err := rows.Scan(&resource, &revision); err != nil {
			t.Fatalf("scan canonical system setting revision: %v", err)
		}
		gotResources = append(gotResources, resource)
		if revision <= 0 {
			t.Errorf("canonical resource %q revision = %d, want a positive revision", resource, revision)
		}
		if rollbackRevision == 0 {
			rollbackRevision = revision
		} else if revision != rollbackRevision {
			t.Errorf("canonical resource %q revision = %d, want shared rollback revision %d", resource, revision, rollbackRevision)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate canonical system setting revisions: %v", err)
	}
	if strings.Join(gotResources, ",") != strings.Join(wantResources, ",") {
		t.Errorf("canonical system setting resources = %v, want %v", gotResources, wantResources)
	}
	if rollbackRevision <= priorMaxRevision {
		t.Errorf(
			"rollback system setting revision = %d, want greater than pre-removal maximum %d",
			rollbackRevision,
			priorMaxRevision,
		)
	}
}

func normalizeConstraintDefinition(value string) string {
	return strings.NewReplacer(" ", "", "\n", "", "\t", "").Replace(strings.ToLower(value))
}
