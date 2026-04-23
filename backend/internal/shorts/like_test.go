package shorts

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestLikePublicShort(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tx := &stubTx{}

	var gotPublicShortID pgtype.UUID
	var gotPutArg sqlc.PutShortLikeParams
	var gotCountShortID pgtype.UUID

	repo := newRepository(
		stubBeginner{tx: tx},
		repositoryStubQueries{},
		func(sqlc.DBTX) queries {
			return repositoryStubQueries{
				getPublicShort: func(_ context.Context, id pgtype.UUID) (sqlc.AppPublicShort, error) {
					gotPublicShortID = id
					return testPublicShortRow(shortID, uuid.New(), uuid.New(), uuid.New(), now, nil, nil, timePtr(now), timePtr(now)), nil
				},
				putShortLike: func(_ context.Context, arg sqlc.PutShortLikeParams) error {
					gotPutArg = arg
					return nil
				},
				deleteShortLike: func(context.Context, sqlc.DeleteShortLikeParams) error {
					t.Fatal("DeleteShortLike() was called during like")
					return nil
				},
				countShortLikes: func(_ context.Context, shortID pgtype.UUID) (int64, error) {
					gotCountShortID = shortID
					return 7, nil
				},
			}
		},
	)

	got, err := repo.LikePublicShort(context.Background(), viewerID, shortID)
	if err != nil {
		t.Fatalf("LikePublicShort() error = %v, want nil", err)
	}
	if !got.HasLiked {
		t.Fatal("LikePublicShort() hasLiked = false, want true")
	}
	if got.LikeCount != 7 {
		t.Fatalf("LikePublicShort() likeCount got %d want %d", got.LikeCount, 7)
	}
	if gotPublicShortID != postgresUUID(shortID) {
		t.Fatalf("LikePublicShort() public short id arg got %v want %v", gotPublicShortID, postgresUUID(shortID))
	}
	if gotPutArg.UserID != postgresUUID(viewerID) {
		t.Fatalf("LikePublicShort() user arg got %v want %v", gotPutArg.UserID, postgresUUID(viewerID))
	}
	if gotPutArg.ShortID != postgresUUID(shortID) {
		t.Fatalf("LikePublicShort() short arg got %v want %v", gotPutArg.ShortID, postgresUUID(shortID))
	}
	if gotCountShortID != postgresUUID(shortID) {
		t.Fatalf("LikePublicShort() count short arg got %v want %v", gotCountShortID, postgresUUID(shortID))
	}
	if !tx.committed {
		t.Fatal("LikePublicShort() committed = false, want true")
	}
	if tx.rolledBack {
		t.Fatal("LikePublicShort() rolledBack = true, want false")
	}
}

func TestUnlikePublicShort(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tx := &stubTx{}

	var gotDeleteArg sqlc.DeleteShortLikeParams

	repo := newRepository(
		stubBeginner{tx: tx},
		repositoryStubQueries{},
		func(sqlc.DBTX) queries {
			return repositoryStubQueries{
				getPublicShort: func(_ context.Context, id pgtype.UUID) (sqlc.AppPublicShort, error) {
					return testPublicShortRow(shortID, uuid.New(), uuid.New(), uuid.New(), now, nil, nil, timePtr(now), timePtr(now)), nil
				},
				putShortLike: func(context.Context, sqlc.PutShortLikeParams) error {
					t.Fatal("PutShortLike() was called during unlike")
					return nil
				},
				deleteShortLike: func(_ context.Context, arg sqlc.DeleteShortLikeParams) error {
					gotDeleteArg = arg
					return nil
				},
				countShortLikes: func(context.Context, pgtype.UUID) (int64, error) {
					return 6, nil
				},
			}
		},
	)

	got, err := repo.UnlikePublicShort(context.Background(), viewerID, shortID)
	if err != nil {
		t.Fatalf("UnlikePublicShort() error = %v, want nil", err)
	}
	if got.HasLiked {
		t.Fatal("UnlikePublicShort() hasLiked = true, want false")
	}
	if got.LikeCount != 6 {
		t.Fatalf("UnlikePublicShort() likeCount got %d want %d", got.LikeCount, 6)
	}
	if gotDeleteArg.UserID != postgresUUID(viewerID) {
		t.Fatalf("UnlikePublicShort() user arg got %v want %v", gotDeleteArg.UserID, postgresUUID(viewerID))
	}
	if gotDeleteArg.ShortID != postgresUUID(shortID) {
		t.Fatalf("UnlikePublicShort() short arg got %v want %v", gotDeleteArg.ShortID, postgresUUID(shortID))
	}
	if !tx.committed {
		t.Fatal("UnlikePublicShort() committed = false, want true")
	}
}

func TestLikePublicShortReturnsNotFound(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tx := &stubTx{}

	repo := newRepository(
		stubBeginner{tx: tx},
		repositoryStubQueries{},
		func(sqlc.DBTX) queries {
			return repositoryStubQueries{
				getPublicShort: func(context.Context, pgtype.UUID) (sqlc.AppPublicShort, error) {
					return sqlc.AppPublicShort{}, pgx.ErrNoRows
				},
			}
		},
	)

	_, err := repo.LikePublicShort(context.Background(), viewerID, shortID)
	if !errors.Is(err, ErrShortNotFound) {
		t.Fatalf("LikePublicShort() error got %v want %v", err, ErrShortNotFound)
	}
	if tx.committed {
		t.Fatal("LikePublicShort() committed = true, want false")
	}
	if !tx.rolledBack {
		t.Fatal("LikePublicShort() rolledBack = false, want true")
	}
}
