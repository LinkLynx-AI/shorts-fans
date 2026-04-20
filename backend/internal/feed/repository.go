package feed

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultPageSize は public short feed の既定 page size です。
const DefaultPageSize = 20

// ErrPublicShortNotFound は対象の public short が存在しないことを表します。
var ErrPublicShortNotFound = errors.New("public short が見つかりません")

// ErrFollowingCursorInvalid は following feed cursor が継続取得に必要な情報を欠いていることを表します。
var ErrFollowingCursorInvalid = errors.New("public short following feed cursor が不正です")

type queries interface {
	GetPublicShortDetailItem(ctx context.Context, arg sqlc.GetPublicShortDetailItemParams) (sqlc.GetPublicShortDetailItemRow, error)
	ListFeedItemsByShortIDs(ctx context.Context, arg sqlc.ListFeedItemsByShortIDsParams) ([]sqlc.ListFeedItemsByShortIDsRow, error)
	ListFollowingPublicFeedItems(ctx context.Context, arg sqlc.ListFollowingPublicFeedItemsParams) ([]sqlc.ListFollowingPublicFeedItemsRow, error)
	ListRecommendedPublicFeedItems(ctx context.Context, arg sqlc.ListRecommendedPublicFeedItemsParams) ([]sqlc.ListRecommendedPublicFeedItemsRow, error)
}

// Repository は fan public short feed/detail 向けの read 操作をまとめます。
type Repository struct {
	queries queries
}

// Cursor は feed keyset pagination 用の cursor です。
type Cursor struct {
	FollowingRemainingShortIDs []uuid.UUID
	PublishedAt                time.Time
	ShortID                    uuid.UUID
}

// CreatorSummary は public short surface の creator 表示情報です。
type CreatorSummary struct {
	AvatarURL   *string
	Bio         string
	DisplayName string
	Handle      string
	ID          uuid.UUID
	IsFollowing bool
}

// ShortSummary は public short surface の short 表示情報です。
type ShortSummary struct {
	Caption                string
	CanonicalMainID        uuid.UUID
	CreatorUserID          uuid.UUID
	ID                     uuid.UUID
	MediaAssetID           uuid.UUID
	PreviewDurationSeconds int64
	PublishedAt            time.Time
}

// UnlockPreview は public short surface の CTA 判定に必要な main 情報です。
type UnlockPreview struct {
	IsOwner             bool
	IsUnlocked          bool
	MainDurationSeconds int64
	PriceJPY            int64
}

// Item は public short feed/detail に共通の read model です。
type Item struct {
	Creator CreatorSummary
	Short   ShortSummary
	Unlock  UnlockPreview
	Viewer  struct {
		IsFollowingCreator bool
		IsPinned           bool
	}
}

// Detail は short detail surface の read model です。
type Detail struct {
	Item   Item
	Viewer struct {
		IsFollowingCreator bool
	}
}

// NewRepository は public short feed repository を構築します。
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{queries: sqlc.New(pool)}
}

// ListRecommended は public short recommended feed の 1 page を返します。
func (r *Repository) ListRecommended(ctx context.Context, viewerUserID *uuid.UUID, cursor *Cursor, limit int) ([]Item, *Cursor, error) {
	params, pageLimit := buildRecommendedPageParams(viewerUserID, cursor, limit)

	rows, err := r.queries.ListRecommendedPublicFeedItems(ctx, params)
	if err != nil {
		return nil, nil, fmt.Errorf("public short recommended feed 取得: %w", err)
	}

	return mapRecommendedPage(rows, pageLimit, "public short recommended feed 取得結果の変換")
}

// ListFollowing は follow 中 creator の public short feed 1 page を返します。
func (r *Repository) ListFollowing(ctx context.Context, viewerUserID uuid.UUID, cursor *Cursor, limit int) ([]Item, *Cursor, error) {
	pageLimit := resolvePageLimit(limit)
	if cursor == nil {
		rankingReferenceAt := currentFollowingRankingReferenceAt()
		rows, err := r.queries.ListFollowingPublicFeedItems(
			ctx,
			buildInitialFollowingPageParams(viewerUserID, rankingReferenceAt),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("public short following feed 取得 viewer=%s: %w", viewerUserID, err)
		}

		return mapInitialFollowingPage(
			rows,
			pageLimit,
			fmt.Sprintf("public short following feed 取得結果の変換 viewer=%s", viewerUserID),
		)
	}

	if err := validateFollowingCursor(cursor); err != nil {
		return nil, nil, fmt.Errorf("public short following feed cursor viewer=%s: %w", viewerUserID, err)
	}

	return r.listFollowingSnapshotPage(ctx, viewerUserID, cursor.FollowingRemainingShortIDs, pageLimit)
}

