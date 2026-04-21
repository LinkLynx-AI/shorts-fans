package shorts

import (
	"context"
	"errors"
	"fmt"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// LikeMutationResult は short like mutation 後の relation state と total count を表します。
type LikeMutationResult struct {
	HasLiked  bool
	LikeCount int64
}

// LikePublicShort は public short への like relation を作成します。
func (r *Repository) LikePublicShort(ctx context.Context, viewerUserID uuid.UUID, shortID uuid.UUID) (LikeMutationResult, error) {
	result, err := r.mutatePublicShortLike(ctx, viewerUserID, shortID, true)
	if err != nil {
		return LikeMutationResult{}, fmt.Errorf("short like 更新 viewer=%s short=%s: %w", viewerUserID, shortID, err)
	}

	return result, nil
}

// UnlikePublicShort は public short への like relation を削除します。
func (r *Repository) UnlikePublicShort(ctx context.Context, viewerUserID uuid.UUID, shortID uuid.UUID) (LikeMutationResult, error) {
	result, err := r.mutatePublicShortLike(ctx, viewerUserID, shortID, false)
	if err != nil {
		return LikeMutationResult{}, fmt.Errorf("short unlike 更新 viewer=%s short=%s: %w", viewerUserID, shortID, err)
	}

	return result, nil
}

func (r *Repository) mutatePublicShortLike(
	ctx context.Context,
	viewerUserID uuid.UUID,
	shortID uuid.UUID,
	shouldLike bool,
) (LikeMutationResult, error) {
	if r == nil || r.beginner == nil {
		return LikeMutationResult{}, fmt.Errorf("short repository pool が初期化されていません")
	}
	if r.newQueries == nil {
		return LikeMutationResult{}, fmt.Errorf("short repository query factory が初期化されていません")
	}

	var result LikeMutationResult
	err := postgres.RunInTx(ctx, r.beginner, func(tx pgx.Tx) error {
		q := r.newQueries(tx)
		shortIDPG := postgres.UUIDToPG(shortID)

		if _, err := q.GetPublicShortByID(ctx, shortIDPG); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrShortNotFound
			}

			return fmt.Errorf("public short 取得 short=%s: %w", shortID, err)
		}

		if shouldLike {
			if err := q.PutShortLike(ctx, sqlc.PutShortLikeParams{
				UserID:  postgres.UUIDToPG(viewerUserID),
				ShortID: shortIDPG,
			}); err != nil {
				return fmt.Errorf("short like 作成 viewer=%s short=%s: %w", viewerUserID, shortID, err)
			}
		} else {
			if err := q.DeleteShortLike(ctx, sqlc.DeleteShortLikeParams{
				UserID:  postgres.UUIDToPG(viewerUserID),
				ShortID: shortIDPG,
			}); err != nil {
				return fmt.Errorf("short like 削除 viewer=%s short=%s: %w", viewerUserID, shortID, err)
			}
		}

		likeCount, err := q.CountShortLikesByShortID(ctx, shortIDPG)
		if err != nil {
			return fmt.Errorf("short like count 取得 short=%s: %w", shortID, err)
		}

		result = LikeMutationResult{
			HasLiked:  shouldLike,
			LikeCount: likeCount,
		}
		return nil
	})
	if err != nil {
		return LikeMutationResult{}, err
	}

	return result, nil
}
