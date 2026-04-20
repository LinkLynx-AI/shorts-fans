package submissionreview

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestGetWorkspaceReviewSurfaceBuildsPackageStates(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainReadyID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainBlockedID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	mainReadyAssetID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	mainBlockedAssetID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	shortApprovedID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	shortRevisionID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	shortBlockedID := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	shortApprovedAssetID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	shortRevisionAssetID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	shortBlockedAssetID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	revisionReasonCode := "caption_needs_adjustment"

	service := &Service{
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getCreatorCapabilityByUserID: func(_ context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					if userID != postgres.UUIDToPG(viewerID) {
						t.Fatalf("GetCreatorCapabilityByUserID() userID got %v want %v", userID, postgres.UUIDToPG(viewerID))
					}
					return sqlc.AppCreatorCapability{UserID: postgres.UUIDToPG(viewerID), State: capabilityStateApproved}, nil
				},
				listMainsByCreatorUserID: func(_ context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppMain, error) {
					if creatorUserID != postgres.UUIDToPG(viewerID) {
						t.Fatalf("ListMainsByCreatorUserID() creatorUserID got %v want %v", creatorUserID, postgres.UUIDToPG(viewerID))
					}
					return []sqlc.AppMain{
						{
							ID:                 postgres.UUIDToPG(mainReadyID),
							CreatorUserID:      postgres.UUIDToPG(viewerID),
							MediaAssetID:       postgres.UUIDToPG(mainReadyAssetID),
							State:              mainStateApprovedForUnlock,
							PriceMinor:         2400,
							CurrencyCode:       "JPY",
							OwnershipConfirmed: true,
							ConsentConfirmed:   true,
						},
						{
							ID:                 postgres.UUIDToPG(mainBlockedID),
							CreatorUserID:      postgres.UUIDToPG(viewerID),
							MediaAssetID:       postgres.UUIDToPG(mainBlockedAssetID),
							State:              mainStateDraft,
							PriceMinor:         0,
							CurrencyCode:       "JPY",
							OwnershipConfirmed: true,
							ConsentConfirmed:   false,
						},
					}, nil
				},
				listShortsByCreatorUserID: func(_ context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppShort, error) {
					if creatorUserID != postgres.UUIDToPG(viewerID) {
						t.Fatalf("ListShortsByCreatorUserID() creatorUserID got %v want %v", creatorUserID, postgres.UUIDToPG(viewerID))
					}
					return []sqlc.AppShort{
						{
							ID:              postgres.UUIDToPG(shortApprovedID),
							CreatorUserID:   postgres.UUIDToPG(viewerID),
							CanonicalMainID: postgres.UUIDToPG(mainReadyID),
							MediaAssetID:    postgres.UUIDToPG(shortApprovedAssetID),
							State:           shortStateApprovedForPublish,
						},
						{
							ID:               postgres.UUIDToPG(shortRevisionID),
							CreatorUserID:    postgres.UUIDToPG(viewerID),
							CanonicalMainID:  postgres.UUIDToPG(mainReadyID),
							MediaAssetID:     postgres.UUIDToPG(shortRevisionAssetID),
							State:            shortStateRevisionRequested,
							ReviewReasonCode: postgres.TextToPG(&revisionReasonCode),
						},
						{
							ID:              postgres.UUIDToPG(shortBlockedID),
							CreatorUserID:   postgres.UUIDToPG(viewerID),
							CanonicalMainID: postgres.UUIDToPG(mainBlockedID),
							MediaAssetID:    postgres.UUIDToPG(shortBlockedAssetID),
							State:           shortStateDraft,
						},
					}, nil
				},
				listMediaAssetsByCreatorUserID: func(_ context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppMediaAsset, error) {
					if creatorUserID != postgres.UUIDToPG(viewerID) {
						t.Fatalf("ListMediaAssetsByCreatorUserID() creatorUserID got %v want %v", creatorUserID, postgres.UUIDToPG(viewerID))
					}
					return []sqlc.AppMediaAsset{
						{ID: postgres.UUIDToPG(mainReadyAssetID), ProcessingState: mediaStateReady},
						{ID: postgres.UUIDToPG(mainBlockedAssetID), ProcessingState: mediaStateReady},
						{ID: postgres.UUIDToPG(shortApprovedAssetID), ProcessingState: mediaStateReady},
						{ID: postgres.UUIDToPG(shortRevisionAssetID), ProcessingState: mediaStateReady},
						{ID: postgres.UUIDToPG(shortBlockedAssetID), ProcessingState: mediaStateReady},
					}, nil
				},
				getMediaAssetByID: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
					switch id {
					case postgres.UUIDToPG(mainReadyAssetID), postgres.UUIDToPG(mainBlockedAssetID), postgres.UUIDToPG(shortApprovedAssetID), postgres.UUIDToPG(shortRevisionAssetID), postgres.UUIDToPG(shortBlockedAssetID):
						return sqlc.AppMediaAsset{ID: id, ProcessingState: mediaStateReady}, nil
					default:
						t.Fatalf("GetMediaAssetByID() unexpected id %v", id)
						return sqlc.AppMediaAsset{}, nil
					}
				},
				getPendingSubmissionReviewIntakeByCanonicalMainID: func(_ context.Context, canonicalMainID pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
				},
				getLatestSubmissionReviewIntakeByCanonicalMainID: func(_ context.Context, canonicalMainID pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					if canonicalMainID == postgres.UUIDToPG(mainReadyID) {
						return sqlc.AppSubmissionReviewIntake{
							ID:     postgres.UUIDToPG(uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")),
							Status: intakeStatusDecisionApplied,
						}, nil
					}
					if canonicalMainID == postgres.UUIDToPG(mainBlockedID) {
						return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
					}
					t.Fatalf("GetLatestSubmissionReviewIntakeByCanonicalMainID() unexpected main %v", canonicalMainID)
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
			}
		},
	}

	got, err := service.GetWorkspaceReviewSurface(context.Background(), viewerID)
	if err != nil {
		t.Fatalf("GetWorkspaceReviewSurface() error = %v, want nil", err)
	}
	if len(got.Packages) != 2 {
		t.Fatalf("GetWorkspaceReviewSurface() packages len got %d want %d", len(got.Packages), 2)
	}
	if len(got.Mains) != 2 {
		t.Fatalf("GetWorkspaceReviewSurface() mains len got %d want %d", len(got.Mains), 2)
	}
	if len(got.Shorts) != 3 {
		t.Fatalf("GetWorkspaceReviewSurface() shorts len got %d want %d", len(got.Shorts), 3)
	}

	packageByMainID := map[uuid.UUID]WorkspaceReviewPackageSummary{}
	for _, summary := range got.Packages {
		packageByMainID[summary.CanonicalMainID] = summary
	}

	readySummary, ok := packageByMainID[mainReadyID]
	if !ok {
		t.Fatalf("GetWorkspaceReviewSurface() missing summary for main %s", mainReadyID)
	}
	if readySummary.ReviewStatus != workspaceReviewPackageStatusChangesRequested {
		t.Fatalf("readySummary.ReviewStatus got %q want %q", readySummary.ReviewStatus, workspaceReviewPackageStatusChangesRequested)
	}
	if readySummary.Readiness != workspaceReviewPackageReadinessReady {
		t.Fatalf("readySummary.Readiness got %q want %q", readySummary.Readiness, workspaceReviewPackageReadinessReady)
	}
	if readySummary.SubmitActionKind != workspaceReviewSubmitActionResubmit {
		t.Fatalf("readySummary.SubmitActionKind got %q want %q", readySummary.SubmitActionKind, workspaceReviewSubmitActionResubmit)
	}
	if readySummary.LinkedShortCount != 2 {
		t.Fatalf("readySummary.LinkedShortCount got %d want %d", readySummary.LinkedShortCount, 2)
	}

	blockedSummary, ok := packageByMainID[mainBlockedID]
	if !ok {
		t.Fatalf("GetWorkspaceReviewSurface() missing summary for main %s", mainBlockedID)
	}
	if blockedSummary.ReviewStatus != workspaceReviewPackageStatusDraft {
		t.Fatalf("blockedSummary.ReviewStatus got %q want %q", blockedSummary.ReviewStatus, workspaceReviewPackageStatusDraft)
	}
	if blockedSummary.Readiness != workspaceReviewPackageReadinessBlocked {
		t.Fatalf("blockedSummary.Readiness got %q want %q", blockedSummary.Readiness, workspaceReviewPackageReadinessBlocked)
	}
	if blockedSummary.SubmitActionKind != workspaceReviewSubmitActionNone {
		t.Fatalf("blockedSummary.SubmitActionKind got %q want %q", blockedSummary.SubmitActionKind, workspaceReviewSubmitActionNone)
	}
	if len(blockedSummary.Blockers) != 2 {
		t.Fatalf("blockedSummary.Blockers len got %d want %d", len(blockedSummary.Blockers), 2)
	}
	if blockedSummary.Blockers[0] != string(readinessBlockerMainPriceMissing) {
		t.Fatalf("blockedSummary.Blockers[0] got %q want %q", blockedSummary.Blockers[0], readinessBlockerMainPriceMissing)
	}
	if blockedSummary.Blockers[1] != string(readinessBlockerConsentMissing) {
		t.Fatalf("blockedSummary.Blockers[1] got %q want %q", blockedSummary.Blockers[1], readinessBlockerConsentMissing)
	}
}

