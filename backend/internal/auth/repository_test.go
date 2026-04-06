package auth

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

type stubQueries struct {
	touchAuthSessionLastSeenByTokenHash func(context.Context, sqlc.TouchAuthSessionLastSeenByTokenHashParams) (sqlc.AppAuthSession, error)
	getCurrentViewerBySessionTokenHash  func(context.Context, string) (sqlc.GetCurrentViewerBySessionTokenHashRow, error)
}

func (s stubQueries) TouchAuthSessionLastSeenByTokenHash(
	ctx context.Context,
	arg sqlc.TouchAuthSessionLastSeenByTokenHashParams,
) (sqlc.AppAuthSession, error) {
	if s.touchAuthSessionLastSeenByTokenHash == nil {
		return sqlc.AppAuthSession{}, nil
	}

	return s.touchAuthSessionLastSeenByTokenHash(ctx, arg)
}

func (s stubQueries) GetCurrentViewerBySessionTokenHash(
	ctx context.Context,
	sessionTokenHash string,
) (sqlc.GetCurrentViewerBySessionTokenHashRow, error) {
	return s.getCurrentViewerBySessionTokenHash(ctx, sessionTokenHash)
}

func TestGetCurrentViewerBySessionTokenHash(t *testing.T) {
	t.Parallel()

	expectedID := uuid.New()
	repository := newRepository(stubQueries{
		getCurrentViewerBySessionTokenHash: func(context.Context, string) (sqlc.GetCurrentViewerBySessionTokenHashRow, error) {
			return sqlc.GetCurrentViewerBySessionTokenHashRow{
				UserID:               pgUUID(expectedID),
				ActiveMode:           "creator",
				CanAccessCreatorMode: true,
			}, nil
		},
	})

	got, err := repository.GetCurrentViewerBySessionTokenHash(context.Background(), "session-token-hash")
	if err != nil {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() error = %v, want nil", err)
	}
	if got.ID != expectedID {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() id got %s want %s", got.ID, expectedID)
	}
	if got.ActiveMode != ActiveModeCreator {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() active mode got %q want %q", got.ActiveMode, ActiveModeCreator)
	}
	if !got.CanAccessCreatorMode {
		t.Fatal("GetCurrentViewerBySessionTokenHash() can access creator mode = false, want true")
	}
}

func TestGetCurrentViewerBySessionTokenHashNotFound(t *testing.T) {
	t.Parallel()

	repository := newRepository(stubQueries{
		getCurrentViewerBySessionTokenHash: func(context.Context, string) (sqlc.GetCurrentViewerBySessionTokenHashRow, error) {
			return sqlc.GetCurrentViewerBySessionTokenHashRow{}, pgx.ErrNoRows
		},
	})

	if _, err := repository.GetCurrentViewerBySessionTokenHash(context.Background(), "session-token-hash"); !errors.Is(err, ErrCurrentViewerNotFound) {
		t.Fatalf("GetCurrentViewerBySessionTokenHash() error got %v want %v", err, ErrCurrentViewerNotFound)
	}
}

func TestGetCurrentViewerBySessionTokenHashRejectsInvalidUUID(t *testing.T) {
	t.Parallel()

	repository := newRepository(stubQueries{
		getCurrentViewerBySessionTokenHash: func(context.Context, string) (sqlc.GetCurrentViewerBySessionTokenHashRow, error) {
			return sqlc.GetCurrentViewerBySessionTokenHashRow{
				UserID:               pgtype.UUID{},
				ActiveMode:           "fan",
				CanAccessCreatorMode: false,
			}, nil
		},
	})

	if _, err := repository.GetCurrentViewerBySessionTokenHash(context.Background(), "session-token-hash"); err == nil {
		t.Fatal("GetCurrentViewerBySessionTokenHash() error = nil, want conversion error")
	}
}

func TestTouchSessionLastSeenByTokenHash(t *testing.T) {
	t.Parallel()

	expectedID := uuid.New()
	now := time.Unix(1710000000, 0).UTC()
	repository := newRepository(stubQueries{
		touchAuthSessionLastSeenByTokenHash: func(context.Context, sqlc.TouchAuthSessionLastSeenByTokenHashParams) (sqlc.AppAuthSession, error) {
			return sqlc.AppAuthSession{
				ID:               pgUUID(expectedID),
				UserID:           pgUUID(uuid.New()),
				ActiveMode:       "fan",
				SessionTokenHash: "session-token-hash",
				ExpiresAt:        pgTime(now.Add(time.Hour)),
				LastSeenAt:       pgTime(now),
				CreatedAt:        pgTime(now),
				UpdatedAt:        pgTime(now),
			}, nil
		},
	})

	got, err := repository.TouchSessionLastSeenByTokenHash(context.Background(), "session-token-hash", now)
	if err != nil {
		t.Fatalf("TouchSessionLastSeenByTokenHash() error = %v, want nil", err)
	}
	if got.ID != expectedID {
		t.Fatalf("TouchSessionLastSeenByTokenHash() id got %s want %s", got.ID, expectedID)
	}
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

func pgTime(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}
