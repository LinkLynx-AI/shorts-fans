package feed

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type stubQueries struct {
	getDetail               func(context.Context, sqlc.GetPublicShortDetailItemParams) (sqlc.GetPublicShortDetailItemRow, error)
	listByShortIDs          func(context.Context, sqlc.ListFeedItemsByShortIDsParams) ([]sqlc.ListFeedItemsByShortIDsRow, error)
	listFollowing           func(context.Context, sqlc.ListFollowingPublicFeedItemsParams) ([]sqlc.ListFollowingPublicFeedItemsRow, error)
	listLegacyRecommended   func(context.Context, sqlc.ListLegacyRecommendedPublicFeedItemsParams) ([]sqlc.ListLegacyRecommendedPublicFeedItemsRow, error)
	listRecommendedShortIDs func(context.Context, sqlc.ListRecommendedPublicFeedShortIDsParams) ([]pgtype.UUID, error)
}

func (s stubQueries) GetPublicShortDetailItem(ctx context.Context, arg sqlc.GetPublicShortDetailItemParams) (sqlc.GetPublicShortDetailItemRow, error) {
	return s.getDetail(ctx, arg)
}

func (s stubQueries) ListFeedItemsByShortIDs(ctx context.Context, arg sqlc.ListFeedItemsByShortIDsParams) ([]sqlc.ListFeedItemsByShortIDsRow, error) {
	return s.listByShortIDs(ctx, arg)
}

func (s stubQueries) ListFollowingPublicFeedItems(ctx context.Context, arg sqlc.ListFollowingPublicFeedItemsParams) ([]sqlc.ListFollowingPublicFeedItemsRow, error) {
	return s.listFollowing(ctx, arg)
}

func (s stubQueries) ListLegacyRecommendedPublicFeedItems(ctx context.Context, arg sqlc.ListLegacyRecommendedPublicFeedItemsParams) ([]sqlc.ListLegacyRecommendedPublicFeedItemsRow, error) {
	return s.listLegacyRecommended(ctx, arg)
}

func (s stubQueries) ListRecommendedPublicFeedShortIDs(ctx context.Context, arg sqlc.ListRecommendedPublicFeedShortIDsParams) ([]pgtype.UUID, error) {
	return s.listRecommendedShortIDs(ctx, arg)
}

func TestNewRepository(t *testing.T) {
	t.Parallel()

	repo := NewRepository(nil)
	if repo == nil {
		t.Fatal("NewRepository() = nil, want repository")
	}
	if repo.queries == nil {
		t.Fatal("NewRepository() queries = nil, want initialized sqlc queries")
	}
}

func TestBuildInitialRecommendedShortIDsParams(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	rankingReferenceAt := time.Unix(1710000000, 0).UTC()

	params := buildInitialRecommendedShortIDsParams(&viewerID, rankingReferenceAt)
	if got, err := postgres.UUIDFromPG(params.ViewerUserID); err != nil || got != viewerID {
		t.Fatalf("buildInitialRecommendedShortIDsParams() viewer got %s err=%v want %s", got, err, viewerID)
	}
	if got, err := postgres.RequiredTimeFromPG(params.RankingReferenceAt); err != nil || !got.Equal(rankingReferenceAt) {
		t.Fatalf("buildInitialRecommendedShortIDsParams() ranking reference got %s err=%v want %s", got, err, rankingReferenceAt)
	}
	if params.LimitCount != int32(RecommendedSnapshotMaxShortIDs) {
		t.Fatalf(
			"buildInitialRecommendedShortIDsParams() limit count got %d want %d",
			params.LimitCount,
			RecommendedSnapshotMaxShortIDs,
		)
	}

	params = buildInitialRecommendedShortIDsParams(nil, rankingReferenceAt)
	if params.ViewerUserID.Valid {
		t.Fatalf("buildInitialRecommendedShortIDsParams() viewer valid got %t want false", params.ViewerUserID.Valid)
	}
	if got, err := postgres.RequiredTimeFromPG(params.RankingReferenceAt); err != nil || !got.Equal(rankingReferenceAt) {
		t.Fatalf("buildInitialRecommendedShortIDsParams() public ranking reference got %s err=%v want %s", got, err, rankingReferenceAt)
	}
	if params.LimitCount != int32(RecommendedSnapshotMaxShortIDs) {
		t.Fatalf(
			"buildInitialRecommendedShortIDsParams() public limit count got %d want %d",
			params.LimitCount,
			RecommendedSnapshotMaxShortIDs,
		)
	}
}

func TestBuildLegacyRecommendedPageParams(t *testing.T) {
	t.Parallel()

	publishedAt := time.Unix(1710000300, 0).UTC()
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	params := buildLegacyRecommendedPageParams(nil, &Cursor{
		PublishedAt: publishedAt,
		ShortID:     shortID,
	}, 10)
	if params.ViewerUserID.Valid {
		t.Fatalf("buildLegacyRecommendedPageParams() viewer valid got %t want false", params.ViewerUserID.Valid)
	}
	if got, err := postgres.RequiredTimeFromPG(params.CursorPublishedAt); err != nil || !got.Equal(publishedAt) {
		t.Fatalf("buildLegacyRecommendedPageParams() publishedAt got %s err=%v want %s", got, err, publishedAt)
	}
	if got, err := postgres.UUIDFromPG(params.CursorShortID); err != nil || got != shortID {
		t.Fatalf("buildLegacyRecommendedPageParams() shortID got %s err=%v want %s", got, err, shortID)
	}
	if params.LimitCount != 11 {
		t.Fatalf("buildLegacyRecommendedPageParams() limit count got %d want %d", params.LimitCount, 11)
	}
}

