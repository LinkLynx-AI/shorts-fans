package creatorregistration

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repositoryQueriesStub struct {
	createCreatorCapability            func(context.Context, sqlc.CreateCreatorCapabilityParams) (sqlc.AppCreatorCapability, error)
	createCreatorProfile               func(context.Context, sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error)
	getCreatorCapabilityByUserID       func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error)
	getCreatorCapabilityByUserIDLocked func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error)
	getCreatorProfileByUserID          func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error)
	getCreatorRegistrationIntakeByUser func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error)
	getUserProfileByUserID             func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error)
	listCreatorRegistrationEvidences   func(context.Context, pgtype.UUID) ([]sqlc.AppCreatorRegistrationEvidence, error)
	listCreatorRegistrationReviewCases func(context.Context, string) ([]sqlc.ListCreatorRegistrationReviewCasesByStateRow, error)
	updateCreatorCapabilityState       func(context.Context, sqlc.UpdateCreatorCapabilityStateParams) (sqlc.AppCreatorCapability, error)
	updateCreatorProfile               func(context.Context, sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error)
	upsertCreatorRegistrationEvidence  func(context.Context, sqlc.UpsertCreatorRegistrationEvidenceParams) (sqlc.AppCreatorRegistrationEvidence, error)
	upsertCreatorRegistrationIntake    func(context.Context, sqlc.UpsertCreatorRegistrationIntakeParams) (sqlc.AppCreatorRegistrationIntake, error)
}

func (s repositoryQueriesStub) CreateCreatorCapability(ctx context.Context, arg sqlc.CreateCreatorCapabilityParams) (sqlc.AppCreatorCapability, error) {
	if s.createCreatorCapability == nil {
		panic("unexpected CreateCreatorCapability call")
	}
	return s.createCreatorCapability(ctx, arg)
}

func (s repositoryQueriesStub) CreateCreatorProfile(ctx context.Context, arg sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
	if s.createCreatorProfile == nil {
		panic("unexpected CreateCreatorProfile call")
	}
	return s.createCreatorProfile(ctx, arg)
}

func (s repositoryQueriesStub) GetCreatorCapabilityByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
	if s.getCreatorCapabilityByUserID == nil {
		panic("unexpected GetCreatorCapabilityByUserID call")
	}
	return s.getCreatorCapabilityByUserID(ctx, userID)
}

func (s repositoryQueriesStub) GetCreatorCapabilityByUserIDForUpdate(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
	if s.getCreatorCapabilityByUserIDLocked != nil {
		return s.getCreatorCapabilityByUserIDLocked(ctx, userID)
	}
	if s.getCreatorCapabilityByUserID != nil {
		return s.getCreatorCapabilityByUserID(ctx, userID)
	}
	panic("unexpected GetCreatorCapabilityByUserIDForUpdate call")
}

func (s repositoryQueriesStub) GetCreatorProfileByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorProfile, error) {
	if s.getCreatorProfileByUserID == nil {
		panic("unexpected GetCreatorProfileByUserID call")
	}
	return s.getCreatorProfileByUserID(ctx, userID)
}

func (s repositoryQueriesStub) GetCreatorRegistrationIntakeByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
	if s.getCreatorRegistrationIntakeByUser == nil {
		panic("unexpected GetCreatorRegistrationIntakeByUserID call")
	}
	return s.getCreatorRegistrationIntakeByUser(ctx, userID)
}

func (s repositoryQueriesStub) GetUserProfileByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppUserProfile, error) {
	if s.getUserProfileByUserID == nil {
		panic("unexpected GetUserProfileByUserID call")
	}
	return s.getUserProfileByUserID(ctx, userID)
}

func (s repositoryQueriesStub) ListCreatorRegistrationEvidencesByUserID(ctx context.Context, userID pgtype.UUID) ([]sqlc.AppCreatorRegistrationEvidence, error) {
	if s.listCreatorRegistrationEvidences == nil {
		panic("unexpected ListCreatorRegistrationEvidencesByUserID call")
	}
	return s.listCreatorRegistrationEvidences(ctx, userID)
}

func (s repositoryQueriesStub) ListCreatorRegistrationReviewCasesByState(ctx context.Context, state string) ([]sqlc.ListCreatorRegistrationReviewCasesByStateRow, error) {
	if s.listCreatorRegistrationReviewCases == nil {
		panic("unexpected ListCreatorRegistrationReviewCasesByState call")
	}
	return s.listCreatorRegistrationReviewCases(ctx, state)
}

func (s repositoryQueriesStub) UpdateCreatorCapabilityState(ctx context.Context, arg sqlc.UpdateCreatorCapabilityStateParams) (sqlc.AppCreatorCapability, error) {
	if s.updateCreatorCapabilityState == nil {
		panic("unexpected UpdateCreatorCapabilityState call")
	}
	return s.updateCreatorCapabilityState(ctx, arg)
}

func (s repositoryQueriesStub) UpdateCreatorProfile(ctx context.Context, arg sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
	if s.updateCreatorProfile == nil {
		panic("unexpected UpdateCreatorProfile call")
	}
	return s.updateCreatorProfile(ctx, arg)
}

func (s repositoryQueriesStub) UpsertCreatorRegistrationEvidence(ctx context.Context, arg sqlc.UpsertCreatorRegistrationEvidenceParams) (sqlc.AppCreatorRegistrationEvidence, error) {
	if s.upsertCreatorRegistrationEvidence == nil {
		panic("unexpected UpsertCreatorRegistrationEvidence call")
	}
	return s.upsertCreatorRegistrationEvidence(ctx, arg)
}

func (s repositoryQueriesStub) UpsertCreatorRegistrationIntake(ctx context.Context, arg sqlc.UpsertCreatorRegistrationIntakeParams) (sqlc.AppCreatorRegistrationIntake, error) {
	if s.upsertCreatorRegistrationIntake == nil {
		panic("unexpected UpsertCreatorRegistrationIntake call")
	}
	return s.upsertCreatorRegistrationIntake(ctx, arg)
}

type repositoryTxBeginnerStub struct {
	begin func(context.Context) (pgx.Tx, error)
}

func (s repositoryTxBeginnerStub) Begin(ctx context.Context) (pgx.Tx, error) {
	return s.begin(ctx)
}

type repositoryTxStub struct {
	commitErr   error
	rollbackErr error
	committed   bool
	rolledBack  bool
}

func (tx *repositoryTxStub) Begin(context.Context) (pgx.Tx, error) { return tx, nil }
func (tx *repositoryTxStub) Commit(context.Context) error {
	tx.committed = true
	return tx.commitErr
}
func (tx *repositoryTxStub) Rollback(context.Context) error {
	tx.rolledBack = true
	return tx.rollbackErr
}
func (tx *repositoryTxStub) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (tx *repositoryTxStub) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (tx *repositoryTxStub) LargeObjects() pgx.LargeObjects                         { return pgx.LargeObjects{} }
func (tx *repositoryTxStub) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (tx *repositoryTxStub) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (tx *repositoryTxStub) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }
func (tx *repositoryTxStub) QueryRow(context.Context, string, ...any) pgx.Row        { return nil }
func (tx *repositoryTxStub) Conn() *pgx.Conn                                         { return nil }

func TestRepositoryGetRegistrationReturnsNilWithoutCapability(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repo := newRepository(repositoryQueriesStub{
		getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
			return testUserProfile(userID), nil
		},
		getCreatorCapabilityByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return sqlc.AppCreatorCapability{}, pgx.ErrNoRows
		},
	})

	registration, err := repo.GetRegistration(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetRegistration() error = %v, want nil", err)
	}
	if registration != nil {
		t.Fatalf("GetRegistration() got %#v, want nil", registration)
	}
}

