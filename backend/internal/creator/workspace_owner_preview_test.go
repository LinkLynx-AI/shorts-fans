package creator

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/media"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type workspacePreviewSignerStub struct{}

func (workspacePreviewSignerStub) PresignGetObject(_ context.Context, bucket string, key string, expires time.Duration) (string, error) {
	return fmt.Sprintf("https://signed.example.com/%s/%s?expires=%d", bucket, key, int(expires.Seconds())), nil
}

func TestListWorkspacePreviewShorts(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("21111111-1111-1111-1111-111111111111")
	shortAID := uuid.MustParse("31111111-1111-1111-1111-111111111111")
	shortBID := uuid.MustParse("32222222-2222-2222-2222-222222222222")
	shortCID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	shortDID := uuid.MustParse("34444444-4444-4444-4444-444444444444")
	assetAID := uuid.MustParse("41111111-1111-1111-1111-111111111111")
	assetBID := uuid.MustParse("42222222-2222-2222-2222-222222222222")
	assetCID := uuid.MustParse("43333333-3333-3333-3333-333333333333")
	assetDID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	delivery := newWorkspacePreviewDelivery(t)
	repo := newRepository(repositoryStubQueries{
		getCapability: func(_ context.Context, gotUserID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			if gotUserID != pgUUID(viewerUserID) {
				t.Fatalf("GetCapability() user got %v want %v", gotUserID, pgUUID(viewerUserID))
			}
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(_ context.Context, gotUserID pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			if gotUserID != pgUUID(viewerUserID) {
				t.Fatalf("GetProfile() user got %v want %v", gotUserID, pgUUID(viewerUserID))
			}
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		listShortsByCreator: func(_ context.Context, gotUserID pgtype.UUID) ([]sqlc.AppShort, error) {
			if gotUserID != pgUUID(viewerUserID) {
				t.Fatalf("ListShortsByCreatorUserID() user got %v want %v", gotUserID, pgUUID(viewerUserID))
			}

			return []sqlc.AppShort{
				testWorkspacePreviewShortRow(shortAID, viewerUserID, mainID, assetAID, now.Add(4*time.Minute), "approved_for_publish"),
				testWorkspacePreviewShortRow(shortBID, viewerUserID, mainID, assetBID, now.Add(3*time.Minute), "draft"),
				testWorkspacePreviewShortRow(shortCID, viewerUserID, mainID, assetCID, now.Add(2*time.Minute), "approved_for_publish"),
				testWorkspacePreviewShortRow(shortDID, viewerUserID, mainID, assetDID, now.Add(time.Minute), "approved_for_publish"),
			}, nil
		},
		getMediaAsset: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			switch id {
			case pgUUID(assetAID):
				return testWorkspacePreviewAssetRow(assetAID, viewerUserID, now, int64Ptr(15000), workspacePreviewAssetReadyState), nil
			case pgUUID(assetBID):
				return testWorkspacePreviewAssetRow(assetBID, viewerUserID, now, int64Ptr(16000), workspacePreviewAssetReadyState), nil
			case pgUUID(assetCID):
				return testWorkspacePreviewAssetRow(assetCID, viewerUserID, now, int64Ptr(17000), "uploaded"), nil
			case pgUUID(assetDID):
				return testWorkspacePreviewAssetRow(assetDID, viewerUserID, now, int64Ptr(18000), workspacePreviewAssetReadyState), nil
			default:
				t.Fatalf("GetMediaAssetByID() id got %v", id)
				return sqlc.AppMediaAsset{}, nil
			}
		},
	})
	repo.delivery = delivery

	firstPage, nextCursor, err := repo.ListWorkspacePreviewShorts(context.Background(), viewerUserID, nil, 1)
	if err != nil {
		t.Fatalf("ListWorkspacePreviewShorts() first page error = %v, want nil", err)
	}
	if len(firstPage) != 1 {
		t.Fatalf("ListWorkspacePreviewShorts() first page len got %d want %d", len(firstPage), 1)
	}
	if firstPage[0].ID != shortAID {
		t.Fatalf("ListWorkspacePreviewShorts() first page id got %s want %s", firstPage[0].ID, shortAID)
	}
	if nextCursor == nil {
		t.Fatal("ListWorkspacePreviewShorts() nextCursor = nil, want non-nil")
	}

	secondPage, secondCursor, err := repo.ListWorkspacePreviewShorts(context.Background(), viewerUserID, nextCursor, 1)
	if err != nil {
		t.Fatalf("ListWorkspacePreviewShorts() second page error = %v, want nil", err)
	}
	if len(secondPage) != 1 {
		t.Fatalf("ListWorkspacePreviewShorts() second page len got %d want %d", len(secondPage), 1)
	}
	if secondPage[0].ID != shortBID {
		t.Fatalf("ListWorkspacePreviewShorts() second page id got %s want %s", secondPage[0].ID, shortBID)
	}
	if secondPage[0].PreviewDurationSeconds != 16 {
		t.Fatalf("ListWorkspacePreviewShorts() second page duration got %d want %d", secondPage[0].PreviewDurationSeconds, 16)
	}
	if secondCursor == nil {
		t.Fatal("ListWorkspacePreviewShorts() second page cursor = nil, want non-nil")
	}

	thirdPage, thirdCursor, err := repo.ListWorkspacePreviewShorts(context.Background(), viewerUserID, secondCursor, 1)
	if err != nil {
		t.Fatalf("ListWorkspacePreviewShorts() third page error = %v, want nil", err)
	}
	if len(thirdPage) != 1 {
		t.Fatalf("ListWorkspacePreviewShorts() third page len got %d want %d", len(thirdPage), 1)
	}
	if thirdPage[0].ID != shortDID {
		t.Fatalf("ListWorkspacePreviewShorts() third page id got %s want %s", thirdPage[0].ID, shortDID)
	}
	if thirdCursor != nil {
		t.Fatalf("ListWorkspacePreviewShorts() third page cursor got %#v want nil", thirdCursor)
	}
}

func TestListWorkspacePreviewMains(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("51111111-1111-1111-1111-111111111111")
	mainAID := uuid.MustParse("61111111-1111-1111-1111-111111111111")
	mainBID := uuid.MustParse("62222222-2222-2222-2222-222222222222")
	mainCID := uuid.MustParse("63333333-3333-3333-3333-333333333333")
	shortLeadNewID := uuid.MustParse("71111111-1111-1111-1111-111111111111")
	shortLeadOldID := uuid.MustParse("72222222-2222-2222-2222-222222222222")
	shortPendingID := uuid.MustParse("73333333-3333-3333-3333-333333333333")
	mainAssetAID := uuid.MustParse("81111111-1111-1111-1111-111111111111")
	mainAssetBID := uuid.MustParse("82222222-2222-2222-2222-222222222222")
	mainAssetCID := uuid.MustParse("83333333-3333-3333-3333-333333333333")
	shortAssetNewID := uuid.MustParse("84444444-4444-4444-4444-444444444444")
	shortAssetOldID := uuid.MustParse("85555555-5555-5555-5555-555555555555")
	shortAssetPendingID := uuid.MustParse("86666666-6666-6666-6666-666666666666")

	delivery := newWorkspacePreviewDelivery(t)
	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		listWorkspacePreviewMainsByCreator: func(_ context.Context, gotUserID pgtype.UUID) ([]sqlc.ListCreatorWorkspacePreviewMainsByCreatorUserIDRow, error) {
			if gotUserID != pgUUID(viewerUserID) {
				t.Fatalf("ListCreatorWorkspacePreviewMainsByCreatorUserID() user got %v want %v", gotUserID, pgUUID(viewerUserID))
			}

			return []sqlc.ListCreatorWorkspacePreviewMainsByCreatorUserIDRow{
				testWorkspacePreviewMainRow(mainAID, viewerUserID, mainAssetAID, now.Add(3*time.Minute), "approved_for_unlock", 2200, workspacePreviewCurrencyCode),
				testWorkspacePreviewMainRow(mainBID, viewerUserID, mainAssetBID, now.Add(2*time.Minute), "revision_requested", 1800, workspacePreviewCurrencyCode),
				testWorkspacePreviewMainRow(mainCID, viewerUserID, mainAssetCID, now.Add(time.Minute), "draft", 1500, workspacePreviewCurrencyCode),
			}, nil
		},
		listShortsByCreator: func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error) {
			return []sqlc.AppShort{
				testWorkspacePreviewShortRow(shortLeadNewID, viewerUserID, mainAID, shortAssetNewID, now.Add(5*time.Minute), "approved_for_publish"),
				testWorkspacePreviewShortRow(shortLeadOldID, viewerUserID, mainAID, shortAssetOldID, now.Add(4*time.Minute), "approved_for_publish"),
				testWorkspacePreviewShortRow(shortPendingID, viewerUserID, mainBID, shortAssetPendingID, now.Add(2*time.Minute), "draft"),
			}, nil
		},
		getMediaAsset: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			switch id {
			case pgUUID(mainAssetAID):
				return testWorkspacePreviewAssetRow(mainAssetAID, viewerUserID, now, int64Ptr(720000), workspacePreviewAssetReadyState), nil
			case pgUUID(mainAssetBID):
				return testWorkspacePreviewAssetRow(mainAssetBID, viewerUserID, now, int64Ptr(600000), workspacePreviewAssetReadyState), nil
			case pgUUID(mainAssetCID):
				return testWorkspacePreviewAssetRow(mainAssetCID, viewerUserID, now, int64Ptr(540000), workspacePreviewAssetReadyState), nil
			case pgUUID(shortAssetNewID):
				return testWorkspacePreviewAssetRow(shortAssetNewID, viewerUserID, now, int64Ptr(15000), workspacePreviewAssetReadyState), nil
			case pgUUID(shortAssetOldID):
				return testWorkspacePreviewAssetRow(shortAssetOldID, viewerUserID, now, int64Ptr(14000), workspacePreviewAssetReadyState), nil
			case pgUUID(shortAssetPendingID):
				return testWorkspacePreviewAssetRow(shortAssetPendingID, viewerUserID, now, int64Ptr(12000), workspacePreviewAssetReadyState), nil
			default:
				t.Fatalf("GetMediaAssetByID() id got %v", id)
				return sqlc.AppMediaAsset{}, nil
			}
		},
	})
	repo.delivery = delivery

	items, nextCursor, err := repo.ListWorkspacePreviewMains(context.Background(), viewerUserID, nil, 10)
	if err != nil {
		t.Fatalf("ListWorkspacePreviewMains() error = %v, want nil", err)
	}
	if len(items) != 2 {
		t.Fatalf("ListWorkspacePreviewMains() len got %d want %d", len(items), 2)
	}
	if items[0].ID != mainAID {
		t.Fatalf("ListWorkspacePreviewMains() id got %s want %s", items[0].ID, mainAID)
	}
	if items[0].LeadShortID != shortLeadNewID {
		t.Fatalf("ListWorkspacePreviewMains() lead short got %s want %s", items[0].LeadShortID, shortLeadNewID)
	}
	if items[0].PriceJpy != 2200 {
		t.Fatalf("ListWorkspacePreviewMains() price got %d want %d", items[0].PriceJpy, 2200)
	}
	if items[1].ID != mainBID {
		t.Fatalf("ListWorkspacePreviewMains() second id got %s want %s", items[1].ID, mainBID)
	}
	if items[1].LeadShortID != shortPendingID {
		t.Fatalf("ListWorkspacePreviewMains() second lead short got %s want %s", items[1].LeadShortID, shortPendingID)
	}
	if items[1].PriceJpy != 1800 {
		t.Fatalf("ListWorkspacePreviewMains() second price got %d want %d", items[1].PriceJpy, 1800)
	}
	if nextCursor != nil {
		t.Fatalf("ListWorkspacePreviewMains() nextCursor got %#v want nil", nextCursor)
	}
}

