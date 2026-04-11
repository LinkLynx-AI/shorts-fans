package creator

import (
	"context"
	"fmt"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
)

// WorkspaceTopMainPerformer は creator workspace top main の read model です。
type WorkspaceTopMainPerformer struct {
	ID          uuid.UUID
	Media       media.VideoPreviewCardAsset
	UnlockCount int64
}

// WorkspaceTopShortPerformer は creator workspace top short の read model です。
type WorkspaceTopShortPerformer struct {
	AttributedUnlockCount int64
	ID                    uuid.UUID
	Media                 media.VideoPreviewCardAsset
}

// WorkspaceTopPerformers は creator workspace 上部の top performers を表します。
type WorkspaceTopPerformers struct {
	TopMain  *WorkspaceTopMainPerformer
	TopShort *WorkspaceTopShortPerformer
}

// GetWorkspaceTopPerformers は current viewer 自身の creator workspace top performers を返します。
func (r *Repository) GetWorkspaceTopPerformers(ctx context.Context, viewerUserID uuid.UUID) (WorkspaceTopPerformers, error) {
	if _, err := r.getApprovedWorkspaceProfile(ctx, viewerUserID); err != nil {
		return WorkspaceTopPerformers{}, err
	}
	if r.delivery == nil {
		return WorkspaceTopPerformers{}, fmt.Errorf("creator workspace top performers 取得 user=%s: delivery is nil", viewerUserID)
	}

	assetCache := map[uuid.UUID]sqlc.AppMediaAsset{}

	topMainCandidates, err := r.queries.ListCreatorWorkspaceTopMainCandidatesByCreatorUserID(ctx, postgres.UUIDToPG(viewerUserID))
	if err != nil {
		return WorkspaceTopPerformers{}, fmt.Errorf("creator workspace top performers 取得 user=%s top main 候補読み込み: %w", viewerUserID, err)
	}
	topMain, err := r.resolveWorkspaceTopMainPerformer(ctx, topMainCandidates, assetCache)
	if err != nil {
		return WorkspaceTopPerformers{}, fmt.Errorf("creator workspace top performers 取得 user=%s top main 解決: %w", viewerUserID, err)
	}

	topShortCandidates, err := r.queries.ListCreatorWorkspaceTopShortCandidatesByCreatorUserID(ctx, postgres.UUIDToPG(viewerUserID))
	if err != nil {
		return WorkspaceTopPerformers{}, fmt.Errorf("creator workspace top performers 取得 user=%s top short 候補読み込み: %w", viewerUserID, err)
	}
	topShort, err := r.resolveWorkspaceTopShortPerformer(ctx, topShortCandidates, assetCache)
	if err != nil {
		return WorkspaceTopPerformers{}, fmt.Errorf("creator workspace top performers 取得 user=%s top short 解決: %w", viewerUserID, err)
	}

	return WorkspaceTopPerformers{
		TopMain:  topMain,
		TopShort: topShort,
	}, nil
}

func (r *Repository) resolveWorkspaceTopMainPerformer(
	ctx context.Context,
	rows []sqlc.ListCreatorWorkspaceTopMainCandidatesByCreatorUserIDRow,
	assetCache map[uuid.UUID]sqlc.AppMediaAsset,
) (*WorkspaceTopMainPerformer, error) {
	for _, row := range rows {
		item, ok, err := r.buildWorkspaceTopMainPerformer(ctx, row, assetCache)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		return &item, nil
	}

	return nil, nil
}

func (r *Repository) resolveWorkspaceTopShortPerformer(
	ctx context.Context,
	rows []sqlc.ListCreatorWorkspaceTopShortCandidatesByCreatorUserIDRow,
	assetCache map[uuid.UUID]sqlc.AppMediaAsset,
) (*WorkspaceTopShortPerformer, error) {
	for _, row := range rows {
		item, ok, err := r.buildWorkspaceTopShortPerformer(ctx, row, assetCache)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		return &item, nil
	}

	return nil, nil
}

