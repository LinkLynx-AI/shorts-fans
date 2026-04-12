package creator

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// WorkspaceShortCaptionMutationResult は workspace short caption 更新結果です。
type WorkspaceShortCaptionMutationResult struct {
	Caption string
	ShortID uuid.UUID
}

// UpdateWorkspaceShortCaption は current viewer 自身の short caption を更新します。
func (r *Repository) UpdateWorkspaceShortCaption(
	ctx context.Context,
	viewerUserID uuid.UUID,
	shortID uuid.UUID,
	caption string,
) (WorkspaceShortCaptionMutationResult, error) {
	if _, err := r.getApprovedWorkspaceProfile(ctx, viewerUserID); err != nil {
		return WorkspaceShortCaptionMutationResult{}, err
	}
	if r.delivery == nil {
		return WorkspaceShortCaptionMutationResult{}, fmt.Errorf(
			"creator workspace short caption 更新 user=%s short=%s: delivery is nil",
			viewerUserID,
			shortID,
		)
	}

	row, err := r.queries.GetShortByID(ctx, postgres.UUIDToPG(shortID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return WorkspaceShortCaptionMutationResult{}, fmt.Errorf(
				"creator workspace short caption 更新 user=%s short=%s: %w",
				viewerUserID,
				shortID,
				ErrWorkspacePreviewNotFound,
			)
		}

		return WorkspaceShortCaptionMutationResult{}, fmt.Errorf(
			"creator workspace short caption 更新 user=%s short=%s short 取得: %w",
			viewerUserID,
			shortID,
			err,
		)
	}

	rowCreatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return WorkspaceShortCaptionMutationResult{}, fmt.Errorf(
			"creator workspace short caption 更新 user=%s short=%s creator 変換: %w",
			viewerUserID,
			shortID,
			err,
		)
	}
	if rowCreatorUserID != viewerUserID {
		return WorkspaceShortCaptionMutationResult{}, fmt.Errorf(
			"creator workspace short caption 更新 user=%s short=%s: %w",
			viewerUserID,
			shortID,
			ErrWorkspacePreviewNotFound,
		)
	}

	assetCache := map[uuid.UUID]sqlc.AppMediaAsset{}
	if _, ok, buildErr := r.resolveWorkspacePreviewShortSource(ctx, row, assetCache); buildErr != nil {
		return WorkspaceShortCaptionMutationResult{}, fmt.Errorf(
			"creator workspace short caption 更新 user=%s short=%s preview 判定: %w",
			viewerUserID,
			shortID,
			buildErr,
		)
	} else if !ok {
		return WorkspaceShortCaptionMutationResult{}, fmt.Errorf(
			"creator workspace short caption 更新 user=%s short=%s: %w",
			viewerUserID,
			shortID,
			ErrWorkspacePreviewNotFound,
		)
	}

	updatedRow, err := r.queries.UpdateShortCaption(ctx, sqlc.UpdateShortCaptionParams{
		Caption: postgres.TextToPG(normalizeWorkspaceShortCaption(caption)),
		ID:      postgres.UUIDToPG(shortID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return WorkspaceShortCaptionMutationResult{}, fmt.Errorf(
				"creator workspace short caption 更新 user=%s short=%s: %w",
				viewerUserID,
				shortID,
				ErrWorkspacePreviewNotFound,
			)
		}

		return WorkspaceShortCaptionMutationResult{}, fmt.Errorf(
			"creator workspace short caption 更新 user=%s short=%s: %w",
			viewerUserID,
			shortID,
			err,
		)
	}

	return WorkspaceShortCaptionMutationResult{
		Caption: resolveWorkspaceShortCaption(updatedRow.Caption),
		ShortID: shortID,
	}, nil
}

func normalizeWorkspaceShortCaption(caption string) *string {
	normalizedCaption := strings.TrimSpace(caption)
	if normalizedCaption == "" {
		return nil
	}

	return &normalizedCaption
}

func resolveWorkspaceShortCaption(value pgtype.Text) string {
	if caption := postgres.OptionalTextFromPG(value); caption != nil {
		return strings.TrimSpace(*caption)
	}

	return ""
}
