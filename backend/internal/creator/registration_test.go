package creator

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestRegisterApprovedCreatorCreatesCapabilityAndProfile(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	displayName := "Mina Rei"
	bio := "close-up shorts"
	tx := &creatorTxStub{}
	beginner := &creatorTxBeginnerStub{tx: tx}

	var gotCapabilityArg sqlc.CreateCreatorCapabilityParams
	var gotProfileArg sqlc.CreateCreatorProfileParams

	repo := &Repository{
		txBeginner: beginner,
		newQueries: func(sqlc.DBTX) queries {
			return repositoryStubQueries{
				getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return sqlc.AppCreatorCapability{}, pgx.ErrNoRows
				},
				createCapability: func(_ context.Context, arg sqlc.CreateCreatorCapabilityParams) (sqlc.AppCreatorCapability, error) {
					gotCapabilityArg = arg
					return sqlc.AppCreatorCapability{
						UserID:                   pgUUID(userID),
						State:                    "approved",
						RejectionReasonCode:      pgText(nil),
						IsResubmitEligible:       false,
						IsSupportReviewRequired:  false,
						SelfServeResubmitCount:   0,
						KycProviderCaseRef:       pgText(nil),
						PayoutProviderAccountRef: pgText(nil),
						SubmittedAt:              pgTime(nil),
						ApprovedAt:               pgTime(timePtr(now)),
						RejectedAt:               pgTime(nil),
						SuspendedAt:              pgTime(nil),
						CreatedAt:                pgTime(timePtr(now)),
						UpdatedAt:                pgTime(timePtr(now.Add(time.Minute))),
					}, nil
				},
				getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return sqlc.AppCreatorProfile{}, pgx.ErrNoRows
				},
				createProfile: func(_ context.Context, arg sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
					gotProfileArg = arg
					return sqlc.AppCreatorProfile{
						UserID:      pgUUID(userID),
						DisplayName: pgText(&displayName),
						Handle:      pgText(nil),
						AvatarUrl:   pgText(nil),
						Bio:         bio,
						PublishedAt: pgTime(nil),
						CreatedAt:   pgTime(timePtr(now)),
						UpdatedAt:   pgTime(timePtr(now.Add(time.Minute))),
					}, nil
				},
			}
		},
	}

	got, err := repo.RegisterApprovedCreator(context.Background(), SelfServeRegistrationInput{
		UserID:      userID,
		DisplayName: "  " + displayName + "  ",
		Bio:         "  " + bio + "  ",
	})
	if err != nil {
		t.Fatalf("RegisterApprovedCreator() error = %v, want nil", err)
	}
	if gotCapabilityArg.UserID != pgUUID(userID) {
		t.Fatalf("RegisterApprovedCreator() capability user arg got %v want %v", gotCapabilityArg.UserID, pgUUID(userID))
	}
	if gotCapabilityArg.State != "approved" {
		t.Fatalf("RegisterApprovedCreator() capability state arg got %q want approved", gotCapabilityArg.State)
	}
	if gotProfileArg.UserID != pgUUID(userID) {
		t.Fatalf("RegisterApprovedCreator() profile user arg got %v want %v", gotProfileArg.UserID, pgUUID(userID))
	}
	if gotProfileArg.DisplayName != pgText(&displayName) {
		t.Fatalf("RegisterApprovedCreator() display name arg got %v want %v", gotProfileArg.DisplayName, pgText(&displayName))
	}
	if gotProfileArg.Bio != bio {
		t.Fatalf("RegisterApprovedCreator() bio arg got %q want %q", gotProfileArg.Bio, bio)
	}
	if got.Profile.Handle != nil {
		t.Fatalf("RegisterApprovedCreator() handle got %v want nil", got.Profile.Handle)
	}
	if got.Profile.AvatarURL != nil {
		t.Fatalf("RegisterApprovedCreator() avatar url got %v want nil", got.Profile.AvatarURL)
	}
	if got.Profile.PublishedAt != nil {
		t.Fatalf("RegisterApprovedCreator() published at got %v want nil", got.Profile.PublishedAt)
	}
	if !beginner.began {
		t.Fatal("RegisterApprovedCreator() began = false, want true")
	}
	if !tx.committed {
		t.Fatal("RegisterApprovedCreator() committed = false, want true")
	}
	if tx.rolledBack {
		t.Fatal("RegisterApprovedCreator() rolledBack = true, want false")
	}
}

