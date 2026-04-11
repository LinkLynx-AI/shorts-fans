package creator

import (
	"context"
	"errors"
	"fmt"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/google/uuid"
)

// ErrCreatorModeUnavailable は creator private workspace を利用できないことを表します。
var ErrCreatorModeUnavailable = errors.New("creator mode is not available")

// WorkspaceOverviewMetrics は creator workspace summary の overview 指標です。
type WorkspaceOverviewMetrics struct {
	GrossUnlockRevenueJpy int64
	UnlockCount           int64
	UniquePurchaserCount  int64
}

// RevisionRequestedSummary は revision requested 件数サマリーです。
type RevisionRequestedSummary struct {
	MainCount  int64
	ShortCount int64
	TotalCount int64
}

// Workspace は creator private workspace summary の read model です。
type Workspace struct {
	Creator                  Profile
	OverviewMetrics          WorkspaceOverviewMetrics
	RevisionRequestedSummary *RevisionRequestedSummary
}

func (r *Repository) getApprovedWorkspaceProfile(ctx context.Context, viewerUserID uuid.UUID) (Profile, error) {
	capability, err := r.GetCapability(ctx, viewerUserID)
	if err != nil {
		if errors.Is(err, ErrCapabilityNotFound) {
			return Profile{}, fmt.Errorf("creator workspace 取得 user=%s: %w", viewerUserID, ErrCreatorModeUnavailable)
		}

		return Profile{}, fmt.Errorf("creator workspace capability 取得 user=%s: %w", viewerUserID, err)
	}
	if capability.State != "approved" {
		return Profile{}, fmt.Errorf("creator workspace 取得 user=%s capability_state=%s: %w", viewerUserID, capability.State, ErrCreatorModeUnavailable)
	}

	profile, err := r.GetProfile(ctx, viewerUserID)
	if err != nil {
		return Profile{}, fmt.Errorf("creator workspace profile 取得 user=%s: %w", viewerUserID, err)
	}

	return profile, nil
}

// GetWorkspace は current viewer 自身の creator workspace summary を返します。
func (r *Repository) GetWorkspace(ctx context.Context, viewerUserID uuid.UUID) (Workspace, error) {
	profile, err := r.getApprovedWorkspaceProfile(ctx, viewerUserID)
	if err != nil {
		return Workspace{}, err
	}

	overviewMetricsRow, err := r.queries.GetCreatorWorkspaceOverviewMetrics(ctx, postgres.UUIDToPG(viewerUserID))
	if err != nil {
		return Workspace{}, fmt.Errorf("creator workspace overview metrics 取得 user=%s: %w", viewerUserID, err)
	}

	revisionSummaryRow, err := r.queries.GetCreatorWorkspaceRevisionRequestedSummary(ctx, postgres.UUIDToPG(viewerUserID))
	if err != nil {
		return Workspace{}, fmt.Errorf("creator workspace revision summary 取得 user=%s: %w", viewerUserID, err)
	}

	revisionSummary := buildRevisionRequestedSummary(revisionSummaryRow.MainCount, revisionSummaryRow.ShortCount)

	return Workspace{
		Creator: profile,
		OverviewMetrics: WorkspaceOverviewMetrics{
			GrossUnlockRevenueJpy: overviewMetricsRow.GrossUnlockRevenueJpy,
			UnlockCount:           overviewMetricsRow.UnlockCount,
			UniquePurchaserCount:  overviewMetricsRow.UniquePurchaserCount,
		},
		RevisionRequestedSummary: revisionSummary,
	}, nil
}

func buildRevisionRequestedSummary(mainCount int64, shortCount int64) *RevisionRequestedSummary {
	totalCount := mainCount + shortCount
	if totalCount == 0 {
		return nil
	}

	return &RevisionRequestedSummary{
		MainCount:  mainCount,
		ShortCount: shortCount,
		TotalCount: totalCount,
	}
}
