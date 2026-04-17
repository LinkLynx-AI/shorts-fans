package devseed

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/auth"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/golang-migrate/migrate/v4"
	pgmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

const integrationPostgresDSNEnv = "POSTGRES_DSN"

func TestRunSeedsBaselineDataIdempotently(t *testing.T) {
	t.Parallel()

	ctx, pool, cleanup := newSeedTestDatabase(t)
	defer cleanup()

	summary, err := Run(ctx, pool)
	if err != nil {
		t.Fatalf("Run() initial error = %v, want nil", err)
	}
	if summary.CreatorUserID != creatorUserID {
		t.Fatalf("Run() creator user id got %s want %s", summary.CreatorUserID, creatorUserID)
	}
	if summary.FanUserID != fanUserID {
		t.Fatalf("Run() fan user id got %s want %s", summary.FanUserID, fanUserID)
	}
	if summary.MainID != mainID {
		t.Fatalf("Run() main id got %s want %s", summary.MainID, mainID)
	}
	if summary.SubmittedReviewUserID != reviewUserID {
		t.Fatalf("Run() submitted review user id got %s want %s", summary.SubmittedReviewUserID, reviewUserID)
	}
	if len(summary.ShortIDs) != len(publicShorts) {
		t.Fatalf("Run() short id count got %d want %d", len(summary.ShortIDs), len(publicShorts))
	}
	if summary.FanSessionToken != fanSessionToken {
		t.Fatalf("Run() fan session token got %q want %q", summary.FanSessionToken, fanSessionToken)
	}
	if summary.CreatorSessionToken != creatorSessionToken {
		t.Fatalf("Run() creator session token got %q want %q", summary.CreatorSessionToken, creatorSessionToken)
	}

	queries := sqlc.New(pool)

	capability, err := queries.GetCreatorCapabilityByUserID(ctx, uuidToPG(creatorUserID))
	if err != nil {
		t.Fatalf("GetCreatorCapabilityByUserID() error = %v, want nil", err)
	}
	if capability.State != "approved" {
		t.Fatalf("GetCreatorCapabilityByUserID() state got %q want %q", capability.State, "approved")
	}

	reviewCapability, err := queries.GetCreatorCapabilityByUserID(ctx, uuidToPG(reviewUserID))
	if err != nil {
		t.Fatalf("GetCreatorCapabilityByUserID() review error = %v, want nil", err)
	}
	if reviewCapability.State != "submitted" {
		t.Fatalf("GetCreatorCapabilityByUserID() review state got %q want %q", reviewCapability.State, "submitted")
	}

	profile, err := queries.GetCreatorProfileByUserID(ctx, uuidToPG(creatorUserID))
	if err != nil {
		t.Fatalf("GetCreatorProfileByUserID() error = %v, want nil", err)
	}
	if !profile.PublishedAt.Valid {
		t.Fatal("GetCreatorProfileByUserID() published_at valid = false, want true")
	}
	if got := textFromPG(profile.DisplayName); got != creatorDisplayName {
		t.Fatalf("GetCreatorProfileByUserID() display_name got %q want %q", got, creatorDisplayName)
	}
	if got := profile.Handle; got != creatorHandle {
		t.Fatalf("GetCreatorProfileByUserID() handle got %q want %q", got, creatorHandle)
	}
	if profile.Bio != creatorBio {
		t.Fatalf("GetCreatorProfileByUserID() bio got %q want %q", profile.Bio, creatorBio)
	}

	publicProfile, err := queries.GetPublicCreatorProfileByHandle(ctx, creatorHandle)
	if err != nil {
		t.Fatalf("GetPublicCreatorProfileByHandle() error = %v, want nil", err)
	}
	if publicProfile.UserID != uuidToPG(creatorUserID) {
		t.Fatalf("GetPublicCreatorProfileByHandle() user_id got %s want %s", publicProfile.UserID, uuidToPG(creatorUserID))
	}

	submittedQueue, err := queries.ListCreatorRegistrationReviewCasesByState(ctx, "submitted")
	if err != nil {
		t.Fatalf("ListCreatorRegistrationReviewCasesByState() error = %v, want nil", err)
	}
	if len(submittedQueue) != 1 {
		t.Fatalf("ListCreatorRegistrationReviewCasesByState() len got %d want 1", len(submittedQueue))
	}
	if submittedQueue[0].UserID != uuidToPG(reviewUserID) {
		t.Fatalf("ListCreatorRegistrationReviewCasesByState() user id got %s want %s", submittedQueue[0].UserID, uuidToPG(reviewUserID))
	}
	if submittedQueue[0].State != "submitted" {
		t.Fatalf("ListCreatorRegistrationReviewCasesByState() state got %q want %q", submittedQueue[0].State, "submitted")
	}

	main, err := queries.GetUnlockableMainByID(ctx, uuidToPG(mainID))
	if err != nil {
		t.Fatalf("GetUnlockableMainByID() error = %v, want nil", err)
	}
	if got := int64FromPG(main.PriceMinor); got != mainPriceMinor {
		t.Fatalf("GetUnlockableMainByID() price_minor got %d want %d", got, mainPriceMinor)
	}

	publicShortRows, err := queries.ListPublicShortsByCreatorUserID(ctx, uuidToPG(creatorUserID))
	if err != nil {
		t.Fatalf("ListPublicShortsByCreatorUserID() error = %v, want nil", err)
	}
	if len(publicShortRows) != len(publicShorts) {
		t.Fatalf("ListPublicShortsByCreatorUserID() len got %d want %d", len(publicShortRows), len(publicShorts))
	}
	if publicShortRows[0].ID != uuidToPG(shortBID) || publicShortRows[1].ID != uuidToPG(shortAID) {
		t.Fatalf("ListPublicShortsByCreatorUserID() ids got [%s %s] want [%s %s]", publicShortRows[0].ID, publicShortRows[1].ID, uuidToPG(shortBID), uuidToPG(shortAID))
	}

	unlock, err := queries.GetMainUnlockByUserIDAndMainID(ctx, sqlc.GetMainUnlockByUserIDAndMainIDParams{
		UserID: uuidToPG(fanUserID),
		MainID: uuidToPG(mainID),
	})
	if err != nil {
		t.Fatalf("GetMainUnlockByUserIDAndMainID() error = %v, want nil", err)
	}
	if got := textFromPG(unlock.PaymentProviderPurchaseRef); got != mainPurchaseRef {
		t.Fatalf("GetMainUnlockByUserIDAndMainID() payment_provider_purchase_ref got %q want %q", got, mainPurchaseRef)
	}

	fanViewer, err := queries.GetCurrentViewerBySessionTokenHash(ctx, auth.HashSessionToken(summary.FanSessionToken))
	if err != nil {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() fan error = %v, want nil", err)
	}
	if fanViewer.UserID != uuidToPG(fanUserID) {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() fan user got %v want %v", fanViewer.UserID, uuidToPG(fanUserID))
	}
	if fanViewer.ActiveMode != "fan" {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() fan active mode got %q want %q", fanViewer.ActiveMode, "fan")
	}
	if fanViewer.CanAccessCreatorMode {
		t.Fatal("GetCurrentViewerBySessionTokenHash() fan can_access_creator_mode = true, want false")
	}

	creatorViewer, err := queries.GetCurrentViewerBySessionTokenHash(ctx, auth.HashSessionToken(summary.CreatorSessionToken))
	if err != nil {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() creator error = %v, want nil", err)
	}
	if creatorViewer.UserID != uuidToPG(creatorUserID) {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() creator user got %v want %v", creatorViewer.UserID, uuidToPG(creatorUserID))
	}
	if creatorViewer.ActiveMode != "creator" {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() creator active mode got %q want %q", creatorViewer.ActiveMode, "creator")
	}
	if !creatorViewer.CanAccessCreatorMode {
		t.Fatal("GetCurrentViewerBySessionTokenHash() creator can_access_creator_mode = false, want true")
	}

	assertCount(t, ctx, pool, "SELECT count(*) FROM app.users", 3)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.user_profiles", 3)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.creator_capabilities", 2)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.creator_profiles", 2)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.creator_registration_intakes", 1)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.creator_registration_evidences", 2)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.media_assets", 3)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.mains", 1)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.shorts", 2)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.main_unlocks", 1)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.creator_follows", 1)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.pinned_shorts", 1)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.auth_sessions", 2)

	if _, err := pool.Exec(ctx, "UPDATE app.creator_profiles SET bio = $1 WHERE user_id = $2", "stale bio", creatorUserID); err != nil {
		t.Fatalf("UPDATE app.creator_profiles stale bio error = %v, want nil", err)
	}

	if _, err := Run(ctx, pool); err != nil {
		t.Fatalf("Run() rerun error = %v, want nil", err)
	}

	profile, err = queries.GetCreatorProfileByUserID(ctx, uuidToPG(creatorUserID))
	if err != nil {
		t.Fatalf("GetCreatorProfileByUserID() after rerun error = %v, want nil", err)
	}
	if profile.Bio != creatorBio {
		t.Fatalf("GetCreatorProfileByUserID() after rerun bio got %q want %q", profile.Bio, creatorBio)
	}

	assertCount(t, ctx, pool, "SELECT count(*) FROM app.users", 3)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.user_profiles", 3)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.media_assets", 3)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.shorts", 2)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.main_unlocks", 1)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.creator_follows", 1)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.pinned_shorts", 1)
	assertCount(t, ctx, pool, "SELECT count(*) FROM app.auth_sessions", 2)
}

