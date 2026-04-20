package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestAdminSubmissionReviewQueriesUseIntakeSnapshot(t *testing.T) {
	ctx, conn, migrator, cleanup := newIntegrationEnvironment(t)
	defer cleanup()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("migrator.Up() error = %v, want nil", err)
	}
	assertMigrationVersion(t, migrator, latestMigrationVersion)

	queries := sqlc.New(conn)
	now := time.Unix(1710000000, 0).UTC()

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

	currentMainAsset, err := createReadyMediaAsset(ctx, queries, creator.ID, "admin-main-current")
	if err != nil {
		t.Fatalf("createReadyMediaAsset(current main) error = %v, want nil", err)
	}
	intakeMainAsset, err := createReadyMediaAsset(ctx, queries, creator.ID, "admin-main-intake")
	if err != nil {
		t.Fatalf("createReadyMediaAsset(intake main) error = %v, want nil", err)
	}
	currentShortAsset, err := createReadyMediaAsset(ctx, queries, creator.ID, "admin-short-current")
	if err != nil {
		t.Fatalf("createReadyMediaAsset(current short) error = %v, want nil", err)
	}
	intakeShortAsset, err := createReadyMediaAsset(ctx, queries, creator.ID, "admin-short-intake")
	if err != nil {
		t.Fatalf("createReadyMediaAsset(intake short) error = %v, want nil", err)
	}

	var mainID pgtype.UUID
	if err := conn.QueryRow(
		ctx,
		`INSERT INTO app.mains (
			creator_user_id,
			media_asset_id,
			state,
			price_minor,
			currency_code,
			ownership_confirmed,
			consent_confirmed
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		creator.ID,
		currentMainAsset.ID,
		"pending_review",
		int64(2400),
		"JPY",
		true,
		true,
	).Scan(&mainID); err != nil {
		t.Fatalf("insert main error = %v, want nil", err)
	}

	var shortID pgtype.UUID
	if err := conn.QueryRow(
		ctx,
		`INSERT INTO app.shorts (
			creator_user_id,
			canonical_main_id,
			media_asset_id,
			caption,
			state
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		creator.ID,
		mainID,
		currentShortAsset.ID,
		"current caption",
		"pending_review",
	).Scan(&shortID); err != nil {
		t.Fatalf("insert short error = %v, want nil", err)
	}

	intake, err := queries.CreateSubmissionReviewIntake(ctx, sqlc.CreateSubmissionReviewIntakeParams{
		CanonicalMainID:    mainID,
		CreatorUserID:      creator.ID,
		Status:             "pending_review",
		SubmitKind:         "initial_submit",
		PreviousIntakeID:   pgtype.UUID{},
		MainMediaAssetID:   intakeMainAsset.ID,
		MainPriceMinor:     1800,
		OwnershipConfirmed: true,
		ConsentConfirmed:   true,
		SubmittedAt:        pgTime(now),
	})
	if err != nil {
		t.Fatalf("CreateSubmissionReviewIntake() error = %v, want nil", err)
	}

	if err := queries.CreateSubmissionReviewIntakeShort(ctx, sqlc.CreateSubmissionReviewIntakeShortParams{
		SubmissionReviewIntakeID: intake.ID,
		ShortID:                  shortID,
		MediaAssetID:             intakeShortAsset.ID,
		Caption:                  pgText("snapshot caption"),
	}); err != nil {
		t.Fatalf("CreateSubmissionReviewIntakeShort() error = %v, want nil", err)
	}

	mainRow, err := queries.GetAdminSubmissionReviewMainByIntakeID(ctx, intake.ID)
	if err != nil {
		t.Fatalf("GetAdminSubmissionReviewMainByIntakeID() error = %v, want nil", err)
	}
	if mainRow.MediaAssetID != intakeMainAsset.ID {
		t.Fatalf("GetAdminSubmissionReviewMainByIntakeID() media asset got %v want %v", mainRow.MediaAssetID, intakeMainAsset.ID)
	}
	if mainRow.PriceMinor != 1800 {
		t.Fatalf("GetAdminSubmissionReviewMainByIntakeID() price got %d want %d", mainRow.PriceMinor, 1800)
	}

	shortRows, err := queries.ListAdminSubmissionReviewShortsByIntakeID(ctx, intake.ID)
	if err != nil {
		t.Fatalf("ListAdminSubmissionReviewShortsByIntakeID() error = %v, want nil", err)
	}
	if len(shortRows) != 1 {
		t.Fatalf("ListAdminSubmissionReviewShortsByIntakeID() len got %d want 1", len(shortRows))
	}
	if shortRows[0].MediaAssetID != intakeShortAsset.ID {
		t.Fatalf("ListAdminSubmissionReviewShortsByIntakeID() media asset got %v want %v", shortRows[0].MediaAssetID, intakeShortAsset.ID)
	}
	if caption := OptionalTextFromPG(shortRows[0].IntakeCaption); caption == nil || *caption != "snapshot caption" {
		t.Fatalf("ListAdminSubmissionReviewShortsByIntakeID() caption got %v want %q", caption, "snapshot caption")
	}
	if mainRow.MediaAssetID == currentMainAsset.ID {
		t.Fatal("GetAdminSubmissionReviewMainByIntakeID() used current main asset, want intake snapshot asset")
	}
	if shortRows[0].MediaAssetID == currentShortAsset.ID {
		t.Fatal("ListAdminSubmissionReviewShortsByIntakeID() used current short asset, want intake snapshot asset")
	}
}