func TestGetShortReviewSurfaceReturnsTargetPackageAndReason(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainAssetID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	shortAssetID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	reasonCode := "caption_needs_adjustment"

	service := &Service{
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getCreatorCapabilityByUserID: func(_ context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					if userID != postgres.UUIDToPG(viewerID) {
						t.Fatalf("GetCreatorCapabilityByUserID() userID got %v want %v", userID, postgres.UUIDToPG(viewerID))
					}
					return sqlc.AppCreatorCapability{UserID: postgres.UUIDToPG(viewerID), State: capabilityStateApproved}, nil
				},
				getShortByID: func(_ context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
					if id != postgres.UUIDToPG(shortID) {
						t.Fatalf("GetShortByID() id got %v want %v", id, postgres.UUIDToPG(shortID))
					}
					return sqlc.AppShort{
						ID:               postgres.UUIDToPG(shortID),
						CreatorUserID:    postgres.UUIDToPG(viewerID),
						CanonicalMainID:  postgres.UUIDToPG(mainID),
						MediaAssetID:     postgres.UUIDToPG(shortAssetID),
						State:            shortStateRevisionRequested,
						ReviewReasonCode: postgres.TextToPG(&reasonCode),
					}, nil
				},
				getMainByID: func(_ context.Context, id pgtype.UUID) (sqlc.AppMain, error) {
					if id != postgres.UUIDToPG(mainID) {
						t.Fatalf("GetMainByID() id got %v want %v", id, postgres.UUIDToPG(mainID))
					}
					return sqlc.AppMain{
						ID:                 postgres.UUIDToPG(mainID),
						CreatorUserID:      postgres.UUIDToPG(viewerID),
						MediaAssetID:       postgres.UUIDToPG(mainAssetID),
						State:              mainStateApprovedForUnlock,
						PriceMinor:         1800,
						CurrencyCode:       "JPY",
						OwnershipConfirmed: true,
						ConsentConfirmed:   true,
					}, nil
				},
				listShortsByCanonicalMainID: func(_ context.Context, canonicalMainID pgtype.UUID) ([]sqlc.AppShort, error) {
					if canonicalMainID != postgres.UUIDToPG(mainID) {
						t.Fatalf("ListShortsByCanonicalMainID() canonicalMainID got %v want %v", canonicalMainID, postgres.UUIDToPG(mainID))
					}
					return []sqlc.AppShort{
						{
							ID:               postgres.UUIDToPG(shortID),
							CreatorUserID:    postgres.UUIDToPG(viewerID),
							CanonicalMainID:  postgres.UUIDToPG(mainID),
							MediaAssetID:     postgres.UUIDToPG(shortAssetID),
							State:            shortStateRevisionRequested,
							ReviewReasonCode: postgres.TextToPG(&reasonCode),
						},
					}, nil
				},
				listMediaAssetsByCreatorUserID: func(_ context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppMediaAsset, error) {
					if creatorUserID != postgres.UUIDToPG(viewerID) {
						t.Fatalf("ListMediaAssetsByCreatorUserID() creatorUserID got %v want %v", creatorUserID, postgres.UUIDToPG(viewerID))
					}
					return []sqlc.AppMediaAsset{
						{ID: postgres.UUIDToPG(mainAssetID), ProcessingState: mediaStateReady},
						{ID: postgres.UUIDToPG(shortAssetID), ProcessingState: mediaStateReady},
					}, nil
				},
				getMediaAssetByID: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
					switch id {
					case postgres.UUIDToPG(mainAssetID), postgres.UUIDToPG(shortAssetID):
						return sqlc.AppMediaAsset{ID: id, ProcessingState: mediaStateReady}, nil
					default:
						t.Fatalf("GetMediaAssetByID() unexpected id %v", id)
						return sqlc.AppMediaAsset{}, nil
					}
				},
				getPendingSubmissionReviewIntakeByCanonicalMainID: func(_ context.Context, canonicalMainID pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					if canonicalMainID != postgres.UUIDToPG(mainID) {
						t.Fatalf("GetPendingSubmissionReviewIntakeByCanonicalMainID() main got %v want %v", canonicalMainID, postgres.UUIDToPG(mainID))
					}
					return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
				},
				getLatestSubmissionReviewIntakeByCanonicalMainID: func(_ context.Context, canonicalMainID pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					if canonicalMainID != postgres.UUIDToPG(mainID) {
						t.Fatalf("GetLatestSubmissionReviewIntakeByCanonicalMainID() main got %v want %v", canonicalMainID, postgres.UUIDToPG(mainID))
					}
					return sqlc.AppSubmissionReviewIntake{
						ID:     postgres.UUIDToPG(uuid.MustParse("66666666-6666-6666-6666-666666666666")),
						Status: intakeStatusDecisionApplied,
					}, nil
				},
			}
		},
	}

	got, err := service.GetShortReviewSurface(context.Background(), viewerID, shortID)
	if err != nil {
		t.Fatalf("GetShortReviewSurface() error = %v, want nil", err)
	}
	if got.Target.Kind != workspaceReviewTargetKindShort {
		t.Fatalf("GetShortReviewSurface() target kind got %q want %q", got.Target.Kind, workspaceReviewTargetKindShort)
	}
	if got.Target.ID != shortID {
		t.Fatalf("GetShortReviewSurface() target id got %s want %s", got.Target.ID, shortID)
	}
	if got.Target.CanonicalMainID != mainID {
		t.Fatalf("GetShortReviewSurface() target canonical main id got %s want %s", got.Target.CanonicalMainID, mainID)
	}
	if got.Review.State != shortStateRevisionRequested {
		t.Fatalf("GetShortReviewSurface() review state got %q want %q", got.Review.State, shortStateRevisionRequested)
	}
	if got.Review.ReasonCode == nil || *got.Review.ReasonCode != reasonCode {
		t.Fatalf("GetShortReviewSurface() review reason got %v want %q", got.Review.ReasonCode, reasonCode)
	}
	if got.Package.ReviewStatus != workspaceReviewPackageStatusChangesRequested {
		t.Fatalf("GetShortReviewSurface() package review status got %q want %q", got.Package.ReviewStatus, workspaceReviewPackageStatusChangesRequested)
	}
	if got.Package.SubmitActionKind != workspaceReviewSubmitActionResubmit {
		t.Fatalf("GetShortReviewSurface() package submit action got %q want %q", got.Package.SubmitActionKind, workspaceReviewSubmitActionResubmit)
	}
}

