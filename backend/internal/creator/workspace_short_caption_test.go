package creator

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

func TestUpdateWorkspaceShortCaption(t *testing.T) {
	t.Parallel()

	now := time.Unix(1710000000, 0).UTC()
	viewerUserID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mainID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	shortID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	mediaAssetID := uuid.MustParse("44444444-4444-4444-4444-444444444444")

	t.Run("updates caption for owner short", func(t *testing.T) {
		t.Parallel()

		var gotUpdateArg sqlc.UpdateShortCaptionParams
		delivery := newWorkspacePreviewDelivery(t)
		repo := newRepository(repositoryStubQueries{
			getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
				return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
			},
			getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
				return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
			},
			getShortByID: func(_ context.Context, gotShortID pgtype.UUID) (sqlc.AppShort, error) {
				if gotShortID != pgUUID(shortID) {
					t.Fatalf("GetShortByID() short id got %v want %v", gotShortID, pgUUID(shortID))
				}
				return testWorkspacePreviewShortRowWithCaption(shortID, viewerUserID, mainID, mediaAssetID, now, "approved_for_publish", stringPtr("old")), nil
			},
			getMediaAsset: func(_ context.Context, gotAssetID pgtype.UUID) (sqlc.AppMediaAsset, error) {
				if gotAssetID != pgUUID(mediaAssetID) {
					t.Fatalf("GetMediaAssetByID() asset id got %v want %v", gotAssetID, pgUUID(mediaAssetID))
				}
				return testWorkspacePreviewAssetRow(mediaAssetID, viewerUserID, now, int64Ptr(16000), workspacePreviewAssetReadyState), nil
			},
			updateShortCaption: func(_ context.Context, arg sqlc.UpdateShortCaptionParams) (sqlc.AppShort, error) {
				gotUpdateArg = arg
				return testWorkspacePreviewShortRowWithCaption(shortID, viewerUserID, mainID, mediaAssetID, now, "approved_for_publish", stringPtr("updated caption")), nil
			},
		})
		repo.delivery = delivery

		result, err := repo.UpdateWorkspaceShortCaption(context.Background(), viewerUserID, shortID, "  updated caption  ")
		if err != nil {
			t.Fatalf("UpdateWorkspaceShortCaption() error = %v, want nil", err)
		}
		if result.ShortID != shortID {
			t.Fatalf("UpdateWorkspaceShortCaption() short id got %s want %s", result.ShortID, shortID)
		}
		if result.Caption != "updated caption" {
			t.Fatalf("UpdateWorkspaceShortCaption() caption got %q want %q", result.Caption, "updated caption")
		}
		if gotUpdateArg.ID != pgUUID(shortID) {
			t.Fatalf("UpdateWorkspaceShortCaption() id arg got %v want %v", gotUpdateArg.ID, pgUUID(shortID))
		}
		if gotUpdateArg.Caption != postgres.TextToPG(stringPtr("updated caption")) {
			t.Fatalf(
				"UpdateWorkspaceShortCaption() caption arg got %#v want %#v",
				gotUpdateArg.Caption,
				postgres.TextToPG(stringPtr("updated caption")),
			)
		}
	})

	t.Run("normalizes blank caption to null and empty string", func(t *testing.T) {
		t.Parallel()

		var gotUpdateArg sqlc.UpdateShortCaptionParams
		delivery := newWorkspacePreviewDelivery(t)
		repo := newRepository(repositoryStubQueries{
			getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
				return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
			},
			getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
				return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
			},
			getShortByID: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
				return testWorkspacePreviewShortRowWithCaption(shortID, viewerUserID, mainID, mediaAssetID, now, "approved_for_publish", stringPtr("old")), nil
			},
			getMediaAsset: func(context.Context, pgtype.UUID) (sqlc.AppMediaAsset, error) {
				return testWorkspacePreviewAssetRow(mediaAssetID, viewerUserID, now, int64Ptr(16000), workspacePreviewAssetReadyState), nil
			},
			updateShortCaption: func(_ context.Context, arg sqlc.UpdateShortCaptionParams) (sqlc.AppShort, error) {
				gotUpdateArg = arg
				return testWorkspacePreviewShortRowWithCaption(shortID, viewerUserID, mainID, mediaAssetID, now, "approved_for_publish", nil), nil
			},
		})
		repo.delivery = delivery

		result, err := repo.UpdateWorkspaceShortCaption(context.Background(), viewerUserID, shortID, "   ")
		if err != nil {
			t.Fatalf("UpdateWorkspaceShortCaption() error = %v, want nil", err)
		}
		if gotUpdateArg.Caption != postgres.TextToPG(nil) {
			t.Fatalf("UpdateWorkspaceShortCaption() blank caption arg got %#v want null", gotUpdateArg.Caption)
		}
		if result.Caption != "" {
			t.Fatalf("UpdateWorkspaceShortCaption() blank result caption got %q want empty string", result.Caption)
		}
	})

	t.Run("maps non-owner short to preview not found", func(t *testing.T) {
		t.Parallel()

		delivery := newWorkspacePreviewDelivery(t)
		repo := newRepository(repositoryStubQueries{
			getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
				return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
			},
			getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
				return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
			},
			getShortByID: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
				return testWorkspacePreviewShortRow(shortID, uuid.MustParse("55555555-5555-5555-5555-555555555555"), mainID, mediaAssetID, now, "approved_for_publish"), nil
			},
		})
		repo.delivery = delivery

		_, err := repo.UpdateWorkspaceShortCaption(context.Background(), viewerUserID, shortID, "updated")
		if !errors.Is(err, ErrWorkspacePreviewNotFound) {
			t.Fatalf("UpdateWorkspaceShortCaption() error got %v want %v", err, ErrWorkspacePreviewNotFound)
		}
	})

	t.Run("maps owned but not previewable short to preview not found", func(t *testing.T) {
		t.Parallel()

		updateCalled := false
		delivery := newWorkspacePreviewDelivery(t)
		repo := newRepository(repositoryStubQueries{
			getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
				return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
			},
			getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
				return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
			},
			getShortByID: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
				return testWorkspacePreviewShortRowWithCaption(shortID, viewerUserID, mainID, mediaAssetID, now, "approved_for_publish", stringPtr("old")), nil
			},
			getMediaAsset: func(context.Context, pgtype.UUID) (sqlc.AppMediaAsset, error) {
				return testWorkspacePreviewAssetRow(mediaAssetID, viewerUserID, now, int64Ptr(16000), "uploaded"), nil
			},
			updateShortCaption: func(context.Context, sqlc.UpdateShortCaptionParams) (sqlc.AppShort, error) {
				updateCalled = true
				return sqlc.AppShort{}, nil
			},
		})
		repo.delivery = delivery

		_, err := repo.UpdateWorkspaceShortCaption(context.Background(), viewerUserID, shortID, "updated")
		if !errors.Is(err, ErrWorkspacePreviewNotFound) {
			t.Fatalf("UpdateWorkspaceShortCaption() error got %v want %v", err, ErrWorkspacePreviewNotFound)
		}
		if updateCalled {
			t.Fatal("UpdateWorkspaceShortCaption() updateCalled = true, want false")
		}
	})

	t.Run("maps missing short to preview not found", func(t *testing.T) {
		t.Parallel()

		delivery := newWorkspacePreviewDelivery(t)
		repo := newRepository(repositoryStubQueries{
			getCapability: func(context.Context, pgtype.UUID) (sqlc.AppCreatorCapability, error) {
				return testCapabilityRow(viewerUserID, now, nil, nil, nil, nil, timePtr(now), nil, nil), nil
			},
			getProfile: func(context.Context, pgtype.UUID) (sqlc.AppCreatorProfile, error) {
				return testProfileRow(viewerUserID, now, stringPtr("Mina Rei"), stringPtr("minarei"), nil, nil), nil
			},
			getShortByID: func(context.Context, pgtype.UUID) (sqlc.AppShort, error) {
				return sqlc.AppShort{}, pgx.ErrNoRows
			},
		})
		repo.delivery = delivery

		_, err := repo.UpdateWorkspaceShortCaption(context.Background(), viewerUserID, shortID, "updated")
		if !errors.Is(err, ErrWorkspacePreviewNotFound) {
			t.Fatalf("UpdateWorkspaceShortCaption() error got %v want %v", err, ErrWorkspacePreviewNotFound)
		}
	})
}