func TestListWorkspacePreviewMainsSkipsUnexpectedCurrency(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("91111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("92222222-2222-2222-2222-222222222222")
	shortID := uuid.MustParse("93333333-3333-3333-3333-333333333333")
	mainAssetID := uuid.MustParse("94444444-4444-4444-4444-444444444444")
	shortAssetID := uuid.MustParse("95555555-5555-5555-5555-555555555555")

	delivery := newWorkspacePreviewDelivery(t)
	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		listWorkspacePreviewMainsByCreator: func(context.Context, pgtype.UUID) ([]sqlc.ListCreatorWorkspacePreviewMainsByCreatorUserIDRow, error) {
			return []sqlc.ListCreatorWorkspacePreviewMainsByCreatorUserIDRow{
				testWorkspacePreviewMainRow(mainID, viewerUserID, mainAssetID, now, "approved_for_unlock", 2200, "USD"),
			}, nil
		},
		listShortsByCreator: func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error) {
			return []sqlc.AppShort{
				testWorkspacePreviewShortRow(shortID, viewerUserID, mainID, shortAssetID, now.Add(time.Minute), "approved_for_publish"),
			}, nil
		},
		getMediaAsset: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			switch id {
			case pgUUID(mainAssetID):
				return testWorkspacePreviewAssetRow(mainAssetID, viewerUserID, now, int64Ptr(720000), workspacePreviewAssetReadyState), nil
			case pgUUID(shortAssetID):
				return testWorkspacePreviewAssetRow(shortAssetID, viewerUserID, now, int64Ptr(15000), workspacePreviewAssetReadyState), nil
			default:
				t.Fatalf("GetMediaAssetByID() id got %v", id)
				return sqlc.AppMediaAsset{}, nil
			}
		},
	})
	repo.delivery = delivery

	items, nextCursor, err := repo.ListWorkspacePreviewMains(context.Background(), viewerUserID, nil, 10)
	if err != nil {
		t.Fatalf("ListWorkspacePreviewMains() error = %v, want nil", err)
	}
	if len(items) != 0 {
		t.Fatalf("ListWorkspacePreviewMains() len got %d want %d", len(items), 0)
	}
	if nextCursor != nil {
		t.Fatalf("ListWorkspacePreviewMains() nextCursor got %#v want nil", nextCursor)
	}
}

