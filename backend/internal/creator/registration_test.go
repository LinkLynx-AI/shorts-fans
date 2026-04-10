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
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestRegisterApprovedCreatorCreatesCapabilityAndProfile(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	displayName := "Mina Rei"
	handle := "minarei"
	bio := "close-up shorts"
	publishedAt := timePtr(now.Add(2 * time.Minute))
	tx := &creatorTxStub{}
	beginner := &creatorTxBeginnerStub{tx: tx}

	var gotCapabilityArg sqlc.CreateCreatorCapabilityParams
	var gotProfileArg sqlc.CreateCreatorProfileParams
	var gotPublishUserID pgtype.UUID

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
						Handle:      handle,
						AvatarUrl:   pgText(nil),
						Bio:         bio,
						PublishedAt: pgTime(nil),
						CreatedAt:   pgTime(timePtr(now)),
						UpdatedAt:   pgTime(timePtr(now.Add(time.Minute))),
					}, nil
				},
				publishProfile: func(_ context.Context, userIDArg pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					gotPublishUserID = userIDArg
					return testProfileRow(userID, now, stringPtr(displayName), stringPtr(handle), nil, publishedAt), nil
				},
			}
		},
	}

	got, err := repo.RegisterApprovedCreator(context.Background(), SelfServeRegistrationInput{
		UserID:      userID,
		DisplayName: "  " + displayName + "  ",
		Handle:      " @" + handle + " ",
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
	if gotProfileArg.Handle != handle {
		t.Fatalf("RegisterApprovedCreator() handle arg got %v want %v", gotProfileArg.Handle, handle)
	}
	if gotProfileArg.PublishedAt != pgTime(nil) {
		t.Fatalf("RegisterApprovedCreator() create published_at arg got %v want %v", gotProfileArg.PublishedAt, pgTime(nil))
	}
	if gotProfileArg.Bio != bio {
		t.Fatalf("RegisterApprovedCreator() bio arg got %q want %q", gotProfileArg.Bio, bio)
	}
	if gotPublishUserID != pgUUID(userID) {
		t.Fatalf("RegisterApprovedCreator() publish user arg got %v want %v", gotPublishUserID, pgUUID(userID))
	}
	if got.Profile.Handle == nil || *got.Profile.Handle != handle {
		t.Fatalf("RegisterApprovedCreator() handle got %v want %v", got.Profile.Handle, handle)
	}
	if got.Profile.AvatarURL != nil {
		t.Fatalf("RegisterApprovedCreator() avatar url got %v want nil", got.Profile.AvatarURL)
	}
	if !reflect.DeepEqual(got.Profile.PublishedAt, publishedAt) {
		t.Fatalf("RegisterApprovedCreator() published at got %v want %v", got.Profile.PublishedAt, publishedAt)
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
	handle := "new_name"
	publishedAt := timePtr(now.Add(3 * time.Hour))
	tx := &creatorTxStub{}
	beginner := &creatorTxBeginnerStub{tx: tx}

	var gotUpdateCapabilityArg sqlc.UpdateCreatorCapabilityStateParams
	var gotUpdateProfileArg sqlc.UpdateCreatorProfileParams
	var gotPublishUserID pgtype.UUID

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
					row := testProfileRow(userID, now, stringPtr(displayName), stringPtr(handle), existingAvatarURL, nil)
					row.Bio = bio
					return row, nil
				},
				publishProfile: func(_ context.Context, userIDArg pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					gotPublishUserID = userIDArg
					row := testProfileRow(userID, now, stringPtr(displayName), stringPtr(handle), existingAvatarURL, publishedAt)
					row.Bio = bio
					return row, nil
				},
			}
		},
	}

	got, err := repo.RegisterApprovedCreator(context.Background(), SelfServeRegistrationInput{
		UserID:      userID,
		DisplayName: displayName,
		Handle:      "@" + handle,
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
	if gotUpdateProfileArg.Handle != handle {
		t.Fatalf("RegisterApprovedCreator() handle update arg got %v want %v", gotUpdateProfileArg.Handle, handle)
	}
	if gotUpdateProfileArg.AvatarUrl != pgText(existingAvatarURL) {
		t.Fatalf("RegisterApprovedCreator() avatar update arg got %v want %v", gotUpdateProfileArg.AvatarUrl, pgText(existingAvatarURL))
	}
	if gotUpdateProfileArg.Bio != bio {
		t.Fatalf("RegisterApprovedCreator() bio update arg got %q want %q", gotUpdateProfileArg.Bio, bio)
	}
	if gotPublishUserID != pgUUID(userID) {
		t.Fatalf("RegisterApprovedCreator() publish user arg got %v want %v", gotPublishUserID, pgUUID(userID))
	}
	if got.Profile.Handle == nil || *got.Profile.Handle != handle {
		t.Fatalf("RegisterApprovedCreator() profile handle got %v want %v", valueOrNil(got.Profile.Handle), handle)
	}
	if !reflect.DeepEqual(got.Profile.AvatarURL, existingAvatarURL) {
		t.Fatalf("RegisterApprovedCreator() profile avatar got %v want %v", got.Profile.AvatarURL, existingAvatarURL)
	}
	if !reflect.DeepEqual(got.Profile.PublishedAt, publishedAt) {
		t.Fatalf("RegisterApprovedCreator() profile published at got %v want %v", got.Profile.PublishedAt, publishedAt)
	}
	if !tx.committed {
		t.Fatal("RegisterApprovedCreator() committed = false, want true")
	}
}

func TestRegisterApprovedCreatorAppliesNewAvatarURLWhenProvided(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.MustParse("26222222-2222-2222-2222-222222222222")
	displayName := "New Name"
	bio := "updated bio"
	handle := "new_name"
	nextAvatarURL := "https://cdn.example.com/avatar-next.png"
	publishedAt := timePtr(now.Add(2 * time.Minute))
	tx := &creatorTxStub{}
	beginner := &creatorTxBeginnerStub{tx: tx}

	var gotCreateProfileArg sqlc.CreateCreatorProfileParams
	var gotUpdateProfileArg sqlc.UpdateCreatorProfileParams

	repo := &Repository{
		txBeginner: beginner,
		newQueries: func(sqlc.DBTX) queries {
			return repositoryStubQueries{
				getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return sqlc.AppCreatorCapability{}, pgx.ErrNoRows
				},
				createCapability: func(context.Context, sqlc.CreateCreatorCapabilityParams) (sqlc.AppCreatorCapability, error) {
					return testCapabilityRow(userID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
				},
				getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return sqlc.AppCreatorProfile{}, pgx.ErrNoRows
				},
				createProfile: func(_ context.Context, arg sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
					gotCreateProfileArg = arg
					return testProfileRow(userID, now, stringPtr(displayName), stringPtr(handle), stringPtr(nextAvatarURL), nil), nil
				},
				updateProfile: func(_ context.Context, arg sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
					gotUpdateProfileArg = arg
					return testProfileRow(userID, now, stringPtr(displayName), stringPtr(handle), stringPtr(nextAvatarURL), nil), nil
				},
				publishProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return testProfileRow(userID, now, stringPtr(displayName), stringPtr(handle), stringPtr(nextAvatarURL), publishedAt), nil
				},
			}
		},
	}

	got, err := repo.RegisterApprovedCreator(context.Background(), SelfServeRegistrationInput{
		AvatarURL:   &nextAvatarURL,
		UserID:      userID,
		DisplayName: displayName,
		Handle:      handle,
		Bio:         bio,
	})
	if err != nil {
		t.Fatalf("RegisterApprovedCreator() error = %v, want nil", err)
	}
	if gotCreateProfileArg.AvatarUrl != pgText(stringPtr(nextAvatarURL)) {
		t.Fatalf("RegisterApprovedCreator() create avatar arg got %v want %v", gotCreateProfileArg.AvatarUrl, pgText(stringPtr(nextAvatarURL)))
	}
	if !reflect.DeepEqual(got.Profile.AvatarURL, stringPtr(nextAvatarURL)) {
		t.Fatalf("RegisterApprovedCreator() profile avatar got %v want %v", got.Profile.AvatarURL, stringPtr(nextAvatarURL))
	}
	if !reflect.DeepEqual(got.Profile.PublishedAt, publishedAt) {
		t.Fatalf("RegisterApprovedCreator() profile published at got %v want %v", got.Profile.PublishedAt, publishedAt)
	}
	if gotUpdateProfileArg.UserID.Valid {
		t.Fatalf("RegisterApprovedCreator() update profile arg got %#v want create path only", gotUpdateProfileArg)
	}
}

