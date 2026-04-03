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

type repositoryStubQueries struct {
	createCapability func(context.Context, sqlc.CreateCreatorCapabilityParams) (sqlc.AppCreatorCapability, error)
	getCapability    func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error)
	updateCapability func(context.Context, sqlc.UpdateCreatorCapabilityStateParams) (sqlc.AppCreatorCapability, error)
	createProfile    func(context.Context, sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error)
	getProfile       func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error)
	getPublicProfile func(context.Context, pgtype.UUID) (sqlc.AppPublicCreatorProfile, error)
	updateProfile    func(context.Context, sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error)
	publishProfile   func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error)
}

func (s repositoryStubQueries) CreateCreatorCapability(ctx context.Context, arg sqlc.CreateCreatorCapabilityParams) (sqlc.AppCreatorCapability, error) {
	if s.createCapability == nil {
		return sqlc.AppCreatorCapability{}, nil
	}
	return s.createCapability(ctx, arg)
}

func (s repositoryStubQueries) GetCreatorCapabilityByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
	if s.getCapability == nil {
		return sqlc.AppCreatorCapability{}, nil
	}
	return s.getCapability(ctx, userID)
}

func (s repositoryStubQueries) UpdateCreatorCapabilityState(ctx context.Context, arg sqlc.UpdateCreatorCapabilityStateParams) (sqlc.AppCreatorCapability, error) {
	if s.updateCapability == nil {
		return sqlc.AppCreatorCapability{}, nil
	}
	return s.updateCapability(ctx, arg)
}

func (s repositoryStubQueries) CreateCreatorProfile(ctx context.Context, arg sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
	if s.createProfile == nil {
		return sqlc.AppCreatorProfile{}, nil
	}
	return s.createProfile(ctx, arg)
}

func (s repositoryStubQueries) GetCreatorProfileByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorProfile, error) {
	if s.getProfile == nil {
		return sqlc.AppCreatorProfile{}, nil
	}
	return s.getProfile(ctx, userID)
}

func (s repositoryStubQueries) GetPublicCreatorProfileByUserID(ctx context.Context, userID pgtype.UUID) (sqlc.AppPublicCreatorProfile, error) {
	if s.getPublicProfile == nil {
		return sqlc.AppPublicCreatorProfile{}, nil
	}
	return s.getPublicProfile(ctx, userID)
}

func (s repositoryStubQueries) UpdateCreatorProfile(ctx context.Context, arg sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
	if s.updateProfile == nil {
		return sqlc.AppCreatorProfile{}, nil
	}
	return s.updateProfile(ctx, arg)
}

func (s repositoryStubQueries) PublishCreatorProfile(ctx context.Context, userID pgtype.UUID) (sqlc.AppCreatorProfile, error) {
	if s.publishProfile == nil {
		return sqlc.AppCreatorProfile{}, nil
	}
	return s.publishProfile(ctx, userID)
}