func TestGetWorkspacePreviewShortDetail(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("0b0b0b0b-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("0c0c0c0c-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("0d0d0d0d-1111-1111-1111-111111111111")
	assetID := uuid.MustParse("0e0e0e0e-1111-1111-1111-111111111111")

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
				testWorkspacePreviewShortRowWithCaption(shortID, viewerUserID, mainID, assetID, now, "approved_for_publish", stringPtr("quiet rooftop preview.")),
			}, nil
		},
		getMediaAsset: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			if id != pgUUID(assetID) {
				t.Fatalf("GetMediaAssetByID() id got %v want %v", id, pgUUID(assetID))
			}
			return testWorkspacePreviewAssetRow(assetID, viewerUserID, now, int64Ptr(16000), workspacePreviewAssetReadyState), nil
		},
	})
	repo.delivery = delivery

	got, err := repo.GetWorkspacePreviewShortDetail(context.Background(), viewerUserID, shortID)
	if err != nil {
		t.Fatalf("GetWorkspacePreviewShortDetail() error = %v, want nil", err)
	}
	if got.Creator.UserID != viewerUserID {
		t.Fatalf("GetWorkspacePreviewShortDetail() creator got %s want %s", got.Creator.UserID, viewerUserID)
	}
	if got.Short.ID != shortID {
		t.Fatalf("GetWorkspacePreviewShortDetail() short id got %s want %s", got.Short.ID, shortID)
	}
	if got.Short.CanonicalMainID != mainID {
		t.Fatalf("GetWorkspacePreviewShortDetail() canonical main id got %s want %s", got.Short.CanonicalMainID, mainID)
	}
	if got.Short.Title != "quiet rooftop preview" {
		t.Fatalf("GetWorkspacePreviewShortDetail() title got %q want %q", got.Short.Title, "quiet rooftop preview")
	}
	if got.Short.Media.URL == "" {
		t.Fatal("GetWorkspacePreviewShortDetail() media url = empty, want playback url")
	}
}