func TestBuildInitialFollowingPageParams(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	rankingReferenceAt := time.Unix(1710003600, 0).UTC()

	params := buildInitialFollowingPageParams(viewerID, rankingReferenceAt, 12)
	if got, err := postgres.UUIDFromPG(params.ViewerUserID); err != nil || got != viewerID {
		t.Fatalf("buildInitialFollowingPageParams() viewer got %s err=%v want %s", got, err, viewerID)
	}
	if got, err := postgres.RequiredTimeFromPG(params.RankingReferenceAt); err != nil || !got.Equal(rankingReferenceAt) {
		t.Fatalf("buildInitialFollowingPageParams() ranking reference got %s err=%v want %s", got, err, rankingReferenceAt)
	}
	if params.DisplayLimitCount != 12 {
		t.Fatalf("buildInitialFollowingPageParams() display limit got %d want %d", params.DisplayLimitCount, 12)
	}
}

func TestValidateRecommendedCursor(t *testing.T) {
	t.Parallel()

	if err := validateRecommendedCursor(nil); err != nil {
		t.Fatalf("validateRecommendedCursor(nil) error = %v, want nil", err)
	}

	validCursor := &Cursor{
		RecommendedRemainingShortIDs: []uuid.UUID{
			uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		},
	}
	if err := validateRecommendedCursor(validCursor); err != nil {
		t.Fatalf("validateRecommendedCursor(valid cursor) error = %v, want nil", err)
	}

	missingRemainingShortIDs := &Cursor{}
	if err := validateRecommendedCursor(missingRemainingShortIDs); !errors.Is(err, ErrRecommendedCursorInvalid) {
		t.Fatalf("validateRecommendedCursor(missing remaining short ids) error got %v want %v", err, ErrRecommendedCursorInvalid)
	}

	mixedState := &Cursor{
		RecommendedRemainingShortIDs: []uuid.UUID{
			uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		},
		FollowingRemainingShortIDs: []uuid.UUID{
			uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		},
	}
	if err := validateRecommendedCursor(mixedState); !errors.Is(err, ErrRecommendedCursorInvalid) {
		t.Fatalf("validateRecommendedCursor(mixed state) error got %v want %v", err, ErrRecommendedCursorInvalid)
	}
}

func TestValidateFollowingCursor(t *testing.T) {
	t.Parallel()

	if err := validateFollowingCursor(nil); err != nil {
		t.Fatalf("validateFollowingCursor(nil) error = %v, want nil", err)
	}

	missingRemainingShortIDs := &Cursor{}
	if err := validateFollowingCursor(missingRemainingShortIDs); !errors.Is(err, ErrFollowingCursorInvalid) {
		t.Fatalf("validateFollowingCursor(missing remaining short ids) error got %v want %v", err, ErrFollowingCursorInvalid)
	}

	nilShortID := &Cursor{
		FollowingRemainingShortIDs: []uuid.UUID{
			uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			uuid.Nil,
		},
	}
	if err := validateFollowingCursor(nilShortID); !errors.Is(err, ErrFollowingCursorInvalid) {
		t.Fatalf("validateFollowingCursor(nil short id) error got %v want %v", err, ErrFollowingCursorInvalid)
	}

	mixedState := &Cursor{
		FollowingRemainingShortIDs: []uuid.UUID{
			uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		},
		PublishedAt: time.Unix(1710000200, 0).UTC(),
		ShortID:     uuid.MustParse("33333333-3333-3333-3333-333333333333"),
	}
	if err := validateFollowingCursor(mixedState); !errors.Is(err, ErrFollowingCursorInvalid) {
		t.Fatalf("validateFollowingCursor(mixed state) error got %v want %v", err, ErrFollowingCursorInvalid)
	}
}

func TestMapFeedItem(t *testing.T) {
	t.Parallel()

	row := makeMapFeedRow()

	item, err := mapFeedItem(row)
	if err != nil {
		t.Fatalf("mapFeedItem() error = %v, want nil", err)
	}
	if item.Creator.AvatarURL == nil || *item.Creator.AvatarURL != "https://cdn.example.com/avatar.jpg" {
		t.Fatalf("mapFeedItem() avatar got %v want avatar url", item.Creator.AvatarURL)
	}
	if item.Creator.DisplayName != "Mina Rei" {
		t.Fatalf("mapFeedItem() displayName got %q want %q", item.Creator.DisplayName, "Mina Rei")
	}
	if item.Short.Caption != "quiet rooftop preview" {
		t.Fatalf("mapFeedItem() caption got %q want %q", item.Short.Caption, "quiet rooftop preview")
	}
	if item.Short.PreviewDurationSeconds != 17 {
		t.Fatalf("mapFeedItem() previewDurationSeconds got %d want %d", item.Short.PreviewDurationSeconds, 17)
	}
	if item.Unlock.MainDurationSeconds != 481 {
		t.Fatalf("mapFeedItem() mainDurationSeconds got %d want %d", item.Unlock.MainDurationSeconds, 481)
	}
	if item.Engagement.LikeCount != 12 {
		t.Fatalf("mapFeedItem() likeCount got %d want %d", item.Engagement.LikeCount, 12)
	}
	if !item.Unlock.IsOwner || !item.Unlock.IsUnlocked || !item.Viewer.IsFollowingCreator || !item.Viewer.IsPinned || !item.Viewer.HasLiked {
		t.Fatalf("mapFeedItem() booleans got owner=%t unlocked=%t following=%t pinned=%t liked=%t want true/true/true/true/true", item.Unlock.IsOwner, item.Unlock.IsUnlocked, item.Viewer.IsFollowingCreator, item.Viewer.IsPinned, item.Viewer.HasLiked)
	}
	if !item.GetPublishedAt().Equal(row.PublishedAt.Time) {
		t.Fatalf("item.GetPublishedAt() got %s want %s", item.GetPublishedAt(), row.PublishedAt.Time)
	}
	if got, err := postgres.UUIDFromPG(row.ID); err != nil || item.GetShortID() != got {
		t.Fatalf("item.GetShortID() got %s err=%v want row id", item.GetShortID(), err)
	}
}

