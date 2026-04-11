package unlock

import (
	"context"
	"errors"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type stubQueries struct{}

func (stubQueries) CreateMainUnlock(context.Context, sqlc.CreateMainUnlockParams) (sqlc.AppMainUnlock, error) {
	return sqlc.AppMainUnlock{}, &pgconn.PgError{
		Code:           "23505",
		ConstraintName: "main_unlocks_pkey",
	}
}

func (stubQueries) EnsureMainUnlock(context.Context, sqlc.EnsureMainUnlockParams) (sqlc.EnsureMainUnlockRow, error) {
	return sqlc.EnsureMainUnlockRow{}, nil
}

func (stubQueries) GetMainUnlockByUserIDAndMainID(context.Context, sqlc.GetMainUnlockByUserIDAndMainIDParams) (sqlc.AppMainUnlock, error) {
	return sqlc.AppMainUnlock{}, nil
}

func (stubQueries) ListUnlockedMainIDsByUserID(context.Context, pgtype.UUID) ([]pgtype.UUID, error) {
	return nil, nil
}

func TestRecordMainUnlockAlreadyUnlocked(t *testing.T) {
	t.Parallel()

	repo := newRepository(stubQueries{})

	_, err := repo.RecordMainUnlock(context.Background(), RecordMainUnlockInput{
		UserID: uuid.New(),
		MainID: uuid.New(),
	})
	if !errors.Is(err, ErrAlreadyUnlocked) {
		t.Fatalf("RecordMainUnlock() error got %v want %v", err, ErrAlreadyUnlocked)
	}
}
