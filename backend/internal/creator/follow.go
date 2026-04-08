package creator

import (
	"context"
	"errors"
	"fmt"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// FollowMutationResult は creator follow mutation 後の relation state を表します。
type FollowMutationResult struct {
	FanCount    int64
	IsFollowing bool
}

// FollowPublicCreator は public creator への follow relation を作成します。
func (r *Repository) FollowPublicCreator(ctx context.Context, viewerUserID uuid.UUID, creatorID string) (FollowMutationResult, error) {
	result, err := r.mutatePublicCreatorFollow(ctx, viewerUserID, creatorID, true)
	if err != nil {
		return FollowMutationResult{}, fmt.Errorf("creator follow 更新 viewer=%s creator=%q: %w", viewerUserID, creatorID, err)
	}

	return result, nil
}

// UnfollowPublicCreator は public creator への follow relation を削除します。
func (r *Repository) UnfollowPublicCreator(ctx context.Context, viewerUserID uuid.UUID, creatorID string) (FollowMutationResult, error) {
	result, err := r.mutatePublicCreatorFollow(ctx, viewerUserID, creatorID, false)
	if err != nil {
		return FollowMutationResult{}, fmt.Errorf("creator unfollow 更新 viewer=%s creator=%q: %w", viewerUserID, creatorID, err)
	}

	return result, nil
}

func (r *Repository) mutatePublicCreatorFollow(ctx context.Context, viewerUserID uuid.UUID, creatorID string, shouldFollow bool) (FollowMutationResult, error) {
	if r == nil || r.txBeginner == nil {
		return FollowMutationResult{}, fmt.Errorf("creator repository pool が初期化されていません")
	}
	if r.newQueries == nil {
		return FollowMutationResult{}, fmt.Errorf("creator repository query factory が初期化されていません")
	}

	var result FollowMutationResult
	err := postgres.RunInTx(ctx, r.txBeginner, func(tx pgx.Tx) error {
		q := r.newQueries(tx)

		creatorUserID, err := getPublicCreatorUserID(ctx, q, creatorID)
		if err != nil {
			return err
		}

		if shouldFollow {
			if err := q.PutCreatorFollow(ctx, sqlc.PutCreatorFollowParams{
				UserID:        postgres.UUIDToPG(viewerUserID),
				CreatorUserID: postgres.UUIDToPG(creatorUserID),
			}); err != nil {
				return fmt.Errorf("creator follow 作成 viewer=%s creator=%s: %w", viewerUserID, creatorUserID, err)
			}
		} else {
			if err := q.DeleteCreatorFollow(ctx, sqlc.DeleteCreatorFollowParams{
				UserID:        postgres.UUIDToPG(viewerUserID),
				CreatorUserID: postgres.UUIDToPG(creatorUserID),
			}); err != nil {
				return fmt.Errorf("creator follow 削除 viewer=%s creator=%s: %w", viewerUserID, creatorUserID, err)
			}
		}

		fanCount, err := q.CountCreatorFollowersByCreatorUserID(ctx, postgres.UUIDToPG(creatorUserID))
		if err != nil {
			return fmt.Errorf("creator follower count 取得 creator=%s: %w", creatorUserID, err)
		}

		result = FollowMutationResult{
			FanCount:    fanCount,
			IsFollowing: shouldFollow,
		}
		return nil
	})
	if err != nil {
		return FollowMutationResult{}, err
	}

	return result, nil
}

func getPublicCreatorUserID(ctx context.Context, q queries, creatorID string) (uuid.UUID, error) {
	creatorUserID, err := ParsePublicID(creatorID)
	if err != nil {
		return uuid.Nil, ErrProfileNotFound
	}

	if _, err := q.GetPublicCreatorProfileByUserID(ctx, postgres.UUIDToPG(creatorUserID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrProfileNotFound
		}

		return uuid.Nil, fmt.Errorf("公開 creator profile 取得 user=%s: %w", creatorUserID, err)
	}

	return creatorUserID, nil
}
