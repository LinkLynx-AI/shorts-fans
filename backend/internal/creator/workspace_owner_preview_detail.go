package creator

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// ErrWorkspacePreviewMainNotFound は対象の owner preview main を解決できないことを表します。
var ErrWorkspacePreviewMainNotFound = errors.New("workspace preview main が見つかりません")

// ErrWorkspacePreviewShortNotFound は対象の owner preview short を解決できないことを表します。
var ErrWorkspacePreviewShortNotFound = errors.New("workspace preview short が見つかりません")

// WorkspacePreviewShortDetailItem は owner preview short detail の read model です。
type WorkspacePreviewShortDetailItem struct {
	Caption                string
	CanonicalMainID        uuid.UUID
	CreatorUserID          uuid.UUID
	ID                     uuid.UUID
	Media                  media.VideoDisplayAsset
	PreviewDurationSeconds int64
}

// WorkspacePreviewMainDetailItem は owner preview main detail の read model です。
type WorkspacePreviewMainDetailItem struct {
	DurationSeconds int64
	ID              uuid.UUID
	Media           media.VideoDisplayAsset
	PriceJpy        int64
}

// WorkspacePreviewShortDetail は owner preview short detail payload を表します。
type WorkspacePreviewShortDetail struct {
	Creator Profile
	Short   WorkspacePreviewShortDetailItem
}

// WorkspacePreviewMainDetail は owner preview main detail payload を表します。
type WorkspacePreviewMainDetail struct {
	Creator    Profile
	EntryShort WorkspacePreviewShortDetailItem
	Main       WorkspacePreviewMainDetailItem
}

// GetWorkspacePreviewShortDetail は current viewer 自身の owner preview short detail を返します。
func (r *Repository) GetWorkspacePreviewShortDetail(
	ctx context.Context,
	viewerUserID uuid.UUID,
	shortID uuid.UUID,
) (WorkspacePreviewShortDetail, error) {
	profile, err := r.getApprovedWorkspaceProfile(ctx, viewerUserID)
	if err != nil {
		return WorkspacePreviewShortDetail{}, err
	}
	if r.delivery == nil {
		return WorkspacePreviewShortDetail{}, fmt.Errorf("creator workspace short preview detail 取得 user=%s: delivery is nil", viewerUserID)
	}

	row, err := r.queries.GetShortByID(ctx, postgres.UUIDToPG(shortID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return WorkspacePreviewShortDetail{}, fmt.Errorf(
				"creator workspace short preview detail 取得 user=%s short=%s: %w",
				viewerUserID,
				shortID,
				ErrWorkspacePreviewShortNotFound,
			)
		}

		return WorkspacePreviewShortDetail{}, fmt.Errorf(
			"creator workspace short preview detail 取得 user=%s short=%s: %w",
			viewerUserID,
			shortID,
			err,
		)
	}

	assetCache := map[uuid.UUID]sqlc.AppMediaAsset{}
	shortDetail, ok, err := r.buildWorkspacePreviewShortDetailItem(ctx, row, assetCache)
	if err != nil {
		return WorkspacePreviewShortDetail{}, fmt.Errorf(
			"creator workspace short preview detail 取得 user=%s short=%s: %w",
			viewerUserID,
			shortID,
			err,
		)
	}
	if !ok || shortDetail.CreatorUserID != viewerUserID {
		return WorkspacePreviewShortDetail{}, fmt.Errorf(
			"creator workspace short preview detail 取得 user=%s short=%s: %w",
			viewerUserID,
			shortID,
			ErrWorkspacePreviewShortNotFound,
		)
	}

	return WorkspacePreviewShortDetail{
		Creator: profile,
		Short:   shortDetail,
	}, nil
}