func TestMapFeedItemNormalizesMissingCaption(t *testing.T) {
	t.Parallel()

	row := makeMapFeedRow()
	row.Caption = pgtype.Text{}

	item, err := mapFeedItem(row)
	if err != nil {
		t.Fatalf("mapFeedItem() error = %v, want nil", err)
	}
	if item.Short.Caption != "" {
		t.Fatalf("mapFeedItem() caption got %q want empty string", item.Short.Caption)
	}

	row = makeMapFeedRow()
	row.Caption = makeText("   ")

	item, err = mapFeedItem(row)
	if err != nil {
		t.Fatalf("mapFeedItem() error = %v, want nil", err)
	}
	if item.Short.Caption != "" {
		t.Fatalf("mapFeedItem() blank caption got %q want empty string", item.Short.Caption)
	}
}

func TestMapFeedItemRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	row := makeMapFeedRow()
	row.IsOwner = "true"
	if _, err := mapFeedItem(row); err == nil {
		t.Fatal("mapFeedItem() error = nil, want invalid bool type")
	}

	row = makeMapFeedRow()
	row.Handle = " "
	if _, err := mapFeedItem(row); err == nil {
		t.Fatal("mapFeedItem() error = nil, want missing handle")
	}

	row = makeMapFeedRow()
	row.ShortDurationMs = pgtype.Int8{}
	if _, err := mapFeedItem(row); err == nil {
		t.Fatal("mapFeedItem() error = nil, want missing short duration")
	}
}

func TestBuildRecommendedSnapshotCursor(t *testing.T) {
	t.Parallel()

	firstID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	secondID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	nextCursor := buildRecommendedSnapshotCursor([]uuid.UUID{firstID, secondID})
	if nextCursor == nil {
		t.Fatal("buildRecommendedSnapshotCursor() = nil, want snapshot cursor")
	}
	if len(nextCursor.RecommendedRemainingShortIDs) != 2 || nextCursor.RecommendedRemainingShortIDs[0] != firstID || nextCursor.RecommendedRemainingShortIDs[1] != secondID {
		t.Fatalf("buildRecommendedSnapshotCursor() got %#v want [%s %s]", nextCursor, firstID, secondID)
	}
}

func TestMapInitialFollowingPageAndCursorHelpers(t *testing.T) {
	t.Parallel()

	firstID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	secondID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	rows := []sqlc.ListFollowingPublicFeedItemsRow{
		makeFollowingRow(firstID, time.Unix(1710000200, 0).UTC(), 765432),
		makeFollowingRow(secondID, time.Unix(1710000100, 0).UTC(), 654321),
	}

	items, nextCursor, err := mapInitialFollowingPage(rows, 1, "following")
	if err != nil {
		t.Fatalf("mapInitialFollowingPage() error = %v, want nil", err)
	}
	if len(items) != 1 {
		t.Fatalf("mapInitialFollowingPage() items len got %d want %d", len(items), 1)
	}
	if nextCursor == nil {
		t.Fatal("mapInitialFollowingPage() next cursor = nil, want snapshot cursor")
	}
	if len(nextCursor.FollowingRemainingShortIDs) != 1 || nextCursor.FollowingRemainingShortIDs[0] != secondID {
		t.Fatalf(
			"mapInitialFollowingPage() next cursor remaining got %#v want [%s]",
			nextCursor.FollowingRemainingShortIDs,
			secondID,
		)
	}

	emptyItems, emptyCursor, err := mapFeedPageCursor([]Item{}, []sqlc.ListLegacyRecommendedPublicFeedItemsRow{}, 1)
	if err != nil {
		t.Fatalf("mapFeedPageCursor() error = %v, want nil", err)
	}
	if len(emptyItems) != 0 || emptyCursor != nil {
		t.Fatalf("mapFeedPageCursor() got items=%#v cursor=%#v want empty nil", emptyItems, emptyCursor)
	}

	if got := lengthOfRows(rows); got != 2 {
		t.Fatalf("lengthOfRows(following) got %d want %d", got, 2)
	}
	publishedAt := time.Unix(1710000200, 0).UTC()
	if got := lengthOfRows([]sqlc.ListLegacyRecommendedPublicFeedItemsRow{makeLegacyRecommendedRow(firstID, publishedAt)}); got != 1 {
		t.Fatalf("lengthOfRows(legacy recommended) got %d want %d", got, 1)
	}
	if got := lengthOfRows(struct{}{}); got != 0 {
		t.Fatalf("lengthOfRows(unknown) got %d want %d", got, 0)
	}
}

func TestMapPageFunctionsWrapMappingErrors(t *testing.T) {
	t.Parallel()

	badRecommended := makeLegacyRecommendedRow(
		uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		time.Unix(1710000200, 0).UTC(),
	)
	badRecommended.Handle = " "
	if _, _, err := mapLegacyRecommendedPage([]sqlc.ListLegacyRecommendedPublicFeedItemsRow{badRecommended}, 1, "recommended"); err == nil {
		t.Fatal("mapLegacyRecommendedPage() error = nil, want wrapped mapping error")
	}

	badFollowing := makeFollowingRow(
		uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		time.Unix(1710000200, 0).UTC(),
		765432,
	)
	badFollowing.Handle = " "
	if _, _, err := mapInitialFollowingPage([]sqlc.ListFollowingPublicFeedItemsRow{badFollowing}, 1, "following"); err == nil {
		t.Fatal("mapInitialFollowingPage() error = nil, want wrapped mapping error")
	}
}

