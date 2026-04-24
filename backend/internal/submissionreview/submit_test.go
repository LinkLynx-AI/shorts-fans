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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type queriesStub struct {
	applySubmissionReviewMainDecision                    func(context.Context, sqlc.ApplySubmissionReviewMainDecisionParams) (sqlc.AppMain, error)
	applySubmissionReviewShortDecision                   func(context.Context, sqlc.ApplySubmissionReviewShortDecisionParams) (sqlc.AppShort, error)
	createSubmissionReviewMainDecision                   func(context.Context, sqlc.CreateSubmissionReviewMainDecisionParams) error
	createSubmissionReviewIntake                         func(context.Context, sqlc.CreateSubmissionReviewIntakeParams) (sqlc.AppSubmissionReviewIntake, error)
	createSubmissionReviewIntakeShort                    func(context.Context, sqlc.CreateSubmissionReviewIntakeShortParams) error
	createSubmissionReviewShortDecision                  func(context.Context, sqlc.CreateSubmissionReviewShortDecisionParams) error
	getCreatorCapabilityByUserID                         func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error)
	getCreatorCapabilityByUserIDForUpdate                func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error)
	getMainByID                                          func(context.Context, pgtype.UUID) (sqlc.AppMain, error)
	getMediaAssetByID                                    func(context.Context, pgtype.UUID) (sqlc.AppMediaAsset, error)
	getLatestSubmissionReviewIntakeByCanonicalMainID     func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error)
	getPendingSubmissionReviewIntakeByCanonicalMainID    func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error)
	getPendingSubmissionReviewIntakeByIDForUpdate        func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error)
	getShortByID                                         func(context.Context, pgtype.UUID) (sqlc.AppShort, error)
	getSubmissionReviewCreatorUserIDByIntakeID           func(context.Context, pgtype.UUID) (pgtype.UUID, error)
	getSubmissionReviewMainByIDForUpdate                 func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error)
	listMediaAssetsByCreatorUserID                       func(context.Context, pgtype.UUID) ([]sqlc.AppMediaAsset, error)
	listMainsByCreatorUserID                             func(context.Context, pgtype.UUID) ([]sqlc.AppMain, error)
	listSubmissionReviewIntakeShortsByIntakeID           func(context.Context, pgtype.UUID) ([]sqlc.AppSubmissionReviewIntakeShort, error)
	listShortsByCanonicalMainID                          func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error)
	listShortsByCreatorUserID                            func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error)
	listSubmissionReviewShortsByCanonicalMainIDForUpdate func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error)
	markSubmissionReviewIntakeDecisionApplied            func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error)
	publishShort                                         func(context.Context, pgtype.UUID) (sqlc.AppShort, error)
	resetSubmissionReviewMainToPending                   func(context.Context, pgtype.UUID) (sqlc.AppMain, error)
	resetSubmissionReviewShortToPending                  func(context.Context, pgtype.UUID) (sqlc.AppShort, error)
	updateMainState                                      func(context.Context, sqlc.UpdateMainStateParams) (sqlc.AppMain, error)
	updateShortState                                     func(context.Context, sqlc.UpdateShortStateParams) (sqlc.AppShort, error)
}

func unexpectedSubmitQuery(name string) {
	panic("unexpected call: " + name)
}

func (s queriesStub) ApplySubmissionReviewMainDecision(ctx context.Context, arg sqlc.ApplySubmissionReviewMainDecisionParams) (sqlc.AppMain, error) {
	if s.applySubmissionReviewMainDecision == nil {
		unexpectedSubmitQuery("ApplySubmissionReviewMainDecision")
	}
	return s.applySubmissionReviewMainDecision(ctx, arg)
}

func (s queriesStub) ApplySubmissionReviewShortDecision(ctx context.Context, arg sqlc.ApplySubmissionReviewShortDecisionParams) (sqlc.AppShort, error) {
	if s.applySubmissionReviewShortDecision == nil {
		unexpectedSubmitQuery("ApplySubmissionReviewShortDecision")
	}
	return s.applySubmissionReviewShortDecision(ctx, arg)
}

func (s queriesStub) CreateSubmissionReviewMainDecision(ctx context.Context, arg sqlc.CreateSubmissionReviewMainDecisionParams) error {
	if s.createSubmissionReviewMainDecision == nil {
		unexpectedSubmitQuery("CreateSubmissionReviewMainDecision")
	}
	return s.createSubmissionReviewMainDecision(ctx, arg)
}

func (s queriesStub) CreateSubmissionReviewIntake(ctx context.Context, arg sqlc.CreateSubmissionReviewIntakeParams) (sqlc.AppSubmissionReviewIntake, error) {
	return s.createSubmissionReviewIntake(ctx, arg)
}

func (s queriesStub) CreateSubmissionReviewIntakeShort(ctx context.Context, arg sqlc.CreateSubmissionReviewIntakeShortParams) error {
	return s.createSubmissionReviewIntakeShort(ctx, arg)
}

func (s queriesStub) CreateSubmissionReviewShortDecision(ctx context.Context, arg sqlc.CreateSubmissionReviewShortDecisionParams) error {
	if s.createSubmissionReviewShortDecision == nil {
		unexpectedSubmitQuery("CreateSubmissionReviewShortDecision")
	}
	return s.createSubmissionReviewShortDecision(ctx, arg)
}

func (s queriesStub) GetCreatorCapabilityByUserIDForUpdate(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
	if s.getCreatorCapabilityByUserIDForUpdate == nil {
		return sqlc.AppCreatorCapability{State: capabilityStateApproved}, nil
	}
	return s.getCreatorCapabilityByUserIDForUpdate(ctx, userID)
}

func (s queriesStub) GetCreatorCapabilityByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
	if s.getCreatorCapabilityByUserID == nil {
		return sqlc.AppCreatorCapability{State: capabilityStateApproved}, nil
	}
	return s.getCreatorCapabilityByUserID(ctx, userID)
}

func (s queriesStub) GetMainByID(ctx context.Context, id pgtype.UUID) (sqlc.AppMain, error) {
	if s.getMainByID == nil {
		unexpectedSubmitQuery("GetMainByID")
	}
	return s.getMainByID(ctx, id)
}

func (s queriesStub) GetMediaAssetByID(ctx context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
	if s.getMediaAssetByID == nil {
		unexpectedSubmitQuery("GetMediaAssetByID")
	}
	return s.getMediaAssetByID(ctx, id)
}

func (s queriesStub) GetLatestSubmissionReviewIntakeByCanonicalMainID(ctx context.Context, canonicalMainID pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
	return s.getLatestSubmissionReviewIntakeByCanonicalMainID(ctx, canonicalMainID)
}

