package postgres

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
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/golang-migrate/migrate/v4"
	pgmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/stdlib"
)

const (
	integrationPostgresDSNEnv = "POSTGRES_DSN"
	latestMigrationVersion    = 16
)

func TestCreatorProfileMigrationsRoundTrip(t *testing.T) {
	ctx, conn, migrator, cleanup := newIntegrationEnvironment(t)
	defer cleanup()

	if err := migrator.Migrate(3); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Migrate(3) error = %v, want nil", err)
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

	if _, err := conn.Exec(
		ctx,
		`INSERT INTO app.creator_profiles (
			user_id,
			display_name,
			avatar_url,
			bio,
			published_at
		) VALUES ($1, $2, $3, $4, $5)`,
		user.ID,
		"draft-profile",
		nil,
		"draft bio",
		nil,
	); err != nil {
		t.Fatalf("Exec(insert creator_profiles at migration 3) error = %v, want nil", err)
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

	var profileCount int
	if err := conn.QueryRow(ctx, "SELECT count(*) FROM app.creator_profiles WHERE user_id = $1", user.ID).Scan(&profileCount); err != nil {
		t.Fatalf("QueryRow(creator_profiles count after down) error = %v, want nil", err)
	}
	if profileCount != 0 {
		t.Fatalf("creator_profiles count after down got %d want %d", profileCount, 0)
	}

	if err := migrator.Migrate(3); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Migrate(3) second run error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, 3)
	assertRelationExists(t, ctx, conn, "app.creator_profile_drafts", false)

	var (
		displayName pgtype.Text
		bio         string
		publishedAt pgtype.Timestamptz
	)
	if err := conn.QueryRow(
		ctx,
		`SELECT display_name, bio, published_at
		FROM app.creator_profiles
		WHERE user_id = $1`,
		user.ID,
	).Scan(&displayName, &bio, &publishedAt); err != nil {
		t.Fatalf("QueryRow(creator_profiles after re-up) error = %v, want nil", err)
	}
	if got := textFromPG(displayName); got != "draft-profile" {
		t.Fatalf("creator_profiles display_name got %q want %q", got, "draft-profile")
	}
	if bio != "draft bio" {
		t.Fatalf("creator_profiles bio got %q want %q", bio, "draft bio")
	}
	if publishedAt.Valid {
		t.Fatalf("creator_profiles published_at valid got %t want false", publishedAt.Valid)
	}
}

