package creatorregistration

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

var (
	// ErrInvalidReviewDecision は未知の review decision が指定されたことを表します。
	ErrInvalidReviewDecision = errors.New("creator registration review decision が不正です")
	// ErrReviewDecisionConflict は現在 state では許可されない遷移が指定されたことを表します。
	ErrReviewDecisionConflict = errors.New("creator registration review decision state conflict")
	// ErrReviewDecisionReasonRequired は rejected decision に reason code が不足していることを表します。
	ErrReviewDecisionReasonRequired = errors.New("creator registration rejection reason code が必要です")
	// ErrReviewDecisionMetadataConflict は rejected metadata の組み合わせが矛盾していることを表します。
	ErrReviewDecisionMetadataConflict = errors.New("creator registration rejection metadata が不正です")
)

// ReviewDecisionInput は reviewer が creator registration に対して適用する decision 入力です。
type ReviewDecisionInput struct {
	Decision                string
	IsResubmitEligible      bool
	IsSupportReviewRequired bool
	ReasonCode              *string
	UserID                  uuid.UUID
}

// ApplyReviewDecision は reviewer の decision を creator registration state に反映します。
func (r *Repository) ApplyReviewDecision(ctx context.Context, input ReviewDecisionInput) (Registration, error) {
	if r == nil || r.txBeginner == nil || r.newQueries == nil {
		return Registration{}, fmt.Errorf("creator registration repository が初期化されていません")
	}

	var registration Registration
	err := postgres.RunInTx(ctx, r.txBeginner, func(tx pgx.Tx) error {
		q := r.newQueries(tx)
		snapshot, err := r.loadSnapshot(ctx, q, input.UserID, false)
		if err != nil {
			return err
		}
		if err := validateReviewDecisionInput(snapshot.capability, input); err != nil {
			return err
		}

		updatedCapability, err := q.UpdateCreatorCapabilityState(
			ctx,
			buildReviewDecisionUpdateParams(snapshot.capability, input),
		)
		if err != nil {
			return fmt.Errorf("creator registration decision 反映 user=%s: %w", input.UserID, err)
		}

		snapshot.capability = &updatedCapability
		registration, err = buildRegistration(snapshot)
		return err
	})
	if err != nil {
		return Registration{}, err
	}

	return registration, nil
}

func buildReviewDecisionUpdateParams(
	capability *sqlc.AppCreatorCapability,
	input ReviewDecisionInput,
) sqlc.UpdateCreatorCapabilityStateParams {
	now := time.Now().UTC()
	decision := strings.TrimSpace(input.Decision)

	base := sqlc.UpdateCreatorCapabilityStateParams{
		KycProviderCaseRef:       capability.KycProviderCaseRef,
		PayoutProviderAccountRef: capability.PayoutProviderAccountRef,
		SubmittedAt:              capability.SubmittedAt,
		UserID:                   postgres.UUIDToPG(input.UserID),
	}

	switch decision {
	case StateApproved:
		return sqlc.UpdateCreatorCapabilityStateParams{
			State:                    StateApproved,
			RejectionReasonCode:      pgtype.Text{},
			IsResubmitEligible:       false,
			IsSupportReviewRequired:  false,
			SelfServeResubmitCount:   0,
			KycProviderCaseRef:       base.KycProviderCaseRef,
			PayoutProviderAccountRef: base.PayoutProviderAccountRef,
			SubmittedAt:              base.SubmittedAt,
			ApprovedAt:               postgres.TimeToPG(&now),
			RejectedAt:               pgtype.Timestamptz{},
			SuspendedAt:              pgtype.Timestamptz{},
			UserID:                   base.UserID,
		}
	case StateRejected:
		return sqlc.UpdateCreatorCapabilityStateParams{
			State:                    StateRejected,
			RejectionReasonCode:      postgres.TextToPG(trimmedOptionalString(input.ReasonCode)),
			IsResubmitEligible:       input.IsResubmitEligible,
			IsSupportReviewRequired:  input.IsSupportReviewRequired,
			SelfServeResubmitCount:   capability.SelfServeResubmitCount,
			KycProviderCaseRef:       base.KycProviderCaseRef,
			PayoutProviderAccountRef: base.PayoutProviderAccountRef,
			SubmittedAt:              base.SubmittedAt,
			ApprovedAt:               pgtype.Timestamptz{},
			RejectedAt:               postgres.TimeToPG(&now),
			SuspendedAt:              pgtype.Timestamptz{},
			UserID:                   base.UserID,
		}
	case StateSuspended:
		return sqlc.UpdateCreatorCapabilityStateParams{
			State:                    StateSuspended,
			RejectionReasonCode:      pgtype.Text{},
			IsResubmitEligible:       false,
			IsSupportReviewRequired:  false,
			SelfServeResubmitCount:   capability.SelfServeResubmitCount,
			KycProviderCaseRef:       base.KycProviderCaseRef,
			PayoutProviderAccountRef: base.PayoutProviderAccountRef,
			SubmittedAt:              base.SubmittedAt,
			ApprovedAt:               capability.ApprovedAt,
			RejectedAt:               pgtype.Timestamptz{},
			SuspendedAt:              postgres.TimeToPG(&now),
			UserID:                   base.UserID,
		}
	default:
		return sqlc.UpdateCreatorCapabilityStateParams{}
	}
}

func validateReviewDecisionInput(capability *sqlc.AppCreatorCapability, input ReviewDecisionInput) error {
	if capability == nil {
		return ErrReviewDecisionConflict
	}

	switch strings.TrimSpace(input.Decision) {
	case StateApproved:
		if capability.State != StateSubmitted {
			return ErrReviewDecisionConflict
		}
		return nil
	case StateRejected:
		if capability.State != StateSubmitted {
			return ErrReviewDecisionConflict
		}
		if trimmedOptionalString(input.ReasonCode) == nil {
			return ErrReviewDecisionReasonRequired
		}
		if input.IsSupportReviewRequired && input.IsResubmitEligible {
			return ErrReviewDecisionMetadataConflict
		}
		return nil
	case StateSuspended:
		if capability.State != StateApproved {
			return ErrReviewDecisionConflict
		}
		return nil
	default:
		return ErrInvalidReviewDecision
	}
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