func TestGetMainReviewSurfaceReturnsMainTargetAndReason(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("10101010-1111-2222-3333-444444444444")
	mainID := uuid.MustParse("20202020-1111-2222-3333-444444444444")
	mainAssetID := uuid.MustParse("30303030-1111-2222-3333-444444444444")
	shortID := uuid.MustParse("40404040-1111-2222-3333-444444444444")
	shortAssetID := uuid.MustParse("50505050-1111-2222-3333-444444444444")
	reasonCode := "main_thumbnail_needs_adjustment"

	service := &Service{
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getCreatorCapabilityByUserID: func(_ context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					if userID != postgres.UUIDToPG(viewerID) {
						t.Fatalf("GetCreatorCapabilityByUserID() userID got %v want %v", userID, postgres.UUIDToPG(viewerID))
					}
					return sqlc.AppCreatorCapability{UserID: postgres.UUIDToPG(viewerID), State: capabilityStateApproved}, nil
				},
				getMainByID: func(_ context.Context, id pgtype.UUID) (sqlc.AppMain, error) {
					if id != postgres.UUIDToPG(mainID) {
						t.Fatalf("GetMainByID() id got %v want %v", id, postgres.UUIDToPG(mainID))
					}
					return sqlc.AppMain{
						ID:                 postgres.UUIDToPG(mainID),
						CreatorUserID:      postgres.UUIDToPG(viewerID),
						MediaAssetID:       postgres.UUIDToPG(mainAssetID),
						State:              mainStateRevisionRequested,
						PriceMinor:         1900,
						CurrencyCode:       "JPY",
						OwnershipConfirmed: true,
						ConsentConfirmed:   true,
						ReviewReasonCode:   postgres.TextToPG(&reasonCode),
					}, nil
				},
				listShortsByCanonicalMainID: func(_ context.Context, canonicalMainID pgtype.UUID) ([]sqlc.AppShort, error) {
					if canonicalMainID != postgres.UUIDToPG(mainID) {
						t.Fatalf("ListShortsByCanonicalMainID() canonicalMainID got %v want %v", canonicalMainID, postgres.UUIDToPG(mainID))
					}
					return []sqlc.AppShort{
						{
							ID:              postgres.UUIDToPG(shortID),
							CreatorUserID:   postgres.UUIDToPG(viewerID),
							CanonicalMainID: postgres.UUIDToPG(mainID),
							MediaAssetID:    postgres.UUIDToPG(shortAssetID),
							State:           shortStateApprovedForPublish,
						},
					}, nil
				},
				listMediaAssetsByCreatorUserID: func(_ context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppMediaAsset, error) {
					if creatorUserID != postgres.UUIDToPG(viewerID) {
						t.Fatalf("ListMediaAssetsByCreatorUserID() creatorUserID got %v want %v", creatorUserID, postgres.UUIDToPG(viewerID))
					}
					return []sqlc.AppMediaAsset{
						{ID: postgres.UUIDToPG(mainAssetID), ProcessingState: mediaStateReady},
						{ID: postgres.UUIDToPG(shortAssetID), ProcessingState: mediaStateReady},
					}, nil
				},
				getMediaAssetByID: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
					switch id {
					case postgres.UUIDToPG(mainAssetID), postgres.UUIDToPG(shortAssetID):
						return sqlc.AppMediaAsset{ID: id, ProcessingState: mediaStateReady}, nil
					default:
						t.Fatalf("GetMediaAssetByID() unexpected id %v", id)
						return sqlc.AppMediaAsset{}, nil
					}
				},
				getPendingSubmissionReviewIntakeByCanonicalMainID: func(_ context.Context, canonicalMainID pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					if canonicalMainID != postgres.UUIDToPG(mainID) {
						t.Fatalf("GetPendingSubmissionReviewIntakeByCanonicalMainID() main got %v want %v", canonicalMainID, postgres.UUIDToPG(mainID))
					}
					return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
				},
				getLatestSubmissionReviewIntakeByCanonicalMainID: func(_ context.Context, canonicalMainID pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					if canonicalMainID != postgres.UUIDToPG(mainID) {
						t.Fatalf("GetLatestSubmissionReviewIntakeByCanonicalMainID() main got %v want %v", canonicalMainID, postgres.UUIDToPG(mainID))
					}
					return sqlc.AppSubmissionReviewIntake{
						ID:     postgres.UUIDToPG(uuid.MustParse("60606060-1111-2222-3333-444444444444")),
						Status: intakeStatusDecisionApplied,
					}, nil
				},
			}
		},
	}

	got, err := service.GetMainReviewSurface(context.Background(), viewerID, mainID)
	if err != nil {
		t.Fatalf("GetMainReviewSurface() error = %v, want nil", err)
	}
	if got.Target.Kind != workspaceReviewTargetKindMain {
		t.Fatalf("GetMainReviewSurface() target kind got %q want %q", got.Target.Kind, workspaceReviewTargetKindMain)
	}
	if got.Target.ID != mainID {
		t.Fatalf("GetMainReviewSurface() target id got %s want %s", got.Target.ID, mainID)
	}
	if got.Target.CanonicalMainID != mainID {
		t.Fatalf("GetMainReviewSurface() target canonical main id got %s want %s", got.Target.CanonicalMainID, mainID)
	}
	if got.Review.State != mainStateRevisionRequested {
		t.Fatalf("GetMainReviewSurface() review state got %q want %q", got.Review.State, mainStateRevisionRequested)
	}
	if got.Review.ReasonCode == nil || *got.Review.ReasonCode != reasonCode {
		t.Fatalf("GetMainReviewSurface() review reason got %v want %q", got.Review.ReasonCode, reasonCode)
	}
	if got.Package.ReviewStatus != workspaceReviewPackageStatusChangesRequested {
		t.Fatalf("GetMainReviewSurface() package review status got %q want %q", got.Package.ReviewStatus, workspaceReviewPackageStatusChangesRequested)
	}
	if got.Package.SubmitActionKind != workspaceReviewSubmitActionResubmit {
		t.Fatalf("GetMainReviewSurface() package submit action got %q want %q", got.Package.SubmitActionKind, workspaceReviewSubmitActionResubmit)
	}
}

