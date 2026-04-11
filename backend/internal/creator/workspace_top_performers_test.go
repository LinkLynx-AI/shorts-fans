package creator

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestGetWorkspaceTopPerformers(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("11111111-3333-3333-3333-333333333333")
	mainInvalidID := uuid.MustParse("21111111-3333-3333-3333-333333333333")
	mainTopID := uuid.MustParse("22222222-3333-3333-3333-333333333333")
	leadShortInvalidID := uuid.MustParse("31111111-3333-3333-3333-333333333333")
	leadShortTopID := uuid.MustParse("32222222-3333-3333-3333-333333333333")
	shortInvalidID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortTopID := uuid.MustParse("34444444-3333-3333-3333-333333333333")
	mainInvalidAssetID := uuid.MustParse("41111111-3333-3333-3333-333333333333")
	mainTopAssetID := uuid.MustParse("42222222-3333-3333-3333-333333333333")
	shortLeadInvalidAssetID := uuid.MustParse("43333333-3333-3333-3333-333333333333")
	shortLeadTopAssetID := uuid.MustParse("44444444-3333-3333-3333-333333333333")
	shortInvalidAssetID := uuid.MustParse("45555555-3333-3333-3333-333333333333")
	shortTopAssetID := uuid.MustParse("46666666-3333-3333-3333-333333333333")

	delivery := newWorkspacePreviewDelivery(t)
	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		listShortsByCreator: func(_ context.Context, gotUserID pgtype.UUID) ([]sqlc.AppShort, error) {
			if gotUserID != pgUUID(viewerUserID) {
				t.Fatalf("ListShortsByCreatorUserID() user got %v want %v", gotUserID, pgUUID(viewerUserID))
			}

			return []sqlc.AppShort{
				testWorkspacePreviewShortRow(leadShortInvalidID, viewerUserID, mainInvalidID, shortLeadInvalidAssetID, now.Add(6*time.Minute), "approved_for_publish"),
				testWorkspacePreviewShortRow(leadShortTopID, viewerUserID, mainTopID, shortLeadTopAssetID, now.Add(5*time.Minute), "approved_for_publish"),
				testWorkspacePreviewShortRow(shortInvalidID, viewerUserID, mainTopID, shortInvalidAssetID, now.Add(4*time.Minute), "approved_for_publish"),
				testWorkspacePreviewShortRow(shortTopID, viewerUserID, mainTopID, shortTopAssetID, now.Add(3*time.Minute), "approved_for_publish"),
			}, nil
		},
		listWorkspaceTopMainCandidates: func(_ context.Context, gotUserID pgtype.UUID) ([]sqlc.ListCreatorWorkspaceTopMainCandidatesByCreatorUserIDRow, error) {
			if gotUserID != pgUUID(viewerUserID) {
				t.Fatalf("ListCreatorWorkspaceTopMainCandidatesByCreatorUserID() user got %v want %v", gotUserID, pgUUID(viewerUserID))
			}

			return []sqlc.ListCreatorWorkspaceTopMainCandidatesByCreatorUserIDRow{
				testWorkspaceTopMainCandidateRow(mainInvalidID, mainInvalidAssetID, now.Add(2*time.Minute), 2400, workspacePreviewCurrencyCode, 300),
				testWorkspaceTopMainCandidateRow(mainTopID, mainTopAssetID, now.Add(time.Minute), 1800, workspacePreviewCurrencyCode, 238),
			}, nil
		},
		listWorkspaceTopShortCandidates: func(_ context.Context, gotUserID pgtype.UUID) ([]sqlc.ListCreatorWorkspaceTopShortCandidatesByCreatorUserIDRow, error) {
			if gotUserID != pgUUID(viewerUserID) {
				t.Fatalf("ListCreatorWorkspaceTopShortCandidatesByCreatorUserID() user got %v want %v", gotUserID, pgUUID(viewerUserID))
			}

			return []sqlc.ListCreatorWorkspaceTopShortCandidatesByCreatorUserIDRow{
				testWorkspaceTopShortCandidateRow(shortInvalidID, mainTopID, shortInvalidAssetID, now.Add(4*time.Minute), 300),
				testWorkspaceTopShortCandidateRow(shortTopID, mainTopID, shortTopAssetID, now.Add(3*time.Minute), 238),
			}, nil
		},
		getMediaAsset: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			switch id {
			case pgUUID(mainInvalidAssetID):
				return testWorkspacePreviewAssetRow(mainInvalidAssetID, viewerUserID, now, int64Ptr(720000), "uploaded"), nil
			case pgUUID(mainTopAssetID):
				return testWorkspacePreviewAssetRow(mainTopAssetID, viewerUserID, now, int64Ptr(600000), workspacePreviewAssetReadyState), nil
			case pgUUID(shortLeadInvalidAssetID):
				return testWorkspacePreviewAssetRow(shortLeadInvalidAssetID, viewerUserID, now, int64Ptr(15000), workspacePreviewAssetReadyState), nil
			case pgUUID(shortLeadTopAssetID):
				return testWorkspacePreviewAssetRow(shortLeadTopAssetID, viewerUserID, now, int64Ptr(16000), workspacePreviewAssetReadyState), nil
			case pgUUID(shortInvalidAssetID):
				return testWorkspacePreviewAssetRow(shortInvalidAssetID, viewerUserID, now, int64Ptr(14000), "uploaded"), nil
			case pgUUID(shortTopAssetID):
				return testWorkspacePreviewAssetRow(shortTopAssetID, viewerUserID, now, int64Ptr(17000), workspacePreviewAssetReadyState), nil
			default:
				t.Fatalf("GetMediaAssetByID() id got %v", id)
				return sqlc.AppMediaAsset{}, nil
			}
		},
	})
	repo.delivery = delivery

	got, err := repo.GetWorkspaceTopPerformers(context.Background(), viewerUserID)
	if err != nil {
		t.Fatalf("GetWorkspaceTopPerformers() error = %v, want nil", err)
	}
	if got.TopMain == nil {
		t.Fatal("GetWorkspaceTopPerformers() TopMain = nil, want non-nil")
	}
	if got.TopMain.ID != mainTopID {
		t.Fatalf("GetWorkspaceTopPerformers() TopMain.ID got %s want %s", got.TopMain.ID, mainTopID)
	}
	if got.TopMain.UnlockCount != 238 {
		t.Fatalf("GetWorkspaceTopPerformers() TopMain.UnlockCount got %d want %d", got.TopMain.UnlockCount, 238)
	}
	if got.TopShort == nil {
		t.Fatal("GetWorkspaceTopPerformers() TopShort = nil, want non-nil")
	}
	if got.TopShort.ID != shortTopID {
		t.Fatalf("GetWorkspaceTopPerformers() TopShort.ID got %s want %s", got.TopShort.ID, shortTopID)
	}
	if got.TopShort.AttributedUnlockCount != 238 {
		t.Fatalf("GetWorkspaceTopPerformers() TopShort.AttributedUnlockCount got %d want %d", got.TopShort.AttributedUnlockCount, 238)
	}
}

