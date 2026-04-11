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
	getDetail       func(context.Context, sqlc.GetPublicShortDetailItemParams) (sqlc.GetPublicShortDetailItemRow, error)
	listFollowing   func(context.Context, sqlc.ListFollowingPublicFeedItemsParams) ([]sqlc.ListFollowingPublicFeedItemsRow, error)
	listRecommended func(context.Context, sqlc.ListRecommendedPublicFeedItemsParams) ([]sqlc.ListRecommendedPublicFeedItemsRow, error)
}

func (s stubQueries) GetPublicShortDetailItem(ctx context.Context, arg sqlc.GetPublicShortDetailItemParams) (sqlc.GetPublicShortDetailItemRow, error) {
	return s.getDetail(ctx, arg)
}

func (s stubQueries) ListFollowingPublicFeedItems(ctx context.Context, arg sqlc.ListFollowingPublicFeedItemsParams) ([]sqlc.ListFollowingPublicFeedItemsRow, error) {
	return s.listFollowing(ctx, arg)
}

func (s stubQueries) ListRecommendedPublicFeedItems(ctx context.Context, arg sqlc.ListRecommendedPublicFeedItemsParams) ([]sqlc.ListRecommendedPublicFeedItemsRow, error) {
	return s.listRecommended(ctx, arg)
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

func TestBuildRecommendedPageParams(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	publishedAt := time.Unix(1710000000, 0).UTC()

	params, limit := buildRecommendedPageParams(&viewerID, &Cursor{
		PublishedAt: publishedAt,
		ShortID:     shortID,
	}, 0)
	if limit != DefaultPageSize {
		t.Fatalf("buildRecommendedPageParams() limit got %d want %d", limit, DefaultPageSize)
	}
	if params.LimitCount != DefaultPageSize+1 {
		t.Fatalf("buildRecommendedPageParams() limit count got %d want %d", params.LimitCount, DefaultPageSize+1)
	}
	if got, err := postgres.UUIDFromPG(params.ViewerUserID); err != nil || got != viewerID {
		t.Fatalf("buildRecommendedPageParams() viewer got %s err=%v want %s", got, err, viewerID)
	}
	if got, err := postgres.RequiredTimeFromPG(params.CursorPublishedAt); err != nil || !got.Equal(publishedAt) {
		t.Fatalf("buildRecommendedPageParams() publishedAt got %s err=%v want %s", got, err, publishedAt)
	}
	if got, err := postgres.UUIDFromPG(params.CursorShortID); err != nil || got != shortID {
		t.Fatalf("buildRecommendedPageParams() cursor short got %s err=%v want %s", got, err, shortID)
	}

	params, limit = buildRecommendedPageParams(nil, nil, 3)
	if limit != 3 {
		t.Fatalf("buildRecommendedPageParams() explicit limit got %d want %d", limit, 3)
	}
	if params.ViewerUserID.Valid {
		t.Fatalf("buildRecommendedPageParams() viewer valid got %t want false", params.ViewerUserID.Valid)
	}
	if params.CursorPublishedAt.Valid {
		t.Fatalf("buildRecommendedPageParams() cursor published valid got %t want false", params.CursorPublishedAt.Valid)
	}
}

func TestBuildFollowingPageParams(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	publishedAt := time.Unix(1710000100, 0).UTC()

	params, limit := buildFollowingPageParams(viewerID, &Cursor{
		PublishedAt: publishedAt,
		ShortID:     shortID,
	}, 1)
	if limit != 1 {
		t.Fatalf("buildFollowingPageParams() limit got %d want %d", limit, 1)
	}
	if params.LimitCount != 2 {
		t.Fatalf("buildFollowingPageParams() limit count got %d want %d", params.LimitCount, 2)
	}
	if got, err := postgres.UUIDFromPG(params.ViewerUserID); err != nil || got != viewerID {
		t.Fatalf("buildFollowingPageParams() viewer got %s err=%v want %s", got, err, viewerID)
	}
	if got, err := postgres.RequiredTimeFromPG(params.CursorPublishedAt); err != nil || !got.Equal(publishedAt) {
		t.Fatalf("buildFollowingPageParams() publishedAt got %s err=%v want %s", got, err, publishedAt)
	}
	if got, err := postgres.UUIDFromPG(params.CursorShortID); err != nil || got != shortID {
		t.Fatalf("buildFollowingPageParams() cursor short got %s err=%v want %s", got, err, shortID)
	}

	params, limit = buildFollowingPageParams(viewerID, nil, 0)
	if limit != DefaultPageSize {
		t.Fatalf("buildFollowingPageParams() default limit got %d want %d", limit, DefaultPageSize)
	}
	if params.LimitCount != DefaultPageSize+1 {
		t.Fatalf("buildFollowingPageParams() limit count got %d want %d", params.LimitCount, DefaultPageSize+1)
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
	if !item.Unlock.IsOwner || !item.Unlock.IsUnlocked || !item.Viewer.IsPinned {
		t.Fatalf("mapFeedItem() booleans got owner=%t unlocked=%t pinned=%t want true/true/true", item.Unlock.IsOwner, item.Unlock.IsUnlocked, item.Viewer.IsPinned)
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

func TestMapRecommendedPage(t *testing.T) {
	t.Parallel()

	firstID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	secondID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	rows := []sqlc.ListRecommendedPublicFeedItemsRow{
		makeRecommendedRow(firstID, time.Unix(1710000200, 0).UTC()),
		makeRecommendedRow(secondID, time.Unix(1710000100, 0).UTC()),
	}

	items, nextCursor, err := mapRecommendedPage(rows, 1, "recommended")
	if err != nil {
		t.Fatalf("mapRecommendedPage() error = %v, want nil", err)
	}
	if len(items) != 1 {
		t.Fatalf("mapRecommendedPage() items len got %d want %d", len(items), 1)
	}
	if nextCursor == nil || nextCursor.ShortID != firstID {
		t.Fatalf("mapRecommendedPage() next cursor got %#v want short %s", nextCursor, firstID)
	}
}

func TestMapFollowingPageAndCursorHelpers(t *testing.T) {
	t.Parallel()

	shortID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	publishedAt := time.Unix(1710000200, 0).UTC()
	rows := []sqlc.ListFollowingPublicFeedItemsRow{
		makeFollowingRow(shortID, publishedAt),
	}

	items, nextCursor, err := mapFollowingPage(rows, 2, "following")
	if err != nil {
		t.Fatalf("mapFollowingPage() error = %v, want nil", err)
	}
	if len(items) != 1 {
		t.Fatalf("mapFollowingPage() items len got %d want %d", len(items), 1)
	}
	if nextCursor != nil {
		t.Fatalf("mapFollowingPage() next cursor got %#v want nil", nextCursor)
	}

	emptyItems, emptyCursor, err := mapFeedPageCursor([]Item{}, []sqlc.ListRecommendedPublicFeedItemsRow{}, 1)
	if err != nil {
		t.Fatalf("mapFeedPageCursor() error = %v, want nil", err)
	}
	if len(emptyItems) != 0 || emptyCursor != nil {
		t.Fatalf("mapFeedPageCursor() got items=%#v cursor=%#v want empty nil", emptyItems, emptyCursor)
	}

	if got := lengthOfRows(rows); got != 1 {
		t.Fatalf("lengthOfRows(following) got %d want %d", got, 1)
	}
	if got := lengthOfRows([]sqlc.ListRecommendedPublicFeedItemsRow{makeRecommendedRow(shortID, publishedAt)}); got != 1 {
		t.Fatalf("lengthOfRows(recommended) got %d want %d", got, 1)
	}
	if got := lengthOfRows(struct{}{}); got != 0 {
		t.Fatalf("lengthOfRows(unknown) got %d want %d", got, 0)
	}
}

func TestMapPageFunctionsWrapMappingErrors(t *testing.T) {
	t.Parallel()

	badRecommended := makeRecommendedRow(
		uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		time.Unix(1710000200, 0).UTC(),
	)
	badRecommended.Handle = " "
	if _, _, err := mapRecommendedPage([]sqlc.ListRecommendedPublicFeedItemsRow{badRecommended}, 1, "recommended"); err == nil {
		t.Fatal("mapRecommendedPage() error = nil, want wrapped mapping error")
	}

	badFollowing := makeFollowingRow(
		uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		time.Unix(1710000200, 0).UTC(),
	)
	badFollowing.Handle = " "
	if _, _, err := mapFollowingPage([]sqlc.ListFollowingPublicFeedItemsRow{badFollowing}, 1, "following"); err == nil {
		t.Fatal("mapFollowingPage() error = nil, want wrapped mapping error")
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
		MainPriceMinor:     makeInt8(1800),
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
			listRecommended: func(_ context.Context, arg sqlc.ListRecommendedPublicFeedItemsParams) ([]sqlc.ListRecommendedPublicFeedItemsRow, error) {
				if got, err := postgres.UUIDFromPG(arg.ViewerUserID); err != nil || got != viewerID {
					t.Fatalf("ListRecommendedPublicFeedItems() viewer got %s err=%v want %s", got, err, viewerID)
				}
				if arg.LimitCount != 2 {
					t.Fatalf("ListRecommendedPublicFeedItems() limit count got %d want %d", arg.LimitCount, 2)
				}

				return []sqlc.ListRecommendedPublicFeedItemsRow{
					makeRecommendedRow(shortID, publishedAt),
				}, nil
			},
			listFollowing: func(_ context.Context, arg sqlc.ListFollowingPublicFeedItemsParams) ([]sqlc.ListFollowingPublicFeedItemsRow, error) {
				if got, err := postgres.UUIDFromPG(arg.ViewerUserID); err != nil || got != viewerID {
					t.Fatalf("ListFollowingPublicFeedItems() viewer got %s err=%v want %s", got, err, viewerID)
				}

				return []sqlc.ListFollowingPublicFeedItemsRow{
					makeFollowingRow(shortID, publishedAt),
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
			listRecommended: func(context.Context, sqlc.ListRecommendedPublicFeedItemsParams) ([]sqlc.ListRecommendedPublicFeedItemsRow, error) {
				return nil, queryErr
			},
			listFollowing: func(context.Context, sqlc.ListFollowingPublicFeedItemsParams) ([]sqlc.ListFollowingPublicFeedItemsRow, error) {
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
				MainPriceMinor:     makeInt8(1800),
				MainDurationMs:     makeInt8(480500),
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
					MainPriceMinor:     makeInt8(1800),
					MainDurationMs:     makeInt8(480500),
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
		AvatarUrl:       makeText("https://cdn.example.com/avatar.jpg"),
		Bio:             "night preview specialist",
		CanonicalMainID: makeUUID("33333333-3333-3333-3333-333333333333"),
		Caption:         makeText("quiet rooftop preview"),
		CreatorUserID:   makeUUID("11111111-1111-1111-1111-111111111111"),
		DisplayName:     makeText("Mina Rei"),
		Handle:          "minarei",
		ID:              makeUUID("22222222-2222-2222-2222-222222222222"),
		IsOwner:         true,
		IsPinned:        true,
		IsUnlocked:      true,
		MainDurationMs:  makeInt8(480500),
		MainPriceMinor:  makeInt8(1800),
		MediaAssetID:    makeUUID("44444444-4444-4444-4444-444444444444"),
		PublishedAt:     makeTimestamp(time.Unix(1710000000, 0).UTC()),
		ShortDurationMs: makeInt8(16500),
	}
}

func makeFollowingRow(shortID uuid.UUID, publishedAt time.Time) sqlc.ListFollowingPublicFeedItemsRow {
	return sqlc.ListFollowingPublicFeedItemsRow{
		ID:              makeUUID(shortID.String()),
		CreatorUserID:   makeUUID("11111111-1111-1111-1111-111111111111"),
		CanonicalMainID: makeUUID("33333333-3333-3333-3333-333333333333"),
		MediaAssetID:    makeUUID("44444444-4444-4444-4444-444444444444"),
		Caption:         makeText("quiet rooftop preview"),
		PublishedAt:     makeTimestamp(publishedAt),
		ShortDurationMs: makeInt8(16500),
		DisplayName:     makeText("Mina Rei"),
		Handle:          "minarei",
		AvatarUrl:       makeText("https://cdn.example.com/avatar.jpg"),
		Bio:             "night preview specialist",
		MainPriceMinor:  makeInt8(1800),
		MainDurationMs:  makeInt8(480500),
		IsPinned:        true,
		IsUnlocked:      false,
		IsOwner:         false,
	}
}

func makeRecommendedRow(shortID uuid.UUID, publishedAt time.Time) sqlc.ListRecommendedPublicFeedItemsRow {
	return sqlc.ListRecommendedPublicFeedItemsRow{
		ID:              makeUUID(shortID.String()),
		CreatorUserID:   makeUUID("11111111-1111-1111-1111-111111111111"),
		CanonicalMainID: makeUUID("33333333-3333-3333-3333-333333333333"),
		MediaAssetID:    makeUUID("44444444-4444-4444-4444-444444444444"),
		Caption:         makeText("quiet rooftop preview"),
		PublishedAt:     makeTimestamp(publishedAt),
		ShortDurationMs: makeInt8(16500),
		DisplayName:     makeText("Mina Rei"),
		Handle:          "minarei",
		AvatarUrl:       makeText("https://cdn.example.com/avatar.jpg"),
		Bio:             "night preview specialist",
		MainPriceMinor:  makeInt8(1800),
		MainDurationMs:  makeInt8(480500),
		IsPinned:        true,
		IsUnlocked:      true,
		IsOwner:         false,
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
