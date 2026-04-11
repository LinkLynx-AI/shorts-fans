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

func TestPinPublicShort(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tx := &stubTx{}

	var gotPublicShortID pgtype.UUID
	var gotPutArg sqlc.PutPinnedShortParams

	repo := newRepository(
		stubBeginner{tx: tx},
		repositoryStubQueries{},
		func(sqlc.DBTX) queries {
			return repositoryStubQueries{
				getPublicShort: func(_ context.Context, id pgtype.UUID) (sqlc.AppPublicShort, error) {
					gotPublicShortID = id
					return testPublicShortRow(shortID, uuid.New(), uuid.New(), uuid.New(), now, nil, nil, timePtr(now), timePtr(now)), nil
				},
				putPinnedShort: func(_ context.Context, arg sqlc.PutPinnedShortParams) error {
					gotPutArg = arg
					return nil
				},
				deletePinnedShort: func(context.Context, sqlc.DeletePinnedShortParams) error {
					t.Fatal("DeletePinnedShort() was called during pin")
					return nil
				},
			}
		},
	)

	got, err := repo.PinPublicShort(context.Background(), viewerID, shortID)
	if err != nil {
		t.Fatalf("PinPublicShort() error = %v, want nil", err)
	}
	if !got.IsPinned {
		t.Fatal("PinPublicShort() isPinned = false, want true")
	}
	if gotPublicShortID != postgresUUID(shortID) {
		t.Fatalf("PinPublicShort() public short id arg got %v want %v", gotPublicShortID, postgresUUID(shortID))
	}
	if gotPutArg.UserID != postgresUUID(viewerID) {
		t.Fatalf("PinPublicShort() user arg got %v want %v", gotPutArg.UserID, postgresUUID(viewerID))
	}
	if gotPutArg.ShortID != postgresUUID(shortID) {
		t.Fatalf("PinPublicShort() short arg got %v want %v", gotPutArg.ShortID, postgresUUID(shortID))
	}
	if !tx.committed {
		t.Fatal("PinPublicShort() committed = false, want true")
	}
	if tx.rolledBack {
		t.Fatal("PinPublicShort() rolledBack = true, want false")
	}
}

func TestUnpinPublicShort(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	shortID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tx := &stubTx{}

	var gotDeleteArg sqlc.DeletePinnedShortParams

	repo := newRepository(
		stubBeginner{tx: tx},
		repositoryStubQueries{},
		func(sqlc.DBTX) queries {
			return repositoryStubQueries{
				getPublicShort: func(_ context.Context, id pgtype.UUID) (sqlc.AppPublicShort, error) {
					return testPublicShortRow(shortID, uuid.New(), uuid.New(), uuid.New(), now, nil, nil, timePtr(now), timePtr(now)), nil
				},
				putPinnedShort: func(context.Context, sqlc.PutPinnedShortParams) error {
					t.Fatal("PutPinnedShort() was called during unpin")
					return nil
				},
				deletePinnedShort: func(_ context.Context, arg sqlc.DeletePinnedShortParams) error {
					gotDeleteArg = arg
					return nil
				},
			}
		},
	)

	got, err := repo.UnpinPublicShort(context.Background(), viewerID, shortID)
	if err != nil {
		t.Fatalf("UnpinPublicShort() error = %v, want nil", err)
	}
	if got.IsPinned {
		t.Fatal("UnpinPublicShort() isPinned = true, want false")
	}
	if gotDeleteArg.UserID != postgresUUID(viewerID) {
		t.Fatalf("UnpinPublicShort() user arg got %v want %v", gotDeleteArg.UserID, postgresUUID(viewerID))
	}
	if gotDeleteArg.ShortID != postgresUUID(shortID) {
		t.Fatalf("UnpinPublicShort() short arg got %v want %v", gotDeleteArg.ShortID, postgresUUID(shortID))
	}
	if !tx.committed {
		t.Fatal("UnpinPublicShort() committed = false, want true")
	}
}

func TestPinPublicShortReturnsNotFound(t *testing.T) {
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

	_, err := repo.PinPublicShort(context.Background(), viewerID, shortID)
	if !errors.Is(err, ErrShortNotFound) {
		t.Fatalf("PinPublicShort() error got %v want %v", err, ErrShortNotFound)
	}
	if tx.committed {
		t.Fatal("PinPublicShort() committed = true, want false")
	}
	if !tx.rolledBack {
		t.Fatal("PinPublicShort() rolledBack = false, want true")
	}
}
