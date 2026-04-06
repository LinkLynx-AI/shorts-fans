package creator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrCapabilityNotFound は対象の creator capability が存在しないことを表します。
var ErrCapabilityNotFound = errors.New("creator capability が見つかりません")

// ErrProfileNotFound は対象の creator profile が存在しないことを表します。
var ErrProfileNotFound = errors.New("creator profile が見つかりません")

// ErrInvalidHandle は creator handle の形式が不正なことを表します。
var ErrInvalidHandle = errors.New("creator handle が不正です")

type queries interface {
	CountCreatorFollowersByCreatorUserID(ctx context.Context, creatorUserID pgtype.UUID) (int64, error)
	CreateCreatorCapability(ctx context.Context, arg sqlc.CreateCreatorCapabilityParams) (sqlc.AppCreatorCapability, error)
	GetCreatorCapabilityByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error)
	UpdateCreatorCapabilityState(ctx context.Context, arg sqlc.UpdateCreatorCapabilityStateParams) (sqlc.AppCreatorCapability, error)
	CountPublicShortsByCreatorUserID(ctx context.Context, creatorUserID pgtype.UUID) (int64, error)
	CreateCreatorProfile(ctx context.Context, arg sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error)
	GetCreatorProfileByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorProfile, error)
	GetPublicCreatorProfileByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppPublicCreatorProfile, error)
	GetPublicCreatorProfileByHandle(ctx context.Context, handle pgtype.Text) (sqlc.AppPublicCreatorProfile, error)
	ListCreatorProfileShortGridItems(ctx context.Context, arg sqlc.ListCreatorProfileShortGridItemsParams) ([]sqlc.ListCreatorProfileShortGridItemsRow, error)
	ListRecentPublicCreatorProfiles(ctx context.Context, arg sqlc.ListRecentPublicCreatorProfilesParams) ([]sqlc.AppPublicCreatorProfile, error)
	SearchPublicCreatorProfiles(ctx context.Context, arg sqlc.SearchPublicCreatorProfilesParams) ([]sqlc.AppPublicCreatorProfile, error)
	UpdateCreatorProfile(ctx context.Context, arg sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error)
	PublishCreatorProfile(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorProfile, error)
}

// Repository は creator 関連の永続化操作を包みます。
type Repository struct {
	queries queries
}