// GetWorkspacePreviewMainDetail は current viewer 自身の owner preview main detail を返します。
func (r *Repository) GetWorkspacePreviewMainDetail(
	ctx context.Context,
	viewerUserID uuid.UUID,
	mainID uuid.UUID,
) (WorkspacePreviewMainDetail, error) {
	profile, err := r.getApprovedWorkspaceProfile(ctx, viewerUserID)
	if err != nil {
		return WorkspacePreviewMainDetail{}, err
	}
	if r.delivery == nil {
		return WorkspacePreviewMainDetail{}, fmt.Errorf("creator workspace main preview detail 取得 user=%s: delivery is nil", viewerUserID)
	}

	row, err := r.queries.GetMainByID(ctx, postgres.UUIDToPG(mainID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return WorkspacePreviewMainDetail{}, fmt.Errorf(
				"creator workspace main preview detail 取得 user=%s main=%s: %w",
				viewerUserID,
				mainID,
				ErrWorkspacePreviewMainNotFound,
			)
		}

		return WorkspacePreviewMainDetail{}, fmt.Errorf(
			"creator workspace main preview detail 取得 user=%s main=%s: %w",
			viewerUserID,
			mainID,
			err,
		)
	}

	mainDetail, ok, err := r.buildWorkspacePreviewMainDetailItem(ctx, row, map[uuid.UUID]sqlc.AppMediaAsset{})
	if err != nil {
		return WorkspacePreviewMainDetail{}, fmt.Errorf(
			"creator workspace main preview detail 取得 user=%s main=%s: %w",
			viewerUserID,
			mainID,
			err,
		)
	}
	if !ok {
		return WorkspacePreviewMainDetail{}, fmt.Errorf(
			"creator workspace main preview detail 取得 user=%s main=%s: %w",
			viewerUserID,
			mainID,
			ErrWorkspacePreviewMainNotFound,
		)
	}

	mainOwnerUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return WorkspacePreviewMainDetail{}, fmt.Errorf(
			"creator workspace main preview detail 取得 user=%s main=%s owner 変換: %w",
			viewerUserID,
			mainID,
			err,
		)
	}
	if mainOwnerUserID != viewerUserID {
		return WorkspacePreviewMainDetail{}, fmt.Errorf(
			"creator workspace main preview detail 取得 user=%s main=%s: %w",
			viewerUserID,
			mainID,
			ErrWorkspacePreviewMainNotFound,
		)
	}

	shortRows, err := r.queries.ListShortsByCreatorUserID(ctx, postgres.UUIDToPG(viewerUserID))
	if err != nil {
		return WorkspacePreviewMainDetail{}, fmt.Errorf(
			"creator workspace main preview detail 取得 user=%s main=%s linked short 読み込み: %w",
			viewerUserID,
			mainID,
			err,
		)
	}

	entryShort, ok, err := r.findWorkspacePreviewEntryShortDetail(ctx, mainID, shortRows, map[uuid.UUID]sqlc.AppMediaAsset{})
	if err != nil {
		return WorkspacePreviewMainDetail{}, fmt.Errorf(
			"creator workspace main preview detail 取得 user=%s main=%s entry short 解決: %w",
			viewerUserID,
			mainID,
			err,
		)
	}
	if !ok {
		return WorkspacePreviewMainDetail{}, fmt.Errorf(
			"creator workspace main preview detail 取得 user=%s main=%s: %w",
			viewerUserID,
			mainID,
			ErrWorkspacePreviewMainNotFound,
		)
	}

	return WorkspacePreviewMainDetail{
		Creator:    profile,
		EntryShort: entryShort,
		Main:       mainDetail,
	}, nil
}

