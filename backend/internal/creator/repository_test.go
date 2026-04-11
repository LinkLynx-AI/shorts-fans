package creator

import (
	"context"
	"errors"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type stubQueries struct {
	getCapability func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error)
	getProfile    func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error)
}

func (s stubQueries) CountCreatorFollowersByCreatorUserID(context.Context, pgtype.UUID) (int64, error) {
	return 0, nil
}

func (s stubQueries) CreateCreatorCapability(context.Context, sqlc.CreateCreatorCapabilityParams) (sqlc.AppCreatorCapability, error) {
	return sqlc.AppCreatorCapability{}, nil
}

func (s stubQueries) GetCreatorCapabilityByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
	return s.getCapability(ctx, userID)
}

func (s stubQueries) GetCreatorWorkspaceOverviewMetrics(context.Context, pgtype.UUID) (sqlc.GetCreatorWorkspaceOverviewMetricsRow, error) {
	return sqlc.GetCreatorWorkspaceOverviewMetricsRow{}, nil
}

func (s stubQueries) ListCreatorWorkspaceTopMainCandidatesByCreatorUserID(context.Context, pgtype.UUID) ([]sqlc.ListCreatorWorkspaceTopMainCandidatesByCreatorUserIDRow, error) {
	return nil, nil
}

func (s stubQueries) ListCreatorWorkspaceTopShortCandidatesByCreatorUserID(context.Context, pgtype.UUID) ([]sqlc.ListCreatorWorkspaceTopShortCandidatesByCreatorUserIDRow, error) {
	return nil, nil
}

func (s stubQueries) ListCreatorWorkspacePreviewMainsByCreatorUserID(context.Context, pgtype.UUID) ([]sqlc.ListCreatorWorkspacePreviewMainsByCreatorUserIDRow, error) {
	return nil, nil
}

func (s stubQueries) GetCreatorWorkspaceRevisionRequestedSummary(context.Context, pgtype.UUID) (sqlc.GetCreatorWorkspaceRevisionRequestedSummaryRow, error) {
	return sqlc.GetCreatorWorkspaceRevisionRequestedSummaryRow{}, nil
}

func (s stubQueries) GetMediaAssetByID(context.Context, pgtype.UUID) (sqlc.AppMediaAsset, error) {
	return sqlc.AppMediaAsset{}, nil
}

func (s stubQueries) UpdateCreatorCapabilityState(context.Context, sqlc.UpdateCreatorCapabilityStateParams) (sqlc.AppCreatorCapability, error) {
	return sqlc.AppCreatorCapability{}, nil
}

func (s stubQueries) CreateCreatorProfile(context.Context, sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
	return sqlc.AppCreatorProfile{}, nil
}

func (s stubQueries) CountPublicShortsByCreatorUserID(context.Context, pgtype.UUID) (int64, error) {
	return 0, nil
}

func (s stubQueries) DeleteCreatorFollow(context.Context, sqlc.DeleteCreatorFollowParams) error {
	return nil
}

func (s stubQueries) GetCreatorProfileByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorProfile, error) {
	return s.getProfile(ctx, userID)
}

func (s stubQueries) GetViewerCreatorFollowState(context.Context, sqlc.GetViewerCreatorFollowStateParams) (bool, error) {
	return false, nil
}

func (s stubQueries) GetPublicCreatorProfileByUserID(context.Context, pgtype.UUID) (sqlc.AppPublicCreatorProfile, error) {
	return sqlc.AppPublicCreatorProfile{}, nil
}

func (s stubQueries) GetPublicCreatorProfileByHandle(context.Context, string) (sqlc.AppPublicCreatorProfile, error) {
	return sqlc.AppPublicCreatorProfile{}, nil
}

func (s stubQueries) ListMainsByCreatorUserID(context.Context, pgtype.UUID) ([]sqlc.AppMain, error) {
	return nil, nil
}

func (s stubQueries) ListCreatorProfileShortGridItems(context.Context, sqlc.ListCreatorProfileShortGridItemsParams) ([]sqlc.ListCreatorProfileShortGridItemsRow, error) {
	return nil, nil
}

func (s stubQueries) ListRecentPublicCreatorProfiles(context.Context, sqlc.ListRecentPublicCreatorProfilesParams) ([]sqlc.AppPublicCreatorProfile, error) {
	return nil, nil
}

func (s stubQueries) ListShortsByCreatorUserID(context.Context, pgtype.UUID) ([]sqlc.AppShort, error) {
	return nil, nil
}

func (s stubQueries) PutCreatorFollow(context.Context, sqlc.PutCreatorFollowParams) error {
	return nil
}

func (s stubQueries) SearchPublicCreatorProfiles(context.Context, sqlc.SearchPublicCreatorProfilesParams) ([]sqlc.AppPublicCreatorProfile, error) {
	return nil, nil
}

func (s stubQueries) UpdateCreatorProfile(context.Context, sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
	return sqlc.AppCreatorProfile{}, nil
}

func (s stubQueries) PublishCreatorProfile(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
	return sqlc.AppCreatorProfile{}, nil
}

func TestGetCapabilityNotFound(t *testing.T) {
	t.Parallel()

	repo := newRepository(stubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return sqlc.AppCreatorCapability{}, pgx.ErrNoRows
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return sqlc.AppCreatorProfile{}, nil
		},
	})

	_, err := repo.GetCapability(context.Background(), uuid.New())
	if !errors.Is(err, ErrCapabilityNotFound) {
		t.Fatalf("GetCapability() error got %v want %v", err, ErrCapabilityNotFound)
	}
}

func TestGetProfileNotFound(t *testing.T) {
	t.Parallel()

	repo := newRepository(stubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return sqlc.AppCreatorCapability{}, nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return sqlc.AppCreatorProfile{}, pgx.ErrNoRows
		},
	})

	_, err := repo.GetProfile(context.Background(), uuid.New())
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("GetProfile() error got %v want %v", err, ErrProfileNotFound)
	}
}
