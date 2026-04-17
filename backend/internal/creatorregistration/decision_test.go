package creatorregistration

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

func TestRepositoryApplyReviewDecisionApprovesSubmittedCapability(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("15151515-1515-1515-1515-151515151515")
	tx := &repositoryTxStub{}
	submittedAt := time.Unix(1710000000, 0).UTC()
	existing := testCapability(userID, StateSubmitted)
	existing.SubmittedAt = postgres.TimeToPG(&submittedAt)
	existing.SelfServeResubmitCount = 2
	reason := "before"
	existing.RejectionReasonCode = postgres.TextToPG(&reason)
	existing.IsResubmitEligible = true
	existing.IsSupportReviewRequired = true
	lockedCapabilityLoaded := false

	repo := &Repository{
		txBeginner: repositoryTxBeginnerStub{
			begin: func(context.Context) (pgx.Tx, error) { return tx, nil },
		},
		newQueries: func(sqlc.DBTX) queries {
			return repositoryQueriesStub{
				getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
					return testUserProfile(userID), nil
				},
				getCreatorCapabilityByUserIDLocked: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					lockedCapabilityLoaded = true
					return existing, nil
				},
				getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return testCreatorProfile(userID, "approved bio"), nil
				},
				getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
					return testIntake(userID), nil
				},
				updateCreatorCapabilityState: func(_ context.Context, arg sqlc.UpdateCreatorCapabilityStateParams) (sqlc.AppCreatorCapability, error) {
					if arg.State != StateApproved {
						t.Fatalf("UpdateCreatorCapabilityState() state got %q want %q", arg.State, StateApproved)
					}
					if arg.SelfServeResubmitCount != 0 {
						t.Fatalf("UpdateCreatorCapabilityState() resubmit count got %d want 0", arg.SelfServeResubmitCount)
					}
					if arg.RejectionReasonCode.Valid {
						t.Fatalf("UpdateCreatorCapabilityState() rejection reason got %#v want invalid", arg.RejectionReasonCode)
					}
					if arg.IsResubmitEligible {
						t.Fatal("UpdateCreatorCapabilityState() isResubmitEligible = true, want false")
					}
					if arg.IsSupportReviewRequired {
						t.Fatal("UpdateCreatorCapabilityState() isSupportReviewRequired = true, want false")
					}
					if !arg.ApprovedAt.Valid {
						t.Fatal("UpdateCreatorCapabilityState() approvedAt invalid, want value")
					}
					if arg.RejectedAt.Valid {
						t.Fatal("UpdateCreatorCapabilityState() rejectedAt valid, want empty")
					}
					if arg.SuspendedAt.Valid {
						t.Fatal("UpdateCreatorCapabilityState() suspendedAt valid, want empty")
					}
					row := existing
					row.State = StateApproved
					row.SelfServeResubmitCount = 0
					row.RejectionReasonCode = pgtype.Text{}
					row.IsResubmitEligible = false
					row.IsSupportReviewRequired = false
					row.ApprovedAt = arg.ApprovedAt
					row.RejectedAt = pgtype.Timestamptz{}
					row.SuspendedAt = pgtype.Timestamptz{}
					return row, nil
				},
			}
		},
	}

	registration, err := repo.ApplyReviewDecision(context.Background(), ReviewDecisionInput{
		Decision: StateApproved,
		UserID:   userID,
	})
	if err != nil {
		t.Fatalf("ApplyReviewDecision() error = %v, want nil", err)
	}
	if registration.State != StateApproved {
		t.Fatalf("ApplyReviewDecision() state got %q want %q", registration.State, StateApproved)
	}
	if !registration.Actions.CanEnterCreatorMode {
		t.Fatal("ApplyReviewDecision() CanEnterCreatorMode = false, want true")
	}
	if registration.Review.ApprovedAt == nil {
		t.Fatal("ApplyReviewDecision() ApprovedAt = nil, want value")
	}
	if registration.Rejection != nil {
		t.Fatalf("ApplyReviewDecision() rejection got %#v want nil", registration.Rejection)
	}
	if !tx.committed {
		t.Fatal("ApplyReviewDecision() committed = false, want true")
	}
	if !lockedCapabilityLoaded {
		t.Fatal("ApplyReviewDecision() did not load capability with row lock")
	}
}