func newSeedTestDatabase(t *testing.T) (context.Context, *pgxpool.Pool, func()) {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv(integrationPostgresDSNEnv))
	if dsn == "" {
		t.Skipf("%s is required for postgres integration tests", integrationPostgresDSNEnv)
	}

	ctx := context.Background()
	baseConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("pgx.ParseConfig() error = %v, want nil", err)
	}

	adminConn, err := connectAdminDatabase(ctx, baseConfig)
	if err != nil {
		t.Fatalf("connectAdminDatabase() error = %v, want nil", err)
	}

	tempDatabaseName := "devseed_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := adminConn.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{tempDatabaseName}.Sanitize()); err != nil {
		adminConn.Close(ctx)
		t.Fatalf("CREATE DATABASE %q error = %v, want nil", tempDatabaseName, err)
	}

	tempConfig := baseConfig.Copy()
	tempConfig.Database = tempDatabaseName
	migrator := newTestMigrator(t, tempConfig)
	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		closeMigrator(t, migrator)
		dropTempDatabase(t, ctx, adminConn, tempDatabaseName)
		adminConn.Close(ctx)
		t.Fatalf("migrator.Up() error = %v, want nil", err)
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		closeMigrator(t, migrator)
		dropTempDatabase(t, ctx, adminConn, tempDatabaseName)
		adminConn.Close(ctx)
		t.Fatalf("pgxpool.ParseConfig() error = %v, want nil", err)
	}
	poolConfig.ConnConfig.Database = tempDatabaseName

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		closeMigrator(t, migrator)
		dropTempDatabase(t, ctx, adminConn, tempDatabaseName)
		adminConn.Close(ctx)
		t.Fatalf("pgxpool.NewWithConfig() error = %v, want nil", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		closeMigrator(t, migrator)
		dropTempDatabase(t, ctx, adminConn, tempDatabaseName)
		adminConn.Close(ctx)
		t.Fatalf("pool.Ping() error = %v, want nil", err)
	}

	cleanup := func() {
		pool.Close()
		closeMigrator(t, migrator)
		dropTempDatabase(t, ctx, adminConn, tempDatabaseName)
		adminConn.Close(ctx)
	}

	return ctx, pool, cleanup
}

