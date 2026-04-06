package creator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestParsePublicID(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	parsed, err := ParsePublicID(FormatPublicID(userID))
	if err != nil {
		t.Fatalf("ParsePublicID() error = %v, want nil", err)
	}
	if parsed != userID {
		t.Fatalf("ParsePublicID() got %s want %s", parsed, userID)
	}
}

func TestParsePublicIDRejectsInvalidValue(t *testing.T) {
	t.Parallel()

	if _, err := ParsePublicID("creator_invalid"); !errors.Is(err, ErrInvalidPublicID) {
		t.Fatalf("ParsePublicID() error got %v want %v", err, ErrInvalidPublicID)
	}
}

func TestGetPublicProfileHeader(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	var gotFollowerCountUserID pgtype.UUID
	var gotShortCountUserID pgtype.UUID

	repo := newRepository(repositoryStubQueries{
		getPublicProfile: func(_ context.Context, userIDArg pgtype.UUID) (sqlc.AppPublicCreatorProfile, error) {
			return testPublicProfileRow(userID, now, stringPtr("Mina Rei"), stringPtr("minarei"), stringPtr("https://cdn.example.com/mina.jpg"), timePtr(now)), nil
		},
		countFollowers: func(_ context.Context, creatorUserID pgtype.UUID) (int64, error) {
			gotFollowerCountUserID = creatorUserID
			return 24, nil
		},
		countPublicShorts: func(_ context.Context, creatorUserID pgtype.UUID) (int64, error) {
			gotShortCountUserID = creatorUserID
			return 2, nil
		},
	})

	header, err := repo.GetPublicProfileHeader(context.Background(), FormatPublicID(userID))
	if err != nil {
		t.Fatalf("GetPublicProfileHeader() error = %v, want nil", err)
	}
	if header.ShortCount != 2 {
		t.Fatalf("GetPublicProfileHeader() short count got %d want %d", header.ShortCount, 2)
	}
	if header.FanCount != 24 {
		t.Fatalf("GetPublicProfileHeader() fan count got %d want %d", header.FanCount, 24)
	}
	if header.IsFollowing {
		t.Fatal("GetPublicProfileHeader() isFollowing = true, want false")
	}
	if gotFollowerCountUserID != pgUUID(userID) {
		t.Fatalf("GetPublicProfileHeader() follower count arg got %v want %v", gotFollowerCountUserID, pgUUID(userID))
	}
	if gotShortCountUserID != pgUUID(userID) {
		t.Fatalf("GetPublicProfileHeader() short count arg got %v want %v", gotShortCountUserID, pgUUID(userID))
	}
}

func TestGetPublicProfileHeaderRejectsInvalidCreatorID(t *testing.T) {
	t.Parallel()

	repo := newRepository(repositoryStubQueries{})

	if _, err := repo.GetPublicProfileHeader(context.Background(), "bad"); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("GetPublicProfileHeader() error got %v want %v", err, ErrProfileNotFound)
	}
}