func TestGetWorkspaceReviewSurfaceReturnsErrorWhenMediaAssetLookupFails(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("61616161-1111-2222-3333-444444444444")
	mainID := uuid.MustParse("71717171-1111-2222-3333-444444444444")
	mainAssetID := uuid.MustParse("81818181-1111-2222-3333-444444444444")

	service := &Service{
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getCreatorCapabilityByUserID: func(_ context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					if userID != postgres.UUIDToPG(viewerID) {
						t.Fatalf("GetCreatorCapabilityByUserID() userID got %v want %v", userID, postgres.UUIDToPG(viewerID))
					}
					return sqlc.AppCreatorCapability{UserID: postgres.UUIDToPG(viewerID), State: capabilityStateApproved}, nil
				},
				listMainsByCreatorUserID: func(_ context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppMain, error) {
					if creatorUserID != postgres.UUIDToPG(viewerID) {
						t.Fatalf("ListMainsByCreatorUserID() creatorUserID got %v want %v", creatorUserID, postgres.UUIDToPG(viewerID))
					}
					return []sqlc.AppMain{
						{
							ID:                 postgres.UUIDToPG(mainID),
							CreatorUserID:      postgres.UUIDToPG(viewerID),
							MediaAssetID:       postgres.UUIDToPG(mainAssetID),
							State:              mainStateDraft,
							PriceMinor:         1800,
							CurrencyCode:       "JPY",
							OwnershipConfirmed: true,
							ConsentConfirmed:   true,
						},
					}, nil
				},
				listShortsByCreatorUserID: func(_ context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppShort, error) {
					if creatorUserID != postgres.UUIDToPG(viewerID) {
						t.Fatalf("ListShortsByCreatorUserID() creatorUserID got %v want %v", creatorUserID, postgres.UUIDToPG(viewerID))
					}
					return nil, nil
				},
				listMediaAssetsByCreatorUserID: func(_ context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppMediaAsset, error) {
					if creatorUserID != postgres.UUIDToPG(viewerID) {
						t.Fatalf("ListMediaAssetsByCreatorUserID() creatorUserID got %v want %v", creatorUserID, postgres.UUIDToPG(viewerID))
					}
					return nil, nil
				},
				getMediaAssetByID: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
					if id != postgres.UUIDToPG(mainAssetID) {
						t.Fatalf("GetMediaAssetByID() id got %v want %v", id, postgres.UUIDToPG(mainAssetID))
					}
					return sqlc.AppMediaAsset{}, errors.New("media lookup failed")
				},
				getPendingSubmissionReviewIntakeByCanonicalMainID: func(_ context.Context, canonicalMainID pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					if canonicalMainID != postgres.UUIDToPG(mainID) {
						t.Fatalf("GetPendingSubmissionReviewIntakeByCanonicalMainID() main got %v want %v", canonicalMainID, postgres.UUIDToPG(mainID))
					}
					return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
				},
				getLatestSubmissionReviewIntakeByCanonicalMainID: func(_ context.Context, canonicalMainID pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					if canonicalMainID != postgres.UUIDToPG(mainID) {
						t.Fatalf("GetLatestSubmissionReviewIntakeByCanonicalMainID() main got %v want %v", canonicalMainID, postgres.UUIDToPG(mainID))
					}
					return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
				},
			}
		},
	}

	_, err := service.GetWorkspaceReviewSurface(context.Background(), viewerID)
	if err == nil {
		t.Fatal("GetWorkspaceReviewSurface() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "media asset load") {
		t.Fatalf("GetWorkspaceReviewSurface() error got %q want media asset load context", err)
	}
}

