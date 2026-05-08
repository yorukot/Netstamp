package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/yorukot/netstamp/internal/config"
)

func TestRunStatusUsesConfigConnectionString(t *testing.T) {
	cfg := testMigrateConfig()
	var stderr bytes.Buffer
	var gotDialect string
	var gotDriver string
	var gotConnectionString string
	var gotDir string
	var pinged bool
	var closed bool
	var statusCalled bool

	runner := migrateRunner{
		stderr: &stderr,
		loadConfig: func() (config.Config, error) {
			return cfg, nil
		},
		setDialect: func(dialect string) error {
			gotDialect = dialect
			return nil
		},
		openDB: func(driver, connectionString string) (*sql.DB, error) {
			gotDriver = driver
			gotConnectionString = connectionString
			return testSQLDB(), nil
		},
		pingDB: func(ctx context.Context, _ *sql.DB) error {
			if _, ok := ctx.Deadline(); !ok {
				t.Fatal("expected ping context to have a deadline")
			}
			pinged = true
			return nil
		},
		closeDB: func(_ *sql.DB) error {
			closed = true
			return nil
		},
		status: func(_ *sql.DB, dir string) error {
			statusCalled = true
			gotDir = dir
			return nil
		},
	}

	code := runner.run([]string{"-command", "status"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", code, stderr.String())
	}
	if gotDialect != "postgres" {
		t.Fatalf("expected postgres dialect, got %q", gotDialect)
	}
	if gotDriver != "pgx" {
		t.Fatalf("expected pgx driver, got %q", gotDriver)
	}
	if gotConnectionString != cfg.Database.ConnectionString() {
		t.Fatalf("expected config connection string %q, got %q", cfg.Database.ConnectionString(), gotConnectionString)
	}
	if gotDir != "db/migrations" {
		t.Fatalf("expected default migration dir, got %q", gotDir)
	}
	if !pinged {
		t.Fatal("expected database ping")
	}
	if !statusCalled {
		t.Fatal("expected status migration command")
	}
	if !closed {
		t.Fatal("expected database close")
	}
}

func TestRunDatabaseConnectionStringFlagOverridesConfig(t *testing.T) {
	var stderr bytes.Buffer
	var gotConnectionString string
	var gotDir string
	var upCalled bool

	runner := migrateRunner{
		stderr: &stderr,
		loadConfig: func() (config.Config, error) {
			return testMigrateConfig(), nil
		},
		setDialect: func(string) error {
			return nil
		},
		openDB: func(_, connectionString string) (*sql.DB, error) {
			gotConnectionString = connectionString
			return testSQLDB(), nil
		},
		pingDB: func(context.Context, *sql.DB) error {
			return nil
		},
		closeDB: func(*sql.DB) error {
			return nil
		},
		up: func(_ *sql.DB, dir string) error {
			upCalled = true
			gotDir = dir
			return nil
		},
	}

	const overrideConnectionString = "postgres://override:secret@db.internal:15432/netstamp?sslmode=require"
	code := runner.run([]string{
		"-database-connection-string", overrideConnectionString,
		"-dir", "custom/migrations",
		"-command", "up",
	})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d with stderr %q", code, stderr.String())
	}
	if gotConnectionString != overrideConnectionString {
		t.Fatalf("expected override connection string %q, got %q", overrideConnectionString, gotConnectionString)
	}
	if gotDir != "custom/migrations" {
		t.Fatalf("expected custom migration dir, got %q", gotDir)
	}
	if !upCalled {
		t.Fatal("expected up migration command")
	}
}

func TestRunReturnsLoadConfigError(t *testing.T) {
	var stderr bytes.Buffer

	runner := migrateRunner{
		stderr: &stderr,
		loadConfig: func() (config.Config, error) {
			return config.Config{}, errors.New("decode config: '' has invalid keys: grpc_addr")
		},
		openDB: func(string, string) (*sql.DB, error) {
			t.Fatal("openDB should not be called when config loading fails")
			return nil, errors.New("openDB should not be called")
		},
	}

	code := runner.run([]string{"-command", "status"})
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "load config: decode config: '' has invalid keys: grpc_addr") {
		t.Fatalf("expected config error in stderr, got %q", stderr.String())
	}
}

func TestRunRejectsUnsupportedMigrationCommandBeforeOpeningDatabase(t *testing.T) {
	var stderr bytes.Buffer

	runner := migrateRunner{
		stderr: &stderr,
		loadConfig: func() (config.Config, error) {
			return testMigrateConfig(), nil
		},
		setDialect: func(string) error {
			t.Fatal("setDialect should not be called for unsupported commands")
			return nil
		},
		openDB: func(string, string) (*sql.DB, error) {
			t.Fatal("openDB should not be called for unsupported commands")
			return nil, errors.New("openDB should not be called")
		},
	}

	code := runner.run([]string{"-command", "sideways"})
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), `migration failed: unsupported migration command "sideways"`) {
		t.Fatalf("expected unsupported command error in stderr, got %q", stderr.String())
	}
}

func TestRunReturnsMigrationError(t *testing.T) {
	var stderr bytes.Buffer
	var closed bool

	runner := migrateRunner{
		stderr: &stderr,
		loadConfig: func() (config.Config, error) {
			return testMigrateConfig(), nil
		},
		setDialect: func(string) error {
			return nil
		},
		openDB: func(string, string) (*sql.DB, error) {
			return testSQLDB(), nil
		},
		pingDB: func(context.Context, *sql.DB) error {
			return nil
		},
		closeDB: func(*sql.DB) error {
			closed = true
			return nil
		},
		down: func(*sql.DB, string) error {
			return errors.New("goose down failed")
		},
	}

	code := runner.run([]string{"-command", "down"})
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stderr.String(), "migration failed: goose down failed") {
		t.Fatalf("expected migration error in stderr, got %q", stderr.String())
	}
	if !closed {
		t.Fatal("expected database close after migration failure")
	}
}

func testMigrateConfig() config.Config {
	return config.Config{
		Database: config.DatabaseConfig{
			Host:     "db.internal",
			Port:     15432,
			User:     "netstamp",
			Password: "secret",
			Name:     "netstamp_test",
			SSLMode:  "require",
		},
	}
}

func testSQLDB() *sql.DB {
	return &sql.DB{}
}
