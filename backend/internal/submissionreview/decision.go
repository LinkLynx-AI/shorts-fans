package submissionreview

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	reviewDecisionApproved          = "approved"
	reviewDecisionRejected          = "rejected"
	reviewDecisionRevisionRequested = "revision_requested"

	reviewDecisionSourceAuto           = "auto"
	reviewDecisionSourceManual         = "manual"
	reviewDecisionSourceManualOverride = "manual_override"
)

var (
	// ErrSubmissionReviewIntakeNotFound は decision apply 対象の pending intake が見つからないことを表します。
	ErrSubmissionReviewIntakeNotFound = errors.New("submission review intake was not found")
	// ErrInvalidReviewDecision は未知の review decision が指定されたことを表します。
	ErrInvalidReviewDecision = errors.New("submission review decision is invalid")
	// ErrInvalidReviewDecisionSource は未知の review decision source が指定されたことを表します。
	ErrInvalidReviewDecisionSource = errors.New("submission review decision source is invalid")
	// ErrReviewDecisionReasonRequired は revision/reject decision に reason code が不足していることを表します。
	ErrReviewDecisionReasonRequired = errors.New("submission review decision reason code is required")
	// ErrReviewDecisionMetadataConflict は decision metadata の組み合わせが矛盾していることを表します。
	ErrReviewDecisionMetadataConflict = errors.New("submission review decision metadata is invalid")
	// ErrSubmissionReviewDecisionTargetsMismatch は pending intake と decision target の対応が崩れていることを表します。
	ErrSubmissionReviewDecisionTargetsMismatch = errors.New("submission review decision targets mismatch")
)

// ReviewDecisionInput は submission package intake へ適用する review decision 入力です。
type ReviewDecisionInput struct {
	IntakeID       uuid.UUID
	DecisionSource string
	MainDecision   *MainReviewDecisionInput
	ShortDecisions []ShortReviewDecisionInput
}

// MainReviewDecisionInput は canonical main に適用する review decision です。
type MainReviewDecisionInput struct {
	Decision   string
	ReasonCode *string
}

// ShortReviewDecisionInput は intake 内 short へ適用する review decision です。
type ShortReviewDecisionInput struct {
	ShortID    uuid.UUID
	Decision   string
	ReasonCode *string
}

type normalizedDecision struct {
	reasonCode  *string
	targetState string
}