func TestRepositoryGetIntakeBuildsEditableSnapshot(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	repo := newRepository(repositoryQueriesStub{
		getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
			return testUserProfile(userID), nil
		},
		getCreatorCapabilityByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapability(userID, StateDraft), nil
		},
		getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testCreatorProfile(userID, "draft bio"), nil
		},
		getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
			return testIntake(userID), nil
		},
		listCreatorRegistrationEvidences: func(context.Context, pgtype.UUID) ([]sqlc.AppCreatorRegistrationEvidence, error) {
			return testEvidenceRows(userID), nil
		},
	})

	intake, err := repo.GetIntake(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetIntake() error = %v, want nil", err)
	}
	if !intake.CanSubmit {
		t.Fatal("GetIntake() CanSubmit = false, want true")
	}
	if intake.IsReadOnly {
		t.Fatal("GetIntake() IsReadOnly = true, want false")
	}
	if intake.CreatorBio != "draft bio" {
		t.Fatalf("GetIntake() CreatorBio got %q want %q", intake.CreatorBio, "draft bio")
	}
	if got := intake.SharedProfile.Handle; got != "creator.handle" {
		t.Fatalf("GetIntake() SharedProfile.Handle got %q want %q", got, "creator.handle")
	}
	if len(intake.Evidences) != 5 {
		t.Fatalf("GetIntake() evidences len got %d want 5", len(intake.Evidences))
	}
}

func TestRepositoryGetIntakeAllowsEligibleRejectedResubmit(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("23232323-2323-2323-2323-232323232323")
	capability := testCapability(userID, StateRejected)
	capability.IsResubmitEligible = true
	capability.SelfServeResubmitCount = 1

	repo := newRepository(repositoryQueriesStub{
		getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
			return testUserProfile(userID), nil
		},
		getCreatorCapabilityByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return capability, nil
		},
		getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testCreatorProfile(userID, "retry bio"), nil
		},
		getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
			return testIntake(userID), nil
		},
		listCreatorRegistrationEvidences: func(context.Context, pgtype.UUID) ([]sqlc.AppCreatorRegistrationEvidence, error) {
			return testEvidenceRows(userID), nil
		},
	})

	intake, err := repo.GetIntake(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetIntake() error = %v, want nil", err)
	}
	if intake.IsReadOnly {
		t.Fatal("GetIntake() IsReadOnly = true, want false for eligible rejected")
	}
	if !intake.CanSubmit {
		t.Fatal("GetIntake() CanSubmit = false, want true for eligible rejected")
	}
}

func TestRepositoryPrepareEvidenceUploadCreatesDraftCapability(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	tx := &repositoryTxStub{}
	created := false
	repo := &Repository{
		txBeginner: repositoryTxBeginnerStub{
			begin: func(context.Context) (pgx.Tx, error) { return tx, nil },
		},
		newQueries: func(sqlc.DBTX) queries {
			return repositoryQueriesStub{
				getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
					return testUserProfile(userID), nil
				},
				getCreatorCapabilityByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return sqlc.AppCreatorCapability{}, pgx.ErrNoRows
				},
				createCreatorCapability: func(_ context.Context, arg sqlc.CreateCreatorCapabilityParams) (sqlc.AppCreatorCapability, error) {
					created = true
					if arg.State != StateDraft {
						t.Fatalf("CreateCreatorCapability() state got %q want %q", arg.State, StateDraft)
					}
					return testCapability(userID, StateDraft), nil
				},
			}
		},
	}

	if err := repo.PrepareEvidenceUpload(context.Background(), userID); err != nil {
		t.Fatalf("PrepareEvidenceUpload() error = %v, want nil", err)
	}
	if !created {
		t.Fatal("PrepareEvidenceUpload() did not create draft capability")
	}
	if !tx.committed {
		t.Fatal("PrepareEvidenceUpload() committed = false, want true")
	}
}

func TestRepositoryPrepareEvidenceUploadAllowsEligibleRejected(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("34343434-3434-3434-3434-343434343434")
	tx := &repositoryTxStub{}
	capability := testCapability(userID, StateRejected)
	capability.IsResubmitEligible = true
	capability.SelfServeResubmitCount = 1
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
					return capability, nil
				},
				getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return testCreatorProfile(userID, "retry bio"), nil
				},
				getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
					return testIntake(userID), nil
				},
			}
		},
	}

	if err := repo.PrepareEvidenceUpload(context.Background(), userID); err != nil {
		t.Fatalf("PrepareEvidenceUpload() error = %v, want nil", err)
	}
	if !tx.committed {
		t.Fatal("PrepareEvidenceUpload() committed = false, want true")
	}
	if !lockedCapabilityLoaded {
		t.Fatal("PrepareEvidenceUpload() did not load capability with row lock")
	}
}

func TestRepositorySaveEvidenceReplacesPreviousObject(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	tx := &repositoryTxStub{}
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
					return testCapability(userID, StateDraft), nil
				},
				getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return sqlc.AppCreatorProfile{}, pgx.ErrNoRows
				},
				getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
					return sqlc.AppCreatorRegistrationIntake{}, pgx.ErrNoRows
				},
				listCreatorRegistrationEvidences: func(context.Context, pgtype.UUID) ([]sqlc.AppCreatorRegistrationEvidence, error) {
					return []sqlc.AppCreatorRegistrationEvidence{
						testEvidenceRow(userID, EvidenceKindGovernmentID, "review-bucket", "creator-registration/evidence/old.png"),
					}, nil
				},
				upsertCreatorRegistrationEvidence: func(_ context.Context, arg sqlc.UpsertCreatorRegistrationEvidenceParams) (sqlc.AppCreatorRegistrationEvidence, error) {
					if arg.Kind != EvidenceKindGovernmentID {
						t.Fatalf("UpsertCreatorRegistrationEvidence() kind got %q want %q", arg.Kind, EvidenceKindGovernmentID)
					}
					return testEvidenceRow(userID, EvidenceKindGovernmentID, arg.StorageBucket, arg.StorageKey), nil
				},
			}
		},
	}

	result, err := repo.SaveEvidence(context.Background(), SaveEvidenceInput{
		FileName:      "government-id.png",
		FileSizeBytes: 321,
		Kind:          EvidenceKindGovernmentID,
		MimeType:      "image/png",
		StorageBucket: "review-bucket",
		StorageKey:    "creator-registration/evidence/new.png",
		UploadedAt:    time.Unix(1710000600, 0).UTC(),
		UserID:        userID,
	})
	if err != nil {
		t.Fatalf("SaveEvidence() error = %v, want nil", err)
	}
	if result.ReplacedObject == nil {
		t.Fatal("SaveEvidence() ReplacedObject = nil, want previous object")
	}
	if got := result.ReplacedObject.Key; got != "creator-registration/evidence/old.png" {
		t.Fatalf("SaveEvidence() replaced key got %q want %q", got, "creator-registration/evidence/old.png")
	}
	if got := result.Evidence.Kind; got != EvidenceKindGovernmentID {
		t.Fatalf("SaveEvidence() Evidence.Kind got %q want %q", got, EvidenceKindGovernmentID)
	}
	if !tx.committed {
		t.Fatal("SaveEvidence() committed = false, want true")
	}
	if !lockedCapabilityLoaded {
		t.Fatal("SaveEvidence() did not load capability with row lock")
	}
}

