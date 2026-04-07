package creator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
)

// DefaultPublicProfileShortGridPageSize は creator profile short grid の既定 page size です。
const DefaultPublicProfileShortGridPageSize = 18

// PublicProfileHeader は creator profile header 用の read model です。
type PublicProfileHeader struct {
	Profile     Profile
	ShortCount  int64
	FanCount    int64
	IsFollowing bool
}

// PublicProfileShort は creator profile short grid 用の read model です。
type PublicProfileShort struct {
	ID                     uuid.UUID
	CreatorUserID          uuid.UUID
	CanonicalMainID        uuid.UUID
	MediaAssetID           uuid.UUID
	MediaURL               string
	PreviewDurationSeconds int64
	PublishedAt            time.Time
}

// PublicProfileShortCursor は creator profile short grid の keyset cursor です。
type PublicProfileShortCursor struct {
	PublishedAt time.Time
	ShortID     uuid.UUID
}

// GetPublicProfileHeader は creator profile header に必要な公開情報を返します。
func (r *Repository) GetPublicProfileHeader(ctx context.Context, creatorID string, viewerUserID *uuid.UUID) (PublicProfileHeader, error) {
	userID, err := ParsePublicID(creatorID)
	if err != nil {
		return PublicProfileHeader{}, fmt.Errorf("公開 creator profile header 取得 creator=%q: %w", creatorID, ErrProfileNotFound)
	}

	profile, err := r.GetPublicProfile(ctx, userID)
	if err != nil {
		return PublicProfileHeader{}, fmt.Errorf("公開 creator profile header 取得 creator=%q: %w", creatorID, err)
	}

	shortCount, err := r.queries.CountPublicShortsByCreatorUserID(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		return PublicProfileHeader{}, fmt.Errorf("公開 creator profile short count 取得 creator=%q: %w", creatorID, err)
	}

	fanCount, err := r.queries.CountCreatorFollowersByCreatorUserID(ctx, postgres.UUIDToPG(userID))
	if err != nil {
		return PublicProfileHeader{}, fmt.Errorf("公開 creator profile follower count 取得 creator=%q: %w", creatorID, err)
	}

	isFollowing := false
	if viewerUserID != nil && *viewerUserID != userID {
		isFollowing, err = r.queries.HasCreatorFollowByUserIDAndCreatorUserID(ctx, sqlc.HasCreatorFollowByUserIDAndCreatorUserIDParams{
			UserID:        postgres.UUIDToPG(*viewerUserID),
			CreatorUserID: postgres.UUIDToPG(userID),
		})
		if err != nil {
			return PublicProfileHeader{}, fmt.Errorf("公開 creator profile follow relation 取得 creator=%q viewer=%s: %w", creatorID, *viewerUserID, err)
		}
	}

	return PublicProfileHeader{
		Profile:     profile,
		ShortCount:  shortCount,
		FanCount:    fanCount,
		IsFollowing: isFollowing,
	}, nil
}

// ListPublicProfileShorts は creator profile short grid の 1 page を返します。
func (r *Repository) ListPublicProfileShorts(ctx context.Context, creatorID string, cursor *PublicProfileShortCursor, limit int) ([]PublicProfileShort, *PublicProfileShortCursor, error) {
	userID, err := ParsePublicID(creatorID)
	if err != nil {
		return nil, nil, fmt.Errorf("公開 creator profile short grid 取得 creator=%q: %w", creatorID, ErrProfileNotFound)
	}

	params, pageLimit := buildPublicProfileShortPageParams(userID, cursor, limit)
	rows, err := r.queries.ListCreatorProfileShortGridItems(ctx, params)
	if err != nil {
		return nil, nil, fmt.Errorf("公開 creator profile short grid 取得 creator=%q: %w", creatorID, err)
	}
	if len(rows) == 0 {
		if _, err := r.GetPublicProfile(ctx, userID); err != nil {
			return nil, nil, fmt.Errorf("公開 creator profile short grid 取得 creator=%q: %w", creatorID, err)
		}

		return []PublicProfileShort{}, nil, nil
	}

	return mapPublicProfileShortPage(rows, pageLimit, fmt.Sprintf("公開 creator profile short grid 取得結果の変換 creator=%q", creatorID))
}

func buildPublicProfileShortPageParams(creatorUserID uuid.UUID, cursor *PublicProfileShortCursor, limit int) (sqlc.ListCreatorProfileShortGridItemsParams, int) {
	if limit <= 0 {
		limit = DefaultPublicProfileShortGridPageSize
	}

	params := sqlc.ListCreatorProfileShortGridItemsParams{
		CreatorUserID: postgres.UUIDToPG(creatorUserID),
		LimitCount:    int32(limit + 1),
	}
	if cursor == nil {
		return params, limit
	}

	params.CursorPublishedAt = postgres.TimeToPG(&cursor.PublishedAt)
	params.CursorShortID = postgres.UUIDToPG(cursor.ShortID)

	return params, limit
}

func mapPublicProfileShortPage(rows []sqlc.ListCreatorProfileShortGridItemsRow, limit int, label string) ([]PublicProfileShort, *PublicProfileShortCursor, error) {
	tiles := make([]PublicProfileShort, 0, min(limit, len(rows)))
	for index, row := range rows {
		if index >= limit {
			break
		}

		tile, err := mapPublicProfileShort(row)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", label, err)
		}

		tiles = append(tiles, tile)
	}

	if len(rows) <= limit {
		return tiles, nil, nil
	}

	lastTile := tiles[len(tiles)-1]

	return tiles, &PublicProfileShortCursor{
		PublishedAt: lastTile.PublishedAt,
		ShortID:     lastTile.ID,
	}, nil
}

func mapPublicProfileShort(row sqlc.ListCreatorProfileShortGridItemsRow) (PublicProfileShort, error) {
	shortID, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return PublicProfileShort{}, fmt.Errorf("公開 creator profile short の short id 変換: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return PublicProfileShort{}, fmt.Errorf("公開 creator profile short の creator user id 変換: %w", err)
	}
	canonicalMainID, err := postgres.UUIDFromPG(row.CanonicalMainID)
	if err != nil {
		return PublicProfileShort{}, fmt.Errorf("公開 creator profile short の canonical main id 変換: %w", err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return PublicProfileShort{}, fmt.Errorf("公開 creator profile short の media asset id 変換: %w", err)
	}
	publishedAt, err := postgres.RequiredTimeFromPG(row.PublishedAt)
	if err != nil {
		return PublicProfileShort{}, fmt.Errorf("公開 creator profile short の published_at 変換: %w", err)
	}

	if !row.PlaybackUrl.Valid || strings.TrimSpace(row.PlaybackUrl.String) == "" {
		return PublicProfileShort{}, fmt.Errorf("公開 creator profile short の playback url がありません")
	}
	if !row.DurationMs.Valid || row.DurationMs.Int64 <= 0 {
		return PublicProfileShort{}, fmt.Errorf("公開 creator profile short の duration_ms がありません")
	}

	return PublicProfileShort{
		ID:                     shortID,
		CreatorUserID:          creatorUserID,
		CanonicalMainID:        canonicalMainID,
		MediaAssetID:           mediaAssetID,
		MediaURL:               strings.TrimSpace(row.PlaybackUrl.String),
		PreviewDurationSeconds: (row.DurationMs.Int64 + 999) / 1000,
		PublishedAt:            publishedAt,
	}, nil
}
