package fanprofile

import (
	"context"
	"errors"
	"fmt"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const overviewTitle = "My archive"

// ErrProfileNotFound は対象の fan profile が存在しないことを表します。
var ErrProfileNotFound = errors.New("fan profile が見つかりません")

type queries interface {
	GetUserByID(ctx context.Context, id pgtype.UUID) (sqlc.AppUser, error)
	CountCreatorFollowsByUserID(ctx context.Context, userID pgtype.UUID) (int64, error)
	CountPinnedShortsByUserID(ctx context.Context, userID pgtype.UUID) (int64, error)
	CountUnlockedMainsByUserID(ctx context.Context, userID pgtype.UUID) (int64, error)
}

// Repository は fan profile overview 関連の永続化操作を包みます。
type Repository struct {
	queries queries
}

// OverviewCounts は fan profile overview の count 群を表します。
type OverviewCounts struct {
	Following    int64
	PinnedShorts int64
	Library      int64
}

// Overview は fan profile overview の read model を表します。
type Overview struct {
	Title  string
	Counts OverviewCounts
}

// NewRepository は pgxpool ベースの fan profile repository を構築します。
func NewRepository(pool *pgxpool.Pool) *Repository {
	return newRepository(sqlc.New(pool))
}

func newRepository(q queries) *Repository {
	return &Repository{queries: q}
}

// GetOverview は fan profile overview の counts-only payload を返します。
func (r *Repository) GetOverview(ctx context.Context, viewerUserID uuid.UUID) (Overview, error) {
	if r == nil || r.queries == nil {
		return Overview{}, fmt.Errorf("fan profile repository が初期化されていません")
	}

	viewerUserIDPG := postgres.UUIDToPG(viewerUserID)
	if _, err := r.queries.GetUserByID(ctx, viewerUserIDPG); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Overview{}, fmt.Errorf("fan profile overview 取得 user=%s: %w", viewerUserID, ErrProfileNotFound)
		}

		return Overview{}, fmt.Errorf("fan profile overview 取得 user=%s: %w", viewerUserID, err)
	}

	followingCount, err := r.queries.CountCreatorFollowsByUserID(ctx, viewerUserIDPG)
	if err != nil {
		return Overview{}, fmt.Errorf("fan profile overview following count 取得 user=%s: %w", viewerUserID, err)
	}

	pinnedShortsCount, err := r.queries.CountPinnedShortsByUserID(ctx, viewerUserIDPG)
	if err != nil {
		return Overview{}, fmt.Errorf("fan profile overview pinned shorts count 取得 user=%s: %w", viewerUserID, err)
	}

	libraryCount, err := r.queries.CountUnlockedMainsByUserID(ctx, viewerUserIDPG)
	if err != nil {
		return Overview{}, fmt.Errorf("fan profile overview library count 取得 user=%s: %w", viewerUserID, err)
	}

	return Overview{
		Title: overviewTitle,
		Counts: OverviewCounts{
			Following:    followingCount,
			PinnedShorts: pinnedShortsCount,
			Library:      libraryCount,
		},
	}, nil
}
