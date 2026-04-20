package submissionreview

import (
	"context"
	"errors"
	"fmt"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	workspaceReviewPackageReadinessBlocked  = "blocked"
	workspaceReviewPackageReadinessConflict = "conflict"
	workspaceReviewPackageReadinessNone     = "none"
	workspaceReviewPackageReadinessReady    = "ready"

	workspaceReviewPackageStatusApproved         = "approved"
	workspaceReviewPackageStatusChangesRequested = "changes_requested"
	workspaceReviewPackageStatusDraft            = "draft"
	workspaceReviewPackageStatusPendingReview    = "pending_review"
	workspaceReviewPackageStatusRejected         = "rejected"

	workspaceReviewSubmitActionNone     = "none"
	workspaceReviewSubmitActionResubmit = "resubmit"
	workspaceReviewSubmitActionSubmit   = "submit"

	workspaceReviewTargetKindMain  = "main"
	workspaceReviewTargetKindShort = "short"
)

const listLatestWorkspaceReviewIntakesByCanonicalMainIDs = `
SELECT DISTINCT ON (canonical_main_id)
  id,
  canonical_main_id,
  creator_user_id,
  status,
  submit_kind,
  previous_intake_id,
  main_media_asset_id,
  main_price_minor,
  ownership_confirmed,
  consent_confirmed,
  submitted_at,
  created_at,
  updated_at
FROM app.submission_review_intakes
WHERE canonical_main_id = ANY($1::uuid[])
ORDER BY canonical_main_id, submitted_at DESC, id DESC
`

const listPendingWorkspaceReviewIntakesByCanonicalMainIDs = `
SELECT
  id,
  canonical_main_id,
  creator_user_id,
  status,
  submit_kind,
  previous_intake_id,
  main_media_asset_id,
  main_price_minor,
  ownership_confirmed,
  consent_confirmed,
  submitted_at,
  created_at,
  updated_at
FROM app.submission_review_intakes
WHERE canonical_main_id = ANY($1::uuid[])
  AND status = 'pending_review'
ORDER BY canonical_main_id, submitted_at DESC, id DESC
`

const listWorkspaceReviewMediaAssetsByIDs = `
SELECT
  id,
  processing_state
FROM app.media_assets
WHERE id = ANY($1::uuid[])
`

// ErrReviewSurfaceNotFound は creator-facing review surface の対象が存在しないことを表します。
var ErrReviewSurfaceNotFound = errors.New("submission review surface was not found")

// WorkspaceReviewSurface は creator workspace dashboard 向け review surface の read model です。
type WorkspaceReviewSurface struct {
	Mains    []WorkspaceReviewMainItem
	Packages []WorkspaceReviewPackageSummary
	Shorts   []WorkspaceReviewShortItem
}

// WorkspaceReviewPackageSummary は canonical main ごとの review / submit 状態を表します。
type WorkspaceReviewPackageSummary struct {
	Blockers         []string
	CanonicalMainID  uuid.UUID
	LinkedShortCount int
	Readiness        string
	ReviewStatus     string
	SubmitActionKind string
}

// WorkspaceReviewMainItem は main tile 用の review state を表します。
type WorkspaceReviewMainItem struct {
	ID    uuid.UUID
	State string
}

// WorkspaceReviewShortItem は short tile 用の review state を表します。
type WorkspaceReviewShortItem struct {
	CanonicalMainID uuid.UUID
	ID              uuid.UUID
	State           string
}

// ItemReviewSurface は main / short detail 用の review surface を表します。
type ItemReviewSurface struct {
	Package WorkspaceReviewPackageSummary
	Review  WorkspaceReviewTargetState
	Target  WorkspaceReviewTarget
}

// WorkspaceReviewTarget は review surface detail の対象 object を表します。
type WorkspaceReviewTarget struct {
	CanonicalMainID uuid.UUID
	ID              uuid.UUID
	Kind            string
}

