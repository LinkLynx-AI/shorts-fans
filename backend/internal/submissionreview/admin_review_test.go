package submissionreview

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type adminReviewQueriesStub struct {
	getCaseSummary func(context.Context, pgtype.UUID) (sqlc.GetAdminSubmissionReviewCaseSummaryByIntakeIDRow, error)
	getMain        func(context.Context, pgtype.UUID) (sqlc.GetAdminSubmissionReviewMainByIntakeIDRow, error)
	listQueue      func(context.Context) ([]sqlc.ListAdminSubmissionReviewQueueRow, error)
	listShorts     func(context.Context, pgtype.UUID) ([]sqlc.ListAdminSubmissionReviewShortsByIntakeIDRow, error)
}

func (s adminReviewQueriesStub) GetAdminSubmissionReviewCaseSummaryByIntakeID(
	ctx context.Context,
	id pgtype.UUID,
) (sqlc.GetAdminSubmissionReviewCaseSummaryByIntakeIDRow, error) {
	return s.getCaseSummary(ctx, id)
}

func (s adminReviewQueriesStub) GetAdminSubmissionReviewMainByIntakeID(
	ctx context.Context,
	id pgtype.UUID,
) (sqlc.GetAdminSubmissionReviewMainByIntakeIDRow, error) {
	return s.getMain(ctx, id)
}

func (s adminReviewQueriesStub) ListAdminSubmissionReviewQueue(ctx context.Context) ([]sqlc.ListAdminSubmissionReviewQueueRow, error) {
	return s.listQueue(ctx)
}

func (s adminReviewQueriesStub) ListAdminSubmissionReviewShortsByIntakeID(
	ctx context.Context,
	id pgtype.UUID,
) ([]sqlc.ListAdminSubmissionReviewShortsByIntakeIDRow, error) {
	return s.listShorts(ctx, id)
}

func TestAdminReviewServiceApplyDecisionDoesNotRequireCaseReload(t *testing.T) {
	t.Parallel()

	intakeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	called := false

	service := &AdminReviewService{
		applyDecision: func(_ context.Context, input ReviewDecisionInput) error {
			called = true
			if input.IntakeID != intakeID {
				t.Fatalf("ApplyDecision() intake id got %s want %s", input.IntakeID, intakeID)
			}
			return nil
		},
	}

	if err := service.ApplyDecision(context.Background(), ReviewDecisionInput{IntakeID: intakeID}); err != nil {
		t.Fatalf("ApplyDecision() error = %v, want nil", err)
	}
	if !called {
		t.Fatal("ApplyDecision() applyDecision stub was not called")
	}
}

type adminReviewSignerStub struct{}

func (adminReviewSignerStub) PresignGetObject(context.Context, string, string, time.Duration) (string, error) {
	return "https://cdn.example.com/signed.m3u8", nil
}

func TestNewAdminReviewServiceDefaultsMediaAccessTTL(t *testing.T) {
	t.Parallel()

	delivery, err := media.NewDelivery(media.DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com",
		MainPrivateBucketName: "main-private",
	}, adminReviewSignerStub{})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	service, err := NewAdminReviewService(AdminReviewServiceConfig{}, &pgxpool.Pool{}, delivery)
	if err != nil {
		t.Fatalf("NewAdminReviewService() error = %v, want nil", err)
	}
	if service.mediaAccessTTL != media.DefaultSignedURLTTL {
		t.Fatalf("NewAdminReviewService() mediaAccessTTL got %s want %s", service.mediaAccessTTL, media.DefaultSignedURLTTL)
	}
	if service.queries == nil {
		t.Fatal("NewAdminReviewService() queries = nil, want initialized service")
	}
	if service.applyDecision == nil {
		t.Fatal("NewAdminReviewService() applyDecision = nil, want initialized service")
	}
}

func TestNewAdminReviewServiceValidatesRequiredDependencies(t *testing.T) {
	t.Parallel()

	delivery, err := media.NewDelivery(media.DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com",
		MainPrivateBucketName: "main-private",
	}, adminReviewSignerStub{})
	if err != nil {
		t.Fatalf("NewDelivery() error = %v, want nil", err)
	}

	if _, err := NewAdminReviewService(AdminReviewServiceConfig{}, nil, delivery); err == nil {
		t.Fatal("NewAdminReviewService() error = nil, want pool validation error")
	}
	if _, err := NewAdminReviewService(AdminReviewServiceConfig{}, &pgxpool.Pool{}, nil); err == nil {
		t.Fatal("NewAdminReviewService() error = nil, want delivery validation error")
	}
}