func (r *Repository) buildWorkspacePreviewShortDetailItem(
	ctx context.Context,
	row sqlc.AppShort,
	assetCache map[uuid.UUID]sqlc.AppMediaAsset,
) (WorkspacePreviewShortDetailItem, bool, error) {
	shortID, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return WorkspacePreviewShortDetailItem{}, false, fmt.Errorf("workspace preview short detail id 変換: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return WorkspacePreviewShortDetailItem{}, false, fmt.Errorf("workspace preview short detail creator user id 変換 short=%s: %w", shortID, err)
	}
	canonicalMainID, err := postgres.UUIDFromPG(row.CanonicalMainID)
	if err != nil {
		return WorkspacePreviewShortDetailItem{}, false, fmt.Errorf("workspace preview short detail canonical main id 変換 short=%s: %w", shortID, err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return WorkspacePreviewShortDetailItem{}, false, fmt.Errorf("workspace preview short detail media asset id 変換 short=%s: %w", shortID, err)
	}

	asset, err := r.getWorkspacePreviewAsset(ctx, mediaAssetID, assetCache)
	if err != nil {
		return WorkspacePreviewShortDetailItem{}, false, fmt.Errorf("workspace preview short detail asset 解決 short=%s: %w", shortID, err)
	}
	if !workspacePreviewAssetReady(asset) {
		return WorkspacePreviewShortDetailItem{}, false, nil
	}
	durationMS := postgres.OptionalInt64FromPG(asset.DurationMs)
	if durationMS == nil {
		return WorkspacePreviewShortDetailItem{}, false, fmt.Errorf("workspace preview short detail duration がありません short=%s", shortID)
	}

	displayAsset, err := r.delivery.ResolveShortDisplayAsset(media.ShortDisplaySource{
		AssetID:    mediaAssetID,
		ShortID:    shortID,
		DurationMS: *durationMS,
	}, media.AccessBoundaryPublic)
	if err != nil {
		return WorkspacePreviewShortDetailItem{}, false, fmt.Errorf("workspace preview short detail display asset 解決 short=%s: %w", shortID, err)
	}

	return WorkspacePreviewShortDetailItem{
		Caption:                normalizeWorkspacePreviewCaption(row.Caption),
		CanonicalMainID:        canonicalMainID,
		CreatorUserID:          creatorUserID,
		ID:                     shortID,
		Media:                  displayAsset,
		PreviewDurationSeconds: displayAsset.DurationSeconds,
	}, true, nil
}

func (r *Repository) buildWorkspacePreviewMainDetailItem(
	ctx context.Context,
	row sqlc.AppMain,
	assetCache map[uuid.UUID]sqlc.AppMediaAsset,
) (WorkspacePreviewMainDetailItem, bool, error) {
	mainID, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return WorkspacePreviewMainDetailItem{}, false, fmt.Errorf("workspace preview main detail id 変換: %w", err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return WorkspacePreviewMainDetailItem{}, false, fmt.Errorf("workspace preview main detail media asset id 変換 main=%s: %w", mainID, err)
	}
	if strings.TrimSpace(row.CurrencyCode) != workspacePreviewCurrencyCode {
		return WorkspacePreviewMainDetailItem{}, false, nil
	}

	asset, err := r.getWorkspacePreviewAsset(ctx, mediaAssetID, assetCache)
	if err != nil {
		return WorkspacePreviewMainDetailItem{}, false, fmt.Errorf("workspace preview main detail asset 解決 main=%s: %w", mainID, err)
	}
	if !workspacePreviewAssetReady(asset) {
		return WorkspacePreviewMainDetailItem{}, false, nil
	}
	durationMS := postgres.OptionalInt64FromPG(asset.DurationMs)
	if durationMS == nil {
		return WorkspacePreviewMainDetailItem{}, false, fmt.Errorf("workspace preview main detail duration がありません main=%s", mainID)
	}

	displayAsset, err := r.delivery.ResolveMainDisplayAsset(ctx, media.MainDisplaySource{
		AssetID:    mediaAssetID,
		MainID:     mainID,
		DurationMS: *durationMS,
	}, media.AccessBoundaryOwner, media.DefaultSignedURLTTL)
	if err != nil {
		return WorkspacePreviewMainDetailItem{}, false, fmt.Errorf("workspace preview main detail display asset 解決 main=%s: %w", mainID, err)
	}

	return WorkspacePreviewMainDetailItem{
		DurationSeconds: displayAsset.DurationSeconds,
		ID:              mainID,
		Media:           displayAsset,
		PriceJpy:        row.PriceMinor,
	}, true, nil
}

func (r *Repository) findWorkspacePreviewEntryShortDetail(
	ctx context.Context,
	mainID uuid.UUID,
	rows []sqlc.AppShort,
	assetCache map[uuid.UUID]sqlc.AppMediaAsset,
) (WorkspacePreviewShortDetailItem, bool, error) {
	for _, row := range rows {
		canonicalMainID, err := postgres.UUIDFromPG(row.CanonicalMainID)
		if err != nil {
			shortID, shortErr := postgres.UUIDFromPG(row.ID)
			if shortErr != nil {
				return WorkspacePreviewShortDetailItem{}, false, fmt.Errorf("workspace preview entry short id / canonical main id 変換: %w", err)
			}
			return WorkspacePreviewShortDetailItem{}, false, fmt.Errorf(
				"workspace preview entry short canonical main id 変換 short=%s: %w",
				shortID,
				err,
			)
		}
		if canonicalMainID != mainID {
			continue
		}

		shortDetail, ok, err := r.buildWorkspacePreviewShortDetailItem(ctx, row, assetCache)
		if err != nil {
			return WorkspacePreviewShortDetailItem{}, false, err
		}
		if ok {
			return shortDetail, true, nil
		}
	}

	return WorkspacePreviewShortDetailItem{}, false, nil
}

func normalizeWorkspacePreviewCaption(value pgtype.Text) string {
	caption := postgres.OptionalTextFromPG(value)
	if caption == nil {
		return ""
	}

	return strings.TrimSpace(*caption)
}
