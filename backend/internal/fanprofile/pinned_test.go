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

func TestListPinnedShorts(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortAID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	shortBID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortCID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	creatorID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	mainID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	mediaID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")

	var gotParams sqlc.ListFanProfilePinnedShortItemsParams
	repo := newRepository(repositoryStubQueries{
		listPinnedShorts: func(_ context.Context, arg sqlc.ListFanProfilePinnedShortItemsParams) ([]sqlc.ListFanProfilePinnedShortItemsRow, error) {
			gotParams = arg

			return []sqlc.ListFanProfilePinnedShortItemsRow{
				testPinnedShortRow(shortAID, creatorID, mainID, mediaID, " after rain preview ", now, "Sora Vale", "soravale", stringPtr("https://cdn.example.com/creator/sora/avatar.jpg"), "after rain"),
				testPinnedShortRow(shortBID, creatorID, mainID, mediaID, "", now.Add(-time.Hour), "Sora Vale", "soravale", nil, "after rain"),
				testPinnedShortRow(shortCID, creatorID, mainID, mediaID, "poolside preview", now.Add(-2*time.Hour), "Sora Vale", "soravale", nil, "after rain"),
			}, nil
		},
	})

	items, nextCursor, err := repo.ListPinnedShorts(context.Background(), viewerID, nil, 2)
	if err != nil {
		t.Fatalf("ListPinnedShorts() error = %v, want nil", err)
	}
	if gotParams.UserID != uuidToPG(viewerID) {
		t.Fatalf("ListPinnedShorts() viewer arg got %v want %v", gotParams.UserID, uuidToPG(viewerID))
	}
	if gotParams.LimitCount != 3 {
		t.Fatalf("ListPinnedShorts() limit got %d want %d", gotParams.LimitCount, 3)
	}
	if len(items) != 2 {
		t.Fatalf("ListPinnedShorts() len got %d want %d", len(items), 2)
	}
	if items[0].ShortID != shortAID {
		t.Fatalf("ListPinnedShorts() first short id got %s want %s", items[0].ShortID, shortAID)
	}
	if items[0].ShortCaption != "after rain preview" {
		t.Fatalf("ListPinnedShorts() first caption got %q want %q", items[0].ShortCaption, "after rain preview")
	}
	if items[1].ShortCaption != "" {
		t.Fatalf("ListPinnedShorts() blank caption got %q want empty string", items[1].ShortCaption)
	}
	if nextCursor == nil {
		t.Fatal("ListPinnedShorts() nextCursor = nil, want non-nil")
	}
	if nextCursor.ShortID != shortBID {
		t.Fatalf("ListPinnedShorts() next cursor short id got %s want %s", nextCursor.ShortID, shortBID)
	}
	if !nextCursor.PinnedAt.Equal(now.Add(-time.Hour)) {
		t.Fatalf("ListPinnedShorts() next cursor pinnedAt got %s want %s", nextCursor.PinnedAt, now.Add(-time.Hour))
	}
}

func TestListPinnedShortsPassesCursor(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	cursor := &PinnedShortCursor{
		PinnedAt: now,
		ShortID:  uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
	}

	var gotParams sqlc.ListFanProfilePinnedShortItemsParams
	repo := newRepository(repositoryStubQueries{
		listPinnedShorts: func(_ context.Context, arg sqlc.ListFanProfilePinnedShortItemsParams) ([]sqlc.ListFanProfilePinnedShortItemsRow, error) {
			gotParams = arg
			return []sqlc.ListFanProfilePinnedShortItemsRow{}, nil
		},
	})

	items, nextCursor, err := repo.ListPinnedShorts(context.Background(), viewerID, cursor, 2)
	if err != nil {
		t.Fatalf("ListPinnedShorts() error = %v, want nil", err)
	}
	if len(items) != 0 {
		t.Fatalf("ListPinnedShorts() len got %d want %d", len(items), 0)
	}
	if nextCursor != nil {
		t.Fatalf("ListPinnedShorts() nextCursor got %#v want nil", nextCursor)
	}
	if gotParams.CursorPinnedAt != timeToPG(&now) {
		t.Fatalf("ListPinnedShorts() cursor pinned_at got %v want %v", gotParams.CursorPinnedAt, timeToPG(&now))
	}
	if gotParams.CursorShortID != uuidToPG(cursor.ShortID) {
		t.Fatalf("ListPinnedShorts() cursor short id got %v want %v", gotParams.CursorShortID, uuidToPG(cursor.ShortID))
	}
}

func TestListPinnedShortsEmpty(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repo := newRepository(repositoryStubQueries{
		listPinnedShorts: func(_ context.Context, arg sqlc.ListFanProfilePinnedShortItemsParams) ([]sqlc.ListFanProfilePinnedShortItemsRow, error) {
			return []sqlc.ListFanProfilePinnedShortItemsRow{}, nil
		},
	})

	items, nextCursor, err := repo.ListPinnedShorts(context.Background(), viewerID, nil, 2)
	if err != nil {
		t.Fatalf("ListPinnedShorts() error = %v, want nil", err)
	}
	if len(items) != 0 {
		t.Fatalf("ListPinnedShorts() len got %d want %d", len(items), 0)
	}
	if nextCursor != nil {
		t.Fatalf("ListPinnedShorts() nextCursor got %#v want nil", nextCursor)
	}
}

func testPinnedShortRow(
	shortID uuid.UUID,
	creatorUserID uuid.UUID,
	canonicalMainID uuid.UUID,
	mediaAssetID uuid.UUID,
	caption string,
	pinnedAt time.Time,
	displayName string,
	handle string,
	avatarURL *string,
	bio string,
) sqlc.ListFanProfilePinnedShortItemsRow {
	return sqlc.ListFanProfilePinnedShortItemsRow{
		ShortID:         uuidToPG(shortID),
		CreatorUserID:   uuidToPG(creatorUserID),
		CanonicalMainID: uuidToPG(canonicalMainID),
		MediaAssetID:    uuidToPG(mediaAssetID),
		Caption:         postgres.TextToPG(optionalString(caption)),
		PinnedAt:        timeToPG(&pinnedAt),
		DurationMs: pgtype.Int8{
			Int64: 17000,
			Valid: true,
		},
		DisplayName: textToPG(displayName),
		Handle:      handle,
		AvatarUrl:   postgres.TextToPG(avatarURL),
		Bio:         bio,
	}
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}