// WorkspaceReviewTargetState は対象 object の審査状態を表します。
type WorkspaceReviewTargetState struct {
	ReasonCode *string
	State      string
}

func (s *Service) readQueries() queries {
	if s == nil || s.newQueries == nil {
		return nil
	}

	return s.newQueries(s.readDB())
}

func (s *Service) readDB() sqlc.DBTX {
	if s == nil {
		return nil
	}

	beginnerDB, ok := any(s.beginner).(sqlc.DBTX)
	if !ok {
		return nil
	}

	return beginnerDB
}

// GetWorkspaceReviewSurface は creator workspace dashboard 用の review surface を返します。
func (s *Service) GetWorkspaceReviewSurface(ctx context.Context, viewerUserID uuid.UUID) (WorkspaceReviewSurface, error) {
	q := s.readQueries()
	if q == nil {
		return WorkspaceReviewSurface{}, fmt.Errorf("submission review service is not initialized")
	}

	if err := ensureApprovedCreatorCapability(ctx, q, viewerUserID); err != nil {
		return WorkspaceReviewSurface{}, err
	}

	mainRows, err := q.ListMainsByCreatorUserID(ctx, postgres.UUIDToPG(viewerUserID))
	if err != nil {
		return WorkspaceReviewSurface{}, fmt.Errorf("submission review surface mains load user=%s: %w", viewerUserID, err)
	}
	shortRows, err := q.ListShortsByCreatorUserID(ctx, postgres.UUIDToPG(viewerUserID))
	if err != nil {
		return WorkspaceReviewSurface{}, fmt.Errorf("submission review surface shorts load user=%s: %w", viewerUserID, err)
	}

	db := s.readDB()
	assetStateCache := map[uuid.UUID]string{}
	if db != nil {
		assetStateCache, err = loadWorkspaceReviewAssetStatesByRows(ctx, db, mainRows, shortRows)
		if err != nil {
			return WorkspaceReviewSurface{}, fmt.Errorf("submission review surface assets load user=%s: %w", viewerUserID, err)
		}
	}
	var pendingIntakesByMainID map[uuid.UUID]*sqlc.AppSubmissionReviewIntake
	var latestIntakesByMainID map[uuid.UUID]*sqlc.AppSubmissionReviewIntake
	if db != nil {
		pendingIntakesByMainID, latestIntakesByMainID, err = loadWorkspaceReviewIntakeMaps(ctx, db, mainRows)
		if err != nil {
			return WorkspaceReviewSurface{}, fmt.Errorf("submission review surface intakes load user=%s: %w", viewerUserID, err)
		}
	}

	packages := make([]WorkspaceReviewPackageSummary, 0, len(mainRows))
	mains := make([]WorkspaceReviewMainItem, 0, len(mainRows))
	shorts := make([]WorkspaceReviewShortItem, 0, len(shortRows))
	shortsByMainID := groupShortsByCanonicalMain(shortRows)

	for _, shortRow := range shortRows {
		shortID, err := postgres.UUIDFromPG(shortRow.ID)
		if err != nil {
			return WorkspaceReviewSurface{}, fmt.Errorf("submission review surface short id parse user=%s: %w", viewerUserID, err)
		}
		canonicalMainID, err := postgres.UUIDFromPG(shortRow.CanonicalMainID)
		if err != nil {
			continue
		}

		shorts = append(shorts, WorkspaceReviewShortItem{
			CanonicalMainID: canonicalMainID,
			ID:              shortID,
			State:           shortRow.State,
		})
	}

	for _, mainRow := range mainRows {
		mainID, err := postgres.UUIDFromPG(mainRow.ID)
		if err != nil {
			return WorkspaceReviewSurface{}, fmt.Errorf("submission review surface main id parse user=%s: %w", viewerUserID, err)
		}

		var summary WorkspaceReviewPackageSummary
		if db != nil {
			summary, err = buildWorkspaceReviewPackageSummaryWithLoadedIntakes(
				ctx,
				q,
				mainRow,
				shortsByMainID[mainID],
				assetStateCache,
				pendingIntakesByMainID[mainID],
				latestIntakesByMainID[mainID],
			)
		} else {
			summary, err = buildWorkspaceReviewPackageSummary(ctx, q, mainRow, shortsByMainID[mainID], assetStateCache)
		}
		if err != nil {
			return WorkspaceReviewSurface{}, fmt.Errorf("submission review surface package build user=%s main=%s: %w", viewerUserID, mainID, err)
		}

		mains = append(mains, WorkspaceReviewMainItem{
			ID:    mainID,
			State: mainRow.State,
		})
		packages = append(packages, summary)
	}

	return WorkspaceReviewSurface{
		Mains:    mains,
		Packages: packages,
		Shorts:   shorts,
	}, nil
}

