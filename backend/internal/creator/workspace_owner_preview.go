package creator

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
)

const (
	workspacePreviewAssetReadyState = "ready"
	workspacePreviewCurrencyCode    = "JPY"
)

// DefaultWorkspacePreviewPageSize は creator workspace owner preview list の既定 page size です。
const DefaultWorkspacePreviewPageSize = 18

// WorkspacePreviewCursor は creator workspace owner preview list の keyset cursor です。
type WorkspacePreviewCursor struct {
	CreatedAt time.Time
	ID        uuid.UUID
}

// WorkspacePreviewShortItem は owner preview 用 short list item の read model です。
type WorkspacePreviewShortItem struct {
	CanonicalMainID        uuid.UUID
	CreatedAt              time.Time
	ID                     uuid.UUID
	Media                  media.VideoPreviewCardAsset
	PreviewDurationSeconds int64
}

// WorkspacePreviewMainItem は owner preview 用 main list item の read model です。
type WorkspacePreviewMainItem struct {
	CreatedAt       time.Time
	DurationSeconds int64
	ID              uuid.UUID
	LeadShortID     uuid.UUID
	Media           media.VideoPreviewCardAsset
	PriceJpy        int64
}

type previewableLinkedShort struct {
	CanonicalMainID uuid.UUID
	CreatedAt       time.Time
	ID              uuid.UUID
}

// ListWorkspacePreviewShorts は current viewer 自身の preview 可能 short 一覧を返します。
func (r *Repository) ListWorkspacePreviewShorts(
	ctx context.Context,
	viewerUserID uuid.UUID,
	cursor *WorkspacePreviewCursor,
	limit int,
) ([]WorkspacePreviewShortItem, *WorkspacePreviewCursor, error) {
	if _, err := r.getApprovedWorkspaceProfile(ctx, viewerUserID); err != nil {
		return nil, nil, err
	}
	if r.delivery == nil {
		return nil, nil, fmt.Errorf("creator workspace short preview 一覧取得 user=%s: delivery is nil", viewerUserID)
	}

	pageLimit := resolveWorkspacePreviewPageLimit(limit)
	shortRows, err := r.queries.ListShortsByCreatorUserID(ctx, postgres.UUIDToPG(viewerUserID))
	if err != nil {
		return nil, nil, fmt.Errorf("creator workspace short preview 一覧取得 user=%s: %w", viewerUserID, err)
	}

	assetCache := map[uuid.UUID]sqlc.AppMediaAsset{}
	items := make([]WorkspacePreviewShortItem, 0, pageLimit+1)
	for _, row := range shortRows {
		item, ok, buildErr := r.buildWorkspacePreviewShortItem(ctx, row, assetCache)
		if buildErr != nil {
			return nil, nil, fmt.Errorf("creator workspace short preview 一覧取得 user=%s: %w", viewerUserID, buildErr)
		}
		if !ok || !workspacePreviewAfterCursor(item.CreatedAt, item.ID, cursor) {
			continue
		}

		items = append(items, item)
		if len(items) > pageLimit {
			break
		}
	}

	pageItems, nextCursor := finalizeWorkspacePreviewShortPage(items, pageLimit)

	return pageItems, nextCursor, nil
}