func TestRepositorySaveIntakeCreatesDraftProfileAndReturnsUpdatedIntake(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	tx := &repositoryTxStub{}
	profile := testUserProfile(userID)
	profile.Handle = "@Creator.Handle"
	var capability *sqlc.AppCreatorCapability
	var creatorProfile *sqlc.AppCreatorProfile
	var intake *sqlc.AppCreatorRegistrationIntake

	repo := &Repository{
		txBeginner: repositoryTxBeginnerStub{
			begin: func(context.Context) (pgx.Tx, error) { return tx, nil },
		},
		newQueries: func(sqlc.DBTX) queries {
			return repositoryQueriesStub{
				getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
					return profile, nil
				},
				getCreatorCapabilityByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					if capability == nil {
						return sqlc.AppCreatorCapability{}, pgx.ErrNoRows
					}
					return *capability, nil
				},
				createCreatorCapability: func(_ context.Context, arg sqlc.CreateCreatorCapabilityParams) (sqlc.AppCreatorCapability, error) {
					row := testCapability(userID, arg.State)
					capability = &row
					return row, nil
				},
				getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					if creatorProfile == nil {
						return sqlc.AppCreatorProfile{}, pgx.ErrNoRows
					}
					return *creatorProfile, nil
				},
				createCreatorProfile: func(_ context.Context, arg sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
					row := sqlc.AppCreatorProfile{
						UserID:      arg.UserID,
						DisplayName: arg.DisplayName,
						AvatarUrl:   arg.AvatarUrl,
						Bio:         arg.Bio,
						Handle:      arg.Handle,
					}
					creatorProfile = &row
					return row, nil
				},
				getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
					if intake == nil {
						return sqlc.AppCreatorRegistrationIntake{}, pgx.ErrNoRows
					}
					return *intake, nil
				},
				upsertCreatorRegistrationIntake: func(_ context.Context, arg sqlc.UpsertCreatorRegistrationIntakeParams) (sqlc.AppCreatorRegistrationIntake, error) {
					row := sqlc.AppCreatorRegistrationIntake{
						UserID:                                  arg.UserID,
						LegalName:                               arg.LegalName,
						BirthDate:                               arg.BirthDate,
						LegalAddress:                            arg.LegalAddress,
						IdentityDocumentType:                    arg.IdentityDocumentType,
						TargetAudienceCategory:                  arg.TargetAudienceCategory,
						HasCoPerformers:                         arg.HasCoPerformers,
						PayoutRecipientType:                     arg.PayoutRecipientType,
						PayoutRecipientName:                     arg.PayoutRecipientName,
						DeclaresNoProhibitedCategory:            arg.DeclaresNoProhibitedCategory,
						AcceptsConsentResponsibility:            arg.AcceptsConsentResponsibility,
						AcceptsAppearanceVerification:           arg.AcceptsAppearanceVerification,
						AcceptsCoPerformerConsentResponsibility: arg.AcceptsCoPerformerConsentResponsibility,
						AcceptsAdultBusinessCompliance:          arg.AcceptsAdultBusinessCompliance,
						ConfirmsInformationMatchesDocuments:     arg.ConfirmsInformationMatchesDocuments,
					}
					intake = &row
					return row, nil
				},
				listCreatorRegistrationEvidences: func(context.Context, pgtype.UUID) ([]sqlc.AppCreatorRegistrationEvidence, error) {
					return testEvidenceRows(userID), nil
				},
			}
		},
	}

	got, err := repo.SaveIntake(context.Background(), SaveIntakeInput{
		AcceptsAdultBusinessCompliance:      true,
		AcceptsAppearanceVerification:       true,
		AcceptsConsentResponsibility:        true,
		BirthDate:                           " 2000-01-02 ",
		ConfirmsInformationMatchesDocuments: true,
		CreatorBio:                          "  updated bio  ",
		DeclaresNoProhibitedCategory:        true,
		IdentityDocumentType:                IdentityDocumentTypeDriverLicense,
		LegalAddress:                        "  Tokyo-to Shibuya-ku 1-2-3  ",
		LegalName:                           "  Creator Name  ",
		PayoutRecipientName:                 "  Creator Biz  ",
		PayoutRecipientType:                 PayoutRecipientTypeBusiness,
		TargetAudienceCategory:              TargetAudienceCategoryGeneralAdult,
		UserID:                              userID,
	})
	if err != nil {
		t.Fatalf("SaveIntake() error = %v, want nil", err)
	}
	if got.CreatorBio != "updated bio" {
		t.Fatalf("SaveIntake() CreatorBio got %q want %q", got.CreatorBio, "updated bio")
	}
	if got.PayoutRecipientType != PayoutRecipientTypeBusiness {
		t.Fatalf("SaveIntake() PayoutRecipientType got %q want %q", got.PayoutRecipientType, PayoutRecipientTypeBusiness)
	}
	if got.LegalAddress != "Tokyo-to Shibuya-ku 1-2-3" {
		t.Fatalf("SaveIntake() LegalAddress got %q want trimmed address", got.LegalAddress)
	}
	if !got.CanSubmit {
		t.Fatal("SaveIntake() CanSubmit = false, want true")
	}
	if creatorProfile == nil || creatorProfile.Handle != "creator.handle" {
		t.Fatalf("SaveIntake() creatorProfile handle got %#v want creator.handle", creatorProfile)
	}
	if !tx.committed {
		t.Fatal("SaveIntake() committed = false, want true")
	}
}

func TestRepositorySaveIntakeUpdatesExistingDraftProfile(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("56565656-5656-5656-5656-565656565656")
	tx := &repositoryTxStub{}
	profile := testUserProfile(userID)
	profile.Handle = "Updated.Creator"
	existingCapability := testCapability(userID, StateDraft)
	existingProfile := testCreatorProfile(userID, "before")
	var savedIntake sqlc.AppCreatorRegistrationIntake
	lockedCapabilityLoaded := false

	repo := &Repository{
		txBeginner: repositoryTxBeginnerStub{
			begin: func(context.Context) (pgx.Tx, error) { return tx, nil },
		},
		newQueries: func(sqlc.DBTX) queries {
			return repositoryQueriesStub{
				getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
					return profile, nil
				},
				getCreatorCapabilityByUserIDLocked: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					lockedCapabilityLoaded = true
					return existingCapability, nil
				},
				getCreatorCapabilityByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return existingCapability, nil
				},
				getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return existingProfile, nil
				},
				updateCreatorProfile: func(_ context.Context, arg sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
					existingProfile = sqlc.AppCreatorProfile{
						UserID:      arg.UserID,
						DisplayName: arg.DisplayName,
						AvatarUrl:   arg.AvatarUrl,
						Bio:         arg.Bio,
						Handle:      arg.Handle,
					}
					return existingProfile, nil
				},
				getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
					if savedIntake.UserID.Valid {
						return savedIntake, nil
					}
					return sqlc.AppCreatorRegistrationIntake{}, pgx.ErrNoRows
				},
				upsertCreatorRegistrationIntake: func(_ context.Context, arg sqlc.UpsertCreatorRegistrationIntakeParams) (sqlc.AppCreatorRegistrationIntake, error) {
					savedIntake = sqlc.AppCreatorRegistrationIntake{
						UserID:                                  arg.UserID,
						LegalName:                               arg.LegalName,
						BirthDate:                               arg.BirthDate,
						LegalAddress:                            arg.LegalAddress,
						IdentityDocumentType:                    arg.IdentityDocumentType,
						TargetAudienceCategory:                  arg.TargetAudienceCategory,
						HasCoPerformers:                         arg.HasCoPerformers,
						PayoutRecipientType:                     arg.PayoutRecipientType,
						PayoutRecipientName:                     arg.PayoutRecipientName,
						DeclaresNoProhibitedCategory:            arg.DeclaresNoProhibitedCategory,
						AcceptsConsentResponsibility:            arg.AcceptsConsentResponsibility,
						AcceptsAppearanceVerification:           arg.AcceptsAppearanceVerification,
						AcceptsCoPerformerConsentResponsibility: arg.AcceptsCoPerformerConsentResponsibility,
						AcceptsAdultBusinessCompliance:          arg.AcceptsAdultBusinessCompliance,
						ConfirmsInformationMatchesDocuments:     arg.ConfirmsInformationMatchesDocuments,
					}
					return savedIntake, nil
				},
				listCreatorRegistrationEvidences: func(context.Context, pgtype.UUID) ([]sqlc.AppCreatorRegistrationEvidence, error) {
					return []sqlc.AppCreatorRegistrationEvidence{}, nil
				},
			}
		},
	}

	intake, err := repo.SaveIntake(context.Background(), SaveIntakeInput{
		AcceptsConsentResponsibility: true,
		BirthDate:                    "2001-02-03",
		CreatorBio:                   "updated bio",
		DeclaresNoProhibitedCategory: true,
		LegalName:                    "Updated Legal",
		PayoutRecipientName:          "Updated Recipient",
		PayoutRecipientType:          PayoutRecipientTypeSelf,
		UserID:                       userID,
	})
	if err != nil {
		t.Fatalf("SaveIntake() update error = %v, want nil", err)
	}
	if intake.CreatorBio != "updated bio" {
		t.Fatalf("SaveIntake() update bio got %q want %q", intake.CreatorBio, "updated bio")
	}
	if intake.PayoutRecipientType != PayoutRecipientTypeSelf {
		t.Fatalf("SaveIntake() update payout type got %q want %q", intake.PayoutRecipientType, PayoutRecipientTypeSelf)
	}
	if intake.CanSubmit {
		t.Fatal("SaveIntake() CanSubmit = true, want false without evidences")
	}
	if existingProfile.Handle != "updated.creator" {
		t.Fatalf("SaveIntake() updated handle got %q want %q", existingProfile.Handle, "updated.creator")
	}
	if !lockedCapabilityLoaded {
		t.Fatal("SaveIntake() did not load capability with row lock")
	}
}

