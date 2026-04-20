package submissionreview

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestApplyDecisionSuccess(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 20, 9, 30, 0, 0, time.UTC)
	intakeID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	creatorUserID := uuid.MustParse("12111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	shortID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	mainAssetID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	shortAssetID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	caption := "snapshot caption"
	tx := &txStub{}
	mainDecisionLogged := false
	shortDecisionLogged := false
	mainDecisionApplied := false
	shortDecisionApplied := false
	intakeClosed := false
	publishCalls := 0

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
		now:      func() time.Time { return now },
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getPendingSubmissionReviewIntakeByIDForUpdate: func(_ context.Context, id pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					if id != postgres.UUIDToPG(intakeID) {
						t.Fatalf("GetPendingSubmissionReviewIntakeByIDForUpdate() id got %v want %v", id, postgres.UUIDToPG(intakeID))
					}
					return sqlc.AppSubmissionReviewIntake{
						ID:                 postgres.UUIDToPG(intakeID),
						CanonicalMainID:    postgres.UUIDToPG(mainID),
						MainMediaAssetID:   postgres.UUIDToPG(mainAssetID),
						MainPriceMinor:     1800,
						OwnershipConfirmed: true,
						ConsentConfirmed:   true,
					}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(_ context.Context, id pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					if id != postgres.UUIDToPG(mainID) {
						t.Fatalf("GetSubmissionReviewMainByIDForUpdate() id got %v want %v", id, postgres.UUIDToPG(mainID))
					}
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
						ID:                   postgres.UUIDToPG(mainID),
						CreatorUserID:        postgres.UUIDToPG(creatorUserID),
						MediaAssetID:         postgres.UUIDToPG(mainAssetID),
						MediaProcessingState: mediaStateReady,
						State:                mainStatePendingReview,
						PriceMinor:           1800,
						OwnershipConfirmed:   true,
						ConsentConfirmed:     true,
					}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(_ context.Context, canonicalMainID pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					if canonicalMainID != postgres.UUIDToPG(mainID) {
						t.Fatalf("ListSubmissionReviewShortsByCanonicalMainIDForUpdate() main id got %v want %v", canonicalMainID, postgres.UUIDToPG(mainID))
					}
					return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{{
						ID:                   postgres.UUIDToPG(shortID),
						MediaAssetID:         postgres.UUIDToPG(shortAssetID),
						MediaProcessingState: mediaStateReady,
						State:                shortStatePendingReview,
						Caption:              postgres.TextToPG(&caption),
					}}, nil
				},
				listSubmissionReviewIntakeShortsByIntakeID: func(_ context.Context, id pgtype.UUID) ([]sqlc.AppSubmissionReviewIntakeShort, error) {
					if id != postgres.UUIDToPG(intakeID) {
						t.Fatalf("ListSubmissionReviewIntakeShortsByIntakeID() id got %v want %v", id, postgres.UUIDToPG(intakeID))
					}
					return []sqlc.AppSubmissionReviewIntakeShort{{
						SubmissionReviewIntakeID: postgres.UUIDToPG(intakeID),
						ShortID:                  postgres.UUIDToPG(shortID),
						MediaAssetID:             postgres.UUIDToPG(shortAssetID),
						Caption:                  postgres.TextToPG(&caption),
					}}, nil
				},
				createSubmissionReviewMainDecision: func(_ context.Context, arg sqlc.CreateSubmissionReviewMainDecisionParams) error {
					mainDecisionLogged = true
					if arg.SubmissionReviewIntakeID != postgres.UUIDToPG(intakeID) {
						t.Fatalf("CreateSubmissionReviewMainDecision() intake id got %v want %v", arg.SubmissionReviewIntakeID, postgres.UUIDToPG(intakeID))
					}
					if arg.MainID != postgres.UUIDToPG(mainID) {
						t.Fatalf("CreateSubmissionReviewMainDecision() main id got %v want %v", arg.MainID, postgres.UUIDToPG(mainID))
					}
					if arg.TargetState != mainStateApprovedForUnlock {
						t.Fatalf("CreateSubmissionReviewMainDecision() target state got %q want %q", arg.TargetState, mainStateApprovedForUnlock)
					}
					if arg.ReasonCode.Valid {
						t.Fatal("CreateSubmissionReviewMainDecision() reason code valid = true, want false")
					}
					if arg.DecisionSource != reviewDecisionSourceManual {
						t.Fatalf("CreateSubmissionReviewMainDecision() decision source got %q want %q", arg.DecisionSource, reviewDecisionSourceManual)
					}
					if got := postgres.OptionalTimeFromPG(arg.DecisionedAt); got == nil || !got.Equal(now) {
						t.Fatalf("CreateSubmissionReviewMainDecision() decisionedAt got %v want %v", got, now)
					}
					return nil
				},
				applySubmissionReviewMainDecision: func(_ context.Context, arg sqlc.ApplySubmissionReviewMainDecisionParams) (sqlc.AppMain, error) {
					mainDecisionApplied = true
					if arg.State != mainStateApprovedForUnlock {
						t.Fatalf("ApplySubmissionReviewMainDecision() state got %q want %q", arg.State, mainStateApprovedForUnlock)
					}
					if got := postgres.OptionalTextFromPG(arg.ReviewDecisionSource); got == nil || *got != reviewDecisionSourceManual {
						t.Fatalf("ApplySubmissionReviewMainDecision() source got %v want %q", got, reviewDecisionSourceManual)
					}
					if got := postgres.OptionalTimeFromPG(arg.ReviewDecisionedAt); got == nil || !got.Equal(now) {
						t.Fatalf("ApplySubmissionReviewMainDecision() decisionedAt got %v want %v", got, now)
					}
					if got := postgres.OptionalTimeFromPG(arg.ApprovedForUnlockAt); got == nil || !got.Equal(now) {
						t.Fatalf("ApplySubmissionReviewMainDecision() approvedAt got %v want %v", got, now)
					}
					return sqlc.AppMain{}, nil
				},
				createSubmissionReviewShortDecision: func(_ context.Context, arg sqlc.CreateSubmissionReviewShortDecisionParams) error {
					shortDecisionLogged = true
					if arg.ShortID != postgres.UUIDToPG(shortID) {
						t.Fatalf("CreateSubmissionReviewShortDecision() short id got %v want %v", arg.ShortID, postgres.UUIDToPG(shortID))
					}
					if arg.TargetState != shortStateApprovedForPublish {
						t.Fatalf("CreateSubmissionReviewShortDecision() target state got %q want %q", arg.TargetState, shortStateApprovedForPublish)
					}
					if arg.DecisionSource != reviewDecisionSourceManual {
						t.Fatalf("CreateSubmissionReviewShortDecision() decision source got %q want %q", arg.DecisionSource, reviewDecisionSourceManual)
					}
					if got := postgres.OptionalTimeFromPG(arg.DecisionedAt); got == nil || !got.Equal(now) {
						t.Fatalf("CreateSubmissionReviewShortDecision() decisionedAt got %v want %v", got, now)
					}
					return nil
				},
				applySubmissionReviewShortDecision: func(_ context.Context, arg sqlc.ApplySubmissionReviewShortDecisionParams) (sqlc.AppShort, error) {
					shortDecisionApplied = true
					if arg.State != shortStateApprovedForPublish {
						t.Fatalf("ApplySubmissionReviewShortDecision() state got %q want %q", arg.State, shortStateApprovedForPublish)
					}
					if got := postgres.OptionalTextFromPG(arg.ReviewDecisionSource); got == nil || *got != reviewDecisionSourceManual {
						t.Fatalf("ApplySubmissionReviewShortDecision() source got %v want %q", got, reviewDecisionSourceManual)
					}
					if got := postgres.OptionalTimeFromPG(arg.ApprovedForPublishAt); got == nil || !got.Equal(now) {
						t.Fatalf("ApplySubmissionReviewShortDecision() approvedAt got %v want %v", got, now)
					}
					if got := postgres.OptionalTimeFromPG(arg.PublishedAt); got != nil {
						t.Fatalf("ApplySubmissionReviewShortDecision() publishedAt got %v want nil", got)
					}
					return sqlc.AppShort{}, nil
				},
				markSubmissionReviewIntakeDecisionApplied: func(_ context.Context, id pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					intakeClosed = true
					if id != postgres.UUIDToPG(intakeID) {
						t.Fatalf("MarkSubmissionReviewIntakeDecisionApplied() id got %v want %v", id, postgres.UUIDToPG(intakeID))
					}
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				getSubmissionReviewCreatorUserIDByIntakeID: func(_ context.Context, id pgtype.UUID) (pgtype.UUID, error) {
					if id != postgres.UUIDToPG(intakeID) {
						t.Fatalf("GetSubmissionReviewCreatorUserIDByIntakeID() intake id got %v want %v", id, postgres.UUIDToPG(intakeID))
					}
					return postgres.UUIDToPG(creatorUserID), nil
				},
				getCreatorCapabilityByUserIDForUpdate: func(_ context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					if userID != postgres.UUIDToPG(creatorUserID) {
						t.Fatalf("GetCreatorCapabilityByUserIDForUpdate() user id got %v want %v", userID, postgres.UUIDToPG(creatorUserID))
					}
					return sqlc.AppCreatorCapability{State: capabilityStateApproved}, nil
				},
				publishShort: func(_ context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
					publishCalls++
					if id != postgres.UUIDToPG(shortID) {
						t.Fatalf("PublishShort() id got %v want %v", id, postgres.UUIDToPG(shortID))
					}
					return sqlc.AppShort{}, nil
				},
			}
		},
	}

	err := service.ApplyDecision(context.Background(), ReviewDecisionInput{
		IntakeID: intakeID,
		MainDecision: &MainReviewDecisionInput{
			Decision: reviewDecisionApproved,
		},
		ShortDecisions: []ShortReviewDecisionInput{{
			ShortID:  shortID,
			Decision: reviewDecisionApproved,
		}},
	})
	if err != nil {
		t.Fatalf("ApplyDecision() error = %v, want nil", err)
	}
	if !mainDecisionLogged || !shortDecisionLogged || !mainDecisionApplied || !shortDecisionApplied || !intakeClosed {
		t.Fatalf(
			"ApplyDecision() completed flags got mainLogged=%t shortLogged=%t mainApplied=%t shortApplied=%t intakeClosed=%t want all true",
			mainDecisionLogged,
			shortDecisionLogged,
			mainDecisionApplied,
			shortDecisionApplied,
			intakeClosed,
		)
	}
	if !tx.committed {
		t.Fatal("ApplyDecision() committed = false, want true")
	}
	if publishCalls != 0 {
		t.Fatalf("ApplyDecision() publishCalls got %d want 0", publishCalls)
	}
}

