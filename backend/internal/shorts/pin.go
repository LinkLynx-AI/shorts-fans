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

// PinMutationResult は short pin mutation 後の relation state を表します。
type PinMutationResult struct {
	IsPinned bool
}

// PinPublicShort は public short への pin relation を作成します。
func (r *Repository) PinPublicShort(ctx context.Context, viewerUserID uuid.UUID, shortID uuid.UUID) (PinMutationResult, error) {
	result, err := r.mutatePublicShortPin(ctx, viewerUserID, shortID, true)
	if err != nil {
		return PinMutationResult{}, fmt.Errorf("short pin 更新 viewer=%s short=%s: %w", viewerUserID, shortID, err)
	}

	return result, nil
}

// UnpinPublicShort は public short への pin relation を削除します。
func (r *Repository) UnpinPublicShort(ctx context.Context, viewerUserID uuid.UUID, shortID uuid.UUID) (PinMutationResult, error) {
	result, err := r.mutatePublicShortPin(ctx, viewerUserID, shortID, false)
	if err != nil {
		return PinMutationResult{}, fmt.Errorf("short unpin 更新 viewer=%s short=%s: %w", viewerUserID, shortID, err)
	}

	return result, nil
}

func (r *Repository) mutatePublicShortPin(
	ctx context.Context,
	viewerUserID uuid.UUID,
	shortID uuid.UUID,
	shouldPin bool,
) (PinMutationResult, error) {
	if r == nil || r.beginner == nil {
		return PinMutationResult{}, fmt.Errorf("short repository pool が初期化されていません")
	}
	if r.newQueries == nil {
		return PinMutationResult{}, fmt.Errorf("short repository query factory が初期化されていません")
	}

	var result PinMutationResult
	err := postgres.RunInTx(ctx, r.beginner, func(tx pgx.Tx) error {
		q := r.newQueries(tx)
		shortIDPG := postgres.UUIDToPG(shortID)

		if _, err := q.GetPublicShortByID(ctx, shortIDPG); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrShortNotFound
			}

			return fmt.Errorf("public short 取得 short=%s: %w", shortID, err)
		}

		if shouldPin {
			if err := q.PutPinnedShort(ctx, sqlc.PutPinnedShortParams{
				UserID:  postgres.UUIDToPG(viewerUserID),
				ShortID: shortIDPG,
			}); err != nil {
				return fmt.Errorf("pinned short 作成 viewer=%s short=%s: %w", viewerUserID, shortID, err)
			}
		} else {
			if err := q.DeletePinnedShort(ctx, sqlc.DeletePinnedShortParams{
				UserID:  postgres.UUIDToPG(viewerUserID),
				ShortID: shortIDPG,
			}); err != nil {
				return fmt.Errorf("pinned short 削除 viewer=%s short=%s: %w", viewerUserID, shortID, err)
			}
		}

		result = PinMutationResult{
			IsPinned: shouldPin,
		}
		return nil
	})
	if err != nil {
		return PinMutationResult{}, err
	}

	return result, nil
}