func TestGetWorkspaceTopPerformersReturnsEmptyWhenNoPreviewableCandidatesRemain(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("51111111-3333-3333-3333-333333333333")
	mainID := uuid.MustParse("61111111-3333-3333-3333-333333333333")
	shortID := uuid.MustParse("71111111-3333-3333-3333-333333333333")
	mainAssetID := uuid.MustParse("81111111-3333-3333-3333-333333333333")
	shortAssetID := uuid.MustParse("91111111-3333-3333-3333-333333333333")

	delivery := newWorkspacePreviewDelivery(t)
	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		listShortsByCreator: func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error) {
			return []sqlc.AppShort{
				testWorkspacePreviewShortRow(shortID, viewerUserID, mainID, shortAssetID, now, "approved_for_publish"),
			}, nil
		},
		listWorkspaceTopMainCandidates: func(context.Context, pgtype.UUID) ([]sqlc.ListCreatorWorkspaceTopMainCandidatesByCreatorUserIDRow, error) {
			return []sqlc.ListCreatorWorkspaceTopMainCandidatesByCreatorUserIDRow{
				testWorkspaceTopMainCandidateRow(mainID, mainAssetID, now, 1800, workspacePreviewCurrencyCode, 12),
			}, nil
		},
		listWorkspaceTopShortCandidates: func(context.Context, pgtype.UUID) ([]sqlc.ListCreatorWorkspaceTopShortCandidatesByCreatorUserIDRow, error) {
			return []sqlc.ListCreatorWorkspaceTopShortCandidatesByCreatorUserIDRow{
				testWorkspaceTopShortCandidateRow(shortID, mainID, shortAssetID, now, 12),
			}, nil
		},
		getMediaAsset: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			switch id {
			case pgUUID(mainAssetID):
				return testWorkspacePreviewAssetRow(mainAssetID, viewerUserID, now, int64Ptr(720000), "uploaded"), nil
			case pgUUID(shortAssetID):
				return testWorkspacePreviewAssetRow(shortAssetID, viewerUserID, now, int64Ptr(15000), "uploaded"), nil
			default:
				t.Fatalf("GetMediaAssetByID() id got %v", id)
				return sqlc.AppMediaAsset{}, nil
			}
		},
	})
	repo.delivery = delivery

	got, err := repo.GetWorkspaceTopPerformers(context.Background(), viewerUserID)
	if err != nil {
		t.Fatalf("GetWorkspaceTopPerformers() error = %v, want nil", err)
	}
	if got.TopMain != nil {
		t.Fatalf("GetWorkspaceTopPerformers() TopMain got %#v want nil", got.TopMain)
	}
	if got.TopShort != nil {
		t.Fatalf("GetWorkspaceTopPerformers() TopShort got %#v want nil", got.TopShort)
	}
}