func TestGetWorkspacePreviewShortDetailReturnsNotFoundWhenShortMissing(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("3b3b3b3b-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("3c3c3c3c-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("3d3d3d3d-1111-1111-1111-111111111111")
	otherShortID := uuid.MustParse("3e3e3e3e-1111-1111-1111-111111111111")
	assetID := uuid.MustParse("3f3f3f3f-1111-1111-1111-111111111111")

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
				testWorkspacePreviewShortRowWithCaption(otherShortID, viewerUserID, mainID, assetID, now, "approved_for_publish", stringPtr("quiet rooftop preview.")),
			}, nil
		},
		getMediaAsset: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			if id != pgUUID(assetID) {
				t.Fatalf("GetMediaAssetByID() id got %v want %v", id, pgUUID(assetID))
			}
			return testWorkspacePreviewAssetRow(assetID, viewerUserID, now, int64Ptr(16000), workspacePreviewAssetReadyState), nil
		},
	})
	repo.delivery = delivery

	_, err := repo.GetWorkspacePreviewShortDetail(context.Background(), viewerUserID, shortID)
	if !errors.Is(err, ErrWorkspacePreviewNotFound) {
		t.Fatalf("GetWorkspacePreviewShortDetail() error got %v want %v", err, ErrWorkspacePreviewNotFound)
	}
}