func TestLoadWorkspaceReviewIntakeMapsAggregatesLatestAndPending(t *testing.T) {
	t.Parallel()

	mainAID := uuid.MustParse("91919191-1111-2222-3333-444444444444")
	mainBID := uuid.MustParse("92929292-1111-2222-3333-444444444444")
	creatorID := uuid.MustParse("93939393-1111-2222-3333-444444444444")
	latestAID := uuid.MustParse("94949494-1111-2222-3333-444444444444")
	latestBID := uuid.MustParse("95959595-1111-2222-3333-444444444444")
	pendingAID := uuid.MustParse("96969696-1111-2222-3333-444444444444")

	db := &workspaceReviewDBTXStub{
		query: func(_ context.Context, query string, args ...any) (pgx.Rows, error) {
			if len(args) != 1 {
				t.Fatalf("Query() args len got %d want %d", len(args), 1)
			}
			gotMainIDs, ok := args[0].([]uuid.UUID)
			if !ok {
				t.Fatalf("Query() mainIDs type got %T want []uuid.UUID", args[0])
			}
			if !reflect.DeepEqual(gotMainIDs, []uuid.UUID{mainAID, mainBID}) {
				t.Fatalf("Query() mainIDs got %v want %v", gotMainIDs, []uuid.UUID{mainAID, mainBID})
			}

			switch query {
			case listLatestWorkspaceReviewIntakesByCanonicalMainIDs:
				return newWorkspaceReviewRows(
					sqlc.AppSubmissionReviewIntake{
						ID:              postgres.UUIDToPG(latestAID),
						CanonicalMainID: postgres.UUIDToPG(mainAID),
						CreatorUserID:   postgres.UUIDToPG(creatorID),
						Status:          intakeStatusDecisionApplied,
						SubmitKind:      submitKindResubmit,
					},
					sqlc.AppSubmissionReviewIntake{
						ID:              postgres.UUIDToPG(latestBID),
						CanonicalMainID: postgres.UUIDToPG(mainBID),
						CreatorUserID:   postgres.UUIDToPG(creatorID),
						Status:          intakeStatusDecisionApplied,
						SubmitKind:      submitKindInitial,
					},
				), nil
			case listPendingWorkspaceReviewIntakesByCanonicalMainIDs:
				return newWorkspaceReviewRows(
					sqlc.AppSubmissionReviewIntake{
						ID:              postgres.UUIDToPG(pendingAID),
						CanonicalMainID: postgres.UUIDToPG(mainAID),
						CreatorUserID:   postgres.UUIDToPG(creatorID),
						Status:          intakeStatusPendingReview,
						SubmitKind:      submitKindResubmit,
					},
				), nil
			default:
				t.Fatalf("Query() unexpected query %q", query)
				return nil, nil
			}
		},
	}

	pendingByMainID, latestByMainID, err := loadWorkspaceReviewIntakeMaps(context.Background(), db, []sqlc.AppMain{
		{ID: postgres.UUIDToPG(mainAID)},
		{ID: postgres.UUIDToPG(mainBID)},
	})
	if err != nil {
		t.Fatalf("loadWorkspaceReviewIntakeMaps() error = %v, want nil", err)
	}
	if len(latestByMainID) != 2 {
		t.Fatalf("latestByMainID len got %d want %d", len(latestByMainID), 2)
	}
	if len(pendingByMainID) != 1 {
		t.Fatalf("pendingByMainID len got %d want %d", len(pendingByMainID), 1)
	}
	if latestByMainID[mainAID] == nil || latestByMainID[mainAID].Status != intakeStatusDecisionApplied {
		t.Fatalf("latestByMainID[mainAID] got %#v want decision_applied", latestByMainID[mainAID])
	}
	if latestByMainID[mainBID] == nil || latestByMainID[mainBID].ID != postgres.UUIDToPG(latestBID) {
		t.Fatalf("latestByMainID[mainBID] got %#v want latest intake id %v", latestByMainID[mainBID], postgres.UUIDToPG(latestBID))
	}
	if pendingByMainID[mainAID] == nil || pendingByMainID[mainAID].Status != intakeStatusPendingReview {
		t.Fatalf("pendingByMainID[mainAID] got %#v want pending_review", pendingByMainID[mainAID])
	}
	if pendingByMainID[mainBID] != nil {
		t.Fatalf("pendingByMainID[mainBID] got %#v want nil", pendingByMainID[mainBID])
	}
}