func TestRepositorySubmitTransitionsToSubmitted(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	tx := &repositoryTxStub{}
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
					return testCapability(userID, StateDraft), nil
				},
				getCreatorCapabilityByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return testCapability(userID, StateDraft), nil
				},
				getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return testCreatorProfile(userID, "draft bio"), nil
				},
				getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
					return testIntake(userID), nil
				},
				listCreatorRegistrationEvidences: func(context.Context, pgtype.UUID) ([]sqlc.AppCreatorRegistrationEvidence, error) {
					return testEvidenceRows(userID), nil
				},
				updateCreatorCapabilityState: func(_ context.Context, arg sqlc.UpdateCreatorCapabilityStateParams) (sqlc.AppCreatorCapability, error) {
					if arg.State != StateSubmitted {
						t.Fatalf("UpdateCreatorCapabilityState() state got %q want %q", arg.State, StateSubmitted)
					}
					row := testCapability(userID, StateSubmitted)
					row.SubmittedAt = arg.SubmittedAt
					return row, nil
				},
			}
		},
	}

	registration, err := repo.Submit(context.Background(), userID)
	if err != nil {
		t.Fatalf("Submit() error = %v, want nil", err)
	}
	if registration.State != StateSubmitted {
		t.Fatalf("Submit() state got %q want %q", registration.State, StateSubmitted)
	}
	if registration.Actions.CanSubmit {
		t.Fatal("Submit() Actions.CanSubmit = true, want false")
	}
	if registration.Review.SubmittedAt == nil {
		t.Fatal("Submit() SubmittedAt = nil, want value")
	}
	if !tx.committed {
		t.Fatal("Submit() committed = false, want true")
	}
	if !lockedCapabilityLoaded {
		t.Fatal("Submit() did not load capability with row lock")
	}
}

func TestRepositorySubmitTransitionsEligibleRejectedToSubmitted(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("67676767-6767-6767-6767-676767676767")
	tx := &repositoryTxStub{}
	rejectedAt := time.Unix(1710000600, 0).UTC()
	existing := testCapability(userID, StateRejected)
	existing.IsResubmitEligible = true
	existing.SelfServeResubmitCount = 1
	existing.RejectedAt = postgres.TimeToPG(&rejectedAt)
	reason := "documents_incomplete"
	existing.RejectionReasonCode = postgres.TextToPG(&reason)
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
					return testCreatorProfile(userID, "retry bio"), nil
				},
				getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
					return testIntake(userID), nil
				},
				listCreatorRegistrationEvidences: func(context.Context, pgtype.UUID) ([]sqlc.AppCreatorRegistrationEvidence, error) {
					return testEvidenceRows(userID), nil
				},
				updateCreatorCapabilityState: func(_ context.Context, arg sqlc.UpdateCreatorCapabilityStateParams) (sqlc.AppCreatorCapability, error) {
					if arg.State != StateSubmitted {
						t.Fatalf("UpdateCreatorCapabilityState() state got %q want %q", arg.State, StateSubmitted)
					}
					if arg.SelfServeResubmitCount != 2 {
						t.Fatalf("UpdateCreatorCapabilityState() resubmit count got %d want %d", arg.SelfServeResubmitCount, 2)
					}
					if arg.RejectionReasonCode.Valid {
						t.Fatalf("UpdateCreatorCapabilityState() rejection reason got %#v want empty", arg.RejectionReasonCode)
					}
					if arg.IsResubmitEligible {
						t.Fatal("UpdateCreatorCapabilityState() isResubmitEligible = true, want false")
					}
					if arg.IsSupportReviewRequired {
						t.Fatal("UpdateCreatorCapabilityState() isSupportReviewRequired = true, want false")
					}
					if arg.RejectedAt.Valid {
						t.Fatal("UpdateCreatorCapabilityState() rejectedAt valid, want empty")
					}
					row := existing
					row.State = StateSubmitted
					row.SelfServeResubmitCount = arg.SelfServeResubmitCount
					row.RejectionReasonCode = pgtype.Text{}
					row.IsResubmitEligible = false
					row.IsSupportReviewRequired = false
					row.SubmittedAt = arg.SubmittedAt
					row.RejectedAt = pgtype.Timestamptz{}
					return row, nil
				},
			}
		},
	}

	registration, err := repo.Submit(context.Background(), userID)
	if err != nil {
		t.Fatalf("Submit() error = %v, want nil", err)
	}
	if registration.State != StateSubmitted {
		t.Fatalf("Submit() state got %q want %q", registration.State, StateSubmitted)
	}
	if registration.Actions.CanResubmit {
		t.Fatal("Submit() Actions.CanResubmit = true, want false after resubmit")
	}
	if registration.Rejection != nil {
		t.Fatalf("Submit() rejection got %#v want nil", registration.Rejection)
	}
	if !tx.committed {
		t.Fatal("Submit() committed = false, want true")
	}
	if !lockedCapabilityLoaded {
		t.Fatal("Submit() did not load capability with row lock")
	}
}

