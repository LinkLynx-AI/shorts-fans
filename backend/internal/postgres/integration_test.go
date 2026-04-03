package postgres

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

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/golang-migrate/migrate/v4"
	pgmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const integrationPostgresDSNEnv = "POSTGRES_DSN"

func TestMigrationsRoundTripLatestRevision(t *testing.T) {
	ctx, conn, migrator, cleanup := newIntegrationEnvironment(t)
	defer cleanup()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Up() error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, 3)

	queries := sqlc.New(conn)
	user, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser() error = %v, want nil", err)
	}

	now := time.Unix(1710000000, 0).UTC()
	_, err = queries.CreateCreatorCapability(ctx, sqlc.CreateCreatorCapabilityParams{
		UserID:                  user.ID,
		State:                   "approved",
		IsResubmitEligible:      false,
		IsSupportReviewRequired: false,
		SelfServeResubmitCount:  0,
		ApprovedAt:              pgTime(now),
	})
	if err != nil {
		t.Fatalf("CreateCreatorCapability() error = %v, want nil", err)
	}

	_, err = queries.CreateCreatorProfile(ctx, sqlc.CreateCreatorProfileParams{
		UserID:      user.ID,
		DisplayName: pgText("draft-profile"),
		Bio:         "draft bio",
	})
	if err != nil {
		t.Fatalf("CreateCreatorProfile() error = %v, want nil", err)
	}

	assertRelationExists(t, ctx, conn, "app.creator_profile_drafts", false)

	if err := migrator.Steps(-1); err != nil {
		t.Fatalf("migrator.Steps(-1) error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, 2)
	assertRelationExists(t, ctx, conn, "app.creator_profile_drafts", true)

	var draftBio string
	if err := conn.QueryRow(
		ctx,
		"SELECT bio FROM app.creator_profile_drafts WHERE user_id = $1 LIMIT 1",
		user.ID,
	).Scan(&draftBio); err != nil {
		t.Fatalf("QueryRow(creator_profile_drafts) error = %v, want nil", err)
	}
	if draftBio != "draft bio" {
		t.Fatalf("creator_profile_drafts bio got %q want %q", draftBio, "draft bio")
	}

	if _, err := queries.GetCreatorProfileByUserID(ctx, user.ID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("GetCreatorProfileByUserID() after down error got %v want %v", err, pgx.ErrNoRows)
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Up() second run error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, 3)
	assertRelationExists(t, ctx, conn, "app.creator_profile_drafts", false)

	profile, err := queries.GetCreatorProfileByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetCreatorProfileByUserID() after re-up error = %v, want nil", err)
	}
	if got := textFromPG(profile.DisplayName); got != "draft-profile" {
		t.Fatalf("GetCreatorProfileByUserID() display_name got %q want %q", got, "draft-profile")
	}
	if profile.Bio != "draft bio" {
		t.Fatalf("GetCreatorProfileByUserID() bio got %q want %q", profile.Bio, "draft bio")
	}
	if profile.PublishedAt.Valid {
		t.Fatalf("GetCreatorProfileByUserID() published_at valid got %t want false", profile.PublishedAt.Valid)
	}
}