func TestGetWorkspaceTopPerformersReturnsTopMainWithoutLinkedShort(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("a1111111-3333-3333-3333-333333333333")
	mainID := uuid.MustParse("a2222222-3333-3333-3333-333333333333")
	mainAssetID := uuid.MustParse("a3333333-3333-3333-3333-333333333333")

	delivery := newWorkspacePreviewDelivery(t)
	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		listWorkspaceTopMainCandidates: func(_ context.Context, gotUserID pgtype.UUID) ([]sqlc.ListCreatorWorkspaceTopMainCandidatesByCreatorUserIDRow, error) {
			if gotUserID != pgUUID(viewerUserID) {
				t.Fatalf("ListCreatorWorkspaceTopMainCandidatesByCreatorUserID() user got %v want %v", gotUserID, pgUUID(viewerUserID))
			}

			return []sqlc.ListCreatorWorkspaceTopMainCandidatesByCreatorUserIDRow{
				testWorkspaceTopMainCandidateRow(mainID, mainAssetID, now, 1800, workspacePreviewCurrencyCode, 238),
			}, nil
		},
		listWorkspaceTopShortCandidates: func(context.Context, pgtype.UUID) ([]sqlc.ListCreatorWorkspaceTopShortCandidatesByCreatorUserIDRow, error) {
			return nil, nil
		},
		getMediaAsset: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			switch id {
			case pgUUID(mainAssetID):
				return testWorkspacePreviewAssetRow(mainAssetID, viewerUserID, now, int64Ptr(600000), workspacePreviewAssetReadyState), nil
			default:
				t.Fatalf("GetMediaAssetByID() id got %v", id)
				return sqlc.AppMediaAsset{}, nil
			}
		},
	})
	repo.delivery = delivery

	got, err := repo.GetWorkspaceTopPerformers(context.Background(), viewerUserID)
	if err != nil {
		t.Fatalf("GetWorkspaceTopPerformers() error = %v, want nil", err)
	}
	if got.TopMain == nil {
		t.Fatal("GetWorkspaceTopPerformers() TopMain = nil, want non-nil")
	}
	if got.TopMain.ID != mainID {
		t.Fatalf("GetWorkspaceTopPerformers() TopMain.ID got %s want %s", got.TopMain.ID, mainID)
	}
	if got.TopMain.UnlockCount != 238 {
		t.Fatalf("GetWorkspaceTopPerformers() TopMain.UnlockCount got %d want %d", got.TopMain.UnlockCount, 238)
	}
	if got.TopShort != nil {
		t.Fatalf("GetWorkspaceTopPerformers() TopShort got %#v want nil", got.TopShort)
	}
}