func TestMapInitialFollowingPageDoesNotMapHiddenRows(t *testing.T) {
	t.Parallel()

	firstID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	secondID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	rows := []sqlc.ListFollowingPublicFeedItemsRow{
		makeFollowingRow(firstID, time.Unix(1710000200, 0).UTC(), 765432),
		makeFollowingRow(secondID, time.Unix(1710000100, 0).UTC(), 654321),
	}
	rows[1].Handle = " "

	items, nextCursor, err := mapInitialFollowingPage(rows, 1, "following")
	if err != nil {
		t.Fatalf("mapInitialFollowingPage() error = %v, want nil", err)
	}
	if len(items) != 1 || items[0].Short.ID != firstID {
		t.Fatalf("mapInitialFollowingPage() items got %#v want first short %s", items, firstID)
	}
	if nextCursor == nil || len(nextCursor.FollowingRemainingShortIDs) != 1 || nextCursor.FollowingRemainingShortIDs[0] != secondID {
		t.Fatalf("mapInitialFollowingPage() next cursor got %#v want remaining [%s]", nextCursor, secondID)
	}
}

func TestMapDetail(t *testing.T) {
	t.Parallel()

	row := sqlc.GetPublicShortDetailItemRow{
		ID:                 makeUUID("22222222-2222-2222-2222-222222222222"),
		CreatorUserID:      makeUUID("11111111-1111-1111-1111-111111111111"),
		CanonicalMainID:    makeUUID("33333333-3333-3333-3333-333333333333"),
		MediaAssetID:       makeUUID("44444444-4444-4444-4444-444444444444"),
		Caption:            makeText("quiet rooftop preview"),
		PublishedAt:        makeTimestamp(time.Unix(1710000000, 0).UTC()),
		ShortDurationMs:    makeInt8(16500),
		DisplayName:        makeText("Mina Rei"),
		Handle:             "minarei",
		AvatarUrl:          makeText("https://cdn.example.com/avatar.jpg"),
		Bio:                "night preview specialist",
		MainPriceMinor:     int64(1800),
		MainDurationMs:     makeInt8(480500),
		IsPinned:           true,
		IsUnlocked:         false,
		IsOwner:            false,
		IsFollowingCreator: true,
	}

	detail, err := mapDetail(row)
	if err != nil {
		t.Fatalf("mapDetail() error = %v, want nil", err)
	}
	if !detail.Viewer.IsFollowingCreator || !detail.Item.Creator.IsFollowing {
		t.Fatalf("mapDetail() following got viewer=%t creator=%t want true/true", detail.Viewer.IsFollowingCreator, detail.Item.Creator.IsFollowing)
	}
	if !detail.GetPublishedAt().Equal(row.PublishedAt.Time) {
		t.Fatalf("detail.GetPublishedAt() got %s want %s", detail.GetPublishedAt(), row.PublishedAt.Time)
	}
	if got, err := postgres.UUIDFromPG(row.ID); err != nil || detail.GetShortID() != got {
		t.Fatalf("detail.GetShortID() got %s err=%v want row id", detail.GetShortID(), err)
	}

	row.IsFollowingCreator = "yes"
	if _, err := mapDetail(row); err == nil {
		t.Fatal("mapDetail() error = nil, want invalid following bool")
	}
}

func TestRepositoryListRecommendedAndFollowing(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	publishedAt := time.Unix(1710000300, 0).UTC()

	repo := &Repository{
		queries: stubQueries{
			listRecommendedShortIDs: func(_ context.Context, arg sqlc.ListRecommendedPublicFeedShortIDsParams) ([]pgtype.UUID, error) {
				if got, err := postgres.UUIDFromPG(arg.ViewerUserID); err != nil || got != viewerID {
					t.Fatalf("ListRecommendedPublicFeedShortIDs() viewer got %s err=%v want %s", got, err, viewerID)
				}
				if !arg.RankingReferenceAt.Valid {
					t.Fatal("ListRecommendedPublicFeedShortIDs() ranking reference valid = false, want true")
				}
				if arg.LimitCount != int32(RecommendedSnapshotMaxShortIDs) {
					t.Fatalf("ListRecommendedPublicFeedShortIDs() limit count got %d want %d", arg.LimitCount, RecommendedSnapshotMaxShortIDs)
				}
				return []pgtype.UUID{
					postgres.UUIDToPG(shortID),
				}, nil
			},
			listByShortIDs: func(_ context.Context, arg sqlc.ListFeedItemsByShortIDsParams) ([]sqlc.ListFeedItemsByShortIDsRow, error) {
				if got, err := postgres.UUIDFromPG(arg.ViewerUserID); err != nil || got != viewerID {
					t.Fatalf("ListFeedItemsByShortIDs() viewer got %s err=%v want %s", got, err, viewerID)
				}
				if len(arg.ShortIds) != 1 {
					t.Fatalf("ListFeedItemsByShortIDs() short ids len got %d want %d", len(arg.ShortIds), 1)
				}
				if got, err := postgres.UUIDFromPG(arg.ShortIds[0]); err != nil || got != shortID {
					t.Fatalf("ListFeedItemsByShortIDs() short got %s err=%v want %s", got, err, shortID)
				}
				return []sqlc.ListFeedItemsByShortIDsRow{
					makeHydratedFollowingRow(shortID, publishedAt),
				}, nil
			},
			listLegacyRecommended: func(context.Context, sqlc.ListLegacyRecommendedPublicFeedItemsParams) ([]sqlc.ListLegacyRecommendedPublicFeedItemsRow, error) {
				t.Fatal("ListLegacyRecommendedPublicFeedItems() was called unexpectedly")
				return nil, nil
			},
			listFollowing: func(_ context.Context, arg sqlc.ListFollowingPublicFeedItemsParams) ([]sqlc.ListFollowingPublicFeedItemsRow, error) {
				if got, err := postgres.UUIDFromPG(arg.ViewerUserID); err != nil || got != viewerID {
					t.Fatalf("ListFollowingPublicFeedItems() viewer got %s err=%v want %s", got, err, viewerID)
				}
				if !arg.RankingReferenceAt.Valid {
					t.Fatal("ListFollowingPublicFeedItems() ranking reference valid = false, want true")
				}

				return []sqlc.ListFollowingPublicFeedItemsRow{
					makeFollowingRow(shortID, publishedAt, 765432),
				}, nil
			},
		},
	}

	recommendedItems, recommendedCursor, err := repo.ListRecommended(context.Background(), &viewerID, nil, 1)
	if err != nil {
		t.Fatalf("ListRecommended() error = %v, want nil", err)
	}
	if len(recommendedItems) != 1 || recommendedCursor != nil {
		t.Fatalf("ListRecommended() got items=%d cursor=%#v want 1 nil", len(recommendedItems), recommendedCursor)
	}

	followingItems, followingCursor, err := repo.ListFollowing(context.Background(), viewerID, nil, 1)
	if err != nil {
		t.Fatalf("ListFollowing() error = %v, want nil", err)
	}
	if len(followingItems) != 1 || followingCursor != nil {
		t.Fatalf("ListFollowing() got items=%d cursor=%#v want 1 nil", len(followingItems), followingCursor)
	}
}