func TestApplyDecisionRejectsMissingShortCoverage(t *testing.T) {
	t.Parallel()

	intakeID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	mainID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	shortID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	mainAssetID := uuid.MustParse("12121212-1212-1212-1212-121212121212")
	shortAssetID := uuid.MustParse("13131313-1313-1313-1313-131313131313")
	tx := &txStub{}
	mutated := false

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
		now:      time.Now,
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getPendingSubmissionReviewIntakeByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{
						ID:                 postgres.UUIDToPG(intakeID),
						CanonicalMainID:    postgres.UUIDToPG(mainID),
						MainMediaAssetID:   postgres.UUIDToPG(mainAssetID),
						MainPriceMinor:     1800,
						OwnershipConfirmed: true,
						ConsentConfirmed:   true,
					}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
						ID:                   postgres.UUIDToPG(mainID),
						MediaAssetID:         postgres.UUIDToPG(mainAssetID),
						MediaProcessingState: mediaStateReady,
						State:                mainStatePendingReview,
						PriceMinor:           1800,
						OwnershipConfirmed:   true,
						ConsentConfirmed:     true,
					}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{{
						ID:                   postgres.UUIDToPG(shortID),
						MediaAssetID:         postgres.UUIDToPG(shortAssetID),
						MediaProcessingState: mediaStateReady,
						State:                shortStatePendingReview,
					}}, nil
				},
				listSubmissionReviewIntakeShortsByIntakeID: func(context.Context, pgtype.UUID) ([]sqlc.AppSubmissionReviewIntakeShort, error) {
					return []sqlc.AppSubmissionReviewIntakeShort{{
						SubmissionReviewIntakeID: postgres.UUIDToPG(intakeID),
						ShortID:                  postgres.UUIDToPG(shortID),
						MediaAssetID:             postgres.UUIDToPG(shortAssetID),
					}}, nil
				},
				createSubmissionReviewMainDecision: func(context.Context, sqlc.CreateSubmissionReviewMainDecisionParams) error {
					mutated = true
					return nil
				},
				createSubmissionReviewShortDecision: func(context.Context, sqlc.CreateSubmissionReviewShortDecisionParams) error {
					mutated = true
					return nil
				},
			}
		},
	}

	err := service.ApplyDecision(context.Background(), ReviewDecisionInput{
		IntakeID: intakeID,
		MainDecision: &MainReviewDecisionInput{
			Decision: reviewDecisionApproved,
		},
	})
	if !errors.Is(err, ErrSubmissionReviewDecisionTargetsMismatch) {
		t.Fatalf("ApplyDecision() error got %v want %v", err, ErrSubmissionReviewDecisionTargetsMismatch)
	}
	if mutated {
		t.Fatal("ApplyDecision() mutated state despite missing short coverage")
	}
	if !tx.rolledBack {
		t.Fatal("ApplyDecision() rolledBack = false, want true")
	}
}