func TestResolveUpdatedAvatarURL(t *testing.T) {
	t.Parallel()

	existingAvatarURL := stringPtr("https://cdn.example.com/existing.png")
	nextAvatarURL := stringPtr("https://cdn.example.com/next.png")

	if got := resolveUpdatedAvatarURL(existingAvatarURL, nil); !reflect.DeepEqual(got, existingAvatarURL) {
		t.Fatalf("resolveUpdatedAvatarURL() got %v want %v", got, existingAvatarURL)
	}
	if got := resolveUpdatedAvatarURL(existingAvatarURL, nextAvatarURL); !reflect.DeepEqual(got, nextAvatarURL) {
		t.Fatalf("resolveUpdatedAvatarURL() got %v want %v", got, nextAvatarURL)
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
		Handle:      "mina",
	}); !errors.Is(err, ErrInvalidDisplayName) {
		t.Fatalf("RegisterApprovedCreator() error got %v want %v", err, ErrInvalidDisplayName)
	}
}

func TestRegisterApprovedCreatorRejectsInvalidHandle(t *testing.T) {
	t.Parallel()

	repo := &Repository{
		txBeginner: &creatorTxBeginnerStub{},
		newQueries: func(sqlc.DBTX) queries {
			return repositoryStubQueries{}
		},
	}

	if _, err := repo.RegisterApprovedCreator(context.Background(), SelfServeRegistrationInput{
		UserID:      uuid.New(),
		DisplayName: "Mina",
		Handle:      "@",
	}); !errors.Is(err, ErrInvalidHandle) {
		t.Fatalf("RegisterApprovedCreator() error got %v want %v", err, ErrInvalidHandle)
	}
}