func TestRegisterApprovedCreatorUpdatesExistingCapabilityAndProfile(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	existingDisplayName := stringPtr("Old Name")
	existingHandle := stringPtr("oldhandle")
	existingAvatarURL := stringPtr("https://cdn.example.com/avatar.jpg")
	displayName := "New Name"
	bio := "updated bio"
	tx := &creatorTxStub{}
	beginner := &creatorTxBeginnerStub{tx: tx}

	var gotUpdateCapabilityArg sqlc.UpdateCreatorCapabilityStateParams
	var gotUpdateProfileArg sqlc.UpdateCreatorProfileParams

	repo := &Repository{
		txBeginner: beginner,
		newQueries: func(sqlc.DBTX) queries {
			return repositoryStubQueries{
				getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					row := testCapabilityRow(userID, now, stringPtr("needs_fix"), nil, nil, timePtr(now.Add(-time.Hour)), nil, nil, nil)
					row.State = "pending"
					return row, nil
				},
				updateCapability: func(_ context.Context, arg sqlc.UpdateCreatorCapabilityStateParams) (sqlc.AppCreatorCapability, error) {
					gotUpdateCapabilityArg = arg
					return testCapabilityRow(userID, now, nil, nil, nil, timePtr(now.Add(-time.Hour)), timePtr(now), nil, nil), nil
				},
				getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					row := testProfileRow(userID, now, existingDisplayName, existingHandle, existingAvatarURL, nil)
					row.Bio = "old bio"
					return row, nil
				},
				updateProfile: func(_ context.Context, arg sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
					gotUpdateProfileArg = arg
					row := testProfileRow(userID, now, stringPtr(displayName), existingHandle, existingAvatarURL, nil)
					row.Bio = bio
					return row, nil
				},
			}
		},
	}

	got, err := repo.RegisterApprovedCreator(context.Background(), SelfServeRegistrationInput{
		UserID:      userID,
		DisplayName: displayName,
		Bio:         bio,
	})
	if err != nil {
		t.Fatalf("RegisterApprovedCreator() error = %v, want nil", err)
	}
	if gotUpdateCapabilityArg.UserID != pgUUID(userID) {
		t.Fatalf("RegisterApprovedCreator() capability update user arg got %v want %v", gotUpdateCapabilityArg.UserID, pgUUID(userID))
	}
	if gotUpdateCapabilityArg.State != "approved" {
		t.Fatalf("RegisterApprovedCreator() capability update state got %q want approved", gotUpdateCapabilityArg.State)
	}
	if gotUpdateCapabilityArg.RejectionReasonCode != pgText(nil) {
		t.Fatalf("RegisterApprovedCreator() rejection reason got %v want nil", gotUpdateCapabilityArg.RejectionReasonCode)
	}
	if gotUpdateProfileArg.UserID != pgUUID(userID) {
		t.Fatalf("RegisterApprovedCreator() profile update user arg got %v want %v", gotUpdateProfileArg.UserID, pgUUID(userID))
	}
	if gotUpdateProfileArg.DisplayName != pgText(&displayName) {
		t.Fatalf("RegisterApprovedCreator() display name update arg got %v want %v", gotUpdateProfileArg.DisplayName, pgText(&displayName))
	}
	if gotUpdateProfileArg.Handle != pgText(existingHandle) {
		t.Fatalf("RegisterApprovedCreator() handle update arg got %v want %v", gotUpdateProfileArg.Handle, pgText(existingHandle))
	}
	if gotUpdateProfileArg.AvatarUrl != pgText(existingAvatarURL) {
		t.Fatalf("RegisterApprovedCreator() avatar update arg got %v want %v", gotUpdateProfileArg.AvatarUrl, pgText(existingAvatarURL))
	}
	if gotUpdateProfileArg.Bio != bio {
		t.Fatalf("RegisterApprovedCreator() bio update arg got %q want %q", gotUpdateProfileArg.Bio, bio)
	}
	if !reflect.DeepEqual(got.Profile.Handle, existingHandle) {
		t.Fatalf("RegisterApprovedCreator() profile handle got %v want %v", got.Profile.Handle, existingHandle)
	}
	if !reflect.DeepEqual(got.Profile.AvatarURL, existingAvatarURL) {
		t.Fatalf("RegisterApprovedCreator() profile avatar got %v want %v", got.Profile.AvatarURL, existingAvatarURL)
	}
	if !tx.committed {
		t.Fatal("RegisterApprovedCreator() committed = false, want true")
	}
}