func (s queriesStub) GetPendingSubmissionReviewIntakeByCanonicalMainID(ctx context.Context, canonicalMainID pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
	return s.getPendingSubmissionReviewIntakeByCanonicalMainID(ctx, canonicalMainID)
}

func (s queriesStub) GetPendingSubmissionReviewIntakeByIDForUpdate(ctx context.Context, id pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
	if s.getPendingSubmissionReviewIntakeByIDForUpdate == nil {
		unexpectedSubmitQuery("GetPendingSubmissionReviewIntakeByIDForUpdate")
	}
	return s.getPendingSubmissionReviewIntakeByIDForUpdate(ctx, id)
}

func (s queriesStub) GetShortByID(ctx context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
	if s.getShortByID == nil {
		unexpectedSubmitQuery("GetShortByID")
	}
	return s.getShortByID(ctx, id)
}

func (s queriesStub) GetSubmissionReviewCreatorUserIDByIntakeID(ctx context.Context, id pgtype.UUID) (pgtype.UUID, error) {
	if s.getSubmissionReviewCreatorUserIDByIntakeID == nil {
		return pgtype.UUID{}, nil
	}
	return s.getSubmissionReviewCreatorUserIDByIntakeID(ctx, id)
}

func (s queriesStub) GetSubmissionReviewMainByIDForUpdate(ctx context.Context, id pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
	return s.getSubmissionReviewMainByIDForUpdate(ctx, id)
}

func (s queriesStub) ListMainsByCreatorUserID(ctx context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppMain, error) {
	if s.listMainsByCreatorUserID == nil {
		unexpectedSubmitQuery("ListMainsByCreatorUserID")
	}
	return s.listMainsByCreatorUserID(ctx, creatorUserID)
}

func (s queriesStub) ListMediaAssetsByCreatorUserID(ctx context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppMediaAsset, error) {
	if s.listMediaAssetsByCreatorUserID == nil {
		unexpectedSubmitQuery("ListMediaAssetsByCreatorUserID")
	}
	return s.listMediaAssetsByCreatorUserID(ctx, creatorUserID)
}

func (s queriesStub) ListSubmissionReviewIntakeShortsByIntakeID(ctx context.Context, submissionReviewIntakeID pgtype.UUID) ([]sqlc.AppSubmissionReviewIntakeShort, error) {
	if s.listSubmissionReviewIntakeShortsByIntakeID == nil {
		unexpectedSubmitQuery("ListSubmissionReviewIntakeShortsByIntakeID")
	}
	return s.listSubmissionReviewIntakeShortsByIntakeID(ctx, submissionReviewIntakeID)
}

func (s queriesStub) ListShortsByCreatorUserID(ctx context.Context, creatorUserID pgtype.UUID) ([]sqlc.AppShort, error) {
	if s.listShortsByCreatorUserID == nil {
		unexpectedSubmitQuery("ListShortsByCreatorUserID")
	}
	return s.listShortsByCreatorUserID(ctx, creatorUserID)
}

func (s queriesStub) ListShortsByCanonicalMainID(ctx context.Context, canonicalMainID pgtype.UUID) ([]sqlc.AppShort, error) {
	if s.listShortsByCanonicalMainID == nil {
		unexpectedSubmitQuery("ListShortsByCanonicalMainID")
	}
	return s.listShortsByCanonicalMainID(ctx, canonicalMainID)
}

func (s queriesStub) ListSubmissionReviewShortsByCanonicalMainIDForUpdate(ctx context.Context, canonicalMainID pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
	return s.listSubmissionReviewShortsByCanonicalMainIDForUpdate(ctx, canonicalMainID)
}

func (s queriesStub) MarkSubmissionReviewIntakeDecisionApplied(ctx context.Context, id pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
	if s.markSubmissionReviewIntakeDecisionApplied == nil {
		unexpectedSubmitQuery("MarkSubmissionReviewIntakeDecisionApplied")
	}
	return s.markSubmissionReviewIntakeDecisionApplied(ctx, id)
}

func (s queriesStub) PublishShort(ctx context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
	if s.publishShort == nil {
		unexpectedSubmitQuery("PublishShort")
	}
	return s.publishShort(ctx, id)
}

func (s queriesStub) ResetSubmissionReviewMainToPending(ctx context.Context, id pgtype.UUID) (sqlc.AppMain, error) {
	if s.resetSubmissionReviewMainToPending == nil {
		unexpectedSubmitQuery("ResetSubmissionReviewMainToPending")
	}
	return s.resetSubmissionReviewMainToPending(ctx, id)
}

func (s queriesStub) ResetSubmissionReviewShortToPending(ctx context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
	if s.resetSubmissionReviewShortToPending == nil {
		unexpectedSubmitQuery("ResetSubmissionReviewShortToPending")
	}
	return s.resetSubmissionReviewShortToPending(ctx, id)
}

type txBeginnerStub struct {
	begin func(context.Context) (pgx.Tx, error)
}

func (s txBeginnerStub) Begin(ctx context.Context) (pgx.Tx, error) {
	return s.begin(ctx)
}

type txStub struct {
	commitErr   error
	rollbackErr error
	committed   bool
	rolledBack  bool
}

func (tx *txStub) Begin(context.Context) (pgx.Tx, error) { return tx, nil }
func (tx *txStub) Commit(context.Context) error {
	tx.committed = true
	return tx.commitErr
}
func (tx *txStub) Rollback(context.Context) error {
	tx.rolledBack = true
	return tx.rollbackErr
}
func (tx *txStub) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (tx *txStub) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (tx *txStub) LargeObjects() pgx.LargeObjects                         { return pgx.LargeObjects{} }
func (tx *txStub) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (tx *txStub) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (tx *txStub) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }
func (tx *txStub) QueryRow(context.Context, string, ...any) pgx.Row        { return nil }
func (tx *txStub) Conn() *pgx.Conn                                         { return nil }

func TestNewService(t *testing.T) {
	t.Parallel()

	service := NewService(nil)
	if service == nil {
		t.Fatal("NewService() = nil, want non-nil")
	}
	if service.newQueries == nil {
		t.Fatal("NewService() newQueries = nil, want non-nil")
	}
}

