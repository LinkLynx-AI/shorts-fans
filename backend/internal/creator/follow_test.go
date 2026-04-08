package creator

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type creatorTxBeginnerStub struct {
	beginErr error
	tx       pgx.Tx
	began    bool
}

func (s *creatorTxBeginnerStub) Begin(context.Context) (pgx.Tx, error) {
	s.began = true
	if s.beginErr != nil {
		return nil, s.beginErr
	}

	return s.tx, nil
}

type creatorTxStub struct {
	commitErr   error
	rollbackErr error
	committed   bool
	rolledBack  bool
}

type creatorRowStub struct {
	err error
}

func (s creatorRowStub) Scan(...any) error {
	return s.err
}

func (s *creatorTxStub) Begin(context.Context) (pgx.Tx, error) {
	return nil, fmt.Errorf("unexpected nested Begin call")
}

func (s *creatorTxStub) Commit(context.Context) error {
	s.committed = true
	return s.commitErr
}

func (s *creatorTxStub) Rollback(context.Context) error {
	s.rolledBack = true
	return s.rollbackErr
}

func (s *creatorTxStub) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, fmt.Errorf("unexpected CopyFrom call")
}

func (s *creatorTxStub) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults {
	return nil
}

func (s *creatorTxStub) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (s *creatorTxStub) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, fmt.Errorf("unexpected Prepare call")
}

func (s *creatorTxStub) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, fmt.Errorf("unexpected Exec call")
}

func (s *creatorTxStub) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, fmt.Errorf("unexpected Query call")
}

func (s *creatorTxStub) QueryRow(context.Context, string, ...any) pgx.Row {
	return creatorRowStub{err: fmt.Errorf("unexpected QueryRow call")}
}

func (s *creatorTxStub) Conn() *pgx.Conn {
	return nil
}

func TestFollowPublicCreator(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	creatorUserID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tx := &creatorTxStub{}
	beginner := &creatorTxBeginnerStub{tx: tx}

	var gotLookupUserID pgtype.UUID
	var gotPutArg sqlc.PutCreatorFollowParams
	var gotCountUserID pgtype.UUID

	repo := &Repository{
		txBeginner: beginner,
		newQueries: func(sqlc.DBTX) queries {
			return repositoryStubQueries{
				getPublicProfile: func(_ context.Context, userID pgtype.UUID) (sqlc.AppPublicCreatorProfile, error) {
					gotLookupUserID = userID
					return testPublicProfileRow(creatorUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), stringPtr("https://cdn.example.com/mina.jpg"), timePtr(now)), nil
				},
				putCreatorFollow: func(_ context.Context, arg sqlc.PutCreatorFollowParams) error {
					gotPutArg = arg
					return nil
				},
				countFollowers: func(_ context.Context, creatorUserID pgtype.UUID) (int64, error) {
					gotCountUserID = creatorUserID
					return 24, nil
				},
			}
		},
	}

	got, err := repo.FollowPublicCreator(context.Background(), viewerID, FormatPublicID(creatorUserID))
	if err != nil {
		t.Fatalf("FollowPublicCreator() error = %v, want nil", err)
	}
	if !got.IsFollowing {
		t.Fatal("FollowPublicCreator() isFollowing = false, want true")
	}
	if got.FanCount != 24 {
		t.Fatalf("FollowPublicCreator() fan count got %d want %d", got.FanCount, 24)
	}
	if gotLookupUserID != pgUUID(creatorUserID) {
		t.Fatalf("FollowPublicCreator() public profile lookup arg got %v want %v", gotLookupUserID, pgUUID(creatorUserID))
	}
	if gotPutArg.UserID != pgUUID(viewerID) {
		t.Fatalf("FollowPublicCreator() put user arg got %v want %v", gotPutArg.UserID, pgUUID(viewerID))
	}
	if gotPutArg.CreatorUserID != pgUUID(creatorUserID) {
		t.Fatalf("FollowPublicCreator() put creator arg got %v want %v", gotPutArg.CreatorUserID, pgUUID(creatorUserID))
	}
	if gotCountUserID != pgUUID(creatorUserID) {
		t.Fatalf("FollowPublicCreator() count arg got %v want %v", gotCountUserID, pgUUID(creatorUserID))
	}
	if !beginner.began {
		t.Fatal("FollowPublicCreator() began = false, want true")
	}
	if !tx.committed {
		t.Fatal("FollowPublicCreator() committed = false, want true")
	}
	if tx.rolledBack {
		t.Fatal("FollowPublicCreator() rolledBack = true, want false")
	}
}

func TestUnfollowPublicCreator(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	creatorUserID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tx := &creatorTxStub{}
	beginner := &creatorTxBeginnerStub{tx: tx}

	var gotDeleteArg sqlc.DeleteCreatorFollowParams

	repo := &Repository{
		txBeginner: beginner,
		newQueries: func(sqlc.DBTX) queries {
			return repositoryStubQueries{
				getPublicProfile: func(_ context.Context, userID pgtype.UUID) (sqlc.AppPublicCreatorProfile, error) {
					return testPublicProfileRow(creatorUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), stringPtr("https://cdn.example.com/mina.jpg"), timePtr(now)), nil
				},
				deleteCreatorFollow: func(_ context.Context, arg sqlc.DeleteCreatorFollowParams) error {
					gotDeleteArg = arg
					return nil
				},
				countFollowers: func(_ context.Context, creatorUserID pgtype.UUID) (int64, error) {
					return 23, nil
				},
			}
		},
	}

	got, err := repo.UnfollowPublicCreator(context.Background(), viewerID, FormatPublicID(creatorUserID))
	if err != nil {
		t.Fatalf("UnfollowPublicCreator() error = %v, want nil", err)
	}
	if got.IsFollowing {
		t.Fatal("UnfollowPublicCreator() isFollowing = true, want false")
	}
	if got.FanCount != 23 {
		t.Fatalf("UnfollowPublicCreator() fan count got %d want %d", got.FanCount, 23)
	}
	if gotDeleteArg.UserID != pgUUID(viewerID) {
		t.Fatalf("UnfollowPublicCreator() delete user arg got %v want %v", gotDeleteArg.UserID, pgUUID(viewerID))
	}
	if gotDeleteArg.CreatorUserID != pgUUID(creatorUserID) {
		t.Fatalf("UnfollowPublicCreator() delete creator arg got %v want %v", gotDeleteArg.CreatorUserID, pgUUID(creatorUserID))
	}
	if !tx.committed {
		t.Fatal("UnfollowPublicCreator() committed = false, want true")
	}
}

func TestFollowPublicCreatorReturnsProfileNotFound(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	creatorUserID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	tx := &creatorTxStub{}
	beginner := &creatorTxBeginnerStub{tx: tx}

	repo := &Repository{
		txBeginner: beginner,
		newQueries: func(sqlc.DBTX) queries {
			return repositoryStubQueries{
				getPublicProfile: func(context.Context, pgtype.UUID) (sqlc.AppPublicCreatorProfile, error) {
					return sqlc.AppPublicCreatorProfile{}, pgx.ErrNoRows
				},
			}
		},
	}

	_, err := repo.FollowPublicCreator(context.Background(), viewerID, FormatPublicID(creatorUserID))
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("FollowPublicCreator() error got %v want %v", err, ErrProfileNotFound)
	}
	if tx.committed {
		t.Fatal("FollowPublicCreator() committed = true, want false")
	}
	if !tx.rolledBack {
		t.Fatal("FollowPublicCreator() rolledBack = false, want true")
	}
}
