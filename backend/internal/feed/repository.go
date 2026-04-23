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

const (
	// DefaultPageSize は public short feed の既定 page size です。
	DefaultPageSize = 20
	// RecommendedSnapshotMaxShortIDs は initial recommended ranking と snapshot continuation で保持する short 上限です。
	RecommendedSnapshotMaxShortIDs = 500
)

// ErrPublicShortNotFound は対象の public short が存在しないことを表します。
var ErrPublicShortNotFound = errors.New("public short が見つかりません")

// ErrRecommendedCursorInvalid は recommended feed cursor が継続取得に必要な情報を欠いていることを表します。
var ErrRecommendedCursorInvalid = errors.New("public short recommended feed cursor が不正です")

// ErrFollowingCursorInvalid は following feed cursor が継続取得に必要な情報を欠いていることを表します。
var ErrFollowingCursorInvalid = errors.New("public short following feed cursor が不正です")

type queries interface {
	GetPublicShortDetailItem(ctx context.Context, arg sqlc.GetPublicShortDetailItemParams) (sqlc.GetPublicShortDetailItemRow, error)
	ListFeedItemsByShortIDs(ctx context.Context, arg sqlc.ListFeedItemsByShortIDsParams) ([]sqlc.ListFeedItemsByShortIDsRow, error)
	ListFollowingPublicFeedItems(ctx context.Context, arg sqlc.ListFollowingPublicFeedItemsParams) ([]sqlc.ListFollowingPublicFeedItemsRow, error)
	ListLegacyRecommendedPublicFeedItems(ctx context.Context, arg sqlc.ListLegacyRecommendedPublicFeedItemsParams) ([]sqlc.ListLegacyRecommendedPublicFeedItemsRow, error)
	ListRecommendedPublicFeedShortIDs(ctx context.Context, arg sqlc.ListRecommendedPublicFeedShortIDsParams) ([]pgtype.UUID, error)
}

// Repository は fan public short feed/detail 向けの read 操作をまとめます。
type Repository struct {
	queries queries
}