func TestRepositoryListRecommendedContinuationHydratesSnapshot(t *testing.T) {
	t.Parallel()

	firstShortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	secondShortID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	publishedAt := time.Unix(1710000300, 0).UTC()

	repo := &Repository{
		queries: stubQueries{
			listByShortIDs: func(_ context.Context, arg sqlc.ListFeedItemsByShortIDsParams) ([]sqlc.ListFeedItemsByShortIDsRow, error) {
				if arg.ViewerUserID.Valid {
					t.Fatalf("ListFeedItemsByShortIDs() viewer valid got %t want false", arg.ViewerUserID.Valid)
				}
				if len(arg.ShortIds) != 1 {
					t.Fatalf("ListFeedItemsByShortIDs() short ids len got %d want %d", len(arg.ShortIds), 1)
				}
				if got, err := postgres.UUIDFromPG(arg.ShortIds[0]); err != nil || got != firstShortID {
					t.Fatalf("ListFeedItemsByShortIDs() first short got %s err=%v want %s", got, err, firstShortID)
				}
				return []sqlc.ListFeedItemsByShortIDsRow{
					makeHydratedFollowingRow(firstShortID, publishedAt),
				}, nil
			},
			listLegacyRecommended: func(context.Context, sqlc.ListLegacyRecommendedPublicFeedItemsParams) ([]sqlc.ListLegacyRecommendedPublicFeedItemsRow, error) {
				t.Fatal("ListLegacyRecommendedPublicFeedItems() was called unexpectedly")
				return nil, nil
			},
		},
	}

	items, nextCursor, err := repo.ListRecommended(context.Background(), nil, &Cursor{
		RecommendedRemainingShortIDs: []uuid.UUID{firstShortID, secondShortID},
	}, 1)
	if err != nil {
		t.Fatalf("ListRecommended() error = %v, want nil", err)
	}
	if len(items) != 1 || items[0].Short.ID != firstShortID {
		t.Fatalf("ListRecommended() got items=%#v want short %s", items, firstShortID)
	}
	if nextCursor == nil || len(nextCursor.RecommendedRemainingShortIDs) != 1 || nextCursor.RecommendedRemainingShortIDs[0] != secondShortID {
		t.Fatalf("ListRecommended() next cursor got %#v want remaining [%s]", nextCursor, secondShortID)
	}
}

func TestRepositoryListRecommendedSupportsLegacyCursor(t *testing.T) {
	t.Parallel()

	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	publishedAt := time.Unix(1710000300, 0).UTC()

	repo := &Repository{
		queries: stubQueries{
			listRecommendedShortIDs: func(context.Context, sqlc.ListRecommendedPublicFeedShortIDsParams) ([]pgtype.UUID, error) {
				t.Fatal("ListRecommendedPublicFeedShortIDs() was called unexpectedly")
				return nil, nil
			},
			listLegacyRecommended: func(_ context.Context, arg sqlc.ListLegacyRecommendedPublicFeedItemsParams) ([]sqlc.ListLegacyRecommendedPublicFeedItemsRow, error) {
				if got, err := postgres.RequiredTimeFromPG(arg.CursorPublishedAt); err != nil || !got.Equal(publishedAt) {
					t.Fatalf("ListLegacyRecommendedPublicFeedItems() cursor publishedAt got %s err=%v want %s", got, err, publishedAt)
				}
				if got, err := postgres.UUIDFromPG(arg.CursorShortID); err != nil || got != shortID {
					t.Fatalf("ListLegacyRecommendedPublicFeedItems() cursor shortID got %s err=%v want %s", got, err, shortID)
				}
				if arg.LimitCount != 2 {
					t.Fatalf("ListLegacyRecommendedPublicFeedItems() limit count got %d want %d", arg.LimitCount, 2)
				}

				return []sqlc.ListLegacyRecommendedPublicFeedItemsRow{
					makeLegacyRecommendedRow(shortID, publishedAt),
				}, nil
			},
		},
	}

	items, nextCursor, err := repo.ListRecommended(context.Background(), nil, &Cursor{
		PublishedAt: publishedAt,
		ShortID:     shortID,
	}, 1)
	if err != nil {
		t.Fatalf("ListRecommended() error = %v, want nil", err)
	}
	if len(items) != 1 || items[0].Short.ID != shortID {
		t.Fatalf("ListRecommended() got items=%#v want short %s", items, shortID)
	}
	if nextCursor != nil {
		t.Fatalf("ListRecommended() next cursor got %#v want nil", nextCursor)
	}
}

