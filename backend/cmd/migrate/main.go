package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/config"
)

func main() {
	command, err := parseCommand(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cfg := config.Load()
	if cfg.PostgresDSN == "" {
		fmt.Fprintln(os.Stderr, "POSTGRES_DSN is required")
		os.Exit(1)
	}

	if err := run(command, cfg.PostgresDSN); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseCommand(args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("usage: go run ./cmd/migrate [up|down|version]")
	}

	switch args[0] {
	case "up", "down", "version":
		return args[0], nil
	default:
		return "", fmt.Errorf("unsupported command %q", args[0])
	}
}

func run(command string, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open database for migrations: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create postgres migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	switch command {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("apply migrations: %w", err)
		}
		return nil
	case "down":
		if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("rollback migration: %w", err)
		}
		return nil
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				fmt.Println("no migrations applied")
				return nil
			}

			return fmt.Errorf("read migration version: %w", err)
		}

		fmt.Printf("%d dirty=%t\n", version, dirty)
		return nil
	default:
		return fmt.Errorf("unsupported command %q", command)
	}
}
