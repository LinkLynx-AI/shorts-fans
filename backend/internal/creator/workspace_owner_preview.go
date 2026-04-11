package creator

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	workspacePreviewAssetReadyState = "ready"
	workspacePreviewCurrencyCode    = "JPY"
)

// ErrWorkspacePreviewNotFound は creator workspace preview 対象が解決できないことを表します。
var ErrWorkspacePreviewNotFound = errors.New("creator workspace preview was not found")

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

// WorkspacePreviewShortSummary は owner preview detail 用 short summary の read model です。
type WorkspacePreviewShortSummary struct {
	CanonicalMainID        uuid.UUID
	Caption                string
	ID                     uuid.UUID
	Media                  media.VideoDisplayAsset
	PreviewDurationSeconds int64
	Title                  string
}

// WorkspacePreviewShortDetail は owner preview 用 short detail の read model です。
type WorkspacePreviewShortDetail struct {
	Creator Profile
	Short   WorkspacePreviewShortSummary
}

// WorkspacePreviewMainSummary は owner preview detail 用 main summary の read model です。
type WorkspacePreviewMainSummary struct {
	DurationSeconds int64
	ID              uuid.UUID
	Media           media.VideoDisplayAsset
	PriceJpy        int64
	Title           string
}

// WorkspacePreviewMainDetail は owner preview 用 main detail の read model です。
type WorkspacePreviewMainDetail struct {
	Creator    Profile
	EntryShort WorkspacePreviewShortSummary
	Main       WorkspacePreviewMainSummary
}

type previewableLinkedShort struct {
	CanonicalMainID uuid.UUID
	CreatedAt       time.Time
	ID              uuid.UUID
}

// GetWorkspacePreviewShortDetail は current viewer 自身の short preview detail を返します。
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
		return WorkspacePreviewShortDetail{}, fmt.Errorf("creator workspace short preview detail 取得 user=%s short=%s: delivery is nil", viewerUserID, shortID)
	}

	row, err := r.queries.GetShortByID(ctx, postgres.UUIDToPG(shortID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return WorkspacePreviewShortDetail{}, fmt.Errorf(
				"creator workspace short preview detail 取得 user=%s short=%s: %w",
				viewerUserID,
				shortID,
				ErrWorkspacePreviewNotFound,
			)
		}
		return WorkspacePreviewShortDetail{}, fmt.Errorf("creator workspace short preview detail 取得 user=%s short=%s: %w", viewerUserID, shortID, err)
	}
	rowCreatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return WorkspacePreviewShortDetail{}, fmt.Errorf("creator workspace short preview detail 取得 user=%s short=%s creator 変換: %w", viewerUserID, shortID, err)
	}
	if rowCreatorUserID != viewerUserID {
		return WorkspacePreviewShortDetail{}, fmt.Errorf(
			"creator workspace short preview detail 取得 user=%s short=%s: %w",
			viewerUserID,
			shortID,
			ErrWorkspacePreviewNotFound,
		)
	}

	assetCache := map[uuid.UUID]sqlc.AppMediaAsset{}
	summary, ok, buildErr := r.buildWorkspacePreviewShortSummary(ctx, row, assetCache)
	if buildErr != nil {
		return WorkspacePreviewShortDetail{}, fmt.Errorf("creator workspace short preview detail 取得 user=%s short=%s: %w", viewerUserID, shortID, buildErr)
	}
	if !ok {
		return WorkspacePreviewShortDetail{}, fmt.Errorf(
			"creator workspace short preview detail 取得 user=%s short=%s: %w",
			viewerUserID,
			shortID,
			ErrWorkspacePreviewNotFound,
		)
	}

	return WorkspacePreviewShortDetail{
		Creator: profile,
		Short:   summary,
	}, nil
}