func TestRepositorySuccessPaths(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.New()
	displayName := stringPtr("alice")
	avatarURL := stringPtr("https://cdn.example.com/avatar.jpg")
	rejectionReason := stringPtr("needs_review")
	kycRef := stringPtr("kyc-1")
	payoutRef := stringPtr("payout-1")
	submittedAt := timePtr(now.Add(time.Hour))
	approvedAt := timePtr(now.Add(2 * time.Hour))
	rejectedAt := timePtr(now.Add(3 * time.Hour))
	suspendedAt := timePtr(now.Add(4 * time.Hour))
	publishedAt := timePtr(now.Add(5 * time.Hour))

	capabilityRow := testCapabilityRow(userID, now, rejectionReason, kycRef, payoutRef, submittedAt, approvedAt, rejectedAt, suspendedAt)
	profileRow := testProfileRow(userID, now, displayName, avatarURL, publishedAt)
	publicProfileRow := testPublicProfileRow(userID, now, displayName, avatarURL, publishedAt)

	var createCapabilityArg sqlc.CreateCreatorCapabilityParams
	var updateCapabilityArg sqlc.UpdateCreatorCapabilityStateParams
	var createProfileArg sqlc.CreateCreatorProfileParams
	var updateProfileArg sqlc.UpdateCreatorProfileParams

	repo := newRepository(repositoryStubQueries{
		createCapability: func(_ context.Context, arg sqlc.CreateCreatorCapabilityParams) (sqlc.AppCreatorCapability, error) {
			createCapabilityArg = arg
			return capabilityRow, nil
		},
		getCapability: func(_ context.Context, gotUserID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			if gotUserID != pgUUID(userID) {
				t.Fatalf("GetCreatorCapabilityByUserID() user got %v want %v", gotUserID, pgUUID(userID))
			}
			return capabilityRow, nil
		},
		updateCapability: func(_ context.Context, arg sqlc.UpdateCreatorCapabilityStateParams) (sqlc.AppCreatorCapability, error) {
			updateCapabilityArg = arg
			return capabilityRow, nil
		},
		createProfile: func(_ context.Context, arg sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
			createProfileArg = arg
			return profileRow, nil
		},
		getProfile: func(_ context.Context, gotUserID pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			if gotUserID != pgUUID(userID) {
				t.Fatalf("GetCreatorProfileByUserID() user got %v want %v", gotUserID, pgUUID(userID))
			}
			return profileRow, nil
		},
		getPublicProfile: func(_ context.Context, gotUserID pgtype.UUID) (sqlc.AppPublicCreatorProfile, error) {
			if gotUserID != pgUUID(userID) {
				t.Fatalf("GetPublicCreatorProfileByUserID() user got %v want %v", gotUserID, pgUUID(userID))
			}
			return publicProfileRow, nil
		},
		updateProfile: func(_ context.Context, arg sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
			updateProfileArg = arg
			return profileRow, nil
		},
		publishProfile: func(_ context.Context, gotUserID pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			if gotUserID != pgUUID(userID) {
				t.Fatalf("PublishCreatorProfile() user got %v want %v", gotUserID, pgUUID(userID))
			}
			return profileRow, nil
		},
	})

	capabilityInput := CreateCapabilityInput{
		UserID:                   userID,
		State:                    "approved",
		RejectionReasonCode:      rejectionReason,
		IsResubmitEligible:       true,
		IsSupportReviewRequired:  true,
		SelfServeResubmitCount:   2,
		KYCProviderCaseRef:       kycRef,
		PayoutProviderAccountRef: payoutRef,
		SubmittedAt:              submittedAt,
		ApprovedAt:               approvedAt,
		RejectedAt:               rejectedAt,
		SuspendedAt:              suspendedAt,
	}

	createdCapability, err := repo.CreateCapability(context.Background(), capabilityInput)
	if err != nil {
		t.Fatalf("CreateCapability() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(createdCapability, wantCapability(userID, now, rejectionReason, kycRef, payoutRef, submittedAt, approvedAt, rejectedAt, suspendedAt)) {
		t.Fatalf("CreateCapability() got %#v want %#v", createdCapability, wantCapability(userID, now, rejectionReason, kycRef, payoutRef, submittedAt, approvedAt, rejectedAt, suspendedAt))
	}
	if createCapabilityArg.UserID != pgUUID(userID) {
		t.Fatalf("CreateCapability() user arg got %v want %v", createCapabilityArg.UserID, pgUUID(userID))
	}

	gotCapability, err := repo.GetCapability(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetCapability() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(gotCapability, createdCapability) {
		t.Fatalf("GetCapability() got %#v want %#v", gotCapability, createdCapability)
	}

	updatedCapability, err := repo.UpdateCapability(context.Background(), UpdateCapabilityInput(capabilityInput))
	if err != nil {
		t.Fatalf("UpdateCapability() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(updatedCapability, createdCapability) {
		t.Fatalf("UpdateCapability() got %#v want %#v", updatedCapability, createdCapability)
	}
	if updateCapabilityArg.UserID != pgUUID(userID) {
		t.Fatalf("UpdateCapability() user arg got %v want %v", updateCapabilityArg.UserID, pgUUID(userID))
	}

	profileInput := CreateProfileInput{
		UserID:      userID,
		DisplayName: displayName,
		AvatarURL:   avatarURL,
		Bio:         "bio",
		PublishedAt: publishedAt,
	}

	createdProfile, err := repo.CreateProfile(context.Background(), profileInput)
	if err != nil {
		t.Fatalf("CreateProfile() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(createdProfile, wantProfile(userID, now, displayName, avatarURL, publishedAt)) {
		t.Fatalf("CreateProfile() got %#v want %#v", createdProfile, wantProfile(userID, now, displayName, avatarURL, publishedAt))
	}
	if createProfileArg.UserID != pgUUID(userID) {
		t.Fatalf("CreateProfile() user arg got %v want %v", createProfileArg.UserID, pgUUID(userID))
	}

	gotProfile, err := repo.GetProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetProfile() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(gotProfile, createdProfile) {
		t.Fatalf("GetProfile() got %#v want %#v", gotProfile, createdProfile)
	}

	gotPublicProfile, err := repo.GetPublicProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetPublicProfile() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(gotPublicProfile, createdProfile) {
		t.Fatalf("GetPublicProfile() got %#v want %#v", gotPublicProfile, createdProfile)
	}

	updatedProfile, err := repo.UpdateProfile(context.Background(), UpdateProfileInput{
		UserID:      userID,
		DisplayName: displayName,
		AvatarURL:   avatarURL,
		Bio:         "bio",
	})
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(updatedProfile, createdProfile) {
		t.Fatalf("UpdateProfile() got %#v want %#v", updatedProfile, createdProfile)
	}
	if updateProfileArg.UserID != pgUUID(userID) {
		t.Fatalf("UpdateProfile() user arg got %v want %v", updateProfileArg.UserID, pgUUID(userID))
	}

	publishedProfile, err := repo.PublishProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("PublishProfile() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(publishedProfile, createdProfile) {
		t.Fatalf("PublishProfile() got %#v want %#v", publishedProfile, createdProfile)
	}
}

func TestRepositoryErrorPaths(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	genericErr := errors.New("query failed")

	repo := newRepository(repositoryStubQueries{
		createCapability: func(context.Context, sqlc.CreateCreatorCapabilityParams) (sqlc.AppCreatorCapability, error) {
			return sqlc.AppCreatorCapability{}, genericErr
		},
		updateCapability: func(context.Context, sqlc.UpdateCreatorCapabilityStateParams) (sqlc.AppCreatorCapability, error) {
			return sqlc.AppCreatorCapability{}, pgx.ErrNoRows
		},
		createProfile: func(context.Context, sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
			return sqlc.AppCreatorProfile{}, genericErr
		},
		getPublicProfile: func(context.Context, pgtype.UUID) (sqlc.AppPublicCreatorProfile, error) {
			return sqlc.AppPublicCreatorProfile{}, pgx.ErrNoRows
		},
		updateProfile: func(context.Context, sqlc.UpdateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
			return sqlc.AppCreatorProfile{}, pgx.ErrNoRows
		},
		publishProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return sqlc.AppCreatorProfile{}, pgx.ErrNoRows
		},
	})

	if _, err := repo.CreateCapability(context.Background(), CreateCapabilityInput{}); !errors.Is(err, genericErr) {
		t.Fatalf("CreateCapability() error got %v want wrapped %v", err, genericErr)
	}
	if _, err := repo.UpdateCapability(context.Background(), UpdateCapabilityInput{UserID: userID}); !errors.Is(err, ErrCapabilityNotFound) {
		t.Fatalf("UpdateCapability() error got %v want %v", err, ErrCapabilityNotFound)
	}
	if _, err := repo.CreateProfile(context.Background(), CreateProfileInput{}); !errors.Is(err, genericErr) {
		t.Fatalf("CreateProfile() error got %v want wrapped %v", err, genericErr)
	}
	if _, err := repo.GetPublicProfile(context.Background(), userID); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("GetPublicProfile() error got %v want %v", err, ErrProfileNotFound)
	}
	if _, err := repo.UpdateProfile(context.Background(), UpdateProfileInput{UserID: userID}); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("UpdateProfile() error got %v want %v", err, ErrProfileNotFound)
	}
	if _, err := repo.PublishProfile(context.Background(), userID); !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("PublishProfile() error got %v want %v", err, ErrProfileNotFound)
	}
}