func TestRepositoryApplyReviewDecisionRejectsSubmittedCapability(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("16161616-1616-1616-1616-161616161616")
	tx := &repositoryTxStub{}
	submittedAt := time.Unix(1710000000, 0).UTC()
	existing := testCapability(userID, StateSubmitted)
	existing.SubmittedAt = postgres.TimeToPG(&submittedAt)
	existing.SelfServeResubmitCount = 1
	reason := " documents_blurry "
	lockedCapabilityLoaded := false

	repo := &Repository{
		txBeginner: repositoryTxBeginnerStub{
			begin: func(context.Context) (pgx.Tx, error) { return tx, nil },
		},
		newQueries: func(sqlc.DBTX) queries {
			return repositoryQueriesStub{
				getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
					return testUserProfile(userID), nil
				},
				getCreatorCapabilityByUserIDLocked: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					lockedCapabilityLoaded = true
					return existing, nil
				},
				getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return testCreatorProfile(userID, "rejected bio"), nil
				},
				getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
					return testIntake(userID), nil
				},
				updateCreatorCapabilityState: func(_ context.Context, arg sqlc.UpdateCreatorCapabilityStateParams) (sqlc.AppCreatorCapability, error) {
					if arg.State != StateRejected {
						t.Fatalf("UpdateCreatorCapabilityState() state got %q want %q", arg.State, StateRejected)
					}
					if !arg.RejectedAt.Valid {
						t.Fatal("UpdateCreatorCapabilityState() rejectedAt invalid, want value")
					}
					if got := arg.RejectionReasonCode.String; got != "documents_blurry" {
						t.Fatalf("UpdateCreatorCapabilityState() reason got %q want %q", got, "documents_blurry")
					}
					if !arg.RejectionReasonCode.Valid {
						t.Fatal("UpdateCreatorCapabilityState() reason invalid, want value")
					}
					if !arg.IsResubmitEligible {
						t.Fatal("UpdateCreatorCapabilityState() isResubmitEligible = false, want true")
					}
					if arg.IsSupportReviewRequired {
						t.Fatal("UpdateCreatorCapabilityState() isSupportReviewRequired = true, want false")
					}
					if arg.SelfServeResubmitCount != 1 {
						t.Fatalf("UpdateCreatorCapabilityState() resubmit count got %d want 1", arg.SelfServeResubmitCount)
					}
					row := existing
					row.State = StateRejected
					row.RejectionReasonCode = arg.RejectionReasonCode
					row.IsResubmitEligible = true
					row.IsSupportReviewRequired = false
					row.RejectedAt = arg.RejectedAt
					row.ApprovedAt = pgtype.Timestamptz{}
					row.SuspendedAt = pgtype.Timestamptz{}
					return row, nil
				},
			}
		},
	}

	registration, err := repo.ApplyReviewDecision(context.Background(), ReviewDecisionInput{
		Decision:           StateRejected,
		IsResubmitEligible: true,
		ReasonCode:         &reason,
		UserID:             userID,
	})
	if err != nil {
		t.Fatalf("ApplyReviewDecision() error = %v, want nil", err)
	}
	if registration.State != StateRejected {
		t.Fatalf("ApplyReviewDecision() state got %q want %q", registration.State, StateRejected)
	}
	if registration.Actions.CanEnterCreatorMode {
		t.Fatal("ApplyReviewDecision() CanEnterCreatorMode = true, want false")
	}
	if registration.Rejection == nil {
		t.Fatal("ApplyReviewDecision() rejection = nil, want value")
	}
	if got := *registration.Rejection.ReasonCode; got != "documents_blurry" {
		t.Fatalf("ApplyReviewDecision() reason got %q want %q", got, "documents_blurry")
	}
	if registration.Rejection.SelfServeResubmitRemain != 1 {
		t.Fatalf("ApplyReviewDecision() remain got %d want 1", registration.Rejection.SelfServeResubmitRemain)
	}
	if !tx.committed {
		t.Fatal("ApplyReviewDecision() committed = false, want true")
	}
	if !lockedCapabilityLoaded {
		t.Fatal("ApplyReviewDecision() did not load capability with row lock")
	}
}