func TestApplyDecisionRejectsReviewStateConflict(t *testing.T) {
	t.Parallel()

	intakeID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	mainID := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	shortID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	mainAssetID := uuid.MustParse("14141414-1414-1414-1414-141414141414")
	shortAssetID := uuid.MustParse("15151515-1515-1515-1515-151515151515")
	tx := &txStub{}

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
		now:      time.Now,
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getPendingSubmissionReviewIntakeByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{
						ID:                 postgres.UUIDToPG(intakeID),
						CanonicalMainID:    postgres.UUIDToPG(mainID),
						MainMediaAssetID:   postgres.UUIDToPG(mainAssetID),
						MainPriceMinor:     1800,
						OwnershipConfirmed: true,
						ConsentConfirmed:   true,
					}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
						ID:                   postgres.UUIDToPG(mainID),
						MediaAssetID:         postgres.UUIDToPG(mainAssetID),
						MediaProcessingState: mediaStateReady,
						State:                mainStateApprovedForUnlock,
						PriceMinor:           1800,
						OwnershipConfirmed:   true,
						ConsentConfirmed:     true,
					}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{{
						ID:                   postgres.UUIDToPG(shortID),
						MediaAssetID:         postgres.UUIDToPG(shortAssetID),
						MediaProcessingState: mediaStateReady,
						State:                shortStateApprovedForPublish,
					}}, nil
				},
				listSubmissionReviewIntakeShortsByIntakeID: func(context.Context, pgtype.UUID) ([]sqlc.AppSubmissionReviewIntakeShort, error) {
					return []sqlc.AppSubmissionReviewIntakeShort{{
						SubmissionReviewIntakeID: postgres.UUIDToPG(intakeID),
						ShortID:                  postgres.UUIDToPG(shortID),
						MediaAssetID:             postgres.UUIDToPG(shortAssetID),
					}}, nil
				},
			}
		},
	}

	err := service.ApplyDecision(context.Background(), ReviewDecisionInput{
		IntakeID: intakeID,
		MainDecision: &MainReviewDecisionInput{
			Decision: reviewDecisionApproved,
		},
		ShortDecisions: []ShortReviewDecisionInput{{
			ShortID:  shortID,
			Decision: reviewDecisionApproved,
		}},
	})
	if !errors.Is(err, ErrReviewStateConflict) {
		t.Fatalf("ApplyDecision() error got %v want %v", err, ErrReviewStateConflict)
	}
	if !tx.rolledBack {
		t.Fatal("ApplyDecision() rolledBack = false, want true")
	}
}

func TestApplyDecisionLeavesExistingApprovedShortsUnpublishedWhenMainBecomesUnlockable(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 20, 10, 15, 0, 0, time.UTC)
	intakeID := uuid.MustParse("10101010-1010-1010-1010-101010101010")
	creatorUserID := uuid.MustParse("10101010-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("20202020-2020-2020-2020-202020202020")
	shortID := uuid.MustParse("30303030-3030-3030-3030-303030303030")
	mainAssetID := uuid.MustParse("40404040-4040-4040-4040-404040404040")
	shortAssetID := uuid.MustParse("50505050-5050-5050-5050-505050505050")
	tx := &txStub{}
	publishCalls := 0

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
		now:      func() time.Time { return now },
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getPendingSubmissionReviewIntakeByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{
						ID:                 postgres.UUIDToPG(intakeID),
						CanonicalMainID:    postgres.UUIDToPG(mainID),
						MainMediaAssetID:   postgres.UUIDToPG(mainAssetID),
						MainPriceMinor:     1800,
						OwnershipConfirmed: true,
						ConsentConfirmed:   true,
					}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
						ID:                   postgres.UUIDToPG(mainID),
						CreatorUserID:        postgres.UUIDToPG(creatorUserID),
						MediaAssetID:         postgres.UUIDToPG(mainAssetID),
						MediaProcessingState: mediaStateReady,
						State:                mainStatePendingReview,
						PriceMinor:           1800,
						OwnershipConfirmed:   true,
						ConsentConfirmed:     true,
					}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{{
						ID:                   postgres.UUIDToPG(shortID),
						MediaAssetID:         postgres.UUIDToPG(shortAssetID),
						MediaProcessingState: mediaStateReady,
						State:                shortStateApprovedForPublish,
						ApprovedForPublishAt: postgres.TimeToPG(&now),
						PublishedAt:          pgtype.Timestamptz{},
					}}, nil
				},
				listSubmissionReviewIntakeShortsByIntakeID: func(context.Context, pgtype.UUID) ([]sqlc.AppSubmissionReviewIntakeShort, error) {
					return []sqlc.AppSubmissionReviewIntakeShort{{
						SubmissionReviewIntakeID: postgres.UUIDToPG(intakeID),
						ShortID:                  postgres.UUIDToPG(shortID),
						MediaAssetID:             postgres.UUIDToPG(shortAssetID),
					}}, nil
				},
				createSubmissionReviewMainDecision: func(context.Context, sqlc.CreateSubmissionReviewMainDecisionParams) error {
					return nil
				},
				applySubmissionReviewMainDecision: func(context.Context, sqlc.ApplySubmissionReviewMainDecisionParams) (sqlc.AppMain, error) {
					return sqlc.AppMain{}, nil
				},
				markSubmissionReviewIntakeDecisionApplied: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				getSubmissionReviewCreatorUserIDByIntakeID: func(_ context.Context, id pgtype.UUID) (pgtype.UUID, error) {
					if id != postgres.UUIDToPG(intakeID) {
						t.Fatalf("GetSubmissionReviewCreatorUserIDByIntakeID() intake id got %v want %v", id, postgres.UUIDToPG(intakeID))
					}
					return postgres.UUIDToPG(creatorUserID), nil
				},
				getCreatorCapabilityByUserIDForUpdate: func(_ context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					if userID != postgres.UUIDToPG(creatorUserID) {
						t.Fatalf("GetCreatorCapabilityByUserIDForUpdate() user id got %v want %v", userID, postgres.UUIDToPG(creatorUserID))
					}
					return sqlc.AppCreatorCapability{State: capabilityStateApproved}, nil
				},
				publishShort: func(_ context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
					publishCalls++
					if id != postgres.UUIDToPG(shortID) {
						t.Fatalf("PublishShort() id got %v want %v", id, postgres.UUIDToPG(shortID))
					}
					return sqlc.AppShort{}, nil
				},
			}
		},
	}

	err := service.ApplyDecision(context.Background(), ReviewDecisionInput{
		IntakeID: intakeID,
		MainDecision: &MainReviewDecisionInput{
			Decision: reviewDecisionApproved,
		},
	})
	if err != nil {
		t.Fatalf("ApplyDecision() error = %v, want nil", err)
	}
	if publishCalls != 0 {
		t.Fatalf("ApplyDecision() publishCalls got %d want 0", publishCalls)
	}
}