// Capability は domain 向けの creator capability レコードです。
type Capability struct {
	UserID                   uuid.UUID
	State                    string
	RejectionReasonCode      *string
	IsResubmitEligible       bool
	IsSupportReviewRequired  bool
	SelfServeResubmitCount   int32
	KYCProviderCaseRef       *string
	PayoutProviderAccountRef *string
	SubmittedAt              *time.Time
	ApprovedAt               *time.Time
	RejectedAt               *time.Time
	SuspendedAt              *time.Time
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

// CreateCapabilityInput は CreateCapability の入力です。
type CreateCapabilityInput struct {
	UserID                   uuid.UUID
	State                    string
	RejectionReasonCode      *string
	IsResubmitEligible       bool
	IsSupportReviewRequired  bool
	SelfServeResubmitCount   int32
	KYCProviderCaseRef       *string
	PayoutProviderAccountRef *string
	SubmittedAt              *time.Time
	ApprovedAt               *time.Time
	RejectedAt               *time.Time
	SuspendedAt              *time.Time
}

// UpdateCapabilityInput は UpdateCapability の入力です。
type UpdateCapabilityInput struct {
	UserID                   uuid.UUID
	State                    string
	RejectionReasonCode      *string
	IsResubmitEligible       bool
	IsSupportReviewRequired  bool
	SelfServeResubmitCount   int32
	KYCProviderCaseRef       *string
	PayoutProviderAccountRef *string
	SubmittedAt              *time.Time
	ApprovedAt               *time.Time
	RejectedAt               *time.Time
	SuspendedAt              *time.Time
}

// Profile は domain 向けの creator profile レコードです。
type Profile struct {
	UserID      uuid.UUID
	DisplayName *string
	Handle      *string
	AvatarURL   *string
	Bio         string
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateProfileInput は CreateProfile の入力です。
type CreateProfileInput struct {
	UserID      uuid.UUID
	DisplayName *string
	Handle      *string
	AvatarURL   *string
	Bio         string
	PublishedAt *time.Time
}

// UpdateProfileInput は UpdateProfile の入力です。
type UpdateProfileInput struct {
	UserID      uuid.UUID
	DisplayName *string
	Handle      *string
	AvatarURL   *string
	Bio         string
}

// NewRepository は pgxpool ベースの creator repository を構築します。
func NewRepository(pool *pgxpool.Pool) *Repository {
	return newRepository(sqlc.New(pool))
}

func newRepository(q queries) *Repository {
	return &Repository{queries: q}
}

// CreateCapability は creator capability を作成します。
func (r *Repository) CreateCapability(ctx context.Context, input CreateCapabilityInput) (Capability, error) {
	row, err := r.queries.CreateCreatorCapability(ctx, sqlc.CreateCreatorCapabilityParams{
		UserID:                   postgres.UUIDToPG(input.UserID),
		State:                    input.State,
		RejectionReasonCode:      postgres.TextToPG(input.RejectionReasonCode),
		IsResubmitEligible:       input.IsResubmitEligible,
		IsSupportReviewRequired:  input.IsSupportReviewRequired,
		SelfServeResubmitCount:   input.SelfServeResubmitCount,
		KycProviderCaseRef:       postgres.TextToPG(input.KYCProviderCaseRef),
		PayoutProviderAccountRef: postgres.TextToPG(input.PayoutProviderAccountRef),
		SubmittedAt:              postgres.TimeToPG(input.SubmittedAt),
		ApprovedAt:               postgres.TimeToPG(input.ApprovedAt),
		RejectedAt:               postgres.TimeToPG(input.RejectedAt),
		SuspendedAt:              postgres.TimeToPG(input.SuspendedAt),
	})
	if err != nil {
		return Capability{}, fmt.Errorf("creator capability 作成: %w", err)
	}

	capability, err := mapCapability(row)
	if err != nil {
		return Capability{}, fmt.Errorf("creator capability 作成結果の変換: %w", err)
	}

	return capability, nil
}

// GetCapability は user ID から creator capability を取得します。
func (r *Repository) GetCapability(ctx context.Context, userID uuid.UUID) (Capability, error) {
	row, err := r.queries.GetCreatorCapabilityByUserID(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Capability{}, fmt.Errorf("creator capability 取得 user=%s: %w", userID, ErrCapabilityNotFound)
		}

		return Capability{}, fmt.Errorf("creator capability 取得 user=%s: %w", userID, err)
	}

	capability, err := mapCapability(row)
	if err != nil {
		return Capability{}, fmt.Errorf("creator capability 取得結果の変換 user=%s: %w", userID, err)
	}

	return capability, nil
}

// UpdateCapability は creator capability を更新します。
func (r *Repository) UpdateCapability(ctx context.Context, input UpdateCapabilityInput) (Capability, error) {
	row, err := r.queries.UpdateCreatorCapabilityState(ctx, sqlc.UpdateCreatorCapabilityStateParams{
		State:                    input.State,
		RejectionReasonCode:      postgres.TextToPG(input.RejectionReasonCode),
		IsResubmitEligible:       input.IsResubmitEligible,
		IsSupportReviewRequired:  input.IsSupportReviewRequired,
		SelfServeResubmitCount:   input.SelfServeResubmitCount,
		KycProviderCaseRef:       postgres.TextToPG(input.KYCProviderCaseRef),
		PayoutProviderAccountRef: postgres.TextToPG(input.PayoutProviderAccountRef),
		SubmittedAt:              postgres.TimeToPG(input.SubmittedAt),
		ApprovedAt:               postgres.TimeToPG(input.ApprovedAt),
		RejectedAt:               postgres.TimeToPG(input.RejectedAt),
		SuspendedAt:              postgres.TimeToPG(input.SuspendedAt),
		UserID:                   postgres.UUIDToPG(input.UserID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Capability{}, fmt.Errorf("creator capability 更新 user=%s: %w", input.UserID, ErrCapabilityNotFound)
		}

		return Capability{}, fmt.Errorf("creator capability 更新 user=%s: %w", input.UserID, err)
	}

	capability, err := mapCapability(row)
	if err != nil {
		return Capability{}, fmt.Errorf("creator capability 更新結果の変換 user=%s: %w", input.UserID, err)
	}

	return capability, nil
}

// CreateProfile は creator profile を作成します。
func (r *Repository) CreateProfile(ctx context.Context, input CreateProfileInput) (Profile, error) {
	handle, err := normalizeStoredHandle(input.Handle)
	if err != nil {
		return Profile{}, fmt.Errorf("creator profile 作成 handle 正規化: %w", err)
	}

	row, err := r.queries.CreateCreatorProfile(ctx, sqlc.CreateCreatorProfileParams{
		UserID:      postgres.UUIDToPG(input.UserID),
		DisplayName: postgres.TextToPG(input.DisplayName),
		Handle:      postgres.TextToPG(handle),
		AvatarUrl:   postgres.TextToPG(input.AvatarURL),
		Bio:         input.Bio,
		PublishedAt: postgres.TimeToPG(input.PublishedAt),
	})
	if err != nil {
		return Profile{}, fmt.Errorf("creator profile 作成: %w", err)
	}

	profile, err := mapProfile(row)
	if err != nil {
		return Profile{}, fmt.Errorf("creator profile 作成結果の変換: %w", err)
	}

	return profile, nil
}

// GetProfile は user ID から creator profile を取得します。
func (r *Repository) GetProfile(ctx context.Context, userID uuid.UUID) (Profile, error) {
	row, err := r.queries.GetCreatorProfileByUserID(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, fmt.Errorf("creator profile 取得 user=%s: %w", userID, ErrProfileNotFound)
		}

		return Profile{}, fmt.Errorf("creator profile 取得 user=%s: %w", userID, err)
	}

	profile, err := mapProfile(row)
	if err != nil {
		return Profile{}, fmt.Errorf("creator profile 取得結果の変換 user=%s: %w", userID, err)
	}

	return profile, nil
}

// GetPublicProfile は user ID から公開中の creator profile を取得します。
func (r *Repository) GetPublicProfile(ctx context.Context, userID uuid.UUID) (Profile, error) {
	row, err := r.queries.GetPublicCreatorProfileByUserID(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, fmt.Errorf("公開 creator profile 取得 user=%s: %w", userID, ErrProfileNotFound)
		}

		return Profile{}, fmt.Errorf("公開 creator profile 取得 user=%s: %w", userID, err)
	}

	profile, err := mapPublicProfile(row)
	if err != nil {
		return Profile{}, fmt.Errorf("公開 creator profile 取得結果の変換 user=%s: %w", userID, err)
	}

	return profile, nil
}

