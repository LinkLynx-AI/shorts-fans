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

const workspaceMainPriceCurrencyCode = "JPY"

// ErrInvalidWorkspaceMainPrice は creator workspace の本編価格が不正なことを表します。
var ErrInvalidWorkspaceMainPrice = errors.New("creator workspace main price is invalid")

// ErrWorkspaceMainNotFound は creator workspace で更新対象の本編を解決できないことを表します。
var ErrWorkspaceMainNotFound = errors.New("creator workspace main was not found")

// WorkspaceMainPrice は creator workspace 本編価格更新結果の read model です。
type WorkspaceMainPrice struct {
	ID       uuid.UUID
	PriceJpy int64
}

// UpdateWorkspaceMainPrice は current viewer 自身の main price を更新します。
func (r *Repository) UpdateWorkspaceMainPrice(
	ctx context.Context,
	viewerUserID uuid.UUID,
	mainID uuid.UUID,
	priceJpy int64,
) (WorkspaceMainPrice, error) {
	if priceJpy <= 0 {
		return WorkspaceMainPrice{}, fmt.Errorf(
			"creator workspace main price 更新 user=%s main=%s: %w",
			viewerUserID,
			mainID,
			ErrInvalidWorkspaceMainPrice,
		)
	}

	if _, err := r.getApprovedWorkspaceProfile(ctx, viewerUserID); err != nil {
		return WorkspaceMainPrice{}, err
	}

	row, err := r.queries.UpdateCreatorWorkspaceMainPrice(ctx, sqlc.UpdateCreatorWorkspaceMainPriceParams{
		MainID:      postgres.UUIDToPG(mainID),
		OwnerUserID: postgres.UUIDToPG(viewerUserID),
		PriceMinor:  priceJpy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return WorkspaceMainPrice{}, fmt.Errorf(
				"creator workspace main price 更新 user=%s main=%s: %w",
				viewerUserID,
				mainID,
				ErrWorkspaceMainNotFound,
			)
		}

		return WorkspaceMainPrice{}, fmt.Errorf("creator workspace main price 更新 user=%s main=%s: %w", viewerUserID, mainID, err)
	}

	updatedMainID, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return WorkspaceMainPrice{}, fmt.Errorf("creator workspace main price 更新 user=%s main=%s id 変換: %w", viewerUserID, mainID, err)
	}

	if row.CurrencyCode != workspaceMainPriceCurrencyCode {
		return WorkspaceMainPrice{}, fmt.Errorf(
			"creator workspace main price 更新 user=%s main=%s unexpected currency=%s",
			viewerUserID,
			mainID,
			row.CurrencyCode,
		)
	}

	return WorkspaceMainPrice{
		ID:       updatedMainID,
		PriceJpy: row.PriceMinor,
	}, nil
}