func TestGetWorkspacePreviewShortDetailReturnsNotFoundWhenShortIsNotPreviewable(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("5b5b5b5b-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("5c5c5c5c-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("5d5d5d5d-1111-1111-1111-111111111111")
	assetID := uuid.MustParse("5e5e5e5e-1111-1111-1111-111111111111")

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
				testWorkspacePreviewShortRowWithCaption(shortID, viewerUserID, mainID, assetID, now, "approved_for_publish", stringPtr("quiet rooftop preview.")),
			}, nil
		},
		getMediaAsset: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			if id != pgUUID(assetID) {
				t.Fatalf("GetMediaAssetByID() id got %v want %v", id, pgUUID(assetID))
			}
			return testWorkspacePreviewAssetRow(assetID, viewerUserID, now, int64Ptr(16000), "uploaded"), nil
		},
	})
	repo.delivery = delivery

	_, err := repo.GetWorkspacePreviewShortDetail(context.Background(), viewerUserID, shortID)
	if !errors.Is(err, ErrWorkspacePreviewNotFound) {
		t.Fatalf("GetWorkspacePreviewShortDetail() error got %v want %v", err, ErrWorkspacePreviewNotFound)
	}
}

func TestGetWorkspacePreviewMainDetail(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("1b1b1b1b-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("1c1c1c1c-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("1d1d1d1d-1111-1111-1111-111111111111")
	mainAssetID := uuid.MustParse("1e1e1e1e-1111-1111-1111-111111111111")
	shortAssetID := uuid.MustParse("1f1f1f1f-1111-1111-1111-111111111111")

	delivery := newWorkspacePreviewDelivery(t)
	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		listMainsByCreator: func(context.Context, pgtype.UUID) ([]sqlc.AppMain, error) {
			return []sqlc.AppMain{
				testWorkspacePreviewMainRecord(mainID, viewerUserID, mainAssetID, now, "approved_for_unlock", 1800, workspacePreviewCurrencyCode),
			}, nil
		},
		listShortsByCreator: func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error) {
			return []sqlc.AppShort{
				testWorkspacePreviewShortRowWithCaption(shortID, viewerUserID, mainID, shortAssetID, now.Add(time.Minute), "approved_for_publish", stringPtr("quiet rooftop preview.")),
			}, nil
		},
		getMediaAsset: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			switch id {
			case pgUUID(mainAssetID):
				return testWorkspacePreviewAssetRow(mainAssetID, viewerUserID, now, int64Ptr(720000), workspacePreviewAssetReadyState), nil
			case pgUUID(shortAssetID):
				return testWorkspacePreviewAssetRow(shortAssetID, viewerUserID, now, int64Ptr(16000), workspacePreviewAssetReadyState), nil
			default:
				t.Fatalf("GetMediaAssetByID() id got %v", id)
				return sqlc.AppMediaAsset{}, nil
			}
		},
	})
	repo.delivery = delivery

	got, err := repo.GetWorkspacePreviewMainDetail(context.Background(), viewerUserID, mainID)
	if err != nil {
		t.Fatalf("GetWorkspacePreviewMainDetail() error = %v, want nil", err)
	}
	if got.Main.ID != mainID {
		t.Fatalf("GetWorkspacePreviewMainDetail() main id got %s want %s", got.Main.ID, mainID)
	}
	if got.Main.Title != "" {
		t.Fatalf("GetWorkspacePreviewMainDetail() title got %q want empty string", got.Main.Title)
	}
	if got.EntryShort.ID != shortID {
		t.Fatalf("GetWorkspacePreviewMainDetail() entry short id got %s want %s", got.EntryShort.ID, shortID)
	}
	if got.EntryShort.Title != "quiet rooftop preview" {
		t.Fatalf("GetWorkspacePreviewMainDetail() entry short title got %q want %q", got.EntryShort.Title, "quiet rooftop preview")
	}
	if got.Main.Media.URL == "" {
		t.Fatal("GetWorkspacePreviewMainDetail() media url = empty, want playback url")
	}
}