func TestListPublicProfileShorts(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	shortAID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	shortBID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortCID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	mainID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	mediaAID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	mediaBID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")
	mediaCID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")

	var gotParams sqlc.ListCreatorProfileShortGridItemsParams
	getPublicProfileCalls := 0
	repo := newRepository(repositoryStubQueries{
		getPublicProfile: func(_ context.Context, userIDArg pgtype.UUID) (sqlc.AppPublicCreatorProfile, error) {
			getPublicProfileCalls++
			return testPublicProfileRow(userID, now, stringPtr("Mina Rei"), stringPtr("minarei"), stringPtr("https://cdn.example.com/mina.jpg"), timePtr(now)), nil
		},
		listProfileShortGrid: func(_ context.Context, arg sqlc.ListCreatorProfileShortGridItemsParams) ([]sqlc.ListCreatorProfileShortGridItemsRow, error) {
			gotParams = arg
			return []sqlc.ListCreatorProfileShortGridItemsRow{
				testPublicProfileShortGridRow(shortAID, userID, mainID, mediaAID, now, "https://cdn.example.com/short-a.mp4", 16000),
				testPublicProfileShortGridRow(shortBID, userID, mainID, mediaBID, now.Add(-time.Hour), "https://cdn.example.com/short-b.mp4", 14000),
				testPublicProfileShortGridRow(shortCID, userID, mainID, mediaCID, now.Add(-2*time.Hour), "https://cdn.example.com/short-c.mp4", 19000),
			}, nil
		},
	})

	shorts, nextCursor, err := repo.ListPublicProfileShorts(context.Background(), FormatPublicID(userID), nil, 2)
	if err != nil {
		t.Fatalf("ListPublicProfileShorts() error = %v, want nil", err)
	}
	if gotParams.CreatorUserID != pgUUID(userID) {
		t.Fatalf("ListPublicProfileShorts() creator arg got %v want %v", gotParams.CreatorUserID, pgUUID(userID))
	}
	if gotParams.LimitCount != 3 {
		t.Fatalf("ListPublicProfileShorts() limit got %d want %d", gotParams.LimitCount, 3)
	}
	if len(shorts) != 2 {
		t.Fatalf("ListPublicProfileShorts() len got %d want %d", len(shorts), 2)
	}
	if shorts[0].PreviewDurationSeconds != 16 {
		t.Fatalf("ListPublicProfileShorts() duration got %d want %d", shorts[0].PreviewDurationSeconds, 16)
	}
	if nextCursor == nil {
		t.Fatal("ListPublicProfileShorts() nextCursor = nil, want non-nil")
	}
	if nextCursor.ShortID != shortBID {
		t.Fatalf("ListPublicProfileShorts() next cursor short id got %s want %s", nextCursor.ShortID, shortBID)
	}
	if getPublicProfileCalls != 0 {
		t.Fatalf("ListPublicProfileShorts() profile lookup calls got %d want %d", getPublicProfileCalls, 0)
	}
}

func TestListPublicProfileShortsEmptyUsesProfileLookup(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	getPublicProfileCalls := 0
	repo := newRepository(repositoryStubQueries{
		getPublicProfile: func(_ context.Context, userIDArg pgtype.UUID) (sqlc.AppPublicCreatorProfile, error) {
			getPublicProfileCalls++
			return testPublicProfileRow(userID, now, stringPtr("Mina Rei"), stringPtr("minarei"), stringPtr("https://cdn.example.com/mina.jpg"), timePtr(now)), nil
		},
		listProfileShortGrid: func(_ context.Context, arg sqlc.ListCreatorProfileShortGridItemsParams) ([]sqlc.ListCreatorProfileShortGridItemsRow, error) {
			return []sqlc.ListCreatorProfileShortGridItemsRow{}, nil
		},
	})

	shorts, nextCursor, err := repo.ListPublicProfileShorts(context.Background(), FormatPublicID(userID), nil, 2)
	if err != nil {
		t.Fatalf("ListPublicProfileShorts() error = %v, want nil", err)
	}
	if len(shorts) != 0 {
		t.Fatalf("ListPublicProfileShorts() len got %d want %d", len(shorts), 0)
	}
	if nextCursor != nil {
		t.Fatalf("ListPublicProfileShorts() nextCursor got %#v want nil", nextCursor)
	}
	if getPublicProfileCalls != 1 {
		t.Fatalf("ListPublicProfileShorts() profile lookup calls got %d want %d", getPublicProfileCalls, 1)
	}
}

func TestListPublicProfileShortsRejectsInvalidCreatorID(t *testing.T) {
	t.Parallel()

	repo := newRepository(repositoryStubQueries{})

	if _, _, err := repo.ListPublicProfileShorts(context.Background(), "bad", nil, 2); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("ListPublicProfileShorts() error got %v want %v", err, ErrProfileNotFound)
	}
}

func testPublicProfileShortGridRow(shortID uuid.UUID, creatorUserID uuid.UUID, mainID uuid.UUID, mediaAssetID uuid.UUID, publishedAt time.Time, playbackURL string, durationMs int64) sqlc.ListCreatorProfileShortGridItemsRow {
	return sqlc.ListCreatorProfileShortGridItemsRow{
		ID:              pgUUID(shortID),
		CreatorUserID:   pgUUID(creatorUserID),
		CanonicalMainID: pgUUID(mainID),
		MediaAssetID:    pgUUID(mediaAssetID),
		PublishedAt:     pgTime(timePtr(publishedAt)),
		PlaybackUrl:     pgText(stringPtr(playbackURL)),
		DurationMs: pgtype.Int8{
			Int64: durationMs,
			Valid: true,
		},
	}
}