// GetDetail は public short detail を返します。
func (r *Repository) GetDetail(ctx context.Context, shortID uuid.UUID, viewerUserID *uuid.UUID) (Detail, error) {
	row, err := r.queries.GetPublicShortDetailItem(ctx, sqlc.GetPublicShortDetailItemParams{
		ShortID:      postgres.UUIDToPG(shortID),
		ViewerUserID: optionalUUIDToPG(viewerUserID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Detail{}, fmt.Errorf("public short detail 取得 short=%s: %w", shortID, ErrPublicShortNotFound)
		}

		return Detail{}, fmt.Errorf("public short detail 取得 short=%s: %w", shortID, err)
	}

	detail, err := mapDetail(row)
	if err != nil {
		return Detail{}, fmt.Errorf("public short detail 取得結果の変換 short=%s: %w", shortID, err)
	}

	return detail, nil
}

func buildRecommendedPageParams(viewerUserID *uuid.UUID, cursor *Cursor, limit int) (sqlc.ListRecommendedPublicFeedItemsParams, int) {
	limit = resolvePageLimit(limit)

	params := sqlc.ListRecommendedPublicFeedItemsParams{
		LimitCount: int32(limit + 1),
	}
	if viewerUserID != nil {
		params.ViewerUserID = optionalUUIDToPG(viewerUserID)
	}
	if cursor == nil {
		return params, limit
	}

	params.CursorPublishedAt = postgres.TimeToPG(&cursor.PublishedAt)
	params.CursorShortID = postgres.UUIDToPG(cursor.ShortID)

	return params, limit
}

func buildInitialFollowingPageParams(viewerUserID uuid.UUID, rankingReferenceAt time.Time) sqlc.ListFollowingPublicFeedItemsParams {
	return sqlc.ListFollowingPublicFeedItemsParams{
		ViewerUserID:       postgres.UUIDToPG(viewerUserID),
		RankingReferenceAt: postgres.TimeToPG(&rankingReferenceAt),
	}
}

func mapRecommendedPage(rows []sqlc.ListRecommendedPublicFeedItemsRow, limit int, label string) ([]Item, *Cursor, error) {
	items := make([]Item, 0, min(limit, len(rows)))
	for index, row := range rows {
		if index >= limit {
			break
		}

		item, err := mapFeedItem(
			mapFeedRow{
				AvatarUrl:          row.AvatarUrl,
				Bio:                row.Bio,
				CanonicalMainID:    row.CanonicalMainID,
				Caption:            row.Caption,
				CreatorUserID:      row.CreatorUserID,
				DisplayName:        row.DisplayName,
				Handle:             row.Handle,
				ID:                 row.ID,
				IsOwner:            row.IsOwner,
				IsPinned:           row.IsPinned,
				IsUnlocked:         row.IsUnlocked,
				IsFollowingCreator: row.IsFollowingCreator,
				MainDurationMs:     row.MainDurationMs,
				MainPriceMinor:     row.MainPriceMinor,
				MediaAssetID:       row.MediaAssetID,
				PublishedAt:        row.PublishedAt,
				ShortDurationMs:    row.ShortDurationMs,
			},
		)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", label, err)
		}

		items = append(items, item)
	}

	return mapFeedPageCursor(items, rows, limit)
}