func TestRepositoryConversionErrors(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	now := time.Unix(1710000000, 0).UTC()
	capabilityRow := testCapabilityRow(userID, now, nil, nil, nil, nil, nil, nil, nil)
	capabilityRow.UserID = pgtype.UUID{}

	profileRow := testProfileRow(userID, now, nil, nil, nil)
	profileRow.UserID = pgtype.UUID{}

	publicProfileRow := testPublicProfileRow(userID, now, nil, nil, nil)
	publicProfileRow.UserID = pgtype.UUID{}

	repo := newRepository(repositoryStubQueries{
		createCapability: func(context.Context, sqlc.CreateCreatorCapabilityParams) (sqlc.AppCreatorCapability, error) {
			return capabilityRow, nil
		},
		createProfile: func(context.Context, sqlc.CreateCreatorProfileParams) (sqlc.AppCreatorProfile, error) {
			return profileRow, nil
		},
		getPublicProfile: func(context.Context, pgtype.UUID) (sqlc.AppPublicCreatorProfile, error) {
			return publicProfileRow, nil
		},
	})

	if _, err := repo.CreateCapability(context.Background(), CreateCapabilityInput{}); err == nil {
		t.Fatal("CreateCapability() error = nil, want conversion error")
	}
	if _, err := repo.CreateProfile(context.Background(), CreateProfileInput{}); err == nil {
		t.Fatal("CreateProfile() error = nil, want conversion error")
	}
	if _, err := repo.GetPublicProfile(context.Background(), userID); err == nil {
		t.Fatal("GetPublicProfile() error = nil, want conversion error")
	}
}

