package creator

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

func TestUpdateWorkspaceMainPrice(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("a1111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("b2222222-2222-2222-2222-222222222222")

	repo := newRepository(repositoryStubQueries{
		getCapability: func(_ context.Context, gotUserID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			if gotUserID != pgUUID(viewerUserID) {
				t.Fatalf("GetCreatorCapabilityByUserID() user got %v want %v", gotUserID, pgUUID(viewerUserID))
			}
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(_ context.Context, gotUserID pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			if gotUserID != pgUUID(viewerUserID) {
				t.Fatalf("GetCreatorProfileByUserID() user got %v want %v", gotUserID, pgUUID(viewerUserID))
			}
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		updateWorkspaceMainPrice: func(_ context.Context, arg sqlc.UpdateCreatorWorkspaceMainPriceParams) (sqlc.UpdateCreatorWorkspaceMainPriceRow, error) {
			if arg.MainID != pgUUID(mainID) {
				t.Fatalf("UpdateCreatorWorkspaceMainPrice() main got %v want %v", arg.MainID, pgUUID(mainID))
			}
			if arg.OwnerUserID != pgUUID(viewerUserID) {
				t.Fatalf("UpdateCreatorWorkspaceMainPrice() owner got %v want %v", arg.OwnerUserID, pgUUID(viewerUserID))
			}
			if arg.PriceMinor != 2400 {
				t.Fatalf("UpdateCreatorWorkspaceMainPrice() price got %d want %d", arg.PriceMinor, 2400)
			}

			return sqlc.UpdateCreatorWorkspaceMainPriceRow{
				CurrencyCode: workspaceMainPriceCurrencyCode,
				ID:           pgUUID(mainID),
				PriceMinor:   2400,
			}, nil
		},
	})

	got, err := repo.UpdateWorkspaceMainPrice(context.Background(), viewerUserID, mainID, 2400)
	if err != nil {
		t.Fatalf("UpdateWorkspaceMainPrice() error = %v, want nil", err)
	}
	if got.ID != mainID {
		t.Fatalf("UpdateWorkspaceMainPrice() id got %s want %s", got.ID, mainID)
	}
	if got.PriceJpy != 2400 {
		t.Fatalf("UpdateWorkspaceMainPrice() price got %d want %d", got.PriceJpy, 2400)
	}
}

func TestUpdateWorkspaceMainPriceRejectsNonPositivePrice(t *testing.T) {
	t.Parallel()

	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			t.Fatal("GetCreatorCapabilityByUserID() should not be called")
			return sqlc.AppCreatorCapability{}, nil
		},
	})

	if _, err := repo.UpdateWorkspaceMainPrice(context.Background(), uuid.New(), uuid.New(), 0); !errors.Is(err, ErrInvalidWorkspaceMainPrice) {
		t.Fatalf("UpdateWorkspaceMainPrice() error got %v want %v", err, ErrInvalidWorkspaceMainPrice)
	}
}

func TestUpdateWorkspaceMainPriceReturnsNotFoundWhenMainMissing(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("c3333333-3333-3333-3333-333333333333")
	mainID := uuid.MustParse("d4444444-4444-4444-4444-444444444444")

	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		updateWorkspaceMainPrice: func(context.Context, sqlc.UpdateCreatorWorkspaceMainPriceParams) (sqlc.UpdateCreatorWorkspaceMainPriceRow, error) {
			return sqlc.UpdateCreatorWorkspaceMainPriceRow{}, pgx.ErrNoRows
		},
	})

	if _, err := repo.UpdateWorkspaceMainPrice(context.Background(), viewerUserID, mainID, 1800); !errors.Is(err, ErrWorkspaceMainNotFound) {
		t.Fatalf("UpdateWorkspaceMainPrice() error got %v want %v", err, ErrWorkspaceMainNotFound)
	}
}

func TestUpdateWorkspaceMainPriceReturnsCreatorModeUnavailableWhenCapabilityMissing(t *testing.T) {
	t.Parallel()

	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return sqlc.AppCreatorCapability{}, pgx.ErrNoRows
		},
	})

	if _, err := repo.UpdateWorkspaceMainPrice(context.Background(), uuid.New(), uuid.New(), 1800); !errors.Is(err, ErrCreatorModeUnavailable) {
		t.Fatalf("UpdateWorkspaceMainPrice() error got %v want %v", err, ErrCreatorModeUnavailable)
	}
}