// ApplyDecision は pending submission review intake に object-level decision を反映します。
func (s *Service) ApplyDecision(ctx context.Context, input ReviewDecisionInput) error {
	if s == nil || s.beginner == nil || s.newQueries == nil {
		return fmt.Errorf("submission review service is not initialized")
	}
	if input.IntakeID == uuid.Nil {
		return ErrSubmissionReviewIntakeNotFound
	}

	source, err := normalizeReviewDecisionSource(input.DecisionSource)
	if err != nil {
		return err
	}

	mainDecision, err := normalizeMainReviewDecision(input.MainDecision)
	if err != nil {
		return err
	}
	shortDecisions, err := normalizeShortReviewDecisions(input.ShortDecisions)
	if err != nil {
		return err
	}

	now := s.now
	if now == nil {
		now = time.Now
	}

	return postgres.RunInTx(ctx, s.beginner, func(tx pgx.Tx) error {
		q := s.newQueries(tx)

		intake, err := q.GetPendingSubmissionReviewIntakeByIDForUpdate(ctx, postgres.UUIDToPG(input.IntakeID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrSubmissionReviewIntakeNotFound
			}
			return fmt.Errorf("submission review decision intake load intake=%s: %w", input.IntakeID, err)
		}

		mainRow, err := q.GetSubmissionReviewMainByIDForUpdate(ctx, intake.CanonicalMainID)
		if err != nil {
			return fmt.Errorf("submission review decision main load intake=%s: %w", input.IntakeID, err)
		}
		shortRows, err := q.ListSubmissionReviewShortsByCanonicalMainIDForUpdate(ctx, intake.CanonicalMainID)
		if err != nil {
			return fmt.Errorf("submission review decision short load intake=%s: %w", input.IntakeID, err)
		}
		intakeShorts, err := q.ListSubmissionReviewIntakeShortsByIntakeID(ctx, intake.ID)
		if err != nil {
			return fmt.Errorf("submission review decision intake short load intake=%s: %w", input.IntakeID, err)
		}

		currentShortsByID := make(map[uuid.UUID]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, len(shortRows))
		for _, row := range shortRows {
			shortID, parseErr := postgres.UUIDFromPG(row.ID)
			if parseErr != nil {
				return fmt.Errorf("submission review decision short id parse intake=%s: %w", input.IntakeID, parseErr)
			}
			currentShortsByID[shortID] = row
		}

		if !matchesIntakeMainSnapshot(mainRow, intake) {
			return ErrSubmissionReviewDecisionTargetsMismatch
		}
		if !isAllowedCurrentMainState(mainRow.State) {
			return ErrReviewStateConflict
		}

		intakeShortsByID := make(map[uuid.UUID]sqlc.AppSubmissionReviewIntakeShort, len(intakeShorts))
		requiredShortDecisionIDs := make(map[uuid.UUID]struct{})
		for _, intakeShort := range intakeShorts {
			shortID, parseErr := postgres.UUIDFromPG(intakeShort.ShortID)
			if parseErr != nil {
				return fmt.Errorf("submission review decision intake short id parse intake=%s: %w", input.IntakeID, parseErr)
			}
			intakeShortsByID[shortID] = intakeShort

			row, ok := currentShortsByID[shortID]
			if !ok {
				return ErrSubmissionReviewDecisionTargetsMismatch
			}
			if !matchesIntakeShortSnapshot(row, intakeShort) {
				return ErrSubmissionReviewDecisionTargetsMismatch
			}
			if !isAllowedCurrentShortState(row.State) {
				return ErrReviewStateConflict
			}
			if row.State == shortStatePendingReview {
				requiredShortDecisionIDs[shortID] = struct{}{}
			}
		}
		for shortID, row := range currentShortsByID {
			if _, ok := intakeShortsByID[shortID]; ok {
				continue
			}
			if row.State != shortStateDraft {
				return ErrReviewStateConflict
			}
		}

		if mainRow.State == mainStatePendingReview && mainDecision == nil {
			return ErrSubmissionReviewDecisionTargetsMismatch
		}
		if mainRow.State != mainStatePendingReview && mainDecision != nil {
			return ErrReviewStateConflict
		}

		if err := validateRequiredShortDecisions(requiredShortDecisionIDs, shortDecisions); err != nil {
			return err
		}
		for shortID := range shortDecisions {
			row, ok := currentShortsByID[shortID]
			if !ok {
				return ErrSubmissionReviewDecisionTargetsMismatch
			}
			if row.State != shortStatePendingReview {
				return ErrReviewStateConflict
			}
		}

		decisionedAt := now().UTC()

		if mainDecision != nil {
			if err := q.CreateSubmissionReviewMainDecision(ctx, sqlc.CreateSubmissionReviewMainDecisionParams{
				SubmissionReviewIntakeID: intake.ID,
				MainID:                   mainRow.ID,
				TargetState:              mainDecision.targetState,
				ReasonCode:               postgres.TextToPG(mainDecision.reasonCode),
				DecisionSource:           source,
				DecisionedAt:             postgres.TimeToPG(&decisionedAt),
			}); err != nil {
				return fmt.Errorf("submission review decision main log create intake=%s: %w", input.IntakeID, err)
			}
			if _, err := q.ApplySubmissionReviewMainDecision(ctx, buildMainDecisionUpdate(mainRow.ID, *mainDecision, source, decisionedAt)); err != nil {
				return fmt.Errorf("submission review decision main apply intake=%s: %w", input.IntakeID, err)
			}
		}

		for shortID, decision := range shortDecisions {
			row := currentShortsByID[shortID]
			if err := q.CreateSubmissionReviewShortDecision(ctx, sqlc.CreateSubmissionReviewShortDecisionParams{
				SubmissionReviewIntakeID: intake.ID,
				ShortID:                  row.ID,
				TargetState:              decision.targetState,
				ReasonCode:               postgres.TextToPG(decision.reasonCode),
				DecisionSource:           source,
				DecisionedAt:             postgres.TimeToPG(&decisionedAt),
			}); err != nil {
				return fmt.Errorf("submission review decision short log create intake=%s short=%s: %w", input.IntakeID, shortID, err)
			}
			if _, err := q.ApplySubmissionReviewShortDecision(ctx, buildShortDecisionUpdate(row.ID, decision, source, decisionedAt)); err != nil {
				return fmt.Errorf("submission review decision short apply intake=%s short=%s: %w", input.IntakeID, shortID, err)
			}
		}

		if mainDecision == nil && len(shortDecisions) == 0 {
			return ErrSubmissionReviewDecisionTargetsMismatch
		}
		if _, err := q.MarkSubmissionReviewIntakeDecisionApplied(ctx, intake.ID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrReviewStateConflict
			}
			return fmt.Errorf("submission review decision intake close intake=%s: %w", input.IntakeID, err)
		}

		return nil
	})
}

func normalizeReviewDecisionSource(raw string) (string, error) {
	source := strings.TrimSpace(raw)
	if source == "" {
		source = reviewDecisionSourceManual
	}

	switch source {
	case reviewDecisionSourceManual, reviewDecisionSourceAuto, reviewDecisionSourceManualOverride:
		return source, nil
	default:
		return "", ErrInvalidReviewDecisionSource
	}
}

func normalizeMainReviewDecision(input *MainReviewDecisionInput) (*normalizedDecision, error) {
	if input == nil {
		return nil, nil
	}

	decision, err := normalizeDecision(input.Decision, mainStateApprovedForUnlock, input.ReasonCode)
	if err != nil {
		return nil, err
	}

	return &decision, nil
}