func TestRegistrationHelpersAndNormalization(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	approved, err := buildRegistration(registrationSnapshot{
		capability:     ptrCapability(testCapability(userID, StateApproved)),
		creatorProfile: ptrCreatorProfile(testCreatorProfile(userID, "bio")),
		userProfile:    testUserProfile(userID),
	})
	if err != nil {
		t.Fatalf("buildRegistration() approved error = %v, want nil", err)
	}
	if !approved.Actions.CanEnterCreatorMode {
		t.Fatal("buildRegistration() approved CanEnterCreatorMode = false, want true")
	}
	if approved.Surface.Kind != "creator_workspace" {
		t.Fatalf("buildRegistration() approved surface got %q want %q", approved.Surface.Kind, "creator_workspace")
	}

	reason := "documents_blurry"
	rejected, err := buildRegistration(registrationSnapshot{
		capability: ptrCapability(sqlc.AppCreatorCapability{
			UserID:                  postgres.UUIDToPG(userID),
			State:                   StateRejected,
			RejectionReasonCode:     postgres.TextToPG(&reason),
			IsResubmitEligible:      true,
			SelfServeResubmitCount:  3,
			RejectedAt:              postgres.TimeToPG(ptrTime(time.Unix(1710001200, 0).UTC())),
			IsSupportReviewRequired: true,
		}),
		userProfile: testUserProfile(userID),
	})
	if err != nil {
		t.Fatalf("buildRegistration() rejected error = %v, want nil", err)
	}
	if rejected.Rejection == nil || rejected.Rejection.SelfServeResubmitRemain != 0 {
		t.Fatalf("buildRegistration() rejected Rejection got %#v want remain 0", rejected.Rejection)
	}
	if rejected.Actions.CanResubmit {
		t.Fatal("buildRegistration() rejected CanResubmit = true, want false when remain is exhausted")
	}
	if rejected.Surface.Kind != "read_only_onboarding" {
		t.Fatalf("buildRegistration() rejected surface got %q want %q", rejected.Surface.Kind, "read_only_onboarding")
	}

	eligibleRejectedCapability := testCapability(userID, StateRejected)
	eligibleRejectedCapability.IsResubmitEligible = true
	eligibleRejectedCapability.SelfServeResubmitCount = 1
	eligibleRejected, err := buildRegistration(registrationSnapshot{
		capability:     ptrCapability(eligibleRejectedCapability),
		creatorProfile: ptrCreatorProfile(testCreatorProfile(userID, "bio")),
		intake:         ptrIntake(testIntake(userID)),
		userProfile:    testUserProfile(userID),
	})
	if err != nil {
		t.Fatalf("buildRegistration() eligible rejected error = %v, want nil", err)
	}
	if !eligibleRejected.Actions.CanResubmit {
		t.Fatal("buildRegistration() eligible rejected CanResubmit = false, want true")
	}

	if _, err := buildRegistration(registrationSnapshot{}); err == nil {
		t.Fatal("buildRegistration() error = nil, want error without capability")
	}

	if err := ensureCapabilityEditable(&sqlc.AppCreatorCapability{State: StateSubmitted}); !errors.Is(err, ErrRegistrationStateConflict) {
		t.Fatalf("ensureCapabilityEditable() error got %v want %v", err, ErrRegistrationStateConflict)
	}
	if err := ensureCapabilityEditable(&sqlc.AppCreatorCapability{
		State:                  StateRejected,
		IsResubmitEligible:     true,
		SelfServeResubmitCount: 1,
	}); err != nil {
		t.Fatalf("ensureCapabilityEditable() eligible rejected error = %v, want nil", err)
	}

	existing := testCapability(userID, StateDraft)
	gotCapability, err := ensureDraftCapability(context.Background(), repositoryQueriesStub{}, userID, &existing)
	if err != nil {
		t.Fatalf("ensureDraftCapability(existing) error = %v, want nil", err)
	}
	if gotCapability.State != StateDraft {
		t.Fatalf("ensureDraftCapability(existing) state got %q want %q", gotCapability.State, StateDraft)
	}

	if _, err := normalizeHandle(" @bad-handle "); !errors.Is(err, ErrInvalidHandle) {
		t.Fatalf("normalizeHandle() error got %v want %v", err, ErrInvalidHandle)
	}
	normalizedHandle, err := normalizeHandle(" @Creator.Name_01 ")
	if err != nil {
		t.Fatalf("normalizeHandle() error = %v, want nil", err)
	}
	if normalizedHandle != "creator.name_01" {
		t.Fatalf("normalizeHandle() got %q want %q", normalizedHandle, "creator.name_01")
	}

	normalizedInput, err := normalizeSaveIntakeInput(SaveIntakeInput{
		AcceptsAdultBusinessCompliance:          true,
		AcceptsAppearanceVerification:           true,
		AcceptsConsentResponsibility:            true,
		AcceptsCoPerformerConsentResponsibility: true,
		BirthDate:                               "2000-02-03",
		ConfirmsInformationMatchesDocuments:     true,
		CreatorBio:                              "  bio  ",
		DeclaresNoProhibitedCategory:            true,
		HasCoPerformers:                         true,
		IdentityDocumentType:                    " driver_license ",
		LegalAddress:                            " Tokyo-to Shibuya-ku 1-2-3 ",
		LegalName:                               "  Legal Name ",
		PayoutRecipientName:                     " Recipient ",
		PayoutRecipientType:                     PayoutRecipientTypeSelf,
		TargetAudienceCategory:                  " general_adult ",
		UserID:                                  userID,
	})
	if err != nil {
		t.Fatalf("normalizeSaveIntakeInput() error = %v, want nil", err)
	}
	if normalizedInput.creatorBio != "bio" {
		t.Fatalf("normalizeSaveIntakeInput() creatorBio got %q want %q", normalizedInput.creatorBio, "bio")
	}
	if normalizedInput.birthDate == nil || normalizedInput.birthDate.Format("2006-01-02") != "2000-02-03" {
		t.Fatalf("normalizeSaveIntakeInput() birthDate got %v want 2000-02-03", normalizedInput.birthDate)
	}
	if normalizedInput.legalAddress != "Tokyo-to Shibuya-ku 1-2-3" {
		t.Fatalf("normalizeSaveIntakeInput() legalAddress got %q want trimmed address", normalizedInput.legalAddress)
	}
	if normalizedInput.identityDocumentType == nil || *normalizedInput.identityDocumentType != IdentityDocumentTypeDriverLicense {
		t.Fatalf("normalizeSaveIntakeInput() identityDocumentType got %v want %s", normalizedInput.identityDocumentType, IdentityDocumentTypeDriverLicense)
	}
	if normalizedInput.targetAudienceCategory == nil || *normalizedInput.targetAudienceCategory != TargetAudienceCategoryGeneralAdult {
		t.Fatalf("normalizeSaveIntakeInput() targetAudienceCategory got %v want %s", normalizedInput.targetAudienceCategory, TargetAudienceCategoryGeneralAdult)
	}
	if _, err := normalizeSaveIntakeInput(SaveIntakeInput{BirthDate: "bad", UserID: userID}); !errors.Is(err, ErrInvalidBirthDate) {
		t.Fatalf("normalizeSaveIntakeInput() bad birth date got %v want %v", err, ErrInvalidBirthDate)
	}
	if _, err := normalizeSaveIntakeInput(SaveIntakeInput{BirthDate: "2020-01-01", UserID: userID}); !errors.Is(err, ErrInvalidBirthDate) {
		t.Fatalf("normalizeSaveIntakeInput() underage birth date got %v want %v", err, ErrInvalidBirthDate)
	}
	if _, err := normalizeSaveIntakeInput(SaveIntakeInput{BirthDate: "2100-01-01", UserID: userID}); !errors.Is(err, ErrInvalidBirthDate) {
		t.Fatalf("normalizeSaveIntakeInput() future birth date got %v want %v", err, ErrInvalidBirthDate)
	}
	if _, err := normalizeSaveIntakeInput(SaveIntakeInput{PayoutRecipientType: "corp", UserID: userID}); !errors.Is(err, ErrInvalidPayoutRecipientTyp) {
		t.Fatalf("normalizeSaveIntakeInput() bad payout type got %v want %v", err, ErrInvalidPayoutRecipientTyp)
	}
	if _, err := normalizeSaveIntakeInput(SaveIntakeInput{LegalAddress: "   ", UserID: userID}); !errors.Is(err, ErrInvalidLegalAddress) {
		t.Fatalf("normalizeSaveIntakeInput() blank address got %v want %v", err, ErrInvalidLegalAddress)
	}
	if _, err := normalizeSaveIntakeInput(SaveIntakeInput{LegalAddress: strings.Repeat("あ", maxLegalAddressLength+1), UserID: userID}); !errors.Is(err, ErrInvalidLegalAddress) {
		t.Fatalf("normalizeSaveIntakeInput() long address got %v want %v", err, ErrInvalidLegalAddress)
	}
	if _, err := normalizeSaveIntakeInput(SaveIntakeInput{IdentityDocumentType: "library_card", UserID: userID}); !errors.Is(err, ErrInvalidIdentityDocumentType) {
		t.Fatalf("normalizeSaveIntakeInput() bad identity type got %v want %v", err, ErrInvalidIdentityDocumentType)
	}
	if _, err := normalizeSaveIntakeInput(SaveIntakeInput{TargetAudienceCategory: "unknown", UserID: userID}); !errors.Is(err, ErrInvalidTargetAudienceCategory) {
		t.Fatalf("normalizeSaveIntakeInput() bad target category got %v want %v", err, ErrInvalidTargetAudienceCategory)
	}
	if _, err := parseBirthDate(" 2001-03-04 "); err != nil {
		t.Fatalf("parseBirthDate() error = %v, want nil", err)
	}

	dateValue := time.Date(2001, 3, 4, 0, 0, 0, 0, time.UTC)
	if got := dateStringFromPG(dateToPG(&dateValue)); got != "2001-03-04" {
		t.Fatalf("dateStringFromPG() got %q want %q", got, "2001-03-04")
	}
	if got := dateStringFromPG(pgtype.Date{}); got != "" {
		t.Fatalf("dateStringFromPG() got %q want empty", got)
	}
	if got := maxInt32(4, 2); got != 4 {
		t.Fatalf("maxInt32() got %d want %d", got, 4)
	}
	if got := optionalTextOrEmpty(ptrCreatorProfile(testCreatorProfile(userID, "bio")), func(row sqlc.AppCreatorProfile) pgtype.Text { return row.DisplayName }); got != "Creator Display" {
		t.Fatalf("optionalTextOrEmpty() got %q want %q", got, "Creator Display")
	}
	if got := stringOrEmpty(ptrCreatorProfile(testCreatorProfile(userID, "bio")), func(row sqlc.AppCreatorProfile) string { return row.Bio }); got != "bio" {
		t.Fatalf("stringOrEmpty() got %q want %q", got, "bio")
	}
}