func TestAdminReviewServiceListCasesMapsQueueRows(t *testing.T) {
	t.Parallel()

	intakeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	creatorUserID := uuid.MustParse("12111111-1111-1111-1111-111111111111")
	submittedAt := time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC)

	service := &AdminReviewService{
		queries: adminReviewQueriesStub{
			listQueue: func(context.Context) ([]sqlc.ListAdminSubmissionReviewQueueRow, error) {
				return []sqlc.ListAdminSubmissionReviewQueueRow{{
					IntakeID:             postgres.UUIDToPG(intakeID),
					CreatorUserID:        postgres.UUIDToPG(creatorUserID),
					AvatarUrl:            postgres.TextToPG(ptrString("https://cdn.example.com/avatar.jpg")),
					CreatorBio:           "quiet rooftop",
					DisplayName:          "Mina Rei",
					Handle:               "minarei",
					MainDecisionRequired: true,
					PendingShortCount:    1,
					ShortCount:           2,
					SubmitKind:           "resubmit",
					SubmittedAt:          postgres.TimeToPG(&submittedAt),
				}}, nil
			},
		},
	}

	items, err := service.ListCases(context.Background())
	if err != nil {
		t.Fatalf("ListCases() error = %v, want nil", err)
	}
	if len(items) != 1 {
		t.Fatalf("ListCases() len got %d want 1", len(items))
	}
	if items[0].IntakeID != intakeID {
		t.Fatalf("ListCases() intake id got %s want %s", items[0].IntakeID, intakeID)
	}
	if items[0].Creator.UserID != creatorUserID {
		t.Fatalf("ListCases() creator id got %s want %s", items[0].Creator.UserID, creatorUserID)
	}
	if items[0].Creator.AvatarURL == nil || *items[0].Creator.AvatarURL != "https://cdn.example.com/avatar.jpg" {
		t.Fatalf("ListCases() avatar got %v want signed avatar", items[0].Creator.AvatarURL)
	}
	if !items[0].MainDecisionRequired {
		t.Fatal("ListCases() MainDecisionRequired = false, want true")
	}
	if items[0].PendingShortCount != 1 || items[0].ShortCount != 2 {
		t.Fatalf("ListCases() pending/short count got %d/%d want 1/2", items[0].PendingShortCount, items[0].ShortCount)
	}
	if items[0].SubmitKind != "resubmit" {
		t.Fatalf("ListCases() submit kind got %q want %q", items[0].SubmitKind, "resubmit")
	}
	if !items[0].SubmittedAt.Equal(submittedAt) {
		t.Fatalf("ListCases() submittedAt got %v want %v", items[0].SubmittedAt, submittedAt)
	}
}

func TestAdminReviewServiceListCasesRequiresInitialization(t *testing.T) {
	t.Parallel()

	var service *AdminReviewService

	if _, err := service.ListCases(context.Background()); err == nil {
		t.Fatal("ListCases() error = nil, want initialization error")
	}
}