func normalizeShortReviewDecisions(inputs []ShortReviewDecisionInput) (map[uuid.UUID]normalizedDecision, error) {
	normalized := make(map[uuid.UUID]normalizedDecision, len(inputs))

	for _, input := range inputs {
		if input.ShortID == uuid.Nil {
			return nil, ErrSubmissionReviewDecisionTargetsMismatch
		}
		if _, exists := normalized[input.ShortID]; exists {
			return nil, ErrSubmissionReviewDecisionTargetsMismatch
		}

		decision, err := normalizeDecision(input.Decision, shortStateApprovedForPublish, input.ReasonCode)
		if err != nil {
			return nil, err
		}
		normalized[input.ShortID] = decision
	}

	return normalized, nil
}

func normalizeDecision(raw string, approvedState string, reasonCode *string) (normalizedDecision, error) {
	decision := strings.TrimSpace(raw)
	trimmedReason := trimmedOptionalString(reasonCode)

	switch decision {
	case reviewDecisionApproved:
		if trimmedReason != nil {
			return normalizedDecision{}, ErrReviewDecisionMetadataConflict
		}
		return normalizedDecision{
			reasonCode:  nil,
			targetState: approvedState,
		}, nil
	case reviewDecisionRevisionRequested, reviewDecisionRejected:
		if trimmedReason == nil {
			return normalizedDecision{}, ErrReviewDecisionReasonRequired
		}
		return normalizedDecision{
			reasonCode:  trimmedReason,
			targetState: decision,
		}, nil
	default:
		return normalizedDecision{}, ErrInvalidReviewDecision
	}
}

func validateRequiredShortDecisions(required map[uuid.UUID]struct{}, actual map[uuid.UUID]normalizedDecision) error {
	if len(required) != len(actual) {
		return ErrSubmissionReviewDecisionTargetsMismatch
	}
	for shortID := range required {
		if _, ok := actual[shortID]; !ok {
			return ErrSubmissionReviewDecisionTargetsMismatch
		}
	}

	return nil
}

func matchesIntakeMainSnapshot(
	mainRow sqlc.GetSubmissionReviewMainByIDForUpdateRow,
	intake sqlc.AppSubmissionReviewIntake,
) bool {
	return mainRow.MediaAssetID == intake.MainMediaAssetID &&
		mainRow.MediaProcessingState == mediaStateReady &&
		mainRow.PriceMinor == intake.MainPriceMinor &&
		mainRow.OwnershipConfirmed == intake.OwnershipConfirmed &&
		mainRow.ConsentConfirmed == intake.ConsentConfirmed
}

func matchesIntakeShortSnapshot(
	shortRow sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow,
	intakeShort sqlc.AppSubmissionReviewIntakeShort,
) bool {
	return shortRow.MediaAssetID == intakeShort.MediaAssetID &&
		shortRow.MediaProcessingState == mediaStateReady &&
		optionalTextEqual(shortRow.Caption, intakeShort.Caption)
}

func isAllowedCurrentMainState(state string) bool {
	return state == mainStatePendingReview || state == mainStateApprovedForUnlock
}

func isAllowedCurrentShortState(state string) bool {
	return state == shortStatePendingReview || state == shortStateApprovedForPublish
}

func buildMainDecisionUpdate(
	mainID pgtype.UUID,
	decision normalizedDecision,
	source string,
	decisionedAt time.Time,
) sqlc.ApplySubmissionReviewMainDecisionParams {
	params := sqlc.ApplySubmissionReviewMainDecisionParams{
		State:                decision.targetState,
		ReviewReasonCode:     postgres.TextToPG(decision.reasonCode),
		ReviewDecisionSource: postgres.TextToPG(&source),
		ReviewDecisionedAt:   postgres.TimeToPG(&decisionedAt),
		ApprovedForUnlockAt:  pgtype.Timestamptz{},
		ID:                   mainID,
	}
	if decision.targetState == mainStateApprovedForUnlock {
		params.ApprovedForUnlockAt = postgres.TimeToPG(&decisionedAt)
	}

	return params
}

func buildShortDecisionUpdate(
	shortID pgtype.UUID,
	decision normalizedDecision,
	source string,
	decisionedAt time.Time,
) sqlc.ApplySubmissionReviewShortDecisionParams {
	params := sqlc.ApplySubmissionReviewShortDecisionParams{
		State:                decision.targetState,
		ReviewReasonCode:     postgres.TextToPG(decision.reasonCode),
		ReviewDecisionSource: postgres.TextToPG(&source),
		ReviewDecisionedAt:   postgres.TimeToPG(&decisionedAt),
		ApprovedForPublishAt: pgtype.Timestamptz{},
		PublishedAt:          pgtype.Timestamptz{},
		ID:                   shortID,
	}
	if decision.targetState == shortStateApprovedForPublish {
		params.ApprovedForPublishAt = postgres.TimeToPG(&decisionedAt)
	}

	return params
}

func trimmedOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func optionalTextEqual(left pgtype.Text, right pgtype.Text) bool {
	leftValue := postgres.OptionalTextFromPG(left)
	rightValue := postgres.OptionalTextFromPG(right)

	switch {
	case leftValue == nil && rightValue == nil:
		return true
	case leftValue == nil || rightValue == nil:
		return false
	default:
		return *leftValue == *rightValue
	}
}
