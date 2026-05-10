package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/yorukot/netstamp/internal/controller/config"
)

func main() {
	os.Exit(run())
}

func run() int {
	return newMigrateRunner(os.Stderr).run(os.Args[1:])
}

type migrateRunner struct {
	stderr io.Writer

	loadConfig func() (config.Config, error)
	setDialect func(string) error
	openDB     func(string, string) (*sql.DB, error)
	pingDB     func(context.Context, *sql.DB) error
	closeDB    func(*sql.DB) error

	status func(*sql.DB, string) error
	up     func(*sql.DB, string) error
	down   func(*sql.DB, string) error
}

func newMigrateRunner(stderr io.Writer) migrateRunner {
	return migrateRunner{
		stderr:     stderr,
		loadConfig: config.Load,
		setDialect: goose.SetDialect,
		openDB:     sql.Open,
		pingDB: func(ctx context.Context, db *sql.DB) error {
			return db.PingContext(ctx)
		},
		closeDB: func(db *sql.DB) error {
			return db.Close()
		},
		status: func(db *sql.DB, dir string) error {
			return goose.Status(db, dir)
		},
		up: func(db *sql.DB, dir string) error {
			return goose.Up(db, dir)
		},
		down: func(db *sql.DB, dir string) error {
			return goose.Down(db, dir)
		},
	}
}

func (r migrateRunner) run(args []string) int {
	r = r.withDefaults()

	flagSet := flag.NewFlagSet("migrate", flag.ContinueOnError)
	flagSet.SetOutput(r.stderr)

	databaseConnectionString := flagSet.String("database-connection-string", "", "PostgreSQL connection string")
	dir := flagSet.String("dir", "db/migrations", "migration directory")
	command := flagSet.String("command", "status", "migration command: status, up, or down")
	if err := flagSet.Parse(args); err != nil {
		return 2
	}

	cfg, err := r.loadConfig()
	if err != nil {
		_, _ = fmt.Fprintf(r.stderr, "load config: %v\n", err)
		return 1
	}

	migrate, ok := r.migrationFunc(*command)
	if !ok {
		_, _ = fmt.Fprintf(r.stderr, "migration failed: unsupported migration command %q\n", *command)
		return 1
	}

	resolvedDatabaseConnectionString := *databaseConnectionString
	if resolvedDatabaseConnectionString == "" {
		resolvedDatabaseConnectionString = cfg.Database.ConnectionString()
	}

	if resolvedDatabaseConnectionString == "" {
		_, _ = fmt.Fprintln(r.stderr, "database connection settings are required")
		return 2
	}

	err = r.setDialect("postgres")
	if err != nil {
		_, _ = fmt.Fprintf(r.stderr, "set migration dialect: %v\n", err)
		return 1
	}

	db, err := r.openDB("pgx", resolvedDatabaseConnectionString)
	if err != nil {
		_, _ = fmt.Fprintf(r.stderr, "open database: %v\n", err)
		return 1
	}
	defer func() {
		if closeErr := r.closeDB(db); closeErr != nil {
			_, _ = fmt.Fprintf(r.stderr, "close database: %v\n", closeErr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = r.pingDB(ctx, db)
	if err != nil {
		_, _ = fmt.Fprintf(r.stderr, "ping database: %v\n", err)
		return 1
	}

	err = migrate(db, *dir)
	if err != nil {
		_, _ = fmt.Fprintf(r.stderr, "migration failed: %v\n", err)
		return 1
	}

	return 0
}

func (r migrateRunner) withDefaults() migrateRunner {
	if r.stderr == nil {
		r.stderr = io.Discard
	}
	if r.loadConfig == nil {
		r.loadConfig = config.Load
	}
	if r.setDialect == nil {
		r.setDialect = goose.SetDialect
	}
	if r.openDB == nil {
		r.openDB = sql.Open
	}
	if r.pingDB == nil {
		r.pingDB = func(ctx context.Context, db *sql.DB) error {
			return db.PingContext(ctx)
		}
	}
	if r.closeDB == nil {
		r.closeDB = func(db *sql.DB) error {
			return db.Close()
		}
	}
	if r.status == nil {
		r.status = func(db *sql.DB, dir string) error {
			return goose.Status(db, dir)
		}
	}
	if r.up == nil {
		r.up = func(db *sql.DB, dir string) error {
			return goose.Up(db, dir)
		}
	}
	if r.down == nil {
		r.down = func(db *sql.DB, dir string) error {
			return goose.Down(db, dir)
		}
	}
	return r
}

func (r migrateRunner) migrationFunc(command string) (func(*sql.DB, string) error, bool) {
	switch command {
	case "status":
		return r.status, true
	case "up":
		return r.up, true
	case "down":
		return r.down, true
	default:
		return nil, false
	}
}
