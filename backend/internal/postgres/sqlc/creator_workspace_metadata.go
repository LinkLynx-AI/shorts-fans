package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const updateCreatorWorkspaceMainPrice = `-- name: UpdateCreatorWorkspaceMainPrice :one
UPDATE app.mains
SET
    price_minor = $1,
    currency_code = 'JPY',
    updated_at = CURRENT_TIMESTAMP
WHERE id = $2
  AND creator_user_id = $3
RETURNING id, creator_user_id, media_asset_id, state, review_reason_code, post_report_state, price_minor, currency_code, ownership_confirmed, consent_confirmed, approved_for_unlock_at, created_at, updated_at
`

type UpdateCreatorWorkspaceMainPriceParams struct {
	PriceMinor    int64
	ID            pgtype.UUID
	CreatorUserID pgtype.UUID
}

func (q *Queries) UpdateCreatorWorkspaceMainPrice(
	ctx context.Context,
	arg UpdateCreatorWorkspaceMainPriceParams,
) (AppMain, error) {
	row := q.db.QueryRow(ctx, updateCreatorWorkspaceMainPrice, arg.PriceMinor, arg.ID, arg.CreatorUserID)
	var i AppMain
	err := row.Scan(
		&i.ID,
		&i.CreatorUserID,
		&i.MediaAssetID,
		&i.State,
		&i.ReviewReasonCode,
		&i.PostReportState,
		&i.PriceMinor,
		&i.CurrencyCode,
		&i.OwnershipConfirmed,
		&i.ConsentConfirmed,
		&i.ApprovedForUnlockAt,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateCreatorWorkspaceShortCaption = `-- name: UpdateCreatorWorkspaceShortCaption :one
UPDATE app.shorts
SET
    caption = $1,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $2
  AND creator_user_id = $3
RETURNING id, creator_user_id, canonical_main_id, media_asset_id, state, review_reason_code, post_report_state, approved_for_publish_at, published_at, created_at, updated_at, caption
`

type UpdateCreatorWorkspaceShortCaptionParams struct {
	Caption       pgtype.Text
	ID            pgtype.UUID
	CreatorUserID pgtype.UUID
}

func (q *Queries) UpdateCreatorWorkspaceShortCaption(
	ctx context.Context,
	arg UpdateCreatorWorkspaceShortCaptionParams,
) (AppShort, error) {
	row := q.db.QueryRow(ctx, updateCreatorWorkspaceShortCaption, arg.Caption, arg.ID, arg.CreatorUserID)
	var i AppShort
	err := row.Scan(
		&i.ID,
		&i.CreatorUserID,
		&i.CanonicalMainID,
		&i.MediaAssetID,
		&i.State,
		&i.ReviewReasonCode,
		&i.PostReportState,
		&i.ApprovedForPublishAt,
		&i.PublishedAt,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Caption,
	)
	return i, err
}