// ListWorkspacePreviewMains は current viewer 自身の preview 可能 main 一覧を返します。
func (r *Repository) ListWorkspacePreviewMains(
	ctx context.Context,
	viewerUserID uuid.UUID,
	cursor *WorkspacePreviewCursor,
	limit int,
) ([]WorkspacePreviewMainItem, *WorkspacePreviewCursor, error) {
	if _, err := r.getApprovedWorkspaceProfile(ctx, viewerUserID); err != nil {
		return nil, nil, err
	}
	if r.delivery == nil {
		return nil, nil, fmt.Errorf("creator workspace main preview 一覧取得 user=%s: delivery is nil", viewerUserID)
	}

	pageLimit := resolveWorkspacePreviewPageLimit(limit)
	mainRows, err := r.queries.ListMainsByCreatorUserID(ctx, postgres.UUIDToPG(viewerUserID))
	if err != nil {
		return nil, nil, fmt.Errorf("creator workspace main preview 一覧取得 user=%s: %w", viewerUserID, err)
	}
	shortRows, err := r.queries.ListShortsByCreatorUserID(ctx, postgres.UUIDToPG(viewerUserID))
	if err != nil {
		return nil, nil, fmt.Errorf("creator workspace main preview 一覧取得 user=%s linked short 読み込み: %w", viewerUserID, err)
	}

	assetCache := map[uuid.UUID]sqlc.AppMediaAsset{}
	leadShortByMainID, err := buildWorkspacePreviewLeadShortMap(shortRows, assetCache, func(assetID uuid.UUID) (sqlc.AppMediaAsset, error) {
		return r.getWorkspacePreviewAsset(ctx, assetID, assetCache)
	})
	if err != nil {
		return nil, nil, fmt.Errorf("creator workspace main preview 一覧取得 user=%s: %w", viewerUserID, err)
	}

	items := make([]WorkspacePreviewMainItem, 0, pageLimit+1)
	for _, row := range mainRows {
		item, ok, buildErr := r.buildWorkspacePreviewMainItem(ctx, row, leadShortByMainID, assetCache)
		if buildErr != nil {
			return nil, nil, fmt.Errorf("creator workspace main preview 一覧取得 user=%s: %w", viewerUserID, buildErr)
		}
		if !ok || !workspacePreviewAfterCursor(item.CreatedAt, item.ID, cursor) {
			continue
		}

		items = append(items, item)
		if len(items) > pageLimit {
			break
		}
	}

	pageItems, nextCursor := finalizeWorkspacePreviewMainPage(items, pageLimit)

	return pageItems, nextCursor, nil
}

func resolveWorkspacePreviewPageLimit(limit int) int {
	if limit <= 0 {
		return DefaultWorkspacePreviewPageSize
	}

	return limit
}

func workspacePreviewAfterCursor(createdAt time.Time, id uuid.UUID, cursor *WorkspacePreviewCursor) bool {
	if cursor == nil {
		return true
	}
	if createdAt.Before(cursor.CreatedAt) {
		return true
	}
	if createdAt.After(cursor.CreatedAt) {
		return false
	}

	return bytes.Compare(id[:], cursor.ID[:]) < 0
}

func finalizeWorkspacePreviewShortPage(
	items []WorkspacePreviewShortItem,
	limit int,
) ([]WorkspacePreviewShortItem, *WorkspacePreviewCursor) {
	if len(items) <= limit {
		return items, nil
	}

	pageItems := items[:limit]
	lastItem := pageItems[len(pageItems)-1]

	return pageItems, &WorkspacePreviewCursor{
		CreatedAt: lastItem.CreatedAt,
		ID:        lastItem.ID,
	}
}

func finalizeWorkspacePreviewMainPage(
	items []WorkspacePreviewMainItem,
	limit int,
) ([]WorkspacePreviewMainItem, *WorkspacePreviewCursor) {
	if len(items) <= limit {
		return items, nil
	}

	pageItems := items[:limit]
	lastItem := pageItems[len(pageItems)-1]

	return pageItems, &WorkspacePreviewCursor{
		CreatedAt: lastItem.CreatedAt,
		ID:        lastItem.ID,
	}
}

func buildWorkspacePreviewLeadShortMap(
	shortRows []sqlc.AppShort,
	assetCache map[uuid.UUID]sqlc.AppMediaAsset,
	getAsset func(uuid.UUID) (sqlc.AppMediaAsset, error),
) (map[uuid.UUID]previewableLinkedShort, error) {
	leadShortByMainID := make(map[uuid.UUID]previewableLinkedShort)

	for _, row := range shortRows {
		previewableShort, ok, err := buildPreviewableLinkedShort(row, assetCache, getAsset)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		if _, exists := leadShortByMainID[previewableShort.CanonicalMainID]; exists {
			continue
		}

		leadShortByMainID[previewableShort.CanonicalMainID] = previewableShort
	}

	return leadShortByMainID, nil
}