func TestCoreQueriesReflectAccessBoundaries(t *testing.T) {
	ctx, conn, migrator, cleanup := newIntegrationEnvironment(t)
	defer cleanup()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Up() error = %v, want nil", err)
	}

	queries := sqlc.New(conn)
	now := time.Unix(1710000000, 0).UTC()

	creator, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser(creator) error = %v, want nil", err)
	}
	buyer, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser(buyer) error = %v, want nil", err)
	}

	_, err = queries.CreateCreatorCapability(ctx, sqlc.CreateCreatorCapabilityParams{
		UserID:                  creator.ID,
		State:                   "approved",
		IsResubmitEligible:      false,
		IsSupportReviewRequired: false,
		SelfServeResubmitCount:  0,
		ApprovedAt:              pgTime(now),
	})
	if err != nil {
		t.Fatalf("CreateCreatorCapability() error = %v, want nil", err)
	}

	unlockableMainAsset, err := createReadyMediaAsset(ctx, queries, creator.ID, "main-unlockable")
	if err != nil {
		t.Fatalf("createReadyMediaAsset(unlockable main) error = %v, want nil", err)
	}
	lockedMainAsset, err := createReadyMediaAsset(ctx, queries, creator.ID, "main-locked")
	if err != nil {
		t.Fatalf("createReadyMediaAsset(locked main) error = %v, want nil", err)
	}
	shortAssetA, err := createReadyMediaAsset(ctx, queries, creator.ID, "short-a")
	if err != nil {
		t.Fatalf("createReadyMediaAsset(short-a) error = %v, want nil", err)
	}
	shortAssetB, err := createReadyMediaAsset(ctx, queries, creator.ID, "short-b")
	if err != nil {
		t.Fatalf("createReadyMediaAsset(short-b) error = %v, want nil", err)
	}
	lockedShortAsset, err := createReadyMediaAsset(ctx, queries, creator.ID, "short-locked")
	if err != nil {
		t.Fatalf("createReadyMediaAsset(short-locked) error = %v, want nil", err)
	}

	unlockableMain, err := queries.CreateMain(ctx, sqlc.CreateMainParams{
		CreatorUserID:       creator.ID,
		MediaAssetID:        unlockableMainAsset.ID,
		State:               "approved_for_unlock",
		PriceMinor:          pgInt64(1200),
		CurrencyCode:        pgText("JPY"),
		OwnershipConfirmed:  true,
		ConsentConfirmed:    true,
		ApprovedForUnlockAt: pgTime(now.Add(time.Hour)),
	})
	if err != nil {
		t.Fatalf("CreateMain(unlockable) error = %v, want nil", err)
	}
	lockedMain, err := queries.CreateMain(ctx, sqlc.CreateMainParams{
		CreatorUserID:      creator.ID,
		MediaAssetID:       lockedMainAsset.ID,
		State:              "draft",
		OwnershipConfirmed: true,
		ConsentConfirmed:   true,
	})
	if err != nil {
		t.Fatalf("CreateMain(locked) error = %v, want nil", err)
	}

	shortA, err := queries.CreateShort(ctx, sqlc.CreateShortParams{
		CreatorUserID:        creator.ID,
		CanonicalMainID:      unlockableMain.ID,
		MediaAssetID:         shortAssetA.ID,
		State:                "approved_for_publish",
		ApprovedForPublishAt: pgTime(now.Add(2 * time.Hour)),
		PublishedAt:          pgTime(now.Add(3 * time.Hour)),
	})
	if err != nil {
		t.Fatalf("CreateShort(shortA) error = %v, want nil", err)
	}
	shortB, err := queries.CreateShort(ctx, sqlc.CreateShortParams{
		CreatorUserID:        creator.ID,
		CanonicalMainID:      unlockableMain.ID,
		MediaAssetID:         shortAssetB.ID,
		State:                "approved_for_publish",
		ApprovedForPublishAt: pgTime(now.Add(4 * time.Hour)),
		PublishedAt:          pgTime(now.Add(5 * time.Hour)),
	})
	if err != nil {
		t.Fatalf("CreateShort(shortB) error = %v, want nil", err)
	}
	lockedShort, err := queries.CreateShort(ctx, sqlc.CreateShortParams{
		CreatorUserID:        creator.ID,
		CanonicalMainID:      lockedMain.ID,
		MediaAssetID:         lockedShortAsset.ID,
		State:                "approved_for_publish",
		ApprovedForPublishAt: pgTime(now.Add(6 * time.Hour)),
		PublishedAt:          pgTime(now.Add(7 * time.Hour)),
	})
	if err != nil {
		t.Fatalf("CreateShort(lockedShort) error = %v, want nil", err)
	}

	gotUnlockableMain, err := queries.GetUnlockableMainByID(ctx, unlockableMain.ID)
	if err != nil {
		t.Fatalf("GetUnlockableMainByID(unlockable) error = %v, want nil", err)
	}
	if gotUnlockableMain.ID != unlockableMain.ID {
		t.Fatalf("GetUnlockableMainByID(unlockable) id got %v want %v", gotUnlockableMain.ID, unlockableMain.ID)
	}
	if _, err := queries.GetUnlockableMainByID(ctx, lockedMain.ID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("GetUnlockableMainByID(locked) error got %v want %v", err, pgx.ErrNoRows)
	}

	gotCanonicalMainID, err := queries.GetCanonicalMainIDByShortID(ctx, shortA.ID)
	if err != nil {
		t.Fatalf("GetCanonicalMainIDByShortID() error = %v, want nil", err)
	}
	if gotCanonicalMainID != unlockableMain.ID {
		t.Fatalf("GetCanonicalMainIDByShortID() got %v want %v", gotCanonicalMainID, unlockableMain.ID)
	}

	canonicalShorts, err := queries.ListShortsByCanonicalMainID(ctx, unlockableMain.ID)
	if err != nil {
		t.Fatalf("ListShortsByCanonicalMainID() error = %v, want nil", err)
	}
	if len(canonicalShorts) != 2 {
		t.Fatalf("ListShortsByCanonicalMainID() len got %d want 2", len(canonicalShorts))
	}
	assertUUIDSet(t, "ListShortsByCanonicalMainID()", []pgtype.UUID{canonicalShorts[0].ID, canonicalShorts[1].ID}, []pgtype.UUID{shortA.ID, shortB.ID})

	publicShorts, err := queries.ListPublicShortsByCreatorUserID(ctx, creator.ID)
	if err != nil {
		t.Fatalf("ListPublicShortsByCreatorUserID() error = %v, want nil", err)
	}
	if len(publicShorts) != 2 {
		t.Fatalf("ListPublicShortsByCreatorUserID() len got %d want 2", len(publicShorts))
	}
	if publicShorts[0].ID != shortB.ID || publicShorts[1].ID != shortA.ID {
		t.Fatalf("ListPublicShortsByCreatorUserID() order got [%v %v] want [%v %v]", publicShorts[0].ID, publicShorts[1].ID, shortB.ID, shortA.ID)
	}
	gotPublicShort, err := queries.GetPublicShortByID(ctx, shortA.ID)
	if err != nil {
		t.Fatalf("GetPublicShortByID(shortA) error = %v, want nil", err)
	}
	if gotPublicShort.ID != shortA.ID {
		t.Fatalf("GetPublicShortByID(shortA) id got %v want %v", gotPublicShort.ID, shortA.ID)
	}
	if _, err := queries.GetPublicShortByID(ctx, lockedShort.ID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("GetPublicShortByID(lockedShort) error got %v want %v", err, pgx.ErrNoRows)
	}

	unlockedMainIDs, err := queries.ListUnlockedMainIDsByUserID(ctx, buyer.ID)
	if err != nil {
		t.Fatalf("ListUnlockedMainIDsByUserID() before purchase error = %v, want nil", err)
	}
	if len(unlockedMainIDs) != 0 {
		t.Fatalf("ListUnlockedMainIDsByUserID() before purchase len got %d want 0", len(unlockedMainIDs))
	}
	if _, err := queries.GetMainUnlockByUserIDAndMainID(ctx, sqlc.GetMainUnlockByUserIDAndMainIDParams{
		UserID: buyer.ID,
		MainID: unlockableMain.ID,
	}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("GetMainUnlockByUserIDAndMainID() before purchase error got %v want %v", err, pgx.ErrNoRows)
	}

	if _, err := queries.CreateMainUnlock(ctx, sqlc.CreateMainUnlockParams{
		UserID: buyer.ID,
		MainID: lockedMain.ID,
	}); err == nil {
		t.Fatal("CreateMainUnlock(locked main) error = nil, want trigger error")
	}

	recordedUnlock, err := queries.CreateMainUnlock(ctx, sqlc.CreateMainUnlockParams{
		UserID:                     buyer.ID,
		MainID:                     unlockableMain.ID,
		PaymentProviderPurchaseRef: pgText("purchase-1"),
		PurchasedAt:                pgTime(now.Add(8 * time.Hour)),
	})
	if err != nil {
		t.Fatalf("CreateMainUnlock(unlockable) error = %v, want nil", err)
	}

	gotUnlock, err := queries.GetMainUnlockByUserIDAndMainID(ctx, sqlc.GetMainUnlockByUserIDAndMainIDParams{
		UserID: buyer.ID,
		MainID: unlockableMain.ID,
	})
	if err != nil {
		t.Fatalf("GetMainUnlockByUserIDAndMainID() after purchase error = %v, want nil", err)
	}
	if gotUnlock.MainID != recordedUnlock.MainID || gotUnlock.UserID != recordedUnlock.UserID {
		t.Fatalf("GetMainUnlockByUserIDAndMainID() got %#v want %#v", gotUnlock, recordedUnlock)
	}

	unlockedMainIDs, err = queries.ListUnlockedMainIDsByUserID(ctx, buyer.ID)
	if err != nil {
		t.Fatalf("ListUnlockedMainIDsByUserID() after purchase error = %v, want nil", err)
	}
	if len(unlockedMainIDs) != 1 || unlockedMainIDs[0] != unlockableMain.ID {
		t.Fatalf("ListUnlockedMainIDsByUserID() after purchase got %#v want [%v]", unlockedMainIDs, unlockableMain.ID)
	}
}