func TestApplyDecisionDoesNotPublishShortUntilMainUnlockable(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 20, 11, 0, 0, 0, time.UTC)
	intakeID := uuid.MustParse("60606060-6060-6060-6060-606060606060")
	mainID := uuid.MustParse("70707070-7070-7070-7070-707070707070")
	shortID := uuid.MustParse("80808080-8080-8080-8080-808080808080")
	mainAssetID := uuid.MustParse("90909090-9090-9090-9090-909090909090")
	shortAssetID := uuid.MustParse("a0a0a0a0-a0a0-a0a0-a0a0-a0a0a0a0a0a0")
	tx := &txStub{}
	publishCalls := 0

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
		now:      func() time.Time { return now },
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getPendingSubmissionReviewIntakeByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{
						ID:                 postgres.UUIDToPG(intakeID),
						CanonicalMainID:    postgres.UUIDToPG(mainID),
						MainMediaAssetID:   postgres.UUIDToPG(mainAssetID),
						MainPriceMinor:     1800,
						OwnershipConfirmed: true,
						ConsentConfirmed:   true,
					}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
						ID:                   postgres.UUIDToPG(mainID),
						MediaAssetID:         postgres.UUIDToPG(mainAssetID),
						MediaProcessingState: mediaStateReady,
						State:                mainStatePendingReview,
						PriceMinor:           1800,
						OwnershipConfirmed:   true,
						ConsentConfirmed:     true,
					}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{{
						ID:                   postgres.UUIDToPG(shortID),
						MediaAssetID:         postgres.UUIDToPG(shortAssetID),
						MediaProcessingState: mediaStateReady,
						State:                shortStatePendingReview,
						PublishedAt:          pgtype.Timestamptz{},
					}}, nil
				},
				listSubmissionReviewIntakeShortsByIntakeID: func(context.Context, pgtype.UUID) ([]sqlc.AppSubmissionReviewIntakeShort, error) {
					return []sqlc.AppSubmissionReviewIntakeShort{{
						SubmissionReviewIntakeID: postgres.UUIDToPG(intakeID),
						ShortID:                  postgres.UUIDToPG(shortID),
						MediaAssetID:             postgres.UUIDToPG(shortAssetID),
					}}, nil
				},
				createSubmissionReviewMainDecision: func(context.Context, sqlc.CreateSubmissionReviewMainDecisionParams) error {
					return nil
				},
				applySubmissionReviewMainDecision: func(context.Context, sqlc.ApplySubmissionReviewMainDecisionParams) (sqlc.AppMain, error) {
					return sqlc.AppMain{}, nil
				},
				createSubmissionReviewShortDecision: func(context.Context, sqlc.CreateSubmissionReviewShortDecisionParams) error {
					return nil
				},
				applySubmissionReviewShortDecision: func(_ context.Context, arg sqlc.ApplySubmissionReviewShortDecisionParams) (sqlc.AppShort, error) {
					if got := postgres.OptionalTimeFromPG(arg.ApprovedForPublishAt); got == nil || !got.Equal(now) {
						t.Fatalf("ApplySubmissionReviewShortDecision() approvedAt got %v want %v", got, now)
					}
					if got := postgres.OptionalTimeFromPG(arg.PublishedAt); got != nil {
						t.Fatalf("ApplySubmissionReviewShortDecision() publishedAt got %v want nil", got)
					}
					return sqlc.AppShort{}, nil
				},
				markSubmissionReviewIntakeDecisionApplied: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				publishShort: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
					publishCalls++
					return sqlc.AppShort{}, nil
				},
			}
		},
	}

	err := service.ApplyDecision(context.Background(), ReviewDecisionInput{
		IntakeID: intakeID,
		MainDecision: &MainReviewDecisionInput{
			Decision: reviewDecisionRevisionRequested,
			ReasonCode: func() *string {
				value := "needs_main_fix"
				return &value
			}(),
		},
		ShortDecisions: []ShortReviewDecisionInput{{
			ShortID:  shortID,
			Decision: reviewDecisionApproved,
		}},
	})
	if err != nil {
		t.Fatalf("ApplyDecision() error = %v, want nil", err)
	}
	if publishCalls != 0 {
		t.Fatalf("ApplyDecision() publishCalls got %d want 0", publishCalls)
	}
}