func mapInitialFollowingPage(rows []sqlc.ListFollowingPublicFeedItemsRow, limit int, label string) ([]Item, *Cursor, error) {
	if len(rows) == 0 {
		return []Item{}, nil, nil
	}

	items := make([]Item, 0, min(limit, len(rows)))
	remainingShortIDs := make([]uuid.UUID, 0, max(0, len(rows)-limit))
	for index, row := range rows {
		if index >= limit {
			shortID, err := postgres.UUIDFromPG(row.ID)
			if err != nil {
				return nil, nil, fmt.Errorf("%s: public short item の short id 変換: %w", label, err)
			}

			remainingShortIDs = append(remainingShortIDs, shortID)
			continue
		}

		item, err := mapFeedItem(
			mapFeedRow{
				AvatarUrl:          row.AvatarUrl,
				Bio:                row.Bio,
				CanonicalMainID:    row.CanonicalMainID,
				Caption:            row.Caption,
				CreatorUserID:      row.CreatorUserID,
				DisplayName:        row.DisplayName,
				Handle:             row.Handle,
				ID:                 row.ID,
				IsOwner:            row.IsOwner,
				IsPinned:           row.IsPinned,
				IsUnlocked:         row.IsUnlocked,
				IsFollowingCreator: row.IsFollowingCreator,
				MainDurationMs:     row.MainDurationMs,
				MainPriceMinor:     row.MainPriceMinor,
				MediaAssetID:       row.MediaAssetID,
				PublishedAt:        row.PublishedAt,
				ShortDurationMs:    row.ShortDurationMs,
			},
		)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: %w", label, err)
		}

		items = append(items, item)
	}

	return items, buildFollowingSnapshotCursor(remainingShortIDs), nil
}

func mapFeedPageCursor[T interface {
	GetPublishedAt() time.Time
	GetShortID() uuid.UUID
}](items []T, rows any, limit int) ([]T, *Cursor, error) {
	switch {
	case len(items) == 0 && lengthOfRows(rows) == 0:
		return []T{}, nil, nil
	case lengthOfRows(rows) <= limit:
		return items, nil, nil
	}

	lastItem := items[len(items)-1]

	return items, &Cursor{
		PublishedAt: lastItem.GetPublishedAt(),
		ShortID:     lastItem.GetShortID(),
	}, nil
}

func lengthOfRows(rows any) int {
	switch typedRows := rows.(type) {
	case []sqlc.ListFeedItemsByShortIDsRow:
		return len(typedRows)
	case []sqlc.ListFollowingPublicFeedItemsRow:
		return len(typedRows)
	case []sqlc.ListRecommendedPublicFeedItemsRow:
		return len(typedRows)
	default:
		return 0
	}
}

func validateFollowingCursor(cursor *Cursor) error {
	if cursor == nil {
		return nil
	}
	if len(cursor.FollowingRemainingShortIDs) == 0 {
		return fmt.Errorf("%w: remaining short ids がありません", ErrFollowingCursorInvalid)
	}
	if !cursor.PublishedAt.IsZero() {
		return fmt.Errorf("%w: following cursor に published_at が含まれています", ErrFollowingCursorInvalid)
	}
	if cursor.ShortID != uuid.Nil {
		return fmt.Errorf("%w: following cursor に short id が含まれています", ErrFollowingCursorInvalid)
	}
	for _, shortID := range cursor.FollowingRemainingShortIDs {
		if shortID == uuid.Nil {
			return fmt.Errorf("%w: remaining short ids に nil short id が含まれます", ErrFollowingCursorInvalid)
		}
	}

	return nil
}

func resolvePageLimit(limit int) int {
	if limit <= 0 {
		return DefaultPageSize
	}

	return limit
}

func currentFollowingRankingReferenceAt() time.Time {
	return time.Now().UTC()
}

func buildFollowingSnapshotCursor(remainingShortIDs []uuid.UUID) *Cursor {
	if len(remainingShortIDs) == 0 {
		return nil
	}

	nextRemainingShortIDs := append([]uuid.UUID(nil), remainingShortIDs...)

	return &Cursor{FollowingRemainingShortIDs: nextRemainingShortIDs}
}

func (r *Repository) listFollowingSnapshotPage(
	ctx context.Context,
	viewerUserID uuid.UUID,
	remainingShortIDs []uuid.UUID,
	limit int,
) ([]Item, *Cursor, error) {
	items := make([]Item, 0, min(limit, len(remainingShortIDs)))
	consumedCount := 0

	for len(items) < limit && consumedCount < len(remainingShortIDs) {
		batchSize := min(limit-len(items), len(remainingShortIDs)-consumedCount)
		batchItems, err := r.hydrateFollowingSnapshotItems(
			ctx,
			viewerUserID,
			remainingShortIDs[consumedCount:consumedCount+batchSize],
		)
		if err != nil {
			return nil, nil, err
		}

		items = append(items, batchItems...)
		consumedCount += batchSize
	}

	return items, buildFollowingSnapshotCursor(remainingShortIDs[consumedCount:]), nil
}

