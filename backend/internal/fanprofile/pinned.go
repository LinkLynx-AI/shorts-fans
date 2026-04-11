package fanprofile

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
)

// DefaultPinnedShortsPageSize は fan profile pinned shorts 一覧の既定 page size です。
const DefaultPinnedShortsPageSize = 18

// PinnedShortItem は private hub 上の pinned short row を表します。
type PinnedShortItem struct {
	CreatorAvatarURL            *string
	CreatorBio                  string
	CreatorDisplayName          string
	CreatorHandle               string
	CreatorUserID               uuid.UUID
	PinnedAt                    time.Time
	ShortCaption                string
	ShortCanonicalMainID        uuid.UUID
	ShortID                     uuid.UUID
	ShortMediaAssetID           uuid.UUID
	ShortPreviewDurationSeconds int64
}

// PinnedShortCursor は pinned shorts 一覧の keyset cursor です。
type PinnedShortCursor struct {
	PinnedAt time.Time
	ShortID  uuid.UUID
}

// ListPinnedShorts は authenticated viewer の pinned short 一覧 1 page を返します。
func (r *Repository) ListPinnedShorts(
	ctx context.Context,
	viewerID uuid.UUID,
	cursor *PinnedShortCursor,
	limit int,
) ([]PinnedShortItem, *PinnedShortCursor, error) {
	params, pageLimit := buildPinnedShortPageParams(viewerID, cursor, limit)

	rows, err := r.queries.ListFanProfilePinnedShortItems(ctx, params)
	if err != nil {
		return nil, nil, fmt.Errorf("fan profile pinned shorts 取得 viewer=%s: %w", viewerID, err)
	}

	return mapPinnedShortPage(rows, pageLimit, fmt.Sprintf("fan profile pinned shorts 取得結果の変換 viewer=%s", viewerID))
}

func buildPinnedShortPageParams(
	viewerID uuid.UUID,
	cursor *PinnedShortCursor,
	limit int,
) (sqlc.ListFanProfilePinnedShortItemsParams, int) {
	if limit <= 0 {
		limit = DefaultPinnedShortsPageSize
	}

	params := sqlc.ListFanProfilePinnedShortItemsParams{
		UserID:     postgres.UUIDToPG(viewerID),
		LimitCount: int32(limit + 1),
	}
	if cursor == nil {
		return params, limit
	}

	params.CursorPinnedAt = postgres.TimeToPG(&cursor.PinnedAt)
	params.CursorShortID = postgres.UUIDToPG(cursor.ShortID)

	return params, limit
}

func mapPinnedShortPage(
	rows []sqlc.ListFanProfilePinnedShortItemsRow,
	limit int,
	label string,
) ([]PinnedShortItem, *PinnedShortCursor, error) {
	items := make([]PinnedShortItem, 0, min(limit, len(rows)))
	for index, row := range rows {
		if index >= limit {
			break
		}

		item, err := mapPinnedShortItem(row)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", label, err)
		}

		items = append(items, item)
	}

	if len(rows) <= limit {
		if len(items) == 0 {
			return []PinnedShortItem{}, nil, nil
		}

		return items, nil, nil
	}

	lastItem := items[len(items)-1]

	return items, &PinnedShortCursor{
		PinnedAt: lastItem.PinnedAt,
		ShortID:  lastItem.ShortID,
	}, nil
}

func mapPinnedShortItem(row sqlc.ListFanProfilePinnedShortItemsRow) (PinnedShortItem, error) {
	shortID, err := postgres.UUIDFromPG(row.ShortID)
	if err != nil {
		return PinnedShortItem{}, fmt.Errorf("pinned short id 変換: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return PinnedShortItem{}, fmt.Errorf("pinned short creator user id 変換: %w", err)
	}
	canonicalMainID, err := postgres.UUIDFromPG(row.CanonicalMainID)
	if err != nil {
		return PinnedShortItem{}, fmt.Errorf("pinned short canonical main id 変換: %w", err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return PinnedShortItem{}, fmt.Errorf("pinned short media asset id 変換: %w", err)
	}
	pinnedAt, err := postgres.RequiredTimeFromPG(row.PinnedAt)
	if err != nil {
		return PinnedShortItem{}, fmt.Errorf("pinned short pinned_at 変換: %w", err)
	}

	if !row.DisplayName.Valid || strings.TrimSpace(row.DisplayName.String) == "" {
		return PinnedShortItem{}, fmt.Errorf("pinned short creator display_name がありません")
	}
	if strings.TrimSpace(row.Handle) == "" {
		return PinnedShortItem{}, fmt.Errorf("pinned short creator handle がありません")
	}
	if !row.DurationMs.Valid || row.DurationMs.Int64 <= 0 {
		return PinnedShortItem{}, fmt.Errorf("pinned short duration_ms がありません")
	}

	caption := ""
	if row.Caption.Valid {
		caption = strings.TrimSpace(row.Caption.String)
	}

	return PinnedShortItem{
		CreatorAvatarURL:            postgres.OptionalTextFromPG(row.AvatarUrl),
		CreatorBio:                  row.Bio,
		CreatorDisplayName:          strings.TrimSpace(row.DisplayName.String),
		CreatorHandle:               strings.TrimSpace(row.Handle),
		CreatorUserID:               creatorUserID,
		PinnedAt:                    pinnedAt,
		ShortCaption:                caption,
		ShortCanonicalMainID:        canonicalMainID,
		ShortID:                     shortID,
		ShortMediaAssetID:           mediaAssetID,
		ShortPreviewDurationSeconds: (row.DurationMs.Int64 + 999) / 1000,
	}, nil
}