func TestApplyDecisionIgnoresExtraCurrentDraftShorts(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 20, 11, 30, 0, 0, time.UTC)
	intakeID := uuid.MustParse("11110000-0000-0000-0000-000000000001")
	creatorUserID := uuid.MustParse("11110000-0000-0000-0000-000000000011")
	mainID := uuid.MustParse("11110000-0000-0000-0000-000000000002")
	intakeShortID := uuid.MustParse("11110000-0000-0000-0000-000000000003")
	extraDraftShortID := uuid.MustParse("11110000-0000-0000-0000-000000000004")
	mainAssetID := uuid.MustParse("11110000-0000-0000-0000-000000000005")
	intakeShortAssetID := uuid.MustParse("11110000-0000-0000-0000-000000000006")
	extraDraftShortAssetID := uuid.MustParse("11110000-0000-0000-0000-000000000007")
	tx := &txStub{}
	publishCalls := 0

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
		now:      func() time.Time { return now },
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getPendingSubmissionReviewIntakeByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{
						ID:                 postgres.UUIDToPG(intakeID),
						CanonicalMainID:    postgres.UUIDToPG(mainID),
						MainMediaAssetID:   postgres.UUIDToPG(mainAssetID),
						MainPriceMinor:     1800,
						OwnershipConfirmed: true,
						ConsentConfirmed:   true,
					}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
						ID:                   postgres.UUIDToPG(mainID),
						CreatorUserID:        postgres.UUIDToPG(creatorUserID),
						MediaAssetID:         postgres.UUIDToPG(mainAssetID),
						MediaProcessingState: mediaStateReady,
						State:                mainStatePendingReview,
						PriceMinor:           1800,
						OwnershipConfirmed:   true,
						ConsentConfirmed:     true,
					}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{
						{
							ID:                   postgres.UUIDToPG(intakeShortID),
							MediaAssetID:         postgres.UUIDToPG(intakeShortAssetID),
							MediaProcessingState: mediaStateReady,
							State:                shortStatePendingReview,
						},
						{
							ID:                   postgres.UUIDToPG(extraDraftShortID),
							MediaAssetID:         postgres.UUIDToPG(extraDraftShortAssetID),
							MediaProcessingState: mediaStateReady,
							State:                shortStateDraft,
						},
					}, nil
				},
				listSubmissionReviewIntakeShortsByIntakeID: func(context.Context, pgtype.UUID) ([]sqlc.AppSubmissionReviewIntakeShort, error) {
					return []sqlc.AppSubmissionReviewIntakeShort{{
						SubmissionReviewIntakeID: postgres.UUIDToPG(intakeID),
						ShortID:                  postgres.UUIDToPG(intakeShortID),
						MediaAssetID:             postgres.UUIDToPG(intakeShortAssetID),
					}}, nil
				},
				createSubmissionReviewMainDecision: func(context.Context, sqlc.CreateSubmissionReviewMainDecisionParams) error {
					return nil
				},
				applySubmissionReviewMainDecision: func(context.Context, sqlc.ApplySubmissionReviewMainDecisionParams) (sqlc.AppMain, error) {
					return sqlc.AppMain{}, nil
				},
				createSubmissionReviewShortDecision: func(context.Context, sqlc.CreateSubmissionReviewShortDecisionParams) error {
					return nil
				},
				applySubmissionReviewShortDecision: func(context.Context, sqlc.ApplySubmissionReviewShortDecisionParams) (sqlc.AppShort, error) {
					return sqlc.AppShort{}, nil
				},
				markSubmissionReviewIntakeDecisionApplied: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				getSubmissionReviewCreatorUserIDByIntakeID: func(_ context.Context, id pgtype.UUID) (pgtype.UUID, error) {
					if id != postgres.UUIDToPG(intakeID) {
						t.Fatalf("GetSubmissionReviewCreatorUserIDByIntakeID() intake id got %v want %v", id, postgres.UUIDToPG(intakeID))
					}
					return postgres.UUIDToPG(creatorUserID), nil
				},
				getCreatorCapabilityByUserIDForUpdate: func(_ context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					if userID != postgres.UUIDToPG(creatorUserID) {
						t.Fatalf("GetCreatorCapabilityByUserIDForUpdate() user id got %v want %v", userID, postgres.UUIDToPG(creatorUserID))
					}
					return sqlc.AppCreatorCapability{State: capabilityStateApproved}, nil
				},
				publishShort: func(_ context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
					publishCalls++
					if id != postgres.UUIDToPG(intakeShortID) {
						t.Fatalf("PublishShort() id got %v want %v", id, postgres.UUIDToPG(intakeShortID))
					}
					return sqlc.AppShort{}, nil
				},
			}
		},
	}

	err := service.ApplyDecision(context.Background(), ReviewDecisionInput{
		IntakeID: intakeID,
		MainDecision: &MainReviewDecisionInput{
			Decision: reviewDecisionApproved,
		},
		ShortDecisions: []ShortReviewDecisionInput{{
			ShortID:  intakeShortID,
			Decision: reviewDecisionApproved,
		}},
	})
	if err != nil {
		t.Fatalf("ApplyDecision() error = %v, want nil", err)
	}
	if publishCalls != 0 {
		t.Fatalf("ApplyDecision() publishCalls got %d want 0", publishCalls)
	}
}