// Cursor は feed keyset pagination 用の cursor です。
type Cursor struct {
	RecommendedRemainingShortIDs []uuid.UUID
	FollowingRemainingShortIDs   []uuid.UUID
	PublishedAt                  time.Time
	ShortID                      uuid.UUID
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

// Engagement は public short の公開 engagement 指標です。
type Engagement struct {
	LikeCount int64
}

// Item は public short feed/detail に共通の read model です。
type Item struct {
	Creator    CreatorSummary
	Engagement Engagement
	Short      ShortSummary
	Unlock     UnlockPreview
	Viewer     struct {
		HasLiked           bool
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
	pageLimit := resolvePageLimit(limit)
	if cursor == nil {
		rankingReferenceAt := currentRecommendedRankingReferenceAt()
		shortIDs, err := r.listRecommendedShortIDs(
			ctx,
			viewerUserID,
			rankingReferenceAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("public short recommended feed 取得: %w", err)
		}

		return r.listRecommendedSnapshotPage(
			ctx,
			viewerUserID,
			shortIDs,
			pageLimit,
			"public short recommended feed initial snapshot hydrate",
		)
	}

	if isLegacyRecommendedCursor(cursor) {
		rows, err := r.queries.ListLegacyRecommendedPublicFeedItems(
			ctx,
			buildLegacyRecommendedPageParams(viewerUserID, cursor, pageLimit),
		)
		if err != nil {
			return nil, nil, fmt.Errorf("public short legacy recommended feed 取得: %w", err)
		}

		return mapLegacyRecommendedPage(rows, pageLimit, "public short legacy recommended feed 取得結果の変換")
	}

	if err := validateRecommendedCursor(cursor); err != nil {
		if viewerUserID == nil {
			return nil, nil, fmt.Errorf("public short recommended feed cursor public: %w", err)
		}

		return nil, nil, fmt.Errorf("public short recommended feed cursor viewer=%s: %w", *viewerUserID, err)
	}

	return r.listRecommendedSnapshotPage(
		ctx,
		viewerUserID,
		cursor.RecommendedRemainingShortIDs,
		pageLimit,
		"public short recommended feed continuation snapshot hydrate",
	)
}

// ListFollowing は follow 中 creator の public short feed 1 page を返します。
func (r *Repository) ListFollowing(ctx context.Context, viewerUserID uuid.UUID, cursor *Cursor, limit int) ([]Item, *Cursor, error) {
	pageLimit := resolvePageLimit(limit)
	if cursor == nil {
		rankingReferenceAt := currentFollowingRankingReferenceAt()
		rows, err := r.queries.ListFollowingPublicFeedItems(
			ctx,
			buildInitialFollowingPageParams(viewerUserID, rankingReferenceAt, pageLimit),
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

func buildInitialRecommendedShortIDsParams(
	viewerUserID *uuid.UUID,
	rankingReferenceAt time.Time,
) sqlc.ListRecommendedPublicFeedShortIDsParams {
	params := sqlc.ListRecommendedPublicFeedShortIDsParams{
		RankingReferenceAt: postgres.TimeToPG(&rankingReferenceAt),
		LimitCount:         int32(RecommendedSnapshotMaxShortIDs),
	}
	if viewerUserID != nil {
		params.ViewerUserID = optionalUUIDToPG(viewerUserID)
	}

	return params
}

func buildLegacyRecommendedPageParams(
	viewerUserID *uuid.UUID,
	cursor *Cursor,
	limit int,
) sqlc.ListLegacyRecommendedPublicFeedItemsParams {
	params := sqlc.ListLegacyRecommendedPublicFeedItemsParams{
		CursorPublishedAt: postgres.TimeToPG(&cursor.PublishedAt),
		CursorShortID:     postgres.UUIDToPG(cursor.ShortID),
		LimitCount:        int32(limit + 1),
	}
	if viewerUserID != nil {
		params.ViewerUserID = optionalUUIDToPG(viewerUserID)
	}

	return params
}

func buildInitialFollowingPageParams(
	viewerUserID uuid.UUID,
	rankingReferenceAt time.Time,
	limit int,
) sqlc.ListFollowingPublicFeedItemsParams {
	return sqlc.ListFollowingPublicFeedItemsParams{
		DisplayLimitCount:  int32(limit),
		ViewerUserID:       postgres.UUIDToPG(viewerUserID),
		RankingReferenceAt: postgres.TimeToPG(&rankingReferenceAt),
	}
}

func mapLegacyRecommendedPage(rows []sqlc.ListLegacyRecommendedPublicFeedItemsRow, limit int, label string) ([]Item, *Cursor, error) {
	if len(rows) == 0 {
		return []Item{}, nil, nil
	}

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
				HasLiked:           row.HasLiked,
				IsPinned:           row.IsPinned,
				IsUnlocked:         row.IsUnlocked,
				IsFollowingCreator: row.IsFollowingCreator,
				LikeCount:          row.LikeCount,
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
				HasLiked:           row.HasLiked,
				IsPinned:           row.IsPinned,
				IsUnlocked:         row.IsUnlocked,
				IsFollowingCreator: row.IsFollowingCreator,
				LikeCount:          row.LikeCount,
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
	case []sqlc.ListLegacyRecommendedPublicFeedItemsRow:
		return len(typedRows)
	case []sqlc.ListFollowingPublicFeedItemsRow:
		return len(typedRows)
	default:
		return 0
	}
}

func isLegacyRecommendedCursor(cursor *Cursor) bool {
	if cursor == nil {
		return false
	}

	return len(cursor.RecommendedRemainingShortIDs) == 0 &&
		len(cursor.FollowingRemainingShortIDs) == 0 &&
		!cursor.PublishedAt.IsZero() &&
		cursor.ShortID != uuid.Nil
}

func validateRecommendedCursor(cursor *Cursor) error {
	if cursor == nil {
		return nil
	}
	if len(cursor.RecommendedRemainingShortIDs) == 0 {
		return fmt.Errorf("%w: remaining short ids がありません", ErrRecommendedCursorInvalid)
	}
	if len(cursor.FollowingRemainingShortIDs) > 0 {
		return fmt.Errorf("%w: following cursor state が混在しています", ErrRecommendedCursorInvalid)
	}
	if !cursor.PublishedAt.IsZero() {
		return fmt.Errorf("%w: recommended cursor に published_at が含まれています", ErrRecommendedCursorInvalid)
	}
	if cursor.ShortID != uuid.Nil {
		return fmt.Errorf("%w: recommended cursor に short id が含まれています", ErrRecommendedCursorInvalid)
	}
	for _, shortID := range cursor.RecommendedRemainingShortIDs {
		if shortID == uuid.Nil {
			return fmt.Errorf("%w: remaining short ids に nil short id が含まれます", ErrRecommendedCursorInvalid)
		}
	}

	return nil
}

func validateFollowingCursor(cursor *Cursor) error {
	if cursor == nil {
		return nil
	}
	if len(cursor.FollowingRemainingShortIDs) == 0 {
		return fmt.Errorf("%w: remaining short ids がありません", ErrFollowingCursorInvalid)
	}
	if len(cursor.RecommendedRemainingShortIDs) > 0 {
		return fmt.Errorf("%w: recommended cursor state が混在しています", ErrFollowingCursorInvalid)
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

func currentRecommendedRankingReferenceAt() time.Time {
	return time.Now().UTC()
}

func buildRecommendedSnapshotCursor(remainingShortIDs []uuid.UUID) *Cursor {
	if len(remainingShortIDs) == 0 {
		return nil
	}

	nextRemainingShortIDs := append([]uuid.UUID(nil), remainingShortIDs...)

	return &Cursor{RecommendedRemainingShortIDs: nextRemainingShortIDs}
}

func buildFollowingSnapshotCursor(remainingShortIDs []uuid.UUID) *Cursor {
	if len(remainingShortIDs) == 0 {
		return nil
	}

	nextRemainingShortIDs := append([]uuid.UUID(nil), remainingShortIDs...)

	return &Cursor{FollowingRemainingShortIDs: nextRemainingShortIDs}
}

func (r *Repository) listRecommendedShortIDs(
	ctx context.Context,
	viewerUserID *uuid.UUID,
	rankingReferenceAt time.Time,
) ([]uuid.UUID, error) {
	rows, err := r.queries.ListRecommendedPublicFeedShortIDs(
		ctx,
		buildInitialRecommendedShortIDsParams(viewerUserID, rankingReferenceAt),
	)
	if err != nil {
		return nil, err
	}

	shortIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		shortID, decodeErr := postgres.UUIDFromPG(row)
		if decodeErr != nil {
			return nil, fmt.Errorf("public short recommended feed short id 変換: %w", decodeErr)
		}

		shortIDs = append(shortIDs, shortID)
	}

	return shortIDs, nil
}

func (r *Repository) listRecommendedSnapshotPage(
	ctx context.Context,
	viewerUserID *uuid.UUID,
	remainingShortIDs []uuid.UUID,
	limit int,
	label string,
) ([]Item, *Cursor, error) {
	items := make([]Item, 0, min(limit, len(remainingShortIDs)))
	consumedCount := 0

	for len(items) < limit && consumedCount < len(remainingShortIDs) {
		batchSize := min(limit-len(items), len(remainingShortIDs)-consumedCount)
		batchItems, err := r.hydrateSnapshotItems(
			ctx,
			viewerUserID,
			remainingShortIDs[consumedCount:consumedCount+batchSize],
			label,
		)
		if err != nil {
			return nil, nil, err
		}

		items = append(items, batchItems...)
		consumedCount += batchSize
	}

	return items, buildRecommendedSnapshotCursor(remainingShortIDs[consumedCount:]), nil
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
		batchItems, err := r.hydrateSnapshotItems(
			ctx,
			&viewerUserID,
			remainingShortIDs[consumedCount:consumedCount+batchSize],
			fmt.Sprintf("public short following snapshot hydrate viewer=%s", viewerUserID),
		)
		if err != nil {
			return nil, nil, err
		}

		items = append(items, batchItems...)
		consumedCount += batchSize
	}

	return items, buildFollowingSnapshotCursor(remainingShortIDs[consumedCount:]), nil
}

func (r *Repository) hydrateSnapshotItems(
	ctx context.Context,
	viewerUserID *uuid.UUID,
	shortIDs []uuid.UUID,
	label string,
) ([]Item, error) {
	rows, err := r.queries.ListFeedItemsByShortIDs(ctx, sqlc.ListFeedItemsByShortIDsParams{
		ViewerUserID: optionalUUIDToPG(viewerUserID),
		ShortIds:     uuidSliceToPG(shortIDs),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
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
				HasLiked:           row.HasLiked,
				IsPinned:           row.IsPinned,
				IsUnlocked:         row.IsUnlocked,
				IsFollowingCreator: row.IsFollowingCreator,
				LikeCount:          row.LikeCount,
				MainDurationMs:     row.MainDurationMs,
				MainPriceMinor:     row.MainPriceMinor,
				MediaAssetID:       row.MediaAssetID,
				PublishedAt:        row.PublishedAt,
				ShortDurationMs:    row.ShortDurationMs,
			},
		)
		if mapErr != nil {
			return nil, fmt.Errorf("%s 取得結果の変換: %w", label, mapErr)
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
	HasLiked           any
	IsOwner            any
	IsPinned           any
	IsUnlocked         any
	IsFollowingCreator any
	LikeCount          any
	MainDurationMs     pgtype.Int8
	MainPriceMinor     any
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
	mainPriceMinor, err := priceMinorFromAny(row.MainPriceMinor)
	if err != nil {
		return Item{}, fmt.Errorf("public short item の main price_minor 変換: %w", err)
	}
	if mainPriceMinor <= 0 {
		return Item{}, fmt.Errorf("public short item の main price_minor がありません")
	}

	likeCount, err := int64FromAny(row.LikeCount)
	if err != nil {
		return Item{}, fmt.Errorf("public short item の like_count 変換: %w", err)
	}
	if likeCount < 0 {
		return Item{}, fmt.Errorf("public short item の like_count が不正です")
	}

	hasLiked, err := boolFromAny(row.HasLiked)
	if err != nil {
		return Item{}, fmt.Errorf("public short item の has_liked 変換: %w", err)
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
		Engagement: Engagement{
			LikeCount: likeCount,
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
			PriceJPY:            mainPriceMinor,
		},
	}
	item.Viewer.HasLiked = hasLiked
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
			HasLiked:           row.HasLiked,
			IsOwner:            row.IsOwner,
			IsPinned:           row.IsPinned,
			IsUnlocked:         row.IsUnlocked,
			IsFollowingCreator: row.IsFollowingCreator,
			LikeCount:          row.LikeCount,
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

func int64FromAny(value any) (int64, error) {
	switch typedValue := value.(type) {
	case int:
		return int64(typedValue), nil
	case int64:
		return typedValue, nil
	case int32:
		return int64(typedValue), nil
	case pgtype.Int8:
		if !typedValue.Valid {
			return 0, nil
		}

		return typedValue.Int64, nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unexpected int64 type %T", value)
	}
}

func priceMinorFromAny(value any) (int64, error) {
	switch typedValue := value.(type) {
	case int64:
		return typedValue, nil
	case pgtype.Int8:
		if !typedValue.Valid {
			return 0, nil
		}

		return typedValue.Int64, nil
	default:
		return 0, fmt.Errorf("unexpected price_minor type %T", value)
	}
}