func (r *Repository) buildWorkspaceTopMainPerformer(
	ctx context.Context,
	row sqlc.ListCreatorWorkspaceTopMainCandidatesByCreatorUserIDRow,
	assetCache map[uuid.UUID]sqlc.AppMediaAsset,
) (WorkspaceTopMainPerformer, bool, error) {
	if row.CurrencyCode != workspacePreviewCurrencyCode {
		return WorkspaceTopMainPerformer{}, false, nil
	}

	mainID, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return WorkspaceTopMainPerformer{}, false, fmt.Errorf("workspace top main id 変換: %w", err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return WorkspaceTopMainPerformer{}, false, fmt.Errorf("workspace top main media asset id 変換 main=%s: %w", mainID, err)
	}

	asset, err := r.getWorkspacePreviewAsset(ctx, mediaAssetID, assetCache)
	if err != nil {
		return WorkspaceTopMainPerformer{}, false, fmt.Errorf("workspace top main asset 解決 main=%s: %w", mainID, err)
	}
	if !workspacePreviewAssetReady(asset) {
		return WorkspaceTopMainPerformer{}, false, nil
	}

	durationMS := postgres.OptionalInt64FromPG(asset.DurationMs)
	if durationMS == nil || *durationMS <= 0 {
		return WorkspaceTopMainPerformer{}, false, fmt.Errorf("workspace top main duration がありません main=%s", mainID)
	}

	previewAsset, err := r.delivery.ResolveMainPreviewCardAsset(ctx, media.MainDisplaySource{
		AssetID:    mediaAssetID,
		MainID:     mainID,
		DurationMS: *durationMS,
	}, media.DefaultSignedURLTTL)
	if err != nil {
		return WorkspaceTopMainPerformer{}, false, fmt.Errorf("workspace top main display asset 解決 main=%s: %w", mainID, err)
	}

	return WorkspaceTopMainPerformer{
		ID:          mainID,
		Media:       previewAsset,
		UnlockCount: row.UnlockCount,
	}, true, nil
}

func (r *Repository) buildWorkspaceTopShortPerformer(
	ctx context.Context,
	row sqlc.ListCreatorWorkspaceTopShortCandidatesByCreatorUserIDRow,
	assetCache map[uuid.UUID]sqlc.AppMediaAsset,
) (WorkspaceTopShortPerformer, bool, error) {
	shortID, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return WorkspaceTopShortPerformer{}, false, fmt.Errorf("workspace top short id 変換: %w", err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return WorkspaceTopShortPerformer{}, false, fmt.Errorf("workspace top short media asset id 変換 short=%s: %w", shortID, err)
	}

	asset, err := r.getWorkspacePreviewAsset(ctx, mediaAssetID, assetCache)
	if err != nil {
		return WorkspaceTopShortPerformer{}, false, fmt.Errorf("workspace top short asset 解決 short=%s: %w", shortID, err)
	}
	if !workspacePreviewAssetReady(asset) {
		return WorkspaceTopShortPerformer{}, false, nil
	}

	durationMS := postgres.OptionalInt64FromPG(asset.DurationMs)
	if durationMS == nil || *durationMS <= 0 {
		return WorkspaceTopShortPerformer{}, false, fmt.Errorf("workspace top short duration がありません short=%s", shortID)
	}

	previewAsset, err := r.delivery.ResolveShortPreviewCardAsset(media.ShortDisplaySource{
		AssetID:    mediaAssetID,
		ShortID:    shortID,
		DurationMS: *durationMS,
	})
	if err != nil {
		return WorkspaceTopShortPerformer{}, false, fmt.Errorf("workspace top short display asset 解決 short=%s: %w", shortID, err)
	}

	return WorkspaceTopShortPerformer{
		AttributedUnlockCount: row.AttributedUnlockCount,
		ID:                    shortID,
		Media:                 previewAsset,
	}, true, nil
}
