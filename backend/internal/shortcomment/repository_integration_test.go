package shortcomment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	pgmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

const shortCommentIntegrationPostgresDSNEnv = "POSTGRES_DSN"

type shortCommentIntegrationScenario struct {
	AuthorUserID uuid.UUID
	CreatorID    uuid.UUID
	MainAssetID  uuid.UUID
	MainID       uuid.UUID
	ShortAssetID uuid.UUID
	ShortID      uuid.UUID
}

func TestRepositoryIntegrationCreateAndListPublicShortComments(t *testing.T) {
	t.Parallel()

	ctx, pool, cleanup := newShortCommentTestDatabase(t)
	defer cleanup()

	repo := NewRepository(pool)
	scenario := seedShortCommentPublicShort(t, ctx, pool)

	created, err := repo.CreateComment(ctx, CreateInput{
		AuthorUserID: scenario.AuthorUserID,
		Body:         "  created comment  ",
		ShortID:      scenario.ShortID,
	})
	if err != nil {
		t.Fatalf("CreateComment() error = %v, want nil", err)
	}
	if created.Body != "created comment" {
		t.Fatalf("CreateComment() body got %q want trimmed body", created.Body)
	}
	if created.Author.DisplayName != "Fan Viewer" {
		t.Fatalf("CreateComment() author display name got %q want Fan Viewer", created.Author.DisplayName)
	}
	if created.Author.Handle != "fanviewer" {
		t.Fatalf("CreateComment() author handle got %q want fanviewer", created.Author.Handle)
	}
	if created.Author.AvatarURL == nil || *created.Author.AvatarURL != "https://cdn.example.com/fan-avatar.jpg" {
		t.Fatalf("CreateComment() avatar got %#v want seeded avatar", created.Author.AvatarURL)
	}

	if _, err := pool.Exec(ctx, `DELETE FROM app.short_comments WHERE id = $1`, created.ID); err != nil {
		t.Fatalf("delete created comment error = %v, want nil", err)
	}

	emptyPage, emptyCursor, err := repo.ListComments(ctx, scenario.ShortID, nil, 2)
	if err != nil {
		t.Fatalf("ListComments(empty public short) error = %v, want nil", err)
	}
	if len(emptyPage) != 0 {
		t.Fatalf("ListComments(empty public short) len got %d want 0", len(emptyPage))
	}
	if emptyCursor != nil {
		t.Fatalf("ListComments(empty public short) cursor got %#v want nil", emptyCursor)
	}

	createdAt := time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
	olderCreatedAt := createdAt.Add(-time.Minute)
	lowTieID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	highTieID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	olderID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	insertShortCommentRow(t, ctx, pool, scenario.ShortID, scenario.AuthorUserID, lowTieID, "same timestamp low id", createdAt)
	insertShortCommentRow(t, ctx, pool, scenario.ShortID, scenario.AuthorUserID, highTieID, "same timestamp high id", createdAt)
	insertShortCommentRow(t, ctx, pool, scenario.ShortID, scenario.AuthorUserID, olderID, "older", olderCreatedAt)

	firstPage, nextCursor, err := repo.ListComments(ctx, scenario.ShortID, nil, 2)
	if err != nil {
		t.Fatalf("ListComments(first page) error = %v, want nil", err)
	}
	if len(firstPage) != 2 {
		t.Fatalf("ListComments(first page) len got %d want 2", len(firstPage))
	}
	if firstPage[0].ID != highTieID || firstPage[1].ID != lowTieID {
		t.Fatalf("ListComments(first page) ids got [%s %s] want [%s %s]", firstPage[0].ID, firstPage[1].ID, highTieID, lowTieID)
	}
	if firstPage[0].Author.Handle != "fanviewer" {
		t.Fatalf("ListComments(first page) author handle got %q want fanviewer", firstPage[0].Author.Handle)
	}
	if nextCursor == nil || nextCursor.CommentID != lowTieID || !nextCursor.CreatedAt.Equal(createdAt) {
		t.Fatalf("ListComments(first page) cursor got %#v want low tie cursor", nextCursor)
	}

	secondPage, finalCursor, err := repo.ListComments(ctx, scenario.ShortID, nextCursor, 2)
	if err != nil {
		t.Fatalf("ListComments(second page) error = %v, want nil", err)
	}
	if len(secondPage) != 1 || secondPage[0].ID != olderID {
		t.Fatalf("ListComments(second page) ids got %#v want [%s]", secondPage, olderID)
	}
	if finalCursor != nil {
		t.Fatalf("ListComments(second page) cursor got %#v want nil", finalCursor)
	}
}

