package feed

import (
	"context"
	"fmt"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type recommendedQueries interface {
	ListRecommendedFeedItems(ctx context.Context, arg sqlc.ListRecommendedFeedItemsParams) ([]sqlc.ListRecommendedFeedItemsRow, error)
}

type recommendedCursor struct {
	PublishedAt time.Time
	ShortID     uuid.UUID
}

type recommendedRecord struct {
	ShortID              uuid.UUID
	CanonicalMainID      uuid.UUID
	CreatorUserID        uuid.UUID
	ShortTitle           string
	ShortCaption         string
	ShortMediaAssetID    uuid.UUID
	ShortPublishedAt     time.Time
	ShortPlaybackURL     string
	ShortDurationSeconds int
	CreatorDisplayName   string
	CreatorHandle        string
	CreatorAvatarURL     *string
	CreatorBio           string
	MainID               uuid.UUID
	MainPriceJPY         int64
	MainDurationSeconds  int
}

// Repository は recommended feed 用の read query を包みます。
type Repository struct {
	queries recommendedQueries
}

// NewRepository は pgxpool ベースの feed repository を構築します。
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: sqlc.New(pool)}
}

func newRepository(q recommendedQueries) *Repository {
	return &Repository{queries: q}
}

func (r *Repository) listRecommended(ctx context.Context, cursor *recommendedCursor, limit int32) ([]recommendedRecord, error) {
	params := sqlc.ListRecommendedFeedItemsParams{
		PageLimit: limit,
	}
	if cursor != nil {
		params.CursorPublishedAt = postgres.TimeToPG(&cursor.PublishedAt)
		params.CursorShortID = postgres.UUIDToPG(cursor.ShortID)
	}

	rows, err := r.queries.ListRecommendedFeedItems(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("recommended feed 一覧取得: %w", err)
	}

	records := make([]recommendedRecord, 0, len(rows))
	for _, row := range rows {
		record, err := mapRecommendedRecord(row)
		if err != nil {
			return nil, fmt.Errorf("recommended feed 一覧取得結果の変換: %w", err)
		}

		records = append(records, record)
	}

	return records, nil
}

func mapRecommendedRecord(row sqlc.ListRecommendedFeedItemsRow) (recommendedRecord, error) {
	shortID, err := postgres.UUIDFromPG(row.ShortID)
	if err != nil {
		return recommendedRecord{}, fmt.Errorf("short id 変換: %w", err)
	}
	canonicalMainID, err := postgres.UUIDFromPG(row.CanonicalMainID)
	if err != nil {
		return recommendedRecord{}, fmt.Errorf("canonical main id 変換: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return recommendedRecord{}, fmt.Errorf("creator user id 変換: %w", err)
	}
	shortMediaAssetID, err := postgres.UUIDFromPG(row.ShortMediaAssetID)
	if err != nil {
		return recommendedRecord{}, fmt.Errorf("short media asset id 変換: %w", err)
	}
	shortPublishedAt, err := postgres.RequiredTimeFromPG(row.ShortPublishedAt)
	if err != nil {
		return recommendedRecord{}, fmt.Errorf("short published_at 変換: %w", err)
	}
	mainID, err := postgres.UUIDFromPG(row.MainID)
	if err != nil {
		return recommendedRecord{}, fmt.Errorf("main id 変換: %w", err)
	}
	if !row.ShortPlaybackUrl.Valid {
		return recommendedRecord{}, fmt.Errorf("short playback url が null です")
	}
	if !row.ShortDurationMs.Valid {
		return recommendedRecord{}, fmt.Errorf("short duration_ms が null です")
	}
	if !row.CreatorDisplayName.Valid {
		return recommendedRecord{}, fmt.Errorf("creator display_name が null です")
	}
	if !row.CreatorHandle.Valid {
		return recommendedRecord{}, fmt.Errorf("creator handle が null です")
	}
	if !row.MainPriceMinor.Valid {
		return recommendedRecord{}, fmt.Errorf("main price_minor が null です")
	}
	if !row.MainDurationMs.Valid {
		return recommendedRecord{}, fmt.Errorf("main duration_ms が null です")
	}

	return recommendedRecord{
		ShortID:              shortID,
		CanonicalMainID:      canonicalMainID,
		CreatorUserID:        creatorUserID,
		ShortTitle:           row.ShortTitle,
		ShortCaption:         row.ShortCaption,
		ShortMediaAssetID:    shortMediaAssetID,
		ShortPublishedAt:     shortPublishedAt,
		ShortPlaybackURL:     row.ShortPlaybackUrl.String,
		ShortDurationSeconds: durationSecondsFromMilliseconds(row.ShortDurationMs.Int64),
		CreatorDisplayName:   row.CreatorDisplayName.String,
		CreatorHandle:        row.CreatorHandle.String,
		CreatorAvatarURL:     postgres.OptionalTextFromPG(row.CreatorAvatarUrl),
		CreatorBio:           row.CreatorBio,
		MainID:               mainID,
		MainPriceJPY:         row.MainPriceMinor.Int64,
		MainDurationSeconds:  durationSecondsFromMilliseconds(row.MainDurationMs.Int64),
	}, nil
}

func durationSecondsFromMilliseconds(milliseconds int64) int {
	if milliseconds <= 0 {
		return 0
	}

	return int((milliseconds + 999) / 1000)
}