func TestRepositoryListFollowingContinuationHydratesSnapshot(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	firstShortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	secondShortID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	publishedAt := time.Unix(1710000300, 0).UTC()

	repo := &Repository{
		queries: stubQueries{
			listByShortIDs: func(_ context.Context, arg sqlc.ListFeedItemsByShortIDsParams) ([]sqlc.ListFeedItemsByShortIDsRow, error) {
				if got, err := postgres.UUIDFromPG(arg.ViewerUserID); err != nil || got != viewerID {
					t.Fatalf("ListFeedItemsByShortIDs() viewer got %s err=%v want %s", got, err, viewerID)
				}
				if len(arg.ShortIds) != 1 {
					t.Fatalf("ListFeedItemsByShortIDs() short ids len got %d want %d", len(arg.ShortIds), 1)
				}
				if got, err := postgres.UUIDFromPG(arg.ShortIds[0]); err != nil || got != firstShortID {
					t.Fatalf("ListFeedItemsByShortIDs() first short got %s err=%v want %s", got, err, firstShortID)
				}

				return []sqlc.ListFeedItemsByShortIDsRow{
					makeHydratedFollowingRow(firstShortID, publishedAt),
				}, nil
			},
			listFollowing: func(context.Context, sqlc.ListFollowingPublicFeedItemsParams) ([]sqlc.ListFollowingPublicFeedItemsRow, error) {
				t.Fatal("ListFollowingPublicFeedItems() was called unexpectedly")
				return nil, nil
			},
		},
	}

	items, nextCursor, err := repo.ListFollowing(context.Background(), viewerID, &Cursor{
		FollowingRemainingShortIDs: []uuid.UUID{firstShortID, secondShortID},
	}, 1)
	if err != nil {
		t.Fatalf("ListFollowing() error = %v, want nil", err)
	}
	if len(items) != 1 || items[0].Short.ID != firstShortID {
		t.Fatalf("ListFollowing() got items=%#v want short %s", items, firstShortID)
	}
	if nextCursor == nil || len(nextCursor.FollowingRemainingShortIDs) != 1 || nextCursor.FollowingRemainingShortIDs[0] != secondShortID {
		t.Fatalf("ListFollowing() next cursor got %#v want remaining [%s]", nextCursor, secondShortID)
	}
}

func TestRepositoryListFollowingContinuationSkipsMissingSnapshotItems(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	missingShortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	availableShortID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	publishedAt := time.Unix(1710000300, 0).UTC()
	callCount := 0

	repo := &Repository{
		queries: stubQueries{
			listByShortIDs: func(_ context.Context, arg sqlc.ListFeedItemsByShortIDsParams) ([]sqlc.ListFeedItemsByShortIDsRow, error) {
				callCount++
				if got, err := postgres.UUIDFromPG(arg.ViewerUserID); err != nil || got != viewerID {
					t.Fatalf("ListFeedItemsByShortIDs() viewer got %s err=%v want %s", got, err, viewerID)
				}
				if len(arg.ShortIds) != 1 {
					t.Fatalf("ListFeedItemsByShortIDs() short ids len got %d want %d", len(arg.ShortIds), 1)
				}
				gotShortID, err := postgres.UUIDFromPG(arg.ShortIds[0])
				if err != nil {
					t.Fatalf("ListFeedItemsByShortIDs() short id decode error = %v, want nil", err)
				}
				switch callCount {
				case 1:
					if gotShortID != missingShortID {
						t.Fatalf("ListFeedItemsByShortIDs() first short got %s want %s", gotShortID, missingShortID)
					}
					return []sqlc.ListFeedItemsByShortIDsRow{}, nil
				case 2:
					if gotShortID != availableShortID {
						t.Fatalf("ListFeedItemsByShortIDs() second short got %s want %s", gotShortID, availableShortID)
					}
					return []sqlc.ListFeedItemsByShortIDsRow{
						makeHydratedFollowingRow(availableShortID, publishedAt),
					}, nil
				default:
					t.Fatalf("ListFeedItemsByShortIDs() callCount got %d want <= 2", callCount)
					return nil, nil
				}
			},
			listFollowing: func(context.Context, sqlc.ListFollowingPublicFeedItemsParams) ([]sqlc.ListFollowingPublicFeedItemsRow, error) {
				t.Fatal("ListFollowingPublicFeedItems() was called unexpectedly")
				return nil, nil
			},
		},
	}

	items, nextCursor, err := repo.ListFollowing(context.Background(), viewerID, &Cursor{
		FollowingRemainingShortIDs: []uuid.UUID{missingShortID, availableShortID},
	}, 1)
	if err != nil {
		t.Fatalf("ListFollowing() error = %v, want nil", err)
	}
	if len(items) != 1 || items[0].Short.ID != availableShortID {
		t.Fatalf("ListFollowing() got items=%#v want short %s", items, availableShortID)
	}
	if nextCursor != nil {
		t.Fatalf("ListFollowing() next cursor got %#v want nil", nextCursor)
	}
}