func TestGetWorkspaceTopPerformersReturnsErrorWhenDeliveryIsNil(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("b1111111-3333-3333-3333-333333333333")

	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
	})

	_, err := repo.GetWorkspaceTopPerformers(context.Background(), viewerUserID)
	if err == nil {
		t.Fatal("GetWorkspaceTopPerformers() error = nil, want non-nil")
	}
	if got, want := err.Error(), "delivery is nil"; !containsString(got, want) {
		t.Fatalf("GetWorkspaceTopPerformers() error got %q want substring %q", got, want)
	}
}

func TestGetWorkspaceTopPerformersReturnsErrorWhenTopMainCandidatesQueryFails(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("b2222222-3333-3333-3333-333333333333")
	wantErr := errors.New("top main query failed")

	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		listWorkspaceTopMainCandidates: func(context.Context, pgtype.UUID) ([]sqlc.ListCreatorWorkspaceTopMainCandidatesByCreatorUserIDRow, error) {
			return nil, wantErr
		},
	})
	repo.delivery = newWorkspacePreviewDelivery(t)

	_, err := repo.GetWorkspaceTopPerformers(context.Background(), viewerUserID)
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetWorkspaceTopPerformers() error got %v want %v", err, wantErr)
	}
	if got, want := err.Error(), "top main 候補読み込み"; !containsString(got, want) {
		t.Fatalf("GetWorkspaceTopPerformers() error got %q want substring %q", got, want)
	}
}

func TestGetWorkspaceTopPerformersReturnsErrorWhenTopShortCandidatesQueryFails(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("b3333333-3333-3333-3333-333333333333")
	wantErr := errors.New("top short query failed")

	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		listWorkspaceTopMainCandidates: func(context.Context, pgtype.UUID) ([]sqlc.ListCreatorWorkspaceTopMainCandidatesByCreatorUserIDRow, error) {
			return nil, nil
		},
		listWorkspaceTopShortCandidates: func(context.Context, pgtype.UUID) ([]sqlc.ListCreatorWorkspaceTopShortCandidatesByCreatorUserIDRow, error) {
			return nil, wantErr
		},
	})
	repo.delivery = newWorkspacePreviewDelivery(t)

	_, err := repo.GetWorkspaceTopPerformers(context.Background(), viewerUserID)
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetWorkspaceTopPerformers() error got %v want %v", err, wantErr)
	}
	if got, want := err.Error(), "top short 候補読み込み"; !containsString(got, want) {
		t.Fatalf("GetWorkspaceTopPerformers() error got %q want substring %q", got, want)
	}
}

func containsString(got string, want string) bool {
	return strings.Contains(got, want)
}

func testWorkspaceTopMainCandidateRow(
	id uuid.UUID,
	mediaAssetID uuid.UUID,
	createdAt time.Time,
	priceMinor int64,
	currencyCode string,
	unlockCount int64,
) sqlc.ListCreatorWorkspaceTopMainCandidatesByCreatorUserIDRow {
	return sqlc.ListCreatorWorkspaceTopMainCandidatesByCreatorUserIDRow{
		ID:           postgres.UUIDToPG(id),
		MediaAssetID: postgres.UUIDToPG(mediaAssetID),
		PriceMinor:   priceMinor,
		CurrencyCode: currencyCode,
		CreatedAt:    postgres.TimeToPG(&createdAt),
		UnlockCount:  unlockCount,
	}
}

func testWorkspaceTopShortCandidateRow(
	id uuid.UUID,
	canonicalMainID uuid.UUID,
	mediaAssetID uuid.UUID,
	createdAt time.Time,
	attributedUnlockCount int64,
) sqlc.ListCreatorWorkspaceTopShortCandidatesByCreatorUserIDRow {
	return sqlc.ListCreatorWorkspaceTopShortCandidatesByCreatorUserIDRow{
		ID:                    postgres.UUIDToPG(id),
		CanonicalMainID:       postgres.UUIDToPG(canonicalMainID),
		MediaAssetID:          postgres.UUIDToPG(mediaAssetID),
		CreatedAt:             postgres.TimeToPG(&createdAt),
		AttributedUnlockCount: attributedUnlockCount,
	}
}