func TestSnapshotAndEvidenceHelpers(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	snapshot := registrationSnapshot{
		capability:     ptrCapability(testCapability(userID, StateDraft)),
		creatorProfile: ptrCreatorProfile(testCreatorProfile(userID, "bio")),
		evidences:      testEvidenceRows(userID),
		intake:         ptrIntake(testIntake(userID)),
		userProfile:    testUserProfile(userID),
	}
	if !isSnapshotComplete(snapshot) {
		t.Fatal("isSnapshotComplete() = false, want true")
	}

	missingHandleSnapshot := snapshot
	missingHandleSnapshot.userProfile.Handle = ""
	if isSnapshotComplete(missingHandleSnapshot) {
		t.Fatal("isSnapshotComplete() = true, want false when handle is empty")
	}

	underageSnapshot := snapshot
	underageIntake := *snapshot.intake
	underageBirthDate := dateOnly(time.Now().UTC().AddDate(-minimumCreatorAgeYears, 0, 1))
	underageIntake.BirthDate = dateToPG(&underageBirthDate)
	underageSnapshot.intake = &underageIntake
	if isSnapshotComplete(underageSnapshot) {
		t.Fatal("isSnapshotComplete() = true, want false when birth date is under 18")
	}

	incompleteCases := []struct {
		name   string
		mutate func(*registrationSnapshot)
	}{
		{
			name: "missing legal address",
			mutate: func(candidate *registrationSnapshot) {
				intake := *candidate.intake
				intake.LegalAddress = ""
				candidate.intake = &intake
			},
		},
		{
			name: "missing identity document type",
			mutate: func(candidate *registrationSnapshot) {
				intake := *candidate.intake
				intake.IdentityDocumentType = pgtype.Text{}
				candidate.intake = &intake
			},
		},
		{
			name: "missing target audience category",
			mutate: func(candidate *registrationSnapshot) {
				intake := *candidate.intake
				intake.TargetAudienceCategory = pgtype.Text{}
				candidate.intake = &intake
			},
		},
		{
			name: "missing appearance verification consent",
			mutate: func(candidate *registrationSnapshot) {
				intake := *candidate.intake
				intake.AcceptsAppearanceVerification = false
				candidate.intake = &intake
			},
		},
		{
			name: "missing document match confirmation",
			mutate: func(candidate *registrationSnapshot) {
				intake := *candidate.intake
				intake.ConfirmsInformationMatchesDocuments = false
				candidate.intake = &intake
			},
		},
		{
			name: "business recipient missing business registration evidence",
			mutate: func(candidate *registrationSnapshot) {
				candidate.evidences = []sqlc.AppCreatorRegistrationEvidence{
					testEvidenceRow(userID, EvidenceKindGovernmentID, "review-bucket", "creator-registration/evidence/government-id.png"),
					testEvidenceRow(userID, EvidenceKindIdentitySelfie, "review-bucket", "creator-registration/evidence/identity-selfie.png"),
					testEvidenceRow(userID, EvidenceKindAddressProof, "review-bucket", "creator-registration/evidence/address-proof.pdf"),
					testEvidenceRow(userID, EvidenceKindPayoutProof, "review-bucket", "creator-registration/evidence/payout-proof.pdf"),
				}
			},
		},
		{
			name: "co performers missing consent responsibility",
			mutate: func(candidate *registrationSnapshot) {
				intake := *candidate.intake
				intake.HasCoPerformers = true
				intake.AcceptsCoPerformerConsentResponsibility = false
				candidate.intake = &intake
				candidate.evidences = append(candidate.evidences, testEvidenceRow(userID, EvidenceKindCoPerformerConsent, "review-bucket", "creator-registration/evidence/co-performer-consent.pdf"))
			},
		},
		{
			name: "co performers missing consent evidence",
			mutate: func(candidate *registrationSnapshot) {
				intake := *candidate.intake
				intake.HasCoPerformers = true
				intake.AcceptsCoPerformerConsentResponsibility = true
				candidate.intake = &intake
			},
		},
	}
	for _, tt := range incompleteCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			candidate := snapshot
			intake := *snapshot.intake
			candidate.intake = &intake
			candidate.evidences = append([]sqlc.AppCreatorRegistrationEvidence{}, snapshot.evidences...)
			tt.mutate(&candidate)

			if isSnapshotComplete(candidate) {
				t.Fatalf("isSnapshotComplete() = true, want false for %s", tt.name)
			}
		})
	}

	uploadedAt := time.Unix(1710001800, 0).UTC()
	evidence, err := mapEvidence(sqlc.AppCreatorRegistrationEvidence{
		FileName:      "government-id.png",
		FileSizeBytes: 512,
		Kind:          EvidenceKindGovernmentID,
		MimeType:      "image/png",
		UploadedAt:    postgres.TimeToPG(&uploadedAt),
	})
	if err != nil {
		t.Fatalf("mapEvidence() error = %v, want nil", err)
	}
	if !evidence.UploadedAt.Equal(uploadedAt) {
		t.Fatalf("mapEvidence() UploadedAt got %s want %s", evidence.UploadedAt, uploadedAt)
	}
	if _, err := mapEvidence(sqlc.AppCreatorRegistrationEvidence{}); err == nil {
		t.Fatal("mapEvidence() error = nil, want invalid uploadedAt error")
	}

	if got := findEvidenceStorageObject(testEvidenceRows(userID), EvidenceKindGovernmentID, "review-bucket", "creator-registration/evidence/government-id.png"); got != nil {
		t.Fatalf("findEvidenceStorageObject() got %#v want nil for unchanged object", got)
	}
	replaced := findEvidenceStorageObject(testEvidenceRows(userID), EvidenceKindGovernmentID, "review-bucket", "creator-registration/evidence/replaced.png")
	if replaced == nil || replaced.Key != "creator-registration/evidence/government-id.png" {
		t.Fatalf("findEvidenceStorageObject() got %#v want previous government_id object", replaced)
	}

	mapped := mapEvidenceList([]sqlc.AppCreatorRegistrationEvidence{
		testEvidenceRow(userID, EvidenceKindGovernmentID, "review-bucket", "creator-registration/evidence/government-id.png"),
		{},
	})
	if len(mapped) != 1 {
		t.Fatalf("mapEvidenceList() len got %d want 1", len(mapped))
	}
}