func (r *Repository) hydrateFollowingSnapshotItems(
	ctx context.Context,
	viewerUserID uuid.UUID,
	shortIDs []uuid.UUID,
) ([]Item, error) {
	rows, err := r.queries.ListFeedItemsByShortIDs(ctx, sqlc.ListFeedItemsByShortIDsParams{
		ViewerUserID: postgres.UUIDToPG(viewerUserID),
		ShortIds:     uuidSliceToPG(shortIDs),
	})
	if err != nil {
		return nil, fmt.Errorf("public short following snapshot hydrate viewer=%s: %w", viewerUserID, err)
	}

	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		item, mapErr := mapFeedItem(
			mapFeedRow{
				AvatarUrl:          row.AvatarUrl,
				Bio:                row.Bio,
				CanonicalMainID:    row.CanonicalMainID,
				Caption:            row.Caption,
				CreatorUserID:      row.CreatorUserID,
				DisplayName:        row.DisplayName,
				Handle:             row.Handle,
				ID:                 row.ID,
				IsOwner:            row.IsOwner,
				IsPinned:           row.IsPinned,
				IsUnlocked:         row.IsUnlocked,
				IsFollowingCreator: row.IsFollowingCreator,
				MainDurationMs:     row.MainDurationMs,
				MainPriceMinor:     row.MainPriceMinor,
				MediaAssetID:       row.MediaAssetID,
				PublishedAt:        row.PublishedAt,
				ShortDurationMs:    row.ShortDurationMs,
			},
		)
		if mapErr != nil {
			return nil, fmt.Errorf(
				"public short following snapshot hydrate 取得結果の変換 viewer=%s: %w",
				viewerUserID,
				mapErr,
			)
		}

		items = append(items, item)
	}

	return items, nil
}

func uuidSliceToPG(values []uuid.UUID) []pgtype.UUID {
	result := make([]pgtype.UUID, 0, len(values))
	for _, value := range values {
		result = append(result, postgres.UUIDToPG(value))
	}

	return result
}

type mapFeedRow struct {
	AvatarUrl          pgtype.Text
	Bio                string
	CanonicalMainID    pgtype.UUID
	Caption            pgtype.Text
	CreatorUserID      pgtype.UUID
	DisplayName        pgtype.Text
	Handle             string
	ID                 pgtype.UUID
	IsOwner            any
	IsPinned           any
	IsUnlocked         any
	IsFollowingCreator any
	MainDurationMs     pgtype.Int8
	MainPriceMinor     int64
	MediaAssetID       pgtype.UUID
	PublishedAt        pgtype.Timestamptz
	ShortDurationMs    pgtype.Int8
}

func mapFeedItem(row mapFeedRow) (Item, error) {
	shortID, err := postgres.UUIDFromPG(row.ID)
	if err != nil {
		return Item{}, fmt.Errorf("public short item の short id 変換: %w", err)
	}
	creatorUserID, err := postgres.UUIDFromPG(row.CreatorUserID)
	if err != nil {
		return Item{}, fmt.Errorf("public short item の creator user id 変換: %w", err)
	}
	canonicalMainID, err := postgres.UUIDFromPG(row.CanonicalMainID)
	if err != nil {
		return Item{}, fmt.Errorf("public short item の canonical main id 変換: %w", err)
	}
	mediaAssetID, err := postgres.UUIDFromPG(row.MediaAssetID)
	if err != nil {
		return Item{}, fmt.Errorf("public short item の media asset id 変換: %w", err)
	}
	publishedAt, err := postgres.RequiredTimeFromPG(row.PublishedAt)
	if err != nil {
		return Item{}, fmt.Errorf("public short item の published_at 変換: %w", err)
	}

	displayName := strings.TrimSpace(row.DisplayName.String)
	handle := strings.TrimSpace(row.Handle)
	caption := ""
	if row.Caption.Valid {
		caption = strings.TrimSpace(row.Caption.String)
	}
	if !row.DisplayName.Valid || displayName == "" {
		return Item{}, fmt.Errorf("public short item の creator display_name がありません")
	}
	if handle == "" {
		return Item{}, fmt.Errorf("public short item の creator handle がありません")
	}
	if !row.ShortDurationMs.Valid || row.ShortDurationMs.Int64 <= 0 {
		return Item{}, fmt.Errorf("public short item の short duration_ms がありません")
	}
	if !row.MainDurationMs.Valid || row.MainDurationMs.Int64 <= 0 {
		return Item{}, fmt.Errorf("public short item の main duration_ms がありません")
	}
	if row.MainPriceMinor <= 0 {
		return Item{}, fmt.Errorf("public short item の main price_minor がありません")
	}

	isOwner, err := boolFromAny(row.IsOwner)
	if err != nil {
		return Item{}, fmt.Errorf("public short item の is_owner 変換: %w", err)
	}
	isPinned, err := boolFromAny(row.IsPinned)
	if err != nil {
		return Item{}, fmt.Errorf("public short item の is_pinned 変換: %w", err)
	}
	isUnlocked, err := boolFromAny(row.IsUnlocked)
	if err != nil {
		return Item{}, fmt.Errorf("public short item の is_unlocked 変換: %w", err)
	}
	isFollowingCreator, err := boolFromAny(row.IsFollowingCreator)
	if err != nil {
		return Item{}, fmt.Errorf("public short item の is_following_creator 変換: %w", err)
	}

	item := Item{
		Creator: CreatorSummary{
			AvatarURL:   postgres.OptionalTextFromPG(row.AvatarUrl),
			Bio:         row.Bio,
			DisplayName: displayName,
			Handle:      handle,
			ID:          creatorUserID,
		},
		Short: ShortSummary{
			Caption:                caption,
			CanonicalMainID:        canonicalMainID,
			CreatorUserID:          creatorUserID,
			ID:                     shortID,
			MediaAssetID:           mediaAssetID,
			PreviewDurationSeconds: (row.ShortDurationMs.Int64 + 999) / 1000,
			PublishedAt:            publishedAt,
		},
		Unlock: UnlockPreview{
			IsOwner:             isOwner,
			IsUnlocked:          isUnlocked,
			MainDurationSeconds: (row.MainDurationMs.Int64 + 999) / 1000,
			PriceJPY:            row.MainPriceMinor,
		},
	}
	item.Viewer.IsFollowingCreator = isFollowingCreator
	item.Viewer.IsPinned = isPinned

	return item, nil
}

