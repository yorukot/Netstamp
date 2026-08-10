//go:build integration

package e2e

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
)

const (
	v010BaselineMigrationFile = "00001_v0_1_0.sql"
	v010BaselineVersion       = int64(1)
	v010IrreversibleError     = "Netstamp v0.1.0 baseline is irreversible"
)

func TestV010BaselineMigrationContract(t *testing.T) {
	adminDatabaseURL := strings.TrimSpace(os.Getenv(testDatabaseURLEnv))
	if adminDatabaseURL == "" {
		t.Skipf("set %s to run backend migration integration tests", testDatabaseURLEnv)
	}

	migrationDir := isolatedV010BaselineMigrationsDir(t)
	adminDB := openAdminDatabase(t, adminDatabaseURL)
	t.Cleanup(func() {
		if err := adminDB.Close(); err != nil {
			t.Errorf("close admin database: %v", err)
		}
	})

	databaseName := "netstamp_baseline_" + randomHex(t, 6)
	createDatabase(t, adminDB, databaseName)
	t.Cleanup(func() {
		dropDatabase(t, adminDB, databaseName)
	})

	databaseURL := databaseURLWithName(t, adminDatabaseURL, databaseName)
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open baseline database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close baseline database: %v", err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping baseline database: %v", err)
	}
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("set goose dialect: %v", err)
	}

	if err := goose.UpContext(ctx, db, migrationDir); err != nil {
		t.Fatalf("apply v0.1.0 baseline: %v", err)
	}
	assertGooseVersion(ctx, t, db, v010BaselineVersion)
	assertV010KeyRelations(ctx, t, db)
	assertRelationAbsent(ctx, t, db, "public.system_setting_revisions")
	assertRelationAbsent(ctx, t, db, "public.system_setting_revisions_rollback_epoch")

	var userID string
	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO users (email, display_name) VALUES ('baseline@example.com', 'Baseline') RETURNING id::text`,
	).Scan(&userID); err != nil {
		t.Fatalf("seed baseline user: %v", err)
	}

	if err := goose.UpContext(ctx, db, migrationDir); err != nil {
		t.Fatalf("reapply v0.1.0 baseline: %v", err)
	}
	assertGooseVersion(ctx, t, db, v010BaselineVersion)
	assertUserExists(ctx, t, db, userID)

	downErr := goose.DownContext(ctx, db, migrationDir)
	if downErr == nil {
		t.Fatal("roll back v0.1.0 baseline succeeded, want irreversible baseline error")
	}
	if !strings.Contains(downErr.Error(), v010IrreversibleError) {
		t.Fatalf("roll back v0.1.0 baseline error = %q, want it to contain %q", downErr, v010IrreversibleError)
	}

	assertGooseVersion(ctx, t, db, v010BaselineVersion)
	assertV010KeyRelations(ctx, t, db)
	assertRelationAbsent(ctx, t, db, "public.system_setting_revisions")
	assertRelationAbsent(ctx, t, db, "public.system_setting_revisions_rollback_epoch")
	assertUserExists(ctx, t, db, userID)
}

func isolatedV010BaselineMigrationsDir(t *testing.T) string {
	t.Helper()

	sourcePath := filepath.Join(migrationsDir(t), v010BaselineMigrationFile)
	contents, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read v0.1.0 baseline migration: %v", err)
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, v010BaselineMigrationFile), contents, 0o600); err != nil {
		t.Fatalf("copy v0.1.0 baseline migration: %v", err)
	}
	return dir
}

func assertGooseVersion(ctx context.Context, t *testing.T, db *sql.DB, want int64) {
	t.Helper()

	got, err := goose.GetDBVersionContext(ctx, db)
	if err != nil {
		t.Fatalf("read Goose migration version: %v", err)
	}
	if got != want {
		t.Fatalf("Goose migration version = %d, want %d", got, want)
	}
}

func assertV010KeyRelations(ctx context.Context, t *testing.T, db *sql.DB) {
	t.Helper()

	for _, relation := range []string{
		"public.users",
		"public.projects",
		"public.checks",
		"public.external_auth_flows",
		"public.system_settings",
		"public.public_status_pages",
		"public.ping_results",
		"public.tcp_results",
		"public.http_results",
		"public.traceroute_results",
		"public.ping_result_rollups_1m",
		"public.traceroute_sampled_runs_1m",
	} {
		var exists bool
		if err := db.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, relation).Scan(&exists); err != nil {
			t.Fatalf("check relation %q: %v", relation, err)
		}
		if !exists {
			t.Errorf("baseline relation %q does not exist", relation)
		}
	}
}

func assertRelationAbsent(ctx context.Context, t *testing.T, db *sql.DB, relation string) {
	t.Helper()

	var exists bool
	if err := db.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, relation).Scan(&exists); err != nil {
		t.Fatalf("check relation %q: %v", relation, err)
	}
	if exists {
		t.Errorf("legacy relation %q exists", relation)
	}
}

func assertUserExists(ctx context.Context, t *testing.T, db *sql.DB, userID string) {
	t.Helper()

	var exists bool
	if err := db.QueryRowContext(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM users WHERE id = $1::uuid)`,
		userID,
	).Scan(&exists); err != nil {
		t.Fatalf("check baseline user: %v", err)
	}
	if !exists {
		t.Error("baseline user was removed")
	}
}
