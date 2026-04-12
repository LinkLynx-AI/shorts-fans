package fanprofile

import (
	"context"
	"errors"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type queriesStub struct {
	getUserByID               func(context.Context, pgtype.UUID) (sqlc.AppUser, error)
	countCreatorFollowsByUser func(context.Context, pgtype.UUID) (int64, error)
	countPinnedShortsByUser   func(context.Context, pgtype.UUID) (int64, error)
	countUnlockedMainsByUser  func(context.Context, pgtype.UUID) (int64, error)
}

func (s queriesStub) GetUserByID(ctx context.Context, id pgtype.UUID) (sqlc.AppUser, error) {
	return s.getUserByID(ctx, id)
}

func (s queriesStub) CountCreatorFollowsByUserID(ctx context.Context, userID pgtype.UUID) (int64, error) {
	return s.countCreatorFollowsByUser(ctx, userID)
}

func (s queriesStub) CountPinnedShortsByUserID(ctx context.Context, userID pgtype.UUID) (int64, error) {
	return s.countPinnedShortsByUser(ctx, userID)
}

func (s queriesStub) CountUnlockedMainsByUserID(ctx context.Context, userID pgtype.UUID) (int64, error) {
	return s.countUnlockedMainsByUser(ctx, userID)
}

func (s queriesStub) ListFanProfileFollowingItems(context.Context, sqlc.ListFanProfileFollowingItemsParams) ([]sqlc.ListFanProfileFollowingItemsRow, error) {
	return nil, nil
}

func (s queriesStub) ListFanProfileLibraryItems(context.Context, sqlc.ListFanProfileLibraryItemsParams) ([]sqlc.ListFanProfileLibraryItemsRow, error) {
	return nil, nil
}

func (s queriesStub) ListFanProfilePinnedShortItems(context.Context, sqlc.ListFanProfilePinnedShortItemsParams) ([]sqlc.ListFanProfilePinnedShortItemsRow, error) {
	return nil, nil
}

func TestGetOverviewPopulated(t *testing.T) {
	t.Parallel()

	expectedUserID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repository := newRepository(queriesStub{
		getUserByID: func(_ context.Context, gotUserID pgtype.UUID) (sqlc.AppUser, error) {
			if gotUserID != postgres.UUIDToPG(expectedUserID) {
				t.Fatalf("GetUserByID() userID got %#v want %#v", gotUserID, postgres.UUIDToPG(expectedUserID))
			}

			return sqlc.AppUser{ID: postgres.UUIDToPG(expectedUserID)}, nil
		},
		countCreatorFollowsByUser: func(_ context.Context, gotUserID pgtype.UUID) (int64, error) {
			if gotUserID != postgres.UUIDToPG(expectedUserID) {
				t.Fatalf("CountCreatorFollowsByUserID() userID got %#v want %#v", gotUserID, postgres.UUIDToPG(expectedUserID))
			}

			return 3, nil
		},
		countPinnedShortsByUser: func(_ context.Context, gotUserID pgtype.UUID) (int64, error) {
			if gotUserID != postgres.UUIDToPG(expectedUserID) {
				t.Fatalf("CountPinnedShortsByUserID() userID got %#v want %#v", gotUserID, postgres.UUIDToPG(expectedUserID))
			}

			return 2, nil
		},
		countUnlockedMainsByUser: func(_ context.Context, gotUserID pgtype.UUID) (int64, error) {
			if gotUserID != postgres.UUIDToPG(expectedUserID) {
				t.Fatalf("CountUnlockedMainsByUserID() userID got %#v want %#v", gotUserID, postgres.UUIDToPG(expectedUserID))
			}

			return 2, nil
		},
	})

	got, err := repository.GetOverview(context.Background(), expectedUserID)
	if err != nil {
		t.Fatalf("GetOverview() error = %v, want nil", err)
	}
	if got.Title != overviewTitle {
		t.Fatalf("GetOverview() title got %q want %q", got.Title, overviewTitle)
	}
	if got.Counts.Following != 3 {
		t.Fatalf("GetOverview() following got %d want %d", got.Counts.Following, 3)
	}
	if got.Counts.PinnedShorts != 2 {
		t.Fatalf("GetOverview() pinnedShorts got %d want %d", got.Counts.PinnedShorts, 2)
	}
	if got.Counts.Library != 2 {
		t.Fatalf("GetOverview() library got %d want %d", got.Counts.Library, 2)
	}
}

func TestGetOverviewEmpty(t *testing.T) {
	t.Parallel()

	viewerUserID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	repository := newRepository(queriesStub{
		getUserByID: func(context.Context, pgtype.UUID) (sqlc.AppUser, error) {
			return sqlc.AppUser{ID: postgres.UUIDToPG(viewerUserID)}, nil
		},
		countCreatorFollowsByUser: func(context.Context, pgtype.UUID) (int64, error) {
			return 0, nil
		},
		countPinnedShortsByUser: func(context.Context, pgtype.UUID) (int64, error) {
			return 0, nil
		},
		countUnlockedMainsByUser: func(context.Context, pgtype.UUID) (int64, error) {
			return 0, nil
		},
	})

	got, err := repository.GetOverview(context.Background(), viewerUserID)
	if err != nil {
		t.Fatalf("GetOverview() error = %v, want nil", err)
	}
	if got.Counts.Following != 0 || got.Counts.PinnedShorts != 0 || got.Counts.Library != 0 {
		t.Fatalf("GetOverview() counts got %#v want all zero", got.Counts)
	}
}

func TestGetOverviewMapsMissingUserToNotFound(t *testing.T) {
	t.Parallel()

	repository := newRepository(queriesStub{
		getUserByID: func(context.Context, pgtype.UUID) (sqlc.AppUser, error) {
			return sqlc.AppUser{}, pgx.ErrNoRows
		},
		countCreatorFollowsByUser: func(context.Context, pgtype.UUID) (int64, error) {
			t.Fatal("CountCreatorFollowsByUserID() was called for missing user")
			return 0, nil
		},
		countPinnedShortsByUser: func(context.Context, pgtype.UUID) (int64, error) {
			t.Fatal("CountPinnedShortsByUserID() was called for missing user")
			return 0, nil
		},
		countUnlockedMainsByUser: func(context.Context, pgtype.UUID) (int64, error) {
			t.Fatal("CountUnlockedMainsByUserID() was called for missing user")
			return 0, nil
		},
	})

	_, err := repository.GetOverview(context.Background(), uuid.MustParse("33333333-3333-3333-3333-333333333333"))
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("GetOverview() error got %v want ErrProfileNotFound", err)
	}
}

func TestGetOverviewPropagatesCountErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("count failed")
	repository := newRepository(queriesStub{
		getUserByID: func(context.Context, pgtype.UUID) (sqlc.AppUser, error) {
			return sqlc.AppUser{}, nil
		},
		countCreatorFollowsByUser: func(context.Context, pgtype.UUID) (int64, error) {
			return 0, expectedErr
		},
		countPinnedShortsByUser: func(context.Context, pgtype.UUID) (int64, error) {
			t.Fatal("CountPinnedShortsByUserID() was called after following error")
			return 0, nil
		},
		countUnlockedMainsByUser: func(context.Context, pgtype.UUID) (int64, error) {
			t.Fatal("CountUnlockedMainsByUserID() was called after following error")
			return 0, nil
		},
	})

	_, err := repository.GetOverview(context.Background(), uuid.MustParse("44444444-4444-4444-4444-444444444444"))
	if !errors.Is(err, expectedErr) {
		t.Fatalf("GetOverview() error got %v want %v", err, expectedErr)
	}
}