// GetMainReviewSurface は current owner の main detail review surface を返します。
func (s *Service) GetMainReviewSurface(ctx context.Context, viewerUserID uuid.UUID, mainID uuid.UUID) (ItemReviewSurface, error) {
	q := s.readQueries()
	if q == nil {
		return ItemReviewSurface{}, fmt.Errorf("submission review service is not initialized")
	}

	if err := ensureApprovedCreatorCapability(ctx, q, viewerUserID); err != nil {
		return ItemReviewSurface{}, err
	}

	mainRow, err := q.GetMainByID(ctx, postgres.UUIDToPG(mainID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ItemReviewSurface{}, ErrReviewSurfaceNotFound
		}
		return ItemReviewSurface{}, fmt.Errorf("submission review surface main load user=%s main=%s: %w", viewerUserID, mainID, err)
	}
	if err := ensureMainOwner(mainRow, viewerUserID); err != nil {
		return ItemReviewSurface{}, err
	}

	shortRows, err := q.ListShortsByCanonicalMainID(ctx, mainRow.ID)
	if err != nil {
		return ItemReviewSurface{}, fmt.Errorf("submission review surface shorts load user=%s main=%s: %w", viewerUserID, mainID, err)
	}

	assetStateCache := map[uuid.UUID]string{}
	if db := s.readDB(); db != nil {
		assetStateCache, err = loadWorkspaceReviewAssetStatesByRows(ctx, db, []sqlc.AppMain{mainRow}, shortRows)
		if err != nil {
			return ItemReviewSurface{}, fmt.Errorf("submission review surface assets load user=%s main=%s: %w", viewerUserID, mainID, err)
		}
	}
	packageShorts := groupShortsByCanonicalMain(shortRows)[mainID]
	summary, err := buildWorkspaceReviewPackageSummary(ctx, q, mainRow, packageShorts, assetStateCache)
	if err != nil {
		return ItemReviewSurface{}, fmt.Errorf("submission review surface package build user=%s main=%s: %w", viewerUserID, mainID, err)
	}

	return ItemReviewSurface{
		Package: summary,
		Review: WorkspaceReviewTargetState{
			ReasonCode: postgres.OptionalTextFromPG(mainRow.ReviewReasonCode),
			State:      mainRow.State,
		},
		Target: WorkspaceReviewTarget{
			CanonicalMainID: mainID,
			ID:              mainID,
			Kind:            workspaceReviewTargetKindMain,
		},
	}, nil
}

