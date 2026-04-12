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

func TestListLibrary(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainAID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainBID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	mainCID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	creatorID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	shortID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	mediaID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")

	var gotParams sqlc.ListFanProfileLibraryItemsParams
	repo := newRepository(repositoryStubQueries{
		listLibrary: func(_ context.Context, arg sqlc.ListFanProfileLibraryItemsParams) ([]sqlc.ListFanProfileLibraryItemsRow, error) {
			gotParams = arg

			return []sqlc.ListFanProfileLibraryItemsRow{
				testLibraryRow(mainAID, creatorID, shortID, mediaID, now, now.Add(-time.Minute), " quiet rooftop preview ", "Mina Rei", "minarei", stringPtr("https://cdn.example.com/creator/mina/avatar.jpg"), "quiet rooftop"),
				testLibraryRow(mainBID, creatorID, shortID, mediaID, now.Add(-time.Hour), now.Add(-time.Hour), "", "Mina Rei", "minarei", nil, "quiet rooftop"),
				testLibraryRow(mainCID, creatorID, shortID, mediaID, now.Add(-2*time.Hour), now.Add(-2*time.Hour), "hotel light preview", "Mina Rei", "minarei", nil, "quiet rooftop"),
			}, nil
		},
	})

	items, nextCursor, err := repo.ListLibrary(context.Background(), viewerID, nil, 2)
	if err != nil {
		t.Fatalf("ListLibrary() error = %v, want nil", err)
	}
	if gotParams.UserID != uuidToPG(viewerID) {
		t.Fatalf("ListLibrary() viewer arg got %v want %v", gotParams.UserID, uuidToPG(viewerID))
	}
	if gotParams.LimitCount != 3 {
		t.Fatalf("ListLibrary() limit got %d want %d", gotParams.LimitCount, 3)
	}
	if len(items) != 2 {
		t.Fatalf("ListLibrary() len got %d want %d", len(items), 2)
	}
	if items[0].MainID != mainAID {
		t.Fatalf("ListLibrary() first main id got %s want %s", items[0].MainID, mainAID)
	}
	if items[0].EntryShortCaption != "quiet rooftop preview" {
		t.Fatalf("ListLibrary() first entry short caption got %q want %q", items[0].EntryShortCaption, "quiet rooftop preview")
	}
	if items[1].EntryShortCaption != "" {
		t.Fatalf("ListLibrary() blank caption got %q want empty string", items[1].EntryShortCaption)
	}
	if items[0].MainDurationSeconds != 480 {
		t.Fatalf("ListLibrary() main duration got %d want %d", items[0].MainDurationSeconds, 480)
	}
	if items[0].EntryShortPreviewDurationSeconds != 16 {
		t.Fatalf("ListLibrary() entry short duration got %d want %d", items[0].EntryShortPreviewDurationSeconds, 16)
	}
	if nextCursor == nil {
		t.Fatal("ListLibrary() nextCursor = nil, want non-nil")
	}
	if nextCursor.MainID != mainBID {
		t.Fatalf("ListLibrary() next cursor main id got %s want %s", nextCursor.MainID, mainBID)
	}
	if !nextCursor.PurchasedAt.Equal(now.Add(-time.Hour)) {
		t.Fatalf("ListLibrary() next cursor purchasedAt got %s want %s", nextCursor.PurchasedAt, now.Add(-time.Hour))
	}
	if !nextCursor.UnlockCreatedAt.Equal(now.Add(-time.Hour)) {
		t.Fatalf("ListLibrary() next cursor unlockCreatedAt got %s want %s", nextCursor.UnlockCreatedAt, now.Add(-time.Hour))
	}
}

