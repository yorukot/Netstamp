package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/yorukot/netstamp/internal/config"
)

func main() {
	os.Exit(run())
}

func run() int {
	databaseConnectionString := flag.String("database-connection-string", "", "PostgreSQL connection string")
	dir := flag.String("dir", "db/migrations", "migration directory")
	command := flag.String("command", "status", "migration command: status, up, or down")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		return 1
	}

	resolvedDatabaseConnectionString := *databaseConnectionString
	if resolvedDatabaseConnectionString == "" {
		resolvedDatabaseConnectionString = cfg.Database.ConnectionString()
	}

	if resolvedDatabaseConnectionString == "" {
		_, _ = fmt.Fprintln(os.Stderr, "database connection settings are required")
		return 2
	}

	err = goose.SetDialect("postgres")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "set migration dialect: %v\n", err)
		return 1
	}

	db, err := sql.Open("pgx", resolvedDatabaseConnectionString)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "open database: %v\n", err)
		return 1
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "ping database: %v\n", err)
		return 1
	}

	switch *command {
	case "status":
		err = goose.Status(db, *dir)
	case "up":
		err = goose.Up(db, *dir)
	case "down":
		err = goose.Down(db, *dir)
	default:
		err = fmt.Errorf("unsupported migration command %q", *command)
	}
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
		return 1
	}

	return 0
}