// GetShortReviewSurface は current owner の short detail review surface を返します。
func (s *Service) GetShortReviewSurface(ctx context.Context, viewerUserID uuid.UUID, shortID uuid.UUID) (ItemReviewSurface, error) {
	q := s.readQueries()
	if q == nil {
		return ItemReviewSurface{}, fmt.Errorf("submission review service is not initialized")
	}

	if err := ensureApprovedCreatorCapability(ctx, q, viewerUserID); err != nil {
		return ItemReviewSurface{}, err
	}

	shortRow, err := q.GetShortByID(ctx, postgres.UUIDToPG(shortID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ItemReviewSurface{}, ErrReviewSurfaceNotFound
		}
		return ItemReviewSurface{}, fmt.Errorf("submission review surface short load user=%s short=%s: %w", viewerUserID, shortID, err)
	}
	if err := ensureShortOwner(shortRow, viewerUserID); err != nil {
		return ItemReviewSurface{}, err
	}

	canonicalMainID, err := postgres.UUIDFromPG(shortRow.CanonicalMainID)
	if err != nil {
		return ItemReviewSurface{}, ErrReviewSurfaceNotFound
	}

	mainRow, err := q.GetMainByID(ctx, shortRow.CanonicalMainID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ItemReviewSurface{}, ErrReviewSurfaceNotFound
		}
		return ItemReviewSurface{}, fmt.Errorf("submission review surface main load user=%s short=%s main=%s: %w", viewerUserID, shortID, canonicalMainID, err)
	}
	if err := ensureMainOwner(mainRow, viewerUserID); err != nil {
		return ItemReviewSurface{}, err
	}

	shortRows, err := q.ListShortsByCanonicalMainID(ctx, shortRow.CanonicalMainID)
	if err != nil {
		return ItemReviewSurface{}, fmt.Errorf("submission review surface shorts load user=%s short=%s: %w", viewerUserID, shortID, err)
	}

	assetStateCache := map[uuid.UUID]string{}
	if db := s.readDB(); db != nil {
		assetStateCache, err = loadWorkspaceReviewAssetStatesByRows(ctx, db, []sqlc.AppMain{mainRow}, shortRows)
		if err != nil {
			return ItemReviewSurface{}, fmt.Errorf("submission review surface assets load user=%s short=%s: %w", viewerUserID, shortID, err)
		}
	}
	packageShorts := groupShortsByCanonicalMain(shortRows)[canonicalMainID]
	summary, err := buildWorkspaceReviewPackageSummary(ctx, q, mainRow, packageShorts, assetStateCache)
	if err != nil {
		return ItemReviewSurface{}, fmt.Errorf("submission review surface package build user=%s short=%s: %w", viewerUserID, shortID, err)
	}

	return ItemReviewSurface{
		Package: summary,
		Review: WorkspaceReviewTargetState{
			ReasonCode: postgres.OptionalTextFromPG(shortRow.ReviewReasonCode),
			State:      shortRow.State,
		},
		Target: WorkspaceReviewTarget{
			CanonicalMainID: canonicalMainID,
			ID:              shortID,
			Kind:            workspaceReviewTargetKindShort,
		},
	}, nil
}

func ensureApprovedCreatorCapability(ctx context.Context, q queries, viewerUserID uuid.UUID) error {
	capability, err := q.GetCreatorCapabilityByUserID(ctx, postgres.UUIDToPG(viewerUserID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCreatorModeUnavailable
		}
		return fmt.Errorf("submission review surface capability load user=%s: %w", viewerUserID, err)
	}
	if capability.State != capabilityStateApproved {
		return ErrCreatorModeUnavailable
	}

	return nil
}

func ensureMainOwner(mainRow sqlc.AppMain, viewerUserID uuid.UUID) error {
	mainCreatorUserID, err := postgres.UUIDFromPG(mainRow.CreatorUserID)
	if err != nil {
		return fmt.Errorf("submission review surface main creator parse user=%s: %w", viewerUserID, err)
	}
	if mainCreatorUserID != viewerUserID {
		return ErrReviewSurfaceNotFound
	}

	return nil
}

func ensureShortOwner(shortRow sqlc.AppShort, viewerUserID uuid.UUID) error {
	shortCreatorUserID, err := postgres.UUIDFromPG(shortRow.CreatorUserID)
	if err != nil {
		return fmt.Errorf("submission review surface short creator parse user=%s: %w", viewerUserID, err)
	}
	if shortCreatorUserID != viewerUserID {
		return ErrReviewSurfaceNotFound
	}

	return nil
}