// UpdateProfile は creator profile を更新します。
func (r *Repository) UpdateProfile(ctx context.Context, input UpdateProfileInput) (Profile, error) {
	handle, err := normalizeStoredHandle(input.Handle)
	if err != nil {
		return Profile{}, fmt.Errorf("creator profile 更新 handle 正規化: %w", err)
	}

	row, err := r.queries.UpdateCreatorProfile(ctx, sqlc.UpdateCreatorProfileParams{
		DisplayName: postgres.TextToPG(input.DisplayName),
		Handle:      postgres.TextToPG(handle),
		AvatarUrl:   postgres.TextToPG(input.AvatarURL),
		Bio:         input.Bio,
		UserID:      postgres.UUIDToPG(input.UserID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, fmt.Errorf("creator profile 更新 user=%s: %w", input.UserID, ErrProfileNotFound)
		}

		return Profile{}, fmt.Errorf("creator profile 更新 user=%s: %w", input.UserID, err)
	}

	profile, err := mapProfile(row)
	if err != nil {
		return Profile{}, fmt.Errorf("creator profile 更新結果の変換 user=%s: %w", input.UserID, err)
	}

	return profile, nil
}

// PublishProfile は creator profile を公開状態にします。
func (r *Repository) PublishProfile(ctx context.Context, userID uuid.UUID) (Profile, error) {
	row, err := r.queries.PublishCreatorProfile(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, fmt.Errorf("creator profile 公開 user=%s: %w", userID, ErrProfileNotFound)
		}

		return Profile{}, fmt.Errorf("creator profile 公開 user=%s: %w", userID, err)
	}

	profile, err := mapProfile(row)
	if err != nil {
		return Profile{}, fmt.Errorf("creator profile 公開結果の変換 user=%s: %w", userID, err)
	}

	return profile, nil
}