func TestSubmitPackageInitialSuccess(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 18, 10, 0, 0, 0, time.UTC)
	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainAssetID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	shortAssetID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	intakeID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	shortCaption := "preview"

	tx := &txStub{}
	createShortSnapshots := 0
	updateShorts := 0

	service := &Service{
		beginner: txBeginnerStub{
			begin: func(context.Context) (pgx.Tx, error) {
				return tx, nil
			},
		},
		now: func() time.Time { return now },
		newQueries: func(db sqlc.DBTX) queries {
			if db != tx {
				t.Fatalf("newQueries() db got %v want %v", db, tx)
			}
			return queriesStub{
				getCreatorCapabilityByUserIDForUpdate: func(_ context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					if userID != postgres.UUIDToPG(viewerID) {
						t.Fatalf("GetCreatorCapabilityByUserIDForUpdate() user id got %v want %v", userID, postgres.UUIDToPG(viewerID))
					}
					return sqlc.AppCreatorCapability{UserID: postgres.UUIDToPG(viewerID), State: capabilityStateApproved}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(_ context.Context, id pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					if id != postgres.UUIDToPG(mainID) {
						t.Fatalf("GetSubmissionReviewMainByIDForUpdate() main id got %v want %v", id, postgres.UUIDToPG(mainID))
					}
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
						ID:                   postgres.UUIDToPG(mainID),
						CreatorUserID:        postgres.UUIDToPG(viewerID),
						MediaAssetID:         postgres.UUIDToPG(mainAssetID),
						State:                mainStateDraft,
						PriceMinor:           1800,
						CurrencyCode:         "JPY",
						OwnershipConfirmed:   true,
						ConsentConfirmed:     true,
						MediaProcessingState: mediaStateReady,
					}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(_ context.Context, canonicalMainID pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					if canonicalMainID != postgres.UUIDToPG(mainID) {
						t.Fatalf("ListSubmissionReviewShortsByCanonicalMainIDForUpdate() main id got %v want %v", canonicalMainID, postgres.UUIDToPG(mainID))
					}
					return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{{
						ID:                   postgres.UUIDToPG(shortID),
						CreatorUserID:        postgres.UUIDToPG(viewerID),
						CanonicalMainID:      postgres.UUIDToPG(mainID),
						MediaAssetID:         postgres.UUIDToPG(shortAssetID),
						State:                shortStateDraft,
						Caption:              postgres.TextToPG(&shortCaption),
						MediaProcessingState: mediaStateReady,
					}}, nil
				},
				getPendingSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
				},
				getLatestSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
				},
				createSubmissionReviewIntake: func(_ context.Context, arg sqlc.CreateSubmissionReviewIntakeParams) (sqlc.AppSubmissionReviewIntake, error) {
					if arg.SubmitKind != submitKindInitial {
						t.Fatalf("CreateSubmissionReviewIntake() submit kind got %q want %q", arg.SubmitKind, submitKindInitial)
					}
					if arg.Status != intakeStatusPendingReview {
						t.Fatalf("CreateSubmissionReviewIntake() status got %q want %q", arg.Status, intakeStatusPendingReview)
					}
					if arg.PreviousIntakeID.Valid {
						t.Fatal("CreateSubmissionReviewIntake() previous intake id valid = true, want false")
					}
					if arg.SubmittedAt != postgres.TimeToPG(&now) {
						t.Fatalf("CreateSubmissionReviewIntake() submittedAt got %#v want %#v", arg.SubmittedAt, postgres.TimeToPG(&now))
					}
					return sqlc.AppSubmissionReviewIntake{ID: postgres.UUIDToPG(intakeID)}, nil
				},
				createSubmissionReviewIntakeShort: func(_ context.Context, arg sqlc.CreateSubmissionReviewIntakeShortParams) error {
					createShortSnapshots++
					if arg.SubmissionReviewIntakeID != postgres.UUIDToPG(intakeID) {
						t.Fatalf("CreateSubmissionReviewIntakeShort() intake id got %v want %v", arg.SubmissionReviewIntakeID, postgres.UUIDToPG(intakeID))
					}
					if arg.ShortID != postgres.UUIDToPG(shortID) {
						t.Fatalf("CreateSubmissionReviewIntakeShort() short id got %v want %v", arg.ShortID, postgres.UUIDToPG(shortID))
					}
					return nil
				},
				resetSubmissionReviewMainToPending: func(_ context.Context, id pgtype.UUID) (sqlc.AppMain, error) {
					if id != postgres.UUIDToPG(mainID) {
						t.Fatalf("ResetSubmissionReviewMainToPending() main id got %v want %v", id, postgres.UUIDToPG(mainID))
					}
					return sqlc.AppMain{ID: postgres.UUIDToPG(mainID), State: mainStatePendingReview}, nil
				},
				resetSubmissionReviewShortToPending: func(_ context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
					updateShorts++
					if id != postgres.UUIDToPG(shortID) {
						t.Fatalf("ResetSubmissionReviewShortToPending() short id got %v want %v", id, postgres.UUIDToPG(shortID))
					}
					return sqlc.AppShort{ID: postgres.UUIDToPG(shortID), State: shortStatePendingReview}, nil
				},
			}
		},
	}

	if err := service.SubmitPackage(context.Background(), viewerID, mainID); err != nil {
		t.Fatalf("SubmitPackage() error = %v, want nil", err)
	}
	if createShortSnapshots != 1 {
		t.Fatalf("SubmitPackage() short snapshots got %d want 1", createShortSnapshots)
	}
	if updateShorts != 1 {
		t.Fatalf("SubmitPackage() updated shorts got %d want 1", updateShorts)
	}
	if !tx.committed {
		t.Fatal("SubmitPackage() committed = false, want true")
	}
}

func TestSubmitPackageReturnsNoOpForPendingIntake(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	tx := &txStub{}
	createCalled := false
	listShortsCalled := false

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
		now:      time.Now,
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getCreatorCapabilityByUserIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return sqlc.AppCreatorCapability{State: capabilityStateApproved}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
						ID:            postgres.UUIDToPG(mainID),
						CreatorUserID: postgres.UUIDToPG(viewerID),
					}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					listShortsCalled = true
					return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{}, nil
				},
				getPendingSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{Status: intakeStatusPendingReview}, nil
				},
				getLatestSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
				},
				createSubmissionReviewIntake: func(context.Context, sqlc.CreateSubmissionReviewIntakeParams) (sqlc.AppSubmissionReviewIntake, error) {
					createCalled = true
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				createSubmissionReviewIntakeShort: func(context.Context, sqlc.CreateSubmissionReviewIntakeShortParams) error {
					createCalled = true
					return nil
				},
				resetSubmissionReviewMainToPending: func(context.Context, pgtype.UUID) (sqlc.AppMain, error) {
					createCalled = true
					return sqlc.AppMain{}, nil
				},
				resetSubmissionReviewShortToPending: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
					createCalled = true
					return sqlc.AppShort{}, nil
				},
			}
		},
	}

	if err := service.SubmitPackage(context.Background(), viewerID, mainID); err != nil {
		t.Fatalf("SubmitPackage() error = %v, want nil", err)
	}
	if createCalled {
		t.Fatal("SubmitPackage() created or updated state despite pending intake")
	}
	if listShortsCalled {
		t.Fatal("SubmitPackage() listed shorts despite pending intake")
	}
}

