package submissionreview

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	capabilityStateApproved = "approved"

	intakeStatusDecisionApplied = "decision_applied"
	intakeStatusPendingReview   = "pending_review"

	mainStateApprovedForUnlock = "approved_for_unlock"
	mainStateDraft             = "draft"
	mainStatePendingReview     = "pending_review"
	mainStateRejected          = "rejected"
	mainStateRevisionRequested = "revision_requested"

	shortStateApprovedForPublish = "approved_for_publish"
	shortStateDraft              = "draft"
	shortStatePendingReview      = "pending_review"
	shortStateRejected           = "rejected"
	shortStateRevisionRequested  = "revision_requested"

	mediaStateReady = "ready"

	submitKindInitial  = "initial_submit"
	submitKindResubmit = "resubmit"
)

// ErrCreatorModeUnavailable は approved creator capability がないため submit flow を使えないことを表します。
var ErrCreatorModeUnavailable = errors.New("creator mode is not available")

// ErrReviewStateConflict は current package state から submit / resubmit できないことを表します。
var ErrReviewStateConflict = errors.New("submission review state conflict")

// ErrSubmissionPackageNotFound は submit 対象の canonical main が存在しないか owner 不一致なことを表します。
var ErrSubmissionPackageNotFound = errors.New("submission package was not found")

type readinessBlocker string

const (
	readinessBlockerLinkedShortMissing readinessBlocker = "linked_short_missing"
	readinessBlockerMainAssetNotReady  readinessBlocker = "main_asset_not_ready"
	readinessBlockerMainPriceMissing   readinessBlocker = "main_price_missing"
	readinessBlockerOwnershipMissing   readinessBlocker = "ownership_missing"
	readinessBlockerConsentMissing     readinessBlocker = "consent_missing"
	readinessBlockerShortAssetNotReady readinessBlocker = "short_asset_not_ready"
)

var readinessBlockerOrder = []readinessBlocker{
	readinessBlockerLinkedShortMissing,
	readinessBlockerMainAssetNotReady,
	readinessBlockerShortAssetNotReady,
	readinessBlockerMainPriceMissing,
	readinessBlockerOwnershipMissing,
	readinessBlockerConsentMissing,
}

// NotReadyError は submission package ready 条件が未成立なことを表します。
type NotReadyError struct {
	Blockers []string
}

func (e *NotReadyError) Error() string {
	if e == nil || len(e.Blockers) == 0 {
		return "submission package is not ready"
	}

	return fmt.Sprintf("submission package is not ready: %s", strings.Join(e.Blockers, ","))
}

type queries interface {
	CreateSubmissionReviewIntake(ctx context.Context, arg sqlc.CreateSubmissionReviewIntakeParams) (sqlc.AppSubmissionReviewIntake, error)
	CreateSubmissionReviewIntakeShort(ctx context.Context, arg sqlc.CreateSubmissionReviewIntakeShortParams) error
	GetCreatorCapabilityByUserIDForUpdate(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error)
	GetLatestSubmissionReviewIntakeByCanonicalMainID(ctx context.Context, canonicalMainID pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error)
	GetPendingSubmissionReviewIntakeByCanonicalMainID(ctx context.Context, canonicalMainID pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error)
	GetSubmissionReviewMainByIDForUpdate(ctx context.Context, id pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error)
	ListSubmissionReviewShortsByCanonicalMainIDForUpdate(ctx context.Context, canonicalMainID pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error)
	UpdateMainState(ctx context.Context, arg sqlc.UpdateMainStateParams) (sqlc.AppMain, error)
	UpdateShortState(ctx context.Context, arg sqlc.UpdateShortStateParams) (sqlc.AppShort, error)
}

type transitionPlan struct {
	PreviousIntakeID pgtype.UUID
	ShortsToPending  []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow
	SubmitKind       string
	UpdateMain       bool
}

// Service は submission package ready 判定と review submit / resubmit を扱います。
type Service struct {
	beginner   postgres.TxBeginner
	now        func() time.Time
	newQueries func(sqlc.DBTX) queries
}

// NewService は pgxpool ベースの submission review service を構築します。
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		beginner: pool,
		now:      time.Now,
		newQueries: func(db sqlc.DBTX) queries {
			return sqlc.New(db)
		},
	}
}