func TestGetWorkspacePreviewMainDetailReturnsNotFoundWhenMainMissing(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("4b4b4b4b-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("4c4c4c4c-1111-1111-1111-111111111111")
	otherMainID := uuid.MustParse("4d4d4d4d-1111-1111-1111-111111111111")
	mainAssetID := uuid.MustParse("4e4e4e4e-1111-1111-1111-111111111111")

	delivery := newWorkspacePreviewDelivery(t)
	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		listMainsByCreator: func(context.Context, pgtype.UUID) ([]sqlc.AppMain, error) {
			return []sqlc.AppMain{
				testWorkspacePreviewMainRecord(otherMainID, viewerUserID, mainAssetID, now, "approved_for_unlock", 1800, workspacePreviewCurrencyCode),
			}, nil
		},
	})
	repo.delivery = delivery

	_, err := repo.GetWorkspacePreviewMainDetail(context.Background(), viewerUserID, mainID)
	if !errors.Is(err, ErrWorkspacePreviewNotFound) {
		t.Fatalf("GetWorkspacePreviewMainDetail() error got %v want %v", err, ErrWorkspacePreviewNotFound)
	}
}

func TestGetWorkspacePreviewMainDetailReturnsNotFoundWithoutPreviewableEntryShort(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("2b2b2b2b-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("2c2c2c2c-1111-1111-1111-111111111111")
	shortID := uuid.MustParse("2d2d2d2d-1111-1111-1111-111111111111")
	mainAssetID := uuid.MustParse("2e2e2e2e-1111-1111-1111-111111111111")
	shortAssetID := uuid.MustParse("2f2f2f2f-1111-1111-1111-111111111111")

	delivery := newWorkspacePreviewDelivery(t)
	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		listMainsByCreator: func(context.Context, pgtype.UUID) ([]sqlc.AppMain, error) {
			return []sqlc.AppMain{
				testWorkspacePreviewMainRecord(mainID, viewerUserID, mainAssetID, now, "approved_for_unlock", 1800, workspacePreviewCurrencyCode),
			}, nil
		},
		listShortsByCreator: func(context.Context, pgtype.UUID) ([]sqlc.AppShort, error) {
			return []sqlc.AppShort{
				testWorkspacePreviewShortRowWithCaption(shortID, viewerUserID, mainID, shortAssetID, now.Add(time.Minute), "approved_for_publish", stringPtr("quiet rooftop preview.")),
			}, nil
		},
		getMediaAsset: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			switch id {
			case pgUUID(mainAssetID):
				return testWorkspacePreviewAssetRow(mainAssetID, viewerUserID, now, int64Ptr(720000), workspacePreviewAssetReadyState), nil
			case pgUUID(shortAssetID):
				return testWorkspacePreviewAssetRow(shortAssetID, viewerUserID, now, int64Ptr(16000), "uploaded"), nil
			default:
				t.Fatalf("GetMediaAssetByID() id got %v", id)
				return sqlc.AppMediaAsset{}, nil
			}
		},
	})
	repo.delivery = delivery

	_, err := repo.GetWorkspacePreviewMainDetail(context.Background(), viewerUserID, mainID)
	if !errors.Is(err, ErrWorkspacePreviewNotFound) {
		t.Fatalf("GetWorkspacePreviewMainDetail() error got %v want %v", err, ErrWorkspacePreviewNotFound)
	}
}