func connectAdminDatabase(ctx context.Context, baseConfig *pgx.ConnConfig) (*pgx.Conn, error) {
	var lastErr error
	for _, databaseName := range uniqueNonEmptyStrings("postgres", baseConfig.Database, "template1") {
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

func newTestMigrator(t *testing.T, config *pgx.ConnConfig) *migrate.Migrate {
	t.Helper()

	db := stdlib.OpenDB(*config)

	driver, err := pgmigrate.WithInstance(db, &pgmigrate.Config{})
	if err != nil {
		db.Close()
		t.Fatalf("postgres.WithInstance() error = %v, want nil", err)
	}

	sourceURL := (&url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(migrationDir(t)),
	}).String()

	migrator, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		db.Close()
		t.Fatalf("migrate.NewWithDatabaseInstance() error = %v, want nil", err)
	}

	return migrator
}

func closeMigrator(t *testing.T, migrator *migrate.Migrate) {
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

func dropTempDatabase(t *testing.T, ctx context.Context, adminConn *pgx.Conn, databaseName string) {
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

func migrationDir(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() ok = false, want true")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "db", "migrations"))
}

func uniqueNonEmptyStrings(values ...string) []string {
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

func assertCount(t *testing.T, ctx context.Context, pool *pgxpool.Pool, query string, want int) {
	t.Helper()

	var got int
	if err := pool.QueryRow(ctx, query).Scan(&got); err != nil {
		t.Fatalf("QueryRow(%q) error = %v, want nil", query, err)
	}
	if got != want {
		t.Fatalf("QueryRow(%q) count got %d want %d", query, got, want)
	}
}

func uuidToPG(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

func textFromPG(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}

	return value.String
}

func int64FromPG(value pgtype.Int8) int64 {
	if !value.Valid {
		return 0
	}

	return value.Int64
}