func TestListLibraryPassesCursor(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	cursor := &LibraryCursor{
		MainID:          uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		PurchasedAt:     now,
		UnlockCreatedAt: now.Add(-time.Minute),
	}

	var gotParams sqlc.ListFanProfileLibraryItemsParams
	repo := newRepository(repositoryStubQueries{
		listLibrary: func(_ context.Context, arg sqlc.ListFanProfileLibraryItemsParams) ([]sqlc.ListFanProfileLibraryItemsRow, error) {
			gotParams = arg
			return []sqlc.ListFanProfileLibraryItemsRow{}, nil
		},
	})

	items, nextCursor, err := repo.ListLibrary(context.Background(), viewerID, cursor, 2)
	if err != nil {
		t.Fatalf("ListLibrary() error = %v, want nil", err)
	}
	if len(items) != 0 {
		t.Fatalf("ListLibrary() len got %d want %d", len(items), 0)
	}
	if nextCursor != nil {
		t.Fatalf("ListLibrary() nextCursor got %#v want nil", nextCursor)
	}
	if gotParams.CursorMainID != uuidToPG(cursor.MainID) {
		t.Fatalf("ListLibrary() cursor main id got %v want %v", gotParams.CursorMainID, uuidToPG(cursor.MainID))
	}
	if gotParams.CursorPurchasedAt != timeToPG(&now) {
		t.Fatalf("ListLibrary() cursor purchased_at got %v want %v", gotParams.CursorPurchasedAt, timeToPG(&now))
	}
	if gotParams.CursorUnlockCreatedAt != timeToPG(&cursor.UnlockCreatedAt) {
		t.Fatalf("ListLibrary() cursor created_at got %v want %v", gotParams.CursorUnlockCreatedAt, timeToPG(&cursor.UnlockCreatedAt))
	}
}

func TestListLibraryEmpty(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repo := newRepository(repositoryStubQueries{
		listLibrary: func(_ context.Context, arg sqlc.ListFanProfileLibraryItemsParams) ([]sqlc.ListFanProfileLibraryItemsRow, error) {
			return []sqlc.ListFanProfileLibraryItemsRow{}, nil
		},
	})

	items, nextCursor, err := repo.ListLibrary(context.Background(), viewerID, nil, 2)
	if err != nil {
		t.Fatalf("ListLibrary() error = %v, want nil", err)
	}
	if len(items) != 0 {
		t.Fatalf("ListLibrary() len got %d want %d", len(items), 0)
	}
	if nextCursor != nil {
		t.Fatalf("ListLibrary() nextCursor got %#v want nil", nextCursor)
	}
}

func testLibraryRow(
	mainID uuid.UUID,
	creatorUserID uuid.UUID,
	entryShortID uuid.UUID,
	entryShortMediaAssetID uuid.UUID,
	purchasedAt time.Time,
	unlockCreatedAt time.Time,
	entryShortCaption string,
	displayName string,
	handle string,
	avatarURL *string,
	bio string,
) sqlc.ListFanProfileLibraryItemsRow {
	return sqlc.ListFanProfileLibraryItemsRow{
		MainID:                    uuidToPG(mainID),
		PurchasedAt:               timeToPG(&purchasedAt),
		UnlockCreatedAt:           timeToPG(&unlockCreatedAt),
		CreatorUserID:             uuidToPG(creatorUserID),
		MainDurationMs:            requiredInt64ToPG(480000),
		DisplayName:               textToPG(displayName),
		Handle:                    handle,
		AvatarUrl:                 postgres.TextToPG(avatarURL),
		Bio:                       bio,
		EntryShortID:              uuidToPG(entryShortID),
		EntryShortCanonicalMainID: uuidToPG(mainID),
		EntryShortCaption:         postgres.TextToPG(optionalString(entryShortCaption)),
		EntryShortMediaAssetID:    uuidToPG(entryShortMediaAssetID),
		EntryShortDurationMs:      requiredInt64ToPG(16000),
	}
}

func requiredInt64ToPG(value int64) pgtype.Int8 {
	return pgtype.Int8{
		Int64: value,
		Valid: true,
	}
}