func TestRepositoryApplyReviewDecisionSuspendsApprovedCapability(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("17171717-1717-1717-1717-171717171717")
	tx := &repositoryTxStub{}
	submittedAt := time.Unix(1710000000, 0).UTC()
	approvedAt := time.Unix(1710003600, 0).UTC()
	existing := testCapability(userID, StateApproved)
	existing.SubmittedAt = postgres.TimeToPG(&submittedAt)
	existing.ApprovedAt = postgres.TimeToPG(&approvedAt)
	lockedCapabilityLoaded := false

	repo := &Repository{
		txBeginner: repositoryTxBeginnerStub{
			begin: func(context.Context) (pgx.Tx, error) { return tx, nil },
		},
		newQueries: func(sqlc.DBTX) queries {
			return repositoryQueriesStub{
				getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
					return testUserProfile(userID), nil
				},
				getCreatorCapabilityByUserIDLocked: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					lockedCapabilityLoaded = true
					return existing, nil
				},
				getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return testCreatorProfile(userID, "approved bio"), nil
				},
				getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
					return testIntake(userID), nil
				},
				updateCreatorCapabilityState: func(_ context.Context, arg sqlc.UpdateCreatorCapabilityStateParams) (sqlc.AppCreatorCapability, error) {
					if arg.State != StateSuspended {
						t.Fatalf("UpdateCreatorCapabilityState() state got %q want %q", arg.State, StateSuspended)
					}
					if !arg.SuspendedAt.Valid {
						t.Fatal("UpdateCreatorCapabilityState() suspendedAt invalid, want value")
					}
					if !arg.ApprovedAt.Valid {
						t.Fatal("UpdateCreatorCapabilityState() approvedAt invalid, want preserved value")
					}
					gotApprovedAt, err := postgres.RequiredTimeFromPG(arg.ApprovedAt)
					if err != nil {
						t.Fatalf("RequiredTimeFromPG() error = %v, want nil", err)
					}
					if !gotApprovedAt.Equal(approvedAt) {
						t.Fatalf("UpdateCreatorCapabilityState() approvedAt got %s want %s", gotApprovedAt, approvedAt)
					}
					row := existing
					row.State = StateSuspended
					row.SuspendedAt = arg.SuspendedAt
					row.RejectionReasonCode = pgtype.Text{}
					row.IsResubmitEligible = false
					row.IsSupportReviewRequired = false
					return row, nil
				},
			}
		},
	}

	registration, err := repo.ApplyReviewDecision(context.Background(), ReviewDecisionInput{
		Decision: StateSuspended,
		UserID:   userID,
	})
	if err != nil {
		t.Fatalf("ApplyReviewDecision() error = %v, want nil", err)
	}
	if registration.State != StateSuspended {
		t.Fatalf("ApplyReviewDecision() state got %q want %q", registration.State, StateSuspended)
	}
	if registration.Actions.CanEnterCreatorMode {
		t.Fatal("ApplyReviewDecision() CanEnterCreatorMode = true, want false")
	}
	if registration.Review.ApprovedAt == nil || !registration.Review.ApprovedAt.Equal(approvedAt) {
		t.Fatalf("ApplyReviewDecision() approvedAt got %v want %s", registration.Review.ApprovedAt, approvedAt)
	}
	if registration.Review.SuspendedAt == nil {
		t.Fatal("ApplyReviewDecision() suspendedAt = nil, want value")
	}
	if !tx.committed {
		t.Fatal("ApplyReviewDecision() committed = false, want true")
	}
	if !lockedCapabilityLoaded {
		t.Fatal("ApplyReviewDecision() did not load capability with row lock")
	}
}