func TestRepositoryIntegrationRejectsNonPublicShort(t *testing.T) {
	t.Parallel()

	ctx, pool, cleanup := newShortCommentTestDatabase(t)
	defer cleanup()

	repo := NewRepository(pool)
	scenario := seedShortCommentPublicShort(t, ctx, pool)
	commentID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	insertShortCommentRow(t, ctx, pool, scenario.ShortID, scenario.AuthorUserID, commentID, "hidden", time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC))

	if _, err := pool.Exec(ctx, `
		UPDATE app.shorts
		SET state = 'removed',
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, scenario.ShortID); err != nil {
		t.Fatalf("make short non-public error = %v, want nil", err)
	}

	if _, _, err := repo.ListComments(ctx, scenario.ShortID, nil, 20); !errors.Is(err, ErrShortNotFound) {
		t.Fatalf("ListComments(non-public short) error got %v want ErrShortNotFound", err)
	}
	if _, err := repo.CreateComment(ctx, CreateInput{
		AuthorUserID: scenario.AuthorUserID,
		Body:         "new hidden comment",
		ShortID:      scenario.ShortID,
	}); !errors.Is(err, ErrShortNotFound) {
		t.Fatalf("CreateComment(non-public short) error got %v want ErrShortNotFound", err)
	}
}

func seedShortCommentPublicShort(t *testing.T, ctx context.Context, pool *pgxpool.Pool) shortCommentIntegrationScenario {
	t.Helper()

	now := time.Date(2026, 4, 20, 10, 0, 0, 0, time.UTC)
	scenario := shortCommentIntegrationScenario{
		AuthorUserID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		CreatorID:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		MainAssetID:  uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		MainID:       uuid.MustParse("44444444-4444-4444-4444-444444444444"),
		ShortAssetID: uuid.MustParse("55555555-5555-5555-5555-555555555555"),
		ShortID:      uuid.MustParse("66666666-6666-6666-6666-666666666666"),
	}

	if _, err := pool.Exec(ctx, `
		INSERT INTO app.users (id)
		VALUES ($1), ($2)
	`, scenario.CreatorID, scenario.AuthorUserID); err != nil {
		t.Fatalf("insert users error = %v, want nil", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO app.creator_capabilities (
			user_id,
			state,
			approved_at
		) VALUES (
			$1,
			'approved',
			$2
		)
	`, scenario.CreatorID, now.Add(-3*time.Hour)); err != nil {
		t.Fatalf("insert creator capability error = %v, want nil", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO app.creator_profiles (
			user_id,
			display_name,
			handle,
			avatar_url,
			bio,
			published_at
		) VALUES (
			$1,
			'Creator One',
			'creatorone',
			NULL,
			'',
			$2
		)
	`, scenario.CreatorID, now.Add(-3*time.Hour)); err != nil {
		t.Fatalf("insert creator profile error = %v, want nil", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO app.user_profiles (
			user_id,
			display_name,
			handle,
			avatar_url
		) VALUES
			($1, 'Creator One', 'creatorone', NULL),
			($2, 'Fan Viewer', 'fanviewer', 'https://cdn.example.com/fan-avatar.jpg')
	`, scenario.CreatorID, scenario.AuthorUserID); err != nil {
		t.Fatalf("insert user profiles error = %v, want nil", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO app.media_assets (
			id,
			creator_user_id,
			processing_state,
			storage_provider,
			storage_bucket,
			storage_key,
			playback_url,
			mime_type,
			duration_ms,
			external_upload_ref
		) VALUES (
			$1,
			$2,
			'ready',
			's3',
			'mock-media-bucket',
			'mock/mains/comment-main.mp4',
			'https://cdn.example.com/mains/comment-main.m3u8',
			'video/mp4',
			180000,
			NULL
		), (
			$3,
			$2,
			'ready',
			's3',
			'mock-media-bucket',
			'mock/shorts/comment-short.mp4',
			'https://cdn.example.com/shorts/comment-short.m3u8',
			'video/mp4',
			18000,
			NULL
		)
	`, scenario.MainAssetID, scenario.CreatorID, scenario.ShortAssetID); err != nil {
		t.Fatalf("insert media assets error = %v, want nil", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO app.mains (
			id,
			creator_user_id,
			media_asset_id,
			state,
			review_reason_code,
			post_report_state,
			price_minor,
			currency_code,
			ownership_confirmed,
			consent_confirmed,
			approved_for_unlock_at
		) VALUES (
			$1,
			$2,
			$3,
			'approved_for_unlock',
			NULL,
			NULL,
			1800,
			'JPY',
			TRUE,
			TRUE,
			$4
		)
	`, scenario.MainID, scenario.CreatorID, scenario.MainAssetID, now.Add(-2*time.Hour)); err != nil {
		t.Fatalf("insert main error = %v, want nil", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO app.shorts (
			id,
			creator_user_id,
			canonical_main_id,
			media_asset_id,
			caption,
			state,
			review_reason_code,
			post_report_state,
			approved_for_publish_at,
			published_at
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			'Comment fixture short',
			'approved_for_publish',
			NULL,
			NULL,
			$5,
			$6
		)
	`, scenario.ShortID, scenario.CreatorID, scenario.MainID, scenario.ShortAssetID, now.Add(-90*time.Minute), now.Add(-80*time.Minute)); err != nil {
		t.Fatalf("insert short error = %v, want nil", err)
	}

	return scenario
}

func insertShortCommentRow(t *testing.T, ctx context.Context, pool *pgxpool.Pool, shortID uuid.UUID, authorUserID uuid.UUID, id uuid.UUID, body string, createdAt time.Time) {
	t.Helper()

	if _, err := pool.Exec(ctx, `
		INSERT INTO app.short_comments (
			id,
			short_id,
			author_user_id,
			body,
			created_at,
			updated_at
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$5
		)
	`, id, shortID, authorUserID, body, createdAt); err != nil {
		t.Fatalf("insert short comment %s error = %v, want nil", id, err)
	}
}

func newShortCommentTestDatabase(t *testing.T) (context.Context, *pgxpool.Pool, func()) {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv(shortCommentIntegrationPostgresDSNEnv))
	if dsn == "" {
		t.Skipf("%s is required for postgres integration tests", shortCommentIntegrationPostgresDSNEnv)
	}

	ctx := context.Background()
	baseConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("pgx.ParseConfig() error = %v, want nil", err)
	}

	adminConn, err := connectShortCommentAdminDatabase(ctx, baseConfig)
	if err != nil {
		t.Fatalf("connectShortCommentAdminDatabase() error = %v, want nil", err)
	}

	tempDatabaseName := "shortcomment_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := adminConn.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{tempDatabaseName}.Sanitize()); err != nil {
		closeShortCommentAdminConn(t, ctx, adminConn)
		t.Fatalf("CREATE DATABASE %q error = %v, want nil", tempDatabaseName, err)
	}

	tempConfig := baseConfig.Copy()
	tempConfig.Database = tempDatabaseName
	migrator := newShortCommentTestMigrator(t, tempConfig)
	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		closeShortCommentMigrator(t, migrator)
		dropShortCommentTempDatabase(t, ctx, adminConn, tempDatabaseName)
		closeShortCommentAdminConn(t, ctx, adminConn)
		t.Fatalf("migrator.Up() error = %v, want nil", err)
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		closeShortCommentMigrator(t, migrator)
		dropShortCommentTempDatabase(t, ctx, adminConn, tempDatabaseName)
		closeShortCommentAdminConn(t, ctx, adminConn)
		t.Fatalf("pgxpool.ParseConfig() error = %v, want nil", err)
	}
	poolConfig.ConnConfig.Database = tempDatabaseName

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		closeShortCommentMigrator(t, migrator)
		dropShortCommentTempDatabase(t, ctx, adminConn, tempDatabaseName)
		closeShortCommentAdminConn(t, ctx, adminConn)
		t.Fatalf("pgxpool.NewWithConfig() error = %v, want nil", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		closeShortCommentMigrator(t, migrator)
		dropShortCommentTempDatabase(t, ctx, adminConn, tempDatabaseName)
		closeShortCommentAdminConn(t, ctx, adminConn)
		t.Fatalf("pool.Ping() error = %v, want nil", err)
	}

	cleanup := func() {
		pool.Close()
		closeShortCommentMigrator(t, migrator)
		dropShortCommentTempDatabase(t, ctx, adminConn, tempDatabaseName)
		closeShortCommentAdminConn(t, ctx, adminConn)
	}

	return ctx, pool, cleanup
}

func connectShortCommentAdminDatabase(ctx context.Context, baseConfig *pgx.ConnConfig) (*pgx.Conn, error) {
	var lastErr error
	for _, databaseName := range shortCommentUniqueNonEmptyStrings("postgres", baseConfig.Database, "template1") {
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

func newShortCommentTestMigrator(t *testing.T, config *pgx.ConnConfig) *migrate.Migrate {
	t.Helper()

	db := stdlib.OpenDB(*config)

	driver, err := pgmigrate.WithInstance(db, &pgmigrate.Config{})
	if err != nil {
		closeShortCommentSQLDB(t, db)
		t.Fatalf("postgres.WithInstance() error = %v, want nil", err)
	}

	sourceURL := (&url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(shortCommentMigrationDir(t)),
	}).String()

	migrator, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		closeShortCommentSQLDB(t, db)
		t.Fatalf("migrate.NewWithDatabaseInstance() error = %v, want nil", err)
	}

	return migrator
}

func closeShortCommentMigrator(t *testing.T, migrator *migrate.Migrate) {
	t.Helper()

	if migrator == nil {
		return
	}

	sourceErr, databaseErr := migrator.Close()
	if sourceErr != nil {
		t.Fatalf("migrator.Close() source error = %v, want nil", sourceErr)
	}
	if databaseErr != nil {
		t.Fatalf("migrator.Close() database error = %v, want nil", databaseErr)
	}
}

func closeShortCommentAdminConn(t *testing.T, ctx context.Context, conn *pgx.Conn) {
	t.Helper()

	if conn == nil {
		return
	}

	if err := conn.Close(ctx); err != nil {
		t.Fatalf("adminConn.Close() error = %v, want nil", err)
	}
}

func closeShortCommentSQLDB(t *testing.T, db *sql.DB) {
	t.Helper()

	if db == nil {
		return
	}

	if err := db.Close(); err != nil {
		t.Fatalf("db.Close() error = %v, want nil", err)
	}
}

func dropShortCommentTempDatabase(t *testing.T, ctx context.Context, adminConn *pgx.Conn, databaseName string) {
	t.Helper()

	if _, err := adminConn.Exec(
		ctx,
		`SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1
			AND pid <> pg_backend_pid()`,
		databaseName,
	); err != nil {
		t.Fatalf("pg_terminate_backend(%q) error = %v, want nil", databaseName, err)
	}

	if _, err := adminConn.Exec(ctx, "DROP DATABASE IF EXISTS "+pgx.Identifier{databaseName}.Sanitize()); err != nil {
		t.Fatalf("DROP DATABASE %q error = %v, want nil", databaseName, err)
	}
}

func shortCommentMigrationDir(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() ok = false, want true")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "db", "migrations"))
}

func shortCommentUniqueNonEmptyStrings(values ...string) []string {
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
