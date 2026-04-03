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

// ErrCapabilityNotFound indicates that the requested creator capability does not exist.
var ErrCapabilityNotFound = errors.New("creator capability not found")

// ErrProfileNotFound indicates that the requested creator profile does not exist.
var ErrProfileNotFound = errors.New("creator profile not found")

type queries interface {
	CreateCreatorCapability(ctx context.Context, arg sqlc.CreateCreatorCapabilityParams) (sqlc.AppCreatorCapability, error)
	GetCreatorCapabilityByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error)
	UpdateCreatorCapabilityState(ctx context.Context, arg sqlc.UpdateCreatorCapabilityStateParams) (sqlc.AppCreatorCapability, error)
	CreateCreatorProfile(ctx context.Context, arg sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error)
	GetCreatorProfileByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorProfile, error)
	GetPublicCreatorProfileByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppPublicCreatorProfile, error)
	UpdateCreatorProfile(ctx context.Context, arg sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error)
	PublishCreatorProfile(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorProfile, error)
}

// Repository wraps creator-related persistence operations.
type Repository struct {
	queries queries
}

// Capability is the domain-facing creator capability record.
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

// CreateCapabilityInput is the input for CreateCapability.
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

// UpdateCapabilityInput is the input for UpdateCapability.
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

// Profile is the domain-facing creator profile record.
type Profile struct {
	UserID      uuid.UUID
	DisplayName *string
	AvatarURL   *string
	Bio         string
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateProfileInput is the input for CreateProfile.
type CreateProfileInput struct {
	UserID      uuid.UUID
	DisplayName *string
	AvatarURL   *string
	Bio         string
	PublishedAt *time.Time
}

// UpdateProfileInput is the input for UpdateProfile.
type UpdateProfileInput struct {
	UserID      uuid.UUID
	DisplayName *string
	AvatarURL   *string
	Bio         string
}

// NewRepository constructs a creator repository backed by pgxpool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return newRepository(sqlc.New(pool))
}

func newRepository(q queries) *Repository {
	return &Repository{queries: q}
}

// CreateCapability creates a creator capability row.
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
		return Capability{}, fmt.Errorf("create creator capability: %w", err)
	}

	capability, err := mapCapability(row)
	if err != nil {
		return Capability{}, fmt.Errorf("create creator capability: %w", err)
	}

	return capability, nil
}

// GetCapability returns a creator capability by user ID.
func (r *Repository) GetCapability(ctx context.Context, userID uuid.UUID) (Capability, error) {
	row, err := r.queries.GetCreatorCapabilityByUserID(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Capability{}, fmt.Errorf("get creator capability %s: %w", userID, ErrCapabilityNotFound)
		}

		return Capability{}, fmt.Errorf("get creator capability %s: %w", userID, err)
	}

	capability, err := mapCapability(row)
	if err != nil {
		return Capability{}, fmt.Errorf("get creator capability %s: %w", userID, err)
	}

	return capability, nil
}

// UpdateCapability updates a creator capability row.
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
			return Capability{}, fmt.Errorf("update creator capability %s: %w", input.UserID, ErrCapabilityNotFound)
		}

		return Capability{}, fmt.Errorf("update creator capability %s: %w", input.UserID, err)
	}

	capability, err := mapCapability(row)
	if err != nil {
		return Capability{}, fmt.Errorf("update creator capability %s: %w", input.UserID, err)
	}

	return capability, nil
}

// CreateProfile creates a creator profile row.
func (r *Repository) CreateProfile(ctx context.Context, input CreateProfileInput) (Profile, error) {
	row, err := r.queries.CreateCreatorProfile(ctx, sqlc.CreateCreatorProfileParams{
		UserID:      postgres.UUIDToPG(input.UserID),
		DisplayName: postgres.TextToPG(input.DisplayName),
		AvatarUrl:   postgres.TextToPG(input.AvatarURL),
		Bio:         input.Bio,
		PublishedAt: postgres.TimeToPG(input.PublishedAt),
	})
	if err != nil {
		return Profile{}, fmt.Errorf("create creator profile: %w", err)
	}

	profile, err := mapProfile(row)
	if err != nil {
		return Profile{}, fmt.Errorf("create creator profile: %w", err)
	}

	return profile, nil
}

