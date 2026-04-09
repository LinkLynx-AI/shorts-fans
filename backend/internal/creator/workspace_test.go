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

func TestGetWorkspace(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	displayName := stringPtr("Mina Rei")
	handle := stringPtr("minarei")
	avatarURL := stringPtr("https://cdn.example.com/mina.jpg")
	approvedAt := timePtr(now.Add(time.Hour))

	var gotMetricsUserID pgtype.UUID
	var gotRevisionUserID pgtype.UUID

	repo := newRepository(repositoryStubQueries{
		getCapability: func(_ context.Context, gotUserID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			if gotUserID != pgUUID(userID) {
				t.Fatalf("GetCapability() user got %v want %v", gotUserID, pgUUID(userID))
			}
			return testCapabilityRow(userID, now, nil, nil, nil, nil, approvedAt, nil, nil), nil
		},
		getProfile: func(_ context.Context, gotUserID pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			if gotUserID != pgUUID(userID) {
				t.Fatalf("GetProfile() user got %v want %v", gotUserID, pgUUID(userID))
			}
			return testProfileRow(userID, now, displayName, handle, avatarURL, nil), nil
		},
		getWorkspaceMetrics: func(_ context.Context, gotUserID pgtype.UUID) (sqlc.GetCreatorWorkspaceOverviewMetricsRow, error) {
			gotMetricsUserID = gotUserID
			return sqlc.GetCreatorWorkspaceOverviewMetricsRow{
				GrossUnlockRevenueJpy: 120000,
				UnlockCount:           238,
				UniquePurchaserCount:  164,
			}, nil
		},
		getRevisionSummary: func(_ context.Context, gotUserID pgtype.UUID) (sqlc.GetCreatorWorkspaceRevisionRequestedSummaryRow, error) {
			gotRevisionUserID = gotUserID
			return sqlc.GetCreatorWorkspaceRevisionRequestedSummaryRow{
				MainCount:  0,
				ShortCount: 1,
			}, nil
		},
	})

	got, err := repo.GetWorkspace(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetWorkspace() error = %v, want nil", err)
	}
	if got.Creator.UserID != userID {
		t.Fatalf("GetWorkspace() creator user id got %s want %s", got.Creator.UserID, userID)
	}
	if got.OverviewMetrics.GrossUnlockRevenueJpy != 120000 {
		t.Fatalf("GetWorkspace() gross unlock revenue got %d want %d", got.OverviewMetrics.GrossUnlockRevenueJpy, 120000)
	}
	if got.OverviewMetrics.UnlockCount != 238 {
		t.Fatalf("GetWorkspace() unlock count got %d want %d", got.OverviewMetrics.UnlockCount, 238)
	}
	if got.OverviewMetrics.UniquePurchaserCount != 164 {
		t.Fatalf("GetWorkspace() unique purchaser count got %d want %d", got.OverviewMetrics.UniquePurchaserCount, 164)
	}
	if got.RevisionRequestedSummary == nil {
		t.Fatal("GetWorkspace() revision summary = nil, want non-nil")
	}
	if got.RevisionRequestedSummary.TotalCount != 1 {
		t.Fatalf("GetWorkspace() total revision count got %d want %d", got.RevisionRequestedSummary.TotalCount, 1)
	}
	if gotMetricsUserID != pgUUID(userID) {
		t.Fatalf("GetWorkspace() metrics arg got %v want %v", gotMetricsUserID, pgUUID(userID))
	}
	if gotRevisionUserID != pgUUID(userID) {
		t.Fatalf("GetWorkspace() revision arg got %v want %v", gotRevisionUserID, pgUUID(userID))
	}
}

func TestGetWorkspaceReturnsNilRevisionSummaryWhenCountsAreZero(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapabilityRow(userID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testProfileRow(userID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		getWorkspaceMetrics: func(context.Context, pgtype.UUID) (sqlc.GetCreatorWorkspaceOverviewMetricsRow, error) {
			return sqlc.GetCreatorWorkspaceOverviewMetricsRow{}, nil
		},
		getRevisionSummary: func(context.Context, pgtype.UUID) (sqlc.GetCreatorWorkspaceRevisionRequestedSummaryRow, error) {
			return sqlc.GetCreatorWorkspaceRevisionRequestedSummaryRow{}, nil
		},
	})

	got, err := repo.GetWorkspace(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetWorkspace() error = %v, want nil", err)
	}
	if got.RevisionRequestedSummary != nil {
		t.Fatalf("GetWorkspace() revision summary got %#v want nil", got.RevisionRequestedSummary)
	}
}

func TestGetWorkspaceReturnsCreatorModeUnavailableWhenCapabilityIsMissing(t *testing.T) {
	t.Parallel()

	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return sqlc.AppCreatorCapability{}, pgx.ErrNoRows
		},
	})

	if _, err := repo.GetWorkspace(context.Background(), uuid.New()); !errors.Is(err, ErrCreatorModeUnavailable) {
		t.Fatalf("GetWorkspace() error got %v want %v", err, ErrCreatorModeUnavailable)
	}
}

func TestGetWorkspaceReturnsCreatorModeUnavailableWhenCapabilityIsNotApproved(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			row := testCapabilityRow(userID, now, nil, nil, nil, nil, nil, nil, nil)
			row.State = "submitted"
			return row, nil
		},
	})

	if _, err := repo.GetWorkspace(context.Background(), userID); !errors.Is(err, ErrCreatorModeUnavailable) {
		t.Fatalf("GetWorkspace() error got %v want %v", err, ErrCreatorModeUnavailable)
	}
}

func TestGetWorkspaceReturnsProfileNotFoundWhenPrivateProfileIsMissing(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapabilityRow(userID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return sqlc.AppCreatorProfile{}, pgx.ErrNoRows
		},
	})

	if _, err := repo.GetWorkspace(context.Background(), userID); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("GetWorkspace() error got %v want %v", err, ErrProfileNotFound)
	}
}