func TestApplyDecisionDoesNotPublishShortWhenPublicGateBlocked(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name            string
		capabilityState string
		mainPostReport  *string
		shortPostReport *string
	}

	tests := []testCase{
		{
			name:            "creator capability not approved",
			capabilityState: "pending_review",
		},
		{
			name:            "main post report blocked",
			capabilityState: capabilityStateApproved,
			mainPostReport: func() *string {
				value := "temporarily_limited"
				return &value
			}(),
		},
		{
			name:            "short post report blocked",
			capabilityState: capabilityStateApproved,
			shortPostReport: func() *string {
				value := "removed"
				return &value
			}(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			now := time.Date(2026, 4, 20, 11, 45, 0, 0, time.UTC)
			intakeID := uuid.New()
			creatorUserID := uuid.New()
			mainID := uuid.New()
			shortID := uuid.New()
			mainAssetID := uuid.New()
			shortAssetID := uuid.New()
			tx := &txStub{}
			publishCalls := 0

			service := &Service{
				beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
				now:      func() time.Time { return now },
				newQueries: func(sqlc.DBTX) queries {
					return queriesStub{
						getPendingSubmissionReviewIntakeByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
							return sqlc.AppSubmissionReviewIntake{
								ID:                 postgres.UUIDToPG(intakeID),
								CanonicalMainID:    postgres.UUIDToPG(mainID),
								MainMediaAssetID:   postgres.UUIDToPG(mainAssetID),
								MainPriceMinor:     1800,
								OwnershipConfirmed: true,
								ConsentConfirmed:   true,
							}, nil
						},
						getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
							return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
								ID:                   postgres.UUIDToPG(mainID),
								CreatorUserID:        postgres.UUIDToPG(creatorUserID),
								MediaAssetID:         postgres.UUIDToPG(mainAssetID),
								MediaProcessingState: mediaStateReady,
								State:                mainStatePendingReview,
								PostReportState:      postgres.TextToPG(tt.mainPostReport),
								PriceMinor:           1800,
								OwnershipConfirmed:   true,
								ConsentConfirmed:     true,
							}, nil
						},
						listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
							return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{{
								ID:                   postgres.UUIDToPG(shortID),
								MediaAssetID:         postgres.UUIDToPG(shortAssetID),
								MediaProcessingState: mediaStateReady,
								State:                shortStatePendingReview,
								PostReportState:      postgres.TextToPG(tt.shortPostReport),
							}}, nil
						},
						listSubmissionReviewIntakeShortsByIntakeID: func(context.Context, pgtype.UUID) ([]sqlc.AppSubmissionReviewIntakeShort, error) {
							return []sqlc.AppSubmissionReviewIntakeShort{{
								SubmissionReviewIntakeID: postgres.UUIDToPG(intakeID),
								ShortID:                  postgres.UUIDToPG(shortID),
								MediaAssetID:             postgres.UUIDToPG(shortAssetID),
							}}, nil
						},
						createSubmissionReviewMainDecision: func(context.Context, sqlc.CreateSubmissionReviewMainDecisionParams) error {
							return nil
						},
						applySubmissionReviewMainDecision: func(context.Context, sqlc.ApplySubmissionReviewMainDecisionParams) (sqlc.AppMain, error) {
							return sqlc.AppMain{}, nil
						},
						createSubmissionReviewShortDecision: func(context.Context, sqlc.CreateSubmissionReviewShortDecisionParams) error {
							return nil
						},
						applySubmissionReviewShortDecision: func(context.Context, sqlc.ApplySubmissionReviewShortDecisionParams) (sqlc.AppShort, error) {
							return sqlc.AppShort{}, nil
						},
						markSubmissionReviewIntakeDecisionApplied: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
							return sqlc.AppSubmissionReviewIntake{}, nil
						},
						getSubmissionReviewCreatorUserIDByIntakeID: func(_ context.Context, id pgtype.UUID) (pgtype.UUID, error) {
							if id != postgres.UUIDToPG(intakeID) {
								t.Fatalf("GetSubmissionReviewCreatorUserIDByIntakeID() intake id got %v want %v", id, postgres.UUIDToPG(intakeID))
							}
							return postgres.UUIDToPG(creatorUserID), nil
						},
						getCreatorCapabilityByUserIDForUpdate: func(_ context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
							if userID != postgres.UUIDToPG(creatorUserID) {
								t.Fatalf("GetCreatorCapabilityByUserIDForUpdate() user id got %v want %v", userID, postgres.UUIDToPG(creatorUserID))
							}
							return sqlc.AppCreatorCapability{State: tt.capabilityState}, nil
						},
						publishShort: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
							publishCalls++
							return sqlc.AppShort{}, nil
						},
					}
				},
			}

			err := service.ApplyDecision(context.Background(), ReviewDecisionInput{
				IntakeID: intakeID,
				MainDecision: &MainReviewDecisionInput{
					Decision: reviewDecisionApproved,
				},
				ShortDecisions: []ShortReviewDecisionInput{{
					ShortID:  shortID,
					Decision: reviewDecisionApproved,
				}},
			})
			if err != nil {
				t.Fatalf("ApplyDecision() error = %v, want nil", err)
			}
			if publishCalls != 0 {
				t.Fatalf("ApplyDecision() publishCalls got %d want 0", publishCalls)
			}
			if !tx.committed {
				t.Fatal("ApplyDecision() committed = false, want true")
			}
		})
	}
}

func TestApplyDecisionRejectsExtraCurrentNonDraftShort(t *testing.T) {
	t.Parallel()

	intakeID := uuid.MustParse("22220000-0000-0000-0000-000000000001")
	mainID := uuid.MustParse("22220000-0000-0000-0000-000000000002")
	intakeShortID := uuid.MustParse("22220000-0000-0000-0000-000000000003")
	extraApprovedShortID := uuid.MustParse("22220000-0000-0000-0000-000000000004")
	mainAssetID := uuid.MustParse("22220000-0000-0000-0000-000000000005")
	intakeShortAssetID := uuid.MustParse("22220000-0000-0000-0000-000000000006")
	extraApprovedShortAssetID := uuid.MustParse("22220000-0000-0000-0000-000000000007")
	tx := &txStub{}
	mutated := false
	approvedAt := time.Date(2026, 4, 20, 8, 0, 0, 0, time.UTC)

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
		now:      time.Now,
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getPendingSubmissionReviewIntakeByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{
						ID:                 postgres.UUIDToPG(intakeID),
						CanonicalMainID:    postgres.UUIDToPG(mainID),
						MainMediaAssetID:   postgres.UUIDToPG(mainAssetID),
						MainPriceMinor:     1800,
						OwnershipConfirmed: true,
						ConsentConfirmed:   true,
					}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
						ID:                   postgres.UUIDToPG(mainID),
						MediaAssetID:         postgres.UUIDToPG(mainAssetID),
						MediaProcessingState: mediaStateReady,
						State:                mainStatePendingReview,
						PriceMinor:           1800,
						OwnershipConfirmed:   true,
						ConsentConfirmed:     true,
					}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{
						{
							ID:                   postgres.UUIDToPG(intakeShortID),
							MediaAssetID:         postgres.UUIDToPG(intakeShortAssetID),
							MediaProcessingState: mediaStateReady,
							State:                shortStatePendingReview,
						},
						{
							ID:                   postgres.UUIDToPG(extraApprovedShortID),
							MediaAssetID:         postgres.UUIDToPG(extraApprovedShortAssetID),
							MediaProcessingState: mediaStateReady,
							State:                shortStateApprovedForPublish,
							ApprovedForPublishAt: postgres.TimeToPG(&approvedAt),
						},
					}, nil
				},
				listSubmissionReviewIntakeShortsByIntakeID: func(context.Context, pgtype.UUID) ([]sqlc.AppSubmissionReviewIntakeShort, error) {
					return []sqlc.AppSubmissionReviewIntakeShort{{
						SubmissionReviewIntakeID: postgres.UUIDToPG(intakeID),
						ShortID:                  postgres.UUIDToPG(intakeShortID),
						MediaAssetID:             postgres.UUIDToPG(intakeShortAssetID),
					}}, nil
				},
				createSubmissionReviewMainDecision: func(context.Context, sqlc.CreateSubmissionReviewMainDecisionParams) error {
					mutated = true
					return nil
				},
				createSubmissionReviewShortDecision: func(context.Context, sqlc.CreateSubmissionReviewShortDecisionParams) error {
					mutated = true
					return nil
				},
			}
		},
	}

	err := service.ApplyDecision(context.Background(), ReviewDecisionInput{
		IntakeID: intakeID,
		MainDecision: &MainReviewDecisionInput{
			Decision: reviewDecisionApproved,
		},
		ShortDecisions: []ShortReviewDecisionInput{{
			ShortID:  intakeShortID,
			Decision: reviewDecisionApproved,
		}},
	})
	if !errors.Is(err, ErrReviewStateConflict) {
		t.Fatalf("ApplyDecision() error got %v want %v", err, ErrReviewStateConflict)
	}
	if mutated {
		t.Fatal("ApplyDecision() mutated state despite extra current non-draft short")
	}
	if !tx.rolledBack {
		t.Fatal("ApplyDecision() rolledBack = false, want true")
	}
}