func TestSubmitPackageReturnsNoOpForConcurrentDuplicateSubmit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		constraintName string
	}{
		{
			name:           "pending main constraint",
			constraintName: submissionReviewPendingMainUniqueConstraint,
		},
		{
			name:           "previous intake constraint",
			constraintName: submissionReviewPreviousIntakeUniqueConstraint,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			viewerID := uuid.MustParse("91919191-9191-9191-9191-919191919191")
			mainID := uuid.MustParse("a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1")
			mainAssetID := uuid.MustParse("b1b1b1b1-b1b1-b1b1-b1b1-b1b1b1b1b1b1")
			shortID := uuid.MustParse("c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1")
			shortAssetID := uuid.MustParse("d1d1d1d1-d1d1-d1d1-d1d1-d1d1d1d1d1d1")
			tx := &txStub{}
			mutatedAfterInsert := false

			service := &Service{
				beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
				now:      time.Now,
				newQueries: func(sqlc.DBTX) queries {
					return queriesStub{
						getCreatorCapabilityByUserIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
							return sqlc.AppCreatorCapability{State: capabilityStateApproved}, nil
						},
						getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
							return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
								ID:                   postgres.UUIDToPG(mainID),
								CreatorUserID:        postgres.UUIDToPG(viewerID),
								MediaAssetID:         postgres.UUIDToPG(mainAssetID),
								State:                mainStateDraft,
								PriceMinor:           1800,
								CurrencyCode:         "JPY",
								OwnershipConfirmed:   true,
								ConsentConfirmed:     true,
								MediaProcessingState: mediaStateReady,
							}, nil
						},
						getPendingSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
							return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
						},
						listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
							return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{{
								ID:                   postgres.UUIDToPG(shortID),
								CreatorUserID:        postgres.UUIDToPG(viewerID),
								CanonicalMainID:      postgres.UUIDToPG(mainID),
								MediaAssetID:         postgres.UUIDToPG(shortAssetID),
								State:                shortStateDraft,
								MediaProcessingState: mediaStateReady,
							}}, nil
						},
						getLatestSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
							return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
						},
						createSubmissionReviewIntake: func(context.Context, sqlc.CreateSubmissionReviewIntakeParams) (sqlc.AppSubmissionReviewIntake, error) {
							return sqlc.AppSubmissionReviewIntake{}, &pgconn.PgError{
								Code:           "23505",
								ConstraintName: tt.constraintName,
							}
						},
						createSubmissionReviewIntakeShort: func(context.Context, sqlc.CreateSubmissionReviewIntakeShortParams) error {
							mutatedAfterInsert = true
							return nil
						},
						resetSubmissionReviewMainToPending: func(context.Context, pgtype.UUID) (sqlc.AppMain, error) {
							mutatedAfterInsert = true
							return sqlc.AppMain{}, nil
						},
						resetSubmissionReviewShortToPending: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
							mutatedAfterInsert = true
							return sqlc.AppShort{}, nil
						},
					}
				},
			}

			if err := service.SubmitPackage(context.Background(), viewerID, mainID); err != nil {
				t.Fatalf("SubmitPackage() error = %v, want nil", err)
			}
			if !tx.rolledBack {
				t.Fatal("SubmitPackage() rolledBack = false, want true")
			}
			if tx.committed {
				t.Fatal("SubmitPackage() committed = true, want false")
			}
			if mutatedAfterInsert {
				t.Fatal("SubmitPackage() mutated state after duplicate submit race")
			}
		})
	}
}

func TestSubmitPackageReturnsNotReadyError(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	tx := &txStub{}

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
		now:      time.Now,
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getCreatorCapabilityByUserIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return sqlc.AppCreatorCapability{State: capabilityStateApproved}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
						ID:                   postgres.UUIDToPG(mainID),
						CreatorUserID:        postgres.UUIDToPG(viewerID),
						State:                mainStateDraft,
						PriceMinor:           0,
						MediaProcessingState: "processing",
					}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					return nil, nil
				},
				getPendingSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
				},
				getLatestSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
				},
				createSubmissionReviewIntake: func(context.Context, sqlc.CreateSubmissionReviewIntakeParams) (sqlc.AppSubmissionReviewIntake, error) {
					t.Fatal("CreateSubmissionReviewIntake() called for not-ready package")
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				createSubmissionReviewIntakeShort: func(context.Context, sqlc.CreateSubmissionReviewIntakeShortParams) error {
					t.Fatal("CreateSubmissionReviewIntakeShort() called for not-ready package")
					return nil
				},
				resetSubmissionReviewMainToPending: func(context.Context, pgtype.UUID) (sqlc.AppMain, error) {
					t.Fatal("ResetSubmissionReviewMainToPending() called for not-ready package")
					return sqlc.AppMain{}, nil
				},
				resetSubmissionReviewShortToPending: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
					t.Fatal("ResetSubmissionReviewShortToPending() called for not-ready package")
					return sqlc.AppShort{}, nil
				},
			}
		},
	}

	err := service.SubmitPackage(context.Background(), viewerID, mainID)
	var notReadyErr *NotReadyError
	if !errors.As(err, &notReadyErr) {
		t.Fatalf("SubmitPackage() error got %v want NotReadyError", err)
	}
	wantBlockers := []string{
		"linked_short_missing",
		"main_asset_not_ready",
		"main_price_missing",
		"ownership_missing",
		"consent_missing",
	}
	if len(notReadyErr.Blockers) != len(wantBlockers) {
		t.Fatalf("SubmitPackage() blockers got %#v want %#v", notReadyErr.Blockers, wantBlockers)
	}
	for i, want := range wantBlockers {
		if notReadyErr.Blockers[i] != want {
			t.Fatalf("SubmitPackage() blocker[%d] got %q want %q", i, notReadyErr.Blockers[i], want)
		}
	}
}

