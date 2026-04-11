package creator

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestGetWorkspacePreviewShortDetail(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("11111111-aaaa-aaaa-aaaa-111111111111")
	shortID := uuid.MustParse("22222222-bbbb-bbbb-bbbb-222222222222")
	mainID := uuid.MustParse("33333333-cccc-cccc-cccc-333333333333")
	assetID := uuid.MustParse("44444444-dddd-dddd-dddd-444444444444")
	caption := "  blue tone の balcony preview。  "

	delivery := newWorkspacePreviewDelivery(t)
	repo := newRepository(repositoryStubQueries{
		getCapability: func(_ context.Context, gotUserID pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			if gotUserID != pgUUID(viewerUserID) {
				t.Fatalf("GetCreatorCapabilityByUserID() user got %v want %v", gotUserID, pgUUID(viewerUserID))
			}
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(_ context.Context, gotUserID pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			if gotUserID != pgUUID(viewerUserID) {
				t.Fatalf("GetCreatorProfileByUserID() user got %v want %v", gotUserID, pgUUID(viewerUserID))
			}
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		getShortByID: func(_ context.Context, gotShortID pgtype.UUID) (sqlc.AppShort, error) {
			if gotShortID != pgUUID(shortID) {
				t.Fatalf("GetShortByID() short got %v want %v", gotShortID, pgUUID(shortID))
			}

			row := testWorkspacePreviewShortRow(shortID, viewerUserID, mainID, assetID, now, "approved_for_publish")
			row.Caption = postgres.TextToPG(&caption)
			return row, nil
		},
		getMediaAsset: func(_ context.Context, id pgtype.UUID) (sqlc.AppMediaAsset, error) {
			if id != pgUUID(assetID) {
				t.Fatalf("GetMediaAssetByID() id got %v want %v", id, pgUUID(assetID))
			}
			return testWorkspacePreviewAssetRow(assetID, viewerUserID, now, int64Ptr(15000), workspacePreviewAssetReadyState), nil
		},
	})
	repo.delivery = delivery

	detail, err := repo.GetWorkspacePreviewShortDetail(context.Background(), viewerUserID, shortID)
	if err != nil {
		t.Fatalf("GetWorkspacePreviewShortDetail() error = %v, want nil", err)
	}
	if detail.Short.ID != shortID {
		t.Fatalf("detail.Short.ID got %s want %s", detail.Short.ID, shortID)
	}
	if detail.Short.CanonicalMainID != mainID {
		t.Fatalf("detail.Short.CanonicalMainID got %s want %s", detail.Short.CanonicalMainID, mainID)
	}
	if detail.Short.Caption != "blue tone の balcony preview。" {
		t.Fatalf("detail.Short.Caption got %q want trimmed caption", detail.Short.Caption)
	}
	if detail.Short.PreviewDurationSeconds != 15 {
		t.Fatalf("detail.Short.PreviewDurationSeconds got %d want %d", detail.Short.PreviewDurationSeconds, 15)
	}
	if !strings.Contains(detail.Short.Media.URL, "/shorts/"+shortID.String()+"/playback.mp4") {
		t.Fatalf("detail.Short.Media.URL got %q want short playback url", detail.Short.Media.URL)
	}
}

func TestGetWorkspacePreviewMainDetail(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("55555555-aaaa-aaaa-aaaa-555555555555")
	mainID := uuid.MustParse("66666666-bbbb-bbbb-bbbb-666666666666")
	entryShortID := uuid.MustParse("77777777-cccc-cccc-cccc-777777777777")
	mainAssetID := uuid.MustParse("88888888-dddd-dddd-dddd-888888888888")
	shortAssetID := uuid.MustParse("99999999-eeee-eeee-eeee-999999999999")
	entryCaption := " rooftop preview "

	delivery := newWorkspacePreviewDelivery(t)
	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		getMainByID: func(_ context.Context, gotMainID pgtype.UUID) (sqlc.AppMain, error) {
			if gotMainID != pgUUID(mainID) {
				t.Fatalf("GetMainByID() main got %v want %v", gotMainID, pgUUID(mainID))
			}
			return sqlc.AppMain{
				ID:            postgres.UUIDToPG(mainID),
				CreatorUserID: postgres.UUIDToPG(viewerUserID),
				MediaAssetID:  postgres.UUIDToPG(mainAssetID),
				PriceMinor:    2200,
				CurrencyCode:  workspacePreviewCurrencyCode,
				CreatedAt:     postgres.TimeToPG(&now),
				UpdatedAt:     postgres.TimeToPG(&now),
			}, nil
		},
		listShortsByCreator: func(_ context.Context, gotUserID pgtype.UUID) ([]sqlc.AppShort, error) {
			if gotUserID != pgUUID(viewerUserID) {
				t.Fatalf("ListShortsByCreatorUserID() user got %v want %v", gotUserID, pgUUID(viewerUserID))
			}
			row := testWorkspacePreviewShortRow(entryShortID, viewerUserID, mainID, shortAssetID, now.Add(time.Minute), "approved_for_publish")
			row.Caption = postgres.TextToPG(&entryCaption)
			return []sqlc.AppShort{row}, nil
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

	detail, err := repo.GetWorkspacePreviewMainDetail(context.Background(), viewerUserID, mainID)
	if err != nil {
		t.Fatalf("GetWorkspacePreviewMainDetail() error = %v, want nil", err)
	}
	if detail.Main.ID != mainID {
		t.Fatalf("detail.Main.ID got %s want %s", detail.Main.ID, mainID)
	}
	if detail.Main.PriceJpy != 2200 {
		t.Fatalf("detail.Main.PriceJpy got %d want %d", detail.Main.PriceJpy, 2200)
	}
	if detail.EntryShort.ID != entryShortID {
		t.Fatalf("detail.EntryShort.ID got %s want %s", detail.EntryShort.ID, entryShortID)
	}
	if detail.EntryShort.Caption != "rooftop preview" {
		t.Fatalf("detail.EntryShort.Caption got %q want trimmed caption", detail.EntryShort.Caption)
	}
	if !strings.Contains(detail.Main.Media.URL, "mains/"+mainID.String()+"/playback.mp4") {
		t.Fatalf("detail.Main.Media.URL got %q want main playback url", detail.Main.Media.URL)
	}
}

func TestUpdateWorkspaceMainPrice(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("aaaaaaaa-1111-1111-1111-aaaaaaaaaaaa")
	mainID := uuid.MustParse("bbbbbbbb-2222-2222-2222-bbbbbbbbbbbb")

	var gotArg sqlc.UpdateCreatorWorkspaceMainPriceParams

	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		updateWorkspaceMainPrice: func(_ context.Context, arg sqlc.UpdateCreatorWorkspaceMainPriceParams) (sqlc.AppMain, error) {
			gotArg = arg
			return sqlc.AppMain{ID: postgres.UUIDToPG(mainID)}, nil
		},
	})

	if err := repo.UpdateWorkspaceMainPrice(context.Background(), viewerUserID, mainID, 2600); err != nil {
		t.Fatalf("UpdateWorkspaceMainPrice() error = %v, want nil", err)
	}
	if gotArg.ID != pgUUID(mainID) {
		t.Fatalf("UpdateCreatorWorkspaceMainPrice() main got %v want %v", gotArg.ID, pgUUID(mainID))
	}
	if gotArg.CreatorUserID != pgUUID(viewerUserID) {
		t.Fatalf("UpdateCreatorWorkspaceMainPrice() creator got %v want %v", gotArg.CreatorUserID, pgUUID(viewerUserID))
	}
	if gotArg.PriceMinor != 2600 {
		t.Fatalf("UpdateCreatorWorkspaceMainPrice() price got %d want %d", gotArg.PriceMinor, 2600)
	}
}

func TestUpdateWorkspaceShortCaptionNormalizesBlankToNil(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("cccccccc-3333-3333-3333-cccccccccccc")
	shortID := uuid.MustParse("dddddddd-4444-4444-4444-dddddddddddd")
	blankCaption := "   "

	var gotArg sqlc.UpdateCreatorWorkspaceShortCaptionParams

	repo := newRepository(repositoryStubQueries{
		getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
			return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
		},
		getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
			return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
		},
		updateWorkspaceShortCaption: func(_ context.Context, arg sqlc.UpdateCreatorWorkspaceShortCaptionParams) (sqlc.AppShort, error) {
			gotArg = arg
			return sqlc.AppShort{ID: postgres.UUIDToPG(shortID)}, nil
		},
	})

	if err := repo.UpdateWorkspaceShortCaption(context.Background(), viewerUserID, shortID, &blankCaption); err != nil {
		t.Fatalf("UpdateWorkspaceShortCaption() error = %v, want nil", err)
	}
	if gotArg.ID != pgUUID(shortID) {
		t.Fatalf("UpdateCreatorWorkspaceShortCaption() short got %v want %v", gotArg.ID, pgUUID(shortID))
	}
	if gotArg.CreatorUserID != pgUUID(viewerUserID) {
		t.Fatalf("UpdateCreatorWorkspaceShortCaption() creator got %v want %v", gotArg.CreatorUserID, pgUUID(viewerUserID))
	}
	if gotArg.Caption.Valid {
		t.Fatalf("UpdateCreatorWorkspaceShortCaption() caption got %#v want invalid/null", gotArg.Caption)
	}
}