func TestRegisterApprovedCreatorRejectsBlankDisplayName(t *testing.T) {
	t.Parallel()

	repo := &Repository{
		txBeginner: &creatorTxBeginnerStub{},
		newQueries: func(sqlc.DBTX) queries {
			return repositoryStubQueries{}
		},
	}

	if _, err := repo.RegisterApprovedCreator(context.Background(), SelfServeRegistrationInput{
		UserID:      uuid.New(),
		DisplayName: "   ",
	}); !errors.Is(err, ErrInvalidDisplayName) {
		t.Fatalf("RegisterApprovedCreator() error got %v want %v", err, ErrInvalidDisplayName)
	}
}

func TestRegisterApprovedCreatorRejectsUninitializedRepository(t *testing.T) {
	t.Parallel()

	var nilRepo *Repository
	if _, err := nilRepo.RegisterApprovedCreator(context.Background(), SelfServeRegistrationInput{
		UserID:      uuid.New(),
		DisplayName: "Mina",
	}); err == nil {
		t.Fatal("RegisterApprovedCreator() nil repo error = nil, want initialization error")
	}

	repoWithoutQueries := &Repository{txBeginner: &creatorTxBeginnerStub{}}
	if _, err := repoWithoutQueries.RegisterApprovedCreator(context.Background(), SelfServeRegistrationInput{
		UserID:      uuid.New(),
		DisplayName: "Mina",
	}); err == nil {
		t.Fatal("RegisterApprovedCreator() nil query factory error = nil, want initialization error")
	}
}

func TestRegisterApprovedCreatorReturnsWrappedErrorWhenCapabilityLookupFails(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	tx := &creatorTxStub{}
	beginner := &creatorTxBeginnerStub{tx: tx}
	expectedErr := errors.New("capability lookup failed")

	repo := &Repository{
		txBeginner: beginner,
		newQueries: func(sqlc.DBTX) queries {
			return repositoryStubQueries{
				getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return sqlc.AppCreatorCapability{}, expectedErr
				},
			}
		},
	}

	_, err := repo.RegisterApprovedCreator(context.Background(), SelfServeRegistrationInput{
		UserID:      userID,
		DisplayName: "Mina",
		Bio:         "bio",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("RegisterApprovedCreator() error got %v want wrapped %v", err, expectedErr)
	}
	if !tx.rolledBack {
		t.Fatal("RegisterApprovedCreator() rolledBack = false, want true")
	}
}

func TestRegisterApprovedCreatorKeepsApprovedCapabilityState(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	tx := &creatorTxStub{}
	beginner := &creatorTxBeginnerStub{tx: tx}
	updateCalled := false

	repo := &Repository{
		txBeginner: beginner,
		newQueries: func(sqlc.DBTX) queries {
			return repositoryStubQueries{
				getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return testCapabilityRow(userID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
				},
				updateCapability: func(context.Context, sqlc.UpdateCreatorCapabilityStateParams) (sqlc.AppCreatorCapability, error) {
					updateCalled = true
					return sqlc.AppCreatorCapability{}, nil
				},
				getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return sqlc.AppCreatorProfile{}, pgx.ErrNoRows
				},
				createProfile: func(context.Context, sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
					return testProfileRow(userID, now, stringPtr("Mina"), nil, nil, nil), nil
				},
			}
		},
	}

	got, err := repo.RegisterApprovedCreator(context.Background(), SelfServeRegistrationInput{
		UserID:      userID,
		DisplayName: "Mina",
		Bio:         "bio",
	})
	if err != nil {
		t.Fatalf("RegisterApprovedCreator() error = %v, want nil", err)
	}
	if updateCalled {
		t.Fatal("RegisterApprovedCreator() update capability called = true, want false")
	}
	if got.Capability.State != "approved" {
		t.Fatalf("RegisterApprovedCreator() capability state got %q want approved", got.Capability.State)
	}
}