func newIntegrationEnvironment(t *testing.T) (context.Context, *pgx.Conn, *migrate.Migrate, func()) {
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

	tempDatabaseName := "sho14_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := adminConn.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{tempDatabaseName}.Sanitize()); err != nil {
		adminConn.Close(ctx)
		t.Fatalf("CREATE DATABASE %q error = %v, want nil", tempDatabaseName, err)
	}

	tempConfig := baseConfig.Copy()
	tempConfig.Database = tempDatabaseName
	tempDSN := tempConfig.ConnString()

	migrator := newTestMigrator(t, tempDSN)

	conn, err := pgx.Connect(ctx, tempDSN)
	if err != nil {
		closeMigrator(t, migrator)
		dropTempDatabase(t, ctx, adminConn, tempDatabaseName)
		adminConn.Close(ctx)
		t.Fatalf("pgx.Connect() error = %v, want nil", err)
	}

	cleanup := func() {
		conn.Close(ctx)
		closeMigrator(t, migrator)
		dropTempDatabase(t, ctx, adminConn, tempDatabaseName)
		adminConn.Close(ctx)
	}

	return ctx, conn, migrator, cleanup
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

func newTestMigrator(t *testing.T, dsn string) *migrate.Migrate {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open() error = %v, want nil", err)
	}

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