// GetProfile returns a creator profile by user ID.
func (r *Repository) GetProfile(ctx context.Context, userID uuid.UUID) (Profile, error) {
	row, err := r.queries.GetCreatorProfileByUserID(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, fmt.Errorf("get creator profile %s: %w", userID, ErrProfileNotFound)
		}

		return Profile{}, fmt.Errorf("get creator profile %s: %w", userID, err)
	}

	profile, err := mapProfile(row)
	if err != nil {
		return Profile{}, fmt.Errorf("get creator profile %s: %w", userID, err)
	}

	return profile, nil
}

// GetPublicProfile returns a public creator profile by user ID.
func (r *Repository) GetPublicProfile(ctx context.Context, userID uuid.UUID) (Profile, error) {
	row, err := r.queries.GetPublicCreatorProfileByUserID(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, fmt.Errorf("get public creator profile %s: %w", userID, ErrProfileNotFound)
		}

		return Profile{}, fmt.Errorf("get public creator profile %s: %w", userID, err)
	}

	profile, err := mapPublicProfile(row)
	if err != nil {
		return Profile{}, fmt.Errorf("get public creator profile %s: %w", userID, err)
	}

	return profile, nil
}

// UpdateProfile updates a creator profile row.
func (r *Repository) UpdateProfile(ctx context.Context, input UpdateProfileInput) (Profile, error) {
	row, err := r.queries.UpdateCreatorProfile(ctx, sqlc.UpdateCreatorProfileParams{
		DisplayName: postgres.TextToPG(input.DisplayName),
		AvatarUrl:   postgres.TextToPG(input.AvatarURL),
		Bio:         input.Bio,
		UserID:      postgres.UUIDToPG(input.UserID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, fmt.Errorf("update creator profile %s: %w", input.UserID, ErrProfileNotFound)
		}

		return Profile{}, fmt.Errorf("update creator profile %s: %w", input.UserID, err)
	}

	profile, err := mapProfile(row)
	if err != nil {
		return Profile{}, fmt.Errorf("update creator profile %s: %w", input.UserID, err)
	}

	return profile, nil
}

// PublishProfile marks a creator profile as public.
func (r *Repository) PublishProfile(ctx context.Context, userID uuid.UUID) (Profile, error) {
	row, err := r.queries.PublishCreatorProfile(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Profile{}, fmt.Errorf("publish creator profile %s: %w", userID, ErrProfileNotFound)
		}

		return Profile{}, fmt.Errorf("publish creator profile %s: %w", userID, err)
	}

	profile, err := mapProfile(row)
	if err != nil {
		return Profile{}, fmt.Errorf("publish creator profile %s: %w", userID, err)
	}

	return profile, nil
}

func mapCapability(row sqlc.AppCreatorCapability) (Capability, error) {
	userID, err := postgres.UUIDFromPG(row.UserID)
	if err != nil {
		return Capability{}, fmt.Errorf("map creator capability user id: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Capability{}, fmt.Errorf("map creator capability created at: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Capability{}, fmt.Errorf("map creator capability updated at: %w", err)
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
		return Profile{}, fmt.Errorf("map creator profile user id: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Profile{}, fmt.Errorf("map creator profile created at: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Profile{}, fmt.Errorf("map creator profile updated at: %w", err)
	}

	return Profile{
		UserID:      userID,
		DisplayName: postgres.OptionalTextFromPG(row.DisplayName),
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
		return Profile{}, fmt.Errorf("map public creator profile user id: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Profile{}, fmt.Errorf("map public creator profile created at: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Profile{}, fmt.Errorf("map public creator profile updated at: %w", err)
	}

	return Profile{
		UserID:      userID,
		DisplayName: postgres.OptionalTextFromPG(row.DisplayName),
		AvatarURL:   postgres.OptionalTextFromPG(row.AvatarUrl),
		Bio:         row.Bio,
		PublishedAt: postgres.OptionalTimeFromPG(row.PublishedAt),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}