func TestSubmitPackageIfReadyIgnoresNotReadyPackage(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return &txStub{}, nil }},
		now:      time.Now,
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getCreatorCapabilityByUserIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return sqlc.AppCreatorCapability{State: capabilityStateApproved}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
						ID:                   postgres.UUIDToPG(mainID),
						CreatorUserID:        postgres.UUIDToPG(viewerID),
						State:                mainStateDraft,
						PriceMinor:           1800,
						OwnershipConfirmed:   true,
						ConsentConfirmed:     true,
						MediaProcessingState: "processing",
					}, nil
				},
				getPendingSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{{
						State:                shortStateDraft,
						MediaProcessingState: mediaStateReady,
					}}, nil
				},
				createSubmissionReviewIntake: func(context.Context, sqlc.CreateSubmissionReviewIntakeParams) (sqlc.AppSubmissionReviewIntake, error) {
					t.Fatal("CreateSubmissionReviewIntake() called for not-ready package")
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				createSubmissionReviewIntakeShort: func(context.Context, sqlc.CreateSubmissionReviewIntakeShortParams) error {
					t.Fatal("CreateSubmissionReviewIntakeShort() called for not-ready package")
					return nil
				},
				resetSubmissionReviewMainToPending: func(context.Context, pgtype.UUID) (sqlc.AppMain, error) {
					t.Fatal("ResetSubmissionReviewMainToPending() called for not-ready package")
					return sqlc.AppMain{}, nil
				},
				resetSubmissionReviewShortToPending: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
					t.Fatal("ResetSubmissionReviewShortToPending() called for not-ready package")
					return sqlc.AppShort{}, nil
				},
			}
		},
	}

	if err := service.SubmitPackageIfReady(context.Background(), viewerID, mainID); err != nil {
		t.Fatalf("SubmitPackageIfReady() error = %v, want nil", err)
	}
}

func TestSubmitPackageIfReadyPropagatesUnexpectedError(t *testing.T) {
	t.Parallel()

	beginErr := errors.New("begin failed")
	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) {
			return nil, beginErr
		}},
		now: time.Now,
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{}
		},
	}

	err := service.SubmitPackageIfReady(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, beginErr) {
		t.Fatalf("SubmitPackageIfReady() error got %v want %v", err, beginErr)
	}
}

func TestSubmitInitialPackageIfReadyIgnoresWorkerOnlyNoOpStates(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("91919191-aaaa-aaaa-aaaa-919191919191")
	ownerID := uuid.MustParse("92929292-aaaa-aaaa-aaaa-929292929292")
	mainID := uuid.MustParse("93939393-aaaa-aaaa-aaaa-939393939393")

	tests := []struct {
		name  string
		build func(t *testing.T) queriesStub
	}{
		{
			name: "missing creator capability",
			build: func(t *testing.T) queriesStub {
				return queriesStub{
					getCreatorCapabilityByUserIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
						return sqlc.AppCreatorCapability{}, pgx.ErrNoRows
					},
					getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
						t.Fatal("GetSubmissionReviewMainByIDForUpdate() called for unavailable creator")
						return sqlc.GetSubmissionReviewMainByIDForUpdateRow{}, nil
					},
				}
			},
		},
		{
			name: "non-approved creator capability",
			build: func(t *testing.T) queriesStub {
				return queriesStub{
					getCreatorCapabilityByUserIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
						return sqlc.AppCreatorCapability{State: "pending_review"}, nil
					},
					getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
						t.Fatal("GetSubmissionReviewMainByIDForUpdate() called for unavailable creator")
						return sqlc.GetSubmissionReviewMainByIDForUpdateRow{}, nil
					},
				}
			},
		},
		{
			name: "missing package",
			build: func(t *testing.T) queriesStub {
				return queriesStub{
					getCreatorCapabilityByUserIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
						return sqlc.AppCreatorCapability{State: capabilityStateApproved}, nil
					},
					getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
						return sqlc.GetSubmissionReviewMainByIDForUpdateRow{}, pgx.ErrNoRows
					},
					listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
						t.Fatal("ListSubmissionReviewShortsByCanonicalMainIDForUpdate() called for missing package")
						return nil, nil
					},
				}
			},
		},
		{
			name: "owner mismatch",
			build: func(t *testing.T) queriesStub {
				return queriesStub{
					getCreatorCapabilityByUserIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
						return sqlc.AppCreatorCapability{State: capabilityStateApproved}, nil
					},
					getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
						return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
							ID:            postgres.UUIDToPG(mainID),
							CreatorUserID: postgres.UUIDToPG(ownerID),
						}, nil
					},
					listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
						t.Fatal("ListSubmissionReviewShortsByCanonicalMainIDForUpdate() called for owner mismatch")
						return nil, nil
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tx := &txStub{}
			service := &Service{
				beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
				now:      time.Now,
				newQueries: func(sqlc.DBTX) queries {
					q := tt.build(t)
					q.createSubmissionReviewIntake = func(context.Context, sqlc.CreateSubmissionReviewIntakeParams) (sqlc.AppSubmissionReviewIntake, error) {
						t.Fatal("CreateSubmissionReviewIntake() called for worker no-op state")
						return sqlc.AppSubmissionReviewIntake{}, nil
					}
					q.createSubmissionReviewIntakeShort = func(context.Context, sqlc.CreateSubmissionReviewIntakeShortParams) error {
						t.Fatal("CreateSubmissionReviewIntakeShort() called for worker no-op state")
						return nil
					}
					q.resetSubmissionReviewMainToPending = func(context.Context, pgtype.UUID) (sqlc.AppMain, error) {
						t.Fatal("ResetSubmissionReviewMainToPending() called for worker no-op state")
						return sqlc.AppMain{}, nil
					}
					q.resetSubmissionReviewShortToPending = func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
						t.Fatal("ResetSubmissionReviewShortToPending() called for worker no-op state")
						return sqlc.AppShort{}, nil
					}
					return q
				},
			}

			if err := service.SubmitInitialPackageIfReady(context.Background(), viewerID, mainID); err != nil {
				t.Fatalf("SubmitInitialPackageIfReady() error = %v, want nil", err)
			}
		})
	}
}