func TestAuthTablesMigrationLatestRevision(t *testing.T) {
	ctx, conn, migrator, cleanup := newIntegrationEnvironment(t)
	defer cleanup()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Up() error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, latestMigrationVersion)

	queries := sqlc.New(conn)
	now := time.Now().UTC()

	user, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser() error = %v, want nil", err)
	}

	identity, err := queries.CreateAuthIdentity(ctx, sqlc.CreateAuthIdentityParams{
		UserID:              user.ID,
		Provider:            "email",
		ProviderSubject:     "fan@example.com",
		EmailNormalized:     pgText("fan@example.com"),
		VerifiedAt:          pgTime(now),
		LastAuthenticatedAt: pgTime(now),
	})
	if err != nil {
		t.Fatalf("CreateAuthIdentity() error = %v, want nil", err)
	}

	secondUser, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser() second user error = %v, want nil", err)
	}

	_, err = queries.CreateAuthIdentity(ctx, sqlc.CreateAuthIdentityParams{
		UserID:              secondUser.ID,
		Provider:            "email",
		ProviderSubject:     "duplicate@example.com",
		EmailNormalized:     pgText("fan@example.com"),
		VerifiedAt:          pgTime(now),
		LastAuthenticatedAt: pgTime(now),
	})
	assertPgConstraintError(t, err, "23505", "idx_auth_identities_email_normalized")

	_, err = queries.CreateAuthIdentity(ctx, sqlc.CreateAuthIdentityParams{
		UserID:              secondUser.ID,
		Provider:            "google",
		ProviderSubject:     "google-subject",
		EmailNormalized:     pgText("fan@example.com"),
		VerifiedAt:          pgTime(now),
		LastAuthenticatedAt: pgTime(now),
	})
	if err != nil {
		t.Fatalf("CreateAuthIdentity() non-email duplicate error = %v, want nil", err)
	}

	_, err = queries.CreateAuthIdentity(ctx, sqlc.CreateAuthIdentityParams{
		UserID:              user.ID,
		Provider:            "cognito",
		ProviderSubject:     "cognito-subject",
		EmailNormalized:     pgText("fan@example.com"),
		VerifiedAt:          pgTime(now),
		LastAuthenticatedAt: pgTime(now),
	})
	if err != nil {
		t.Fatalf("CreateAuthIdentity() cognito bridge error = %v, want nil", err)
	}

	_, err = queries.CreateAuthIdentity(ctx, sqlc.CreateAuthIdentityParams{
		UserID:              secondUser.ID,
		Provider:            "cognito",
		ProviderSubject:     "second-cognito-subject",
		EmailNormalized:     pgText("fan@example.com"),
		VerifiedAt:          pgTime(now),
		LastAuthenticatedAt: pgTime(now),
	})
	assertPgConstraintError(t, err, "23505", "idx_auth_identities_cognito_email_normalized")

	_, err = queries.CreateAuthIdentity(ctx, sqlc.CreateAuthIdentityParams{
		UserID:              secondUser.ID,
		Provider:            "email",
		ProviderSubject:     "missing-email@example.com",
		VerifiedAt:          pgTime(now),
		LastAuthenticatedAt: pgTime(now),
	})
	assertPgConstraintError(t, err, "23514", "auth_identities_email_provider_requires_email_check")

	gotIdentity, err := queries.GetAuthIdentityByProviderAndSubject(ctx, sqlc.GetAuthIdentityByProviderAndSubjectParams{
		Provider:        "email",
		ProviderSubject: "fan@example.com",
	})
	if err != nil {
		t.Fatalf("GetAuthIdentityByProviderAndSubject() error = %v, want nil", err)
	}
	if gotIdentity.ID != identity.ID {
		t.Fatalf("GetAuthIdentityByProviderAndSubject() id got %v want %v", gotIdentity.ID, identity.ID)
	}

	updatedIdentity, err := queries.RecordAuthIdentityAuthentication(ctx, sqlc.RecordAuthIdentityAuthenticationParams{
		ID:                  identity.ID,
		EmailNormalized:     pgText("fan@example.com"),
		VerifiedAt:          pgTime(now),
		LastAuthenticatedAt: pgTime(now.Add(time.Minute)),
	})
	if err != nil {
		t.Fatalf("RecordAuthIdentityAuthentication() error = %v, want nil", err)
	}
	if !updatedIdentity.LastAuthenticatedAt.Time.Equal(now.Add(time.Minute)) {
		t.Fatalf("RecordAuthIdentityAuthentication() last_authenticated_at got %s want %s", updatedIdentity.LastAuthenticatedAt.Time, now.Add(time.Minute))
	}

	challenge, err := queries.CreateAuthLoginChallenge(ctx, sqlc.CreateAuthLoginChallengeParams{
		Provider:           "email",
		ProviderSubject:    "fan@example.com",
		EmailNormalized:    pgText("fan@example.com"),
		ChallengeTokenHash: "challenge-token-hash",
		Purpose:            "login",
		ExpiresAt:          pgTime(now.Add(10 * time.Minute)),
		AttemptCount:       0,
	})
	if err != nil {
		t.Fatalf("CreateAuthLoginChallenge() error = %v, want nil", err)
	}

	gotChallenge, err := queries.GetLatestPendingAuthLoginChallengeByProviderAndSubject(ctx, sqlc.GetLatestPendingAuthLoginChallengeByProviderAndSubjectParams{
		Provider:        "email",
		ProviderSubject: "fan@example.com",
	})
	if err != nil {
		t.Fatalf("GetLatestPendingAuthLoginChallengeByProviderAndSubject() error = %v, want nil", err)
	}
	if gotChallenge.ID != challenge.ID {
		t.Fatalf("GetLatestPendingAuthLoginChallengeByProviderAndSubject() id got %v want %v", gotChallenge.ID, challenge.ID)
	}

	challenge, err = queries.IncrementAuthLoginChallengeAttemptCount(ctx, challenge.ID)
	if err != nil {
		t.Fatalf("IncrementAuthLoginChallengeAttemptCount() error = %v, want nil", err)
	}
	if challenge.AttemptCount != 1 {
		t.Fatalf("IncrementAuthLoginChallengeAttemptCount() attempt_count got %d want 1", challenge.AttemptCount)
	}

	challenge, err = queries.ConsumeAuthLoginChallenge(ctx, sqlc.ConsumeAuthLoginChallengeParams{
		ID:         challenge.ID,
		ConsumedAt: pgTime(now.Add(2 * time.Minute)),
	})
	if err != nil {
		t.Fatalf("ConsumeAuthLoginChallenge() error = %v, want nil", err)
	}
	if !challenge.ConsumedAt.Valid {
		t.Fatal("ConsumeAuthLoginChallenge() consumed_at valid = false, want true")
	}
	if _, err := queries.ConsumeAuthLoginChallenge(ctx, sqlc.ConsumeAuthLoginChallengeParams{
		ID:         challenge.ID,
		ConsumedAt: pgTime(now.Add(3 * time.Minute)),
	}); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("ConsumeAuthLoginChallenge() second call error got %v want %v", err, pgx.ErrNoRows)
	}

	session, err := queries.CreateAuthSession(ctx, sqlc.CreateAuthSessionParams{
		UserID:                user.ID,
		ActiveMode:            "fan",
		SessionTokenHash:      "session-token-hash",
		ExpiresAt:             pgTime(now.Add(24 * time.Hour)),
		RecentAuthenticatedAt: pgTime(now),
	})
	if err != nil {
		t.Fatalf("CreateAuthSession() error = %v, want nil", err)
	}
	if !session.RecentAuthenticatedAt.Time.Equal(now) {
		t.Fatalf("CreateAuthSession() recent_authenticated_at got %s want %s", session.RecentAuthenticatedAt.Time, now)
	}

	gotSession, err := queries.GetActiveAuthSessionByTokenHash(ctx, "session-token-hash")
	if err != nil {
		t.Fatalf("GetActiveAuthSessionByTokenHash() error = %v, want nil", err)
	}
	if gotSession.ID != session.ID {
		t.Fatalf("GetActiveAuthSessionByTokenHash() id got %v want %v", gotSession.ID, session.ID)
	}

	touchedByToken, err := queries.TouchAuthSessionLastSeenByTokenHash(ctx, sqlc.TouchAuthSessionLastSeenByTokenHashParams{
		LastSeenAt:       pgTime(now.Add(4 * time.Minute)),
		SessionTokenHash: "session-token-hash",
	})
	if err != nil {
		t.Fatalf("TouchAuthSessionLastSeenByTokenHash() error = %v, want nil", err)
	}
	if !touchedByToken.LastSeenAt.Time.Equal(now.Add(4 * time.Minute)) {
		t.Fatalf("TouchAuthSessionLastSeenByTokenHash() last_seen_at got %s want %s", touchedByToken.LastSeenAt.Time, now.Add(4*time.Minute))
	}

	refreshedSession, err := queries.RefreshAuthSessionRecentAuthenticatedAtByTokenHash(
		ctx,
		sqlc.RefreshAuthSessionRecentAuthenticatedAtByTokenHashParams{
			SessionTokenHash:      "session-token-hash",
			RecentAuthenticatedAt: pgTime(now.Add(5 * time.Minute)),
		},
	)
	if err != nil {
		t.Fatalf("RefreshAuthSessionRecentAuthenticatedAtByTokenHash() error = %v, want nil", err)
	}
	if !refreshedSession.RecentAuthenticatedAt.Time.Equal(now.Add(5 * time.Minute)) {
		t.Fatalf(
			"RefreshAuthSessionRecentAuthenticatedAtByTokenHash() recent_authenticated_at got %s want %s",
			refreshedSession.RecentAuthenticatedAt.Time,
			now.Add(5*time.Minute),
		)
	}

	currentViewer, err := queries.GetCurrentViewerBySessionTokenHash(ctx, "session-token-hash")
	if err != nil {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() before capability error = %v, want nil", err)
	}
	if currentViewer.UserID != user.ID {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() user id got %v want %v", currentViewer.UserID, user.ID)
	}
	if currentViewer.ActiveMode != "fan" {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() active mode got %q want %q", currentViewer.ActiveMode, "fan")
	}
	if currentViewer.CanAccessCreatorMode {
		t.Fatal("GetCurrentViewerBySessionTokenHash() can_access_creator_mode = true, want false before capability")
	}

	_, err = queries.CreateCreatorCapability(ctx, sqlc.CreateCreatorCapabilityParams{
		UserID:                  user.ID,
		State:                   "approved",
		IsResubmitEligible:      false,
		IsSupportReviewRequired: false,
		SelfServeResubmitCount:  0,
		ApprovedAt:              pgTime(now),
	})
	if err != nil {
		t.Fatalf("CreateCreatorCapability() for current viewer bootstrap error = %v, want nil", err)
	}

	session, err = queries.TouchAuthSession(ctx, sqlc.TouchAuthSessionParams{
		ID:         session.ID,
		ActiveMode: "creator",
		LastSeenAt: pgTime(now.Add(5 * time.Minute)),
	})
	if err != nil {
		t.Fatalf("TouchAuthSession() error = %v, want nil", err)
	}
	if session.ActiveMode != "creator" {
		t.Fatalf("TouchAuthSession() active_mode got %q want %q", session.ActiveMode, "creator")
	}

	currentViewer, err = queries.GetCurrentViewerBySessionTokenHash(ctx, "session-token-hash")
	if err != nil {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() after capability error = %v, want nil", err)
	}
	if currentViewer.ActiveMode != "creator" {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() creator active mode got %q want %q", currentViewer.ActiveMode, "creator")
	}
	if !currentViewer.CanAccessCreatorMode {
		t.Fatal("GetCurrentViewerBySessionTokenHash() can_access_creator_mode = false, want true after capability")
	}

	_, err = queries.RevokeActiveAuthSessionByTokenHash(ctx, sqlc.RevokeActiveAuthSessionByTokenHashParams{
		RevokedAt:        pgTime(now.Add(6 * time.Minute)),
		SessionTokenHash: "session-token-hash",
	})
	if err != nil {
		t.Fatalf("RevokeActiveAuthSessionByTokenHash() error = %v, want nil", err)
	}
	if _, err := queries.GetActiveAuthSessionByTokenHash(ctx, "session-token-hash"); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("GetActiveAuthSessionByTokenHash() after revoke error got %v want %v", err, pgx.ErrNoRows)
	}

	if err := migrator.Migrate(12); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Migrate(12) error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, 12)
	assertRelationExists(t, ctx, conn, "app.auth_identities", true)
	assertRelationExists(t, ctx, conn, "app.auth_sessions", true)
	assertRelationExists(t, ctx, conn, "app.auth_login_challenges", true)

	legacySessionCreatedAt := now.Add(-2 * time.Hour)
	if _, err := conn.Exec(
		ctx,
		`INSERT INTO app.auth_sessions (
			user_id,
			active_mode,
			session_token_hash,
			expires_at,
			last_seen_at,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		secondUser.ID,
		"fan",
		"legacy-session-token-hash",
		now.Add(24*time.Hour),
		legacySessionCreatedAt.Add(10*time.Minute),
		legacySessionCreatedAt,
		legacySessionCreatedAt,
	); err != nil {
		t.Fatalf("Exec(insert legacy auth session before migration 13) error = %v, want nil", err)
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Up() second run error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, latestMigrationVersion)
	assertRelationExists(t, ctx, conn, "app.auth_identities", true)
	assertRelationExists(t, ctx, conn, "app.auth_sessions", true)
	assertRelationExists(t, ctx, conn, "app.auth_login_challenges", true)

	var backfilledRecentAuthenticatedAt time.Time
	if err := conn.QueryRow(
		ctx,
		`SELECT recent_authenticated_at
		FROM app.auth_sessions
		WHERE session_token_hash = $1`,
		"legacy-session-token-hash",
	).Scan(&backfilledRecentAuthenticatedAt); err != nil {
		t.Fatalf("QueryRow(legacy auth session recent_authenticated_at) error = %v, want nil", err)
	}
	if !backfilledRecentAuthenticatedAt.Equal(legacySessionCreatedAt) {
		t.Fatalf(
			"legacy auth session recent_authenticated_at got %s want %s",
			backfilledRecentAuthenticatedAt,
			legacySessionCreatedAt,
		)
	}

	thirdUser, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser() third user error = %v, want nil", err)
	}

	_, err = queries.CreateAuthIdentity(ctx, sqlc.CreateAuthIdentityParams{
		UserID:              thirdUser.ID,
		Provider:            "email",
		ProviderSubject:     "reapplied@example.com",
		EmailNormalized:     pgText("fan@example.com"),
		VerifiedAt:          pgTime(now),
		LastAuthenticatedAt: pgTime(now),
	})
	assertPgConstraintError(t, err, "23505", "idx_auth_identities_email_normalized")
}

func TestCreatorFollowQueriesAreIdempotent(t *testing.T) {
	ctx, conn, migrator, cleanup := newIntegrationEnvironment(t)
	defer cleanup()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Up() error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, latestMigrationVersion)

	queries := sqlc.New(conn)
	now := time.Unix(1710000000, 0).UTC()

	viewer, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser() viewer error = %v, want nil", err)
	}
	creatorUser, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser() creator error = %v, want nil", err)
	}

	_, err = queries.CreateCreatorCapability(ctx, sqlc.CreateCreatorCapabilityParams{
		UserID:                  creatorUser.ID,
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
		UserID:      creatorUser.ID,
		DisplayName: pgText("creator"),
		Handle:      "creator",
		Bio:         "public bio",
		PublishedAt: pgTime(now),
	})
	if err != nil {
		t.Fatalf("CreateCreatorProfile() error = %v, want nil", err)
	}

	if err := queries.PutCreatorFollow(ctx, sqlc.PutCreatorFollowParams{
		UserID:        viewer.ID,
		CreatorUserID: creatorUser.ID,
	}); err != nil {
		t.Fatalf("PutCreatorFollow() first call error = %v, want nil", err)
	}
	if err := queries.PutCreatorFollow(ctx, sqlc.PutCreatorFollowParams{
		UserID:        viewer.ID,
		CreatorUserID: creatorUser.ID,
	}); err != nil {
		t.Fatalf("PutCreatorFollow() second call error = %v, want nil", err)
	}

	isFollowing, err := queries.GetViewerCreatorFollowState(ctx, sqlc.GetViewerCreatorFollowStateParams{
		UserID:        viewer.ID,
		CreatorUserID: creatorUser.ID,
	})
	if err != nil {
		t.Fatalf("GetViewerCreatorFollowState() after follow error = %v, want nil", err)
	}
	if !isFollowing {
		t.Fatal("GetViewerCreatorFollowState() after follow got false want true")
	}

	fanCount, err := queries.CountCreatorFollowersByCreatorUserID(ctx, creatorUser.ID)
	if err != nil {
		t.Fatalf("CountCreatorFollowersByCreatorUserID() after follow error = %v, want nil", err)
	}
	if fanCount != 1 {
		t.Fatalf("CountCreatorFollowersByCreatorUserID() after follow got %d want %d", fanCount, 1)
	}

	if err := queries.DeleteCreatorFollow(ctx, sqlc.DeleteCreatorFollowParams{
		UserID:        viewer.ID,
		CreatorUserID: creatorUser.ID,
	}); err != nil {
		t.Fatalf("DeleteCreatorFollow() first call error = %v, want nil", err)
	}
	if err := queries.DeleteCreatorFollow(ctx, sqlc.DeleteCreatorFollowParams{
		UserID:        viewer.ID,
		CreatorUserID: creatorUser.ID,
	}); err != nil {
		t.Fatalf("DeleteCreatorFollow() second call error = %v, want nil", err)
	}

	isFollowing, err = queries.GetViewerCreatorFollowState(ctx, sqlc.GetViewerCreatorFollowStateParams{
		UserID:        viewer.ID,
		CreatorUserID: creatorUser.ID,
	})
	if err != nil {
		t.Fatalf("GetViewerCreatorFollowState() after unfollow error = %v, want nil", err)
	}
	if isFollowing {
		t.Fatal("GetViewerCreatorFollowState() after unfollow got true want false")
	}

	fanCount, err = queries.CountCreatorFollowersByCreatorUserID(ctx, creatorUser.ID)
	if err != nil {
		t.Fatalf("CountCreatorFollowersByCreatorUserID() after unfollow error = %v, want nil", err)
	}
	if fanCount != 0 {
		t.Fatalf("CountCreatorFollowersByCreatorUserID() after unfollow got %d want %d", fanCount, 0)
	}
}

func TestRecommendationFoundationMigrationRoundTrip(t *testing.T) {
	ctx, conn, migrator, cleanup := newIntegrationEnvironment(t)
	defer cleanup()

	if err := migrator.Migrate(15); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Migrate(15) error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, 15)
	assertRelationExists(t, ctx, conn, "app.recommendation_events", true)
	assertRelationExists(t, ctx, conn, "app.recommendation_viewer_short_features", true)
	assertRelationExists(t, ctx, conn, "app.recommendation_viewer_creator_features", true)
	assertRelationExists(t, ctx, conn, "app.recommendation_viewer_main_features", true)
	assertRelationExists(t, ctx, conn, "app.recommendation_short_global_features", true)

	if err := migrator.Steps(-1); err != nil {
		t.Fatalf("migrator.Steps(-1) error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, 14)
	assertRelationExists(t, ctx, conn, "app.recommendation_events", false)
	assertRelationExists(t, ctx, conn, "app.recommendation_viewer_short_features", false)
	assertRelationExists(t, ctx, conn, "app.recommendation_viewer_creator_features", false)
	assertRelationExists(t, ctx, conn, "app.recommendation_viewer_main_features", false)
	assertRelationExists(t, ctx, conn, "app.recommendation_short_global_features", false)

	if err := migrator.Migrate(15); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Migrate(15) error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, 15)
	assertRelationExists(t, ctx, conn, "app.recommendation_events", true)
	assertRelationExists(t, ctx, conn, "app.recommendation_viewer_short_features", true)
	assertRelationExists(t, ctx, conn, "app.recommendation_viewer_creator_features", true)
	assertRelationExists(t, ctx, conn, "app.recommendation_viewer_main_features", true)
	assertRelationExists(t, ctx, conn, "app.recommendation_short_global_features", true)
}

func TestRecommendationQueriesEnforceIdentityConsistency(t *testing.T) {
	ctx, conn, migrator, cleanup := newIntegrationEnvironment(t)
	defer cleanup()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Up() error = %v, want nil", err)
	}

	queries := sqlc.New(conn)
	now := time.Unix(1710000000, 0).UTC()
	fixture := createRecommendationFixture(t, ctx, queries, now)

	_, err := queries.InsertRecommendationEvent(ctx, sqlc.InsertRecommendationEventParams{
		ViewerUserID:    fixture.ViewerUserID,
		EventKind:       "impression",
		CreatorUserID:   fixture.CreatorUserID,
		CanonicalMainID: fixture.AlternateMainID,
		ShortID:         fixture.ShortID,
		OccurredAt:      pgTime(now.Add(5 * time.Minute)),
		IdempotencyKey:  "recommendation-short-mismatch",
	})
	assertPgConstraintError(t, err, "23503", "recommendation_events_short_identity_fkey")

	_, err = queries.InsertRecommendationEvent(ctx, sqlc.InsertRecommendationEventParams{
		ViewerUserID:   fixture.ViewerUserID,
		EventKind:      "view_start",
		CreatorUserID:  fixture.CreatorUserID,
		OccurredAt:     pgTime(now.Add(7 * time.Minute)),
		IdempotencyKey: "recommendation-invalid-shape",
	})
	assertPgConstraintError(t, err, "23514", "recommendation_events_payload_shape_check")

	_, err = queries.UpsertRecommendationViewerMainFeatures(ctx, sqlc.UpsertRecommendationViewerMainFeaturesParams{
		ViewerUserID:          fixture.ViewerUserID,
		CanonicalMainID:       fixture.CanonicalMainID,
		CreatorUserID:         fixture.OtherCreatorUserID,
		ImpressionCount:       1,
		LastImpressionAt:      pgTime(now.Add(10 * time.Minute)),
		ViewStartCount:        0,
		ViewCompletionCount:   0,
		RewatchLoopCount:      0,
		MainClickCount:        0,
		UnlockConversionCount: 0,
	})
	assertPgConstraintError(t, err, "23503", "recommendation_viewer_main_features_main_identity_fkey")
}

func TestRecommendationAggregateUpsertsRejectExistingRowIdentityConflict(t *testing.T) {
	ctx, conn, migrator, cleanup := newIntegrationEnvironment(t)
	defer cleanup()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Up() error = %v, want nil", err)
	}

	queries := sqlc.New(conn)
	now := time.Unix(1710000000, 0).UTC()
	fixture := createRecommendationFixture(t, ctx, queries, now)
	initialAt := now.Add(30 * time.Minute)

	rowsAffected, err := queries.UpsertRecommendationViewerShortFeatures(ctx, sqlc.UpsertRecommendationViewerShortFeaturesParams{
		ViewerUserID:          fixture.ViewerUserID,
		ShortID:               fixture.ShortID,
		CreatorUserID:         fixture.CreatorUserID,
		CanonicalMainID:       fixture.CanonicalMainID,
		ImpressionCount:       1,
		LastImpressionAt:      pgTime(initialAt),
		ViewStartCount:        0,
		ViewCompletionCount:   0,
		RewatchLoopCount:      0,
		MainClickCount:        0,
		UnlockConversionCount: 0,
	})
	if err != nil {
		t.Fatalf("UpsertRecommendationViewerShortFeatures() initial error = %v, want nil", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("UpsertRecommendationViewerShortFeatures() initial rows affected got %d want %d", rowsAffected, 1)
	}

	rowsAffected, err = queries.UpsertRecommendationViewerShortFeatures(ctx, sqlc.UpsertRecommendationViewerShortFeaturesParams{
		ViewerUserID:          fixture.ViewerUserID,
		ShortID:               fixture.ShortID,
		CreatorUserID:         fixture.CreatorUserID,
		CanonicalMainID:       fixture.AlternateMainID,
		ImpressionCount:       1,
		LastImpressionAt:      pgTime(initialAt.Add(time.Minute)),
		ViewStartCount:        0,
		ViewCompletionCount:   0,
		RewatchLoopCount:      0,
		MainClickCount:        0,
		UnlockConversionCount: 0,
	})
	if err != nil {
		t.Fatalf("UpsertRecommendationViewerShortFeatures() mismatch error = %v, want nil", err)
	}
	if rowsAffected != 0 {
		t.Fatalf("UpsertRecommendationViewerShortFeatures() mismatch rows affected got %d want %d", rowsAffected, 0)
	}

	shortRows, err := queries.ListRecommendationViewerShortFeaturesByViewerAndShortIDs(ctx, sqlc.ListRecommendationViewerShortFeaturesByViewerAndShortIDsParams{
		ViewerUserID: fixture.ViewerUserID,
		ShortIds:     []pgtype.UUID{fixture.ShortID},
	})
	if err != nil {
		t.Fatalf("ListRecommendationViewerShortFeaturesByViewerAndShortIDs() error = %v, want nil", err)
	}
	if len(shortRows) != 1 || shortRows[0].ImpressionCount != 1 {
		t.Fatalf("ListRecommendationViewerShortFeaturesByViewerAndShortIDs() got %#v", shortRows)
	}

	rowsAffected, err = queries.UpsertRecommendationViewerMainFeatures(ctx, sqlc.UpsertRecommendationViewerMainFeaturesParams{
		ViewerUserID:          fixture.ViewerUserID,
		CanonicalMainID:       fixture.CanonicalMainID,
		CreatorUserID:         fixture.CreatorUserID,
		ImpressionCount:       1,
		LastImpressionAt:      pgTime(initialAt),
		ViewStartCount:        0,
		ViewCompletionCount:   0,
		RewatchLoopCount:      0,
		MainClickCount:        0,
		UnlockConversionCount: 0,
	})
	if err != nil {
		t.Fatalf("UpsertRecommendationViewerMainFeatures() initial error = %v, want nil", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("UpsertRecommendationViewerMainFeatures() initial rows affected got %d want %d", rowsAffected, 1)
	}

	rowsAffected, err = queries.UpsertRecommendationViewerMainFeatures(ctx, sqlc.UpsertRecommendationViewerMainFeaturesParams{
		ViewerUserID:          fixture.ViewerUserID,
		CanonicalMainID:       fixture.CanonicalMainID,
		CreatorUserID:         fixture.OtherCreatorUserID,
		ImpressionCount:       1,
		LastImpressionAt:      pgTime(initialAt.Add(time.Minute)),
		ViewStartCount:        0,
		ViewCompletionCount:   0,
		RewatchLoopCount:      0,
		MainClickCount:        0,
		UnlockConversionCount: 0,
	})
	if err != nil {
		t.Fatalf("UpsertRecommendationViewerMainFeatures() mismatch error = %v, want nil", err)
	}
	if rowsAffected != 0 {
		t.Fatalf("UpsertRecommendationViewerMainFeatures() mismatch rows affected got %d want %d", rowsAffected, 0)
	}

	mainRows, err := queries.ListRecommendationViewerMainFeaturesByViewerAndMainIDs(ctx, sqlc.ListRecommendationViewerMainFeaturesByViewerAndMainIDsParams{
		ViewerUserID:     fixture.ViewerUserID,
		CanonicalMainIds: []pgtype.UUID{fixture.CanonicalMainID},
	})
	if err != nil {
		t.Fatalf("ListRecommendationViewerMainFeaturesByViewerAndMainIDs() error = %v, want nil", err)
	}
	if len(mainRows) != 1 || mainRows[0].ImpressionCount != 1 {
		t.Fatalf("ListRecommendationViewerMainFeaturesByViewerAndMainIDs() got %#v", mainRows)
	}

	rowsAffected, err = queries.UpsertRecommendationShortGlobalFeatures(ctx, sqlc.UpsertRecommendationShortGlobalFeaturesParams{
		ShortID:               fixture.ShortID,
		CreatorUserID:         fixture.CreatorUserID,
		CanonicalMainID:       fixture.CanonicalMainID,
		ImpressionCount:       1,
		LastImpressionAt:      pgTime(initialAt),
		ViewStartCount:        0,
		ViewCompletionCount:   0,
		RewatchLoopCount:      0,
		MainClickCount:        0,
		UnlockConversionCount: 0,
	})
	if err != nil {
		t.Fatalf("UpsertRecommendationShortGlobalFeatures() initial error = %v, want nil", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("UpsertRecommendationShortGlobalFeatures() initial rows affected got %d want %d", rowsAffected, 1)
	}

	rowsAffected, err = queries.UpsertRecommendationShortGlobalFeatures(ctx, sqlc.UpsertRecommendationShortGlobalFeaturesParams{
		ShortID:               fixture.ShortID,
		CreatorUserID:         fixture.CreatorUserID,
		CanonicalMainID:       fixture.AlternateMainID,
		ImpressionCount:       1,
		LastImpressionAt:      pgTime(initialAt.Add(time.Minute)),
		ViewStartCount:        0,
		ViewCompletionCount:   0,
		RewatchLoopCount:      0,
		MainClickCount:        0,
		UnlockConversionCount: 0,
	})
	if err != nil {
		t.Fatalf("UpsertRecommendationShortGlobalFeatures() mismatch error = %v, want nil", err)
	}
	if rowsAffected != 0 {
		t.Fatalf("UpsertRecommendationShortGlobalFeatures() mismatch rows affected got %d want %d", rowsAffected, 0)
	}

	globalRows, err := queries.ListRecommendationShortGlobalFeaturesByShortIDs(ctx, []pgtype.UUID{fixture.ShortID})
	if err != nil {
		t.Fatalf("ListRecommendationShortGlobalFeaturesByShortIDs() error = %v, want nil", err)
	}
	if len(globalRows) != 1 || globalRows[0].ImpressionCount != 1 {
		t.Fatalf("ListRecommendationShortGlobalFeaturesByShortIDs() got %#v", globalRows)
	}
}

func TestRecommendationDuplicateEventSequenceDoesNotDoubleCount(t *testing.T) {
	ctx, conn, migrator, cleanup := newIntegrationEnvironment(t)
	defer cleanup()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Up() error = %v, want nil", err)
	}

	queries := sqlc.New(conn)
	now := time.Unix(1710000000, 0).UTC()
	fixture := createRecommendationFixture(t, ctx, queries, now)
	occurredAt := now.Add(15 * time.Minute)

	insertParams := sqlc.InsertRecommendationEventParams{
		ViewerUserID:    fixture.ViewerUserID,
		EventKind:       "impression",
		CreatorUserID:   fixture.CreatorUserID,
		CanonicalMainID: fixture.CanonicalMainID,
		ShortID:         fixture.ShortID,
		OccurredAt:      pgTime(occurredAt),
		IdempotencyKey:  "recommendation-dup-1",
	}

	firstRow, err := queries.InsertRecommendationEvent(ctx, insertParams)
	if err != nil {
		t.Fatalf("InsertRecommendationEvent() first call error = %v, want nil", err)
	}

	rowsAffected, err := queries.UpsertRecommendationViewerShortFeatures(ctx, sqlc.UpsertRecommendationViewerShortFeaturesParams{
		ViewerUserID:          fixture.ViewerUserID,
		ShortID:               fixture.ShortID,
		CreatorUserID:         fixture.CreatorUserID,
		CanonicalMainID:       fixture.CanonicalMainID,
		ImpressionCount:       1,
		LastImpressionAt:      pgTime(occurredAt),
		ViewStartCount:        0,
		ViewCompletionCount:   0,
		RewatchLoopCount:      0,
		MainClickCount:        0,
		UnlockConversionCount: 0,
	})
	if err != nil {
		t.Fatalf("UpsertRecommendationViewerShortFeatures() first call error = %v, want nil", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("UpsertRecommendationViewerShortFeatures() first call rows affected got %d want %d", rowsAffected, 1)
	}

	_, err = queries.InsertRecommendationEvent(ctx, insertParams)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("InsertRecommendationEvent() duplicate call error got %v want %v", err, pgx.ErrNoRows)
	}

	secondRow, err := queries.GetRecommendationEventByViewerAndIdempotencyKey(ctx, sqlc.GetRecommendationEventByViewerAndIdempotencyKeyParams{
		ViewerUserID:   fixture.ViewerUserID,
		IdempotencyKey: insertParams.IdempotencyKey,
	})
	if err != nil {
		t.Fatalf("GetRecommendationEventByViewerAndIdempotencyKey() error = %v, want nil", err)
	}
	if secondRow.ID != firstRow.ID {
		t.Fatalf("GetRecommendationEventByViewerAndIdempotencyKey() id got %v want %v", secondRow.ID, firstRow.ID)
	}

	rows, err := queries.ListRecommendationViewerShortFeaturesByViewerAndShortIDs(ctx, sqlc.ListRecommendationViewerShortFeaturesByViewerAndShortIDsParams{
		ViewerUserID: fixture.ViewerUserID,
		ShortIds:     []pgtype.UUID{fixture.ShortID},
	})
	if err != nil {
		t.Fatalf("ListRecommendationViewerShortFeaturesByViewerAndShortIDs() error = %v, want nil", err)
	}
	if len(rows) != 1 {
		t.Fatalf("ListRecommendationViewerShortFeaturesByViewerAndShortIDs() len got %d want %d", len(rows), 1)
	}
	if rows[0].ImpressionCount != 1 {
		t.Fatalf("ListRecommendationViewerShortFeaturesByViewerAndShortIDs() impression_count got %d want %d", rows[0].ImpressionCount, 1)
	}
	if !rows[0].LastImpressionAt.Valid || !rows[0].LastImpressionAt.Time.Equal(occurredAt) {
		t.Fatalf("ListRecommendationViewerShortFeaturesByViewerAndShortIDs() last_impression_at got %#v want %s", rows[0].LastImpressionAt, occurredAt)
	}
}

func TestRecommendationViewerShortFeatureUpsertKeepsLatestTimestamp(t *testing.T) {
	ctx, conn, migrator, cleanup := newIntegrationEnvironment(t)
	defer cleanup()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Up() error = %v, want nil", err)
	}

	queries := sqlc.New(conn)
	now := time.Unix(1710000000, 0).UTC()
	fixture := createRecommendationFixture(t, ctx, queries, now)
	latestAt := now.Add(20 * time.Minute)
	olderAt := now.Add(10 * time.Minute)

	rowsAffected, err := queries.UpsertRecommendationViewerShortFeatures(ctx, sqlc.UpsertRecommendationViewerShortFeaturesParams{
		ViewerUserID:          fixture.ViewerUserID,
		ShortID:               fixture.ShortID,
		CreatorUserID:         fixture.CreatorUserID,
		CanonicalMainID:       fixture.CanonicalMainID,
		ImpressionCount:       1,
		LastImpressionAt:      pgTime(latestAt),
		ViewStartCount:        0,
		ViewCompletionCount:   0,
		RewatchLoopCount:      0,
		MainClickCount:        0,
		UnlockConversionCount: 0,
	})
	if err != nil {
		t.Fatalf("UpsertRecommendationViewerShortFeatures() latest event error = %v, want nil", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("UpsertRecommendationViewerShortFeatures() latest event rows affected got %d want %d", rowsAffected, 1)
	}

	rowsAffected, err = queries.UpsertRecommendationViewerShortFeatures(ctx, sqlc.UpsertRecommendationViewerShortFeaturesParams{
		ViewerUserID:          fixture.ViewerUserID,
		ShortID:               fixture.ShortID,
		CreatorUserID:         fixture.CreatorUserID,
		CanonicalMainID:       fixture.CanonicalMainID,
		ImpressionCount:       1,
		LastImpressionAt:      pgTime(olderAt),
		ViewStartCount:        0,
		ViewCompletionCount:   0,
		RewatchLoopCount:      0,
		MainClickCount:        0,
		UnlockConversionCount: 0,
	})
	if err != nil {
		t.Fatalf("UpsertRecommendationViewerShortFeatures() older event error = %v, want nil", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("UpsertRecommendationViewerShortFeatures() older event rows affected got %d want %d", rowsAffected, 1)
	}

	rows, err := queries.ListRecommendationViewerShortFeaturesByViewerAndShortIDs(ctx, sqlc.ListRecommendationViewerShortFeaturesByViewerAndShortIDsParams{
		ViewerUserID: fixture.ViewerUserID,
		ShortIds:     []pgtype.UUID{fixture.ShortID},
	})
	if err != nil {
		t.Fatalf("ListRecommendationViewerShortFeaturesByViewerAndShortIDs() error = %v, want nil", err)
	}
	if len(rows) != 1 {
		t.Fatalf("ListRecommendationViewerShortFeaturesByViewerAndShortIDs() len got %d want %d", len(rows), 1)
	}
	if rows[0].ImpressionCount != 2 {
		t.Fatalf("ListRecommendationViewerShortFeaturesByViewerAndShortIDs() impression_count got %d want %d", rows[0].ImpressionCount, 2)
	}
	if !rows[0].LastImpressionAt.Valid || !rows[0].LastImpressionAt.Time.Equal(latestAt) {
		t.Fatalf("ListRecommendationViewerShortFeaturesByViewerAndShortIDs() last_impression_at got %#v want %s", rows[0].LastImpressionAt, latestAt)
	}
}

func TestMediaAssetProcessingMigrationLatestRevision(t *testing.T) {
	ctx, conn, migrator, cleanup := newIntegrationEnvironment(t)
	defer cleanup()

	if err := migrator.Migrate(4); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Migrate(4) error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, 4)

	queries := sqlc.New(conn)
	now := time.Unix(1710000000, 0).UTC()

	creator, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser() error = %v, want nil", err)
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

	processingAsset, err := queries.CreateMediaAsset(ctx, sqlc.CreateMediaAssetParams{
		CreatorUserID:   creator.ID,
		ProcessingState: "processing",
		StorageProvider: "s3",
		StorageBucket:   "test-bucket",
		StorageKey:      "processing-asset",
		MimeType:        "video/mp4",
	})
	if err != nil {
		t.Fatalf("CreateMediaAsset(processing) error = %v, want nil", err)
	}
	if processingAsset.ProcessingState != "processing" {
		t.Fatalf("CreateMediaAsset(processing) state got %q want %q", processingAsset.ProcessingState, "processing")
	}
	if processingAsset.PlaybackUrl.Valid {
		t.Fatalf("CreateMediaAsset(processing) playback_url valid got %t want false", processingAsset.PlaybackUrl.Valid)
	}

	readyAsset, err := queries.UpdateMediaAssetProcessingState(ctx, sqlc.UpdateMediaAssetProcessingStateParams{
		ID:              processingAsset.ID,
		ProcessingState: "ready",
		PlaybackUrl:     pgText("https://cdn.example.com/processing-asset.m3u8"),
		DurationMs:      pgInt64(1000),
	})
	if err != nil {
		t.Fatalf("UpdateMediaAssetProcessingState(ready before down) error = %v, want nil", err)
	}
	if readyAsset.ProcessingState != "ready" {
		t.Fatalf("UpdateMediaAssetProcessingState(ready before down) state got %q want %q", readyAsset.ProcessingState, "ready")
	}

	if err := migrator.Steps(-1); err != nil {
		t.Fatalf("migrator.Steps(-1) error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, 3)

	_, err = queries.CreateMediaAsset(ctx, sqlc.CreateMediaAssetParams{
		CreatorUserID:   creator.ID,
		ProcessingState: "processing",
		StorageProvider: "s3",
		StorageBucket:   "test-bucket",
		StorageKey:      "processing-asset-v3",
		MimeType:        "video/mp4",
	})
	if err == nil {
		t.Fatal("CreateMediaAsset(processing) at version 3 error = nil, want constraint error")
	}

	if err := migrator.Migrate(4); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Migrate(4) second run error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, 4)

	processingAssetAgain, err := queries.CreateMediaAsset(ctx, sqlc.CreateMediaAssetParams{
		CreatorUserID:   creator.ID,
		ProcessingState: "processing",
		StorageProvider: "s3",
		StorageBucket:   "test-bucket",
		StorageKey:      "processing-asset-v4",
		MimeType:        "video/mp4",
	})
	if err != nil {
		t.Fatalf("CreateMediaAsset(processing) after re-up error = %v, want nil", err)
	}
	if processingAssetAgain.ProcessingState != "processing" {
		t.Fatalf("CreateMediaAsset(processing) after re-up state got %q want %q", processingAssetAgain.ProcessingState, "processing")
	}
}

func TestCreatorProfileHandleQueriesLatestRevision(t *testing.T) {
	ctx, conn, migrator, cleanup := newIntegrationEnvironment(t)
	defer cleanup()

	if err := migrator.Migrate(3); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Migrate(3) error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, 3)

	queries := sqlc.New(conn)
	now := time.Unix(1710000000, 0).UTC()

	creatorA, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser(creatorA) error = %v, want nil", err)
	}
	creatorB, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser(creatorB) error = %v, want nil", err)
	}
	creatorPrivate, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser(creatorPrivate) error = %v, want nil", err)
	}

	for _, creatorID := range []pgtype.UUID{creatorA.ID, creatorB.ID, creatorPrivate.ID} {
		if _, err := queries.CreateCreatorCapability(ctx, sqlc.CreateCreatorCapabilityParams{
			UserID:                  creatorID,
			State:                   "approved",
			IsResubmitEligible:      false,
			IsSupportReviewRequired: false,
			SelfServeResubmitCount:  0,
			ApprovedAt:              pgTime(now),
		}); err != nil {
			t.Fatalf("CreateCreatorCapability(%v) error = %v, want nil", creatorID, err)
		}
	}

	if _, err := conn.Exec(
		ctx,
		`INSERT INTO app.creator_profiles (user_id, display_name, avatar_url, bio, published_at)
		VALUES ($1, $2, $3, $4, $5), ($6, $7, $8, $9, $10), ($11, $12, $13, $14, $15)`,
		creatorA.ID,
		"Mina Rei",
		"https://cdn.example.com/creator/mina-a/avatar.jpg",
		"bio-a",
		pgTime(now),
		creatorB.ID,
		"Mina Rei",
		"https://cdn.example.com/creator/mina-b/avatar.jpg",
		"bio-b",
		pgTime(now.Add(-time.Hour)),
		creatorPrivate.ID,
		"Private Mina",
		"https://cdn.example.com/creator/mina-private/avatar.jpg",
		"bio-private",
		pgtype.Timestamptz{},
	); err != nil {
		t.Fatalf("INSERT creator_profiles(version3) error = %v, want nil", err)
	}

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Up() to latest error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, latestMigrationVersion)

	profileA, err := queries.GetCreatorProfileByUserID(ctx, creatorA.ID)
	if err != nil {
		t.Fatalf("GetCreatorProfileByUserID(creatorA) error = %v, want nil", err)
	}
	profileB, err := queries.GetCreatorProfileByUserID(ctx, creatorB.ID)
	if err != nil {
		t.Fatalf("GetCreatorProfileByUserID(creatorB) error = %v, want nil", err)
	}
	if profileA.Handle == profileB.Handle {
		t.Fatalf("backfilled handles got duplicate %q want unique handles", profileA.Handle)
	}
	var primaryCreatorID pgtype.UUID
	switch {
	case profileA.Handle == "minarei" && strings.HasPrefix(profileB.Handle, "minarei_"):
		primaryCreatorID = creatorA.ID
	case profileB.Handle == "minarei" && strings.HasPrefix(profileA.Handle, "minarei_"):
		primaryCreatorID = creatorB.ID
	default:
		t.Fatalf("backfilled handles got creatorA=%q creatorB=%q want one exact and one suffixed handle", profileA.Handle, profileB.Handle)
	}

	gotByHandle, err := queries.GetPublicCreatorProfileByHandle(ctx, "minarei")
	if err != nil {
		t.Fatalf("GetPublicCreatorProfileByHandle() error = %v, want nil", err)
	}
	if gotByHandle.UserID != primaryCreatorID {
		t.Fatalf("GetPublicCreatorProfileByHandle() user got %v want %v", gotByHandle.UserID, primaryCreatorID)
	}

	privateProfile, err := queries.GetCreatorProfileByUserID(ctx, creatorPrivate.ID)
	if err != nil {
		t.Fatalf("GetCreatorProfileByUserID(creatorPrivate) error = %v, want nil", err)
	}
	if !strings.HasPrefix(privateProfile.Handle, "privatemina") {
		t.Fatalf("creatorPrivate handle got %q want prefix %q", privateProfile.Handle, "privatemina")
	}
	if privateProfile.PublishedAt.Valid {
		t.Fatalf("creatorPrivate published_at valid got %t want false", privateProfile.PublishedAt.Valid)
	}

	recentProfiles, err := queries.ListRecentPublicCreatorProfiles(ctx, sqlc.ListRecentPublicCreatorProfilesParams{
		LimitCount: 10,
	})
	if err != nil {
		t.Fatalf("ListRecentPublicCreatorProfiles() error = %v, want nil", err)
	}
	if len(recentProfiles) != 2 {
		t.Fatalf("ListRecentPublicCreatorProfiles() len got %d want %d", len(recentProfiles), 2)
	}
	if recentProfiles[0].UserID != creatorA.ID || recentProfiles[1].UserID != creatorB.ID {
		t.Fatalf("ListRecentPublicCreatorProfiles() order got [%v %v] want [%v %v]", recentProfiles[0].UserID, recentProfiles[1].UserID, creatorA.ID, creatorB.ID)
	}

	filteredProfiles, err := queries.SearchPublicCreatorProfiles(ctx, sqlc.SearchPublicCreatorProfilesParams{
		DisplayNameQuery:  pgText("mina"),
		HandlePrefixQuery: "mina",
		LimitCount:        10,
	})
	if err != nil {
		t.Fatalf("SearchPublicCreatorProfiles() error = %v, want nil", err)
	}
	if len(filteredProfiles) != 2 {
		t.Fatalf("SearchPublicCreatorProfiles() len got %d want %d", len(filteredProfiles), 2)
	}

	creatorC, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser(creatorC) error = %v, want nil", err)
	}
	creatorD, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser(creatorD) error = %v, want nil", err)
	}
	for _, creatorID := range []pgtype.UUID{creatorC.ID, creatorD.ID} {
		if _, err := queries.CreateCreatorCapability(ctx, sqlc.CreateCreatorCapabilityParams{
			UserID:                  creatorID,
			State:                   "approved",
			IsResubmitEligible:      false,
			IsSupportReviewRequired: false,
			SelfServeResubmitCount:  0,
			ApprovedAt:              pgTime(now),
		}); err != nil {
			t.Fatalf("CreateCreatorCapability(%v) after latest error = %v, want nil", creatorID, err)
		}
	}
	if _, err := queries.CreateCreatorProfile(ctx, sqlc.CreateCreatorProfileParams{
		UserID:      creatorC.ID,
		DisplayName: pgText("A Under"),
		Handle:      "a_b",
		Bio:         "bio-c",
		PublishedAt: pgTime(now.Add(-2 * time.Hour)),
	}); err != nil {
		t.Fatalf("CreateCreatorProfile(creatorC) error = %v, want nil", err)
	}
	if _, err := queries.CreateCreatorProfile(ctx, sqlc.CreateCreatorProfileParams{
		UserID:      creatorD.ID,
		DisplayName: pgText("AB Plain"),
		Handle:      "ab",
		Bio:         "bio-d",
		PublishedAt: pgTime(now.Add(-3 * time.Hour)),
	}); err != nil {
		t.Fatalf("CreateCreatorProfile(creatorD) error = %v, want nil", err)
	}

	literalPercentProfiles, err := queries.SearchPublicCreatorProfiles(ctx, sqlc.SearchPublicCreatorProfilesParams{
		DisplayNameQuery:  pgText(`\%`),
		HandlePrefixQuery: "",
		LimitCount:        10,
	})
	if err != nil {
		t.Fatalf("SearchPublicCreatorProfiles(literal percent) error = %v, want nil", err)
	}
	if len(literalPercentProfiles) != 0 {
		t.Fatalf("SearchPublicCreatorProfiles(literal percent) len got %d want %d", len(literalPercentProfiles), 0)
	}

	literalUnderscoreProfiles, err := queries.SearchPublicCreatorProfiles(ctx, sqlc.SearchPublicCreatorProfilesParams{
		DisplayNameQuery:  pgText(`a\_`),
		HandlePrefixQuery: `a\_`,
		LimitCount:        10,
	})
	if err != nil {
		t.Fatalf("SearchPublicCreatorProfiles(literal underscore) error = %v, want nil", err)
	}
	if len(literalUnderscoreProfiles) != 1 {
		t.Fatalf("SearchPublicCreatorProfiles(literal underscore) len got %d want %d", len(literalUnderscoreProfiles), 1)
	}
	if literalUnderscoreProfiles[0].UserID != creatorC.ID {
		t.Fatalf("SearchPublicCreatorProfiles(literal underscore) user got %v want %v", literalUnderscoreProfiles[0].UserID, creatorC.ID)
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
		PriceMinor:          1200,
		CurrencyCode:        "JPY",
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
		PriceMinor:         1200,
		CurrencyCode:       "JPY",
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

func TestPaymentQueriesLatestRevision(t *testing.T) {
	ctx, conn, migrator, cleanup := newIntegrationEnvironment(t)
	defer cleanup()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Up() error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, latestMigrationVersion)

	queries := sqlc.New(conn)
	now := time.Unix(1710000000, 0).UTC()

	buyer, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser(buyer) error = %v, want nil", err)
	}
	creator, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser(creator) error = %v, want nil", err)
	}

	if _, err := queries.CreateCreatorCapability(ctx, sqlc.CreateCreatorCapabilityParams{
		UserID:                  creator.ID,
		State:                   "approved",
		IsResubmitEligible:      false,
		IsSupportReviewRequired: false,
		SelfServeResubmitCount:  0,
		ApprovedAt:              pgTime(now),
	}); err != nil {
		t.Fatalf("CreateCreatorCapability() error = %v, want nil", err)
	}

	mainAsset, err := createReadyMediaAsset(ctx, queries, creator.ID, "payment-main")
	if err != nil {
		t.Fatalf("createReadyMediaAsset(main) error = %v, want nil", err)
	}
	shortAsset, err := createReadyMediaAsset(ctx, queries, creator.ID, "payment-short")
	if err != nil {
		t.Fatalf("createReadyMediaAsset(short) error = %v, want nil", err)
	}

	main, err := queries.CreateMain(ctx, sqlc.CreateMainParams{
		CreatorUserID:       creator.ID,
		MediaAssetID:        mainAsset.ID,
		State:               "approved_for_unlock",
		PriceMinor:          1800,
		CurrencyCode:        "JPY",
		OwnershipConfirmed:  true,
		ConsentConfirmed:    true,
		ApprovedForUnlockAt: pgTime(now.Add(time.Hour)),
	})
	if err != nil {
		t.Fatalf("CreateMain() error = %v, want nil", err)
	}
	short, err := queries.CreateShort(ctx, sqlc.CreateShortParams{
		CreatorUserID:        creator.ID,
		CanonicalMainID:      main.ID,
		MediaAssetID:         shortAsset.ID,
		State:                "approved_for_publish",
		ApprovedForPublishAt: pgTime(now.Add(2 * time.Hour)),
		PublishedAt:          pgTime(now.Add(3 * time.Hour)),
	})
	if err != nil {
		t.Fatalf("CreateShort() error = %v, want nil", err)
	}

	method, err := queries.UpsertUserPaymentMethod(ctx, sqlc.UpsertUserPaymentMethodParams{
		UserID:                    buyer.ID,
		Provider:                  "ccbill",
		ProviderPaymentTokenRef:   "token-initial",
		ProviderPaymentAccountRef: "account-1",
		Brand:                     "visa",
		Last4:                     "4242",
		LastUsedAt:                pgTime(now),
	})
	if err != nil {
		t.Fatalf("UpsertUserPaymentMethod(initial) error = %v, want nil", err)
	}
	if method.Provider != "ccbill" || method.Brand != "visa" || method.Last4 != "4242" {
		t.Fatalf("UpsertUserPaymentMethod(initial) got %#v", method)
	}

	updatedMethod, err := queries.UpsertUserPaymentMethod(ctx, sqlc.UpsertUserPaymentMethodParams{
		UserID:                    buyer.ID,
		Provider:                  "ccbill",
		ProviderPaymentTokenRef:   "token-updated",
		ProviderPaymentAccountRef: "account-1",
		Brand:                     "mastercard",
		Last4:                     "4444",
		LastUsedAt:                pgTime(now.Add(time.Minute)),
	})
	if err != nil {
		t.Fatalf("UpsertUserPaymentMethod(update) error = %v, want nil", err)
	}
	if updatedMethod.ID != method.ID {
		t.Fatalf("UpsertUserPaymentMethod(update) id got %v want %v", updatedMethod.ID, method.ID)
	}
	if updatedMethod.ProviderPaymentTokenRef != "token-updated" || updatedMethod.Brand != "mastercard" || updatedMethod.Last4 != "4444" {
		t.Fatalf("UpsertUserPaymentMethod(update) got %#v", updatedMethod)
	}

	listedMethods, err := queries.ListUserPaymentMethodsByUserID(ctx, buyer.ID)
	if err != nil {
		t.Fatalf("ListUserPaymentMethodsByUserID() error = %v, want nil", err)
	}
	if len(listedMethods) != 1 || listedMethods[0].ID != method.ID {
		t.Fatalf("ListUserPaymentMethodsByUserID() got %#v want single method %v", listedMethods, method.ID)
	}

	touchedMethod, err := queries.TouchUserPaymentMethodLastUsedAt(ctx, sqlc.TouchUserPaymentMethodLastUsedAtParams{
		LastUsedAt: pgTime(now.Add(2 * time.Minute)),
		ID:         method.ID,
		UserID:     buyer.ID,
	})
	if err != nil {
		t.Fatalf("TouchUserPaymentMethodLastUsedAt() error = %v, want nil", err)
	}
	if !touchedMethod.LastUsedAt.Time.Equal(now.Add(2 * time.Minute)) {
		t.Fatalf("TouchUserPaymentMethodLastUsedAt() last_used_at got %s want %s", touchedMethod.LastUsedAt.Time, now.Add(2*time.Minute))
	}

	attempt, err := queries.CreateMainPurchaseAttempt(ctx, sqlc.CreateMainPurchaseAttemptParams{
		UserID:                  buyer.ID,
		MainID:                  main.ID,
		FromShortID:             short.ID,
		Provider:                "ccbill",
		PaymentMethodMode:       "saved_card",
		UserPaymentMethodID:     method.ID,
		ProviderPaymentTokenRef: "token-updated",
		IdempotencyKey:          "purchase-attempt-1",
		Status:                  "processing",
		RequestedPriceJpy:       1800,
		RequestedCurrencyCode:   392,
		AcceptedAge:             true,
		AcceptedTerms:           true,
	})
	if err != nil {
		t.Fatalf("CreateMainPurchaseAttempt() error = %v, want nil", err)
	}
	if attempt.UserPaymentMethodID != method.ID || attempt.Status != "processing" {
		t.Fatalf("CreateMainPurchaseAttempt() got %#v", attempt)
	}

	inflight, err := queries.GetLatestInflightMainPurchaseAttemptByUserIDAndMainIDForUpdate(ctx, sqlc.GetLatestInflightMainPurchaseAttemptByUserIDAndMainIDForUpdateParams{
		UserID: buyer.ID,
		MainID: main.ID,
	})
	if err != nil {
		t.Fatalf("GetLatestInflightMainPurchaseAttemptByUserIDAndMainIDForUpdate() error = %v, want nil", err)
	}
	if inflight.ID != attempt.ID {
		t.Fatalf("GetLatestInflightMainPurchaseAttemptByUserIDAndMainIDForUpdate() id got %v want %v", inflight.ID, attempt.ID)
	}

	_, err = queries.CreateMainPurchaseAttempt(ctx, sqlc.CreateMainPurchaseAttemptParams{
		UserID:                  buyer.ID,
		MainID:                  main.ID,
		FromShortID:             short.ID,
		Provider:                "ccbill",
		PaymentMethodMode:       "saved_card",
		UserPaymentMethodID:     method.ID,
		ProviderPaymentTokenRef: "token-updated",
		IdempotencyKey:          "purchase-attempt-2",
		Status:                  "processing",
		RequestedPriceJpy:       1800,
		RequestedCurrencyCode:   392,
		AcceptedAge:             true,
		AcceptedTerms:           true,
	})
	assertPgConstraintError(t, err, "23505", "idx_main_purchase_attempts_user_main_inflight")

	processedAt := now.Add(4 * time.Minute)
	attempt, err = queries.UpdateMainPurchaseAttemptOutcome(ctx, sqlc.UpdateMainPurchaseAttemptOutcomeParams{
		Status:                   "succeeded",
		FailureReason:            pgtype.Text{},
		PendingReason:            pgtype.Text{},
		ProviderPurchaseRef:      pgText("purchase-ref-1"),
		ProviderTransactionRef:   pgText("txn-1"),
		ProviderSessionRef:       pgText("session-1"),
		ProviderPaymentUniqueRef: pgText("payment-unique-1"),
		ProviderDeclineCode:      pgtype.Int4{},
		ProviderDeclineText:      pgtype.Text{},
		ProviderProcessedAt:      pgTime(processedAt),
		ID:                       attempt.ID,
	})
	if err != nil {
		t.Fatalf("UpdateMainPurchaseAttemptOutcome() error = %v, want nil", err)
	}
	if attempt.Status != "succeeded" || attempt.ProviderPurchaseRef.String != "purchase-ref-1" {
		t.Fatalf("UpdateMainPurchaseAttemptOutcome() got %#v", attempt)
	}
	if !attempt.ProviderProcessedAt.Time.Equal(processedAt) {
		t.Fatalf("UpdateMainPurchaseAttemptOutcome() provider_processed_at got %s want %s", attempt.ProviderProcessedAt.Time, processedAt)
	}

	gotAttempt, err := queries.GetMainPurchaseAttemptByID(ctx, attempt.ID)
	if err != nil {
		t.Fatalf("GetMainPurchaseAttemptByID() error = %v, want nil", err)
	}
	if gotAttempt.ID != attempt.ID {
		t.Fatalf("GetMainPurchaseAttemptByID() id got %v want %v", gotAttempt.ID, attempt.ID)
	}

	gotAttempt, err = queries.GetMainPurchaseAttemptByIDForUpdate(ctx, attempt.ID)
	if err != nil {
		t.Fatalf("GetMainPurchaseAttemptByIDForUpdate() error = %v, want nil", err)
	}
	if gotAttempt.ID != attempt.ID {
		t.Fatalf("GetMainPurchaseAttemptByIDForUpdate() id got %v want %v", gotAttempt.ID, attempt.ID)
	}

	gotAttempt, err = queries.GetMainPurchaseAttemptByIdempotencyKeyForUpdate(ctx, "purchase-attempt-1")
	if err != nil {
		t.Fatalf("GetMainPurchaseAttemptByIdempotencyKeyForUpdate() error = %v, want nil", err)
	}
	if gotAttempt.ID != attempt.ID {
		t.Fatalf("GetMainPurchaseAttemptByIdempotencyKeyForUpdate() id got %v want %v", gotAttempt.ID, attempt.ID)
	}

	gotAttempt, err = queries.GetMainPurchaseAttemptByProviderPurchaseRefForUpdate(ctx, pgText("purchase-ref-1"))
	if err != nil {
		t.Fatalf("GetMainPurchaseAttemptByProviderPurchaseRefForUpdate() error = %v, want nil", err)
	}
	if gotAttempt.ID != attempt.ID {
		t.Fatalf("GetMainPurchaseAttemptByProviderPurchaseRefForUpdate() id got %v want %v", gotAttempt.ID, attempt.ID)
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

	migrator := newTestMigrator(t, tempConfig)

	conn, err := pgx.ConnectConfig(ctx, tempConfig)
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

func assertPgConstraintError(t *testing.T, err error, wantCode string, wantConstraint string) {
	t.Helper()

	if err == nil {
		t.Fatalf("error = nil, want pg error code %s constraint %s", wantCode, wantConstraint)
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("error type got %T want *pgconn.PgError", err)
	}
	if pgErr.Code != wantCode {
		t.Fatalf("pg error code got %q want %q", pgErr.Code, wantCode)
	}
	if pgErr.ConstraintName != wantConstraint {
		t.Fatalf("pg constraint got %q want %q", pgErr.ConstraintName, wantConstraint)
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

type recommendationFixture struct {
	ViewerUserID       pgtype.UUID
	CreatorUserID      pgtype.UUID
	OtherCreatorUserID pgtype.UUID
	CanonicalMainID    pgtype.UUID
	AlternateMainID    pgtype.UUID
	ShortID            pgtype.UUID
}

func createRecommendationFixture(t *testing.T, ctx context.Context, queries *sqlc.Queries, now time.Time) recommendationFixture {
	t.Helper()

	viewer, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser(viewer) error = %v, want nil", err)
	}
	creator, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser(creator) error = %v, want nil", err)
	}
	otherCreator, err := queries.CreateUser(ctx)
	if err != nil {
		t.Fatalf("CreateUser(otherCreator) error = %v, want nil", err)
	}

	for _, creatorUserID := range []pgtype.UUID{creator.ID, otherCreator.ID} {
		if _, err := queries.CreateCreatorCapability(ctx, sqlc.CreateCreatorCapabilityParams{
			UserID:                  creatorUserID,
			State:                   "approved",
			IsResubmitEligible:      false,
			IsSupportReviewRequired: false,
			SelfServeResubmitCount:  0,
			ApprovedAt:              pgTime(now),
		}); err != nil {
			t.Fatalf("CreateCreatorCapability(%v) error = %v, want nil", creatorUserID, err)
		}
	}

	mainAsset, err := createReadyMediaAsset(ctx, queries, creator.ID, "recommendation-main")
	if err != nil {
		t.Fatalf("createReadyMediaAsset(main) error = %v, want nil", err)
	}
	alternateMainAsset, err := createReadyMediaAsset(ctx, queries, creator.ID, "recommendation-main-alt")
	if err != nil {
		t.Fatalf("createReadyMediaAsset(alternate main) error = %v, want nil", err)
	}
	shortAsset, err := createReadyMediaAsset(ctx, queries, creator.ID, "recommendation-short")
	if err != nil {
		t.Fatalf("createReadyMediaAsset(short) error = %v, want nil", err)
	}

	canonicalMain, err := queries.CreateMain(ctx, sqlc.CreateMainParams{
		CreatorUserID:       creator.ID,
		MediaAssetID:        mainAsset.ID,
		State:               "approved_for_unlock",
		PriceMinor:          1200,
		CurrencyCode:        "JPY",
		OwnershipConfirmed:  true,
		ConsentConfirmed:    true,
		ApprovedForUnlockAt: pgTime(now.Add(time.Hour)),
	})
	if err != nil {
		t.Fatalf("CreateMain(canonical) error = %v, want nil", err)
	}
	alternateMain, err := queries.CreateMain(ctx, sqlc.CreateMainParams{
		CreatorUserID:       creator.ID,
		MediaAssetID:        alternateMainAsset.ID,
		State:               "approved_for_unlock",
		PriceMinor:          900,
		CurrencyCode:        "JPY",
		OwnershipConfirmed:  true,
		ConsentConfirmed:    true,
		ApprovedForUnlockAt: pgTime(now.Add(2 * time.Hour)),
	})
	if err != nil {
		t.Fatalf("CreateMain(alternate) error = %v, want nil", err)
	}

	short, err := queries.CreateShort(ctx, sqlc.CreateShortParams{
		CreatorUserID:        creator.ID,
		CanonicalMainID:      canonicalMain.ID,
		MediaAssetID:         shortAsset.ID,
		State:                "approved_for_publish",
		ApprovedForPublishAt: pgTime(now.Add(3 * time.Hour)),
		PublishedAt:          pgTime(now.Add(4 * time.Hour)),
	})
	if err != nil {
		t.Fatalf("CreateShort() error = %v, want nil", err)
	}

	return recommendationFixture{
		ViewerUserID:       viewer.ID,
		CreatorUserID:      creator.ID,
		OtherCreatorUserID: otherCreator.ID,
		CanonicalMainID:    canonicalMain.ID,
		AlternateMainID:    alternateMain.ID,
		ShortID:            short.ID,
	}
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