func TestRepositoryApplyReviewDecisionValidatesInput(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("18181818-1818-1818-1818-181818181818")
	existing := testCapability(userID, StateSubmitted)
	lockedCapabilityLoaded := false
	repo := &Repository{
		txBeginner: repositoryTxBeginnerStub{
			begin: func(context.Context) (pgx.Tx, error) { return &repositoryTxStub{}, nil },
		},
		newQueries: func(sqlc.DBTX) queries {
			return repositoryQueriesStub{
				getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
					return testUserProfile(userID), nil
				},
				getCreatorCapabilityByUserIDLocked: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					lockedCapabilityLoaded = true
					return existing, nil
				},
				getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return testCreatorProfile(userID, "bio"), nil
				},
				getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
					return testIntake(userID), nil
				},
			}
		},
	}

	if _, err := repo.ApplyReviewDecision(context.Background(), ReviewDecisionInput{
		Decision: "unknown",
		UserID:   userID,
	}); !errors.Is(err, ErrInvalidReviewDecision) {
		t.Fatalf("ApplyReviewDecision() invalid decision got %v want %v", err, ErrInvalidReviewDecision)
	}
	if !lockedCapabilityLoaded {
		t.Fatal("ApplyReviewDecision() validation did not load capability with row lock")
	}

	blankReason := "   "
	if _, err := repo.ApplyReviewDecision(context.Background(), ReviewDecisionInput{
		Decision:   StateRejected,
		ReasonCode: &blankReason,
		UserID:     userID,
	}); !errors.Is(err, ErrReviewDecisionReasonRequired) {
		t.Fatalf("ApplyReviewDecision() blank reason got %v want %v", err, ErrReviewDecisionReasonRequired)
	}

	reason := "support_only"
	if _, err := repo.ApplyReviewDecision(context.Background(), ReviewDecisionInput{
		Decision:                StateRejected,
		IsResubmitEligible:      true,
		IsSupportReviewRequired: true,
		ReasonCode:              &reason,
		UserID:                  userID,
	}); !errors.Is(err, ErrReviewDecisionMetadataConflict) {
		t.Fatalf("ApplyReviewDecision() conflicting metadata got %v want %v", err, ErrReviewDecisionMetadataConflict)
	}

	if _, err := repo.ApplyReviewDecision(context.Background(), ReviewDecisionInput{
		Decision: StateSuspended,
		UserID:   userID,
	}); !errors.Is(err, ErrRegistrationStateConflict) {
		t.Fatalf("ApplyReviewDecision() invalid transition got %v want %v", err, ErrRegistrationStateConflict)
	}
}

func TestRepositoryApplyReviewDecisionRequiresExistingCapability(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("19191919-1919-1919-1919-191919191919")
	repo := &Repository{
		txBeginner: repositoryTxBeginnerStub{
			begin: func(context.Context) (pgx.Tx, error) { return &repositoryTxStub{}, nil },
		},
		newQueries: func(sqlc.DBTX) queries {
			return repositoryQueriesStub{
				getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
					return testUserProfile(userID), nil
				},
				getCreatorCapabilityByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return sqlc.AppCreatorCapability{}, pgx.ErrNoRows
				},
			}
		},
	}

	if _, err := repo.ApplyReviewDecision(context.Background(), ReviewDecisionInput{
		Decision: StateApproved,
		UserID:   userID,
	}); !errors.Is(err, ErrRegistrationIncomplete) {
		t.Fatalf("ApplyReviewDecision() missing capability got %v want %v", err, ErrRegistrationIncomplete)
	}
}