func assertMigrationVersion(t *testing.T, migrator *migrate.Migrate, want uint) {
	t.Helper()

	got, dirty, err := migrator.Version()
	if err != nil {
		t.Fatalf("migrator.Version() error = %v, want nil", err)
	}
	if dirty {
		t.Fatal("migrator.Version() dirty = true, want false")
	}
	if got != want {
		t.Fatalf("migrator.Version() got %d want %d", got, want)
	}
}

func assertRelationExists(t *testing.T, ctx context.Context, conn *pgx.Conn, relationName string, want bool) {
	t.Helper()

	var regclass pgtype.Text
	if err := conn.QueryRow(ctx, "SELECT to_regclass($1)::text", relationName).Scan(&regclass); err != nil {
		t.Fatalf("to_regclass(%q) error = %v, want nil", relationName, err)
	}
	if regclass.Valid != want {
		t.Fatalf("to_regclass(%q) valid got %t want %t", relationName, regclass.Valid, want)
	}
}

func createReadyMediaAsset(ctx context.Context, queries *sqlc.Queries, creatorUserID pgtype.UUID, key string) (sqlc.AppMediaAsset, error) {
	return queries.CreateMediaAsset(ctx, sqlc.CreateMediaAssetParams{
		CreatorUserID:   creatorUserID,
		ProcessingState: "ready",
		StorageProvider: "s3",
		StorageBucket:   "test-bucket",
		StorageKey:      key,
		PlaybackUrl:     pgText("https://cdn.example.com/" + key + ".m3u8"),
		MimeType:        "video/mp4",
		DurationMs:      pgInt64(1000),
	})
}

func assertUUIDSet(t *testing.T, label string, got []pgtype.UUID, want []pgtype.UUID) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("%s len got %d want %d", label, len(got), len(want))
	}

	gotSet := make(map[[16]byte]struct{}, len(got))
	for _, id := range got {
		gotSet[id.Bytes] = struct{}{}
	}

	for _, id := range want {
		if _, ok := gotSet[id.Bytes]; !ok {
			t.Fatalf("%s missing id %v in %#v", label, id, got)
		}
	}
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

func pgText(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: true}
}

func textFromPG(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func pgTime(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func pgInt64(value int64) pgtype.Int8 {
	return pgtype.Int8{Int64: value, Valid: true}
}
