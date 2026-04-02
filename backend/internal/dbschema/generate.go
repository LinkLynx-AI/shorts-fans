package dbschema

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"gopkg.in/yaml.v3"
)

const (
	defaultAdminDatabase = "postgres"
	defaultTempPrefix    = "schema_export"
)

// Options はスキーマ YAML 生成に必要な入力を表します。
type Options struct {
	AppEnv       string
	PostgresDSN  string
	MigrationDir string
	OutputPath   string
}

// Generate は migration から一時 DB を構築し、人間が読みやすい YAML スキーマを出力します。
func Generate(ctx context.Context, opts Options) (err error) {
	if strings.TrimSpace(opts.PostgresDSN) == "" {
		return fmt.Errorf("POSTGRES_DSN is required")
	}
	if strings.EqualFold(strings.TrimSpace(opts.AppEnv), "production") {
		return fmt.Errorf("schema export is disabled in production")
	}

	migrationDir, err := filepath.Abs(opts.MigrationDir)
	if err != nil {
		return fmt.Errorf("resolve migration dir: %w", err)
	}

	outputPath, err := filepath.Abs(opts.OutputPath)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}

	migrations, err := listMigrationFiles(migrationDir)
	if err != nil {
		return err
	}

	baseConfig, err := pgx.ParseConfig(opts.PostgresDSN)
	if err != nil {
		return fmt.Errorf("parse POSTGRES_DSN: %w", err)
	}

	adminConn, err := connectAdmin(ctx, baseConfig)
	if err != nil {
		return err
	}
	defer adminConn.Close(ctx)

	tempDatabaseName, err := newTempDatabaseName(defaultTempPrefix)
	if err != nil {
		return fmt.Errorf("build temp database name: %w", err)
	}

	if err := createDatabase(ctx, adminConn, tempDatabaseName); err != nil {
		return err
	}

	cleanupTempDatabase := func() error {
		return dropDatabase(ctx, adminConn, tempDatabaseName)
	}
	defer func() {
		cleanupErr := cleanupTempDatabase()
		if cleanupErr == nil {
			return
		}
		if err == nil {
			err = cleanupErr
			return
		}
		err = errors.Join(err, cleanupErr)
	}()

	tempConfig := baseConfig.Copy()
	tempConfig.Database = tempDatabaseName
	tempDSN := tempConfig.ConnString()

	if err := applyMigrations(tempDSN, migrationDir); err != nil {
		return err
	}

	catalog, err := inspectCatalog(ctx, tempDSN)
	if err != nil {
		return err
	}

	document := buildDocument(documentSource{
		Generator:    "go run ./cmd/schema",
		MigrationDir: filepath.Clean(opts.MigrationDir),
		Migrations:   migrations,
	}, catalog)

	content, err := marshalDocument(document)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(outputPath, content, 0o644); err != nil {
		return fmt.Errorf("write schema file: %w", err)
	}

	return nil
}

func listMigrationFiles(migrationDir string) ([]string, error) {
	entries, err := filepath.Glob(filepath.Join(migrationDir, "*.up.sql"))
	if err != nil {
		return nil, fmt.Errorf("list migration files: %w", err)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no up migrations found in %s", migrationDir)
	}

	sort.Strings(entries)

	migrations := make([]string, 0, len(entries))
	for _, entry := range entries {
		migrations = append(migrations, filepath.Base(entry))
	}

	return migrations, nil
}

func connectAdmin(ctx context.Context, baseConfig *pgx.ConnConfig) (*pgx.Conn, error) {
	candidates := uniqueStrings(defaultAdminDatabase, baseConfig.Database, "template1")

	var lastErr error
	for _, databaseName := range candidates {
		adminConfig := baseConfig.Copy()
		adminConfig.Database = databaseName

		conn, err := pgx.ConnectConfig(ctx, adminConfig)
		if err == nil {
			return conn, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf("connect admin database: %w", lastErr)
}

func createDatabase(ctx context.Context, adminConn *pgx.Conn, databaseName string) error {
	if _, err := adminConn.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{databaseName}.Sanitize()); err != nil {
		return fmt.Errorf("create temp database %q: %w", databaseName, err)
	}

	return nil
}

func dropDatabase(ctx context.Context, adminConn *pgx.Conn, databaseName string) error {
	if _, err := adminConn.Exec(
		ctx,
		`SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1
			AND pid <> pg_backend_pid()`,
		databaseName,
	); err != nil {
		return fmt.Errorf("terminate temp database sessions for %q: %w", databaseName, err)
	}

	if _, err := adminConn.Exec(ctx, "DROP DATABASE IF EXISTS "+pgx.Identifier{databaseName}.Sanitize()); err != nil {
		return fmt.Errorf("drop temp database %q: %w", databaseName, err)
	}

	return nil
}

func applyMigrations(tempDSN string, migrationDir string) (err error) {
	db, err := sql.Open("pgx", tempDSN)
	if err != nil {
		return fmt.Errorf("open temp database for migrations: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create postgres migration driver: %w", err)
	}

	sourceURL := (&url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(migrationDir),
	}).String()

	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}
	defer func() {
		sourceErr, databaseErr := m.Close()
		if err == nil && sourceErr != nil {
			err = fmt.Errorf("close migration source: %w", sourceErr)
		}
		if err == nil && databaseErr != nil {
			err = fmt.Errorf("close migration database: %w", databaseErr)
		}
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations to temp database: %w", err)
	}

	return nil
}

func marshalDocument(doc schemaDocument) ([]byte, error) {
	content, err := yaml.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("marshal schema yaml: %w", err)
	}

	header := "# Code generated by go run ./cmd/schema. DO NOT EDIT.\n\n"
	return append([]byte(header), content...), nil
}

func newTempDatabaseName(prefix string) (string, error) {
	var randomBytes [4]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	return fmt.Sprintf("%s_%s", prefix, strings.ToLower(hex.EncodeToString(randomBytes[:]))), nil
}

func uniqueStrings(values ...string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}

		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result
}