func TestWorkspacePreviewHelpers(t *testing.T) {
	t.Parallel()

	if got := resolveWorkspacePreviewPageLimit(0); got != DefaultWorkspacePreviewPageSize {
		t.Fatalf("resolveWorkspacePreviewPageLimit(0) got %d want %d", got, DefaultWorkspacePreviewPageSize)
	}
	if got := resolveWorkspacePreviewPageLimit(5); got != 5 {
		t.Fatalf("resolveWorkspacePreviewPageLimit(5) got %d want %d", got, 5)
	}

	now := time.Unix(1710000000, 0).UTC()
	mainID := uuid.MustParse("96666666-6666-6666-6666-666666666666")
	shortID := uuid.MustParse("97777777-7777-7777-7777-777777777777")
	items := []WorkspacePreviewMainItem{
		{
			CreatedAt:       now,
			DurationSeconds: 720,
			ID:              mainID,
			LeadShortID:     shortID,
			PriceJpy:        2200,
		},
		{
			CreatedAt:       now.Add(-time.Minute),
			DurationSeconds: 600,
			ID:              uuid.MustParse("98888888-8888-8888-8888-888888888888"),
			LeadShortID:     shortID,
			PriceJpy:        1800,
		},
	}

	firstPage, nextCursor := finalizeWorkspacePreviewMainPage(items, 1)
	if len(firstPage) != 1 {
		t.Fatalf("finalizeWorkspacePreviewMainPage() page len got %d want %d", len(firstPage), 1)
	}
	if nextCursor == nil {
		t.Fatal("finalizeWorkspacePreviewMainPage() nextCursor = nil, want non-nil")
	}
	if !nextCursor.CreatedAt.Equal(items[0].CreatedAt) || nextCursor.ID != items[0].ID {
		t.Fatalf("finalizeWorkspacePreviewMainPage() nextCursor got %#v want item[0] cursor", nextCursor)
	}

	fullPage, finalCursor := finalizeWorkspacePreviewMainPage(items, 2)
	if len(fullPage) != 2 {
		t.Fatalf("finalizeWorkspacePreviewMainPage() full page len got %d want %d", len(fullPage), 2)
	}
	if finalCursor != nil {
		t.Fatalf("finalizeWorkspacePreviewMainPage() finalCursor got %#v want nil", finalCursor)
	}

	assetID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
	cache := map[uuid.UUID]sqlc.AppMediaAsset{}
	callCount := 0

	firstAsset, err := resolveWorkspacePreviewMediaAsset(assetID, cache, func(uuid.UUID) (sqlc.AppMediaAsset, error) {
		callCount++
		return testWorkspacePreviewAssetRow(assetID, shortID, now, int64Ptr(15000), workspacePreviewAssetReadyState), nil
	})
	if err != nil {
		t.Fatalf("resolveWorkspacePreviewMediaAsset() first error = %v, want nil", err)
	}
	secondAsset, err := resolveWorkspacePreviewMediaAsset(assetID, cache, func(uuid.UUID) (sqlc.AppMediaAsset, error) {
		callCount++
		return sqlc.AppMediaAsset{}, errors.New("unexpected call")
	})
	if err != nil {
		t.Fatalf("resolveWorkspacePreviewMediaAsset() second error = %v, want nil", err)
	}
	if callCount != 1 {
		t.Fatalf("resolveWorkspacePreviewMediaAsset() callCount got %d want %d", callCount, 1)
	}
	if firstAsset.ID != secondAsset.ID {
		t.Fatalf("resolveWorkspacePreviewMediaAsset() cached asset id got %v want %v", secondAsset.ID, firstAsset.ID)
	}
}