func TestSubmitPackageResubmitKeepsApprovedObjects(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 18, 12, 0, 0, 0, time.UTC)
	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	mainAssetID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortApprovedID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	shortRevisionID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	latestIntakeID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	newIntakeID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	updateShorts := 0
	updateMainCalled := false

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return &txStub{}, nil }},
		now:      func() time.Time { return now },
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getCreatorCapabilityByUserIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return sqlc.AppCreatorCapability{State: capabilityStateApproved}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
						ID:                   postgres.UUIDToPG(mainID),
						CreatorUserID:        postgres.UUIDToPG(viewerID),
						MediaAssetID:         postgres.UUIDToPG(mainAssetID),
						State:                mainStateApprovedForUnlock,
						PriceMinor:           1800,
						CurrencyCode:         "JPY",
						OwnershipConfirmed:   true,
						ConsentConfirmed:     true,
						MediaProcessingState: mediaStateReady,
					}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{
						{
							ID:                   postgres.UUIDToPG(shortApprovedID),
							CreatorUserID:        postgres.UUIDToPG(viewerID),
							CanonicalMainID:      postgres.UUIDToPG(mainID),
							State:                shortStateApprovedForPublish,
							MediaProcessingState: mediaStateReady,
						},
						{
							ID:                   postgres.UUIDToPG(shortRevisionID),
							CreatorUserID:        postgres.UUIDToPG(viewerID),
							CanonicalMainID:      postgres.UUIDToPG(mainID),
							State:                shortStateRevisionRequested,
							MediaProcessingState: mediaStateReady,
						},
					}, nil
				},
				getPendingSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
				},
				getLatestSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{
						ID:     postgres.UUIDToPG(latestIntakeID),
						Status: intakeStatusDecisionApplied,
					}, nil
				},
				createSubmissionReviewIntake: func(_ context.Context, arg sqlc.CreateSubmissionReviewIntakeParams) (sqlc.AppSubmissionReviewIntake, error) {
					if arg.SubmitKind != submitKindResubmit {
						t.Fatalf("CreateSubmissionReviewIntake() submit kind got %q want %q", arg.SubmitKind, submitKindResubmit)
					}
					if arg.PreviousIntakeID != postgres.UUIDToPG(latestIntakeID) {
						t.Fatalf("CreateSubmissionReviewIntake() previous intake id got %v want %v", arg.PreviousIntakeID, postgres.UUIDToPG(latestIntakeID))
					}
					return sqlc.AppSubmissionReviewIntake{ID: postgres.UUIDToPG(newIntakeID)}, nil
				},
				createSubmissionReviewIntakeShort: func(context.Context, sqlc.CreateSubmissionReviewIntakeShortParams) error {
					return nil
				},
				resetSubmissionReviewMainToPending: func(context.Context, pgtype.UUID) (sqlc.AppMain, error) {
					updateMainCalled = true
					return sqlc.AppMain{}, nil
				},
				resetSubmissionReviewShortToPending: func(_ context.Context, id pgtype.UUID) (sqlc.AppShort, error) {
					updateShorts++
					if id != postgres.UUIDToPG(shortRevisionID) {
						t.Fatalf("ResetSubmissionReviewShortToPending() short id got %v want %v", id, postgres.UUIDToPG(shortRevisionID))
					}
					return sqlc.AppShort{}, nil
				},
			}
		},
	}

	if err := service.SubmitPackage(context.Background(), viewerID, mainID); err != nil {
		t.Fatalf("SubmitPackage() error = %v, want nil", err)
	}
	if updateMainCalled {
		t.Fatal("SubmitPackage() updated approved main during resubmit")
	}
	if updateShorts != 1 {
		t.Fatalf("SubmitPackage() updated shorts got %d want 1", updateShorts)
	}
}

func TestSubmitInitialPackageIfReadyDoesNotResubmitRevisionPackage(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("88888888-1111-1111-1111-888888888888")
	mainID := uuid.MustParse("99999999-1111-1111-1111-999999999999")
	mainAssetID := uuid.MustParse("aaaaaaaa-1111-1111-1111-aaaaaaaaaaaa")
	shortApprovedID := uuid.MustParse("bbbbbbbb-1111-1111-1111-bbbbbbbbbbbb")
	shortRevisionID := uuid.MustParse("cccccccc-1111-1111-1111-cccccccccccc")
	latestIntakeID := uuid.MustParse("dddddddd-1111-1111-1111-dddddddddddd")
	mutated := false

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return &txStub{}, nil }},
		now:      time.Now,
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getCreatorCapabilityByUserIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return sqlc.AppCreatorCapability{State: capabilityStateApproved}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
						ID:                   postgres.UUIDToPG(mainID),
						CreatorUserID:        postgres.UUIDToPG(viewerID),
						MediaAssetID:         postgres.UUIDToPG(mainAssetID),
						State:                mainStateApprovedForUnlock,
						PriceMinor:           1800,
						CurrencyCode:         "JPY",
						OwnershipConfirmed:   true,
						ConsentConfirmed:     true,
						MediaProcessingState: mediaStateReady,
					}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{
						{
							ID:                   postgres.UUIDToPG(shortApprovedID),
							CreatorUserID:        postgres.UUIDToPG(viewerID),
							CanonicalMainID:      postgres.UUIDToPG(mainID),
							State:                shortStateApprovedForPublish,
							MediaProcessingState: mediaStateReady,
						},
						{
							ID:                   postgres.UUIDToPG(shortRevisionID),
							CreatorUserID:        postgres.UUIDToPG(viewerID),
							CanonicalMainID:      postgres.UUIDToPG(mainID),
							State:                shortStateRevisionRequested,
							MediaProcessingState: mediaStateReady,
						},
					}, nil
				},
				getPendingSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
				},
				getLatestSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{
						ID:     postgres.UUIDToPG(latestIntakeID),
						Status: intakeStatusDecisionApplied,
					}, nil
				},
				createSubmissionReviewIntake: func(context.Context, sqlc.CreateSubmissionReviewIntakeParams) (sqlc.AppSubmissionReviewIntake, error) {
					mutated = true
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				createSubmissionReviewIntakeShort: func(context.Context, sqlc.CreateSubmissionReviewIntakeShortParams) error {
					mutated = true
					return nil
				},
				resetSubmissionReviewMainToPending: func(context.Context, pgtype.UUID) (sqlc.AppMain, error) {
					mutated = true
					return sqlc.AppMain{}, nil
				},
				resetSubmissionReviewShortToPending: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
					mutated = true
					return sqlc.AppShort{}, nil
				},
			}
		},
	}

	if err := service.SubmitInitialPackageIfReady(context.Background(), viewerID, mainID); err != nil {
		t.Fatalf("SubmitInitialPackageIfReady() error = %v, want nil", err)
	}
	if mutated {
		t.Fatal("SubmitInitialPackageIfReady() created a resubmit intake or changed state")
	}
}

