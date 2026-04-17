package creatorregistration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestReviewServiceListCasesValidatesState(t *testing.T) {
	t.Parallel()

	service := &ReviewService{
		repository: newRepository(repositoryQueriesStub{}),
		signEvidenceURL: func(context.Context, string, string, time.Duration) (string, error) {
			return "", nil
		},
	}

	if _, err := service.ListCases(context.Background(), "draft"); !errors.Is(err, ErrInvalidReviewState) {
		t.Fatalf("ListCases() error got %v want %v", err, ErrInvalidReviewState)
	}
}

func TestReviewServiceListCasesReturnsQueueItems(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	submittedAt := time.Date(2026, 4, 18, 9, 0, 0, 0, time.UTC)
	avatarURL := "https://cdn.example.com/avatar.jpg"

	service := &ReviewService{
		repository: newRepository(repositoryQueriesStub{
			listCreatorRegistrationReviewCases: func(context.Context, string) ([]sqlc.ListCreatorRegistrationReviewCasesByStateRow, error) {
				return []sqlc.ListCreatorRegistrationReviewCasesByStateRow{
					{
						AvatarUrl:   postgres.TextToPG(&avatarURL),
						CreatorBio:  "quiet rooftop",
						DisplayName: "Creator Display",
						Handle:      "creator.handle",
						LegalName:   "Creator Legal",
						State:       StateSubmitted,
						SubmittedAt: postgres.TimeToPG(&submittedAt),
						UserID:      postgres.UUIDToPG(userID),
					},
				}, nil
			},
		}),
		signEvidenceURL: func(context.Context, string, string, time.Duration) (string, error) {
			return "", nil
		},
	}

	items, err := service.ListCases(context.Background(), StateSubmitted)
	if err != nil {
		t.Fatalf("ListCases() error = %v, want nil", err)
	}
	if len(items) != 1 {
		t.Fatalf("ListCases() len got %d want 1", len(items))
	}
	if items[0].UserID != userID {
		t.Fatalf("ListCases() user id got %s want %s", items[0].UserID, userID)
	}
	if items[0].LegalName != "Creator Legal" {
		t.Fatalf("ListCases() legal name got %q want %q", items[0].LegalName, "Creator Legal")
	}
	if items[0].Review.SubmittedAt == nil || !items[0].Review.SubmittedAt.Equal(submittedAt) {
		t.Fatalf("ListCases() submittedAt got %v want %s", items[0].Review.SubmittedAt, submittedAt)
	}
}

func TestReviewServiceGetCaseReturnsSignedEvidenceURLs(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	capability := testCapability(userID, StateSubmitted)
	submittedAt := time.Date(2026, 4, 18, 9, 0, 0, 0, time.UTC)
	capability.SubmittedAt = postgres.TimeToPG(&submittedAt)
	governmentURL := "https://signed.example.com/government"
	payoutURL := "https://signed.example.com/payout"

	service := &ReviewService{
		evidenceAccessTTL: 15 * time.Minute,
		repository: newRepository(repositoryQueriesStub{
			getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
				return testUserProfile(userID), nil
			},
			getCreatorCapabilityByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
				return capability, nil
			},
			getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
				return testCreatorProfile(userID, "quiet rooftop"), nil
			},
			getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
				return testIntake(userID), nil
			},
			listCreatorRegistrationEvidences: func(context.Context, pgtype.UUID) ([]sqlc.AppCreatorRegistrationEvidence, error) {
				return testEvidenceRows(userID), nil
			},
		}),
		signEvidenceURL: func(_ context.Context, bucket string, key string, _ time.Duration) (string, error) {
			switch key {
			case "creator-registration/evidence/government-id.png":
				return governmentURL, nil
			case "creator-registration/evidence/payout-proof.pdf":
				return payoutURL, nil
			default:
				return "", errors.New("unexpected key")
			}
		},
	}

	reviewCase, err := service.GetCase(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetCase() error = %v, want nil", err)
	}
	if reviewCase.State != StateSubmitted {
		t.Fatalf("GetCase() state got %q want %q", reviewCase.State, StateSubmitted)
	}
	if reviewCase.Intake.LegalName != "Creator Legal" {
		t.Fatalf("GetCase() legal name got %q want %q", reviewCase.Intake.LegalName, "Creator Legal")
	}
	if len(reviewCase.Evidences) != 2 {
		t.Fatalf("GetCase() evidence len got %d want 2", len(reviewCase.Evidences))
	}
	if reviewCase.Evidences[0].AccessURL != governmentURL {
		t.Fatalf("GetCase() government access url got %q want %q", reviewCase.Evidences[0].AccessURL, governmentURL)
	}
	if reviewCase.Evidences[1].AccessURL != payoutURL {
		t.Fatalf("GetCase() payout access url got %q want %q", reviewCase.Evidences[1].AccessURL, payoutURL)
	}
}

func TestReviewServiceGetCaseHidesDraftCases(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	service := &ReviewService{
		evidenceAccessTTL: 15 * time.Minute,
		repository: newRepository(repositoryQueriesStub{
			getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
				return testUserProfile(userID), nil
			},
			getCreatorCapabilityByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
				return testCapability(userID, StateDraft), nil
			},
			getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
				return testCreatorProfile(userID, "quiet rooftop"), nil
			},
			getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
				return testIntake(userID), nil
			},
			listCreatorRegistrationEvidences: func(context.Context, pgtype.UUID) ([]sqlc.AppCreatorRegistrationEvidence, error) {
				return testEvidenceRows(userID), nil
			},
		}),
		signEvidenceURL: func(context.Context, string, string, time.Duration) (string, error) {
			t.Fatal("signEvidenceURL() called, want draft case to be rejected before signing")
			return "", nil
		},
	}

	if _, err := service.GetCase(context.Background(), userID); !errors.Is(err, ErrReviewCaseNotFound) {
		t.Fatalf("GetCase() error got %v want %v", err, ErrReviewCaseNotFound)
	}
}

func TestReviewServiceApplyDecisionHidesDraftCases(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	reasonCode := "documents_blurry"

	service := &ReviewService{
		repository: newRepository(repositoryQueriesStub{
			getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
				return testUserProfile(userID), nil
			},
			getCreatorCapabilityByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
				return testCapability(userID, StateDraft), nil
			},
			getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
				return testCreatorProfile(userID, "quiet rooftop"), nil
			},
			getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
				return testIntake(userID), nil
			},
		}),
		signEvidenceURL: func(context.Context, string, string, time.Duration) (string, error) {
			t.Fatal("signEvidenceURL() called, want decision path to stop before detail reload")
			return "", nil
		},
	}

	_, err := service.ApplyDecision(context.Background(), ReviewDecisionInput{
		Decision:   StateRejected,
		ReasonCode: &reasonCode,
		UserID:     userID,
	})
	if !errors.Is(err, ErrReviewCaseNotFound) {
		t.Fatalf("ApplyDecision() error got %v want %v", err, ErrReviewCaseNotFound)
	}
}