func newWorkspacePreviewDelivery(t *testing.T) *media.Delivery {
	t.Helper()

	delivery, err := media.NewDelivery(media.DeliveryConfig{
		ShortPublicBaseURL:    "https://cdn.example.com/shorts",
		MainPrivateBucketName: "main-private-bucket",
	}, workspacePreviewSignerStub{})
	if err != nil {
		t.Fatalf("media.NewDelivery() error = %v, want nil", err)
	}

	return delivery
}

func testWorkspacePreviewMainRow(
	id uuid.UUID,
	creatorID uuid.UUID,
	mediaAssetID uuid.UUID,
	createdAt time.Time,
	state string,
	priceMinor int64,
	currencyCode string,
) sqlc.ListCreatorWorkspacePreviewMainsByCreatorUserIDRow {
	return sqlc.ListCreatorWorkspacePreviewMainsByCreatorUserIDRow{
		ID:            postgres.UUIDToPG(id),
		CreatorUserID: postgres.UUIDToPG(creatorID),
		MediaAssetID:  postgres.UUIDToPG(mediaAssetID),
		State:         state,
		PriceMinor:    priceMinor,
		CurrencyCode:  currencyCode,
		CreatedAt:     postgres.TimeToPG(&createdAt),
		UpdatedAt:     postgres.TimeToPG(&createdAt),
	}
}

func testWorkspacePreviewMainRecord(
	id uuid.UUID,
	creatorID uuid.UUID,
	mediaAssetID uuid.UUID,
	createdAt time.Time,
	state string,
	priceMinor int64,
	currencyCode string,
) sqlc.AppMain {
	return sqlc.AppMain{
		ID:            postgres.UUIDToPG(id),
		CreatorUserID: postgres.UUIDToPG(creatorID),
		MediaAssetID:  postgres.UUIDToPG(mediaAssetID),
		State:         state,
		PriceMinor:    priceMinor,
		CurrencyCode:  currencyCode,
		CreatedAt:     postgres.TimeToPG(&createdAt),
		UpdatedAt:     postgres.TimeToPG(&createdAt),
	}
}

func testWorkspacePreviewShortRow(
	id uuid.UUID,
	creatorID uuid.UUID,
	canonicalMainID uuid.UUID,
	mediaAssetID uuid.UUID,
	createdAt time.Time,
	state string,
) sqlc.AppShort {
	return sqlc.AppShort{
		ID:              postgres.UUIDToPG(id),
		CreatorUserID:   postgres.UUIDToPG(creatorID),
		CanonicalMainID: postgres.UUIDToPG(canonicalMainID),
		MediaAssetID:    postgres.UUIDToPG(mediaAssetID),
		State:           state,
		CreatedAt:       postgres.TimeToPG(&createdAt),
		UpdatedAt:       postgres.TimeToPG(&createdAt),
	}
}

func testWorkspacePreviewShortRowWithCaption(
	id uuid.UUID,
	creatorID uuid.UUID,
	canonicalMainID uuid.UUID,
	mediaAssetID uuid.UUID,
	createdAt time.Time,
	state string,
	caption *string,
) sqlc.AppShort {
	row := testWorkspacePreviewShortRow(id, creatorID, canonicalMainID, mediaAssetID, createdAt, state)
	row.Caption = postgres.TextToPG(caption)

	return row
}

func testWorkspacePreviewAssetRow(
	id uuid.UUID,
	creatorID uuid.UUID,
	createdAt time.Time,
	durationMS *int64,
	processingState string,
) sqlc.AppMediaAsset {
	return sqlc.AppMediaAsset{
		ID:              postgres.UUIDToPG(id),
		CreatorUserID:   postgres.UUIDToPG(creatorID),
		ProcessingState: processingState,
		DurationMs:      postgres.Int64ToPG(durationMS),
		CreatedAt:       postgres.TimeToPG(&createdAt),
		UpdatedAt:       postgres.TimeToPG(&createdAt),
	}
}

func int64Ptr(value int64) *int64 {
	return &value
}