func TestResolveWorkspaceReviewPackageStatusRequiresPendingIntake(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		mainState     string
		pendingExists bool
		shortRows     []sqlc.AppShort
		want          string
	}{
		{
			name:          "short pending without intake falls back to draft",
			mainState:     mainStateDraft,
			pendingExists: false,
			shortRows: []sqlc.AppShort{
				{State: shortStatePendingReview},
			},
			want: workspaceReviewPackageStatusDraft,
		},
		{
			name:          "main pending without intake falls back to draft",
			mainState:     mainStatePendingReview,
			pendingExists: false,
			shortRows: []sqlc.AppShort{
				{State: shortStateApprovedForPublish},
			},
			want: workspaceReviewPackageStatusDraft,
		},
		{
			name:          "pending intake keeps pending review",
			mainState:     mainStatePendingReview,
			pendingExists: true,
			shortRows: []sqlc.AppShort{
				{State: shortStateApprovedForPublish},
			},
			want: workspaceReviewPackageStatusPendingReview,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := resolveWorkspaceReviewPackageStatus(tt.mainState, tt.shortRows, tt.pendingExists)
			if got != tt.want {
				t.Fatalf("resolveWorkspaceReviewPackageStatus() got %q want %q", got, tt.want)
			}
		})
	}
}

func TestLoadWorkspaceReviewAssetStatesByRowsLoadsExactAssetIDs(t *testing.T) {
	t.Parallel()

	mainAssetID := uuid.MustParse("a1a1a1a1-1111-2222-3333-444444444444")
	shortAssetID := uuid.MustParse("b2b2b2b2-1111-2222-3333-444444444444")

	db := &workspaceReviewDBTXStub{
		query: func(_ context.Context, query string, args ...any) (pgx.Rows, error) {
			if query != listWorkspaceReviewMediaAssetsByIDs {
				t.Fatalf("Query() query got %q want %q", query, listWorkspaceReviewMediaAssetsByIDs)
			}
			if len(args) != 1 {
				t.Fatalf("Query() args len got %d want %d", len(args), 1)
			}

			gotAssetIDs, ok := args[0].([]uuid.UUID)
			if !ok {
				t.Fatalf("Query() assetIDs type got %T want []uuid.UUID", args[0])
			}
			if !reflect.DeepEqual(gotAssetIDs, []uuid.UUID{mainAssetID, shortAssetID}) {
				t.Fatalf("Query() assetIDs got %v want %v", gotAssetIDs, []uuid.UUID{mainAssetID, shortAssetID})
			}

			return newWorkspaceReviewMediaAssetRows(
				sqlc.AppMediaAsset{ID: postgres.UUIDToPG(mainAssetID), ProcessingState: mediaStateReady},
				sqlc.AppMediaAsset{ID: postgres.UUIDToPG(shortAssetID), ProcessingState: "processing"},
			), nil
		},
	}

	assetStates, err := loadWorkspaceReviewAssetStatesByRows(
		context.Background(),
		db,
		[]sqlc.AppMain{
			{MediaAssetID: postgres.UUIDToPG(mainAssetID)},
		},
		[]sqlc.AppShort{
			{MediaAssetID: postgres.UUIDToPG(shortAssetID)},
			{MediaAssetID: postgres.UUIDToPG(shortAssetID)},
		},
	)
	if err != nil {
		t.Fatalf("loadWorkspaceReviewAssetStatesByRows() error = %v, want nil", err)
	}

	if !reflect.DeepEqual(assetStates, map[uuid.UUID]string{
		mainAssetID:  mediaStateReady,
		shortAssetID: "processing",
	}) {
		t.Fatalf("loadWorkspaceReviewAssetStatesByRows() got %v want %v", assetStates, map[uuid.UUID]string{
			mainAssetID:  mediaStateReady,
			shortAssetID: "processing",
		})
	}
}