func TestApplyDecisionRejectsSnapshotMismatch(t *testing.T) {
	t.Parallel()

	intakeID := uuid.MustParse("aaaaaaaa-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("bbbbbbbb-2222-2222-2222-222222222222")
	shortID := uuid.MustParse("cccccccc-3333-3333-3333-333333333333")
	mainAssetID := uuid.MustParse("dddddddd-4444-4444-4444-444444444444")
	currentShortAssetID := uuid.MustParse("eeeeeeee-5555-5555-5555-555555555555")
	snapshotShortAssetID := uuid.MustParse("ffffffff-6666-6666-6666-666666666666")
	tx := &txStub{}
	mutated := false

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
		now:      time.Now,
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getPendingSubmissionReviewIntakeByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{
						ID:                 postgres.UUIDToPG(intakeID),
						CanonicalMainID:    postgres.UUIDToPG(mainID),
						MainMediaAssetID:   postgres.UUIDToPG(mainAssetID),
						MainPriceMinor:     1800,
						OwnershipConfirmed: true,
						ConsentConfirmed:   true,
					}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
						ID:                   postgres.UUIDToPG(mainID),
						MediaAssetID:         postgres.UUIDToPG(mainAssetID),
						MediaProcessingState: mediaStateReady,
						State:                mainStatePendingReview,
						PriceMinor:           1800,
						OwnershipConfirmed:   true,
						ConsentConfirmed:     true,
					}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{{
						ID:                   postgres.UUIDToPG(shortID),
						MediaAssetID:         postgres.UUIDToPG(currentShortAssetID),
						MediaProcessingState: mediaStateReady,
						State:                shortStatePendingReview,
					}}, nil
				},
				listSubmissionReviewIntakeShortsByIntakeID: func(context.Context, pgtype.UUID) ([]sqlc.AppSubmissionReviewIntakeShort, error) {
					return []sqlc.AppSubmissionReviewIntakeShort{{
						SubmissionReviewIntakeID: postgres.UUIDToPG(intakeID),
						ShortID:                  postgres.UUIDToPG(shortID),
						MediaAssetID:             postgres.UUIDToPG(snapshotShortAssetID),
					}}, nil
				},
				createSubmissionReviewMainDecision: func(context.Context, sqlc.CreateSubmissionReviewMainDecisionParams) error {
					mutated = true
					return nil
				},
				createSubmissionReviewShortDecision: func(context.Context, sqlc.CreateSubmissionReviewShortDecisionParams) error {
					mutated = true
					return nil
				},
			}
		},
	}

	err := service.ApplyDecision(context.Background(), ReviewDecisionInput{
		IntakeID: intakeID,
		MainDecision: &MainReviewDecisionInput{
			Decision: reviewDecisionApproved,
		},
		ShortDecisions: []ShortReviewDecisionInput{{
			ShortID:  shortID,
			Decision: reviewDecisionApproved,
		}},
	})
	if !errors.Is(err, ErrSubmissionReviewDecisionTargetsMismatch) {
		t.Fatalf("ApplyDecision() error got %v want %v", err, ErrSubmissionReviewDecisionTargetsMismatch)
	}
	if mutated {
		t.Fatal("ApplyDecision() mutated state despite snapshot mismatch")
	}
	if !tx.rolledBack {
		t.Fatal("ApplyDecision() rolledBack = false, want true")
	}
}

func TestApplyDecisionRejectsSnapshotReadinessMismatch(t *testing.T) {
	t.Parallel()

	intakeID := uuid.MustParse("abababab-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("bcbcbcbc-2222-2222-2222-222222222222")
	shortID := uuid.MustParse("cdcdcdcd-3333-3333-3333-333333333333")
	mainAssetID := uuid.MustParse("dededede-4444-4444-4444-444444444444")
	shortAssetID := uuid.MustParse("efefefef-5555-5555-5555-555555555555")
	tx := &txStub{}
	mutated := false

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
		now:      time.Now,
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getPendingSubmissionReviewIntakeByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{
						ID:                 postgres.UUIDToPG(intakeID),
						CanonicalMainID:    postgres.UUIDToPG(mainID),
						MainMediaAssetID:   postgres.UUIDToPG(mainAssetID),
						MainPriceMinor:     1800,
						OwnershipConfirmed: true,
						ConsentConfirmed:   true,
					}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
						ID:                   postgres.UUIDToPG(mainID),
						MediaAssetID:         postgres.UUIDToPG(mainAssetID),
						MediaProcessingState: mediaStateReady,
						State:                mainStatePendingReview,
						PriceMinor:           1800,
						OwnershipConfirmed:   true,
						ConsentConfirmed:     true,
					}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{{
						ID:                   postgres.UUIDToPG(shortID),
						MediaAssetID:         postgres.UUIDToPG(shortAssetID),
						MediaProcessingState: "processing_failed",
						State:                shortStatePendingReview,
					}}, nil
				},
				listSubmissionReviewIntakeShortsByIntakeID: func(context.Context, pgtype.UUID) ([]sqlc.AppSubmissionReviewIntakeShort, error) {
					return []sqlc.AppSubmissionReviewIntakeShort{{
						SubmissionReviewIntakeID: postgres.UUIDToPG(intakeID),
						ShortID:                  postgres.UUIDToPG(shortID),
						MediaAssetID:             postgres.UUIDToPG(shortAssetID),
					}}, nil
				},
				createSubmissionReviewMainDecision: func(context.Context, sqlc.CreateSubmissionReviewMainDecisionParams) error {
					mutated = true
					return nil
				},
				createSubmissionReviewShortDecision: func(context.Context, sqlc.CreateSubmissionReviewShortDecisionParams) error {
					mutated = true
					return nil
				},
			}
		},
	}

	err := service.ApplyDecision(context.Background(), ReviewDecisionInput{
		IntakeID: intakeID,
		MainDecision: &MainReviewDecisionInput{
			Decision: reviewDecisionApproved,
		},
		ShortDecisions: []ShortReviewDecisionInput{{
			ShortID:  shortID,
			Decision: reviewDecisionApproved,
		}},
	})
	if !errors.Is(err, ErrSubmissionReviewDecisionTargetsMismatch) {
		t.Fatalf("ApplyDecision() error got %v want %v", err, ErrSubmissionReviewDecisionTargetsMismatch)
	}
	if mutated {
		t.Fatal("ApplyDecision() mutated state despite readiness mismatch")
	}
	if !tx.rolledBack {
		t.Fatal("ApplyDecision() rolledBack = false, want true")
	}
}