// SubmitPackage は current package を initial submit または resubmit します。
func (s *Service) SubmitPackage(ctx context.Context, viewerUserID uuid.UUID, mainID uuid.UUID) error {
	if s == nil || s.beginner == nil || s.newQueries == nil {
		return fmt.Errorf("submission review service is not initialized")
	}

	return postgres.RunInTx(ctx, s.beginner, func(tx pgx.Tx) error {
		q := s.newQueries(tx)

		capability, err := q.GetCreatorCapabilityByUserIDForUpdate(ctx, postgres.UUIDToPG(viewerUserID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("submission package submit user=%s: %w", viewerUserID, ErrCreatorModeUnavailable)
			}
			return fmt.Errorf("submission package submit capability load user=%s: %w", viewerUserID, err)
		}
		if capability.State != capabilityStateApproved {
			return fmt.Errorf(
				"submission package submit user=%s capability_state=%s: %w",
				viewerUserID,
				capability.State,
				ErrCreatorModeUnavailable,
			)
		}

		mainRow, err := q.GetSubmissionReviewMainByIDForUpdate(ctx, postgres.UUIDToPG(mainID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("submission package submit main=%s user=%s: %w", mainID, viewerUserID, ErrSubmissionPackageNotFound)
			}
			return fmt.Errorf("submission package submit main load main=%s user=%s: %w", mainID, viewerUserID, err)
		}

		mainCreatorUserID, err := postgres.UUIDFromPG(mainRow.CreatorUserID)
		if err != nil {
			return fmt.Errorf("submission package submit main creator parse main=%s: %w", mainID, err)
		}
		if mainCreatorUserID != viewerUserID {
			return fmt.Errorf("submission package submit main=%s user=%s: %w", mainID, viewerUserID, ErrSubmissionPackageNotFound)
		}

		if _, err := q.GetPendingSubmissionReviewIntakeByCanonicalMainID(ctx, mainRow.ID); err == nil {
			return nil
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("submission package submit pending intake load main=%s user=%s: %w", mainID, viewerUserID, err)
		}

		shortRows, err := q.ListSubmissionReviewShortsByCanonicalMainIDForUpdate(ctx, mainRow.ID)
		if err != nil {
			return fmt.Errorf("submission package submit shorts load main=%s user=%s: %w", mainID, viewerUserID, err)
		}

		if blockers := assessReadyBlockers(mainRow, shortRows); len(blockers) > 0 {
			return &NotReadyError{Blockers: blockers}
		}

		latestIntake, latestExists, err := loadLatestIntake(ctx, q, mainRow.ID)
		if err != nil {
			return fmt.Errorf("submission package submit latest intake load main=%s user=%s: %w", mainID, viewerUserID, err)
		}

		transition, err := determineTransition(mainRow, shortRows, latestExists, latestIntake)
		if err != nil {
			return fmt.Errorf("submission package submit state transition main=%s user=%s: %w", mainID, viewerUserID, err)
		}

		submittedAt := s.now().UTC()
		intake, err := q.CreateSubmissionReviewIntake(ctx, sqlc.CreateSubmissionReviewIntakeParams{
			CanonicalMainID:    mainRow.ID,
			CreatorUserID:      mainRow.CreatorUserID,
			Status:             intakeStatusPendingReview,
			SubmitKind:         transition.SubmitKind,
			PreviousIntakeID:   transition.PreviousIntakeID,
			MainMediaAssetID:   mainRow.MediaAssetID,
			MainPriceMinor:     mainRow.PriceMinor,
			OwnershipConfirmed: mainRow.OwnershipConfirmed,
			ConsentConfirmed:   mainRow.ConsentConfirmed,
			SubmittedAt:        postgres.TimeToPG(&submittedAt),
		})
		if err != nil {
			return fmt.Errorf("submission package submit intake create main=%s user=%s: %w", mainID, viewerUserID, err)
		}

		for _, shortRow := range shortRows {
			if err := q.CreateSubmissionReviewIntakeShort(ctx, sqlc.CreateSubmissionReviewIntakeShortParams{
				SubmissionReviewIntakeID: intake.ID,
				ShortID:                  shortRow.ID,
				MediaAssetID:             shortRow.MediaAssetID,
				Caption:                  shortRow.Caption,
			}); err != nil {
				return fmt.Errorf("submission package submit intake short create main=%s short=%s: %w", mainID, shortRow.ID, err)
			}
		}

		if transition.UpdateMain {
			if _, err := q.UpdateMainState(ctx, buildPendingReviewMainUpdate(mainRow)); err != nil {
				return fmt.Errorf("submission package submit main state update main=%s user=%s: %w", mainID, viewerUserID, err)
			}
		}

		for _, shortRow := range transition.ShortsToPending {
			if _, err := q.UpdateShortState(ctx, buildPendingReviewShortUpdate(shortRow)); err != nil {
				return fmt.Errorf("submission package submit short state update main=%s short=%s: %w", mainID, shortRow.ID, err)
			}
		}

		return nil
	})
}

func loadLatestIntake(
	ctx context.Context,
	q queries,
	canonicalMainID pgtype.UUID,
) (sqlc.AppSubmissionReviewIntake, bool, error) {
	intake, err := q.GetLatestSubmissionReviewIntakeByCanonicalMainID(ctx, canonicalMainID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.AppSubmissionReviewIntake{}, false, nil
		}
		return sqlc.AppSubmissionReviewIntake{}, false, err
	}

	return intake, true, nil
}

