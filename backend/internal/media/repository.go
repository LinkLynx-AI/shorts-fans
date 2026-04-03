package media

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

// ErrAssetNotFound indicates that the requested media asset does not exist.
var ErrAssetNotFound = errors.New("media asset not found")

type queries interface {
	CreateMediaAsset(ctx context.Context, arg sqlc.CreateMediaAssetParams) (sqlc.AppMediaAsset, error)
	GetMediaAssetByID(ctx context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error)
	ListMediaAssetsByCreatorUserID(ctx context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppMediaAsset, error)
	UpdateMediaAssetProcessingState(ctx context.Context, arg sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error)
}

// Repository wraps media-related persistence operations.
type Repository struct {
	queries queries
}

// Asset is the domain-facing media asset record.
type Asset struct {
	ID                uuid.UUID
	CreatorUserID     uuid.UUID
	ProcessingState   string
	StorageProvider   string
	StorageBucket     string
	StorageKey        string
	PlaybackURL       *string
	MimeType          string
	DurationMS        *int64
	ExternalUploadRef *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// CreateAssetInput is the input for CreateAsset.
type CreateAssetInput struct {
	CreatorUserID     uuid.UUID
	ProcessingState   string
	StorageProvider   string
	StorageBucket     string
	StorageKey        string
	PlaybackURL       *string
	MimeType          string
	DurationMS        *int64
	ExternalUploadRef *string
}

// UpdateAssetProcessingStateInput is the input for UpdateAssetProcessingState.
type UpdateAssetProcessingStateInput struct {
	ID                uuid.UUID
	ProcessingState   string
	PlaybackURL       *string
	DurationMS        *int64
	ExternalUploadRef *string
}

// NewRepository constructs a media repository backed by pgxpool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return newRepository(sqlc.New(pool))
}

func newRepository(q queries) *Repository {
	return &Repository{queries: q}
}

// CreateAsset creates a media asset row.
func (r *Repository) CreateAsset(ctx context.Context, input CreateAssetInput) (Asset, error) {
	row, err := r.queries.CreateMediaAsset(ctx, sqlc.CreateMediaAssetParams{
		CreatorUserID:     postgres.UUIDToPG(input.CreatorUserID),
		ProcessingState:   input.ProcessingState,
		StorageProvider:   input.StorageProvider,
		StorageBucket:     input.StorageBucket,
		StorageKey:        input.StorageKey,
		PlaybackUrl:       postgres.TextToPG(input.PlaybackURL),
		MimeType:          input.MimeType,
		DurationMs:        postgres.Int64ToPG(input.DurationMS),
		ExternalUploadRef: postgres.TextToPG(input.ExternalUploadRef),
	})
	if err != nil {
		return Asset{}, fmt.Errorf("create media asset: %w", err)
	}

	asset, err := mapAsset(row)
	if err != nil {
		return Asset{}, fmt.Errorf("create media asset: %w", err)
	}

	return asset, nil
}

// GetAsset returns a media asset by ID.
func (r *Repository) GetAsset(ctx context.Context, id uuid.UUID) (Asset, error) {
	row, err := r.queries.GetMediaAssetByID(ctx, postgres.UUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Asset{}, fmt.Errorf("get media asset %s: %w", id, ErrAssetNotFound)
		}

		return Asset{}, fmt.Errorf("get media asset %s: %w", id, err)
	}

	asset, err := mapAsset(row)
	if err != nil {
		return Asset{}, fmt.Errorf("get media asset %s: %w", id, err)
	}

	return asset, nil
}

// ListAssetsByCreator returns media assets owned by the creator.
func (r *Repository) ListAssetsByCreator(ctx context.Context, creatorUserID uuid.UUID) ([]Asset, error) {
	rows, err := r.queries.ListMediaAssetsByCreatorUserID(ctx, postgres.UUIDToPG(creatorUserID))
	if err != nil {
		return nil, fmt.Errorf("list media assets for creator %s: %w", creatorUserID, err)
	}

	assets := make([]Asset, 0, len(rows))
	for _, row := range rows {
		asset, err := mapAsset(row)
		if err != nil {
			return nil, fmt.Errorf("list media assets for creator %s: %w", creatorUserID, err)
		}

		assets = append(assets, asset)
	}

	return assets, nil
}

// UpdateAssetProcessingState updates the media processing fields.
func (r *Repository) UpdateAssetProcessingState(ctx context.Context, input UpdateAssetProcessingStateInput) (Asset, error) {
	row, err := r.queries.UpdateMediaAssetProcessingState(ctx, sqlc.UpdateMediaAssetProcessingStateParams{
		ProcessingState:   input.ProcessingState,
		PlaybackUrl:       postgres.TextToPG(input.PlaybackURL),
		DurationMs:        postgres.Int64ToPG(input.DurationMS),
		ExternalUploadRef: postgres.TextToPG(input.ExternalUploadRef),
		ID:                postgres.UUIDToPG(input.ID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Asset{}, fmt.Errorf("update media asset %s: %w", input.ID, ErrAssetNotFound)
		}

		return Asset{}, fmt.Errorf("update media asset %s: %w", input.ID, err)
	}

	asset, err := mapAsset(row)
	if err != nil {
		return Asset{}, fmt.Errorf("update media asset %s: %w", input.ID, err)
	}

	return asset, nil
}

func mapAsset(row sqlc.AppMediaAsset) (Asset, error) {
	id, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return Asset{}, fmt.Errorf("map media asset id: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return Asset{}, fmt.Errorf("map media asset creator user id: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Asset{}, fmt.Errorf("map media asset created at: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Asset{}, fmt.Errorf("map media asset updated at: %w", err)
	}

	return Asset{
		ID:                id,
		CreatorUserID:     creatorUserID,
		ProcessingState:   row.ProcessingState,
		StorageProvider:   row.StorageProvider,
		StorageBucket:     row.StorageBucket,
		StorageKey:        row.StorageKey,
		PlaybackURL:       postgres.OptionalTextFromPG(row.PlaybackUrl),
		MimeType:          row.MimeType,
		DurationMS:        postgres.OptionalInt64FromPG(row.DurationMs),
		ExternalUploadRef: postgres.OptionalTextFromPG(row.ExternalUploadRef),
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
	}, nil
}