func TestRepositoryUtilityFallbacks(t *testing.T) {
	t.Parallel()

	repo := NewRepository(&pgxpool.Pool{})
	if repo == nil || repo.queries == nil || repo.newQueries == nil {
		t.Fatalf("NewRepository(non-nil) got %#v want initialized repository", repo)
	}
	if got := dateToPG(nil); got.Valid {
		t.Fatalf("dateToPG(nil) got valid=%t want false", got.Valid)
	}
	if got := maxInt32(1, 2); got != 2 {
		t.Fatalf("maxInt32() got %d want %d", got, 2)
	}
	if got := optionalTextOrEmpty[sqlc.AppCreatorProfile](nil, func(sqlc.AppCreatorProfile) pgtype.Text { return pgtype.Text{} }); got != "" {
		t.Fatalf("optionalTextOrEmpty(nil) got %q want empty", got)
	}
	if got := stringOrEmpty[sqlc.AppCreatorProfile](nil, func(sqlc.AppCreatorProfile) string { return "unexpected" }); got != "" {
		t.Fatalf("stringOrEmpty(nil) got %q want empty", got)
	}
	if got := findEvidenceStorageObject([]sqlc.AppCreatorRegistrationEvidence{{
		Kind:          EvidenceKindGovernmentID,
		StorageBucket: "",
		StorageKey:    "",
	}}, EvidenceKindGovernmentID, "review-bucket", "current-key"); got != nil {
		t.Fatalf("findEvidenceStorageObject() got %#v want nil for empty storage object", got)
	}
}

func TestLoadSnapshotAndUpsertDraftProfileErrors(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	repo := newRepository(repositoryQueriesStub{
		getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
			return sqlc.AppUserProfile{}, pgx.ErrNoRows
		},
	})
	if _, err := repo.loadSnapshot(context.Background(), repo.queries, userID, false, false); !errors.Is(err, ErrSharedProfileNotFound) {
		t.Fatalf("loadSnapshot() error got %v want %v", err, ErrSharedProfileNotFound)
	}

	if _, err := upsertDraftProfile(context.Background(), repositoryQueriesStub{}, sqlc.AppUserProfile{Handle: "creator"}, "bio", nil); !errors.Is(err, ErrInvalidDisplayName) {
		t.Fatalf("upsertDraftProfile() blank display name got %v want %v", err, ErrInvalidDisplayName)
	}

	profile := testUserProfile(userID)
	duplicateErr := &pgconn.PgError{Code: "23505", ConstraintName: creatorProfilesHandleUniqueConstraint}
	_, err := upsertDraftProfile(context.Background(), repositoryQueriesStub{
		createCreatorProfile: func(context.Context, sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
			return sqlc.AppCreatorProfile{}, duplicateErr
		},
	}, profile, "bio", nil)
	if !errors.Is(err, ErrHandleAlreadyTaken) {
		t.Fatalf("upsertDraftProfile() duplicate error got %v want %v", err, ErrHandleAlreadyTaken)
	}
}

func TestRepositoryGuardsAndConflicts(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("12121212-1212-1212-1212-121212121212")
	if NewRepository(nil) == nil {
		t.Fatal("NewRepository(nil) = nil, want non-nil")
	}

	if _, err := (&Repository{}).GetRegistration(context.Background(), userID); err == nil {
		t.Fatal("GetRegistration() error = nil, want initialization error")
	}
	if _, err := (&Repository{}).GetIntake(context.Background(), userID); err == nil {
		t.Fatal("GetIntake() error = nil, want initialization error")
	}
	if err := (&Repository{}).PrepareEvidenceUpload(context.Background(), userID); err == nil {
		t.Fatal("PrepareEvidenceUpload() error = nil, want initialization error")
	}
	if _, err := (&Repository{}).SaveEvidence(context.Background(), SaveEvidenceInput{UserID: userID}); err == nil {
		t.Fatal("SaveEvidence() error = nil, want initialization error")
	}
	if _, err := (&Repository{}).SaveIntake(context.Background(), SaveIntakeInput{UserID: userID}); err == nil {
		t.Fatal("SaveIntake() error = nil, want initialization error")
	}
	if _, err := (&Repository{}).Submit(context.Background(), userID); err == nil {
		t.Fatal("Submit() error = nil, want initialization error")
	}

	tx := &repositoryTxStub{}
	conflictRepo := &Repository{
		txBeginner: repositoryTxBeginnerStub{
			begin: func(context.Context) (pgx.Tx, error) { return tx, nil },
		},
		newQueries: func(sqlc.DBTX) queries {
			return repositoryQueriesStub{
				getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
					return testUserProfile(userID), nil
				},
				getCreatorCapabilityByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return testCapability(userID, StateSubmitted), nil
				},
				getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return sqlc.AppCreatorProfile{}, pgx.ErrNoRows
				},
				getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
					return sqlc.AppCreatorRegistrationIntake{}, pgx.ErrNoRows
				},
				listCreatorRegistrationEvidences: func(context.Context, pgtype.UUID) ([]sqlc.AppCreatorRegistrationEvidence, error) {
					return []sqlc.AppCreatorRegistrationEvidence{}, nil
				},
			}
		},
	}
	if err := conflictRepo.PrepareEvidenceUpload(context.Background(), userID); !errors.Is(err, ErrRegistrationStateConflict) {
		t.Fatalf("PrepareEvidenceUpload() error got %v want %v", err, ErrRegistrationStateConflict)
	}
	if !tx.rolledBack {
		t.Fatal("PrepareEvidenceUpload() rolledBack = false, want true")
	}

	incompleteRepo := &Repository{
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
	if _, err := incompleteRepo.Submit(context.Background(), userID); !errors.Is(err, ErrRegistrationIncomplete) {
		t.Fatalf("Submit() error got %v want %v", err, ErrRegistrationIncomplete)
	}
}

func TestUpsertDraftProfileUpdateAndBuildSharedProfileFallback(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("13131313-1313-1313-1313-131313131313")
	profile := testUserProfile(userID)
	existing := testCreatorProfile(userID, "before")
	updated, err := upsertDraftProfile(context.Background(), repositoryQueriesStub{
		updateCreatorProfile: func(_ context.Context, arg sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
			return sqlc.AppCreatorProfile{
				UserID:      arg.UserID,
				DisplayName: arg.DisplayName,
				AvatarUrl:   arg.AvatarUrl,
				Bio:         arg.Bio,
				Handle:      arg.Handle,
			}, nil
		},
	}, profile, "after", &existing)
	if err != nil {
		t.Fatalf("upsertDraftProfile() update error = %v, want nil", err)
	}
	if updated.Bio != "after" {
		t.Fatalf("upsertDraftProfile() update bio got %q want %q", updated.Bio, "after")
	}

	shared := buildSharedProfile(sqlc.AppUserProfile{
		UserID:      pgtype.UUID{Valid: false},
		DisplayName: "Creator Display",
		Handle:      "creator.handle",
	})
	if shared.UserID != uuid.Nil {
		t.Fatalf("buildSharedProfile() UserID got %s want nil UUID", shared.UserID)
	}
}