func mapDetail(row sqlc.GetPublicShortDetailItemRow) (Detail, error) {
	item, err := mapFeedItem(
		mapFeedRow{
			AvatarUrl:          row.AvatarUrl,
			Bio:                row.Bio,
			CanonicalMainID:    row.CanonicalMainID,
			Caption:            row.Caption,
			CreatorUserID:      row.CreatorUserID,
			DisplayName:        row.DisplayName,
			Handle:             row.Handle,
			ID:                 row.ID,
			IsOwner:            row.IsOwner,
			IsPinned:           row.IsPinned,
			IsUnlocked:         row.IsUnlocked,
			IsFollowingCreator: row.IsFollowingCreator,
			MainDurationMs:     row.MainDurationMs,
			MainPriceMinor:     row.MainPriceMinor,
			MediaAssetID:       row.MediaAssetID,
			PublishedAt:        row.PublishedAt,
			ShortDurationMs:    row.ShortDurationMs,
		},
	)
	if err != nil {
		return Detail{}, err
	}

	detail := Detail{
		Item: item,
	}
	isFollowingCreator, err := boolFromAny(row.IsFollowingCreator)
	if err != nil {
		return Detail{}, fmt.Errorf("public short detail の is_following_creator 変換: %w", err)
	}
	detail.Item.Creator.IsFollowing = isFollowingCreator
	detail.Item.Viewer.IsFollowingCreator = isFollowingCreator
	detail.Viewer.IsFollowingCreator = isFollowingCreator

	return detail, nil
}

// GetPublishedAt は cursor 生成用に published_at を返します。
func (i Item) GetPublishedAt() time.Time {
	return i.Short.PublishedAt
}

// GetShortID は cursor 生成用に short id を返します。
func (i Item) GetShortID() uuid.UUID {
	return i.Short.ID
}

// GetPublishedAt は cursor 生成用に published_at を返します。
func (d Detail) GetPublishedAt() time.Time {
	return d.Item.Short.PublishedAt
}

// GetShortID は cursor 生成用に short id を返します。
func (d Detail) GetShortID() uuid.UUID {
	return d.Item.Short.ID
}

func optionalUUIDToPG(value *uuid.UUID) pgtype.UUID {
	if value == nil {
		return pgtype.UUID{}
	}

	return postgres.UUIDToPG(*value)
}

func boolFromAny(value any) (bool, error) {
	switch typedValue := value.(type) {
	case bool:
		return typedValue, nil
	case nil:
		return false, nil
	default:
		return false, fmt.Errorf("unexpected bool type %T", value)
	}
}