func TestMapFunctionsRejectInvalidRows(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	userID := uuid.New()

	invalidCapability := testCapabilityRow(userID, now, nil, nil, nil, nil, nil, nil, nil)
	invalidCapability.CreatedAt = pgtype.Timestamptz{}
	if _, err := mapCapability(invalidCapability); err == nil {
		t.Fatal("mapCapability() error = nil, want conversion error")
	}

	invalidProfile := testProfileRow(userID, now, nil, nil, nil)
	invalidProfile.CreatedAt = pgtype.Timestamptz{}
	if _, err := mapProfile(invalidProfile); err == nil {
		t.Fatal("mapProfile() error = nil, want conversion error")
	}

	invalidPublicProfile := testPublicProfileRow(userID, now, nil, nil, nil)
	invalidPublicProfile.CreatedAt = pgtype.Timestamptz{}
	if _, err := mapPublicProfile(invalidPublicProfile); err == nil {
		t.Fatal("mapPublicProfile() error = nil, want conversion error")
	}
}

func testCapabilityRow(userID uuid.UUID, now time.Time, rejectionReason *string, kycRef *string, payoutRef *string, submittedAt *time.Time, approvedAt *time.Time, rejectedAt *time.Time, suspendedAt *time.Time) sqlc.AppCreatorCapability {
	return sqlc.AppCreatorCapability{
		UserID:                   pgUUID(userID),
		State:                    "approved",
		RejectionReasonCode:      pgText(rejectionReason),
		IsResubmitEligible:       true,
		IsSupportReviewRequired:  true,
		SelfServeResubmitCount:   2,
		KycProviderCaseRef:       pgText(kycRef),
		PayoutProviderAccountRef: pgText(payoutRef),
		SubmittedAt:              pgTime(submittedAt),
		ApprovedAt:               pgTime(approvedAt),
		RejectedAt:               pgTime(rejectedAt),
		SuspendedAt:              pgTime(suspendedAt),
		CreatedAt:                pgTime(timePtr(now)),
		UpdatedAt:                pgTime(timePtr(now.Add(time.Minute))),
	}
}

func testProfileRow(userID uuid.UUID, now time.Time, displayName *string, avatarURL *string, publishedAt *time.Time) sqlc.AppCreatorProfile {
	return sqlc.AppCreatorProfile{
		UserID:      pgUUID(userID),
		DisplayName: pgText(displayName),
		AvatarUrl:   pgText(avatarURL),
		Bio:         "bio",
		PublishedAt: pgTime(publishedAt),
		CreatedAt:   pgTime(timePtr(now)),
		UpdatedAt:   pgTime(timePtr(now.Add(time.Minute))),
	}
}

func testPublicProfileRow(userID uuid.UUID, now time.Time, displayName *string, avatarURL *string, publishedAt *time.Time) sqlc.AppPublicCreatorProfile {
	return sqlc.AppPublicCreatorProfile{
		UserID:      pgUUID(userID),
		DisplayName: pgText(displayName),
		AvatarUrl:   pgText(avatarURL),
		Bio:         "bio",
		PublishedAt: pgTime(publishedAt),
		CreatedAt:   pgTime(timePtr(now)),
		UpdatedAt:   pgTime(timePtr(now.Add(time.Minute))),
	}
}

func wantCapability(userID uuid.UUID, now time.Time, rejectionReason *string, kycRef *string, payoutRef *string, submittedAt *time.Time, approvedAt *time.Time, rejectedAt *time.Time, suspendedAt *time.Time) Capability {
	return Capability{
		UserID:                   userID,
		State:                    "approved",
		RejectionReasonCode:      rejectionReason,
		IsResubmitEligible:       true,
		IsSupportReviewRequired:  true,
		SelfServeResubmitCount:   2,
		KYCProviderCaseRef:       kycRef,
		PayoutProviderAccountRef: payoutRef,
		SubmittedAt:              submittedAt,
		ApprovedAt:               approvedAt,
		RejectedAt:               rejectedAt,
		SuspendedAt:              suspendedAt,
		CreatedAt:                now,
		UpdatedAt:                now.Add(time.Minute),
	}
}

func wantProfile(userID uuid.UUID, now time.Time, displayName *string, avatarURL *string, publishedAt *time.Time) Profile {
	return Profile{
		UserID:      userID,
		DisplayName: displayName,
		AvatarURL:   avatarURL,
		Bio:         "bio",
		PublishedAt: publishedAt,
		CreatedAt:   now,
		UpdatedAt:   now.Add(time.Minute),
	}
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

func pgText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func pgTime(value *time.Time) pgtype.Timestamptz {
	if value == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *value, Valid: true}
}

func stringPtr(value string) *string {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}