func TestRegisterApprovedCreatorRejectsUninitializedRepository(t *testing.T) {
	t.Parallel()

	var nilRepo *Repository
	if _, err := nilRepo.RegisterApprovedCreator(context.Background(), SelfServeRegistrationInput{
		UserID:      uuid.New(),
		DisplayName: "Mina",
		Handle:      "mina",
	}); err == nil {
		t.Fatal("RegisterApprovedCreator() nil repo error = nil, want initialization error")
	}

	repoWithoutQueries := &Repository{txBeginner: &creatorTxBeginnerStub{}}
	if _, err := repoWithoutQueries.RegisterApprovedCreator(context.Background(), SelfServeRegistrationInput{
		UserID:      uuid.New(),
		DisplayName: "Mina",
		Handle:      "mina",
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
		Handle:      "mina",
		Bio:         "bio",
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("RegisterApprovedCreator() error got %v want wrapped %v", err, expectedErr)
	}
	if !tx.rolledBack {
		t.Fatal("RegisterApprovedCreator() rolledBack = false, want true")
	}
}

func TestRegisterApprovedCreatorReturnsWrappedErrorWhenPublishFails(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("38333333-3333-3333-3333-333333333333")
	tx := &creatorTxStub{}
	beginner := &creatorTxBeginnerStub{tx: tx}
	expectedErr := errors.New("publish failed")

	repo := &Repository{
		txBeginner: beginner,
		newQueries: func(sqlc.DBTX) queries {
			return repositoryStubQueries{
				getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return sqlc.AppCreatorCapability{}, pgx.ErrNoRows
				},
				createCapability: func(context.Context, sqlc.CreateCreatorCapabilityParams) (sqlc.AppCreatorCapability, error) {
					return testCapabilityRow(userID, time.Unix(1710000000, 0).UTC(), nil, nil, nil, nil, timePtr(time.Unix(1710000000, 0).UTC()), nil, nil), nil
				},
				getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return sqlc.AppCreatorProfile{}, pgx.ErrNoRows
				},
				createProfile: func(context.Context, sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
					return testProfileRow(userID, time.Unix(1710000000, 0).UTC(), stringPtr("Mina"), stringPtr("mina"), nil, nil), nil
				},
				publishProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return sqlc.AppCreatorProfile{}, expectedErr
				},
			}
		},
	}

	_, err := repo.RegisterApprovedCreator(context.Background(), SelfServeRegistrationInput{
		UserID:      userID,
		DisplayName: "Mina",
		Handle:      "mina",
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
	publishedAt := timePtr(now.Add(2 * time.Minute))

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
					return testProfileRow(userID, now, stringPtr("Mina"), stringPtr("mina"), nil, nil), nil
				},
				publishProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return testProfileRow(userID, now, stringPtr("Mina"), stringPtr("mina"), nil, publishedAt), nil
				},
			}
		},
	}

	got, err := repo.RegisterApprovedCreator(context.Background(), SelfServeRegistrationInput{
		UserID:      userID,
		DisplayName: "Mina",
		Handle:      "mina",
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
	if !reflect.DeepEqual(got.Profile.PublishedAt, publishedAt) {
		t.Fatalf("RegisterApprovedCreator() profile published at got %v want %v", got.Profile.PublishedAt, publishedAt)
	}
}

func TestRegisterApprovedCreatorMapsDuplicateHandle(t *testing.T) {
	t.Parallel()

	userID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	tx := &creatorTxStub{}
	beginner := &creatorTxBeginnerStub{tx: tx}

	repo := &Repository{
		txBeginner: beginner,
		newQueries: func(sqlc.DBTX) queries {
			return repositoryStubQueries{
				getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
					return sqlc.AppCreatorCapability{}, pgx.ErrNoRows
				},
				createCapability: func(context.Context, sqlc.CreateCreatorCapabilityParams) (sqlc.AppCreatorCapability, error) {
					return testCapabilityRow(userID, time.Unix(1710000000, 0).UTC(), nil, nil, nil, nil, timePtr(time.Unix(1710000000, 0).UTC()), nil, nil), nil
				},
				getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
					return sqlc.AppCreatorProfile{}, pgx.ErrNoRows
				},
				createProfile: func(context.Context, sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
					return sqlc.AppCreatorProfile{}, &pgconn.PgError{Code: "23505", ConstraintName: creatorProfilesHandleUniqueConstraint}
				},
			}
		},
	}

	_, err := repo.RegisterApprovedCreator(context.Background(), SelfServeRegistrationInput{
		UserID:      userID,
		DisplayName: "Mina",
		Handle:      "mina",
		Bio:         "bio",
	})
	if !errors.Is(err, ErrHandleAlreadyTaken) {
		t.Fatalf("RegisterApprovedCreator() error got %v want %v", err, ErrHandleAlreadyTaken)
	}
}

func valueOrNil(value *string) any {
	if value == nil {
		return nil
	}

	return *value
}