func TestGetShortReviewSurfaceUsesPreloadedAssetStatesWhenDBAvailable(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("c3c3c3c3-1111-2222-3333-444444444444")
	mainID := uuid.MustParse("d4d4d4d4-1111-2222-3333-444444444444")
	mainAssetID := uuid.MustParse("e5e5e5e5-1111-2222-3333-444444444444")
	shortID := uuid.MustParse("f6f6f6f6-1111-2222-3333-444444444444")
	shortAssetID := uuid.MustParse("07070707-1111-2222-3333-444444444444")
	reasonCode := "caption_needs_adjustment"

	db := &workspaceReviewDBTXStub{
		query: func(_ context.Context, query string, args ...any) (pgx.Rows, error) {
			if query != listWorkspaceReviewMediaAssetsByIDs {
				t.Fatalf("Query() query got %q want %q", query, listWorkspaceReviewMediaAssetsByIDs)
			}
			if len(args) != 1 {
				t.Fatalf("Query() args len got %d want %d", len(args), 1)
			}

			gotAssetIDs, ok := args[0].([]uuid.UUID)
			if !ok {
				t.Fatalf("Query() assetIDs type got %T want []uuid.UUID", args[0])
			}
			if !reflect.DeepEqual(gotAssetIDs, []uuid.UUID{mainAssetID, shortAssetID}) {
				t.Fatalf("Query() assetIDs got %v want %v", gotAssetIDs, []uuid.UUID{mainAssetID, shortAssetID})
			}

			return newWorkspaceReviewMediaAssetRows(
				sqlc.AppMediaAsset{ID: postgres.UUIDToPG(mainAssetID), ProcessingState: mediaStateReady},
				sqlc.AppMediaAsset{ID: postgres.UUIDToPG(shortAssetID), ProcessingState: mediaStateReady},
			), nil
		},
	}

	service := &Service{
		beginner: db,
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getCreatorCapabilityByUserID: func(_ context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					if userID != postgres.UUIDToPG(viewerID) {
						t.Fatalf("GetCreatorCapabilityByUserID() userID got %v want %v", userID, postgres.UUIDToPG(viewerID))
					}
					return sqlc.AppCreatorCapability{UserID: postgres.UUIDToPG(viewerID), State: capabilityStateApproved}, nil
				},
				getShortByID: func(_ context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
					if id != postgres.UUIDToPG(shortID) {
						t.Fatalf("GetShortByID() id got %v want %v", id, postgres.UUIDToPG(shortID))
					}
					return sqlc.AppShort{
						ID:               postgres.UUIDToPG(shortID),
						CreatorUserID:    postgres.UUIDToPG(viewerID),
						CanonicalMainID:  postgres.UUIDToPG(mainID),
						MediaAssetID:     postgres.UUIDToPG(shortAssetID),
						State:            shortStateRevisionRequested,
						ReviewReasonCode: postgres.TextToPG(&reasonCode),
					}, nil
				},
				getMainByID: func(_ context.Context, id pgtype.UUID) (sqlc.AppMain, error) {
					if id != postgres.UUIDToPG(mainID) {
						t.Fatalf("GetMainByID() id got %v want %v", id, postgres.UUIDToPG(mainID))
					}
					return sqlc.AppMain{
						ID:                 postgres.UUIDToPG(mainID),
						CreatorUserID:      postgres.UUIDToPG(viewerID),
						MediaAssetID:       postgres.UUIDToPG(mainAssetID),
						State:              mainStateApprovedForUnlock,
						PriceMinor:         1800,
						CurrencyCode:       "JPY",
						OwnershipConfirmed: true,
						ConsentConfirmed:   true,
					}, nil
				},
				listShortsByCanonicalMainID: func(_ context.Context, canonicalMainID pgtype.UUID) ([]sqlc.AppShort, error) {
					if canonicalMainID != postgres.UUIDToPG(mainID) {
						t.Fatalf("ListShortsByCanonicalMainID() canonicalMainID got %v want %v", canonicalMainID, postgres.UUIDToPG(mainID))
					}
					return []sqlc.AppShort{
						{
							ID:               postgres.UUIDToPG(shortID),
							CreatorUserID:    postgres.UUIDToPG(viewerID),
							CanonicalMainID:  postgres.UUIDToPG(mainID),
							MediaAssetID:     postgres.UUIDToPG(shortAssetID),
							State:            shortStateRevisionRequested,
							ReviewReasonCode: postgres.TextToPG(&reasonCode),
						},
					}, nil
				},
				getMediaAssetByID: func(context.Context, pgtype.UUID) (sqlc.AppMediaAsset, error) {
					t.Fatal("GetMediaAssetByID() should not be called when asset states are preloaded")
					return sqlc.AppMediaAsset{}, nil
				},
				getPendingSubmissionReviewIntakeByCanonicalMainID: func(_ context.Context, canonicalMainID pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					if canonicalMainID != postgres.UUIDToPG(mainID) {
						t.Fatalf("GetPendingSubmissionReviewIntakeByCanonicalMainID() main got %v want %v", canonicalMainID, postgres.UUIDToPG(mainID))
					}
					return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
				},
				getLatestSubmissionReviewIntakeByCanonicalMainID: func(_ context.Context, canonicalMainID pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					if canonicalMainID != postgres.UUIDToPG(mainID) {
						t.Fatalf("GetLatestSubmissionReviewIntakeByCanonicalMainID() main got %v want %v", canonicalMainID, postgres.UUIDToPG(mainID))
					}
					return sqlc.AppSubmissionReviewIntake{
						ID:     postgres.UUIDToPG(uuid.MustParse("18181818-1111-2222-3333-444444444444")),
						Status: intakeStatusDecisionApplied,
					}, nil
				},
			}
		},
	}

	got, err := service.GetShortReviewSurface(context.Background(), viewerID, shortID)
	if err != nil {
		t.Fatalf("GetShortReviewSurface() error = %v, want nil", err)
	}
	if got.Package.SubmitActionKind != workspaceReviewSubmitActionResubmit {
		t.Fatalf("GetShortReviewSurface() package submit action got %q want %q", got.Package.SubmitActionKind, workspaceReviewSubmitActionResubmit)
	}
}