func groupShortsByCanonicalMain(shortRows []sqlc.AppShort) map[uuid.UUID][]sqlc.AppShort {
	grouped := make(map[uuid.UUID][]sqlc.AppShort, len(shortRows))
	for _, shortRow := range shortRows {
		canonicalMainID, err := postgres.UUIDFromPG(shortRow.CanonicalMainID)
		if err != nil {
			continue
		}
		grouped[canonicalMainID] = append(grouped[canonicalMainID], shortRow)
	}

	return grouped
}

func buildWorkspaceReviewPackageSummary(
	ctx context.Context,
	q queries,
	mainRow sqlc.AppMain,
	shortRows []sqlc.AppShort,
	assetStateCache map[uuid.UUID]string,
) (WorkspaceReviewPackageSummary, error) {
	pendingIntake, pendingExists, err := loadPendingIntake(ctx, q, mainRow.ID)
	if err != nil {
		return WorkspaceReviewPackageSummary{}, err
	}
	latestIntake, latestExists, err := loadLatestIntake(ctx, q, mainRow.ID)
	if err != nil {
		return WorkspaceReviewPackageSummary{}, err
	}

	return buildWorkspaceReviewPackageSummaryWithLoadedIntakes(
		ctx,
		q,
		mainRow,
		shortRows,
		assetStateCache,
		intakePointer(pendingExists, pendingIntake),
		intakePointer(latestExists, latestIntake),
	)
}

func buildWorkspaceReviewPackageSummaryWithLoadedIntakes(
	ctx context.Context,
	q queries,
	mainRow sqlc.AppMain,
	shortRows []sqlc.AppShort,
	assetStateCache map[uuid.UUID]string,
	pendingIntake *sqlc.AppSubmissionReviewIntake,
	latestIntake *sqlc.AppSubmissionReviewIntake,
) (WorkspaceReviewPackageSummary, error) {
	mainID, err := postgres.UUIDFromPG(mainRow.ID)
	if err != nil {
		return WorkspaceReviewPackageSummary{}, fmt.Errorf("submission review surface main id parse: %w", err)
	}

	mainTransitionRow, err := buildWorkspaceReviewMainTransitionRow(ctx, q, mainRow, assetStateCache)
	if err != nil {
		return WorkspaceReviewPackageSummary{}, err
	}
	shortTransitionRows, err := buildWorkspaceReviewShortTransitionRows(ctx, q, shortRows, assetStateCache)
	if err != nil {
		return WorkspaceReviewPackageSummary{}, err
	}

	pendingExists := pendingIntake != nil
	latestExists := latestIntake != nil
	resolvedPendingIntake := dereferenceWorkspaceReviewIntake(pendingIntake)
	resolvedLatestIntake := dereferenceWorkspaceReviewIntake(latestIntake)

	reviewStatus := resolveWorkspaceReviewPackageStatus(mainRow.State, shortRows, pendingExists)
	blockers := assessReadyBlockers(mainTransitionRow, shortTransitionRows)
	readiness, submitActionKind, err := resolveWorkspaceReviewPackageAction(mainTransitionRow, shortTransitionRows, blockers, pendingExists, resolvedPendingIntake, latestExists, resolvedLatestIntake)
	if err != nil {
		return WorkspaceReviewPackageSummary{}, err
	}

	return WorkspaceReviewPackageSummary{
		Blockers:         blockers,
		CanonicalMainID:  mainID,
		LinkedShortCount: len(shortRows),
		Readiness:        readiness,
		ReviewStatus:     reviewStatus,
		SubmitActionKind: submitActionKind,
	}, nil
}