func TestRepositoryGetRegistrationAndWriteConflicts(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("14141414-1414-1414-1414-141414141414")
	repo := newRepository(repositoryQueriesStub{
		getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
			return testUserProfile(userID), nil
		},
		getCreatorCapabilityByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapability(userID, StateSubmitted), nil
		},
		getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testCreatorProfile(userID, "draft bio"), nil
		},
		getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
			return testIntake(userID), nil
		},
	})

	registration, err := repo.GetRegistration(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetRegistration() error = %v, want nil", err)
	}
	if registration == nil || registration.State != StateSubmitted {
		t.Fatalf("GetRegistration() got %#v want submitted registration", registration)
	}
	if registration.Actions.CanSubmit {
		t.Fatal("GetRegistration() Actions.CanSubmit = true, want false")
	}

	tx := &repositoryTxStub{}
	conflictRepo := &Repository{
		txBeginner: repositoryTxBeginnerStub{
			begin: func(context.Context) (pgx.Tx, error) { return tx, nil },
		},
		newQueries: func(sqlc.DBTX) queries {
			return repositoryQueriesStub{
				getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
					return testUserProfile(userID), nil
				},
				getCreatorCapabilityByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return testCapability(userID, StateSubmitted), nil
				},
				getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return sqlc.AppCreatorProfile{}, pgx.ErrNoRows
				},
				getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
					return sqlc.AppCreatorRegistrationIntake{}, pgx.ErrNoRows
				},
				listCreatorRegistrationEvidences: func(context.Context, pgtype.UUID) ([]sqlc.AppCreatorRegistrationEvidence, error) {
					return []sqlc.AppCreatorRegistrationEvidence{}, nil
				},
			}
		},
	}

	if _, err := conflictRepo.SaveIntake(context.Background(), SaveIntakeInput{
		CreatorBio: "bio",
		UserID:     userID,
	}); !errors.Is(err, ErrRegistrationStateConflict) {
		t.Fatalf("SaveIntake() conflict error got %v want %v", err, ErrRegistrationStateConflict)
	}
	if _, err := conflictRepo.SaveEvidence(context.Background(), SaveEvidenceInput{
		FileName:      "government-id.png",
		FileSizeBytes: 1,
		Kind:          EvidenceKindGovernmentID,
		MimeType:      "image/png",
		StorageBucket: "review-bucket",
		StorageKey:    "tmp/government-id.png",
		UploadedAt:    time.Unix(1710000000, 0).UTC(),
		UserID:        userID,
	}); !errors.Is(err, ErrRegistrationStateConflict) {
		t.Fatalf("SaveEvidence() conflict error got %v want %v", err, ErrRegistrationStateConflict)
	}

	supportOnlyTx := &repositoryTxStub{}
	supportOnlyCapability := testCapability(userID, StateRejected)
	supportOnlyCapability.IsSupportReviewRequired = true
	supportOnlyCapability.SelfServeResubmitCount = 2

	supportOnlyRepo := &Repository{
		txBeginner: repositoryTxBeginnerStub{
			begin: func(context.Context) (pgx.Tx, error) { return supportOnlyTx, nil },
		},
		newQueries: func(sqlc.DBTX) queries {
			return repositoryQueriesStub{
				getUserProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppUserProfile, error) {
					return testUserProfile(userID), nil
				},
				getCreatorCapabilityByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return supportOnlyCapability, nil
				},
				getCreatorProfileByUserID: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return testCreatorProfile(userID, "blocked bio"), nil
				},
				getCreatorRegistrationIntakeByUser: func(context.Context, pgtype.UUID) (sqlc.AppCreatorRegistrationIntake, error) {
					return testIntake(userID), nil
				},
				listCreatorRegistrationEvidences: func(context.Context, pgtype.UUID) ([]sqlc.AppCreatorRegistrationEvidence, error) {
					return testEvidenceRows(userID), nil
				},
			}
		},
	}

	if _, err := supportOnlyRepo.SaveIntake(context.Background(), SaveIntakeInput{
		CreatorBio: "bio",
		UserID:     userID,
	}); !errors.Is(err, ErrRegistrationStateConflict) {
		t.Fatalf("SaveIntake() support-only rejected error got %v want %v", err, ErrRegistrationStateConflict)
	}
}

func testUserProfile(userID uuid.UUID) sqlc.AppUserProfile {
	avatarURL := "https://cdn.example.com/avatar.jpg"
	return sqlc.AppUserProfile{
		UserID:      postgres.UUIDToPG(userID),
		DisplayName: "Creator Display",
		Handle:      "creator.handle",
		AvatarUrl:   postgres.TextToPG(&avatarURL),
	}
}

func testCapability(userID uuid.UUID, state string) sqlc.AppCreatorCapability {
	return sqlc.AppCreatorCapability{
		UserID: postgres.UUIDToPG(userID),
		State:  state,
	}
}

func testCreatorProfile(userID uuid.UUID, bio string) sqlc.AppCreatorProfile {
	displayName := "Creator Display"
	avatarURL := "https://cdn.example.com/avatar.jpg"
	return sqlc.AppCreatorProfile{
		UserID:      postgres.UUIDToPG(userID),
		DisplayName: postgres.TextToPG(&displayName),
		AvatarUrl:   postgres.TextToPG(&avatarURL),
		Bio:         bio,
		Handle:      "creator.handle",
	}
}

func testIntake(userID uuid.UUID) sqlc.AppCreatorRegistrationIntake {
	birthDate := time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)
	payoutType := PayoutRecipientTypeBusiness
	identityDocumentType := IdentityDocumentTypeDriverLicense
	targetAudienceCategory := TargetAudienceCategoryGeneralAdult
	return sqlc.AppCreatorRegistrationIntake{
		UserID:                              postgres.UUIDToPG(userID),
		LegalName:                           "Creator Legal",
		BirthDate:                           dateToPG(&birthDate),
		LegalAddress:                        "Tokyo-to Shibuya-ku 1-2-3",
		IdentityDocumentType:                postgres.TextToPG(&identityDocumentType),
		TargetAudienceCategory:              postgres.TextToPG(&targetAudienceCategory),
		HasCoPerformers:                     false,
		PayoutRecipientType:                 postgres.TextToPG(&payoutType),
		PayoutRecipientName:                 "Creator Biz",
		DeclaresNoProhibitedCategory:        true,
		AcceptsConsentResponsibility:        true,
		AcceptsAppearanceVerification:       true,
		AcceptsAdultBusinessCompliance:      true,
		ConfirmsInformationMatchesDocuments: true,
	}
}

func testEvidenceRows(userID uuid.UUID) []sqlc.AppCreatorRegistrationEvidence {
	return []sqlc.AppCreatorRegistrationEvidence{
		testEvidenceRow(userID, EvidenceKindGovernmentID, "review-bucket", "creator-registration/evidence/government-id.png"),
		testEvidenceRow(userID, EvidenceKindIdentitySelfie, "review-bucket", "creator-registration/evidence/identity-selfie.png"),
		testEvidenceRow(userID, EvidenceKindAddressProof, "review-bucket", "creator-registration/evidence/address-proof.pdf"),
		testEvidenceRow(userID, EvidenceKindPayoutProof, "review-bucket", "creator-registration/evidence/payout-proof.pdf"),
		testEvidenceRow(userID, EvidenceKindBusinessRegistration, "review-bucket", "creator-registration/evidence/business-registration.pdf"),
	}
}

func testEvidenceRow(userID uuid.UUID, kind string, bucket string, key string) sqlc.AppCreatorRegistrationEvidence {
	uploadedAt := time.Unix(1710000000, 0).UTC()
	return sqlc.AppCreatorRegistrationEvidence{
		UserID:        postgres.UUIDToPG(userID),
		Kind:          kind,
		FileName:      kind + ".png",
		MimeType:      "image/png",
		FileSizeBytes: 256,
		StorageBucket: bucket,
		StorageKey:    key,
		UploadedAt:    postgres.TimeToPG(&uploadedAt),
	}
}

func ptrCapability(value sqlc.AppCreatorCapability) *sqlc.AppCreatorCapability {
	return &value
}

func ptrCreatorProfile(value sqlc.AppCreatorProfile) *sqlc.AppCreatorProfile {
	return &value
}

func ptrIntake(value sqlc.AppCreatorRegistrationIntake) *sqlc.AppCreatorRegistrationIntake {
	return &value
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