// GetWorkspacePreviewMainDetail は current viewer 自身の main preview detail を返します。
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
		return WorkspacePreviewMainDetail{}, fmt.Errorf("creator workspace main preview detail 取得 user=%s main=%s: delivery is nil", viewerUserID, mainID)
	}

	mainRows, err := r.queries.ListMainsByCreatorUserID(ctx, postgres.UUIDToPG(viewerUserID))
	if err != nil {
		return WorkspacePreviewMainDetail{}, fmt.Errorf("creator workspace main preview detail 取得 user=%s main=%s: %w", viewerUserID, mainID, err)
	}

	assetCache := map[uuid.UUID]sqlc.AppMediaAsset{}
	var mainSummary WorkspacePreviewMainSummary
	mainFound := false
	for _, row := range mainRows {
		summary, ok, buildErr := r.buildWorkspacePreviewMainSummary(ctx, row, assetCache)
		if buildErr != nil {
			return WorkspacePreviewMainDetail{}, fmt.Errorf("creator workspace main preview detail 取得 user=%s main=%s: %w", viewerUserID, mainID, buildErr)
		}
		if !ok || summary.ID != mainID {
			continue
		}

		mainSummary = summary
		mainFound = true
		break
	}
	if !mainFound {
		return WorkspacePreviewMainDetail{}, fmt.Errorf(
			"creator workspace main preview detail 取得 user=%s main=%s: %w",
			viewerUserID,
			mainID,
			ErrWorkspacePreviewNotFound,
		)
	}

	shortRows, err := r.queries.ListShortsByCreatorUserID(ctx, postgres.UUIDToPG(viewerUserID))
	if err != nil {
		return WorkspacePreviewMainDetail{}, fmt.Errorf("creator workspace main preview detail 取得 user=%s main=%s linked short 読み込み: %w", viewerUserID, mainID, err)
	}

	for _, row := range shortRows {
		entryShort, ok, buildErr := r.buildWorkspacePreviewShortSummary(ctx, row, assetCache)
		if buildErr != nil {
			return WorkspacePreviewMainDetail{}, fmt.Errorf("creator workspace main preview detail 取得 user=%s main=%s: %w", viewerUserID, mainID, buildErr)
		}
		if !ok || entryShort.CanonicalMainID != mainID {
			continue
		}

		return WorkspacePreviewMainDetail{
			Creator:    profile,
			EntryShort: entryShort,
			Main:       mainSummary,
		}, nil
	}

	return WorkspacePreviewMainDetail{}, fmt.Errorf(
		"creator workspace main preview detail 取得 user=%s main=%s: %w",
		viewerUserID,
		mainID,
		ErrWorkspacePreviewNotFound,
	)
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
	mainRows, err := r.queries.ListCreatorWorkspacePreviewMainsByCreatorUserID(ctx, postgres.UUIDToPG(viewerUserID))
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
	row sqlc.ListCreatorWorkspacePreviewMainsByCreatorUserIDRow,
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