func buildWorkspaceReviewMainTransitionRow(
	ctx context.Context,
	q queries,
	mainRow sqlc.AppMain,
	assetStateCache map[uuid.UUID]string,
) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
	mediaProcessingState, err := loadWorkspaceReviewMediaProcessingState(ctx, q, mainRow.MediaAssetID, assetStateCache)
	if err != nil {
		return sqlc.GetSubmissionReviewMainByIDForUpdateRow{}, err
	}

	return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
		ID:                   mainRow.ID,
		CreatorUserID:        mainRow.CreatorUserID,
		MediaAssetID:         mainRow.MediaAssetID,
		State:                mainRow.State,
		ReviewReasonCode:     mainRow.ReviewReasonCode,
		PostReportState:      mainRow.PostReportState,
		PriceMinor:           mainRow.PriceMinor,
		CurrencyCode:         mainRow.CurrencyCode,
		OwnershipConfirmed:   mainRow.OwnershipConfirmed,
		ConsentConfirmed:     mainRow.ConsentConfirmed,
		ApprovedForUnlockAt:  mainRow.ApprovedForUnlockAt,
		CreatedAt:            mainRow.CreatedAt,
		UpdatedAt:            mainRow.UpdatedAt,
		MediaProcessingState: mediaProcessingState,
	}, nil
}

func buildWorkspaceReviewShortTransitionRows(
	ctx context.Context,
	q queries,
	shortRows []sqlc.AppShort,
	assetStateCache map[uuid.UUID]string,
) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
	rows := make([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, 0, len(shortRows))
	for _, shortRow := range shortRows {
		mediaProcessingState, err := loadWorkspaceReviewMediaProcessingState(ctx, q, shortRow.MediaAssetID, assetStateCache)
		if err != nil {
			return nil, err
		}

		rows = append(rows, sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{
			ID:                   shortRow.ID,
			CreatorUserID:        shortRow.CreatorUserID,
			CanonicalMainID:      shortRow.CanonicalMainID,
			MediaAssetID:         shortRow.MediaAssetID,
			State:                shortRow.State,
			ReviewReasonCode:     shortRow.ReviewReasonCode,
			PostReportState:      shortRow.PostReportState,
			ApprovedForPublishAt: shortRow.ApprovedForPublishAt,
			PublishedAt:          shortRow.PublishedAt,
			CreatedAt:            shortRow.CreatedAt,
			UpdatedAt:            shortRow.UpdatedAt,
			Caption:              shortRow.Caption,
			MediaProcessingState: mediaProcessingState,
		})
	}

	return rows, nil
}

func loadWorkspaceReviewAssetStatesByRows(
	ctx context.Context,
	db sqlc.DBTX,
	mainRows []sqlc.AppMain,
	shortRows []sqlc.AppShort,
) (map[uuid.UUID]string, error) {
	if db == nil {
		return map[uuid.UUID]string{}, nil
	}

	assetIDs := make([]uuid.UUID, 0, len(mainRows)+len(shortRows))
	seenAssetIDs := make(map[uuid.UUID]struct{}, len(mainRows)+len(shortRows))
	for _, mainRow := range mainRows {
		assetID, err := postgres.UUIDFromPG(mainRow.MediaAssetID)
		if err != nil {
			return nil, fmt.Errorf("submission review surface main media asset id parse: %w", err)
		}
		if _, exists := seenAssetIDs[assetID]; exists {
			continue
		}

		seenAssetIDs[assetID] = struct{}{}
		assetIDs = append(assetIDs, assetID)
	}
	for _, shortRow := range shortRows {
		assetID, err := postgres.UUIDFromPG(shortRow.MediaAssetID)
		if err != nil {
			return nil, fmt.Errorf("submission review surface short media asset id parse: %w", err)
		}
		if _, exists := seenAssetIDs[assetID]; exists {
			continue
		}

		seenAssetIDs[assetID] = struct{}{}
		assetIDs = append(assetIDs, assetID)
	}

	return loadWorkspaceReviewAssetStatesByAssetIDs(ctx, db, assetIDs)
}

