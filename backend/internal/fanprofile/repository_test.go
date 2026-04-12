package fanprofile

import (
	"context"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type repositoryStubQueries struct {
	listFollowing    func(context.Context, sqlc.ListFanProfileFollowingItemsParams) ([]sqlc.ListFanProfileFollowingItemsRow, error)
	listLibrary      func(context.Context, sqlc.ListFanProfileLibraryItemsParams) ([]sqlc.ListFanProfileLibraryItemsRow, error)
	listPinnedShorts func(context.Context, sqlc.ListFanProfilePinnedShortItemsParams) ([]sqlc.ListFanProfilePinnedShortItemsRow, error)
}

func (s repositoryStubQueries) GetUserByID(context.Context, pgtype.UUID) (sqlc.AppUser, error) {
	return sqlc.AppUser{}, nil
}

func (s repositoryStubQueries) CountCreatorFollowsByUserID(context.Context, pgtype.UUID) (int64, error) {
	return 0, nil
}

func (s repositoryStubQueries) CountPinnedShortsByUserID(context.Context, pgtype.UUID) (int64, error) {
	return 0, nil
}

func (s repositoryStubQueries) CountUnlockedMainsByUserID(context.Context, pgtype.UUID) (int64, error) {
	return 0, nil
}

func (s repositoryStubQueries) ListFanProfileFollowingItems(ctx context.Context, arg sqlc.ListFanProfileFollowingItemsParams) ([]sqlc.ListFanProfileFollowingItemsRow, error) {
	if s.listFollowing == nil {
		return nil, nil
	}

	return s.listFollowing(ctx, arg)
}

func (s repositoryStubQueries) ListFanProfileLibraryItems(ctx context.Context, arg sqlc.ListFanProfileLibraryItemsParams) ([]sqlc.ListFanProfileLibraryItemsRow, error) {
	if s.listLibrary == nil {
		return nil, nil
	}

	return s.listLibrary(ctx, arg)
}

func (s repositoryStubQueries) ListFanProfilePinnedShortItems(ctx context.Context, arg sqlc.ListFanProfilePinnedShortItemsParams) ([]sqlc.ListFanProfilePinnedShortItemsRow, error) {
	if s.listPinnedShorts == nil {
		return nil, nil
	}

	return s.listPinnedShorts(ctx, arg)
}

func TestListFollowing(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	creatorAID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	creatorBID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	creatorCID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	var gotParams sqlc.ListFanProfileFollowingItemsParams
	repo := newRepository(repositoryStubQueries{
		listFollowing: func(_ context.Context, arg sqlc.ListFanProfileFollowingItemsParams) ([]sqlc.ListFanProfileFollowingItemsRow, error) {
			gotParams = arg

			return []sqlc.ListFanProfileFollowingItemsRow{
				testFollowingRow(creatorAID, "Aoi N", "aoina", stringPtr("https://cdn.example.com/creator/aoi/avatar.jpg"), "soft light", now),
				testFollowingRow(creatorBID, "Mina Rei", "minarei", nil, "quiet rooftop", now.Add(-time.Hour)),
				testFollowingRow(creatorCID, "Sora Vale", "soravale", nil, "after rain", now.Add(-2*time.Hour)),
			}, nil
		},
	})

	items, nextCursor, err := repo.ListFollowing(context.Background(), viewerID, nil, 2)
	if err != nil {
		t.Fatalf("ListFollowing() error = %v, want nil", err)
	}
	if gotParams.UserID != uuidToPG(viewerID) {
		t.Fatalf("ListFollowing() viewer arg got %v want %v", gotParams.UserID, uuidToPG(viewerID))
	}
	if gotParams.LimitCount != 3 {
		t.Fatalf("ListFollowing() limit got %d want %d", gotParams.LimitCount, 3)
	}
	if len(items) != 2 {
		t.Fatalf("ListFollowing() len got %d want %d", len(items), 2)
	}
	if items[0].CreatorUserID != creatorAID {
		t.Fatalf("ListFollowing() first creator id got %s want %s", items[0].CreatorUserID, creatorAID)
	}
	if items[0].AvatarURL == nil || *items[0].AvatarURL != "https://cdn.example.com/creator/aoi/avatar.jpg" {
		t.Fatalf("ListFollowing() first avatar got %#v want avatar url", items[0].AvatarURL)
	}
	if nextCursor == nil {
		t.Fatal("ListFollowing() nextCursor = nil, want non-nil")
	}
	if nextCursor.CreatorUserID != creatorBID {
		t.Fatalf("ListFollowing() next cursor creator id got %s want %s", nextCursor.CreatorUserID, creatorBID)
	}
	if !nextCursor.FollowedAt.Equal(now.Add(-time.Hour)) {
		t.Fatalf("ListFollowing() next cursor followedAt got %s want %s", nextCursor.FollowedAt, now.Add(-time.Hour))
	}
}

func TestListFollowingPassesCursor(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	cursor := &FollowingCursor{
		CreatorUserID: uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		FollowedAt:    now,
	}

	var gotParams sqlc.ListFanProfileFollowingItemsParams
	repo := newRepository(repositoryStubQueries{
		listFollowing: func(_ context.Context, arg sqlc.ListFanProfileFollowingItemsParams) ([]sqlc.ListFanProfileFollowingItemsRow, error) {
			gotParams = arg
			return []sqlc.ListFanProfileFollowingItemsRow{}, nil
		},
	})

	items, nextCursor, err := repo.ListFollowing(context.Background(), viewerID, cursor, 2)
	if err != nil {
		t.Fatalf("ListFollowing() error = %v, want nil", err)
	}
	if len(items) != 0 {
		t.Fatalf("ListFollowing() len got %d want %d", len(items), 0)
	}
	if nextCursor != nil {
		t.Fatalf("ListFollowing() nextCursor got %#v want nil", nextCursor)
	}
	if gotParams.CursorFollowedAt != timeToPG(&now) {
		t.Fatalf("ListFollowing() cursor followed_at got %v want %v", gotParams.CursorFollowedAt, timeToPG(&now))
	}
	if gotParams.CursorCreatorUserID != uuidToPG(cursor.CreatorUserID) {
		t.Fatalf("ListFollowing() cursor creator user id got %v want %v", gotParams.CursorCreatorUserID, uuidToPG(cursor.CreatorUserID))
	}
}

func TestListFollowingEmpty(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	repo := newRepository(repositoryStubQueries{
		listFollowing: func(_ context.Context, arg sqlc.ListFanProfileFollowingItemsParams) ([]sqlc.ListFanProfileFollowingItemsRow, error) {
			return []sqlc.ListFanProfileFollowingItemsRow{}, nil
		},
	})

	items, nextCursor, err := repo.ListFollowing(context.Background(), viewerID, nil, 2)
	if err != nil {
		t.Fatalf("ListFollowing() error = %v, want nil", err)
	}
	if len(items) != 0 {
		t.Fatalf("ListFollowing() len got %d want %d", len(items), 0)
	}
	if nextCursor != nil {
		t.Fatalf("ListFollowing() nextCursor got %#v want nil", nextCursor)
	}
}

func testFollowingRow(
	creatorUserID uuid.UUID,
	displayName string,
	handle string,
	avatarURL *string,
	bio string,
	followedAt time.Time,
) sqlc.ListFanProfileFollowingItemsRow {
	return sqlc.ListFanProfileFollowingItemsRow{
		CreatorUserID: uuidToPG(creatorUserID),
		DisplayName:   textToPG(displayName),
		Handle:        handle,
		AvatarUrl:     postgres.TextToPG(avatarURL),
		Bio:           bio,
		FollowedAt:    timeToPG(&followedAt),
	}
}

func stringPtr(value string) *string {
	return &value
}

func textToPG(value string) pgtype.Text {
	return pgtype.Text{
		String: value,
		Valid:  true,
	}
}

func timeToPG(value *time.Time) pgtype.Timestamptz {
	return postgres.TimeToPG(value)
}

func uuidToPG(value uuid.UUID) pgtype.UUID {
	return postgres.UUIDToPG(value)
}
