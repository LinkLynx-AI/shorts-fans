package fanprofile

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultFollowingPageSize は fan profile following 一覧の既定 page size です。
const DefaultFollowingPageSize = 20

type queries interface {
	ListFanProfileFollowingItems(ctx context.Context, arg sqlc.ListFanProfileFollowingItemsParams) ([]sqlc.ListFanProfileFollowingItemsRow, error)
}

// Repository は fan profile private hub 向けの read 操作をまとめます。
type Repository struct {
	queries queries
}

// FollowingItem は private hub 上の followed creator row を表します。
type FollowingItem struct {
	AvatarURL     *string
	Bio           string
	CreatorUserID uuid.UUID
	DisplayName   string
	FollowedAt    time.Time
	Handle        string
}

// FollowingCursor は following 一覧の keyset cursor です。
type FollowingCursor struct {
	CreatorUserID uuid.UUID
	FollowedAt    time.Time
}

// NewRepository は pgxpool ベースの fan profile repository を構築します。
func NewRepository(pool *pgxpool.Pool) *Repository {
	return newRepository(sqlc.New(pool))
}

func newRepository(q queries) *Repository {
	return &Repository{queries: q}
}

// ListFollowing は authenticated viewer の followed creator 一覧 1 page を返します。
func (r *Repository) ListFollowing(ctx context.Context, viewerID uuid.UUID, cursor *FollowingCursor, limit int) ([]FollowingItem, *FollowingCursor, error) {
	params, pageLimit := buildFollowingPageParams(viewerID, cursor, limit)

	rows, err := r.queries.ListFanProfileFollowingItems(ctx, params)
	if err != nil {
		return nil, nil, fmt.Errorf("fan profile following 取得 viewer=%s: %w", viewerID, err)
	}

	return mapFollowingPage(rows, pageLimit, fmt.Sprintf("fan profile following 取得結果の変換 viewer=%s", viewerID))
}

func buildFollowingPageParams(viewerID uuid.UUID, cursor *FollowingCursor, limit int) (sqlc.ListFanProfileFollowingItemsParams, int) {
	if limit <= 0 {
		limit = DefaultFollowingPageSize
	}

	params := sqlc.ListFanProfileFollowingItemsParams{
		UserID:     postgres.UUIDToPG(viewerID),
		LimitCount: int32(limit + 1),
	}
	if cursor == nil {
		return params, limit
	}

	params.CursorFollowedAt = postgres.TimeToPG(&cursor.FollowedAt)
	params.CursorCreatorUserID = postgres.UUIDToPG(cursor.CreatorUserID)

	return params, limit
}

func mapFollowingPage(rows []sqlc.ListFanProfileFollowingItemsRow, limit int, label string) ([]FollowingItem, *FollowingCursor, error) {
	items := make([]FollowingItem, 0, min(limit, len(rows)))
	for index, row := range rows {
		if index >= limit {
			break
		}

		item, err := mapFollowingItem(row)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", label, err)
		}

		items = append(items, item)
	}

	if len(rows) <= limit {
		if len(items) == 0 {
			return []FollowingItem{}, nil, nil
		}

		return items, nil, nil
	}

	lastItem := items[len(items)-1]
	return items, &FollowingCursor{
		CreatorUserID: lastItem.CreatorUserID,
		FollowedAt:    lastItem.FollowedAt,
	}, nil
}

func mapFollowingItem(row sqlc.ListFanProfileFollowingItemsRow) (FollowingItem, error) {
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return FollowingItem{}, fmt.Errorf("followed creator user id 変換: %w", err)
	}

	followedAt, err := postgres.RequiredTimeFromPG(row.FollowedAt)
	if err != nil {
		return FollowingItem{}, fmt.Errorf("followed_at 変換: %w", err)
	}

	if !row.DisplayName.Valid || strings.TrimSpace(row.DisplayName.String) == "" {
		return FollowingItem{}, fmt.Errorf("followed creator display_name がありません")
	}
	if !row.Handle.Valid || strings.TrimSpace(row.Handle.String) == "" {
		return FollowingItem{}, fmt.Errorf("followed creator handle がありません")
	}

	return FollowingItem{
		AvatarURL:     postgres.OptionalTextFromPG(row.AvatarUrl),
		Bio:           row.Bio,
		CreatorUserID: creatorUserID,
		DisplayName:   strings.TrimSpace(row.DisplayName.String),
		FollowedAt:    followedAt,
		Handle:        strings.TrimSpace(row.Handle.String),
	}, nil
}
