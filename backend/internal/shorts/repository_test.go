package shorts

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type stubQueries struct {
	createMain  func(context.Context, sqlc.CreateMainParams) (sqlc.AppMain, error)
	createShort func(context.Context, sqlc.CreateShortParams) (sqlc.AppShort, error)
}

func (s stubQueries) CreateMain(ctx context.Context, arg sqlc.CreateMainParams) (sqlc.AppMain, error) {
	return s.createMain(ctx, arg)
}

func (s stubQueries) GetMainByID(context.Context, pgtype.UUID) (sqlc.AppMain, error) {
	return sqlc.AppMain{}, nil
}

func (s stubQueries) ListMainsByCreatorUserID(context.Context, pgtype.UUID) ([]sqlc.AppMain, error) {
	return nil, nil
}

func (s stubQueries) UpdateMainState(context.Context, sqlc.UpdateMainStateParams) (sqlc.AppMain, error) {
	return sqlc.AppMain{}, nil
}

func (s stubQueries) GetUnlockableMainByID(context.Context, pgtype.UUID) (sqlc.AppUnlockableMain, error) {
	return sqlc.AppUnlockableMain{}, nil
}

func (s stubQueries) CreateShort(ctx context.Context, arg sqlc.CreateShortParams) (sqlc.AppShort, error) {
	return s.createShort(ctx, arg)
}

func (s stubQueries) GetShortByID(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
	return sqlc.AppShort{}, nil
}

func (s stubQueries) ListShortsByCreatorUserID(context.Context, pgtype.UUID) ([]sqlc.AppShort, error) {
	return nil, nil
}

func (s stubQueries) UpdateShortState(context.Context, sqlc.UpdateShortStateParams) (sqlc.AppShort, error) {
	return sqlc.AppShort{}, nil
}

func (s stubQueries) PublishShort(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
	return sqlc.AppShort{}, nil
}

func (s stubQueries) ListPublicShortsByCreatorUserID(context.Context, pgtype.UUID) ([]sqlc.AppPublicShort, error) {
	return nil, nil
}

func (s stubQueries) GetPublicShortByID(context.Context, pgtype.UUID) (sqlc.AppPublicShort, error) {
	return sqlc.AppPublicShort{}, nil
}

func (s stubQueries) ListShortsByCanonicalMainID(context.Context, pgtype.UUID) ([]sqlc.AppShort, error) {
	return nil, nil
}

func (s stubQueries) GetCanonicalMainIDByShortID(context.Context, pgtype.UUID) (pgtype.UUID, error) {
	return pgtype.UUID{}, nil
}

type stubBeginner struct {
	tx pgx.Tx
}

func (s stubBeginner) Begin(context.Context) (pgx.Tx, error) {
	return s.tx, nil
}

type stubTx struct {
	committed  bool
	rolledBack bool
}

func (tx *stubTx) Begin(context.Context) (pgx.Tx, error) { return tx, nil }
func (tx *stubTx) Commit(context.Context) error          { tx.committed = true; return nil }
func (tx *stubTx) Rollback(context.Context) error        { tx.rolledBack = true; return nil }
func (tx *stubTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (tx *stubTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (tx *stubTx) LargeObjects() pgx.LargeObjects                         { return pgx.LargeObjects{} }
func (tx *stubTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (tx *stubTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (tx *stubTx) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }
func (tx *stubTx) QueryRow(context.Context, string, ...any) pgx.Row        { return nil }
func (tx *stubTx) Conn() *pgx.Conn                                         { return nil }

func TestCreateMainWithShortsRequiresShorts(t *testing.T) {
	t.Parallel()

	repo := newRepository(nil, stubQueries{
		createMain: func(context.Context, sqlc.CreateMainParams) (sqlc.AppMain, error) {
			return sqlc.AppMain{}, nil
		},
		createShort: func(context.Context, sqlc.CreateShortParams) (sqlc.AppShort, error) {
			return sqlc.AppShort{}, nil
		},
	}, nil)

	_, err := repo.CreateMainWithShorts(context.Background(), CreateMainWithShortsInput{})
	if !errors.Is(err, ErrLinkedShortsRequired) {
		t.Fatalf("CreateMainWithShorts() error got %v want %v", err, ErrLinkedShortsRequired)
	}
}

func TestCreateMainWithShortsRollsBackOnShortFailure(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	mainID := postgresUUID(uuid.New())
	mainMediaID := uuid.New()
	creatorID := uuid.New()
	shortMediaID := uuid.New()

	tx := &stubTx{}
	repo := newRepository(
		stubBeginner{tx: tx},
		stubQueries{
			createMain: func(context.Context, sqlc.CreateMainParams) (sqlc.AppMain, error) {
				return sqlc.AppMain{
					ID:                 mainID,
					CreatorUserID:      postgresUUID(creatorID),
					MediaAssetID:       postgresUUID(mainMediaID),
					State:              "draft",
					OwnershipConfirmed: true,
					ConsentConfirmed:   true,
					CreatedAt:          timestamp(now),
					UpdatedAt:          timestamp(now),
				}, nil
			},
			createShort: func(context.Context, sqlc.CreateShortParams) (sqlc.AppShort, error) {
				return sqlc.AppShort{}, errors.New("short insert failed")
			},
		},
		func(sqlc.DBTX) queries {
			return stubQueries{
				createMain: func(context.Context, sqlc.CreateMainParams) (sqlc.AppMain, error) {
					return sqlc.AppMain{
						ID:                 mainID,
						CreatorUserID:      postgresUUID(creatorID),
						MediaAssetID:       postgresUUID(mainMediaID),
						State:              "draft",
						OwnershipConfirmed: true,
						ConsentConfirmed:   true,
						CreatedAt:          timestamp(now),
						UpdatedAt:          timestamp(now),
					}, nil
				},
				createShort: func(context.Context, sqlc.CreateShortParams) (sqlc.AppShort, error) {
					return sqlc.AppShort{}, errors.New("short insert failed")
				},
			}
		},
	)

	_, err := repo.CreateMainWithShorts(context.Background(), CreateMainWithShortsInput{
		Main: CreateMainInput{
			CreatorUserID:      creatorID,
			MediaAssetID:       mainMediaID,
			State:              "draft",
			OwnershipConfirmed: true,
			ConsentConfirmed:   true,
		},
		Shorts: []CreateLinkedShortInput{
			{
				MediaAssetID: shortMediaID,
				State:        "draft",
			},
		},
	})
	if err == nil {
		t.Fatal("CreateMainWithShorts() error = nil, want rollback path")
	}
	if !tx.rolledBack {
		t.Fatal("CreateMainWithShorts() rollback = false, want true")
	}
	if tx.committed {
		t.Fatal("CreateMainWithShorts() committed = true, want false")
	}
}

func postgresUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

func timestamp(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}
