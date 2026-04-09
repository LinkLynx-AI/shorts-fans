package creator

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
)

// ErrInvalidDisplayName は creator registration の display name が不正なことを表します。
var ErrInvalidDisplayName = errors.New("creator display name が不正です")

// SelfServeRegistrationInput は self-serve creator registration の入力です。
type SelfServeRegistrationInput struct {
	Bio         string
	DisplayName string
	Handle      string
	UserID      uuid.UUID
}

// SelfServeRegistrationResult は self-serve creator registration の結果です。
type SelfServeRegistrationResult struct {
	Capability Capability
	Profile    Profile
}

// RegisterApprovedCreator は self-serve creator registration を即時 approved で登録します。
func (r *Repository) RegisterApprovedCreator(
	ctx context.Context,
	input SelfServeRegistrationInput,
) (SelfServeRegistrationResult, error) {
	if r == nil || r.txBeginner == nil {
		return SelfServeRegistrationResult{}, fmt.Errorf("creator repository pool が初期化されていません")
	}
	if r.newQueries == nil {
		return SelfServeRegistrationResult{}, fmt.Errorf("creator repository query factory が初期化されていません")
	}

	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		return SelfServeRegistrationResult{}, ErrInvalidDisplayName
	}
	handle, err := normalizeRequiredHandle(input.Handle)
	if err != nil {
		return SelfServeRegistrationResult{}, err
	}

	bio := strings.TrimSpace(input.Bio)
	now := time.Now().UTC()
	var result SelfServeRegistrationResult

	err = postgres.RunInTx(ctx, r.txBeginner, func(tx pgx.Tx) error {
		q := r.newQueries(tx)

		capability, err := upsertApprovedCapability(ctx, q, input.UserID, now)
		if err != nil {
			return err
		}

		profile, err := upsertPrivateProfile(ctx, q, input.UserID, displayName, handle, bio)
		if err != nil {
			return err
		}

		result = SelfServeRegistrationResult{
			Capability: capability,
			Profile:    profile,
		}
		return nil
	})
	if err != nil {
		return SelfServeRegistrationResult{}, fmt.Errorf("creator self-serve registration user=%s: %w", input.UserID, err)
	}

	return result, nil
}

func upsertApprovedCapability(ctx context.Context, q queries, userID uuid.UUID, now time.Time) (Capability, error) {
	existingRow, err := q.GetCreatorCapabilityByUserID(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return Capability{}, fmt.Errorf("creator capability 取得 user=%s: %w", userID, err)
		}

		createdRow, createErr := q.CreateCreatorCapability(ctx, sqlc.CreateCreatorCapabilityParams{
			UserID:                   postgres.UUIDToPG(userID),
			State:                    "approved",
			RejectionReasonCode:      postgres.TextToPG(nil),
			IsResubmitEligible:       false,
			IsSupportReviewRequired:  false,
			SelfServeResubmitCount:   0,
			KycProviderCaseRef:       postgres.TextToPG(nil),
			PayoutProviderAccountRef: postgres.TextToPG(nil),
			SubmittedAt:              postgres.TimeToPG(nil),
			ApprovedAt:               postgres.TimeToPG(&now),
			RejectedAt:               postgres.TimeToPG(nil),
			SuspendedAt:              postgres.TimeToPG(nil),
		})
		if createErr != nil {
			return Capability{}, fmt.Errorf("creator capability 作成 user=%s: %w", userID, createErr)
		}

		capability, mapErr := mapCapability(createdRow)
		if mapErr != nil {
			return Capability{}, fmt.Errorf("creator capability 作成結果の変換 user=%s: %w", userID, mapErr)
		}

		return capability, nil
	}

	existingCapability, err := mapCapability(existingRow)
	if err != nil {
		return Capability{}, fmt.Errorf("creator capability 取得結果の変換 user=%s: %w", userID, err)
	}

	if existingCapability.State == "approved" {
		return existingCapability, nil
	}

	updatedRow, err := q.UpdateCreatorCapabilityState(ctx, sqlc.UpdateCreatorCapabilityStateParams{
		State:                    "approved",
		RejectionReasonCode:      postgres.TextToPG(nil),
		IsResubmitEligible:       false,
		IsSupportReviewRequired:  false,
		SelfServeResubmitCount:   0,
		KycProviderCaseRef:       postgres.TextToPG(existingCapability.KYCProviderCaseRef),
		PayoutProviderAccountRef: postgres.TextToPG(existingCapability.PayoutProviderAccountRef),
		SubmittedAt:              postgres.TimeToPG(existingCapability.SubmittedAt),
		ApprovedAt:               postgres.TimeToPG(&now),
		RejectedAt:               postgres.TimeToPG(nil),
		SuspendedAt:              postgres.TimeToPG(nil),
		UserID:                   postgres.UUIDToPG(userID),
	})
	if err != nil {
		return Capability{}, fmt.Errorf("creator capability 更新 user=%s: %w", userID, err)
	}

	capability, err := mapCapability(updatedRow)
	if err != nil {
		return Capability{}, fmt.Errorf("creator capability 更新結果の変換 user=%s: %w", userID, err)
	}

	return capability, nil
}

func upsertPrivateProfile(ctx context.Context, q queries, userID uuid.UUID, displayName string, handle string, bio string) (Profile, error) {
	existingRow, err := q.GetCreatorProfileByUserID(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, fmt.Errorf("creator profile 取得 user=%s: %w", userID, err)
		}

		createdRow, createErr := q.CreateCreatorProfile(ctx, sqlc.CreateCreatorProfileParams{
			UserID:      postgres.UUIDToPG(userID),
			DisplayName: postgres.TextToPG(&displayName),
			Handle:      handle,
			AvatarUrl:   postgres.TextToPG(nil),
			Bio:         bio,
			PublishedAt: postgres.TimeToPG(nil),
		})
		if createErr != nil {
			createErr = mapProfileWriteError(createErr)
			return Profile{}, fmt.Errorf("creator profile 作成 user=%s: %w", userID, createErr)
		}

		profile, mapErr := mapProfile(createdRow)
		if mapErr != nil {
			return Profile{}, fmt.Errorf("creator profile 作成結果の変換 user=%s: %w", userID, mapErr)
		}

		return profile, nil
	}

	existingProfile, err := mapProfile(existingRow)
	if err != nil {
		return Profile{}, fmt.Errorf("creator profile 取得結果の変換 user=%s: %w", userID, err)
	}

	updatedRow, err := q.UpdateCreatorProfile(ctx, sqlc.UpdateCreatorProfileParams{
		DisplayName: postgres.TextToPG(&displayName),
		Handle:      handle,
		AvatarUrl:   postgres.TextToPG(existingProfile.AvatarURL),
		Bio:         bio,
		UserID:      postgres.UUIDToPG(userID),
	})
	if err != nil {
		err = mapProfileWriteError(err)
		return Profile{}, fmt.Errorf("creator profile 更新 user=%s: %w", userID, err)
	}

	profile, err := mapProfile(updatedRow)
	if err != nil {
		return Profile{}, fmt.Errorf("creator profile 更新結果の変換 user=%s: %w", userID, err)
	}

	return profile, nil
}