func TestSubmitPackageRejectsReopenFromRejectedState(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return &txStub{}, nil }},
		now:      time.Now,
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getCreatorCapabilityByUserIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return sqlc.AppCreatorCapability{State: capabilityStateApproved}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
						ID:                   postgres.UUIDToPG(mainID),
						CreatorUserID:        postgres.UUIDToPG(viewerID),
						State:                mainStateRejected,
						PriceMinor:           1800,
						CurrencyCode:         "JPY",
						OwnershipConfirmed:   true,
						ConsentConfirmed:     true,
						MediaProcessingState: mediaStateReady,
					}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					return []sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{{
						State:                shortStateApprovedForPublish,
						MediaProcessingState: mediaStateReady,
					}}, nil
				},
				getPendingSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
				},
				getLatestSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					return sqlc.AppSubmissionReviewIntake{Status: intakeStatusDecisionApplied}, nil
				},
				createSubmissionReviewIntake: func(context.Context, sqlc.CreateSubmissionReviewIntakeParams) (sqlc.AppSubmissionReviewIntake, error) {
					t.Fatal("CreateSubmissionReviewIntake() called for rejected state")
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				createSubmissionReviewIntakeShort: func(context.Context, sqlc.CreateSubmissionReviewIntakeShortParams) error {
					t.Fatal("CreateSubmissionReviewIntakeShort() called for rejected state")
					return nil
				},
				resetSubmissionReviewMainToPending: func(context.Context, pgtype.UUID) (sqlc.AppMain, error) {
					t.Fatal("ResetSubmissionReviewMainToPending() called for rejected state")
					return sqlc.AppMain{}, nil
				},
				resetSubmissionReviewShortToPending: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
					t.Fatal("ResetSubmissionReviewShortToPending() called for rejected state")
					return sqlc.AppShort{}, nil
				},
			}
		},
	}

	if err := service.SubmitPackage(context.Background(), viewerID, mainID); !errors.Is(err, ErrReviewStateConflict) {
		t.Fatalf("SubmitPackage() error got %v want %v", err, ErrReviewStateConflict)
	}
}

func TestNotReadyErrorError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  *NotReadyError
		want string
	}{
		{
			name: "nil error",
			err:  nil,
			want: "submission package is not ready",
		},
		{
			name: "empty blockers",
			err:  &NotReadyError{},
			want: "submission package is not ready",
		},
		{
			name: "with blockers",
			err: &NotReadyError{
				Blockers: []string{"main_asset_not_ready", "ownership_missing"},
			},
			want: "submission package is not ready: main_asset_not_ready,ownership_missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.err.Error(); got != tt.want {
				t.Fatalf("NotReadyError.Error() got %q want %q", got, tt.want)
			}
		})
	}
}

func TestSubmitPackageReturnsInitializationError(t *testing.T) {
	t.Parallel()

	service := &Service{}

	err := service.SubmitPackage(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("SubmitPackage() error = nil, want non-nil")
	}
	if got, want := err.Error(), "submission review service is not initialized"; got != want {
		t.Fatalf("SubmitPackage() error got %q want %q", got, want)
	}
}

func TestSubmitPackageRejectsMissingCreatorCapability(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("21212121-2121-2121-2121-212121212121")
	mainID := uuid.MustParse("31313131-3131-3131-3131-313131313131")
	tx := &txStub{}
	mutated := false

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
		now:      time.Now,
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getCreatorCapabilityByUserIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return sqlc.AppCreatorCapability{}, pgx.ErrNoRows
				},
				getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					mutated = true
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					mutated = true
					return nil, nil
				},
				getPendingSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					mutated = true
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				getLatestSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					mutated = true
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				createSubmissionReviewIntake: func(context.Context, sqlc.CreateSubmissionReviewIntakeParams) (sqlc.AppSubmissionReviewIntake, error) {
					mutated = true
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				createSubmissionReviewIntakeShort: func(context.Context, sqlc.CreateSubmissionReviewIntakeShortParams) error {
					mutated = true
					return nil
				},
				resetSubmissionReviewMainToPending: func(context.Context, pgtype.UUID) (sqlc.AppMain, error) {
					mutated = true
					return sqlc.AppMain{}, nil
				},
				resetSubmissionReviewShortToPending: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
					mutated = true
					return sqlc.AppShort{}, nil
				},
			}
		},
	}

	if err := service.SubmitPackage(context.Background(), viewerID, mainID); !errors.Is(err, ErrCreatorModeUnavailable) {
		t.Fatalf("SubmitPackage() error got %v want %v", err, ErrCreatorModeUnavailable)
	}
	if mutated {
		t.Fatal("SubmitPackage() mutated state despite missing creator capability")
	}
}

func TestSubmitPackageRejectsNonApprovedCreatorCapability(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("41414141-4141-4141-4141-414141414141")
	mainID := uuid.MustParse("51515151-5151-5151-5151-515151515151")
	tx := &txStub{}
	mutated := false

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
		now:      time.Now,
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getCreatorCapabilityByUserIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return sqlc.AppCreatorCapability{State: "pending"}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					mutated = true
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					mutated = true
					return nil, nil
				},
				getPendingSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					mutated = true
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				getLatestSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					mutated = true
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				createSubmissionReviewIntake: func(context.Context, sqlc.CreateSubmissionReviewIntakeParams) (sqlc.AppSubmissionReviewIntake, error) {
					mutated = true
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				createSubmissionReviewIntakeShort: func(context.Context, sqlc.CreateSubmissionReviewIntakeShortParams) error {
					mutated = true
					return nil
				},
				resetSubmissionReviewMainToPending: func(context.Context, pgtype.UUID) (sqlc.AppMain, error) {
					mutated = true
					return sqlc.AppMain{}, nil
				},
				resetSubmissionReviewShortToPending: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
					mutated = true
					return sqlc.AppShort{}, nil
				},
			}
		},
	}

	if err := service.SubmitPackage(context.Background(), viewerID, mainID); !errors.Is(err, ErrCreatorModeUnavailable) {
		t.Fatalf("SubmitPackage() error got %v want %v", err, ErrCreatorModeUnavailable)
	}
	if mutated {
		t.Fatal("SubmitPackage() mutated state despite non-approved creator capability")
	}
}