func loadWorkspaceReviewAssetStatesByAssetIDs(
	ctx context.Context,
	db sqlc.DBTX,
	assetIDs []uuid.UUID,
) (map[uuid.UUID]string, error) {
	if db == nil || len(assetIDs) == 0 {
		return map[uuid.UUID]string{}, nil
	}

	rows, err := db.Query(ctx, listWorkspaceReviewMediaAssetsByIDs, assetIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assetStateCache := make(map[uuid.UUID]string, len(assetIDs))
	for rows.Next() {
		assetID, processingState, err := scanWorkspaceReviewMediaAssetState(rows)
		if err != nil {
			return nil, err
		}

		assetStateCache[assetID] = processingState
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return assetStateCache, nil
}

func loadWorkspaceReviewIntakeMaps(
	ctx context.Context,
	db sqlc.DBTX,
	mainRows []sqlc.AppMain,
) (map[uuid.UUID]*sqlc.AppSubmissionReviewIntake, map[uuid.UUID]*sqlc.AppSubmissionReviewIntake, error) {
	if db == nil || len(mainRows) == 0 {
		return map[uuid.UUID]*sqlc.AppSubmissionReviewIntake{}, map[uuid.UUID]*sqlc.AppSubmissionReviewIntake{}, nil
	}

	mainIDs := make([]uuid.UUID, 0, len(mainRows))
	for _, row := range mainRows {
		mainID, err := postgres.UUIDFromPG(row.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("submission review surface main id parse: %w", err)
		}
		mainIDs = append(mainIDs, mainID)
	}

	latestByMainID, err := loadWorkspaceReviewIntakeMap(ctx, db, listLatestWorkspaceReviewIntakesByCanonicalMainIDs, mainIDs)
	if err != nil {
		return nil, nil, err
	}
	pendingByMainID, err := loadWorkspaceReviewIntakeMap(ctx, db, listPendingWorkspaceReviewIntakesByCanonicalMainIDs, mainIDs)
	if err != nil {
		return nil, nil, err
	}

	return pendingByMainID, latestByMainID, nil
}

func loadWorkspaceReviewIntakeMap(
	ctx context.Context,
	db sqlc.DBTX,
	query string,
	mainIDs []uuid.UUID,
) (map[uuid.UUID]*sqlc.AppSubmissionReviewIntake, error) {
	if len(mainIDs) == 0 {
		return map[uuid.UUID]*sqlc.AppSubmissionReviewIntake{}, nil
	}

	rows, err := db.Query(ctx, query, mainIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	intakesByMainID := make(map[uuid.UUID]*sqlc.AppSubmissionReviewIntake, len(mainIDs))
	for rows.Next() {
		intake, err := scanWorkspaceReviewIntake(rows)
		if err != nil {
			return nil, err
		}

		canonicalMainID, err := postgres.UUIDFromPG(intake.CanonicalMainID)
		if err != nil {
			return nil, fmt.Errorf("submission review surface intake canonical main id parse: %w", err)
		}

		intakeCopy := intake
		intakesByMainID[canonicalMainID] = &intakeCopy
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return intakesByMainID, nil
}

func scanWorkspaceReviewIntake(row pgx.Row) (sqlc.AppSubmissionReviewIntake, error) {
	var intake sqlc.AppSubmissionReviewIntake
	err := row.Scan(
		&intake.ID,
		&intake.CanonicalMainID,
		&intake.CreatorUserID,
		&intake.Status,
		&intake.SubmitKind,
		&intake.PreviousIntakeID,
		&intake.MainMediaAssetID,
		&intake.MainPriceMinor,
		&intake.OwnershipConfirmed,
		&intake.ConsentConfirmed,
		&intake.SubmittedAt,
		&intake.CreatedAt,
		&intake.UpdatedAt,
	)
	if err != nil {
		return sqlc.AppSubmissionReviewIntake{}, err
	}

	return intake, nil
}

func scanWorkspaceReviewMediaAssetState(row pgx.Row) (uuid.UUID, string, error) {
	var assetID pgtype.UUID
	var processingState string
	if err := row.Scan(&assetID, &processingState); err != nil {
		return uuid.Nil, "", err
	}

	parsedAssetID, err := postgres.UUIDFromPG(assetID)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("submission review surface media asset id parse: %w", err)
	}

	return parsedAssetID, processingState, nil
}

func intakePointer(exists bool, intake sqlc.AppSubmissionReviewIntake) *sqlc.AppSubmissionReviewIntake {
	if !exists {
		return nil
	}

	intakeCopy := intake
	return &intakeCopy
}

func dereferenceWorkspaceReviewIntake(intake *sqlc.AppSubmissionReviewIntake) sqlc.AppSubmissionReviewIntake {
	if intake == nil {
		return sqlc.AppSubmissionReviewIntake{}
	}

	return *intake
}

func loadWorkspaceReviewMediaProcessingState(
	ctx context.Context,
	q queries,
	mediaAssetID pgtype.UUID,
	assetStateCache map[uuid.UUID]string,
) (string, error) {
	id, err := postgres.UUIDFromPG(mediaAssetID)
	if err != nil {
		return "", fmt.Errorf("submission review surface media asset id parse: %w", err)
	}
	if cached, ok := assetStateCache[id]; ok {
		return cached, nil
	}

	asset, err := q.GetMediaAssetByID(ctx, mediaAssetID)
	if err != nil {
		return "", fmt.Errorf("submission review surface media asset load asset=%s: %w", id, err)
	}

	assetStateCache[id] = asset.ProcessingState
	return asset.ProcessingState, nil
}

func loadPendingIntake(
	ctx context.Context,
	q queries,
	canonicalMainID pgtype.UUID,
) (sqlc.AppSubmissionReviewIntake, bool, error) {
	intake, err := q.GetPendingSubmissionReviewIntakeByCanonicalMainID(ctx, canonicalMainID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.AppSubmissionReviewIntake{}, false, nil
		}
		return sqlc.AppSubmissionReviewIntake{}, false, err
	}

	return intake, true, nil
}

func resolveWorkspaceReviewPackageStatus(
	mainState string,
	shortRows []sqlc.AppShort,
	pendingExists bool,
) string {
	if pendingExists {
		return workspaceReviewPackageStatusPendingReview
	}

	hasRejected := mainState == mainStateRejected
	hasRevisionRequested := mainState == mainStateRevisionRequested
	allApproved := mainState == mainStateApprovedForUnlock

	for _, shortRow := range shortRows {
		switch shortRow.State {
		case shortStateRejected:
			hasRejected = true
			allApproved = false
		case shortStateRevisionRequested:
			hasRevisionRequested = true
			allApproved = false
		case shortStateApprovedForPublish:
			// keep
		default:
			allApproved = false
		}
	}

	switch {
	case hasRejected:
		return workspaceReviewPackageStatusRejected
	case hasRevisionRequested:
		return workspaceReviewPackageStatusChangesRequested
	case allApproved:
		return workspaceReviewPackageStatusApproved
	default:
		return workspaceReviewPackageStatusDraft
	}
}

func resolveWorkspaceReviewPackageAction(
	mainRow sqlc.GetSubmissionReviewMainByIDForUpdateRow,
	shortRows []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow,
	blockers []string,
	pendingExists bool,
	_ sqlc.AppSubmissionReviewIntake,
	latestExists bool,
	latestIntake sqlc.AppSubmissionReviewIntake,
) (string, string, error) {
	if pendingExists {
		return workspaceReviewPackageReadinessNone, workspaceReviewSubmitActionNone, nil
	}
	if len(blockers) > 0 {
		return workspaceReviewPackageReadinessBlocked, workspaceReviewSubmitActionNone, nil
	}

	transition, err := determineTransition(mainRow, shortRows, latestExists, latestIntake)
	if err != nil {
		if errors.Is(err, ErrReviewStateConflict) {
			return workspaceReviewPackageReadinessConflict, workspaceReviewSubmitActionNone, nil
		}
		return "", "", err
	}

	submitActionKind := workspaceReviewSubmitActionSubmit
	if transition.SubmitKind == submitKindResubmit {
		submitActionKind = workspaceReviewSubmitActionResubmit
	}

	return workspaceReviewPackageReadinessReady, submitActionKind, nil
}