func TestApplyDecisionRejectsSnapshotLinkedStateDrift(t *testing.T) {
	t.Parallel()

	mainAssetID := uuid.MustParse("11112222-3333-4444-5555-666677778888")
	shortAssetID := uuid.MustParse("9999aaaa-bbbb-cccc-dddd-eeeeffff0000")

	tests := []struct {
		name          string
		mainState     string
		shortState    string
		mainDecision  *MainReviewDecisionInput
		shortDecision []ShortReviewDecisionInput
	}{
		{
			name:       "main drifted to draft",
			mainState:  mainStateDraft,
			shortState: shortStatePendingReview,
			shortDecision: []ShortReviewDecisionInput{{
				ShortID:  uuid.MustParse("10101010-2020-3030-4040-505050505050"),
				Decision: reviewDecisionApproved,
			}},
		},
		{
			name:         "short drifted to draft",
			mainState:    mainStatePendingReview,
			shortState:   shortStateDraft,
			mainDecision: &MainReviewDecisionInput{Decision: reviewDecisionApproved},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			intakeID := uuid.New()
			mainID := uuid.New()
			shortID := uuid.MustParse("10101010-2020-3030-4040-505050505050")
			tx := &txStub{}
			mutated := false

			service := &Service{
				beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
				now:      time.Now,
				newQueries: func(sqlc.DBTX) queries {
					return queriesStub{
						getPendingSubmissionReviewIntakeByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
							return sqlc.AppSubmissionReviewIntake{
								ID:                 postgres.UUIDToPG(intakeID),
								CanonicalMainID:    postgres.UUIDToPG(mainID),
								MainMediaAssetID:   postgres.UUIDToPG(mainAssetID),
								MainPriceMinor:     1800,
								OwnershipConfirmed: true,
								ConsentConfirmed:   true,
							}, nil
						},
						getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
							return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
								ID:                   postgres.UUIDToPG(mainID),
								MediaAssetID:         postgres.UUIDToPG(mainAssetID),
								MediaProcessingState: mediaStateReady,
								State:                tt.mainState,
								PriceMinor:           1800,
								OwnershipConfirmed:   true,
								ConsentConfirmed:     true,
							}, nil
						},
						listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
							return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{{
								ID:                   postgres.UUIDToPG(shortID),
								MediaAssetID:         postgres.UUIDToPG(shortAssetID),
								MediaProcessingState: mediaStateReady,
								State:                tt.shortState,
							}}, nil
						},
						listSubmissionReviewIntakeShortsByIntakeID: func(context.Context, pgtype.UUID) ([]sqlc.AppSubmissionReviewIntakeShort, error) {
							return []sqlc.AppSubmissionReviewIntakeShort{{
								SubmissionReviewIntakeID: postgres.UUIDToPG(intakeID),
								ShortID:                  postgres.UUIDToPG(shortID),
								MediaAssetID:             postgres.UUIDToPG(shortAssetID),
							}}, nil
						},
						createSubmissionReviewMainDecision: func(context.Context, sqlc.CreateSubmissionReviewMainDecisionParams) error {
							mutated = true
							return nil
						},
						createSubmissionReviewShortDecision: func(context.Context, sqlc.CreateSubmissionReviewShortDecisionParams) error {
							mutated = true
							return nil
						},
					}
				},
			}

			err := service.ApplyDecision(context.Background(), ReviewDecisionInput{
				IntakeID:       intakeID,
				MainDecision:   tt.mainDecision,
				ShortDecisions: tt.shortDecision,
			})
			if !errors.Is(err, ErrReviewStateConflict) {
				t.Fatalf("ApplyDecision() error got %v want %v", err, ErrReviewStateConflict)
			}
			if mutated {
				t.Fatal("ApplyDecision() mutated state despite snapshot-linked state drift")
			}
			if !tx.rolledBack {
				t.Fatal("ApplyDecision() rolledBack = false, want true")
			}
		})
	}
}

func TestApplyDecisionValidation(t *testing.T) {
	t.Parallel()

	reasonCode := "needs_fix"
	tests := []struct {
		name  string
		input ReviewDecisionInput
		want  error
	}{
		{
			name: "invalid source",
			input: ReviewDecisionInput{
				IntakeID:       uuid.New(),
				DecisionSource: "robot",
			},
			want: ErrInvalidReviewDecisionSource,
		},
		{
			name: "approved with reason",
			input: ReviewDecisionInput{
				IntakeID: uuid.New(),
				MainDecision: &MainReviewDecisionInput{
					Decision:   reviewDecisionApproved,
					ReasonCode: &reasonCode,
				},
			},
			want: ErrReviewDecisionMetadataConflict,
		},
		{
			name: "revision requested without reason",
			input: ReviewDecisionInput{
				IntakeID: uuid.New(),
				MainDecision: &MainReviewDecisionInput{
					Decision: reviewDecisionRevisionRequested,
				},
			},
			want: ErrReviewDecisionReasonRequired,
		},
	}

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return &txStub{}, nil }},
		now:      time.Now,
		newQueries: func(sqlc.DBTX) queries {
			t.Fatal("newQueries() should not be called for validation failure")
			return queriesStub{}
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := service.ApplyDecision(context.Background(), tt.input)
			if !errors.Is(err, tt.want) {
				t.Fatalf("ApplyDecision() error got %v want %v", err, tt.want)
			}
		})
	}
}