func assessReadyBlockers(
	mainRow sqlc.GetSubmissionReviewMainByIDForUpdateRow,
	shortRows []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow,
) []string {
	seen := map[readinessBlocker]struct{}{}
	ordered := make([]string, 0, len(readinessBlockerOrder))
	add := func(blocker readinessBlocker) {
		if _, ok := seen[blocker]; ok {
			return
		}
		seen[blocker] = struct{}{}
		ordered = append(ordered, string(blocker))
	}

	if len(shortRows) == 0 {
		add(readinessBlockerLinkedShortMissing)
	}
	if mainRow.MediaProcessingState != mediaStateReady {
		add(readinessBlockerMainAssetNotReady)
	}
	for _, shortRow := range shortRows {
		if shortRow.MediaProcessingState != mediaStateReady {
			add(readinessBlockerShortAssetNotReady)
			break
		}
	}
	if mainRow.PriceMinor <= 0 {
		add(readinessBlockerMainPriceMissing)
	}
	if !mainRow.OwnershipConfirmed {
		add(readinessBlockerOwnershipMissing)
	}
	if !mainRow.ConsentConfirmed {
		add(readinessBlockerConsentMissing)
	}

	slices.SortStableFunc(ordered, func(a string, b string) int {
		return slices.Index(readinessBlockerOrder, readinessBlocker(a)) - slices.Index(readinessBlockerOrder, readinessBlocker(b))
	})

	return ordered
}

func determineTransition(
	mainRow sqlc.GetSubmissionReviewMainByIDForUpdateRow,
	shortRows []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow,
	latestExists bool,
	latestIntake sqlc.AppSubmissionReviewIntake,
) (transitionPlan, error) {
	allShortsDraft := true
	hasRejected := mainRow.State == mainStateRejected
	hasRevisionRequested := mainRow.State == mainStateRevisionRequested
	hasPendingReview := mainRow.State == mainStatePendingReview
	hasUnsupportedState := !isAllowedMainState(mainRow.State)
	shortsToPending := make([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, 0, len(shortRows))

	for _, shortRow := range shortRows {
		switch shortRow.State {
		case shortStateDraft:
			shortsToPending = append(shortsToPending, shortRow)
		case shortStateRevisionRequested:
			hasRevisionRequested = true
			shortsToPending = append(shortsToPending, shortRow)
			allShortsDraft = false
		case shortStatePendingReview:
			hasPendingReview = true
			allShortsDraft = false
		case shortStateRejected:
			hasRejected = true
			allShortsDraft = false
		case shortStateApprovedForPublish:
			allShortsDraft = false
		default:
			hasUnsupportedState = true
			allShortsDraft = false
		}
		if shortRow.State != shortStateDraft {
			allShortsDraft = false
		}
	}

	allDraft := mainRow.State == mainStateDraft && allShortsDraft
	if allDraft {
		return transitionPlan{
			ShortsToPending: shortsToPending,
			SubmitKind:      submitKindInitial,
			UpdateMain:      true,
		}, nil
	}

	if hasPendingReview || hasRejected || hasUnsupportedState {
		return transitionPlan{}, ErrReviewStateConflict
	}
	if !hasRevisionRequested || !latestExists || latestIntake.Status != intakeStatusDecisionApplied {
		return transitionPlan{}, ErrReviewStateConflict
	}

	return transitionPlan{
		PreviousIntakeID: latestIntake.ID,
		ShortsToPending:  shortsToPending,
		SubmitKind:       submitKindResubmit,
		UpdateMain:       mainRow.State == mainStateDraft || mainRow.State == mainStateRevisionRequested,
	}, nil
}

func isAllowedMainState(state string) bool {
	switch state {
	case mainStateDraft, mainStateRevisionRequested, mainStateApprovedForUnlock:
		return true
	default:
		return false
	}
}

func buildPendingReviewMainUpdate(row sqlc.GetSubmissionReviewMainByIDForUpdateRow) sqlc.UpdateMainStateParams {
	return sqlc.UpdateMainStateParams{
		State:               mainStatePendingReview,
		ReviewReasonCode:    pgtype.Text{},
		PostReportState:     row.PostReportState,
		PriceMinor:          row.PriceMinor,
		CurrencyCode:        row.CurrencyCode,
		OwnershipConfirmed:  row.OwnershipConfirmed,
		ConsentConfirmed:    row.ConsentConfirmed,
		ApprovedForUnlockAt: pgtype.Timestamptz{},
		ID:                  row.ID,
	}
}

func buildPendingReviewShortUpdate(row sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow) sqlc.UpdateShortStateParams {
	return sqlc.UpdateShortStateParams{
		State:                shortStatePendingReview,
		ReviewReasonCode:     pgtype.Text{},
		PostReportState:      row.PostReportState,
		ApprovedForPublishAt: pgtype.Timestamptz{},
		PublishedAt:          pgtype.Timestamptz{},
		ID:                   row.ID,
	}
}