func TestSubmitPackageRejectsOwnerMismatch(t *testing.T) {
	t.Parallel()

	viewerID := uuid.MustParse("61616161-6161-6161-6161-616161616161")
	ownerID := uuid.MustParse("71717171-7171-7171-7171-717171717171")
	mainID := uuid.MustParse("81818181-8181-8181-8181-818181818181")
	tx := &txStub{}
	mutated := false

	service := &Service{
		beginner: txBeginnerStub{begin: func(context.Context) (pgx.Tx, error) { return tx, nil }},
		now:      time.Now,
		newQueries: func(sqlc.DBTX) queries {
			return queriesStub{
				getCreatorCapabilityByUserIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return sqlc.AppCreatorCapability{State: capabilityStateApproved}, nil
				},
				getSubmissionReviewMainByIDForUpdate: func(context.Context, pgtype.UUID) (sqlc.GetSubmissionReviewMainByIDForUpdateRow, error) {
					return sqlc.GetSubmissionReviewMainByIDForUpdateRow{
						ID:            postgres.UUIDToPG(mainID),
						CreatorUserID: postgres.UUIDToPG(ownerID),
					}, nil
				},
				listSubmissionReviewShortsByCanonicalMainIDForUpdate: func(context.Context, pgtype.UUID) ([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, error) {
					mutated = true
					return nil, nil
				},
				getPendingSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					mutated = true
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				getLatestSubmissionReviewIntakeByCanonicalMainID: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
					mutated = true
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				createSubmissionReviewIntake: func(context.Context, sqlc.CreateSubmissionReviewIntakeParams) (sqlc.AppSubmissionReviewIntake, error) {
					mutated = true
					return sqlc.AppSubmissionReviewIntake{}, nil
				},
				createSubmissionReviewIntakeShort: func(context.Context, sqlc.CreateSubmissionReviewIntakeShortParams) error {
					mutated = true
					return nil
				},
				resetSubmissionReviewMainToPending: func(context.Context, pgtype.UUID) (sqlc.AppMain, error) {
					mutated = true
					return sqlc.AppMain{}, nil
				},
				resetSubmissionReviewShortToPending: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
					mutated = true
					return sqlc.AppShort{}, nil
				},
			}
		},
	}

	if err := service.SubmitPackage(context.Background(), viewerID, mainID); !errors.Is(err, ErrSubmissionPackageNotFound) {
		t.Fatalf("SubmitPackage() error got %v want %v", err, ErrSubmissionPackageNotFound)
	}
	if mutated {
		t.Fatal("SubmitPackage() mutated state despite owner mismatch")
	}
}

func TestLoadLatestIntake(t *testing.T) {
	t.Parallel()

	mainID := postgres.UUIDToPG(uuid.MustParse("abababab-abab-abab-abab-abababababab"))
	latestID := postgres.UUIDToPG(uuid.MustParse("cdcdcdcd-cdcd-cdcd-cdcd-cdcdcdcdcdcd"))

	tests := []struct {
		name         string
		load         func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error)
		wantExists   bool
		wantErr      error
		wantIntakeID pgtype.UUID
	}{
		{
			name: "latest exists",
			load: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
				return sqlc.AppSubmissionReviewIntake{ID: latestID}, nil
			},
			wantExists:   true,
			wantIntakeID: latestID,
		},
		{
			name: "not found",
			load: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
				return sqlc.AppSubmissionReviewIntake{}, pgx.ErrNoRows
			},
			wantExists: false,
		},
		{
			name: "unexpected error",
			load: func(context.Context, pgtype.UUID) (sqlc.AppSubmissionReviewIntake, error) {
				return sqlc.AppSubmissionReviewIntake{}, errors.New("boom")
			},
			wantErr: errors.New("boom"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			q := queriesStub{
				getLatestSubmissionReviewIntakeByCanonicalMainID: tt.load,
			}

			gotIntake, gotExists, err := loadLatestIntake(context.Background(), q, mainID)
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Fatalf("loadLatestIntake() error got %v want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("loadLatestIntake() error = %v, want nil", err)
			}
			if gotExists != tt.wantExists {
				t.Fatalf("loadLatestIntake() exists got %t want %t", gotExists, tt.wantExists)
			}
			if gotIntake.ID != tt.wantIntakeID {
				t.Fatalf("loadLatestIntake() intake id got %v want %v", gotIntake.ID, tt.wantIntakeID)
			}
		})
	}
}

func TestDetermineTransition(t *testing.T) {
	t.Parallel()

	latestID := postgres.UUIDToPG(uuid.MustParse("dededede-dede-dede-dede-dededededede"))

	tests := []struct {
		name                 string
		mainState            string
		shortStates          []string
		latestExists         bool
		latestStatus         string
		wantSubmitKind       string
		wantUpdateMain       bool
		wantPendingShorts    int
		wantPreviousIntakeID pgtype.UUID
		wantErr              error
	}{
		{
			name:              "initial submit from all draft",
			mainState:         mainStateDraft,
			shortStates:       []string{shortStateDraft, shortStateDraft},
			wantSubmitKind:    submitKindInitial,
			wantUpdateMain:    true,
			wantPendingShorts: 2,
		},
		{
			name:      "pending review conflicts",
			mainState: mainStatePendingReview,
			shortStates: []string{
				shortStateApprovedForPublish,
			},
			wantErr: ErrReviewStateConflict,
		},
		{
			name:      "revision requested without latest decision conflicts",
			mainState: mainStateRevisionRequested,
			shortStates: []string{
				shortStateApprovedForPublish,
			},
			wantErr: ErrReviewStateConflict,
		},
		{
			name:                 "resubmit reopens revision requested objects only",
			mainState:            mainStateApprovedForUnlock,
			shortStates:          []string{shortStateApprovedForPublish, shortStateRevisionRequested},
			latestExists:         true,
			latestStatus:         intakeStatusDecisionApplied,
			wantSubmitKind:       submitKindResubmit,
			wantUpdateMain:       false,
			wantPendingShorts:    1,
			wantPreviousIntakeID: latestID,
		},
		{
			name:         "rejected short conflicts",
			mainState:    mainStateApprovedForUnlock,
			shortStates:  []string{shortStateRejected},
			latestExists: true,
			latestStatus: intakeStatusDecisionApplied,
			wantErr:      ErrReviewStateConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			shortRows := make([]sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow, 0, len(tt.shortStates))
			for _, state := range tt.shortStates {
				shortRows = append(shortRows, sqlc.ListSubmissionReviewShortsByCanonicalMainIDForUpdateRow{
					State: state,
				})
			}

			got, err := determineTransition(
				sqlc.GetSubmissionReviewMainByIDForUpdateRow{State: tt.mainState},
				shortRows,
				tt.latestExists,
				sqlc.AppSubmissionReviewIntake{
					ID:     latestID,
					Status: tt.latestStatus,
				},
			)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("determineTransition() error got %v want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("determineTransition() error = %v, want nil", err)
			}
			if got.SubmitKind != tt.wantSubmitKind {
				t.Fatalf("determineTransition() submit kind got %q want %q", got.SubmitKind, tt.wantSubmitKind)
			}
			if got.UpdateMain != tt.wantUpdateMain {
				t.Fatalf("determineTransition() update main got %t want %t", got.UpdateMain, tt.wantUpdateMain)
			}
			if len(got.ShortsToPending) != tt.wantPendingShorts {
				t.Fatalf("determineTransition() shorts to pending got %d want %d", len(got.ShortsToPending), tt.wantPendingShorts)
			}
			if got.PreviousIntakeID != tt.wantPreviousIntakeID {
				t.Fatalf("determineTransition() previous intake id got %v want %v", got.PreviousIntakeID, tt.wantPreviousIntakeID)
			}
		})
	}
}