func buildPreviewableLinkedShort(
	row sqlc.AppShort,
	assetCache map[uuid.UUID]sqlc.AppMediaAsset,
	getAsset func(uuid.UUID) (sqlc.AppMediaAsset, error),
) (previewableLinkedShort, bool, error) {
	shortID, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return previewableLinkedShort{}, false, fmt.Errorf("preview linked short id 変換: %w", err)
	}
	canonicalMainID, err := postgres.UUIDFromPG(row.CanonicalMainID)
	if err != nil {
		return previewableLinkedShort{}, false, fmt.Errorf("preview linked short canonical main id 変換 short=%s: %w", shortID, err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return previewableLinkedShort{}, false, fmt.Errorf("preview linked short media asset id 変換 short=%s: %w", shortID, err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return previewableLinkedShort{}, false, fmt.Errorf("preview linked short created_at 変換 short=%s: %w", shortID, err)
	}

	asset, err := resolveWorkspacePreviewMediaAsset(mediaAssetID, assetCache, getAsset)
	if err != nil {
		return previewableLinkedShort{}, false, fmt.Errorf("preview linked short asset 解決 short=%s: %w", shortID, err)
	}
	if !workspacePreviewAssetReady(asset) {
		return previewableLinkedShort{}, false, nil
	}

	return previewableLinkedShort{
		CanonicalMainID: canonicalMainID,
		CreatedAt:       createdAt,
		ID:              shortID,
	}, true, nil
}

func (r *Repository) buildWorkspacePreviewShortItem(
	ctx context.Context,
	row sqlc.AppShort,
	assetCache map[uuid.UUID]sqlc.AppMediaAsset,
) (WorkspacePreviewShortItem, bool, error) {
	shortID, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return WorkspacePreviewShortItem{}, false, fmt.Errorf("workspace preview short id 変換: %w", err)
	}
	canonicalMainID, err := postgres.UUIDFromPG(row.CanonicalMainID)
	if err != nil {
		return WorkspacePreviewShortItem{}, false, fmt.Errorf("workspace preview short canonical main id 変換 short=%s: %w", shortID, err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return WorkspacePreviewShortItem{}, false, fmt.Errorf("workspace preview short media asset id 変換 short=%s: %w", shortID, err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return WorkspacePreviewShortItem{}, false, fmt.Errorf("workspace preview short created_at 変換 short=%s: %w", shortID, err)
	}

	asset, err := r.getWorkspacePreviewAsset(ctx, mediaAssetID, assetCache)
	if err != nil {
		return WorkspacePreviewShortItem{}, false, fmt.Errorf("workspace preview short asset 解決 short=%s: %w", shortID, err)
	}
	if !workspacePreviewAssetReady(asset) {
		return WorkspacePreviewShortItem{}, false, nil
	}

	durationMS := postgres.OptionalInt64FromPG(asset.DurationMs)
	if durationMS == nil || *durationMS <= 0 {
		return WorkspacePreviewShortItem{}, false, fmt.Errorf("workspace preview short duration がありません short=%s", shortID)
	}

	previewAsset, err := r.delivery.ResolveShortPreviewCardAsset(media.ShortDisplaySource{
		AssetID:    mediaAssetID,
		ShortID:    shortID,
		DurationMS: *durationMS,
	})
	if err != nil {
		return WorkspacePreviewShortItem{}, false, fmt.Errorf("workspace preview short display asset 解決 short=%s: %w", shortID, err)
	}

	return WorkspacePreviewShortItem{
		CanonicalMainID:        canonicalMainID,
		CreatedAt:              createdAt,
		ID:                     shortID,
		Media:                  previewAsset,
		PreviewDurationSeconds: previewAsset.DurationSeconds,
	}, true, nil
}

func (r *Repository) buildWorkspacePreviewMainItem(
	ctx context.Context,
	row sqlc.AppMain,
	leadShortByMainID map[uuid.UUID]previewableLinkedShort,
	assetCache map[uuid.UUID]sqlc.AppMediaAsset,
) (WorkspacePreviewMainItem, bool, error) {
	if row.CurrencyCode != workspacePreviewCurrencyCode {
		return WorkspacePreviewMainItem{}, false, nil
	}

	mainID, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return WorkspacePreviewMainItem{}, false, fmt.Errorf("workspace preview main id 変換: %w", err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return WorkspacePreviewMainItem{}, false, fmt.Errorf("workspace preview main media asset id 変換 main=%s: %w", mainID, err)
	}
	createdAt, err := postgres.RequiredTimeFromPG(row.CreatedAt)
	if err != nil {
		return WorkspacePreviewMainItem{}, false, fmt.Errorf("workspace preview main created_at 変換 main=%s: %w", mainID, err)
	}

	leadShort, ok := leadShortByMainID[mainID]
	if !ok {
		return WorkspacePreviewMainItem{}, false, nil
	}

	asset, err := r.getWorkspacePreviewAsset(ctx, mediaAssetID, assetCache)
	if err != nil {
		return WorkspacePreviewMainItem{}, false, fmt.Errorf("workspace preview main asset 解決 main=%s: %w", mainID, err)
	}
	if !workspacePreviewAssetReady(asset) {
		return WorkspacePreviewMainItem{}, false, nil
	}

	durationMS := postgres.OptionalInt64FromPG(asset.DurationMs)
	if durationMS == nil || *durationMS <= 0 {
		return WorkspacePreviewMainItem{}, false, fmt.Errorf("workspace preview main duration がありません main=%s", mainID)
	}

	previewAsset, err := r.delivery.ResolveMainPreviewCardAsset(ctx, media.MainDisplaySource{
		AssetID:    mediaAssetID,
		MainID:     mainID,
		DurationMS: *durationMS,
	}, media.DefaultSignedURLTTL)
	if err != nil {
		return WorkspacePreviewMainItem{}, false, fmt.Errorf("workspace preview main display asset 解決 main=%s: %w", mainID, err)
	}

	return WorkspacePreviewMainItem{
		CreatedAt:       createdAt,
		DurationSeconds: previewAsset.DurationSeconds,
		ID:              mainID,
		LeadShortID:     leadShort.ID,
		Media:           previewAsset,
		PriceJpy:        row.PriceMinor,
	}, true, nil
}

func workspacePreviewAssetReady(asset sqlc.AppMediaAsset) bool {
	durationMS := postgres.OptionalInt64FromPG(asset.DurationMs)

	return asset.ProcessingState == workspacePreviewAssetReadyState &&
		durationMS != nil &&
		*durationMS > 0
}

func resolveWorkspacePreviewMediaAsset(
	mediaAssetID uuid.UUID,
	assetCache map[uuid.UUID]sqlc.AppMediaAsset,
	getAsset func(uuid.UUID) (sqlc.AppMediaAsset, error),
) (sqlc.AppMediaAsset, error) {
	if cachedAsset, ok := assetCache[mediaAssetID]; ok {
		return cachedAsset, nil
	}

	asset, err := getAsset(mediaAssetID)
	if err != nil {
		return sqlc.AppMediaAsset{}, err
	}

	assetCache[mediaAssetID] = asset
	return asset, nil
}

func (r *Repository) getWorkspacePreviewAsset(
	ctx context.Context,
	mediaAssetID uuid.UUID,
	assetCache map[uuid.UUID]sqlc.AppMediaAsset,
) (sqlc.AppMediaAsset, error) {
	return resolveWorkspacePreviewMediaAsset(mediaAssetID, assetCache, func(id uuid.UUID) (sqlc.AppMediaAsset, error) {
		asset, err := r.queries.GetMediaAssetByID(ctx, postgres.UUIDToPG(id))
		if err != nil {
			return sqlc.AppMediaAsset{}, fmt.Errorf("media asset 読み込み id=%s: %w", id, err)
		}

		return asset, nil
	})
}