type workspaceReviewDBTXStub struct {
	begin func(context.Context) (pgx.Tx, error)
	query func(context.Context, string, ...any) (pgx.Rows, error)
}

func (s *workspaceReviewDBTXStub) Begin(ctx context.Context) (pgx.Tx, error) {
	if s.begin != nil {
		return s.begin(ctx)
	}
	return nil, nil
}

func (s *workspaceReviewDBTXStub) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (s *workspaceReviewDBTXStub) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return s.query(ctx, sql, args...)
}

func (s *workspaceReviewDBTXStub) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	return nil
}

type workspaceReviewRows struct {
	index   int
	intakes []sqlc.AppSubmissionReviewIntake
}

func newWorkspaceReviewRows(intakes ...sqlc.AppSubmissionReviewIntake) *workspaceReviewRows {
	return &workspaceReviewRows{
		index:   -1,
		intakes: intakes,
	}
}

func (r *workspaceReviewRows) Close() {}

func (r *workspaceReviewRows) Err() error { return nil }

func (r *workspaceReviewRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }

func (r *workspaceReviewRows) FieldDescriptions() []pgconn.FieldDescription { return nil }

func (r *workspaceReviewRows) Next() bool {
	r.index++
	return r.index < len(r.intakes)
}

func (r *workspaceReviewRows) Scan(dest ...any) error {
	current := r.intakes[r.index]
	values := []any{
		current.ID,
		current.CanonicalMainID,
		current.CreatorUserID,
		current.Status,
		current.SubmitKind,
		current.PreviousIntakeID,
		current.MainMediaAssetID,
		current.MainPriceMinor,
		current.OwnershipConfirmed,
		current.ConsentConfirmed,
		current.SubmittedAt,
		current.CreatedAt,
		current.UpdatedAt,
	}

	for index, destValue := range dest {
		reflect.ValueOf(destValue).Elem().Set(reflect.ValueOf(values[index]))
	}

	return nil
}

func (r *workspaceReviewRows) Values() ([]any, error) { return nil, nil }

func (r *workspaceReviewRows) RawValues() [][]byte { return nil }

func (r *workspaceReviewRows) Conn() *pgx.Conn { return nil }

type workspaceReviewMediaAssetRows struct {
	assets []sqlc.AppMediaAsset
	index  int
}

func newWorkspaceReviewMediaAssetRows(assets ...sqlc.AppMediaAsset) *workspaceReviewMediaAssetRows {
	return &workspaceReviewMediaAssetRows{
		assets: assets,
		index:  -1,
	}
}

func (r *workspaceReviewMediaAssetRows) Close() {}

func (r *workspaceReviewMediaAssetRows) Err() error { return nil }

func (r *workspaceReviewMediaAssetRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }

func (r *workspaceReviewMediaAssetRows) FieldDescriptions() []pgconn.FieldDescription { return nil }

func (r *workspaceReviewMediaAssetRows) Next() bool {
	r.index++
	return r.index < len(r.assets)
}

func (r *workspaceReviewMediaAssetRows) Scan(dest ...any) error {
	current := r.assets[r.index]
	values := []any{
		current.ID,
		current.ProcessingState,
	}

	for index, destValue := range dest {
		reflect.ValueOf(destValue).Elem().Set(reflect.ValueOf(values[index]))
	}

	return nil
}

func (r *workspaceReviewMediaAssetRows) Values() ([]any, error) { return nil, nil }

func (r *workspaceReviewMediaAssetRows) RawValues() [][]byte { return nil }

func (r *workspaceReviewMediaAssetRows) Conn() *pgx.Conn { return nil }