func TestRepositoryErrorWrapping(t *testing.T) {
	t.Parallel()

	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	queryErr := errors.New("db down")

	repo := &Repository{
		queries: stubQueries{
			getDetail: func(_ context.Context, arg sqlc.GetPublicShortDetailItemParams) (sqlc.GetPublicShortDetailItemRow, error) {
				if got, err := postgres.UUIDFromPG(arg.ShortID); err != nil || got != shortID {
					t.Fatalf("GetPublicShortDetailItem() short got %s err=%v want %s", got, err, shortID)
				}

				return sqlc.GetPublicShortDetailItemRow{}, pgx.ErrNoRows
			},
			listRecommendedShortIDs: func(context.Context, sqlc.ListRecommendedPublicFeedShortIDsParams) ([]pgtype.UUID, error) {
				return nil, queryErr
			},
			listFollowing: func(context.Context, sqlc.ListFollowingPublicFeedItemsParams) ([]sqlc.ListFollowingPublicFeedItemsRow, error) {
				return nil, queryErr
			},
			listByShortIDs: func(context.Context, sqlc.ListFeedItemsByShortIDsParams) ([]sqlc.ListFeedItemsByShortIDsRow, error) {
				return nil, queryErr
			},
		},
	}

	if _, _, err := repo.ListRecommended(context.Background(), nil, nil, 1); !errors.Is(err, queryErr) {
		t.Fatalf("ListRecommended() error got %v want wrapped %v", err, queryErr)
	}
	if _, _, err := repo.ListFollowing(context.Background(), uuid.MustParse("11111111-1111-1111-1111-111111111111"), nil, 1); !errors.Is(err, queryErr) {
		t.Fatalf("ListFollowing() error got %v want wrapped %v", err, queryErr)
	}
	if _, _, err := repo.ListFollowing(
		context.Background(),
		uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		&Cursor{FollowingRemainingShortIDs: []uuid.UUID{shortID}},
		1,
	); !errors.Is(err, queryErr) {
		t.Fatalf("ListFollowing(snapshot) error got %v want wrapped %v", err, queryErr)
	}
	if _, err := repo.GetDetail(context.Background(), shortID, nil); !errors.Is(err, ErrPublicShortNotFound) {
		t.Fatalf("GetDetail() error got %v want wrapped %v", err, ErrPublicShortNotFound)
	}

	repo.queries = stubQueries{
		getDetail: func(context.Context, sqlc.GetPublicShortDetailItemParams) (sqlc.GetPublicShortDetailItemRow, error) {
			return sqlc.GetPublicShortDetailItemRow{}, queryErr
		},
	}
	if _, err := repo.GetDetail(context.Background(), shortID, nil); !errors.Is(err, queryErr) {
		t.Fatalf("GetDetail() error got %v want wrapped %v", err, queryErr)
	}

	repo.queries = stubQueries{
		getDetail: func(context.Context, sqlc.GetPublicShortDetailItemParams) (sqlc.GetPublicShortDetailItemRow, error) {
			row := sqlc.GetPublicShortDetailItemRow{
				ID:                 makeUUID("22222222-2222-2222-2222-222222222222"),
				CreatorUserID:      makeUUID("11111111-1111-1111-1111-111111111111"),
				CanonicalMainID:    makeUUID("33333333-3333-3333-3333-333333333333"),
				MediaAssetID:       makeUUID("44444444-4444-4444-4444-444444444444"),
				Caption:            makeText("quiet rooftop preview"),
				PublishedAt:        makeTimestamp(time.Unix(1710000000, 0).UTC()),
				ShortDurationMs:    makeInt8(16500),
				DisplayName:        makeText("Mina Rei"),
				Handle:             "minarei",
				MainPriceMinor:     1800,
				MainDurationMs:     makeInt8(480500),
				LikeCount:          4,
				HasLiked:           true,
				IsPinned:           true,
				IsUnlocked:         true,
				IsOwner:            true,
				IsFollowingCreator: true,
			}
			row.MediaAssetID = pgtype.UUID{}

			return row, nil
		},
	}
	if _, err := repo.GetDetail(context.Background(), shortID, nil); err == nil {
		t.Fatal("GetDetail() error = nil, want mapping error")
	}
}

func TestRepositoryGetDetailSuccess(t *testing.T) {
	t.Parallel()

	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	repo := &Repository{
		queries: stubQueries{
			getDetail: func(context.Context, sqlc.GetPublicShortDetailItemParams) (sqlc.GetPublicShortDetailItemRow, error) {
				return sqlc.GetPublicShortDetailItemRow{
					ID:                 makeUUID("22222222-2222-2222-2222-222222222222"),
					CreatorUserID:      makeUUID("11111111-1111-1111-1111-111111111111"),
					CanonicalMainID:    makeUUID("33333333-3333-3333-3333-333333333333"),
					MediaAssetID:       makeUUID("44444444-4444-4444-4444-444444444444"),
					Caption:            makeText("quiet rooftop preview"),
					PublishedAt:        makeTimestamp(time.Unix(1710000000, 0).UTC()),
					ShortDurationMs:    makeInt8(16500),
					DisplayName:        makeText("Mina Rei"),
					Handle:             "minarei",
					AvatarUrl:          makeText("https://cdn.example.com/avatar.jpg"),
					Bio:                "night preview specialist",
					MainPriceMinor:     1800,
					MainDurationMs:     makeInt8(480500),
					LikeCount:          4,
					HasLiked:           true,
					IsPinned:           true,
					IsUnlocked:         false,
					IsOwner:            false,
					IsFollowingCreator: true,
				}, nil
			},
		},
	}

	detail, err := repo.GetDetail(context.Background(), shortID, nil)
	if err != nil {
		t.Fatalf("GetDetail() error = %v, want nil", err)
	}
	if detail.Item.Short.ID != shortID {
		t.Fatalf("GetDetail() short id got %s want %s", detail.Item.Short.ID, shortID)
	}
	if !detail.Viewer.IsFollowingCreator {
		t.Fatal("GetDetail() following = false, want true")
	}
}

func TestOptionalUUIDToPGAndBoolFromAny(t *testing.T) {
	t.Parallel()

	if got := optionalUUIDToPG(nil); got.Valid {
		t.Fatalf("optionalUUIDToPG(nil) got valid=%t want false", got.Valid)
	}

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	if got, err := postgres.UUIDFromPG(optionalUUIDToPG(&viewerID)); err != nil || got != viewerID {
		t.Fatalf("optionalUUIDToPG() got %s err=%v want %s", got, err, viewerID)
	}

	for _, testCase := range []struct {
		name    string
		value   any
		want    bool
		wantErr bool
	}{
		{name: "true", value: true, want: true},
		{name: "false", value: false, want: false},
		{name: "nil", value: nil, want: false},
		{name: "invalid", value: "true", wantErr: true},
	} {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := boolFromAny(testCase.value)
			if testCase.wantErr {
				if err == nil {
					t.Fatal("boolFromAny() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("boolFromAny() error = %v, want nil", err)
			}
			if got != testCase.want {
				t.Fatalf("boolFromAny() got %t want %t", got, testCase.want)
			}
		})
	}
}