func mapCapability(row sqlc.AppCreatorCapability) (Capability, error) {
	userID, err := postgres.UUIDFromPG(row.UserID)
	if err != nil {
		return Capability{}, fmt.Errorf("creator capability の user id 変換: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Capability{}, fmt.Errorf("creator capability の created_at 変換: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Capability{}, fmt.Errorf("creator capability の updated_at 変換: %w", err)
	}

	return Capability{
		UserID:                   userID,
		State:                    row.State,
		RejectionReasonCode:      postgres.OptionalTextFromPG(row.RejectionReasonCode),
		IsResubmitEligible:       row.IsResubmitEligible,
		IsSupportReviewRequired:  row.IsSupportReviewRequired,
		SelfServeResubmitCount:   row.SelfServeResubmitCount,
		KYCProviderCaseRef:       postgres.OptionalTextFromPG(row.KycProviderCaseRef),
		PayoutProviderAccountRef: postgres.OptionalTextFromPG(row.PayoutProviderAccountRef),
		SubmittedAt:              postgres.OptionalTimeFromPG(row.SubmittedAt),
		ApprovedAt:               postgres.OptionalTimeFromPG(row.ApprovedAt),
		RejectedAt:               postgres.OptionalTimeFromPG(row.RejectedAt),
		SuspendedAt:              postgres.OptionalTimeFromPG(row.SuspendedAt),
		CreatedAt:                createdAt,
		UpdatedAt:                updatedAt,
	}, nil
}

func mapProfile(row sqlc.AppCreatorProfile) (Profile, error) {
	userID, err := postgres.UUIDFromPG(row.UserID)
	if err != nil {
		return Profile{}, fmt.Errorf("creator profile の user id 変換: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Profile{}, fmt.Errorf("creator profile の created_at 変換: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Profile{}, fmt.Errorf("creator profile の updated_at 変換: %w", err)
	}

	return Profile{
		UserID:      userID,
		DisplayName: postgres.OptionalTextFromPG(row.DisplayName),
		Handle:      postgres.OptionalTextFromPG(row.Handle),
		AvatarURL:   postgres.OptionalTextFromPG(row.AvatarUrl),
		Bio:         row.Bio,
		PublishedAt: postgres.OptionalTimeFromPG(row.PublishedAt),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

func mapPublicProfile(row sqlc.AppPublicCreatorProfile) (Profile, error) {
	userID, err := postgres.UUIDFromPG(row.UserID)
	if err != nil {
		return Profile{}, fmt.Errorf("公開 creator profile の user id 変換: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Profile{}, fmt.Errorf("公開 creator profile の created_at 変換: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Profile{}, fmt.Errorf("公開 creator profile の updated_at 変換: %w", err)
	}

	return Profile{
		UserID:      userID,
		DisplayName: postgres.OptionalTextFromPG(row.DisplayName),
		Handle:      postgres.OptionalTextFromPG(row.Handle),
		AvatarURL:   postgres.OptionalTextFromPG(row.AvatarUrl),
		Bio:         row.Bio,
		PublishedAt: postgres.OptionalTimeFromPG(row.PublishedAt),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}
