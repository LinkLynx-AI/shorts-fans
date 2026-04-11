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
)

// UpdateWorkspaceMainPrice は current viewer 自身の main price を更新します。
func (r *Repository) UpdateWorkspaceMainPrice(
	ctx context.Context,
	viewerUserID uuid.UUID,
	mainID uuid.UUID,
	priceJpy int64,
) error {
	if priceJpy <= 0 {
		return fmt.Errorf("creator workspace main price 更新 user=%s main=%s: price must be positive", viewerUserID, mainID)
	}
	if _, err := r.getApprovedWorkspaceProfile(ctx, viewerUserID); err != nil {
		return err
	}

	_, err := r.queries.UpdateCreatorWorkspaceMainPrice(ctx, sqlc.UpdateCreatorWorkspaceMainPriceParams{
		CreatorUserID: postgres.UUIDToPG(viewerUserID),
		ID:            postgres.UUIDToPG(mainID),
		PriceMinor:    priceJpy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf(
				"creator workspace main price 更新 user=%s main=%s: %w",
				viewerUserID,
				mainID,
				ErrWorkspacePreviewMainNotFound,
			)
		}

		return fmt.Errorf("creator workspace main price 更新 user=%s main=%s: %w", viewerUserID, mainID, err)
	}

	return nil
}

// UpdateWorkspaceShortCaption は current viewer 自身の short caption を更新します。
func (r *Repository) UpdateWorkspaceShortCaption(
	ctx context.Context,
	viewerUserID uuid.UUID,
	shortID uuid.UUID,
	caption *string,
) error {
	if _, err := r.getApprovedWorkspaceProfile(ctx, viewerUserID); err != nil {
		return err
	}

	normalizedCaption := normalizeWorkspaceCaptionInput(caption)
	_, err := r.queries.UpdateCreatorWorkspaceShortCaption(ctx, sqlc.UpdateCreatorWorkspaceShortCaptionParams{
		Caption:       postgres.TextToPG(normalizedCaption),
		CreatorUserID: postgres.UUIDToPG(viewerUserID),
		ID:            postgres.UUIDToPG(shortID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf(
				"creator workspace short caption 更新 user=%s short=%s: %w",
				viewerUserID,
				shortID,
				ErrWorkspacePreviewShortNotFound,
			)
		}

		return fmt.Errorf("creator workspace short caption 更新 user=%s short=%s: %w", viewerUserID, shortID, err)
	}

	return nil
}

func normalizeWorkspaceCaptionInput(caption *string) *string {
	if caption == nil {
		return nil
	}

	normalized := strings.TrimSpace(*caption)
	if normalized == "" {
		return nil
	}

	return &normalized
}