func TestAdminReviewServiceGetCaseDisablesDecisionFlagsForAppliedIntake(t *testing.T) {
	t.Parallel()

	intakeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	creatorUserID := uuid.MustParse("12111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainAssetID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	shortID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortAssetID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	submittedAt := time.Date(2026, 4, 20, 9, 0, 0, 0, time.UTC)
	videoAsset := media.VideoDisplayAsset{
		DurationSeconds: 18,
		ID:              shortAssetID,
		Kind:            "video",
		PosterURL:       "https://cdn.example.com/poster.jpg",
		URL:             "https://cdn.example.com/video.m3u8",
	}

	service := &AdminReviewService{
		queries: adminReviewQueriesStub{
			getCaseSummary: func(_ context.Context, id pgtype.UUID) (sqlc.GetAdminSubmissionReviewCaseSummaryByIntakeIDRow, error) {
				if id != postgres.UUIDToPG(intakeID) {
					t.Fatalf("GetAdminSubmissionReviewCaseSummaryByIntakeID() id got %v want %v", id, postgres.UUIDToPG(intakeID))
				}
				return sqlc.GetAdminSubmissionReviewCaseSummaryByIntakeIDRow{
					IntakeID:           postgres.UUIDToPG(intakeID),
					Status:             "decision_applied",
					SubmitKind:         "initial_submit",
					CanonicalMainID:    postgres.UUIDToPG(mainID),
					CreatorUserID:      postgres.UUIDToPG(creatorUserID),
					MainMediaAssetID:   postgres.UUIDToPG(mainAssetID),
					MainPriceMinor:     1800,
					OwnershipConfirmed: true,
					ConsentConfirmed:   true,
					SubmittedAt:        postgres.TimeToPG(&submittedAt),
					DisplayName:        "Mina Rei",
					Handle:             "minarei",
					CreatorBio:         "quiet rooftop",
				}, nil
			},
			getMain: func(_ context.Context, id pgtype.UUID) (sqlc.GetAdminSubmissionReviewMainByIntakeIDRow, error) {
				if id != postgres.UUIDToPG(intakeID) {
					t.Fatalf("GetAdminSubmissionReviewMainByIntakeID() id got %v want %v", id, postgres.UUIDToPG(intakeID))
				}
				return sqlc.GetAdminSubmissionReviewMainByIntakeIDRow{
					MainID:               postgres.UUIDToPG(mainID),
					MediaAssetID:         postgres.UUIDToPG(mainAssetID),
					State:                "pending_review",
					PriceMinor:           1800,
					CurrencyCode:         "JPY",
					DurationMs:           postgres.Int64ToPG(ptrInt64(18000)),
					MediaProcessingState: mediaStateReady,
				}, nil
			},
			listShorts: func(_ context.Context, id pgtype.UUID) ([]sqlc.ListAdminSubmissionReviewShortsByIntakeIDRow, error) {
				if id != postgres.UUIDToPG(intakeID) {
					t.Fatalf("ListAdminSubmissionReviewShortsByIntakeID() id got %v want %v", id, postgres.UUIDToPG(intakeID))
				}
				return []sqlc.ListAdminSubmissionReviewShortsByIntakeIDRow{{
					ShortID:              postgres.UUIDToPG(shortID),
					MediaAssetID:         postgres.UUIDToPG(shortAssetID),
					IntakeCaption:        postgres.TextToPG(ptrString("snapshot caption")),
					State:                "pending_review",
					DurationMs:           postgres.Int64ToPG(ptrInt64(18000)),
					MediaProcessingState: mediaStateReady,
				}}, nil
			},
		},
		resolveMain: func(_ context.Context, source media.MainDisplaySource, _ media.AccessBoundary, _ time.Duration) (media.VideoDisplayAsset, error) {
			if source.AssetID != mainAssetID {
				t.Fatalf("resolveMain() asset id got %s want %s", source.AssetID, mainAssetID)
			}
			return videoAsset, nil
		},
		resolveShort: func(source media.ShortDisplaySource, _ media.AccessBoundary) (media.VideoDisplayAsset, error) {
			if source.AssetID != shortAssetID {
				t.Fatalf("resolveShort() asset id got %s want %s", source.AssetID, shortAssetID)
			}
			return videoAsset, nil
		},
	}

	reviewCase, err := service.GetCase(context.Background(), intakeID)
	if err != nil {
		t.Fatalf("GetCase() error = %v, want nil", err)
	}
	if reviewCase.Main.DecisionRequired {
		t.Fatal("GetCase() main.DecisionRequired = true, want false for decision_applied intake")
	}
	if len(reviewCase.Shorts) != 1 {
		t.Fatalf("GetCase() shorts len got %d want 1", len(reviewCase.Shorts))
	}
	if reviewCase.Shorts[0].DecisionRequired {
		t.Fatal("GetCase() short.DecisionRequired = true, want false for decision_applied intake")
	}
}

func TestAdminReviewServiceGetCaseRejectsUnknownIntake(t *testing.T) {
	t.Parallel()

	service := &AdminReviewService{
		queries: adminReviewQueriesStub{
			getCaseSummary: func(context.Context, pgtype.UUID) (sqlc.GetAdminSubmissionReviewCaseSummaryByIntakeIDRow, error) {
				return sqlc.GetAdminSubmissionReviewCaseSummaryByIntakeIDRow{}, pgx.ErrNoRows
			},
		},
	}

	_, err := service.GetCase(context.Background(), uuid.MustParse("11111111-1111-1111-1111-111111111111"))
	if !errors.Is(err, ErrAdminReviewCaseNotFound) {
		t.Fatalf("GetCase() error got %v want %v", err, ErrAdminReviewCaseNotFound)
	}
}

func ptrInt64(value int64) *int64 {
	return &value
}

func ptrString(value string) *string {
	return &value
}
