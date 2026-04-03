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

// ErrAssetNotFound は対象の media asset が存在しないことを表します。
var ErrAssetNotFound = errors.New("media asset が見つかりません")

type queries interface {
	CreateMediaAsset(ctx context.Context, arg sqlc.CreateMediaAssetParams) (sqlc.AppMediaAsset, error)
	GetMediaAssetByID(ctx context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error)
	ListMediaAssetsByCreatorUserID(ctx context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppMediaAsset, error)
	UpdateMediaAssetProcessingState(ctx context.Context, arg sqlc.UpdateMediaAssetProcessingStateParams) (sqlc.AppMediaAsset, error)
}

// Repository は media 関連の永続化操作を包みます。
type Repository struct {
	queries queries
}

// Asset は domain 向けの media asset レコードです。
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

// CreateAssetInput は CreateAsset の入力です。
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

// UpdateAssetProcessingStateInput は UpdateAssetProcessingState の入力です。
type UpdateAssetProcessingStateInput struct {
	ID                uuid.UUID
	ProcessingState   string
	PlaybackURL       *string
	DurationMS        *int64
	ExternalUploadRef *string
}

// NewRepository は pgxpool ベースの media repository を構築します。
func NewRepository(pool *pgxpool.Pool) *Repository {
	return newRepository(sqlc.New(pool))
}

func newRepository(q queries) *Repository {
	return &Repository{queries: q}
}

// CreateAsset は media asset を作成します。
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
		return Asset{}, fmt.Errorf("media asset 作成: %w", err)
	}

	asset, err := mapAsset(row)
	if err != nil {
		return Asset{}, fmt.Errorf("media asset 作成結果の変換: %w", err)
	}

	return asset, nil
}

// GetAsset は ID から media asset を取得します。
func (r *Repository) GetAsset(ctx context.Context, id uuid.UUID) (Asset, error) {
	row, err := r.queries.GetMediaAssetByID(ctx, postgres.UUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Asset{}, fmt.Errorf("media asset 取得 id=%s: %w", id, ErrAssetNotFound)
		}

		return Asset{}, fmt.Errorf("media asset 取得 id=%s: %w", id, err)
	}

	asset, err := mapAsset(row)
	if err != nil {
		return Asset{}, fmt.Errorf("media asset 取得結果の変換 id=%s: %w", id, err)
	}

	return asset, nil
}

// ListAssetsByCreator は creator が所有する media asset 一覧を返します。
func (r *Repository) ListAssetsByCreator(ctx context.Context, creatorUserID uuid.UUID) ([]Asset, error) {
	rows, err := r.queries.ListMediaAssetsByCreatorUserID(ctx, postgres.UUIDToPG(creatorUserID))
	if err != nil {
		return nil, fmt.Errorf("media asset 一覧取得 creator=%s: %w", creatorUserID, err)
	}

	assets := make([]Asset, 0, len(rows))
	for _, row := range rows {
		asset, err := mapAsset(row)
		if err != nil {
			return nil, fmt.Errorf("media asset 一覧取得結果の変換 creator=%s: %w", creatorUserID, err)
		}

		assets = append(assets, asset)
	}

	return assets, nil
}

// UpdateAssetProcessingState は media asset の processing 状態を更新します。
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
			return Asset{}, fmt.Errorf("media asset 更新 id=%s: %w", input.ID, ErrAssetNotFound)
		}

		return Asset{}, fmt.Errorf("media asset 更新 id=%s: %w", input.ID, err)
	}

	asset, err := mapAsset(row)
	if err != nil {
		return Asset{}, fmt.Errorf("media asset 更新結果の変換 id=%s: %w", input.ID, err)
	}

	return asset, nil
}

func mapAsset(row sqlc.AppMediaAsset) (Asset, error) {
	id, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return Asset{}, fmt.Errorf("media asset の id 変換: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return Asset{}, fmt.Errorf("media asset の creator user id 変換: %w", err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return Asset{}, fmt.Errorf("media asset の created_at 変換: %w", err)
	}
	updatedAt, err := postgres.RequiredTimeFromPG(row.UpdatedAt)
	if err != nil {
		return Asset{}, fmt.Errorf("media asset の updated_at 変換: %w", err)
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