func (r *Repository) buildWorkspacePreviewShortSummary(
	ctx context.Context,
	row sqlc.AppShort,
	assetCache map[uuid.UUID]sqlc.AppMediaAsset,
) (WorkspacePreviewShortSummary, bool, error) {
	shortID, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return WorkspacePreviewShortSummary{}, false, fmt.Errorf("workspace preview short detail id 変換: %w", err)
	}
	canonicalMainID, err := postgres.UUIDFromPG(row.CanonicalMainID)
	if err != nil {
		return WorkspacePreviewShortSummary{}, false, fmt.Errorf("workspace preview short detail canonical main id 変換 short=%s: %w", shortID, err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return WorkspacePreviewShortSummary{}, false, fmt.Errorf("workspace preview short detail media asset id 変換 short=%s: %w", shortID, err)
	}

	asset, err := r.getWorkspacePreviewAsset(ctx, mediaAssetID, assetCache)
	if err != nil {
		return WorkspacePreviewShortSummary{}, false, fmt.Errorf("workspace preview short detail asset 解決 short=%s: %w", shortID, err)
	}
	if !workspacePreviewAssetReady(asset) {
		return WorkspacePreviewShortSummary{}, false, nil
	}

	durationMS := postgres.OptionalInt64FromPG(asset.DurationMs)
	if durationMS == nil || *durationMS <= 0 {
		return WorkspacePreviewShortSummary{}, false, fmt.Errorf("workspace preview short detail duration がありません short=%s", shortID)
	}

	displayAsset, err := r.delivery.ResolveShortDisplayAsset(media.ShortDisplaySource{
		AssetID:    mediaAssetID,
		ShortID:    shortID,
		DurationMS: *durationMS,
	}, media.AccessBoundaryOwner)
	if err != nil {
		return WorkspacePreviewShortSummary{}, false, fmt.Errorf("workspace preview short detail display asset 解決 short=%s: %w", shortID, err)
	}

	caption := ""
	if captionValue := postgres.OptionalTextFromPG(row.Caption); captionValue != nil {
		caption = strings.TrimSpace(*captionValue)
	}

	return WorkspacePreviewShortSummary{
		CanonicalMainID:        canonicalMainID,
		Caption:                caption,
		ID:                     shortID,
		Media:                  displayAsset,
		PreviewDurationSeconds: displayAsset.DurationSeconds,
		Title:                  normalizeWorkspacePreviewTitleFromCaption(caption),
	}, true, nil
}

func (r *Repository) buildWorkspacePreviewMainSummary(
	ctx context.Context,
	row sqlc.AppMain,
	assetCache map[uuid.UUID]sqlc.AppMediaAsset,
) (WorkspacePreviewMainSummary, bool, error) {
	if row.CurrencyCode != workspacePreviewCurrencyCode {
		return WorkspacePreviewMainSummary{}, false, nil
	}

	mainID, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return WorkspacePreviewMainSummary{}, false, fmt.Errorf("workspace preview main detail id 変換: %w", err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return WorkspacePreviewMainSummary{}, false, fmt.Errorf("workspace preview main detail media asset id 変換 main=%s: %w", mainID, err)
	}

	asset, err := r.getWorkspacePreviewAsset(ctx, mediaAssetID, assetCache)
	if err != nil {
		return WorkspacePreviewMainSummary{}, false, fmt.Errorf("workspace preview main detail asset 解決 main=%s: %w", mainID, err)
	}
	if !workspacePreviewAssetReady(asset) {
		return WorkspacePreviewMainSummary{}, false, nil
	}

	durationMS := postgres.OptionalInt64FromPG(asset.DurationMs)
	if durationMS == nil || *durationMS <= 0 {
		return WorkspacePreviewMainSummary{}, false, fmt.Errorf("workspace preview main detail duration がありません main=%s", mainID)
	}

	displayAsset, err := r.delivery.ResolveMainDisplayAsset(ctx, media.MainDisplaySource{
		AssetID:    mediaAssetID,
		MainID:     mainID,
		DurationMS: *durationMS,
	}, media.AccessBoundaryOwner, media.DefaultSignedURLTTL)
	if err != nil {
		return WorkspacePreviewMainSummary{}, false, fmt.Errorf("workspace preview main detail display asset 解決 main=%s: %w", mainID, err)
	}

	return WorkspacePreviewMainSummary{
		DurationSeconds: displayAsset.DurationSeconds,
		ID:              mainID,
		Media:           displayAsset,
		PriceJpy:        row.PriceMinor,
		// main title is not persisted in the current backend model, so keep the
		// contract shape without inventing a derived title here.
		Title: "",
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

func normalizeWorkspacePreviewTitleFromCaption(caption string) string {
	normalized := strings.TrimSpace(caption)

	return strings.TrimRightFunc(normalized, func(r rune) bool {
		switch r {
		case '。', '.', '!', '?':
			return true
		default:
			return false
		}
	})
}