func makeMapFeedRow() mapFeedRow {
	return mapFeedRow{
		AvatarUrl:          makeText("https://cdn.example.com/avatar.jpg"),
		Bio:                "night preview specialist",
		CanonicalMainID:    makeUUID("33333333-3333-3333-3333-333333333333"),
		Caption:            makeText("quiet rooftop preview"),
		CreatorUserID:      makeUUID("11111111-1111-1111-1111-111111111111"),
		DisplayName:        makeText("Mina Rei"),
		Handle:             "minarei",
		HasLiked:           true,
		ID:                 makeUUID("22222222-2222-2222-2222-222222222222"),
		IsOwner:            true,
		IsPinned:           true,
		IsUnlocked:         true,
		IsFollowingCreator: true,
		LikeCount:          12,
		MainDurationMs:     makeInt8(480500),
		MainPriceMinor:     int64(1800),
		MediaAssetID:       makeUUID("44444444-4444-4444-4444-444444444444"),
		PublishedAt:        makeTimestamp(time.Unix(1710000000, 0).UTC()),
		ShortDurationMs:    makeInt8(16500),
	}
}

func makeFollowingRow(shortID uuid.UUID, publishedAt time.Time, rankScore int64) sqlc.ListFollowingPublicFeedItemsRow {
	return sqlc.ListFollowingPublicFeedItemsRow{
		ID:                 makeUUID(shortID.String()),
		CreatorUserID:      makeUUID("11111111-1111-1111-1111-111111111111"),
		CanonicalMainID:    makeUUID("33333333-3333-3333-3333-333333333333"),
		MediaAssetID:       makeUUID("44444444-4444-4444-4444-444444444444"),
		Caption:            makeText("quiet rooftop preview"),
		PublishedAt:        makeTimestamp(publishedAt),
		ShortDurationMs:    makeInt8(16500),
		DisplayName:        makeText("Mina Rei"),
		Handle:             "minarei",
		AvatarUrl:          makeText("https://cdn.example.com/avatar.jpg"),
		Bio:                "night preview specialist",
		MainPriceMinor:     1800,
		MainDurationMs:     makeInt8(480500),
		LikeCount:          4,
		HasLiked:           true,
		IsPinned:           true,
		IsUnlocked:         false,
		IsOwner:            false,
		IsFollowingCreator: true,
		RankScore:          rankScore,
	}
}

func makeHydratedFollowingRow(shortID uuid.UUID, publishedAt time.Time) sqlc.ListFeedItemsByShortIDsRow {
	return sqlc.ListFeedItemsByShortIDsRow{
		ID:                 makeUUID(shortID.String()),
		CreatorUserID:      makeUUID("11111111-1111-1111-1111-111111111111"),
		CanonicalMainID:    makeUUID("33333333-3333-3333-3333-333333333333"),
		MediaAssetID:       makeUUID("44444444-4444-4444-4444-444444444444"),
		Caption:            makeText("quiet rooftop preview"),
		PublishedAt:        makeTimestamp(publishedAt),
		ShortDurationMs:    makeInt8(16500),
		DisplayName:        makeText("Mina Rei"),
		Handle:             "minarei",
		AvatarUrl:          makeText("https://cdn.example.com/avatar.jpg"),
		Bio:                "night preview specialist",
		MainPriceMinor:     1800,
		MainDurationMs:     makeInt8(480500),
		LikeCount:          4,
		HasLiked:           true,
		IsPinned:           true,
		IsUnlocked:         false,
		IsOwner:            false,
		IsFollowingCreator: true,
	}
}

func makeLegacyRecommendedRow(shortID uuid.UUID, publishedAt time.Time) sqlc.ListLegacyRecommendedPublicFeedItemsRow {
	return sqlc.ListLegacyRecommendedPublicFeedItemsRow{
		ID:                 makeUUID(shortID.String()),
		CreatorUserID:      makeUUID("11111111-1111-1111-1111-111111111111"),
		CanonicalMainID:    makeUUID("33333333-3333-3333-3333-333333333333"),
		MediaAssetID:       makeUUID("44444444-4444-4444-4444-444444444444"),
		Caption:            makeText("quiet rooftop preview"),
		PublishedAt:        makeTimestamp(publishedAt),
		ShortDurationMs:    makeInt8(16500),
		DisplayName:        makeText("Mina Rei"),
		Handle:             "minarei",
		AvatarUrl:          makeText("https://cdn.example.com/avatar.jpg"),
		Bio:                "night preview specialist",
		MainPriceMinor:     1800,
		MainDurationMs:     makeInt8(480500),
		LikeCount:          4,
		HasLiked:           true,
		IsPinned:           true,
		IsUnlocked:         true,
		IsOwner:            false,
		IsFollowingCreator: true,
	}
}

func makeUUID(value string) pgtype.UUID {
	id := uuid.MustParse(value)

	return pgtype.UUID{
		Bytes: [16]byte(id),
		Valid: true,
	}
}

func makeText(value string) pgtype.Text {
	return pgtype.Text{
		String: value,
		Valid:  true,
	}
}

func makeTimestamp(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  value,
		Valid: true,
	}
}

func makeInt8(value int64) pgtype.Int8 {
	return pgtype.Int8{
		Int64: value,
		Valid: true,
	}
}
